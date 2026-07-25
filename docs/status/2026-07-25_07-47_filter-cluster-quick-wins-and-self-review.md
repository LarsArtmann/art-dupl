# Status Report — Filter Cluster Quick Wins + Self-Review

**Date:** 2026-07-25 07:47
**Session scope:** Filter code optimization, self-review, lint guard fix, planning
**Branch:** fork

---

## A) FULLY DONE

### 1. Filter Marker Unification + Optimization (TODO lines 24, 26, 27)
- **Extracted `matchedGeneratedCategory`** (`cmd/util.go`) — single source of truth for templ/sqlc/protobuf marker matching. Both `allowsContent` and `filterExcludedGenerated` delegate to it. Eliminated the duplicated marker switch.
- **`bytes.Contains` fast-path** — regular files (~95% of files) now return after a single `bytes.Contains(content, []byte("Code generated"))` with zero `string(content)` allocation. Escape analysis confirms stack-allocated needles.
- **Benchmarks prove it:** `BenchmarkMatchedGeneratedCategory_RegularFile` = **21ns, 0 allocs/op**.

### 2. Missing Test Coverage (TODO line 43)
- **Added `cmd/filter_stats_test.go`** — dedicated coverage for `RecordWithSource`, `SourceBreakdown()`, defensive-copy guarantees, nil-receiver safety.
- **Found and fixed a real nil-panic bug**: `Breakdown()`/`SourceBreakdown()` panicked on nil receiver because `s.byReason`/`s.bySource` were evaluated as arguments before the nil-checked `copyMapUnderLock` ran. Refactored to closure pattern.

### 3. Lint Guard Fix (CRITICAL — was broken)
- **Commit `8ea7f1aa`** re-enabled `exhaustruct` + `tagliatelle` in `.golangci.yml`, breaking the Nix CI guard (`flake.nix:199` `disabled-linters` derivation).
- **Removed both linters** from enable list and config blocks. Guard passes again.
- **NOTE: The auto-git daemon REVERTED this fix once already** (reformatted file 2-space→4-space AND re-added the lines). Fixed again in working tree. This is fragile — the daemon may revert again.

### 4. Drift Detection Tests
- **`TestMarkersMatchGogenfilter`** — cross-validates our hardcoded markers against gogenfilter's actual detection. Catches silent breakage if gogenfilter changes marker strings.
- **`TestMockgenStringerMarkersMatchGogenfilter`** — documents the mockgen/stringer contract.

### 5. Documentation Corrections
- **AGENTS.md**: Corrected SQLC filename patterns (gogenfilter matches `models.go`, `querier.go`, `query.sql.go`, `batch.go`, `*.sql.go` — NOT `_sqlc.go` as documented).
- **TODO_LIST.md**: Corrected lazy-reading TODO to clearly state it's blocked by upstream gogenfilter API.
- **Self-review report**: `docs/reviews/2026-07-25_07-39_brutal-self-review.html` with D2 execution graph.

---

## B) PARTIALLY DONE

### Filter path optimization
- Marker matching: **DONE** (bytes.Contains + early-exit + unified helper).
- File reading: **NOT DONE** — `shouldIncludeFile` still reads content upfront when includes are active. Blocked by gogenfilter API (`FilterDetailed` reads internally but doesn't return content).

### Test coverage for filter stats
- Source tracking, defensive copies, nil safety: **DONE**.
- Removed redundant `TestSourceBreakdown` (duplicate of `TestFilterStatsSourceBreakdown` in `filter_includes_test.go:347`).

---

## C) NOT STARTED (from TODO_LIST.md)

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 1 | `--no-actionability` flag | 30min | HIGH |
| 2 | `--explain` flag | 2h | HIGH |
| 3 | `--diff-report baseline` mode | 2h | HIGH |
| 4 | HTML report improvements | 1h | MED |
| 5 | YAML config (`.artdupl.yml`) | 2h | MED |
| 6 | Refactor `generatorIncludes` struct | 1h | LOW |
| 7 | Interface-method-aware suppression | 3h | MED |
| 8 | `--recommend-threshold` | 2h | LOW-MED |
| 9 | Templ Phase 3: expression normalization | 2h | LOW |
| 10 | Split `printer/` into sub-packages | LARGE | HIGH |
| 11 | Push defense-in-depth to gogenfilter | Upstream | MED |
| 12 | Branded `NodeType int32` | HIGH RISK | MED |
| 13 | Hide `syntax/golang` behind facade | Blocked | MED |

---

## D) TOTALLY FUCKED UP

### 1. Auto-git daemon reverted my lint fix
- I removed `exhaustruct`+`tagliatelle` from `.golangci.yml`.
- The auto-git daemon reformatted the file (2-space→4-space indentation) AND re-added both linters.
- **I didn't catch this until the status report investigation.** My "fix" was silently undone.
- Fixed again in working tree, but this will keep happening unless the daemon is configured to not touch `.golangci.yml`, or I commit immediately after fixing.

### 2. I didn't run `nix run .#lint` / `nix flake check` initially
- AGENTS.md explicitly mandates using `flake.nix` for all build/task automation.
- I ran `golangci-lint` manually instead, which doesn't execute the `disabled-linters` Nix derivation.
- The broken guard was invisible to my manual lint run.

### 3. Created a redundant test (split brain)
- `TestSourceBreakdown` in `filter_stats_test.go` duplicated `TestFilterStatsSourceBreakdown` in `filter_includes_test.go:347`.
- Removed it, but I should have checked for existing coverage before writing.

---

## E) WHAT WE SHOULD IMPROVE

### Process
1. **Always run Nix checks** — `nix build .#checks.x86_64-linux.lint` and `nix build .#checks.x86_64-linux.disabled-linters`, not just `golangci-lint`.
2. **Check for existing test coverage before writing new tests** — grep for the function name in `_test.go` files first.
3. **Commit immediately after fixing** — the auto-git daemon can revert uncommitted changes.
4. **Run `git diff` before declaring done** — verify working tree matches expectations.

### Architecture
5. **Marker constants duplicate gogenfilter** — mitigated with drift-detection test, but the real fix is upstream (expose marker constants from gogenfilter).
6. **`generatorIncludes` 6-bool struct** — adequate after `matchedGeneratedCategory` refactor. Not worth the 50-call-site churn right now.
7. **`FilterStats` nil-safety** — all read methods are now nil-safe. This should be a package-wide convention.

### Testing
8. **No benchmarks for config/filter/baseline paths** — added filter benchmarks this session, but config loading and baseline operations have zero benchmark coverage.
9. **No property-based tests** — the marker matching is a pure function that would benefit from property-based testing (e.g., "any file without 'Code generated' never matches").

---

## F) NEXT 50 THINGS TO DO

### Tier 1 — 1% effort, 51% impact (DO FIRST)
1. `--no-actionability` flag — disable actionability filtering, fixes BDD fragility
2. `--explain` flag — explain WHY a clone was reported (method, pattern, actionability)
3. `--diff-report <baseline>` — show new/resolved/suppressed clones vs baseline
4. Fix auto-git daemon reverting `.golangci.yml` — investigate why it re-adds removed linters
5. Add `disabled-linters` check to CI pipeline (not just local Nix)
6. Add filter benchmarks to `nix build .#checks.x86_64-linux.bench`

### Tier 2 — 4% effort, 64% impact
7. HTML report: `--html-output <file>` flag
8. HTML report: TTY auto-detection (stdout=terminal → text, otherwise HTML)
9. HTML report: stable `id` attributes on clone groups for deep-linking
10. YAML config support (`.artdupl.yml`) via `go-faster/yaml`
11. YAML config: add `MarshalYAML`/`UnmarshalYAML` hooks to domain enums
12. Config file discovery (`.artdupl.yml` / `.artdupl.json` in cwd or parent dirs)
13. `--config <path>` flag for explicit config file
14. Refactor `generatorIncludes` to `map[GenerationCategory]bool` (if new categories are planned)
15. Add `GenerationCategory` domain enum type (follow existing `DetectionMode` pattern)
16. Baseline diff: implement `baseline.Diff(old, new)` returning `New/Resolved/Unchanged` sets
17. `--diff-report` JSON output format for CI integration
18. Property-based test for `matchedGeneratedCategory` (any input without header → false)

### Tier 3 — 20% effort, 80% impact
19. Interface-method-aware suppression at all thresholds
20. `--recommend-threshold` based on codebase size + test-to-production ratio
21. Templ Phase 3: expression normalization (`{ id.String() }` vs `{ groupID.String() }`)
22. Watch mode (`--watch`) — re-run on file change
23. SARIF output improvements for GitHub Code Scanning
24. Exit code 4 (clones found) as distinct from 1 (error) for CI integration
25. `--min-lines` filter for suppression (already in `SuppressionConfig`, wire to CLI)
26. Cache format versioning (bump on incompatible `Node` changes)
27. Progress output: ETA calculation for large codebases
28. `--include-generated` accepts comma-separated values (`--include-generated sqlc,templ`)

### Code Quality
29. Split `printer/` into sub-packages (blocked by circular dep on `StatsPrinter`)
30. Move `Printer`/`ReadFile`/`StatsPrinter` interfaces to `printer/base/`
31. Extract HTML printer to `printer/html/`
32. Extract JSON printer to `printer/json/`
33. Extract text printer to `printer/text/`
34. Extract SARIF printer to `printer/sarif/`
35. `printer/actionability.go` still imports `syntax.Node` directly — decouple via `domain.ProcessedClone`
36. Eliminate `printer.simpleJSONClone` — fold into `JSONClone`
37. Consolidate `printer.CloneGroup` / `CloneGroupView` / `pkg/artdupl.CloneGroup` (legitimately different DTOs, but verify)
38. Add `context.Context` to `Printer` interface methods (currently context-less)

### Detection
39. Type-aware detection: support `--incremental` mode (currently incompatible)
40. Type-aware detection: cache `go/packages` results between runs
41. Semantic mode: encode `SelectStmt` column names (prevent `SELECT a FROM t` matching `SELECT b FROM t`)
42. Semantic mode: encode `RangeClause` context (table names are API surface)
43. Add detection for Go generics instantiation patterns (`Foo[int]` vs `Foo[string]`)
44. Structural mode: add option to ignore comments (currently always included)

### Infrastructure
45. GitHub Actions CI workflow (run `nix flake check` on PR)
46. Pre-commit hook running `golangci-lint fmt` (prevent formatting churn)
47. Dependabot / nix flake update automation
48. Add `gosec` + `govulncheck` to Nix checks
49. Release automation via GitHub Actions + `nix build` for versioned binaries
50. Architecture decision record (ADR) for the `generatorIncludes` design decision

---

## G) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **The auto-git daemon keeps reverting `.golangci.yml`** (re-adding `exhaustruct`+`tagliatelle` after I remove them). Is this a configured formatter hook? Should I add `.golangci.yml` to a "do not auto-format" list, or is there a different root cause?

2. **`GOEXPERIMENT=jsonv2` is required** but `gopls` warns `version_cmd.go:49` uses `json.Marshal` which "requires go1.27" (we're on go1.26.5). Is this a real compatibility issue or a gopls false positive? The build works fine with `GOEXPERIMENT=jsonv2`.

3. **Should `--diff-report` be a subcommand** (`art-dupl diff <baseline>`) or a flag on the root command (`art-dupl --diff-report <baseline> ./...`)? The former is cleaner for CI; the latter matches the existing `--baseline` flag pattern. I can't tell which UX you prefer without asking.
