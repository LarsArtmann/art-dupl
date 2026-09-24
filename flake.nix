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
      url = "github:LarsArtmann/gogenfilter?rev=7183352045a350140ae848bb7e84cd8a64e9a23a";
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
      version = "0.7.0";

      gogenfilterGoMod = builtins.readFile "${gogenfilter}/go.mod";
      gogenfilterGoSum = builtins.readFile "${gogenfilter}/go.sum";

      mkPackage =
        pkgs:
        let
          # buildGoModule must run the SAME go the module requires
          # (go.mod pins go 1.27.1; see .buildflow.yml for why the pin is
          # deliberate); nixpkgs' default go lags behind.
          buildGoModule = pkgs.buildGoModule.override { go = pkgs.go_1_27; };
        in
        buildGoModule {
          pname = "art-dupl";
          inherit version;

          nativeBuildInputs = [ pkgs.templ ];

          src = lib.cleanSource ./.;

          vendorHash = "sha256-Efi0Wxc2fKRRxIAPlST0M39pXco1pxZNGmvhc901hcw=";
          proxyVendor = true;

          overrideModAttrs = _old: {
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
          goPkg = pkgs.go_1_27;

          # goimports and templ shell out to `go` for module/import
          # resolution. The treefmt sandbox has no network, so a go that is
          # older than go.mod's requirement cannot download the toolchain and
          # every format run dies. Wrap them so they see the same go the
          # module requires, with GOTOOLCHAIN=local (nothing to download).
          withGo127 =
            pkg: bin:
            pkgs.symlinkJoin {
              name = "${bin}-go127";
              paths = [ pkg ];
              nativeBuildInputs = [ pkgs.makeWrapper ];
              postBuild = ''
                wrapProgram "$out/bin/${bin}" \
                  --prefix PATH : ${goPkg}/bin \
                  --set GOTOOLCHAIN local
              '';
            };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            # Generated code is machine output and must not be formatted:
            # gofumpt rewrites templ's `var x = ...` bindings inside
            # report_templ.go, making the treefmt check permanently red after
            # every `templ generate` (mirrors .golangci.yml's _templ\.go$
            # formatter exclusion).
            settings.excludes = [ "*_templ.go" ];
            programs = {
              gofumpt.enable = true;
              goimports.package = withGo127 pkgs.goimports "goimports";
              templ.package = withGo127 pkgs.templ "templ";
              nixfmt.enable = true;
            };
          };

          checks = {
            format = config.treefmt.build.check self;
            build = config.packages.default;

            # Fast vendorHash drift check: forces realization of the goModules
            # FOD. If vendorHash doesn't match go.sum, the FOD fails with a
            # clear hash mismatch error — before any Go code compiles.
            vendor-hash = pkgs.runCommand "vendor-hash" { } ''
              echo "vendor hash verified: ${config.packages.default.goModules}"
              touch $out
            '';

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

            # arch-lint (TODO #40, 2026-09-23): the 2026-09-18 arch-lint break
            # only surfaced in CI. go/packages needs a Go toolchain + writable
            # module/build cache at runtime, which the sandbox DOES provide via
            # buildGoModule's env (same pattern as the `lint` check above) —
            # checkPhase runs after the default build, so stdlib export data is
            # warm. CI still runs it too (workflow pinning go1.27.1 explicitly).
            arch-lint = config.packages.default.overrideAttrs (old: {
              name = "${old.pname}-arch-lint";
              nativeBuildInputs = old.nativeBuildInputs ++ [ pkgs.go-arch-lint ];
              GOCACHE = "/tmp/go-arch-build";
              doCheck = true;
              checkPhase = ''
                runHook preCheck
                go-arch-lint check
                runHook postCheck
              '';
              installPhase = ''
                touch $out
              '';
            });

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
                  # Gate = the tool's DEFAULT-threshold verdict on its own
                  # source. (Threshold 1 is a diagnostic level: since
                  # ADR-0023's nested-statement emission, ~46 genuine small
                  # pairs surface there; tracking their cleanup lives in
                  # TODO_LIST. The default-threshold promise is what users
                  # experience and what this gate enforces.)
                  output=$(art-dupl -t 5 --plumbing . 2>/dev/null) || true
                  if [ -n "$output" ]; then
                    echo "FAIL: art-dupl detected duplication in its own source at the default threshold 5:" >&2
                    echo "$output" >&2
                    exit 1
                  fi
                  echo "OK: art-dupl self-scan emits 0 lines at the default threshold 5"
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

            # Allocation-regression gate: guarded suffixtree benchmarks must
            # stay within scripts/alloc-budgets.txt (allocs/op, ±1 tolerance).
            # Allocation counts are deterministic; wall time is not.
            alloc-gate = config.packages.default.overrideAttrs (old: {
              name = "${old.pname}-alloc-gate";
              doCheck = true;
              nativeBuildInputs = old.nativeBuildInputs ++ [ pkgs.gawk ];
              checkPhase = ''
                runHook preCheck
                bash scripts/check-alloc-regression.sh
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
