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
    { self, nixpkgs, gogenfilter }:
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

          vendorHash = "sha256-rvzXCdcx2EwxjBYicMkEYV6z/ExZlRQhH09ys4WB/Ms=";

          overrideModAttrs = old: {
            preBuild = ''
              mkdir -p dummy
              cat > dummy/go.mod << 'DUMMYEOF'
              module github.com/LarsArtmann/gogenfilter
              go 1.26
              DUMMYEOF
              echo 'package gogenfilter' > dummy/dummy.go
              go mod edit -replace=github.com/LarsArtmann/gogenfilter=./dummy
            '';
          };

          preBuild = ''
            cp -rL vendor vendor-tmp
            chmod -R u+w vendor-tmp
            rm -rf vendor
            mv vendor-tmp vendor
            rm -rf vendor/github.com/LarsArtmann/gogenfilter
            mkdir -p vendor/github.com/LarsArtmann/gogenfilter
            cp -r ${gogenfilter}/. vendor/github.com/LarsArtmann/gogenfilter/
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
