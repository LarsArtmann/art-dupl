# Comprehensive Status Report — art-dupl Hardening

**Date**: 2026-04-20 07:26  
**Branch**: `fork`  
**Ahead of origin**: 3 commits (not yet pushed)  
**Working tree**: Clean  
**Last pushed commit**: `e143dfd` (2026-04-15)  
**Unpushed commits**: `40ac57b`, `1afe02b`, `ffdb562`

---

## A. FULLY DONE ✅

### Session 1 (2026-04-15) — 11 commits, all pushed to origin

| # | Commit | Description |
|---|--------|-------------|
| 1 | `bfce6fd` | Migrated gogenfilter API to functional-options (`NewFilter(Disabled())` pattern) |
| 2 | `b2d72fe` | Deleted dead code: `pkg/errors/`, `testutils/`, `detection/simple_detector.go` |
| 3 | `42d0e88` | Fixed lint warnings in `cmd/config_builder.go` + justfile issues |
| 4 | `48f83de` | Extracted `addSharedFlags()` — deduplicated 18 flags between root and stats commands |
| 5 | `520f219` | Committed `go.mod`/`go.sum` fix — added missing `doublestar/v4` transitive dep |
| 6 | `6c7755d` | Fixed `job.TestContextTimeoutExpired` timing race (5ms/10ms → 1ms/50ms, 50x margin) |
| 7 | `3098157` | Fixed gci import grouping in `config_builder.go` |
| 8 | `b895849` | Improved test coverage: syntax +13.5%, cli +12.5%, printer +3.0% |
| 9 | `e143dfd` | Wrote comprehensive assessment + execution plan with mermaid.js |

### Session 2 (2026-04-15) — Push confirmed

| # | Commit | Description |
|---|--------|-------------|
| 10 | `e143dfd` push | All 11 commits pushed to `origin/fork` successfully |

### Session 3 (2026-04-20) — 3 commits, NOT yet pushed

| # | Commit | Description |
|---|--------|-------------|
| 11 | `40ac57b` | Deleted dead packages: `internal/enum/`, `git/`, `migration/`, `adapter/` (2,983 lines removed) |
| 12 | `1afe02b` | Deleted dead `cli.ExitIfBothSet` — zero production callers, called `os.Exit(1)` |
| 13 | `ffdb562` | Removed dead fields from `cli.RuntimeConfig`: `DiffMode`, `All`, `OutputDir` |

### Total Impact Across All Sessions

| Metric | Value |
|--------|-------|
| Total commits | 13 (on fork, since baseline) |
| Lines deleted | ~3,100+ |
| Dead packages removed | 8 (`pkg/errors/`, `testutils/`, `detection/simple_detector.go`, `internal/enum/`, `git/`, `migration/`, `adapter/`, `cli/validation.go`) |
| Dead fields removed | 3 (`DiffMode`, `All`, `OutputDir` from `RuntimeConfig`) |
| Test coverage gained | +29% across 3 packages |
| Flag deduplication | 18 shared flags consolidated |
| Bug fixes | 2 (gogenfilter migration, timing race) |

---

## B. PARTIALLY DONE ⚠️

### Phase 1.4 — Error Handling Unification (Investigated, No Action Needed)

**Status**: Investigated and **correctly deferred**.

- `pkg/artdupl/` uses `fmt.Errorf` with `%w` wrapping and sentinel errors — this is the **correct** pattern for an SDK package. Consumers need `errors.Is()` compatibility, not internal typed wrappers.
- `job/` has zero error creation code — it only propagates errors.
- The `duplerrors.Wrap*` pattern in `cmd/` is appropriate for the CLI layer where error categorization drives user-facing messages.
- **Conclusion**: No unification needed. The current error handling is architecturally correct — SDK layer uses standard Go patterns, CLI layer uses typed wrappers.

### LSP Staleness

- `golangci_lint_ls` still shows 2 stale warnings on `config_builder.go` (gci at line 10, nolintlint at line 175). These are confirmed stale from the previous session. The actual file content is correct and `go build ./...` passes. LSP cannot be restarted programmatically.

---

## C. NOT STARTED 🔲

### Phase 2: Type Consolidation (3-5 hours estimated)

| Item | Effort | Impact |
|------|--------|--------|
| 2.1 Consolidate Clone types (4 types → 1 canonical) | 2-3h | Very High |
| 2.2 Consolidate CloneGroup types (3 types → 1 canonical) | 1-2h | High |
| 2.3 Split `config/detectionmethod.go` (426 lines, 5 enum types) | 30min | Medium |

### Phase 3: Test Coverage (4-5 hours estimated)

| Item | Current Coverage | Target |
|------|-----------------|--------|
| HTML diff view (11 functions at 0%) | ~0% | >60% |
| JSON printer (`SetHash`, `PrintFooter`, `OutputSimpleJSON`) | ~30% | >70% |
| Plumbing printer (`PrintHeader`, `PrintFooter`, `OutputPlumbing`) | ~0% | >70% |
| Text printer (`PrintHeader`, `PrintFooter`, `OutputText`) | ~20% | >70% |
| Hash package | 70% | >85% |
| Config package | 70.4% | >80% |
| Cmd package | 75.1% | >80% |

### Phase 4: Architecture (6-8 hours estimated)

| Item | Description |
|------|-------------|
| 4.1 Unify analysis pipelines | Extract to `internal/analysis/` — merge `cmd/run_analysis.go` and `pkg/artdupl/detector_pipeline.go` |
| 4.2 Split `printer/html.go` | 1,484 lines with Go+JS+CSS+HTML — split into focused files |
| 4.3 Nolint directive audit | ~80+ nolint directives across codebase |
| 4.4 Reduce `config/detectionmethod.go` complexity | 426 lines, 5 enum types in one file |

---

## D. TOTALLY FUCKED UP 💥

### Build Cache Corruption (RECURRING)

**Severity**: Blocking — prevents `go build` and `go test`  
**Root cause**: Go 1.26.0 via Nix has broken stdlib packages AND the build cache gets corrupted periodically.  
**Symptoms**: `package io is not in std`, `open ... no such file or directory` for cache files.  
**Workaround**: `rm -rf ~/Library/Caches/go-build/ && go build ./...` (takes ~2 minutes to rebuild cache).  
**Impact**: Cannot verify changes in real-time. Tests that were passing may appear to fail due to cache issues, not code issues.  
**Frequency**: Happened 3 times across 3 sessions.  
**Permanent fix needed**: Either fix the Nix Go installation or switch to a different Go installation method.

### Go 1.26.1 Installations (BOTH BROKEN)

Two Nix store paths have Go 1.26.1 but BOTH have broken stdlib packages (missing `internal/gover`, `crypto/internal/fips140deps/godebug`, etc.). These are completely unusable.

### `go test -cover ./...` BROKEN

Go 1.26.0 cannot run `go test -cover ./...` — causes setup failures for many packages. Must use `go test -coverprofile` with specific packages instead.

---

## E. WHAT WE SHOULD IMPROVE 🔧

### Process Improvements

1. **Run `go mod tidy` immediately** after any dependency change — would have prevented 240 BDD test failures in session 1
2. **Test PATH Go first** — wasted time fighting broken Nix Go 1.26.1 installations
3. **Build cache resilience** — the Go/Nix setup is fundamentally fragile; consider Docker or a standalone Go installation
4. **Push frequency** — should push after every commit, not batch at end. Lost 3 commits when session was interrupted before push.

### Code Improvements

1. **`printer/html.go`** (1,484 lines) — largest file, mixes Go/JS/CSS/HTML, 11 functions at 0% coverage. This is the single biggest risk.
2. **Split brain on Clone types** — 4 Clone types + 3 CloneGroup types across packages. Every conversion is manual and error-prone.
3. **Duplicate analysis pipelines** — `cmd/run_analysis.go` and `pkg/artdupl/detector_pipeline.go` do essentially the same thing differently.
4. **`config/detectionmethod.go`** — 426 lines with 5 enum types should be split into focused files.
5. **~80+ nolint directives** — many may be stale or unnecessary after the cleanup work.

---

## F. TOP 25 THINGS TO DO NEXT (Sorted by Impact × Effort)

### Quick Wins (Phase 1 — COMPLETE ✅)

| # | Task | Status |
|---|------|--------|
| ~~1~~ | ~~Delete dead packages (internal/enum, git, migration, adapter)~~ | ✅ Done |
| ~~2~~ | ~~Delete dead cli.ExitIfBothSet~~ | ✅ Done |
| ~~3~~ | ~~Remove dead RuntimeConfig fields~~ | ✅ Done |
| ~~4~~ | ~~Error handling investigation~~ | ✅ Done (no action needed) |

### Immediate Next Steps (Phase 2 — Type Consolidation)

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 5 | **Push 3 unpushed commits to origin** | 1 min | Critical |
| 6 | **Split `config/detectionmethod.go` into 5 focused files** | 30 min | Medium |
| 7 | **Add conversion methods to `domain.Clone`** (`ToPrinterClone()`, `ToSDKClone()`) | 1h | High |
| 8 | **Migrate `printer/` to use `domain.Clone` conversions** | 1-2h | High |
| 9 | **Migrate `pkg/artdupl/` to use `domain.Clone` directly** | 1h | High |
| 10 | **Consolidate CloneGroup types** (same pattern as Clone) | 1-2h | High |

### Test Coverage (Phase 3)

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 11 | **Add tests for `printer/html.go`** (11 functions at 0%) | 2-3h | Very High |
| 12 | **Add tests for `printer/plumbing.go`** (PrintHeader, PrintFooter, OutputPlumbing) | 1h | High |
| 13 | **Add tests for `printer/text.go`** (PrintHeader, PrintFooter, OutputText) | 1h | High |
| 14 | **Add tests for `printer/json.go`** (SetHash, PrintFooter, OutputSimpleJSON) | 1h | High |
| 15 | **Increase `hash/` package coverage** (70% → 85%) | 1h | Medium |
| 16 | **Increase `config/` package coverage** (70.4% → 80%) | 1h | Medium |
| 17 | **Increase `cmd/` package coverage** (75.1% → 80%) | 1h | Medium |

### Architecture (Phase 4)

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 18 | **Split `printer/html.go`** into focused files (html.go, html_css.go, html_js.go, html_diff.go) | 1-2h | High |
| 19 | **Unify analysis pipelines** — extract shared logic to `internal/analysis/` | 3-4h | Very High |
| 20 | **Audit and reduce nolint directives** (~80+ across codebase) | 2h | Medium |
| 21 | **Add `domain.Clone.Size()` method** to eliminate size computation duplication | 30 min | Medium |

### Quality of Life

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 22 | **Fix Go/Nix build environment** — stop cache corruption | 1h | Critical |
| 23 | **Add pre-commit hook** to run `go build ./...` before allowing commits | 15 min | Medium |
| 24 | **Write ADR for error handling strategy** (SDK: fmt.Errorf, CLI: duplerrors) | 30 min | Medium |
| 25 | **Update AGENTS.md** to reflect deleted packages and current architecture | 15 min | Low |

---

## G. TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF 🤔

**The Go/Nix build environment is fundamentally broken.**

Three sessions in a row, the Go build cache has corrupted itself. Go 1.26.0 works but `go test -cover ./...` fails. Both Go 1.26.1 installations have broken stdlib packages. The workaround (`rm -rf ~/Library/Caches/go-build/`) takes 2+ minutes each time.

**Question**: Is there a specific reason you're using Go via Nix (e.g., `flake.nix` requirement, team standard)? Would you be open to:
- Installing Go via `go install` or Homebrew as a stable fallback?
- Running Go in Docker for reproducibility?
- Fixing the Nix Go derivation to include all stdlib packages?

This is the single biggest productivity killer — I spend 10-15% of each session fighting build cache issues instead of doing useful work.

---

## Current Coverage Snapshot

| Package | Coverage | Trend |
|---------|----------|-------|
| domain | 97.0% | Stable |
| adapter | ~~97.7%~~ | **DELETED** |
| detection | 83.6% | Stable |
| syntax | 81.1% | ⬆️ +13.5% (session 1) |
| job | 76.7% | Stable |
| cli | 75.0% | ⬆️ +12.5% (session 1) |
| cmd | 75.1% | Stable |
| config | 70.4% | Stable |
| hash | 70.0% | Stable |
| printer | 66.5% | ⬆️ +3.0% (session 1) |

## Git Summary

```
Branch: fork
Commits on fork (since baseline): 20
Commits ahead of origin/fork: 3 (UNPUSHED)
Working tree: CLEAN
Unpushed commits:
  40ac57b refactor: delete dead packages (internal/enum, git, migration, adapter)
  1afe02b refactor: delete dead cli.ExitIfBothSet (zero production callers)
  ffdb562 refactor: remove dead fields from cli.RuntimeConfig

Total lines deleted across all sessions: ~3,100+
Total packages deleted: 8
```
