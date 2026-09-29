#!/usr/bin/env bash
# check-intent.sh — concurrent-session tripwire.
#
# WHY: concurrent sessions and the auto-commit daemon are the norm in this
# repo. The 2026-09-29 jsonutil incident: a foreign working-tree edit rode a
# docs sweep into a pushed commit and nobody named it. This script makes the
# naming mechanical: a session declares the files it intends to touch in a
# manifest; anything else in `git status` gets named UNEXPECTED before a
# commit can sweep it into history.
#
# Usage:
#   scripts/check-intent.sh             # report; exit 0 always (advisory)
#   scripts/check-intent.sh --strict    # exit 1 when unexpected files exist
#
# Manifest: `.check-intent` in the repo root (override: CHECK_INTENT_FILE env).
# One glob-or-path per line (bash pattern semantics, `*` allowed); `#` comments
# and blank lines ignored. The manifest is a LOCAL session artifact — it is
# gitignored, never committed. Opt-in: without a manifest the tripwire only
# reminds you it is not enforced (and --strict fails, since nothing is covered).

set -euo pipefail
cd "$(dirname "$0")/.."

strict=0
[ "${1:-}" = "--strict" ] || [ -z "${1:-}" ] || { echo "usage: check-intent.sh [--strict]" >&2; exit 2; }
[ "${1:-}" = "--strict" ] && strict=1

manifest="${CHECK_INTENT_FILE:-.check-intent}"

status=$(git status --porcelain --untracked-files=all)
if [ -z "$status" ]; then
	echo "intent tripwire: working tree clean"
	exit 0
fi

if [ ! -f "$manifest" ]; then
	echo "intent tripwire: no manifest at $manifest — nothing is enforced"
	[ "$strict" -eq 1 ] && exit 1
	exit 0
fi

unexpected=0
while IFS= read -r entry; do
	[ -n "$entry" ] || continue
	path=${entry:3}
	path=${path#\"}; path=${path%\"}
	path=${path//\"\\\"\"/\"}
	matched=0
	while IFS= read -r pat; do
		[ -n "$pat" ] || continue
		case "$pat" in \#*) continue ;; esac
		# shellcheck disable=SC2254
		case "$path" in
			$pat) matched=1; break ;;
		esac
	done < "$manifest"
	if [ "$matched" -eq 0 ]; then
		echo "UNEXPECTED: $entry"
		unexpected=$((unexpected + 1))
	fi
done <<< "$status"

if [ "$unexpected" -eq 0 ]; then
	echo "intent tripwire: every dirty file is covered by $manifest"
	exit 0
fi

echo "intent tripwire: $unexpected file(s) outside the declared intent — name them before committing (foreign session? stop and ask)"
[ "$strict" -eq 1 ] && exit 1
exit 0
