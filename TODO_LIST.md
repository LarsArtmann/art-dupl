# TODO List

**Last Updated:** 2026-08-16

Actionable items for the next 2-4 weeks. Completed work lives in `CHANGELOG.md`.
This file is OPEN work only — no completed, rejected, or resolved items.

Master plan `docs/planning/2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md`
is complete (24/26 done or no-go-verified). Remaining items below.

---

## HIGH Priority

### Timing evidence for ADR-0022 (parked on machine load)

**Source:** master plan T2.2 + T23; `docs/status/2026-08-16_12-27_master-plan-wave2-flake-green.md`.
**Entry criterion:** machine idle — `uptime` load sustained < ~4, no foreign nixbld/govulncheck/service jobs.
Never compare pinned and unpinned runs taken hours apart (machine state dominates; only interleaved A/B counts).

- [ ] **T2.2: Interleaved pinned/unpinned A/B** — `taskset -c 0-7,16-23` vs full machine, ≥6 alternations on `FindDuplOver/threshold_10` + `par4/tokens_10000`; annotate `docs/benchmarks/baseline-2026-08-16-v3_notes.md` (replaces the "still open, see TODO_LIST" note).
- [ ] **T23: `perf stat` cache-miss A/B vs `23fa1b4f`** — `git worktree add /tmp/artdupl-23fa1b4f 23fa1b4f`; counters `cache-references,cache-misses,LLC-load-misses` on suffixtree benches; verdict into baseline notes + ADR-0022 addendum.
- [ ] Pinned `benchstat` for `BenchmarkFindTranBoundary8/9`; record in `docs/benchmarks/` and as a comment on `linearScanMax` justifying the cutoff.

## MEDIUM Priority

### Wave-3 quality follow-ups

**Source:** `docs/status/2026-08-16_13-13_master-plan-wave3-signal-hardening.md` §f.

- [ ] **BDD/PTY test for HTML auto-write**: current tests inject the TTY probe; a real-PTY end-to-end test (via `script(1)` or a pty lib) would lock the full path.
- [ ] **Decide + document `--quiet` vs zero-match warning**: should `WarnUnmatchedExcludePatterns` respect `--quiet`? Currently unconditional.
- [ ] **Symmetric zero-match warning for `--include-pattern`**: same misconfig trap as `--exclude-pattern` (T12), opposite direction.
- [ ] **Coverage script: filter test-only packages**: `scripts/check-coverage.sh` reports 0.0% rows for test-only packages — noise.
- [ ] **TESTING.md: property/parity-test conventions**: document the reference-implementation parity pattern (T20) + coverage-guard-in-test pattern (assert the interesting branch actually ran).
- [ ] **a11y: focus style for `.anchor-link`**: anchor perm links have hover styling but no visible keyboard focus ring.
- [ ] **JSON output: expose stable anchor id**: cross-format linking (HTML deep link ↔ JSON record) needs the `AnchorID` in JSON too.
- [ ] **`--html-out`: mkdir -p parent dir** on demand instead of erroring on a missing directory.

### Carry-over questions

- [ ] **Corpus drift 2665 → 2670**: which feedback-corpus clone count grew by 5 between validations?
- [ ] **Budget ratchet**: `MemoryUsage` measures 4783 < 4802 budget — tighten the budget or leave headroom for map-iteration-order variance?

## PARKED: Explicit Entry Criteria (not amnesia)

Items live in ROADMAP/DEFERRED until their trigger fires. Triggers mirror plan §5.

| Item (detail in ROADMAP / ADR)                                                            | Entry criterion                                                   |
| ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Suffix Array + LCP detector (ADR-0020)                                                    | T1 stage split shows suffix tree ≥30% of wall clock on real repos |
| Per-file offset map (8N→2N memory)                                                        | T1 + memory profile on a 100k-file corpus                         |
| Winnowing pre-filter                                                                      | A user actually hits 100k-file scale                              |
| `int32` arena indices, `[]Pos` pool                                                       | Profile shows pointer-chasing / search allocs dominant again      |
| Threshold cliff, `.art-duplignore`, `--ci-gate`, `--diff-baseline`, test-aware thresholds | Post-T10 corpus numbers define which UX lever pays first          |
| TS/Python support, LSP, watch mode, ML actionability                                      | Explicit user pull                                                |
| TypeAwareData restructure, branded `NodeType`, `syntax/golang` facade                     | Breaking-change windows only (major version)                      |

### Architecturally constrained (DEFERRED)

- [ ] **Branded `NodeType int32`**: prevents cross-package constant collision, but touches the gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: blocked by import cycle (`syntax/golang` imports `syntax` for Node; `printer/actionability` imports `syntax/golang` for AST constants).
- [ ] **`sync.Pool` for `[]*Node` stream slices**: NO-GO, measured 2026-08-16 — after the T13 arena, serialization costs 2 allocs/file (~0.03% of run allocs); pooling would add cross-package lifetime plumbing for no measurable win. Re-evaluate only if serialization allocations regress.
- [ ] **`sync.Pool` for contextList `[]Pos` slices**: rejected in ADR-0022 — slices transfer between contextLists via `append` (which may reallocate); lifetime tracking would out-complex the savings.
- [ ] **Restructure `TypeAwareData` so `EraseHash` is collection-level**: per-entry on `PreloadedAST`, validated at runtime with a warning (`job/incremental.go::SetTypeAwareData`). Collection-level type would enforce the invariant at compile time. Breaking change to `syntax/golang/typeinfo.go`.
