# Status Report — Master Plan Wave 3: Signal & Hardening

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 30 items verified resolved, 8 open (wave-3 landed; final gates + T2.2 A/B completed 2026-09-22; micro-tails open).

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 30 items verified resolved, 8 open (wave-3 landed; final gates + T2.2 A/B completed 2026-09-22; micro-tails open).

**Date**: 2026-08-16 13:13 CEST
**Branch**: `fork` @ `7f98ed10` (daemon commits ahead of session start `8c1f75cf`)
**Plan**: `docs/planning/2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md`
**Session scope**: resumed mid-T12 (build broken) → T12, T19, T16, T17, T18, T20, T21, T22 complete; T24 just started (interrupted by user stop).

**Overall**: 24 of 26 plan tasks complete or no-go-verified. Remaining: T24 (doc polish, 3 small items), T25/T26 (docs sync), T2.2/T23 (timing evidence — blocked on machine load), AGENTS.md updates, final gates.

---

## a) FULLY DONE (this session)

| Task                                           | What shipped                                                                                                                                                                                                                                                                                                                                                                                                    | Verified by                                                                                                                  |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **T12** `--exclude-pattern` zero-match warning | Import fix (build was broken), warning via `WarnUnmatchedExcludePatterns`, tracking wired into walk + stdin + dump-tokens paths, 5 unit tests, HOW_TO_USE "Pattern Semantics" section, flag help rewritten                                                                                                                                                                                                      | E2E all 4 semantics (regex-style warns, matched glob silent, path-glob silent, stdin warns); `go test ./cmd/`; lint 0 issues |
| **T19** cleanup trio                           | 19.1 verified `benchmarkFindTranMethod` **NOT dead** (3 call sites) → kept, no deletion (plan said "verify unused → delete"; verification failed, so delete is wrong). 19.2 `b.Loop()` ×3 in `parallel_bench_test.go`. 19.3 fixed AGENTS "Workers routing" bullet — old text was factually BACKWARDS ("Never use >1, that sends 0 to sequential"); truth: `1` is the only sequential value, `0` = auto/parallel | grep call sites; bench smoke runs; code read of `normalizeWorkerCount` + gating                                              |
| **T16** coverage baseline                      | `scripts/check-coverage.sh` (snapshot + `--print`), `docs/benchmarks/coverage-baseline.txt` (**75.2% total**), README table row                                                                                                                                                                                                                                                                                 | script executed, output reviewed; all packages green at capture time                                                         |
| **T17** TTY HTML auto-write                    | `resolveHTMLOutput`/`openHTMLOutput` rewrite: non-HTML formats always stdout (kills the old silent empty-file bug when `--html-out` was set with `--json`); explicit flag wins; TTY → auto-write `art-dupl-report.html` + stderr notice; piped → stdout. `charmbracelet/x/term` promoted to direct dep. 5 tests (`html_output_test.go`)                                                                         | PTY E2E via `script -qec`: notice + 19KB file; pipe E2E: HTML on stdout; explicit flag on PTY: no notice, no default file    |
| **T18** stable display IDs                     | `groupAnchorID(hash, num)` (sanitize + `group-` prefix + positional fallback for degenerate empty hash), `AnchorID` on `CloneGroupView`, templ `id` + clickable `#` permalink with `stopPropagation`, CSS, goldens regenerated. 4 tests incl. reorder-stability and uniqueness                                                                                                                                  | unit tests + E2E real output shows `id="group-e0f6093241ba9931"` + matching `href`                                           |
| **T20** slice-vs-map property test             | `tran_parity_test.go`: naive `referenceFindTran`, BFS state enumeration, invariant checks (strictly ascending + unique keys) and parity for present/near-miss/far probes. 5 adversarial deterministic streams + **1000 seeded random streams** (alphabets 2–41). Coverage guard asserts the binary-search branch was actually exercised                                                                         | all green; guard nonzero                                                                                                     |
| **T21** fuzz hardening                         | High-fanout seeds (full-byte ramps 64/200, interleaved ramp) added to `FuzzFindDuplOver`; NEW `FuzzTranLookupSemantics` fuzz target verifying findTran-vs-reference SEMANTICS (not just no-panic)                                                                                                                                                                                                               | 30s run: **2.69M execs, 0 failures**, 74 interesting inputs                                                                  |
| **T22** linearScanMax boundary bench           | `BenchmarkFindTranBoundary8`/`Boundary9` (exact-8 / exact-9 transitions via direct `addTran` fixture), `TestBoundaryStatesUseExpectedBranch` fixture-straddle guard                                                                                                                                                                                                                                             | 8: ~57–60ns/op (16 probes), 9: ~69–75ns/op (18 probes) — per-probe costs comparable, cutoff at 8 remains justified           |

Prior-session work carried and staged: T5.1b (flake check ALL GREEN), T13 (serialization 10,015→2 allocs), T15 (syntax budgets + gate), T11 (run-scoped cache stats), status report 12-27.

## b) PARTIALLY DONE

- **T24** suffixtree doc polish — **just started, interrupted**:
  - 24.1 package doc refresh: new text drafted, **edit NOT applied** (first attempt failed: auto-commit daemon touched `suffixtree.go` mtime mid-edit; re-read confirms content unchanged, diff vs HEAD empty — redo the edit).
  - 24.2 `maxStackKeys`/`getAll` dedup evaluation: analysis done (position lists are disjoint by construction — each Pos is filed under its unique preceding token; dedup unnecessary) — **not yet documented in code**.
  - 24.3 `BenchmarkMemoryUsage` magic number `10000`: **not yet named**.
- **AGENTS.md updates owed**: T13 arena serialization (incl. field-preservation hazard), T11 run-scoped cache stats, T12 warning, alloc-gate syntax/ coverage — none written yet (daemon's own AGENTS edits are staged separately, not mine to judge).

## c) NOT STARTED

- **T25**: TODO_LIST checkbox sync + CHANGELOG entries (T7–T12, T13, T15, T17–T18, T19–T22).
- **T26**: parked-tier revisit triggers into TODO_LIST.
- **T2.2**: interleaved pinned/unpinned A/B + v3-notes annotation — deferred: foreign load (nixbld1 C++ builds, govulncheck, service.test) had machine at load-74 at 12:38, now ~27 and falling.
- **T23**: `perf stat` LLC-miss A/B vs `23fa1b4f` worktree — same deferral.
- **Final gates**: full build + tests + `-race` + full-repo lint + alloc gate + `nix flake check` (must stage everything first — ~40 min).

## d) TOTALLY FUCKED UP (self-caught, all fixed before staging)

1. **Broken first draft of `TestBoundaryStatesUseExpectedBranch`** — wrote a half-finished loop calling `findTran(nil, ...)` with dead code; caught it on review and rewrote cleanly.
2. `tran_parity_test.go` first literal `tokensOf({...})` — invalid Go array-literal syntax; compile error, fixed with `[]TokenValue{...}`.
3. `html_anchor_test.go` first draft had dummy `var _ = bytes.MinRead` import-keeper hacks — removed (embarrassing pattern).
4. First multiedit on `run_flags.go` imports guessed the import block wrong (1 of 2 edits failed) — fixed by reading actual imports.
5. wsl_v5/gci/golines lint failures in first `tran_parity_test.go` — restructured.
6. T24.1 edit failed against daemon-touched file — pending redo, not lost.

Nothing broken shipped; every failure was caught by compile/test/lint/self-review before staging.

## e) WHAT WE SHOULD IMPROVE

- **Write tests file-complete on first pass** — 3 of my 6 fuckups were "drafted then immediately rewrote". Think the whole file through before writing.
- **HTML TTY auto-write lacks a BDD test** — unit tests cover the injected decision function; the real TTY branch was only manually PTY-tested.
- **`FuzzTranLookupSemantics` corpus**: 74 interesting inputs from the 30s run live in build cache, NOT in `testdata/fuzz/` — commit a seed corpus so CI regress runs cover the interesting space.
- **Boundary bench numbers are unpinned** (machine loaded) — record a pinned benchstat when idle.
- **Coverage baseline rows for test-only packages** (`internal/testutil` 0.0%, `testhelpers` 0.0%, `examples` 0.0%) add noise — consider a filter in the script.
- **T12 warning ignores `--quiet`** — warnings print unconditionally (consistent with other warnings, but worth a deliberate decision).

## f) NEXT — ordered queue (up to 50)

**Finish the wave:**

~~1. T24.1: apply the drafted package-doc refresh on `suffixtree.go`~~ done at 2026-08-16_17-37 a1 — applied next session
~~2. T24.2: document the getAll disjointness argument as a comment~~ done at 2026-08-16_17-37 a2 — comment in suffixtree/dupl.go
~~3. T24.3: `const memoryUsageTokens = 10000` (+ Few/ManyTokens in `memory_bench_test.go`)~~ done at 2026-08-16_17-37 a3 — const named; re-named with Few/Many constants 2026-09-28
~~4. Commit FuzzTranLookupSemantics seed corpus to `testdata/fuzz/`~~ done — suffixtree/testdata/fuzz/FuzzTranLookupSemantics/ verified present
~~5. AGENTS.md: T13 arena serializer bullet (+ serial() field-preservation hazard cross-ref)~~ done at 2026-08-16_17-37 a5 — bullet live in AGENTS
~~6. AGENTS.md: T11 run-scoped cache stats bullet~~ done at 2026-08-16_17-37 a5 — bullet live in AGENTS
~~7. AGENTS.md: T12 exclude-pattern warning bullet~~ done at 2026-08-16_17-37 a5 — bullet live in AGENTS
~~8. AGENTS.md: alloc-gate syntax/ coverage bullet~~ done at 2026-08-16_17-37 a5 — alloc-gate syntax/ coverage in AGENTS
~~9. AGENTS.md: T17 TTY auto-write bullet~~ done at 2026-08-16_17-37 a5 — HTML output resolution bullet in AGENTS
~~10. AGENTS.md: T18 stable anchors bullet~~ done at 2026-08-16_17-37 a5 — stable anchors bullet in AGENTS
~~11. T25.1: TODO_LIST checkbox sync (T1–T22 done states)~~ done at 2026-08-16_17-37 a6 — TODO_LIST rewritten open-only
~~12. T25.2: CHANGELOG entries (T7–T12, T13/T15, T17–T18, T19–T22)~~ done at 2026-08-16_17-37 a8 — CHANGELOG entries landed
~~13. T26.1: parked-tier revisit triggers (DEFERRED/ROADMAP → entry criteria)~~ done at 2026-08-16_17-37 a7 — parked-tier entry criteria table in TODO_LIST
~~14. Full-repo `golangci-lint run --timeout 5m ./...` (only touched packages done)~~ done — full-repo lint green in the completion wave; standard gate since
~~15. Full `go build ./... && go test ./...`~~ done at 2026-08-16_17-37 a11 — full build+test green; race now runs on every flake check (AGENTS 2026-08-16)
~~16. `go test -race ./...`~~ done — race cadence on every nix flake check (AGENTS 2026-08-16 decision)
~~17. `scripts/check-alloc-regression.sh` (gate green re-confirm)~~ done — scripts/check-alloc-regression.sh wired as the flake alloc-gate check
~~18. Stage everything, then `nix flake check` in background (~40 min)~~ done — T5.1b flake check ALL GREEN; CI flake lane green since

**Timing evidence (when machine idle):**
~~19. T2.2: interleaved P/U A/B (`FindDuplOver/threshold_10`, `par4/tokens_10000`, 6 alternations)~~ done — docs/benchmarks/pinned-unpinned-2026-09-22.txt (30 samples/arm)
~~20. T2.2: annotate `baseline-2026-08-16-v3_notes.md` (replace "still open, see TODO_LIST")~~ done — baseline-2026-08-16-v3_notes.md carries the pinning verdict
~~21. T23: `git worktree add /tmp/artdupl-23fa1b4f 23fa1b4f`~~ done — A/B performed 2026-09-22 (see pinned-unpinned file header)
~~22. T23: `perf stat -e cache-references,cache-misses,LLC-load-misses` A/B on suffixtree bench~~ done — ADR-0022 addendum "Cache-counter A/B vs 23fa1b4f" (docs/adr/0022:157)
~~23. T23: record verdict in `baseline-2026-08-16-v3_notes.md` + ADR-0022 addendum~~ done — ADR-0022:142 addendum with perf stat evidence
~~24. Pinned `benchstat` run for Boundary8/9, record in docs/benchmarks~~ done — docs/benchmarks/linear-scan-boundary-2026-09-22.txt (pinned, 20 samples/size)
~~25. Document 8-vs-9 result as a comment on `linearScanMax`~~ done — linearScanMax doc comment cites the measured 8-vs-9 numbers (suffixtree/findtran.go:3-19)

**Quality follow-ups:**
~~26. BDD/PTY test for HTML auto-write~~ resolved by design — the TTY probe is injected and unit-tested (AGENTS "HTML output resolution"); the real branch is a third-party one-liner, manually PTY-verified at ship
~~27. Decide `--quiet` vs T12 warning semantics (document choice)~~ done — decision recorded: diagnostics print unconditionally; --quiet suppresses progress only (AGENTS --quiet + include/exclude warning bullets)
~~28. Symmetric zero-match warning for `--include-pattern`~~ done — WarnUnmatchedIncludePatterns shipped (AGENTS include-pattern zero-match warning)
~~29. Coverage script: filter test-only packages~~ done — scripts/check-coverage.sh:29 excludes test-only packages
~~30. Consider coverage trend note in TESTING.md~~ done this pass — TESTING.md Coverage section now documents the trend baseline workflow
~~31. TESTING.md: property/parity-test conventions section~~ done — TESTING.md "Property and Parity Test Conventions" section
~~32. a11y: focus style for `.anchor-link`~~ done — .anchor-link:focus-visible rule in printer/html_template.go:102
~~33. JSON output: expose stable anchor id for cross-format linking~~ done — JSONClone.AnchorID serialized as anchor_id (printer/json.go:27)
~~34. `--html-out`: mkdir -p parent dir on demand~~ done — output dir MkdirAll in cmd/run_all_modes.go:27
~~35. Check `website/package.json` pre-existing modification (not mine — read before judging)~~ done at 2026-08-16_17-37 a10 — daemon commit 9d7f9813, deliberate deps bump
~~36. Verify goldens diff is CSS+anchor-only (review staged `printer/testdata`)~~ done at 2026-08-16_17-37 a9 — goldens render header+footer only by construction
~~37. Re-check corpus count 2665→2670 drift question (carry-over)~~ done — resolved inline in 2026-08-16_17-37 f22 (concurrent go-cqrs-lite edits; deterministic)
~~38. Budget-ratchet decision: 4783 < 4802 headroom (carry-over)~~ done — budgets re-recorded deterministically in scripts/alloc-budgets.txt after the time-seed fix (FewTokens=23, ManyTokens=44)

**Then:** wave-3 status report refresh, park T23 verdict if inconclusive, resume ROADMAP triage.

## g) QUESTIONS FOR YOU (max 3)

1. **Machine load**: foreign `nixbld1` C++ builds + `govulncheck` + `service.test` had the box at load-74 (now ~27). Are those yours and roughly when does it free up? T2.2/T23 timing evidence is parked until idle — or say the word and I run them noisy with caveats.
2. **Coverage gate ambition**: baseline is trend-only (75.2%). Do you want a per-package floor later (flake check), or keep it advisory permanently?
3. **T12 warning under `--quiet`**: currently warnings always print (like other warnings). Suppress under `--quiet`, or keep unconditional because it points at where output went / a misconfiguration?
