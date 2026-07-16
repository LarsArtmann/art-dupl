# Changelog

All notable changes to art-dupl will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`--quiet`/`-q` flag**: Suppresses non-essential status output (progress messages, profiling notices). Clone results are still printed to stdout.
- **`--no-color` flag**: Explicitly disables colored output. Complements the `NO_COLOR` environment variable.
- **Typed exit codes**: `ExitCodeForError` maps errors to exit codes: 0=success, 1=general error, 2=config/validation error, 3=internal error, 130=interrupted (SIGINT). Enables CI pipelines to distinguish failure modes.
- **`SuppressionConfig` struct**: Groups `SuppressTestLow`, `TestThreshold`, and `MinLines` into a single value, eliminating 3-parameter function signatures prone to argument-swap bugs.
- **86+ new tests**: Integration tests for actionability patterns via `EvaluateActionabilityWithLabel` (error wrapping, cobra command, builder callback, table-driven test with non-testing receiver), exit code tests, and existing unit test coverage backfilled.
- **`--min-lines` flag**: Suppresses clone groups spanning fewer than N source lines (0 = disabled). Complementary filter to `--threshold`.
- **`--dump-tokens` debug flag**: Outputs the serialized token stream (filename, position, type, semantic hash, name) without running clone detection. Essential for debugging false positives/negatives.
- **KeyValueExpr field name encoding**: Struct field names in composite literals (`Point{X:1}` vs `Size{W:1}`) now encoded into the KeyValueExpr node Type via `encodeSemanticType`. Field names are API surface, not local variables. 3 tests added.
- **Error wrapping detection for 2-stmt bodies**: `isReturnOrWrappedReturn` now handles `log.Print(err); return err` pattern via `isLogOrPrintStmt` + `isLoggingMethod`.
- **Templ semantic mode**: HTML element tag names (`<a>`, `<div>`, `<button>`), attribute names (`href`, `class`, `hx-get`), and component callee names (`@demoSection(...)`) are now encoded into node Types via `syntax.EncodeSemanticType()`. Templ detection was previously purely structural — every element had the same token. Eliminated 84% of templ false positives on real projects (SwettySwipperWeb: 31→5 groups, DiscordSync: 10→5).
- **Statement-level tokenization for templ**: Each HTML element subtree is now fingerprinted as a single composite token (same mechanism as Go statements). Threshold now counts duplicated HTML _elements_, not arbitrary AST nodes. Sentinel nodes between files fix suffix-tree maximal-repeat detection.
- **Literal value normalization in semantic mode**: Semantic mode now hashes BasicLit KIND (STRING, INT, FLOAT) instead of VALUE. This enables Type-2 clone detection where only literal values differ — the most common real-world duplication pattern. Exact mode still hashes verbatim values.
- **Generic type parameter alpha-normalization**: Type parameters (`T`, `U` in generics) are now declared in the per-function symbol table. `func Map[T any]()` and `func Filter[U any]()` with the same body now match as clones.
- **Lock+Defer Unlock actionability pattern**: `m.Lock(); defer m.Unlock()` and `m.RLock(); defer m.RUnlock()` 2-statement patterns now suppressed as idiomatic Go boilerplate. Added `isAcquireMethod()` helper and `RUnlock` to cleanup methods.
- **Node.Fingerprint field**: Separate `Fingerprint int32` field on `syntax.Node` for statement-level composite hashes. `Val()` routes between `Fingerprint` (statements) and `Type` (non-statements). Fits in existing struct padding (still 64B aligned).
- **`--include-generated` flag**: Unified generated-code inclusion (`sqlc`, `templ`, `protobuf`, `mockgen`, `stringer`, `generic`, `all`). Replaces the separate `--include-sqlc`, `--include-templ`, `--include-protobuf`, `--include-mockgen`, `--include-stringer`, and `--include-generic` flags.
- **Three detection modes**: `--semantic` (default), `--exact`, `--structural` replace the former `Semantic`/`Exact` bool flags via a single `Config.DetectionMode` enum (ADR-0007).
- **Baseline CI gating**: `art-dupl baseline` records accepted clones; `art-dupl check` reports only new clones and exits 1 for CI gates.
- **Parallel incremental parsing**: `ParseIncrementalParallel` worker pool with `singleflight.Group` deduplication for byte-identical files.
- **Cache eviction**: `--max-cache-entries` flag with LRU-style `Prune` eviction.
- **HTML collapse controls**: "Collapse All" / "Expand All" toolbar buttons on clone groups.
- **SARIF rule metadata**: `precision`, `problem.severity`, and `tags` properties for GitHub Code Scanning / SonarQube.
- **Two new actionability patterns**: assign+error-check (`err := f(); if err != nil { return }`) and single-CallExpr (`errors.New("foo")`) now classified as NonActionable.
- **Performance regression tests**: `TestPerfRegressionSerialize` + `TestPerfRegressionHashSeq` with threshold gating in `flake.nix` `checks.bench`.
- **JSON config migration shim**: `Config.UnmarshalJSON` converts legacy `"semantic": false` → `"detectionMode": "exact"`.
- **`CloneRef` value object**: Shared `domain.CloneRef` (Filename, LineStart, LineEnd, Fragment) embedded across `ProcessedClone` and `pkg/artdupl.Clone` to eliminate field-name drift.
- **Generic sort comparator factory**: `sortGroupsByCriteria[T]` unifies 3 parallel 4-criteria sort implementations.
- **Astro + Starlight documentation website**: Complete public documentation site deployed to `art-dupl.lars.software` with landing page, 13 Starlight docs pages, brand theming, and Firebase hosting with security headers.
- **ADRs 0005-0008**: Split-brain type unification, non-destructive serial, detection mode enum, semantic encoding layout.
- **`encoding/json/v2` migration**: All JSON marshaling uses `encoding/json/v2` (requires `GOEXPERIMENT=jsonv2`, set in `flake.nix`).

### Changed

- **Default threshold raised from 1 to 5**: Filters trivial single-statement duplicates while catching meaningful cloning. Users can lower to 3 for more sensitivity or raise to 10+ for noise reduction. Validated across 15 real-world projects at 100% precision.
- **Templ detection now semantic-aware**: Previously purely structural (every HTML element shared the same token). Now distinguishes elements by tag name, attribute names, and component callee names.
- **Cache keys migrated from SHA-1 to SHA-256**: `CacheVersion` bumped to 2. Old caches are automatically invalidated.
- **ValueSpec/TypeSpec declarations fingerprinted as single composite tokens**: Each `var`/`const`/`type` declaration is now a single token, preventing partial expression prefix matching in declaration-only files.
- **Non-destructive AST serialization**: `serial()` shallow-copies each node before writing Type/Owns; the original tree is never mutated (ADR-0006).
- **`omitzero` instead of `omitempty`** for custom `MarshalJSON` enum types (json/v2 compatibility).
- **3 anti-idiomatic linters disabled**: `exhaustruct` (50 false positives, incompatible with Go zero-value design), `gochecknoglobals` (27 false positives, flags sync.Pool/lookup maps/test fixtures), `recvcheck` (9 false positives, string enums legitimately mix pointer/value receivers).
- **Two high-complexity functions refactored**: `evaluateActionabilityDetailed` (gocyclo 17→3, table-driven) and `applyPatternLabel` (gocyclo 17→2, map lookup).
- **Builder callback threshold lowered**: From 3 to 2 calls for more aggressive FP suppression of builder/fluent API patterns.
- **`isReturnOrWrappedReturn` extended**: Now handles 2-statement error handling bodies (`log.Print(err); return err`) in addition to single-statement returns.
- **HOW_TO_USE.md threshold table updated**: Recommendations now reflect default=5 (was project-size-based with 10-50 range).
- **11 stale docs removed**: SIMD (4 files), STATICPOOL, EXECUTION_PLAN, IMPROVEMENT_PLAN, MODERNIZATION_FINAL_REPORT, code-quality-improvements, enum-consolidation-plan, phase0-validation-safety-report.

### Fixed

- **`--min-lines` only checked first clone**: `shouldSuppressGroup` was comparing only `group.Clones[0].LineCount()` against the threshold. Fixed by extracting `minCloneLineCount()` which walks ALL clones and returns the smallest. A group is suppressed if ANY clone falls below the threshold.
- **`dumpTokensOutput` untestable**: Refactored to accept `io.Writer` instead of hardcoding `os.Stdout`, enabling unit tests.
- **gocyclo on `runCmd`**: Extracted `parseOutputFormat()` and `runStandardAnalysis()` to reduce cyclomatic complexity below threshold.
- **Lint config gaps**: Added `exhaustruct` and `gochecknoglobals` to `_test.go` exclusions, and `cobra.Command` to `exhaustruct` exclude list (30+ fields, intentional partial initialization).
- **`containsTRunCall` overmatching**: Function now verifies the receiver Ident matches common `*testing.T` variable names (t, tt, tc, test, etc.) instead of matching any `.Run` selector.
- **Cobra detection overmatching**: `isCommandLiteral` now verifies the SelectorExpr receiver Ident is "cobra" or "fang", not just any selector named "Command".
- **`assertionMethodNames` global variable**: Replaced with `isAssertionMethod()` switch function (gochecknoglobals compliance).
- **`isWrappingCall` per-call map allocation**: Replaced with `isWrappingCallName()` switch function (performance).
- **Node.Fingerprint corrupting BaseType for statement nodes**: `serial()` was overwriting `node.Type` with the fingerprint hash, making `DecodeBaseType()` return garbage for ALL statement-level matches. Every actionability pattern that checked `BaseType == golang.IfStmt` etc. was silently broken. Fixed by adding separate `Fingerprint` field (commit `930b91a`).
- **Templ callee identity loss**: `transformTemplElementExpression` created bare `ComponentRender` nodes with no name encoding. Every `@call()` in every templ file produced an identical suffix-tree token, causing false positives in demo files. Fixed by encoding callee name via `extractCalleeName()` (commit `23a3b03`).

### Removed

- **Idiom category deleted**: The `domain.CategoryIdiom` classification was fundamentally broken — the two-layer classification architecture (per-clone `ClassifyClone` then group-level `EvaluateActionabilityWithLabel`) created a contradictory state where clones were labeled "idiom — typically not actionable" but had `Actionability=Actionable` (overwritten by the group evaluator). The 14 AST-based actionability patterns already handle genuinely non-actionable clones with precision. Small clones are now classified by their actual AST type (function, struct, unknown, etc.) and the actionability patterns determine whether they're worth fixing.
- **`--since` flag removed**: Was a dead stub that was accepted and stored but never read by any analysis code. Git-diff file selection is not implemented; only content-hash caching via `--incremental` works.

### Deprecated

- **Per-generator `--include-*` flags**: `--include-sqlc`, `--include-templ`, `--include-protobuf`, `--include-mockgen`, `--include-stringer`, and `--include-generic` are now hidden aliases. They still work but print a deprecation warning pointing to `--include-generated`.

## [0.3.0] - 2026-06-12

### Added

- **CSV output with proper escaping**: Stats CSV now uses `encoding/csv` for correct field quoting and escaping
- **`--output-file` flag for stats**: Write stats output to a file instead of stdout
- **Clone actionability classification**: AST-based non-actionable pattern detection
  - Interface method implementations, test scaffolding, interface satisfaction, Go builtin patterns
  - `domain.CloneActionability` typed enum: `actionable` vs `non-actionable`
- **`domain.HealthScore` typed enum**: A-F health grade with `IsValid()`, `String()`, JSON marshaling
- **`domain.CloneCategory.IsValid()/String()`**: Validation and stringer consistent with `CloneSeverity`/`HealthScore`
- **`domain.ClonePriority.IsValid()/String()`**: Same pattern, enables type-safe priority comparisons
- **Semantic operator encoding**: Binary, unary, increment, and assignment operators hashed into semantic node type
- **Idiom category**: Clones with <5 tokens classified as `idiom` (near-zero actionability)
- **HTML printer migrated to `templ`**: Report generation via type-safe templ components instead of `fmt.Fprintf`
- **SDK documentation**: Comprehensive godoc for `pkg/artdupl/` public API
- **BDD tests for `--only` and `--include-generic`**: Stats subcommand integration coverage
- **`--simple-json` CLI flag**: Simpler JSON output format with `score=impact`

### Fixed

- **`priorityScore()` now uses `domain.ClonePriority`**: Eliminates fragile raw string comparisons
- **`TopCloneGroup` uses typed domain values**: `domain.ClonePriority` and `domain.CloneCategory` instead of `string`
- **Documentation accuracy**: Fixed incorrect "Semantic OFF by default" (it's ON), removed dead SIMD/string-interning claims
- **Semantic classification accuracy**: Operator encoding and idiom category reduce false positives
- **GoReleaser nix install**: Removed unnecessary `-r` flag from `cp` command
- **All lint issues to zero**: godoclint, gci, wsl whitespace warnings eliminated

### Changed

- **`GetCategoryEmoji` complexity 16→2**: Map lookup replaces 14-case switch
- **Clone classification decoupled from `syntax/golang`**: `printer/clone_classify.go` no longer imports AST internals
- **Printer test deduplication**: Shared `boolTestCase` runner eliminates test table duplication
- **Populated `docs/DOMAIN_LANGUAGE.md`**: 20 glossary terms, 9 value objects, 6 bounded contexts

### Removed

- **`internal/simd/` dead code**: 163 lines with 2 stale TODOs — never shipped

## [0.2.0] - 2026-05-21

### Added

- **Smart filtering**: Protobuf, mockgen, stringer, and generic (`--include-*`) generated code detection
- **`--include-generic` filter**: Catch-all for any tool following Go's `Code generated by` convention

### Fixed

- **`os.Exit(1)` test killer**: `statError` removed, `validatePaths` added
- **BDD `prepareSubcommandArgs` bug**: 9 hidden test failures from flag-value parsing
- **21 lint issues → 0**: errcheck, goconst, exhaustruct, err113, gci, golines, gocyclo
- **Nix lint sandbox**: Set `GOLANGCI_LINT_CACHE` in `flake.nix`
- **`buildJSONData` complexity 16→<10**: 6 extracted helpers
- **CI consolidation**: 3 overlapping workflows → single `ci.yml`
- **SDK version**: Hardcoded → `runtime/debug.ReadBuildInfo()`

### Removed

- **`internal/simd/` dead code** and `hashSeqSIMD()` indirection

## [0.1.0] - 2026-05-03

### Added

- **SARIF output format**: SARIF 2.1.0 for GitHub Advanced Security / CodeQL integration
- **Multi-detection methods**: Hash, suffix-tree, or combined via goroutines
- **Templ file support**: Full `.templ` analysis via pure Go parser
- **Statistics subcommand**: `art-dupl stats` with text, JSON, CSV, health grade
- **Configuration file support**: JSON-based `dupl.json` with reflection-based config merging
- **`detection.MethodDetector` interface**: Pluggable detector registry
- **Domain-driven types**: `LineNumber`, `Threshold`, `TokenCount`, `Filepath` with validation
- **Rich error types**: 11 `DuplError` types with context, wrapping, and categorization
- **BDD test suite**: Ginkgo/Gomega behavior-driven tests
- **Professional CLI**: Fang/Cobra with auto-completion, man pages, version info
- **Semantic detection ON by default**: `--structural` flag disables it
  - Methods with different receivers distinguished (e.g., `CrushMode.IsValid` vs `SafetyMode.IsValid`)
  - Extended to `FuncDecl` (receiver + name) and `TypeSpec` (type name)
- **Major file splits**: `cmd/run.go` → 5, `printer/stats.go` → 7, `domain/` → 6, `pkg/artdupl/` → 5
- **O(1) suffix tree transitions**: Map-based lookup replaces O(n) linear search

### Fixed

- **Gosec security violations**: G115 integer overflow, G301/G304/G306 file permissions
- **JSON output inconsistencies**: `detection_method` vs `detection_methods`
- **Double-counting**: Fixed in `TotalDuplicateLines`, added unique duplicate lines metric
- **`SortByTotalTokens` bug**: Incorrect sorting in `printer/sorter.go` and `printer/text.go`

## [0.0.1] - Initial Fork

### Added

- Fork from original `dupl` project
- Go AST-based structural clone detection using suffix tree algorithm (Ukkonen's)
- XXH3 streaming hash-based detection (~20x faster than SHA-256)
- Basic CLI with threshold configuration
- Text and HTML output formats
- Vendor directory exclusion
- CGO-free builds: Pure Go implementation

---

_Last updated: 2026-07-16_
