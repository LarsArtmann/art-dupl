#!/bin/bash

# Verify linting changes by running golangci-lint and showing issue counts

set -e

echo "========================================="
echo "Running golangci-lint..."
echo "========================================="
echo ""

# Run linter and capture output
OUTPUT=$(golangci-lint run 2>&1)

# Show summary
echo "$OUTPUT" | grep -E "^[0-9]+ issues" || echo "No issue summary found"

# Show issue breakdown
echo ""
echo "========================================="
echo "Issue Breakdown:"
echo "========================================="
echo "$OUTPUT" | grep "^\*" || echo "No issues found"

echo ""
echo "========================================="
echo "Verification complete!"
echo "========================================="
