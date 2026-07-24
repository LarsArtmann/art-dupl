# TODO List

**Last Updated:** 2026-07-25

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md`.
Items here are OPEN work only; no completed, rejected, or resolved items.

---

## HIGH Priority

### Code Quality

- [ ] **Split `printer/` into sub-packages**: ~29 source files / ~3500+ lines. Blocked by circular dep: core `printer.go` references `StatsPrinter`. Clean split requires moving `Printer`/`ReadFile`/`StatsPrinter` interfaces to a separate base package.

---

## MEDIUM Priority

### Filtering and Generated Code

- [ ] **Push defense-in-depth into gogenfilter**: `filterExcludedGenerated` in `cmd/util.go` patches a gap where filename-gated category filters miss files without expected suffixes (`_templ.go`, `_sqlc.go`, `*.pb.go`). The proper fix is making gogenfilter's category filters content-based as a fallback, or making `FilterDetailed` always run a generic content check. Track: upstream issue/PR against `github.com/LarsArtmann/gogenfilter`.
- [ ] **Refactor `generatorIncludes` struct**: 6 boolean fields (SQLC, Templ, Protobuf, Mockgen, Stringer, Generic) with shotgun surgery on every new category. Consider a `map[string]struct{}` or bitfield.
- [ ] **Unify `allowsContent` and `filterExcludedGenerated`**: Both switch on the same content markers with opposite polarity. Extract a single `categorizeContent` function.
- [ ] **Lazy content reading**: `shouldIncludeFile` reads content upfront for every file when includes are active, even if filename-based filter would catch it. Read only when filename check doesn't match.
- [ ] **Use `bytes.Contains` instead of `string(content)`**: Avoid heap allocation in `filterExcludedGenerated` and `allowsContent` for large files.
- [ ] **Early-exit optimization**: If file content doesn't contain "Code generated", skip all marker checks in `filterExcludedGenerated`.

### CLI and UX

- [ ] **YAML config file support**: `.artdupl.yml` parser alongside existing JSON support.
- [ ] **`--diff-report <baseline>` mode**: Show only new/suppressed/resolved clones vs baseline. Enables the extract-verify-improve loop without manual JSON diffing.
- [ ] **`--explain` flag**: Explain WHY a clone was reported (which detection method, what pattern matched, why it's actionable). Aids triage.
- [ ] **HTML report improvements**: File output flag, TTY auto-detection, stable `id` attributes on clone groups for deep-linking.
- [ ] **`--recommend-threshold`**: Auto-suggest threshold based on codebase size and test-to-production ratio.

### Detection and Filtering

- [ ] **Templ Phase 3: expression normalization**: Normalize `{ id.String() }` vs `{ groupID.String() }` in templ expressions. Deemed low impact at threshold 5 but would improve sensitivity.
- [ ] **Interface-method-aware suppression**: Detect method signatures matching interface declarations at all thresholds, not just the current `interface-implementation` pattern.

### Code Hygiene

_All Code Hygiene items completed — see `CHANGELOG.md`._

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
