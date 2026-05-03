# Status Report — 2026-05-03 02:13

## Summary

Comprehensive audit + critical bug fixes + dead code elimination + feature gap closure.

---

## ✅ FULLY DONE

| # | Task | Commit | Impact |
|---|------|--------|--------|
| 1 | **Fix SortByTotalTokens bug** — was falling through to SortBySize in `printer/sorter.go:39` and `printer/text.go:221` | `30a5a32` | CRITICAL — sorting was silently wrong |
| 2 | **Fix semantic flag description** — was "already the default" but config default is `false` | `30a5a32` | HIGH — user-facing misinformation |
| 3 | **Fix structural flag description** — was "opt-out from default" but structural IS the default | `30a5a32` | HIGH — confusing UX |
| 4 | **Fix TestFindProjectRoot** — removed `maxDepth=10`, fixed test markers | `30a5a32` | MEDIUM — pre-existing CI failure |
| 5 | **Delete dead cli/ package** — moved `DefaultThreshold` to `config/` | `8f8aa18` | LOW — removed confusion |
| 6 | **Add --simple-json CLI flag** — feature was inaccessible from CLI | `c6346cb` | MEDIUM — feature gap closed |
| 7 | **Rewrite FEATURES.md** — 280+ features with honest status indicators | `eaff254` | HIGH — accurate feature inventory |
| 8 | **Fix AGENTS.md** — semantic default, SARIF, sorting, filtering accuracy | `eaff254` | HIGH — docs accuracy |
| 9 | **Fix TODO_LIST.md** — removed stale TODO, added README tasks | `eaff254` | MEDIUM — actionable todos |
| 10 | **Fix 3 lint issues** — prealloc + wsl_v5 in BDD tests | `eaff254` | LOW — clean lint |
| 11 | **Add 11 BDD tests** — SARIF, total-tokens, --only, --diff, protobuf/mockgen | `eaff254` | HIGH — coverage for untested features |
| 12 | **Architecture diagrams** — current + improved mermaid graphs | `eaff254` | MEDIUM — visual documentation |
| 13 | **Execution plan** — Pareto breakdown with mermaid graph | `3996fb6` | MEDIUM — structured plan |

---

## ⚠️ PARTIALLY DONE

| Task | What's Done | What Remains |
|------|-------------|--------------|
| **README.md update** | Identified ALL inaccuracies (semantic, templ, --exclude-templ, missing SARIF/--only/--diff) | Actual file edits not applied yet — 8 inaccuracies remain |
| **Architecture improvements** | Full analysis with concrete findings | `printer/format.go` and `printer/sort_type.go` type aliases not deleted yet |
| **Sort consolidation** | Identified 4 duplicate switch blocks | Not consolidated yet — needs generic dispatch |

---

## ❌ NOT STARTED

| # | Task | Work | Impact |
|---|------|------|--------|
| 1 | Delete `printer/format.go` type alias — use `config.OutputFormat` directly | 5 files | MEDIUM |
| 2 | Delete `printer/sort_type.go` type alias — use `config.SortCriteria` directly | 13 files | MEDIUM |
| 3 | Update README.md — fix 8 inaccuracies + add missing features | 1 file | HIGH |
| 4 | Extract `DetectionConfig` from `config.Config` god struct | 5+ files | HIGH |
| 5 | Define `MethodDetector` interface in `detection/` | 4 files | HIGH |
| 6 | Refactor MultiDetector to use registry pattern | 4 files | HIGH |
| 7 | Wire TODO/Legacy detectors to CLI | `cmd/`, `detection/` | MEDIUM |
| 8 | Consolidate 4 sorting switch blocks into 1 generic dispatch | 3 files | HIGH |
| 9 | SDK parity with CLI — add filtering, incremental, workers, semantic to SDK | 10+ files | HIGH |

---

## 💥 TOTALLY FUCKED UP — Nothing!

All changes verified:
- **22/22 packages pass tests**
- **0 lint issues**
- **0 regressions**
- **Previously-failing `internal/utils` now passes**

---

## 🔧 WHAT WE SHOULD IMPROVE

### Critical
1. **README.md is stale** — 8 inaccuracies including wrong flag names and wrong defaults
2. **SortByTotalTokens was silently broken** — fixed in this session, but was present for unknown duration

### High Priority
3. **`printer/format.go` and `printer/sort_type.go`** — pure type aliases adding indirection, should use `config.*` directly
4. **`config.Config` is a 30-field god struct** — `MultiDetector` receives it but only reads 1 field (`DetectionMethods`)
5. **SDK diverges from CLI** — 12+ features missing from SDK (filtering, incremental, workers, semantic, diff, sort)
6. **4 duplicate sorting switch blocks** — same logic across 4 different data representations

### Medium Priority
7. **`domain/` types barely used** — only `detection/todos.go` uses `Filepath`, `LineNumber`, `CloneSeverity`
8. **TODO/Legacy detectors are dead code** — implemented but not wired to CLI
9. **BDD test schema validation** — SARIF test checks only `"version"` and `"runs"`, not full SARIF 2.1.0 schema

---

## 🏆 Top 25 Next Steps (sorted by impact/effort)

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 1 | Update README.md — fix all 8 inaccuracies | 30min | HIGH |
| 2 | Delete `printer/format.go` type alias | 15min | MEDIUM |
| 3 | Delete `printer/sort_type.go` type alias | 20min | MEDIUM |
| 4 | Extract `DetectionConfig` from `config.Config` | 25min | HIGH |
| 5 | Update `MultiDetector` to take `DetectionMethods` directly | 10min | HIGH |
| 6 | Define `MethodDetector` interface in `detection/` | 15min | HIGH |
| 7 | Refactor `MultiDetector` to use registry pattern | 20min | HIGH |
| 8 | Consolidate 4 sorting switch blocks | 25min | HIGH |
| 9 | Wire TODO/Legacy detectors to CLI | 15min | MEDIUM |
| 10 | Add `--todos` and `--legacy` CLI flags | 10min | MEDIUM |
| 11 | Add BDD test for SARIF schema validation | 15min | MEDIUM |
| 12 | Add BDD test for `--workers` parallel parsing | 10min | MEDIUM |
| 13 | Add BDD test for `--since` git-aware incremental | 15min | MEDIUM |
| 14 | Add BDD test for `--timeout` | 10min | LOW |
| 15 | Fix `--diff` format validation error message | 5min | LOW |
| 16 | Use `encoding/csv` for CSV output | 20min | MEDIUM |
| 17 | Consolidate `domain` types into wider codebase | 60min | MEDIUM |
| 18 | Add SDK support for `--semantic` flag | 15min | MEDIUM |
| 19 | Add SDK support for `--workers` flag | 15min | MEDIUM |
| 20 | Add SDK support for `--filter-generated` | 20min | MEDIUM |
| 21 | Add SDK support for `--incremental` mode | 25min | MEDIUM |
| 22 | Unify CLI and SDK pipelines | 60min | HIGH |
| 23 | Replace `pkg/artdupl` type aliases with independent types | 30min | MEDIUM |
| 24 | Extract `ProcessedClone` DTO for Printer interface | 45min | HIGH |
| 25 | Remove `cli/` from AGENTS.md module structure | 5min | LOW |

---

## ❓ Top #1 Question

**Should `config.Config.Semantic` default to `true` (making the flag description accurate) or stay `false` (backward compatibility)?**

The flag description in `cmd/flags.go` originally said "already the default" which implies it was INTENDED to default to `true`. The `config/config.go` comment says "off for backward compatibility." This is a product decision — changing the default changes behavior for all existing users.

---

## Commits This Session

| Commit | Message |
|--------|---------|
| `eaff254` | docs + test: comprehensive audit suite — features, architecture, quality, BDD, docs |
| `3996fb6` | docs(plan): comprehensive execution plan with Pareto breakdown |
| `30a5a32` | fix: SortByTotalTokens bug + semantic flag + TestFindProjectRoot |
| `8f8aa18` | refactor: delete dead cli/ package, move DefaultThreshold to config/ |
| `c6346cb` | feat(flags): add --simple-json CLI flag for simplified JSON output |

## Build Status

```
22/22 packages PASS
0 lint issues
0 test failures
Previously-failing internal/utils now PASSES
```
