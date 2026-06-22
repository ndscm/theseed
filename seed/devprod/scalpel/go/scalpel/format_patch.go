// Package scalpel parses git format-patch files into editable structures so
// individual file diffs can be picked, dropped, moved, or rewritten before
// the patch is re-rendered and re-applied.
//
// A format-patch file (exported with --no-stat --no-signature) has the shape:
//
//	From <sha> <date>
//	From: <author>
//	Subject: [PATCH] ...
//	...                           # rfc2822-style header lines
//	<blank line>                  # header/body separator
//	<commit message>
//	---                           # end-of-message marker emitted by format-patch
//	diff --git a/... b/...        # one block per changed file
//	  <mode lines>                # e.g. "new file mode", "index abcd..ef01"
//	  --- a/... | /dev/null
//	  +++ b/... | /dev/null
//	  @@ hunk @@
//	  ... hunk lines ...
//	diff --git ...                # repeats
//	...
//	<trailing blank line>
//
// The parser splits on that first blank line, then walks the body looking for
// 'diff --git' lines to cut it into per-file Diff blocks.
package scalpel

import (
	"regexp"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// needsQuoting mirrors git's default core.quotePath=true rule from quote.c:
// wrap in double quotes when any byte is a control char, DEL, high-bit, `"`, or
// `\`. Without this, unconditionally quoting breaks round-trip equality (parse
// -> render on a plain ASCII path should be a no-op).
func needsQuoting(path string) bool {
	for _, ch := range path {
		if ch < 0x20 || ch == 0x22 || ch == 0x5c || ch == 0x7f || ch >= 0x80 {
			return true
		}
	}
	return false
}

// stripPath strips a prefix like "a/" or "b/" from a path, handling optional
// surrounding quotes.
func stripPath(prefix string, line string) (string, error) {
	path := strings.Trim(line, `"`)
	if !strings.HasPrefix(path, prefix) {
		return "", seederr.WrapErrorf("expected prefix %q: %v", prefix, line)
	}
	return path[len(prefix):], nil
}

// formatPath renders "<prefix><path>", wrapped in quotes iff git would quote it.
func formatPath(prefix string, path string) string {
	if needsQuoting(path) {
		return `"` + prefix + path + `"`
	}
	return prefix + path
}

// parseDiffLine parses "diff --git a/... b/..." into (aPath, bPath).
func parseDiffLine(line string) (string, string, error) {
	const header = "diff --git "
	if !strings.HasPrefix(line, header) {
		return "", "", seederr.WrapErrorf("no diff --git header found: %v", line)
	}
	rest := line[len(header):]
	// Paths may be independently quoted (for spaces / non-ASCII).
	// Cases: a/x b/y | "a/x" "b/y" | "a/x" b/y | a/x "b/y"
	aRaw := ""
	bRaw := ""
	if strings.HasPrefix(rest, `"`) {
		closeA := strings.IndexByte(rest[1:], '"')
		if closeA == -1 {
			return "", "", seederr.WrapErrorf("cannot parse diff --git line: %v", line)
		}
		closeA++ // shift back into rest's index space
		aRaw = rest[:closeA+1]
		bRaw = rest[closeA+2:]
	} else {
		space := strings.IndexByte(rest, ' ')
		if space == -1 {
			return "", "", seederr.WrapErrorf("cannot parse diff --git line: %v", line)
		}
		aRaw = rest[:space]
		bRaw = rest[space+1:]
	}
	a, err := stripPath("a/", aRaw)
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	b, err := stripPath("b/", bRaw)
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	return a, b, nil
}

// FormatDiff is a single "diff --git" block within a patch.
type FormatDiff struct {
	diffLine string

	a string
	b string

	modeLines []string
	binary    bool

	// Text diff fields.
	minusLine string
	plusLine  string
	hunkLines []string

	// Binary diff fields.
	binaryLines []string
}

func (d *FormatDiff) matchA(aPatterns []*regexp.Regexp) bool {
	for _, p := range aPatterns {
		if p.MatchString(d.a) {
			return true
		}
	}
	return false
}

func (d *FormatDiff) matchB(bPatterns []*regexp.Regexp) bool {
	for _, p := range bPatterns {
		if p.MatchString(d.b) {
			return true
		}
	}
	return false
}

// render re-emits the block in the same order parseDiff consumed it. The
// trailing "\n" guarantees each Diff ends on a line boundary when concatenated
// in Patch.Render.
func (d *FormatDiff) render() string {
	result := []string{d.diffLine}
	result = append(result, d.modeLines...)
	if d.binary {
		result = append(result, d.binaryLines...)
	} else {
		if d.minusLine != "" {
			result = append(result, d.minusLine)
		}
		if d.plusLine != "" {
			result = append(result, d.plusLine)
		}
		result = append(result, d.hunkLines...)
	}
	return strings.Join(result, "\n") + "\n"
}

// parseFormatDiff parses a diff block into its components and verifies paths.
//
// Walks the block linearly:
//
//	diff --git ...        (lines[0], also consumed for paths)
//	<mode lines>          (index / mode / similarity / rename from|to / ...)
//	<then either>
//	    GIT binary patch  -> binaryLines holds the rest
//	<or>
//	    --- <path>        -> minusLine
//	    +++ <path>        -> plusLine
//	    @@ ... @@         -> start of hunkLines
func parseFormatDiff(lines []string) (*FormatDiff, error) {
	d := &FormatDiff{}
	d.diffLine = lines[0]
	a, b, err := parseDiffLine(d.diffLine)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	d.a = a
	d.b = b

	// Collect mode lines up to the first '--- ' or 'GIT binary patch' marker.
	i := 1
	for i < len(lines) &&
		!strings.HasPrefix(lines[i], "--- ") &&
		!strings.HasPrefix(lines[i], "GIT binary patch") {
		d.modeLines = append(d.modeLines, lines[i])
		i++
	}

	// Binary diff: git emits 'GIT binary patch' followed by base85 blocks. We
	// don't parse the body — preserve it verbatim for faithful re-rendering.
	if i < len(lines) && strings.HasPrefix(lines[i], "GIT binary patch") {
		d.binary = true
		d.binaryLines = append(d.binaryLines, lines[i:]...)
		return d, nil
	}

	// Text diff: --- and +++ lines (may be /dev/null or quoted for non-ASCII).
	// Cross-check that the paths on '---'/'+++' agree with the 'diff --git'
	// header, except when the side is /dev/null (add/delete).
	if i < len(lines) && strings.HasPrefix(lines[i], "--- ") {
		d.minusLine = lines[i]
		if d.minusLine != "--- /dev/null" {
			minusPath, err := stripPath("a/", d.minusLine[len("--- "):])
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			if minusPath != d.a {
				return nil, seederr.WrapErrorf("--- path mismatch: %v != %v", minusPath, d.a)
			}
		}
		i++
	}
	if i < len(lines) && strings.HasPrefix(lines[i], "+++ ") {
		d.plusLine = lines[i]
		if d.plusLine != "+++ /dev/null" {
			plusPath, err := stripPath("b/", d.plusLine[len("+++ "):])
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			if plusPath != d.b {
				return nil, seederr.WrapErrorf("+++ path mismatch: %v != %v", plusPath, d.b)
			}
		}
		i++
	}

	// Everything left is @@-hunks. Keep them as raw lines — editors like
	// replaceHunkLines operate on these in place.
	d.hunkLines = append(d.hunkLines, lines[i:]...)
	return d, nil
}

// FormatPatch is a parsed representation of a git format-patch file.
type FormatPatch struct {
	name         string
	headerLines  []string
	messageLines []string
	diffs        []*FormatDiff
}

// PickDiff keeps only file diffs matching any pattern, removing the rest.
func (p *FormatPatch) PickDiff(aPatterns []*regexp.Regexp, bPatterns []*regexp.Regexp) {
	kept := []*FormatDiff{}
	for _, d := range p.diffs {
		ma := d.matchA(aPatterns)
		mb := d.matchB(bPatterns)
		if ma && mb {
			kept = append(kept, d)
			continue
		}
		switch {
		case d.a == d.b:
			seedlog.Infof("%v: Dropped %q", p.name, d.a)
		case ma && !mb:
			seedlog.Warnf("%v: Dropped %q -> [%q]", p.name, d.a, d.b)
		case !ma && mb:
			seedlog.Warnf("%v: Dropped [%q] -> %q", p.name, d.a, d.b)
		default:
			seedlog.Infof("%v: Dropped %q -> %q", p.name, d.a, d.b)
		}
	}
	p.diffs = kept
}

// DropDiff removes file diffs matching any pattern in place.
func (p *FormatPatch) DropDiff(aPatterns []*regexp.Regexp, bPatterns []*regexp.Regexp) {
	kept := []*FormatDiff{}
	for _, d := range p.diffs {
		ma := d.matchA(aPatterns)
		mb := d.matchB(bPatterns)
		if !ma || !mb {
			kept = append(kept, d)
			continue
		}
		if d.a == d.b {
			seedlog.Infof("%v: Dropped %q", p.name, d.a)
		} else {
			seedlog.Infof("%v: Dropped %q -> %q", p.name, d.a, d.b)
		}
	}
	p.diffs = kept
}

// MoveDiff rewrites the paths of file diffs matching both patterns in place.
func (p *FormatPatch) MoveDiff(
	aMatchPattern *regexp.Regexp,
	bMatchPattern *regexp.Regexp,
	aReplace string,
	bReplace string,
) {
	for _, d := range p.diffs {
		ma := d.matchA([]*regexp.Regexp{aMatchPattern})
		mb := d.matchB([]*regexp.Regexp{bMatchPattern})
		if !ma || !mb {
			continue
		}
		newA := aMatchPattern.ReplaceAllString(d.a, aReplace)
		newB := bMatchPattern.ReplaceAllString(d.b, bReplace)
		if d.a == d.b && newA == newB {
			seedlog.Infof("%v: Moved %q => %q", p.name, d.a, newA)
		} else {
			seedlog.Infof("%v: Moved %q -> %q => %q -> %q", p.name, d.a, d.b, newA, newB)
		}
		d.diffLine = "diff --git " + formatPath("a/", newA) + " " + formatPath("b/", newB)
		if d.minusLine != "" && d.minusLine != "--- /dev/null" {
			d.minusLine = "--- " + formatPath("a/", newA)
		}
		if d.plusLine != "" && d.plusLine != "+++ /dev/null" {
			d.plusLine = "+++ " + formatPath("b/", newB)
		}
		for i, ml := range d.modeLines {
			if strings.HasPrefix(ml, "rename from ") {
				d.modeLines[i] = "rename from " + formatPath("", newA)
			} else if strings.HasPrefix(ml, "rename to ") {
				d.modeLines[i] = "rename to " + formatPath("", newB)
			}
		}
		d.a = newA
		d.b = newB
	}
}

// Render reassembles the patch as:  header\n  \n  message\n  \n  diff-blocks.
// The two blank lines match git's own format-patch output: one after the
// header and one separating the message from the first diff.
func (p *FormatPatch) Render() string {
	result := strings.Builder{}
	result.WriteString(strings.Join(p.headerLines, "\n"))
	result.WriteString("\n")
	result.WriteString("\n")
	if len(p.messageLines) > 0 {
		result.WriteString(strings.Join(p.messageLines, "\n"))
		result.WriteString("\n")
	}
	result.WriteString("\n")
	for _, d := range p.diffs {
		result.WriteString(d.render())
	}
	return result.String()
}

// ParseFormatPatch parses a format-patch file. name is used only for log messages.
func ParseFormatPatch(name string, text string) (*FormatPatch, error) {
	// Strip a single trailing newline so the split below doesn't yield a
	// spurious empty tail line. The blank separator between sections is still
	// preserved inside text.
	text = strings.TrimSuffix(text, "\n")

	// The first blank line separates the rfc2822-style header from the body
	// (commit message + diffs). headerText excludes that blank.
	headerText := text
	if idx := strings.Index(text, "\n\n"); idx >= 0 {
		headerText = text[:idx]
	}
	p := &FormatPatch{name: name}
	header := strings.TrimSpace(headerText)
	if header != "" {
		p.headerLines = strings.Split(header, "\n")
	}

	// Body: everything after the blank separator. Use strings.Split — NOT a
	// splitlines-equivalent — because diff hunks can contain \r (CRLF source
	// files), \v, \f, and Unicode line separators as DATA, and splitting on
	// those would incorrectly break those lines apart.
	bodyText := ""
	if bodyStart := len(headerText) + 1; bodyStart <= len(text) {
		bodyText = text[bodyStart:]
	}
	bodyLines := strings.Split(bodyText, "\n")

	// Every per-file block starts with a 'diff --git ' line. Collect their
	// indices so we can slice the body into (message, diff-1, diff-2, ...).
	diffStarts := []int{}
	for i, line := range bodyLines {
		if strings.HasPrefix(line, "diff --git ") {
			diffStarts = append(diffStarts, i)
		}
	}

	// Commit message occupies the body up to the first diff. If there are no
	// diffs (empty commit), the whole body is the message.
	messageEnd := len(bodyLines)
	if len(diffStarts) > 0 {
		messageEnd = diffStarts[0]
	}
	// Collapse into a stripped, normalized list — the message is small and
	// doesn't carry binary content, so splitting on newlines is safe here.
	message := strings.TrimSpace(strings.Join(bodyLines[:messageEnd], "\n"))
	if message != "" {
		p.messageLines = strings.Split(message, "\n")
	}

	if len(diffStarts) == 0 {
		return p, nil
	}

	// Slice each [start, nextStart) range into its own Diff block.
	for j, start := range diffStarts {
		end := len(bodyLines)
		if j+1 < len(diffStarts) {
			end = diffStarts[j+1]
		}
		d, err := parseFormatDiff(bodyLines[start:end])
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		p.diffs = append(p.diffs, d)
	}
	return p, nil
}
