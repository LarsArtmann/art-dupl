# TODO List

**Last Updated:** 2026-09-29

Actionable items for the next 2-4 weeks. Completed work lives in `CHANGELOG.md`.
This file is OPEN work only — no completed, rejected, or resolved items.

Master plan `docs/planning/2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md`
is fully resolved (audited 2026-09-28): 24 of 26 tasks executed, 2 no-go-verified by
evidence — **T7** (defer-cleanup pattern: covered by the existing raii-defer/defer-call
patterns; adding an arbitrary-resource variant measured zero benefit at real
over-suppression risk, corpus re-validation 2026-09-23) and **T14** (`[]*Node`
stream-slice `sync.Pool`: measured ~0.03% of run allocations after the T13 arena —
`CHANGELOG.md` "Arena node serialization" + PARKED tier below). Nothing open.
Remaining items below are from later plans.

The 2026-09-28 docs-health harvest was EXECUTED 2026-09-29: the v0.8.0 release
carries the dead-directive detector, the Go 1.27 selector-key fix, the go-line
pin gate, the SARIF→go-finding integration test, the provider test-gap bundle,
the Go-1.27 dogfood script, and all doc gaps. Remaining items are cross-repo
or external by nature.

---

## MEDIUM Priority

### BuildFlow core-lane follow-ups (harvested 2026-09-25)

**Source:** `docs/planning/2026-09-24_18-04_artdupl-core-lane-buildflow-inversion.md` phases 2–3 and
`docs/status/2026-09-25_05-00_artdupl-core-lane-live-jscpd-demotion-ghost-purge.md` (BuildFlow side).

- [ ] **Provider threshold knob (D1)** — toolsdk upstream: options/config channel on `Spec` so BuildFlow users can raise/lower the fixed 5-statement threshold without a release; prototype + PR in go-finding.
- [ ] **toolsdk Spec.Timeout mapping (D1)** — BuildFlow maps no timeout from Spec; measure art-dupl at BuildFlow workspace scale (33-module fixture benchmark, numbers into docs/benchmarks/) and decide whether a Spec-level timeout field is needed upstream.
- [x] ~~Warnings budget for art-dupl in BuildFlow's own .buildflow.yml (D2)~~ DONE 2026-09-29: observed 1217 findings via the fresh binary's provider lane; ceiling set to 1460 (+20%, single-run seed noted in the config — replace with the 14-day telemetry max after soak). Render collapse verified: `art-dupl 1217 finding(s) reported - within budget (1460)`. Committed on BuildFlow master locally; push with the next BuildFlow release train.
- [ ] **Lane-overlap measurement (D3, post-soak)** — after the fleet picks up the BuildFlow release: query jscpd non-Go findings vs art-dupl Go findings per repo; verdict on whether the backup lane pulls weight.
- [ ] **Fleet rollout of the core-lane pattern** — the same blank-import + registration-test + jscpd-lane-split recipe applies to any other BuildFlow consumer repo with Go duplication needs.

### Cross-repo / external (remaining from earlier harvests)

- [ ] **Sweep gogenfilter consumers to v3.6.1** — 9 consumers SKIPPED with pre-existing red baselines still carry ≤v3.6.0: branching-flow, BuildFlow, auto-deduplicate (build failures), erraudit, go-filewatcher, Cyberdom, overview, project-discovery-daemon (test failures). Fixing those baselines is its own task per repo; go-filewatcher's `TestFilterGeneratedCode_SingleFilters/SQLC` failure may be a stale gogenfilter-behavior assumption worth checking first.
- [ ] **Fleet audit: `encoding/json/v2` imports / `format:` tags on Go 1.27** (#14) — art-dupl in-repo slice is DONE (v0.8.0: baseline + config_migrate migrated, jsonv2gate allowlist shrunk to enum/test-support). Check every OTHER LarsArtmann repo on Go 1.27 for the breakage class (go.dev/issue/71631).
- [ ] **Fleet audit: `filepath.Separator` matching + `strings.Split(_, ":")` path parsing** (#15) — art-dupl in-repo sweep is DONE (2026-09-29: zero Separator-matching or colon-path-parsing sites; crawl slash-normalization invariant holds). Sweep the OTHER repos (the Windows bug class that cost seven CI cycles).
- [ ] **Branch protection with required checks + failure notifications** (#18) — red CI sat unnoticed for 4 days; needs owner action on GitHub settings.
- [ ] **Root-cause Windows exe-start `ProcessState nil`; un-skip `TestExitCodes_Process`** (#4) — 3-attempt retry insufficient, runner refuses freshly built exes; logic covered by `TestExitCodeForError` meanwhile. Needs a Windows CI runner experiment.
- [ ] **go-paperless: tag + release the findByName consolidation** (#20) — pushed 2026-09-19 (`04c32dc`), CI verifying; needs CHANGELOG + version decision via go-release.

### BuildFlow upstream (discovered 2026-09-23)

- [ ] Align `go-version-auto-configure` (raises the go line to the toolchain on `go get -u`) with `go-mod-normalize` (canonicalizes patch pins away) — the two dispositions fought through repo go.mods 5x on 2026-09-19/23 (BuildFlow preflight `workspace/go-line-flipflop`), and the flip recurred 2026-09-28 via auto-commit before the new pin gate caught the class. Worked around in art-dupl via `.buildflow.yml` `skip_steps`; fleet fix belongs upstream.
- [ ] `golangci-lint-auto-configure` re-adds fleet-default linters that repos deliberately banned (exhaustruct, tagliatelle here) on every pipeline run, fighting per-repo guard scripts — the chronic nix `disabled-linters` red since 2026-09-19. Worked around via `skip_steps`; consider a "respect existing ban list" heuristic upstream.
- [x] ~~Rebuild the BuildFlow binary~~ DONE 2026-09-29: `nix build .` green at current HEAD, `nix run .#reinstall` reports the new build; the PATH binary (`~/.nix-profile`, root-owned) still resolves the old e881e96 until the next SystemNix/home-manager rebuild — use `~/projects/BuildFlow/result/bin/buildflow` meanwhile. Original symptom (go-structure-linter false "unknown field" from go1.26 analysis against a go1.27 `go list`) to re-verify after the profile swap.
- [ ] Re-verify `nix-hash-fix` — PARTIALLY verified 2026-09-29: the tool now runs (not the historical 36/36 block), detected a REAL staleness (new `ultraviolet` pseudo-version dep without a hash update), and its `--fix` ran `go mod tidy` (GOWORK=off) across all three module dirs; the vendorHash recompute itself is blocked on the concurrent session's in-flight dependency changes settling. Re-run `-s nix-hash-fix --fix` once the tree stabilizes and commit the hash.
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
