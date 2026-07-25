# TODO List

**Last Updated:** 2026-07-26

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md`.
Items here are OPEN work only; no completed, rejected, or resolved items.

---

## HIGH Priority

### Code Quality

- [ ] **Split `printer/` into sub-packages**: ~29 source files / ~3500+ lines. Blocked by circular dep: core `printer.go` references `StatsPrinter`. Clean split requires moving `Printer`/`ReadFile`/`StatsPrinter` interfaces to a separate base package.

---

## MEDIUM Priority

### Filtering and Generated Code

- [ ] **Push defense-in-depth into gogenfilter**: `filterExcludedGenerated` in `cmd/util.go` patches a gap where filename-gated category filters miss files without expected patterns (`_templ.go`, `models.go`/`querier.go`/`query.sql.go`/`batch.go`/`*.sql.go` for SQLC, `*.pb.go` for protobuf). The proper fix is making gogenfilter's category filters content-based as a fallback, or making `FilterDetailed` always run a generic content check. Track: upstream issue/PR against `github.com/LarsArtmann/gogenfilter`.
- [ ] **Refactor `generatorIncludes` struct**: 6 boolean fields (SQLC, Templ, Protobuf, Mockgen, Stringer, Generic) with shotgun surgery on every new category. Consider a `map[string]struct{}` or bitfield.
- [x] **Unify `allowsContent` and `filterExcludedGenerated`**: Both switched on the same content markers with opposite polarity. Extracted a single `matchedGeneratedCategory` helper (now the one source of truth for templ/sqlc/protobuf marker matching) plus a `categoryIncluded` reason→field mapping.
- [ ] **Lazy content reading**: `shouldIncludeFile` reads content upfront for every file when includes are active, even if filename-based filter would catch it. Read only when filename check doesn't match. **Blocked**: `gogenfilter.FilterDetailed` reads content internally but doesn't return it, so avoiding the upfront read causes a double-read for regular files (the ~90% case). The current `FilterDetailedWithContent(path, content)` approach is optimal for the common case. Fix requires an upstream gogenfilter API change (return content from FilterDetailed, or expose a filename-only pre-check). The marker-matching path itself already early-exits on files lacking the `"Code generated"` header via `bytes.Contains`, avoiding the `string(content)` copy for regular files.
- [x] **Use `bytes.Contains` instead of `string(content)`** + **Early-exit optimization**: Done together inside `matchedGeneratedCategory` and `allowsContent`. The common case (non-generated files) now returns after a single `bytes.Contains(content, []byte("Code generated"))` with no `string(content)` allocation; escape analysis confirms the constant-needle `[]byte` conversions stay on the stack.

### CLI and UX

- [ ] **YAML config file support**: `.artdupl.yml` parser alongside existing JSON support.
- [ ] **`--diff-report <baseline>` mode**: Show only new/suppressed/resolved clones vs baseline. Enables the extract-verify-improve loop without manual JSON diffing.
- [x] **`--explain` flag**: Implemented in `printer/text.go::writeExplanation`. Adds an explanation line after each clone group header showing clone type, actionability (+ pattern label via new `NonActionablePattern` field on `CloneClassification`), category, token/line counts, and extractability estimate. The `ExplainSetter` interface is wired in `cmd/run_flags.go`. BDD test in `bdd/actionability_test.go`.
- [ ] **HTML report improvements**: File output flag, TTY auto-detection, stable `id` attributes on clone groups for deep-linking.
- [ ] **`--recommend-threshold`**: Auto-suggest threshold based on codebase size and test-to-production ratio.

### Detection and Filtering

- [ ] **Templ Phase 3: expression normalization**: Normalize `{ id.String() }` vs `{ groupID.String() }` in templ expressions. Deemed low impact at threshold 5 but would improve sensitivity.
- [ ] **Interface-method-aware suppression**: Detect method signatures matching interface declarations at all thresholds, not just the current `interface-implementation` pattern.

### Code Hygiene

- [x] **`SetFilterSourceStats` unit test**: Added `cmd/filter_stats_test.go` with dedicated coverage for `SourceBreakdown()`, `RecordWithSource` source attribution, defensive-copy guarantees, and nil-receiver safety (the latter uncovered and fixed a real nil-panic bug in `Breakdown`/`SourceBreakdown` where `s.byReason`/`s.bySource` were evaluated as args before the nil-checked `copyMapUnderLock` ran — refactored to a closure pattern matching `withReadLock`).
- [x] **`--no-actionability` flag**: Implemented. Gates actionability filtering via `if semantic && !suppression.NoActionability` in `cmd/run_output.go`. `SuppressionConfig.NoActionability` wired from `config.Config.NoActionability`. BDD tests in `bdd/actionability_test.go` verify guard-clause clones appear with the flag.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
