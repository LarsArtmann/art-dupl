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
templ generate     # run templ generate (required before build if .templ files changed)
go build ./...     # build all packages (requires GOEXPERIMENT=jsonv2 — see below)
go test ./...      # run all tests
golangci-lint run --timeout 5m ./...  # lint
nix flake check    # reproducible CI (includes templ generate in preBuild)
```

> **`GOEXPERIMENT=jsonv2` is required.** The project migrated to `encoding/json/v2`.
> The `flake.nix` devShell sets this automatically. For non-Nix workflows, export it
> manually: `export GOEXPERIMENT=jsonv2`. Convention: use `omitzero` (not `omitempty`)
> on custom `MarshalJSON` types, and `format:nano` for `time.Duration` fields.

## Architecture

```
cmd/        CLI (root, stats, baseline, check, version) via Fang/Cobra
config/     Config management, typed enums, reflection-based merge
detection/  MultiDetector dispatches to MethodDetector implementations
suffixtree/ Core suffix tree algorithm on AST tokens
syntax/     AST handling (golang/ + templ/)
hash/       Rolling hash-based detection
job/        Orchestrates parse → serialize → build tree (threads golang.DetectionMode)
printer/    Output formatting (text, HTML, JSON, plumbing, SARIF, stats)
domain/     Value objects (ProcessedClone, enums, Extractability, validation sentinels)
baseline/   Baseline recording + CI check file format (Load/Save/Add/Has)
errors/     6 typed error types, stack traces, JSON marshaling
cache/      File-based AST caching with SHA-256 content hashing + LRU eviction
pkg/artdupl/ Public SDK (Detector interface) — independent types, no config aliases
pkg/enum/   Shared enum helpers (MarshalJSON, UnmarshalJSON, Parse)
```

## Critical Conventions

- **Idiomatic Go only** — no Result[T], Option[T], railway-oriented programming. Standard `(T, error)` returns.
- **Three detection modes** (`config.DetectionMode`): `semantic` (default, alpha-normalized — detects Type 2 renamed clones), `exact` (verbatim name hashing — Type 1 copy-paste only), `structural` (AST shape only). Config stores a single `DetectionMode` enum (not bools). CLI flags map via `config_builder.go`; job package threads `golang.DetectionMode` (not bool). See ADR-0007.
- **Alpha-normalization** (`syntax/golang/normalizer.go`): per-function symbol table canonicalizes locals (params, receiver, **type parameters (generics)**, body vars, **closure/FuncLit params and locals**) to v0/v1/... before hashing. `o.Name` keeps original for clone-type classification; `o.Type` uses the canonical name. Selectors/field names/types are NOT canonicalized (API surface). The symbol table is flat (no nested-scope shadowing).
- **Literal normalization** (`normalizesLiterals()` on `DetectionMode`): Semantic mode normalizes `BasicLit` VALUES to their KIND (STRING, INT, FLOAT) before hashing. This enables Type-2 clone detection where only literal values differ (e.g., `errors.New("foo")` matches `errors.New("bar")`). Exact mode hashes literal values verbatim (Type-1 only). Package-level identifiers are NOT normalized, so error definitions with different names (`ErrFoo` vs `ErrBar`) do NOT collapse.
- **Semantic encoding layout**: `[24-bit identifier/operator hash][8-bit base AST node type]`. Both Exact and Semantic modes hash identifiers (`hashesIdentifiers()`); Structural ignores them. Consumers comparing node types must use `golang.DecodeBaseType(node.Type)` — never compare raw `node.Type` against `golang.*` constants.
- **Operator-token encoding policy** (`syntax/golang/transform.go`): statement/decl/operator tokens that change semantics MUST be encoded via `encodeSemanticType` so siblings don't collapse. Currently covered: `AssignStmt.Tok`, `IncDecStmt.Tok`, `BinaryExpr.Op`, `UnaryExpr.Op`, `BranchStmt.Tok` (break/continue/goto/fallthrough), `GenDecl.Tok` (var/const/type/import), `ChanType.Dir` (chan/<-chan/chan<-, via `chanDirString` since `ast.ChanDir` has no `String()`), `KeyValueExpr.Key` (field names in struct literals — `Point{X:1}` ≠ `Size{W:1}`). When adding a new AST case, encode its distinguishing token/operator the same way.
- **KeyValueExpr field name encoding** (`syntax/golang/transform.go`): In semantic/exact mode, struct field names in composite literals are encoded into the KeyValueExpr Type via `encodeSemanticType`. This prevents `Point{X:1}` from matching `Size{W:1}` — field names are API surface, not local variables. Alpha-normalization does NOT affect this encoding because it reads the raw AST `n.Key.(*ast.Ident).Name` directly, not the (potentially normalized) child Ident node.
- **Templ callee name encoding** (`syntax/templ/transform_components.go`): `extractCalleeName()` splits at first `(` to get the function/method name from Go template call expressions. The callee name is then encoded via `EncodeSemanticType` so `{{ComponentA()}}` ≠ `{{ComponentB()}}`. Applied to both `transformTemplElementExpression` and `transformCallTemplateExpression`. **Gotcha**: `TemplElementExpression` (wraps `{{ <tag> }}` HTML elements) is different from `CallTemplateExpression` (wraps `{{ GoExpr }}` Go expressions). When debugging templ false positives, check BOTH transformer functions — the AST node type hierarchy in templ is: `Template` > [`TemplElementExpression` | `CallTemplateExpression` | `CSSExpression` | etc.].
- **Clone type classification** (`printer/clone_processor.go`): Type 1 (exact, identical names), Type 2 (renamed, detected via alpha-normalization), Type 3 (near-miss). Walks full subtree via `collectNamesPreOrder` (direct Children walk) — NOT `syntax.Serialize` (which destructively mutates node Types).
- **Statement-level tokenization**: `serial()` fingerprints entire statement subtrees into single composite tokens (`Statement=true` flag on `BlockStmt` children, `GenDecl` spec children). The fingerprint is stored in `Node.Fingerprint` (NOT `Node.Type`), so `DecodeBaseType(Type)` returns the correct base AST type for actionability analysis. `Val()` returns `Fingerprint` for statement nodes, `Type` for non-statement nodes. This means threshold counts duplicated STATEMENTS, not arbitrary AST nodes. ValueSpec (var/const) and TypeSpec (type) declarations are each fingerprinted as single tokens.
- **Only `.go` and `.templ` files** are processed by default. Vendor excluded by default (`--vendor` to include).
- **Generated code filtered by default** (sqlc, templ, protobuf, mockgen, stringer, generic catch-all). Override with `--include-generated <category>` (`sqlc`, `templ`, `protobuf`, `mockgen`, `stringer`, `generic`, `all`). **FilterGeneric override**: `--include-generated sqlc` disables `FilterSQLC` but `FilterGeneric` (any "Code generated by") still catches the file. `shouldIncludeFile` in `cmd/util.go` checks the file's generation comment via `generatorIncludes.allowsFile` and overrides the generic filter for explicitly-included categories. Uses content-only detection (not gogenfilter's `Is*Generated` which is filename-gated for SQLC/templ/protobuf). The old per-generator `--include-sqlc`, `--include-templ`, etc. flags are deprecated hidden aliases that still work with a warning.
- **Suffix tree uses O(1) map-based transition lookup** (optimized from O(n) linear search).
- **Cache eviction**: `cache.Prune(maxEntries)` evicts oldest entries by modification time. `Config.MaxCacheEntries` (0 = unlimited) wires into `IncrementalParser`.
- **Errors** use typed hierarchy from `errors/` package. Wrap with `duplerrors.Wrap*`. No panics for expected errors. `DuplError` no longer captures `debug.Stack()` — it was never consumed and added runtime overhead on every error path.
- **Config merging** is reflection-based — adding Config fields requires no merge code changes.
- **BDD tests** use Ginkgo/Gomega in `bdd/`. Helpers: `NewBDDTestSetupForGinkgo()`, `RunArtDupl()`, `CreateDuplicateFiles()`, `RunArtDuplOnDir()`, `RunArtDuplWithStdin()` — all in `internal/testutil/bdd.go`.
- **SDK type independence**: `pkg/artdupl` has **ZERO imports** of `config/` and `errors/`. Uses `domain` types via aliases (`DetectionMethod = domain.DetectionMethod`, `ErrInvalidThreshold = domain.ErrInvalidThreshold`). Logger aliased to `pkg/logger.Logger`. The `detection` package uses `[]domain.DetectionMethod` (typed, not plain `[]string`). Enforced via `.go-arch-lint.yml`.
- **Shared types in domain**: `DetectionMethod`, `ErrInvalidThreshold`, `ErrThresholdTooLarge`, `FileReaderFunc`, and `Logger` are defined once in `domain`/`pkg/logger` and aliased by `config`, `pkg/artdupl`, `detection`, and `printer`. See ADR-0005. `config.Config.Timeout` is `time.Duration` (not `int` seconds). `Fragment` is `string` everywhere (not `[]byte`).
- **Context propagation**: All pipeline goroutines (detection, job/parse, job/buildtree, job/incremental, job/ParseIncrementalParallel, SDK streaming) use `select { case ch <- v: case <-ctx.Done(): return }` for every channel send. Never use the "check-then-send" pattern (`select { case <-ctx.Done(): ...; default: }` followed by a bare `ch <- v`) — it races. `collectResults` in `job/parse.go` and `collectIncrementalResults` in `job/incremental.go` accept `ctx` as first param. `sendCtx[T]` in `job/sendctx.go` centralizes the pattern. `BuildTree`'s `done` channel is buffered(1). Only `cmd/run_crawl.go` file feeders lack ctx (stdin/filepath.Walk are inherently blocking; process exits on cancellation).
- **Enum pattern**: Domain enums use `pkg/enum` shared helpers (`MarshalJSON`, `UnmarshalJSON`, `Parse`). Each enum has `IsValid()` and `String()`. ClonePriority has `Rank()` for ordinal comparison. Config enums (`SortCriteria`, `OutputFormat`, `DiffMode`) still live in `config/` but use the same helpers.
- **`Node.Clone()`**: Deep-copies a `syntax.Node` subtree. Used on the incremental cache-hit path (`job/incremental.go`) to prevent data races when multiple goroutines hit the same cached content hash.
- **`MethodDetector.Name()`**: Interface method returns a human-readable detection-method name. Replaces the former `detName` type switch — new detectors just implement `Name()`.
- **Lint config**: `makezero: always: true` means all `make([]T, n)` with `n > 0` are flagged. Use `make([]T, 0, n)` + `append`, or add `//nolint:makezero` for genuine index-fill patterns (binary buffers, DP matrices, `copy()` targets). `wrapcheck` ignores `pkg/enum` via `ignore-package-globs`.

## Nix Flake — Private Dependency Pattern

`gogenfilter` is a private dep handled via two-phase dummy/replace in `flake.nix`:

1. `builtins.readFile "${gogenfilter}/go.mod"` reads real go.mod at eval time
2. `overrideModAttrs`: dummy dir with real go.mod/go.sum + `-replace` to `./dummy`
3. `preBuild`: swap dummy for real gogenfilter, patch `vendor/modules.txt`

**When gogenfilter changes:** update `rev=`, set `vendorHash=""`, run `nix build`, copy the correct hash.

## Known Limitations

- **Printer ↔ syntax.Node coupling**: Printer interface uses `domain.ProcessedCloneGroup`, but `actionability.go` still imports `syntax.Node` directly for pattern evaluation. `clone_processor.go` is the bridge point. The `everySequenceMatch` helper was extracted to reduce duplication.
- **Parallel Clone types**: `printer.CloneGroup`, `printer.JSONClone`, `pkg/artdupl.Clone`, `pkg/artdupl.CloneGroup`, `domain.ProcessedClone(Group)`. Fragment unified to `string`. Field names aligned (`LineStart`/`LineEnd` canonical). `printer.CloneGroup.Files` renamed to `Clones`. `StartPos`/`EndPos` added to `domain.ProcessedClone`. Types remain separate for Printer/SDK DTO independence — see `docs/research/SPLIT-BRAIN.html` for the full analysis. **`domain.CloneRef`** value object (Filename, LineStart, LineEnd, Fragment + `LineCount()`) is embedded in `domain.ProcessedClone` and `pkg/artdupl.Clone` to eliminate field-name drift. Struct literals must use `CloneRef: CloneRef{...}` / `CloneRef: domain.CloneRef{...}` nested initialization.
- **ConstantCSSProperty Pos=0,End=0**: upstream `a-h/templ` limitation (no Range field). Mitigated by inheriting parent CSSTemplate range.
- **Templ has no semantic mode**: `syntax/templ/` matching is purely structural — no identifier/operator encoding.
- **RESOLVED: destructive serialization + Type-2 classification** (2026-07-01): `syntax.Serialize`→`serial()` now shallow-copies each node before writing `Type`/`Owns` (`syntax/syntax.go` clone at line ~99), so the original tree is never mutated and serialization is idempotent. `printer/clone_processor.go::classifyCloneType` walks `node.Children` directly via `collectNamesPreOrder` (line ~110) — it no longer calls `syntax.Serialize`, so the Type-1-vs-Type-2 (renamed) distinction is preserved for statement-rooted fragments. Verified by idempotency + non-mutation tests. See ADR-0006 and `docs/status/2026-07-01_01-34_post-execution-sprint.md`.
- **RESOLVED: incremental cache-miss aliasing + parallel parsing** (2026-07-01): `job/incremental.go::parseFile` now deep-clones (`deepCloneNodes`) before storing in cache, so the cache-miss and hit paths both return independent node slices. `singleflight.Group` (keyed by content hash) deduplicates concurrent parses of byte-identical files (vendored/generated copies). `ParseIncrementalParallel(ctx, fchan, workers)` adds a worker pool (`startIncrementalWorkers` + reused `feedFiles` + `collectIncrementalResults`); `cmd/run_analysis.go` dispatches to it when `cfg.Workers > 1`. Each caller of `group.Do` deep-clones the shared result and stamps its own `Filename` via `cloneWithFilename`, so no two files share mutated nodes. Covered by 9 `-race` tests in `job/incremental_parallel_test.go` (identical-content singleflight, mutation isolation, sequential parity, cancellation).
- **Default threshold is 5** (`config.DefaultThreshold`). Filters trivial single-statement duplicates while catching meaningful cloning. Users can lower to 3 for more sensitivity or raise to 10+ for noise reduction.
- **Actionability patterns** (`printer/actionability*.go`): Non-actionable patterns suppress clone groups that represent Go boilerplate rather than real duplication. Detected patterns: signature-only, interface-implementation, RAII-defer (single defer **and** Lock+Defer Unlock / RLock+Defer RUnlock 2-stmt sequence), error-propagation, error-wrapping, assertion-chain, cobra-boilerplate, test-data-pair, table-driven-test, test-scaffolding, data-dominated, describe-table, builder-callback, **assign+error-check** (2-stmt `err := ...; if err != nil { return ... }`), **single-call-expression** (lone `CallExpr` like `errors.New("foo")`). A group is NonActionable only when EVERY clone matches the same pattern. Actionability filtering runs in `cmd/run_output.go` only when `--semantic` is active.
- **Sort comparator factory**: `sortGroupsByCriteria[T]` + `GroupMetrics[T]` in `printer/sort_unified.go` provides a single 4-criteria switch (Size/Occurrence/Hash/TotalTokens) via generic extractors. All group-level sorts go through this factory — never duplicate the switch. `SortProcessedClonesByCriteria` handles within-group (individual clone) sorting separately.
- **No type-aware detection (go/types)**: art-dupl is purely syntax-based. It cannot distinguish `a.String()` (where `a` is `time.Time`) from `b.String()` (where `b` is `*big.Int`). Integrating `go/types` as an opt-in `--type-aware` mode is the highest-impact future improvement — it would eliminate the `same-method-name-different-receiver-type` class of false positives. Requires `golang.org/x/tools/go/packages` for type checking (10-100x slower than parsing alone). NOT YET IMPLEMENTED. Other researched libraries: `go/constant` (not needed — `Kind.String()` is equivalent for semantic mode), `ast/astutil` (not useful — no pattern matching API).
- **`filepath.Walk` context cancellation**: `handleWalkEntry` in `cmd/run_crawl.go` checks `ctx.Err()` at each entry and returns the context error to abort the walk. `crawlDirectoryWithOpts` suppresses `context.Canceled`/`context.DeadlineExceeded` from stderr.
