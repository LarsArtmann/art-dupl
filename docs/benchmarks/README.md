# Benchmark Baselines

This directory contains committed benchmark result snapshots for regression detection.

## Files

| File | Date | CPU | Notes |
|------|------|-----|-------|
| `baseline-2026-08-10.txt` | 2026-08-10 | AMD Ryzen AI MAX+ 395 (32 threads) | Re-run with count=5 for statistical significance. First run was thermally throttled (2x slower). |
| `baseline-2026-08-16.txt` | 2026-08-16 | AMD Ryzen AI MAX+ 395 (32 threads) | count=10 for suffixtree and syntax packages. Includes cache line optimizations (Node/CloneNode field reordering + stack buffer in walkTrans/getAll). Allocation reduction: FindDuplOver -0.29% allocs/op (~21 fewer per call, ~400 fewer at 10k tokens). Timing comparison with baseline-2026-08-10 is inconclusive due to thermal throttling (baseline had only 5 samples, ±∞ CI). |
| `baseline-2026-08-16-v2.txt` | 2026-08-16 | AMD Ryzen AI MAX+ 395 (32 threads) | count=10, suffixtree only. Major allocation reduction: contextList sync.Pool + posList elimination + state arena + tree back-pointer removal. **Search path**: FindDuplOver -78.5% allocs/op (7158→1543), par4/tokens_10000 -78.5% allocs (142761→30798), -97.6% bytes (7.4MB→175KB). **Construction**: STreeUpdate -30% allocs (arena eliminates per-state heap allocs). Timing still thermally noisy but consistently faster (par4/10k: 2.5ms→1.3ms). |
| `baseline-2026-08-16-v3.txt` | 2026-08-16 | AMD Ryzen AI MAX+ 395 (32 threads) | count=10, full suffixtree suite. Slice-based transitions (sorted `[]tran` values replacing `map[TokenValue]*tran`) + lazy allocation + 512-state arena blocks (ADR-0022). **Construction**: STreeUpdate/tokens_2000 -67% allocs (6214→2046), -39% bytes; tokens_100 -70% bytes. MemoryUsageManyTokens 10067→44 allocs. Search allocs unchanged (1543 — inherent `[]Pos`/Match output); map CPU share fell ~32%→~14%. See `baseline-2026-08-16-v3_notes.md`. |
| `CPU_TOPOLOGY.md` | 2026-08-16 | AMD Ryzen AI MAX+ 395 | CPU topology analysis + CCX-pinning A/B benchmarks: one-CCX `taskset` beats full machine by ~25-30% on the parallel search. Explains the 1-NUMA-node/2-die design and how to run affinity-aware benchmarks. |

## How to Compare Against Baseline

```bash
# Run benchmarks and compare
go test ./suffixtree/... ./syntax/... ./syntax/golang/... -bench=. -benchmem -count=3 -run='^$' > /tmp/bench_current.txt
benchstat baseline-2026-08-10.txt /tmp/bench_current.txt
```

## Generating a New Baseline

When making significant performance changes, generate a new baseline:

```bash
GOEXPERIMENT=jsonv2 go test \
  ./suffixtree/... ./syntax/... ./syntax/golang/... \
  ./cache/... ./printer/actionability/... ./cmd/... ./pkg/format/... ./printer/... \
  -bench=. -benchmem -count=5 -run='^$' \
  > docs/benchmarks/baseline-$(date +%Y-%m-%d).txt
```
