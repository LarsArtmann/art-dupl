# Status Report: art-dupl — Comprehensive Audit & Roadmap

**Date:** 2026-04-30 21:29
**Branch:** fork (commit 358c317)
**Author:** Crush (GLM-5.1) + Lars

---

## Executive Summary

Deep audit of the art-dupl codebase after completing the Nix private dependency fix. All tests pass (24/24 packages), `go vet` clean, LSP info-only diagnostics. The codebase is in good shape overall but has significant technical debt in `printer/html.go` (1484 lines), missing CI, stale TODO list, and several type safety improvements available.

---

## A) FULLY DONE

| Item                      | Details                                                                          |
| ------------------------- | -------------------------------------------------------------------------------- |
| Nix flake private dep fix | Two-phase dummy/replace pattern with auto-synced go.mod via `builtins.readFile`  |
| Full Nix verification     | `nix build`, `nix flake check` (including test derivation), `nix run` all pass   |
| AGENTS.md updated         | Private dependency pattern documented with update workflow                       |
| All Go tests passing      | 24/24 packages PASS, 0 FAIL                                                      |
| `go vet` clean            | No issues                                                                        |
| Dead packages removed     | `lib/`, `migration/`, `adapter/` fully removed from source (refs only in docs)   |
| BDD test suite            | 16+ test files covering CLI, filtering, sorting, detection, config, plumbing     |
| Domain types              | Strong typing with value objects (`LineNumber`, `Threshold`, `CloneCount`, etc.) |
| Config enums              | Well-structured typed enums with generic helpers                                 |
| Error types               | Full `DuplError` hierarchy with typed constructors                               |
| go.mod clean              | Minimal direct deps (11), idiomatic Go 1.26.0                                    |

## B) PARTIALLY DONE

| Item                     | Current State                                 | Remaining                                                                                                                 |
| ------------------------ | --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| TODO_LIST.md             | Exists but stale (last updated 2026-04-05)    | References files that no longer exist (detector.go 546L, run.go 528L, stats.go 727L, clone.go 495L, domain_types.go 525L) |
| CI/CD                    | No `.github/workflows/` directory exists      | Need full CI pipeline                                                                                                     |
| Code quality enforcement | `.golangci.yml` is comprehensive (85 linters) | But never runs automatically                                                                                              |
| Type safety in domain    | Most value objects typed                      | Stringly-typed timestamps, `OutputFormat string`, `Only string` remain                                                    |
| Enum consistency         | Config enums use generic helpers              | Domain enums use switch-based `IsValid()` — inconsistent pattern                                                          |

## C) NOT STARTED

| Item                                                                 | Impact                                         | Effort |
| -------------------------------------------------------------------- | ---------------------------------------------- | ------ |
| CONTRIBUTING.md                                                      | Referenced in go.mod comment but doesn't exist | Low    |
| CI pipeline (GitHub Actions)                                         | No automated testing or building               | Medium |
| `printer/html.go` refactoring (1484 lines → multiple files)          | Highest code health improvement                | High   |
| Type safety: `time.Time` for timestamps                              | Prevents invalid date states                   | Low    |
| Type safety: `config.OutputFormat` for `domain.Options.OutputFormat` | Consistency                                    | Low    |
| Type safety: `config.FileType` for `config.Config.Only`              | Consistency                                    | Low    |
| `GetStatsData() any` → concrete return type                          | Type safety                                    | Medium |
| `fmt.Errorf` with `//nolint:err113` → typed errors                   | Lint compliance                                | Low    |
| `errors.As` → `errors.AsType` (3 instances in errors/types.go)       | Modernize per gopls hints                      | Low    |
| LSP hints: unnecessary type args, unused params                      | Code quality                                   | Low    |
| 304 status report files in docs/status/                              | Cleanup / archival                             | Low    |
| Justfile vendor hash update recipe                                   | Developer experience                           | Low    |
| SIMD TODOs (6 items)                                                 | Performance                                    | High   |

## D) TOTALLY FUCKED UP

| Item                 | Severity | Details                                   |
| -------------------- | -------- | ----------------------------------------- |
| **Nothing critical** | —        | All tests pass, code compiles, tool works |

Minor concerns:

- `docs/status/` has 304 files — excessive archive that slows searches
- `.golangci.yml` references non-existent `lib/` and `migration/` in exclusion rules
- TODO_LIST.md references files that were already split/refactored
- No CI means no automated quality gates on push

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **`printer/html.go` at 1484 lines** — the single biggest code health issue. Extract into:
   - `printer/html_template.go` (CSS/JS templates)
   - `printer/html_builder.go` (HTML generation)
   - `printer/html.go` (main struct + Print method)

2. **Stringly-typed timestamps** in domain:
   - `Analysis.CreatedAt string` → `time.Time`
   - `Analysis.CompletedAt *string` → `*time.Time`
   - `Repository.LastIndexed string` → `time.Time`
   - `SourceFile` fields using `string` where `Filepath`/`Hash` domain types exist

3. **Duplicate domain models**: `domain.Clone` (strong types) vs `artdupl.Clone` (primitives) — no explicit conversion layer

4. **Enum inconsistency**: Config enums use `enum_helpers.go` generics; domain enums use switch statements

### Type Safety

5. `domain.Options.OutputFormat` is `string` but should be `config.OutputFormat`
6. `config.Config.Only` is `string` but should be `config.FileType`
7. `GetStatsData() any` → return a concrete stats struct
8. `printer/format.go:28` and `printer/sort_type.go:26` use `//nolint:err113` — should use typed errors

### Developer Experience

9. No CI pipeline — zero automated quality gates
10. No CONTRIBUTING.md despite being referenced in go.mod
11. 304 status reports in `docs/status/` — needs archival policy
12. Justfile missing vendor hash update recipe
13. `.golangci.yml` references dead packages (`lib/`, `migration/`)

### Modernization

14. `errors.As` → `errors.AsType[*DuplError]` (3 instances, gopls hint)
15. Remove unnecessary type arguments (10+ instances, gopls hints)
16. Fix unused parameters in tests (gopls hints)

---

## F) Top #25 Things to Get Done Next

Sorted by **Impact × Effort** (high impact / low effort first):

| #  | Task                                                                     | Impact | Effort | Rationale                                 |
| -- | ------------------------------------------------------------------------ | ------ | ------ | ----------------------------------------- |
| 1  | Create CONTRIBUTING.md                                                   | Medium | Low    | Referenced in go.mod, quick win           |
| 2  | Fix `domain.Options.OutputFormat` → `config.OutputFormat`                | Medium | Low    | 1-line type fix, improves type safety     |
| 3  | Fix `config.Config.Only` → `config.FileType`                             | Medium | Low    | 1-line type fix                           |
| 4  | Use `time.Time` for domain timestamps                                    | Medium | Low    | Replace 3-4 string fields                 |
| 5  | Replace `errors.As` with `errors.AsType` (3 instances)                   | Low    | Low    | gopls modernization hint                  |
| 6  | Fix `//nolint:err113` — use typed errors                                 | Medium | Low    | 2 files, cleaner linting                  |
| 7  | Clean `.golangci.yml` dead package refs                                  | Low    | Low    | Remove `lib/`, `migration/` exclusions    |
| 8  | Update TODO_LIST.md with current state                                   | Medium | Low    | Currently stale, references deleted files |
| 9  | Add justfile recipe for vendor hash updates                              | Medium | Low    | Developer workflow improvement            |
| 10 | Archive old docs/status/ files (keep last 30 days)                       | Low    | Low    | 304 files → ~10                           |
| 11 | Fix LSP hints: unnecessary type args, unused params                      | Low    | Low    | 15+ instances in test code                |
| 12 | Add basic CI pipeline (build + test + lint)                              | High   | Medium | No automated quality gates                |
| 13 | Refactor `printer/html.go` (1484L → 3 files)                             | High   | Medium | Largest file, multiple 80+ line functions |
| 14 | Split `detection/todos.go` (TodoDetector + LegacyDetector)               | Medium | Low    | Two unrelated detectors in one file       |
| 15 | Add explicit conversion `domain.Clone` → `artdupl.Clone`                 | Medium | Medium | Bridge between internal/SDK types         |
| 16 | Unify enum patterns (domain → config generic helpers)                    | Medium | Medium | Consistency across codebase               |
| 17 | Make `GetStatsData()` return concrete type                               | Medium | Medium | Replace `any` with typed struct           |
| 18 | Add `Repository.Path` as `Filepath` domain type                          | Low    | Low    | Use existing domain type                  |
| 19 | Refactor `cmd/run_analysis.go` (447L)                                    | Medium | Medium | Multiple responsibilities                 |
| 20 | Refactor `printer/diff.go` (398L)                                        | Medium | Medium | Long functions                            |
| 21 | Split `config/config.go` (391L)                                          | Medium | Medium | Config + validation mixed                 |
| 22 | Type safety in `syntax/syntax.go:132` — TODO comment about int positions | Medium | Medium | Existing TODO acknowledged                |
| 23 | Refactor `syntax/golang/transform.go` (355L, 300L switch)                | Medium | High   | Giant switch statement                    |
| 24 | Implement SIMD TODOs (6 items)                                           | Medium | High   | Performance improvement                   |
| 25 | Create GitHub Actions CI with multi-OS matrix                            | High   | High   | Full CI coverage                          |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should we prioritize the CI pipeline (#12) or the html.go refactoring (#13)?**

- CI pipeline gives automated quality gates on every push but requires GitHub Actions config + potentially SSH key setup for the private gogenfilter dep
- html.go refactoring is the single biggest code health improvement (1484L file with 5 functions over 80 lines each) but is a large mechanical change

**My recommendation:** CI first. The codebase is stable (all tests pass, no vet issues). CI protects against regressions while we refactor. The html.go split can then be done with confidence that CI will catch breakage.

**Blocking question for CI:** Will GitHub Actions need SSH key access to clone the private gogenfilter dependency? If so, we need to set up a deploy key or PAT as a repository secret before CI will work.

---

## Verification Commands

```bash
# All passing as of this report:
nix build                                          # ✅
nix flake check                                    # ✅ (including test derivation)
nix run . -- --version                             # ✅ art-dupl version 358c317-dirty
go test -count=1 ./...                             # ✅ 24/24 packages PASS
go vet ./...                                       # ✅ clean
```

## Codebase Stats

- **Go packages:** 24
- **Go files:** ~210
- **Test files:** ~90
- **BDD test files:** 16
- **Lines of code:** ~3427 in the 6 largest files alone
- **Direct dependencies:** 11
- **Linters configured:** ~85
- **Status reports:** 304 (archival needed)
