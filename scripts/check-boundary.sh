#!/usr/bin/env bash
# check-boundary.sh — the task-boundary gate.
#
# WHY: the 2026-09-29 jsonutil incident — a foreign working-tree edit rode a
# docs sweep into a pushed commit and lived ~90 minutes because task boundaries
# ran scoped test runs, and stale `go test` caches masked the breakage for days
# before that. This script is the non-negotiable gate at every task boundary.
#
# Usage:
#   scripts/check-boundary.sh          # fast profile (every task boundary)
#   scripts/check-boundary.sh --full   # complete race suite (close-outs)
#
# Fast profile: build + vet + race on the wire-sensitive packages
# (internal/ — jsonutil, jsonv2gate — plus config and cmd, which carries the
# docs count gate). Uses -count=1 so stale caches cannot mask a regression.
# Full profile: the complete CGO-enabled race suite, also -count=1.

set -euo pipefail
cd "$(dirname "$0")/.."

mode="fast"
if [ "${1:-}" = "--full" ]; then
	mode="full"
elif [ -n "${1:-}" ]; then
	echo "unknown argument: $1 (supported: --full)" >&2
	exit 2
fi

start=$SECONDS

if [ "$mode" = "full" ]; then
	echo "== boundary gate (full): build + vet + race ./... =="
	go build ./...
	go vet ./...
	CGONOTE=""
	if ! CGO_ENABLED=1 go test -count=1 -race ./...; then
		echo "BOUNDARY GATE FAILED (full race suite)" >&2
		exit 1
	fi
else
	echo "== boundary gate (fast): build + vet + race on internal/config/cmd =="
	go build ./...
	go vet ./...
	if ! CGO_ENABLED=1 go test -count=1 -race ./internal/... ./config/... ./cmd/...; then
		echo "BOUNDARY GATE FAILED (fast profile)" >&2
		echo "note: this profile covers internal/, config/, cmd/ — for close-outs run scripts/check-boundary.sh --full" >&2
		exit 1
	fi
fi

echo "boundary gate OK ($((SECONDS - start))s, $mode profile)"
