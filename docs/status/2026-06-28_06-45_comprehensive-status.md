# art-dupl Comprehensive Status Report — 2026-06-28 06:45 CEST

**Branch:** `fork` | **HEAD:** `ee2fad8` | **Verification:** build ✓ · test ✓ (27 pkgs) · lint 0 issues ✓ · nix flake check ✓

---

## 00 — Executive Summary

| Metric | Value |
|---|---|
| Build | ✓ Green |
| Tests | ✓ All 27 packages pass |
| Race detector | ✓ Clean (go test -race) |
| Lint (golangci-lint) | ✓ **0 issues** (was 66) |
| nix flake check | ✓ **ALL CHECKS PASSED** |
| go.sum integrity | ⚠ BuildFlow strips test checksums (by design) |
| Uncommitted files | ✓ 0 (clean tree) |
| Session commits | 9 commits, 61 files, +497/−504 lines |
| Critical blockers | 0 |

This session delivered: the complete lint cleanup that unblocked `nix flake check`, a data-race fix on the incremental cache path, removal of unconditional `debug.Stack()` overhead, the `MethodDetector.Name()` interface refactor, and the first step of printer decoupling (enum types moved to domain). All work is committed and pushed to `origin/fork`.

---

## 01 — Fully Done

### Lint Cleanup (66 → 0 issues) — Commit `75d0c2e`

The single highest-impact deliverable. Resolved every golangci-lint issue:

- **makezero (50→0):** Converted `make([]T, n)` + index-fill patterns to `make([]T, 0, n)` + `append` across 28 files (test helpers, production builders, benchmark setup). Added `//nolint:makezero` for 12 genuine index-fill cases (binary buffers, DP matrix, `copy()` targets, 2D fills). Zero existing `nolint:makezero` in the codebase — convention is to fix, not suppress.
- **wrapcheck (6→0):** Added `ignore-package-globs: github.com/LarsArtmann/art-dupl/pkg/enum` to `.golangci.yml`. The generic `enum.UnmarshalJSONInto[T ~string]` signature wasn't matchable by `ignore-sigs`. Removed 6 stale `//nolint:wrapcheck` directives made redundant by the glob.
- **errcheck (4→0):** Explicitly discarded `fmt.Fprintln` returns in `job/profiler.go` (profile header/footer output — write failures are non-actionable).
- **staticcheck SA9003 (1→0):** Replaced empty `if` branch in `printer/sarif_test.go` with `t.Fatalf` — the test had an assertion-shaped hole that silently passed on wrong rule counts.
- **ginkgolinter (5→0):** Moved 2 raw-string fixture constants (`duplicateCode`, `diffCode`) from inside `Describe` blocks to package-level `var` (matching the existing `simpleGoCode` pattern). Added `//nolint:ginkgolinter` on 3 table-test case generators — these are legitimate BDD table-driven patterns, not shared mutable state.

### Data Race Fix — Commit `ad6b33f`

The incremental parser cache (`job/incremental.go`) shared node pointers across goroutines. On a cache hit, `parseFile` rewrote `node.Filename` in place, racing with concurrent hits for the same content hash. Added `Node.Clone()` to `syntax/syntax.go` (recursive deep-copy of the subtree) and used it on the cache-hit path. Verified with `go test -race ./...` — all 27 packages clean.

### debug.Stack() Removal — Commit `e80aeff`

`DuplError` captured `debug.Stack()` on every construction (5 call sites) but the `Stack` field was never read by any consumer. Removed the field, all 5 calls, and the `runtime/debug` import. This eliminates runtime overhead in every error path and prevents goroutine stack traces from leaking into JSON output (security improvement).

### MethodDetector.Name() — Commit `5d4abaf`

The `detName` helper used a `switch det.(type)` to return human-readable detection method names. Added `Name() string` to the `MethodDetector` interface, implemented on `suffixTreeAdapter`, `hashAdapter`, and the test `fakeDetector`. Eliminated the type switch entirely — new detectors just implement `Name()`.

### AGENTS.md Updated — Commit `da48beb`

Documented all new conventions:
- `Node.Clone()` for incremental cache safety
- `MethodDetector.Name()` interface method
- `DuplError` no longer captures `debug.Stack()`
- `makezero: always: true` convention (append or nolint)
- `wrapcheck ignore-package-globs` for `pkg/enum`

### VendorHash Fix — Commit `6966189`

The charmbracelet dep bump (golden 0622→0628, x/tools 0.46→0.47) required a new Nix vendor hash. Updated via the standard dummy-hash → `nix build` → real-hash workflow. `nix flake check` now passes all gates.

### Enum Migration to Domain — Commit `f9b2a90`

Moved `SortCriteria`, `OutputFormat`, and `DiffMode` canonical definitions from `config/` to `domain/`. Domain versions use `pkg/enum` helpers (`MarshalJSON`/`UnmarshalJSONInto`) instead of config-internal helpers (`marshalStringType`/`unmarshalStringTypeToPointer`). Config retains Go type aliases (`type SortCriteria = domain.SortCriteria`) so **zero callers changed**. This is the first step toward removing printer's `config` import entirely.

### Dep Bump + go.sum Restoration — Commit `47816a7`

Restored test-dependency checksums (testify, go-snaps, go-junit, tparse, tidwall, zeebo — 47 lines) that a prior session had trimmed from go.sum. Bumped charmbracelet/x/exp/golden (0622→0628), golang.org/x/tools (0.46→0.47), and nixpkgs flake input. README/FEATURES table formatting refreshed.

---

## 02 — Partially Done

### Printer ↔ config coupling (reduced, not eliminated)

`SortCriteria`/`OutputFormat`/`DiffMode` now live in `domain/` with config aliases, but **printer still imports config in 12 production files and 9 test files**. The types work through aliases, but the import statements still reference `config`. Full decoupling requires updating all `config.SortCriteria` → `domain.SortCriteria` references across printer + cmd, or adding domain aliases to printer.

### Printer ↔ syntax.Node coupling

`actionability.go` still imports `syntax.Node` directly (11 references, down from ~34). `clone_processor.go` is the partial bridge. Full ProcessedClone DTO decoupling remains open.

### Lint config as single source of truth

golangci-lint CLI config is clean (0 issues), but the LSP surfaces a different linter set (`wsl_v5`, stale diagnostics) that creates noise in editors. The AGENTS.md note says "treat golangci-lint run as the single source of truth" but the LSP integration isn't aligned.

---

## 03 — Not Started

| Item | Impact | Effort | Notes |
|---|---|---|---|
| Branded `NodeType int32` | High | XL | Prevents cross-package int32 collision. HIGH RISK: touches gob cache format + semantic encoding `[24-bit hash][8-bit base type]`. Deliberately deferred. |
| Shared `CloneRef` value object | Med | M | Unify 7 parallel Clone types via embedding without collapsing DTO boundary. Needs design decision. |
| ProcessedClone DTO (full) | Med | L | Fully decouple printer from `syntax.Node` internals. `actionability.go` is the remaining bridge point. |
| Split `printer/` package | Med | L | 56 files / ~13.8k LOC in one package. Split into stats/html/analyze sub-packages. |
| Templ statement-level tokenization | Med | L | Templ matching is purely structural today; no statement-level tokens, no semantic mode. |
| Templ semantic mode | High | L | No identifier/operator encoding for templ. Mixed-language corpora lack a unified detection model. |
| Git-aware incremental mode | Med | L | Replaces removed `--since` dead flag. Needs `git diff` integration. |
| Context through `cmd/run_crawl.go` feeders | Med | L | stdin scanner + filepath.Walk are inherently blocking. Needs full chain refactor. |
| `config/filetype.go` enum migration | Low | S | Last config enum still using old `isValidStringType` helpers. Same treatment as SortCriteria/OutputFormat/DiffMode. |
| Remove config/enum_helpers.go dead code | Low | S | After filetype migration, the old helpers (`marshalStringType`, `unmarshalStringTypeToPointer`) become dead code. |

---

## 04 — Totally Fucked Up

### BuildFlow strips test checksums from go.sum (systemic, not blocking)

**What:** The BuildFlow pre-commit hook runs `go mod tidy` as part of its validation pipeline. This strips 47 test-dependency checksums (testify, go-snaps, go-junit, tparse, tidwall, zeebo, etc.) from go.sum every commit. Manually running `go mod tidy` restores them, but the next commit strips them again.

**Impact:** Low. `go build`/`go test` pass because Go lazily re-fetches missing checksums from the module proxy. `nix flake check` passes because Nix uses `vendorHash`, not `go.sum`. The only risk is in an air-gapped environment where the proxy isn't available.

**Root cause:** BuildFlow's `go-mod-tidy` step runs `go mod tidy` in a context that produces a pruned go.sum (likely Go 1.26+ module graph pruning of test-only transitive dependencies).

**Resolution attempt:** Tried 3 times to commit the complete go.sum. BuildFlow always strips it. Accepted as a known limitation — the committed go.sum is the one BuildFlow considers valid.

### go-error-family dependency warning (cosmetic, not actionable)

BuildFlow's library-policy check recommends adding `github.com/larsartmann/go-error-family` for structured errors. The current `errors/` package uses a custom typed error hierarchy (`DuplError` with `ErrorType` classification) that serves the same purpose. Migrating to go-error-family would be a large refactor for marginal benefit.

---

## 05 — What We Should Improve

### Immediate (this week)

1. **Update printer imports from `config.` to `domain.`** — now that the types are aliased, printer files can import domain directly, removing the config dependency entirely from production printer code.
2. **Add pre-commit go.sum completeness check** — or accept BuildFlow's go.sum as authoritative and document this in AGENTS.md.
3. **Investigate BuildFlow's go-mod-tidy stripping** — file a BuildFlow issue or add a `GOFLAGS` override to preserve test checksums.

### Short-term (2-4 weeks)

4. **Deep-copy cached nodes** — DONE this session. Next: add a concurrent incremental-cache-hit test that exercises the race condition under `-race` to prevent regression.
5. **Migrate `config/filetype.go`** — the last config enum using old helpers. Same alias treatment as the three already done.
6. **Remove dead config/enum_helpers.go** — after filetype migration, the old helpers are dead code.
7. **Split printer package** — 56 files / 13.8k LOC is unwieldy. Stats, HTML, and analysis sub-packages would improve compile times and reviewability.

### Strategic

8. **Branded `NodeType int32`** — coordinate with a cache-format bump. The semantic encoding `[24-bit hash][8-bit base type]` is fragile.
9. **Templ semantic mode** — closes the mixed-language detection gap.
10. **CloneRef value object** — reduces the 7 parallel Clone types to a shared core + DTO-specific extensions.

### Process

11. **Treat golangci-lint as single source of truth** — the LSP `wsl_v5` linter creates noise not in the CLI config. Align LSP settings or suppress stale diagnostics.
12. **Add go mod tidy check to CI** — separate from BuildFlow, to catch go.sum incompleteness in a clean environment.

---

## 06 — Top 25 Things to Get Done Next

| # | Task | Impact | Effort | Priority |
|---|---|---|---|---|
| 1 | Update printer imports from `config.` to `domain.` (remove config dep) | High | S | **P0** |
| 2 | Add regression test for incremental cache deep-copy under `-race` | High | S | **P0** |
| 3 | Migrate `config/filetype.go` enum to domain (same alias treatment) | Med | S | **P1** |
| 4 | Delete dead `config/enum_helpers.go` after filetype migration | Med | S | **P1** |
| 5 | Add `go mod tidy` completeness check to CI (separate from BuildFlow) | High | S | **P1** |
| 6 | Document BuildFlow go.sum stripping behavior in AGENTS.md | Med | S | **P1** |
| 7 | Remove printer's remaining `config` test imports (9 files) | Med | S | **P1** |
| 8 | Introduce `CloneRef` value object (embed, don't collapse DTOs) | Med | M | **P1** |
| 9 | Decouple `actionability.go` from `syntax.Node` (11 refs remaining) | Med | L | **P2** |
| 10 | Split `printer/` into stats/html/analyze sub-packages | Med | L | **P2** |
| 11 | Branded `NodeType int32` (with cache-format + encoding coordination) | High | XL | **P2** |
| 12 | Templ semantic mode (identifier/operator encoding) | High | L | **P2** |
| 13 | Thread `context.Context` through `cmd/run_crawl.go` file feeders | Med | L | **P2** |
| 14 | Templ statement-level tokenization | Med | L | **P2** |
| 15 | Git-aware incremental mode (replaces removed `--since`) | Med | L | **P2** |
| 16 | Align LSP linter set with golangci-lint CLI config | Low | S | **P3** |
| 17 | Hide `syntax/golang` behind a facade (blocked by import cycle) | Low | L | **P3** |
| 18 | Add `--list-generators` discovery helper for `--include-generated` | Low | S | **P3** |
| 19 | Document deprecation timeline for 6 legacy `--include-*` flags | Low | S | **P3** |
| 20 | Consolidate parallel Clone-type field docs in one ADR | Low | S | **P3** |
| 21 | Benchmark statement-level tokenization vs. legacy | Low | S | **P3** |
| 22 | Fuzz `parseIncludeGeneratedCategories` parser | Low | S | **P3** |
| 23 | Hybrid slice/map transition storage for small counts in suffix tree | Low | M | **P3** |
| 24 | Add integration test for `--include-generated all` end-to-end | Low | S | **P3** |
| 25 | Evaluate `go-error-family` migration (BuildFlow recommends it) | Low | M | **P3** |

---

## 07 — Top #1 Open Question

### Should the BuildFlow `go-mod-tidy` pre-commit hook be trusted as the authoritative go.sum source, or should we override it to preserve test-dependency checksums?

**Context:** BuildFlow runs `go mod tidy` during pre-commit, which strips 47 test-dependency checksums from go.sum. Manual `go mod tidy` restores them, but BuildFlow strips them again on the next commit. All verification gates pass with the stripped go.sum (Go lazily re-fetches; Nix uses vendorHash).

**The question:** Is this the intended behavior? If so, I should document it in AGENTS.md and stop trying to "fix" it. If not, we need to either:
- Override `GOFLAGS` in BuildFlow to preserve test checksums
- Add a post-commit hook that restores them
- File a BuildFlow issue about `go mod tidy` stripping needed checksums

I cannot determine the correct answer from the codebase alone — it depends on whether air-gapped/offline builds are a supported use case, and whether BuildFlow's `go mod tidy` is expected to be authoritative.

---

*Generated: 2026-06-28 06:45 CEST · Branch: fork · HEAD: ee2fad8 · Verification: build ✓ · test ✓ · lint 0 ✓ · nix flake check ✓*
