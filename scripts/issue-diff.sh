#!/bin/bash

# Compare linter issues between runs

set -e

BASELINE=${1:-424}

echo "========================================="
echo "Linter Issue Comparison"
echo "========================================="
echo ""

# Run linter and count issues
OUTPUT=$(golangci-lint run 2>&1)
TOTAL=$(echo "$OUTPUT" | grep -E "^[0-9]+ issues" | grep -oE "^[0-9]+")

echo "Baseline: $BASELINE issues"
echo "Current:  $TOTAL issues"
echo ""

if [ "$TOTAL" -lt "$BASELINE" ]; then
    CHANGE=$((BASELINE - TOTAL))
    echo "✅ IMPROVED: -$CHANGE issues"
elif [ "$TOTAL" -gt "$BASELINE" ]; then
    CHANGE=$((TOTAL - BASELINE))
    echo "❌ REGRESSED: +$CHANGE issues"
else
    echo "➡️  NO CHANGE: $TOTAL issues"
fi

echo ""
echo "========================================="
echo "Issue Breakdown:"
echo "========================================="
echo "$OUTPUT" | grep "^\*" | sort -t: -k2 -rn || echo "No issues found"

echo ""
echo "========================================="
echo "Complete!"
echo "========================================="
