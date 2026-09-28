# docs/status/archived/ — Archive Regimes and Gates

Point-in-time snapshots (status reports, plans) whose items are fully resolved
or explicitly routed. Archived ≠ deleted: every file here is a historical
record; nothing in it is re-litigated.

**Last counted: 2026-09-28 — 329 files, 35 in regime A, 294 in regime B.**

## The two regimes

### Regime A — inline-annotated (the standard since 2026-09-28)

Every numbered item in the file carries an inline verdict:

```markdown
12. ~~Fix warmup store pollution~~ done at `a7b8159`
```

Variants: `done at <hash/date — evidence>`, `Won't implement — <reason>`,
`parked — <where>`. Open items are left bare — absence of a marker IS the
"open" signal. Files qualify for archiving only when EVERY item is resolved
or routed; the annotation happens BEFORE the `git mv`.

### Regime B — legacy (pre-2026-09-28, 294 files)

Files archived before the strikethrough regime existed. Their items carry NO
verdicts; a reader cannot tell from the file whether it shipped. The pass
that consolidated them (commit `32f12cad`) deliberately did NOT retrofit
annotations — striking items without per-item verification would fabricate
evidence, which is worse than silence.

## Retrofit policy (the g1 decision, current state)

Retrofitting regime B is **not scheduled**. It would cost roughly 2–3 focused
docs-health sessions (≈30 files per session at the verified per-item protocol:
read item → grep/CHANGELOG-verify → strike with evidence → check-rows gate).
The verdict on whether that calendar is ever spent rests with Lars (open
question g1 in `docs/status/2026-09-28_21-54_docs-health-full-audit-pass.md`).
Until an explicit go, the honest reading instruction for regime B files is:
**treat every unmarked item as status-unknown, not as open** — route anything
that matters into TODO_LIST instead of trusting the file.

## Gates (run over every batch after annotating)

```bash
# Presence: every file in the touched set carries at least one strikethrough
grep -rLn '~~' docs/status/archived/   # must print NOTHING for regime-A scope

# Uniformity: no PARTIAL table rows (mixed struck/unstruck) — the planted-miss class
python3 ~/.config/crush/skills/docs-health/assets/check-rows.py <file...>
```

Batch annotation tooling: `scripts/annotate-status-items.py` (atomic per file;
spec grammar documented in its header). ALWAYS dry-run a new file shape on a
/tmp copy first.

## Relationship to `docs/archive/` (top-level)

Two directories, two jobs — do NOT merge them:

| Directory               | Holds                                                            | Example                                                      |
| ----------------------- | ---------------------------------------------------------------- | ------------------------------------------------------------ |
| `docs/archive/`         | Superseded **design documents and proposals** (not dated status) | `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md`, `BDD_TESTS_REVIEW.md` |
| `docs/status/archived/` | Dated **point-in-time snapshots** (reports, plans, audits)       | `2026-08-16_13-13_wave-report.md`                            |

The naming is an unfortunate near-collision (`archive` vs `archived`) kept
because both directories are established and linked from other files; the
distinction above is what matters, not the rename.
