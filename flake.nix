{
  description = "art-dupl — Fast, type-safe code duplication detector for Go projects";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    gogenfilter = {
      url = "git+ssh://git@github.com/LarsArtmann/gogenfilter?rev=8788d6c7732b51abadc6cc97d5b5e6f91243b040";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      gogenfilter,
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      # gogenfilter's go.mod and go.sum are read at Nix evaluation time so the
      # dummy is always in sync — no manual maintenance when deps change.
      gogenfilterGoMod = builtins.readFile "${gogenfilter}/go.mod";
      gogenfilterGoSum = builtins.readFile "${gogenfilter}/go.sum";

      mkPackage =
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          goPkg = pkgs.go_1_26 or pkgs.go;
          version = "0.2.0";
        in
        pkgs.buildGoModule {
          pname = "art-dupl";
          inherit version;

          go = goPkg;

          nativeBuildInputs = [ pkgs.templ ];

          src = pkgs.lib.cleanSource ./.;

          # Private Go modules can't be fetched inside the Nix sandbox (no SSH).
          # Strategy: gogenfilter is pre-fetched as a flake input (via SSH during
          # evaluation). The goModules derivation uses a dummy local replace so it
          # can vendor all public deps without network access to the private repo.
          # The main build then swaps in the real gogenfilter from the flake input.
          vendorHash = "sha256-0T74NtbVIYkyOdPvMDrSUvioDOyGVV9XjxT4jtAczB0=";

          overrideModAttrs = old: {
            preBuild = ''
              mkdir -p dummy
              cat > dummy/go.mod << 'DUMMYEOF'
              ${gogenfilterGoMod}
              DUMMYEOF
              cat > dummy/go.sum << 'DUMMYEOF'
              ${gogenfilterGoSum}
              DUMMYEOF
              echo 'package gogenfilter
              import (
                _ "github.com/bmatcuk/doublestar/v4"
                _ "github.com/go-faster/yaml"
              )' > dummy/dummy.go
              go mod edit -replace=github.com/LarsArtmann/gogenfilter/v3=./dummy
              go mod tidy
            '';
          };

          preBuild = ''
            templ generate
            chmod -R u+w vendor
            rm -rf vendor/github.com/LarsArtmann/gogenfilter/v3
            mkdir -p vendor/github.com/LarsArtmann/gogenfilter/v3
            cp -r ${gogenfilter}/. vendor/github.com/LarsArtmann/gogenfilter/v3/
            sed -i 's|=> ./dummy|=> ./vendor/github.com/LarsArtmann/gogenfilter/v3|' vendor/modules.txt
            go mod edit -replace=github.com/LarsArtmann/gogenfilter/v3=./vendor/github.com/LarsArtmann/gogenfilter/v3
          '';

          ldflags = [
            "-s"
            "-w"
            "-X github.com/LarsArtmann/art-dupl/cmd.Version=${version}"
            "-X github.com/LarsArtmann/art-dupl/cmd.Commit=${self.rev or "dirty"}"
            "-X github.com/LarsArtmann/art-dupl/cmd.Date=unknown"
          ];

          env.CGO_ENABLED = 0;

          subPackages = [ "cmd/art-dupl" ];

          # Tests are run via checks, not during build
          doCheck = false;

          meta = with pkgs.lib; {
            description = "Fast, type-safe code duplication detector for Go projects";
            longDescription = ''
              art-dupl is a modern code duplication detection tool for Go source files.
              It analyzes abstract syntax trees (ASTs) to find structural code clones
              while ignoring literal values. Supports multiple detection algorithms,
              professional CLI with auto-completion, and comprehensive output formats
              including text, HTML, JSON, and plumbing.
            '';
            homepage = "https://github.com/LarsArtmann/art-dupl";
            license = licenses.mit;
            mainProgram = "art-dupl";
            maintainers = [ ];
            platforms = platforms.all;
          };
        };
    in
    {
      # Packages: the main binary
      packages = forAllSystems (system: {
        default = mkPackage system;
        art-dupl = mkPackage system;
      });

      # Apps: run directly with `nix run github:LarsArtmann/art-dupl`
      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/art-dupl";
        };
        art-dupl = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/art-dupl";
        };
      });

      # Dev shell: full development environment
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          goPkg = pkgs.go_1_26 or pkgs.go;
        in
        {
          default = pkgs.mkShell {
            name = "art-dupl-dev";

            packages = with pkgs; [
              goPkg
              golangci-lint
              just
              bc
              git
              gopls
              templ
            ];

            env = {
              CGO_ENABLED = 0;
              GOTOOLCHAIN = "local";
              GOWORK = "off";
              GOPRIVATE = "github.com/LarsArtmann/*";
            };

            shellHook = ''
              echo ""
              echo "art-dupl development shell"
              echo "=========================="
              echo "Go:            $(go version)"
              echo "golangci-lint: $(golangci-lint --version | head -1)"
              echo "just:          $(just --version)"
              echo ""
              echo "Run 'just --list' to see available recipes"
              echo ""
            '';
          };
        }
      );

      # Checks: run with `nix flake check`
      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          goPkg = pkgs.go_1_26 or pkgs.go;
          pkg = self.packages.${system}.default;
        in
        {
          # Build check: verifies the package compiles
          build = pkg;

          # Test check: runs the test suite
          test = pkg.overrideAttrs (old: {
            name = "${old.pname}-test";
            doCheck = true;
            checkPhase = ''
              runHook preCheck
              go test ./...
              runHook postCheck
            '';
            installPhase = ''
              touch $out
            '';
          });

          # Lint check: runs golangci-lint
          lint = pkg.overrideAttrs (old: {
            name = "${old.pname}-lint";
            nativeBuildInputs = old.nativeBuildInputs ++ [ pkgs.golangci-lint ];
            GOCACHE = "/tmp/go-build";
            GOLANGCI_LINT_CACHE = "/tmp/golangci-lint-cache";
            buildPhase = ''
              runHook preBuild
              golangci-lint run --timeout 5m ./...
              runHook postBuild
            '';
            installPhase = ''
              touch $out
            '';
          });

          # Format check: verifies Go code is formatted
          fmt = pkgs.runCommand "art-dupl-fmt" { nativeBuildInputs = [ goPkg ]; } ''
            cd ${pkgs.lib.cleanSource ./.}
            test -z "$(gofmt -l .)" || (echo "Unformatted files:"; gofmt -l .; exit 1)
            touch $out
          '';
        }
      );

      # Overlay: use with `pkgs.art-dupl` after applying overlay
      overlays.default = final: prev: {
        art-dupl = self.packages.${prev.stdenv.hostPlatform.system}.default;
      };

      # Formatter: `nix fmt` formats .nix files
      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt);
    };
}
