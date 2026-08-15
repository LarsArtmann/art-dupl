# Feedback: Language-specific idiom noise at threshold 1 in `upd`

**Date:** 2026-07-15
**Project:** upd — Go CLI for bumping NPM dependencies (`github.com/LarsArtmann/upd`, ~2 500 LOC)
**Command:** `art-dupl --semantic --sort total-tokens -t 1`
**Goal:** Find real, fixable duplication with zero false positives and zero false negatives.

> **🔄 PARTIALLY ADDRESSED (updated 2026-07-16):** This feedback was analyzed and several suggestions were implemented in the **2026-07-16 semantic precision hardening sessions** (commits `930b91a`, `61aeca8`). **Implemented:** (1) Actionability patterns expanded — 15+ patterns now correctly work (the `Fingerprint` bug that broke them was fixed); (2) Literal value normalization — semantic mode now hashes BasicLit KIND not VALUE, detecting Type-2 clones with different literal values; (3) Generic type parameter alpha-normalization. **Not yet implemented:** `--test-threshold` flag (still feedback #1 request), helper-call-site detection (needs call-graph resolution), `--dump-tokens` debug flag, fixability scoring. The threshold floor of 3 was effectively superseded by raising the default to 5. At `-t 5`, the `upd` project produces 0 clone groups (already clean).

## Session Summary

Ran `art-dupl` at `-t 1` on a small, well-maintained Go project. The report surfaced **16 clone groups**. After manual review, only **2 were genuinely harmful** and worth extracting:

1. **Repeated blank-line writes in `PrintUsage`** — 4 anonymous `_, _ = fmt.Fprintln(w)` calls. Extracted into `usageBlankLine(w)`.
2. **Repeated table-border draws in `renderUpgradeTable`** — 3 identical `borderChars(...) + writeBorder(...)` sequences. Extracted into `Renderer.renderBorder`.

After the refactor, the re-run showed **15 clone groups** (down from 16). The border group disappeared; the blank-line calls now show up as 4 occurrences of `usageBlankLine(w)` — intentional helper reuse, not harmful duplication.

## What Remains (False Positives / Intentional Idioms)

The remaining 15 groups are single-statement Go idioms that are not fixable without making the code worse:

| Pattern                                | Examples                                                    | Count | Verdict                                        |
| -------------------------------------- | ----------------------------------------------------------- | ----- | ---------------------------------------------- |
| `return nil`                           | `cmd/upd/main.go`, `packagejson.go`, `render.go`, `diff.go` | 6     | Idiomatic terminal statement.                  |
| `return false`                         | `config.go`, `pnpm.go`, `engine.go`, `manifest.go`           | 4+    | Standard boolean guard result.                 |
| `return true`                          | `engine.go`, `manifest.go`                                  | 3     | Same as above.                                 |
| `return updates, errors`               | `engine.go`                                                 | 2     | Named-return pair in different functions.      |
| `var x []string`                       | `benchmark_test.go`, `packagejson.go`                       | 2     | Variable declaration shape.                    |
| `err := json.Unmarshal(raw, &v)`       | `pnpm.go`, `packagejson.go`                                  | 3     | Standard JSON parsing pattern; targets differ. |
| `_, err := dec.ReadToken()`            | `packagejson.go`                                            | 2     | JSON decoder streaming pattern.                |
| `seconds, err := strconv.Atoi(header)` | `pnpm.go`, `progress.go`                                     | 2     | Standard string parsing.                       |
| `if r.noColor { return text }`         | `render.go`                                                 | 2     | Color-guard early return; semantics differ.    |
| `cfg := DefaultConfig()`               | `config.go`, `config_test.go`                               | 2     | Test setup pattern.                            |

## Why These Are False Positives for a Fixability Goal

### One-liner idioms are not duplication

A `return nil` appears hundreds of times in any Go project. Treating it as a clone would imply extracting a `returnNil()` helper, which is absurd and would violate Go conventions.

### Helper call sites are intentional reuse, not clones

After extracting `usageBlankLine(w)`, art-dupl correctly reports the 4 call sites as structurally identical. This is **not a bug** — it is the tool accurately reflecting that the helper is consumed in multiple places. However, for a "fixable duplication" goal, these call sites should not be surfaced as actionable clones. They are the _result_ of deduplication, not new duplication.

### Language-specific constructs share unavoidable shape

`json.Unmarshal(raw, &v)` followed by error handling is the canonical Go JSON parsing pattern. The target struct, error message, and surrounding logic differ, but the unmarshal call itself is identical. Without type-flow awareness, the tool cannot distinguish "same operation, different data" from "same operation, same abstraction".

## Suggestions for Zero-FP/FN

### 1. Expand actionability patterns to cover one-liner idioms

The existing `single-call-expression` pattern is close, but should be extended to cover entire statements that are universally idiomatic in Go:

- `return nil` / `return false` / `return true` / `return 0` / `return ""`
- `var x []T` / `var x map[K]V` / `var x string` / `var x error`
- `_, _ = fmt.FprintX(w, ...)` (error suppression in printers)
- `err := json.Unmarshal(raw, &v)` (when the only variable is the target)
- `seconds, err := strconv.Atoi(...)` / `n, err := strconv.Atoi(...)`
- `tok, err := dec.ReadToken()` / `objTok, err := dec.ReadToken()`
- `t.Run(tt.name, func(t *testing.T) { ... })` fragments in table-driven tests

A group should be marked `NonActionable` if every clone is one of these atomic idioms, even at threshold 1.

### 2. Distinguish helper-consumption clones from anonymous clones

When a clone group consists entirely of calls to the _same user-defined function or method_, treat it as `NonActionable` or a new category `helper-reuse`. For example:

```go
usageBlankLine(w)        // 4 call sites in config.go
r.renderBorder("top", ...) // 3 call sites in render.go
```

This requires call-graph or identifier resolution. Without it, the tool will always re-report clones after a successful extraction, making the "zero false negatives" goal feel unattainable.

### 3. Add threshold-aware output warnings

At `-t 1`, the false-positive rate is near 100% because the tool is matching single statements. A report summary should warn:

```
⚠️  Threshold 1 is below the recommended minimum (5).
   Expect a high false-positive rate from single-statement Go idioms.
```

This sets realistic expectations and aligns with the existing skill guidance that recommends `-t 5`.

### 4. Consider a "fixability score" instead of binary classification

Binary `Actionable` / `NonActionable` is too coarse for zero-FP/FN. A scored model could weigh:

- **Size of clone** (larger = more likely real duplication)
- **Number of occurrences** (2 occurrences of a 2-line pattern is weaker than 5 occurrences of a 5-line pattern)
- **Idiom density** (how many tokens are in the known-idiom allow-list)
- **Cross-package vs. same-file** (cross-package duplication is more likely real)
- **Whether the fragment is inside a helper function** (helper call sites are low-value)

Only clones above a fixability threshold would be reported as actionable, while others get a lower priority or are hidden entirely.

## Metrics

| Metric                               | Value                                         |
| ------------------------------------ | --------------------------------------------- |
| Clone groups at `-t 1` (initial)     | 16                                            |
| Clone groups after deduplication     | 15                                            |
| Real duplications found              | 2                                             |
| Extracted helpers                    | 2 (`usageBlankLine`, `Renderer.renderBorder`) |
| False positives / intentional idioms | ~15 groups                                    |
| Estimated FP rate at `-t 1`          | ~93%                                          |
| FP rate at `-t 5` (estimated)        | Likely 0–1 group                              |

## Conclusion

`art-dupl` correctly detected the structural clones. The problem is not detection accuracy — it is **classification of what is worth fixing**. For a zero-FP/FN goal, the tool needs more aggressive, language-specific idiom filtering and a way to recognize that helper call sites are intentional reuse.

At `-t 5`, this project is likely already clean. The value of running at `-t 1` is in finding tiny helpers that `-t 5` misses, but only if the noise can be filtered. The biggest improvement would be recognizing that **single-statement Go idioms are not duplication**, even when they appear in many files.

---

**Author:** Crush (kimi-for-coding) on behalf of the `upd` deduplication review.
