# Comprehensive Status Report — 2026-04-30

**Generated:** 2026-04-30 22:08 CEST
**Branch:** fork (1 commit ahead of origin/fork)
**Go Version:** 1.26.0
**Total Packages:** 28 (all passing)
**Total Source Files:** ~236 Go files, ~24,000+ lines
**Nix Flake:** All checks passed

---

## a) FULLY DONE ✓

### Dependency Update — gogenfilter v2.1
- **gogenfilter** updated from pseudo-version `v0.1.1-0.20260424213812-5957230e34ed` → `v0.2.1-0.20260430195342-edf3d8d84a2a` (tag `v2.1`, commit `edf3d8d`)
- `flake.nix` rev updated, `vendorHash` regenerated (`sha256-Zri2zWjmhMbBGe9pt88H3XkKh8RZmc5DqigBE7nCL/s=`)
- `flake.lock` updated (gogenfilter input rev 5957230e → edf3d8d, revCount 249 → 254)
- `go.mod`, `go.sum`, `vendor/` all regenerated
- Build: ✓ | Tests: ✓ | Nix flake check: ✓

### Recent Session Accomplishments (last 20 commits)
| Commit | Description |
|--------|-------------|
| `6b3aabf` | Extract inline CSS/JS to external files for caching |
| `528432c` | Simplify flag-to-config mapping, derive ReportMetadata from Config |
| `eb30754` | Unify printer.SortBy with config.SortCriteria via type alias |
| `6cc7256` | Resolve all lint warnings (goconst, unused, nolintlint, tparallel, golines, revive, nlreturn, prealloc) |
| `b9e33dc` | Extract test constants and improve error handling |
| `358c317` | Auto-sync dummy go.mod from gogenfilter at eval time |
| `652b325` | Eliminate vendor/ requirement for private gogenfilter dependency |
| `6658fa9` | Replace local fileStat with FileStatMixin |
| `9bfa62f` | Unify SARIFRegion with LineRangeMixin |
| `81aafb7` | Add Firebase hosting and architecture documentation |
| `6ded3b5` | Add landing page for art-dupl |
| `6e0e47f` | Extract helper methods to reduce duplication |
| `b457ff9` | Apply consistent formatting to flake.nix |
| `028c420` | Add Nix flake support for reproducible builds |

### Test Coverage (Fresh Run)
| Package | Coverage |
|---------|----------|
| cli | 100.0% |
| pkg/format | 100.0% |
| pkg/position | 100.0% |
| domain | 97.0% |
| hash | 96.7% |
| internal/simd | 95.8% |
| config | 93.9% |
| syntax/golang | 93.9% |
| suffixtree | 91.0% |
| internal/utils | 93.3% |
| errors | 89.4% |
| pkg/logger | 87.5% |
| cache | 87.4% |
| printer | 85.9% |
| syntax/templ | 85.3% |
| pkg/artdupl | 86.4% |
| detection | 83.6% |
| syntax | 81.1% |
| job | 76.7% |
| cmd | 76.3% |
| bdd | 70.0% |
| internal/filtertest | 50.0% |
| cmd/art-dupl | 0.0% (main package) |
| examples | 0.0% (examples) |

---

## b) PARTIALLY DONE ⚠️

### Landing Page / Site
- `site/index.html` exists with OG image meta tags, Twitter card updated to `summary_large_image`
- `site/og-image.svg`, `site/robots.txt`, `site/sitemap.xml` created but **unstaged**
- `.github/workflows/deploy-site.yml` created but **unstaged**
- Firebase hosting configured but deployment workflow not yet tested through CI

### SIMD Optimization
- Skeleton code in `internal/simd/simd.go` and `syntax/hash_simd.go`
- 7 TODOs for actual SIMD implementation (ARM64, x86)
- Build tags set (`goexperiment.simd`) but no real vectorized code yet

### Documentation Cleanup
- ~300+ status reports in `docs/status/` (many are historical, could be archived)
- FEATURES.md last updated 2026-02-12 (may be stale)
- TODO_LIST.md last updated 2026-04-05 (partially addressed by recent refactors)

---

## c) NOT STARTED ○

From TODO_LIST.md and FEATURES.md:
1. **TokenValue type with validation** (HIGH priority) — type-safe token representation
2. **Proper CSV output** via `encoding/csv` (MEDIUM) — current CSV is manual string building
3. **README update** with semantic detection defaults (MEDIUM)
4. **`--profile` flag implementation** — listed as experimental, no code
5. **`--timeout` flag implementation** — listed as experimental, no code
6. **Godoc on all exported symbols** (LOW)
7. **File splitting** — `printer/html.go` (1484 lines), `printer/stats.go` (727 lines), `domain/clone.go` (495 lines)
8. **GitHub Release automation** — no release workflow, no goreleaser
9. **Cross-platform binary distribution** — no CI artifacts published
10. **Real SIMD implementations** for hot hash paths

---

## d) TOTALLY FUCKED UP 💥

### Minor Issues
1. **Unstaged vendor/ directory** — `vendor/` is in `.gitignore`-like limbo; tracked by git but appears as untracked. Need to decide: commit vendor/ or .gitignore it properly.
2. **1 gci formatting issue** in `cmd/config_builder.go:214` — not critical but should be fixed.
3. **gogenfilter tag `v2.1`** — The module path is `github.com/LarsArtmann/gogenfilter` (no `/v2` suffix), but the tag is `v2.1`. Go's semver treats this as a major version 2 module without the required `/v2` path suffix. This works now via pseudo-version but could break if Go tooling enforces strict semver in the future. **Recommendation:** Retag as `v0.2.1` or add `/v2` to the module path in gogenfilter.

---

## e) WHAT WE SHOULD IMPROVE 📈

### Architecture
- **`printer/html.go` at 1484 lines** — single biggest file, should be split into sub-packages or at least 3-4 files
- **Type safety in `syntax/syntax.go:132`** — documented TODO for int-based positions, should use typed integers
- **300+ status reports** in `docs/status/` — archive old ones, keep only recent
- **`examples/` package has 0% coverage** — either test them or remove

### Process
- **No CI release pipeline** — nix builds work locally but no artifact publishing
- **FEATURES.md stale** — last updated Feb 2026, 3 months behind
- **No changelog** — CHANGELOG.md doesn't exist
- **Site deployment untested** — deploy workflow created but never triggered

### Code Quality
- **`internal/filtertest` at 50% coverage** — below project standard
- **`bdd` at 70% coverage** — could benefit from more scenario coverage
- **7 SIMD TODOs** — either implement or remove the dead scaffolding
- **Phantom types / panic conditions** — identified in previous audit (545 phantom type issues, 532 panic conditions)

---

## f) Top 25 Things We Should Get Done Next

### Critical (P0)
1. Fix gogenfilter semver tagging (`v2.1` → `v0.2.1` or module path `/v2`)
2. Commit and push current work (gogenfilter update + site improvements)
3. Fix `cmd/config_builder.go:214` gci formatting issue
4. Add `vendor/` to `.gitignore` or properly track it

### High (P1)
5. Split `printer/html.go` (1484 lines) into 3-4 focused files
6. Implement `TokenValue` type with validation (TODO_LIST HIGH)
7. Update `FEATURES.md` to reflect current state (3 months stale)
8. Create `CHANGELOG.md` with recent release history
9. Add GitHub Actions release workflow (goreleaser or nix-based)
10. Test Firebase deploy workflow end-to-end

### Medium (P2)
11. Update README with semantic detection defaults
12. Implement proper CSV output via `encoding/csv`
13. Split `printer/stats.go` (727 lines)
14. Split `domain/clone.go` (495 lines)
15. Increase `internal/filtertest` coverage above 80%
16. Increase `bdd` coverage above 80%
17. Archive old `docs/status/` reports (>30 days)
18. Remove or implement SIMD scaffolding (7 TODOs)

### Low (P3)
19. Add godoc to all exported symbols
20. Implement `--profile` flag
21. Implement `--timeout` flag
22. Add cross-platform binary builds to CI
23. Add `examples/` tests or convert to documentation
24. Audit and fix phantom type issues (545 identified)
25. Audit and eliminate unnecessary panic conditions (532 identified)

---

## g) Top #1 Question

**What is the intended future of the `vendor/` directory?**

Currently `vendor/` exists locally (populated by `go mod vendor`) and is shown as an untracked directory in git status. The Nix build uses a custom vendor strategy (dummy replace → real gogenfilter swap). The Go toolchain build (`go build ./...`) uses vendor directly. This creates a split:

- If we `.gitignore` vendor/ → Go builds require `go mod vendor` first, Nix builds unaffected
- If we commit vendor/ → reproducible builds without network, but large diffs on dep updates
- Current state → untracked, best of neither world

**Which approach should we standardize on?**

---

## Build & Test Summary

| Check | Status |
|-------|--------|
| `go build ./...` | ✓ PASS |
| `go test ./...` | ✓ 28/28 packages pass |
| `nix build` | ✓ PASS |
| `nix flake check` | ✓ All checks passed |
| Test coverage | 85%+ average across core packages |
| Lint (gci) | ⚠ 1 issue in cmd/config_builder.go |

---

## Uncommitted Changes (Current Working Tree)

| File | Status | Description |
|------|--------|-------------|
| `flake.nix` | Modified | gogenfilter rev + vendorHash update |
| `flake.lock` | Modified | gogenfilter input lock update |
| `go.mod` | Modified | gogenfilter pseudo-version update |
| `go.sum` | Modified | gogenfilter checksums |
| `site/index.html` | Modified | OG image meta tags + Twitter card upgrade |
| `.github/workflows/deploy-site.yml` | Untracked | Firebase deploy workflow |
| `site/og-image.svg` | Untracked | Open Graph image |
| `site/robots.txt` | Untracked | SEO robots file |
| `site/sitemap.xml` | Untracked | SEO sitemap |
| `vendor/` | Untracked | Go module vendor directory |
