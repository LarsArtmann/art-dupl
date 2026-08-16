# CPU Topology & Affinity Findings (AMD Ryzen AI MAX+ 395)

Investigation date: 2026-08-16. Answers two questions:

1. Can we exploit CCX/L3-domain locality for the parallel suffix-tree search?
2. Why does the OS report a single CPU/NUMA node despite two compute dies?

All facts below were verified on this machine via sysfs and A/B benchmarks
(commands to reproduce at the bottom).

## Topology Facts (verified via sysfs)

| Topology level  | Count   | Details                                                     |
| --------------- | ------- | ----------------------------------------------------------- |
| Socket/package  | 1       | `physical_package_id=0` for all CPUs                        |
| Die (CCD)       | 2       | die 0 = CPUs `0-7,16-23`, die 1 = CPUs `8-15,24-31`         |
| L3 (LLC) domain | 2       | 32 MB each; `shared_cpu_list` matches the die split exactly |
| NUMA node       | 1       | `node0` covers CPUs `0-31`; LPDDR5X, UMA                    |
| Cores / threads | 16 / 32 | 2 threads per core (SMT)                                    |

The kernel is not blind to the two dies: `die_cpus_list` and the L3
`shared_cpu_list` both expose the split, and the scheduler's LLC (MC)
sched domains follow it. The single-CPU view is only true at the package
and NUMA level — and that is correct (see next section).

## Why a Single NUMA Node Is Correct Here

NUMA describes **memory** asymmetry, not core asymmetry. On Strix Halo the
LPDDR5X memory controllers sit on the shared SoC die, not on either CCD, so
both compute dies are equidistant from all DRAM. There is no "local vs
remote memory" distinction, so firmware publishes an SRAT with a single
proximity domain and Linux exposes one node.

- Contrast: EPYC in NPS4 mode attaches memory controllers per fabric
  quadrant → real NUMA nodes appear. Even 1P EPYC defaults to NPS1
  (single node) despite 4–12 CCDs.
- The asymmetry that DOES exist (cross-die L3 traffic pays an Infinity
  Fabric hop, ~10–20 ns) is cache-shaped, and the kernel models it at the
  die/LLC sched-domain level, not the NUMA level.
- Faking NUMA nodes would be worse than none: the NUMA balancer would
  migrate pages between domains chasing a ~10–20 ns delta against ~100 ns
  DRAM latency — churn with no payoff.

## Benchmark Evidence: CCX Pinning Wins

`BenchmarkFindDuplOverParallel`, `tokens_10000`, median ns/op of 5 runs
(`GOEXPERIMENT=jsonv2`, two independent run orderings to control for
thermal effects):

| Config                           | HW threads | parN | Median           |
| -------------------------------- | ---------- | ---- | ---------------- |
| `taskset -c 0-7,16-23` (one CCX) | 16 (SMT)   | 16   | **1.71–1.72 ms** |
| `taskset -c 0-7` (8 physical)    | 8          | 8    | ~1.83 ms         |
| Unpinned (both CCXes)            | 32         | 32   | ~2.17–2.31 ms    |

Findings:

- **One CCX with half the threads beats the full machine by ~25–30%.**
  Stable across both run orderings.
- SMT adds only ~6–8% (16 logical vs 8 physical on the same L3).
- Why: the search is L3-latency bound (map traversal over a read-only
  tree). The tree is read-shared, so with both CCXes active each die holds
  a duplicate copy in its L3 and misses pay the fabric hop. Constraining
  to one CCX keeps the working set in a single 32 MB L3.

Caveats:

- This machine thermally throttles under sustained benchmark load
  (documented in `README.md`; one unpinned run degraded 9.2 → 19.3 ms over
  three consecutive iterations of `seq/tokens_10000`). Medians and
  reversed-order reruns are used to control for this.
- Bench trees are small (10k tokens, fit in L3). A real-corpus A/B should
  confirm before hardcoding any default.

## Practical Guidance

**Do:**

```bash
# Pin to one CCX (fastest on this machine for the search phase)
taskset -c 0-7,16-23 art-dupl ...

# GOMAXPROCS follows the affinity mask automatically
```

**Don't:**

- `numactl --membind` / `--cpunodebind` — no-op or meaningless on this
  UMA machine.
- In-process pinned thread pools (`LockOSThread` + `sched_setaffinity`) —
  they fight the Go scheduler and lose work-stealing; the tree walk is
  fine as plain goroutines under an external affinity mask.
- Expect Go to do this itself: goroutine→core affinity does not exist, and
  the Go allocator has no locality concept.

**Proposed follow-ups (not yet implemented):**

1. `--cpu-affinity <cpuspec>` CLI flag — `sched_setaffinity` at startup;
   `GOMAXPROCS` adapts automatically.
2. Per-worker match batching in `walkTrans` (`suffixtree/dupl.go`) to cut
   cross-core channel traffic — the only shared write in the hot loop.

## Reproducing

```bash
# Topology inspection
lscpu
cat /sys/devices/system/cpu/cpu{0,8}/topology/{die_id,die_cpus_list,physical_package_id}
cat /sys/devices/system/cpu/cpu0/cache/index3/{size,shared_cpu_list}

# A/B benchmark (note: /tmp may be a nearly-full tmpfs; redirect TMPDIR)
export GOEXPERIMENT=jsonv2 TMPDIR="$HOME/.cache/artdupl-tmp"
taskset -c 0-7,16-23 go test ./suffixtree -bench=FindDuplOverParallel -run='^$' -count=5
taskset -c 0-7        go test ./suffixtree -bench=FindDuplOverParallel -run='^$' -count=5
                      go test ./suffixtree -bench=FindDuplOverParallel -run='^$' -count=5
```
