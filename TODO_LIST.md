# TODO List

**Last Updated:** 2026-09-23

Actionable items for the next 2-4 weeks. Completed work lives in `CHANGELOG.md`.
This file is OPEN work only — no completed, rejected, or resolved items.

Master plan `docs/planning/2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md`
is complete (24/26 done or no-go-verified). Remaining items below.

---

## MEDIUM Priority

### v0.7.0 release follow-ups (harvested 2026-09-19)

**Source:** `docs/status/2026-09-19_06-41_v0.7.0-release-go1.27-coherence-ci-recovery.md` section (f);
`(#N)` = that report's task number. Items already done during harvest (pkg.go.dev render, ADR-0024,
FEATURES/AGENTS/HOW_TO_USE updates, jsonutil unit tests, gogenfilter CHANGELOG entry, performance.yml
toolchain pins, Windows stdin-test skips) are NOT listed — they live in the CHANGELOG when released.

**Done 2026-09-23 (removed per the no-completed-items rule; details: CHANGELOG [Unreleased]):**
#12 self-clean ledger, #16 pre-release-check.sh, #19 auto-tag workflow, #22 corpus re-baseline,
#24 stale-shell doctor, #29 separator tables, #30 exhaustruct noise, #31 monthly self-scan routine,
#32 errors.AsType, #35 2026-08 VERIFY pass, #38 t.Chdir, #39 tagalign/nestif, #40 arch-lint in flake,
#48 dump-tokens templ verification.

- [ ] **Sweep gogenfilter consumers to v3.6.1** — DONE 2026-09-19 for all 14 clean consumers (8 direct + 6 indirect, incl. vendor sync in oxlint-auto-configure). **Follow-up:** 9 consumers were SKIPPED with pre-existing red baselines and still carry ≤v3.6.0 — branching-flow, BuildFlow, auto-deduplicate (build failures), erraudit, go-filewatcher, Cyberdom, overview, project-discovery-daemon (test failures). Fixing those baselines is its own task per repo; go-filewatcher's `TestFilterGeneratedCode_SingleFilters/SQLC` failure may be a stale gogenfilter-behavior assumption worth checking first.
- [ ] **Fleet audit: `encoding/json/v2` imports / `format:` tags on Go 1.27** (#14) — same breakage class as art-dupl's blocker (go.dev/issue/71631); check every LarsArtmann repo on Go 1.27. **art-dupl in-repo slice (2026-09-23):** ~30 files still import `encoding/json/v2`/`jsontext` directly (mostly `_test.go` and enum/string-only payloads — tolerated per AGENTS.md); migrate any file that must marshal Duration-bearing types to the v1 API (the `internal/jsonutil` regression class).
- [ ] **Fleet audit: `filepath.Separator` matching + `strings.Split(_, ":")` path parsing** (#15) — the Windows bug class that cost seven CI cycles; sweep test and product code.
- [ ] **Branch protection with required checks + failure notifications** (#18) — red CI sat unnoticed for 4 days; needs owner action on GitHub settings.
- [ ] **Root-cause Windows exe-start `ProcessState nil`; un-skip `TestExitCodes_Process`** (#4) — 3-attempt retry insufficient, runner refuses freshly built exes; logic covered by `TestExitCodeForError` meanwhile.
- [ ] **go-paperless: tag + release the findByName consolidation** (#20) — pushed 2026-09-19 (`04c32dc`), CI verifying; needs CHANGELOG + version decision via go-release.

### BuildFlow upstream (discovered 2026-09-23)

- [ ] Align `go-version-auto-configure` (raises the go line to the toolchain on `go get -u`) with `go-mod-normalize` (canonicalizes patch pins away) — the two dispositions fought through repo go.mods 5x on 2026-09-19/23 (BuildFlow preflight `workspace/go-line-flipflop`). Worked around in art-dupl via `.buildflow.yml` `skip_steps`; fleet fix belongs upstream.
- [ ] `golangci-lint-auto-configure` re-adds fleet-default linters that repos deliberately banned (exhaustruct, tagliatelle here) on every pipeline run, fighting per-repo guard scripts — the chronic nix `disabled-linters` red since 2026-09-19. Worked around via `skip_steps`; consider a "respect existing ban list" heuristic upstream.
- [ ] Rebuild the BuildFlow binary (stale: built at `7e1fbfe`, repo at later HEAD) — `go-structure-linter` currently emits false "unknown field" errors because its analysis packages are go1.26 against a go1.27 `go list`.
- [ ] Re-verify `nix-hash-fix` now that treefmt/disabled-linters are green — it failed 36/36 historically because unrelated red checks blocked the pipeline; confirm the vendorHash update for the 2026-09-22/23 go.sum bumps landed.
- [ ] pma auto-commit daemon blind spot: mid-edit files were committed unformatted several times during 2026-09-22/23 sessions; consider a post-format hook or debounce (BuildFlow skill anti-pattern note).

## PARKED: Explicit Entry Criteria (not amnesia)

Items live in ROADMAP/DEFERRED until their trigger fires. Triggers mirror
plan §5.

| Item (detail in ROADMAP / ADR)                                                            | Entry criterion                                                   |
| ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Suffix Array + LCP detector (ADR-0020)                                                    | T1 stage split shows suffix tree ≥30% of wall clock on real repos |
| Per-file offset map (8N→2N memory)                                                        | T1 + memory profile on a 100k-file corpus                         |
| Winnowing pre-filter                                                                      | A user actually hits 100k-file scale                              |
| int32 arena indices, []Pos pool                                                           | Profile shows pointer-chasing / search allocs dominant again      |
| Threshold cliff, `.art-duplignore`, `--ci-gate`, `--diff-baseline`, test-aware thresholds | Post-T10 corpus numbers define which UX lever pays first          |
| TS/Python support, LSP, watch mode, ML actionability                                      | Explicit user pull                                                |
| TypeAwareData restructure, branded NodeType, syntax/golang facade                         | Breaking-change windows only (major version)                      |

### Architecturally constrained (DEFERRED)

- [ ] Branded `NodeType int32`: prevents cross-package constant collision, but touches the gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] Hide `syntax/golang` behind facade: blocked by import cycle (`syntax/golang` imports `syntax` for Node; `printer/actionability` imports `syntax/golang` for AST constants).
- [ ] `sync.Pool` for `[]*Node` stream slices: NO-GO, measured 2026-08-16 — after the T13 arena, serialization costs 2 allocs/file (~0.03% of run allocs); pooling would add cross-package lifetime plumbing for no measurable win. Re-evaluate only if serialization allocations regress.
- [ ] `sync.Pool` for contextList `[]Pos` slices: rejected in ADR-0022 — slices transfer between contextLists via `append` (which may reallocate); lifetime tracking would out-complex the savings.
- [ ] Restructure `TypeAwareData` so `EraseHash` is collection-level: per-entry on `PreloadedAST`, validated at runtime with a warning (`job/incremental.go::SetTypeAwareData`). Collection-level type would enforce the invariant at compile time. Breaking change to `syntax/golang/typeinfo.go`.
