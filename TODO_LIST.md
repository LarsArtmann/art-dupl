# TODO List

**Last Updated:** 2026-07-19

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md`.

---

## ✅ Recently Completed (Tier 1 Feedback Sprint — 2026-07-19)

From `docs/feedback/2026-07-19_critical_evaluation_of_feedback.md`:

- [x] **Text output: one-line code preview per group** — `printer/text.go` `previewFirstLine()` adds `| <first source line>` after each clone in text and `--rich-text` modes. Plumbing unchanged. 7 unit tests in `printer/text_test.go`.
- [x] **Skill doc: recommend `-t 25` for test-heavy libraries** — `deduplicate-code/SKILL.md` now has a "Test-heavy libraries" note with `-t 25` and `--exclude-pattern '*_test.go'` guidance.
- [x] **SortCloneGroups direct test coverage** — `printer/sort_unified_test.go::TestSortCloneGroups_PublicAPI` (4 subtests) covers the public wrapper that delegates to the shared `cloneGroupMetrics` var.
- [x] **Acceptance comments on deliberate test fixtures** — 3 test files (`bdd/exit_codes_test.go`, `printer/overlap_test.go`, `printer/semantic_precision_test.go`) now carry `// art-dupl: accepted: <rationale>` markers so future dedup runs surface them as intentional.

---

## 🔴 Deferred — Architecturally Constrained

These items are blocked by fundamental design constraints and cannot be resolved without significant architectural changes.

- [ ] **Split `printer/` into sub-packages** (stats, html, analyze) — ~29 source files / ~3500+ lines. Requires interface inversion: core `printer.go` references `StatsPrinter`, creating circular deps with any sub-package. Clean split requires moving `Printer`/`ReadFile`/`StatsPrinter` interfaces to a separate base package + extracting shared test helpers from `_test.go` files + handling `.(*stats)` type assertions to unexported types.
- [ ] **Branded `NodeType int32`** — A single `syntax.NodeType` type does NOT prevent cross-package constant value collision (golang and templ constants would share the same type). The proper fix requires per-package `NodeType` types (`golang.NodeType`, `templ.NodeType`), which is even more invasive. Current 8-bit shared encoding space is intentional (see ADR-0008). **HIGH RISK**: touches gob cache format.
- [ ] **Hide `syntax/golang` behind facade** — **BLOCKED** by import cycle (syntax/golang imports syntax for Node type).

---

## 🟡 MEDIUM Priority

### Code Quality

- [ ] Implement hybrid slice/map transition storage for small transition counts in suffix tree (deferred — map already O(1)).

### Testing

- [x] **Fix `RunArtDuplWithStdin` to exercise real stdin** — Refactored to inject a pipe reader via `os.Stdin` replacement + `--files` flag, exercising `feedFromStdin` end-to-end.
- [x] **Add `feedFromStdin` filter integration test** — `TestStdinFeed_FiltersGeneratedCode` verifies `gogenfilter.Filter` + `FileTypeGo` end-to-end through stdin.
- [x] **Add stderr suppression test for `feedFromStdin`** — `TestStdinFeed_StderrSuppressedOnCancel` verifies the `ctx.Err() == nil` guard.
- [x] **Add stdin timeout cancellation test** — `TestStdinFeed_TimeoutUnblocksScanner` verifies `context.WithTimeout` unblocks the scanner.
- [x] **Progress output for long runs** — `cmd/progress.go` wraps file channel with periodic count to stderr (every 5s + final count). Suppressed by `--quiet` and `ARTDUPL_NO_PROGRESS=1`.
- [x] **True exit code integration test** — `TestExitCodes_Process` builds real binary and checks actual process exit codes.
- [x] **Test `--workers 0` auto-detection** — Fixed `> 1` → `!= 1` routing bug (0 was falling to sequential). Added `TestWorkers_AutoDetection`.
- [x] **Test `NewProcessedCloneGroup` constructor** — `TestNewProcessedCloneGroup` verifies TokenCount computation, empty clones, and Validate() pass.

### CLI & UX

- [ ] **`--help` text audit** — Verify every flag description is accurate and includes valid values.
- [ ] **YAML config file support** — `.artdupl.yml` parser alongside existing JSON support.
- [ ] **Cross-link `docs/ACTIONABILITY_PATTERNS.md`** — Link from HOW_TO_USE.md and AGENTS.md.
- [ ] **SDK: expose `ExitCodeForError` and `VersionInfo`** — Consider adding to `pkg/artdupl` for library consumers.
- [ ] **Shell completion end-to-end test** — Verify `art-dupl completion bash` produces valid bash script.
- [ ] **Deprecation warning for `--semantic`** — It's the default now; flag is redundant but not deprecated.

### Performance

- [ ] **Profile-guided optimization** — Run with `--profile` and check for hot spots.
- [ ] **Memory usage for large repos** — Test on 10000+ files.
