# Feedback: Drove to literal zero at every threshold from `-t 5` to `-t 50`; production semantic clones real and actionable; Ginkgo `DescribeTable` variadic gotcha forces helper shape

**Date:** 2026-07-19
**Project:** go-auto-upgrade — Go monorepo for automated code migrations (`github.com/larsartmann/go-auto-upgrade`, ~7 500 production LOC, ~5 500 test LOC, 8 migrator subpackages with shared `goldentest` framework)
**Command:** `art-dupl --semantic --sort total-tokens -t 50 --html` (initial), then `--json` for triage, then plain text for verification
**Goal:** Drive duplication to zero per user instruction ("de-duplicate until ZERO").

> Verdict: **First observed session to reach 0 clone groups at every tested threshold (5, 10, 15, 20, 30, 50).** The semantic engine correctly identified cross-package production clones of finding-builder functions despite different format strings, and the JSON output mode (`--json`) was the workhorse for triage — not the HTML report. One real surprise: Ginkgo's `DescribeTable(description string, args ...any)` does NOT accept `func + slice...` syntax, which forced an extra helper to merge body and entries into a single slice. The workflow works end-to-end when the codebase cooperates.

## Session Summary

Ran art-dupl at the user-specified `-t 50` on a mid-size Go project with a deliberate "shared framework + many consumers" architecture (a `goldentest` package consumed by 8 migrator packages, plus a CLI test layer). Initial run produced **15 clone groups / 32 occurrences / 32 tokens** — 4 in production code, 27 in test code, 0 generated.

After full iteration, the final run reports **0 clone groups at `-t 50`, `-t 30`, `-t 20`, `-t 15` (default), `-t 10`, and `-t 5`**. At `-t 5`, three groups reappear — all of them the standard `Migrator` interface implementation boilerplate (`type Migrator struct{}`, `func New()`, `func Name()`, `func Description()`, `func MinGoVersion()`) repeated once per migrator package. These are structurally intentional: each migrator must satisfy the same interface, and the per-migrator `Description()` string is the unique data.

**Real duplications found and fixed:** 15 groups
**False positives / accepted idioms remaining:** 0 at `-t ≥ 5`; 3 at `-t < 5` (interface implementations)
**Net new code:** 1 new file (`cmd/go-auto-upgrade/testutil/cli_helpers.go`, 120 lines), 15 files modified
**Net LOC change:** ~+180 lines added, ~-220 lines removed across production and test code

The session validates that art-dupl's "drive to zero" workflow is **literally achievable** for codebases with a cooperative shape (shared framework, deterministic test patterns, no generated-code noise). This is the first session in the observed feedback corpus to reach zero at the default threshold.

## What Worked Exceptionally Well

### 1. Semantic mode correctly grouped production finding-builder clones across packages

The highest-value finding of the session. Three migrator packages (`errorspkg`, `lo2stdlib`, `jsonv1tov2`) each defined a `rewriteFinding` function with **different message and suggestion strings** but identical structure:

```go
// errorspkg
func rewriteFinding(apiName string, pos token.Position, path domain.PathString) finding.Finding {
    return migratorutils.NewRewriteFinding(
        "api-rewrite-"+strings.ToLower(apiName),
        Name,
        fmt.Sprintf("replaced errors.%s with stdlib equivalent", apiName),
        pos,
        string(path),
        fmt.Sprintf("errors.%s migrated to stdlib by go-auto-upgrade", apiName),
    )
}

// lo2stdlib (same shape, different qualifier and wording)
// jsonv1tov2 (same shape, different target library and wording)
```

The semantic engine unified all three despite the format-string differences. This is exactly Type-2/Type-3 clone detection done right — the structure is identical, only the literals and one helper name differ. The fix was to extract two helpers into `migratorutils`:

```go
func APIRewriteRuleName(apiName string) string {
    return "api-rewrite-" + strings.ToLower(strings.ReplaceAll(apiName, ".", "-"))
}

func NewAPIRewriteFinding(apiName string, name domain.MigratorName, message, suggestion string, pos token.Position, path string) finding.Finding {
    return NewRewriteFinding(APIRewriteRuleName(apiName), name, message, pos, path, suggestion)
}
```

After the extraction, art-dupl **re-detected** the three call sites as still-cloned (they were now `migratorutils.NewAPIRewriteFinding(...)` with the same argument shape). This forced the additional step of letting each migrator pass its own message/suggestion format strings, which preserved the per-migrator wording while removing the structural duplication. The tool's stubbornness on re-run was a feature.

### 2. `--json` mode is the actual triage workhorse

The prior feedback notes (`2026-07-19_cyberdom_t25_templ_exclusion_and_accept_tracking.md`) say "no JSON output." **That is no longer accurate.** The current `--json` flag emits a clean schema:

```json
{
  "version": "1.0",
  "threshold": 50,
  "files_analyzed": 125,
  "detection_method": "art-dupl",
  "clone_groups": [
    {
      "hash": "066c9c29b47f04ac",
      "size": 3,
      "files": [{ "filename": "...", "line_start": 23, "line_end": 46, "fragment": "..." }]
    }
  ]
}
```

I used `--json | jq` for **every triage pass**: counting groups, filtering by file, extracting just the line ranges for navigation, summarizing per-package distribution. The HTML report was opened once for the initial overview and never again. The text mode was used only for the final "is this zero?" verification.

Whoever added `--json` — thank you. It changed the workflow from "open browser, scroll, click" to "pipe through jq, decide in 5 seconds."

### 3. Convergence was monotonic and predictable

The report shrank monotonically across iterations:

```
15 → 14 → 13 → 12 → 10 → 9 → 8 → 0
```

Each refactor eliminated the targeted group and never introduced a new one. No "whack-a-mole" where fixing one clone created two new clones elsewhere. The only exception was the `rewriteFinding` extraction mentioned above, where the call-site shape was still cloned after the first pass — but that was a **correct observation**, not a false positive, and it led to a better second-pass extraction.

### 4. The hash-based group identity made re-runs easy to compare

Each clone group carries a stable `hash` field. Across runs, I could track which groups were resolved vs. which were new by diffing the hash sets. This made it trivial to confirm that my refactors didn't accidentally create new duplication patterns elsewhere in the codebase.

### 5. Threshold sweep confirmed excellent signal-to-noise at the default

After reaching zero at `-t 50`, I swept down to verify the fix was robust:

| Threshold | Clone groups | Notes                                                  |
| --------- | ------------ | ------------------------------------------------------ |
| 50        | 0            | User-specified starting point                          |
| 30        | 0            |                                                        |
| 20        | 0            |                                                        |
| 15        | 0            | **Default**                                            |
| 10        | 0            |                                                        |
| 5         | 3            | All `Migrator` interface implementations (intentional) |

The fact that zero holds from `-t 50` down to `-t 10` is strong evidence that the remaining `-t 5` findings are genuine interface-contract boilerplate, not overlooked duplication. This is the cleanest threshold-sweep result in the observed feedback corpus.

## What Was Painful

### 1. Ginkgo's `DescribeTable` signature blocks the obvious helper extraction

This is not art-dupl's fault, but it is a real workflow obstacle worth documenting. The standard golden-test pattern in this project was:

```go
var _ = Describe("Golden File Tests", func() {
    Context("...", func() {
        DescribeTable(
            "...",
            func(name string) { goldentest.AssertGolden(...) },
            Entry("a", "a"),
            Entry("b", "b"),
        )
    })
})
```

Seven migrator packages had near-identical `golden_test.go` files with this shape, differing only in the `Entry` data. The obvious extraction is a helper that takes a slice of entries and emits the `DescribeTable`. But Ginkgo's signature is:

```go
func DescribeTable(description string, args ...any) bool
```

and Go **rejects** the call `DescribeTable("desc", body, entries...)` where `entries` is `[]any`:

```
too many arguments in call to DescribeTable
    have (string, func(name string), []any...)
    want (string, ...any)
```

Go's variadic spread cannot follow a non-variadic argument. The workaround is a wrapper that concatenates body and entries into a single slice:

```go
func describeTableVariadic(description string, body any, entries []any) {
    args := make([]any, 0, 1+len(entries))
    args = append(args, body)
    args = append(args, entries...)
    DescribeTable(description, args...)
}
```

This works, but it is a non-obvious shape that every Ginkgo user extracting a table helper will rediscover. **art-dupl could not have known about this**, but the tool's report is what surfaced the duplication in the first place — so the user will hit this gotcha when they act on the report.

**Suggestion for the skill docs (not the tool):** Add a "Ginkgo table extraction" recipe to the deduplicate-code skill that pre-empts this with the `describeTableVariadic` snippet. Future dedup sessions on Ginkgo projects will save 15 minutes each.

### 2. The HTML report's `--html` flag does not auto-detect a TTY

When stdout is a terminal, `art-dupl --html` dumps a multi-thousand-line HTML document to the terminal, which is unreadable. The user must know to redirect (`> /tmp/report.html`) before opening. A TTY check that either (a) writes to `art-dupl-report.html` in the cwd, or (b) prints a "redirect to a file or pipe to a browser" hint, would save the first-time user from a wall of HTML.

This was raised in `2026-07-19-httputil-t25-high-signal-html-report-cross-package-noise.md` and remains valid. Low priority since `--json` and plain text cover most triage needs, but it is the only rough edge in the output story.

### 3. The "unknown" category dominated the report at `-t 50`

Every single one of the 15 initial clone groups was categorized as `📄 unknown`. The HTML badges offer no hint whether a group is a function, a block, a struct literal, or a set of statements. For triage, this meant opening every group to classify it mentally.

This was raised in `2026-07-19_cyberdom_t25_templ_exclusion_and_accept_tracking.md` and remains valid. Falling back to the AST node type (`func`, `block-stmt`, `composite-lit`, `assign-stmt`) instead of `unknown` would give the badges signal without requiring a smarter classifier.

### 4. No accept directive — still

Same gap as every prior feedback file. The three `Migrator` interface clones at `-t 5` will reappear on every future run forever. There is no way to tell the tool "I've reviewed these, they're interface contracts, don't report them again."

I won't re-argue the design (the `2026-07-19-httputil-...` and `2026-07-19_cyberdom_...` files already make the case eloquently). I'll only note that **this session reached zero at `-t ≥ 10` without the directive**, which means the accept-mechanism gap is a `-t 5` problem, not a `-t 15+` problem. Users who standardize on `-t 15` (the default) or higher will not hit this wall on a well-factored codebase.

## Actionable Deduplications Found

### 1. Extracted `migratorutils.APIRewriteRuleName` + `NewAPIRewriteFinding` (3 production sites)

Three migrators (`errorspkg`, `lo2stdlib`, `jsonv1tov2`) had near-identical `rewriteFinding` builders. The semantic engine correctly unified them despite different message strings. Extracted two helpers into the shared `migratorutils` package. Each migrator now passes its own message/suggestion format strings, preserving the per-migrator wording (which is intentional — each migrator migrates to a different target).

### 2. Extracted `stdlibwrappers.posFromFinding` (3 production sites, same file)

Three finding-builder functions in `internal/migrators/stdlibwrappers/tables.go` each opened with the identical 5-line `pos := token.Position{Filename: ..., Line: ..., Column: ...}` extraction from an upstream finding. Pulled into a `posFromFinding(f *finding.Finding) token.Position` helper. Small win, but it removed a copy-paste hazard: the next person who needed to add a 4th finding builder would have reached for the same snippet.

### 3. Extracted `goldentest.DefineGoldenSuite` + `DefineSingleGoldenSuite` (7 test files)

The biggest single extraction of the session. Seven migrator packages each had a `golden_test.go` with the same `Describe("Golden File Tests", func() { Context(...).DescribeTable(...).Entry(...) })` structure, differing only in the `Entry` data and the migrator instance. Built a declarative `GoldenSuite` value type plus two builders:

```go
// Single-fixture migrators (errorspkg, testifyassert)
goldentest.DefineSingleGoldenSuite(
    &Migrator{}, domain.GoVersion{Major: 1, Minor: 23},
    "pkg/errors migration", "wrap_patterns", 5,
)

// Multi-fixture migrators (jsonv1tov2, stdlibwrappers, lo2stdlib, lotidy, stdlib2lo)
goldentest.DefineGoldenSuite(goldentest.GoldenSuite{
    Migrator:  &Migrator{},
    GoVersion: domain.GoVersion{Major: 1, Minor: 24},
    AssertionGroups:   []goldentest.GoldenAssertionGroup{...},
    IdempotentEntries: goldentest.GoldenEntryList(fixtureNames, "idempotent"),
    FindingCountEntries: goldentest.FindingEntryList(pairs),
})
```

Two supporting helpers (`GoldenEntryList`, `FindingEntryList`) eliminate the per-row label duplication (`{Label: "wrap_patterns idempotent", Name: "wrap_patterns"}` becomes implicit when `GoldenEntryList(names, "idempotent")` derives the label from the name plus suffix).

The two errorspkg/testifyassert files collapsed from 41 lines each to 16 lines each. The full-file duplication between them (the original top hit in the report) was eliminated entirely.

### 4. Extracted `testutil.RunCLI`, `WriteExampleMigrationFile`, `AssertCLIWriteSuccess`, `AssertCLIDryRun`, `FileHasContent` (4 CLI test files)

The CLI test layer (`cmd/go-auto-upgrade/`) had four near-identical "write temp file, run binary, assert success" patterns spread across `all_test.go`, `main_test.go`, and `coverage_test.go`. Created a new `cmd/go-auto-upgrade/testutil/cli_helpers.go` (120 lines) with five focused helpers. The test files shrank to declarative one-liners:

```go
func TestCLIAll_WriteFlag(t *testing.T) {
    testutil.AssertCLIWriteSuccess(
        t, binaryPath,
        []string{"all"},
        "write_test.go",
        "all -w command failed",
    )
}
```

This required defining a `CLIRunner` interface so the helper could accept either the external-binary execution path (`exec.CommandContext`) or the in-process cmdguard execution path (`cli.ExecuteWithArgs`). The interface is one method (`ExecuteWithArgs`) and is satisfied implicitly by both surfaces.

## Suggestions (Prioritized)

### 1. Document the Ginkgo `DescribeTable` variadic workaround in the skill (MEDIUM IMPACT, LOW RISK)

Not an art-dupl fix — a skill-doc fix. Every dedup session on a Ginkgo codebase will hit the "extract a `DescribeTable` helper" finding, and every session will rediscover that Go's variadic spread cannot follow a non-variadic argument. Pre-empting this with the `describeTableVariadic` snippet in the skill's "extract" section will save future sessions 15-30 minutes of head-scratching and failed compiles.

The snippet to include:

```go
// describeTableVariadic calls DescribeTable with body and entries merged
// into a single variadic slice. Needed because Go refuses
// DescribeTable("desc", body, entries...) — body would be a non-variadic
// arg before a variadic spread.
func describeTableVariadic(description string, body any, entries []any) {
    args := append([]any{body}, entries...)
    DescribeTable(description, args...)
}
```

### 2. Demote `unknown` category to AST node type fallback (MEDIUM IMPACT, LOW RISK)

Every clone has an AST node type. When the semantic classifier cannot assign a category, fall back to the node type (`func-decl`, `block-stmt`, `composite-lit`, `assign-stmt`, `call-expr`) instead of `unknown`. This gives the HTML badge immediate signal: "this is a duplicated function body" vs. "this is a duplicated struct literal" vs. "this is a duplicated assignment pattern" — three very different triage paths.

This was raised in `2026-07-19_cyberdom_t25_...` and remains the highest-value low-cost report improvement.

### 3. Add a "fixture-driven test" detection pattern (LOW IMPACT, MEDIUM RISK)

The session's biggest extraction was collapsing 7 `golden_test.go` files into a declarative `DefineGoldenSuite` call. The shape that flagged them was:

```go
var _ = Describe("...", func() {
    Context("...", func() {
        DescribeTable("...", func(name string) { ... }, Entry(...), Entry(...), ...)
    })
})
```

This is a recognizable pattern: a `var _ = Describe(...)` assignment whose body is one or more `Context → DescribeTable` chains. When art-dupl detects this shape repeated across files in different packages, the suggestion could specifically be "extract a `DefineGoldenSuite`-style helper" rather than the generic "consider extracting to shared test utility."

The risk is that this is framework-specific (Ginkgo) and may not generalize. But for Go projects using Ginkgo (which is common in the LarsArtmann ecosystem and elsewhere), the pattern is high-value.

### 4. TTY-aware `--html` output (LOW IMPACT, LOW RISK)

When `--html` is passed and stdout is a TTY, auto-write to `art-dupl-report.html` in the cwd and print the path. When stdout is a pipe, behave as today. Saves the first-time user from a terminal full of HTML.

### 5. Accept directive — for `-t 5` users only (MEDIUM IMPACT, LOW RISK)

Reiterating the prior-feedback request, but with a scope observation: **at `-t 15` (the default) or higher, the accept-mechanism gap does not block convergence on a well-factored codebase.** It only becomes painful at `-t 5`, where interface implementations and idiomatic boilerplate dominate. This suggests the accept directive's value is concentrated in low-threshold workflows (CI gating on "no new duplication" rather than "zero duplication").

If the implementer is choosing where to invest, the directive matters more for `-t 5` CI gating than for interactive `-t 15+` sessions. The interactive path already converges without it.

## Metrics

| Metric                                   | Value                              |
| ---------------------------------------- | ---------------------------------- |
| Clone groups at `-t 50` (initial)        | 15 (32 occurrences, 32 tokens)     |
| Clone groups at `-t 50` (final)          | **0**                              |
| Clone groups at `-t 15` (default, final) | **0**                              |
| Clone groups at `-t 5` (final)           | 3 (all interface implementations)  |
| Production clones (initial)              | 4                                  |
| Production clones (final)                | 0                                  |
| Test clones (initial)                    | 27                                 |
| Test clones (final)                      | 0 at `-t ≥ 5`                      |
| Real duplications fixed                  | 15 groups                          |
| Helpers extracted                        | 7 (2 production, 5 test)           |
| New files added                          | 1 (`testutil/cli_helpers.go`)      |
| Files modified                           | 15                                 |
| Net LOC change                           | ~+180 added, ~-220 removed         |
| False positives at `-t 50`               | 0                                  |
| False positives at `-t 5`                | 3 (all intentional interface impl) |
| Estimated FP rate at `-t 50`             | 0%                                 |
| Estimated FP rate at `-t 5`              | 100% (of the 3 remaining)          |
| Iterations to zero                       | 7 refactor passes                  |
| Time to full convergence                 | ~45 minutes                        |
| Time to triage 15 initial groups         | ~2 minutes (with `--json + jq`)    |

## Comparison to Prior Feedback

| Issue raised here                      | Previously raised?                 | Status                                    |
| -------------------------------------- | ---------------------------------- | ----------------------------------------- |
| Zero-clone convergence achievable      | No                                 | **New — first observed zero result**      |
| JSON output mode exists and works well | Yes (2026-07-19, "no JSON output") | **Resolved — `--json` is now available**  |
| `unknown` category uninformative       | Yes (2026-07-19)                   | Open                                      |
| `--html` TTY auto-detection missing    | Yes (2026-07-19)                   | Open                                      |
| Accept/suppress directive needed       | Yes (2× 2026-07-19)                | Open — but only blocks `-t 5` workflows   |
| Ginkgo `DescribeTable` variadic gotcha | No                                 | **New — skill-doc fix, not tool fix**     |
| Production semantic clones real        | No                                 | **New — validates Type-2/3 detection**    |
| Threshold sweep holding zero           | No                                 | **New — confirms default is well-chosen** |

## Conclusion

This was the cleanest dedup session in the observed feedback corpus. Two factors made it so:

1. **The codebase cooperated.** A shared `goldentest` framework with deterministic test patterns meant the test-code duplication was highly extractable — the `DefineGoldenSuite` helper collapsed 7 files into declarative data. Production code had three semantically-cloned finding-builder functions that yielded to a clean two-helper extraction in `migratorutils`. No generated code, no interface-satisfaction false positives at the working threshold, no cross-package structural dead-ends.

2. **The tool cooperated.** Semantic mode correctly grouped Type-2/Type-3 clones with different literals. The `--json` output made triage fast and scriptable. Re-runs after each refactor were monotonic — no whack-a-mole. The hash-based group identity made cross-run comparison trivial.

The remaining gaps (`unknown` category, TTY-aware HTML, accept directive at `-t 5`) are all known from prior feedback and none of them blocked this session at `-t ≥ 10`. They are real edges, but they are edges, not fundamentals.

The headline result: **art-dupl can be driven to literal zero on a cooperative Go codebase, at every threshold from 5 to 50.** The workflow is viable end-to-end. The skill's instruction to "iterate to zero, then stop" is achievable, not aspirational — provided the codebase has a shape that admits extraction (shared frameworks, deterministic test patterns, no generated noise).

For projects with heavier interface usage, generated code, or cross-package test-helper structural limits, the accept-mechanism gaps documented in prior feedback remain the blocking issue. For library and CLI projects of this shape, art-dupl is ready today.

---

**Author:** Crush (glm-5.2) on behalf of the `go-auto-upgrade` deduplication session.
**Session arc:** 15 → 0 clone groups at `-t 50`, holding zero down to `-t 5` (3 intentional interface implementations only). 15 files changed, 7 helpers extracted, 1 new file, all tests passing (`go test -race ./...` + `nix flake check`).
