"""export_requirements converts a uv.lock into a requirements.txt, following
`uv export --format requirements.txt`.

It reimplements the relevant parts of uv's exporter
(uv/crates/uv/src/commands/project/export.rs and
uv/crates/uv-lock/src/lock/export/): it discovers the workspace roots, walks the
locked dependency graph, computes the environment marker under which each
package is reachable, and renders the result in requirements.txt form (with
hashes and `# via` annotations).

Unlike `uv export`, which flattens the resolution assuming every package is
installed together into a single site-packages, this tool preserves the extras
activated on each package (e.g. `fastmcp-slim[client,server]==x`). Bazel's
rules_python rebuilds the dependency graph from those extras and pulls only the
required optional dependencies, so dropping them (as `uv export` does) would
under-specify the graph. uv deliberately omits extras from `requirements.txt`
output and has declined to add them, which is why we reimplement the export
here rather than shelling out to `uv export`:
https://github.com/astral-sh/uv/issues/12072

Known simplifications relative to uv:
  - Marker reachability parses each dependency's marker into disjunctive normal
    form (treating each comparison as an opaque atom), combines them with
    boolean and/or distributed to canonical DNF, and reduces by clause
    subsumption. Individual comparisons are not semantically simplified
    (e.g. `python < '3.13' or python >= '3.13'` is not collapsed to true).
  - Conflict markers and PEP 508 extra-activation gating are not modeled: a
    package's activated extras are the union requested by any reachable
    dependent, which can over-activate under narrow markers.
  - Default dependency groups cannot be read from pyproject.toml, so the `dev`
    group is included (matching uv's built-in default).
"""

import asyncio
import collections
import dataclasses
import sys
import tomllib

import seed.infra.flag.python.seed_flag as seed_flag
import seed.infra.init.python.seed_init as seed_init

arg_uv_lock = seed_flag.define_positional(
    "uv_lock", "", help="path to the uv.lock to convert"
)

# ---------------------------------------------------------------------------
# Lockfile model and parsing
# ---------------------------------------------------------------------------

# Source kinds in priority order; the first key present on a `[package].source`
# table determines how the requirement is rendered.
SOURCE_KINDS = ["editable", "virtual", "directory", "path", "git", "url", "registry"]


@dataclasses.dataclass
class Dep:
    """A single edge in the lockfile: a reference to another package by name,
    gated by an optional marker, optionally activating extras."""

    name: str
    marker: str
    extras: list[str]


@dataclasses.dataclass
class Package:
    """One `[[package]]` entry from the lockfile."""

    name: str = ""
    version: str = ""
    source_kind: str = ""
    source_path: str = ""
    source: dict[str, str] = dataclasses.field(default_factory=dict)
    deps: list[Dep] = dataclasses.field(default_factory=list)
    opt_deps: dict[str, list[Dep]] = dataclasses.field(
        default_factory=dict
    )  # extra -> deps
    dev_deps: dict[str, list[Dep]] = dataclasses.field(
        default_factory=dict
    )  # group -> deps
    hashes: list[str] = dataclasses.field(default_factory=list)

    def is_root(self) -> bool:
        """Whether the package is a workspace root (a member the user is
        developing), as opposed to a locked third-party dependency."""
        return self.source_kind in ("editable", "virtual")


def parse_source(raw: dict) -> tuple[dict[str, str], str, str]:
    """Extract the string-valued source dataclasses.fields and resolve the source kind/path
    from a `[package].source` table."""
    source = {key: value for key, value in raw.items() if isinstance(value, str)}
    for kind in SOURCE_KINDS:
        if kind in source:
            return source, kind, source[kind]
    return source, "", ""


def convert_deps(items: list | None) -> list[Dep]:
    return [
        Dep(item.get("name", ""), item.get("marker", ""), item.get("extra", []) or [])
        for item in items or []
    ]


def convert_package(raw: dict) -> Package:
    source, kind, path = parse_source(raw.get("source", {}) or {})

    hashes = []
    sdist = raw.get("sdist")
    if sdist and sdist.get("hash"):
        hashes.append(sdist["hash"])
    for wheel in raw.get("wheels", []) or []:
        if wheel.get("hash"):
            hashes.append(wheel["hash"])

    return Package(
        name=raw.get("name", ""),
        version=raw.get("version", ""),
        source_kind=kind,
        source_path=path,
        source=source,
        deps=convert_deps(raw.get("dependencies")),
        opt_deps={
            extra: convert_deps(deps)
            for extra, deps in (raw.get("optional-dependencies") or {}).items()
        },
        dev_deps={
            group: convert_deps(deps)
            for group, deps in (raw.get("dev-dependencies") or {}).items()
        },
        hashes=hashes,
    )


def parse_lock(data: bytes) -> list[Package]:
    """Decode a uv.lock into the internal package representation."""
    lock = tomllib.loads(data.decode("utf-8"))
    raw_packages = lock.get("package") or []
    if not raw_packages:
        raise ValueError("no packages found; is this a uv.lock?")
    return [convert_package(raw) for raw in raw_packages]


# ---------------------------------------------------------------------------
# Markers
# ---------------------------------------------------------------------------


@dataclasses.dataclass
class Marker:
    """A disjunction (OR) of conjunctions (AND) of opaque marker-string atoms:
    a disjunctive normal form. The empty conjunction (a clause with no atoms) is
    true, so is_true collapses the whole marker to true; a marker with no clauses
    is false."""

    is_true: bool = False
    clauses: list[list[str]] = dataclasses.field(
        default_factory=list
    )  # each clause is sorted, deduped


def marker_true() -> Marker:
    return Marker(is_true=True)


def marker_false() -> Marker:
    return Marker()


def marker_is_false(m: Marker) -> bool:
    return not m.is_true and not m.clauses


def marker_key(m: Marker) -> str:
    """A canonical string identity, used for fixpoint comparison."""
    if m.is_true:
        return "\x01true"
    if marker_is_false(m):
        return "\x01false"
    parts = sorted("\x00".join(clause) for clause in m.clauses)
    return "\x02".join(parts)


def marker_render(m: Marker) -> str:
    """Render the marker as a PEP 508 expression, or "" when it is true."""
    if m.is_true or marker_is_false(m):
        return ""
    parts = sorted(" and ".join(clause) for clause in m.clauses)
    if len(parts) == 1:
        return parts[0]
    # Parenthesize a disjunct only when it binds looser than the surrounding
    # `or`, i.e. when it itself contains a top-level `and`/`or`.
    wrapped = [f"({p})" if " and " in p or " or " in p else p for p in parts]
    return " or ".join(wrapped)


def norm_clause(atoms: list[str]) -> list[str]:
    """Sort and dedupe the atoms of a conjunction."""
    return sorted(set(atoms))


def is_subset(a: list[str], b: list[str]) -> bool:
    return set(a) <= set(b)


def reduce_or(clauses: list[list[str]]) -> list[list[str]]:
    """Dedupe identical clauses and drop any clause subsumed by a strictly
    broader one (a subset clause implies the superset, so the superset is
    redundant in a disjunction)."""
    seen = set()
    uniq = []
    for clause in clauses:
        key = "\x00".join(clause)
        if key not in seen:
            seen.add(key)
            uniq.append(clause)
    kept = []
    for i, clause in enumerate(uniq):
        subsumed = any(
            i != j and len(other) < len(clause) and is_subset(other, clause)
            for j, other in enumerate(uniq)
        )
        if not subsumed:
            kept.append(clause)
    return kept


def and_clauses(a: list[list[str]], b: list[list[str]]) -> list[list[str]]:
    """Distribute AND over two disjunctive-normal-form clause sets, producing
    the Cartesian product of conjunctions."""
    if not a:
        return b
    if not b:
        return a
    return [norm_clause(ca + cb) for ca in a for cb in b]


def marker_and(a: Marker, b: Marker) -> Marker:
    """Conjoin two markers, distributing AND over OR so the result stays in
    disjunctive normal form (matching uv's canonical marker rendering)."""
    if a.is_true:
        return b
    if b.is_true:
        return a
    if marker_is_false(a) or marker_is_false(b):
        return marker_false()
    return Marker(clauses=reduce_or(and_clauses(a.clauses, b.clauses)))


def marker_or(a: Marker, b: Marker) -> Marker:
    if a.is_true or b.is_true:
        return marker_true()
    if marker_is_false(a):
        return b
    if marker_is_false(b):
        return a
    return Marker(clauses=reduce_or(a.clauses + b.clauses))


def is_marker_boundary(c: str) -> bool:
    """Whether c delimits a marker keyword (`and`/`or`)."""
    return c in " \t()"


def marker_keyword_at(s: str, i: int) -> tuple[str, int]:
    """Return the boolean keyword starting at s[i] (with its length), or ("", 0)
    if none begins there at a word boundary."""
    for kw in ("and", "or"):
        n = len(kw)
        if s[i : i + n] == kw:
            before = i == 0 or is_marker_boundary(s[i - 1])
            after = i + n == len(s) or is_marker_boundary(s[i + n])
            if before and after:
                return kw, n
    return "", 0


def tokenize_marker(s: str) -> list[str]:
    """Split a PEP 508 marker into a flat token stream of atoms (opaque
    comparisons), the keywords `and`/`or`, and parentheses. Quoted values are
    kept intact so operators inside them are not mistaken for keywords."""
    toks = []
    buf = []

    def flush():
        atom = "".join(buf).strip()
        if atom:
            toks.append(atom)
        buf.clear()

    i, n = 0, len(s)
    while i < n:
        c = s[i]
        if c in ("'", '"'):
            j = i + 1
            while j < n and s[j] != c:
                j += 1
            if j < n:
                j += 1  # include the closing quote
            buf.append(s[i:j])
            i = j
            continue
        if c in ("(", ")"):
            flush()
            toks.append(c)
            i += 1
            continue
        kw, length = marker_keyword_at(s, i)
        if kw:
            flush()
            toks.append(kw)
            i += length
            continue
        buf.append(c)
        i += 1
    flush()
    return toks


class MarkerParser:
    """A recursive-descent parser over the token stream that yields disjunctive
    normal form (an OR of AND-clauses)."""

    def __init__(self, toks: list[str]):
        self.toks = toks
        self.pos = 0

    def peek(self) -> str:
        return self.toks[self.pos] if self.pos < len(self.toks) else ""

    def next(self) -> str:
        tok = self.peek()
        self.pos += 1
        return tok

    def parse_or(self) -> list[list[str]]:
        clauses = self.parse_and()
        while self.peek() == "or":
            self.next()
            clauses = clauses + self.parse_and()
        return clauses

    def parse_and(self) -> list[list[str]]:
        clauses = self.parse_term()
        while self.peek() == "and":
            self.next()
            clauses = and_clauses(clauses, self.parse_term())
        return clauses

    def parse_term(self) -> list[list[str]]:
        if self.peek() == "(":
            self.next()
            inner = self.parse_or()
            if self.peek() == ")":
                self.next()
            return inner
        atom = self.next()
        if atom == "":
            return []
        return [[atom]]


def marker_from_string(s: str) -> Marker:
    """Parse a lockfile marker string into disjunctive normal form; an empty
    string is the always-true marker."""
    if not s.strip():
        return marker_true()
    clauses = MarkerParser(tokenize_marker(s)).parse_or()
    if not clauses:
        return marker_true()
    return Marker(clauses=reduce_or(clauses))


# ---------------------------------------------------------------------------
# Dependency graph resolution
# ---------------------------------------------------------------------------


@dataclasses.dataclass
class Exportable:
    """A package selected for export together with its reachability marker, the
    extras activated on it, and the packages that depend on it."""

    pkg: Package
    marker: Marker
    extras: list[str]
    dependents: list[str]


def marker_reachability(
    root: int, adj: dict[int, list[tuple[int, Marker]]]
) -> dict[int, Marker]:
    """Propagate markers from the root through the graph. Each node's marker is
    the disjunction, over every path from the root, of the conjunction of the
    edge markers along that path. Markers only grow toward true, and the atom
    set is finite, so the fixpoint terminates."""
    reach = {root: marker_true()}
    queue = collections.deque([root])
    while queue:
        u = queue.popleft()
        um = reach[u]
        for to, mark in adj.get(u, []):
            cand = marker_and(um, mark)
            cur = reach.get(to)
            nxt = marker_or(cur if cur is not None else marker_false(), cand)
            if cur is None or marker_key(nxt) != marker_key(cur):
                reach[to] = nxt
                queue.append(to)
    del reach[root]
    return reach


def root_indices(packages: list[Package]) -> list[int]:
    """Return the workspace root packages. Editable/virtual packages are always
    roots; if a lockfile has none (e.g. a flat, non-project lock), packages that
    nothing depends on are treated as roots instead."""
    roots = [idx for idx, p in enumerate(packages) if p.is_root()]
    if roots:
        return roots

    depended = set()
    for p in packages:
        for dep in p.deps:
            depended.add(dep.name)
        for deps in p.opt_deps.values():
            for dep in deps:
                depended.add(dep.name)
    return [idx for idx, p in enumerate(packages) if p.name not in depended]


def source_rank(p: Package) -> int:
    """Order editables first, then path/directory requirements, then named
    packages, matching uv's RequirementComparator."""
    if p.source_kind == "editable":
        return 0
    if p.source_kind in ("path", "directory"):
        return 1
    return 2


def resolve(packages: list[Package]) -> list[Exportable]:
    """Walk the locked dependency graph from the workspace roots and return the
    exportable packages, sorted for output."""
    by_name: dict[str, list[int]] = {}
    for idx, p in enumerate(packages):
        by_name.setdefault(p.name, []).append(idx)

    roots = root_indices(packages)

    # The virtual root node has index len(packages); all workspace roots and
    # their dependency groups connect to it.
    root = len(packages)
    adj: dict[int, list[tuple[int, Marker]]] = {}
    incoming: dict[int, set[str]] = {}

    def add_edge(frm: int, to: int, marker_str: str):
        adj.setdefault(frm, []).append((to, marker_from_string(marker_str)))
        if frm != root:
            incoming.setdefault(to, set()).add(packages[frm].name)

    # Track the extras activated on each package. Unlike a flat `uv export`,
    # rules_python rebuilds the dependency graph from these, so they must be
    # preserved (e.g. `fastmcp-slim[client,server]`). uv itself refuses to emit
    # extras in `requirements.txt`: https://github.com/astral-sh/uv/issues/12072
    activated: dict[int, set[str]] = {}

    def record_extras(to: int, extras: list[str]):
        if extras:
            activated.setdefault(to, set()).update(extras)

    queue = collections.deque()
    seen = set()

    def enqueue(idx: int, extra: str):
        if (idx, extra) not in seen:
            seen.add((idx, extra))
            queue.append((idx, extra))

    def visit_dep(frm: int, dep: Dep):
        for di in by_name.get(dep.name, []):
            add_edge(frm, di, dep.marker)
            record_extras(di, dep.extras)
            enqueue(di, "")
            for extra in dep.extras:
                enqueue(di, extra)

    for r in roots:
        add_edge(root, r, "")
        enqueue(r, "")

        # Development dependencies connect directly to the root: they may be
        # installed without installing the workspace package itself. Only the
        # `dev` group is exported, matching uv's built-in default.
        for dep in packages[r].dev_deps.get("dev", []):
            visit_dep(root, dep)

    while queue:
        idx, extra = queue.popleft()
        p = packages[idx]
        deps = p.opt_deps.get(extra, []) if extra else p.deps
        for dep in deps:
            visit_dep(idx, dep)

    reach = marker_reachability(root, adj)

    nodes = []
    for idx, p in enumerate(packages):
        m = reach.get(idx)
        if m is None or marker_is_false(m):
            continue
        if p.source_kind == "virtual":
            continue
        nodes.append(
            Exportable(
                pkg=p,
                marker=m,
                extras=sorted(activated.get(idx, set())),
                dependents=sorted(incoming.get(idx, set())),
            )
        )

    nodes.sort(
        key=lambda node: (source_rank(node.pkg), node.pkg.name, node.pkg.version)
    )
    return nodes


# ---------------------------------------------------------------------------
# Rendering
# ---------------------------------------------------------------------------


def anchor_path(path: str) -> str:
    """Render a lockfile path as a requirements.txt entry: absolute paths become
    file URLs, relative paths are anchored at the current directory."""
    if path == "":
        return "."
    if path.startswith("/"):
        return "file://" + path
    if path == "." or path.startswith("./") or path.startswith("../"):
        return path
    return "./" + path


def git_url(p: Package) -> str:
    """Reconstruct a PEP 508 git URL from the lockfile source. The lock stores
    the reference and commit separately; uv re-encodes them into the URL."""
    url = p.source_path
    if not url.startswith("git+"):
        url = "git+" + url
    rev = p.source.get("rev", "")
    if rev:
        url += "@" + rev
    sub = p.source.get("subdirectory", "")
    if sub:
        url += "#subdirectory=" + sub
    return url


def extras_suffix(extras: list[str]) -> str:
    """Render the PEP 508 `[extra1,extra2]` suffix, or "" when there are none.
    The extras are expected pre-sorted."""
    if not extras:
        return ""
    return "[" + ",".join(extras) + "]"


def requirement_string(p: Package, extras: list[str]) -> str:
    """Render the left-hand side of a requirement line for a package, following
    its source type and including any activated extras so rules_python can
    reconstruct the dependency graph."""
    suffix = extras_suffix(extras)
    if p.source_kind == "registry":
        return f"{p.name}{suffix}=={p.version}"
    if p.source_kind == "editable":
        return "-e " + anchor_path(p.source_path)
    if p.source_kind in ("path", "directory"):
        return anchor_path(p.source_path)
    if p.source_kind == "git":
        return f"{p.name}{suffix} @ {git_url(p)}"
    if p.source_kind == "url":
        return f"{p.name}{suffix} @ {p.source_path}"
    if p.version:
        return f"{p.name}{suffix}=={p.version}"
    return p.name + suffix


def render(out, nodes: list[Exportable]):
    parts = [
        "# This file was autogenerated by uv via the following command:\n",
        "#    bazel run //seed/devprod/python/translate_uv_lock:export_requirements -- $(pwd)/uv.lock\n",
    ]
    for node in nodes:
        line = requirement_string(node.pkg, node.extras)

        marker = marker_render(node.marker)
        if marker:
            line += f" ; {marker}"

        for h in sorted(node.pkg.hashes):
            line += f" \\\n    --hash={h}"
        line += "\n"
        parts.append(line)

        dependents = node.dependents
        if len(dependents) == 1:
            parts.append(f"    # via {dependents[0]}\n")
        elif len(dependents) > 1:
            parts.append("    # via\n")
            for d in dependents:
                parts.append(f"    #   {d}\n")

    out.write("".join(parts))


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------


async def main():
    seed_init.initialize()
    lock_path = arg_uv_lock.get()
    with open(lock_path, "rb") as f:
        data = f.read()
    render(sys.stdout, resolve(parse_lock(data)))


if __name__ == "__main__":
    asyncio.run(main())
