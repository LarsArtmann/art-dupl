# AGENTS.md - art-dupl

Go tool for finding code clones via suffix tree + hash-based detection on ASTs. Multi-method, multi-format output.

## What Stays Here

Enduring context that's hard to discover from code. For everything else, see the right file:

| Need                   | File                      |
| ---------------------- | ------------------------- |
| How to use the CLI     | `HOW_TO_USE.md`           |
| Testing practices      | `TESTING.md`              |
| Feature inventory      | `FEATURES.md`             |
| Open work              | `TODO_LIST.md`            |
| Long-term ideas        | `ROADMAP.md`              |
| SDK design             | `SDK_DESIGN.md`           |
| Architecture decisions | `docs/adr/`               |
| Domain language        | `docs/DOMAIN_LANGUAGE.md` |

## Build & Test

```bash
just generate      # run templ generate (required before build if .templ files changed)
just build          # → dist/art-dupl (includes templ generate)
just test           # tests with coverage
just check          # lint
just ci             # format + lint + test (includes templ generate)
nix flake check     # reproducible CI (includes templ generate in preBuild)
```

## Architecture

```
cmd/        CLI (root, stats, version) via Fang/Cobra
config/     Config management, typed enums, reflection-based merge
detection/  MultiDetector dispatches to MethodDetector implementations
suffixtree/ Core suffix tree algorithm on AST tokens
syntax/     AST handling (golang/ + templ/)
hash/       Rolling hash-based detection
job/        Orchestrates parse → serialize → build tree
printer/    Output formatting (text, HTML, JSON, plumbing, SARIF, stats)
domain/     Value objects (Filepath, LineNumber, CloneSeverity)
pkg/artdupl/ Public SDK (Detector interface)
```

## Critical Conventions

- **Idiomatic Go only** — no Result[T], Option[T], railway-oriented programming. Standard `(T, error)` returns.
- **Semantic matching is default** (`config.DefaultConfig.Semantic = true`). `--structural` disables it. Semantic is now _faster_ than structural (was 9x slower before optimization).
- **Only `.go` and `.templ` files** are processed by default. Vendor excluded by default (`--vendor` to include).
- **Generated code filtered by default** (sqlc, templ, protobuf, mockgen, stringer). Override with `--include-*` flags.
- **Suffix tree uses O(1) map-based transition lookup** (optimized from O(n) linear search).
- **Errors** use typed hierarchy from `errors/` package. Wrap with `duplerrors.Wrap*`. No panics for expected errors.
- **Config merging** is reflection-based — adding Config fields requires no merge code changes.
- **BDD tests** use Ginkgo/Gomega in `bdd/`. Helpers: `NewBDDTestSetupForGinkgo()`, `RunArtDupl()`, `CreateDuplicateFiles()`, `RunArtDuplOnDir()`, `RunArtDuplWithStdin()` — all in `internal/testutil/bdd.go`.

## Nix Flake — Private Dependency Pattern

`gogenfilter` is a private dep handled via two-phase dummy/replace in `flake.nix`:

1. `builtins.readFile "${gogenfilter}/go.mod"` reads real go.mod at eval time
2. `overrideModAttrs`: dummy dir with real go.mod/go.sum + `-replace` to `./dummy`
3. `preBuild`: swap dummy for real gogenfilter, patch `vendor/modules.txt`

**When gogenfilter changes:** update `rev=`, set `vendorHash=""`, run `nix build`, copy the correct hash.

## Known Limitations

- **Printer ↔ syntax.Node coupling**: `Printer.PrintClones(dups [][]*syntax.Node)` — all 6 printers depend on AST internals. Fix requires ProcessedClone DTO (111 test call sites). Tracked in TODO_LIST.md.
- **Three parallel Clone types**: `printer.clone`, `printer.CloneGroup`, `pkg/artdupl.Clone`, `domain.ProcessedClone` — consolidation blocked on Printer DTO change.
- **ConstantCSSProperty Pos=0,End=0**: upstream `a-h/templ` limitation (no Range field). Mitigated by inheriting parent CSSTemplate range.
