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

### 5. ~~No JSON or machine-readable output~~ RESOLVED

> **Update (2026-07-19, same day):** `--json` flag already exists and works (verified in the `go-auto-upgrade` feedback session: `art-dupl --json` was the triage workhorse). The claim below was based on a stale `version dev` binary. The `--simple-json` variant provides a simpler schema with `score=impact`.

~~For triage at scale (the `-t 15` run showed 103 groups; `-t 5` would show far more), I wanted to pipe output through `jq` to filter by category, group by file, count occurrences per package, etc. The text mode is human-formatted; the HTML is presentation-formatted. Neither is structured.~~

~~A `--format json` flag emitting `[{group_id, category, priority, occurrences: [{file, line, snippet}], tokens, suggestion}]` would unlock scripted triage and CI integration.~~

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
      {
        "file": "ai/claude/client.go",
        "start_line": 50,
        "end_line": 55,
        "snippet": "..."
      }
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
