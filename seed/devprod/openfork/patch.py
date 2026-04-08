"""Parser for git format-patch files.

A format-patch file (exported with --no-stat --no-signature) has the shape:

    From <sha> <date>
    From: <author>
    Subject: [PATCH] ...
    ...                           # rfc2822-style header lines
    <blank line>                  # header/body separator
    <commit message>
    ---                           # end-of-message marker emitted by format-patch
    diff --git a/... b/...        # one block per changed file
      <mode lines>                # e.g. "new file mode", "index abcd..ef01"
      --- a/... | /dev/null
      +++ b/... | /dev/null
      @@ hunk @@
      ... hunk lines ...
    diff --git ...                # repeats
    ...
    <trailing blank line>

The parser splits on that first blank line, then walks the body looking for
'diff --git' lines to cut it into per-file Diff blocks.
"""

import logging
import re

logger = logging.getLogger(__name__)


class Diff:
    """A single diff --git block within a patch."""

    @staticmethod
    def _strip_path(prefix: str, line: str) -> str:
        """Strip a prefix like 'a/' or 'b/' from a path, handling optional quotes."""
        path = line.strip('"')
        if not path.startswith(prefix):
            raise RuntimeError(f"expected prefix {prefix!r}: {line}")
        return path[len(prefix) :]

    @staticmethod
    def _needs_quoting(path: str) -> bool:
        # Mirrors git's default core.quotePath=true rule from quote.c: wrap in
        # double quotes when any byte is a control char, DEL, high-bit, `"`,
        # or `\`. Without this, unconditionally quoting breaks round-trip
        # equality (parse -> render on a plain ASCII path should be a no-op).
        for ch in path:
            o = ord(ch)
            if o < 0x20 or o == 0x22 or o == 0x5c or o == 0x7f or o >= 0x80:
                return True
        return False

    @staticmethod
    def _format_path(prefix: str, path: str) -> str:
        """Render '<prefix><path>', wrapped in quotes iff git would quote it."""
        if Diff._needs_quoting(path):
            return f'"{prefix}{path}"'
        return f"{prefix}{path}"

    @staticmethod
    def _parse_diff_line(line: str) -> tuple[str, str]:
        """Parse 'diff --git a/... b/...' into (a_path, b_path)."""
        if not line.startswith("diff --git "):
            raise RuntimeError(f"no diff --git header found: {line}")
        rest = line[len("diff --git ") :]
        # Paths may be independently quoted (for spaces / non-ASCII).
        # Cases: a/x b/y | "a/x" "b/y" | "a/x" b/y | a/x "b/y"
        if rest.startswith('"'):
            close_a = rest.find('"', 1)
            if close_a == -1:
                raise RuntimeError(f"cannot parse diff --git line: {line}")
            a_raw = rest[: close_a + 1]
            b_raw = rest[close_a + 2 :]
        else:
            space = rest.find(" ")
            if space == -1:
                raise RuntimeError(f"cannot parse diff --git line: {line}")
            a_raw = rest[:space]
            b_raw = rest[space + 1 :]
        return Diff._strip_path("a/", a_raw), Diff._strip_path("b/", b_raw)

    diff_line: str
    mode_lines: list[str]
    binary: bool

    # Text diff fields
    minus_line: str
    plus_line: str
    hunk_lines: list[str]

    # Binary diff fields
    binary_lines: list[str]

    a: str
    b: str

    def _parse(self, lines: list[str]) -> None:
        """Parse a diff block into its components and verify paths.

        Walks the block linearly:
          diff --git ...        (lines[0], already consumed for path)
          <mode lines>          (index / mode / similarity / rename from|to / ...)
          <then either>
              GIT binary patch  -> binary_lines holds the rest
          <or>
              --- <path>        -> minus_line
              +++ <path>        -> plus_line
              @@ ... @@         -> start of hunk_lines
        """
        self.diff_line = lines[0]
        self.a, self.b = self._parse_diff_line(self.diff_line)

        # Mode lines (index, new/deleted file mode, similarity index, etc.)
        self.mode_lines = []
        self.binary = False
        self.minus_line = ""
        self.plus_line = ""
        self.hunk_lines = []
        self.binary_lines = []

        # Collect mode lines up to the first '--- ' or 'GIT binary patch' marker.
        i = 1
        while (
            i < len(lines)
            and not lines[i].startswith("--- ")
            and not lines[i].startswith("GIT binary patch")
        ):
            self.mode_lines.append(lines[i])
            i += 1

        # Binary diff: git emits 'GIT binary patch' followed by base85 blocks.
        # We don't parse the body — preserve it verbatim for faithful re-rendering.
        if i < len(lines) and lines[i].startswith("GIT binary patch"):
            self.binary = True
            self.binary_lines = lines[i:]
            return

        # Text diff: --- and +++ lines (may be /dev/null or quoted for non-ASCII).
        # Cross-check that the paths on '---'/'+++' agree with the 'diff --git'
        # header, except when the side is /dev/null (add/delete).
        if i < len(lines) and lines[i].startswith("--- "):
            self.minus_line = lines[i]
            if self.minus_line != "--- /dev/null":
                minus_path = self._strip_path("a/", self.minus_line[len("--- ") :])
                if minus_path != self.a:
                    raise RuntimeError(f"--- path mismatch: {minus_path} != {self.a}")
            i += 1
        if i < len(lines) and lines[i].startswith("+++ "):
            self.plus_line = lines[i]
            if self.plus_line != "+++ /dev/null":
                plus_path = self._strip_path("b/", self.plus_line[len("+++ ") :])
                if plus_path != self.b:
                    raise RuntimeError(f"+++ path mismatch: {plus_path} != {self.b}")
            i += 1

        # Everything left is @@-hunks. Keep them as raw lines — editors like
        # replace_hunk_lines operate on these in place.
        self.hunk_lines = lines[i:]

    def __init__(self, lines: list[str]) -> None:
        self._parse(lines)

    def render(self) -> str:
        # Re-emit the block in the same order _parse consumed it. The trailing
        # "\n" guarantees each Diff ends on a line boundary when concatenated
        # in Patch.render.
        result = [self.diff_line]
        result.extend(self.mode_lines)
        if self.binary:
            result.extend(self.binary_lines)
        else:
            if self.minus_line:
                result.append(self.minus_line)
            if self.plus_line:
                result.append(self.plus_line)
            result.extend(self.hunk_lines)
        return "\n".join(result) + "\n"

    def match_a(self, a_patterns: list[re.Pattern[str]]) -> bool:
        return any(p.match(self.a) for p in a_patterns)

    def match_b(self, b_patterns: list[re.Pattern[str]]) -> bool:
        return any(p.match(self.b) for p in b_patterns)

    def replace_hunk_lines(
        self, line_symbol: str, pattern: re.Pattern[str], replace: str
    ) -> None:
        # Hunk lines are prefixed by ' ' (context), '-' (removed), '+' (added),
        # or '\' (e.g. "\ No newline at end of file"). We only rewrite lines
        # whose prefix matches line_symbol, and we pattern-match on the content
        # *after* the prefix so the caller's regex sees the real source text.
        #
        # A newline in `replace` would split one hunk line into two without
        # updating the @@ header counts, which git am rejects as "corrupt
        # patch" — fail loudly here instead of letting it through.
        if "\n" in replace:
            raise ValueError("replace must not contain newlines")
        for i, line in enumerate(self.hunk_lines):
            if line[:1] == line_symbol and pattern.search(line[1:]):
                self.hunk_lines[i] = line_symbol + pattern.sub(replace, line[1:])


class Patch:
    """Parsed representation of a git format-patch file."""

    name: str
    header_lines: list[str]
    message_lines: list[str]
    diffs: list[Diff]

    def __init__(self, name: str, text: str) -> None:
        self.name = name
        # Strip a single trailing newline so split() doesn't yield a spurious
        # empty tail line. The blank separator between sections is still
        # preserved inside `text`.
        if text.endswith("\n"):
            text = text[:-1]

        # The first blank line separates the rfc2822-style header from the
        # body (commit message + diffs). header_text excludes that blank.
        header_text = text.split("\n\n", 1)[0]
        self.header_lines = header_text.strip().splitlines()

        # Body: everything after the blank separator. Use split("\n") — NOT
        # splitlines() — because diff hunks can contain \r (CRLF source files),
        # \v, \f, and Unicode line separators as DATA, and splitlines() would
        # incorrectly break those lines apart.
        body_lines = text[len(header_text) + 1 :].split("\n")

        self.diffs = []

        # Every per-file block starts with a 'diff --git ' line. Collect their
        # indices so we can slice the body into (message, diff-1, diff-2, ...).
        diff_starts = [
            i for i, line in enumerate(body_lines) if line.startswith("diff --git ")
        ]
        # Commit message occupies the body up to the first diff. If there are
        # no diffs (empty commit), the whole body is the message.
        message_lines = body_lines[: diff_starts[0]] if diff_starts else body_lines
        # Collapse into a stripped, normalized list — the message is small and
        # doesn't carry binary content, so splitlines() is safe here.
        self.message_lines = "\n".join(message_lines).strip().splitlines()

        if not diff_starts:
            return

        # Slice each [start, next_start) range into its own Diff block.
        self.diffs = []
        for j, start in enumerate(diff_starts):
            end = diff_starts[j + 1] if j + 1 < len(diff_starts) else len(body_lines)
            self.diffs.append(Diff(body_lines[start:end]))

    def pick_diff(
        self, a_patterns: list[re.Pattern[str]], b_patterns: list[re.Pattern[str]]
    ) -> None:
        """Keep only file diffs matching any pattern, remove the rest."""
        kept: list[Diff] = []
        for d in self.diffs:
            ma = d.match_a(a_patterns)
            mb = d.match_b(b_patterns)
            if ma and mb:
                kept.append(d)
            else:
                if d.a == d.b:
                    logger.info("%s: Dropped %r", self.name, d.a)
                else:
                    if ma and not mb:
                        logger.warning("%s: Dropped %r -> [%r]", self.name, d.a, d.b)
                    elif not ma and mb:
                        logger.warning("%s: Dropped [%r] -> %r", self.name, d.a, d.b)
                    else:
                        logger.info("%s: Dropped %r -> %r", self.name, d.a, d.b)
        self.diffs = kept

    def pick_file(self, patterns: list[re.Pattern[str]]) -> None:
        self.pick_diff(patterns, patterns)

    def drop_diff(
        self, a_patterns: list[re.Pattern[str]], b_patterns: list[re.Pattern[str]]
    ) -> None:
        """Remove file diffs matching any pattern in-place."""
        kept: list[Diff] = []
        for d in self.diffs:
            ma = d.match_a(a_patterns)
            mb = d.match_b(b_patterns)
            if ma and mb:
                if d.a == d.b:
                    logger.info("%s: Dropped %r", self.name, d.a)
                else:
                    logger.info("%s: Dropped %r -> %r", self.name, d.a, d.b)
            else:
                kept.append(d)
        self.diffs = kept

    def drop_file(self, patterns: list[re.Pattern[str]]) -> None:
        self.drop_diff(patterns, patterns)

    def move_diff(
        self,
        a_match_pattern: re.Pattern[str],
        b_match_pattern: re.Pattern[str],
        a_replace: str,
        b_replace: str,
    ) -> None:
        """Move file diffs matching any pattern in-place."""
        kept: list[Diff] = []
        for d in self.diffs:
            ma = d.match_a([a_match_pattern])
            mb = d.match_b([b_match_pattern])
            if ma and mb:
                new_a = a_match_pattern.sub(a_replace, d.a)
                new_b = b_match_pattern.sub(b_replace, d.b)
                if d.a == d.b and new_a == new_b:
                    logger.info("%s: Moved %r => %r", self.name, d.a, new_a)
                else:
                    logger.info(
                        "%s: Moved %r -> %r => %r -> %r",
                        self.name,
                        d.a,
                        d.b,
                        new_a,
                        new_b,
                    )
                d.diff_line = (
                    f"diff --git {Diff._format_path('a/', new_a)}"
                    f" {Diff._format_path('b/', new_b)}"
                )
                if d.minus_line and d.minus_line != "--- /dev/null":
                    d.minus_line = f"--- {Diff._format_path('a/', new_a)}"
                if d.plus_line and d.plus_line != "+++ /dev/null":
                    d.plus_line = f"+++ {Diff._format_path('b/', new_b)}"
                d.mode_lines = [
                    (
                        f"rename from {Diff._format_path('', new_a)}"
                        if ml.startswith("rename from ")
                        else f"rename to {Diff._format_path('', new_b)}"
                        if ml.startswith("rename to ")
                        else ml
                    )
                    for ml in d.mode_lines
                ]
                d.a = new_a
                d.b = new_b
            kept.append(d)
        self.diffs = kept

    def move_file(
        self,
        match_pattern: re.Pattern[str],
        replace: str,
    ) -> None:
        self.move_diff(match_pattern, match_pattern, replace, replace)

    def replace_diff_hunk_lines(
        self,
        a_patterns: list[re.Pattern[str]],
        b_patterns: list[re.Pattern[str]],
        search_patterns: list[re.Pattern[str]],
        replace: str,
    ) -> None:
        """Replace lines in file diffs matching any pattern in-place."""
        kept: list[Diff] = []
        for d in self.diffs:
            ma = d.match_a(a_patterns)
            mb = d.match_b(b_patterns)
            if ma and mb:
                for search_pattern in search_patterns:
                    d.replace_hunk_lines(" ", search_pattern, replace)
            if ma:
                for search_pattern in search_patterns:
                    d.replace_hunk_lines("-", search_pattern, replace)
            if mb:
                for search_pattern in search_patterns:
                    d.replace_hunk_lines("+", search_pattern, replace)
            kept.append(d)
        self.diffs = kept

    def replace_hunk_lines(
        self,
        path_patterns: list[re.Pattern[str]],
        search_patterns: list[re.Pattern[str]],
        replace: str,
    ) -> None:
        self.replace_diff_hunk_lines(
            path_patterns, path_patterns, search_patterns, replace
        )

    def render(self) -> str:
        # Reassemble as:  header\n  \n  message\n  \n  diff-blocks
        # The two blank lines match git's own format-patch output: one after
        # the header and one separating the message from the first diff.
        result = (
            ("\n".join(self.header_lines) + "\n")
            + "\n"
            + ("\n".join(self.message_lines) + "\n" if self.message_lines else "")
            + "\n"
            + ("".join([d.render() for d in self.diffs]))
        )
        return result
