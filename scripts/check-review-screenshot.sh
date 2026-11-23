#!/usr/bin/env bash
# Screenshot-freshness guard for the fleet-review dashboard (plan T22, f27).
#
# docs/review/libs/index-screenshot.png must reflect the CURRENT
# docs/review/libs/index.html. The stamp file records the dashboard's SHA-256
# at regen time; any dashboard edit without a matching regen fails --check.
#
# Modes:
#   --check (default)  verify the stamp matches the dashboard's current hash
#   --regen            recapture the screenshot with headless chromium and
#                      restamp (local-only; needs chromium, e.g. via
#                      `nix shell nixpkgs#chromium -c bash scripts/check-review-screenshot.sh --regen`)
#
# Regen recipe (what --regen automates, plus the manual trim):
#   1. chromium --headless --disable-gpu --no-sandbox --hide-scrollbars \
#        --screenshot=docs/review/libs/index-screenshot.png \
#        --window-size=1440,4300 file://$PWD/docs/review/libs/index.html
#   2. trim trailing background below the footer:
#      magick ... -crop 1440x<content-bottom-rounded-up-to-8>+0+0 +repage
#      (find the footer bottom edge from a bottom-strip render)
#   3. sha256sum docs/review/libs/index.html > docs/review/libs/index-screenshot.sha256
set -euo pipefail
cd "$(dirname "$0")/.."
HTML=docs/review/libs/index.html
SHOT=docs/review/libs/index-screenshot.png
STAMP=docs/review/libs/index-screenshot.sha256

die() { echo "SCREENSHOT STALE: $*" >&2; exit 1; }

[ -f "$HTML" ] || die "$HTML missing"
[ -f "$SHOT" ] || die "$SHOT missing"
[ -f "$STAMP" ] || die "$STAMP missing"

case "${1:---check}" in
  --check)
    expected=$(awk '/^[0-9a-f]{64}/ { print $1; exit }' "$STAMP")
    [ -n "$expected" ] || die "stamp file has no sha256 line"
    actual=$(sha256sum "$HTML" | cut -d' ' -f1)
    if [ "$actual" != "$expected" ]; then
      die "$HTML changed (sha $actual) after the screenshot was taken (stamp $expected).
     Regenerate: nix shell nixpkgs#chromium -c bash scripts/check-review-screenshot.sh --regen
     then trim the bottom per the recipe in scripts/check-review-screenshot.sh and re-stamp."
    fi
    echo "screenshot stamp OK ($SHOT matches $HTML)"
    ;;
  --regen)
    command -v chromium >/dev/null || die "chromium not on PATH (run via: nix shell nixpkgs#chromium -c bash $0 --regen)"
    chromium --headless --disable-gpu --no-sandbox --hide-scrollbars \
      --screenshot="$SHOT" --window-size=1440,4300 \
      "file://$(pwd)/$HTML" 2>/dev/null
    echo "captured $SHOT at 1440x4300"
    echo "NEXT (manual): trim the trailing background below the footer, then:"
    echo "  sha256sum $HTML > $STAMP"
    ;;
  *)
    echo "usage: $0 [--check|--regen]" >&2
    exit 2
    ;;
esac
