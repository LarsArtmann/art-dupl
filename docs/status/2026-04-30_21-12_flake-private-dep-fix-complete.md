# Status Report: 2026-04-30 21:12

**Session Focus**: Fix flake.nix private dependency handling — eliminate vendor/ requirement

---

## a) FULLY DONE

### 1. Nix Flake Private Dependency Fix (`flake.nix`)

**Problem**: `vendorHash = null` required a `vendor/` directory that didn't exist. The original comment said "Vendored dependencies are required because gogenfilter is a private repository." But there was no vendor/ on disk and no vendor/ in git — the build was broken.

**Root Cause**: `gogenfilter` is a private GitHub repo. Nix builds run in a sandbox with no network access and no SSH credentials, so `go mod download` cannot fetch it during `buildGoModule`.

**Solution Implemented**: Two-phase approach that eliminates the need for `vendor/` entirely:

1. **Flake input** (`flake.nix:6-9`): `gogenfilter` is pre-fetched via SSH during flake evaluation (outside sandbox), locked in `flake.lock`
2. **`overrideModAttrs`** (`flake.nix:46-72`): In the goModules derivation (fixed-output, sandboxed), a dummy module with matching go.mod dependencies replaces gogenfilter via `go mod edit -replace`. This lets `go mod vendor` include all public deps + gogenfilter's transitive deps without network access
3. **`preBuild`** (`flake.nix:74-81`): In the main build (regular derivation, can reference store paths), the dummy is swapped for the real gogenfilter from the flake input, `vendor/modules.txt` patched, and `go.mod` updated with matching replace directive
4. **`vendorHash`** (`flake.nix:42-45`): Proper SRI hash of the vendored output — no more `null`

**Verification**:
- `nix build` — produces working binary (`art-dupl version 6658fa9-dirty`)
- `nix flake check --no-build` — all checks pass
- `nix run . -- --help` — runs correctly

### 2. Research Conducted

- Analyzed `buildGoModule` internals (goModules derivation, fixed-output constraints, preBuild ordering, overrideModAttrs mechanics)
- Explored 5+ failed approaches before finding the working one (direct replace in fixed-output, vendor directory permission issues, modules.txt consistency)
- Verified SSH access to gogenfilter and exact commit pinning

---

## b) PARTIALLY DONE

### Nothing partially done this session.

---

## c) NOT STARTED

1. **`nix flake check` with tests** — `nix flake check` passes metadata but the `test` check derivation hasn't been run (would run `go test ./...` inside Nix)
2. **AGENTS.md update** — Should document the new flake private dependency pattern for future sessions
3. **CI/CD integration** — The `flake.lock` change (new gogenfilter input) needs to be pushed and CI needs SSH key for gogenfilter
4. **Cross-platform build verification** — Only tested on `x86_64-linux`; darwin builds untested
5. **`printer/stats_visualization.go`** — Unrelated file showing as modified in git status; not investigated

---

## d) TOTALLY FUCKED UP

### Nothing fucked up. The build works.

**However**, the solution has a known fragility:

- The dummy go.mod in `overrideModAttrs` must exactly match gogenfilter's real go.mod (dependencies list). If gogenfilter adds a new dependency, the dummy must be updated AND `vendorHash` must be recomputed.
- This is a manual step that could be automated with a comment or script.

### Failed Approaches (for the record)

| # | Approach | Why It Failed |
|---|----------|---------------|
| 1 | `vendorHash = null` + `go mod vendor` in source tree | No vendor/ existed; broken by design |
| 2 | `go mod edit -replace` pointing to `${gogenfilter}` Nix store path directly in goModules | Fixed-output derivations cannot reference other store paths |
| 3 | Dummy with empty go.mod (just `package gogenfilter`) | `go mod vendor` needs actual package files to resolve imports |
| 4 | `cp -r vendor vendor-tmp` then modify | Symlinks from Nix store are read-only; cp -r preserves them |
| 5 | `cp -rL vendor vendor-tmp` then modify | Still failed on nested read-only files; `rm -rf` couldn't delete |
| 6 | Main `preBuild` inherited by goModules | preBuild is inherited; created vendor/ in goModules, triggering "vendor folder exists" check |
| 7 | Replace in main go.mod without fixing vendor/modules.txt | "inconsistent vendoring" — modules.txt said `./dummy`, go.mod said `./vendor/...` |

---

## e) WHAT WE SHOULD IMPROVE

1. **Dummy go.mod maintenance burden** — Extract gogenfilter's go.mod into a let binding or separate file so it's not duplicated inline
2. **vendorHash recomputation** — Could write a justfile recipe: `just update-vendor-hash` that sets to empty and rebuilds
3. **flake.lock staleness** — No CI check that flake.lock is up to date
4. **Comment quality** — The overrideModAttrs section needs a comment explaining WHY the dummy go.mod must match the real one
5. **Dev shell** — Dev shell doesn't reference gogenfilter; developers still need SSH access for local `go mod download`
6. **The `printer/stats_visualization.go` modification** — Untracked change sitting in the working tree; should be investigated or reverted

---

## f) Top 25 Things We Should Get Done Next

### High Priority (Build/Infrastructure)

1. Run `nix flake check` (full, including test derivation) to verify tests pass in Nix sandbox
2. Investigate and commit or revert the `printer/stats_visualization.go` change
3. Add a justfile recipe for `update-vendor-hash` to automate hash recomputation
4. Update AGENTS.md with the private dependency flake pattern documentation
5. Push changes to remote and verify CI passes
6. Extract the dummy go.mod into a Nix `let` binding to reduce maintenance burden
7. Add CI check for flake.lock staleness (GitHub Action)
8. Verify cross-platform builds (at minimum `nix flake check --all-systems`)
9. Ensure CI has SSH key configured for gogenfilter private repo access

### Medium Priority (Code Quality)

10. Run `just ci` (format, lint, test) to verify Go codebase is clean
11. Run `just check-coverage` to verify 80%+ threshold
12. Clean up old status reports in `docs/status/` — there are 200+ files, many stale
13. Review `flake.nix` for any other improvements (e.g., `proxyVendor = true` for faster builds)
14. Add `nix run .#test` as a convenience app for running tests via Nix
15. Document the two-phase dummy/replace pattern in a Nix-specific doc

### Lower Priority (Nice-to-Have)

16. Explore `gomod2nix` as an alternative to the dummy approach
17. Add `nix develop` documentation to README.md
18. Create a `.github/dependabot.yml` for nix flake updates
19. Add shell completion generation to the Nix build
20. Benchmark Nix build time vs justfile build time
21. Investigate Nix caching for CI (cachix or GitHub Actions cache)
22. Add `nix profile install` support documentation
23. Create a Homebrew formula update script tied to Nix build
24. Explore generating the dummy go.mod from the actual gogenfilter flake input
25. Evaluate migrating from justfile to pure Nix task runner (taskforge, etc.)

---

## g) Top #1 Question I Cannot Figure Out Myself

**Is there a way to extract the go.mod from the gogenfilter flake input at Nix evaluation time?**

The dummy go.mod in `overrideModAttrs` is a fragile copy of gogenfilter's real go.mod. If I could read `${gogenfilter}/go.mod` during Nix evaluation and use its contents to generate the dummy dynamically, the whole thing would be self-maintaining. The problem is:

1. `${gogenfilter}` is a Nix store path available at build time, but I need its contents during **evaluation** (to set `vendorHash`)
2. `builtins.readFile` could work but the go.mod parsing in Nix would be complex
3. Alternatively: could I run a Nix derivation that extracts the go.mod and feeds it back? But that creates a circular dependency

This would eliminate the entire maintenance burden of keeping the dummy in sync.

---

## Files Changed

| File | Change |
|------|--------|
| `flake.nix` | +34/-8 — Added gogenfilter flake input, two-phase private dep handling, proper vendorHash |
| `flake.lock` | Updated with gogenfilter input (SSH-pinned to rev 5957230e34ed) |

## Build Verification

```
$ nix build                       # SUCCESS — produces result/bin/art-dupl
$ ./result/bin/art-dupl --version # art-dupl version 6658fa9-dirty
$ nix flake check --no-build      # all checks passed
$ nix run . -- --help             # SUCCESS — shows help text
```
