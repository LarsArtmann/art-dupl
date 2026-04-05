# Changelog

All notable changes to art-dupl will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### 2026-03-28

### Added

- **SARIF output format**: Security tool integration with SARIF 2.1.0
  - SARIFOutput, SARIFRun, SARIFTool, SARIFResult structures
  - Proper level mapping (error/warning/note based on clone size)
  - Duplicate hash filtering to avoid redundant results
  - Invocation timing information
  - `--sarif` flag for CLI access

### Fixed

- **Gosec security violations**: G115 integer overflow, G301/G304/G306 file permissions
- **Cyclomatic complexity**: Fixed in critical functions
- **Cognitive complexity**: Fixed in critical functions (gocognit)
- **Funlen lint issue**: Split buildSuffixTree into smaller functions
- **JSON output inconsistencies**: detection_method vs detection_methods
- **Double-counting**: Fixed in TotalDuplicateLines, added unique duplicate lines metric
- **Migration tests**: 0% → 83.1% coverage

### Changed

- **Semantic detection now OFF by default**: Restores backward compatibility with structural-only matching
  - Use `--semantic` flag to enable semantic-aware detection (matches by identifier names)
  - Methods with different receiver types are distinguished when semantic is enabled (e.g., `CrushMode.IsValid` vs `SafetyMode.IsValid`)
  - Extended semantic hashing to `FuncDecl` (receiver type + method name) and `TypeSpec` (type name)
  - `--structural` flag deprecated (now the default behavior)

### Technical

- **Multi-detection methods**: hash, art-dupl, or combined execution
- **internal/testutil package**: Shared BDD helpers for Ginkgo/Gomega tests
- **Lint issues**: 60+ → 0 resolved

### Changed

### Added

- **Templ file support**: Full analysis of `.templ` template files (templ.guide) using pure Go parser (github.com-a-h-templ/parser/v2)
- **Multi-language architecture**: Language-agnostic AST representation supporting both Go and Templ
- **Comprehensive BDD test suite**: Ginkgo/Gomega behavior-driven tests for CLI commands, configuration, filtering, and stats
- **Smart filtering**: Automatic detection and filtering of SQLC and Templ generated code
- **Configuration file support**: JSON-based configuration for team consistency (`dupl.json`)
- **Multiple output formats**: Text, HTML, JSON, and plumbing formats
- **Statistics subcommand**: Aggregated duplication metrics with multiple output formats (`art-dupl stats`)
- **Semantic-aware detection**: Content-aware matching with `--semantic` flag to opt-in
  - FuncDecl: Combines receiver type name + function name into node hash
  - TypeSpec: Encodes type name into node hash
  - Ident/SelectorExpr: Already supported (identifier references, method calls)
- **Professional CLI**: Built with Fang/Cobra framework with auto-completion and version info
- **Custom pattern filtering**: Include/exclude patterns for fine-grained file control
- **Streaming detection**: Non-blocking results for large projects via channels

### Changed

- **Major refactoring**: Split large files into focused modules
  - `cmd/run.go` → 5 modules (analysis, crawl, flags, output, all_modes)
  - `printer/stats.go` → 7 modules (collector, formatter, health, recommendations, styles, visualization)
  - `domain/domain_types.go` → 6 files (file, id, metadata, metric, validation, helpers)
  - `pkg/artdupl/detector.go` → 5 modules (conversion, pipeline, utils, validation)
- **Improved stats accuracy**: Better line counting and simplified plumbing output
- **Domain-driven types**: Strong type safety with `LineNumber`, `Threshold`, `TokenCount`, etc.
- **Error handling**: Rich `DuplError` types with context, wrapping, and categorization

### Technical

- **CGO-free builds**: No C dependencies required, pure Go implementation
- **SIMD optimizations**: Vectorized transition search for improved performance
- **Memory efficiency**: String interning pool and optimized node struct layout
- **Test coverage**: Comprehensive unit, integration, and BDD tests

## [1.0.0] - Initial Fork

### Added

- Fork from original `dupl` project
- Go AST-based structural clone detection using suffix tree algorithm
- Basic CLI with threshold configuration
- Text and HTML output formats
- Vendor directory exclusion

---

## Feature Overview

| Feature                        | Status           |
| ------------------------------ | ---------------- |
| Go file analysis               | ✅ Full          |
| .templ file analysis           | ✅ Full          |
| Hash-based detection           | ✅ Full          |
| Suffix-tree detection          | ✅ Full          |
| Multi-method detection         | ✅ Full          |
| Text output                    | ✅ Full          |
| HTML output                    | ✅ Full          |
| JSON output                    | ✅ Full          |
| Plumbing output                | ✅ Full          |
| Stats subcommand               | ✅ Full          |
| Configuration files            | ✅ Full          |
| Smart filtering (SQLC/templ)   | ✅ Full          |
| Sorting (size/occurrence/hash) | ✅ Full          |
| BDD test suite                 | ✅ Comprehensive |

## Supported Languages

| Language | Extension | Support Level |
| -------- | --------- | ------------- |
| Go       | `.go`     | Full analysis |
| Templ    | `.templ`  | Full analysis |

---

_This changelog was generated based on git commit history._
