# TODO List

**Last Updated: 2026-07-16**

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md`.

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

- [ ] **Fix `RunArtDuplWithStdin` to exercise real stdin** — BDD helper (`internal/testutil/bdd_runners.go:109`) converts stdin content to CLI args, never exercising the `feedFromStdin` code path. Now that `feedFromStdin` accepts `io.ReadCloser`, refactor the helper to inject a pipe reader for real stdin coverage.
- [ ] **Add `feedFromStdin` filter integration test** — Current unit tests pass `nil` filter and `""` file type. Add a test with an active `gogenfilter.Filter` and `FileTypeGo` to verify the filtering pipeline works end-to-end through stdin.
- [ ] **Add stderr suppression test for `feedFromStdin`** — The `ctx.Err() == nil` guard on `run_crawl.go:83` (suppressing scanner errors from forced close) is untested.
- [ ] **Add stdin timeout cancellation test** — Verify `context.WithTimeout` expiry unblocks the scanner while reading.
- [ ] **Progress output for long runs** — Add file-count or spinner progress to stderr when not `--quiet`.
- [ ] **True exit code integration test** — Run `exec.Command` and check actual process exit code, not just `ExitCodeForError` return value.
- [ ] **Test `--workers 0` auto-detection** — Verify 0 defaults to NumCPU.
- [ ] **Test `NewProcessedCloneGroup` constructor** — Direct unit test for the constructor computing `TokenCount`.

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
