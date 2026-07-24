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

- Update `cmd/version.go` default values if not using ldflags
- Update `flake.nix` version string if applicable
- Verify `art-dupl version` prints the correct version

## 4. Tag and Sign

```bash
git tag -s v0.X.0 -m "art-dupl v0.X.0"
git tag -v v0.X.0              # verify signature
```

## 5. Build and Verify Binary

```bash
nix build                     # produces versioned binary
./result/bin/art-dupl version  # verify version string
```

## 6. Push and Release

```bash
git push origin v0.X.0        # push the signed tag
gh release create v0.X.0 --title "art-dupl v0.X.0" --notes-from-tag
```

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
