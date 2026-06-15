# Naming Review Report — 2026-06-15

## Executive Summary

- 229 Go files reviewed (48k lines)
- **0 Critical** issues (no lying names, no hidden mutations)
- **1 High** issue (fixed: `cnt` abbreviation → `count`)
- **3 Medium** issues (split-brain Clone type names, `priorityData` vagueness)
- **2 Low** issues (Data suffixes on view models — acceptable pattern)

## Fixed This Session

| #   | File             | Line | Identifier | Issue                       | Fix     |
| --- | ---------------- | ---- | ---------- | --------------------------- | ------- |
| 1   | syntax/syntax.go | 228+ | `cnt`      | Abbreviation in 4 functions | `count` |

## Domain Alignment Issues

### Clone Type Proliferation (Split-Brain Risk)

The codebase has many Clone-related types serving different layers. The names are mostly domain-aligned but the sheer count creates confusion:

| Type                          | Package | Purpose              |
| ----------------------------- | ------- | -------------------- |
| `domain.ProcessedClone`       | domain  | Canonical clone type |
| `domain.ProcessedCloneGroup`  | domain  | Canonical group type |
| `printer.CloneGroup`          | printer | JSON output          |
| `printer.JSONClone`           | printer | JSON output          |
| `printer.SimpleJSONClone`     | printer | Simple JSON output   |
| `printer.SimpleCloneGroup`    | printer | Simple JSON output   |
| `printer.CloneGroupDiff`      | printer | Diff rendering       |
| `printer.CloneDiff`           | printer | Diff rendering       |
| `printer.CloneWithContent`    | printer | Diff rendering       |
| `printer.CloneOccurrenceData` | printer | HTML view model      |
| `printer.CloneGroupViewData`  | printer | HTML view model      |
| `printer.TopCloneGroup`       | printer | Stats                |
| `pkg/artdupl.CloneGroup`      | SDK     | Public API           |

**Key risk:** `printer.CloneGroup` and `pkg/artdupl.CloneGroup` share a name but are different types. This is tracked in TODO_LIST.md as the "Four parallel Clone types" consolidation.

## Medium Issues

| #   | File                      | Identifier     | Issue                           | Suggestion            |
| --- | ------------------------- | -------------- | ------------------------------- | --------------------- |
| 1   | domain/processed_clone.go | `priorityData` | Vague "Data" suffix             | `priorityWeights`     |
| 2   | printer/html_views.go     | `*ViewData`    | Multiple types with Data suffix | Consider `*ViewModel` |
| 3   | printer/file_processor.go | `FileInfo`     | Borderline generic              | Acceptable in context |

## Strengths (Excellent Naming)

- **Domain types**: `CloneSeverity`, `CloneCategory`, `CloneActionability`, `ClonePriority` — all domain-aligned, strong types
- **Predicate functions**: All `is*` functions in `actionability.go` correctly return bool and are precisely named
- **No Impl/Base/Abstract prefixes**: Clean architecture naming
- **No I-prefix interfaces**: Follows Go conventions
- **Config enums**: Strong typed enums with `Parse*` and `IsValid()` methods
- **Error sentinels**: Well-named, consistent `Err*` prefix pattern
- **No euphemisms or lying names**: All function names accurately describe behavior

## Consistency

- **No synonym drift**: Consistent use of "clone" (not "duplicate/copy/repeat") throughout
- **Consistent verb tense**: Detection methods consistently named `Find*`, processors consistently `Process*`
- **No cutesy names**: Professional naming throughout
