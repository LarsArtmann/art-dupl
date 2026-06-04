# art-dupl — Full Status Report

**Date:** 2026-06-04 23:25  
**Branch:** fork  
**HEAD:** `94db410` refactor(docs),chore(deps): rewrite AGENTS.md to concise enduring-context guide and bump gogenfilter v3.0.2 → v3.1.0

---

## Executive Summary

art-dupl is **healthy and fully functional**. All CI gates pass: build, test, lint (0 issues), fmt, race detector. The codebase is at **Health Score A** (0.69% duplication ratio). Today's session resolved two latent issues: a stale `vendorHash` blocking `nix build`, and a broken lint check that had never worked in the Nix sandbox.

---

## A) Fully Done ✅

### Core Product
- **All 7 output formats**: text, HTML, JSON, simple-JSON, plumbing, SARIF, CSV (stats)
- **Multi-method detection**: suffix tree (art-dupl), hash-based, both simultaneously
- **2 languages**: Go (45 node types), Templ (28 node types) — both fully functional
- **Smart filtering**: SQLC, templ, protobuf, mockgen, stringer, generic generated code
- **Statistics subcommand**: text, JSON, CSV with health grade (A-F)
- **SDK**: `pkg/artdupl/` with `Detector` interface, stream support, validation
- **Professional CLI**: Fang framework, shell completion, man pages

### Build & CI
- **Nix flake**: build, test, lint, fmt checks all green (`nix flake check` passes)
- **Go modules**: v1.26.3, all deps current
- **golangci-lint**: 0 issues
- **Race detector**: All tests pass with `-race`
- **BDD tests**: Ginkgo/Gomega suite green

### Today's Fixes (this session)
1. **vendorHash refreshed** — was stale after `8ef4740` updated transitive deps, `nix build` was broken
2. **Lint check fixed** — `pkgs.runCommand` → `pkg.overrideAttrs` (inherits vendor, GOCACHE), was never green in sandbox
3. **gogenfilter bumped** v3.0.2 → v3.1.0 — pin, vendorHash, go.mod, go.sum, flake.lock all updated
4. **AGENTS.md rewritten** — 883L monolith → 68L concise enduring-context guide

### Self-Analysis (art-dupl on itself)
| Metric | Value |
|--------|-------|
| Files scanned | 205 |
| Clone groups | 59 |
| Total clones | 147 |
| Duplication ratio | 0.69% |
| Health score | **A** |
| Analysis time | 70ms |

---

## B) Partially Done 🔧

| Item | Status | Details |
|------|--------|---------|
| **CSV clone output** | Stats CSV works; general clone CSV uses manual formatting, not `encoding/csv` | Low priority |
| **SIMD optimizations** | `internal/simd` deleted; 6 SIMD TODOs remain in `syntax/hash_simd.go` | Framework gone, items orphaned |
| **TODO Detector** | `TodoDetector` implemented in `detection/todos.go` but not exposed via CLI (`-m todo` doesn't work) | Code exists, no CLI wiring |
| **Legacy Detector** | `LegacyDetector` implemented but not exposed via CLI | Same as above |
| **Printer ↔ syntax.Node decoupling** | `domain.ProcessedClone` DTO exists, `Printer.PrintClones` still takes `[][]*syntax.Node` | 111 test call sites block migration |

---

## C) Not Started 📋

From TODO_LIST.md, never started:

| Item | Priority |
|------|----------|
| Introduce ProcessedClone DTO to decouple Printer from syntax.Node (111 test call sites) | HIGH |
| Consolidate three parallel Clone types | HIGH |
| Implement TokenValue type with validation | HIGH |
| Implement CSV output format properly using `encoding/csv` | MEDIUM |
| Unify enum patterns: domain enums should use config's generic helpers | MEDIUM |
| Decouple `printer/clone_classify.go` from `syntax/golang` direct import | MEDIUM |
| Add `--output-file` flag to stats subcommand | MEDIUM |
| Split `printer/stats_test.go` (975L → 3 files) | MEDIUM |
| Refactor `syntax/golang/transform.go` (369L, 300L switch statement) | LOW |
| Create `domain.HealthScore` typed enum | LOW |
| Write SDK documentation for `pkg/artdupl/` | LOW |
| Add BDD test for `stats --only templ` and `--only go` | LOW |
| Add BDD test for `--include-generic` end-to-end | LOW |
| Add fuzz tests for templ parser edge cases | LOW |
| Add ADR for semantic-as-default and reflection-based config merge | LOW |
| Validate GoReleaser release config | LOW |

---

## D) Totally Fucked Up 💥

| Issue | Severity | Details |
|-------|----------|---------|
| **Lint check was silently broken since creation** | HIGH | `e738896` (May 22) added lint check via `pkgs.runCommand` — never worked in sandbox (no vendor, no GOCACHE). CI was running lint via GitHub Actions (go, not nix), so CI was green. But `nix flake check` silently skipped lint or failed. **Fixed today.** |
| **vendorHash drifted for 3 commits** | MEDIUM | Commit `8ef4740` bumped transitive deps but didn't update `vendorHash`. `nix build` was broken from Jun 4 13:04 until today's fix. Nobody noticed because CI uses `go build`, not `nix build`. |
| **6 orphaned SIMD TODOs** | LOW | `internal/simd/` was deleted but 6 SIMD references remain in `syntax/hash_simd.go`. Dead code that will never be implemented. |
| **TODO/Legacy detectors: code exists, no CLI** | LOW | 2 detection methods fully implemented but users can't access them. `-m todo` and `-m legacy` would need enum registration + CLI wiring. |

---

## E) What We Should Improve 🎯

### Architecture
1. **Printer↔syntax.Node coupling** — the single biggest architectural debt. All 6 printer implementations depend on AST internals. `ProcessedClone` DTO exists but isn't used.
2. **Three parallel Clone types** — `printer.clone`, `printer.CloneGroup`/`printer.JSONClone`, `pkg/artdupl.Clone`. Consolidate after Printer DTO migration.
3. **`printer/clone_classify.go` imports `syntax/golang` directly** — language-specific coupling breaks multi-language ambition.

### Build & CI
4. **CI doesn't run `nix flake check`** — only runs `go test`, `golangci-lint`, `go build`. The Nix checks (including lint and fmt) are untested in CI. If vendorHash drifts again, CI won't catch it.
5. **`-race` requires CGO** — `CGO_ENABLED=0` in flake.nix means `nix` test check can't run race detector. Only runs locally with manual `CGO_ENABLED=1`.

### Code Quality
6. **`syntax/golang/transform.go`** — 369 lines with a 300-line switch statement. Hard to maintain, should use a lookup table.
7. **`printer/stats_test.go`** — 975 lines, should be split into 3 focused test files.
8. **Orphaned SIMD TODOs** — clean up or delete.

### Documentation
9. **AGENTS.md was an 883-line monolith** — fixed today (→68L), but the rewrite lost some detail. Verify nothing critical was dropped.
10. **SDK docs** — `pkg/artdupl/` has no dedicated documentation.
11. **ADR gaps** — No ADR for semantic-as-default decision or reflection-based config merge.

---

## F) Top 25 Things to Do Next

### 🔴 Critical (Architecture Debt)
| # | Item | Impact | Effort |
|---|------|--------|--------|
| 1 | Decouple Printer from syntax.Node via ProcessedClone DTO | Unblocks Clone consolidation, multi-language | High (111 test sites) |
| 2 | Consolidate three parallel Clone types into one | Eliminates confusion, reduces code | High (depends on #1) |
| 3 | Wire TODO and Legacy detectors to CLI (`-m todo`, `-m legacy`) | Users get 2 more detection methods for free | Low |
| 4 | Add `nix flake check` to CI workflow | Catches vendorHash drift, Nix sandbox issues | Low |
| 5 | Fix orphaned SIMD TODOs — delete or document as won't-fix | Removes dead code noise | Trivial |

### 🟡 Important (Quality & Maintainability)
| # | Item | Impact | Effort |
|---|------|--------|--------|
| 6 | Implement CSV output using `encoding/csv` | Proper RFC 4180 compliance | Medium |
| 7 | Refactor `syntax/golang/transform.go` (369L → lookup table) | Maintainability | Medium |
| 8 | Split `printer/stats_test.go` (975L → 3 files) | Test readability | Low |
| 9 | Decouple `printer/clone_classify.go` from `syntax/golang` | Multi-language readiness | Medium |
| 10 | Implement TokenValue type with validation | Type safety | Medium |
| 11 | Add `--output-file` flag to stats subcommand | User experience | Low |
| 12 | Unify enum patterns across domain and config | Consistency | Medium |

### 🟢 Nice to Have (Polish)
| # | Item | Impact | Effort |
|---|------|--------|--------|
| 13 | Write SDK documentation for `pkg/artdupl/` | Adoption | Medium |
| 14 | Add ADR: semantic-as-default decision | Decision record | Low |
| 15 | Add ADR: reflection-based config merge | Decision record | Low |
| 16 | Create `domain.HealthScore` typed enum | Type safety | Low |
| 17 | Add BDD test: `stats --only templ` and `stats --only go` | Coverage | Low |
| 18 | Add BDD test: `--include-generic` end-to-end | Coverage | Low |
| 19 | Add fuzz tests for templ parser edge cases | Robustness | Medium |
| 20 | Validate GoReleaser release config | Release reliability | Low |
| 21 | Fix remaining LSP hints (unused params, unnecessary type args) | Clean code | Trivial |
| 22 | Consider adding `--since` integration tests | Incremental feature confidence | Medium |
| 23 | Add `--diff-mode` (inline vs side-by-side) to HTML output | User experience | Low |
| 24 | Investigate `ConstantCSSProperty` Pos=0 upstream fix (a-h/templ) | Accuracy | External dep |
| 25 | Add memory allocation benchmarks for large repos | Performance visibility | Medium |

---

## G) Top #1 Question I Can't Figure Out Myself

**Should TODO/Legacy detectors be promoted to full CLI-accessible detection methods, or are they experimental prototypes that should be removed?**

The code in `detection/todos.go` is complete and tested — `TodoDetector` finds TODO/FIXME/HACK comments, `LegacyDetector` finds legacy patterns. Both are wired through `MultiDetector.FindDuplOver()` internally. But:
- They're marked `DEFINED_ONLY` in FEATURES.md
- The config enum `DetectionMethod` doesn't include `todo` or `legacy`
- No CLI flags expose them
- No BDD tests cover them as user-facing features

Are these intended as v3.0 features, or should they be cleaned up? This is a product decision that affects the TODO list prioritization.

---

## Build & Test Matrix

| Check | Status | Notes |
|-------|--------|-------|
| `nix build .#art-dupl` | ✅ PASS | Binary: 7.1MB, version 0.2.0 |
| `nix flake check` (build, test, lint, fmt) | ✅ PASS | All 6 derivations green |
| `go test ./...` | ✅ PASS | 25 packages, 0 failures |
| `go test -race ./...` | ✅ PASS | Requires `CGO_ENABLED=1` |
| `golangci-lint run` | ✅ PASS | 0 issues |
| `gofmt -l .` | ✅ PASS | No unformatted files |
| `nix develop` (devShell) | ✅ PASS | Go 1.26.3, golangci-lint 2.12.2, just 1.51.0 |

## Dependency Versions

| Dependency | Version | Notes |
|------------|---------|-------|
| Go | 1.26.3 | Aligned: go.mod, flake.nix, .golangci.yml |
| gogenfilter | v3.1.0 | Pinned at rev `8788d6c` |
| nixpkgs | `nixos-unstable` (May 31) | 4 days old |
| golangci-lint | 2.12.2 | Via nixpkgs |
| lipgloss/v2 | 2.0.3 | |
| cobra | 1.10.2 | |
| fang | 1.0.0 | |
| templ | 0.3.1020 | |
| ginkgo/v2 | 2.29.0 | |

---

_Generated by art-dupl session on 2026-06-04_
