# Status Report: ADR-0022 Timing Evidence — T2.2 Pinning A/B, T23 perf stat, linearScanMax Boundary

**Date:** 2026-09-22 22:40 · **Branch:** fork (HEAD `332aef7b`, working tree clean) ·
**Scope:** this session only (the HIGH-priority "Timing evidence for ADR-0022" block from TODO_LIST)

All three parked tasks were executed, measured, documented, and verified. The
headline: **the 2026-08-16 "pin to one CCX, win 25-30%" claim is obsolete on the
slice layout — the direction reversed** (unpinned par32 is 3.9% faster than
pinned par16, p=0.011), and **T23's hardware counters confirm ADR-0022's cache
claim** (cache-misses −59.2%, references −67.8%, <3% round variance).

---

## a) FULLY DONE

1. **T2.2 — interleaved pinned/unpinned A/B** (spec'd benches): 6 alternations/arm,
   `-count=5` (30 samples/arm), compiled test binary, `GOEXPERIMENT=jsonv2`,
   GOMAXPROCS-suffix normalization for benchstat. Results: threshold_10
   117.5µs±19% vs 114.2µs±4% (~, p=0.146; pinned CIs ~5× tighter); par4/10k
   987µs±7% vs 921µs±11% (−6.7%, p=0.080, not significant).
2. **T2.2 supplement — NumCPU-matched A/B** (the actual configuration behind the
   25-30% claim: worker count follows the affinity mask): par32 full machine
   463.1µs±2% vs par16 one-CCX 481.3µs±4% — **unpinned wins +3.9% (p=0.011)**.
   Bench names (par16/par32) normalized to `parNumCPU` for benchstat grouping.
3. **T2.2 interpretation**: CPU_TOPOLOGY.md's claim was measured on the map
   layout (its own text: "map traversal over a read-only tree") —
   L3-latency-bound pointer chasing. ADR-0022's contiguous layout made search
   scale across workers (par32: ~2.2ms → 463µs, ~4.7×), obsoleting pinning for
   speed. Pinning remains the right protocol for regression benchmarks (CIs).
4. **T23 — perf stat cache-counter A/B vs `23fa1b4f`** (worktree built at
   `/tmp/artdupl-23fa1b4f`, compared, removed): both arms pinned to CCX0,
   interleaved ×4, 3 benches × count=5 per invocation. cache-references
   35.81e9 → 11.55e9 (−67.8%); cache-misses 6.07e9 → 2.48e9 (−59.2%); miss
   RATE rose 16.9%→21.4% (fewer lines touched overall; survivors are output
   allocs + GC). Same-window timing medians: construction −41%, sequential
   search −52%, par4/10k −63%.
5. **T23 platform forensics**: `LLC-load-misses` has no openable counterpart
   (no `amd_nb`/`l3` PMU registered; perf-listed `l3_cache/*` events fail to
   open). Documented `cache-misses` (raw 0x964) as the LLC-bound stand-in.
6. **Boundary — `linearScanMax` cutoff evidence**: pinned benchstat for
   Boundary8/9 (spec) **plus** new Boundary4/16 fixtures (scope extension,
   justified: brackets the crossover) — per-probe 2.87ns (n=4) / 4.23 (n=8) /
   3.46 (n=9) / 3.85 (n=16). Crossover sits between 4 and 8; cutoff 8 keeps
   the early-exit scan for the 2-5-transition majority at ≤~18%/probe cost on
   6-8-fanout states. Numbers recorded on the constant's doc comment
   (`suffixtree/findtran.go`).
7. **Provenance correction (discovered)**: the committed v2 baseline file's
   1,543 search allocs came from an UNCOMMITTED mid-overhaul state; the actual
   last map-layout commit (`23fa1b4f`) measures 7,158. Documented in notes +
   ADR — the A/B measures the full `081e347f` overhaul (pool+slices+arena).
8. **Evidence committed**: `docs/benchmarks/pinned-unpinned-2026-09-22.txt`
   (358 lines, per-round ambient load embedded), `linear-scan-boundary-2026-09-22.txt`.
9. **Docs updated everywhere the stale claim lived**: v3 notes (T2.2+T23+boundary
   sections; "still open" placeholder replaced), ADR-0022 addendum, CPU_TOPOLOGY.md
   (supersession banner, revised Do/Don't, dropped `--cpu-affinity` follow-up),
   benchmarks README (rules of thumb + 2 new file rows), AGENTS.md affinity
   paragraph. Sweep found no residue in HOW_TO_USE.md / README.md / website / ADR-0019.
10. **Living docs synced**: TODO_LIST (HIGH section closed, dated 2026-09-22),
    CHANGELOG entry (also removed a stale duplicate `### Fixed/Nothing yet` block).
11. **Verification**: `go build ./...`, `go test ./suffixtree/`,
    `CGO_ENABLED=1 go test -race -run TestBoundary`, new benches compile+run
    (0 allocs), `golangci-lint ./suffixtree/` clean; alloc-gate regex confirmed
    to exclude the new fixtures (no budget file change needed). Worktree pruned.

## b) PARTIALLY DONE

1. **T2.2 evidence quality**: complete and internally consistent, but measured
   under ambient load 8-19 (see d1). The reversal direction is robust (residual
   bias runs AGAINST the unpinned arm, which won anyway), but the exact +3.9%
   magnitude carries a load asterisk. No idle-window confirmation run exists.
2. **T23 counter attribution**: process-total counters (setup + calibration
   included), not per-bench/per-op. Both arms used an identical protocol so the
   comparison is valid, but finer attribution was achievable (one bench per
   process, or counters normalized per completed op).
3. **Reproducibility tooling**: full protocol is documented in prose, but the
   actual runner scripts (`/tmp/artdupl-timing/*.sh`) were NOT committed — they
   are volatile. Evidence files are committed; the executable protocol is not.

## c) NOT STARTED

1. Idle-window re-validation of the pinning reversal (a load<4 window never
   opened during 12+ min of polling; one DID open at 22:26 — see d1).
2. Full `nix flake check` (canonical gate: full suite + race + alloc-gate);
   I ran proportionate targeted checks instead (comments+benches+docs change).
3. Committing the A/B runner scripts under `scripts/`.
4. Fresh full-benchmark baseline generation (the new evidence is targeted, not
   a new committed baseline file for the whole suite).

## d) TOTALLY FUCKED UP (honest)

1. **I noticed a quiet window and didn't use it.** The boundary bench ran at
   load ~3.4 (22:26) — the entry criterion actually opened ~10 minutes after I
   gave up waiting and ran T2.2 under load 8-19. Correct move at 22:26: re-run
   the headline parNumCPU A/B in that window for a clean confirmation. I ran
   the (already-scheduled) boundary bench, noticed the load in the output, and
   moved on to documentation instead of capitalizing. The reversal claim now
   rests on load-disclosed interleaved data when a cleaner confirmation was
   sitting right there.
2. **Wrong provenance shipped in a commit.** The evidence-file header initially
   claimed "fork@dd7d211c (clean tree)" — wrong: the working tree was at
   `3f839885` by build time (daemon commits). Root cause: I trusted the stale
   session-start git snapshot instead of capturing `git rev-parse HEAD` at
   build time. Caught and corrected (with the alloc-column identity argument),
   but the wrong version was auto-committed as `2b300abe` for ~9 minutes.
3. **Notes claimed data the raw file didn't contain.** First assembled evidence
   file lacked the per-round ambient loads ("recorded per round in the raw
   file" was false until I regenerated it 9 minutes later; loads had gone to a
   separate log). Same auto-commit window as above. Fixed by regeneration.
4. **Documented an inference as fact.** "cache-misses maps to AMD L2-miss
   events (select 0x64)" — I verified the raw config (0x964) but the umask
   semantics come from AMD PPR knowledge, not from this Zen 5 machine's event
   table (perf list shows only generic names). Well-founded, but the docs state
   it more confidently than the verification supports.
5. Minor: `go install benchstat` sandbox denial cost a round trip; using
   `go run ...@latest` also means the analysis toolchain isn't pinned (the
   module cache holds whatever @latest resolved to today).

## e) WHAT WE SHOULD IMPROVE (in me / in process)

1. **Capture provenance at build time**: evidence headers should be generated
   with `git rev-parse HEAD` + `git status --porcelain` embedded by the runner
   script, never hand-written from memory or stale snapshots.
2. **Idle-window watcher**: a script that polls `/proc/loadavg`, waits for N
   consecutive samples < threshold, then auto-fires the A/B suite — eliminates
   the entire "parked on machine load" task class (this task sat a month).
3. **When a criterion you're missing suddenly opens mid-session — use it.**
   Generalize: re-check blocked-entry conditions whenever ANY measurement
   reveals the environment changed, not just at scheduled poll points.
4. **Commit runner scripts with evidence**: raw outputs without the exact
   scripts are reproducible only in prose. `scripts/bench/` (or similar) should
   hold t22_ab.sh, t23_perf.sh, boundary_bench.sh.
5. **Per-op counter normalization** for perf evidence (counters ÷ completed
   benchmark iterations, or one bench per process) — stronger than totals.
6. **Pin the analysis toolchain** (benchstat version via a tools go.mod or
   vendored path) so committed evidence is regenerable byte-comparably.
7. **Verify AMD event semantics** against the Zen 5 PPR (or switch to IBS
   events) before citing select 0x64 semantics as fact — or soften the wording
   to "raw config 0x964 (AMD core-PMU cache events)".
8. **Run `nix flake check` before declaring benchmark/test-file changes done**,
   even when the diff looks comment-only — it is the project's canonical gate
   and I know it.

## f) NEXT (up to 50, roughly priority-ordered)

**Close out this session's threads:**
1. Re-run the parNumCPU A/B in a genuine idle window (overnight watcher or
   manual when load<4) to confirm the reversal magnitude without the asterisk.
2. Same for threshold_10 ±CI comparison (pinned-tighter-CIs claim, n=30).
3. Commit A/B runner scripts (`/tmp/artdupl-timing/*.sh`) into `scripts/bench/`.
4. Run full `nix flake check` on the current tree (race + alloc-gate + arch-lint).
5. Verify/soften the AMD event 0x64 umask semantics wording in the notes/ADR.
6. Build the load-waiting wrapper (polls loadavg, fires suite when criterion
   holds) so "parked on machine load" tasks never park again.
7. Decide and record: bench protocol for future committed baselines stays
   pinned-only (stability) even though unpinned is faster — one sentence in
   benchmarks README already says this; confirm it's the standing rule.
8. Cross-check `baseline-2026-09-14-nested-tokens.txt` provenance (pinned or
   not?) — pinned-vs-unpinned baselines must not be benchstat-compared.
9. Consider a fresh full-suite pinned baseline file (2026-09-22) since the
   search got measurably faster since the last committed baseline.
10. Sanity-check that `parN` worker-count naming (`runtime.NumCPU()`) in
    bench output doesn't confuse CI alloc-gate parsing across machines
    (gate regex already excludes it — confirm no CI lane runs `-bench .`).

**ADR-0022 / suffixtree follow-through:**
11. Retune decision: `linearScanMax` 8→4 or 6 (crossover data exists now;
    needs a real-corpus fanout histogram to size the win — see #12).
12. Instrument real-corpus state-fanout distribution (the 2-10 claim comes
    from synthetic 10k-token trees) to settle #11 quantitatively.
13. Per-worker match batching in `walkTrans` (CPU_TOPOLOGY follow-up #2,
    still potentially valid) — now measurable with the pinned-CI protocol.
14. Re-visit "int32 arena indices / []Pos pool" parked item — the new
    par32-scaling data point (463µs, alloc-bound tail) may change its entry
    criterion evaluation.
15. perf stat per-op counters old-vs-new as a proper benchmark (one bench per
    process) to complete T23's attribution refinement.
16. Consider adding the perf-counter A/B to the same scripts/ set with
    `perf stat -r` repetition so it's re-runnable on layout changes.

**Stale-doc hygiene noticed this session:**
17. CHANGELOG had a stale duplicate `### Fixed / Nothing yet` block (fixed this
    session) — sweep other sections for template leftovers.
18. The 2026-08-16 status reports reference "T2.2 still open" — docs-health
    VERIFY pass item (#35 in TODO_LIST) now covers one more resolved claim.
19. master plan wave2 doc lists T2.2/T23 under "Not Started" (historical
    snapshot — annotate, don't rewrite, per docs-health convention).
20. TODO_LIST "Corpus re-baseline post-v0.7.0" (#22) — this session's timing
    data makes that more urgent (search 4.7× faster than AGENTS corpus-era
    assumptions in the topology section).

**Existing TODO_LIST MEDIUM items this session's work unblocks or informs:**
21. Fleet audit `encoding/json/v2` on Go 1.27 (existing #14 — unaffected, keep).
22. `scripts/pre-release-check.sh` (existing #16).
23. Branch protection + failure notifications (existing #18).
24. Windows exe-start `ProcessState nil` root-cause (existing #4).
25. `--dump-tokens` positions on templ files (existing #48).
26. arch-lint in local `nix flake check` (existing #40) — this session relied
    on manual lint; would have been covered by it.
27. Stale-shell graceful failure (existing #24).
28. Self-clean decision ledger (existing #12).
29. Small quality batch (existing #29-#39 grab-bag).
30. gogenfilter consumer sweep to v3.6.1 for the 9 red-baseline repos (existing).

**Methodology hardening (general):**
31. Add a "benchmark evidence checklist" to TESTING.md: interleave arms,
    record ambient load per round, embed provenance at build time, normalize
    GOMAXPROCS suffixes, disclose entry-criterion violations.
32. Teach the BDD/testutil layer nothing — but consider a tiny
    `scripts/bench/normalize.sh` for the -N suffix stripping (reused 4× today).
33. Poll `git log` before writing provenance by hand in any future session.
34. Consider `nice -n 19` on benchmark binaries vs foreign load (untested;
    may reduce descheduling interference without touching other sessions).
35. Consider `perf stat` paranoid=2 workarounds documentation (kernel-space
    excluded — we noted `:u`; write it into the perf script header).
36. Boundary fixtures: consider a fixture at n=5/6 to narrow the crossover
    bracket if #11 ever gets serious.
37. If linearScanMax is retuned, the layout/alloc tests pin nothing about it —
    add the fanout histogram as a test fixture instead of prose.
38. Reconcile ADR-0022's Results table "FindDuplOver 1,543 (ADR-0019 era)"
    row with the now-measured 7,158 at the actual map-era commit (one-line
    addendum footnote — the addendum already covers it; make the table honest).
39. Add benchstat to the devShell (flake) so `go install` sandbox issues
    don't recur and the version is pinned by the flake.
40. Consider committing the /tmp binaries' build recipe (a `just`-less
    `scripts/bench/build.sh`) so evidence binaries are one command away.

**Website / external (lower priority):**
41. Website perf page (if any) — check whether the 25-30% pinning claim leaked
    into published content (grep found none in website/src today — confirmed
    clean; only re-check if content changes).
42. Nothing else external this session touched.

**Optional depth (only if evidence is questioned):**
43. IBS-based miss attribution (upstream cache level per-instruction) if the
    cache-counter story ever needs instruction-level proof.
44. Repeat T23 with `-benchtime` fixed-op-count mode so old/new run IDENTICAL
    iteration counts (removes the per-op normalization question entirely).
45. Measure CCX1 (8-15,24-31) as the pinned arm to rule out asymmetric
    foreign-load placement favoring CCX0 (cheap, closes the last confound).
46. Two independent observers (fresh binaries, different benchtime seeds) for
    the reversal before quoting +3.9% anywhere external.
47. Document why `par2` was excluded from the A/B (spec listed two benches;
    the supplement added parN — par2/seq remain unmeasured today).
48. If the reversal survives idle-window confirmation, update
    docs/benchmarks/README.md "How to Compare" example (still shows unpinned
    compare against committed baselines — should be pinned compare).
49. Consider a tiny `benchstat` wrapper that fails CI when pinned-vs-unpinned
    files are compared (guard the new standing rule mechanically).
50. Celebrate: the month-parked HIGH block is closed — don't let the next one
    park; the watcher (#6) is the systemic fix.

## g) QUESTIONS (cannot figure out myself; max 3)

1. **Is load-disclosed evidence sufficient to close T2.2 permanently, or do
   you want an idle-window confirmation run** (e.g., the overnight watcher
   fires parNumCPU A/B at load<4) before the reversal is treated as final in
   external-facing material? I'm confident in the direction; the +3.9%
   magnitude is what carries the asterisk.
2. **Retune `linearScanMax` (8 → 4/6)?** The measured crossover argues 8 is
   mildly suboptimal for 6-8-fanout states; the win is bounded (~18%/probe on
   a small state class) and the current constant is now documented and
   justified. Changing it is a behavior change with no production complaint —
   your call whether the optimization is worth the churn.
3. **Standing bench protocol**: keep all committed baselines and CI gates
   pinned-only (my recommendation: yes — tight CIs, comparability with
   history), and if so, should the A/B runner scripts + idle-watcher be
   committed as permanent tooling under `scripts/bench/`?

---

**Evidence index:** `docs/benchmarks/pinned-unpinned-2026-09-22.txt` ·
`docs/benchmarks/linear-scan-boundary-2026-09-22.txt` ·
`docs/benchmarks/baseline-2026-08-16-v3_notes.md` (3 new sections) ·
`docs/adr/0022-suffixtree-data-layout.md` (addendum) ·
`docs/benchmarks/CPU_TOPOLOGY.md` (supersession) · runner scripts (volatile):
`/tmp/artdupl-timing/{t22_ab,t22n_ab,t23_perf,boundary_bench}.sh`
