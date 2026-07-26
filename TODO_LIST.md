# TODO List

**Last Updated:** 2026-07-26

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md` (`[Unreleased]`).
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### Correctness

- [ ] **Repo-wide audit for duplicated `errors.New("...")` sentinels**: The `ErrInvalidDetectionMode` bug had two distinct `errors.New(...)` pointers with identical messages, so `errors.Is` across packages silently returned `false`. Scan the whole codebase for other instances of this pattern and consolidate to single definitions aliased to `domain`. Source: `docs/status/2026-07-25_19-37_scanner-bug-followup.md`. Partially addressed by `domain/cross_package_alias_test.go` (verifies existing aliases), but a full audit for undiscovered duplicates is still needed.

### Code Quality

- [ ] **Split `printer/` into sub-packages**: ~29 source files / ~3500+ lines. Blocked by circular dep: core `printer.go` references `StatsPrinter`. Clean split requires moving `Printer`/`ReadFile`/`StatsPrinter` interfaces to a separate base package.

---

## MEDIUM Priority

### Filtering and Generated Code

- [ ] **Push defense-in-depth into gogenfilter**: `filterExcludedGenerated` in `cmd/util.go` patches a gap where filename-gated category filters miss files without expected patterns. The proper fix is making gogenfilter's category filters content-based as a fallback. Track: upstream issue/PR against `github.com/LarsArtmann/gogenfilter`.
- [ ] **Lazy content reading**: `shouldIncludeFile` reads content upfront for every file when includes are active, even if the filename-based filter would catch it. **Blocked**: `gogenfilter.FilterDetailed` reads content internally but doesn't return it, so avoiding the upfront read causes a double-read for regular files (~90% case). Fix requires an upstream gogenfilter API change. The marker-matching path already early-exits on files lacking the `"Code generated"` header via `bytes.Contains`.

### Detection and Filtering

- [ ] **Type-aware interface-method detection**: The `interface-method` actionability pattern (pattern #20) currently uses a static name list. The deeper, type-aware variant (using `go/types` to verify a FuncDecl actually implements an interface method) is tracked in ROADMAP.

### CI and Infrastructure

- [ ] **Local pre-commit git hook**: The auto-commit daemon re-adds `exhaustruct`/`tagliatelle` periodically. The GitHub workflow (`.github/workflows/lint-config-guard.yml`) only runs on push/PR, not on local daemon commits. A `.git/hooks/pre-commit` hook or a `.pre-commit-config.yaml` would stop the regression at commit time locally. The `disabled-linters` Nix check catches it in `nix flake check`, but the daemon bypasses CI.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
