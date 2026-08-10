# Roadmap

**No timeline.** Aspirational items for future consideration. Items graduate to `TODO_LIST.md` when they become actionable.

---

## Language Support

- [ ] **TypeScript/JavaScript support**: Would require a TS AST parser (Babel, swc, or tree-sitter). The detection pipeline (suffix tree, hash, actionability) is language-agnostic; only the AST-to-Node transformer needs per-language implementation.
- [ ] **Python support**: Same architecture as TypeScript: a Python AST-to-Node transformer. The `syntax/` package is designed for multi-language extension.
- [ ] **Cross-language type matching (Go and Templ)**: Investigate whether type information from Go can improve Templ clone detection accuracy.

## IDE and Tooling

- [ ] **LSP server mode**: Run art-dupl as a Language Server, surfacing clones as diagnostics in VS Code / GoLand. Would enable real-time clone detection during development.
- [ ] **Watch mode**: Continuous monitoring with incremental detection. Re-run only on changed files. Would pair with `--incremental` caching.
- [ ] **Web UI dashboard**: Compile to WASM for in-browser clone detection with an interactive report interface.

## Architecture

- [ ] **Plugin architecture for detection methods**: Allow third-party detectors beyond suffix-tree and hash. The `detection.MethodDetector` interface supports this but is not documented as a public extension point.
- [x] **Parallel suffix tree search**: Ukkonen's construction is inherently sequential (active-point state propagation), so true parallel construction is impossible without a fundamentally different algorithm. Instead, `FindDuplOverParallel` parallelizes the DFS search phase across root-level subtrees, giving 1.5-3.4x speedup depending on tree size. The `--search-workers` CLI flag and `Options.SearchWorkers` SDK field control worker count (0=sequential, >1=parallel).
- [x] **Memory-compact suffix tree storage**: Changed `STree.data` from `[]Token` (16 bytes/interface) to `[]TokenValue` (4 bytes/int32), reducing pointer-array memory by 75%. The tree no longer retains references to original Token objects; the external `[]*syntax.Node` in `BuildTree` is the sole copy, needed for `FindSyntaxUnits` position-indexed lookups.
- [ ] **Suffix Array + LCP as alternative to Ukkonen's**: Implement SA-IS construction + Kasai LCP array + maximal-repeat scan as a new `MethodDetector`. Scientifically superior for this use case: 30-44% less memory (contiguous int arrays vs pointer-heavy tree), 2-5x faster search (cache locality), parallelizable construction (impossible with Ukkonen's), and simpler code (~150 lines vs ~250). Requires alphabet remapping pass (TokenValue int32 range -> compact [0,sigma)). Migration is low-risk via `MethodDetector` interface — both detectors can coexist for A/B benchmarking. See ADR-0020 and `docs/research/algorithmic-alternatives-for-clone-detection.md`.
- [ ] **Per-file position offset map (eliminate `[]*syntax.Node`)**: The `[]*syntax.Node` in `BuildTree` (8N bytes) is the dominant memory cost — larger than the tree structure itself. Replace with `(fileOffset, localPosition)` pairs so each file's nodes can be GC'd after search. Would reduce the irreducible memory floor from 8N to ~2N bytes. Requires restructuring `FindSyntaxUnits` output to carry position metadata instead of direct `*Node` pointers.
- [ ] **Winnowing pre-filter for extreme scale**: For 100K+ file codebases, add a two-phase pipeline: Phase 1 uses Winnowing (Schleimer et al. 2003, O(n) with sparse O(n/(w+1)) memory) to identify clone candidate regions. Phase 2 indexes only candidates with the suffix tree/array. Detection guarantee for clones >= t = w + k - 1 tokens. Industry standard (MOSS uses this). Would extend scalability to SourcererCC-class workloads.

## Quality and Intelligence

- [~] **Interface-aware suppression**: Detect method signatures that implement an interface contract and suppress them as structural duplication, not actionable cloning. Same-package interface detection via `go/types` is implemented (the `interface-method` pattern checks the `InterfaceMethod` flag set by the transformer when `--type-aware` is active). Cross-package and stdlib interfaces still rely on the static name list (`commonInterfaceMethodNames`). Full call-graph analysis remains future work.
- [x] **Configurable actionability patterns**: `--disable-pattern <label>` and `--list-patterns` implemented. 23 patterns currently active; property engine adds second-pass analysis.
- [ ] **ML-based actionability classification**: Train a model on labeled clone data to predict whether a clone is actionable, replacing the rule-based actionability patterns. Would handle edge cases the 22 current patterns miss.
- [~] **Fixability score**: Property-based extractability engine implemented (ADR-0017) with 4 properties + confidence scoring. Three-tier output (actionable / low-confidence / non-actionable). Confidence values need calibration against real-world data.
- [ ] **Nested-scope shadowing in alpha-normalization**: Current symbol table is flat (no nested-scope shadowing). Proper lexical scoping would improve Type-2 clone accuracy in deeply nested code.
- [ ] **Type narrowing for interface-typed variables**: If a local has an interface type, two variables with the same interface type match even if their concrete types differ. Could add concrete-type awareness via flow analysis.

## Performance

- [ ] **Incremental type checking**: Only re-check changed packages in type-aware mode. Currently type-aware mode loads all files for `go/packages`.
- [ ] **Caching for type-checking results**: Similar to the AST cache, cache `go/types` results keyed by content hash and detection mode.
- [x] **Type-aware + incremental integration**: `--type-aware` combined with `--incremental` now works. The `IncrementalParser` threads `typeInfos` via `SetTypeAwareData()`.

## UX and Defaults

- [ ] **Type-aware + suggest-generics as default (with `--fast` escape hatch)**: Long-term goal: make the most informative analysis the default, with a `--fast` flag that disables type-checking for quick syntactic scans. Revisit AFTER: (1) precision filtering brings `--suggest-generics` from 12.5% to >50% (surface 3 real candidates on DiscordSync, not 24), (2) progress output during the ~100x slower type-checking phase (57s dead-air on DiscordSync), (3) real-world validation on 3-5 codebases confirms signal-to-noise ratio. See `docs/status/2026-08-10_03-37_suggest-generics-e2e-validation-discordsync.md` for the E2E measurements and `docs/feedback/new/2026-08-10_discordsync_type-aware-false-negatives-generics-extraction-candidates.md` for the original analysis.

## Documentation and Adoption

- [ ] **SARIF output validation**: Validate emitted SARIF JSON against GitHub's official schema validator library in CI.
- [ ] **Performance optimization guide**: Document `--workers`, `--search-workers`, `--incremental`, `--cache-dir` tuning for different codebase sizes.
- [ ] **awesome-go submission**: Submit to awesome-go list once stable v1.0 is tagged.

---

_These items are aspirational and have no committed timeline. They represent potential future directions based on user needs, feedback sessions, and project evolution._
