# Status Report — Full Arc: v0.7.1/v0.7.2 Release, BuildFlow Core Lane Live, Phase-2 Hardening

**Date:** 2026-09-25 09:43 CEST
**Repos:** `/home/lars/projects/art-dupl` (branch `fork`, clean, pushed to `origin/fork` at `69107492`) and `/home/lars/projects/BuildFlow` (branch `master`, clean, ahead of origin, NOT pushed).
**Scope:** This session only — execution of the entire SUPERB plan (`docs/planning/2026-09-24_18-04_artdupl-core-lane-buildflow-inversion.md`): A1–A8 (BuildFlow core lane), B1/B2/B4/B5/B6/B7, C1–C6, D4. Two CONCURRENT sessions worked the same repos in overlapping windows (a fan-out-verification session and an early-morning session that added `TestArtDuplProviderRegistered` and bumped BuildFlow to art-dupl v0.7.2); all collisions were reconciled and are called out below.

---

## Headline

**art-dupl is now the LIVE core duplication detector in BuildFlow, verified end-to-end; jscpd is the demoted non-Go backup lane; the entire `dupl-check`/`dupl_threshold`/`--semantic` ghost-config surface is deleted; the provider is production-hardened (.gitignore, concurrency, fixtures, gates); and a new AST gate caught and fixed a real latent Duration-in-v2 bug.** All gates that CAN run are green on both repos. The only red anywhere is environmental: `/run/binfmt` missing on this host blocks every sandboxed nix build (BuildFlow gotcha #190, needs one root command).

---

## a) FULLY DONE

1. **Release v0.7.1 cut and pushed** (A2): fixed the pre-release gate's cgo bug (`-race` requires `CGO_ENABLED=1`, matching the flake), repaired art-dupl's stale `vendorHash` (the CI-red blocker), cut the CHANGELOG section, waited for green CI on the exact commit, pushed annotated tag, verified `proxy.golang.org` serves `v0.7.1` and a clean-dir consumer `go get` works. v0.7.2 subsequently landed from the parallel session; BuildFlow pins v0.7.2.
2. **BuildFlow preflight** (A1): consistency script ALL GREEN (24 flake pins, go.work floor, build+vet root/tools/execution, quiescence); baseline captured (`list providers` showed jscpd, NO art-dupl — the ghost verdict confirmed at runtime).
3. **Wiring live** (A3): blank import in `sdk_imports.go` (comment adjusted by the concurrent session, reconciled), `require github.com/LarsArtmann/art-dupl` (v0.7.1 by me → v0.7.2 by them), `go work vendor`, vendorHash repair, build+vet green in root/tools/execution. Smoke run found 1210+ duplicate findings on BuildFlow's own tree — the detector is real.
4. **Compliance rail** (A4): `ToolArtDupl` constant added; consistency tests green; `TestArtDuplProviderRegistered` landed via the concurrent session (I verified it matches the detector-only contract rather than duplicating it).
5. **jscpd Go-lane demotion** (A5): `goFilePattern` out of `jscpdTriggerPatterns`, `go` out of `jscpdFormats`; backup-lane policy in the header comment; all four jscpd integration tests rewritten to JS fixtures and passing against the REAL jscpd binary (0.18s runs, not skips); report fixtures converted `"go"` → `"javascript"`; new `TestJscpd_ExcludesGoLane` pins the exclusion AND asserts art-dupl stays registered; live verification: Go-only repo → "no tools matched", JS-only repo → jscpd finds the clone.
6. **Ghost purge end-to-end** (A6): `ToolDuplCheck` constant, `providerlessToolMeta` row, `domain/config/dupl_threshold.go` (whole type), `config/koanf.go` key/env/defaults, `unknown_keys.go`, `materialize.go` block, `defaults.go`, `constants.DefaultDuplThreshold`, `--dupl-threshold` + `--semantic` CLI flags, `parseCoreConfigValues` middle return (now `(FileSize, Severity, error)`) with all error-context plumbing, `config show` display, `config init` + `setup` templates, validate block, `model.Config` field + validation entry, telemetry `duplThreshold` field + ctor param + `ToMap` key + doc line, `language.ToolDupl`, and both wrong install hints (`doctor_cmd.go` + `flake.nix` → `github.com/LarsArtmann/art-dupl`). Zero references remain outside git history.
7. **E2E verified against a real fixture repo** (A7): findings render via `--format finding` (rule `art-dupl/duplicate-code`, snippet, both occurrences, GroupID-based ids) and in the summary; severity **warning** stays below the default fail-on=error gate; result cache hits on unchanged trees and invalidates on file change (discovered: adding `go.mod` alone does NOT invalidate — keying is on analyzed-file content); `warnings_budget: {art-dupl: 0}` produces `[OVER budget 0 by 2]` + render collapse.
8. **Docs consistency** (A8): `docs --check` green after provider-count 115→116 updates across README/TODO_LIST/FEATURES/AGENTS (109 constructors + 7 toolsdk); AGENTS gotcha **#191** (core/backup lane split, full purge inventory, E2E ops notes) and gotcha **#73** pipeline addendum written; the gotcha-numbering audit guard caught my first insertion ordering and I fixed it — the ratchet worked.
9. **Provider .gitignore support** (B1): matcher extracted `cmd/gitignore.go` → `internal/gitignore` (git mv + cmd shim keeps call sites stable); arch-lint `gitignore` component added (cmd + provider mayDependOn); **found and fixed a doc-vs-code lie**: `LoadGitignore`'s comment claimed a down-walk for nested `.gitignore` files that the implementation never did — wrote `LoadTree(root)` doing real git-style nested discovery; provider crawl honors it with directory-level pruning; tests cover pattern+negation+nested override and the nil-matcher fallback.
10. **Concurrency + determinism suite** (B2): 8-goroutine parallel `Detect` under `-race` on one working dir; GroupID stability across independent runs (`TestFindingsFromGroups_GroupIDStableAcrossRuns`); suite green under `-race`.
11. **Provider polish** (B4): explicit self-documenting no-op `HealthCheck` in the Spec (pure-Go detector, nothing to probe); Spec Description now states the fixed contract ("semantic mode, threshold 5 statements (not configurable via the SDK)"); `stripEmptyMetadata` full-classification-key-set test (7 keys asserted gone, 3 interchange keys asserted kept); forward-slash/native-separator `IsIgnored` parity test; gogenfilter content-only test (generated file with unexpected filename filtered, hand-written survives).
12. **Multi-module fixture** (B6): 3 modules + workspace scaffolding through the full toolsdk path; cross-module clone MUST be found (it is); vendor copy, generated file, and dot-dir file MUST NOT appear (they don't).
13. **f#4 AST gate** (C1): new `internal/jsonv2gate` scanner test — fails on any production file importing `encoding/json/v2`/`jsontext` that also declares a wire-facing (`json:` tag ≠ `-`) `time.Duration` field; other production v2 importers require a reviewed rationale in the allowlist table; **canary-verified failing then passing** (a gate never observed failing is untested); wired via the default suite so CI + `nix flake check` get it with zero plumbing; AGENTS.md documents the gate and how to extend it.
14. **The gate earned its keep immediately**: it flagged `pkg/artdupl/types.go` — `Timeout time.Duration json:"timeout"` next to a direct v2 import, the exact latent "no default representation" crash class from 2026-09-23/24. Migrated to the v1 API (identical aux-struct pattern); tests green.
15. **Classification decision** (B5): ADR-0025 written — SDK findings stay classification-free (logic lives in the output layer by design, no consumer needs the keys, severity safety already solved at the boundary via the v0.7.2 cap, reversal is cheap).
16. **Monthly self-scan + SDK dogfood** (B7/C3): `scripts/self-scan.sh` ran 29 shown → **one group was MY OWN new code** (duplicated `len(rules)==0` tail in `LoadGitignore`/`LoadTree`) → extracted `matcherFromRules` → 28 shown; ledger section appended with per-group dispositions; then ran the provider path against art-dupl itself via a consumer-style module: **179 findings / 72 distinct groups** (GroupID-stable) vs CLI's 60 detected / 0 shown at `-t 5` — delta is crawl policy (provider deliberately scans `*_test.go`, CLI default-ignores them), documented in the ledger.
17. **Evidence re-verification** (C2): every durable go-line claim re-checked with plain rg (`go.mod 1.27.1`, `.golangci.yml 1.27.1`, flake `go_1_27`); "nicate" exists only as the quarantined incident description, no corrupted claims persist in docs; no corrections needed.
18. **Flipflop skip verified** (C4): `.buildflow.yml` `skip_steps: [go-mod-normalize]` present with rationale; the pin survived every pipeline run this session (`go 1.27.1` intact after multiple `buildflow -s` + preflight invocations); outcome recorded in TODO_LIST.
19. **HARVEST + docs** (C5/D4): TODO_LIST gained the "BuildFlow core-lane follow-ups (harvested 2026-09-25)" section with explicit done-not-re-listed inventory; FEATURES provider row updated (gitignore, severity cap, ADR-0025, LIVE status); HOW_TO_USE gained an SDK/provider section; README gained the toolsdk provider subsection with the blank-import snippet, the `ErrNoDuplicatesFound` → empty-findings contract (also documented on `FindClones` itself), and the BuildFlow consumer #1 story.
20. **lessons.md committed** (C6): the `rg -r` recidivism lesson written into crush-config `references/lessons.md` by commit (`9960540`), with the rule, the re-verification ritual, and the meta-lesson (summaries don't survive task load).
21. **Final gates**: art-dupl full suite (`go test -count=1 ./...`) green, `golangci-lint run ./...` 0 issues, `go-arch-lint check` green, AGENTS accuracy pass (gitignore bullet + provider bullet), pushed to `origin/fork`. BuildFlow full workspace suite green, audit ratchets green, `docs --check` green, golangci 0 issues in all touched modules, everything committed.

## b) PARTIALLY DONE

1. **BuildFlow nix gates** — environmentally blocked, not code-blocked: `/run/binfmt` is missing on this host, so `nix build .` and `nix run .#update-vendor-hash` fail with `getting attributes of path "/run/binfmt"` (gotcha #190, hit in real time — it ALSO broke art-dupl's previously-cached nix build when my source changes forced a fresh build). Root go.sum was repaired standalone (`GOWORK=off go mod tidy` — the v0.7.2 bump had updated only tools/go.sum) and `go work vendor` re-ran, but the **vendorHash after that tidy is UNVERIFIED**; first nix-capable run should expect one `update-vendor-hash` cycle. The one-time root fix is `sudo mkdir -p /run/binfmt` (tmpfs: lost on reboot) or SystemNix `boot.binfmt.emulatedSystems` (permanent).
2. **BuildFlow push** — all changes committed (daemon heuristic commits + my AGENTS/status-doc commits) but `master` is NOT pushed; needs your go-ahead, ideally after the nix gate can run once.
3. **System buildflow binary still stale** — PATH binary is NixOS-profile-owned (`buildflow-7e1fbfe`); I built and used `/tmp/bf-test` for all E2E work, but a fresh dogfood install requires your NixOS switch (which also currently can't evaluate nix builds until binfmt is fixed). Under the stale binary, a `warnings_budget` false-positive `unknown_keys` warning appears — schema is present in the tree, artifact of the stale binary.
4. **D1/D2/D3/D6 phase-3 items** — routed into TODO_LIST with precise entry criteria rather than half-executed: threshold knob + Spec.Timeout upstream (needs go-finding work), warnings_budget for art-dupl in BuildFlow's own `.buildflow.yml` (needs the stale binary replaced first), fleet lane-overlap query (needs post-soak), fleet oddments.
5. **D5 json/v2 migration slice** — materially advanced but not closed: the gate mechanizes the invariant and the 6 remaining production importers are allowlisted with reviewed rationale; the ~22 `_test.go` importers remain deliberately tolerated (string-only payloads) and are still tracked under the existing fleet-audit TODO item.
6. **B3 workspace-scale perf/timeout benchmark** — not executed; routed into the D1 Spec.Timeout TODO item (they share the fixture and the measurement).

## c) NOT STARTED (this session)

1. D1 toolsdk upstream issues/prototype (options channel, Timeout mapping, HealthCheck helper).
2. D2 warnings_budget entry in BuildFlow's own `.buildflow.yml` (blocked by the stale system binary).
3. D3 fleet lane-overlap measurement (post-soak by definition).
4. D6 fleet oddments triage (gogenfilter 9 skipped consumers, filepath.Separator sweep, branch protection, go-paperless release, Windows exe-start probe).
5. Fleet rollout of the core-lane blank-import pattern to other consumer repos.
6. README/HOW_TO_USE updates for the CLI-side story of the new lane split (docs cover the SDK side; the CLI quick-start still describes threshold flags a BuildFlow user will never touch — harmless but could note the pipeline defaults).

## d) TOTALLY FUCKED UP

1. **I built an E2E fixture that art-dupl at its DEFAULT threshold correctly does not flag — then went hunting for a wiring bug.** A 10-line two-branch clone counts as 4 composite statements (ADR-0023 subsumed trimming), below `-t 5`; I only believed the tool after art-dupl's own CLI also reported 0 and a threshold sweep showed the unit counting 4. The wiring was fine the whole time; my fixture was wrong twice (the first version was even smaller). Cost: several debug loops across two repos, plus a `--format finding` run against a 100%-cache-hit empty result that nearly sent me into the result-cache code. The lesson is now in gotcha #191 (fixtures need ≥6 flat statements), but it was in my head the whole time as ADR-0023 semantics I didn't apply.
2. **Inherited-environment stumble**: art-dupl's direnv sets `GOWORK=off`, so `go work vendor` in BuildFlow failed with "no go.work file found" and I ran it twice before checking `go env GOWORK`. Two dead calls from not checking the env first.
3. **I nearly duplicated `TestArtDuplProviderRegistered`** — my edit failed only because the concurrent session had already committed the test minutes earlier. The collision was avoided by luck (tool error), not by process: a one-line grep for the symbol before writing would have been the discipline.
4. **Compile-by-trial against go-finding types**: four rounds of branded-type/arity errors (`RelatedRef.Position` not `.Location`, `FilePath` casts, `NewCLIUsage` arity in FOUR call sites across two files, `Spec.Detect` being an interface not a function). One read of `go-finding/finding.go` and the toolsdk Spec would have prevented all of them; I fixed forward instead of reading first.
5. **The jsonv2gate file Frankenstein moment**: I patched the test file with sed/python and the edit tool interchangeably, leaving mixed identifiers (`importsV2`/`risk`/`durationRisk`) across three repair rounds before rewriting the file cleanly. Rewrite-first was the obvious move after the second failed patch.
6. **First jscpd lane-check was aimed wrong**: I ran `buildflow -s jscpd` from the BuildFlow root instead of the fixture, misread "41 files, 2 formats" as the tool firing on my fixture (it was scanning BuildFlow's own tree — which incidentally proved Go exclusion, but by accident, not by verification).
7. **Sloppy first-draft test code**: a dead `_ = clone` var, two over-long fixture strings golines flagged, and a `writeFile` name collision with an existing package helper (caught by vet). All cleaned, but each was preventable by looking at the surrounding package before writing.
8. **Gotcha-numbering insertion violated the repo's own audit guard** (191 placed before 190) — the guard caught it, which is good, but I inserted into a numbered sequence without checking the sequencing test that exists precisely for this.

## e) WHAT WE SHOULD IMPROVE

1. **Concurrent-session protocol worked but was luck-adjacent**: three sessions in overlapping windows on the same repos produced zero conflicts ONLY because the meta-guards (registration-test meta-guard, numbering audit, consistency tests, edge snapshot) caught every collision. Make a symbol-existence grep a mandatory pre-write step for any test/doc the plan says "add" — the plan cannot know what a parallel session already did.
2. **Fixture-first discipline**: validate any E2E fixture against the CORE tool (art-dupl CLI) BEFORE testing the integration layer; a fixture that cannot fire poisons every downstream conclusion.
3. **Read the interchange types before writing consumer code**: go-finding's `Finding`/`RelatedRef`/`Position` shapes and `toolsdk.Spec` field types are small; one read each eliminates the compile-by-trial loop that ate the most time this session.
4. **One edit channel per file**: alternating python/heredoc and the edit tool caused three "modified since read" conflicts; pick per file and stay with it.
5. **The binfmt environment fix should be permanent, not per-reboot**: `boot.binfmt.emulatedSystems = ["aarch64-linux"]` in SystemNix recreates `/run/binfmt` at every activation; the mkdir stopgap dies with tmpfs. Until then, every nix build on this host is a coin flip and vendorHash trust is deferred (gotcha #190 corollary).
6. **Result-cache keying deserves user-facing docs**: "analyzed-file content only; manifest/`go.mod` changes don't invalidate a detector's cached result; use `--no-result-cache-for`" is currently gotcha-only knowledge that cost me a debug loop.
7. **The provider's `_test.go` crawl policy should live in the Spec description**, not just the ledger — it is a visible 60-vs-72 group delta between CLI and pipeline channels and will be "discovered" repeatedly otherwise.
8. **Status-report evidence hygiene held**: plain-rg-only discipline was kept all session (zero `rg -r` uses), and the C2 re-verification pass confirmed no corrupted claims persisted — keep the ritual now that the lesson is durably committed.

## f) Up to 50 things we should get done next

**P0 — unblock the gates (needs you/root)**
1. `sudo mkdir -p /run/binfmt` (instant stopgap) and/or SystemNix `boot.binfmt.emulatedSystems` (permanent) — every nix gate on this host depends on it.
2. After binfmt: run `nix run .#update-vendor-hash` in BuildFlow and `nix build .` in both repos; expect exactly one hash cycle from my post-tidy go.sum change.
3. Push BuildFlow `master` once the nix gate is green (7+ commits ahead, includes the lane inversion).
4. Rebuild/install the system buildflow binary via NixOS switch (kills the stale-binary `unknown_keys` false positive and dogfood BLOCKING warnings).
5. Confirm BuildFlow CI green on the pushed HEAD (nix jobs were blocked locally; runners have working binfmt).

**P1 — core-lane completion**
6. D2: telemetry-derived `warnings_budget: {art-dupl: N}` in BuildFlow's own `.buildflow.yml` (the tree shows 1210+ findings; gotcha #184 pattern).
7. D1a: toolsdk upstream issue + prototype — options/config channel so the 5-statement threshold is configurable without a release.
8. D1b: toolsdk `Spec.Timeout` → BuildFlow timeout mapping gap (measure art-dupl at 33-module workspace scale first — B3 benchmark, numbers into `docs/benchmarks/`).
9. D1c: toolsdk explicit no-op HealthCheck helper (art-dupl hand-rolled one; upstream it).
10. Provider Spec Description: document the `_test.go`-inclusive crawl policy (the 60-vs-72 CLI delta).
11. Release v0.8.0 with the B1–B6 hardening (gitignore, gate, fixtures, ADR-0025) so BuildFlow pins the hardened provider; then bump BuildFlow and re-run its E2E.
12. Post-release: `pkg.go.dev` render check for `pkg/provider` + `internal/gitignore` doc rendering.
13. Re-run BuildFlow's own gotcha-#73 baseline: the pipeline now surfaces clone findings continuously; re-verify the manual-sweep numbers still match the ledger after the provider's `_test.go`-inclusive crawl is visible in dogfood.
14. Add `art-dupl` to BuildFlow's `list providers` snapshot test/docs example counts (docs --check covers claims; a registry-snapshot test may need the count bump ratcheted).
15. Verify result-cache invalidation ALSO covers provider version bumps (binaryHash path, gotcha #122) so v0.7.2→v0.8.0 upgrades don't serve stale findings.

**P2 — fleet**
16. Fleet rollout recipe doc: blank import + registration test + jscpd lane split + `TestJscpd_ExcludesGoLane` mirror — one copy-paste section for any BuildFlow consumer repo.
17. D3: fleet lane-overlap query (jscpd non-Go findings vs art-dupl Go findings per repo, post-soak) → verdict on whether jscpd's lane should shrink further.
18. gogenfilter: sweep the 9 skipped consumers still on ≤v3.6.0 (existing TODO item; go-filewatcher's SQLC test failure flagged as the first check).
19. Fleet `filepath.Separator` sweep (existing TODO #15).
20. Branch-protection checklist for the repos flagged in TODO (#18).
21. go-paperless release decision (#20).
22. Windows exe-start probe (#4).
23. Check whether any fleet `.buildflow.yml` still carries `dupl_threshold:` (now an honest unknown-key warning) and clean them.
24. Announce the lane split in the BuildFlow docs site (provider count changed; the duplication story changed materially).
25. Consider a `deprecatedToolAliases` entry for `dupl-check` in BuildFlow if any fleet config names it (purge assumed none exist — verify, don't assume).

**P3 — art-dupl polish**
26. Migrate the ~22 tolerated `_test.go` direct v2 importers to the v1 API (gate already prevents new ones; this is debt cleanup).
27. `baseline/baseline.go` + `config/config_migrate.go`: reassess whether the two non-test production v2 allowlist entries can be migrated (baseline is a wire path; migrate if golden tests confirm byte-identity).
28. Provider: Windows live-crawl test (slash parity is unit-tested; a `GOOS=windows` compile + CI matrix check closes the loop).
29. Provider: timeout/size-bound test at BuildFlow workspace scale (shared fixture with item 8).
30. `--explain`/actionability parity note: document in HOW_TO_USE that pipeline findings lack classification by design (ADR-0025 is in docs/adr; surface it where users hit the delta).
31. Self-scan cadence: next sweep due ~2026-10-22 (28 shown, all accepted classes; the ledger's table held).
32. Extend `TestJscpd_ExcludesGoLane`'s sibling idea: an art-dupl-side test pinning that the provider's Spec NEVER gains a `go`-format jscpd overlap (cross-repo guard is impossible; a doc-level invariant may be enough — decide).
33.HOW_TO_USE: add the composite-statement fixture caveat for anyone writing E2E fixtures (the gotcha #191 lesson, user-facing).
34. SDK_DESIGN.md: reflect ADR-0025 + the provider in the SDK architecture diagram.
35. FEATURES.md: the "WORTH CONSIDERING" section may need the lane-split story folded in (verify on next docs-health pass).
36. CHANGELOG: the B1–B6 work is still under `[Unreleased]`-equivalent (post-v0.7.2); keep it curated for the v0.8.0 cut.
37. Benchmarks: commit the multi-module fixture generator if B3 lands (docs/benchmarks note per repo convention).
38. Consider surfacing the provider's gitignore behavior in `--format finding` metadata (e.g. `art-dupl/gitignore-filtered-count`) — optional observability.
39. Audit ratchet parity: BuildFlow's `TestModuleFanOut_*` AST ratchets now cover art-dupl indirectly; art-dupl could mirror the same AST-scan style for its own Spec invariants (Trigger language list, no `go` in jscpd — that one lives BuildFlow-side only).

**P4 — process**
40. Add the symbol-existence pre-check to the personal pre-write checklist (concurrent sessions are the norm, not the exception).
41. Write the "fixture must fire in the core tool first" rule into the E2E section of TESTING.md.
42. Prune `/tmp` fixture repos (`dup-e2e`, `go-only`, `js-only`, `dupfixture`, `sdkself`, `release-verify`, `/tmp/bf-test`) when the disk sweep next runs.
43. crush-config: home-manager activation to deploy the new lessons.md entry (commit exists; delivery layer needs the switch).
44. Record the three-session-collision experience as a lessons.md candidate IF it recurs (guards worked; no action yet).
45. Revisit gotcha #190's SystemNix fix as a proper PR to the SystemNix repo (needs your authorization — it's a system config repo).
46. TODO_LIST: the harvested D1/D2/D3 items need owners/dates at the next planning pass (they are entry-criteria-gated, not abandoned).
47. Check whether BuildFlow's `verify-config` output should list skipped steps (the go-mod-normalize skip is invisible in its output — made C4 verification harder than needed).
48. art-dupl: consider a CHANGELOG link from the provider Spec Description (consumers see it in `--list`).
49. Sweep stale LSP diagnostics cache after this session's many file moves (git mv + rewrites) — gopls may serve pre-move state.
50. Close the loop on the 2026-09-24_13-30 report's ANNOTATE item: that report's "coexist" framing is superseded by the core/backup verdict; a one-line annotation pointing at the 2026-09-25 reports would finish the docs-health arc.

## g) Questions I cannot figure out myself

1. **Push policy for BuildFlow `master`**: push now with the nix gates pending (runners have working binfmt, so CI would be the verification), or hold until you've run the binfmt fix + one local `nix build .`/`update-vendor-hash` cycle? The vendorHash after my root go.sum repair is the one unverified artifact in the tree.
2. **The binfmt root fix**: do you want me to draft the SystemNix change (`boot.binfmt.emulatedSystems = ["aarch64-linux"]` per gotcha #190) as a ready-to-apply module diff for your review, or do you own NixOS config changes end-to-end yourself?
3. **Provider test-file policy**: should the provider's crawl SKIP `*_test.go` like the CLI's default (aligning the 72-vs-60 group delta between pipeline and CLI channels), or keep scanning tests (jscpd parity, catches test-code clones) as implemented? I documented it, but it is a real product decision I can't make from evidence alone.

---

*Point-in-time snapshot; this report's (f) section is the HARVEST input (items 1–25 already routed into TODO_LIST/ROADMAP where durable). Waiting for instructions.*
