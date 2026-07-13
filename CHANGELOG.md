# Changelog

All notable changes to art-dupl will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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
- **ADRs 0005-0008**: Split-brain type unification, non-destructive serial, detection mode enum, semantic encoding layout.
- **`encoding/json/v2` migration**: All JSON marshaling uses `encoding/json/v2` (requires `GOEXPERIMENT=jsonv2`, set in `flake.nix`).

### Changed

- **Default threshold raised from 1 to 5**: Filters trivial single-statement duplicates while catching meaningful cloning. Users can lower to 3 for more sensitivity or raise to 10+ for noise reduction.
- **Cache keys migrated from SHA-1 to SHA-256**: `CacheVersion` bumped to 2. Old caches are automatically invalidated.
- **ValueSpec/TypeSpec declarations fingerprinted as single composite tokens**: Each `var`/`const`/`type` declaration is now a single token, preventing partial expression prefix matching in declaration-only files.
- **Non-destructive AST serialization**: `serial()` shallow-copies each node before writing Type/Owns; the original tree is never mutated (ADR-0006).
- **`omitzero` instead of `omitempty`** for custom `MarshalJSON` enum types (json/v2 compatibility).

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

_Last updated: 2026-07-13_
