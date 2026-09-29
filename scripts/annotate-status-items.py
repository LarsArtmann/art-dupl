#!/usr/bin/env python3
"""Batch inline annotator for docs-health sweeps.

Applies `~~<original line>~~ <verdict>` to numbered list lines, numbered
markdown table rows, and `- [ ]` checkbox lines. Atomic per file: writes only
if EVERY spec matched exactly one line (a partial spec set aborts with no
write). Refuses to touch lines already containing ~~. Never renumbers.

Usage:
  annotate-status-items.py <file> <specfile>            strike mode (default)
  annotate-status-items.py --verify <file> <specfile>   resolve + print, write NOTHING
  annotate-status-items.py --emit-keys <file> <lineno> [<lineno> ...]
                                                        print ready-to-paste spec keys

Specfile lines: `<key>\t<verdict>` (tab-separated; `#` lines and blanks ignored).
An empty verdict strikes only.

Key forms:
  12            numbered line starting "12." or table row "| 12 |"
  f7            prefixed numbering like "f7." (1-3 letter prefix allowed)
  A#3           "A#3." style keys
  7@substring    item 7 disambiguated: the line must contain <substring>
  any:substring  any line containing <substring> (use sparingly — last resort)
  - [ ] checkbox lines match by literal substring of the checkbox text

AMBIGUITY RULES:
  - Duplicate keys in a specfile are a HARD ERROR (the silent dict-overwrite
    mis-struck 4 files in the June batch — a bare numeric key quietly replaced
    its earlier verdict). Never rely on "last one wins".
  - Keys are first-unstruck-match-wins. If a file has MULTIPLE numbered lists
    (lists restarting at 1), a bare `1` hits the FIRST unstruck "1." line —
    always disambiguate with `1@<distinct substring>`.
  - Already-struck lines are doubly protected: `~~1. ...~~` wrapping does not
    match the number patterns, and any line containing ~~ is skipped outright.

Workflow:
  1. Generate keys mechanically, never from memory:
       annotate-status-items.py --emit-keys <file> 12 13 14 >> spec.tsv
     (stdout = one `N@substring` key per line, paste-ready; stderr = the
     matched-line previews for eyeballing).
  2. Dry-run a new file shape first — copy to /tmp, run the spec against the
     copy, inspect, then run against the real file.
  3. `--verify` before every real run: prints `lineNo: matched line` per key
     and writes nothing.
  4. After a sweep, verify with:
       grep -c '~~' <file>          # presence: every intended strike landed
       check-rows.py <file>         # uniformity: no PARTIAL table rows (skill asset)

Origin: extracted from the 2026-09-28 docs-health full-audit pass
(~140 reports classified, 26 files annotated) — persisted out of /tmp so the
next pass does not start from zero. Hardened 2026-09-29 (SUPERB v2 M03:
dup-key refusal, --verify, --emit-keys — fixtures in scripts/check-annotator.sh).
See docs/planning/2026-09-28_21-57_SUPERB-docs-health-completion-plan.md (M00).
"""

import pathlib
import re
import sys

BANNED_VERDICTS = {"open", "todo", "pending", "later", "tbd"}


def fail(msg):
    print(f"FAIL {msg}")
    sys.exit(1)


def load_specs(spec_path):
    """Parse the specfile; refuse duplicate keys before any file is read."""
    specs = {}
    order = []
    seen_lines = {}
    for lineno, raw in enumerate(
        pathlib.Path(spec_path).read_text().splitlines(), start=1
    ):
        if not raw.strip() or raw.startswith("#"):
            continue
        key, _, verdict = raw.partition("\t")
        key = key.strip()
        if key in specs:
            fail(
                f"{pathlib.Path(spec_path).name}:{lineno}: duplicate spec key "
                f"{key!r} (first defined at line {seen_lines[key]}); a repeated "
                f"key silently overwrites in dict-based specs — disambiguate "
                f"with `@substring` or merge the verdicts"
            )
        specs[key] = verdict.strip()
        order.append(key)
        seen_lines[key] = lineno
    if not specs:
        fail(f"{pathlib.Path(spec_path).name}: specfile has no key lines")
    for k, v in specs.items():
        if v.lower().strip(" .") in BANNED_VERDICTS:
            fail(
                f"{pathlib.Path(spec_path).name}: spec {k!r} has verdict "
                f"{v!r}; open items stay BARE (absence is the open signal) — "
                f"omit the key instead"
            )
    return specs, order


NUM_PAT = re.compile(
    r"^(\s*)(?:[-*]\s*)?([A-Za-z]{0,3}\d{1,3}[a-z]?|[A-Za-z]+#\d+|\d+)\.\s"
)
ROW_PAT = re.compile(r"^\|\s*([A-Za-z]{0,3}\d{1,3}[a-z]?)\s*\|")
CB_PAT = re.compile(r"^(\s*)- \[ \] (.*)$")


def resolve(lines, specs):
    """Return {key: line_index} for every spec key, or the missing keys."""
    hits = {}
    for i, line in enumerate(lines):
        if "~~" in line:
            continue
        matched = False
        m = NUM_PAT.match(line)
        if not m:
            m = ROW_PAT.match(line)
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
        m = CB_PAT.match(line)
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
    missing = [k for k in specs if k not in hits]
    return hits, missing


def emit_keys(path, linenos):
    """Print paste-ready `N@substring` keys for the given 1-based line numbers."""
    lines = pathlib.Path(path).read_text().splitlines()
    for arg in linenos:
        try:
            no = int(arg)
        except ValueError:
            fail(f"--emit-keys: line number {arg!r} is not an integer")
        if not 1 <= no <= len(lines):
            fail(f"--emit-keys: line {no} out of range (file has {len(lines)} lines)")
        line = lines[no - 1]
        m = NUM_PAT.match(line) or ROW_PAT.match(line)
        if m:
            tok = m.group(2) if m.lastindex and m.lastindex >= 2 else m.group(1)
        else:
            mcb = CB_PAT.match(line)
            if mcb:
                tok = None
            else:
                print(
                    f"--emit-keys: line {no} is not a numbered row, table row, "
                    f"or checkbox; use an `any:` key manually if warranted",
                    file=sys.stderr,
                )
                continue
        if tok is not None:
            rest = line[m.end() :].strip()
        else:
            tok = None
            rest = mcb.group(2).strip()
        rest = rest.split("\t")[0].strip()
        substr = rest[:32].strip()
        if not substr:
            fail(f"--emit-keys: line {no} has no text after its marker")
        if tok is None:
            key = substr
        else:
            key = f"{tok}@{substr}"
        print(key)
        print(f"{no}: {line.strip()[:100]}", file=sys.stderr)


def main():
    args = sys.argv[1:]
    if not args:
        print(__doc__)
        sys.exit(2)
    if args[0] == "--emit-keys":
        if len(args) < 3:
            fail("usage: --emit-keys <file> <lineno> [<lineno> ...]")
        emit_keys(args[1], args[2:])
        return
    verify = args[0] == "--verify"
    if verify:
        args = args[1:]
    if len(args) != 2:
        fail("usage: [--verify] <file> <specfile> | --emit-keys <file> <lineno...>")
    path = pathlib.Path(args[0])
    specs, _ = load_specs(args[1])
    lines = path.read_text().splitlines(keepends=True)
    hits, missing = resolve(lines, specs)
    if missing:
        fail(f"{path.name}: unmatched specs: {missing}")
    if verify:
        for key, idx in sorted(hits.items(), key=lambda kv: kv[1]):
            print(f"{idx + 1}: {lines[idx].rstrip()}")
        print(f"VERIFY OK {path.name}: {len(hits)} keys resolved (nothing written)")
        return
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
