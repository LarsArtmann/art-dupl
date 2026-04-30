{
  description = "art-dupl — Fast, type-safe code duplication detector for Go projects";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    gogenfilter = {
      url = "git+ssh://git@github.com/LarsArtmann/gogenfilter?rev=5957230e34ed14cde1999d7a379ef71a6989e3a1";
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

      mkPackage =
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          goPkg = pkgs.go_1_26 or pkgs.go;
          version = self.shortRev or self.dirtyShortRev or "unknown";
        in
        pkgs.buildGoModule {
          pname = "art-dupl";
          inherit version;

          go = goPkg;

          src = pkgs.lib.cleanSource ./.;

          # Private Go modules can't be fetched inside the Nix sandbox (no SSH).
          # Strategy: gogenfilter is pre-fetched as a flake input (via SSH during
          # evaluation). The goModules derivation uses a dummy local replace so it
          # can vendor all public deps without network access to the private repo.
          # The main build then swaps in the real gogenfilter from the flake input.
          vendorHash = "sha256-+fDTxFD/4M4ba/NEQ54Ge3p1dcXerUmCaFFV9kk4Wpk=";

          overrideModAttrs = old: {
            preBuild = ''
              mkdir -p dummy
              cat > dummy/go.mod << 'DUMMYEOF'
              module github.com/LarsArtmann/gogenfilter
              go 1.26.0
              require (
                github.com/bmatcuk/doublestar/v4 v4.10.0
                github.com/go-faster/yaml v0.4.6
              )
              require (
                github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
                github.com/go-faster/errors v0.7.1 // indirect
                github.com/go-faster/jx v1.2.0 // indirect
                github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
                github.com/segmentio/asm v1.2.1 // indirect
                go.uber.org/multierr v1.11.0 // indirect
                golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90 // indirect
                golang.org/x/sys v0.43.0 // indirect
              )
              DUMMYEOF
              echo 'package gogenfilter
              import (
                _ "github.com/bmatcuk/doublestar/v4"
                _ "github.com/go-faster/yaml"
              )' > dummy/dummy.go
              go mod edit -replace=github.com/LarsArtmann/gogenfilter=./dummy
            '';
          };

          preBuild = ''
            chmod -R u+w vendor
            rm -rf vendor/github.com/LarsArtmann/gogenfilter
            mkdir -p vendor/github.com/LarsArtmann/gogenfilter
            cp -r ${gogenfilter}/. vendor/github.com/LarsArtmann/gogenfilter/
            sed -i 's|=> ./dummy|=> ./vendor/github.com/LarsArtmann/gogenfilter|' vendor/modules.txt
            go mod edit -replace=github.com/LarsArtmann/gogenfilter=./vendor/github.com/LarsArtmann/gogenfilter
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
            ];

            env.CGO_ENABLED = 0;

            shellHook = ''
              export GOTOOLCHAIN=local
              export GOWORK=off
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
        }
      );

      # Overlay: use with `pkgs.art-dupl` after applying overlay
      overlays.default = final: prev: {
        art-dupl = self.packages.${prev.system}.default;
      };

      # Formatter: `nix fmt` formats .nix files
      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt);
    };
}
