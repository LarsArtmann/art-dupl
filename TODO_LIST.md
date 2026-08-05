# TODO List

**Last Updated:** 2026-08-05

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md` (`[Unreleased]`).
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### Architecture

- [ ] **Inject output writers into production code**: The root cause of the `cmd` test race condition (2026-08-02). 23+ direct `fmt.Fprintf(os.Stderr, ...)` calls and 2 `fmt.Printf` calls in `cmd/` bypass cobra's injectable writer system. Route through `cmd.OutOrStdout()`/`cmd.ErrOrStderr()` or accept `io.Writer` params. Eliminates the `CaptureStdoutStderr` global-state race surface and makes all tests safely parallelizable. Source: `docs/status/2026-08-02_01-05_race-condition-fix-cmd-tests.md` section E1.

### Property Engine Follow-up

- [ ] **Calibrate confidence values against real-world data**: The property engine uses hardcoded thresholds (0.8 actionable, 0.5 low-confidence, 60% helper-dominance ratio). Run on 3-5 OSS Go projects, measure FP/FN rates, and tune thresholds empirically.
- [ ] **Add property engine labels to `--list-patterns`**: The 4 property labels (`property-engine`, `property-control-flow`, `property-roi`, `property-parameterizable`) are invisible to `AllActionabilityPatterns()`. Either register them or document as internal-only.

### Correctness

- [ ] **Add regression test for lazy-read nil-content invariant**: `FilterDetailedAndContent` returns `nil` content when phase-1 filename detection catches a file. This is the core performance guarantee of the v3.4.0 migration but has no test that directly asserts it. If someone reverts to `FilterDetailedWithContent`, no test catches the regression.

### Code Quality

- [ ] **Remove `tagliatelle` from `.golangci.yml` enable list (recurring)**: The auto-committer keeps re-adding `tagliatelle` (and possibly `exhaustruct`) to `.golangci.yml:108`. The project's own guard script (`scripts/check-disabled-linters.sh`) bans it and fails `nix flake check`. This is a recurring whack-a-mole — investigate the root cause (why does the auto-committer re-add it?) and consider making the guard script auto-fix instead of just failing.

---

## MEDIUM Priority

### Testing

- [ ] **Add `FuzzFindDuplOverParallel` fuzz test**: The parallel suffix tree search path has correctness and property tests but no randomized fuzz test. Mirror the existing `FuzzFindDuplOver` to ensure no panic and channel-always-closes on arbitrary input.
- [ ] **Parameterize property tests for parallel search**: The property tests (`TestProperty_*`) only test sequential `FindDuplOver`. Extend them to also run against `FindDuplOverParallel` to guarantee the mathematical invariants hold for both paths.
- [ ] **Add BDD test exercising `--search-workers`**: No end-to-end test covers parallel search. Add a scenario to `bdd/` that runs analysis with `--search-workers 4` and verifies identical results to sequential.
- [ ] **Add dedicated FuncLit flag-reset test**: The transformer resets `enclosingInterfaceMethod = false` in the FuncLit case (closures are not interface methods), but no test verifies closures inside interface method bodies don't inherit the flag.
- [ ] **Add SDK test for `InterfaceMethod` flag**: `pkg/artdupl` has `detector_type_aware_test.go` but no test verifies `InterfaceMethod` flows through the SDK boundary to `pkg/artdupl.Clone`.
- [ ] **Add extractability engine integration test**: `isFormatSpecifierDifference` unit tests call the function directly. No integration test exercises `EvaluateExtractability` end-to-end with format-specifier-differing `CloneNode` trees.

### Documentation

- [ ] **Document `--search-workers` in HOW_TO_USE.md**: The flag is registered, functional, and wired through CLI → config → detection → SDK, but not documented in the user guide.

### CI and Infrastructure

- [ ] **Wire `go-arch-lint` into `nix flake check`**: `.go-arch-lint.yml` defines the `actionability`/`stats` printer sub-package boundaries (verified enforceable via negative test), but **no Nix check actually runs `go-arch-lint`** — the boundaries are config-only, not CI-enforced. Standalone `go-arch-lint check` exits 1 due to **13 pre-existing violations** unrelated to the printer split (missing `baseline` component; undeclared deps: `cli-commands`→`baseline`/`testutil`/`syntax-golang`, `sdk`→`syntax-golang`, `printer`→`baseline`, `syntax-golang`→`domain`). Adding the CI check requires either fixing these config gaps (mostly legitimate couplings to add to `mayDependOn`) or refactoring the one smell (`cmd/run_all_modes.go` prod file imports `internal/testutil`).

### Code Cleanup

- [ ] **Remove stale "defense-in-depth" comments**: 4 references in `bdd/filter_features_test.go` (lines 461, 477, 502, 535) and 2 in production code (`printer/stats_data.go:87`, `printer/stats/stats_collector.go:42`) describe a mechanism removed in the gogenfilter v3.4.0 migration. The tests pass but the comments are misleading.
- [ ] **Remove `go.mod` local replace directive**: `replace github.com/LarsArtmann/gogenfilter/v3 => /home/lars/projects/gogenfilter` works for Nix (which overrides it) and local dev, but breaks for anyone cloning the repo without the local path. Should point to the tagged `v3.4.0` version.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
