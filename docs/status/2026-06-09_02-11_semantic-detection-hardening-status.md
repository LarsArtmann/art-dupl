# Status Report: Semantic Detection Hardening

**Date:** 2026-06-09 02:11
**Branch:** fork (up to date with origin)
**Commits this session:** 3 (48537b2, 8e170e8, 68d3d6e)
**Test status:** 22/22 packages PASS
**Lint status:** 3 pre-existing godoclint warnings (not our code), 0 new issues
**Build status:** `just build` PASS, `nix build` FAIL (vendorHash stale)

---

## A) FULLY DONE

### Session: Semantic Detection Hardening (2026-06-09)

| # | Task | Commit | Impact |
|---|------|--------|--------|
| 1 | Fix classification bug — decode base type before category switch | 68d3d6e | CRITICAL — every FuncDecl/Ident/SelectorExpr was `unknown` in default mode |
| 2 | Fix actionability Type comparisons via baseTypeOf | 68d3d6e | Prevents breakage from operator encoding |
| 3 | Encode operators into semantic types (BinaryExpr, UnaryExpr, IncDecStmt, AssignStmt) | 68d3d6e | Eliminates inverse-condition false positives |
| 4 | Fix hashSeq truncation (1 byte → 4 bytes per node) | 68d3d6e | Fixes incorrect group merging |
| 5 | Add CategoryIdiom for <5 token clones | 68d3d6e | Direct triage improvement |
| 6 | BDD tests for operator differences and inverse conditions | 68d3d6e | Confidence in semantic mode |
| 7 | Update AGENTS.md with semantic encoding conventions | 68d3d6e | Documentation |
| 8 | Add real-world feedback document from overview dedup session | 48537b2 | Feedback loop |
| 9 | Add comprehensive Pareto execution plan | 8e170e8 | Planning |

### Previously Done (project-level)

- 14 TODO items completed in 2026-05-23 batch
- 3 ADRs (map-based transition, semantic default, reflection-based config merge)
- 7 output formats (text, HTML, JSON, plumbing, SARIF, stats, CSV)
- 2 language support (Go, Templ)
- 5 detection methods
- Smart filtering (sqlc, templ, protobuf, mockgen, stringer)
- SDK/programmatic API in `pkg/artdupl/`
- Actionability filtering (signature-only, defer RAII, error propagation)
- Health score grading (A through F)

---

## B) PARTIALLY DONE

| Item | Status | What's Missing |
|------|--------|----------------|
| `nix build` | `vendorHash` stale | Need to update hash in flake.nix: set `vendorHash=""`, run `nix build`, copy correct hash |
| Semantic mode for Templ | No encoding applied | `syntax/templ/` matching is purely structural — no identifier/operator encoding |
| SDK docs | `SDK_DESIGN.md` exists | No usage examples, no godoc rendering |

---

## C) NOT STARTED

From TODO_LIST.md (18 open items):

| Priority | Item |
|----------|------|
| HIGH | Introduce ProcessedClone DTO to decouple Printer from syntax.Node |
| HIGH | Consolidate three parallel Clone types (printer.clone, printer.CloneGroup, pkg/artdupl.Clone, domain.ProcessedClone) |
| HIGH | Implement TokenValue type with validation |
| MED | CSV output using `encoding/csv` |
| MED | Unify enum patterns |
| MED | SIMD memory layouts / string interning |
| MED | Decouple printer/clone_classify.go from syntax/golang |
| MED | `--output-file` flag for stats subcommand |
| MED | Split printer/stats_test.go (975L → 3 files) |
| LOW | Refactor syntax/golang/transform.go (369L) |
| LOW | Fix remaining LSP hints |
| LOW | Create domain.HealthScore typed enum |
| LOW | Write SDK documentation |
| LOW | BDD tests for `--only templ`, `--only go`, `--include-generic` |
| LOW | Fuzz tests for templ parser |
| LOW | Validate GoReleaser config |

From ROADMAP.md (not started, aspirational):

- TypeScript/JavaScript support
- Python support
- Watch mode
- GitHub Actions templates, pre-commit hooks
- Performance baseline benchmarks

From feedback (2026-06-08-overview-dedup-session.md):

- Suppress test-only low-priority clones (reiterated from 2026-06-04)
- Test-table pattern detection (CompositeLit multi-clone in same file)
- Cross-reference with existing helpers ("this clone already has a function")
- Semantic mode granularity levels (`--semantic=strict`)

---

## D) TOTALLY FUCKED UP

| Issue | Severity | Details |
|-------|----------|---------|
| `nix build` broken | HIGH | `vendorHash` mismatch: `sha256-0T74N...` vs `sha256-p8mld...`. Has been broken since before this session. Build works via `just build`. |
| Pre-existing classification bug | WAS CRITICAL | **Fixed this session (68d3d6e).** Every FuncDecl/Ident/SelectorExpr clone was classified as `unknown` in semantic mode (the default). Existed since semantic encoding was introduced. |
| 3 godoclint warnings | LOW | `hash/doc.go`, `printer/stats.go`, `syntax/golang/doc.go` — "package has more than one godoc". Pre-existing, not ours. |
| `cmd/run_crawl.go:45` LSP warning | LOW | `bufio.Scanner` missing `sc.Err()` check after scan loop. Pre-existing. |
| `charmbracelet/fang` v1 | LOW | Library policy says use `fang/v2` — currently on v1. Pre-existing. |
| `sha1_hash` in cache/file_cache.go | LOW | Library policy flags SHA-1 usage. Pre-existing. |

---

## E) WHAT WE SHOULD IMPROVE

### Architecture
1. **Printer ↔ syntax.Node coupling** — All 6 printers depend on `[]*syntax.Node`. ProcessedClone DTO is the fix (111 test call sites to update).
2. **Three parallel Clone types** — `printer.clone`, `printer.CloneGroup`, `pkg/artdupl.Clone`, `domain.ProcessedClone`. Consolidation blocked on Printer DTO change.
3. **Decouple clone_classify.go from syntax/golang** — Classification imports golang node types directly. Should use a language-agnostic interface.

### False Positive Reduction
4. **Suppress test-only low-priority clones** — 76% of reported clones at threshold 15 were test noise. Filter at output level.
5. **Test-table pattern detection** — Multiple clones in same file within CompositeLit → classify as `test-data`.
6. **Inverse-condition detection** — Now mitigated by operator encoding, but a "inverse of line X" note would be even better.

### Tooling
7. **Fix nix build** — Update vendorHash. 5-minute fix.
8. **hashSeq unit tests** — Currently ZERO unit tests (only benchmarks). Should add basic coverage.
9. **Fuzz tests for templ parser** — Important for robustness with untrusted input.

### Performance
10. **SIMD memory layouts** — String interning and cache-aligned Node arrays for large codebases.
11. **hashSeq benchmark regression check** — CI should fail if hashing regresses significantly.

### Quality of Life
12. **`--output-file` flag for stats** — Write stats to file instead of stdout.
13. **SDK documentation** — pkg/artdupl/ needs usage examples.
14. **Watch mode** — Re-run on file changes (from roadmap).

---

## F) Top 25 Things to Get Done Next

Sorted by impact × effort (Pareto ordering):

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | **Fix nix vendorHash** — set `""`, build, copy hash | HIGH | 5min | Tooling |
| 2 | **Suppress test-only low-priority clones** in output | HIGH | 30min | False Positive |
| 3 | **Add hashSeq unit tests** (currently zero) | HIGH | 30min | Testing |
| 4 | **Decouple clone_classify from syntax/golang** via interface | HIGH | 2h | Architecture |
| 5 | **ProcessedClone DTO** — decouple printer from syntax.Node | HIGH | 4h | Architecture |
| 6 | **Consolidate three Clone types** into one | HIGH | 3h | Architecture |
| 7 | **Implement TokenValue type with validation** | MED | 1h | Type Safety |
| 8 | **Test-table pattern detection** in classification | MED | 1h | False Positive |
| 9 | **Cross-reference with existing helpers** detection | MED | 4h | Feature |
| 10 | **Semantic mode for Templ** — identifier/operator encoding | MED | 3h | Feature |
| 11 | **Fix 3 godoclint warnings** — duplicate package docs | LOW | 10min | Quality |
| 12 | **Fix bufio scanner `sc.Err()` check** in run_crawl.go | LOW | 5min | Quality |
| 13 | **Migrate fang v1 → fang/v2** | LOW | 1h | Dependencies |
| 14 | **Replace SHA-1 in file_cache.go** with SHA-256 | LOW | 15min | Security |
| 15 | **CSV output using encoding/csv** | MED | 2h | Feature |
| 16 | **`--output-file` flag for stats subcommand** | LOW | 30min | Feature |
| 17 | **Split printer/stats_test.go** (975L → 3 files) | LOW | 30min | Quality |
| 18 | **Refactor syntax/golang/transform.go** (369L switch) | LOW | 2h | Quality |
| 19 | **Create domain.HealthScore typed enum** | LOW | 15min | Type Safety |
| 20 | **Unify enum patterns** across codebase | LOW | 1h | Quality |
| 21 | **Write SDK documentation** with examples | MED | 2h | Documentation |
| 22 | **BDD tests for `--only templ`, `--only go`, `--include-generic`** | MED | 1h | Testing |
| 23 | **Fuzz tests for templ parser** | MED | 2h | Testing |
| 24 | **Watch mode** — re-run on file changes | HIGH | 4h | Feature |
| 25 | **Performance baseline benchmarks** — CI regression tracking | MED | 2h | Performance |

---

## G) Top #1 Question I Cannot Figure Out Myself

**What is the target audience priority for art-dupl right now?**

Is the priority:
- **(A)** CLI tool for developers running manually on their projects (improve output, reduce noise, better triage)?
- **(B)** CI/CD integration for automated clone detection in pipelines (exit codes, SARIF, machine-readable output)?
- **(C)** SDK/library for other tools to consume programmatically (SDK docs, stable API, ProcessedClone DTO)?

This matters because:
- (A) → suppress test noise, improve HTML report, add idiom filtering
- (B) → fix nix build, add `--exit-code` flag, improve SARIF, GitHub Action template
- (C) → ProcessedClone DTO first, stabilize pkg/artdupl API, add integration tests

The TODO_LIST.md has items for all three paths, but the next sprint should probably focus on one.

---

## Metrics Snapshot

| Metric | Value |
|--------|-------|
| Go source files | 135 |
| Test files | 91 |
| Total packages | 22 |
| Test packages passing | 22/22 (100%) |
| Avg test coverage | ~87% |
| Open TODOs (HIGH) | 3 |
| Open TODOs (MED) | 6 |
| Open TODOs (LOW) | 9 |
| ADRs | 3 |
| Languages supported | 2 (Go, Templ) |
| Output formats | 7 |
| Detection methods | 5 |
| Go version | 1.26.3 |
| Direct dependencies | 11 |
| Lint issues (new) | 0 |
| Lint issues (pre-existing) | 3 godoclint + 1 LSP warning |
| nix build | BROKEN (vendorHash stale) |
| just build | PASS |
| `dist/art-dupl` binary | Built and functional |
