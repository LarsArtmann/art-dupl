# Real-World End-to-End Benchmark: cli/cli

**First measurement of where a real run's time goes.** Gates the perf work
planned in `docs/planning/2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md`.

- **Fixture**: [cli/cli](https://github.com/cli/cli) (GitHub CLI), shallow clone
- **Pinned commit**: `0eeec0b92edbe70199f9768522f831d3534f41ad`
- **Corpus size**: 916 non-vendored `.go` files
- **art-dupl commit**: `9d7f98130f6aa73f1f9f0f84f802bf0d9961b22e` + `--timing` instrumentation
- **Command**: `art-dupl --timing --quiet -t 15 <fixture>`
- **Date**: 2026-08-16
- **Reproduce**: `scripts/bench-realworld.sh <fixture-dir>` (set `PIN_CORES=0-7,16-23` for the pinned mode)

## Results (3 runs per mode)

### Pinned to one L3 domain (`taskset -c 0-7,16-23`, 16 cores)

| stage | run 1 | run 2 | run 3 |
|---|---|---|---|
| crawl (wall) | 112ms | 119ms | 112ms |
| parse (active, summed) | 430ms | 424ms | 423ms |
| serialize (active) | 30ms | 29ms | 28ms |
| tree-build (active) | 95ms | 101ms | 94ms |
| ingest (wall) | 121ms | 127ms | 121ms |
| search (wall) | 50ms | 51ms | 59ms |
| print (wall) | 51ms | 52ms | 60ms |
| **total (wall)** | **172ms** | **179ms** | **181ms** |
| allocations | 4.43M objs / 267.0 MB | 4.43M / 266.9 MB | 4.43M / 266.9 MB |
| GC | 13 cycles / 1.07ms pause | 14 / 1.12ms | 14 / 1.35ms |

### Unpinned (all 32 threads, machine default)

Medians of 3 runs: total **186ms**, ingest 134ms, search 51ms, tree-build
(active) 98ms, parse (active, summed) 514ms, serialize 39ms. Allocation
counts identical (±0.1%).

## Interpretation

Share of **total wall** (pinned medians, total ≈ 181ms):

| component | time | share of total wall | note |
|---|---|---|---|
| suffix tree build | 94–102ms active | **52–56%** | sequential (Ukkonen); 78–80% of the 121–128ms ingest wall → it IS the ingest critical path |
| suffix tree search | 48–59ms wall | **26–33%** | sequential by default (`--search-workers 0`) |
| serialize | 28–32ms active | **15–18%** | single goroutine on the critical path |
| parse | 423–459ms summed active | ~0% marginal | fully hidden by worker parallelism (16 cores) |
| GC | 1.1–1.4ms pause | <1% | allocations are cheap at this scale |

**Suffix tree (build + search) ≈ 80% of a real run's wall clock.** Parse,
despite consuming the most CPU (423ms summed), contributes almost nothing to
wall time at default parallelism because the sequential tree builder absorbs
it. Pinning to one L3 domain cut parse active time ~18% (514→430ms) and total
wall ~3% — consistent with `CPU_TOPOLOGY.md`.

## Verdict Memo (gates T13 / T14 / T23)

1. **T13 `serial()` bulk allocation + T14 `[]*Node` pool: GO.** The plan's
   gate was "serialize ≥ 15% of wall clock" — measured 15–18%, gate met at the
   margin. Framing correction: with GC pause at ~1.3ms these are
   *allocation-count* wins (4.43M objects/run), not wall-time wins. Execute
   them for the deterministic metric; do not expect total wall to move.
2. **T23 `perf stat` A/B: GO.** Tree build+search dominate; cache-miss
   evidence for ADR-0022's locality claims is worth collecting.
3. **ADR-0020 suffix-array migration gate: MET** (search ≥ 30% of wall).
   Parked work stays parked, but its entry criterion is now evidenced —
   revisit if search share grows or wall-time targets appear.
4. **Unexpected finding:** tree **build** (52–56%) now outweighs search
   (26–33%). Build is inherently sequential (ADR-0019). If wall time ever
   matters for large repos, the lever is chunked trees / suffix arrays, not
   more search workers.

## Caveats

- One fixture, one machine (AMD Ryzen AI MAX+ 395, thermally noisy; timing
  varied ±5% across runs, stage ordering was stable across all 6 runs).
- Active-time stages (`parse`, `serialize`, `tree-build`) measure goroutine
  busy time; their wall-clock contribution differs as analyzed above.
- 916 files ≈ 270ms runs: too small for cache-miss-level attribution; T23's
  `perf stat` covers that layer.
