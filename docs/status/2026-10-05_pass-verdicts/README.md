# 2026-10-05 Docs-Health Pass — Verdict Record

Durable copy of the pass record that previously lived only in `/tmp/dh/`
(rescued 2026-10-06; the 09-28 pass lost its annotator the same way — this
dir is the vow kept).

## Contents

- `verdicts-all.txt` — concatenation of verdict waves v1–v8 (4,244 lines,
  ~2,500 per-item verdicts). Block format:

  ```
  FILE <relative/path>
  CLASS ARCHIVE|KEEP|SKIP|EMPTY
  MOVE archived|done|reviews|none
  FIX <exact old line> ||= <exact new line>
  BANNER <one-line resolution note>
  AUDIT <resolved> <open> [open-items-summary]
  ITEMS:
  <key>\t<DONE|OPEN|WONT|SUPERSEDED>\t<evidence>
  ```

  Keys are agent-derived (e.g. `f1`, `b@T19`, `1@substring`) and do NOT
  always match the annotator grammar — that mismatch is why per-item strikes
  failed wholesale (see the pass report's §d1). When re-keying for the
  A-lite→A conversion, ALWAYS regenerate keys with
  `scripts/annotate-status-items.py --emit-keys <file> <lineno...>` against
  the CURRENT line numbers; the keys in this file are wave-time only.

- `scripts/docs-health-driver.py` (repo root `scripts/`) — the driver that
  consumed these blocks: banners under the first heading, FIX literal
  replaces, item strikes via the annotator, `git mv` for archive moves.

## Provenance

- Pass report: `docs/status/2026-10-06_01-42_docs-health-pass-self-review.md`
- Archive manifest: `docs/status/archived/MANIFEST-2026-10-05.md`
- Sweep commits: `d34aad4b` (194 files: annotations + 55 `git mv`s),
  `cc67cf0a` (living docs + manifest)

## Status

Kept for auditability of the 55 archive moves and as the verdict source for
the A-lite→regime-A re-keying follow-up (P0#3). Long-term retention is an
open owner question (§g Q2 of the pass report); deleting this dir only after
that answer says so.
