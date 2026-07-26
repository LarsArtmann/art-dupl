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
      url = "github:LarsArtmann/gogenfilter?rev=6c1baabdaf12709d3ea45779e93b1a4a8f3cb898";
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
      inherit (nixpkgs) lib;

      # Bump this for each release (see RELEASE.md step 3).
      # go build without ldflags still reports "dev"; nix build injects this.
      version = "0.5.1";

      gogenfilterGoMod = builtins.readFile "${gogenfilter}/go.mod";
      gogenfilterGoSum = builtins.readFile "${gogenfilter}/go.sum";

      mkPackage =
        pkgs:
        let
          inherit (pkgs) buildGoModule;
        in
        buildGoModule {
          pname = "art-dupl";
          inherit version;

          nativeBuildInputs = [ pkgs.templ ];

          src = lib.cleanSource ./.;

          vendorHash = "sha256-EkOKxXPBn+BGFXKhrILOh3MpwnnIn31qrTKg3iu5fCI=";
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

          env = {
            CGO_ENABLED = 0;
            GOEXPERIMENT = "jsonv2";
          };

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
            maintainers = [
              {
                name = "Lars Artmann";
                github = "LarsArtmann";
              }
            ];
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

            race = config.packages.default.overrideAttrs (old: {
              name = "${old.pname}-race";
              doCheck = true;
              checkPhase = ''
                runHook preCheck
                CGO_ENABLED=1 go test -race ./...
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

            disabled-linters = pkgs.runCommand "art-dupl-disabled-linters" { } ''
              cd ${pkgs.lib.cleanSource ./.}
              bash scripts/check-disabled-linters.sh .golangci.yml
              touch $out
            '';

            self-test =
              pkgs.runCommand "art-dupl-self-test"
                {
                  nativeBuildInputs = [
                    config.packages.default
                    pkgs.templ
                    goPkg
                  ];
                  GOEXPERIMENT = "jsonv2";
                }
                ''
                  cp -r --no-preserve=mode ${pkgs.lib.cleanSource ./.}/* .
                  templ generate
                  output=$(art-dupl -t 1 --plumbing . 2>/dev/null) || true
                  if [ -n "$output" ]; then
                    echo "FAIL: art-dupl detected duplication in its own source at threshold 1:" >&2
                    echo "$output" >&2
                    exit 1
                  fi
                  echo "OK: art-dupl self-scan emits 0 lines at threshold 1"
                  touch $out
                '';

            sarif-validate = config.packages.default.overrideAttrs (old: {
              name = "${old.pname}-sarif-validate";
              doCheck = true;
              checkPhase = ''
                runHook preCheck
                go test -run='TestSARIF' -v ./printer/...
                runHook postCheck
              '';
              installPhase = ''
                touch $out
              '';
            });

            bench = config.packages.default.overrideAttrs (old: {
              name = "${old.pname}-bench";
              doCheck = true;
              checkPhase = ''
                runHook preCheck
                go test -run='^TestPerfRegression' -v ./syntax/...
                runHook postCheck
              '';
              installPhase = ''
                touch $out
              '';
            });
          };

          packages = {
            default = mkPackage pkgs;
            art-dupl = mkPackage pkgs;
          };

          apps = {
            default = {
              type = "app";
              program = "${self.packages.${pkgs.stdenv.system}.default}/bin/art-dupl";
              meta.description = "Run art-dupl code clone detection";
            };
            art-dupl = {
              type = "app";
              program = "${self.packages.${pkgs.stdenv.system}.default}/bin/art-dupl";
              meta.description = "Run art-dupl code clone detection";
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
                GOEXPERIMENT = "jsonv2";
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
              GOEXPERIMENT = "jsonv2";
            };
          };
        };

      flake.overlays.default = final: _prev: {
        art-dupl = mkPackage final;
      };
    };
}
