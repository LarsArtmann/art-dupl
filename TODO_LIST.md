# TODO List

**Last Updated:** 2026-09-18

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

### Detection granularity follow-ups (ADR-0023, 2026-09-14)

*(go-paperless follow-ups #16-#20 completed 2026-09-18: find-by-name family
consolidated into `findByName[T]`, counting-server fixture extracted to
`newNoRequestServer`, doc-example/client-test clones accepted with rationale,
type-aware-suppressed groups verified gone — type-aware ≡ semantic on the
current tree. Default `-t 2` run now shows 0 groups. Details: CHANGELOG.)*

- [ ] **Release ADR-0023**: minor version bump, tag, `go get` + pkg.go.dev verification. CHANGELOG migration note already written. (report #46)

### Carry-over questions

*(none — corpus drift 2665→2670 resolved 2026-09-18: concurrent go-cqrs-lite
edits during the engine release wave; art-dupl verified deterministic via
back-to-back byte-identical runs. See annotations in the 2026-08-16 status
reports.)*

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
