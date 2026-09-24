# Status Report — BuildFlow art-dupl Comparison: Ghost Wiring, Core-vs-Backup Verdict

**Date:** 2026-09-24 17:00 CEST
**Repo:** `/home/lars/projects/art-dupl` (branch `fork`, clean tree — daemon committed all session work by report time)
**Scope:** This session only: the user-directed BuildFlow comparison ("View and compare the current implementation of art-dupl in /home/lars/projects/BuildFlow"), the f#1 comment fix, and what the comparison exposed. Prior-session state is referenced only where the same arc continues.

---

## Headline

**BuildFlow has NO art-dupl implementation — only ghost wiring — and its only live duplication detector is jscpd.**
The user then delivered the architecture correction that reframes everything:

> **art-dupl is the CORE duplication detector in BuildFlow; jscpd is the BACKUP for languages art-dupl does not support yet** (JS/TS, Python, Rust, Bash, YAML, ...).

So the correct end state is not "coexist as peers" (my session's original recommendation — wrong priority): it is **promote art-dupl to the Go+templ lane and demote jscpd to the non-Go fallback lane**. Until BuildFlow blank-imports `pkg/provider`, art-dupl remains a ghost system: shipped, tested, lint-clean, and delivering zero value in the product it was built for.

---

## a) FULLY DONE

1. **BuildFlow art-dupl implementation surveyed with a complete evidence chain** (all clean `rg`, all claims file:line-referenced):
   - No art-dupl provider has ever existed: `git log --all -- 'tools/providers/dupl*'` → empty; no factory anywhere.
   - The live duplication detector is **jscpd** (`tools/providers/jscpd_provider.go:77`): external CLI, JSON report parsed from a temp file, **hardcoded `--min-lines 10 --min-tokens 80`** (`:107`), severity warning→error at 50+ lines, 2min timeout, not fan-out (scans `.` once). The hardcoded 80-token bar means BuildFlow's current Go duplication detection is far weaker than art-dupl's default (5 statements, Type 1/2/3 semantic).
2. **Ghost-wiring inventory** (config that validates, telemetrizes, and reaches no consumer):
   - `ToolDuplCheck = "dupl-check"` (`domain/config/tool_name.go:127`) — orphaned constant + metadata row + build-mode blocklist case; no provider.
   - `dupl_threshold` config key: min 10 / max 50 / default **30 tokens** (`domain/config/dupl_threshold.go`), materialized (`config/materialize.go:77`), validated (`internal/cli/config_cmd_validate.go:82`), sent to telemetry — **read by nobody**.
   - `DuplSemantic` / `--semantic` flag (`internal/cli/config_flags.go:34`) — read by nobody.
   - `language.ToolDupl = "dupl"` (`language/language_constants.go:34`) — a third distinct name.
   - **Three-name split brain**: `dupl` / `dupl-check` / `art-dupl`, none of them wired.
3. **Wrong install hints found (×2)**: `internal/cli/doctor_cmd.go:226` and `flake.nix:543` both say `go install github.com/afontainized/art-dupl@latest` — the org does not exist; the real module is `github.com/LarsArtmann/art-dupl`. Any user following the hint installs nothing.
4. **SDK consumption path verified ready** (the 7-tool precedent):
   - `tools/providers/sdk_imports.go`: 7 blank imports, art-dupl absent — the wiring is literally one import line.
   - `ToolFromSpec` (`sdk_registry.go:45`): wraps Detect via `detectorWithFileCandidates` (branching-flow-specific; art-dupl ignores it harmlessly) + `detectorWithWorkDir` (WorkingDir agreement + `Scanning` progress line, gotchas #12/#83).
   - **`healthCheckFromSDK(nil)` → `dtool.NoOpHealthCheck()`** (`sdk_registry.go:185-190`) — verified: art-dupl's Spec sets no HealthCheck and is still safe under `Tool.Validate()`. No bug.
5. **f#1 lying comment fixed** (`pkg/provider/provider.go`): both sites now state the truth — GroupID/severity/positions/snippets come from the shared `printer/finding` adapter, but classification metadata keys (`art-dupl/clone-type`, `-category`, `-priority`, `-actionability`, `-generics-*`) are **absent** because the SDK pipeline never computes them; the old "metadata byte-identical with the CLI's finding output" claim was false. `go test -count=1 ./pkg/provider/` ok, `golangci-lint run ./pkg/provider/` → 0 issues.
6. **toolsdk version path verified**: BuildFlow `toolsdk v1.13.0 // indirect`; art-dupl requires `v1.13.1`; **`toolsdk/v1.13.1` tag confirmed on the go-finding remote** (`git ls-remote`) — MVS resolution of the blank import is evidence-backed, not assumed.
7. **BuildFlow not mutated.** The only blocked verification (runtime `list providers`) was skipped rather than running `go mod tidy` in a repo I'm not authorized to touch; static evidence is conclusive on its own.

## b) PARTIALLY DONE

1. **Runtime confirmation of the ghost status** — `go run ./cmd/buildflow list providers` fails with "updates to go.mod needed" (BuildFlow tree unstable, daemon churn). Static evidence (no constructor exists + no blank import) proves `dupl-check` cannot be registered, but the runtime list check itself did not execute.
2. **g1 (wiring/naming decision)** — answered with facts + a recommendation, but the recommendation's core premise ("jscpd and art-dupl coexist") was **wrong priority**; corrected by the user mid-report to core/backup. The decision frame now exists; the decision itself still pends.
3. **Provider production-readiness for a CORE lane** — the provider is correct for its shipped scope (semantic mode, fixed threshold, own crawl with gogenfilter) but was built as "an additional detector": no .gitignore support, no classification metadata, no explicit HealthCheck, no timeout/size-bound testing at BuildFlow workspace scale. Core-lane duty raises the bar on all four.
4. **Prior session's open debts carried**: f#4 AST-scanner gate (ban direct `encoding/json/v2` in Duration/wire paths), re-verification of "ln"-tainted evidence from the earlier `rg -r` artifacts, go-line flip-flop skip-gate effectiveness — all untouched this session.

## c) NOT STARTED

1. **BuildFlow wiring** (the entire point of the provider): blank import in `sdk_imports.go`, direct go.mod require, `*ProviderRegistered` regression test (gotchas #149/#156).
2. **jscpd demotion for the Go lane**: remove `goFilePattern` from `jscpdTriggerPatterns` and `go` from `jscpdFormats` so jscpd keeps only languages art-dupl lacks. templ was never a jscpd format — no templ decision needed.
3. **Ghost config removal** in BuildFlow: `dupl-check` constant + metadata row + build-mode case; `dupl_threshold`/`DuplSemantic`/`--semantic` plumbing end-to-end (flag → koanf → materialize → validate → telemetry); `tool_paths` doc example referencing `"dupl"`.
4. **Install-hint fixes** ×2 (`doctor_cmd.go:226`, `flake.nix:543`).
5. **art-dupl release tag (g2)** so BuildFlow's go.mod and `providerVersion()` resolve real values instead of `dev`.
6. **g3 (go-output standing decision)** — untouched; recommendation stands (go-output renders BuildFlow-side; art-dupl stays renderer-free per ADR-0024 wire pins).
7. **Provider .gitignore support** (parity with jscpd's `--gitignore`; the matcher lives in art-dupl's `cmd` layer, not the SDK).
8. **Threshold configurability** — toolsdk Spec has no options channel; core-lane users may want a knob (requires a go-finding upstream change).

## d) TOTALLY FUCKED UP

1. **The architecture inversion (user-confirmed)**: BuildFlow's duplication detection is exactly backwards from intent. jscpd — token-based, 80-token hardcoded bar for Go, no semantic mode, no templ — is the ONLY dup detector in the pipeline, while art-dupl (the core tool, purpose-built for Go+templ, Type 1/2/3 semantic) is a ghost system. The tool with the strongest detection sits out; the weakest carries the lane.
2. **My recommendation repeated the inversion**: I proposed "coexist as parallel detectors" — framing jscpd as art-dupl's peer. Wrong. Core/backup is the truth, and I should have asked about relative priority instead of assuming parity. Corrected by the user; recorded here so the lesson sticks.
3. **BuildFlow's art-dupl story is rotten end-to-end**: three name aliases (`dupl`/`dupl-check`/`art-dupl`), a fully-plumbed config surface (`dupl_threshold` 10–50, `--semantic`) that reaches no consumer, telemetry importing a value nothing uses, and an install hint for a nonexistent org — replicated in two places. This is the gotcha-#147 ghost-config disease in its purest form, predating this session; the comparison exposed it, nothing more.
4. **I used `rg -r` AGAIN** (third+ offense across sessions) — `rg -rn "dupl"` corrupted jscpd output ("nicate") and I briefly read fabricated tokens before quarantining it. The prior session's summary carries this as a CRITICAL lesson and I still relapsed once. This is a process-discipline failure, not a knowledge gap: the lesson exists and was ignored under task load.
5. **The provider remains a ghost system on the art-dupl side too**: one session after shipping, its only consumer is its own test file. Every day unwired is value deferred; the wiring is one line plus the compliance rail (require + regression test).

## e) WHAT WE SHOULD IMPROVE

1. **Lesson durability**: recorded lessons do not survive task load when they live only in conversation summaries. Cross-project process lessons (the `rg -r` ban) belong in the crush-config repo's `references/lessons.md` (per global AGENTS.md rules) — a file that loads with context, not a summary that gets skimmed.
2. **Verify-then-claim**: I wrote "MVS would pick v1.13.1 anyway" before checking the tag; it happened to be true (now verified), but the order was wrong. The `verify-external-claims` skill exists precisely for this and wasn't consulted for the BuildFlow-facing claims.
3. **Ask about priority, don't assume parity**: when comparing two tools in one domain, the relative-priority question (core vs backup vs peer) is a user-decision I should have surfaced explicitly instead of defaulting to "coexist".
4. **Sibling-repo hygiene check before recommending integration**: BuildFlow's go.mod needed `go mod tidy`; a 10-second preflight (per BuildFlow's own gotcha #181) would have flagged that before I tried `go run`.
5. **Provider self-documentation**: make the HealthCheck explicit in the Spec (BuildFlow's NoOp fallback covers it, but explicit is self-documenting for the next SDK consumer), and state the fixed-threshold contract in the Spec description so BuildFlow users aren't surprised by the absent knob.
6. **Ghost-system audits on both sides**: art-dupl's provider (unwired) and BuildFlow's dupl config (unconsumed) are the same disease from opposite ends — a periodic "who consumes this?" pass would have caught both.

## f) Up to 50 things we should get done next

> Brainstorm ranked by impact, not a commitment list; `docs-health` HARVEST applies extra routing rigor (many items are ROADMAP fuel; BuildFlow-side items need explicit authorization). P0 = make the core real; P1 = hardening + carried debts; P2 = BuildFlow-side; P3 = polish/upstream.

**P0 — promote art-dupl to the core lane**
1. Cut the art-dupl release tag (g2) so BuildFlow pins resolve real versions and `providerVersion()` stops reporting `dev`.
2. Authorize + execute BuildFlow wiring: blank import `_ "github.com/LarsArtmann/art-dupl/pkg/provider"` in `sdk_imports.go`.
3. Add the art-dupl require to BuildFlow go.mod (direct; MVS lands `toolsdk v1.13.1`), `go work vendor`, `nix run .#update-vendor-hash`.
4. Add the `*ProviderRegistered` regression test in `sdk_imports_test.go` (mandated by `TestAllSDKToolsHaveProviderRegisteredTest`).
5. Demote jscpd for Go: remove `goFilePattern` from `jscpdTriggerPatterns` and `go` from `jscpdFormats` (`jscpd_provider.go`) — jscpd becomes the non-Go backup lane.
6. Kill the ghost config end-to-end: `dupl-check` constant, metadata row, build-mode case, `dupl_threshold` key/flag/materialize/validate/telemetry, `DuplSemantic`/`--semantic`, `tool_paths` doc example.
7. Fix the install hints: `doctor_cmd.go:226` + `flake.nix:543` → `go install github.com/LarsArtmann/art-dupl@latest`.
8. One canonical name (`art-dupl`) everywhere; decide whether `dupl-check` needs a `deprecatedToolAliases` entry for any fleet config referencing it.
9. Update BuildFlow's `tool_metadata.go` row (`"dupl-check"`, LanguageAny) to the real tool: `art-dupl`, Go+templ scope, real description.
10. E2E: run BuildFlow against a fixture repo and verify art-dupl findings render in the summary's detect-only section (gotcha #86 path).

**P1 — core-lane hardening (art-dupl side)**
11. Provider .gitignore support (port/downshare `GitignoreMatcher` or adopt a matcher lib) — parity with jscpd's `--gitignore`.
12. Threshold policy decision: fixed semantic-5 acceptable for the core lane, or a toolsdk options channel (go-finding upstream change)?
13. Classification metadata: decide whether the SDK should compute clone-type/category/priority/actionability (CLI-parity findings) or stay classification-free with the provider documenting why.
14. Explicit HealthCheck in the Spec (self-documenting; check toolsdk for a no-op helper).
15. Integration test against a BuildFlow-shaped fixture: 33-module workspace, vendor/, generated files, templ files, dot-dirs.
16. Concurrency test: parallel `Detect` calls (BuildFlow result-cache + parallel-step scenarios) — the SDK must be race-clean under `DetectorFromFinding` wrapping.
17. Performance/timeout: art-dupl's suffix tree at BuildFlow-workspace scale vs BuildFlow's `default_step_timeout` — `ToolFromSpec` maps no Timeout from Spec, so the default applies; measure and decide whether a Spec-level timeout is needed upstream.
18. Verify art-dupl findings stay below the fail-on=error default gate (severity ladder) so a clone-heavy repo doesn't hard-fail pipelines unconfigured (gotcha #154/#156 interaction).
19. Verify the audit ratchets pass with the new spec: not in `moduleScopedGoToolSpecs` (no go/packages — correct), no fan-out audits fire.
20. Result-cache interplay: confirm gotcha #122's binaryHash invalidation covers in-process detector upgrades (art-dupl version bump → fresh results).
21. Self-scan via the SDK path on art-dupl itself (the provider is a new consumer of pkg/artdupl — dogfood it).
22. Windows invariant: provider crawl uses `filepath.WalkDir` — verify slash-normalization parity with the CLI's crawl-path invariant.
23. gogenfilter version parity: provider's `gogenfilter.FilterAll` vs cmd's marker table — pin and test the content-only detection behavior.
24. Add GroupID determinism test across provider runs (stable anchors depend on it).
25. Test `stripEmptyMetadata` drops ALL zero-valued classification keys (assert the full key set, not a sample).

**P1 — carried debts from the prior session**
26. f#4: AST-scanner test banning direct `encoding/json/v2` imports in Duration/wire paths (pattern: `TestNoDuplicateErrorNewMessages`).
27. Re-verify all "ln"-tainted evidence from the earlier `rg -r` artifacts with plain rg (still never done).
28. Verify the go-line flip-flop skip-gate (`.buildflow.yml` skip of `go-mod-normalize`) is actually effective in BuildFlow.
29. Monthly self-scan cadence check (`scripts/self-scan.sh` + ledger entry — due since 2026-09-22 decision).
30. TODO_LIST fleet-audit entry: migrate remaining ~30 direct `encoding/json/v2` files that carry Duration/wire concerns.
31. FEATURES.md/HOW_TO_USE.md: confirm the provider is documented as an SDK integration feature (FEATURES row done last session; HOW_TO_USE unchecked).
32. Annotate the prior status report (2026-09-24_13-30) with this session's corrections (docs-health ANNOTATE mode — the g1 answer supersedes its "coexist" framing).
33. HARVEST this report's (f) list into TODO_LIST/ROADMAP with routing rigor.
34. Record the `rg -r` recidivism as a cross-project lesson in crush-config `references/lessons.md` (by commit, not in-session write).

**P2 — BuildFlow-side follow-ups (need authorization)**
35. BuildFlow preflight before any wiring session: `go mod tidy` state, `nix run .#deps`, tree quiescence (their gotcha #181).
36. BuildFlow docs: their AGENTS.md gotcha #73 (manual `-t` sweeps) should gain the pipeline-tool story once wired.
37. Warnings-budget entry for art-dupl in BuildFlow's own `.buildflow.yml` (gotcha #184 pattern).
38. Decide dedup policy between lanes: with disjoint languages (Go+templ vs rest), no DependsOn or cross-suppression needed — write that down so nobody re-adds overlap later.
39. Update BuildFlow's `language/language_constants.go::ToolDupl` ("dupl") — delete or rename to match the canonical name.
40. Confirm the telemetry path: `dupl_threshold` removal touches `CLIUsage.ToMap`/`configMap` (gotcha #25 rule: only `cliFlagsMap`) — do it in the same change as item 6.
41. `buildflow docs --check` after tool changes (their doc-consistency gate).
42. Post-wiring: re-baseline gotcha #73's manual sweep numbers — the pipeline now surfaces the same signal continuously.
43. Consider jscpd `.jscpd.json` config-path branch (`jscpd_provider.go` respects a repo config) — document that Go files leaving jscpd's format list don't break existing repo configs.
44. g3 ratification: go-output renders BuildFlow-side; art-dupl stays renderer-free (standing recommendation, untouched).

**P3 — polish / upstream**
45. toolsdk upstream proposal: options/config channel for detector specs (threshold knobs without code changes).
46. toolsdk upstream: explicit HealthCheck helper + optional Timeout field consideration (BuildFlow maps neither today).
47. art-dupl README: document the toolsdk provider as an SDK integration (BuildFlow is consumer #1).
48. pkg/artdupl docs: the `ErrNoDuplicatesFound` → empty-findings contract for SDK consumers.
49. Cross-check exclusion parity: BuildFlow `constants.SkipDirNames`/FileIndex skips vs provider's crawl skips (dot-dirs, vendor, node_modules, examples/demo/demos) — document divergences deliberately or align.
50. After both lanes ship: one fleet-wide query comparing jscpd-vs-art-dupl finding overlap on non-Go languages to validate the backup lane is actually pulling weight (or demote jscpd further).

## g) Questions I cannot figure out myself

1. **Wiring authorization + fleet-config policy**: May I execute the BuildFlow side in a follow-up session (items 2–10, 35–43)? And within it: is deleting `dupl_threshold`/`--semantic` acceptable even though fleet `.buildflow.yml` files carrying those keys will start warning as unknown keys (gotcha #147's honest-noise tradeoff), or do you want a deprecation window?
2. **Threshold for the core lane**: is the provider's fixed semantic threshold (5 statements) acceptable for BuildFlow's primary dup detector, or do you want it configurable — which requires a toolsdk Spec extension upstream (go-finding change, not art-dupl-local)?
3. **Release timing + g3**: cut the art-dupl tag now (BuildFlow's pin and `providerVersion()` resolve real values immediately), or bundle with the .gitignore/classification hardening first? And do you ratify the standing go-output decision (renders BuildFlow-side; art-dupl stays renderer-free)?

---

*Point-in-time snapshot; per skill contract this report's (f) section is the HARVEST input for TODO_LIST/ROADMAP. Waiting for instructions.*
