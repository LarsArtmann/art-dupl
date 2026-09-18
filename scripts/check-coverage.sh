#!/usr/bin/env bash
# Snapshot per-package test coverage into docs/benchmarks/coverage-baseline.txt.
#
# Usage:
#   scripts/check-coverage.sh              # write the baseline snapshot
#   scripts/check-coverage.sh --print      # print to stdout only (no file write)
#
# The baseline is a committed snapshot for trend-watching, NOT a CI gate:
# coverage percentages move with test additions and refactors; a drop is a
# signal to look, not a failure. Regenerate after deliberate test changes:
#   scripts/check-coverage.sh && git add docs/benchmarks/coverage-baseline.txt
set -euo pipefail

cd "$(dirname "$0")/.."

export GOEXPERIMENT=jsonv2

# Coverage rows for these packages are noise, not signal:
#   - cmd/art-dupl: main package, exercised end-to-end by the BDD exec tests
#   - examples: doc-example code
#   - internal/testhelpers, internal/testutil: test support code for other tests
# Excluding them at the go test level also keeps the coverprofile clean.
run_coverage() {
	echo "# art-dupl coverage baseline — regenerate with scripts/check-coverage.sh"
	echo "# date: $(date -u +%Y-%m-%dT%H:%M:%SZ) — $(go version | cut -d' ' -f3)"
	echo "# test-only packages (cmd/art-dupl, examples, internal/testhelpers, internal/testutil) are excluded"
	echo "#"
	local pkgs
	pkgs=$(go list ./... | grep -vE "/(cmd/art-dupl|examples|internal/testhelpers|internal/testutil)$")
	go test -cover ${pkgs} -coverprofile=/tmp/artdupl-cover.out
	echo "#"
	echo -n "# total: "
	go tool cover -func=/tmp/artdupl-cover.out | tail -1
	rm -f /tmp/artdupl-cover.out
}

if [[ "${1:-}" == "--print" ]]; then
	run_coverage
else
	run_coverage >docs/benchmarks/coverage-baseline.txt
	echo "wrote docs/benchmarks/coverage-baseline.txt"
fi
