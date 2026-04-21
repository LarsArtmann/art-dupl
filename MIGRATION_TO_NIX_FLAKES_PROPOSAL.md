# Migration to Nix Flakes — Proposal

**Date:** 2026-04-21  
**Status:** Draft  
**Scope:** Replace ad-hoc toolchain management with reproducible Nix Flakes + direnv  

---

## 1. Executive Summary

art-dupl currently has **no Nix configuration** — no `flake.nix`, no `shell.nix`, no `.envrc`. Development relies on globally-installed Go, golangci-lint, and just. This has caused recurring issues:

- **Build cache corruption** (3 sessions in a row, per `docs/status/2026-04-20_07-26_phase1-complete-status.md`)
- **Broken Nix-managed Go 1.26.1** with missing stdlib packages (`internal/gover`, `crypto/internal/fips140deps/godebug`)
- **Go version mismatch** between gopls (1.26.0) and CLI (1.26.1)
- **No reproducible onboarding** — each developer must manually install the right tool versions

A Nix Flakes migration would:

1. **Pin all toolchain versions** (Go, golangci-lint, just, ginkgo, benchstat) in a lockfile
2. **Provide one-command onboarding** (`nix develop` or automatic via direnv)
3. **Eliminate "works on my machine"** across Linux, macOS, and CI
4. **Enable reproducible builds** with `nix build`
5. **Support cross-compilation** natively through Nix

---

## 2. Current State Analysis

### 2.1 Build System Inventory

| Component | Current Tool | Version | Source |
|-----------|-------------|---------|--------|
| Build runner | `just` (Justfile) | Unknown | Global install |
| Alternative runner | `make` (Makefile) | Unknown | Global install |
| Go compiler | `go` | 1.26.0 (system) / 1.26.1 (broken) | Nix system profile |
| Linter | `golangci-lint` | v1.64.6 (CI pinned) | Global install |
| Test framework | `ginkgo` | v2.28.1 (via go.mod) | `go install` |
| Benchmark comparison | `benchstat` | latest | `go install` |
| Fuzz testing | `go test -fuzz` | Built-in | Go stdlib |

### 2.2 Justfile Recipes to Preserve

The Justfile is the primary build interface (95% of commands per AGENTS.md). All recipes must continue to work inside `nix develop`:

| Recipe | Purpose | Nix Considerations |
|--------|---------|-------------------|
| `default` | clean → check → test → build | All deps must be in PATH |
| `clean` | Remove dist/, cover.out | No changes needed |
| `test` | `go test -v -cover ./...` | Go must be correct version |
| `check` | `golangci-lint run` | golangci-lint in devShell |
| `build` | `go build -ldflags "-s -w" -trimpath` | CGO_ENABLED=0 |
| `test-race` | Race detector | Works with Nix Go |
| `coverage` | Coverage report | Works with Nix Go |
| `bench` | Benchmarks | Works with Nix Go |
| `check-coverage` | 80% threshold check | Requires `bc` |
| `test-fuzz` / `test-fuzz-long` | Fuzz testing | Works with Nix Go |
| `build-all` | Cross-compile 3 platforms | Requires Nix cross-compilation or keep as-is |
| `ci` | fmt → check → test | All tools in devShell |
| `fmt` | `gofmt -s -w .` | Go includes gofmt |

### 2.3 Makefile Specifics

The Makefile adds `GOEXPERIMENT=jsonv2` to all commands. This is **not** in the Justfile. The Nix build should decide whether to incorporate this or deprecate the Makefile.

### 2.4 Shell Scripts

| Script | Purpose | Nix Impact |
|--------|---------|------------|
| `scripts/issue-diff.sh` | Compare linter issue counts across runs | Requires `golangci-lint` in PATH |
| `scripts/verify-lint.sh` | Run linter and show issue breakdown | Requires `golangci-lint` in PATH |

### 2.5 CI Workflows

Four GitHub Actions workflows currently use `actions/setup-go@v5` and `extractions/setup-just@v2`. After Nix migration, CI can optionally use `nix develop` for full parity, but this is **not required** initially — CI can continue using current actions.

| Workflow | Trigger | Go Versions | OS Matrix | Key Steps |
|----------|---------|-------------|-----------|-----------|
| `build.yml` | Push/PR to main/fork | oldstable, stable | ubuntu, macos, windows | test + build |
| `checks.yml` | Push/PR to main/fork | stable | ubuntu | lint + test + race test |
| `art-dupl.yml` | Push/PR to main/fork | 1.26rc2 | ubuntu | build + test + race + coverage + self-analysis |
| `performance.yml` | Push/PR + daily cron | stable | ubuntu | benchmarks + memory profiling |

### 2.6 Go Module Specifics

- **Go version:** `go 1.26.0` in `go.mod`
- **CGO:** Disabled (`CGO_ENABLED: 0` in all CI workflows)
- **Local replace directive:** `replace github.com/LarsArtmann/gogenfilter => ../gogenfilter`
  - This means `nix build` needs access to a sibling directory or a vendored gogenfilter
  - For the `flake.nix`, we should handle this via `src` that includes the sibling or via flake inputs
- **No vendoring:** Dependencies fetched via `go mod download`

---

## 3. Proposed Architecture

### 3.1 File Structure

```
art-dupl/
├── flake.nix              # Flake definition (NEW)
├── flake.lock             # Pinned dependency versions (GENERATED)
├── .envrc                 # direnv integration (NEW)
├── Justfile               # Unchanged — works inside nix develop
├── Makefile               # Unchanged — works inside nix develop
├── scripts/               # Unchanged — works inside nix develop
├── go.mod                 # Unchanged
└── go.sum                 # Unchanged
```

### 3.2 Flake Outputs

```
flake.nix
├── packages.<system>.art-dupl        # Production binary (buildGoModule)
├── packages.<system>.default         # Alias → art-dupl
├── devShells.<system>.default        # Development environment
├── checks.<system>.test              # Run tests via Nix
├── checks.<system>.lint              # Run linter via Nix
├── checks.<system>.build             # Build verification
├── formatter.<system>                # gofmt via Nix
└── apps.<system>.art-dupl            # Run directly via nix run
```

### 3.3 DevShell Contents

The development shell must provide everything needed for the Justfile to work:

| Package | Purpose | nixpkgs Name |
|---------|---------|--------------|
| Go 1.26+ | Compiler, gofmt, go test | `go_1_26` or `go` |
| golangci-lint | Linting | `golangci-lint` |
| just | Build runner | `just` |
| ginkgo | BDD test runner | `ginkgo` (via `go install` in shellHook) |
| benchstat | Benchmark comparison | `benchstat` (via `go install` in shellHook) |
| bc | Coverage threshold math | `bc` |
| git | Version control | `git` |
| gopls | LSP (optional) | `gopls` |

### 3.4 Build Configuration

```nix
buildGoModule {
  pname = "art-dupl";
  version = "unstable";  # or derived from git

  # ldflags matching Justfile: -s -w -trimpath
  ldflags = [ "-s" "-w" ];
  trimDeps = true;

  # Static binary (no CGO)
  env.CGO_ENABLED = 0;

  # Build the CLI entrypoint
  subPackages = [ "cmd/art-dupl" ];

  # Disable tests in build derivation (run via checks instead)
  doCheck = false;
}
```

---

## 4. Implementation Plan

### Phase 1: Minimum Viable Flake (1-2 hours)

**Goal:** Dev shell that makes `just` recipes work with pinned dependencies.

#### Step 1.1: Create `flake.nix`

```nix
{
  description = "art-dupl — Fast code duplication detector for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        # Go version matching go.mod requirement
        goPkg = pkgs.go_1_26;
      in
      {
        packages = {
          default = self.packages.${system}.art-dupl;

          art-dupl = pkgs.buildGoModule {
            pname = "art-dupl";
            version = "0.0.0-unstable";

            src = pkgs.lib.cleanSourceWith {
              src = pkgs.lib.cleanSource ./.;
              filter = name: type:
                let
                  baseName = baseNameOf name;
                in
                  !(
                    # Exclude non-source directories
                    pkgs.lib.hasSuffix "dist" baseName
                    || pkgs.lib.hasSuffix "coverage.html" baseName
                    || pkgs.lib.hasSuffix "cover.out" baseName
                  );
            };

            vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Update after first build

            ldflags = [ "-s" "-w" ];
            env.CGO_ENABLED = 0;
            subPackages = [ "cmd/art-dupl" ];
            doCheck = false;

            meta = with pkgs.lib; {
              description = "Fast code duplication detector for Go projects";
              homepage = "https://github.com/LarsArtmann/art-dupl";
              license = licenses.mit;
              mainProgram = "art-dupl";
            };
          };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            goPkg
            golangci-lint
            just
            bc
            git
            gopls
          ];

          env.CGO_ENABLED = 0;

          shellHook = ''
            echo "art-dupl development environment"
            echo "Go: $(go version)"
            echo "golangci-lint: $(golangci-lint version)"
            echo "just: $(just --version)"
            echo ""
            echo "Available: just --list"
          '';
        };
      }
    );
}
```

#### Step 1.2: Create `.envrc`

```bash
# .envrc
if ! has nix_direnv_version || ! nix_direnv_version 3.0.0; then
  source_url "https://raw.githubusercontent.com/nix-community/nix-direnv/3.0.0/direnvrc" "sha256-21TMnI2xWX7HkSKj3GgmO8M8QZpRSgkFwYdGMeSiOig="
fi

use flake
```

#### Step 1.3: Update `.gitignore`

Add:
```
.direnv/
```

#### Step 1.4: Fix `vendorHash`

```bash
# Set fake hash, attempt build, copy real hash from error
nix build .#art-dupl
# Error will show the correct hash — paste it into flake.nix
```

#### Step 1.5: Verify DevShell

```bash
# Enter dev shell
nix develop

# Verify all tools
go version           # Should show go1.26.x
golangci-lint version
just --version

# Run full CI
just ci
```

### Phase 2: Nix Checks (30 min)

**Goal:** Run test/lint via `nix flake check` for CI parity.

```nix
checks = {
  test = pkgs.runCommand "art-dupl-tests" { } ''
    ${goPkg}/bin/go test -v ./...
    touch $out
  '';

  lint = pkgs.runCommand "art-dupl-lint" { } ''
    ${pkgs.golangci-lint}/bin/golangci-lint run --timeout 5m
    touch $out
  '';

  build = self.packages.${system}.art-dupl;
};
```

### Phase 3: Handle Local Replace Directive (30 min - 1 hour)

**Problem:** `go.mod` has `replace github.com/LarsArtmann/gogenfilter => ../gogenfilter`

**Options (in order of preference):**

#### Option A: Flake Input for gogenfilter (Recommended)

```nix
inputs = {
  nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  flake-utils.url = "github:numtide/flake-utils";
  gogenfilter.url = "github:LarsArtmann/gogenfilter";  # or path:../gogenfilter
  gogenfilter.flake = false;  # It's a Go module, not a flake
};

# In buildGoModule:
postConfigure = ''
  # Replace the local path with the flake input path
  rm -rf vendor/github.com/LarsArtmann/gogenfilter
  ln -s ${gogenfilter} vendor/github.com/LarsArtmann/gogenfilter
'';
```

#### Option B: Path Input (for local development)

```nix
inputs = {
  gogenfilter.url = "path:../gogenfilter";
  gogenfilter.flake = false;
};
```

#### Option C: Remove Replace Directive (if gogenfilter is published)

If gogenfilter is published to a Go module proxy, remove the replace directive entirely and reference it normally. This is the cleanest option if the module is stable.

#### Option D: Vendored Dependencies

Run `go mod vendor` and set `vendorHash = null` in buildGoModule. This locks all dependencies including gogenfilter in the vendor directory.

### Phase 4: CI Integration (30 min)

**Goal:** Use Nix in GitHub Actions for full reproducibility.

#### Option A: Nix-based CI (Full Reproducibility)

```yaml
# .github/workflows/nix-ci.yml
name: Nix CI

on:
  push:
    branches: [main, master, fork]
  pull_request:
    branches: [main, master, fork]

jobs:
  nix-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: cachix/install-nix-action@v30
        with:
          extra_nix_config: |
            experimental-features = nix-command flakes

      - uses: cachix/cachix-action@v15
        with:
          name: art-dupl
          signingKey: ${{ secrets.CACHIX_SIGNING_KEY }}

      - name: Check flake
        run: nix flake check

      - name: Build
        run: nix build

      - name: Run
        run: ./result/bin/art-dupl --version
```

#### Option B: Hybrid CI (Recommended Initially)

Keep existing CI workflows unchanged. Add Nix CI as an **additional** workflow. This provides a gradual migration path without breaking existing pipelines.

### Phase 5: Advanced Features (Optional)

#### 5.1 Cross-Compilation

Replace `just build-all` with Nix cross-compilation:

```nix
packages = {
  art-dupl-linux-amd64 = self.packages.${system}.art-dupl.overrideAttrs (old: {
    GOOS = "linux";
    GOARCH = "amd64";
  });

  art-dupl-darwin-amd64 = self.packages.${system}.art-dupl.overrideAttrs (old: {
    GOOS = "darwin";
    GOARCH = "amd64";
  });

  art-dupl-windows-amd64 = self.packages.${system}.art-dupl.overrideAttrs (old: {
    GOOS = "windows";
    GOARCH = "amd64";
  });
};
```

#### 5.2 Version Injection

Derive version from git tag or commit hash:

```nix
version = if (self ? shortRev) then self.shortRev else "dirty";

ldflags = [
  "-s" "-w"
  "-X main.version=${version}"
  "-X main.commit=${self.rev or "dirty"}"
];
```

#### 5.3 GOEXPERIMENT=jsonv2 Support

For the Makefile's `GOEXPERIMENT=jsonv2`:

```nix
env.GOEXPERIMENT = "jsonv2";  # In buildGoModule
# Or in devShell:
shellHook = ''
  export GOEXPERIMENT=jsonv2
'';
```

> **Decision needed:** Should the Nix build use `GOEXPERIMENT=jsonv2`? The Justfile does not use it, only the Makefile does. Recommend **not** including it in the default build unless there's a specific reason.

#### 5.4 Formatter

```nix
formatter = pkgs.nixpkgs-fmt;  # For .nix files
# Go formatting stays with `just fmt` (gofmt)
```

#### 5.5 Binary Cache (Cachix)

Set up [cachix.org](https://cachix.org) to cache Nix builds for the project. This avoids rebuilding from scratch on every CI run and for every developer.

---

## 5. gogenfilter Local Module — Deep Dive

The `replace` directive is the trickiest part of this migration:

```
replace github.com/LarsArtmann/gogenfilter => ../gogenfilter
```

### Impact Assessment

| Scenario | `nix develop` | `nix build` | CI |
|----------|---------------|-------------|-----|
| Replace directive active | Works (local path accessible) | Fails (sandboxed build) | Fails (no sibling dir) |
| Remove replace, use published module | Works | Works | Works |
| Flake input with path | Works | Works | Fails (no sibling) |
| Flake input with GitHub URL | Works | Works | Works |

### Recommendation

1. **Short term:** Use `path:../gogenfilter` as a flake input for local development
2. **Long term:** Publish gogenfilter to GitHub and remove the replace directive

The flake can support both modes:

```nix
inputs = {
  gogenfilter.url = "github:LarsArtmann/gogenfilter";
  # For local development, override with:
  # nix build --override-input gogenfilter path:../gogenfilter
};
```

---

## 6. Known Issues to Address

### 6.1 Go Build Cache Corruption

The recurring cache corruption documented in status reports is likely caused by:

1. **Multiple Go versions** on PATH simultaneously (system Nix Go vs auto-downloaded toolchain)
2. **Nix store Go** with broken stdlib packages
3. **go.work** forcing a different Go version than `go.mod`

**How Nix Flakes fixes this:**

- The devShell provides **exactly one Go version** in PATH
- No system Go conflicts — the shell is isolated
- `GOTOOLCHAIN=local` prevents auto-downloading different Go versions
- The `go.work` issue can be managed by setting `GOWORK=off` in the devShell

```nix
shellHook = ''
  export GOTOOLCHAIN=local
  export GOWORK=off
'';
```

### 6.2 Go Version Pinning

Current `go.mod` says `go 1.26.0` but the project has been using 1.26.1. The flake should:

1. Use the **same Go version** as `go.mod` specifies
2. Set `GOTOOLCHAIN=local` to prevent auto-downloads
3. Update `go.mod` if the team wants to standardize on a different version

### 6.3 gopls Version Mismatch

This should be resolved automatically — the devShell provides matching `go` and `gopls` from the same nixpkgs commit.

---

## 7. Migration Checklist

### Pre-Migration

- [ ] Verify Go version in `go.mod` matches desired Nix Go version
- [ ] Decide on gogenfilter strategy (publish to GitHub, flake input, or vendor)
- [ ] Decide whether `GOEXPERIMENT=jsonv2` should be in the Nix build
- [ ] Ensure `nix` is installed with flakes enabled

### Phase 1: Minimum Viable Flake

- [ ] Create `flake.nix` with devShell and buildGoModule
- [ ] Create `.envrc` for direnv integration
- [ ] Add `.direnv/` to `.gitignore`
- [ ] Fix `vendorHash` by running `nix build` and copying correct hash
- [ ] Verify `just ci` works inside `nix develop`
- [ ] Verify `nix build` produces working binary

### Phase 2: Nix Checks

- [ ] Add `checks` for test, lint, and build
- [ ] Verify `nix flake check` passes

### Phase 3: gogenfilter Integration

- [ ] Add gogenfilter as flake input
- [ ] Handle the replace directive in buildGoModule
- [ ] Verify build works with both local and remote gogenfilter

### Phase 4: CI Integration

- [ ] Add Nix CI workflow (hybrid mode initially)
- [ ] (Optional) Set up Cachix for binary cache
- [ ] (Optional) Migrate existing workflows to use `nix develop`

### Phase 5: Polish

- [ ] Add cross-compilation packages
- [ ] Add version injection from git
- [ ] Add formatter for `.nix` files
- [ ] Update AGENTS.md with Nix workflow documentation
- [ ] Update README with Nix onboarding instructions

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| gogenfilter replace directive breaks `nix build` | High | High | Flake input or publish module |
| Go 1.26 not yet in nixpkgs stable | Medium | Medium | Use nixos-unstable or overlay |
| CI slowdown from Nix installation | Low | Low | Cachix cache, hybrid CI approach |
| Team members unfamiliar with Nix | Medium | Medium | direnv makes it transparent; fallback to manual installs |
| flake.lock grows stale | Low | Low | Automated renovate/dependabot for flake.lock |
| Build sandbox blocks local paths | High | Medium | Flake inputs instead of local paths |

---

## 9. Alternatives Considered

### 9.1 No Nix (Status Quo)

Continue with manual tool installation. **Rejected** — recurring cache corruption and version mismatch issues prove this is unreliable.

### 9.2 Docker-based Development

Use a Dockerfile with all tools pre-installed. **Rejected** — heavier weight, slower startup, no direnv integration, cross-platform issues on macOS.

### 9.3 devbox (Jetpack)

Uses Nix under the hood but with simpler `devbox.json` config. **Viable alternative** but less flexible than raw flakes for the buildGoModule packaging use case.

### 9.4 asdf / mise

Version manager approach with `.tool-versions`. **Partial solution** — only handles tool versions, not reproducible builds or packaging.

---

## 10. Recommended First Step

```bash
# 1. Create flake.nix and .envrc (see Phase 1 above)
# 2. Enter the dev shell
nix develop

# 3. Verify everything works
just ci

# 4. If it works, commit
git add flake.nix flake.lock .envrc .gitignore
```

The entire Phase 1 can be validated in under an hour. If the devShell works correctly with `just ci`, the migration is already providing value — even without `nix build` or Nix CI.
