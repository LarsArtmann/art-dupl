#!/usr/bin/env python3
"""Batch inline annotator for docs-health sweeps.

Applies `~~<original line>~~ <verdict>` to numbered list lines, numbered
markdown table rows, and `- [ ]` checkbox lines. Atomic per file: writes only
if EVERY spec matched exactly one line (a partial spec set aborts with no
write). Refuses to touch lines already containing ~~. Never renumbers.

Usage: annotate-status-items.py <file> <specfile>
Specfile lines: `<key>\t<verdict>` (tab-separated; `#` lines and blanks ignored).
An empty verdict strikes only.

Key forms:
  12            numbered line starting "12." or table row "| 12 |"
  f7            prefixed numbering like "f7." (1-3 letter prefix allowed)
  A#3           "A#3." style keys
  7@substring    item 7 disambiguated: the line must contain <substring>
  any:substring  any line containing <substring> (use sparingly — last resort)
  - [ ] checkbox lines match by literal substring of the checkbox text

AMBIGUITY RULE: keys are first-unstruck-match-wins. If a file has MULTIPLE
numbered lists (lists restarting at 1), a bare `1` hits the FIRST unstruck
"1." line — always disambiguate with `1@<distinct substring>`. Already-struck
lines are doubly protected: `~~1. ...~~` wrapping does not match the number
patterns, and any line containing ~~ is skipped outright.

Workflow: ALWAYS dry-run a new file shape first — copy the file to /tmp, run
the spec against the copy, inspect, then run against the real file. After a
sweep, verify with:
  grep -c '~~' <file>          # presence: every intended strike landed
  check-rows.py <file>         # uniformity: no PARTIAL table rows (skill asset)

Origin: extracted from the 2026-09-28 docs-health full-audit pass
(~140 reports classified, 26 files annotated) — persisted out of /tmp so the
next pass does not start from zero. See
docs/planning/2026-09-28_21-57_SUPERB-docs-health-completion-plan.md (M00).
"""

import pathlib
import re
import sys


def main():
    path = pathlib.Path(sys.argv[1])
    specs = {}
    order = []
    for raw in pathlib.Path(sys.argv[2]).read_text().splitlines():
        if not raw.strip() or raw.startswith("#"):
            continue
        key, _, verdict = raw.partition("\t")
        key = key.strip()
        specs[key] = verdict.strip()
        order.append(key)
    lines = path.read_text().splitlines(keepends=True)
    num_pat = re.compile(
        r"^(\s*)(?:[-*]\s*)?([A-Za-z]{0,3}\d{1,3}[a-z]?|[A-Za-z]+#\d+|\d+)\.\s"
    )
    row_pat = re.compile(r"^\|\s*([A-Za-z]{0,3}\d{1,3}[a-z]?)\s*\|")
    cb_pat = re.compile(r"^(\s*)- \[ \] (.*)$")
    hits = {}
    for i, line in enumerate(lines):
        if "~~" in line:
            continue
        matched = False
        m = num_pat.match(line)
        if not m:
            m = row_pat.match(line)
        if m:
            tok = m.group(2) if m.lastindex and m.lastindex >= 2 else m.group(1)
            if "@" not in tok:
                for key in specs:
                    if key in hits:
                        continue
                    if "@" in key:
                        tok_req, _, substr = key.partition("@")
                        if tok == tok_req and substr in line:
                            hits[key] = i
                            matched = True
                            break
                    elif tok == key:
                        hits[key] = i
                        matched = True
                        break
        if matched:
            continue
        m = cb_pat.match(line)
        if m:
            for key in specs:
                if key in hits:
                    continue
                if key in m.group(2):
                    hits[key] = i
                    matched = True
                    break
        if matched:
            continue
        for key in specs:
            if key in hits:
                continue
            if key.startswith("any:") and key[4:] in line:
                hits[key] = i
                break
    banned = {"open", "todo", "pending", "later", "tbd"}
    for k, v in specs.items():
        if v.lower().strip(" .") in banned:
            print(f"FAIL {path.name}: spec {k!r} has verdict {v!r}; open items stay BARE (absence is the open signal) — omit the key instead")
            sys.exit(1)
    missing = [k for k in order if k not in hits]
    if missing:
        print(f"FAIL {path.name}: unmatched specs: {missing}")
        sys.exit(1)
    for key, idx in hits.items():
        line = lines[idx]
        nl = "\n" if line.endswith("\n") else ""
        stripped = line.rstrip("\n")
        verdict = specs[key]
        new = f"~~{stripped}~~ {verdict}".rstrip()
        lines[idx] = new + nl
    path.write_text("".join(lines))
    print(f"OK {path.name}: annotated {len(hits)} items")


if __name__ == "__main__":
    main()
