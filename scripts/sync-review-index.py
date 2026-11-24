#!/usr/bin/env python3
"""Derive-and-verify docs/review/README.md version cells from report headers.

Every per-lib report (docs/review/libs/*.md) carries a header table with the
exact checkout it audited (tag + full/partial SHA + review Date). The index
table in docs/review/README.md repeats (tag @ shortsha) and the review Date
per row — the only mechanically derivable cells, and historically the drift
class (f33). This script DERIVES the expected tag/sha tokens from each
report header, verifies the README cell carries exactly them, and verifies
the Date column equals the report's Date field. Everything else in the cell
(tag dates, annotations like "; hub/ untagged") is prose and untouched.

Modes:
  --check (default)  report drift, exit 1 on any
  --fix              rewrite the drifted tokens/Date cell in place

Wire-in: the flake `citation-sweep` check runs `--check` so index drift
fails CI alongside stale citations.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
LIBS_DIR = REPO / "docs" / "review" / "libs"
README = REPO / "docs" / "review" / "README.md"

TAG_RE = re.compile(r"\b(?:toolsdk/)?v\d+\.\d+\.\d+(?:[-.\w]*-g[0-9a-f]{7,})?")
SHA_RE = re.compile(r"\b[0-9a-f]{7,40}\b")


def header_fields(report: Path) -> dict[str, str]:
    """Parse the report's header table into field -> value."""
    fields: dict[str, str] = {}
    for line in report.read_text().splitlines()[:30]:
        m = re.match(r"^\|\s*([^|]+?)\s*\|\s*(.+?)\s*\|\s*$", line)
        if m:
            fields[m.group(1).strip()] = m.group(2).strip()
    return fields


def derive_tag_sha(header_value: str) -> tuple[str, str]:
    """Derive (tag, shortsha) from a report header cell value."""
    tags = TAG_RE.findall(header_value)
    if not tags:
        raise ValueError(f"no version tag found in: {header_value[:80]}")
    if any(t.startswith("toolsdk/") for t in tags):
        tag = next(t for t in tags if t.startswith("toolsdk/"))
        pos = header_value.find(tag)
        after = header_value[pos + len(tag) :]
        m = SHA_RE.search(after)
        if not m:
            raise ValueError(f"toolsdk tag {tag} carries no following sha")
        return tag, m.group(0)[:7]
    tag = tags[0]
    # Prefer the header's full checkout SHA (40 hex), excluding any sha that
    # is part of the tag's own -g suffix; fall back to the g-suffix.
    for m in SHA_RE.finditer(header_value):
        tok = m.group(0)
        if len(tok) == 40 and tok not in tag:
            return tag, tok[:8]
    m = re.search(r"-g([0-9a-f]{7,})$", tag)
    if m:
        return tag, m.group(1)[:8]
    m = SHA_RE.search(header_value)
    if not m:
        raise ValueError(f"tag {tag} but no sha in: {header_value[:80]}")
    return tag, m.group(0)[:8]


CELL_RE = re.compile(r"^`([^`]+)`\s+@\s+`([0-9a-f]+)`\s*(.*)$")


def check_cell(cell: str, tag: str, sha: str) -> str | None:
    """Return a rewritten cell, or None when the cell already matches.

    Canonical form: `` `TAG` @ `SHA` `` plus any trailing prose (annotation
    parentheticals) preserved verbatim. The SHA is the describe suffix when
    the tag carries one, else the first 8 hex of the header's checkout SHA.
    """
    m = CELL_RE.match(cell)
    if not m:
        raise ValueError(
            f"unparseable version cell (expected `TAG` @ `SHA` prose): {cell}"
        )
    existing_tag, existing_sha, prose = m.group(1), m.group(2), m.group(3).strip()
    if existing_tag == tag and existing_sha == sha:
        return None
    canonical = f"`{tag}` @ `{sha}`"
    if prose:
        canonical += f" {prose}"
    return canonical


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--check", action="store_true", help="verify only (default)")
    ap.add_argument("--fix", action="store_true", help="rewrite drifted cells")
    ap.add_argument("--root", type=Path, default=REPO, help="repo root override")
    args = ap.parse_args()
    root: Path = args.root
    readme_path = root / "docs" / "review" / "README.md"
    libs_dir = root / "docs" / "review" / "libs"

    rows = re.findall(
        r"^\|\s*(\S[^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|"
        r"\s*([^|]*?)\s*\|\s*\[libs/([^]]+\.md)\]",
        readme_path.read_text(),
        re.MULTILINE,
    )
    if not rows:
        print("FATAL: no index rows parsed from", readme_path)
        return 2

    problems: list[str] = []
    fixes: list[tuple[str, str, str]] = []  # (lib, old_line, new_line)
    for lib, version_cell, date_cell, _verdict, _score, report_name in rows:
        report = libs_dir / report_name
        if not report.exists():
            problems.append(f"{lib}: report {report_name} missing")
            continue
        fields = header_fields(report)
        header_value = fields.get("Library checkout reviewed") or fields.get(
            "Version reviewed"
        )
        if not header_value:
            problems.append(f"{lib}: no checkout/Version-reviewed header field")
            continue
        try:
            tag, sha = derive_tag_sha(header_value)
        except ValueError as e:
            problems.append(f"{lib}: {e}")
            continue
        new_cell = check_cell(version_cell, tag, sha)
        if new_cell is not None:
            problems.append(
                f"{lib}: version cell '{version_cell}' != derived '{tag} @ {sha}'"
            )
            fixes.append((lib, version_cell, new_cell))
        report_date = re.sub(r"[^0-9-]", "", fields.get("Date", ""))
        if report_date and date_cell != report_date:
            problems.append(
                f"{lib}: Date column '{date_cell}' != report Date '{report_date}'"
            )
            fixes.append((lib, date_cell, report_date))

    if not problems:
        print(f"review-index sync OK: {len(rows)} rows derived from report headers")
        return 0

    for p in problems:
        print("DRIFT:", p)
    if not args.fix:
        print(f"{len(problems)} drift finding(s); run with --fix to repair")
        return 1

    text = readme_path.read_text()
    for lib, old, new in fixes:
        text = text.replace(old, new, 1)
    readme_path.write_text(text)
    print(f"fixed {len(fixes)} cell(s); re-run with --check to verify")
    return 0


if __name__ == "__main__":
    sys.exit(main())
