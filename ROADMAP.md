# Roadmap

**No timeline.** Aspirational items for future consideration. Items graduate to `TODO_LIST.md` when they become actionable.

---

## Language Support

- **TypeScript/JavaScript support**: Would require a TS AST parser (Babel, swc, or tree-sitter). The detection pipeline (suffix tree, hash, actionability) is language-agnostic; only the AST-to-Node transformer needs per-language implementation.
- **Python support**: Same architecture as TypeScript: a Python AST-to-Node transformer. The `syntax/` package is designed for multi-language extension.
- **Cross-language type matching (Go and Templ)**: Investigate whether type information from Go can improve Templ clone detection accuracy.

## IDE and Tooling

- **LSP server mode**: Run art-dupl as a Language Server, surfacing clones as diagnostics in VS Code / GoLand. Would enable real-time clone detection during development.
- **Watch mode**: Continuous monitoring with incremental detection. Re-run only on changed files. Would pair with `--incremental` caching.
- **Web UI dashboard**: Compile to WASM for in-browser clone detection with an interactive report interface.

## Architecture

- **Plugin architecture for detection methods**: Allow third-party detectors beyond suffix-tree and hash. The `detection.MethodDetector` interface supports this but is not documented as a public extension point.
- **Suffix Array + LCP as alternative to Ukkonen's**: Implement SA-IS construction + Kasai LCP array + maximal-repeat scan as a new `MethodDetector`. Scientifically superior for this use case: 30-44% less memory (contiguous int arrays vs pointer-heavy tree), 2-5x faster search (cache locality), parallelizable construction (impossible with Ukkonen's), and simpler code (~150 lines vs ~250). Requires alphabet remapping pass (TokenValue int32 range -> compact [0,sigma)). Migration is low-risk via `MethodDetector` interface — both detectors can coexist for A/B benchmarking. See ADR-0020 and `docs/research/algorithmic-alternatives-for-clone-detection.md`.
- **Per-file position offset map (eliminate `[]*syntax.Node`)**: The `[]*syntax.Node` in `BuildTree` (8N bytes) is the dominant memory cost — larger than the tree structure itself. Replace with `(fileOffset, localPosition)` pairs so each file's nodes can be GC'd after search. Would reduce the irreducible memory floor from 8N to ~2N bytes. Requires restructuring `FindSyntaxUnits` output to carry position metadata instead of direct `*Node` pointers.
- **Winnowing pre-filter for extreme scale**: For 100K+ file codebases, add a two-phase pipeline: Phase 1 uses Winnowing (Schleicher et al. 2003, O(n) with sparse O(n/(w+1)) memory) to identify clone candidate regions. Phase 2 indexes only candidates with the suffix tree/array. Detection guarantee for clones >= t = w + k - 1 tokens. Industry standard (MOSS uses this). Would extend scalability to SourcererCC-class workloads.
- **Type-3 structural clone detection via CFG matching**: Detect near-miss clones where statements are reordered, interleaved with non-duplicated code, or have small structural differences. Current suffix-tree matching requires exact token sequences. Would use control-flow graph fingerprinting or call-sequence matching. Highest effort, highest value improvement. Identified by art-dupl-threshold-cliff and licenseforge feedback.

## Quality and Intelligence

- **Interface-aware suppression (cross-package)**: Same-package interface detection via `go/types` is implemented (the `interface-method` pattern). Cross-package and stdlib interfaces still rely on the static name list (`commonInterfaceMethodNames`). Full call-graph analysis remains future work.
- **Configurable actionability patterns**: `--disable-pattern <label>` and `--list-patterns` implemented. 29 denylist patterns + 4 property-engine labels currently active.
- **ML-based actionability classification**: Train a model on labeled clone data to predict whether a clone is actionable, replacing the rule-based actionability patterns. Would handle edge cases the 29 current patterns miss.
- **Fixability score**: Property-based extractability engine implemented (ADR-0017) with 4 properties + confidence scoring. Three-tier output (actionable / low-confidence / non-actionable). Confidence values need calibration against real-world data. Could evolve into a continuous fixability score instead of binary Actionable/NonActionable.
- **Nested-scope shadowing in alpha-normalization**: Current symbol table is flat (no nested-scope shadowing). Proper lexical scoping would improve Type-2 clone accuracy in deeply nested code.
- **Type narrowing for interface-typed variables**: If a local has an interface type, two variables with the same interface type match even if their concrete types differ. Could add concrete-type awareness via flow analysis.
- **Irreducibility class per group**: Classify each group by WHY it can't be reduced: semantic (real duplication), idiomatic (language convention), structurally-bound (framework constraint), or threshold-noise. Would give users actionable context instead of a binary verdict. (samber-do-auditlog feedback)
- **Helper-call-site detection via call-graph**: Detect when duplication is already an invocation of a shared helper (the extraction IS the helper call, not more duplication). Requires call-graph resolution. (upd, discordsync feedback)

## Performance

- **Incremental type checking**: Only re-check changed packages in type-aware mode. Currently type-aware mode loads all files for `go/packages`.
- **Caching for type-checking results**: Similar to the AST cache, cache `go/types` results keyed by content hash and detection mode.
- **In-memory LRU cache layer**: Add in-process cache on top of `FileCache` to avoid redundant gob deserialization on hot paths. Biggest perf win for cache-heavy workflows.

## UX and Defaults

- **Threshold cliff mitigation**: Binary on/off between t=2 and t=3 causes abrupt jumps in clone count. Approaches: fractional thresholds (token-based), `--explain-threshold` mode showing what each value produces, or recommending a range. (art-dupl-threshold-cliff, licenseforge feedback)
- **`.art-duplignore` config file support**: Project-level ignore patterns for clone detection, parallel to `.gitignore`. (false-positive-report feedback)
- **Test-file-aware thresholds**: Different default thresholds for production vs test vs testdata code. Test boilerplate (`t.Parallel()`, `if got != want`) is structurally duplicated by design. (licenseforge feedback)
- **`--suggest-extraction` dry-run mode**: Show the proposed helper signature for a clone group without actually extracting. (go-etag feedback)
- **`--ci-gate` mode**: Exit non-zero only on actionable clones (not non-actionable boilerplate). Stricter CI gating than current `check` mode. (go-etag feedback)
- **`--diff-baseline` mode**: Show what changed since the last baseline scan — new, suppressed, and resolved clone groups. (go-cqrs-lite feedback)
- **Type-aware + suggest-generics as default (with `--fast` escape hatch)**: Long-term goal: make the most informative analysis the default. Revisit AFTER: (1) precision filtering brings `--suggest-generics` from 12.5% to >50%, (2) progress output during the ~100x slower type-checking phase, (3) real-world validation on 3-5 codebases confirms signal-to-noise ratio.

## Documentation and Adoption

- **SARIF output validation**: Validate emitted SARIF JSON against GitHub's official schema validator library in CI.
- **Performance optimization guide**: Document `--workers`, `--search-workers`, `--incremental`, `--cache-dir` tuning for different codebase sizes.
- **awesome-go submission**: Submit to awesome-go list once stable v1.0 is tagged.

---

_These items are aspirational and have no committed timeline. They represent potential future directions based on user needs, feedback sessions, and project evolution._
