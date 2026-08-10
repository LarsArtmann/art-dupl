# Feedback: `art-dupl` — Threshold Cliff, Actionability False Positives on Type Aliases & Helper Calls, Missed Structural Pipeline Duplication

**Date:** 2026-08-07
**Tool Version:** `art-dupl` v0.6.1-c170f4d
**Tested On:** `~/projects/file-and-image-renamer` — 224 Go files, 218 with type information
**Detection Mode:** `--type-aware --sort total-tokens`
**Commits this session:** `ce40612` (dedup constants/injector/healthd), `7a1eb0b` (dedup CLI banners/providers), `b418ca1` (formatting)

---

## TL;DR

Ran a full deduplication sprint using `art-dupl -t 1` against the `file-and-image-renamer` codebase. **The tool correctly identified 5 real clone groups that were eliminated**, but the remaining 4 groups expose three tool-level gaps:

1. **Actionability mislabeling** — Every remaining clone is labeled "actionable" with "fix: Review and extract common logic", but at least 2 are genuinely NOT actionable (canonical type-alias re-export and helper-call-with-different-args). The `type-alias-block` and `single-call-expression` patterns exist in `--list-patterns` but never fire for these cases.

2. **Threshold cliff** — At `-t 1`: 9 actionable groups. At `-t 2`: 3 groups. At `-t 3`: **0 groups**. The recommended threshold (`-t 5`) also shows 0. There is zero useful signal between "noise flood" and "silence". The tool needs either finer-grained thresholds (statement fractions) or a `--explain-why-suppressed` mode that shows what was filtered at each level.

3. **Missed structural duplication** — The single largest maintenance burden in the codebase (`rename_explain.go` and `rename_process.go` share a 50+ line quality → AI → clean → combine → improvement pipeline) is invisible to art-dupl because the surrounding code differs in print formatting and branching. The tool catches the 1-line `basename := filepath.Base(imagePath)` overlap but misses the 50-line structural clone around it.

**Scorecard:** 5 real catches (eliminated), 4 remaining (2 false-positive "actionable", 2 borderline), 1 major structural clone missed.

---

## Detection Tally

### Phase 1: Initial scan (before refactoring)

```
art-dupl --type-aware --sort total-tokens -t 1
→ 9 clone groups, 21 clones total
```

| #   | Clone Group                                                | Locations                                    | Verdict            | Resolution                                                      |
| --- | ---------------------------------------------------------- | -------------------------------------------- | ------------------ | --------------------------------------------------------------- |
| 1   | `fmt.Println(title); fmt.Println(strings.Repeat(...))` × 5 | deadletter, health, query_display, explain, quality_only | **REAL CATCH**     | Extracted to existing `printHeading` helper                     |
| 2   | `DefaultDebounceDelay = 2 * time.Second` etc. × 2         | filechange/config.go, pkg/constants/timeouts.go | **REAL CATCH**     | Removed dead constants from pkg/constants (filechange is now a separate module) |
| 3   | `for _, inv := range invokers { ... }` × 2                | injector/eager_invoke.go (two functions)     | **REAL CATCH**     | Extracted `runInvokers` helper                                  |
| 4   | `filters := views.ParseEventFilters(...); data := s.buildEventsData(...)` × 2 | healthd/events_api.go (two handlers)        | **REAL CATCH**     | Extracted `loadEventsData` helper                               |
| 5   | `name := do.NameOf[T](); if c.isOverridden(name) { return }` × 2 | injector/providers.go (provide + provideValue) | **REAL CATCH**     | Extracted `skipIfOverridden[T]` helper                          |
| 6   | `printHeading("X", "=", 50)` × 2                           | query_display.go (two different headings)    | **FALSE POSITIVE** | Helper already used; different args = different domain content  |
| 7   | `basename := filepath.Base(imagePath)` × 2                 | rename_explain.go, rename_process.go         | **BORDERLINE**     | 1-line idiomatic Go; structural pipeline around it is the real dup |
| 8   | `d.processed[path] = time.Now(); d.mu.Unlock()` × 2        | filechange/cleanup.go, filechange/processing.go | **BORDERLINE**    | Critical-section tail vs test helper; different lock contexts   |
| 9   | `ErrorType = domainerrors.ErrorType` × 2                   | deadletter/deadletter.go, provider/provider.go | **FALSE POSITIVE** | Canonical type-alias re-export per domain-driven design pattern |

**Scorecard:** 5 real catches (eliminated), 2 false positives, 2 borderline, 0 missed at this threshold.

### Phase 2: Post-refactoring scan

```
art-dupl --type-aware --sort total-tokens -t 1
→ 4 clone groups (groups 6–9 above remain)
```

### Phase 3: Threshold sweep

| Threshold | Clone Groups | Notes                                             |
| --------- | ------------ | ------------------------------------------------- |
| `-t 1`    | 9 → 4        | Default actionability filtering; groups 6–9 remain |
| `-t 2`    | 3            | Groups 6–8 only; group 9 (type alias) drops out    |
| `-t 3`    | **0**        | Complete silence                                   |
| `-t 5`    | **0**        | Recommended threshold for this codebase size       |

### Phase 4: Raw scan (no actionability filtering)

```
art-dupl --type-aware --sort total-tokens -t 1 --no-actionability
→ 102 clone groups
```

The actionability filter removes 93 groups (102 → 9). This is effective, but the remaining 9 include 2 false positives that the filter should have caught.

---

## Issue 1: Actionability False Positives — `type-alias-block` Pattern Never Fires

### The clone

```go
// pkg/deadletter/deadletter.go:17
type ErrorType = domainerrors.ErrorType

// pkg/provider/provider.go:31
type ErrorType = domainerrors.ErrorType
```

### art-dupl's verdict

```
explain: type-1 | actionable | unknown | 1 tokens, 1 lines
fix: Review and extract common logic
```

### Why this is wrong

This is a **canonical type-alias re-export** — a deliberate Go pattern where a package exposes a type from a domain package so consumers don't need to import the domain package directly. The project's AGENTS.md explicitly documents this pattern:

> **`domain.FilenameGenerator` is the canonical interface** — implemented by `provider.VisionAdapter`

Both `deadletter` and `provider` re-export `ErrorType` and its constants from `domain/errors` because:
- `deadletter.Entry` serializes error types in its JSON schema
- `provider.Error` wraps `domainerrors.TypedError` for consumer-facing error handling

Extracting these into a shared location would either (a) force every consumer to import `domain/errors` directly (breaking the facade pattern), or (b) require a new shared package that both re-export from (indirection for zero benefit).

### Root cause: pattern exists but doesn't match

`art-dupl --list-patterns` includes `type-alias-block`:

```
type-alias-block
```

But the pattern apparently only fires for multi-line alias blocks (multiple `const` or `type` declarations in a group), not for a single-line `type ErrorType = domainerrors.ErrorType`. The pattern should also suppress single-line type alias re-exports, which are equally intentional.

### Suggested fix

Extend the `type-alias-block` pattern (or add a `type-alias-reexport` pattern) to match:
- `type X = pkg.Y` (single-line type alias)
- Especially when the RHS is a qualified identifier (`pkg.Type`), which indicates a cross-package re-export rather than a same-package alias

---

## Issue 2: Actionability False Positives — Helper Call With Different Arguments

### The clone

```go
// cmd/file-renamer/query_display.go:45-46
printHeading("Hash Database Statistics", "=", 50)
fmt.Printf("Total Unique Files:   %d\n", totalFiles)

// cmd/file-renamer/query_display.go:66-67
printHeading("History Log Statistics", "=", 50)
fmt.Printf("Total Operations:   %d\n", total)
```

### art-dupl's verdict

```
explain: type-2 | actionable | unknown | 2 tokens, 2 lines
fix: Review and extract common logic
```

### Why this is wrong

The clone is a **type-2** (semantic — renamed variables). But the "different names" here are not semantic clones of each other — they are **different domain content**. `"Hash Database Statistics"` and `"History Log Statistics"` are different section headings for different data sources. `totalFiles` and `total` are different variables from different stores.

The two call sites already USE the extracted helper (`printHeading`). The "duplication" art-dupl sees is the call shape `printHeading(stringLiteral, "=", 50)` followed by a `fmt.Printf` — but the literal content is unique domain data, not a renamed clone.

This is the "helper call with different arguments" pattern: two call sites of the same function with different meaningful arguments. The `single-call-expression` pattern exists in `--list-patterns` but apparently doesn't suppress this because the call is followed by a `fmt.Printf` (2 statements, not 1).

### Suggested fix

Either:
1. Extend `single-call-expression` to cover "N call sites of the same function with all-different string-literal arguments"
2. Add a `helper-call-different-args` pattern that suppresses when the only shared tokens are the function name and constant arguments (like `"="` and `50`), while all variable content (the heading text) differs

---

## Issue 3: Threshold Cliff — No Useful Signal Between t=1 and t=3

### The problem

| Threshold | Result   |
| --------- | -------- |
| t=1       | 9 groups |
| t=2       | 3 groups |
| t=3       | 0 groups |
| t=4       | 0 groups |
| t=5       | 0 groups |

There is a **binary cliff** between t=2 and t=3. The tool goes from "3 clone groups" to "completely clean" with a single threshold step. This means:

1. **Users who want to push duplication to zero** have no useful threshold between "noisy" (t=1, 9 groups including false positives) and "silent" (t=3, nothing).
2. **The recommended threshold** (`--recommend-threshold` suggests t=5 for this codebase) hides everything — including the 5 real catches that were eliminated this session. Those 5 groups existed at t=1 but not at t=5.
3. **CI gates** using the recommended threshold would never fire on this codebase, giving false confidence that there's no duplication.

### Why this happens

The `-t` flag counts duplicated **statements**. Most of the real duplication in this codebase was 2–4 statements long:
- `fmt.Println(title); fmt.Println(strings.Repeat(...))` — 2 statements
- `name := do.NameOf[T](); if c.isOverridden(name) { return }` — 3 statements
- `for _, inv := range invokers { ... return ... }` — 4 statements (loop + body)

At t=5, all of these are below the threshold and invisible. At t=1, single-statement clones like `t.Helper()` produce 34-member clone groups.

### Suggested improvements

1. **Fractional thresholds** — Allow `-t 2.5` or use a separate `--min-tokens` flag that operates on token count, not statement count. The `--explain` output already reports token counts (e.g., "2 tokens, 2 lines").
2. **`--show-suppressed` for intermediate thresholds** — The `--show-suppressed` flag exists but only shows actionability-suppressed clones, not threshold-suppressed ones. A `--explain-threshold` mode that says "at t=5, these N groups were suppressed because they have fewer than 5 statements" would help users calibrate.
3. **Recommend a threshold range, not a single value** — Instead of "Recommended threshold: 5", suggest "Recommended threshold: 3–5 for CI gates; use t=1–2 for periodic deep audits with `--explain` to filter false positives manually".

---

## Issue 4: Missed Structural Duplication — The `rename_explain.go` / `rename_process.go` Pipeline

### The real duplication

`rename_explain.go:explainFile()` and `rename_process.go:processFile()` implement the **same processing pipeline** with different output formatting:

```
1. basename := filepath.Base(imagePath)
2. quality := renamer.AssessFilenameQuality(basename)
3. printQualityAssessment(basename, quality)
4. quality gate check (RecommendationKeep → skip)
5. dataURL, err := renamer.ImageToBase64(imagePath)
6. description, usage, err := s.gen.GenerateFilename(ctx, dataURL, s.cfg.Prompt)
7. s.totalTokens += usage.TotalTokens
8. description = renamer.CleanFilename(description)
9. CorrectTypos
10. DetectGarbageImage
11. empty check
12. ext := strings.ToLower(filepath.Ext(imagePath))
13. originalName := strings.TrimSuffix(basename, ext)
14. CombineNames / CombineNamesWithDecision
15. AnalyzeFinalImprovement
```

This is a **15-step pipeline** duplicated across two files. Steps 1–14 are semantically identical — the only differences are:
- Print formatting (`fmt.Printf("→ Quality gate: WOULD SKIP")` vs `fmt.Println("✓ Filename is already well-named")`)
- The branching after step 14 (explain mode prints the decision trace; process mode prompts for y/n and executes the rename)

### What art-dupl catches

Only step 1:
```
basename := filepath.Base(imagePath)
```

This is a **1-statement clone** (type-1, 2 tokens). Everything else in the pipeline is invisible because the variable names, print statements, and branching differ enough to break token-level matching.

### What art-dupl should catch

The structural pipeline — 15 steps in the same order, calling the same `renamer.*` functions, with the same data flow. This is a **type-3 clone** (semantic/structural — same logic, different syntax). The tool's `--semantic` mode (alpha-normalization) helps with renamed variables but doesn't help when the surrounding print/branch code creates enough token divergence to break the matching window.

### Impact

This is the **single largest maintenance burden** in the codebase. When `renamer.CombineNames` changes its signature, both files must be updated in lockstep. When a new AI safety check is added (e.g., `DetectGarbageImage`), it must be added to both pipelines. The 1-statement `basename` clone is noise; the 15-step pipeline clone is the real problem.

### Suggested improvements

1. **Control-flow graph (CFG) matching** — Beyond token/AST matching, compare the sequence of function calls in each function body. Two functions that call `AssessFilenameQuality → ImageToBase64 → GenerateFilename → CleanFilename → CorrectTypos → DetectGarbageImage → CombineNames → AnalyzeFinalImprovement` in the same order are structural clones regardless of the interleaving print statements.
2. **Call-sequence fingerprinting** — Hash the ordered list of external function calls (qualified by package) in each function body. Two functions with the same call-sequence fingerprint are structural clones even if the surrounding code differs.
3. **Data-flow matching** — Track the data dependencies: `basename` flows into `AssessFilenameQuality` which produces `quality` which flows into `CombineNames`. Two functions with the same data-flow DAG are structural clones.

---

## Issue 5: `--no-actionability` Raw Results Show Filtering Effectiveness But Also Gaps

### The numbers

| Mode                     | Clone Groups |
| ------------------------ | ------------ |
| t=1 (actionability on)   | 9            |
| t=1 (actionability off)  | 102          |
| Filtering ratio          | 91%          |

The actionability filter removes 93 of 102 groups. This is **effective**. However, examining the raw output reveals patterns that the filter catches but that could be useful signal in some contexts:

### Useful groups hidden by actionability filtering

| Raw Clone                                                        | Count | Pattern             | Hidden because...                          |
| ---------------------------------------------------------------- | ----- | ------------------- | ------------------------------------------ |
| `t.Helper()`                                                     | 34    | test-scaffolding    | Correctly suppressed — test boilerplate    |
| `t.Parallel()`                                                   | 51    | test-scaffolding    | Correctly suppressed — test boilerplate    |
| `container := injector.NewContainer(injector.WithConfig(cfg))`   | 2     | single-call-expression | Correctly suppressed — DI wiring          |
| `cfg := config.LoadOrDefault()`                                  | 2     | single-call-expression | Correctly suppressed — config loading     |
| `if apiKey == "" { return ..., fmt.Errorf(...) }`                | 2     | guard-clause        | Correctly suppressed — guard pattern       |
| `os.ReadFile(path)` / `os.UserHomeDir()` / `os.Getenv(...)`      | 6     | single-simple-statement | Correctly suppressed — stdlib calls     |

### Gaps in the filtered set (false positives that survive filtering)

| Surviving Clone                             | Should be filtered?            | Pattern that should fire         |
| ------------------------------------------- | ------------------------------ | -------------------------------- |
| `ErrorType = domainerrors.ErrorType`        | **Yes** — type alias re-export | `type-alias-block` (doesn't fire) |
| `printHeading("X", "=", 50)` with diff X   | **Yes** — helper call, diff args | `single-call-expression` (doesn't fire) |

---

## What Worked Well

### 1. `--explain` output is excellent

The per-group categorization is clear and actionable:

```
explain: type-1 | actionable | assignment | 2 tokens, 2 lines
fix: Review and extract common logic
```

The **type-1 / type-2** distinction, **category** label, and **token count** give enough information to triage without reading the source. The fix suggestion ("Review and extract common logic") is generic but appropriate for actionable clones.

### 2. Actionability filtering is 91% effective

Removing `t.Helper()`, `t.Parallel()`, guard clauses, and single stdlib calls from the default view is the right call. Without it, t=1 produces 102 groups — almost all noise.

### 3. `--recommend-threshold` correctly identifies codebase size

For 225 Go files, recommending t=5 is reasonable for CI gates. The problem is that it also hides the t=1–4 signal that's useful for periodic deep audits.

### 4. `--type-aware` prevents cross-type false matches

The type-aware mode correctly avoids matching `time.Time.String()` against `*big.Int.String()` by encoding static type information into the hash. This is valuable for a codebase with heavy use of generics and interfaces.

### 5. Multi-module support works correctly

The tool correctly analyzed both the main module and the `filechange/` sub-module (separate `go.mod`) without configuration. Files from both modules appeared in the results with correct paths.

### 6. `--list-patterns` shows a thoughtful pattern set

The 28 suppressible patterns cover the common Go idioms that produce false positives:
- `table-driven-test`, `test-scaffolding`, `testdata-pair` — test boilerplate
- `guard-clause`, `error-propagation`, `assign-error-check` — control flow
- `raii-defer`, `interface-implementation` — language patterns
- `cobra-boilerplate`, `templ-rendering-idiom` — framework patterns

The gaps (type-alias-block not firing, single-call-expression not firing) are fixable within this framework.

---

## Summary of Suggested Tool Improvements

| #    | Issue                                              | Severity | Effort  |
| ---- | -------------------------------------------------- | -------- | ------- |
| 1    | `type-alias-block` pattern doesn't fire for single-line aliases | Medium   | Low     |
| 2    | `single-call-expression` doesn't suppress helper calls with different args | Medium   | Medium  |
| 3    | Threshold cliff between t=2 and t=3 (binary on/off) | High     | High    |
| 4    | No structural/type-3 clone detection (call-sequence matching) | High     | High    |
| 5    | `--recommend-threshold` gives single value, not range | Low      | Low     |
| 6    | No `--explain-threshold` to show what's suppressed at each level | Low      | Medium  |
| 7    | Fix suggestion is generic ("Review and extract")   | Low      | Low     |

### Priority recommendation

1. **Fix issues 1–2 first** (pattern matching gaps) — these are low-effort fixes to existing patterns that eliminate 2 of 4 false positives in this codebase.
2. **Investigate issue 3** (threshold cliff) — this affects every user who runs `art-dupl -t 1` and gets noise, or `art-dupl -t 5` and gets silence. Consider token-based thresholds.
3. **Research issue 4** (structural clone detection) — this is the biggest value-add but also the biggest effort. Call-sequence fingerprinting is the lightest-weight approach that would catch the `rename_explain.go` / `rename_process.go` pipeline clone.

---

## Appendix: Full Command Reference

All commands run during this session:

```bash
# Initial scan (9 groups)
art-dupl --type-aware --sort total-tokens -t 1

# Post-refactor scan (4 groups)
art-dupl --type-aware --sort total-tokens -t 1

# Threshold sweep
art-dupl --type-aware --sort total-tokens -t 1   # 9 groups
art-dupl --type-aware --sort total-tokens -t 2   # 3 groups
art-dupl --type-aware --sort total-tokens -t 3   # 0 groups
art-dupl --type-aware --sort total-tokens -t 5   # 0 groups

# Raw scan without actionability filtering (102 groups)
art-dupl --type-aware --sort total-tokens -t 1 --no-actionability

# Explain output (categorization + fix suggestions)
art-dupl --type-aware --sort total-tokens -t 1 --explain

# Recommended threshold
art-dupl --recommend-threshold   # → 5

# Available patterns
art-dupl --list-patterns         # 28 patterns

# JSON output (structured metadata)
art-dupl --type-aware --sort total-tokens -t 1 --json
```

---

## Resolution (2026-08-10)

**IDENTIFIED.** Threshold cliff → ROADMAP (fractional thresholds, --explain-threshold). Type-3 structural clone detection → ROADMAP (CFG matching, highest effort/highest value).
