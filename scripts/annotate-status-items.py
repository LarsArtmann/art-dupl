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
import sys, re, pathlib

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
    num_pat = re.compile(r"^(\s*)(?:[-*]\s*)?([A-Za-z]{0,3}\d{1,3}[a-z]?|[A-Za-z]+#\d+|\d+)\.\s")
    row_pat = re.compile(r"^\|\s*([A-Za-z]{0,3}\d{1,3}[a-z]?)\s*\|")
    cb_pat = re.compile(r"^(\s*)- \[ \] (.*)$")
    hits = {}
    for i, line in enumerate(lines):
        if "~~" in line:
            continue
        m = num_pat.match(line)
        if not m:
            m = row_pat.match(line)
        if m:
            tok = m.group(2) if m.lastindex and m.lastindex >= 2 else m.group(1)
            if "@" in tok:
                continue
            for key in specs:
                if key in hits:
                    continue
                if "@" in key:
                    tok_req, _, substr = key.partition("@")
                    if tok == tok_req and substr in line:
                        hits[key] = i
                        break
                elif tok == key:
                    hits[key] = i
                    break
            continue
        m = cb_pat.match(line)
        if m:
            for key in specs:
                if key in hits:
                    continue
                if key in m.group(2):
                    hits[key] = i
                    break
            continue
        for key in specs:
            if key in hits:
                continue
            if key.startswith("any:") and key[4:] in line:
                hits[key] = i
                break
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
