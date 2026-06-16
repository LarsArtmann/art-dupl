# Comprehensive Status Report — 2026-06-16 16:38

**Branch:** `fork` (pushed to origin)
**Go:** 1.26.3 linux/amd64
**Codebase:** 237 Go files, ~49,755 LOC, 101 modules
**Tests:** 246 BDD specs + extensive unit/integration tests
**Build:** ✅ Clean | **Lint:** ✅ 0 issues | **Non-BDD Tests:** ✅ All pass

---

## A) FULLY DONE ✅

### This Session's Commits

| Commit | Description |
|--------|-------------|
| `583e0e5` | docs: update TODO_LIST, FEATURES, AGENTS with Findings pipeline completion |
| `5cdc4ed` | fix(detection): use select on channel sends to prevent goroutine deadlock |
| `b0a5b16` | feat: wire PrintFindings through pipeline, fix vendorHash, refactor interner |

### Critical Feature — Findings Pipeline (was BROKEN, now FULLY FUNCTIONAL)

The #1 critical bug in this codebase: `FindFindings` was implemented and tested but **never called from any CLI path**. Users running `art-dupl --detection-methods todos` got zero output. This is now fixed end-to-end:

- ✅ `Printer.PrintFindings([]domain.Finding)` added to Printer interface
- ✅ Implemented in all 6 printers: Text, JSON, Plumbing, HTML, SARIF, Stats
- ✅ `executeAnalysis` now returns `<-chan domain.Finding` alongside `<-chan syntax.Match`
- ✅ `runCmd` drains findings channel and calls `p.PrintFindings(findings)`
- ✅ `runAllModes` passes findings to all output format files
- ✅ JSON output includes `findings` array in `JSONOutput` struct
- ✅ SARIF output converts findings to `SARIFResult` entries with `findingLevel` mapping
- ✅ Goroutine deadlock fixed: `select` on channel sends for ctx cancellation
- ✅ Verified end-to-end: `art-dupl --detection-methods todos /tmp/test/` produces output

### Code Quality

- ✅ `InternFilename` wired into all 4 transformer construction sites:
  - `syntax/golang/parse.go:43` — Go AST transformer
  - `syntax/templ/parser.go:41` — Templ AST transformer
  - `syntax/syntax.go:96` — `NewSyntheticFileNode`
  - `job/incremental.go:131` — Cache-hit path
- ✅ `forcetypeassert` in `intern.go` fixed (replaced `sync.Map` with `RWMutex+map` for type safety)
- ✅ Fuzz tests for templ parser (`FuzzParseBytes` — 2M+ executions, no panics)
- ✅ Extracted `spawnCloneDetection` and `spawnFindingDetection` from `executeAnalysis`
- ✅ All callers of `executeAnalysis` updated (runCmd, runAllModes, stats, integration tests)
- ✅ All 25 BuildFlow pre-commit hook steps pass

### Previously Completed (Prior Sessions — 2026-06-15 to 2026-06-16)

- ✅ `MethodDetector` interface with polymorphic dispatch via adapters
- ✅ `domain.Finding` type + `FindFindings` pipeline
- ✅ `--suppress-test-low` and `--test-threshold` CLI flags
- ✅ `context.Context` propagation through all detectors and suffix tree
- ✅ SDK type independence (`pkg/artdupl` has own types, no config aliases)
- ✅ Shared `pkg/enum` package, unified domain enums with JSON marshaling
- ✅ `ClonePriority` replaces `CloneSeverity` (collapsed 2 types into 1)
- ✅ Actionability classification (11 non-actionable patterns: test data, table-driven, scaffolding, defer, error propagation, etc.)
- ✅ DescribeTable and builder/callback detection patterns
- ✅ String interning infrastructure (`InternFilename`)
- ✅ Fuzz tests for suffix tree (`FuzzFindDuplOver`, `FuzzFindDuplOverCancellation`)
- ✅ Tier 1-3 quick fixes (20+ items: SDK safety, detection cleanup, domain types)

---

## B) PARTIALLY DONE 🟡

### Findings Pipeline — Output Quality

The pipeline is wired and produces output, but the **text output format is basic**:

```
📋 Findings (3):
  /tmp/main.go:3 [todo] fix this later
  /tmp/main.go:4 [todo] this is broken
  /tmp/main.go:8 [todo] temporary workaround
```

**What's missing:**
- No priority badges or color coding in text output
- No grouping by file in text output
- JSON findings are included but not in SimpleJSON format
- SARIF findings lack `Rule` definitions in the tool driver (only results)

### InternFilename — Partial Coverage

Wired into 4 transformer construction sites, but the interner uses a simple `RWMutex+map`. For truly large codebases (10k+ files), a `sync.Map` with `LoadOrStore` would be lock-free on the read path. The current implementation was chosen to fix the `forcetypeassert` lint issue.

### Documentation Completeness

- ✅ TODO_LIST.md updated
- ✅ FEATURES.md updated
- ✅ AGENTS.md updated
- ❌ No ADR for the Findings pipeline architecture decision
- ❌ HOW_TO_USE.md not updated with `--detection-methods todos` examples

---

## C) NOT STARTED ❌

### HIGH Priority Architecture Refactors (Deferred — multi-session)

1. **ProcessedClone DTO decoupling** — Printer still imports `syntax.Node` directly in `actionability.go`. `clone_processor.go` is the bridge. Requires redesigning the actionability pattern evaluation to work on serializable data instead of AST nodes.

2. **Consolidate three parallel Clone types** — `printer.CloneGroup` (JSON DTO), `pkg/artdupl.Clone` (SDK DTO), `domain.ProcessedClone` (canonical internal DTO). Each has different field names and shapes. Consolidation blocked on Printer/SDK DTO design decisions.

3. **Split `printer/` package** — 50+ files in one package. Should split into `printer/clone/`, `printer/stats/`, `printer/actionability/`, etc. Large mechanical refactor with no behavior change.

4. **Type-strengthen ProcessedClone** — `Filename string → domain.Filepath`, `LineStart/LineEnd int → domain.LineNumber`. ~25 consumer sites need updating. Low risk but tedious.

### Architecturally Blocked

5. **`syntax/golang` facade** — BLOCKED by import cycle (`syntax/golang` imports `syntax` for `Node` type, so `syntax` cannot re-export `golang` symbols). Would require splitting `Node` into a separate package.

### MEDIUM Priority

6. **Refactor `actionability.go`** (623 lines) — Patterns already extracted into 30 functions, but file is still large. Could split into `actionability_test_patterns.go`, `actionability_defer_patterns.go`, etc.

7. **Hybrid slice/map transition storage** — Suffix tree transitions use map for O(1) lookup. For nodes with few transitions (<8), a slice would be more cache-friendly. Low value since map is already O(1).

---

## D) TOTALLY FUCKED UP 💥

### BDD Test Suite Timeout

The BDD suite (264 specs) **times out at 600s**. Individual specs pass instantly, and subsets of ~50 pass in seconds. The full suite hangs — likely a goroutine leak or resource exhaustion from running 264 specs that each spawn the full analysis pipeline.

**Impact:** Cannot run `go test ./...` without manual timeout or excluding BDD.
**Root cause hypothesis:** Each BDD spec compiles and runs art-dupl in-process, spawning goroutines. The findings pipeline adds another goroutine per run. If any spec doesn't drain the findings channel, the goroutine blocks forever (though we fixed the select-on-send pattern, there may be a path where the channel is never consumed).
**Pre-existing:** This timeout existed before our changes — verified by stashing our changes and reproducing.

### Stale LSP/gopls Diagnostics

The IDE shows 4-11 phantom errors at all times (wrong arg counts, undefined types). These are **stale gopls cache** — `go build ./...` and `golangci-lint run ./...` both produce 0 issues. The gopls index doesn't pick up changes to function signatures. Requires `:LspRestart` or `go clean -cache`.

### Pre-existing Binaries in Repo

BuildFlow flags 4 binaries that are NOT in git but exist in the working directory: `art-dupl`, `bdd/art-dupl-filter_features-test`, `dist/art-dupl`, `result`. These are build artifacts that should be in `.gitignore`.

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Architecture

1. **Break the import cycle** — The `syntax` ↔ `syntax/golang` cycle prevents clean facade design. Extract `syntax.Node` into a `syntax/types` or `ast/types` package that both can import.

2. **Single Clone type** — Having 3 parallel types for the same concept (clone group) is a maintenance burden. Pick one canonical type and use adapters at boundaries.

3. **Printer package decomposition** — 50 files in `printer/` is a code smell. Split by concern: `output/text/`, `output/json/`, `output/html/`, `output/sarif/`, `analyze/actionability/`.

### Testing

4. **Fix BDD timeout** — Profile the BDD suite to find the hanging spec(s). Likely a goroutine leak. Use `runtime.NumGoroutine()` in tests to detect leaks.

5. **Integration test for findings** — Add a BDD spec that verifies `--detection-methods todos` produces expected output in text, JSON, and SARIF formats.

6. **Benchmark findings pipeline** — Measure overhead of the findings goroutine on clone detection performance.

### Code Quality

7. **Delete `CloneSeverity` type alias** — The deprecated aliases in `domain/types_severity.go` cause `exhaustive` lint false positives. Schedule deletion for next breaking version.

8. **Templ semantic mode** — Templ matching is purely structural. Adding identifier/operator encoding (like Go has) would reduce false positives.

9. **Consistent error wrapping** — The findings path uses `fmt.Errorf` in some places and `duplerrors.Wrap` in others. Standardize.

### Developer Experience

10. **HOW_TO_USE.md update** — Document `--detection-methods todos,legacy` and what the findings output looks like.

11. **`.gitignore` for build artifacts** — Add `art-dupl`, `dist/`, `result` to `.gitignore`.

12. **gopls reliability** — Document the stale cache issue in AGENTS.md with the workaround (`go clean -cache && go build`).

---

## F) Top 25 Things to Do Next

| # | Task | Impact | Effort | Priority |
|---|------|--------|--------|----------|
| 1 | **Fix BDD test suite timeout** (profile goroutine leaks) | Critical | Medium | P0 |
| 2 | **Delete CloneSeverity type aliases** (fixes exhaustive lint false positive) | High | Low | P0 |
| 3 | **Add BDD spec for findings output** (verify --detection-methods todos works in all formats) | High | Low | P0 |
| 4 | **Add findings to SimpleJSON output** | Medium | Low | P1 |
| 5 | **Add SARIF rule definitions for findings** (tool.driver.rules) | Medium | Low | P1 |
| 6 | **Improve text findings output** (group by file, priority badges) | Medium | Low | P1 |
| 7 | **Update HOW_TO_USE.md** with findings examples | Medium | Low | P1 |
| 8 | **Add `.gitignore` entries** for build artifacts | Low | Trivial | P1 |
| 9 | **Add ADR-0005** for Findings pipeline architecture | Medium | Low | P1 |
| 10 | **Type-strengthen ProcessedClone** (Filename→Filepath, Lines→LineNumber) | Medium | Medium | P2 |
| 11 | **Decouple actionability.go from syntax.Node** (ProcessedClone DTO) | High | High | P2 |
| 12 | **Consolidate 3 Clone types** into 1 canonical | High | High | P2 |
| 13 | **Split printer/ package** into sub-packages | Medium | High | P2 |
| 14 | **Break syntax/syntax-golang import cycle** (extract Node to shared package) | High | High | P2 |
| 15 | **Add templ semantic mode** (identifier/operator hashing) | Medium | High | P2 |
| 16 | **Benchmark findings pipeline overhead** | Medium | Low | P2 |
| 17 | **Add goleak** to unit tests (goroutine leak detection) | Medium | Low | P2 |
| 18 | **Refactor actionability.go** into sub-files by pattern category | Low | Medium | P3 |
| 19 | **Implement hybrid slice/map transition storage** | Low | Medium | P3 |
| 20 | **Add `--findings-only` flag** (skip clone detection, just findings) | Medium | Low | P3 |
| 21 | **Add severity filtering** (`--min-priority medium` to filter findings) | Medium | Low | P3 |
| 22 | **Add findings to stats output** (count by type, priority distribution) | Low | Low | P3 |
| 23 | **Add custom TODO patterns** (`--todo-patterns "BUG,PERF,SECURITY"`) | Medium | Medium | P3 |
| 24 | **Add legacy pattern customization** (`--legacy-patterns "pkg.OldFunc"`) | Medium | Medium | P3 |
| 25 | **Cache findings results** in incremental mode | Low | Medium | P3 |

---

## G) Top Question I Cannot Figure Out 🔴

**Why does the BDD test suite hang when running all 264 specs, but individual specs and small batches complete in milliseconds?**

I've verified:
- Individual specs pass (`--ginkgo.focus="should detect duplicates"` → PASS in 0.003s)
- The goroutine deadlock fix (select on channel sends) is in place
- Stashing our changes and running on the clean tree also hangs
- `go test ./bdd/... -ginkgo.dry-run` completes instantly
- No obvious shared state between specs

The hang is pre-existing — it existed before this session's changes. But it blocks `go test ./...` from completing, which is a significant developer experience problem. The most likely cause is either:

1. **Goroutine accumulation** — Each spec spawns goroutines (clone detection, findings detection) that don't fully drain before the next spec starts. Over 264 specs, goroutines pile up.
2. **File handle exhaustion** — Each spec creates temp files and runs analysis. On Linux, file descriptor limits could be hit.
3. **Ginkgo serial execution deadlock** — Some specs may depend on shared state (package-level variables) that gets corrupted when run in sequence.

**What I'd need to figure this out:** Run the BDD suite with `-ginkgo.progress` and `-ginkgo.v` to see exactly which spec hangs, then examine its setup/teardown. A `runtime.NumGoroutine()` check before/after each spec would reveal if goroutines are leaking.

---

## Session Metrics

| Metric | Value |
|--------|-------|
| Commits this session | 3 (pushed to origin) |
| Files changed | 15+ |
| Lines added | ~400 |
| Lines removed | ~100 |
| Tests passing | 21/21 non-BDD packages ✅ |
| Lint issues | 0 ✅ |
| Build status | Clean ✅ |
| BuildFlow | 25/25 steps pass ✅ |
| Critical bugs fixed | 2 (Findings pipeline, goroutine deadlock) |
