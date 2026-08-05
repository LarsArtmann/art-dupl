#!/usr/bin/env bash
# Calibration script: runs art-dupl on Go projects and collects confidence data.
#
# Usage: ./scripts/calibrate-confidence.sh <project-dir> [<project-dir> ...]
#
# Outputs a summary of clone groups, their actionability classification, and
# confidence values. Use this to evaluate whether the extractability engine's
# thresholds need tuning.

set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: $0 <project-dir> [<project-dir> ...]" >&2
    exit 1
fi

BINARY="${ARTDUPL_BINARY:-$(go build -o /tmp/art-dupl-calibrate ./cmd/art-dupl/ && echo /tmp/art-dupl-calibrate)}"

total_projects=0
total_clones=0
total_actionable=0
total_non_actionable=0

echo "# art-dupl Confidence Calibration Report"
echo "# Date: $(date -I)"
echo "# Threshold: 5, Mode: semantic"
echo ""

for dir in "$@"; do
    if [ ! -d "$dir" ]; then
        echo "WARN: $dir not found, skipping" >&2
        continue
    fi

    total_projects=$((total_projects + 1))
    proj_name=$(basename "$dir")

    output=$("$BINARY" --threshold 5 --semantic --explain --quiet "$dir" 2>/dev/null || true)

    clone_count=$(echo "$output" | grep -c "^  explain:" || true)
    actionable_count=$(echo "$output" | grep -c "actionable" || true)
    non_actionable_count=$(echo "$output" | grep -c "non-actionable" || true)

    total_clones=$((total_clones + clone_count))
    total_actionable=$((total_actionable + actionable_count))
    total_non_actionable=$((total_non_actionable + non_actionable_count))

    echo "## $proj_name"
    echo "- Clone groups: $clone_count"
    echo "- Actionable: $actionable_count"
    echo "- Non-actionable: $non_actionable_count"

    if [ "$clone_count" -gt 0 ]; then
        echo "$output" | grep "^  explain:" | while IFS= read -r line; do
            echo "  - $line"
        done
    fi

    echo ""
done

echo "---"
echo "## Summary"
echo "- Projects tested: $total_projects"
echo "- Total clone groups: $total_clones"
echo "- Actionable: $total_actionable"
echo "- Non-actionable: $total_non_actionable"
