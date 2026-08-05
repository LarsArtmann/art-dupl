# Status Report: TODO List Execution Sprint

**Date:** 2026-08-05 18:21
**Session scope:** Execute the 3 open TODO_LIST.md items + self-review

---

## a) FULLY DONE

### Task 1 (HIGH): Large-Scale Confidence Calibration

**What was done:**

- Built art-dupl binary and ran `--threshold 5 --semantic --explain` on **6 projects with 4,764 Go files**
- Captured all **87 clone groups** with file locations and first-line previews
- **Manually labeled every single group** as TRUE POSITIVE or FALSE POSITIVE by reading source code
- Verified ambiguous labels (type definitions vs coincidental fields, type aliases vs re-exports) by reading actual source at each location
- Computed precision metrics per-project and overall
- Wrote comprehensive report: `docs/calibration/confidence-calibration-large-2026-08-05.md`
- Corrected the old report's false "0% false positives" claim with a SUPERSEDED header
- Converted 4 specific findings into actionable TODO items

**Key result:** **86.2% precision** (75 TP / 12 FP). The initial "0% FPs" claim was wrong — it assumed all "actionable" classifications were correct without manual verification. Confidence _values_ are well-calibrated; the fix opportunity is in suppression _patterns_, not thresholds.

**Files created/modified:**

- `docs/calibration/confidence-calibration-large-2026-08-05.md` (new, comprehensive)
- `docs/calibration/confidence-thresholds-2026-08-05.md` (SUPERSEDED header added)
- `TODO_LIST.md` (calibration item replaced with 4 new HIGH items)
- `CHANGELOG.md` (entry added)

### Task 2 (MEDIUM): go-arch-lint CI Integration

**What was done:**

- Created `.github/workflows/arch-lint.yml` — GitHub Actions workflow that installs go-arch-lint v1.16.0 and runs `go-arch-lint check`
- Triggers on push/PR to main/master/fork when `.go` files, `.go-arch-lint.yml`, or the workflow itself changes
- Includes templ generation + module download steps (go-arch-lint needs compiled packages)
- Verified `go-arch-lint check` passes locally with 0 violations
- Updated `flake.nix` comment to reference the new workflow instead of the old "run manually" instruction
- Fixed stale CHANGELOG entry that falsely claimed go-arch-lint ran in `nix flake check`

**Files created/modified:**

- `.github/workflows/arch-lint.yml` (new)
- `flake.nix` (comment updated, 2 lines)
- `CHANGELOG.md` (corrected entry)

### Task 3 (MEDIUM): Auto-Fix Guard Script Test

**What was done:**

- Created `scripts/check_disabled_linters_test.go` with **6 test cases**:
  1. `TestAutoFixRemovesBannedLinters` — core TODO: verifies `sed -i` removes exhaustruct+tagliatelle, preserves valid linters, exits 0
  2. `TestAutoFixRemovesSettingsBlock` — verifies orphaned `exhaustruct:` settings key is stripped
  3. `TestReadOnlyFailsWithBannedLinter` — verifies read-only files fail with exit 1 and "read-only" message
  4. `TestCleanFilePasses` — verifies clean configs pass with OK message, unmodified
  5. `TestCommentsMentioningBannedLintersAllowed` — verifies explanatory comments are NOT flagged or stripped
  6. `TestMissingConfigFails` — verifies missing file errors
- Fixed lint issues: `CommandContext`, `wsl_v5` whitespace, `nonamedreturns`, `staticcheck` error-last convention
- All tests pass, 0 lint issues

**Files created:**

- `scripts/check_disabled_linters_test.go` (new, 200 lines)

### Bonus: Fixed Broken `nix flake check`

**Discovered during verification:** The auto-commit daemon had re-added `tagliatelle` to `.golangci.yml` (commit `d7d1d96e`). This broke `nix flake check` — the disabled-linters Nix check failed because the sandbox is read-only and can't auto-fix. Removed the linter manually.

**This proves the guard script works exactly as designed** — it caught the regression. The auto-commit daemon is the exact threat model described in the workflow comments.

---

## b) PARTIALLY DONE

### Nothing

All 3 tasks were completed end-to-end.

---

## c) NOT STARTED (from this session's scope)

None — all 3 TODO items were addressed. The 4 new HIGH priority items created from calibration findings are documented but intentionally not started (they require code changes to the actionability engine, which is new work, not a TODO execution task).

---

## d) TOTALLY FUCKED UP

### Near-miss: Shipping without running `nix flake check`

**What happened:** I ran `go build`, `go test`, and `golangci-lint` — all passed. But I almost wrote the status report without running `nix flake check`. When I finally ran it, it **failed** because the auto-commit daemon had re-added `tagliatelle` to `.golangci.yml` during this session.

**Root cause:** The auto-commit daemon (described in `~/.config/crush/AGENTS.md`) continuously commits changes. It re-added a banned linter while I was working. My test suite didn't catch this because `go test` and `golangci-lint` don't run the Nix disabled-linters check.

**Lesson:** Always run `nix flake check` as the final verification gate, not just `go test` + `golangci-lint`. The Nix checks test things the Go toolchain doesn't.

### Stale flake.nix comment

The `flake.nix` comment referenced a status report (`docs/status/2026-08-05_17-33_todo-list-execution-sprint.md`) and said "run manually in devShell." This was stale — the file was from a prior session and the guidance was outdated. Fixed to point to the new GitHub Actions workflow.

### Prior session's false claim

The old calibration report (`confidence-thresholds-2026-08-05.md`) claimed "0% false positives" based on **no manual verification** — it just trusted the tool's "actionable" classification. This is the kind of "trophy-case marking" that the verify-external-claims skill exists to prevent. The new report corrects this.

---

## e) WHAT WE SHOULD IMPROVE

1. **The auto-commit daemon is a reliability hazard.** It re-added a banned linter mid-session, breaking `nix flake check`. The guard script caught it, but only because I ran the Nix check manually. Consider: (a) making the guard script a pre-commit hook, or (b) having the daemon run `nix flake check` before committing.

2. **Calibration measured precision but NOT recall.** All 87 found clones were "actionable" — 0 were suppressed by the tool. We don't know how many actionable clones were wrongly suppressed (false negatives). The `--show-suppressed` flag (new TODO item) would fix this.

3. **The calibration script (`calibrate-confidence.sh`) parses `--explain` output via grep.** This is fragile — if the explain format changes, the counts break silently. Consider a JSON output mode for machine consumption.

4. **LSP diagnostics were stale throughout the session.** The LSP reported 21 warnings on the test file that didn't exist when `golangci-lint` ran. This caused confusion and wasted cycles. The LSP cache needs invalidation after edits.

5. **No CI runs on the `fork` branch for the arch-lint workflow.** The workflow triggers on `main/master/fork` but has never been tested in actual GitHub Actions. The `go install go-arch-lint@v1.16.0` step might fail if the version doesn't exist on GitHub.

6. **The test file lives in `scripts/` as a Go package.** This means `go build ./...` now compiles a package in a directory that previously had no `.go` files. This works but is unconventional — most projects keep tests adjacent to the code they test or in a dedicated `internal/` dir.

7. **Labeling was done by a single person (me) in one pass.** Inter-rater reliability is unknown. A second pass or a different labeler might disagree on borderline cases (e.g., duplicated type definitions in different packages — are they TP or "intentional domain reuse"?).

---

## f) Up to 50 Things to Get Done Next

### From calibration findings (HIGH impact, eliminates FPs)

1. **Exclude demo/example directories by default** — `examples/`, `demo/`, `test_plugin/`, `test_temp/`. Eliminates 4/12 FPs.
2. **Add `--include-examples` CLI flag** to override the demo exclusion.
3. **Extend single-declaration pattern to type-alias blocks** — detect `type(...)` groups with 2+ `X = pkg.Y` aliases. Eliminates 3/12 FPs.
4. **Add interface-assertion actionability pattern** — `_ I = (*T)(nil)`. Eliminates 1 FP.
5. **Widen test-helper-delegate to `testing.B`** — `b.Helper()` delegate. Eliminates 1 FP.
6. **Add `--show-suppressed` flag** to surface suppressed groups for recall measurement.
7. **Add JSON output to calibration script** for machine-parseable results.
8. **Re-run calibration after implementing items 1-5** to verify precision improvement.

### Actionability engine improvements

9. **Add report-generation-idiom pattern** — two report builders writing different strings with similar structure (3 borderline FPs).
10. **Add size-based confidence penalty** for clones under 10 tokens (the borderline zone identified in calibration).
11. **Test the actionability patterns against the 87 labeled clones** as a regression suite.
12. **Document the actionability pattern priority order** in a decision tree diagram.
13. **Add property-based testing for actionability patterns** (input: random AST, output: consistent classification).

### CI/Infrastructure

14. **Add `nix flake check` as a pre-commit hook** to catch auto-commit daemon regressions before they land.
15. **Test the arch-lint workflow in actual GitHub Actions** — verify `go install go-arch-lint@v1.16.0` works.
16. **Pin go-arch-lint version in a Nix devShell alias** as a fallback for local development.
17. **Add a CI step that runs `scripts/check_disabled-linters.sh`** on every push (not just when `.golangci.yml` changes), since the daemon can modify it at any time.
18. **Add coverage reporting to the scripts test** — verify all guard script branches are exercised.
19. **Create a `make lint-ci` target** that runs golangci-lint + arch-lint + disabled-linters in sequence.

### Calibration methodology

20. **Run calibration on a monorepo** (1000+ files in one project) to test scaling.
21. **Run calibration on non-Go projects** to verify templ detection (if any templ projects exist at scale).
22. **Have a second person label a subset** of the 87 clones for inter-rater reliability.
23. **Compute Cohen's kappa** between the tool's classification and human labels.
24. **Track precision over time** — re-run calibration after each actionability pattern change.
25. **Add a "ground truth" dataset** of labeled clones to the repo for regression testing.
26. **Measure false negative rate** by running with `--no-actionability` and comparing to default.

### Code quality

27. **Move the scripts test to `internal/scripts/`** or add a build tag to exclude from production builds.
28. **Add integration test for the full calibration pipeline** (build → run → parse → report).
29. **Refactor `calibrate-confidence.sh` to use `--json` output** instead of grep-parsing explain text.
30. **Add type safety to the guard script** — validate YAML structure, not just regex patterns.
31. **Add a `--dry-run` flag to the guard script** to preview fixes without applying.
32. **Document the guard script's regex behavior** in a testable specification.

### Documentation

33. **Update AGENTS.md** with the calibration methodology and 86.2% precision finding.
34. **Update HOW_TO_USE.md** with the `--show-suppressed` flag (once implemented).
35. **Add a "Calibration" section to the README** explaining how to evaluate false positive rates.
36. **Document the FP categories** in `docs/ACTIONABILITY_PATTERNS.md`.
37. **Write an ADR for the demo-directory exclusion decision.**
38. **Write an ADR for the type-alias-block suppression decision.**

### Testing

39. **Add BDD test for demo-directory exclusion.**
40. **Add BDD test for type-alias-block suppression.**
41. **Add BDD test for interface-assertion pattern.**
42. **Add test for `testing.B` helper delegate.**
43. **Add race test for the calibration script** (parallel file access).
44. **Add fuzz test for the guard script regex** (malformed YAML inputs).
45. **Add snapshot test for explain output format** (detect format drift).

### Performance

46. **Benchmark calibration on 10K+ files** to measure analysis time.
47. **Profile the actionability evaluation** on large clone sets.
48. **Add caching for actionability pattern evaluation** (currently re-evaluates per clone).

### Architecture

49. **Consider a plugin system for actionability patterns** (instead of hardcoded switch).
50. **Extract calibration tooling into a standalone subcommand** (`art-dupl calibrate`).

---

## g) Questions (cannot figure out myself)

### 1. Should the 4 new HIGH priority TODO items be implemented now or deferred?

The calibration produced 4 concrete actionability-engine improvements (demo exclusion, alias-block pattern, interface-assertion pattern, `--show-suppressed` flag). Each eliminates a measured false positive category. Implementing all 4 would raise precision from 86.2% to ~94%. But they require code changes to the detection/actionability engine — is that in scope for this sprint, or should they wait for a dedicated actionability-engine sprint?

### 2. Is the auto-commit daemon's re-adding of `tagliatelle` a known issue or a bug?

The daemon commit `d7d1d96e` ("chore: re-add tagliatelle linter") happened during this session and broke `nix flake check`. This is the exact scenario the guard script was built for. But: is the daemon supposed to respect the `.golangci.yml` disabled-linters policy? Or is it expected to make independent formatting/linting decisions that the guard catches?

### 3. Should the calibration ground-truth labels (87 groups) be committed to the repo?

The manual labels (75 TP / 12 FP with per-group rationale) are valuable as a regression test dataset. But they reference external projects (`picoclaw`, `CreditReformBilanzampel`, etc.) that may change or be deleted. Should I: (a) commit the labels as a JSON/CSV dataset with file locations, (b) create synthetic test cases that reproduce each FP category, or (c) keep the labels only in the report as documentation?

---

## Verification Summary

| Check                                              | Status              |
| -------------------------------------------------- | ------------------- |
| `go build ./...`                                   | PASS                |
| `go test ./scripts/...`                            | PASS (6/6 tests)    |
| `go test ./cmd/...`                                | PASS                |
| `golangci-lint run ./...`                          | PASS (0 issues)     |
| `nix build .#checks.x86_64-linux.disabled-linters` | PASS                |
| `go-arch-lint check` (local)                       | PASS (0 violations) |
| `bash scripts/check-disabled-linters.sh`           | PASS                |

All green after fixing the tagliatelle regression mid-session.
