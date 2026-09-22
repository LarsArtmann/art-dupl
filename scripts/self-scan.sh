#!/usr/bin/env bash
# Monthly self-scan routine: art-dupl scanning its own source at -t 1.
#
# The CI/nix gate only covers -t 5 (default threshold). This script runs the
# documented monthly -t 1 --type-aware self-scan so new clones are judged
# while they are small, not re-litigated from scratch later.
#
# Routine (first run of each month):
#   1. scripts/self-scan.sh                — canonical scan, prints the count
#   2. Judge every shown group: extract harmful clones, accept intentional
#      ones with a rationale (see the dedup-code skill).
#   3. Record per-group decisions in docs/SELF_CLEAN_LEDGER.md (append a new
#      dated section; keep prior sections for history).
#
# A scan failure (build error, bad flag) fails the script; a rising
# shown-count is a judgment call for a human, not a mechanical failure.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "== art-dupl monthly self-scan: -t 1 --type-aware ==" >&2
# go run ./cmd/art-dupl is the canonical dev channel (the `art-dupl` binary on
# PATH may predate the current tree, see docs/status/2026-09-22_21-47).
go run ./cmd/art-dupl -t 1 --type-aware .

echo
echo "== Next steps (monthly routine) ==" >&2
echo "1. Judge every shown group above: extract harmful, accept intentional." >&2
echo "2. Append decisions to docs/SELF_CLEAN_LEDGER.md under today's date." >&2
echo "3. Deep dive when needed: add --no-actionability --show-suppressed." >&2
exit 0
