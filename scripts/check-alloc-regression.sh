#!/usr/bin/env bash
# Allocation-regression gate: fails when any guarded benchmark's allocs/op
# exceeds its committed budget by more than the +1 tolerance.
#
# Allocation counts are the near-deterministic perf metric on this project
# (CPU timings are thermally noisy); the gate turns "we cut allocations X%"
# into an enforced invariant. Counts for the suffixtree benchmarks fluctuate
# within a small noise band (measured 2026-09-19: seq/tokens_10000 drew
# 30742-30745 across runs on the SAME machine and toolchain — goroutine- and
# runtime-internal allocation noise, NOT machine variance; GOMAXPROCS=1 does
# not stabilize it). The gate therefore runs each benchmark suite with
# -count=3 and compares the MINIMUM observed allocs/op against the budget:
# the min converges to the deterministic floor, so budgets are stable and a
# real regression still trips the +1 headroom. Budget values in the file are
# the FLOOR, not the ceiling — a single-draw capture above the floor shows up
# as "improved" info lines, never as a failure.
#
# Usage:
#   scripts/check-alloc-regression.sh            # use default budgets
#   scripts/check-alloc-regression.sh other.txt  # custom budget file
#
# Budget file format (tab- or space-separated, # comments):
#   BenchmarkName/Sub  maxAllocsPerOp
#
# The benchmark name matches WITHOUT the trailing -<procs> suffix.
# When you deliberately change allocations, update the budget file in the
# same commit and note why.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
budget_file="${1:-$repo_root/scripts/alloc-budgets.txt}"

if [[ ! -f "$budget_file" ]]; then
	echo "error: budget file $budget_file not found" >&2
	exit 2
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

bench_out="$tmp/bench.txt"

(cd "$repo_root" && GOEXPERIMENT=jsonv2 go test ./suffixtree/ -run '^$' \
	-bench '^(BenchmarkSTreeUpdate|BenchmarkFindDuplOver|BenchmarkMemoryUsage)' \
	-benchmem -count=3 | tee "$bench_out") >&2

(cd "$repo_root" && GOEXPERIMENT=jsonv2 go test ./syntax/ -run '^$' \
	-bench '^BenchmarkSerialize' \
	-benchmem -count=3 | tee -a "$bench_out") >&2

fail=0
improved=0

while IFS=$'\t' read -r bench budget _note; do
	[[ -z "$bench" || "$bench" == \#* ]] && continue

	# Strip optional -procs suffix from budget entries.
	bench="${bench%%-*}"

	# Find ALL allocs/op draws for this benchmark across the -count=3 runs
	# and keep the minimum: the min converges to the deterministic floor,
	# filtering the run-to-run allocation noise documented in the header.
	actual="$(
		awk -v b="$bench" \
			'{ name = $1; sub(/-[0-9]+$/, "", name); if (name == b) {for (i = 2; i <= NF; i++) if ($(i+1) == "allocs/op") {print $i}}}' \
			"$bench_out" | sort -n | head -1
	)"

	if [[ -z "$actual" ]]; then
		# Benchmarks with 0 allocs omit the column entirely.
		actual=0
	fi

	limit=$((budget + 1))

	if ((actual > limit)); then
		echo "FAIL: $bench allocs/op $actual exceeds budget $budget (tolerance +1)" >&2
		fail=1
	elif ((actual < budget)); then
		echo "info: $bench improved: $actual < budget $budget — ratchet the budget down" >&2
		improved=1
	else
		echo "ok:   $bench $actual <= $limit" >&2
	fi
done < <(sed 's/[[:space:]]\+/	/g; s/^\t//' "$budget_file")

if ((fail)); then
	exit 1
fi

if ((improved)); then
	echo "note: at least one benchmark beat its budget; consider tightening scripts/alloc-budgets.txt" >&2
fi

echo "allocation gate passed"
