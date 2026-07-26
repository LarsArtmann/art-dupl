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

- Update `version` in `flake.nix` (the explicit string, NOT `self.rev`)
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
gh release create vX.Y.0 --title "art-dupl vX.Y.0" --generate-notes
```

Use `--generate-notes` or `--notes-file <changelog-section>` for full release notes.
Do NOT use `--notes-from-tag` (too terse).```

## 7. Post-Release

- Verify the GitHub Release page renders correctly
- Verify the release assets are downloadable
- Move the CHANGELOG section from "Unreleased" to the version heading
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
