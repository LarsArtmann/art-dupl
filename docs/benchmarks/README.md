# Benchmark Baselines

This directory contains committed benchmark result snapshots for regression detection.

## Files

| File | Date | CPU | Notes |
|------|------|-----|-------|
| `baseline-2026-08-10.txt` | 2026-08-10 | AMD Ryzen AI MAX+ 395 (32 threads) | Re-run with count=5 for statistical significance. First run was thermally throttled (2x slower). |
| `baseline-2026-08-16.txt` | 2026-08-16 | AMD Ryzen AI MAX+ 395 (32 threads) | count=10 for suffixtree and syntax packages. Includes cache line optimizations (Node/CloneNode field reordering + stack buffer in walkTrans/getAll). Allocation reduction: FindDuplOver -0.29% allocs/op (~21 fewer per call, ~400 fewer at 10k tokens). Timing comparison with baseline-2026-08-10 is inconclusive due to thermal throttling (baseline had only 5 samples, ±∞ CI). |

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
