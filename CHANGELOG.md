# Changelog

All notable changes to art-dupl will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Templ file support**: Full analysis of `.templ` template files (templ.guide) using tree-sitter integration with CGO bindings
- **Multi-language architecture**: Language-agnostic AST representation supporting both Go and Templ
- **Comprehensive BDD test suite**: Ginkgo/Gomega behavior-driven tests for CLI commands, configuration, filtering, and stats
- **Smart filtering**: Automatic detection and filtering of SQLC and Templ generated code
- **Configuration file support**: JSON-based configuration for team consistency (`dupl.json`)
- **Multiple output formats**: Text, HTML, JSON, and plumbing formats
- **Statistics subcommand**: Aggregated duplication metrics with multiple output formats (`art-dupl stats`)
- **Sorting options**: By size, occurrence, or hash
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

- **CGO required**: Build with `CGO_ENABLED=1` for tree-sitter integration
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
