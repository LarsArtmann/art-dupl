# SARIF output-path allocation baseline (2026-10-11)

Question (harvested as G12 after the M5-M8 wire corrections): do the
per-group SARIF additions (`relatedLocations` fan-out via
`sarifSiblingLocation`, conditional `category` property) add allocation
churn on a hot path?

Verdict: **the SARIF path is output-layer only** — it runs once per clone
group at print time, after detection; the detection hot path is guarded
separately by the suffixtree/syntax alloc budgets
(`scripts/alloc-budgets.txt`, flake `alloc-gate`). This benchmark pins the
output-path cost so future wire changes have a benchstat baseline.

Benchmark: `BenchmarkSARIFPrintClonesMulti` (`printer/sarif_bench_test.go`)

- one clone group, 3 occurrences in distinct files, fragments present
  (snippet + sibling-link paths hot), full PrintHeader/PrintClones/
  PrintFooter cycle including the final JSON marshal of the whole document.

Machine: AMD RYZEN AI MAX+ 395, go1.27.1, -count=5, 2026-10-11:

```
BenchmarkSARIFPrintClonesMulti-32   ~2121-3088 iterations
   ns/op:  399,718 - 700,266   (timing noisy; thermally shared machine)
   B/op:   219,481 - 226,418
   allocs/op: 252 - 253        (stable)
```

Allocs/op is the deterministic signal (per repo convention): ~253
allocations per full-document SARIF emit at 3 occurrences. No regression
gate is attached — if a future wire change moves this number materially,
re-run with `-count=5` and judge; do not add the number to
`alloc-budgets.txt` (output-layer cost is proportional to result count and
not perf-critical).
