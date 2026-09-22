# TODO List

**Last Updated:** 2026-09-19

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

_(go-paperless follow-ups #16-#20 completed 2026-09-18: find-by-name family
consolidated into `findByName[T]`, counting-server fixture extracted to
`newNoRequestServer`, doc-example/client-test clones accepted with rationale,
type-aware-suppressed groups verified gone — type-aware ≡ semantic on the
current tree. Default `-t 2` run now shows 0 groups. Details: CHANGELOG.)_

_(Release completed 2026-09-18: v0.7.0 tagged at 6dd99aec, module proxy
propagated, `go get github.com/LarsArtmann/art-dupl@v0.7.0` verified with a
compile-and-run consumer, GitHub Release published.)_

### Carry-over questions

_(none — corpus drift 2665→2670 resolved 2026-09-18: concurrent go-cqrs-lite
edits during the engine release wave; art-dupl verified deterministic via
back-to-back byte-identical runs. See annotations in the 2026-08-16 status
reports.)_

### v0.7.0 release follow-ups (harvested 2026-09-19)

**Source:** `docs/status/2026-09-19_06-41_v0.7.0-release-go1.27-coherence-ci-recovery.md` section (f);
`(#N)` = that report's task number. Items already done during harvest (pkg.go.dev render, ADR-0024,
FEATURES/AGENTS/HOW_TO_USE updates, jsonutil unit tests, gogenfilter CHANGELOG entry, performance.yml
toolchain pins, Windows stdin-test skips) are NOT listed — they live in the CHANGELOG when released.

- [ ] **Sweep gogenfilter consumers to v3.6.1** — DONE 2026-09-19 for all 14 clean consumers (8 direct + 6 indirect, incl. vendor sync in oxlint-auto-configure). **Follow-up:** 9 consumers were SKIPPED with pre-existing red baselines and still carry ≤v3.6.0 — branching-flow, BuildFlow, auto-deduplicate (build failures), erraudit [formerly hierarchical-errors - same repo, listed twice in error until 2026-09-22], go-filewatcher, Cyberdom, overview, project-discovery-daemon (test failures). Fixing those baselines is its own task per repo; go-filewatcher's `TestFilterGeneratedCode_SingleFilters/SQLC` failure may be a stale gogenfilter-behavior assumption worth checking first.
- [ ] **Fleet audit: `encoding/json/v2` imports / `format:` tags on Go 1.27** (#14) — same breakage class as art-dupl's blocker (go.dev/issue/71631); check every LarsArtmann repo on Go 1.27.
- [ ] **Fleet audit: `filepath.Separator` matching + `strings.Split(_, ":")` path parsing** (#15) — the Windows bug class that cost seven CI cycles; sweep test and product code.
- [ ] **`scripts/pre-release-check.sh`** (#16) — codify `git ls-remote` tag-collision check, CI-green gate, and toolchain pinning (go-release skill Phase 0–4).
- [ ] **Branch protection with required checks + failure notifications** (#18) — red CI sat unnoticed for 4 days; needs owner action on GitHub settings.
- [ ] **Investigate the "Auto-tag on version change" workflow** (#19) — ensure it cannot fight manual release tags (v0.7.0 collision class).
- [ ] **go-paperless: tag + release the findByName consolidation** (#20) — pushed 2026-09-19 (`04c32dc`), CI verifying; needs CHANGELOG + version decision via go-release.
- [ ] **Root-cause Windows exe-start `ProcessState nil`; un-skip `TestExitCodes_Process`** (#4) — 3-attempt retry insufficient, runner refuses freshly built exes; logic covered by `TestExitCodeForError` meanwhile. (The stdin-cancel flakes were fixed 2026-09-19: the three `TestStdinFeed_*` tests are Windows-skipped with documented rationale, and the windows Test lane runs without `GOEXPERIMENT=jsonv2` to dodge the fatal-GC-crash class — CI green since `ea05eb1e`.)
- [ ] **Corpus re-baseline post-v0.7.0** (#22) — AGENTS.md corpus numbers predate ADR-0023 + Go 1.27.1; fresh validation pass then refresh numbers.
- [ ] **Self-clean decision ledger** (#12) — durable per-group record for the `-t 1` self-scan (true current count: **40 groups**, re-measured 2026-09-19; the report's "44" was a file-count heuristic). Decision + rationale per group, so future sessions don't re-litigate.
- [ ] **Make stale-shell failures graceful** (#24) — detect local go < go.mod requirement and print an actionable message (dart of the `GOTOOLCHAIN=local` trap).
- [ ] **Verify `--dump-tokens` positions on a templ file** (#48) — the position work targeted Go files; templ ranges inherit differently.
- [ ] **`nix flake check` should cover arch-lint locally** (#40) — the 2026-09-18 arch-lint break only surfaced in CI.
- [ ] __docs-health VERIFY pass over 2026-08-_ status reports_* (#35) — several claims now stale post-ADR-0023/1.27.
- [ ] **Small quality batch** — both-separator table cases for `shouldSkipPath` (#29); `t.Chdir` sweep (#38); confirm tagalign/nestif additions are wanted (#39); exhaustruct bdd-exclusion noise check (#30); monthly `-t 1` self-scan routine documented (#31); evaluate `errors.AsType[E]` in errors/marshal.go (#32).

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
