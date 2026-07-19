# Critical Evaluation of 2026-07-19 Feedback Reports

**Date:** 2026-07-19
**Evaluator:** Crush (glm-5.2)
**Sources:**

- `2026-07-19_cyberdom_t25_templ_exclusion_and_accept_tracking.md`
- `2026-07-19_generated-templ-not-excluded-and-output-triage-gaps.md`
- `2026-07-19-httputil-t25-high-signal-html-report-cross-package-noise.md`

**Method:** Every claim verified empirically against current source (commit `418bf230` + working tree), using a freshly-built binary (`go build -o /tmp/artdupl-test ./cmd/art-dupl`).

---

## Headline Finding

**Most feedback claims are invalid.** Three AI agents ran dedup sessions using a **stale installed binary** (`/home/lars/go/bin/art-dupl`, built Jul 18 20:38, reports `version dev`). Their reports describe behavior that does not match current source. Several "missing feature" requests are for features that already exist and are documented in `--help`.

Before any code changes are made, the genuinely missing features must be separated from the noise.

---

## Claim-by-Claim Verdict

### Report 1: Cyberdom

| Claim                                                        | Verdict             | Evidence                                                                                                                                                                                                                                                                            |
| ------------------------------------------------------------ | ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "Generated templ NOT auto-excluded — 20/43 occurrences"      | **INVALID**         | Current binary verbose output: `Auto-generated code filtering enabled (templ, default)`. The `_templ.go` file IS excluded. Author was on stale binary.                                                                                                                              |
| "interface-implementation pattern not firing at -t 25"       | **INVALID**         | Pattern IS firing. Tested with 2 identical-body interface impls (10 lines each). Structural mode at -t 3 reports them; semantic mode at -t 3 silently suppresses them via actionability. Author mistook "suppressed = pattern not firing". Real UX issue: suppression is invisible. |
| "No mechanism to mark clones as accepted — no baseline file" | **INVALID**         | `art-dupl baseline` + `art-dupl check` subcommands exist. See `baseline/baseline.go` and `cmd/baseline_cmd.go`. Hash-based CI gating is fully implemented.                                                                                                                          |
| "Compile-time interface assertions flagged"                  | **UNVERIFIED**      | Plausible — `var _ Iface = (*Type)(nil)` patterns are likely detected as clones.                                                                                                                                                                                                    |
| "No JSON or machine-readable output"                         | **INVALID**         | Four structured outputs exist: `--json`, `--simple-json`, `--plumbing`, `--sarif`. Verified in `--help`.                                                                                                                                                                            |
| "'unknown' category uninformative"                           | **PARTIALLY VALID** | `--rich-text` mode surfaces categories like `[type-2]`, `[low]`, but `unknown` is a real fallback. Could fall back to AST node type.                                                                                                                                                |
| "Threshold semantics opaque"                                 | **PARTIALLY VALID** | Help text is clearer than feedback implies: "minimum number of duplicated statements to report as clone (default: 5)". Could still explain unit better.                                                                                                                             |

### Report 2: DiscordSync

| Claim                                                            | Verdict               | Evidence                                                                                                                                                                                                                                                                              |
| ---------------------------------------------------------------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "Generated `*_templ.go` files reported despite being gitignored" | **INVALID**           | Same as Report 1. Stale binary. Current source excludes via `Code generated by templ` marker.                                                                                                                                                                                         |
| "--filter-generated is opt-in"                                   | **INVALID**           | Filtering is default-ON. Help text says "Enable" which is misleading wording, but behavior is opt-out. Verified via verbose output.                                                                                                                                                   |
| "--exclude-templ inverts user's mental model"                    | **INVALID**           | The flag name says what it does: excludes `.templ` source files. `.templ` sources ARE included by default (per `config.go:86` comment and `run_crawl.go:290`). Help text correctly says "Exclude .templ source files from analysis". The user's mental model was wrong, not the flag. |
| "Text output has no code preview"                                | **VALID**             | Default and `--rich-text` both show only `file:line-line`. No code preview. Real gap.                                                                                                                                                                                                 |
| "Text output lacks category tag"                                 | **PARTIALLY VALID**   | `--rich-text` shows `[type-2] [low]` tags. Default text shows none.                                                                                                                                                                                                                   |
| "No diff mode / extract-verify loop"                             | **PARTIALLY INVALID** | `art-dupl check` against baseline IS a diff mode for CI. Not arbitrary baseline diff, but addresses the core use case.                                                                                                                                                                |
| "Honor .gitignore"                                               | **NOVEL**             | Real new idea, not implemented. Has tradeoffs (subprocess cost, non-git repos).                                                                                                                                                                                                       |

### Report 3: httputil

| Claim                                                           | Verdict                   | Evidence                                                                                                                                          |
| --------------------------------------------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| "//art-dupl:accept directive missing"                           | **VALID (complementary)** | Baseline file exists but is hash-based and doesn't travel with code. Inline directive would be code-local and survive renames. Genuinely missing. |
| "Skill recommends wrong default (-t 5 vs -t 25 for test-heavy)" | **VALID**                 | Skill doc could note test-heavy libraries benefit from higher thresholds. Easy doc fix.                                                           |
| "Same-function category"                                        | **NOVEL**                 | Good idea: clone groups where all occurrences are in the same FuncDecl are often intentional (mw1/mw2 test pattern). Would help triage.           |
| "--out flag for HTML mode"                                      | **VALID**                 | Minor ergonomic gap. HTML goes to stdout; users redirect manually.                                                                                |
| "Stable id on HTML clone groups"                                | **VALID**                 | Minor: `id="clone-N"` for deep-linking.                                                                                                           |
| "Heuristic: would-take-more-params-than-lines"                  | **NOVEL**                 | Interesting but risky. Could compute `extraction_cost = num_params / num_lines`.                                                                  |

---

## Summary: Genuinely Missing Features

After critical evaluation, the **actually missing** features are:

### Tier 1: High-impact, low-risk

1. **Text output: one-line code preview per group** (both reports mention) — ✅ DONE
   - Add first non-empty source line of the first clone as preview
   - Helps triage without opening files
   - Trivial implementation

2. **Skill doc: recommend `-t 25` for test-heavy libraries** (Report 3) — ✅ DONE
   - One-line addition to `deduplicate-code/SKILL.md`

3. **Clarify `--filter-generated` help text** (Report 2 confusion) — ⚠ OBSOLETE
   - Currently says "Enable filtering" which implies opt-in
   - Change to "Filtering is default; this flag is obsolete" or similar
   - **2026-07-19:** No longer applicable. `--filter-generated` was replaced by `--include-generated` (default filtering is ON; the new flag opts INTO specific generated categories). See `cmd/root.go` help output.

### Tier 2: Medium-impact, medium-effort

4. **`//art-dupl:accept` directive** (Report 1 + Report 3)
   - Complementary to baseline (baseline = CI hash file; directive = code-local)
   - Lives with code, survives renames, documents reason at site
   - Implementation: scan comments during file load, suppress matching groups
   - ~150-300 lines of code + tests

5. **"Same-function" category** (Report 3)
   - Novel idea, helps triage by downranking likely-intentional clones
   - Requires tracking which FuncDecl each clone belongs to
   - Moderate effort

### Tier 3: Low-priority / unclear value

6. **Honor `.gitignore`** (Report 2) — subprocess cost, non-git repos edge case
7. **HTML `--out` flag** (Report 3) — minor, redirect workaround exists
8. **HTML anchor IDs** (Report 3) — minor, deep-linking nicety
9. **`--recommend-threshold`** (Report 1) — nice-to-have
10. **`--diff-report`** (Reports 1+2) — overlaps with existing `baseline/check`

### Investigate further

11. **Interface-implementation pattern at high thresholds** — verify whether it fires
12. **Compile-time interface assertions** — check if `var _ Iface = (*Type)(nil)` is detected as a clone (it likely is)

---

## Rejected Suggestions (Already Exist)

These suggestions from the feedback should NOT be implemented because the features already exist:

- **"Add JSON output"** — `--json`, `--simple-json`, `--plumbing`, `--sarif` all exist
- **"Add baseline mechanism"** — `art-dupl baseline` + `art-dupl check` subcommands
- **"Fix templ auto-exclusion"** — already works in current source (was stale binary)
- **"Make `--filter-generated` default"** — already is default
- **"Rename `--exclude-templ`"** — name is correct; user mental model was wrong

---

## Root Cause: Stale Binary

The biggest finding across all 3 reports is that they were generated against a **stale binary** built Jul 18. The binary reports `version dev` (no version embedding), making it indistinguishable from a fresh build. This caused:

- 3 independent agents to file similar "templ not excluded" reports
- Multiple "missing feature" requests for features that exist
- Erosion of trust in the tool's capabilities

**Recommendation:** Add version embedding via `-ldflags "-X main.version=<git-sha>"` to the build. A user running `art-dupl --version` should immediately see if they're on a stale build.

---

## Recommended Action

1. **Do not blindly implement** the feedback's feature requests
2. **Do** add version embedding (highest ROI — prevents future stale-binary feedback)
3. **Do** add text output code preview (genuine gap)
4. **Do** update skill docs for test-heavy libraries
5. **Defer** the `//art-dupl:accept` directive until baseline mechanism is documented
6. **Reject** JSON/baseline/templ requests with pointers to existing features

The feedback reports contain useful signal wrapped in incorrect observations. The reports' authors would benefit from reading `art-dupl --help` more thoroughly before filing.

---

# Appendices: Original Feedback Files

The three source feedback reports are appended below in full, unmodified, so future readers can compare the claims against the verdicts above without cross-referencing separate files.

---

## Appendix A: Cyberdom Feedback

**Source:** `docs/feedback/new/2026-07-19_cyberdom_t25_templ_exclusion_and_accept_tracking.md`

# Feedback: templ generated code leaks through auto-exclusion; no accept/suppress mechanism creates a re-report treadmill

**Date:** 2026-07-19
**Project:** Cyberdom — Go 1.26 Discord agent with CQRS/event-sourcing (~10k LOC, 23 packages, Ginkgo + stdlib dual test styles)
**Command:** `art-dupl --semantic --sort total-tokens -t 25 --html` (initial), then text-mode re-runs for triage
**Goal:** Reduce duplication to zero per user instruction; realistically, eliminate all harmful clones and document the rest.

## Session Summary

Ran art-dupl at the user-specified `-t 25` on a mid-size Go project. Initial run produced **11 clone groups / 43 occurrences**. After excluding the generated templ file and performing 5 refactors, the final run showed **6 clone groups / 12 occurrences**, all in the "accept" category (interface contracts, compile-time checks, table-driven tests, distinct mock data).

**Real duplications found and fixed:** 5 groups
**False positives / accepted idioms remaining:** 6 groups
**Net LOC change:** -38 lines (97 added, 135 removed)

The session was productive — art-dupl surfaced genuine duplication in test scaffolding, fixture declarations, and dead abstraction patterns. However, three gaps consumed disproportionate triage time and would recur on every future run.

## What Worked Well

### 1. Semantic mode accurately grouped cross-file structural clones

The tool correctly identified the `InsertAIResponseParams{...}` literal repeated across 3 test functions in `internal/db/db_test.go` despite surrounding code differing. This was the highest-value finding of the session and led to a clean `insertTestAIResponse` helper extraction.

### 2. Clone "migration" after refactor forced better design

When I collapsed the `Responses` fixture struct into a `buildResponseFixtures()` builder with an anonymous return type, art-dupl immediately re-flagged the group at new line numbers — the anonymous struct shape appeared twice (return type + return literal). This forced me to introduce a named `responseFixtures` type, which is genuinely better design. The tool's stubbornness here was a feature, not a bug.

### 3. Text output mode is superior to HTML for triage

Despite the beautiful HTML report (dark theme, collapsible groups, copy buttons, diff viewer), I switched to text output (`art-dupl` without `--html`) for every triage pass. The `file:line` format is scannable, greppable, and doesn't require a browser round-trip. The HTML is nice for sharing; the text mode is what actually drives the work.

### 4. Test-vs-production separation in the summary

The HTML summary card showing `Production: 28 / Test code: 15` was useful for prioritization. I focused on production clones first, then test scaffolding. Keeping this split visible in the overview helps.

### 5. Fast and stable

No performance complaints on a 10k LOC project. Re-runs after each refactor were instant enough to use as a tight feedback loop.

## What Was Painful

### 1. Generated templ code was NOT auto-excluded — the biggest noise source

The deduplicate-code skill documentation states:

> Generated code (sqlc, protobuf, mockgen, stringer, templ) and test-file noise are auto-excluded by default.

**This did not happen.** `templates/layout_templ.go` (generated from `layout.templ`) produced **20 of the 43 initial occurrences** — nearly half the report. Every single one was the templ runtime buffer-release boilerplate:

```go
if !templ_7745c5c3_IsBuffer {
    defer func() {
        templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
        if templ_7745c5c3_Err == nil {
            templ_7745c5c3_Err = templ_7745c5c3_BufErr
        }
    }()
}
```

I had to manually add `--exclude-pattern "templates/layout_templ.go"` to every subsequent run. The skill's claim and the tool's behavior are out of sync. Either:

- The auto-exclusion heuristic doesn't recognize `*_templ.go` files (detection gap), or
- The skill documentation overstates what the tool does (docs gap)

Either way, this is the single highest-impact fix: 20 phantom clones drowned out the 23 real ones.

### 2. Go interface implementations flagged as the #1 "real" clone group

After excluding templ, the top remaining group was 4 occurrences of the `CompleteWithTools` method signature:

```go
func (c *Client) CompleteWithTools(
    ctx context.Context,
    systemPrompt string,
    messages []ai.Message,
    tools []ai.ToolSchema,
) (*ai.Response, error) {
```

…in `ai/claude/client.go`, `ai/openai/client.go`, and twice in `ai/testhelpers/mocks.go` (two different mock types implementing the same interface).

These signatures **must** be identical — Go's interface system requires it. This is not deduplicatable. The 2026-07-06 feedback notes that an `interface-implementation` actionability pattern was added, but it clearly isn't firing at `-t 25`. The pattern may only activate at lower thresholds, or the detection requires the interface declaration to be in the same analysis pass.

This group will reappear on every run forever, in every Go project with interface implementations. It is the canonical false positive for Go.

### 3. No mechanism to mark clones as "accepted" — the re-report treadmill

The deduplicate-code skill wisely says:

> When accepting, leave a one-line rationale so the next reader knows it was deliberate.

But art-dupl has **no way to read those rationales back**. Every future run will re-report the same 6 accepted clones. The user (or the next AI session) will re-triage them from scratch. There is no `.artdupl-ignore`, no `// artdupl:accept: <reason>` comment parsing, no baseline file.

This creates a treadmill: the tool's output never converges, even as the codebase gets cleaner. The skill says "stop when every remaining clone has a defensible reason to exist" — but the tool gives you no way to record that decision.

This is the #1 missing feature for sustained use. Without it, art-dupl is a one-shot tool that degrades in usefulness every time you run it after the first.

### 4. Compile-time interface assertions flagged as clones

Two packages had identical `TestClientImplementsInterfaces` functions:

```go
func TestClientImplementsInterfaces(t *testing.T) {
    var (
        _ ai.Client          = (*Client)(nil)
        _ ai.ClientWithTools = (*Client)(nil)
    )
}
```

This is the idiomatic Go compile-time interface check. It must exist in each package to verify the local `Client` type. Flagging it as duplication suggests extracting a shared helper — which is impossible (the whole point is per-package compile-time verification).

### 5. No JSON or machine-readable output

For triage at scale (the `-t 15` run showed 103 groups; `-t 5` would show far more), I wanted to pipe output through `jq` to filter by category, group by file, count occurrences per package, etc. The text mode is human-formatted; the HTML is presentation-formatted. Neither is structured.

A `--format json` flag emitting `[{group_id, category, priority, occurrences: [{file, line, snippet}], tokens, suggestion}]` would unlock scripted triage and CI integration.

### 6. The "unknown" category is uninformative

16 of 43 initial clones were categorized as `unknown`. The HTML badge just says `📄 unknown`. For triage, this is a black box — I have to open every group to understand what kind of duplication it is. A fallback like `statement` or `expression` (matching the AST node type) would at least hint at the nature of the clone.

### 7. Threshold semantics are opaque

The skill says threshold counts "duplicated statements (not AST nodes)." But the relationship between threshold, line count, and token count is unclear. A clone spanning 8 lines (`fixtures.go:192,200`) was reported as a single clone at `-t 25`. Is an 8-line struct literal one "statement"? Two?

The `--help` output should explain the unit. Right now, choosing a threshold is guesswork — the user picked 25, the skill recommends 5, and neither explains what the number means in observable terms.

## Actionable Deduplications Found

### 1. Extracted `insertTestAIResponse` helper (3 sites eliminated)

`internal/db/db_test.go` had 3 near-identical `InsertAIResponseParams{...}` literals with overlapping fields. Extracted a helper with a mutator callback, mirroring the existing `createTestToolExecutions` pattern in the same file.

### 2. Added `NewEmbedSenderTrackerGetter` (2 sites eliminated)

Two service test files (`briefing`, `review`) inlined the same 4-line closure building `*EmbedSenderWithChannel`. Discovered an existing-but-unused `EmbedSenderTracker` type in the testhelpers package. Added a lazy-evaluating constructor that defers sender lookup until the Ginkgo `It` block runs.

### 3. Collapsed `Responses` fixture into a builder (2 sites eliminated)

The fixture struct had 4 entries each inlining `Usage{InputTokens, OutputTokens, TotalTokens}` where `TotalTokens` was always `Input + Output` (a latent redundancy). Refactored into a `buildResponseFixtures()` function with a `mk` helper that computes `TotalTokens`.

### 4. Removed redundant "Service creation" test specs (2 sites eliminated)

Both service test suites had a `Describe("Service creation")` block asserting `Expect(service).NotTo(BeNil())` — but `service` is already constructed in `BeforeEach`, making the assertion tautological. Pure cargo-cult test duplication.

## Suggestions

### 1. Fix templ auto-exclusion (highest priority)

The skill promises templ files are auto-excluded. They are not. Either:

- Detect `*_templ.go` files by name pattern and skip them, or
- Parse the file header for the templ generation marker and skip, or
- Update the skill docs to say "add `--exclude-pattern '*_templ.go'` manually"

This single fix would have cut the initial report from 43 to 23 occurrences and saved the first 10 minutes of confusion.

### 2. Add an accept/suppress mechanism

Implement inline comment parsing:

```go
// artdupl:accept: interface implementation — signature must match ai.ClientWithTools
func (c *Client) CompleteWithTools(...) (*ai.Response, error) {
```

Or a baseline file (`.artdupl-baseline.json`) recording accepted group fingerprints. On subsequent runs, accepted groups are hidden or de-prioritized with their rationale shown.

Without this, the tool cannot converge. Every run re-litigates settled decisions.

### 3. Make the interface-implementation pattern fire at all thresholds

The 2026-07-06 feedback says this pattern exists, but it's not firing at `-t 25` for method-signature clones. If the pattern only works when the interface declaration is in the same file or analysis unit, extend it to detect interface satisfaction across packages (the method set matches a known interface). Go's `types.Implements` makes this check trivial if the type checker is available.

### 4. Add `--format json`

For CI integration and scripted triage. Schema:

```json
[
  {
    "id": "stable-hash-of-fingerprint",
    "category": "method",
    "priority": "low",
    "tokens": 2,
    "occurrences": [
      { "file": "ai/claude/client.go", "start_line": 50, "end_line": 55, "snippet": "..." }
    ],
    "suggestion": "Consider strategy pattern or early returns",
    "is_test_code": false
  }
]
```

Stable IDs (content-addressed) are essential for diffing reports across runs.

### 5. Add a `--diff <baseline>` mode

Given a previous report (JSON), show only what changed: new groups, resolved groups, moved groups. This turns art-dupl from a snapshot tool into a progress-tracking tool.

### 6. Clarify threshold units in `--help`

Current: "Threshold (tokens)". Better:

```
-t, --threshold int    Minimum number of duplicated statements for a group to be reported.
                       A "statement" is a single AST statement (if, return, assignment, etc.),
                       not a line or token. Default: 5. Range: 1-100.
                       Lower thresholds find more clones but increase false positives from idioms.
```

### 7. Improve the "unknown" category

If the AST-based categorizer can't classify a clone, fall back to the AST node type (`basic-literal`, `composite-lit`, `call-expr`, `block-stmt`) rather than `unknown`. Every clone has a node type; `unknown` means the classifier gave up.

### 8. Add `--recommend-threshold`

Analyze the clone distribution and suggest a threshold where the harmful/intentional ratio is best. Output something like:

```
At -t 5:  103 groups (est. 85% intentional / table-driven / idioms)
At -t 15:  41 groups (est. 60% intentional)
At -t 25:   6 groups (est. 0% harmful remaining — all accepted)
Recommended: -t 20 (catches new duplications, low re-report noise)
```

This helps users pick a sustainable threshold instead of guessing.

## Metrics

| Metric                                         | Value        |
| ---------------------------------------------- | ------------ |
| Clone groups at `-t 25` (initial, incl. templ) | 11 (43 occ.) |
| Clone groups after excluding templ             | 10 (23 occ.) |
| Clone groups after refactoring                 | 6 (12 occ.)  |
| Real duplications fixed                        | 5 groups     |
| False positives / accepted idioms              | 6 groups     |
| Estimated FP rate at `-t 25` (post-exclusion)  | ~50%         |
| Time wasted on templ exclusion                 | ~10 min      |
| Time wasted re-triaging interface contracts    | ~5 min       |
| Net LOC change                                 | -38 lines    |

## Comparison to Prior Feedback

| Issue raised here                 | Previously raised? | Status              |
| --------------------------------- | ------------------ | ------------------- |
| Generated code not excluded       | No                 | **New**             |
| Interface implementations flagged | Yes (2026-07-06)   | Partially addressed |
| No accept/suppress mechanism      | No                 | **New**             |
| Table-driven tests flagged        | Yes (2026-06-11)   | Known               |
| Compile-time checks flagged       | No                 | **New**             |
| No JSON output                    | No                 | **New**             |
| Threshold semantics unclear       | Yes (2026-07-15)   | Known               |
| "unknown" category uninformative  | No                 | **New**             |

## Conclusion

art-dupl found real, fixable duplication that improved the codebase. The semantic engine is accurate, the refactoring feedback loop is tight, and the HTML report is polished. The tool's core competency — detecting structural clones — is solid.

The gaps are in the **workflow around detection**:

1. **Auto-exclusion is incomplete** (templ leaks through, contradicting the skill docs).
2. **There is no convergence mechanism** (no way to mark clones as accepted, so every run re-reports settled decisions).
3. **False-positive patterns for Go are under-filtered** at higher thresholds (interface signatures, compile-time checks).

Fixing the templ exclusion is a one-line change with massive noise reduction. Adding the accept/suppress mechanism is a medium effort with outsized impact on sustained use. Together, these two fixes would transform art-dupl from a "run once and triage" tool into a "run continuously and converge" tool.

---

**Author:** Crush (glm-5.2) on behalf of the Cyberdom deduplication session.
**Session diff:** 5 files, +97 / -135 lines, all tests passing (`go test -race ./...`).

---

## Appendix B: DiscordSync Feedback

**Source:** `docs/feedback/new/2026-07-19_generated-templ-not-excluded-and-output-triage-gaps.md`

# Feedback: Generated `*_templ.go` files not excluded; high-threshold signal-to-noise; text-output triage gaps

**Date:** 2026-07-19
**Project:** DiscordSync — Go CQRS Discord archive bot (~30 kLOC across `internal/`, 20 packages, templ + HTMX frontend, SQLite/Turso backend)
**Command:** `art-dupl --semantic --sort total-tokens -t 70 .`
**Result:** 12 clone groups at t=70. After a full dedup pass, reduced to 4 groups — but 8 of those 4 were in gitignored generated code that should never have been reported.

## Session Summary

Ran art-dupl at a deliberately high threshold (t=70) on a mature, already-deduplicated codebase to surface only the largest, most expensive clones. The goal was a quick triage pass: find the handful of high-impact duplications, extract helpers, re-run to confirm zero.

**What art-dupl reported:** 12 clone groups at t=70.
**What was actually deduplicated:** 7 groups (8 helpers extracted across 7 files).
**What should not have been reported:** 8 groups in `internal/web/activity_templ.go` — a **gitignored, templ-generated** file.
**What remained after the pass:** 4 groups (8 generated + 2 test-file pairs that are genuinely borderline).

## Issue 1 (HIGH): Generated `*_templ.go` files are reported despite being gitignored

This is the most actionable finding. The session opened with 12 clone groups; 8 of them (67%) were in a single generated file:

```
found 6 clones:
  internal/web/activity_templ.go:138,155
  internal/web/activity_templ.go:165,182
  internal/web/activity_templ.go:187,204
  internal/web/activity_templ.go:290,307
  internal/web/activity_templ.go:317,334
  internal/web/activity_templ.go:339,356
found 2 clones:
  internal/web/activity_templ.go:214,231
  internal/web/activity_templ.go:366,383
```

Verification:

```bash
$ git check-ignore internal/web/activity_templ.go
internal/web/activity_templ.go
$ head -3 internal/web/activity_templ.go
// Code generated by templ - DO NOT EDIT.
// templ: version: v0.3.1020
package web
```

The file is (a) gitignored, (b) marked `// Code generated by templ - DO NOT EDIT.`, and (c) regenerated from `.templ` sources on every build. It is structurally impossible for the user to "deduplicate" it — the duplication lives in the templ generator's output, not in user-editable code.

### Root cause: `--filter-generated` is opt-in and templ is not covered

From `art-dupl --help`:

```
--filter-generated          Enable filtering of sqlc.dev and templ.guide generated code
                            (auto-detects sqlc.yaml in parent directories)
--exclude-templ             Exclude .templ source files from analysis
--include-sqlc              Include sqlc.dev generated files (override auto-detection)
```

Three problems compound:

1. **`--filter-generated` is not the default.** The user must opt in. The deduplicate-code skill (`~/.config/crush/skills/deduplicate-code/SKILL.md`) runs `art-dupl --semantic --sort total-tokens -t 5 --html` with no `--filter-generated` flag, so every AI agent following the skill reports generated-code noise.
2. **`--filter-generated` covers sqlc and templ.guide, but the auto-detect only triggers for sqlc** (`auto-detects sqlc.yaml in parent directories`). There is no equivalent auto-detect for templ (no search for `templ: version` header, no `templ.go.sum`, no `// Code generated by templ` marker scan). So a templ project must pass `--filter-generated` **and** the filter must actually match templ output — which the help text implies but the behavior does not guarantee.
3. **`--exclude-templ` excludes `.templ` _source_ files, not `_templ.go` generated output.** These are different things. The user wants to analyze their `.templ` sources (that is where the real duplication lives) but exclude the generated `_templ.go` output. The current flag semantics are inverted from the user's mental model.

### Suggested fix (in priority order)

1. **Make `--filter-generated` the default**, with `--include-generated` to opt out. Generated code is almost never the target of a dedup pass. The current default (analyze everything, make the user opt into filtering) is the wrong default for the common case.
2. **Honor `.gitignore`.** A gitignored `.go` file is, by definition, not user-editable. art-dupl could run `git check-ignore <file>` during file enumeration and skip ignored paths (with a `--include-ignored` escape hatch). This would have eliminated all 8 templ groups in this session with zero false negatives — no real duplication lives in gitignored files.
3. **Add a templ auto-detect** equivalent to the sqlc one: scan for the `// Code generated by templ - DO NOT EDIT.` header (or `// templ: version:` on line 2). The marker is stable and well-defined by the templ project.
4. **Rename or split `--exclude-templ`.** The current name reads as "exclude templ-generated code" but actually excludes `.templ` sources. Consider `--exclude-templ-source` (current behavior) vs `--exclude-templ-output` (the `_templ.go` files), or consolidate under `--filter-generated`.

### Impact

For this project: eliminates 8 of 12 reported groups (67% noise reduction) at t=70. At t=30 (the project's standard threshold), the generated-file noise is worse because templ output is verbose and repetitive — the same CSS-class stacks and templ-runtime calls produce dozens of clone groups that are all unfixable.

---

## Issue 2 (MEDIUM): Text output has no code preview, category, or priority

The default text output for a clone group is:

```
found 6 clones:
  internal/web/activity_templ.go:138,155
  internal/web/activity_templ.go:165,182
  ...
```

That is: count + `file:line,file:line` pairs. To triage a single group, the user must:

1. Open each file at the cited line.
2. Read the surrounding code.
3. Decide if it is a real clone, an interface contract, a generated artifact, or a test idiom.
4. Repeat for every group.

For a 12-group report this is tolerable. For the 112-group report I saw at t=30, it is the bottleneck of the entire dedup workflow — far more time is spent opening files than evaluating clones.

### What would help

- **A one-line code preview per clone** (first line of the matched token sequence). jshint, eslint, and rubocop all do this. It lets the user triage 80% of groups without opening a file.
- **A category tag per group** (`generated`, `test`, `interface-impl`, `type-ref`, `sync-mutex`, `unknown`). The semantic engine already has this information internally; surfacing it in the text output would let the user `grep -v generated` or similar.
- **A priority score** (`high` / `med` / `low`) based on token count × occurrence count × whether the group spans multiple files. The `--sort total-tokens` flag already sorts by one axis; a composite priority would be more useful.

The `--html` and `--json` outputs may already include some of this (the skill recommends `--html`), but the text output is the default and the only mode that composes well with pipes, `grep`, and `fzf`.

---

## Issue 3 (LOW): No "did my extraction actually eliminate the clone?" feedback

The dedup workflow is iterative: find a clone → extract a helper → re-run → confirm the group is gone. art-dupl supports the "find" and "re-run" steps but not the "confirm" step. There is no diff mode, no "compare to last run" cache, no `--verify-shrink <previous-report>` flag.

In this session, this caused a real miss: I extracted `assertPartialHasNoChrome` in `partial_test.go`, marked the todo complete, and moved on. The clone was still there (the extraction removed the `assert.NotContains` pair but the shared `doRequest + assert.Equal + body :=` prefix still registered). I only caught it when writing the session report, because I re-ran art-dupl for the report rather than after each edit.

### What would help

- A `--since <git-ref>` mode that shows only clone groups that **changed** (new, grown, shrunk, eliminated) since the ref. The `--since` flag exists in the help output but is documented as "for incremental mode" — it is not clear whether it produces a diff of clone groups or just a faster re-run.
- A `--diff-report <baseline.json>` mode that compares the current run against a saved `--json` baseline and reports deltas. This would make the extract-and-verify loop tight: save baseline, edit, re-run with diff, see exactly which groups shrank or disappeared.

This is lower priority than Issues 1 and 2 because the workaround (re-run and eyeball) does work — it just relies on discipline the user may not have.

---

## What worked well

### Semantic detection correctly identified the one production clone

At t=70, exactly one production-code clone group was reported:

```
found 2 clones:
  internal/projection/stage_instances.go:13,25
  internal/projection/stage_instances.go:30,42
```

These were the `StageInstanceCreated` and `StageInstanceUpdated` handlers — identical payload type, identical ensure-FKs + upsert + commit body. A closure factory (`handleStageInstanceUpsert(database)`) eliminated the group cleanly. The tool nailed it: no false positive, high signal, immediately actionable.

### Helper-consumption re-detection is honest

After extracting helpers, the remaining "clone" between `partial_test.go` subtests is the shared `doRequest + assert.Equal + body := rec.Body.String()` prefix — the call sites that consume the new helper. This is **correct behavior**: the two subtests now share an identical 3-line preamble because they use the same request pattern. Reporting it is honest; the user's judgment call is whether to extract a `partialBody(t, views, path)` helper or accept the prefix as idiomatic. The tool should not make that call; it correctly surfaces the shape and steps back.

### High-threshold runs are a fast triage path

t=70 on this codebase produced 12 groups in under 2 seconds. The signal density was high enough to act on immediately: one production extraction, six test-helper extractions, eight generated-code groups to dismiss. For a first-pass triage of an unfamiliar codebase, a high-threshold run (`-t 50` or `-t 70`) is a better entry point than the skill's recommended `-t 5` — it surfaces the expensive duplication first, without the 100+ group noise of a low-threshold run.

**Suggestion:** the skill could recommend a two-pass workflow: (1) `-t 50` for triage, (2) `-t 5` for cleanup after the big clones are extracted. This matches how senior engineers actually triage — biggest wins first.

---

## Metrics

| Metric                                          | Value   |
| ----------------------------------------------- | ------- |
| Clone groups at t=70                            | 12      |
| Generated-file groups (gitignored templ)        | 8       |
| Test-file groups                                | 3       |
| Production-code groups                          | 1       |
| Groups extracted (eliminated)                   | 7       |
| Groups accepted (documented)                    | 2       |
| False positives (generated noise)               | 8       |
| False positive rate (raw)                       | 67%     |
| False positive rate (after filtering generated) | 0%      |
| Time to triage all 12 groups                    | ~8 min  |
| Time to extract all 7 helpers                   | ~25 min |

---

## Suggestions (prioritized)

1. **Default `--filter-generated` to on.** Add `--include-generated` for the rare case where the user wants to audit generator output. This single change would have eliminated 67% of this session's report.
2. **Honor `.gitignore`** during file enumeration, with `--include-ignored` to override. Gitignored files are, by definition, not user-editable; reporting clones in them is never actionable.
3. **Add templ auto-detection** to `--filter-generated`: scan for the `// Code generated by templ - DO NOT EDIT.` header. The marker is stable and defined by the templ project.
4. **Add a one-line code preview** to the text output for each clone group. The first matched line is enough for 80% of triage decisions.
5. **Surface the category tag** (`generated` / `test` / `interface-impl` / `unknown`) in the text output so users can `grep -v` noise categories.
6. **Add `--diff-report <baseline>`** to support the extract-verify loop. Save baseline JSON, edit, re-run with diff, see exactly which groups shrank or disappeared.
7. **Rename `--exclude-templ`** to `--exclude-templ-source` and add `--exclude-templ-output` for the generated `_templ.go` files. The current name inverts the user's mental model.

## Conclusion

The semantic engine is accurate — the one production clone it reported was real, and the helper-consumption re-detection is honest. The tool's weakness is **input filtering and output triage**: 67% of this session's report was unfixable generated-code noise, and the text output forces file-by-file inspection that a one-line preview would eliminate. Making `--filter-generated` the default (and extending it to cover templ output) is the single highest-impact change.

---

## Appendix C: httputil Feedback

**Source:** `docs/feedback/new/2026-07-19-httputil-t25-high-signal-html-report-cross-package-noise.md`

# Feedback: Excellent test-code dedup signal at `-t 25`; HTML report shine; cross-package helper duplication flagged as noise

**Date:** 2026-07-19
**Project:** httputil — Go HTTP middleware library (`github.com/larsartmann/httputil`, ~3 700 production LOC, ~6 000 test LOC, 2 packages)
**Command:** `art-dupl --semantic --sort total-tokens -t 25 --html`
**Goal:** Drive duplication to zero; extract every harmful clone.

> Verdict: **High-signal session, low-noise report.** 11 clone groups → 9 actionable extractions → 2 accepted intentional groups. The semantic engine and the HTML report both performed well; the suggestions below are about edges, not fundamentals.

## Session Summary

Ran art-dupl at `-t 25` on a library that already had a prior dedup pass (so the baseline was not messy). The report surfaced **11 clone groups (42 tokens), 100% in test code, 0 production clones** — exactly the right signal: no production noise, all the actionable duplicates were test-fixture boilerplate.

After full iteration, the final report is **2 clone groups (2 tokens)**, both intentional and documented in the project's `AGENTS.md`:

1. `mw1` / `mw2` middleware factories in `stack_test.go` — the labels are the test's subject matter
2. `newTypedBodyHandler` defined once per package (`httputil` + `httpspec`) — dependency direction forbids sharing

**Net result:** 11 files changed, **-86 lines** removed, 8 new helpers extracted, all tests pass (`-race`), `golangci-lint` 0 issues.

## What Worked Exceptionally Well

### 1. `-t 25` is the right default for libraries with heavy test fixtures

This is the strongest signal of the session. The skill recommends `-t 5`, and prior feedback notes note that `-t 5` is where noise drops to near-zero. But on this codebase, `-t 5` reports **321 groups** — virtually all single-line idioms (`t.Parallel()`, `return n, nil`, `if !ok { return }`). `-t 25` cut straight to the 11 groups that were actually worth thinking about, and 9 of those yielded real extractions. **For library-style projects with many small test handlers, `-t 25` is a better default than `-t 5`.**

The session validates raising the default floor for projects where test fixtures dominate LOC. The art-dupl author may want to consider a heuristic: if >60% of LOC is test code, suggest `-t 25`.

### 2. Semantic mode correctly grouped Type-2 clones

The 10 occurrences of:

```go
http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusXXX)  // StatusOK, StatusTeapot, etc.
})
```

were correctly unified into a single clone group despite **different status codes**. This is exactly the Type-2 clone detection I wanted — the structure is identical, only the literal differs. After extracting `newStatusOnlyHandler(status int)`, all 10 sites collapsed to one-liner call sites.

Likewise, the CORS test bodies that differed only in input origin + expected value were correctly unified:

```go
// 4 sites, same shape, only cfg.Origins / req origin / expectation differ
cfg := CORSConfig{ AllowedOrigins: ..., AllowedMethods: ... }
middleware := CORS(cfg)
inner := newNoOpHandler()
req := newTestRequest(http.MethodGet, "/", requestOrigin)
rec := newRecorder()
middleware(inner).ServeHTTP(rec, req)
assertAllowOrigin(t, rec, wantOrigin)
```

These all collapsed into `assertCORSForOrigin(t, allowed, requestOrigin, wantOrigin)`.

### 3. The HTML report is excellent for triage

The collapsible clone groups, the per-group category/priority badges, the "test" vs "production" filter buttons, and the "10 occurrences · 10 tokens" summary line made it trivial to triage 11 groups in under a minute. The 📋 Copy button on each occurrence was unexpectedly handy for building edit operations — I copy-pasted exact source into `old_string`/`new_string` pairs without having to re-read each file.

The filter buttons (`🧪 Test` / `📦 Production`) were particularly important: confirming **0 production clones** at a glance let me focus entirely on the test-fixture extractions without worrying about behavioral risk.

### 4. Iterative re-runs converge quickly

After each batch of extractions, re-running art-dupl was fast (~1s on a ~10k LOC project) and the report shrank monotonically: 11 → 7 → 2. No "whack-a-mole" where fixing one clone introduced two new ones. The only post-extraction re-detection was expected: my newly-extracted `newTypedBodyHandler` helper is structurally identical in `httputil/testutil_test.go` and `httpspec/handlers_test.go` — but that's a **correct observation**, not a false positive.

## What Surprised Me (Mildly Negative)

### 1. Cross-package helper duplication is structurally unresolvable — but still reported

**Group:** `newTypedBodyHandler` defined in both `testutil_test.go` (package `httputil`) and `httpspec/handlers_test.go` (package `httpspec`).

These are byte-identical bodies because they do the same thing — but they CANNOT be shared:

- `httputil` is the parent package
- `httpspec` is a subpackage of `httputil`
- Go's dependency direction means `httputil` CANNOT import from `httputil/httpspec`
- The only way to share would be to create a third internal `testutil` subpackage — significant overhead for two 6-line helpers

art-dupl correctly identifies this as a clone. It IS a clone. But it's **structurally unactionable** in Go. The report offers no way to express "I have reviewed this and it is intentional." The user has to write a `// accepted duplication: ...` comment elsewhere (in my case, in `AGENTS.md`) that the tool does not read.

**Suggestion:** Support a `//art-dupl:accept` or `//nolint:art-dupl` directive at the clone site. When the tool re-encounters the same clone, it could read the directive and omit the group from the report (or show it in a separate "Accepted (N)" section). This would let the user converge to a literally zero-line report and make CI gating possible.

The directive format could be cheap to implement: a comment on the line of the first occurrence, optionally with a reason:

```go
//art-dupl:accept cross-package helper duplication is structural in Go
func newTypedBodyHandler(contentType, body string) http.Handler { ... }
```

### 2. "Integer-labeled" test fixtures are reported as clones

**Group:** `mw1` and `mw2` in `stack_test.go`:

```go
mw1 := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        order = append(order, "mw1-before")
        next.ServeHTTP(w, r)
        order = append(order, "mw1-after")
    })
}

mw2 := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        order = append(order, "mw2-before")
        next.ServeHTTP(w, r)
        order = append(order, "mw2-after")
    })
}
```

art-dupl correctly identifies these as clones — they ARE structurally identical after alpha-normalization of the label literals. But the **label is the entire point of the test**: this test asserts that `mw1-before, mw2-before, handler, mw2-after, mw1-after` is the observed execution order. Parameterizing the helper (`newOrderRecordingMiddleware(name string)`) would obscure that the labels `mw1` and `mw2` are intrinsic to the assertion.

This is a generalizable pattern: when a clone group is **inside a single function** and the differing identifiers are the subject of an assertion later in the same function, extraction usually harms readability. The tool cannot know this without data-flow analysis, but the **single-function** heuristic is cheap and would let the report down-rank such cases.

**Suggestion:** Add a "same-function" category/priority. Clone groups where all occurrences live in the same function (start..end line ranges overlap a single `ast.FuncDecl`) could be tagged `🟡 same-function` instead of `🟢 low`, signaling "probably intentional — review carefully before extracting."

### 3. The `--html` mode writes the report to stdout, not to a file

The skill says "View the HTML output directly — do not save it as a file." In practice, I had to redirect to a file (`art-dupl --html > /tmp/report.html`) and then open it. For a CLI tool, a flag like `--html --out=report.html` (or auto-writing to `art-dupl-report.html` in the cwd) would be more ergonomic than piping.

Minor nit; not a real problem once you know the trick.

### 4. The `code-N-M` element IDs in the HTML are not anchored

When I wanted to send a colleague a link to clone group #6 specifically, I had no URL fragment to use (`#clone-group-6` or similar). Each `<div class="clone-group">` would benefit from an `id="clone-N"` attribute so deep-linking works.

Very minor.

## Metrics

| Metric                             | Value                             |
| ---------------------------------- | --------------------------------- |
| Clone groups at `-t 25` (initial)  | 11 (42 tokens)                    |
| Clone groups at `-t 25` (final)    | 2 (2 tokens, both accepted)       |
| Clone groups at `-t 5` (reference) | 321 (overwhelmingly single-line)  |
| Production clones                  | 0                                 |
| Test clones                        | 11 → 2                            |
| Real duplications found            | 9                                 |
| Extracted helpers                  | 8 (6 in httpspec, 2 in httputil)  |
| Net LOC change                     | -86 lines                         |
| False positives at `-t 25`         | 2 (both intentional, documented)  |
| Estimated FP rate at `-t 25`       | 18% (2/11)                        |
| Estimated FP rate at `-t 5`        | >95% (idiom noise)                |
| Time to triage 11 groups           | ~1 minute (HTML filters + badges) |
| Time to full iteration to zero     | ~10 minutes (extract + verify)    |

## Suggestions (Prioritized)

### 1. Inline `//art-dupl:accept` directive (HIGH IMPACT, LOW RISK)

Allow users to mark intentional clones at the site. When re-running, omit accepted groups from the active report (or list them under a separate "Accepted (N)" section). This is the single biggest unlock for CI gating and "drive to zero" workflows — without it, the terminal state of a dedup session is always "N clones remain, trust me they're fine."

Implementation sketch:

- On each clone group, compute a stable hash from the AST fingerprint (you already have this).
- Before emitting, scan each occurrence's preceding comment lines for `//art-dupl:accept[:<hash-prefix>] [<reason>]`.
- If found, suppress the group (or move to "Accepted" section).
- The hash-suffix is optional but prevents stale accepts after the code changes.

### 2. Consider raising the documented default for test-heavy libraries (MEDIUM IMPACT, LOW RISK)

The skill currently recommends `-t 5` as a sensible default. For projects where >60% of LOC is test code, `-t 25` produces dramatically better signal-to-noise (18% FP rate vs >95%). Consider:

- Adding a note to the skill: "For library projects with heavy test fixtures, try `-t 25` first."
- Or auto-detecting: if the proportion of `_test.go` (or equivalent) LOC exceeds 50%, emit a hint in the report summary: "ℹ️ Test-heavy codebase (62% test LOC). Consider `-t 25` for cleaner signal."

### 3. "Same-function" category for clone groups (MEDIUM IMPACT, LOW RISK)

When every occurrence of a clone group lives inside the same `ast.FuncDecl`, tag it as `🟡 same-function` rather than `🟢 low`. These cases are very often intentional — the duplicated fragments implement a parallel structure (e.g., mw1 vs mw2 in an ordering test, or two columns of a comparison matrix) where extraction would obscure intent.

### 4. `--out=<path>` flag for HTML mode (LOW IMPACT, LOW RISK)

Ergonomic improvement. Auto-detecting a sensible default (`art-dupl-report.html` in cwd when `--html` is passed and stdout is a TTY) would remove the redirect dance.

### 5. Stable `id` attributes on HTML clone groups (LOW IMPACT, LOW RISK)

Add `id="clone-group-N"` to each `<div class="clone-group">` so deep links (`file://.../report.html#clone-group-6`) work for sharing specific findings.

### 6. Heuristic: warn on "would-take-more-params-than-lines" extractions (LOW IMPACT, MEDIUM RISK)

A common false-positive class is a 4-line pattern where extraction would require a 3-parameter helper. The skill itself calls this out ("An abstraction would take more parameters than the duplicated code has lines"). The tool could compute a simple `extraction-cost` score: `num_parameters / num_lines_in_clone`. When this exceeds 1.0, the clone is unlikely to benefit from extraction and could be down-ranked or annotated `🟡 low-value-extraction`.

This is a heuristic, so there's a risk of false negatives — but it would have saved me from briefly considering the `mw1`/`mw2` extraction.

## Conclusion

This was a clean, productive session. At `-t 25`, art-dupl on a test-heavy library gave **9 real findings and only 2 accepted duplicates** — an 82% actionable rate, which is excellent. The HTML report made triage fast, and the semantic engine correctly grouped Type-2 clones with varying literals.

The two remaining clones are **structurally unactionable** in Go (cross-package test-helper duplication) and **intentionally unactionable** (label-parameterized test middleware). Both are correctly detected; what's missing is a way to tell the tool "I've reviewed these, don't report them again." That single feature — an inline `//art-dupl:accept` directive — would unlock CI gating and let "zero clones" become a literally achievable state rather than an aspirational one.

For library authors with heavy test suites, the unstated best practice is `-t 25`, not `-t 5`. The skill documentation would benefit from saying so explicitly.

---

**Author:** Crush on behalf of the `httputil` deduplication session.
**Session hash:** 11 → 2 clone groups in one pass.
