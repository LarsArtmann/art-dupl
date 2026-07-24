#!/usr/bin/env bash
# CI guard: verify that disabled linters are not re-enabled in .golangci.yml.
#
# The auto-committer has re-added exhaustruct and tagliatelle to the enable
# list multiple times. This script fails the build if they appear in the
# linters.enable section or as orphaned exclusion rules.
#
# Usage: ./scripts/check-disabled-linters.sh
# Exit codes: 0 = pass, 1 = disabled linter found

set -euo pipefail

CONFIG="${1:-.golangci.yml}"

if [ ! -f "$CONFIG" ]; then
    echo "ERROR: $CONFIG not found" >&2
    exit 1
fi

# Disabled linters that must never appear in enable or exclusions
DISABLED_LINTERS="exhaustruct tagliatelle"

FAILED=0

for linter in $DISABLED_LINTERS; do
    # Check if the linter appears anywhere in the config file
    if grep -q "$linter" "$CONFIG"; then
        echo "FAIL: '$linter' found in $CONFIG" >&2
        echo "  This linter has been intentionally disabled." >&2
        echo "  Remove it from both the enable list and exclusion rules." >&2
        FAILED=1
    fi
done

if [ "$FAILED" -eq 0 ]; then
    echo "OK: no disabled linters in $CONFIG"
fi

exit "$FAILED"
