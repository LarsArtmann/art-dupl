# Status Report — Master Plan Wave 3: Signal & Hardening

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

1. T24.1: apply the drafted package-doc refresh on `suffixtree.go`
2. T24.2: document the getAll disjointness argument as a comment
3. T24.3: `const memoryUsageTokens = 10000` (+ Few/ManyTokens in `memory_bench_test.go`)
4. Commit FuzzTranLookupSemantics seed corpus to `testdata/fuzz/`
5. AGENTS.md: T13 arena serializer bullet (+ serial() field-preservation hazard cross-ref)
6. AGENTS.md: T11 run-scoped cache stats bullet
7. AGENTS.md: T12 exclude-pattern warning bullet
8. AGENTS.md: alloc-gate syntax/ coverage bullet
9. AGENTS.md: T17 TTY auto-write bullet
10. AGENTS.md: T18 stable anchors bullet
11. T25.1: TODO_LIST checkbox sync (T1–T22 done states)
12. T25.2: CHANGELOG entries (T7–T12, T13/T15, T17–T18, T19–T22)
13. T26.1: parked-tier revisit triggers (DEFERRED/ROADMAP → entry criteria)
14. Full-repo `golangci-lint run --timeout 5m ./...` (only touched packages done)
15. Full `go build ./... && go test ./...`
16. `go test -race ./...`
17. `scripts/check-alloc-regression.sh` (gate green re-confirm)
18. Stage everything, then `nix flake check` in background (~40 min)

**Timing evidence (when machine idle):**
19. T2.2: interleaved P/U A/B (`FindDuplOver/threshold_10`, `par4/tokens_10000`, 6 alternations)
20. T2.2: annotate `baseline-2026-08-16-v3_notes.md` (replace "still open, see TODO_LIST")
21. T23: `git worktree add /tmp/artdupl-23fa1b4f 23fa1b4f`
22. T23: `perf stat -e cache-references,cache-misses,LLC-load-misses` A/B on suffixtree bench
23. T23: record verdict in `baseline-2026-08-16-v3_notes.md` + ADR-0022 addendum
24. Pinned `benchstat` run for Boundary8/9, record in docs/benchmarks
25. Document 8-vs-9 result as a comment on `linearScanMax`

**Quality follow-ups:**
26. BDD/PTY test for HTML auto-write
27. Decide `--quiet` vs T12 warning semantics (document choice)
28. Symmetric zero-match warning for `--include-pattern`
29. Coverage script: filter test-only packages
30. Consider coverage trend note in TESTING.md
31. TESTING.md: property/parity-test conventions section
32. a11y: focus style for `.anchor-link`
33. JSON output: expose stable anchor id for cross-format linking
34. `--html-out`: mkdir -p parent dir on demand
35. Check `website/package.json` pre-existing modification (not mine — read before judging)
36. Verify goldens diff is CSS+anchor-only (review staged `printer/testdata`)
37. Re-check corpus count 2665→2670 drift question (carry-over)
38. Budget-ratchet decision: 4783 < 4802 headroom (carry-over)

**Then:** wave-3 status report refresh, park T23 verdict if inconclusive, resume ROADMAP triage.

## g) QUESTIONS FOR YOU (max 3)

1. **Machine load**: foreign `nixbld1` C++ builds + `govulncheck` + `service.test` had the box at load-74 (now ~27). Are those yours and roughly when does it free up? T2.2/T23 timing evidence is parked until idle — or say the word and I run them noisy with caveats.
2. **Coverage gate ambition**: baseline is trend-only (75.2%). Do you want a per-package floor later (flake check), or keep it advisory permanently?
3. **T12 warning under `--quiet`**: currently warnings always print (like other warnings). Suppress under `--quiet`, or keep unconditional because it points at where output went / a misconfiguration?
