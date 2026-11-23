# Library Review Index

 Point-in-time utilization/adoption reviews of first-party (LarsArtmann)
libraries against art-dupl. One row per report: library, version reviewed,
date, verdict. Re-reviews UPDATE the row and bump the version-reviewed column
(each report carries its own header table with the exact checkout SHAs it
audited). **The dashboard (`libs/index.html`) is the CANONICAL verdict home**
for the run — stat cards, verdict table (score bands, per-row group
provenance), top opportunities, fleet argument, verify appendix. The index
table below is a derived convenience view; on any disagreement the dashboard
and the per-report headers win.

Citation gate: every report's `file:line` citations are machine-verified by
`scripts/citation-sweep.py` (art-dupl-relative, `<lib>/`-qualified cross-repo,
absolute). Exit 0 is the ship gate; re-run it after any report edit.

## Fleet run 2026-11-22/23 (art-dupl @ `cee67b4a` for A/B, `fb062def` for C)

| Library                                                       | Version reviewed (local checkout)                  | Date       | Verdict                                   | Score + inputs          | Report                                  |
| ------------------------------------------------------------- | -------------------------------------------------- | ---------- | ----------------------------------------- | ----------------------- | --------------------------------------- |
| go-finding (core)                                             | `v1.16.0` @ `5f568e058` (2026-11-22)               | 2026-11-22 | IN USE — 3 shipped capabilities unused    | 82/100 (15.5/19)        | [libs/go-finding.md](libs/go-finding.md) |
| go-finding/toolsdk                                            | `toolsdk/v1.16.0` @ `9c675c7` (2026-10-11)         | 2026-11-22 | IN USE — current pin                      | 88/100 (7/8)            | [libs/go-finding-toolsdk.md](libs/go-finding-toolsdk.md) |
| linter-autoconfigure-sdk                                      | `v0.8.1-7-gbfe96d7` @ `bfe96d7f` (2026-11-20)      | 2026-11-22 | NOT-ADOPTED (value lands host-side)       | 0/100 (0/10)            | [libs/linter-autoconfigure-sdk.md](libs/linter-autoconfigure-sdk.md) |
| go-policy-dsl                                                 | `v0.4.0-1-gee88a84` @ `ee88a84a` (2026-11-20)      | 2026-11-22 | NOT-ADOPTED (no policy author/engine)     | 0/100 (0/8)             | [libs/go-policy-dsl.md](libs/go-policy-dsl.md) |
| go-dag-app                                                    | `v0.1.1-1-gc2b6c99` @ `c2b6c99a` (2026-11-20)      | 2026-11-23 | NOT-ADOPTED (no DI/DAG layer)             | 0/100 (0 applicable)    | [libs/go-dag-app.md](libs/go-dag-app.md) |
| go-linter-sdk                                                 | `v0.6.1-8-g1694dcb` @ `1694dcb2` (2026-11-22)      | 2026-11-23 | NOT-ADOPTED (dual-SDK = split-brain)      | 50/100 (0.5/1)          | [libs/go-linter-sdk.md](libs/go-linter-sdk.md) |
| go-compliance                                                 | `v0.2.0-10-g01028e6` @ `01028e64` (2026-11-22)     | 2026-11-23 | NOT-ADOPTED (art-dupl is the supplier)    | 0/100 (0/1)             | [libs/go-compliance.md](libs/go-compliance.md) |

## Dashboards

| Dashboard                                        | Date       | Scope                          |
| ------------------------------------------------ | ---------- | ------------------------------ |
| [libs/index.html](libs/index.html)               | 2026-11-23 | The 7-library fleet run above  |

Related: `docs/research/2026-10-05_go-finding-deep-dive.html` (+ §05–§06
ledgers) remains canonical for WHY each historical go-finding decline was
verdict-ed; the 2026-11-22/23 fleet run supersedes it for CURRENT adoption
truth. Open work routed at `TODO_LIST.md` **T20** (go-finding adoption
deltas) and the struck **G8**/**T07b** rows.

## Refresh triggers

- go-finding: re-audit on every module train (per-module tags — check core,
  toolsdk, analysis, pipeline independently); immediately after T20 lands.
- The five NOT-ADOPTED libraries: re-audit only when their report's
  RECONSIDER-WHEN trigger fires (each report names it in its Verdict section).
