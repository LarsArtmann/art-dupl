#!/usr/bin/env python3
"""Citation sweep for docs/review/libs/ fleet-run reports.

Extracts every `path:line` (and `path:start-end`) citation from the report
Markdown files and verifies that the file exists (resolving art-dupl-relative,
`<lib>/`-qualified cross-repo, submodule-qualified (`toolsdk/...` -> go-finding),
`~`-absolute, and bare-absolute forms) and that the cited line range fits the
file. En-dashes in ranges are normalized.

An optional content anchor extends a citation: `path:line ~token` additionally
verifies that `token` appears on a line inside the cited range — existence and
bounds alone cannot catch a semantically stale citation (code moved within
bounds). Use anchors on high-stakes citations.

Modes:
  (default)            sweep docs/review/libs/*.md + docs/review/README.md +
                       docs/research/README.md (explicit args override)
  --self-test          run hermetic fixture probes (regex, resolution, stale
                       detection, anchor detection, ambiguous basenames) —
                       catches a broken script itself; run in gates
  --skip-missing-roots sandbox mode: citations whose only candidate roots are
                       absent (the nix sandbox has no sibling checkouts) are
                       SKIPPED, not failed; art-dupl-internal citations still
                       verify

Exit 0 = every citation resolves (or is skipped). Exit 1 = stale/broken
citations (printed).
"""

from __future__ import annotations

import re
import sys
import tempfile
from pathlib import Path

# Repo root derived from THIS script's location (scripts/ parent), not a
# hardcoded host path: the nix citation-sweep check cds into the cleaned
# store source, where the host checkout does not exist.
ART_DUPL = Path(__file__).resolve().parent.parent
PROJECTS = Path.home() / "projects"
LIB_ROOTS = {
    "go-finding": PROJECTS / "go-finding",
    "linter-autoconfigure-sdk": PROJECTS / "linter-autoconfigure-sdk",
    "go-policy-dsl": PROJECTS / "go-policy-dsl",
    "go-dag-app": PROJECTS / "go-dag-app",
    "go-linter-sdk": PROJECTS / "go-linter-sdk",
    "go-compliance": PROJECTS / "go-compliance",
}

# Submodule-qualified prefixes resolve under their workspace parent when the
# bare form misses everywhere. "cmd" sits behind art-dupl in resolution
# order, so art-dupl's own cmd/ keeps winning for its files.
SUBMODULE_PARENTS = {
    "toolsdk": "go-finding",
    "analysis": "go-finding",
    "pipeline": "go-finding",
    "cmd": "go-finding",
}

CITE_RE = re.compile(
    r"([\w./~+-]+\.(?:go|md|yml|yaml|json|sh|py|templ|mod|html|ts))"
    r":(\d+)(?:[\u2013-](\d+))?"
    r"(?:\s+~([^\s|`]+))?"
)


def file_line_count(path: Path) -> int:
    try:
        with path.open("rb") as fh:
            return sum(1 for _ in fh)
    except OSError:
        return -1


def read_lines(path: Path) -> list[str]:
    try:
        return path.read_text(encoding="utf-8", errors="replace").splitlines()
    except OSError:
        return []


SKIP_DIRS = {"node_modules", ".git", "vendor", "dist", ".nix", "result"}


def rglob_basenames(root: Path, name: str) -> list[Path]:
    hits: list[Path] = []
    for candidate in root.rglob(name):
        if any(part in SKIP_DIRS for part in candidate.parts):
            continue
        hits.append(candidate)
    return hits


def resolve(raw: str) -> tuple[Path | None, str]:
    """Resolve a citation path. Returns (resolved, note)."""
    p = Path(raw).expanduser()
    if p.is_absolute():
        return (p if p.exists() else None), ""
    for prefix, root in sorted(LIB_ROOTS.items(), key=lambda kv: -len(kv[0])):
        if raw == prefix or raw.startswith(prefix + "/"):
            candidate = root / raw[len(prefix) :].lstrip("/")
            return (candidate if candidate.exists() else None), ""
    parent = SUBMODULE_PARENTS.get(raw.split("/")[0])
    if parent is not None and parent in LIB_ROOTS:
        candidate = LIB_ROOTS[parent] / raw
        if candidate.exists():
            return candidate, f"submodule-resolved under {parent}"
    candidate = ART_DUPL / raw
    if candidate.exists():
        return candidate, ""
    for root in LIB_ROOTS.values():
        candidate = root / raw
        if candidate.exists():
            return candidate, ""
    if "/" not in raw:
        hits = rglob_basenames(ART_DUPL, raw)
        seen = {str(h) for h in hits}
        for lib_root in LIB_ROOTS.values():
            for h in rglob_basenames(lib_root, raw):
                seen.add(str(h))
        if len(seen) == 1:
            return Path(next(iter(seen))), "bare-basename"
        if len(seen) > 1:
            return None, "AMBIGUOUS basename: " + ", ".join(sorted(seen))
    return None, ""


def _only_missing_roots(path_raw: str, missing_roots: set[str]) -> bool:
    """True when this citation cannot resolve in the current environment for
    root-absence reasons (sandbox mode). Absolute/tilde citations are
    environment-dependent: skip them when the target is absent here — the
    local run (no missing roots) still verifies them against the host tree."""
    p = Path(path_raw).expanduser()
    if p.is_absolute():
        return not p.exists()
    for prefix in sorted(LIB_ROOTS, key=len, reverse=True):
        if path_raw == prefix or path_raw.startswith(prefix + "/"):
            return prefix in missing_roots
    parent = SUBMODULE_PARENTS.get(path_raw.split("/")[0])
    if parent is not None:
        return parent in missing_roots
    if (ART_DUPL / path_raw).exists():
        return False
    # Bare relative paths (with or without /) may live under any lib root;
    # in sandbox mode a miss with absent roots is presumed root-missing.
    # The local run (no missing roots) never skips and catches real staleness.
    return bool(missing_roots)


def sweep(
    report: Path, skip_missing_roots: bool = False
) -> tuple[list[str], list[str]]:
    """Returns (failures, skipped). Only failures drive the exit code."""
    failures: list[str] = []
    skipped: list[str] = []
    missing_roots = {name for name, root in LIB_ROOTS.items() if not root.is_dir()}
    for lineno, line in enumerate(report.read_text(encoding="utf-8").splitlines(), 1):
        for m in CITE_RE.finditer(line):
            path_raw = m.group(1)
            start = int(m.group(2))
            end_raw = m.group(3)
            end = int(end_raw) if end_raw else start
            anchor = m.group(4)
            cite = f"{path_raw}:{start}{('-' + str(end)) if end_raw else ''}"

            if skip_missing_roots and _only_missing_roots(path_raw, missing_roots):
                skipped.append(
                    f"{report.name}:{lineno}  {cite}  -> SKIPPED (root missing)"
                )
                continue

            resolved, note = resolve(path_raw)
            if resolved is None:
                failures.append(
                    f"{report.name}:{lineno}  {cite}  -> FILE NOT FOUND {note}"
                )
                continue
            count = file_line_count(resolved)
            if count >= 0 and (
                start < 1 or end < start or start > count or end > count
            ):
                failures.append(
                    f"{report.name}:{lineno}  {cite}  -> {resolved} has {count} lines"
                )
                continue
            if anchor:
                lines = read_lines(resolved)
                window = lines[start - 1 : end]
                if window and not any(anchor in l for l in window):
                    failures.append(
                        f"{report.name}:{lineno}  {cite}  -> CONTENT ANCHOR MISS (~{anchor} not in cited range of {resolved})"
                    )
    return failures, skipped


def self_test() -> int:
    """Hermetic probes against a synthetic fixture tree. Raises on any failure."""
    probes: list[str] = []
    global ART_DUPL, LIB_ROOTS

    def check(label: str, condition: bool) -> None:
        if not condition:
            raise AssertionError(f"probe FAILED: {label}")
        probes.append(label)

    with tempfile.TemporaryDirectory() as td:
        root = Path(td)
        repo = root / "repo"
        libs_go = root / "libs" / "go-finding"
        libs_other = root / "libs" / "other-lib"
        for d in (repo / "pkg", repo / "cmd", libs_go / "toolsdk", libs_other):
            d.mkdir(parents=True)

        x_go = repo / "pkg" / "x.go"
        x_go.write_text(
            "\n".join(f"line {i}" for i in range(1, 31)).replace(
                "line 12", "line 12 ANCHOR_TOKEN"
            )
            + "\n"
        )
        (libs_go / "report.go").write_text(
            "\n".join(f"l{i}" for i in range(1, 201)).replace(
                "l140", "l140 SetRunProperty"
            )
            + "\n"
        )
        (libs_go / "toolsdk" / "options.go").write_text(
            "\n".join(f"o{i}" for i in range(1, 181)) + "\n"
        )
        (libs_go / "nested").mkdir()
        (libs_other / "nested").mkdir()
        (libs_go / "nested" / "dup.go").write_text("dup\n")
        (libs_other / "nested" / "dup.go").write_text("dup\n")
        (libs_go / "nested" / "unique.go").write_text("u\n")
        (repo / "cmd" / "x.go").write_text("artdupl side wins\n")

        report = repo / "r.md"

        old_art, old_libs = ART_DUPL, LIB_ROOTS
        ART_DUPL = repo
        LIB_ROOTS = {"go-finding": libs_go, "other-lib": libs_other}
        try:
            report.write_text("see pkg/x.go:12 ~ANCHOR_TOKEN\n")
            check("anchor hit passes", sweep(report) == ([], []))

            report.write_text("see pkg/x.go:12 ~NOT_THERE\n")
            check(
                "anchor miss detected",
                any("CONTENT ANCHOR MISS" in f for f in sweep(report)[0]),
            )

            report.write_text("see pkg/x.go:5-9\n")
            check("range in bounds passes", sweep(report) == ([], []))

            report.write_text("see pkg/x.go:25-40\n")
            check("range out of bounds detected", len(sweep(report)[0]) == 1)

            report.write_text("see go-finding/report.go:140 ~SetRunProperty\n")
            check("lib-root prefix resolves with anchor", sweep(report) == ([], []))

            report.write_text("see toolsdk/options.go:180\n")
            check(
                "submodule prefix resolves",
                sweep(report) == ([], [])
                and "submodule-resolved" in resolve("toolsdk/options.go")[1],
            )

            report.write_text("see go-finding/report.go:140 ~SetRunProperty\n")
            failures, skipped = sweep(report, skip_missing_roots=True)
            check(
                "skip-missing-roots: present root still verifies",
                failures == [] and skipped == [],
            )

            LIB_ROOTS = {"go-finding": root / "no-such-root"}
            failures, skipped = sweep(report, skip_missing_roots=True)
            check(
                "skip-missing-roots: absent root -> skipped, not failed",
                failures == [] and len(skipped) == 1,
            )
            failures, _ = sweep(report, skip_missing_roots=False)
            check("without flag, absent root still fails", len(failures) == 1)
            LIB_ROOTS = {"go-finding": libs_go, "other-lib": libs_other}

            report.write_text("see cmd/x.go:1\n")
            check(
                "cmd falls through to art-dupl",
                resolve("cmd/x.go")[0] == repo / "cmd" / "x.go",
            )

            report.write_text("see dup.go:1\n")
            res, note = resolve("dup.go")
            check(
                "ambiguous basename rejected",
                res is None and note.startswith("AMBIGUOUS"),
            )

            report.write_text("see unique.go:1\n")
            check(
                "bare unique basename resolves",
                resolve("unique.go")[1] == "bare-basename",
            )

            report.write_text("en–dash range see pkg/x.go:3–6\n")
            check("en-dash range parses", sweep(report) == ([], []))

            report.write_text("| `pkg/x.go:12 ~ANCHOR_TOKEN` | table cell |\n")
            check("backticked table citation parses", sweep(report) == ([], []))

            m1 = CITE_RE.search("a.go:7")
            m2 = CITE_RE.search("a.go:7-9 ~tok")
            check(
                "CITE_RE groups intact",
                m1.group(2) == "7"
                and m1.group(4) is None
                and m2.group(3) == "9"
                and m2.group(4) == "tok",
            )

            check(
                "skip-missing-roots: all roots present -> no skip",
                not _only_missing_roots("go-finding/report.go", set()),
            )
            check(
                "skip-missing-roots: absent root targeted -> skip",
                _only_missing_roots("go-finding/report.go", {"go-finding"}),
            )
            check(
                "skip-missing-roots: present absolute never skipped",
                not _only_missing_roots(
                    str((repo / "pkg" / "x.go").absolute()), {"go-finding"}
                ),
            )
            check(
                "skip-missing-roots: absent absolute skips",
                _only_missing_roots("/etc/hosts-definitely-absent", {"go-finding"}),
            )
            check(
                "predicate flags absent absolute; only sandbox mode consults it",
                _only_missing_roots("/etc/hosts-definitely-absent", set()),
            )
            check(
                "skip-missing-roots: art-dupl slashed still verified",
                not _only_missing_roots("pkg/x.go", {"go-finding"}),
            )
            check(
                "skip-missing-roots: bare slashed lib path skips",
                _only_missing_roots("audit/settings.go", {"go-finding"}),
            )
            check(
                "no missing roots: bare slashed never skips",
                not _only_missing_roots("audit/settings.go", set()),
            )
        finally:
            ART_DUPL, LIB_ROOTS = old_art, old_libs

    print(f"SELF-TEST OK ({len(probes)} probes)")
    return 0


def main() -> int:
    args = sys.argv[1:]
    if "--self-test" in args:
        args.remove("--self-test")
        try:
            return self_test()
        except AssertionError as exc:
            print(f"SELF-TEST FAILED: {exc}")
            return 1

    skip_missing_roots = "--skip-missing-roots" in args
    args = [a for a in args if a != "--skip-missing-roots"]

    targets: list[Path] = []
    for arg in args:
        p = Path(arg)
        if p.is_dir():
            targets.extend(sorted(p.glob("*.md")))
        else:
            targets.append(p)
    if not targets:
        targets = sorted((ART_DUPL / "docs/review/libs").glob("*.md"))
        targets += [
            ART_DUPL / "docs/review/README.md",
            ART_DUPL / "docs/research/README.md",
        ]

    total_failures = 0
    for report in targets:
        failures, skipped = sweep(report, skip_missing_roots)
        status = "OK" if not failures else f"{len(failures)} STALE"
        if skipped:
            status += f" ({len(skipped)} skipped)"
        print(f"{report}: {status}")
        for failure in failures:
            print(f"  {failure}")
        for skip in skipped:
            print(f"  {skip}")
        total_failures += len(failures)
    print(
        f"\ncitation sweep: {len(targets)} report(s), {total_failures} stale citation(s)"
    )
    return 1 if total_failures else 0


if __name__ == "__main__":
    sys.exit(main())
