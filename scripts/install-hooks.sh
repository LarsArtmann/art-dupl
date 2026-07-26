#!/usr/bin/env bash
# Re-installs the disabled-linters guard into .git/hooks/pre-commit.
#
# The buildflow daemon auto-generates the pre-commit hook and periodically
# overwrites it. This script appends our guard so that exhaustruct/tagliatelle
# re-enabling is caught at commit time.
#
# Run: bash scripts/install-hooks.sh
# Safe to re-run (idempotent).

set -euo pipefail

HOOK=".git/hooks/pre-commit"
MARKER="# art-dupl disabled-linters guard"

if [ ! -f "$HOOK" ]; then
    echo "No pre-commit hook found at $HOOK"
    echo "Install buildflow first (buildflow precommit install), then re-run this script."
    exit 1
fi

if grep -qF "$MARKER" "$HOOK"; then
    echo "Guard already present in $HOOK (idempotent, nothing to do)."
    exit 0
fi

cat >> "$HOOK" << 'GUARD'

# art-dupl disabled-linters guard
if [ -f .golangci.yml ]; then
    bash scripts/check-disabled-linters.sh .golangci.yml
fi
GUARD

echo "Guard appended to $HOOK"
