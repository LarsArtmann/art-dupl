# Status Report: Quality Sprint Complete — Next Phase Planning

**Date:** 2026-04-30 22:40
**Branch:** fork (commit 40c6c5b)
**Author:** Crush (GLM-5.1) + Lars
**State:** Clean working tree, all tests pass (24/24), build OK

---

## A) FULLY DONE

### Infrastructure & Build

<!-- check-rows:ledger -->
| Item                        | Commit    | Details                                            |
| --------------------------- | --------- | -------------------------------------------------- |
| Nix flake private dep fix   | `652b325` | Two-phase dummy/replace pattern for gogenfilter    |
| Auto-synced dummy go.mod    | `358c317` | `builtins.readFile` reads real go.mod at eval time |
| Vendor/ removed from git    | `3cf84f5` | Added to .gitignore                                |
| CONTRIBUTING.md created     | `6d1bcfe` | Full contributor guide                             |
| TODO_LIST.md updated        | `6d1bcfe` | Removed stale refs, added current items            |
| Execution plan with mermaid | `bcf9632` | docs/planning/2026-04-30_21-39_...                 |
| AGENTS.md updated           | `e13997b` | Architecture decisions documented                  |

### Error Handling Modernization

<!-- check-rows:ledger -->
| Item                            | Commit    | Details                                                         |
| ------------------------------- | --------- | --------------------------------------------------------------- |
| `errors.As` → `errors.AsType`   | `4498343` | 3 instances in errors/types.go (Go 1.24+ idiom)                 |
| All `//nolint:err113` removed   | `4498343` | Replaced with typed errors (EnumValidationError, sentinel + %w) |
| Dead `migration/` rules removed | `4498343` | 6 blocks from .golangci.yml                                     |

### Type Safety Improvements

<!-- check-rows:ledger -->
| Item                                  | Commit    | Details                                                     |
| ------------------------------------- | --------- | ----------------------------------------------------------- |
| `Config.Only` → `config.FileType`     | `8aae71c` | Used existing FileType enum, eliminated matchesOnlyFilter() |
| `Analysis.CreatedAt` → `time.Time`    | `d0e5a2d` | String → time.Time, validation uses IsZero()                |
| `Analysis.CompletedAt` → `*time.Time` | `d0e5a2d` | *string → *time.Time                                        |
| Unused `DetectionOptions` removed     | `40c6c5b` | Dead code cleanup                                           |

### Structural Improvements

<!-- check-rows:ledger -->
| Item                        | Commit    | Details                                     |
| --------------------------- | --------- | ------------------------------------------- |
| `printer/html.go` split     | `24d1902` | 1484L → 4 files (364L + 523L + 315L + 301L) |
| Test constant extraction    | `b9e33dc` | Repeated string literals → named constants  |
| Error check on json.Marshal | `b9e33dc` | Was discarding error in test                |

### From Parallel Sessions (not by this conversation)

<!-- check-rows:ledger -->
| Item                              | Commit                          | Details                            |
| --------------------------------- | ------------------------------- | ---------------------------------- |
| gogenfilter updated to v2.1       | `2a2e9f7`                       | Dependency update                  |
| SortBy unified with SortCriteria  | `eb30754`                       | Type alias pattern                 |
| Flag-to-config mapping simplified | `528432c`                       | ReportMetadata derived from Config |
| Site improvements                 | `a78c31a`, `4b600b2`, `345670c` | Landing page, visualizations       |
| Comprehensive lint fixes          | `6cc7256`                       | goconst, unused, nolintlint, etc.  |
| Go formatting applied             | `1533da5`                       | Consistent across codebase         |

---

## B) PARTIALLY DONE

<!-- check-rows:ledger -->
| Item                  | Current State                                 | Remaining                                                                       |
| --------------------- | --------------------------------------------- | ------------------------------------------------------------------------------- |
| CI pipeline           | `.github/workflows/` exists with 5 workflows  | `art-dupl.yml` uses `go-version: "1.26rc2"` — should be `stable`                |
| Type safety in domain | Threshold, LineNumber, TokenCount, etc. typed | Repository.Path, SourceFile.Path still `string`                                 |
| Enum consistency      | Config uses generic helpers                   | Domain enums still use switch statements                                        |
| LSP diagnostics       | Build passes, go vet clean                    | gopls shows stale "duplicate" errors from html.go split (cache issue, not real) |

---

## C) NOT STARTED

### High Impact, Low Effort

| # | Item                                                     | Effort | Impact                |
| - | -------------------------------------------------------- | ------ | --------------------- |
| ~~1~~ done — moot — `GlobalPool` removed from the codebase | ~~Fix `GlobalPool()` data race (`sync.OnceValue`)~~ | ~~5min~~ | ~~High — race condition~~ |
| ~~2~~ done — `strings.HasSuffix` in `config/filetype.go` (dead helper removed at 5342b858) | ~~Replace hand-rolled `hasSuffix` with `strings.HasSuffix`~~ | ~~2min~~ | ~~Low — code clarity~~ |
| ~~3~~ done — workflows pin `GO_VERSION`/`go-version-file` (`.github/workflows/`) | ~~Fix CI `go-version: "1.26rc2"` → `stable`~~ | ~~1min~~ | ~~Medium~~ |
| ~~4~~ done — moot — `Repository` type no longer exists | ~~Type `Repository.Path` as `domain.Filepath`~~ | ~~15min~~ | ~~Medium~~ |
| ~~5~~ done — moot — `GetStatsData` removed; `buildJSONData` returns concrete `jsonStatsOutput` | ~~Fix `GetStatsData() any` → concrete return type~~ | ~~20min~~ | ~~Medium~~ |
| ~~6~~ done — `errors.Join` used for multi-error paths (e.g. `internal/testutil/capture.go`) | ~~Use `errors.Join` for multi-rule validation~~ | ~~10min~~ | ~~Low~~ |

### Medium Impact, Medium Effort

| #  | Item                                                          | Effort | Impact              |
| -- | ------------------------------------------------------------- | ------ | ------------------- |
| ~~7~~ done — moot — `NodeToClone` helper removed | ~~Move `NodeToClone` out of `domain/` (break domain→syntax dep)~~ | ~~30min~~ | ~~High — architecture~~ |
| ~~8~~ done — no setter methods; concrete `stats` API in `printer/stats/stats_formatter.go` | ~~Simplify `StatsPrinter` interface (8 setters → option struct)~~ | ~~30min~~ | ~~Medium — API design~~ |
| ~~9~~ done — `printer/sort_unified.go` uses `slices.SortFunc` | ~~`sort.Interface` → `slices.SortFunc`~~ | ~~15min~~ | ~~Low — modernize~~ |
| ~~10~~ done — `pkg/enum/enum.go` + `config/enum_helpers.go` | ~~Unify enum patterns (domain → config generic helpers)~~ | ~~60min~~ | ~~Medium~~ |
| ~~11~~ done — moot — `detection/todos.go` removed | ~~Split `detection/todos.go` (TodoDetector + LegacyDetector)~~ | ~~20min~~ | ~~Low~~ |
| ~~12~~ done — now 337L with `config_io.go`/`config_validate.go` extracted | ~~Split `config/config.go` (391L)~~ | ~~30min~~ | ~~Low~~ |
| ~~13~~ done — SIMD design documented in `hash/` notes + ADR-0019/0022 | ~~Resolve or document 6 SIMD TODOs~~ | ~~30min~~ | ~~Low~~ |

### High Impact, High Effort

<!-- check-rows:ledger -->
| #  | Item                                                           | Effort | Impact |
| -- | -------------------------------------------------------------- | ------ | ------ |
| ~~14~~ **Superseded — ADR-0005 split-brain type unification.** | ~~Evaluate `domain.Clone` — is it dead code in runtime pipeline?~~ | ~~60min~~ | ~~High~~ |
| 15 | Split `syntax/golang/transform.go` (355L, 300L switch)         | 120min | Medium |
| 16 | Split `cmd/run_analysis.go` (447L)                             | 60min  | Medium |
| ~~17~~ done — `suffixtree/suffixtree.go` `TokenValue` | ~~Implement TokenValue type (TODO #1 in TODO_LIST.md)~~ | ~~180min~~ | ~~High~~ |
| ~~18~~ **Superseded — docs-health archival pass 2026-11 (378 files now in `docs/status/archived/`).** | ~~Archive 304 old docs/status/ files~~ | ~~15min~~ | ~~Low~~ |

---

## D) TOTALLY FUCKED UP

**Nothing.** The codebase is in good shape:

- All 24 packages compile and pass tests
- `go vet` is clean
- `nix build` and `nix flake check` pass
- No build-breaking issues

### Known Quirks (not broken, just annoying)

- **gopls cache** shows phantom "duplicate declaration" errors in `printer/html.go` vs `html_diff.go`/`html_summary.go`. This is a gopls caching issue — `go build` compiles cleanly. Restarting gopls resolves it.
- **LSP hints** (info-level): unnecessary type args, unused params in tests — cosmetic only.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture Debt

1. ~~**`domain/` imports `syntax/`** — violates clean architecture. Domain should be innermost layer, unaware of infrastructure. `domain/conversion.go` should move to a `convert/` or `adapter/` package.~~ done — moot — `domain/` no longer imports `syntax/` in non-test code

2. ~~**Three parallel `Clone` types** — `domain.Clone` (strong types), `artdupl.Clone` (SDK primitives), `printer.clone` (internal). Domain.Clone is never used in the runtime pipeline (SDK converts directly from syntax.Node). This is either dead code that should be removed, or the pipeline should flow through domain types.~~ **Superseded — ADR-0005 split-brain type unification.**

3. ~~**Triple threshold validation** — domain, config, and SDK each validate threshold independently with different rules. Should consolidate to `domain.NewThreshold()` as single source of truth.~~ **Superseded — ADR-0005 split-brain type unification.**

4. ~~**`StatsPrinter` has 8 setter methods** — replace with `SetStats(StatsConfig)` option struct.~~ done — no setter methods; concrete `stats` API in `printer/stats/stats_formatter.go`

### Type Safety Gaps

5. ~~`Repository.Path` is `string` but `Filepath` type exists in domain~~ done — moot — `Repository` type removed
6. ~~`SourceFile.Path` is `string` but `Filepath` type exists~~ done — moot — `SourceFile` type removed
7. ~~`GetStatsData()` returns `any` — should return concrete type~~ done — moot — `GetStatsData` removed
8. ~~`buildJSONData()` returns `any` — should return `jsonStatsOutput`~~ done — `buildJSONData` returns concrete `jsonStatsOutput`
9. ~~Domain enums use switch-based `IsValid()` — should use config's generic map-based pattern~~ done — `pkg/enum/enum.go` + `config/enum_helpers.go`

### Concurrency Bug

10. ~~**`GlobalPool()` has a data race** — `if globalPool != nil` check before `sync.Once.Do` is unsynchronized. Use `sync.OnceValue` (Go 1.21+).~~ done — moot — `GlobalPool` removed

### Minor Code Quality

11. ~~`hasSuffix()` in `config/filetype.go` reinvents `strings.HasSuffix` — unnecessary~~ done — `strings.HasSuffix` in `config/filetype.go`
12. ~~`sort.Interface` pattern in `printer/common.go` — use `slices.SortFunc` + `cmp.Compare`~~ done — `printer/sort_unified.go` uses `slices.SortFunc`
13. ~~`validateFields` in `domain/validation.go` returns only first error — use `errors.Join`~~ done — `errors.Join` idioms in place; `domain/validation.go` removed

---

## F) Top #25 Things to Get Done Next

Sorted by impact × effort (highest ROI first):

<!-- check-rows:ledger -->
| #  | Task                                                           | Impact | Effort | ROI   |
| -- | -------------------------------------------------------------- | ------ | ------ | ----- |
| ~~1~~ done — moot — `GlobalPool` removed | ~~Fix `GlobalPool()` data race with `sync.OnceValue`~~ | ~~High~~ | ~~5min~~ | ~~★★★★★~~ |
| ~~2~~ done — workflows pin `GO_VERSION`/`go-version-file` | ~~Fix CI go-version `1.26rc2` → `stable`~~ | ~~Medium~~ | ~~1min~~ | ~~★★★★★~~ |
| ~~3~~ done — `strings.HasSuffix` in `config/filetype.go` | ~~Replace `hasSuffix` with `strings.HasSuffix`~~ | ~~Low~~ | ~~2min~~ | ~~★★★★~~ |
| ~~4~~ done — `errors.Join` idioms in place; `domain/validation.go` removed | ~~Use `errors.Join` for `validateFields`~~ | ~~Medium~~ | ~~10min~~ | ~~★★★★~~ |
| ~~5~~ done — moot — types removed | ~~Type `Repository.Path` and `SourceFile.Path` as `Filepath`~~ | ~~Medium~~ | ~~15min~~ | ~~★★★★~~ |
| ~~6~~ done — moot — `SourceFile` type removed | ~~Type `SourceFile.Hash` as `domain.Hash`~~ | ~~Medium~~ | ~~10min~~ | ~~★★★★~~ |
| ~~7~~ done — moot — `GetStatsData` removed | ~~Fix `GetStatsData()` return type~~ | ~~Medium~~ | ~~20min~~ | ~~★★★~~ |
| ~~8~~ done — returns concrete `jsonStatsOutput` (`printer/stats/stats_formatter.go:412`) | ~~Fix `buildJSONData()` return type~~ | ~~Medium~~ | ~~15min~~ | ~~★★★~~ |
| ~~9~~ done — option-based API in `printer/stats/stats_formatter.go` | ~~Simplify `StatsPrinter` with option struct~~ | ~~Medium~~ | ~~30min~~ | ~~★★★~~ |
| ~~10~~ done — moot — `NodeToClone` removed | ~~Move `NodeToClone` out of `domain/`~~ | ~~High~~ | ~~30min~~ | ~~★★★~~ |
| ~~11~~ done — moot — `detection/todos.go` removed | ~~Split `detection/todos.go` into separate files~~ | ~~Low~~ | ~~20min~~ | ~~★★~~ |
| ~~12~~ done — `printer/sort_unified.go` uses `slices.SortFunc` | ~~Modernize `sort.Interface` → `slices.SortFunc`~~ | ~~Low~~ | ~~15min~~ | ~~★★~~ |
| ~~13~~ **Superseded — ADR-0005 split-brain type unification.** | ~~Evaluate and document `domain.Clone` dead code question~~ | ~~High~~ | ~~60min~~ | ~~★★~~ |
| ~~14~~ **Superseded — ADR-0005 split-brain type unification.** | ~~Consolidate threshold validation to single source~~ | ~~Medium~~ | ~~30min~~ | ~~★★~~ |
| ~~15~~ done — `pkg/enum/enum.go` + `config/enum_helpers.go` | ~~Unify enum patterns (domain → config generics)~~ | ~~Medium~~ | ~~60min~~ | ~~★★~~ |
| ~~16~~ done — now 337L with extracted modules | ~~Split `config/config.go` (391L)~~ | ~~Low~~ | ~~30min~~ | ~~★★~~ |
| ~~17~~ **Superseded — docs-health archival pass 2026-11.** | ~~Archive old docs/status/ (304 files → ~10)~~ | ~~Low~~ | ~~15min~~ | ~~★~~ |
| ~~18~~ done — SIMD design documented in `hash/` notes + ADR-0019/0022 | ~~Resolve/document 6 SIMD TODOs~~ | ~~Low~~ | ~~30min~~ | ~~★~~ |
| 19 | Split `cmd/run_analysis.go` (447L)                             | Low    | 60min  | ★     |
| 20 | Split `syntax/golang/transform.go` (355L switch)               | Low    | 120min | ★     |
| ~~21~~ done — `suffixtree/suffixtree.go` `TokenValue` | ~~Implement TokenValue type (TODO #1 in TODO_LIST)~~ | ~~High~~ | ~~180min~~ | ~~★~~ |
| ~~22~~ **Won't implement — Nix flake + BuildFlow are the adopted tooling; no justfile.** | ~~Add justfile recipe for vendor hash updates~~ | ~~Medium~~ | ~~15min~~ | ~~★~~ |
| ~~23~~ done at `c0261563` | ~~Fix remaining LSP hints (unnecessary type args, unused params)~~ | ~~Low~~ | ~~20min~~ | ~~★~~ |
| ~~24~~ done — moot — `Repository` type removed | ~~Type `Repository.LastIndexed` as `time.Time`~~ | ~~Low~~ | ~~10min~~ | ~~★~~ |
| ~~25~~ **Superseded — ADR-0005 split-brain type unification.** | ~~Investigate `domain.Clone` → `artdupl.Clone` conversion layer~~ | ~~Medium~~ | ~~60min~~ | ~~★~~ |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should `domain.Clone` / `domain.CloneGroup` be removed as dead code, or should the runtime pipeline be refactored to flow through them?**

Evidence that they're dead code in the runtime:

- The SDK (`pkg/artdupl/detector_conversion.go`) converts `syntax.Node` → `artdupl.Clone` directly
- The printer uses its own internal `clone` struct
- `domain.Clone` is only used in `domain/conversion.go` (NodeToClone) and domain tests
- The runtime pipeline never touches domain.Clone

However:

- They represent the "ideal" type-safe model
- They could serve as the canonical type if we refactor the pipeline
- Removing them feels like discarding good architecture work

**This is a product/architecture decision that depends on the project's direction.** If the project is moving toward a clean architecture with domain as the core, we should refactor the pipeline to use domain types. If the project prioritizes shipping with the current SDK-focused approach, we should remove the unused domain types to reduce cognitive load.

---

## Verification

```bash
go test -count=1 ./...   # 24/24 PASS
go build ./...           # BUILD OK
go vet ./...             # clean
git status               # clean working tree
```

## Session Stats

- **Commits by this session:** 8 (4498343 through 40c6c5b)
- **Commits by parallel sessions:** 12 (6cc7256 through 365f338)
- **Files modified:** ~30 unique files
- **Lines changed:** ~+600/-200 net (excluding auto-formatting)
- **Tests:** 24/24 passing, 0 failures
- **Build:** Clean (go build, nix build, nix flake check)
