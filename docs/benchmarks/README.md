# Benchmark Baselines

This directory contains committed benchmark result snapshots for regression detection.

## Files

| File                                  | Date       | CPU                                | Notes                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ------------------------------------- | ---------- | ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `baseline-2026-08-10.txt`             | 2026-08-10 | AMD Ryzen AI MAX+ 395 (32 threads) | Re-run with count=5 for statistical significance. First run was thermally throttled (2x slower).                                                                                                                                                                                                                                                                                                                                                            |
| `baseline-2026-08-16.txt`             | 2026-08-16 | AMD Ryzen AI MAX+ 395 (32 threads) | count=10 for suffixtree and syntax packages. Includes cache line optimizations (Node/CloneNode field reordering + stack buffer in walkTrans/getAll). Allocation reduction: FindDuplOver -0.29% allocs/op (~21 fewer per call, ~400 fewer at 10k tokens). Timing comparison with baseline-2026-08-10 is inconclusive due to thermal throttling (baseline had only 5 samples, ±∞ CI).                                                                         |
| `baseline-2026-08-16-v2.txt`          | 2026-08-16 | AMD Ryzen AI MAX+ 395 (32 threads) | count=10, suffixtree only. Major allocation reduction: contextList sync.Pool + posList elimination + state arena + tree back-pointer removal. **Search path**: FindDuplOver -78.5% allocs/op (7158→1543), par4/tokens_10000 -78.5% allocs (142761→30798), -97.6% bytes (7.4MB→175KB). **Construction**: STreeUpdate -30% allocs (arena eliminates per-state heap allocs). Timing still thermally noisy but consistently faster (par4/10k: 2.5ms→1.3ms).     |
| `baseline-2026-08-16-v3.txt`          | 2026-08-16 | AMD Ryzen AI MAX+ 395 (32 threads) | count=10, full suffixtree suite. Slice-based transitions (sorted `[]tran` values replacing `map[TokenValue]*tran`) + lazy allocation + 512-state arena blocks (ADR-0022). **Construction**: STreeUpdate/tokens_2000 -67% allocs (6214→2046), -39% bytes; tokens_100 -70% bytes. MemoryUsageManyTokens 10067→44 allocs. Search allocs unchanged (1543 — inherent `[]Pos`/Match output); map CPU share fell ~32%→~14%. See `baseline-2026-08-16-v3_notes.md`. |
| `coverage-baseline.txt`               | 2026-08-16 | —                                  | Per-package `go test -cover` snapshot (total 75.2%). Regenerate with `scripts/check-coverage.sh` (or `--print` for stdout). A trend signal, not a gate: investigate drops, don't chase increases.                                                                                                                                                                                                                                                           |
| `CPU_TOPOLOGY.md`                     | 2026-08-16 | AMD Ryzen AI MAX+ 395              | CPU topology analysis + CCX-pinning A/B benchmarks (map-layout era: one-CCX `taskset` beat full machine by ~25-30%; **superseded 2026-09-22** — on the slice layout unpinned par32 is ~4% faster, see `baseline-2026-08-16-v3_notes.md`). Explains the 1-NUMA-node/2-die design and how to run affinity-aware benchmarks.                                                                                                                                   |
| `pinned-unpinned-2026-09-22.txt`      | 2026-09-22 | AMD Ryzen AI MAX+ 395              | T2.2 interleaved pinned/unpinned A/B (30 samples/arm) on the ADR-0022 slice layout: pinning no longer wins (par32 full machine +3.9% vs pinned par16, p=0.011); pinned runs have ~5x tighter CIs. Ambient load 8-19 disclosed per round.                                                                                                                                                                                                                    |
| `linear-scan-boundary-2026-09-22.txt` | 2026-09-22 | AMD Ryzen AI MAX+ 395              | Pinned `FindTranBoundary` 4/8/9/16 (20 samples/size): per-probe lookup cost brackets the linear/binary crossover between 4 and 8, justifying `linearScanMax = 8`. Recorded on the constant's doc comment in `suffixtree/findtran.go`.                                                                                                                                                                                                                       |

## How to Compare Against Baseline

```bash
# Run benchmarks and compare
go test ./suffixtree/... ./syntax/... ./syntax/golang/... -bench=. -benchmem -count=3 -run='^$' > /tmp/bench_current.txt
benchstat baseline-2026-08-10.txt /tmp/bench_current.txt
```

## Timing Benchmarks: Pin the Cores (Required)

The dev machine (AMD Ryzen AI MAX+ 395) is thermally noisy and has 2 L3 domains.
For any benchmark whose **timing** columns matter, pin to one CCX and prefer
allocation columns when the run is unpinned:

```bash
# One L3 domain (CCX0): CPUs 0-7 + 16-23
taskset -c 0-7,16-23 go test ./suffixtree/ -run '^$' -bench . -benchmem -count=10 > /tmp/bench_pinned.txt
```

Rules of thumb (measured 2026-08-16, revised 2026-09-22, see `CPU_TOPOLOGY.md`):

- **Parallel search benches**: pin for CI tightness, not speed — on the
  current (slice) layout the full machine is marginally FASTER (par32 vs
  pinned par16, +3.9%, p=0.011); the old 25-30% pinning win was a property
  of the superseded map layout. Pin when comparing against committed
  baselines (they were captured pinned).
- **Sequential benches**: pinning removes cross-CCX scheduler migration noise;
  single-core benches are valid either way, but pinned runs have tighter CIs
  (±4% vs ±19% measured on threshold_10).
- **Allocs/op columns are deterministic** regardless of pinning — when in doubt,
  compare allocations, not nanoseconds.
- Compare only pinned-vs-pinned or unpinned-vs-unpinned runs; `benchstat`
  cannot normalize affinity. Interleave A/B rounds (never compare runs taken
  hours apart) and normalize the `-GOMAXPROCS` suffix before benchstat.

## Generating a New Baseline

When making significant performance changes, generate a new baseline:

```bash
GOEXPERIMENT=jsonv2 go test \
  ./suffixtree/... ./syntax/... ./syntax/golang/... \
  ./cache/... ./printer/actionability/... ./cmd/... ./pkg/format/... ./printer/... \
  -bench=. -benchmem -count=5 -run='^$' \
  > docs/benchmarks/baseline-$(date +%Y-%m-%d).txt
```
