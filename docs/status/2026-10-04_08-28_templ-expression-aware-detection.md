# Status Report: Expression-Aware Templ Detection Overhaul

**Date:** 2026-10-04 08:28 CEST
**Session task:** "I feel like our templ integration is meh!" — READ/UNDERSTAND/RESEARCH/REFLECT, break into steps, execute + verify one at a time.
**Working tree at report time:** clean (auto-commit daemon has committed all session work as `chore: auto-commit` commits on `fork`).
**Verification state at report time:** build green; full suite green as of the classify-split; targeted suites re-run green after the last change (printer/actionability, syntax/*); corpus re-checked. Full-repo lint was 0 issues before the classify split; printer package lint not re-run after it (listed in next tasks).

---

## a) Fully done and verified

### Diagnosis (all empirically confirmed before any code)

1. `--exact` mode dropped ALL names for templ files (`mode == Semantic` bool squeeze in `job/file_parser.go`) — `<div>` ≡ `<section>`.
2. All embedded Go expressions were invisible: `if user.IsAdmin` ≡ `if user.IsCompletelyUnrelatedCondition` (false Type-1 clones), `{ item.Name }` ≡ `{ item.DeletedAt }`, component-call args invisible, `{{ goCode }}` invisible, top-level Go code skipped entirely.
3. `if X { A } else { B }` ≡ `if X { A; B }` (else structure flattened; else-if links flattened into siblings).
4. Switch: `default:` ≡ `case "…":` (ComponentSwitchDefaultCase existed, never used); `fallthrough` typed as a case body.
5. Conditional attributes (`<p if x { class="a" } else { class="b" }>`): condition AND both branches completely invisible.
6. CSS: property names invisible (`color: red` ≡ `background: blue`); CSS declaration not a unit → duplicate clone groups printed.
7. Component/CSS declarations hashed the whole signature string instead of the declared name.
8. Regex normalizer mis-normalized lowercase package names (`display.Card` → `v0.Card`, aliased with `widgets.Card` across files).
9. templ clone categories in `--explain` decoded against the parallel golang enum → `unknown`/wrong categories.

### Implemented (7 planned steps, all executed)

| Step | What                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Key files                                                                                            | Verification                                                                                                                                               |
| ---- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | `golang.DetectionMode` threaded through templ parser; names encoded iff `HashesIdentifiers()` (exact=verbatim, semantic=normalized, structural=off); arch-lint `syntax-templ` → `syntax-golang`                                                                                                                                                                                                                                                             | `syntax/templ/parser.go`, `job/file_parser.go`, `.go-arch-lint.yml`                                  | `mode_plumbing_test.go` (4 tests); CLI: exact mode 0 groups on div-vs-section                                                                              |
| 2    | `golang.ParseSnippet` bridge: 7 snippet kinds (Expression, IfCond, ForClause, SwitchTag, CaseList, Statements, GoFile); synthetic-file wrapping with offset padding; `normalizer.collectSnippetLocals` (free lowercase idents declared in source order; callees/selector fields/comp-lit keys/predeclared ids verbatim)                                                                                                                                     | `syntax/golang/snippet.go`, `syntax/golang/normalizer.go`                                            | `snippet_test.go` (15 tests: normalization, exact/structural, positions, statement marks, import filtering, error fallback)                                |
| 3    | Transformer upgrades: conditions parsed to child tokens; else containers + statement-marked else-if links; `ComponentElseStatement=100` / `ComponentFallthroughStatement=101` (above golang 0-53 range so `syntax.IsStatementContainer` can't collide); default-case typing; cond-attr traversal; CSS property names + statement composites; decl-name hashing; callee verbatim + parsed args; GoCode statement splicing; top-level Go declarations spliced | `syntax/templ/transform*.go`, `syntax/templ/snippet.go`, `syntax/templ/types.go`, `syntax/syntax.go` | `behavior_test.go` (~20 token-stream tests); CLI checks: false-positive guards gone, Type-2 positives intact, CSS dup output gone, else structure distinct |
| 4    | Retired regex normalizer (`normalize.go` + tests deleted; `symbols` field removed)                                                                                                                                                                                                                                                                                                                                                                          | `syntax/templ/`                                                                                      | intent coverage ported to AST-level tests                                                                                                                  |
| 5    | `CacheVersion` 4 → 5 (templ serialized stream changed; v4 entries would reproduce expression-blind results)                                                                                                                                                                                                                                                                                                                                                 | `cache/file_cache.go`                                                                                | cache + job tests green                                                                                                                                    |
| 6    | Docs: AGENTS.md templ bullet rewritten (expression-aware contract), CacheVersion note, jsonutil recurrence #4 recorded; FEATURES.md templ row updated                                                                                                                                                                                                                                                                                                       | `AGENTS.md`, `FEATURES.md`                                                                           | docs-health count gate passed (29 templ cases unchanged)                                                                                                   |
| 7    | Verification: full suite `-count=1` green; full-repo golangci-lint **0 issues**; arch-lint OK; `scripts/check-boundary.sh` fast profile **OK (43s)**; `-race` green on syntax/*, job, cmd; self-scan `-t 5` **0 shown**; corpus: templ-components 108 groups (10 true whole-file `_sources` pairs), go-sse 17 shown (baseline 16), go-cqrs-lite 347 shown (baseline 411)                                                                                    | —                                                                                                    | see Next Tasks for the re-checks after the last change                                                                                                     |

### Bonus fix (foreign-edit incident, not part of the plan)

- **jsonutil v1-only invariant broken a 4th time** by auto-commit `758646e5` (2026-10-03, predating this session): `internal/jsonutil/jsonutil.go` + `config/config_migrate.go` re-migrated to direct `encoding/json/v2` → Duration refusal + wrong empty-indent output; `jsonv2gate` named both files. Restored both from `758646e5^`; jsonv2gate, jsonutil, config, provider-golden all green again. Recurrence #4 documented in AGENTS.md.

### Polish (discovered during verification, completed)

- **templ clone categories + suggestions** (`printer/actionability/clone_classify.go`): the compiler proved the golang/templ enums collide numerically in a merged switch (duplicate cases), so `nodeTypeToCategory` is split into `goNodeTypeToCategory` / `templNodeTypeToCategory`, disambiguated by `.templ` filename. Added `templSuggestion` ("Extract to a shared templ component" etc.). arch-lint: `actionability` → `syntax-templ` added. Verified end-to-end: templ if-group now `conditional` + templ-specific fix line; component-call group `call`.

---

## b) Partially done / in flight

1. **Post-classify-split full re-verification.** The category split was the LAST code change; after it only `printer/actionability` tests + build + `--explain` CLI were re-run. Full-repo lint and the full test suite have NOT been re-run since. Expectation: green (change is additive), but unverified.
2. **Unit tests for `templNodeTypeToCategory` / `templSuggestion`** — none exist yet; behavior is only covered end-to-end via `--explain` scratch checks.
3. **BDD-level templ false-positive guards** — planned in the original breakdown (`bdd/templ_clone_detection_test.go` additions), NOT done. The behavior is covered at the syntax/token-stream level (`syntax/templ/behavior_test.go`) instead, so the BDD layer adds redundancy rather than coverage.
4. **`nix flake check`** (the full CI mirror: race lane, alloc-gate, vendor-hash) — not run this session. Boundary fast profile + targeted `-race` were run instead. Alloc budgets should be unaffected (no suffixtree/Go-parser hot-path changes budgeted), but the templ snippet path adds go/parser calls per expression with no benchmark watching it.

## c) Not started (deliberate scope cuts, candidates for later)

1. **ADR for the expression-aware templ detection** — the change is architecturally significant (new golang→templ dependency direction, new snippet bridge, enum extension policy "values above 100"); AGENTS.md has the convention text but no ADR-0027.
2. **Templ parse performance benchmark** — `ParseSnippet` invokes `go/parser` per embedded expression; no baseline exists, so regressions would be invisible (the alloc-gate only guards suffixtree/syntax paths).
3. **`ScriptTemplate` parameter expressions** — name is encoded; `Parameters` expression not parsed (scripts' JS bodies intentionally unparsed).
4. **`SpreadAttributes` real-spread syntax verification** — my probe file accidentally used `attrs={...}` (a named attribute); true `{ expr... }` spread gets expression children by code path but was never exercised.
5. **Snippet-GoFile early-offset clamping test** — top-level Go code in the first ~10 bytes of a file clamps positions to 0 (documented in `shiftTree`); untested edge.
6. **Templ-tuned priority thresholds** — markup clones inherit Go function/loop priority curves; possibly fine, never evaluated.
7. **Category/priority for mixed Go+templ groups** — a group spanning a `.go` and a `.templ` file classifies by the FIRST instance's filename; rare, undefined, untested.

## d) Totally fucked up (and fixed, with lessons)

1. **`go build -o /tmp/art-dupl-test ./cmd`** produced an **ar archive** (multi-package dir) instead of an executable; every subsequent "art-dupl behavior" observation was actually mvdan/sh failing to run the archive. Cost ~6 tool calls of confusion. Lesson: the main package is `./cmd/art-dupl`; verify `file $(which binary)` when output looks impossible.
2. **behavior_test.go harness parsed bare component bodies** without wrapping them in a package/component shell — every stream compared a garbage 2-token tree; half the assertions "passed" for the wrong reason. Caught by inspecting the truncated streams (`A=[25 206088211]`). Lesson: assert on stream CONTENT (length floors) in helpers, not just equality.
3. **Merged golang+templ category switch** — compiler caught duplicate-case collisions (templ Element(4) == golang BasicLit(4), etc.). Would have shipped partially-wrong categories for templ AND made the switch order load-bearing for Go files. Lesson: the parallel-enum overlap is not theoretical; any value-keyed switch must be file-kind-scoped.
4. **Snippet wrapper bugs** (missing func-closing brace; GoFile wrongly function-wrapped; CaseClause looked up in the func body instead of inside the SwitchStmt; empty-result fallback turning import-only GoCode opaque) — all caught by the snippet tests + one debug probe.
5. **Environment:** `/mnt/buildcache` corrupted mid-session (stale -d entries, even for stdlib); worked around with a session-local `GOCACHE=/tmp/go-build-cache-artdupl`. Not ours, but any later "weird link error" should check there first.
6. **Tooling friction (minor):** two edit attempts rejected on read-before-edit staleness; one `var _ = strings.TrimSpace` hack briefly committed to transform.go before removal; golines/wsl findings needed two passes because gofmt ≠ golines.

## e) What we should improve

1. **The auto-commit daemon remains the top recurring incident source** (jsonutil recurrence #4 this session, 4 total). The jsonv2gate catches this specific class, but the general pattern — daemon commits in-flight/foreign edits with heuristic messages — has no general gate. A "tests green at commit time" hook or a daemon exclusion list for `internal/jsonutil`-style invariant files would kill the class.
2. **Corpus baselines are stale after this change** — go-sse 16→17, go-cqrs-lite 411→347. The deltas mix real behavior change (templ expressions) with corpus drift (repos are actively edited). A fresh documented baseline + a note in AGENTS.md would prevent future sessions from misreading the drift.
3. **Templ detection now has zero performance visibility** — add a templ parse benchmark + budget before the next optimization pass, not after.
4. **Behavior tests live at two layers** (syntax/templ token streams + golang snippet tests) with the BDD layer thin on templ; decide which layer owns "user-visible templ behavior" and say so in TESTING.md.
5. **`file`-type sanity for CLI smoke tests** — half the session's false alarms came from running a non-executable; the boundary script could `file`-check its built binary in dev workflows.

## f) Next tasks (prioritized, concrete)

1. Re-run full-repo `golangci-lint run ./...` (last run predates the classify split).
2. Re-run `go test ./... -count=1` full suite (same reason).
3. Add unit tests: `templNodeTypeToCategory` (each templ type → expected category), `templSuggestion` (each category branch + fallback), and a Go-side regression pin (`goNodeTypeToCategory(golang.BasicLit)` stays `expression`).
4. Run `nix flake check` (race lane + alloc-gate + vendor-hash) to close the CI-mirror gap.
5. Add `BenchmarkTemplParse` (templ file with N expressions, ParseBytesWithMode in semantic mode) + commit baseline under `docs/benchmarks/`.
6. Write ADR-0027: expression-aware templ detection (motivation: the 7 diagnosed bug classes; the ParseSnippet bridge contract; enum-value policy ≥100; arch-lint dependency change; fallback semantics).
7. Add BDD guard: two templ files differing only in an if-condition must produce 0 groups at `-t 1` (the user-visible version of behavior_test.go's guard).
8. Update `docs/benchmarks/` notes + re-document corpus numbers (go-sse 17, go-cqrs-lite 347) with a dated re-baseline entry.
9. Verify true spread-attribute syntax `{ attrs... }` produces expression children (probe + test).
10. Test SnippetGoFile position clamping edge (top-level Go in the first 10 bytes).
11. Decide + document the owning layer for templ behavior tests (TESTING.md).
12. Check `TODO_LIST.md` for stale templ entries that this session resolved (e.g. anything referencing "templ has no semantic mode") and close them.
13. Escalate or accept the jsonv2 recurrence pattern (tie to question 3 below).
14. Consider `--dump-tokens` showing decoded names for templ statement composites (currently bare `11`/`12`).
15. Templ-tuned priority evaluation: sample 20 templ clone groups from templ-components, check priority ordering feels right.
16. Mixed Go+templ group classification: define + test the rule (currently first-instance filename wins, undocumented).
17. `ScriptTemplate.Parameters` expression parsing (small, symmetric with components).
18. Sweep FEATURES.md "Limitations" section for other stale templ claims (line 285 area mentions "Go and Templ Only" — verify wording still accurate).
19. Self-scan ledger: not due until 2026-10-22, but the new templ code itself should be judged then (`scripts/self-scan.sh`) — put on calendar.
20. Evaluate caching `ParseSnippet` results per (content-hash, kind) if benchmarks show templ parse cost regressed (only with data).
21. Website/docs: the project website's templ feature page (if exists) should mention expression-aware detection as a differentiator.
22. Upstream a-h/templ: `ConstantCSSProperty` still has no Range (Pos inherits parent) — check newer templ versions for a fix; bump if available.

## g) Questions I cannot answer myself

1. **Corpus re-baseline policy:** should I re-baseline go-sse/go-cqrs-lite/templ-components numbers NOW (documenting the templ-change delta explicitly), or wait for the next scheduled docs-health pass (2026-10-28)? The drifted numbers will confuse the next status report either way.
2. **`_sources/` vendoring noise:** templ-components shows 10 whole-file duplicate pairs (`cmd/tc/_sources/X.templ` vs `X.templ`) at the DEFAULT threshold. Should art-dupl gain a default exclusion (or a `--include-vendored-sources` style flag) for such vendored copy directories, or is surfacing them correct behavior?
3. **jsonv2 recurrence escalation:** after the 4th identical incident, is the catch-and-restore loop (jsonv2gate + manual restore from parent) acceptable as the permanent control, or do you want me to take it to BuildFlow upstream (the daemon packaging in-flight edits is the root cause) — e.g. a daemon pre-commit test hook?

---

## Reflection answers

### What did I forget?

- To re-run the FULL suite + full lint after the very last code change (the classify split) — only targeted suites were re-run. Items 1–2 in Next Tasks.
- The BDD layer entirely — my original plan promised BDD additions and I silently substituted syntax-level tests without noting the substitution until writing this report.
- The templ parse performance dimension: I added per-expression `go/parser` cost with no benchmark watching it. The project's own history (alloc-gate rationale) says exactly this gets forgotten.
- No ADR. AGENTS.md got the facts, but this repo's convention for architectural decisions is `docs/adr/`, and a future session reading only ADRs will miss WHY the templ enum now jumps to 100.
- `TODO_LIST.md` was never consulted — resolved templ items there (if any) remain open.

### What could I have done better?

- **Verify the binary before blaming the code.** The ar-archive blunder cost real time and produced misleading evidence ("art-dupl parses itself"); a `file` check or a smoke `version` call would have ended it in one step.
- **Write the test harness correctly the first time** — the behavior tests initially compared garbage streams because bodies weren't wrapped; a `t.Fatal` floor on stream length inside the helper would have caught it immediately instead of via confusing diffs.
- **Probe upstream types BEFORE writing the first draft**, not after: the ScriptTemplate field-name compile error and the CaseClause-inside-SwitchStmt bug both came from coding against remembered type shapes. The probe program I eventually wrote should have been step zero.
- **Batch lint hygiene into the writing pass** — wsl/golines/varnamelen findings needed two extra passes; matching the repo's blank-line idiom while writing would have saved a cycle.
- **Treat the merged category switch with suspicion from the start** — the parallel-enum overlap is DOCUMENTED in AGENTS.md ("values must stay unused by syntax/templ's parallel raw enum"); I should have anticipated the collision instead of letting the compiler find it.

### What could I still improve (forward-looking)?

- **Kill the jsonv2 incident class structurally** (question 3) — four recurrences means the control is detection-only, not prevention.
- **Performance telemetry for templ**: benchmark + optional snippet-parse caching, decided by data.
- **Corpus freshness**: a small script that re-runs the three corpus checks and diffs against the documented numbers would make drift visible in minutes instead of rediscovered per session.
- **Layer ownership for detection-behavior tests** so the next language integration (if any) starts with the right test pyramid instead of reinventing it.
- **Upstream templ tracking**: newer a-h/templ versions may fix ConstantCSSProperty ranges; a quarterly dependency check would keep the known-limitations list honest.
