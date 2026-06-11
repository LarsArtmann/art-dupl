# Comprehensive Status Report — 2026-06-11

> **Date:** 2026-06-11 13:57
> **Branch:** fork (up to date with origin)
> **Lint:** 0 issues | **Tests:** 25/25 pass | **Working tree:** clean

---

## A) FULLY DONE ✅

### This Session (5 commits by previous assistant, 5 by current session)

| Commit | Description |
|--------|-------------|
| `8889a7a` | Fix all 3 godoclint warnings → zero lint issues |
| `6538918` | Comprehensive status report |
| `d2952f9` | GCI/WSL lint fixes |
| `76d1db4` | TODO_LIST.md update |
| `2583add` | GetCategoryEmoji complexity 16→2 |
| `49ae11a` | FuncType-based interface implementation detector |
| `b4ea5f8` | HealthScore typed enum |
| `60f402d` | Decouple clone_classify.go from syntax/golang |
| `9741041` | `--output-file` flag + scanner.Err() fix |
| `a371919` | CSV output using `encoding/csv` |
| `e854653` | BDD tests for `--only` and `--include-generic` |
| `e82d6b5` | SDK godoc documentation |
| `04b9acd` | GoReleaser nix install fix |

### Stable, Working Features

- **25/25 test packages** passing with 0 failures
- **0 lint issues** via `golangci-lint run`
- **7 output formats**: text, HTML, JSON, simple-JSON, plumbing, SARIF, CSV
- **2 languages**: Go (full AST) + Templ (pure Go parser)
- **2 detection methods**: suffix tree (art-dupl) + hash-based (XXH3)
- **Smart filtering**: sqlc, templ, protobuf, mockgen, stringer, generic generated code
- **Stats subcommand**: text/JSON/CSV formats, A-F health grade, recommendations
- **Professional CLI**: Fang/Cobra, auto-completion, version info
- **GoReleaser**: multi-arch builds (linux/darwin/windows, amd64/arm64), Docker, Homebrew, Nix, NFPM
- **CI/CD**: GitHub Actions release workflow with cosign signing + SBOMs

---

## B) PARTIALLY DONE ⚠️

### 1. Documentation Drift — Multiple Files Out of Date

| File | Last Updated | Issue |
|------|-------------|-------|
| `FEATURES.md` | 2026-05-02 | CSV still marked "PARTIALLY_FUNCTIONAL", `--output-file` not documented, actionability patterns missing |
| `CHANGELOG.md` | 2026-03-28 | No entries for: HealthScore type, `--output-file`, CSV refactor, actionability patterns, exhaustive switch fix |
| `docs/DOMAIN_LANGUAGE.md` | Never | Empty template with placeholder "Example Term" — never populated |
| `TODO_LIST.md` | 2026-06-11 | MEDIUM items "CSV encoding/csv" and "--output-file" still marked unchecked despite being done |

### 2. Enum Unification — Analyzed but Not Executed

**What was found:**
- `domain/CloneSeverity` and `domain/HealthScore` have hand-written JSON marshal/unmarshal boilerplate (~40 lines each)
- `config/enum_helpers.go` has generic helpers (`marshalStringType`, `unmarshalStringTypeToPointer`, `isValidStringType`)
- `domain/helpers.go` has **parallel** generic helpers (`marshalStringID`, `unmarshalWithValidation`)
- Domain types **cannot** import config (layer violation: domain is core, config is infrastructure)
- Config's helpers return `null` for invalid values (omitempty-style), domain's helpers return **errors** for invalid values (strict)
- **Decision was to skip** — semantic difference means unification would lose meaning

**What SHOULD have been done instead:** Extract shared helpers to `internal/enum/` package that both `domain/` and `config/` can import, with both strict and permissive modes.

### 3. Actionability Layer Violation — Partially Fixed

- `printer/clone_classify.go` was decoupled from `syntax/golang` (moved `nodeTypeNames` map)
- **But** `printer/actionability.go` STILL imports `syntax/golang` directly (line 9-10)
- **And** `printer/clone_processor.go` STILL imports `syntax/golang` directly (line 9)
- All 8 actionability detectors use Go-specific AST constants
- The layer violation was acknowledged but not fully resolved

---

## C) NOT STARTED ❌

### HIGH Priority (from TODO_LIST.md)

1. **ProcessedClone DTO** — Decouple printer from `syntax.Node` (111 test call sites)
2. **Clone type consolidation** — 4 parallel types: `printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup`, `domain.ProcessedClone`
3. **TokenValue type** — Validated type for suffix tree tokens

### MEDIUM Priority

4. **Extract shared enum helpers** to `internal/enum/` (consolidate domain/helpers.go + config/enum_helpers.go)
5. **Add missing methods to domain enums** — `CloneCategory`, `ClonePriority`, `CloneActionability` all lack `IsValid()`/`String()`/JSON methods
6. **Fix `priorityScore()` in printer/stats.go** — uses raw strings instead of `domain.ClonePriority` constants
7. **Fix LSP stale diagnostic cache** — exhaustive/cyclop/godoclint warnings from stale LSP, not real code issues

### LOW Priority

8. **Transform.go table-driven refactor** — 369L, 300L switch. Skipped — each case has unique logic, table would be harder to read.
9. **String interning** — Referenced in docs but never implemented. `syntax.Node.Filename` duplicated across thousands of nodes.
10. **Populate `docs/DOMAIN_LANGUAGE.md`** — Still an empty template
11. **Fuzz tests for templ parser** edge cases
12. **Wire TODO/Legacy detectors** to CLI (implemented but not exposed)

---

## D) TOTALLY FUCKED UP 💥

### 1. CHANGELOG.md is a Lie

- Claims "Semantic detection now OFF by default" — but `config.DefaultConfig.Semantic = true` (ON by default since the previous session changed it)
- Claims "SIMD optimizations: Vectorized transition search" — but `internal/simd/` was deleted, SIMD code was removed
- Claims "Memory efficiency: String interning pool" — no interning pool exists (was fabricated, then deleted)
- Has duplicate `### Changed`, `### Added`, `### Technical` sections
- No entries for 2+ months of work (May-June 2026)

### 2. FEATURES.md Semantic Detection Stale

- Doesn't mention semantic is now the DEFAULT
- CSV output still marked "PARTIALLY_FUNCTIONAL" despite being fully refactored with `encoding/csv`

### 3. LSP Diagnostic Cache is Corrupted

The gopls/golangci-lint LSP shows 6 warnings that are **not real**:
- `ErrInvalidHealthScore` undefined — code compiles fine, `just check` = 0 issues
- `GetCategoryEmoji` cyclop 16 — already refactored to map lookup, complexity is 2
- `exhaustive` missing cases — all 9 PatternLabel values are handled
- `godoclint` "package has more than one godoc" — references `doc.go` that doesn't exist
- `zeebo/xxh3` missing go.sum — builds fine with vendored deps

These are stale LSP diagnostics, not real issues. LSP cache needs clearing.

### 4. `docs/DOMAIN_LANGUAGE.md` Never Populated

Empty template with "Example Term" placeholders. Every AI session sees this file and ignores it because it has no content. This defeats the purpose of domain-driven design documentation.

---

## E) WHAT WE SHOULD IMPROVE

### Type Models

1. **`domain.CloneCategory`** — 14 constants, zero methods. No `IsValid()`, no `String()`, no JSON marshaling. Same package as `CloneSeverity` which has all of these.
2. **`domain.ClonePriority`** — 4 constants, only display methods. Missing `IsValid()`/`String()`/JSON.
3. **`domain.CloneActionability`** — 2 constants, zero methods. Completely bare.
4. **`printer/stats.go:priorityScore()`** — Maps raw strings `"critical"→4` instead of using `domain.ClonePriority`. Fragile.
5. **`domain/helpers.go`** duplicates `config/enum_helpers.go`** — Both have generic string marshaling helpers. Should be in shared `internal/enum/`.

### Architecture

6. **`printer/` is a god package** — 51 files mixing output formatting with business logic (classification, actionability, sorting, clone processing).
7. **4 parallel Clone types** — `printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup`, `domain.ProcessedClone`. Same concept, different shapes.
8. **Layer violations** — `printer/actionability.go` and `printer/clone_processor.go` import `syntax/golang` directly.
9. **`detection/todo_detector.go`** re-parses files with `go/parser` instead of using the existing `syntax/` layer.

### Code Reuse

10. **`BuildCloneGroups()`** in `printer/groups.go` duplicates `collectMatchesIntoGroups()` in `pkg/artdupl/detector_pipeline.go`.
11. **Threshold validation** duplicated between `config/config_validate.go` and `pkg/artdupl/types.go`.
12. **`isTestFile()`** in `printer/clone_classify.go` is domain-level logic trapped in the printer package.

### Libraries

13. **`golang.org/x/exp`** in go.mod but `slices`/`maps` packages not used (manual dedup, manual nil filtering).
14. **`pkg/format/hash.go`** — Single 17-line file, questionable package value.

---

## F) Top #25 Things We Should Get Done Next

Sorted by **Impact × Effort⁻¹** (highest ROI first):

### Tier 1: High Impact, Low Effort (Quick Wins)

| # | Task | Impact | Effort | Why |
|---|------|--------|--------|-----|
| 1 | Update `TODO_LIST.md` — mark completed items | Medium | 5min | Simple housekeeping, prevents confusion |
| 2 | Update `CHANGELOG.md` — fix lies, add entries | High | 20min | Document is actively misleading |
| 3 | Update `FEATURES.md` — CSV, `--output-file`, actionability | Medium | 10min | Stale since May |
| 4 | Populate `docs/DOMAIN_LANGUAGE.md` | High | 15min | Empty template = dead documentation |
| 5 | Add `IsValid()/String()` to `CloneCategory`, `ClonePriority`, `CloneActionability` | Medium | 15min | Consistency with `CloneSeverity`/`HealthScore` |
| 6 | Fix `priorityScore()` to use `domain.ClonePriority` constants | Low | 5min | Eliminates fragile raw strings |
| 7 | Restart LSP to clear stale diagnostic cache | Low | 1min | 6 phantom warnings |

### Tier 2: High Impact, Medium Effort (Architecture)

| # | Task | Impact | Effort | Why |
|---|------|--------|--------|-----|
| 8 | Extract shared enum helpers to `internal/enum/` | Medium | 30min | Eliminates parallel implementations |
| 9 | Decouple `printer/actionability.go` from `syntax/golang` | High | 1h | Last layer violation in printer |
| 10 | Extract `printer/clone_classify.go` + `actionability.go` to `detection/classify/` | High | 2h | Business logic should not live in printer |
| 11 | Consolidate Clone types: define canonical `domain.ProcessedClone` | High | 4h | 4 types for same concept is madness |
| 12 | Add `--output-file` support to root command (not just stats) | Medium | 30min | Users expect this on all subcommands |

### Tier 3: High Impact, High Effort (Major Refactor)

| # | Task | Impact | Effort | Why |
|---|------|--------|--------|-----|
| 13 | ProcessedClone DTO — decouple printer from `syntax.Node` | Very High | 8h | 111 test call sites, biggest coupling issue |
| 14 | TokenValue type with validation | High | 4h | Type safety for suffix tree tokens |
| 15 | Wire TODO/Legacy detectors to CLI | Medium | 2h | Implemented but inaccessible to users |
| 16 | Fix `detection/todo_detector.go` to use existing `syntax/` parsing | Medium | 2h | Eliminates redundant re-parsing |

### Tier 4: Nice to Have

| # | Task | Impact | Effort | Why |
|---|------|--------|--------|-----|
| 17 | Use `x/exp/slices` for dedup/filtering in `config/detection_method.go` | Low | 15min | Already in go.mod |
| 18 | String interning for `syntax.Node.Filename` | Medium | 2h | Memory optimization for large codebases |
| 19 | Consolidate `pkg/format/hash.go` into `pkg/format/` or inline it | Low | 10min | Single 17-line file, questionable package |
| 20 | Fuzz tests for templ parser edge cases | Medium | 3h | Coverage for parser robustness |
| 21 | Move `isTestFile()` from printer to domain | Low | 15min | Domain logic in wrong package |
| 22 | Deduplicate `BuildCloneGroups()` and `collectMatchesIntoGroups()` | Low | 30min | Same function in two packages |
| 23 | Add `--verbose` output for stats `--output-file` path confirmation | Low | 10min | User experience |
| 24 | Investigate `charm.land/fang/v2` migration (from v1) | Low | 2h | Library policy warning in CI |
| 25 | Implement proper general clone CSV output (not just stats) | Medium | 2h | Stats CSV is done, clone CSV is not |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should the ProcessedClone DTO refactor (Tier 3, item #13) be done BEFORE or AFTER the Clone type consolidation (item #11)?**

Rationale: The ProcessedClone DTO introduces a clean boundary between detection (`syntax.Node`) and output (`ProcessedClone`). But there are 4 parallel Clone types that should be consolidated into one canonical form. Doing ProcessedClone first means the consolidated type would need to be redesigned to match the DTO. Doing consolidation first means touching 111 test call sites twice.

**My recommendation:** Do ProcessedClone DTO first — it's the bigger win (breaks the `syntax.Node` coupling), and the consolidation naturally follows because the DTO becomes the canonical type.

---

## Commit History This Session

```
04b9acd fix(goreleaser): remove unnecessary -r flag from nix install cp command
e82d6b5 docs(sdk): write comprehensive godoc for pkg/artdupl SDK package
e854653 test(bdd): add BDD tests for --only and --include-generic with stats
a371919 refactor(printer): use encoding/csv for proper CSV escaping in stats output
9741041 feat(stats): add --output-file flag and fix unchecked scanner.Err()
```

Previous session (already pushed):
```
8889a7a fix(lint): eliminate all 3 godoclint warnings to achieve zero lint issues
6538918 docs(status): add comprehensive review and planning report
d2952f9 fix(lint): resolve gci alignment and wsl whitespace issues
76d1db4 docs: update TODO_LIST.md with completed items from session
2583add refactor(domain): reduce GetCategoryEmoji cyclomatic complexity 16→2
49ae11a fix(printer): improve test scaffolding detector with relaxed signals
4d1d664 feat(printer): add granular non-actionable pattern detection with AST-based classification
7dc69d6 docs: add false-positive patterns report and update planning/status docs
```
