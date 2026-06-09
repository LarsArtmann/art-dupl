{
  description = "art-dupl — Fast, type-safe code duplication detector for Go projects";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    gogenfilter = {
      url = "git+ssh://git@github.com/LarsArtmann/gogenfilter?rev=8788d6c7732b51abadc6cc97d5b5e6f91243b040";
      flake = false;
    };
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      systems,
      treefmt-nix,
      gogenfilter,
    }:
    let
      lib = nixpkgs.lib;

      version = "0.2.0";

      gogenfilterGoMod = builtins.readFile "${gogenfilter}/go.mod";
      gogenfilterGoSum = builtins.readFile "${gogenfilter}/go.sum";

      mkPackage =
        pkgs:
        let
          buildGoModule = pkgs.buildGoModule;
        in
        buildGoModule {
          pname = "art-dupl";
          inherit version;

          nativeBuildInputs = [ pkgs.templ ];

          src = lib.cleanSource ./.;

          vendorHash = "sha256-p8mldrn+sJYbpswh29zdEfxsqdBunwOmhWX+vTPZh1U=";
          proxyVendor = true;

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
            mkdir -p gogenfilter-real
            cp -r ${gogenfilter}/. gogenfilter-real/
            go mod edit -replace=github.com/LarsArtmann/gogenfilter/v3=./gogenfilter-real
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

          doCheck = false;

          meta = with lib; {
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
            maintainers = [ lib.maintainers.larsartmann ];
            platforms = platforms.all;
          };
        };
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          ...
        }:
        let
          goPkg = pkgs.go_1_26;
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              templ.enable = true;
              nixfmt.enable = true;
            };
          };

          checks = {
            format = config.treefmt.build.check self;
            build = config.packages.default;

            test = config.packages.default.overrideAttrs (old: {
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

            lint = config.packages.default.overrideAttrs (old: {
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

            fmt = pkgs.runCommand "art-dupl-fmt" { nativeBuildInputs = [ goPkg ]; } ''
              cd ${pkgs.lib.cleanSource ./.}
              test -z "$(gofmt -l .)" || (echo "Unformatted files:"; gofmt -l .; exit 1)
              touch $out
            '';
          };

          packages = {
            default = mkPackage pkgs;
            art-dupl = mkPackage pkgs;
          };

          apps = {
            default = {
              type = "app";
              program = "${self.packages.${pkgs.stdenv.system}.default}/bin/art-dupl";
            };
            art-dupl = {
              type = "app";
              program = "${self.packages.${pkgs.stdenv.system}.default}/bin/art-dupl";
            };
          };

          devShells = {
            default = pkgs.mkShell {
              name = "art-dupl-dev";

              packages = with pkgs; [
                goPkg
                golangci-lint
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
                echo ""
              '';
            };

            ci = pkgs.mkShellNoCC {
              packages = [
                goPkg
                pkgs.golangci-lint
                pkgs.templ
              ];

              GOWORK = "off";
              GOPRIVATE = "github.com/LarsArtmann/*";
            };
          };
        };

      flake.overlays.default = final: _prev: {
        art-dupl = mkPackage final;
      };
    };
}
