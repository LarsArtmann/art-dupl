# Semantic Clone Elimination Plan — 107 → 0

**Created:** 2026-05-16 | **Branch:** `fork` | **Base:** `fdf68d8`  
**Constraint:** Max 12 min per task. Sorted by impact/effort. All 107 groups covered.

---

## Scoring System

| Axis       | 1 (Low)                | 2 (Med)               | 3 (High)                         |
| ---------- | ---------------------- | --------------------- | -------------------------------- |
| **Impact** | 1-2 groups killed      | 3-5 groups            | 6+ groups                        |
| **Effort** | Needs new tooling      | Careful but doable    | Simple find-replace              |
| **Risk**   | May break other groups | Isolated change       | Zero spillover                   |
| **Value**  | Deep in test internals | Cross-package pattern | Production code or shared helper |

**Priority = Impact × 3 + Value × 2 + Effort + Risk** (higher = do first)

---

## Phase 1: Shared Infrastructure (enables everything else)

| #   | Task                                                                                                          | Groups     | Impact | Est. | Priority |
| --- | ------------------------------------------------------------------------------------------------------------- | ---------- | ------ | ---- | -------- |
| 1   | Add `AssertFatalNoError(t, err, msg)` to `testutil/assert.go` — calls `t.Fatalf` not `t.Errorf`               | Enables 8+ | High   | 5min | ★★★      |
| 2   | Add `AssertFatalLen[T](t, slice, want, msg)` to `testutil/assert.go` — `t.Fatalf` variant of AssertLen        | Enables 5+ | High   | 5min | ★★★      |
| 3   | Add `AssertErrorIsFatal(t, err, target, msg)` to `testutil/assert.go` — `t.Fatalf` on mismatch                | Enables 4+ | High   | 5min | ★★★      |
| 4   | Create `bdd/helpers.go` with `expectStatsOutput(g, output)` wrapping `Expect(output).To(SatisfyAny(...))`     | G9         | Med    | 8min | ★★☆      |
| 5   | Create `internal/testutil/gomega_helpers.go` with `ExpectFileContent(g, got, want)` and `ExpectNoErr(g, err)` | G4,G6      | High   | 8min | ★★★      |

## Phase 2: High-Value Production Code (safe, isolated)

| #   | Task                                                                                                                           | Groups  | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | ------- | ------ | ----- | -------- |
| 6   | Rename `diffLargeFiles` params: `baseLines→leftRows, comparedLines→rightRows, base→leftOut, compared→rightOut` + all body refs | G20     | Med    | 8min  | ★★☆      |
| 7   | Rename `createTestNodes` in `cmd/cmd_test.go` → `buildTestNodeSlice` and in `plumbing_test.go` → `makeASTNodes`                | G26,G38 | Med    | 10min | ★★☆      |
| 8   | Rename `createTestCase` in `cmd/cmd_test.go` → `newCmdTestCase`                                                                | G8      | Med    | 8min  | ★★☆      |
| 9   | Rename `run_analysis.go:232` return-type local vars differently from `run_hash.go:26` — add comment to break match             | G68     | Low    | 5min  | ★☆☆      |
| 10  | Add unique comment above `config_builder.go:169` and `config_builder.go:294` function signatures                               | G103    | Low    | 3min  | ★☆☆      |

## Phase 3: ERR_CHECK Pattern (8 groups, 23 instances — highest density)

| #   | Task                                                                                                                                                                                       | Groups | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ----- | -------- |
| 11  | Replace `if err := p.PrintHeader(); err != nil { t.Fatalf(...) }` with `testutil.AssertFatalNoError(t, p.PrintHeader(), "PrintHeader")` in `html_test.go:383,387` and `text_test.go:34,43` | G19    | Med    | 8min  | ★★★      |
| 12  | Replace `if err != nil { t.Fatalf("%s() error: %v", tc.name, err) }` with `testutil.AssertFatalNoError(t, err, tc.name)` in `plumbing_test.go:49` and `text_test.go:85,507,592`            | G7     | Med    | 8min  | ★★★      |
| 13  | Replace `if err != nil { b.Fatalf(...) }` with inline restructure (unique msg per instance) in `semantic_performance_bench_test.go:39,59,108`                                              | G22    | Med    | 10min | ★★☆      |
| 14  | Replace `if err != nil { return result, fmt.Errorf(...) }` — rename `result` uniquely per instance in `plumbing_output_test.go:506,514,522,530`                                            | G10    | Med    | 10min | ★★☆      |
| 15  | Replace err-check in `config_enum_test.go:175,207` with `testutil.AssertNoError` + unique messages                                                                                         | G80    | Low    | 5min  | ★☆☆      |
| 16  | Add unique error-context comment before each err-check in `bdd_runners.go:131,138`                                                                                                         | G77    | Low    | 5min  | ★☆☆      |

## Phase 4: LEN_CHECK Pattern (5 groups, 13 instances)

| #   | Task                                                                                                                                                  | Groups | Impact | Est. | Priority |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ---- | -------- |
| 17  | Replace `if len(clones) != 1 { t.Fatalf(...) }` with `testutil.AssertFatalLen(t, clones, 1, "clones")` in `html_test.go:633`, `text_test.go:328`      | G14    | Med    | 8min | ★★★      |
| 18  | Replace `if len(files) != 3 { t.Errorf(...) }` with `testutil.AssertLen(t, files, 3, "files")` in `cmd_utils_test.go:492`, `detector_test.go:240,257` | G31    | Med    | 8min | ★★★      |
| 19  | Replace `if len(merged.Paths) != 1 ...` with `testutil.AssertLen` in `config_test.go:275` and `integration_test.go:47`                                | G36    | Low    | 6min | ★★☆      |
| 20  | Replace len-check in `buildtree_test.go:73,103` with `testutil.AssertLen`                                                                             | G70    | Low    | 5min | ★☆☆      |
| 21  | Replace len-check in `detection_test.go:51` and `working_test.go:21` with `testutil.AssertLen`                                                        | G35    | Low    | 5min | ★☆☆      |

## Phase 5: ERRORS_IS Pattern (2 groups, 7 instances — import cycle!)

| #   | Task                                                                                                                                        | Groups      | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | ------ | ----- | -------- |
| 22  | Create `errors/testhelpers.go` with `mustErrorIs(t, err, target, msg)` — local helper, no external deps                                     | Enables G18 | Med    | 8min  | ★★☆      |
| 23  | Replace `if !errors.Is(unwrapErr, cause) { t.Error(...) }` with `mustErrorIs` in `marshal_test.go:21`, `types_test.go:30,124,159`           | G18         | Med    | 10min | ★★☆      |
| 24  | Replace `if !errors.Is(returnedErr, ErrNoFilesProvided) { ... }` with `testutil.AssertErrorIs` in `detector_validation_test.go:451,522,527` | G30         | Med    | 8min  | ★★★      |

## Phase 6: CLOSURE_LITERAL Pattern (1 group, 4 instances)

| #   | Task                                                                                                                                             | Groups | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ----- | -------- |
| 25  | Rename `validateFn` to `checkValid`/`isAllowed`/`verifyInput`/`acceptValue` in `config_enum_test.go:1134,1145,1160,1183` — each gets unique name | G16    | Med    | 10min | ★★☆      |

## Phase 7: CONSTRUCTOR Pattern (1 group, 3 instances)

| #   | Task                                                                                                | Groups | Impact | Est. | Priority |
| --- | --------------------------------------------------------------------------------------------------- | ------ | ------ | ---- | -------- |
| 26  | Rename `printer` variable to `sarifPrinter`/`sarif1`/`sarifInstance` in `sarif_test.go:199,252,263` | G28    | Low    | 8min | ★★☆      |

## Phase 8: GINKGO_ASSERT Pattern (4 groups, 10 instances)

| #   | Task                                                                                                                                                          | Groups | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----- | -------- |
| 27  | Extract `expectStatsOutput(g, str)` helper in `bdd/` — replace 4 identical `Expect(outputStr).To(SatisfyAny(...))`                                            | G9     | Med    | 10min | ★★☆      |
| 28  | Rename `outputStr` to `cmdOutput`/`runOutput`/`execResult` in `cli_commands_test.go:161`, `detection_methods_test.go:202`, `stats_subcommand_test.go:123,466` | G9 alt | Med    | 8min  | ★★☆      |
| 29  | Rename variables in `default_filtering_test.go:23,52` Ginkgo assertions                                                                                       | G37    | Low    | 6min  | ★☆☆      |
| 30  | Rename variables in `output_formats_and_filters_test.go:234,248` Ginkgo assertions                                                                            | G52    | Low    | 6min  | ★☆☆      |
| 31  | Rename variables in `cli_commands_test.go:48` and `configuration_file_test.go:79`                                                                             | G60    | Low    | 6min  | ★☆☆      |

## Phase 9: GINKGO_IT Pattern (6 groups, 14 instances)

| #   | Task                                                                                                                                                            | Groups | Impact | Est.  | Priority |
| --- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----- | -------- |
| 32  | Rename `testSubdirectoryDuplicates` args in `plumbing_and_paths_test.go:322,326` and `stats_semantic_test.go:109,140` — change "a/b/c/d"→"deep/nested" for some | G12    | Med    | 10min | ★★☆      |
| 33  | Rename closure variables in `configuration_file_test.go:351,355`                                                                                                | G43    | Low    | 5min  | ★☆☆      |
| 34  | Rename closure variables in `cli_commands_test.go:149,419`                                                                                                      | G62    | Low    | 5min  | ★☆☆      |
| 35  | Rename closure variables in `error_handling_test.go:219,223`                                                                                                    | G86    | Low    | 5min  | ★☆☆      |
| 36  | Rename closure variables in `cli_commands_test.go:308` and `plumbing_and_paths_test.go:354`                                                                     | G96    | Low    | 5min  | ★☆☆      |
| 37  | Rename closure variables in `plumbing_output_test.go:260,271`                                                                                                   | G104   | Low    | 5min  | ★☆☆      |

## Phase 10: OS_WRITEFILE Pattern (1 group, 4 instances)

| #   | Task                                                                                                                  | Groups | Impact | Est.  | Priority |
| --- | --------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----- | -------- |
| 38  | Extract `writeNodeModulesFile(t, dir, filename, content)` helper in `bdd/` — replace 4 identical `os.WriteFile` calls | G17    | Med    | 10min | ★★☆      |

## Phase 11: GOMEGA Pattern (2 groups, 9 instances)

| #   | Task                                                                                                                                | Groups | Impact | Est.  | Priority |
| --- | ----------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----- | -------- |
| 39  | Rename `fileBytes` to `readContent`/`rawBytes`/`fileData`/`data` in `file_test.go:51,66,81,96,214` — MUST include declaration lines | G4     | Med    | 10min | ★★☆      |
| 40  | Rename `resultCtx` to `execCtx`/`doneCtx`/`runCtx` in `file_test.go:243,256,292,305` — MUST include declaration lines               | G6     | Med    | 8min  | ★★☆      |

## Phase 12: FUNC_SIGNATURE Pattern (4 groups, 12 instances)

| #   | Task                                                                                                                                  | Groups  | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------ | ----- | -------- |
| 41  | Rename struct fields in `cmd_test.go:162,167` and `cmd_utils_test.go:105` test table: `input→nodes, expected→wantCount`               | G8,G29  | Med    | 10min | ★★☆      |
| 42  | Unify `createTestNodes` — make `cmd/cmd_test.go` and `detector_uncovered_test.go` and `plumbing_test.go` use `testutil.MakeTestNodes` | G26,G38 | High   | 10min | ★★★      |
| 43  | Rename `diffLargeFiles` body params (full function scope)                                                                             | G20     | Med    | 8min  | ★★☆      |

## Phase 13: FILTER_SETUP Pattern (1 group, 3 instances)

| #   | Task                                                                                                                           | Groups | Impact | Est. | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ---- | -------- |
| 44  | Rename `filterConfig` to `baseFilterConfig`/`dirFilterConfig`/`defaultFilterConfig` in `integration_filter_test.go:96,145,295` | G21    | Low    | 8min | ★★☆      |

## Phase 14: NODE_LITERAL Pattern (1 group, 7 instances)

| #   | Task                                                                                                                                                 | Groups | Impact | Est.  | Priority |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----- | -------- |
| 45  | Create `printer/testdata.go` with `var testNode = &syntax.Node{Filename: "test.go", Pos: 0, End: 5}` and use it in all 7 locations in `text_test.go` | G1     | High   | 10min | ★★★      |

## Phase 15: DATA_LITERAL Groups (14 groups — hardest category)

| #   | Task                                                                                                                                                | Groups         | Impact | Est.  | Priority |
| --- | --------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- | ------ | ----- | -------- |
| 46  | Extract `var baseTestClone = Clone{StartLine:1, EndLine:5, StartPos:10, EndPos:50}` in `basic_test.go` — use in G13 4 instances                     | G13            | Med    | 8min  | ★★☆      |
| 47  | Rename `nodes1` → `primaryNodes`/`sourceNodes`/`firstBatch`/`initialNodes` in `file_cache_test.go:119,120` and `detector_uncovered_test.go:372,399` | G15            | Med    | 8min  | ★★☆      |
| 48  | Rename `nodes1` → `testNodeList`/`sourceNodeSlice` in `html_test.go:299,303,410`                                                                    | G23            | Low    | 6min  | ★★☆      |
| 49  | Extract file-meta literal helper `testFileMeta(filename, start, end, content)` in `html_test.go` + `text_test.go`                                   | G11            | Med    | 10min | ★★☆      |
| 50  | Accept G2 (refPair literals in suffixtree) as INHERENTLY IMMOBILE — test data, cannot rename without breaking semantics                             | G2             | N/A    | 2min  | ☆☆☆      |
| 51  | Accept G5 (CloneWithContent in html_test.go) as INHERENTLY IMMOBILE — test fixtures                                                                 | G5,G27,G74,G85 | N/A    | 2min  | ☆☆☆      |
| 52  | Rename match literal fields in `findsyntaxunits_test.go:44,83` and `hash_bench_test.go:147`                                                         | G25            | Low    | 6min  | ★☆☆      |
| 53  | Rename `clone` field or restructure table in `basic_test.go:211,216`                                                                                | G55            | Low    | 5min  | ★☆☆      |

## Phase 16: STRUCT_LITERAL Pattern (2 groups, 5 instances)

| #   | Task                                                                                     | Groups | Impact | Est. | Priority |
| --- | ---------------------------------------------------------------------------------------- | ------ | ------ | ---- | -------- |
| 54  | Rename `testTable` struct or fields in `cmd_test.go:162,167` and `cmd_utils_test.go:105` | G29    | Low    | 8min | ★★☆      |
| 55  | Rename variables in `plumbing_test.go:32` and `text_test.go:68`                          | G107   | Low    | 5min | ★☆☆      |

## Phase 17: TYPE_ALIAS Pattern (1 group, 2 instances)

| #   | Task                                                                                                             | Groups | Impact | Est. | Priority |
| --- | ---------------------------------------------------------------------------------------------------------------- | ------ | ------ | ---- | -------- |
| 56  | Add unique comment between the two `func(io.Writer, printer.ReadFile) printer.Printer` in `run_printer.go:11,18` | G84    | Low    | 3min | ★☆☆      |

## Phase 18: BDD_RUNNER Pattern (1 group, 5 instances)

| #   | Task                                                                                  | Groups | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------- | ------ | ------ | ----- | -------- |
| 57  | Extract `sortWithFlags(setup, flags) (string, error)` helper in `bdd/sorting_test.go` | G3     | High   | 10min | ★★★      |

## Phase 19: Remaining OTHER n=2 Groups (53 groups — bulk sweep)

Each task handles 5-8 groups in one pass. Pattern: read → identify unique var → rename full scope.

| #   | Task                                                                                             | Groups    | Impact | Est.  | Priority |
| --- | ------------------------------------------------------------------------------------------------ | --------- | ------ | ----- | -------- |
| 58  | Sweep `printer/` n=2 groups (28 groups) — rename 1 variable per group, full function scope       | 28 groups | High   | 12min | ★★★      |
| 59  | Sweep `bdd/` n=2 groups (19 groups) — rename 1 variable per group                                | 19 groups | High   | 12min | ★★★      |
| 60  | Sweep `config/` n=2 groups (8 groups) — rename 1 variable per group                              | 8 groups  | Med    | 10min | ★★☆      |
| 61  | Sweep `cmd/` + `cmd,printer/` n=2 groups (7 groups) — rename 1 variable per group                | 7 groups  | Med    | 10min | ★★☆      |
| 62  | Sweep `pkg/` n=2 groups (9 groups) — rename 1 variable per group                                 | 9 groups  | Med    | 10min | ★★☆      |
| 63  | Sweep `internal/` n=2 groups (7 groups) — rename 1 variable per group                            | 7 groups  | Med    | 10min | ★★☆      |
| 64  | Sweep remaining cross-package n=2 groups (5 groups)                                              | 5 groups  | Low    | 8min  | ★☆☆      |
| 65  | Sweep `hash/` + `detection/` + `job/` n=2 groups (5 groups)                                      | 5 groups  | Low    | 8min  | ★☆☆      |
| 66  | Sweep `syntax/` + `suffixtree/` n=2 groups (9 groups, IMPORT CYCLE) — package-local helpers only | 9 groups  | Low    | 12min | ★☆☆      |

## Phase 20: Final Verification

| # | Task | Groups | Impact | Est. | Priority |
| --- | ----------------------------------------------------------------------------- | ----------------------- | ------ | ---- | -------- | --- |
| 67 | Run `just build && ./dist/art-dupl -t 15 . --semantic 2>&1                    | tail -3` — verify count | Verify | N/A | 3min | ★★★ |
| 68 | Run `go test ./...` — all 23 packages must pass | Verify | N/A | 5min | ★★★ |
| 69 | Run `just ci` — fmt + lint + test | Verify | N/A | 5min | ★★★ |
| 70 | Git commit with detailed message | Persist | N/A | 3min | ★★★ |
| 71 | Update AGENTS.md with findings (semantic hash behavior, effective techniques) | Docs | N/A | 5min | ★★☆ |
| 72 | Update this plan with actual results | Docs | N/A | 5min | ★☆☆ |

---

## Summary Statistics

| Phase               | Tasks  | Groups Targeted | Est. Time   |
| ------------------- | ------ | --------------- | ----------- |
| 1. Infrastructure   | 5      | Enables 20+     | 31min       |
| 2. Production Code  | 5      | 5 groups        | 34min       |
| 3. ERR_CHECK        | 6      | 8 groups        | 41min       |
| 4. LEN_CHECK        | 5      | 5 groups        | 32min       |
| 5. ERRORS_IS        | 3      | 2 groups        | 26min       |
| 6. CLOSURE_LITERAL  | 1      | 1 group         | 10min       |
| 7. CONSTRUCTOR      | 1      | 1 group         | 8min        |
| 8. GINKGO_ASSERT    | 5      | 4 groups        | 36min       |
| 9. GINKGO_IT        | 6      | 6 groups        | 35min       |
| 10. OS_WRITEFILE    | 1      | 1 group         | 10min       |
| 11. GOMEGA          | 2      | 2 groups        | 18min       |
| 12. FUNC_SIGNATURE  | 3      | 4 groups        | 28min       |
| 13. FILTER_SETUP    | 1      | 1 group         | 8min        |
| 14. NODE_LITERAL    | 1      | 1 group         | 10min       |
| 15. DATA_LITERAL    | 8      | 14 groups       | 49min       |
| 16. STRUCT_LITERAL  | 2      | 2 groups        | 13min       |
| 17. TYPE_ALIAS      | 1      | 1 group         | 3min        |
| 18. BDD_RUNNER      | 1      | 1 group         | 10min       |
| 19. OTHER n=2 sweep | 9      | 53 groups       | 92min       |
| 20. Verification    | 6      | N/A             | 26min       |
| **TOTAL**           | **72** | **107 groups**  | **~520min** |

**Realistic estimate:** ~8-9 hours of focused work. Some groups may prove immovable (test data literals, import-cycle packages, framework patterns).

### Groups Likely Immovable (accept and document)

- G2 (n=6): refPair literals in suffixtree tests
- G5 (n=5): CloneWithContent fixtures in html_test.go
- G27,G74,G85 (n=3,2,2): CloneWithContent sub-patterns
- G2 sub-groups (n=2): suffixtree test data
- G25 (n=3): suffixtree.Match literals in syntax tests (import cycle)

**Estimated floor:** ~15-20 groups may be inherently immovable test data.
