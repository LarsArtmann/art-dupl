#!/usr/bin/env bash
# Allocation-regression gate: fails when any guarded benchmark's allocs/op
# exceeds its committed budget by more than the ±1 tolerance.
#
# Allocation counts are the deterministic perf metric on this project (CPU
# timings are thermally noisy); the gate turns "we cut allocations X%" into an
# enforced invariant.
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
	-benchmem -count=1 | tee "$bench_out") >&2

(cd "$repo_root" && GOEXPERIMENT=jsonv2 go test ./syntax/ -run '^$' \
	-bench '^BenchmarkSerialize' \
	-benchmem -count=1 | tee -a "$bench_out") >&2

fail=0
improved=0

while IFS=$'\t' read -r bench budget _note; do
	[[ -z "$bench" || "$bench" == \#* ]] && continue

	# Strip optional -procs suffix from budget entries.
	bench="${bench%%-*}"

	# Find the allocs/op field: the field immediately before "allocs/op".
	# The bench output name carries a -<procs> suffix; strip it before
	# comparing so the match is exact (not prefix-loose).
	actual="$(
		awk -v b="$bench" \
			'{ name = $1; sub(/-[0-9]+$/, "", name); if (name == b) {for (i = 2; i <= NF; i++) if ($(i+1) == "allocs/op") {print $i; exit}}}' \
			"$bench_out"
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
