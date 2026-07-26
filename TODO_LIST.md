# TODO List

**Last Updated:** 2026-07-26

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md` (`[Unreleased]`).
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### Correctness

- [ ] **Repo-wide audit for duplicated `errors.New("...")` sentinels**: The `ErrInvalidDetectionMode` bug had two distinct `errors.New(...)` pointers with identical messages, so `errors.Is` across packages silently returned `false`. Scan the whole codebase for other instances of this pattern and consolidate to single definitions aliased to `domain`. Source: `docs/status/2026-07-25_19-37_scanner-bug-followup.md`.

### Code Quality

- [ ] **Split `printer/` into sub-packages**: ~29 source files / ~3500+ lines. Blocked by circular dep: core `printer.go` references `StatsPrinter`. Clean split requires moving `Printer`/`ReadFile`/`StatsPrinter` interfaces to a separate base package.

---

## MEDIUM Priority

### Filtering and Generated Code

- [ ] **Push defense-in-depth into gogenfilter**: `filterExcludedGenerated` in `cmd/util.go` patches a gap where filename-gated category filters miss files without expected patterns. The proper fix is making gogenfilter's category filters content-based as a fallback. Track: upstream issue/PR against `github.com/LarsArtmann/gogenfilter`.
- [ ] **Refactor `generatorIncludes` struct**: 6 boolean fields (SQLC, Templ, Protobuf, Mockgen, Stringer, Generic) with shotgun surgery on every new category. Consider a `map[string]struct{}` or bitfield.
- [ ] **Lazy content reading**: `shouldIncludeFile` reads content upfront for every file when includes are active, even if the filename-based filter would catch it. **Blocked**: `gogenfilter.FilterDetailed` reads content internally but doesn't return it, so avoiding the upfront read causes a double-read for regular files (~90% case). Fix requires an upstream gogenfilter API change. The marker-matching path already early-exits on files lacking the `"Code generated"` header via `bytes.Contains`.

### CLI and UX

- [ ] **YAML config file support**: `.artdupl.yml` parser alongside existing JSON support. Caveat: config enums use `MarshalJSON`/`UnmarshalJSON` hooks; YAML round-trip needs validation. Library choice: `go-faster/yaml` (not `gopkg.io/yaml.v3`).
- [ ] **`--diff-report <baseline>` mode**: Show only new/suppressed/resolved clones vs baseline. Enables the extract-verify-improve loop without manual JSON diffing.
- [ ] **HTML report improvements**: File output flag, TTY auto-detection, stable `id` attributes on clone groups for deep-linking.
- [ ] **`--recommend-threshold`**: Auto-suggest threshold based on codebase size and test-to-production ratio.
- [ ] **SARIF actionability metadata**: Emit the matched actionability pattern as `rule.tags` and clone type as a rule property. Closes the parity gap where JSON has `non_actionable_pattern` but SARIF consumers (GitHub Advanced Security) do not.

### Detection and Filtering

- [ ] **Systematic ExprStmt-wrapping audit**: `isSingleCallExpression` missed `ExprStmt`-wrapped `CallExpr` (fixed). Audit ALL 18 actionability patterns for the same wrapping gap and extract a shared `unwrapExprStmt(n)` helper. Source: `docs/status/2026-07-25_04-08_actionability-exprstmt-gap-fix.md`.
- [ ] **Templ Phase 3: expression normalization**: Normalize `{ id.String() }` vs `{ groupID.String() }` in templ expressions. Deemed low impact at threshold 5 but would improve sensitivity.
- [ ] **Interface-method-aware suppression**: Detect method signatures matching interface declarations at all thresholds, not just the current `interface-implementation` pattern. (The deeper, type-aware variant is tracked in ROADMAP.)

### Code Hygiene

- [ ] **Fix stale `Threshold: 15` in examples**: `examples/examples_sdk_demo.go:182` still hardcodes `15` instead of `pkg/artdupl.DefaultThreshold` (5). The code-hygiene sprint fixed 17 other occurrences across test fixtures but missed this one.
- [ ] **Convert remaining `switch name` predicates to `slices.Contains`**: `isCleanupMethod`, `isLoggingMethod`, `isAssertionMethod`, `isWrappingCallName` in `printer/actionability*.go` still use the switch form. Only the two art-dupl self-flagged (`isAcquireMethod`, `isTestingVarName`) were converted. Low urgency but consistency argues for converting all four.

### CI and Infrastructure

- [ ] **CI self-test gate**: Add `test -z "$(art-dupl -t 1 --plumbing .)"` as a gate so the tool's own zero-duplication invariant is enforced. Decision needed: Nix check, GitHub workflow, or both. Source: `docs/status/2026-07-25_19-37_scanner-bug-followup.md`.
- [ ] **Pre-receive/CI gate for `.golangci.yml`**: The auto-commit daemon has re-added `exhaustruct`/`tagliatelle` 7+ times. `scripts/check-disabled-linters.sh` catches it in `nix flake check`, but the daemon's commit path bypasses CI. A pre-receive hook or a GitHub workflow that rejects commits touching `.golangci.yml` with forbidden linters is the durable fix.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
