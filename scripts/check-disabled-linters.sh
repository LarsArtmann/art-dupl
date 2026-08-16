#!/usr/bin/env bash
# CI guard + auto-fixer: verify that disabled linters are not re-enabled in .golangci.yml.
#
# The auto-committer has re-added exhaustruct and tagliatelle to the enable
# list multiple times. This script detects them and, when the file is writable,
# auto-removes them so the fix is immediate. In read-only contexts (Nix sandbox,
# CI containers) it fails the build instead.
#
# Usage: ./scripts/check-disabled-linters.sh [config-file]
# Exit codes: 0 = pass (or auto-fixed), 1 = disabled linter found and not fixable

set -euo pipefail

CONFIG="${1:-.golangci.yml}"

if [ ! -f "$CONFIG" ]; then
	echo "ERROR: $CONFIG not found" >&2
	exit 1
fi

# Disabled linters that must never appear in enable or exclusion sections.
DISABLED_LINTERS="exhaustruct tagliatelle"

FAILED=0

for linter in $DISABLED_LINTERS; do
	# Check for the linter as an enabled entry ("- lintername") or as a settings key ("lintername:")
	# Comments mentioning the name are allowed for documentation purposes.
	if grep -qE "^[[:space:]]*-[[:space:]]+${linter}\b|^[[:space:]]*${linter}:" "$CONFIG"; then
		# Attempt auto-fix: remove the offending line(s) if the file is writable.
		if [ -w "$CONFIG" ]; then
			# Remove lines that enable the linter ("- lintername").
			sed -i "/^[[:space:]]*-[[:space:]]*${linter}\b/d" "$CONFIG"
			# Remove orphaned settings blocks ("lintername:").
			sed -i "/^[[:space:]]*${linter}:/d" "$CONFIG"
			echo "WARN: auto-removed '${linter}' from $CONFIG" >&2
			echo "  This linter is intentionally disabled (see CHANGELOG / AGENTS.md)." >&2
		else
			echo "FAIL: '${linter}' is enabled or configured in $CONFIG (read-only, cannot auto-fix)" >&2
			echo "  This linter has been intentionally disabled." >&2
			echo "  Remove it from both the enable list and any settings blocks." >&2
			echo "  (Comments mentioning it are fine — only enable/settings entries trigger this guard.)" >&2
			FAILED=1
		fi
	fi
done

if [ "$FAILED" -eq 0 ]; then
	echo "OK: no disabled linters in $CONFIG"
fi

exit "$FAILED"
