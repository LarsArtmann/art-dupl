# Status Report — Wave 3 Completion: T24/T25/T26 + Final Gates (partial)

**Date:** 2026-08-16 17:37 CEST
**Branch:** `fork` @ `7f98ed10`
**Machine:** idle again — load 2.30/2.49/3.32 (foreign nixbld/govulncheck load from 12:38 is GONE; T2.2/T23 are now unblocked)
**Scope of this report:** this session only (resumed from `docs/status/2026-08-16_13-13_master-plan-wave3-signal-hardening.md`).

---

## a) FULLY DONE (this session)

1. **T24.1 — suffixtree package-doc refresh** (`suffixtree/suffixtree.go:1-16`): applied the drafted 8-line package comment. Documents arena-allocated states, value-transitions, Ukkonen sequential construction, `FindDuplOverParallel` disjoint-subtree dispatch, complexity, and the parity-test enforcement pointer. (The prior session's failed edit — daemon-touched mtime — retried cleanly on the first attempt this time.)
2. **T24.2 — `getAll()` disjointness comment** (`suffixtree/dupl.go`): documents WHY no dedup is needed: every `Pos` is filed under exactly one map key (its unique preceding token), at leaf coverage and through `append`'s per-key merge, so concatenated position lists are disjoint by construction.
3. **T24.3 — memory-bench magic numbers** (`suffixtree/suffixtree_bench_test.go`, `memory_bench_test.go`): `const memoryUsageTokens = 10000` named and used by `BenchmarkMemoryUsage` AND `benchmarkMemoryUsage` (Few/ManyTokens). Bonus consistency: memory_bench_test.go converted `b.N`→`b.Loop()` (matching T19.2's sweep; removed the now-redundant `b.ResetTimer()`). Verified by running all three benchmarks.
4. **Fuzz seed corpus committed** — `suffixtree/testdata/fuzz/FuzzTranLookupSemantics/` (74 files, 300K) copied from the build cache; seeds now run in every plain `go test` (verified: `FuzzTranLookupSemantics/*` subtests PASS).
5. **AGENTS.md — 6 bullets** owed from the plan: (1) rewrote the `serial()` field-preservation hazard bullet around the T13 `nodeSerializer` arena (hazard now lives in the struct literal; `countSerializedNodes`/`serial` must traverse identically); (2) run-scoped cache stats (T11) semantics; (3) exclude-pattern zero-match warning (T12) + why `*_gen.go` matches but anchored globs may not; (4) alloc-gate extended to `syntax/` (T15) with budgets file path; (5) HTML output resolution order (T17); (6) stable anchor derivation (T18).
6. **T25.1 — TODO_LIST sync**: fully rewritten. Removed everything completed by the plan (HTML remainder trio, correctness hardening trio, cache stats, exclude UX, coverage baseline, suffixtree follow-ups T13/T14/T15/T20/T21/T22, doc trio, cleanup trio). What remains is honest open work only.
7. **T26.1 — parked-tier revisit triggers**: TODO_LIST now carries the explicit entry-criteria table (mirrors plan §5) + an "Architecturally constrained (DEFERRED)" section including the measured T14 NO-GO with its re-evaluation trigger.
8. **T25.2 — CHANGELOG entries**: ~16 Added (`--timing` stage report, 3 patterns + T10 corpus numbers, run-scoped cache stats, zero-match warning, coverage baseline, TTY auto-write, stable anchors, parity property tests, fuzz seeds corpus, boundary benchmarks, Clear() race test, alloc gate), 3 Changed (arena serialization with numbers, b.Loop() sweep, plus existing layout bullet), 4 Fixed (htmlprinter race, ReadMemStats self-test dupe, alloc-gate name-matching bug, cmd test parallelism race). All cross-checked against actual commits/labels (e.g. verified `testmain-boilerplate`/`embed-directive` labels, T7's evidence-based NO-matcher verdict).
9. **Goldens diff mystery RESOLVED** (follow-up #36 from the 13-13 report): the golden tests render only `PrintHeader`+`PrintFooter` — no clone groups — so a CSS-only diff is correct by construction. Anchor markup in the group body is covered by `printer/html_anchor_test.go` (derivation, stability, deep-linkability, uniqueness).
10. **`website/package.json` mystery RESOLVED** (follow-up #35): NOT an uncommitted mod anymore — the daemon committed it as `9d7f9813` (deliberate deps bump, 07:22). Working tree is clean for `website/`; nothing rides along.
11. **Full build + full test suite GREEN**: `GOEXPERIMENT=jsonv2 go build ./...` + `go test ./...` — zero failures across all 24 packages (incl. new fuzz seeds).

## b) PARTIALLY DONE

- **Final gates**: build+tests done (11). Still to run: full-repo `golangci-lint`, `go test -race ./...`, `scripts/check-alloc-regression.sh`, staged `nix flake check` (~40 min). The `.go` files edited this session (`suffixtree.go`, `dupl.go`, 2 bench files) have NOT been individually linted or gofmt-checked yet.

## c) NOT STARTED (this session, by design — queued next)

- T2.2 interleaved pinned/unpinned A/B + `baseline-2026-08-16-v3_notes.md` annotation — **now unblocked** (load 2.3).
- T23 `perf stat` LLC A/B vs `23fa1b4f` worktree + ADR-0022 addendum — now unblocked.
- Pinned benchstat for `Boundary8/9` + `linearScanMax` justification comment.
- Wave-3 quality follow-ups (BDD/PTY HTML test, `--quiet` semantics, include-pattern warning, coverage-script filtering, TESTING.md conventions, a11y focus ring, JSON anchor id, `--html-out` mkdir -p).
- Carry-over questions: corpus drift 2665→2670, budget ratchet 4783<4802.

## d) TOTALLY FUCKED UP (honest)

1. **AGENTS.md multiedit clobbered a bullet**: my third edit's `old_string` was the entire "Clone categories" bullet (intended as an insertion anchor) — the replacement deleted it. Caught one tool call later by scanning the diff, restored immediately. Lesson: for insert-before/after operations, put the anchor bullet in BOTH old and new strings, not just old.
2. **First `getAll()` comment draft was subtly wrong**: referenced "see append/start below" as if `start` were a method — it's a local variable in a closure 50 lines away. Rewrote to describe the mechanism (leaf coverage + per-key merge) rather than point at a symbol. Lesson: comments that cite symbols must cite symbols that exist at the reader's vantage point.
3. **CHANGELOG T18 entry shipped a garbled phrase** ("click-to-copy-free plain anchors") in my first pass — pure word salad; fixed to describe the real behavior (stopPropagation so the header toggle is unaffected).
4. **Report-writing order**: I ran the full test suite but did NOT kick off lint/race/flake before the user's STOP — the gates are the exact thing left dangling, and staging for flake check still hasn't happened. A tighter session would have started the 40-min flake check first (after staging), run lint/race in its shadow, and been fully done.

## e) WHAT WE SHOULD IMPROVE

- **Golden tests cover only the report shell**: `TestHTMLOutputGolden` renders header+footer, never a clone group. All body-level HTML (anchors, diff views, code blocks) is golden-untested. A golden with one real group would have made the T18 change's HTML diff reviewable in the golden itself.
- **Fuzz corpus was one `cp` away from being lost**: 74 interesting inputs lived only in the build cache (`/mnt/buildcache`!) until today. Any fuzz target whose seeds matter should get its corpus committed the day the target lands. Consider a tiny check: fuzz targets with zero `testdata/fuzz/<Name>` seeds are suspect.
- **The `--timing`/profiler duplication class**: the canonical self-test caught the ReadMemStats dupe — good — but the gate only runs on flake check (40 min). The threshold-1 self-scan is cheap enough to be a unit test.
- **Doc-sync tasks (T25/T26) keep landing LAST**: two sessions in a row the living docs lagged the code by hours. Cheaper: write the CHANGELOG bullet the moment a task completes, not in a batch.

## f) NEXT — ordered queue (up to 50, realistic for next session(s))

1. `gofmt -l` on edited files; fix if dirty
2. Full-repo `golangci-lint run --timeout 5m ./...`
3. `go test -race ./...`
4. `scripts/check-alloc-regression.sh` re-confirm green
5. `git add -A`, then background `nix flake check` (~40 min)
6. **T2.2** interleaved A/B: `taskset -c 0-7,16-23` vs unpinned, ≥6 alternations, `FindDuplOver/threshold_10` + `par4/tokens_10000`
7. Annotate `docs/benchmarks/baseline-2026-08-16-v3_notes.md` with the verdict (replaces "still open")
8. **T23** `git worktree add /tmp/artdupl-23fa1b4f 23fa1b4f`
9. T23 `perf stat -e cache-references,cache-misses,LLC-load-misses` A/B, pinned
10. T23 verdict → baseline notes + ADR-0022 addendum
11. Pinned benchstat `Boundary8/9` → `docs/benchmarks/`
12. Justification comment on `linearScanMax` citing the measured 8-vs-9 numbers
13. BDD/PTY test for HTML auto-write (real `script(1)` or pty lib)
14. Decide + document `--quiet` vs `WarnUnmatchedExcludePatterns` (currently unconditional)
15. Symmetric zero-match warning for `--include-pattern`
16. `scripts/check-coverage.sh`: filter test-only packages (0.0% rows are noise)
17. TESTING.md: property/parity-test conventions section (reference-scan pattern, coverage-guard-in-test pattern)
18. a11y: visible focus style for `.anchor-link`
19. Expose `AnchorID` in JSON output for cross-format linking
20. `--html-out`: mkdir -p parent dir
21. Extend HTML goldens with one clone group (covers anchors/diff markup)
22. Investigate corpus drift 2665→2670 (which 5 clone groups appeared?)
23. Decide budget ratchet: tighten `MemoryUsage` 4802→~4785 or keep headroom
24. Consider wiring the threshold-1 self-scan as a fast unit test, not only flake check
25. Update `docs/status/2026-08-16_13-13...` cross-ref or mark superseded by this report
26. Re-run `benchstat` on committed baselines after T2.2 verdicts; commit refreshed numbers
27. Sweep for other daemon-committed surprises: `git log --since="2026-08-16 13:00" --oneline` review
28. Master-plan closure note in the plan doc (26/26 or explicitly-parked status)

_(28 items — the queue from the 13-13 report's §f items 26–38 is fully represented above; nothing beyond it is invented.)_

## g) Questions I CANNOT answer myself

1. **Timing work now or later?** The machine is idle (load 2.3) — T2.2 + T23 are finally unblocked. Run them as the next block (they take ~30–60 min of mostly-waiting and pair well with the 40-min flake check), or do you want the machine for something else first?
2. **Coverage baseline destiny**: keep `coverage-baseline.txt` advisory forever, or promote per-package floors into `nix flake check` once numbers stabilize (e.g. suffixtree ≥95%, syntax ≥80%)?
3. **T12 warning under `--quiet`**: currently the zero-match `--exclude-pattern` warning prints unconditionally. Suppress under `--quiet` (consistent with all other progress output) or keep unconditional (it's a misconfig signal, arguably worth shouting)?
