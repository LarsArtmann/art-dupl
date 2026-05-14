# Comprehensive Execution Plan — art-dupl

**Generated:** 2026-05-14 09:52 UTC
**Clone groups:** 125 remaining | **Go LOC:** 45,035 | **Tests:** 22/22 pass

---

## Scoring Formula

`Score = (Importance × Impact × CustomerValue × 10) / EffortMinutes`

| Dimension         | Range   | Meaning                             |
| ----------------- | ------- | ----------------------------------- |
| **Importance**    | 1-10    | Architectural criticality           |
| **Impact**        | 1-10    | Maintainability/cleanup improvement |
| **CustomerValue** | 1-10    | Direct user-facing benefit          |
| **Effort**        | ≤12 min | Max 12 minutes per task             |

---

## All Tasks Sorted by Score

### PHASE 1: Critical Architecture (Score 400-1000)

| #   | Task                                                                               | File(s)                           | Clone Group             | I×Im×CV | Effort | Score    | Deps   |
| --- | ---------------------------------------------------------------------------------- | --------------------------------- | ----------------------- | ------- | ------ | -------- | ------ |
| 1   | **Define ProcessedClone DTO type** — filename, lineStart/End, fragment, hash, size | `domain/processed_clone.go` (new) | architectural           | 800     | 10m    | **800**  | —      |
| 2   | **Change Printer interface**: `PrintClones(groups []ProcessedCloneGroup) error`    | `printer/printer.go`              | 6-file PrintClones loop | 800     | 8m     | **1000** | #1     |
| 3   | **Create NodeToProcessed conversion function** — extract from `ProcessNodeRange()` | `printer/sorter.go` or `syntax/`  | 6-file PrintClones loop | 720     | 10m    | **720**  | #1     |
| 4   | **Convert TextPrinter** to use `ProcessedCloneGroup`                               | `printer/text.go:73`              | 6-file PrintClones loop | 450     | 12m    | **450**  | #2, #3 |
| 5   | **Convert HTMLPrinter** to use `ProcessedCloneGroup`                               | `printer/html.go:138`             | 6-file PrintClones loop | 450     | 12m    | **450**  | #2, #3 |
| 6   | **Convert JSONPrinter** to use `ProcessedCloneGroup`                               | `printer/json.go:113`             | 6-file PrintClones loop | 450     | 12m    | **450**  | #2, #3 |
| 7   | **Convert SARIFPrinter** to use `ProcessedCloneGroup`                              | `printer/sarif.go:157`            | 6-file PrintClones loop | 450     | 12m    | **450**  | #2, #3 |
| 8   | **Convert PlumbingPrinter** to use `ProcessedCloneGroup`                           | `printer/plumbing.go:24`          | 6-file PrintClones loop | 450     | 12m    | **450**  | #2, #3 |
| 9   | **Convert StatsPrinter** to use `ProcessedCloneGroup`                              | `printer/stats.go:91`             | 6-file PrintClones loop | 450     | 12m    | **450**  | #2, #3 |
| 10  | **Update callers** (cmd/run_printer.go, cmd/run_analysis.go)                       | `cmd/run_*.go`                    | 6-file PrintClones loop | 486     | 10m    | **486**  | #2-#9  |
| 11  | **Consolidate pkg/artdupl.Clone** to embed ProcessedClone                          | `pkg/artdupl/`                    | 3 parallel Clone types  | 486     | 10m    | **486**  | #1     |
| 12  | **Consolidate printer.CloneGroup** to use ProcessedClone                           | `printer/`                        | 3 parallel Clone types  | 450     | 8m     | **450**  | #1     |
| 13  | **Define ProcessedCloneGroup type** — slice of ProcessedClone                      | `domain/`                         | architectural           | 600     | 8m     | **600**  | #1     |
| 14  | **Extract SortNodesByCriteria + PrepareClonesInfo shared logic**                   | `printer/sorter.go`               | sorter.go:70,86,96,125  | 213     | 12m    | **213**  | —      |
| 15  | **Extract getSortKey helper** for 4 sort functions                                 | `printer/sorter.go`               | sorter.go:70,86,96,125  | 123     | 12m    | **123**  | —      |

### PHASE 2: Type Safety & Core Refactoring (Score 80-500)

| #   | Task                                                                          | File(s)                                                            | Clone Group                             | I×Im×CV | Effort | Score   | Deps     |
| --- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------ | --------------------------------------- | ------- | ------ | ------- | -------- |
| 16  | **Multi-language architecture** — define LanguageParser interface             | `syntax/`                                                          | parse.go/parser.go                      | 504     | 10m    | **504** | —        |
| 17  | **Implement TokenValue type** with `<, ==` methods                            | `domain/`, `suffixtree/`, `syntax/`                                | architectural                           | 473     | 12m    | **473** | —        |
| 18  | **Update text printer tests** (111 call sites)                                | `printer/text_test.go`                                             | 6-file PrintClones loop                 | 213     | 12m    | **213** | #2-#9    |
| 19  | **Update html printer tests**                                                 | `printer/html_test.go`                                             | 6-file PrintClones loop                 | 213     | 12m    | **213** | #2-#9    |
| 20  | **Update json/sarif/plumbing/stats printer tests**                            | `printer/*_test.go`                                                | 6-file PrintClones loop                 | 213     | 12m    | **213** | #2-#9    |
| 21  | **Extract generic detector base** (TodoDetector + LegacyDetector)             | `detection/`                                                       | issue_helpers.go:57, legacy:30, todo:40 | 147     | 10m    | **147** | —        |
| 22  | **Extract regex compile pattern** (2 identical blocks)                        | `detection/todo_detector.go:33`, `detection/legacy_detector.go:25` | —                                       | 126     | 8m     | **126** | —        |
| 23  | **Add domain type tests** (Filepath, LineNumber validation)                   | `domain/*_test.go` (new)                                           | 0% coverage                             | 140     | 10m    | **140** | —        |
| 24  | **Add Clone.IsValid() tests + CloneSeverity tests**                           | `domain/*_test.go`                                                 | 0% coverage                             | 140     | 10m    | **140** | #23      |
| 25  | **Implement CSV output via encoding/csv**                                     | `printer/csv.go` or `printer/stats.go`                             | CSV manual formatting                   | 125     | 12m    | **125** | —        |
| 26  | **Refactor syntax/golang/transform.go** 300L switch                           | `syntax/golang/transform.go`                                       | transform.go 355L                       | 105     | 12m    | **105** | —        |
| 27  | **Wire TodoDetector + LegacyDetector to CLI** — add `-m todo` and `-m legacy` | `cmd/`, `detection/`                                               | DEFINED_ONLY status                     | 100     | 10m    | **100** | —        |
| 28  | **Unify enum patterns** — domain enums use config ParseEnum/MarshalJSON       | `domain/`, `config/`                                               | enum inconsistencies                    | 90      | 12m    | **90**  | —        |
| 29  | **Performance benchmarks** for large repo scanning                            | `*_bench_test.go` or `benchmark/`                                  | no baselines                            | 83      | 12m    | **83**  | —        |
| 30  | **Fix domain/ package 0% test coverage**                                      | `domain/*_test.go`                                                 | 0% coverage                             | 82      | 12m    | **82**  | #23, #24 |
| 31  | **String interning for identifiers**                                          | `syntax/` or `suffixtree/`                                         | memory optimization                     | 75      | 10m    | **75**  | —        |
| 32  | **Enhanced error context** in job/ and detection/                             | `job/`, `detection/`                                               | error wrapping gaps                     | 75      | 10m    | **75**  | —        |

### PHASE 3: Test Code Deduplication (Score 9-50)

| #   | Task                                                             | File(s)                                               | Clone Group                            | I×Im×CV | Effort | Score  | Deps |
| --- | ---------------------------------------------------------------- | ----------------------------------------------------- | -------------------------------------- | ------- | ------ | ------ | ---- |
| 33  | **Extract assertFieldEquals[T]** helper for detector types tests | `pkg/artdupl/detector_types_test.go`                  | 13-clone group:139-141,172-174,229-231 | 22      | 12m    | **22** | —    |
| 34  | **Extract assertFieldEquals for detector_uncovered_test.go**     | `pkg/artdupl/detector_uncovered_test.go`              | 106-108,160-162                        | 24      | 10m    | **24** | #33  |
| 35  | **Extract assertFieldEquals for detector_validation_test.go**    | `pkg/artdupl/detector_validation_test.go`             | 268-270,362-364,406-408                | 24      | 10m    | **24** | #33  |
| 36  | **Table-drive bdd/sorting_test.go** RunArtDuplWithFlags calls    | `bdd/sorting_test.go`                                 | 5 clones:98-265                        | 17      | 12m    | **17** | —    |
| 37  | **Extract assertThresholdEquals for config_enum_test.go**        | `config/config_enum_test.go`                          | 3 clones:139-141,172-174,229-231       | 17      | 12m    | **17** | —    |
| 38  | **Extract assertConfigFieldEquals for config_test.go**           | `config/config_test.go`                               | 81-88                                  | 20      | 10m    | **20** | —    |
| 39  | **Extract assertErrorContains helper** for 9-file clone          | `cmd/`, `config/`, `job/`, `pkg/position/`, `syntax/` | 9 clones:36-38,63-65,89-91             | 18      | 10m    | **18** | —    |
| 40  | **Extract assertHTMLOutputContains helper**                      | `printer/html_test.go`                                | 4 clones:897,950 + text:450,451        | 17      | 12m    | **17** | —    |
| 41  | **Extract assertBufferEmpty/Contains helpers**                   | `printer/*_test.go`, `cmd/`                           | 6 clones:text_utils:19-67,cmd:236      | 20      | 10m    | **20** | —    |
| 42  | **Extract assertCloneValid helper** for basic_test.go            | `pkg/artdupl/basic_test.go`                           | 208,212,227,234                        | 20      | 10m    | **20** | —    |
| 43  | **Extract assertFilterResult helper** for filtertest             | `internal/filtertest/`                                | 5 clones:96-298                        | 20      | 10m    | **20** | —    |
| 44  | **Table-drive bdd/plumbing_output_test.go**                      | `bdd/plumbing_output_test.go`                         | 260-280,506-532                        | 18      | 10m    | **18** | —    |
| 45  | **Extract assertStatsOutput helper** for cmd_test.go             | `cmd/cmd_test.go`                                     | 20-24,154-163                          | 20      | 10m    | **20** | —    |
| 46  | **Deduplicate hash/detector_test.go** assertion patterns         | `hash/detector_test.go`                               | 123-412,280-364                        | 12      | 10m    | **12** | —    |
| 47  | **Deduplicate printer/sarif_test.go** output assertions          | `printer/sarif_test.go`                               | 155-272                                | 12      | 10m    | **12** | —    |
| 48  | **Extract assertFileContains helper** for internal/utils tests   | `internal/utils/file_test.go`                         | 51-214,243-305                         | 15      | 10m    | **15** | —    |

### PHASE 4: Cleanup & Polish (Score 0-50)

| #   | Task                                                                  | File(s)                                                  | Clone Group             | I×Im×CV | Effort | Score  | Deps   |
| --- | --------------------------------------------------------------------- | -------------------------------------------------------- | ----------------------- | ------- | ------ | ------ | ------ |
| 49  | **Fix 15 LSP unused-write hints** in detector_types_test.go           | `pkg/artdupl/detector_types_test.go`                     | LSP diagnostics         | 0       | 10m    | **0**  | —      |
| 50  | **Fix LSP infertypeargs hints** (config/enum_helpers.go:78)           | `config/enum_helpers.go`                                 | LSP diagnostics         | 0       | 8m     | **0**  | —      |
| 51  | **Fix LSP unusedparams hint** (detector_validation_test.go:406)       | `pkg/artdupl/detector_validation_test.go`                | LSP diagnostics         | 0       | 8m     | **0**  | —      |
| 52  | **Fix golangci-lint wsl_v5 in text_test.go**                          | `printer/text_test.go`                                   | lint warnings           | 0       | 8m     | **0**  | —      |
| 53  | **Fix golangci-lint godot in html_test.go**                           | `printer/html_test.go`                                   | lint warnings           | 0       | 5m     | **0**  | —      |
| 54  | **Implement 6 SIMD TODOs**                                            | `internal/simd/simd.go`, `syntax/hash_simd.go`           | SIMD placeholder        | 27      | 12m    | **27** | —      |
| 55  | **Unify semantic default**: align config `false` with flag docs       | `cmd/`, `config/`                                        | Known Limitation        | 30      | 8m     | **30** | —      |
| 56  | **Extract shared parser interface** golang/parse.go + templ/parser.go | `syntax/golang/parse.go:27`, `syntax/templ/parser.go:19` | parse.go/parser.go      | 72      | 10m    | **72** | —      |
| 57  | **Deduplicate cmd/config_builder.go** method call patterns            | `cmd/config_builder.go:169,277,294`                      | 3 clones                | 50      | 10m    | **50** | —      |
| 58  | **Deduplicate cmd/stats.go** code blocks                              | `cmd/stats.go:138-159`                                   | 2 clones                | 50      | 8m     | **50** | —      |
| 59  | **Deduplicate suffixtree unicode test data**                          | `suffixtree/suffixtree_test.go`                          | 180-185 unicode         | 9       | 10m    | **9**  | —      |
| 60  | **Deduplicate detection/working_test.go + detection_test.go**         | `detection/`                                             | detection_test.go:51-63 | 15      | 10m    | **15** | —      |
| 61  | **Update AGENTS.md** — remove references to deleted packages          | `AGENTS.md`                                              | stale docs              | 16      | 10m    | **16** | —      |
| 62  | **Archive 322 docs/status/ files**                                    | `docs/status/`                                           | 304+ files              | 0       | 12m    | **0**  | —      |
| 63  | **Document ProcessedClone DTO decision**                              | `docs/` or `AGENTS.md`                                   | architecture doc        | 38      | 8m     | **38** | #1-#12 |
| 64  | **Property-based tests** for suffix tree core                         | `suffixtree/`                                            | testing gap             | 18      | 10m    | **18** | —      |
| 65  | **Memory layout optimization** for SIMD structures                    | `internal/simd/`                                         | performance             | 32      | 10m    | **32** | —      |

---

## Clone Group → Task Mapping

| Clone Group                 | Files                                                                                                        | Count | Task #          | Estimated Fix                    |
| --------------------------- | ------------------------------------------------------------------------------------------------------------ | ----- | --------------- | -------------------------------- |
| 6-file PrintClones loop     | printer/text.go:73, json.go:113, sarif.go:157, html.go:138, stats.go:91, plumbing.go:24                      | 6     | #1-#12, #18-#20 | Requires ProcessedClone DTO      |
| sort switch patterns        | printer/sorter.go:70,86,96,125                                                                               | 4     | #14-#15         | Extract shared sort key fn       |
| detector type assertions    | pkg/artdupl/detector_types_test.go (7 locations)                                                             | 13    | #33             | Generic assertFieldEquals helper |
| validation field assertions | pkg/artdupl/detector_validation_test.go (5 locations)                                                        | 5     | #35             | assertFieldEquals                |
| BDD sorting setup           | bdd/sorting_test.go:98-265                                                                                   | 5     | #36             | Table-drive test                 |
| filter assertions           | internal/filtertest/ (5 locations)                                                                           | 5     | #43             | assertFilterResult helper        |
| parser signatures           | syntax/golang/parse.go:27, syntax/templ/parser.go:19                                                         | 2     | #56             | Shared interface                 |
| unicode test data           | suffixtree/suffixtree_test.go:180-185                                                                        | 2-6   | #59             | Extract test data helper         |
| error asserts (9 files)     | cmd/, config/, job/, pkg/position/, printer/, syntax/                                                        | 9     | #39             | assertErrorContains helper       |
| config threshold checks     | config/config_enum_test.go:139-231                                                                           | 3     | #37             | assertThresholdEquals            |
| regression callbacks        | pkg/artdupl/detector_validation_test.go:445-541                                                              | 5     | #35             | assertFieldEquals                |
| html test patterns          | printer/html_test.go (5 locations)                                                                           | 5     | #40             | assertHTMLOutputContains         |
| text utils assertions       | printer/text_utils_test.go:19-76                                                                             | 6     | #41             | assertBuffer helpers             |
| detector uncovered          | pkg/artdupl/detector_uncovered_test.go:106-162                                                               | 2     | #34             | assertFieldEquals                |
| basic clone assertions      | pkg/artdupl/basic_test.go:208-234                                                                            | 4     | #42             | assertCloneValid                 |
| config error asserts        | config/config_enum_test.go:1155-1204                                                                         | 4     | #37             | assertThresholdEquals            |
| html buffer checks          | config/config_test.go:437-445, printer/html_test.go:573, text_test.go:315, syntax/syntax_test.go:282-297     | 8     | #39, #41        | Shared buffer helpers            |
| sarif output asserts        | printer/sarif_test.go:155-272                                                                                | 6+    | #47             | assertSARIOutput                 |
| cmd test patterns           | cmd/cmd_test.go:20-230                                                                                       | 6+    | #45             | assertStatsOutput helpers        |
| hash detector asserts       | hash/detector_test.go:123-412                                                                                | 6+    | #46             | assertHashOutput                 |
| detection test/working      | detection/detection_test.go:51-63, detection/working_test.go:21-32                                           | 2     | #60             | Shared setup                     |
| plumbing output             | bdd/plumbing_output_test.go:260-532                                                                          | 6+    | #44             | Table-drive                      |
| utils file asserts          | internal/utils/file_test.go:51-305                                                                           | 8+    | #48             | assertFile helpers               |
| examples test asserts       | examples/examples_test.go:271-273, detector_uncovered:215-217, detector_validation:362-408, text_utils:19-67 | 6     | #39-#41         | Shared assert helpers            |
| printer groups asserts      | printer/groups_test.go:82-124                                                                                | 6     | #47             | assertGroupOutput                |
| diff line asserts           | printer/diff.go:133-182                                                                                      | 3     | —               | Minor inline patterns            |
| cmd stats asserts           | cmd/cmd_test.go:154-163, cmd_utils_test.go:107-111                                                           | 4     | #45             | Shared output helpers            |
| cache test asserts          | cache/file_cache_test.go:119-120, detector_uncovered:389-416                                                 | 4     | #35             | assertFileContent                |
| semantic perf bench         | bdd/semantic_performance_bench_test.go:20-110                                                                | 3     | —               | Table benchmark data             |
| html clone group            | printer/html_test.go:259-358                                                                                 | 3     | #40             | assertHTMLOutput                 |
| printer diff tests          | printer/diff_test.go:291-449                                                                                 | 6     | —               | Test-specific                    |
| internal testutil           | internal/testutil/tabletest.go:41-83                                                                         | 2     | —               | Helper functions                 |
| format hash asserts         | job/parse_parallel_test.go:154-156, pkg/format/hash_test.go:29-31                                            | 2     | #39             | assertErrorContains              |
| stats test empty            | printer/stats_test.go:858,893                                                                                | 2     | —               | Test-specific                    |
| legacy/todo regex           | detection/legacy_detector.go:25, detection/todo_detector.go:33                                               | 2     | #22             | Extract compile fn               |
| config enum error           | config/config_enum_test.go:981-992                                                                           | 2     | #37             | assertThresholdEquals            |
| detector integration        | pkg/artdupl/detector_integration_test.go:164, detector_test.go:186                                           | 2     | —               | Copy pattern (OK)                |
| stats styles                | printer/stats_styles.go:31-32, cmd/art-dupl/main.go:28-30                                                    | 3     | —               | Import aliases (OK)              |
| run printer calls           | cmd/run_printer.go:15,30,40,14,27                                                                            | 3     | —               | Sequential calls (OK)            |
| run analysis/hash           | cmd/run_analysis.go:232, cmd/run_hash.go:26                                                                  | 2     | —               | Error return pattern (OK)        |
| run flags/stats             | cmd/run_flags.go:41, cmd/stats.go:65                                                                         | 2     | —               | Flag def pattern (OK)            |
| job setup patterns          | job/incremental_test.go:264-270, job/parse_parallel_test.go:58-64                                            | 2     | #60             | Shared test data helper          |
| syntax units                | syntax/syntax_test.go:312-328, syntax/findsyntaxunits_test.go:107-145                                        | 4     | #39             | assertErrorContains              |
| bdd plumbing                | bdd/plumbing_output_test.go:89-112                                                                           | 2     | #44             | Table-drive                      |
| cli commands                | bdd/cli_commands_test.go:161-421                                                                             | 4     | —               | BDD setup pattern (OK)           |
| config test configPath      | config/config_test.go:35,71                                                                                  | 2     | —               | Simple check (OK)                |
| error types                 | errors/marshal_test.go:21-161                                                                                | 6     | —               | Error test pattern (OK)          |
| artdupl basic               | pkg/artdupl/basic_test.go:138-149                                                                            | 2     | #42             | assertCloneValid                 |
| detector types              | pkg/artdupl/detector_types_test.go:30-45,212-218                                                             | 4     | #33             | assertFieldEquals                |
| simd test                   | internal/simd/simd_test.go:96,129                                                                            | 2     | —               | Slice check (OK)                 |
| cmd/cmd_utils               | cmd/cmd_utils_test.go:132-139,496-498                                                                        | 4     | #45             | Shared assertion                 |
| BDD error handling          | bdd/error_handling_test.go:219-230                                                                           | 2     | #44             | Table-drive                      |
| BDD output formats          | bdd/output_formats_and_filters_test.go:55-260                                                                | 4     | —               | BDD scenario pattern (OK)        |
| BDD default filtering       | bdd/default_filtering_test.go:23-78                                                                          | 2     | #36             | Shared BDD setup                 |
| BDD sem tests               | bdd/stats_semantic_test.go:109-142                                                                           | 2     | —               | BDD setup pattern (OK)           |
| config enum defaults        | config/config_enum_test.go:681-1126                                                                          | 2     | —               | Validation test pattern (OK)     |
| filtertest assertions       | internal/filtertest/assertions.go:17-54                                                                      | 2     | —               | Already has helpers              |
| cmd stats tests             | cmd/stats_integration_test.go:15,28                                                                          | 2     | —               | Simple assert (OK)               |
| examples sdk                | examples/examples_sdk_demo.go:70,138                                                                         | 2     | —               | Similar asserts (OK)             |
| cmd runner                  | internal/testutil/bdd_runners.go:131-144                                                                     | 2     | —               | Helper variants (OK)             |
| detection tests             | detection/detection_test.go:123-180                                                                          | 2     | —               | Table-driven (OK)                |
| testutil clone              | cmd/cmd_test.go:20, pkg/artdupl/detector_uncovered_test.go:28, printer/plumbing_test.go:183                  | 3     | #60             | Shared test data                 |
| printer sorter data         | printer/text_test.go:282-285, text_utils_test.go:122-125                                                     | 2     | —               | Test data (OK)                   |
| printer text                | printer/text_test.go:241-252                                                                                 | 2     | —               | Test data (OK)                   |
| pkg det uncovered           | pkg/artdupl/detector_uncovered_test.go:253,449                                                               | 2     | —               | Test feed (OK)                   |
| BDD cli                     | bdd/cli_commands_test.go:308-356, plumbing_and_paths_test.go:354-356                                         | 2     | —               | BDD pattern (OK)                 |
| filter tests                | internal/filtertest/user_scenario_test.go:213-240                                                            | 4     | —               | Test assertions (OK)             |
| syntax units tests          | syntax/findsyntaxunits_test.go:130-145                                                                       | 2     | —               | Table-driven (OK)                |
| printer stats               | printer/stats_test.go:774-779,71-208                                                                         | 4     | —               | Test setup (OK)                  |
| BBD templ                   | bdd/templ_clone_detection_test.go:62-73                                                                      | 2     | —               | BDD pattern (OK)                 |
| config test io              | config/config_test.go:299-319                                                                                | 2     | —               | Validation test (OK)             |
| printer diff                | printer/diff_test.go:291-327                                                                                 | 3     | —               | Diff assertion (OK)              |
| html test subs              | printer/html_test.go:1109-1120                                                                               | 2     | —               | HTML assertions (OK)             |
| suffixtree bench            | suffixtree/suffixtree_bench_test.go:144-225                                                                  | 2     | —               | Benchmark (OK)                   |

---

## Recommended Execution Order

**Sprint 1 (Sessions 1-3): Critical Architecture**

1. #2 Define ProcessedClone DTO interface (8m)
2. #3 Create NodeToProcessed conversion (10m)
3. #4 Convert TextPrinter (12m)
4. #5 Convert HTMLPrinter (12m)
5. #6 Convert JSONPrinter (12m)
6. #7 Convert SARIFPrinter (12m)
7. #8 Convert PlumbingPrinter (12m)
8. #9 Convert StatsPrinter (12m)
9. #10 Update callers (10m)
10. #1 Define types (10m) — can be parallel with #4-#9

**Sprint 2 (Sessions 4-5): Type Consolidation** 11. #11 Consolidate pkg/artdupl.Clone (10m) 12. #12 Consolidate printer.CloneGroup (8m) 13. #13 Define ProcessedCloneGroup (8m) 14. #14 Extract SortNodesByCriteria (12m) 15. #15 Extract sort key helper (12m)

**Sprint 3 (Sessions 6-7): Core Type Safety** 16. #17 TokenValue type (12m) 17. #16 Multi-language architecture (10m) 18. #25 CSV encoding/csv (12m) 19. #23-#24 domain/ tests (10m each)

**Sprint 4 (Sessions 8-9): Test Helpers** 20. #33-#35 assertFieldEquals (10-12m each) 21. #36 Table-drive BDD sorting (12m) 22. #37-#38 Config test helpers (12m each)

**Sprint 5 (Sessions 10-11): Cleanup** 23. #49-#53 LSP/Lint fixes (5-10m each) 24. #54 SIMD TODOs (12m) 25. #55 Semantic default (8m) 26. #62 Archive status files (12m)

---

_Total tasks: 65 | Est. sessions: ~15 | Zero tasks >12min_
