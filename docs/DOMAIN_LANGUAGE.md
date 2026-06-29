# Domain Language

A **Unified Language** for art-dupl — shared across Developer, Contributor, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.

## Glossary

| Term             | Definition                                                                       | Context                           |
| ---------------- | -------------------------------------------------------------------------------- | --------------------------------- |
| Clone            | A detected duplicate code fragment, defined by file and line range               | Core detection output             |
| Clone Group      | Two or more Clones sharing the same structural or semantic pattern               | All output formats                |
| Token            | A single AST node serialized into an integer type for matching                   | Suffix tree / hash input          |
| Threshold        | Minimum number of consecutive Tokens to consider a Clone (default: 15)           | CLI `--threshold` flag            |
| Semantic Mode    | Matching by AST structure + identifier/operator names (default ON)               | Detection method selection        |
| Structural Mode  | Matching by AST shape only, ignoring identifier/operator names                   | `--structural` flag               |
| Detection Method | Algorithm used to find Clones: suffix tree, hash, or both                        | `--method` / `-m` flag            |
| Suffix Tree      | Ukkonen's algorithm on serialized AST tokens, finds repeated substrings          | Core algorithm                    |
| Hash Detection   | Rolling XXH3 hash on AST token sequences, content-addressed dedup                | Alternative detection method      |
| Multi-Detection  | Running both suffix tree and hash detection in parallel                          | `-m "hash,art-dupl"`              |
| Category         | Classification of clone code type: function, method, struct, etc.                | `domain.CloneCategory`            |
| Priority         | How important a clone is to address: critical, high, medium, low                 | `domain.ClonePriority`            |
| Actionability    | Whether a clone can realistically be deduplicated                                | `domain.CloneActionability`       |
| Health Score     | A-F grade for codebase duplication health                                        | `domain.HealthScore`              |
| Severity         | Impact level of a clone: low, medium, high, critical                             | `domain.CloneSeverity`            |
| Idiom            | Clone with <5 tokens AND ≤5 lines — structural artifact, near-zero actionability | Classification category           |
| Non-Actionable   | Clone that follows idiomatic Go patterns, not worth deduplicating                | Interface impls, test scaffolding |
| Smart Filtering  | Automatic exclusion of generated code (sqlc, templ, protobuf, etc.)              | File selection pipeline           |
| Output Format    | Presentation mode: text, HTML, JSON, CSV, plumbing, SARIF, simple-json           | `--format` flag                   |
| Stats            | Aggregated duplication metrics subcommand                                        | `art-dupl stats`                  |
| ProcessedClone   | Decoupled DTO representing a clone fragment for printer output                   | `domain.ProcessedClone`           |
| Incremental Mode | Only analyze files changed since a reference point                               | `--incremental`, `--since`        |

## Entities

Objects with identity and lifecycle.

| Term        | Definition                                                      | Context                  |
| ----------- | --------------------------------------------------------------- | ------------------------ |
| AST         | Abstract Syntax Tree — parsed representation of Go/Templ source | Suffix tree / hash input |
| Suffix Tree | Data structure built from serialized ASTs for clone search      | Core algorithm           |
| Clone Group | A cluster of duplicate fragments sharing a pattern              | Primary detection output |

## Value Objects

Immutable objects defined by attributes.

| Term               | Definition                                          | Context                     |
| ------------------ | --------------------------------------------------- | --------------------------- |
| Filepath           | Validated file path string                          | `domain.Filepath`           |
| LineNumber         | Validated 1-based line number (uint16)              | `domain.LineNumber`         |
| CloneCategory      | Category enum: function, method, test, struct, etc. | `domain.CloneCategory`      |
| ClonePriority      | Priority enum: critical, high, medium, low          | `domain.ClonePriority`      |
| CloneActionability | Actionability enum: actionable, non-actionable      | `domain.CloneActionability` |
| CloneSeverity      | Severity enum: low, medium, high, critical          | `domain.CloneSeverity`      |
| HealthScore        | Health grade enum: A, B, C, D, F                    | `domain.HealthScore`        |
| Threshold          | Minimum token sequence size for clone detection     | `config.DefaultThreshold`   |
| TokenCount         | Number of AST tokens in a clone fragment            | Stats / output              |

## Events

Things that happen in the domain.

| Term             | Definition                                    | Context                 |
| ---------------- | --------------------------------------------- | ----------------------- |
| CloneDetected    | A clone group was found by a detection method | Detection pipeline      |
| AnalysisComplete | All files parsed, all detections finished     | Stats / output pipeline |
| FileFiltered     | A file was excluded by smart filtering        | File selection pipeline |

## Commands

Actions the system can perform.

| Term            | Definition                                        | Context               |
| --------------- | ------------------------------------------------- | --------------------- |
| DetectClones    | Run detection on a set of files                   | CLI `art-dupl`        |
| ShowStats       | Display aggregated duplication statistics         | CLI `art-dupl stats`  |
| ClassifyClone   | Assign category, priority, actionability to clone | Printer pipeline      |
| FilterFiles     | Apply smart filtering to input file list          | Job pipeline          |
| BuildSuffixTree | Construct suffix tree from serialized ASTs        | Suffix tree algorithm |

## Bounded Contexts

Subsystems with distinct vocabulary.

| Context        | Description                                                         |
| -------------- | ------------------------------------------------------------------- |
| Detection      | Finding duplicate code via suffix tree, hash, or multi-method       |
| Classification | Assigning category, priority, and actionability to clones           |
| Output         | Formatting and presenting results in text, HTML, JSON, etc.         |
| Configuration  | Merging defaults, config files, and CLI flags                       |
| Filtering      | Selecting which files to analyze (smart filtering, include/exclude) |
| Stats          | Aggregating metrics and computing health scores                     |
| SDK            | Public programmatic API for embedding art-dupl in other tools       |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
