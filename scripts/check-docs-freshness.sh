#!/usr/bin/env bash
# Docs-health freshness gate: warns when a living doc's "Last Updated"
# stamp drifts too far behind the repo's HEAD date, so documentation rot
# is visible before it lies. Strict mode (STRICT=1) exits non-zero — wire
# into `nix flake check` only after one clean cycle in warn mode (the
# gate never observed failing is untested; Verschlimmbesserung guard
# from docs/planning/2026-09-28_21-57_SUPERB-docs-health-completion-plan.md M13).
#
# Usage: scripts/check-docs-freshness.sh [--strict]
#   --strict   exit 1 on any stale doc (CI mode; default is warn-only)

set -uo pipefail

MAX_AGE_DAYS=30
strict=0
[[ "${1:-}" == "--strict" ]] && strict=1

head_date="$(git log -1 --format=%cd --date=short)"
head_ts="$(date -d "$head_date" +%s 2>/dev/null || date -j -f '%Y-%m-%d' "$head_date" +%s)"
stale=0

check_doc() {
	local file="$1"
	[[ -f "$file" ]] || return 0
	local stamp
	stamp="$(grep -m1 -oE 'Last Updated\*\*:?[~ ]*[0-9]{4}-[0-9]{2}-[0-9]{2}|Last Updated:?\*\* [0-9]{4}-[0-9]{2}-[0-9]{2}|"Last Updated:"? [0-9]{4}-[0-9]{2}-[0-9]{2}' "$file" 2>/dev/null | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | head -1)"
	if [[ -z "$stamp" ]]; then
		echo "WARN $file: no Last-Updated stamp found (add one: **Last Updated:** YYYY-MM-DD)"
		return 0
	fi
	local ts age
	ts="$(date -d "$stamp" +%s 2>/dev/null || date -j -f '%Y-%m-%d' "$stamp" +%s)"
	age=$(( (head_ts - ts) / 86400 ))
	if (( age > MAX_AGE_DAYS )); then
		echo "STALE $file: Last Updated $stamp is $age days behind HEAD ($head_date)"
		stale=$((stale + 1))
	fi
}

for doc in FEATURES.md TODO_LIST.md ROADMAP.md README.md AGENTS.md HOW_TO_USE.md; do
	check_doc "$doc"
done

# Count drift is NOT this script's job: the Go count gate
# (cmd/docs_health_counts_test.go) re-derives documented numbers from code
# and fails CI on drift — run `go test -run 'Counts|CountGate' ./cmd/`.

if (( stale > 0 )); then
	if (( strict )); then
		echo "FAIL: $stale stale living doc(s) (strict mode)" >&2
		exit 1
	fi
	echo "WARN: $stale stale living doc(s) — run a docs-health pass (warn-only; use --strict to enforce)"
fi
exit 0
