# Status Report — Printer Package Decoupling Sprint

> **Date:** 2026-07-26 17:20
> **Session scope:** Execute the 3-phase printer package decoupling plan from `docs/planning/2026-07-26_16-31_PRINTER-PACKAGE-DECOUPLING.md`
> **Quality gate:** `nix flake check` passed all 11 checks at the end of the session

---

## a) FULLY DONE

### Phase 1 — Extract `printer/actionability/` (the 1% → 51%)

- **17 files moved** via `git mv` (8 prod + 9 test files) from `printer/` to `printer/actionability/`
- **Package declarations changed** on all 17 files: `package printer` → `package actionability`
- **Exported `applyPatternLabel`** → `ApplyPatternLabel` (was unexported, called from root's `clone_processor.go`)
- **`printer/clone_processor.go` rewired**: imports `printer/actionability`, qualifies all calls (`actionability.EvaluateActionabilityWithLabel`, `actionability.ClassifyClone`, `actionability.ApplyPatternLabel`, `actionability.PatternNone`)
- **`cmd/run_output.go` rewired**: `actionability.EvaluateActionabilityWithDisabled`, `actionability.PatternLabel` on `SuppressionConfig.DisabledPatterns`
- **`cmd/run_flags.go` rewired**: `actionability.ListActionabilityPatterns`, `actionability.PatternLabel` in `buildDisabledPatternSet`
- **`cmd/diff_report.go` rewired**: `actionability.EvaluateActionabilityWithDisabled`
- **`printer/semantic_precision_test.go` rewired**: `actionability.EvaluateActionability`
- **Type aliases kept in root** (`CloneCategory`, `ClonePriority`, `CloneClassification` in `clone_processor.go`) — needed by `html*.go` format printers
- **`TestPriorityHigher` test split**: moved to `printer/priority_higher_test.go` (it tests root's `priorityHigher` from `html.go`)
- **Fixed pre-existing `exhaustive` lint error** in `getSuggestion` switch (missing cases added)
- **Result**: 1,670 prod LOC extracted into a clean leaf package (depends only on domain + syntax)

### Phase 2 — Extract `printer/stats/` (completes the 4% → 64%)

- **12 files moved** via `git mv` (7 prod + 5 test files) from `printer/` to `printer/stats/`
- **3 golden test files moved** to `printer/stats/testdata/` (`TestTextOutputGolden.golden`, `TestStatsCSVOutputGolden.golden`, `TestStatsJSONOutputGolden.golden`)
- **Package declarations changed**: `package printer` → `package stats`
- **Root printer types qualified** in all stats prod files: `printer.Printer`, `printer.ReadFile`, `printer.StatsConfig`, `printer.StatsView`, `printer.TopCloneGroup`
- **`stats_data.go` KEPT IN ROOT** — `StatsView`/`TopCloneGroup` are referenced by root interfaces (`GetStatsView() *StatsView` in `printer.go`). Moving them would create a circular dependency.
- **`cmd/run_printer.go` rewired**: `stats.NewStats` instead of `printer.NewStats`
- **`cmd/stats.go` rewired**: `stats.NewStats` instead of `printer.NewStats`
- **Local test helpers created** in `printer/stats/helpers_test.go` (`mockReadFile`, `processTestNodes`, `printTestClones`, `testFileContent`)
- **Health constants duplicated** in `printer/groups_test.go` (`healthSmall`, `healthMedium`, `healthLarge`, `healthHuge`) — they moved to `stats_health.go` in the stats package
- **Result**: 1,303 prod LOC extracted. Stats imports root (for interface satisfaction); root does NOT import stats.

### Phase 3 — Make boundaries durable (completes the 20% → 80%)

- **`.go-arch-lint.yml` updated**:
  - Added `actionability` component (`in: printer/actionability/**`) with deps: `domain, syntax, syntax-golang, errors, pkg-utils`
  - Added `stats` component (`in: printer/stats/**`) with deps: `domain, config, errors, pkg-utils, printer`
  - Updated `printer` deps to ADD `actionability` (root now imports the sub-package)
  - Updated `cli-commands` deps to ADD `actionability, stats`
  - These rules enforce that actionability is a leaf (cannot depend on printer/stats/formats) and stats cannot depend on actionability
- **`AGENTS.md` updated**: Architecture tree now shows `printer/`, `printer/actionability/`, `printer/stats/` as separate lines. File references updated (`printer/actionability*.go` → `printer/actionability/`, `printer/clone_classify.go` → `printer/actionability/clone_classify.go`).
- **`CHANGELOG.md` updated**: Added entry under `[Unreleased] > Changed` describing the split.
- **`TODO_LIST.md` updated**: Replaced the old "blocked by circular dep" item with a Phase 4 deferred item.

### Guardrail: Linter regression fixed (temporarily)

- Removed `exhaustruct` and `tagliatelle` from `.golangci.yml` (re-added by the daemon between sessions)
- `scripts/check-disabled-linters.sh` passed at the time

### Verification

- `GOEXPERIMENT=jsonv2 go build ./...` — passes
- `GOEXPERIMENT=jsonv2 go test -count=1 ./...` — all 26 packages pass
- `nix flake check` — all 11 checks pass (build, test, race, lint, fmt, treefmt, self-test, arch-lint, bench, SARIF validate, disabled-linters)
- `pkg/artdupl` SDK — zero `printer/` imports (verified)

---

## b) PARTIALLY DONE

### Documentation updates — 80% done

- `AGENTS.md` architecture tree updated ✅
- `AGENTS.md` file references updated ✅
- `CHANGELOG.md` entry added ✅
- `TODO_LIST.md` updated ✅ but **uncommitted** (working tree dirty — daemon may or may not commit it)
- **Plan doc `Status:` line NOT updated** — still says `PLANNING — awaiting execution`. Should say `DONE — executed 2026-07-26`.
- **Plan doc success criteria checkboxes NOT checked** — all 8 boxes still `[ ]`

### Arch-lint enforcement — 70% done

- Component definitions added ✅
- Dependency rules added ✅
- **But**: The arch-lint check only ran as part of `nix flake check` — I did NOT run the arch-lint standalone or inspect its output in detail to confirm the new rules actually catch violations (e.g., if someone in actionability imports printer)

---

## c) NOT STARTED

### Phase 4 — Format printer extraction (DEFERRED)

- text/json/html/sarif/plumbing → sub-packages
- Documented as deferred in `TODO_LIST.md`
- Not part of this sprint per the plan

### Test helper deduplication

- `mockReadFile` is now duplicated in 3 places: `printer/json_test.go`, `printer/stats/helpers_test.go`, (original was shared)
- `testFileContent` constant duplicated in 2 places: `printer/html_golden_test.go`, `printer/stats/helpers_test.go`
- `healthSmall`/`healthMedium`/`healthLarge`/`healthHuge` constants duplicated: `printer/groups_test.go` + `printer/stats/stats_health.go`
- Should be consolidated into a shared test utility, but this is post-sprint cleanup

---

## d) TOTALLY FUCKED UP

### The daemon re-added `exhaustruct` and `tagliatelle` AGAIN

At the end of the session, `.golangci.yml` is **dirty** — the daemon has re-added both disabled linters back into the enable list + the `exhaustruct` settings block. The `.golangci.yml` diff is +400/-395 lines (the daemon rewrote the entire file). `scripts/check-disabled-linters.sh` **FAILS** right now.

The last `nix flake check` passed because it built from the **committed** state (commit `7e4e621b`), which had the clean `.golangci.yml`. But the working tree is dirty, and the next daemon commit will re-introduce the linters.

**This is the #1 issue to fix immediately.**

### Messy git history — 13 daemon commits in 1 hour

The daemon committed 13 times during this session, often mid-edit. The commit messages are generic and sometimes wrong (e.g., `feat(actionability): enhance clone classification with semantic precision` for what was a mechanical file move). The history is not atomic — Phase 1, 2, and 3 changes are spread across interleaved daemon commits rather than clean per-phase commits.

### `TODO_LIST.md` change is uncommitted

My edit to `TODO_LIST.md` (replacing the old "blocked by circular dep" item with the Phase 4 deferred item) is sitting in the working tree, uncommitted. The daemon may pick it up or it may get lost.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Commit the guardrail FIRST and separately** — I removed `exhaustruct`/`tagliatelle` at the start but didn't commit it as an isolated change. The daemon re-added them mid-session. Should have committed immediately.
2. **Use `git mv` for all file moves** — I did this correctly, but the daemon's auto-commits sometimes used regular file operations, potentially losing history.
3. **Update the plan doc status immediately** — The plan doc still says "PLANNING". Should have been the first thing updated after execution started.
4. **Don't duplicate test helpers** — I created `printer/stats/helpers_test.go` with copies of `mockReadFile`, `processTestNodes`, `printTestClones`, and `testFileContent`. These should have been extracted to a shared `internal/testutil/` or `printer/internal/` test helper package. Go doesn't allow sharing test helpers across packages easily, but `internal/` directories work.
5. **Run `nix flake check` one final time after all changes settle** — I ran it after Phase 3 and it passed, but the daemon has since dirtied `.golangci.yml`. Should re-verify.
6. **Don't rely on the daemon to commit your work** — It commits with wrong messages and at wrong times. For refactors like this, squashing into clean per-phase commits would be better.

### Code quality improvements

7. **The `healthSmall`/`healthMedium`/`healthLarge`/`healthHuge` constants are now duplicated** — They live in `printer/stats/stats_health.go` (prod) and `printer/groups_test.go` (test, in root). These should be exported from stats or moved to domain if they're domain concepts.
8. **Type aliases in root `clone_processor.go`** — `CloneCategory`, `ClonePriority`, `CloneClassification` are `domain.*` aliases kept for backward compat with `html*.go`. These should eventually be replaced with direct `domain.*` references.
9. **`stats_data.go` stayed in root** — The plan wanted to rename `StatsView` → `stats.View`, but I kept it in root to avoid the circular dependency. This means `printer/stats/` imports root for `printer.StatsView`, which is correct but means the stats package isn't fully self-contained.
10. **`TestPriorityHigher` was orphaned to root** — It tests `priorityHigher` from `html.go` but was originally in `clone_classify_test.go`. Now it's in a standalone `priority_higher_test.go` file. Works, but the test organization is slightly worse.

---

## f) Next 50 Things to Get Done

### Immediate (blocking / quality gate)

1. **Fix `.golangci.yml` — remove daemon-re-added `exhaustruct`/`tagliatelle`** (blocking: `check-disabled-linters.sh` fails right now)
2. **Commit the `TODO_LIST.md` change** (sitting in working tree)
3. **Update plan doc `Status:` to `DONE — executed 2026-07-26`**
4. **Check all 8 success criteria boxes in the plan doc**
5. **Re-run `nix flake check` after fixing `.golangci.yml`** to confirm green

### Deduplication cleanup (introduced this session)

6. **Extract `mockReadFile` to a shared test helper** — duplicated in `printer/json_test.go` and `printer/stats/helpers_test.go`
7. **Extract `testFileContent` constant** — duplicated in `printer/html_golden_test.go` and `printer/stats/helpers_test.go`
8. **Consolidate `healthSmall`/`healthMedium`/`healthLarge`/`healthHuge`** — duplicated in `printer/groups_test.go` and `printer/stats/stats_health.go`. Export from stats or move to domain.
9. **Consider a `printer/internal/testutil/` package** for shared test helpers across printer sub-packages

### Actionability package hardening

10. **Export `CloneCategory`/`ClonePriority` type aliases from actionability** — currently test files use `domain.CloneCategory` directly, but the aliases existed before for ergonomics
11. **Add a package-level doc comment** to `printer/actionability/` explaining its role as a leaf package
12. **Verify arch-lint actually catches actionability→printer violations** (write a negative test or manually try importing printer from actionability)

### Stats package hardening

13. **Add a package-level doc comment** to `printer/stats/` explaining it imports root for interface satisfaction
14. **Consider whether `StatsView` should move to `domain/`** — it's a pure data struct with only `domain.*` types + primitives. Moving it would eliminate the stats→root import dependency.
15. **Consider whether `StatsConfig` should move to `domain/`** or `config/` — same reasoning
16. **Add an arch-lint negative test** for stats→actionability imports

### Arch-lint / CI improvements

17. **Run go-arch-lint standalone** to verify the new component rules work (not just via nix flake check)
18. **Add the daemon linter regression as a pre-commit hook** (`.git/hooks/pre-commit`) — the Nix check only catches it in CI, not locally when the daemon commits
19. **Consider a `.pre-commit-config.yaml`** for the linter guard
20. **Add an arch-lint test that asserts component count** (fails if someone adds/removes a component without updating the test)

### Format printer extraction (Phase 4 — evaluate later)

21. **Assess root printer LOC after settling** — is ~3,830 LOC still too large?
22. **If yes, extract `printer/text/`** — text.go + text_utils + text_golden_test
23. **Extract `printer/json/`** — json.go + json_test.go
24. **Extract `printer/html/`** — html*.go + report.templ + report_templ.go
25. **Extract `printer/sarif/`** — sarif.go + sarif_test.go
26. **Extract `printer/plumbing/`** — plumbing.go + plumbing_test.go
27. **Each extraction needs to import root for `Printer` interface** — same pattern as stats

### General codebase health

28. **Replace em-dashes with commas** in any new code (AGENTS.md convention)
29. **Audit the 13 daemon commits** from this session — squash if appropriate, or at least verify no unintended changes
30. **Update `FEATURES.md`** if the package structure changed how features are described
31. **Update `ROADMAP.md`** — the printer split was a roadmap item, mark it done
32. **Consider adding `printer/actionability/` to the SDK** — `pkg/artdupl` could expose actionability evaluation directly
33. **Review if `clone_classify.go` belongs in actionability** — it's classification logic, but maybe it should be in `domain/` or its own package
34. **Add integration test** that runs art-dupl on its own source with the new package structure
35. **Verify the self-test (`art-dupl` on its own source) still detects real duplication** and doesn't flag the new package boundaries as clones

### Documentation

36. **Write an ADR for the printer split** (ADR-0017 or similar) — document the decision to keep interfaces in root, the import direction (root→actionability, stats→root), and the deferral of Phase 4
37. **Update `docs/DOMAIN_LANGUAGE.md`** if any domain terms shifted packages
38. **Update `SDK_DESIGN.md`** if the SDK's relationship to printer changed (it didn't, but worth verifying)
39. **Add the new packages to the README module tree** if there is one
40. **Document the `healthSmall` constant duplication** as known debt or fix it

### Testing improvements

41. **Add a test that asserts `printer/actionability/` has zero imports of `printer/`** (arch-lint negative test)
42. **Add a test that asserts `pkg/artdupl/` has zero imports of `printer/`** (already verified manually, should be permanent)
43. **Add race detection tests for the new package boundaries** (goroutines crossing package boundaries)
44. **Run `go test -race ./printer/...`** specifically to catch any cross-package data races introduced by the split
45. **Verify golden test outputs are byte-identical** before and after the split (behavior-preserving verification)

### Technical debt

46. **The type aliases in `clone_processor.go`** (`CloneCategory`, `ClonePriority`, `CloneClassification`) should be replaced with direct `domain.*` references across `html*.go` — mechanical cleanup
47. **`priorityHigher` in `html.go`** could move to `domain/` as a method on `ClonePriority` — it's a comparison utility
48. **The `FileStatMixin` type** in `stats_formatter.go` is only used within stats — confirm it's not referenced elsewhere
49. **The `StyleMixin` type** in `stats_styles.go` is only used within stats — confirm it's not referenced elsewhere
50. **Squash the 13 daemon commits** from this session into 3 clean commits (one per phase) for a readable history

---

## g) Questions (3)

### Q1: Should I squash the 13 daemon commits into 3 clean per-phase commits?

The daemon committed 13 times during this session with generic/wrong messages. The alternative is `git rebase -i HEAD~13` and squash into 3 commits: "Phase 1: extract actionability", "Phase 2: extract stats", "Phase 3: arch-lint + docs". This would make the history readable but rewrites history that's already been pushed (if the daemon pushed). **I cannot determine if the daemon has already pushed these commits to remote** — can you check, and should I squash?

### Q2: Should `StatsView`/`StatsConfig` move to `domain/`?

Currently `stats_data.go` stays in root `printer/` because `StatsView` is referenced by the `StatsPrinter` interface in `printer.go`. But `StatsView` is a pure data struct containing only `domain.*` types + primitives. Moving it to `domain/` would eliminate the stats→root import dependency entirely, making stats a true leaf. The tradeoff: `StatsConfig` contains `config.OutputFormat`, so it can't go to domain without a circular dependency. **Should I move `StatsView` to domain and keep `StatsConfig` in root, or leave both in root?** I can reason about either approach but the tradeoff is non-trivial.

### Q3: Should the `healthSmall`/`healthMedium`/`healthLarge`/`healthHuge` constants be exported from `domain/`?

These severity bucket labels ("small", "medium", "large", "huge") are now duplicated between `printer/stats/stats_health.go` (prod) and `printer/groups_test.go` (root test). They represent a concept that spans both packages (health score calculation + group sorting). Moving them to `domain/` (e.g., as `domain.HealthSmall`, `domain.HealthMedium`) would eliminate the duplication. **Or should they stay as package-private implementation details in both places?** The duplication is only 4 string constants, but it's the kind of thing that drifts.

---

## Summary

The printer package decoupling sprint executed Phases 1-3 successfully. Two clean leaf packages were extracted (`printer/actionability/` at 1,670 LOC, `printer/stats/` at 1,303 LOC), reducing the root `printer/` from ~6,600 → ~3,830 LOC. All 11 `nix flake checks` passed. The boundaries are enforced by `.go-arch-lint.yml`.

**However**: The daemon re-added the disabled linters to `.golangci.yml` at the end of the session (working tree is dirty), several test helpers were duplicated rather than shared, the plan doc status wasn't updated, and the git history is messy (13 daemon commits). These are all fixable but need immediate attention.
