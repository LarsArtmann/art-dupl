# TODO List

**Last Updated:** 2026-07-26

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md` (`[Unreleased]`).
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### Correctness

- [x] **Repo-wide audit for duplicated `errors.New("...")` sentinels**: COMPLETED 2026-07-26. Found 2 semantic duplicates in `pkg/artdupl/`: `ErrCloneLineEndBeforeStart` (was a distinct pointer from `domain.ErrLineEndBeforeStart`) and `ErrUnsupportedMethod` (was distinct from `domain.ErrInvalidDetectionMethod`). Both now aliased to domain. Added AST-based auto-scanner test (`TestNoDuplicateErrorNewMessages`) that scans ALL production `.go` files for duplicate `errors.New("literal")` messages — self-maintaining, no hardcoded list. Expanded `TestAliasedSentinelsAreIdentical` from 10 to 12 cases.

### Code Quality

- [x] **Extract format printers into sub-packages (Phase 4)**: EVALUATED AND DEFERRED 2026-07-26. Root `printer/` is 3,830 LOC (excluding generated `report_templ.go`). Coupling too tight for safe extraction: `json.go` alone has 25 refs to shared types (`CloneGroup`, `JSONClone`, `toJSONClone`, `SortCloneGroups`). Extraction would require a `printer/types/` package for shared infra (~1,300 LOC churn across every format file). Format set (text/json/html/sarif/plumbing) is stable — YAGNI. Re-evaluate if: root exceeds 5,000 LOC, a new format is added, or shared types stabilize.

---

## MEDIUM Priority

### Filtering and Generated Code

- [ ] **Push defense-in-depth into gogenfilter**: `filterExcludedGenerated` in `cmd/util.go` patches a gap where filename-gated category filters miss files without expected patterns. The proper fix is making gogenfilter's category filters content-based as a fallback. Track: upstream issue/PR against `github.com/LarsArtmann/gogenfilter`.
- [ ] **Lazy content reading**: `shouldIncludeFile` reads content upfront for every file when includes are active, even if the filename-based filter would catch it. **Blocked**: `gogenfilter.FilterDetailed` reads content internally but doesn't return it, so avoiding the upfront read causes a double-read for regular files (~90% case). Fix requires an upstream gogenfilter API change. The marker-matching path already early-exits on files lacking the `"Code generated"` header via `bytes.Contains`.

### Detection and Filtering

- [ ] **Type-aware interface-method detection**: The `interface-method` actionability pattern (pattern #20) currently uses a static name list. The deeper, type-aware variant (using `go/types` to verify a FuncDecl actually implements an interface method) is tracked in ROADMAP.

### CI and Infrastructure

- [ ] **Local pre-commit git hook**: The auto-commit daemon re-adds `exhaustruct`/`tagliatelle` periodically. The GitHub workflow (`.github/workflows/lint-config-guard.yml`) only runs on push/PR, not on local daemon commits. A `.git/hooks/pre-commit` hook or a `.pre-commit-config.yaml` would stop the regression at commit time locally. The `disabled-linters` Nix check catches it in `nix flake check`, but the daemon bypasses CI.
- [ ] **Wire `go-arch-lint` into `nix flake check`**: `.go-arch-lint.yml` defines the `actionability`/`stats` printer sub-package boundaries (verified enforceable via negative test), but **no Nix check actually runs `go-arch-lint`** — the boundaries are config-only, not CI-enforced. Standalone `go-arch-lint check` exits 1 due to **13 pre-existing violations** unrelated to the printer split (missing `baseline` component; undeclared deps: `cli-commands`→`baseline`/`testutil`/`syntax-golang`, `sdk`→`syntax-golang`, `printer`→`baseline`, `syntax-golang`→`domain`). Adding the CI check requires either fixing these config gaps (mostly legitimate couplings to add to `mayDependOn`) or refactoring the one smell (`cmd/run_all_modes.go` prod file imports `internal/testutil`).

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
