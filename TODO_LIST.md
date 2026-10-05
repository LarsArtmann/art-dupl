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

- [ ] **Cut v0.8.0 (owner-gated)** — CHANGELOG `## [0.8.0] - 2026-09-29` complete, worktree rehearsal green; execution recipe parked at `docs/planning/2026-09-29_02-45_v0.8.0-release-park.md` (tag message, notes extraction, sequence; STOP gate = full suite + CI green on pushed HEAD). Same batch: go-paperless v0.4.3 (#20) and the branch-protection settings (#18).

### BuildFlow core-lane follow-ups (harvested 2026-09-25)

**Source:** `docs/planning/2026-09-24_18-04_artdupl-core-lane-buildflow-inversion.md` phases 2–3 and
`docs/status/2026-09-25_05-00_artdupl-core-lane-live-jscpd-demotion-ghost-purge.md` (BuildFlow side).

- [x] ~~**Provider threshold knob (D1) — art-dupl side**~~ DONE 2026-10-05: toolsdk options channel merged upstream (PR [#41](https://github.com/LarsArtmann/go-finding/pull/41), tagged `toolsdk/v1.14.0` — already the fork pin): `Spec.Options` declarations + `WithOptions`/`OptionsFromContext` per-run channel + `Spec.ValidateOptions` (unknown names fail loudly; kind-only checks; registration-time validation). Landed on fork via cherry-pick of `bb8b925e` (the parked `feat/provider-threshold-knob` was RED on its own `toolsdk v1.13.1` pin — its tests use v1.14.0 APIs; the cherry-pick re-based it onto fork's pins and completed the Spec Options block that the old auto-commit had lost). Coverage: default-parity (no option → 5), raised (20), typo-reject, range-reject 0 AND 1001 — all in `TestDetectThresholdOption`. Remaining (cross-repo, T18): BuildFlow maps consumer config to `WithOptions`.
- [x] ~~**Suppression surfacing — `emit-suppressed-accepted` provider option (plan T06)**~~ DONE 2026-10-05: `internal/accept` extracted from `cmd/accept_directive.go` (cmd keeps aliases; arch-lint `accept` component, cmd+provider mayDependOn); `printer/finding` adapter gained `Options.EmitSuppressedAccepted` + `Options.Accepted` predicate (adapter stays decoupled from `internal/accept`); accepted groups → `Suppression{Kind: InSource, Rule: RuleCloneDetected, Reason: "//art-dupl:accept directive"}` on every finding of the group. Provider declares the `emit-suppressed-accepted` bool option (default false = byte-identical, no scanning — pinned by `TestEmitSuppressedAcceptedOffIsByteIdentical`); SARIF round-trip incl. suppression via `ToSARIFWithOpts(WithIncludeSuppressed())` pinned by `TestSuppressionRoundTripsViaSARIFIncludeSuppressed` (G7); E2E by `TestDetectEmitSuppressedAcceptedOption` (real fixture with a directive, wrong-kind ValidateOptions reject). Default-flip decision deferred: revisit after BuildFlow consumer feedback (status report §g Q2).
- [x] ~~toolsdk Spec.Timeout mapping (D1)~~ DONE 2026-09-29: measured `BenchmarkDetectWorkspaceScale` (new, env-gated, hermetic by default) over the real BuildFlow tree — 33 modules, 7,958 Go files, 1,217 findings — full Detect in **650 ms** / 295 MB allocs, ~100x under BuildFlow's 60 s budget. Verdict: **no Spec.Timeout field needed** — the SDK honors ctx cancellation end to end, so consumers bound with `context.WithTimeout`; recorded with reproduction steps in `docs/benchmarks/toolsdk-detect-workspace-scale-2026-09-29.txt`.
- [x] ~~Warnings budget for art-dupl in BuildFlow's own .buildflow.yml (D2)~~ DONE 2026-09-29: observed 1217 findings via the fresh binary's provider lane; ceiling set to 1460 (+20%, single-run seed noted in the config — replace with the 14-day telemetry max after soak). Render collapse verified: `art-dupl 1217 finding(s) reported - within budget (1460)`. Committed on BuildFlow master locally; push with the next BuildFlow release train.
- [ ] **Lane-overlap measurement (D3, post-soak)** — after the fleet picks up the BuildFlow release: query jscpd non-Go findings vs art-dupl Go findings per repo; verdict on whether the backup lane pulls weight.
- [ ] **Fleet rollout of the core-lane pattern** — the same blank-import + registration-test + jscpd-lane-split recipe applies to any other BuildFlow consumer repo with Go duplication needs.

### Cross-repo / external (remaining from earlier harvests)

- [ ] **Sweep gogenfilter consumers to v3.6.1** — 9 consumers SKIPPED with pre-existing red baselines still carry ≤v3.6.0: branching-flow, BuildFlow, auto-deduplicate (build failures), erraudit, go-filewatcher, hierarchical-errors, Cyberdom, overview, project-discovery-daemon (test failures). Per-repo checklist (2026-09-29 diagnosis, M17): (1) raise the repo's `go` line to 1.27 FIRST — go-filewatcher's baseline does not even build on a 1.27 toolchain (`json.Marshal requires go1.27 (file is go1.26)`), so its `TestFilterGeneratedCode_SingleFilters/SQLC` "behavior" failure is UNVERIFIED — the go-line bump may clear the whole class; (2) bump `gogenfilter/v3` to v3.6.1; (3) `go mod vendor` where vendor/ exists (oxlint-auto-configure lesson: post-bump build catches stale `vendor/modules.txt`); (4) `go test ./...` green before push (F11: no bump on a red baseline — fix the baseline first when the fix IS the bump). The 09-19 report's list omitted hierarchical-errors; it is in the 9.
- [ ] **Fleet audit: `encoding/json/v2` imports / `format:` tags on Go 1.27** (#14) — art-dupl in-repo slice is DONE (v0.8.0: baseline + config_migrate migrated, jsonv2gate allowlist shrunk to enum/test-support). Check every OTHER LarsArtmann repo on Go 1.27 for the breakage class (go.dev/issue/71631).
- [ ] **Fleet audit: `filepath.Separator` matching + `strings.Split(_, ":")` path parsing** (#15) — art-dupl in-repo sweep is DONE (2026-09-29: zero Separator-matching or colon-path-parsing sites; crawl slash-normalization invariant holds). Sweep the OTHER repos (the Windows bug class that cost seven CI cycles).
- [ ] **Branch protection with required checks + failure notifications** (#18) — red CI sat unnoticed for 4 days; needs owner action on GitHub settings. Recipe (2026-09-29, SUPERB M20): (1) **art-dupl** — Settings → Branches → Add classic rule for `main` AND `fork`: _Require status checks to pass_, pick gates by exact job name: `Lint`, `Test (ubuntu-latest)`, `Test (macos-latest)`, `Test (windows-latest)` (matrix names appear in the picker only after the branch's first CI run), `Coverage`, `Nix Flake Check`, `Self-Analysis`, `Verify package boundaries`, `Check for new clones`, `Performance Regression`, `Verify no forbidden linters`; enable _Include administrators_; leave `fail-fast: false` in ci.yml as is so one red OS can't mask the others. (2) **Fleet** — same rule per repo with THEIR gate names; 11 repos have workflows (2026-09-29): go-sse, go-cqrs-lite, go-paperless, DiscordSync, go-filewatcher, gogenfilter, crush-config, BuildFlow, templ-components, go-error-family, monitor365. (3) **Notifications** — GitHub Settings → Notifications → Actions → _Send notifications for failed workflows_ (emails the pusher; the auto-daemon pushes as Lars, so red reaches him) plus a weekly `gh run list --limit 20` sweep until required checks make merge-on-red impossible.
- [ ] **Windows exe-start fix: confirm on windows CI** (#4) — ROOT-CAUSED 2026-09-29 (5th round, stdlib-source proof): `go build -o art-dupl-test` writes an extensionless file on windows (cmd/go appends `.exe` only without `-o`), and os/exec `findExecutable` refuses extension-less absolute paths (PATHEXT resolution → ErrNotFound); the discarded `_ = cmd.Run()` error surfaced as the misleading "ProcessState is nil". NOT Defender/AV — the probe's red (36509027090) was invalid evidence anyway (it excluded `RUNNER_TEMP` while the binary lives under `%TMP%`). Fixed: `-o art-dupl-test.exe`, start-errors now fatal via `errors.AsType[*exec.ExitError]`, windows skip deleted, Defender probe job deleted. Remaining: one green windows-lane `TestExitCodes_Process` run to confirm, then this line can be closed.
- [ ] **go-paperless: tag + release the findByName consolidation** (#20) — pushed 2026-09-19 (`04c32dc`), CI verifying; v0.4.3 decided (two fixes, no API surface); follow-the-recipe memo: `docs/planning/2026-09-29_03-10_go-paperless-release-memo.md` (verify repo state first — memo reflects 2026-09-29 HEAD).

### BuildFlow upstream (discovered 2026-09-23)

- [ ] Align `go-version-auto-configure` (raises the go line to the toolchain on `go get -u`) with `go-mod-normalize` (canonicalizes patch pins away) — the two dispositions fought through repo go.mods 5x on 2026-09-19/23 (BuildFlow preflight `workspace/go-line-flipflop`). **ROOT-CAUSED 2026-09-29 (observed live)**: there were TWO rewriters, not one — besides `go-mod-normalize`, the FORMAT pipeline's `go-version-auto-configure:repair` step also strips the patch even in fast build mode (that was the live flip source: every `buildflow format` by any session rewrote the line, and the auto-daemon packaged it). art-dupl now skips BOTH steps (verified: pin survives a full format run); the full diagnosis + knob design (`CanonicalizeOptions.RespectPatchFloor` + `formIssues` gate + the missing provider-config channel) is recorded on BuildFlow TODO BF2, where the fleet fix belongs.
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
