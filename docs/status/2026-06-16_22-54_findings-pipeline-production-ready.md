# Comprehensive Status Report — 2026-06-16 22:54

**Branch:** `fork` (pushed to origin, up to date)
**Go:** 1.26.3 linux/amd64
**Codebase:** 236 Go files, ~49,994 LOC, 101 modules, 246 BDD specs
**Commits today:** 28 | **Total commits:** 1,451
**Build:** ✅ Clean | **Lint:** ✅ 0 issues | **Tests:** ✅ 21/21 non-BDD packages pass

---

## A) FULLY DONE ✅

### Findings Pipeline — End to End (3 rounds of self-review)

The critical bug that started this session: `FindFindings` was implemented but **never called from any CLI path**. Users running `art-dupl --detection-methods todos` got zero output. Now fully fixed through 3 self-review iterations:

**Round 1 — Core wiring:**

- ✅ `Printer.PrintFindings([]domain.Finding)` added to Printer interface
- ✅ Implemented in all 6 printers: Text, JSON, Plumbing, HTML, SARIF, Stats
- ✅ `executeAnalysis` returns `<-chan domain.Finding` alongside clone channel
- ✅ `runCmd` and `runAllModes` drain findings and pass to printers
- ✅ JSON output includes `findings` array in `JSONOutput` struct
- ✅ SARIF output converts findings to `SARIFResult` entries with `findingLevel`
- ✅ Goroutine deadlock fixed: `select` on channel sends for ctx cancellation

**Round 2 — Bug fixes from self-review:**

- ✅ Nil channel deadlock fix: `collectFindings(nil)` blocks forever in hash-only mode — guarded with nil check
- ✅ JSON plumbing fix: `PrintFindings` was called AFTER `OutputJSON` flushed — findings passed into `printDupls` so they're set before JSON flush
- ✅ Goroutine leak fix: `stats.go` and integration tests now drain `findingChan` instead of discarding with `_`
- ✅ Dead code removal: Deleted `domain/types_severity.go` (all `CloneSeverity` aliases were zero-usage dead code causing exhaustive lint false positives)

**Round 3 — Quality hardening:**

- ✅ `FindingType` enum gets `pkg/enum` methods (MarshalJSON/UnmarshalJSON/ParseFindingType) matching ClonePriority pattern
- ✅ `ErrInvalidLineNumber` sentinel added — `Finding.Validate` was reusing semantically wrong `ErrLineEndBeforeStart`
- ✅ SARIF rule definitions added for `art-dupl/todo` and `art-dupl/legacy` (previously results referenced undefined rules — SARIF spec violation)
- ✅ Text footer now includes `"Found total N findings."` matching clone groups summary
- ✅ SARIF level strings extracted as constants (`sarifLevelError`, `sarifLevelWarning`, `sarifLevelNote`)
- ✅ Dead `ErrInvalidCloneSeverity` sentinel deleted
- ✅ 7 new tests: FindingType_IsValid, FindingType_String, ParseFindingType, FindingType_MarshalJSON, FindingType_UnmarshalJSON, FindingType_UnmarshalJSON_Invalid, Finding_Validate (4 sub-tests)

### Code Quality Features

- ✅ `InternFilename` wired into all 4 transformer construction sites (golang parse, templ parse, NewSyntheticFileNode, incremental cache-hit path)
- ✅ `forcetypeassert` in `intern.go` fixed (replaced `sync.Map` with `RWMutex+map` for type safety)
- ✅ Templ parser fuzz tests (`FuzzParseBytes` — 2M+ executions, no panics)
- ✅ Suffix tree fuzz tests (`FuzzFindDuplOver`, `FuzzFindDuplOverCancellation`)
- ✅ Integration tests: `TestExecuteAnalysis_FindingsPipeline` (2 cases: todos produce findings, hash-only returns nil channel)

### Commits This Session (28 total, 15 in this reporting window)

| Commit    | Description                                                           |
| --------- | --------------------------------------------------------------------- |
| `17f799f` | test(domain): FindingType enum tests + Finding.Validate               |
| `8a4a453` | refactor(domain): delete dead ErrInvalidCloneSeverity                 |
| `70bb3a9` | feat(printer): SARIF rule definitions + text footer count             |
| `73bb46b` | refactor(domain): FindingType pkg/enum methods + ErrInvalidLineNumber |
| `bc12630` | test(cmd): integration test for findings pipeline                     |
| `b473f54` | fix: drain findingChan, delete CloneSeverity aliases                  |
| `911784b` | fix(cmd): pass findings to printDupls (JSON/SARIF fix)                |
| `1677d80` | fix(cmd): guard collectFindings against nil channel                   |
| `63281c1` | docs(status): comprehensive status report                             |
| `583e0e5` | docs: update TODO_LIST, FEATURES, AGENTS                              |
| `5cdc4ed` | fix(detection): select on channel sends (deadlock fix)                |
| `b0a5b16` | feat: wire PrintFindings through pipeline + InternFilename            |

---

## B) PARTIALLY DONE 🟡

### Findings Text Output

The text output is functional but basic:

```
📋 Findings (3):
  /tmp/main.go:3 [todo] fix this later
  /tmp/main.go:4 [todo] this is broken
  /tmp/main.go:8 [todo] temporary workaround

Found total 3 findings.
```

**Missing:** No grouping by file, no priority badges/colors, no rich text mode support for findings.

### Documentation

- ✅ AGENTS.md updated with Finding pipeline architecture
- ✅ TODO_LIST.md updated with all completed items
- ✅ FEATURES.md updated (TODO/Legacy detection, JSON, SARIF descriptions)
- ❌ HOW_TO_USE.md does NOT document `--detection-methods todos,legacy`
- ❌ README.md does NOT mention TODO/legacy detection methods

### SARIF Completeness

- ✅ Rule definitions exist for `duplicate-code`, `todo`, `legacy`
- ✅ Results are generated for both clones and findings
- ❌ No `runs[].results[].properties` with custom metadata
- ❌ Finding-level doesn't map from `ClonePriority` for legacy findings (they all get `"note"`)

---

## C) NOT STARTED ❌

### Architecture (Multi-session refactors — documented in TODO_LIST.md)

1. **ProcessedClone DTO decoupling** — `actionability.go` still imports `syntax.Node` directly
2. **Consolidate three parallel Clone types** — `printer.CloneGroup`, `pkg/artdupl.Clone`, `domain.ProcessedClone`
3. **Split `printer/` into sub-packages** — 50+ files in one package
4. **Type-strengthen ProcessedClone** — `Filename string → Filepath`, `LineStart/End int → LineNumber` (~25 sites)
5. **`syntax/golang` facade** — BLOCKED by import cycle

### Code Quality

6. **Refactor `printDupls` signature** — 11 parameters, should use a struct (deferred — high churn, stable interface)
7. **`LegacyIssue` Tags field** — Asymmetric with `TodoIssue` which has Tags (low value — no natural tag source)
8. **HOW_TO_USE.md** — No detection-methods or findings documentation
9. **README.md** — Missing todos/legacy in detection methods table

### Testing

10. **BDD spec for findings output** — No Ginkgo spec verifies `--detection-methods todos` produces expected output across formats
11. **BDD test suite timeout** — 264 specs hang at 600s (pre-existing, likely goroutine accumulation)

---

## D) TOTALLY FUCKED UP 💥

### BDD Test Suite Timeout (Pre-existing)

The BDD suite (264 specs) times out at 600s. Individual specs pass in milliseconds. The full suite hangs — likely goroutine accumulation from 264 specs each spawning the full analysis pipeline. This is **pre-existing** (verified by stashing changes and reproducing).

**Impact:** Cannot run `go test ./...` without excluding BDD.
**Root cause hypothesis:** Goroutine leak per spec, compounding over 264 runs.

### Stale LSP/gopls Cache (Persistent)

IDE shows 4-11 phantom errors at all times. `go build` and `golangci-lint run` produce 0 issues. Requires `:LspRestart` or `go clean -cache`.

### Build Artifacts in Working Directory

BuildFlow flags 4 untracked binaries: `art-dupl`, `dist/art-dupl`, `result`, `bdd/art-dupl-filter_features-test`. These are NOT in git but exist locally. `.gitignore` covers them, but the linter still complains.

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Architecture

1. **Break the import cycle** — Extract `syntax.Node` into `syntax/types` so `syntax/golang` can import it without circular dependency. Unblocks the facade refactor.

2. **Single Clone type** — 3 parallel types for the same concept is maintenance debt. Pick `domain.ProcessedClone` as canonical, use adapters at JSON/SDK boundaries.

3. **Printer package decomposition** — 50 files in `printer/` is a smell. Split: `output/text/`, `output/json/`, `output/html/`, `output/sarif/`, `analyze/actionability/`.

4. **`printDupls` struct refactor** — 11 params → struct with config fields. Currently deferred due to churn.

### Testing

5. **Fix BDD timeout** — Profile with `-ginkgo.progress` to find hanging spec. Add `runtime.NumGoroutine()` checks.

6. **Add goleak** — `go.uber.org/goleak` in TestMain to catch goroutine leaks automatically.

7. **BDD findings spec** — Verify `--detection-methods todos` produces output in text, JSON, SARIF.

### Code Quality

8. **HOW_TO_USE.md** — Document findings output, detection methods, examples.

9. **Rich text findings** — Group by file, priority badges, consistent with clone rich text mode.

10. **Templ semantic mode** — Add identifier/operator encoding to reduce false positives.

---

## F) Top 25 Things to Do Next

| #   | Task                                                               | Impact   | Effort | Priority |
| --- | ------------------------------------------------------------------ | -------- | ------ | -------- |
| 1   | **Fix BDD test suite timeout** (profile goroutine leaks)           | Critical | Medium | P0       |
| 2   | **HOW_TO_USE.md**: Add detection-methods + findings sections       | High     | Low    | P0       |
| 3   | **README.md**: Add todos/legacy to detection methods table         | High     | Low    | P0       |
| 4   | **Add BDD spec for findings** (verify todos output in all formats) | High     | Low    | P1       |
| 5   | **Rich text findings** (group by file, priority badges)            | Medium   | Low    | P1       |
| 6   | **Add goleak** to unit tests                                       | Medium   | Low    | P1       |
| 7   | **printDupls struct refactor** (11 params → config struct)         | Medium   | Medium | P2       |
| 8   | **Type-strengthen ProcessedClone** (Filepath, LineNumber)          | Medium   | Medium | P2       |
| 9   | **Break syntax import cycle** (extract Node to shared types)       | High     | High   | P2       |
| 10  | **Consolidate 3 Clone types** into 1 canonical                     | High     | High   | P2       |
| 11  | **Decouple actionability.go from syntax.Node**                     | High     | High   | P2       |
| 12  | **Split printer/ package** into sub-packages                       | Medium   | High   | P2       |
| 13  | **Templ semantic mode** (identifier/operator hashing)              | Medium   | High   | P2       |
| 14  | **Benchmark findings pipeline overhead**                           | Medium   | Low    | P2       |
| 15  | **Add `--findings-only` flag** (skip clone detection)              | Medium   | Low    | P3       |
| 16  | **Add severity filtering** (`--min-priority medium`)               | Medium   | Low    | P3       |
| 17  | **Custom TODO patterns** (`--todo-patterns "BUG,PERF"`)            | Medium   | Medium | P3       |
| 18  | **Custom legacy patterns** (`--legacy-patterns "pkg.OldFunc"`)     | Medium   | Medium | P3       |
| 19  | **Cache findings** in incremental mode                             | Low      | Medium | P3       |
| 20  | **Findings in stats output** (count by type/priority)              | Low      | Low    | P3       |
| 21  | **Hybrid slice/map transition storage**                            | Low      | Medium | P3       |
| 22  | **Refactor actionability.go** into sub-files                       | Low      | Medium | P3       |
| 23  | **Add ADR-0005** for Findings pipeline architecture                | Low      | Low    | P3       |
| 24  | **LegacyIssue Tags field** for consistency                         | Low      | Low    | P3       |
| 25  | **Improve SARIF finding-level mapping** for legacy                 | Low      | Low    | P3       |

---

## G) Top Question I Cannot Figure Out 🔴

**Why does the BDD test suite (264 specs) hang ONLY when running ALL specs together, but individual specs and subsets complete in milliseconds?**

Verified facts:

- Individual specs: PASS in 0.003s
- Subset of 50: PASS in seconds
- Full suite of 264: HANGS at 600s timeout
- `git stash` (clean tree): Same hang — pre-existing
- Dry run: Completes instantly
- Each spec uses `RunArtDupl()` which executes the tool in-process

The goroutine leak we fixed (undrained `findingChan`) was a contributing factor for non-BDD tests, but the BDD suite was hanging before our changes. The most likely remaining cause: each BDD spec calls the full CLI in-process via `Execute()`, which spawns goroutines. Over 264 specs without process isolation, goroutines or file handles accumulate.

**What would resolve this:** Either (a) add `runtime.NumGoroutine()` assertions between specs, or (b) run BDD specs in parallel worker processes via `ginkgo -p`, or (c) profile with `pprof` to identify the exact resource that leaks.

---

## Session Metrics

| Metric               | Value                                                                       |
| -------------------- | --------------------------------------------------------------------------- |
| Commits this session | 28                                                                          |
| Files changed        | ~20                                                                         |
| Lines added          | ~600                                                                        |
| Lines removed        | ~150                                                                        |
| Tests passing        | 21/21 non-BDD ✅                                                            |
| Lint issues          | 0 ✅                                                                        |
| Build status         | Clean ✅                                                                    |
| BuildFlow            | 34/34 steps pass ✅                                                         |
| Critical bugs fixed  | 4 (nil channel deadlock, JSON plumbing, goroutine deadlock, goroutine leak) |
| Self-review rounds   | 3                                                                           |
| New tests added      | 9 (7 domain + 2 integration)                                                |
