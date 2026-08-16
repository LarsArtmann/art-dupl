# Status Report — 2026-05-06 10:26

## Session Focus

Fix `*_templ.go` generated files passing through unfiltered, and clean up the filtering pipeline architecture.

## a) FULLY DONE

### 1. Root cause identified and fixed — `*_templ.go` filtering default

**Commit:** `45da310` — `fix: revert templ filtering default — *_templ.go files filtered again`

Commit `e9e65a5` flipped `IncludeTempl` default from `false` to `true`, making `*_templ.go` generated files pass through unfiltered. The commit conflated two concerns:

- `.templ` **source** files (user-written, should always be analyzed)
- `*_templ.go` **generated** files (templ compiler output, should be filtered by default)

The fix:

- `IncludeTempl` default: `true` → `false` (filtered by default, like sqlc/protobuf)
- Renamed `--exclude-templ` → `--include-templ` (matches `--include-sqlc`, `--include-protobuf` pattern)
- `.templ` source files remain always included via `isSourceFile()` regardless of this flag
- `*_templ.go` generated files are now filtered by default via `gogenfilter.FilterTempl`

### 2. Config merge bug fixed — missing fields

**Commit:** `cb6289f` — `fix: add missing IncludeProtobuf/IncludeMockgen/IncludeStringer to config merge`

`IncludeProtobuf`, `IncludeMockgen`, `IncludeStringer` were silently ignored when set in JSON config files. Only CLI flags could set them. Added all three to `mergeConfig()` with the standard bool pattern.

### 3. Dead `--filter-generated` flag removed

**Commit:** `f91292f` — `refactor: remove dead --filter-generated flag`

`FilterGenerated` was stored in config but **never read by `setupFilter()`**. It was only used for an HTML badge (cosmetic). All generated code filtering is already controlled by individual `--include-<generator>` flags. Removed from:

- `Config` struct
- `FlagValues` struct
- `cmd/flags.go`
- `cmd/config_builder.go`
- `config/config_merge.go`
- `printer.ReportMetadata`
- `printer/html.go` badge rendering
- All unit tests (`cmd/cmd_test.go`, `printer/html_test.go`)

### 4. BDD tests updated (9 files)

**Commit:** `d32b64e` — `test(bdd): update tests to reflect *_templ.go filtering default`

Updated all BDD tests across 9 files:

| File                                             | Changes                                                                                                                       |
| ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| `bdd/default_filtering_test.go`                  | Removed `--exclude-templ` from 4 test invocations, renamed test names, added `--include-templ` for inclusion tests            |
| `bdd/filter_features_test.go`                    | Removed `--filter-generated` from 3 invocations, removed `--exclude-templ` from 1, added `--include-templ` for inclusion test |
| `bdd/stats_command_test.go`                      | Updated helper calls, renamed tests, added `--include-templ` for inclusion tests                                              |
| `bdd/stats_subcommand_test.go`                   | Removed `filter-generated` DescribeTable entry                                                                                |
| `bdd/plumbing_output_test.go`                    | Removed `--filter-generated` flag, renamed test                                                                               |
| `bdd/templ_clone_detection_test.go`              | Removed `--exclude-templ`, renamed test, updated comments                                                                     |
| `bdd/output_formats_and_filters_test.go`         | Removed `--filter-generated` from 2 invocations, renamed 2 tests                                                              |
| `bdd/configuration_file_test.go`                 | Removed `filterGenerated` from JSON configs, changed `excludeTempl` to `includeTempl`                                         |
| `internal/filtertest/integration_filter_test.go` | Renamed sub-test                                                                                                              |

### 5. All tests pass — 22/22 suites

```
ok  github.com/LarsArtmann/art-dupl/bdd
ok  github.com/LarsArtmann/art-dupl/cache
ok  github.com/LarsArtmann/art-dupl/cmd
ok  github.com/LarsArtmann/art-dupl/config
ok  github.com/LarsArtmann/art-dupl/detection
ok  github.com/LarsArtmann/art-dupl/errors
ok  github.com/LarsArtmann/art-dupl/examples
ok  github.com/LarsArtmann/art-dupl/hash
ok  github.com/LarsArtmann/art-dupl/internal/configtest
ok  github.com/LarsArtmann/art-dupl/internal/filtertest
ok  github.com/LarsArtmann/art-dupl/internal/simd
ok  github.com/LarsArtmann/art-dupl/internal/utils
ok  github.com/LarsArtmann/art-dupl/job
ok  github.com/LarsArtmann/art-dupl/pkg/artdupl
ok  github.com/LarsArtmann/art-dupl/pkg/format
ok  github.com/LarsArtmann/art-dupl/pkg/logger
ok  github.com/LarsArtmann/art-dupl/pkg/position
ok  github.com/LarsArtmann/art-dupl/printer
ok  github.com/LarsArtmann/art-dupl/suffixtree
ok  github.com/LarsArtmann/art-dupl/syntax
ok  github.com/LarsArtmann/art-dupl/syntax/golang
ok  github.com/LarsArtmann/art-dupl/syntax/templ
```

Zero failures. Build clean. Zero remaining references to `--exclude-templ`, `--filter-generated`, `filterGenerated`, or `FilterGenerated`.

### 6. SARIF hash fix (pre-existing, caught during session)

**Commit:** `a098597` — `fix: SARIF output now uses real clone hashes instead of position strings`

### 7. Semantic mode improvements (pre-existing, caught during session)

**Commit:** `88b05a5` — `feat(semantic): add interface context tracking for semantic mode`
**Commit:** `0f96ed7` — `fix(semantic): encode interface field types as distinct semantic tokens`

---

## b) PARTIALLY DONE

Nothing — all tasks started were completed.

---

## c) NOT STARTED

### Architecture improvements identified but not implemented:

1. **Printer ↔ syntax.Node coupling** — `Printer.PrintClones(dups [][]*syntax.Node)` forces all 6 implementations to depend on AST internals. Should introduce `ProcessedClone` DTO.
2. **Three parallel Clone types** — `printer.clone`, `pkg/artdupl.Clone`, and different `CloneGroup` types should be consolidated.
3. **gogenfilter v3 upgrade** — Current `v0.2.1` pseudo-version. v3 adds FilterGoEnum, FilterOapi, FilterDeepcopy, FilterWire, FilterMoq.
4. **AGENTS.md update** — Should reflect all filtering pipeline changes from this session.

---

## d) TOTALLY FUCKED UP (honest self-review)

1. **First attempt thrashed** — Made partial edits across many files without compiling between steps, chased stale LSP errors, mixed production and test changes in the same batch. Had to `git checkout -- .` and start over.
2. **Didn't verify before editing** — On the first attempt, I renamed `IncludeTempl` to `IncludeTemplGenerated` which created cascading LSP errors across the entire codebase. The simpler approach (keep the field name, just change the default) was obviously better.
3. **Should have planned before executing** — I jumped into editing before fully understanding the scope of changes needed across BDD tests. The agent-based analysis phase was good, but I should have completed the full plan before making the first edit.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements:

1. **Always build after every edit batch** — Not just at the end. Each logical change should compile before moving to the next.
2. **Plan ALL edits before making the first one** — Read all files, map all changes, then execute. Don't discover dependencies mid-edit.
3. **Don't trust LSP diagnostics** — They can be stale. Use `go build ./...` as the source of truth.
4. **Commit after each self-contained change** — Smaller commits are easier to review and revert.

### Code improvements identified:

1. **`setupFilter()` always creates a `gogenfilter.Filter`** — Even when no filter options are set (empty `filterOptions` + no patterns). The `NewFilter()` call at the bottom creates a no-op filter. Should return `nil` when no filtering is needed.
2. **Bool flag merge pattern** — `applyBooleanFlags` only sets `true` values via the loop. The special-case handling for `IncludeTempl` (now removed) was a band-aid. Consider tracking which flags were explicitly set.
3. **`cmd/config_builder.go` still has `cfg.IncludeTempl = flags.IncludeTempl`** — This unconditional assignment is no longer needed now that `IncludeTempl` defaults to `false` like all others. The standard loop handles it correctly.

---

## f) Top 25 Things to Do Next

| #  | Priority | Task                                                                                                                                                                  | Impact                 | Effort |
| -- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- | ------ |
| 1  | P0       | Update AGENTS.md with filtering pipeline changes                                                                                                                      | Docs accuracy          | LOW    |
| 2  | P0       | Remove unnecessary `cfg.IncludeTempl = flags.IncludeTempl` workaround in config_builder.go                                                                            | Code cleanliness       | LOW    |
| 3  | P1       | Update HOW_TO_USE.md to reflect `--include-templ` instead of `--exclude-templ`                                                                                        | User-facing docs       | LOW    |
| 4  | P1       | Update README.md flag examples                                                                                                                                        | User-facing docs       | LOW    |
| 5  | P1       | Upgrade gogenfilter to v3 — gain FilterGoEnum, FilterOapi, FilterDeepcopy, FilterWire, FilterMoq                                                                      | Better detection       | MED    |
| 6  | P1       | Add `--include-go-enum`, `--include-oapi`, etc. flags for new gogenfilter detectors                                                                                   | Feature parity         | MED    |
| 7  | P2       | Return `nil` from `setupFilter()` when no filter options are set                                                                                                      | Performance            | LOW    |
| 8  | P2       | Printer DTO refactor — introduce `ProcessedClone` to decouple from AST                                                                                                | Architecture           | HIGH   |
| 9  | P2       | Consolidate three Clone types (`printer.clone`, `pkg/artdupl.Clone`)                                                                                                  | Code health            | HIGH   |
| 10 | P2       | Move `printer/clone_classify.go` language-specific constants out of printer                                                                                           | Language extensibility | MED    |
| 11 | P2       | Add integration test for `art-dupl --semantic .` on real project                                                                                                      | Confidence             | MED    |
| 12 | P3       | Fix `templ_clone_detection_test.go` "exclude .templ files" test — it tests `.templ` source exclusion but `--include-templ` only controls `*_templ.go` generated files | Test accuracy          | LOW    |
| 13 | P3       | Consider `--only templ` flag for filtering to `.templ` source files only                                                                                              | UX completeness        | LOW    |
| 14 | P3       | Add flag description consistency check to CI                                                                                                                          | Prevent future drift   | LOW    |
| 15 | P3       | Track which flags were explicitly set in config_builder (replace bool loop pattern)                                                                                   | Correctness            | MED    |
| 16 | P4       | Add `--filter-generic` flag for `gogenfilter.FilterGeneric` (catches any "Code generated by" comment)                                                                 | Feature                | LOW    |
| 17 | P4       | Fix `sqlc.yaml` auto-detection — flag description mentions it but no runtime detection exists                                                                         | Honest UX              | MED    |
| 18 | P4       | Add E2E test that runs `art-dupl` on its own source code and validates output                                                                                         | Meta-testing           | MED    |
| 19 | P4       | Nix flake: update gogenfilter rev after v3 upgrade                                                                                                                    | Build system           | LOW    |
| 20 | P4       | Consider renaming `IncludeTempl` to `IncludeTemplGenerated` for clarity                                                                                               | Naming                 | LOW    |
| 21 | P5       | SARIF: add fingerprinting for GitHub CodeQL integration quality                                                                                                       | Security tooling       | MED    |
| 22 | P5       | Stats subcommand: add filter reason breakdown (how many templ/sqlc/protobuf filtered)                                                                                 | Observability          | MED    |
| 23 | P5       | Add `--filter-report` flag to output what was filtered and why                                                                                                        | Debug UX               | MED    |
| 24 | P5       | Incremental analysis: wire filter through cache invalidation                                                                                                          | Performance            | HIGH   |
| 25 | P5       | Add fuzz tests for gogenfilter integration edge cases                                                                                                                 | Robustness             | MED    |

---

## g) Top #1 Question I Cannot Figure Out Myself

**The `templ_clone_detection_test.go` line 243 test:** It creates `display.templ` (a `.templ` SOURCE file) and expects it to NOT appear in output. But `--include-templ` only controls `*_templ.go` GENERATED files — `.templ` source files are always included by `isSourceFile()`. The test passes because `display.templ` content has no duplicate (single file, no match), not because of any filtering.

**Question:** Is this test intentionally testing that `.templ` source files with no duplicates don't appear in output (trivially true), or was it meant to test `*_templ.go` generated file exclusion but accidentally creates `.templ` source files instead? Should I fix the test to actually create `*_templ.go` files and test generated code exclusion, or update the test name to accurately describe what it tests?

---

## Commits This Session (newest first)

```
0f96ed7 fix(semantic): encode interface field types as distinct semantic tokens
88b05a5 feat(semantic): add interface context tracking for semantic mode
1df70a0 style: fix alignment in config/config.go
d32b64e test(bdd): update tests to reflect *_templ.go filtering default
a098597 fix: SARIF output now uses real clone hashes instead of position strings
f91292f refactor: remove dead --filter-generated flag
cb6289f fix: add missing IncludeProtobuf/IncludeMockgen/IncludeStringer to config merge
45da310 fix: revert templ filtering default — *_templ.go files filtered again
```

## Build & Test Status

- **Build:** Clean (`go build ./...` passes)
- **Tests:** 22/22 suites pass
- **Lint:** Not run this session
- **Uncommitted changes:** 1 file (`syntax/golang/transform.go` — blank line addition, trivial)
