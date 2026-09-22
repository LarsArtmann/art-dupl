#!/usr/bin/env bash
# Stale-shell doctor: detects the "GOTOOLCHAIN=local + older go" trap.
#
# Symptom: every go command fails with
#   go.mod requires go >= 1.27.x (running go 1.26.y; GOTOOLCHAIN=local)
# Cause: the active go binary predates go.mod's requirement. On this fleet
# that almost always means a stale direnv/nix shell after a flake bump.
#
# Usage: scripts/go-env-doctor.sh        # check + actionable guidance
# Exit codes: 0 = toolchain satisfies go.mod; 1 = stale/mismatch (message printed)
set -euo pipefail
cd "$(dirname "$0")/.."

required=$(awk '/^go [0-9]/ {print $2; exit}' go.mod)
if [ -z "$required" ]; then
	echo "go-env-doctor: no 'go <version>' directive found in go.mod" >&2
	exit 1
fi

if ! command -v go >/dev/null 2>&1; then
	echo "go-env-doctor: no go binary on PATH." >&2
	echo "  Fix: enter the project shell — 'direnv reload' (or 'nix develop')." >&2
	exit 1
fi

actual=$(go env GOVERSION 2>/dev/null || go version | awk '{print $3}')
gotoolchain=$(go env GOTOOLCHAIN 2>/dev/null || echo "?")

ver_ge() {
	# go1.27.1-style semver compare (major.minor.patch)
	[ "$(printf '%s\n%s\n' "$2" "$1" | sort -V | head -1)" = "$2" ]
}

if ver_ge "${actual#go}" "$required"; then
	echo "OK: $actual satisfies go.mod (requires go $required, GOTOOLCHAIN=$gotoolchain)"
	exit 0
fi

echo "STALE SHELL: active go is $actual but go.mod requires go >= $required." >&2
echo "  Why it matters: with GOTOOLCHAIN=$gotoolchain the toolchain will not" >&2
echo "  auto-upgrade, so builds, tests, and gopls all fail on version checks." >&2
echo "  Fix (pick the first that applies):" >&2
echo "    1. direnv reload            # re-enter the flake devShell (usual case)" >&2
echo "    2. nix develop              # if not using direnv" >&2
echo "    3. GOTOOLCHAIN=auto go ...  # one-off: let go fetch the required toolchain" >&2
echo "  Verify with: scripts/go-env-doctor.sh" >&2
exit 1
