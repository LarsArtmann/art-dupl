# Benchmark Baselines

This directory contains committed benchmark result snapshots for regression detection.

## Files

| File | Date | CPU | Notes |
|------|------|-----|-------|
| `baseline-2026-08-10.txt` | 2026-08-10 | AMD Ryzen AI MAX+ 395 (32 threads) | First baseline after fixing crashing benchmarks (FindTranSmall, TestAndSplit) and cache key isolation bug |

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
  -bench=. -benchmem -count=1 -run='^$' \
  > docs/benchmarks/baseline-$(date +%Y-%m-%d).txt
```
