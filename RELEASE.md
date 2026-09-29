# Release Checklist

Follow this checklist for every release. No exceptions.

## 1. Pre-Release Verification

Run ALL of these. If any fails, fix before proceeding.

```bash
export GOEXPERIMENT=jsonv2

templ generate                # regenerate templ files
go build ./...                # must compile clean
go test ./...                 # all tests pass
CGO_ENABLED=1 go test -race ./...  # race detector clean
golangci-lint run --timeout 5m ./...  # zero lint issues
nix flake check               # all 8 checks pass
```

## 2. CHANGELOG Update

- Add the new version section at the top of `CHANGELOG.md`
- List all user-facing changes (features, fixes, breaking changes)
- Add the "Compare" footer link: `[Full changelog](https://github.com/LarsArtmann/art-dupl/compare/vPREV...vNEW)`
- Verify no `Unreleased` section leaks into the release

## 3. Version Bump

- Update `version` in `flake.nix` (the explicit string, NOT `self.rev`). **Known stale-bump trap (found 2026-09-29):** the flake said `0.7.0` while v0.7.1/v0.7.2 were already tagged — the bump step was skipped twice. Verify the CURRENT flake version matches the LAST tag before bumping, or you cannot tell a stale flake from a fresh one.
- Verify `art-dupl version` prints the correct version after `nix build`
- **Do NOT commit via the daemon** — stage manually: `git add CHANGELOG.md flake.nix && git commit -m "chore(release): cut vX.Y.Z"`

## 4. Tag and Sign

```bash
git tag -s v0.X.0 -m "art-dupl v0.X.0"
git tag -v v0.X.0              # verify signature
```

## 5. Build and Verify Binary

```bash
nix build                     # produces versioned binary
./result/bin/art-dupl version  # MUST print the version string, not a commit hash
```

## 6. Push and Release

```bash
git push --follow-tags origin fork  # atomic branch + tag push
gh release create vX.Y.0 --title "art-dupl vX.Y.0" --notes-file <notes-file>
```

Draft the notes FILE from the CHANGELOG section BEFORE creating the release
(preferred over `--generate-notes`, which produced the v0.7.2 generic-boilerplate
body). Do NOT use `--notes-from-tag` (too terse).

**Verify the rendered body actually contains the changes** (v0.7.2 shipped with an
empty `## Changelog` stub and pure install boilerplate — the gap was caught three
days later). If the body is generic, backfill without clobbering the install
section: fetch the body (`gh release view vX.Y.Z --json body -q .body`), replace
the `## Changelog` stub with the CHANGELOG section's content, write back with
`gh release edit vX.Y.Z --notes-file`.

## 7. Post-Release

- Verify the GitHub Release page renders correctly
- Verify the release assets are downloadable
- Reopen an empty `## [Unreleased]` section at the CHANGELOG top (with its compare link def `[Unreleased]: https://github.com/LarsArtmann/art-dupl/compare/vX.Y.Z...HEAD`)
- Update `ROADMAP.md` if any items were completed
- Announce in relevant channels

## Quality Gate Reminders

- **Never skip `go test -race`**. The v0.4.0 release shipped without it.
- **Never skip `nix flake check`**. It catches formatting and reproducibility issues.
- **Always verify the tag signature**. Unsigned or unverified tags break trust.
- **Always test the binary** from `nix build`, not just `go build`. The Nix build includes ldflags for version embedding.
- **Verify `art-dupl version` prints the version string** (e.g. `0.5.1`), NOT a commit hash. If it prints a hash, the `flake.nix` version string was not bumped.
- **Verify all new `--flag` entries appear in `HOW_TO_USE.md`**. Use: `grep -F -- '--flagname' HOW_TO_USE.md`
- **Commit the CHANGELOG + version bump manually** as `chore(release): cut vX.Y.Z`. Do NOT rely on the daemon for release commits.
- **Use `git push --follow-tags`** for atomic branch + tag push.
- **Pre-commit hook**: if buildflow reinstalled the hook, re-apply the guard: `bash scripts/install-hooks.sh`
```
