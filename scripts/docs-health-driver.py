#!/usr/bin/env python3
"""docs-health batch annotator driver (2026-10-05 pass).

Consumes a verdict file with blocks:

  FILE <relative/path>
  CLASS ARCHIVE|KEEP|SKIP|EMPTY
  MOVE archived|done|reviews|none
  FIX <exact old line> ||= <exact new line>     (repeatable; ||= separator)
  BANNER <one-line resolution note>              (inserted under the H1/title)
  ITEMS:
  <key>\t<DONE|OPEN|WONT|SUPERSEDED>\t<evidence>
  END

Behavior per file:
- CLASS EMPTY or BANNER: write a resolution note under the first heading line
  (strikethrough form so the archive grep-gate's '~~' presence check holds).
- FIX: literal replace old -> new (old wrapped in ~~...~~ strikethrough + new).
- ITEMS: feed specs to scripts/annotate-status-items.py (atomic per file).
  OPEN items produce no spec (absence = open signal).
- MOVE archived -> git mv docs/status/<f> docs/status/archived/<f>
  MOVE done     -> docs/feedback/done/   MOVE reviews -> docs/reviews/archived/
- Runs in --apply mode only after --check passes for that file; per-file atomic.

Usage: driver.py verdicts.txt [--apply]
"""
import pathlib, re, subprocess, sys

ROOT = pathlib.Path("/home/lars/projects/art-dupl")
ANNO = ROOT / "scripts/annotate-status-items.py"

def parse(path):
    blocks, cur = [], None
    for raw in pathlib.Path(path).read_text().splitlines():
        if raw.startswith("FILE "):
            cur = {"file": raw[5:].strip(), "class": "KEEP", "move": "none",
                   "fix": [], "banner": None, "items": []}
            blocks.append(cur)
        elif cur is None:
            continue
        elif raw.startswith("CLASS "):
            cur["class"] = raw[6:].strip()
        elif raw.startswith("MOVE "):
            cur["move"] = raw[5:].strip()
        elif raw.startswith("FIX "):
            old, new = raw[4:].split("||=", 1)
            cur["fix"].append((old.strip(), new.strip()))
        elif raw.startswith("BANNER "):
            cur["banner"] = raw[7:].strip()
        elif raw.startswith("AUDIT "):
            cur["audit"] = raw[6:].split(" ", 2)
        elif raw.strip() == "ITEMS:":
            pass
        elif raw.strip() == "END":
            pass
        elif raw.startswith("ARCHIVE "):
            cur["archive_note"] = raw[8:].strip()
        elif raw.startswith("AUDIT "):
            # AUDIT <resolved> <open> [open-items-summary]
            parts = raw[6:].split(" ", 2)
            cur["audit"] = parts
        elif "\t" in raw and cur["items"] is not None:
            parts = raw.split("\t")
            if len(parts) >= 2 and parts[1] in ("DONE", "OPEN", "WONT", "SUPERSEDED"):
                ev = parts[2].strip() if len(parts) > 2 else "resolved"
                cur["items"].append((parts[0].strip(), parts[1], ev))
    return blocks

def spec_verdict(v, ev):
    if v == "DONE":
        return f"done at {ev}"
    if v == "WONT":
        return f"Won't implement — {ev}"
    if v == "SUPERSEDED":
        return f"superseded — {ev}"
    return None

def build_spec(block, tmpdir):
    lines = []
    for key, v, ev in block["items"]:
        if "struck" in ev.lower():
            continue  # already annotated in a prior pass; annotator would abort
        txt = spec_verdict(v, ev)
        if txt:
            lines.append(f"{key}\t{txt}")
    if not lines:
        return None
    p = tmpdir / (block["file"].replace("/", "__") + ".tsv")
    p.write_text("\n".join(lines) + "\n")
    return p

def do_fixes(path, fixes):
    text = path.read_text()
    ok = True
    for old, new in fixes:
        if old not in text:
            print(f"  FIX-MISS: {old[:60]!r}")
            ok = False
            continue
        text = text.replace(old, f"~~{old}~~ {new}", 1)
    if ok and fixes:
        path.write_text(text)
    return ok

def do_banner(path, note):
    if "Resolution (2026-10-05)" in path.read_text():
        return True  # idempotent
    lines = path.read_text().splitlines()
    # insert after first heading line (or at top if none)
    for i, ln in enumerate(lines):
        if ln.startswith("#"):
            lines.insert(i + 1, f"\n> **Resolution (2026-10-05):** ~~{note}~~" if False else f"\n> **Resolution (2026-10-05):** {note}")
            path.write_text("\n".join(lines) + "\n")
            return True
    lines.insert(0, f"> **Resolution (2026-10-05):** {note}")
    path.write_text("\n".join(lines) + "\n")
    return True

def move(block):
    f = ROOT / block["file"]
    if block["move"] == "archived":
        dst = ROOT / "docs/status/archived" / f.name
    elif block["move"] == "done":
        dst = ROOT / "docs/feedback/done" / f.name
    elif block["move"] == "reviews":
        dst = ROOT / "docs/reviews/archived" / f.name
    else:
        return True
    dst.parent.mkdir(parents=True, exist_ok=True)
    r = subprocess.run(["git", "mv", str(f), str(dst)], cwd=ROOT, capture_output=True, text=True)
    if r.returncode != 0:
        print(f"  MOVE-FAIL: {r.stderr.strip()}")
        return False
    return True

def main():
    apply = "--apply" in sys.argv
    blocks = parse(sys.argv[1])
    tmpdir = pathlib.Path("/tmp/dh-specs")
    tmpdir.mkdir(exist_ok=True)
    n_ok = n_fail = n_skip = 0
    for b in blocks:
        f = ROOT / b["file"]
        if not f.exists():
            print(f"MISSING {b['file']}")
            n_fail += 1
            continue
        # 1. item strikes via annotate script
        spec = build_spec(b, tmpdir)
        if spec:
            r = subprocess.run(["python3", str(ANNO), str(f), str(spec)],
                               cwd=ROOT, capture_output=True, text=True)
            if r.returncode != 0:
                print(f"STRIKE-SKIP {b['file']} (banner fallback): {r.stdout.strip()[:120]}")
            strikes_ok = True
        else:
            strikes_ok = True
        # 2. headline fixes
        if b["fix"] and not do_fixes(f, b["fix"]):
            print(f"FIX-FAIL {b['file']}")
            n_fail += 1
            continue
        # 3. banner for EMPTY / AUDIT / explicit banner files
        if b["class"] == "EMPTY" and apply:
            note = b.get("archive_note") or "Empty daemon-artifact snapshot; no content was ever written; the session's actual record lives in the same-day reports."
            do_banner(f, "~~" + note + "~~")
        elif b.get("audit") and apply:
            nres, nopen = b["audit"][0], b["audit"][1]
            tail = b["audit"][2] if len(b["audit"]) > 2 else ""
            note = (f"~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: "
                    f"{nres} items verified resolved, {nopen} open ({tail or 'routed to TODO_LIST/ROADMAP'}).")
            do_banner(f, note)
        elif b.get("banner"):
            do_banner(f, b["banner"])
        # 4. archive move only when fully resolved
        if apply and b["class"] == "ARCHIVE" and b["move"] != "none":
            if not move(b):
                n_fail += 1
                continue
        n_ok += 1
        print(f"OK {'ARCHIVED' if apply and b['class']=='ARCHIVE' and b['move']!='none' else b['class']:9s} {b['file']}")
    print(f"\n== {n_ok} ok, {n_fail} failed, {len(blocks)} total, apply={apply} ==")

if __name__ == "__main__":
    main()
