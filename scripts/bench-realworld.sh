#!/usr/bin/env bash
# Real-world end-to-end benchmark with stage-time split.
#
# Runs the art-dupl binary over an external repository fixture with --timing
# and collects per-stage wall/active times plus allocation counters.
#
# Usage:
#   scripts/bench-realworld.sh [fixture-dir]
#
# Environment:
#   RUNS       number of timed runs per mode (default 3)
#   PIN_CORES  cores for taskset pinning, e.g. "0-7,16-23" (default: unpinned)
#   OUT        output markdown file (default docs/benchmarks/realworld-<name>.md)
#   BINARY     prebuilt binary to benchmark (default: temp build of ./cmd/art-dupl)
#
# CPU pinning matters on multi-CCX AMD chips: the parallel search is
# L3-latency bound, so pinning to one L3 domain beats the full machine.
# See docs/benchmarks/CPU_TOPOLOGY.md.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture="${1:-${FIXTURE:-}}"
runs="${RUNS:-3}"

if [[ -z "$fixture" ]]; then
	echo "error: pass a fixture directory (or set FIXTURE)" >&2
	exit 2
fi

if [[ ! -d "$fixture" ]]; then
	echo "error: fixture directory $fixture does not exist" >&2
	exit 2
fi

fixture_abs="$(cd "$fixture" && pwd)"
fixture_name="$(basename "$fixture_abs")"
fixture_sha="$(git -C "$fixture_abs" rev-parse HEAD 2>/dev/null || echo "not-a-git-repo")"
out="${OUT:-$repo_root/docs/benchmarks/realworld-$fixture_name.md}"

if [[ -z "${BINARY:-}" ]]; then
	binary="$(mktemp -d)/art-dupl"
	(cd "$repo_root" && go build -o "$binary" ./cmd/art-dupl)
else
	binary="$BINARY"
fi

run_mode() {
	local label="$1"
	shift
	local runner=("$@")

	echo "=== $label ($runs runs) ===" >&2

	for i in $(seq 1 "$runs"); do
		"${runner[@]}" \
			--timing --quiet -t 15 "$fixture_abs" \
			>/dev/null 2>/tmp/art-dupl-bench-$i.err

		echo "--- run $i done ---" >&2
	done

	{
		echo "### $label"
		echo
		echo '```'
		for i in $(seq 1 "$runs"); do
			echo "run $i:"
			sed 's/^/  /' /tmp/art-dupl-bench-$i.err
		done
		echo '```'
		echo
	} >>"$out.body"

	# Keep the last run's raw timing for the summary parser.
	cp /tmp/art-dupl-bench-"$runs".err /tmp/art-dupl-bench-last-"$label".timing
}

echo "# Real-world benchmark: $fixture_name" >"$out"
{
	echo
	echo "- **Fixture**: $fixture_abs"
	echo "- **Pinned commit**: \`$fixture_sha\`"
	echo "- **art-dupl commit**: \`$(git -C "$repo_root" rev-parse HEAD)\`"
	echo "- **Date**: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	echo "- **Runs per mode**: $runs"
	echo "- **Cores**: ${PIN_CORES:-all (unpinned)}"
	echo "- **Command**: \`art-dupl --timing --quiet -t 15 <fixture>\`"
	echo
	echo "Stage semantics: \`parse\`/\`serialize\`/\`tree-build\` sum ACTIVE time across"
	echo "workers (can exceed wall on parallel runs); \`ingest\`/\`search\`/\`print\`/\`total\`"
	echo "are wall-clock phases that overlap where the pipeline streams."
} >>"$out"

rm -f "$out.body"

if [[ -n "${PIN_CORES:-}" ]]; then
	run_mode "pinned ($PIN_CORES)" taskset -c "$PIN_CORES" "$binary"
else
	run_mode "unpinned" "$binary"
fi

{
	echo "## Aggregated medians"
	echo
	echo "| stage | median |"
	echo "|---|---|"
	for stage in crawl parse serialize tree-build ingest search print total; do
		medians=()
		for f in /tmp/art-dupl-bench-last-*.timing; do
			v=$(awk -v s="$stage" '$1==s {print $2}' "$f")
			[[ -n "$v" ]] && medians+=("$v")
		done
		if [[ ${#medians[@]} -gt 0 ]]; then
			sorted=($(printf '%s\n' "${medians[@]}" | sort))
			echo "| $stage | ${sorted[$(( ${#sorted[@]} / 2 ))]} |"
		fi
	done
	echo
	echo "### Allocations (last run)"
	echo
	echo '```'
	grep 'allocations:' /tmp/art-dupl-bench-last-*.timing | tail -1 || true
	echo '```'
} >>"$out.body"

cat "$out.body" >>"$out"
rm -f "$out.body"
rm -f /tmp/art-dupl-bench-*.err /tmp/art-dupl-bench-*.timing

echo "wrote $out" >&2
