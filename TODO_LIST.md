# TODO List

**Last Updated:** 2026-08-05

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md` (`[Unreleased]`).
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### Property Engine Follow-up

- [ ] **Calibrate confidence values against real-world data**: The property engine uses hardcoded thresholds (0.8 actionable, 0.5 low-confidence, 60% helper-dominance ratio). Run on 3-5 OSS Go projects, measure FP/FN rates, and tune thresholds empirically.
- [ ] **Add property engine labels to `--list-patterns`**: The 4 property labels (`property-engine`, `property-control-flow`, `property-roi`, `property-parameterizable`) are invisible to `AllActionabilityPatterns()`. Either register them or document as internal-only.

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

- [x] **Type-aware interface-method detection**: COMPLETED 2026-08-05. The `interface-method` actionability pattern now uses `go/types` to verify a FuncDecl satisfies a same-package interface when `--type-aware` is active. The `golang.IsInterfaceMethod()` function scans the package scope for interfaces with a matching method name and checks `types.Implements` for both value and pointer receivers. The `InterfaceMethod` flag is propagated from FuncDecl to body statement nodes via the transformer (same save/restore pattern as `EnclosingReturnArity`) because FuncDecl nodes are never clone roots in Go files (structural filter in `FindSyntaxUnits` rejects non-Statement roots). The pattern matcher has two paths: path 1 checks FuncDecl roots (edge case, uses static name list), path 2 checks statement-level roots (normal Go files, uses propagated flag). The static stdlib name list (`commonInterfaceMethodNames`) remains as a fallback. Also fixed a pre-existing bug where FuncDecl nodes did not have `Name` set by the transformer. Cross-package interface scanning remains future work (ROADMAP).

### CI and Infrastructure

- [x] **Local pre-commit git hook**: COMPLETED 2026-07-26. The `scripts/check-disabled-linters.sh` guard is now appended to `.git/hooks/pre-commit` after the buildflow step via `scripts/install-hooks.sh` (idempotent, re-run if buildflow overwrites the hook). The guard fails the commit if `exhaustruct` or `tagliatelle` are re-added to `.golangci.yml`.
- [ ] **Wire `go-arch-lint` into `nix flake check`**: `.go-arch-lint.yml` defines the `actionability`/`stats` printer sub-package boundaries (verified enforceable via negative test), but **no Nix check actually runs `go-arch-lint`** — the boundaries are config-only, not CI-enforced. Standalone `go-arch-lint check` exits 1 due to **13 pre-existing violations** unrelated to the printer split (missing `baseline` component; undeclared deps: `cli-commands`→`baseline`/`testutil`/`syntax-golang`, `sdk`→`syntax-golang`, `printer`→`baseline`, `syntax-golang`→`domain`). Adding the CI check requires either fixing these config gaps (mostly legitimate couplings to add to `mayDependOn`) or refactoring the one smell (`cmd/run_all_modes.go` prod file imports `internal/testutil`).

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
