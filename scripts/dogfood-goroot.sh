#!/usr/bin/env bash
# Go-1.27 product-correctness dogfood: run the CLI over GOROOT packages that
# exercise the newest syntax and the type checker's hardest corners.
#
# - src/math/rand/v2: ships a GENERIC METHOD ((*Rand).N[Int], legalized in
#   Go 1.27 #77273) — parse, normalize, serialize, and both go/types-backed
#   loaders (--type-aware, --suggest-generics) must tolerate it.
# - src/go/types: alias-heavy (gotypesalias is permanently on in 1.27 —
#   go/types always produces Alias nodes now) — type-aware hashing must not
#   crash or degenerate on alias-dense code.
#
# Exit 0 = all four runs completed with a clean type load. Findings counts
# are informational (stdlib drifts constantly); the gate is crash-freedom
# and the loader ✅. Run from the repo root.

set -euo pipefail

cd "$(dirname "$0")/.."
GOROOT="$(go env GOROOT)"
BIN="$(mktemp /tmp/art-dupl-dogfood.XXXXXX)"
trap 'rm -f "$BIN"' EXIT

go build -o "$BIN" ./cmd/art-dupl

fail=0

run() {
    local label="$1"; shift
    echo "=== $label ==="
    if ! out="$("$@" 2>&1)"; then
        echo "FAILED: $label"; echo "$out" | tail -5; fail=1; return
    fi
    if ! echo "$out" | grep -q "✅"; then
        echo "TYPE LOAD INCOMPLETE: $label"; echo "$out" | tail -5; fail=1; return
    fi
    echo "$out" | grep -E "Detected|found .* clones" | tail -2 || true
}

run "rand/v2 plain"              "$BIN" -t 1 --no-actionability "$GOROOT/src/math/rand/v2"
run "rand/v2 --type-aware"       "$BIN" --type-aware -t 1 --no-actionability "$GOROOT/src/math/rand/v2"
run "rand/v2 --suggest-generics" "$BIN" --suggest-generics -t 1 --no-actionability "$GOROOT/src/math/rand/v2"
run "go/types alias-heavy --type-aware" "$BIN" --type-aware -t 2 --no-actionability "$GOROOT/src/go/types"

if [ "$fail" -ne 0 ]; then
    echo "DOGFOOD FAILED" >&2
    exit 1
fi

echo "DOGFOOD OK"
