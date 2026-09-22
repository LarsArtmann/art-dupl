#!/usr/bin/env bash
# Pre-release gate: codifies the go-release skill's Phase 0-4 checks plus the
# v0.7.0 release-wave lessons (tag collision, red CI, toolchain pinning).
#
# Usage:
#   scripts/pre-release-check.sh [--version vX.Y.Z] [--race] [--lint] [--skip-ci]
#
# Checks (in order):
#   1.  Clean git tree                       (no uncommitted changes)
#   2.  Toolchain pinning                    (go-env-doctor: active go >= go.mod)
#   3.  No local `replace` directives        (poison in published tags)
#   4.  No pseudo-version placeholders       (v0.0.0-...-000000000000)
#   5.  go mod tidy leaves go.mod/go.sum unchanged
#   6.  go mod verify
#   7.  go build ./... && go vet ./...
#   8.  go test ./...                        (--race adds the race detector)
#   9.  golangci-lint via buildflow          (--lint)
#   10. Tag collision: --version must not exist on the remote
#       (git ls-remote; a colliding tag is UNRECOVERABLE once the proxy sees it)
#   11. CI green on the default branch       (gh; --skip-ci to override explicitly)
set -euo pipefail
cd "$(dirname "$0")/.."

RACE=0
LINT=0
SKIP_CI=0
VERSION=""

while [ $# -gt 0 ]; do
	case "$1" in
	--version)
		VERSION="${2:?--version needs a value}"
		shift 2
		;;
	--race)
		RACE=1
		shift
		;;
	--lint)
		LINT=1
		shift
		;;
	--skip-ci)
		SKIP_CI=1
		shift
		;;
	*)
		echo "unknown flag: $1" >&2
		exit 2
		;;
	esac
done

fail() {
	echo "FAIL: $1" >&2
	echo "Pre-release gate blocked. Fix the above before tagging." >&2
	exit 1
}

say() { printf '  %s\n' "$1"; }

echo "== pre-release-check =="

# 1. Clean tree
[ -z "$(git status --porcelain)" ] || fail "working tree is dirty — commit or stash first"
say "clean tree"

# 2. Toolchain
scripts/go-env-doctor.sh >/dev/null || fail "stale toolchain (run scripts/go-env-doctor.sh for guidance)"
say "toolchain satisfies go.mod"

# 3. replace directives
if grep -qE '^replace' go.mod; then
	grep -nE '^replace' go.mod >&2
	fail "go.mod has replace directives — remove or move to go.work before tagging"
fi
say "no replace directives"

# 4. pseudo-version placeholders
if grep -rE 'v0\.0\.0-00010101000000-000000000000' --include=go.mod --include=go.sum .; then
	fail "pseudo-version placeholder found"
fi
say "no pseudo-version placeholders"

# 5. tidy idempotence
cp go.mod /tmp/prc-go.mod
cp go.sum /tmp/prc-go.sum
go mod tidy
diff -q go.mod /tmp/prc-go.mod >/dev/null || fail "go mod tidy changes go.mod — commit the result first"
diff -q go.sum /tmp/prc-go.sum >/dev/null || fail "go mod tidy changes go.sum — commit the result first"
say "go mod tidy clean"

# 6. verify
go mod verify >/dev/null || fail "go mod verify failed"
say "go mod verify ok"

# 7-8. build, vet, test
go build ./... || fail "go build failed"
go vet ./... || fail "go vet failed"
say "build + vet ok"
if [ "$RACE" -eq 1 ]; then
	go test -race -count=1 ./... || fail "race tests failed"
else
	go test -count=1 ./... || fail "tests failed"
fi
say "tests ok"

# 9. lint
if [ "$LINT" -eq 1 ]; then
	buildflow -s golangci-lint || fail "golangci-lint findings remain"
	say "lint ok"
fi

# 10. tag collision
if [ -n "$VERSION" ]; then
	echo "$VERSION" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$' ||
		fail "'$VERSION' is not a valid semver tag (expected vX.Y.Z)"
	if git ls-remote --exit-code --tags origin "refs/tags/${VERSION}" >/dev/null 2>&1; then
		fail "tag $VERSION already exists on the remote — pick a new version; NEVER move a tag"
	fi
	say "$VERSION does not collide with the remote"
else
	say "SKIP: pass --version vX.Y.Z to run the tag-collision check"
fi

# 11. CI green on default branch
if [ "$SKIP_CI" -eq 1 ]; then
	say "SKIP: CI gate overridden by --skip-ci"
elif command -v gh >/dev/null 2>&1; then
	default_branch=$(gh repo view --json defaultBranchRef --jq .defaultBranchRef.name)
	red=$(gh run list --branch "$default_branch" --limit 5 --json conclusion \
		--jq '[.[] | select(.conclusion == "failure" or .conclusion == "cancelled")] | length')
	[ "$red" -eq 0 ] || fail "$red of the last 5 runs on $default_branch are red — green CI is a release prerequisite"
	say "CI green on $default_branch (last 5 runs)"
else
	fail "gh CLI not available — cannot verify CI state (use --skip-ci to override consciously)"
fi

echo "PASS: all pre-release checks green."
[ -n "$VERSION" ] || echo "Reminder: re-run with --version vX.Y.Z before tagging."
