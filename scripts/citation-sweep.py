#!/usr/bin/env python3
"""Citation sweep for docs/review/libs/ fleet-run reports.

Extracts every `path:line` (and `path:start-end`) citation from the report
Markdown files and verifies that the file exists (resolving art-dupl-relative,
`<lib>/`-qualified cross-repo, `~`-absolute, and bare-absolute forms) and that
the cited line range fits the file. En-dashes in ranges are normalized.

Exit 0 = every citation resolves. Exit 1 = stale/broken citations (printed).
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

ART_DUPL = Path("/home/lars/projects/art-dupl")
PROJECTS = Path.home() / "projects"
LIB_ROOTS = {
    "go-finding": PROJECTS / "go-finding",
    "linter-autoconfigure-sdk": PROJECTS / "linter-autoconfigure-sdk",
    "go-policy-dsl": PROJECTS / "go-policy-dsl",
    "go-dag-app": PROJECTS / "go-dag-app",
    "go-linter-sdk": PROJECTS / "go-linter-sdk",
    "go-compliance": PROJECTS / "go-compliance",
}

CITE_RE = re.compile(
    r"([\w./~+-]+\.(?:go|md|yml|yaml|json|sh|py|templ|mod|html|ts))"
    r":(\d+)(?:\u2013|-)(\d+)?|"
    r"([\w./~+-]+\.(?:go|md|yml|yaml|json|sh|py|templ|mod|html|ts)):(\d+)"
)


def file_line_count(path: Path) -> int:
    try:
        with path.open("rb") as fh:
            return sum(1 for _ in fh)
    except OSError:
        return -1


SKIP_DIRS = {"node_modules", ".git", "vendor", "dist", ".nix", "result"}


def rglob_basenames(root: Path, name: str) -> list[Path]:
    hits: list[Path] = []
    for candidate in root.rglob(name):
        if any(part in SKIP_DIRS for part in candidate.parts):
            continue
        hits.append(candidate)
    return hits


def resolve(raw: str) -> tuple[Path | None, str):
    """Resolve a citation path. Returns (resolved, note)."""
    p = Path(raw).expanduser()
    if p.is_absolute():
        return (p if p.exists() else None), ""
    for prefix, root in sorted(LIB_ROOTS.items(), key=lambda kv: -len(kv[0])):
        if raw == prefix or raw.startswith(prefix + "/"):
            candidate = root / raw[len(prefix):].lstrip("/")
            return (candidate if candidate.exists() else None), ""
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


def sweep(report: Path) -> list[str]:
    failures: list[str] = []
    for lineno, line in enumerate(report.read_text(encoding="utf-8").splitlines(), 1):
        for m in CITE_RE.finditer(line):
            path_raw = m.group(1) or m.group(4)
            start = int(m.group(2) or m.group(5))
            end_raw = m.group(3)
            end = int(end_raw) if end_raw else start
            resolved, note = resolve(path_raw)
            if resolved is None:
                failures.append(
                    f"{report.name}:{lineno}  {path_raw}:{start}"
                    f"{('-' + str(end)) if end_raw else ''}  -> FILE NOT FOUND {note}"
                )
                continue
            count = file_line_count(resolved)
            if count >= 0 and (start < 1 or end < start or start > count or end > count):
                failures.append(
                    f"{report.name}:{lineno}  {path_raw}:{start}"
                    f"{('-' + str(end)) if end_raw else ''}  -> {resolved} has {count} lines"
                )
    return failures


def main() -> int:
    args = sys.argv[1:]
    targets: list[Path] = []
    for arg in args:
        p = Path(arg)
        if p.is_dir():
            targets.extend(sorted(p.glob("*.md")))
        else:
            targets.append(p)
    if not targets:
        targets = sorted((ART_DUPL / "docs/review/libs").glob("*.md"))

    total_failures = 0
    for report in targets:
        failures = sweep(report)
        status = "OK" if not failures else f"{len(failures)} STALE"
        print(f"{report}: {status}")
        for failure in failures:
            print(f"  {failure}")
        total_failures += len(failures)
    print(f"\ncitation sweep: {len(targets)} report(s), {total_failures} stale citation(s)")
    return 1 if total_failures else 0


if __name__ == "__main__":
    sys.exit(main())
