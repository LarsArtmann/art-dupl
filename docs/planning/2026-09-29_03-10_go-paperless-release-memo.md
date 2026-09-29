# go-paperless Release Memo — v0.4.3 (findByName consolidation + gosec fix)

**Date:** 2026-09-29 (SUPERB v2 M19; execution is Lars's call in that repo — this is a follow-the-recipe memo)
**Repo state verified 2026-09-29:** `main == origin/main` (0 ahead), consolidation `04c32dc` (client.go + example_test.go, findByName) pushed; `[Unreleased]` carries the gosec G120 suppression fix (`ba786b8`); `go build ./...` green on the local 1.27 toolchain (with GOEXPERIMENT unset); flake `version = "0.4.2"`; last tag `v0.4.2` (2026-09-23).

## Recommended version: v0.4.3

Both unreleased entries are fixes (client behavior repair + CI tooling); no new API surface.

## Exact sequence (go-release skill: verify → CHANGELOG → bump → tag → push → pkg.go.dev)

```bash
cd ~/projects/go-paperless
env -u GOEXPERIMENT CGO_ENABLED=1 go test ./...   # 1. full suite, cache-busting
# 2. CHANGELOG: rename `## [Unreleased]` -> `## [0.4.3] - <date>`, open a fresh
#    empty [Unreleased], add the `[0.4.3]` compare link def at the bottom.
# 3. flake.nix: version = "0.4.2" -> "0.4.3" (line ~49).
git add CHANGELOG.md flake.nix
git commit -m "chore(release): cut v0.4.3"        # NOT via the daemon
git tag -s v0.4.3 -m "findByName consolidation + standalone-gosec suppression fix"
nix build 2>/dev/null && ./result/bin/*/version || true   # 4. versioned binary if flake builds
git push --follow-tags origin main                # 5. atomic branch + tag
gh release create v0.4.3 --title "go-paperless v0.4.3" --notes-file <(awk '/^## \[0.4.3\]/{f=1;next} /^## \[/{f=0} f' CHANGELOG.md)
# 6. verify the rendered release body lists BOTH fixes (the v0.7.2 art-dupl lesson:
#    generic notes ship silently); check pkg.go.dev picks up the tag.
```

## Entry in the art-dupl ledger

TODO_LIST #20 strikes when the tag lands (`git -C ~/projects/go-paperless tag -l v0.4.3` non-empty AND the GH release renders the two fixes).
