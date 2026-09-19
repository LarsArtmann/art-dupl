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

## Detection Granularity (ADR-0023 follow-ups)

Shipped 2026-09-14: nested-block statements (loop/if/switch/select bodies, else-if links) are emitted as individual tokens; subsumed units are trimmed at `FindSyntaxUnits`. Remaining granularity work, deliberately deferred:

- **FuncLit/closure-body emission**: statements inside closures (`t.Run(..., func(t){...})`, `defer func(){...}()`, `go func(){...}()`) are still fingerprinted as part of the enclosing statement — the same masking class ADR-0023 fixed for control-flow blocks. Deferred because test-heavy repos would see a large noise-profile change; needs its own corpus validation and possibly new actionability patterns. (source report #40/#44 context)
- **GenDecl spec-level sharing**: `var (...)` blocks with partially-shared specs match only as whole declarations; specs stay composite-only because a single-spec decl would double-count one source statement. (report #33)
- **LabeledStmt.Stmt interiors**: labeled statements are composite-only; interiors are masked (rare shape).
- **Cross-composite statement-head sharing**: identical statement heads appearing at the START of different composite statement types (e.g. same first statement in an if vs a for) — probe whether header tokens should participate. (report #34)
- **maxChildren truncation as a second false-negative class**: children beyond index `maxChildren` are dropped from the stream; very long functions/blocks could silently lose detectable clones. (report #35)
- **Seed `--suggest-generics` from sub-statement analysis**: with nested emission, generics candidates can now be found at statement level inside composite bodies; annotate those specifically. (report #40)

## Platform and Ecosystem (harvested 2026-09-19)

**Source:** `docs/status/2026-09-19_06-41_v0.7.0-release-go1.27-coherence-ci-recovery.md` (#N = that report's task number).

- **Real Windows stdin cancellation** (#5, gated on the platform-investment question): `feedFromStdin`'s close-on-cancel relies on `os.File.Close` unblocking a pending `Read`, which Windows `os.Pipe` does not guarantee — three stdin tests are Windows-skipped because of it. A real fix means a thread-based reader or `CancelIoEx`; a supported-Windows decision would also add a windows-required CI lane. Until the platform question is answered, skip-and-document is the correct resting state.
- **Go 1.28 watch item: `GOEXPERIMENT=jsonv2` retirement** (#50): the flake sets the flag; a graduated/retired flag change would alter engine behavior under the v1 API. Output bytes are experiment-independent by construction (ADR-0024), so the risk is confined to build breakage — verify `nix build` on the first 1.28 beta.
- **treefmt withGo127 wrappers → upstream env option** (#33): the symlinkJoin wrappers around goimports/templ exist because the treefmt sandbox has no network for toolchain download; replace with an upstream treefmt env option if one lands.
- **Type-aware ≡ semantic re-verification on go-paperless** (#43): the contract held on the 2026-09-18 tree; re-run type-aware after any future client.go change.
- **`//nolint:dupl`-style parallel tests in go-paperless client_test.go** (#44): documented as intentional; converting to table tests needs owner sign-off.
- **`anchor_id` in minimal/plumbing JSON** (#46): deliberately excluded today; expose only if cross-format linking demand appears from `--simple-json` consumers.
- **Irreducible `-t 1` residue**: after the 2026-09-18 self-clean (47→44, re-measured at 40 shown groups on 2026-09-19), remaining groups are idioms or deliberate mirrors marked `//art-dupl:accept`; revisit only when new patterns land.

## Quality and Intelligence

- **Interface-aware suppression (cross-package)**: Same-package interface detection via `go/types` is implemented (the `interface-method` pattern). Cross-package and stdlib interfaces still rely on the static name list (`commonInterfaceMethodNames`). Full call-graph analysis remains future work.
- **Configurable actionability patterns**: `--disable-pattern <label>` and `--list-patterns` implemented. 30 denylist patterns + 4 property-engine labels currently active.
- **ML-based actionability classification**: Train a model on labeled clone data to predict whether a clone is actionable, replacing the rule-based actionability patterns. Would handle edge cases the 30 current patterns miss.
- **Fixability score**: Property-based extractability engine implemented (ADR-0017) with 4 properties + confidence scoring. Three-tier output (actionable / low-confidence / non-actionable). Confidence values need calibration against real-world data. Could evolve into a continuous fixability score instead of binary Actionable/NonActionable.
- **Nested-scope shadowing in alpha-normalization**: Current symbol table is flat (no nested-scope shadowing). Proper lexical scoping would improve Type-2 clone accuracy in deeply nested code.
- **Type narrowing for interface-typed variables**: If a local has an interface type, two variables with the same interface type match even if their concrete types differ. Could add concrete-type awareness via flow analysis.
- **Irreducibility class per group**: Classify each group by WHY it can't be reduced: semantic (real duplication), idiomatic (language convention), structurally-bound (framework constraint), or threshold-noise. Would give users actionable context instead of a binary verdict. (samber-do-auditlog feedback)
- **Helper-call-site detection via call-graph**: Detect when duplication is already an invocation of a shared helper (the extraction IS the helper call, not more duplication). Requires call-graph resolution. (upd, discordsync feedback)
- **Golden-corpus recall harness**: run detection over a curated set of known clones (including loop-skeleton shapes) and report recall %. Turns "did the granularity change help?" into a number. (report #42)
- **Type-aware value heuristic**: report how many type-aware-killed groups were actionable, so the 24x runtime cost becomes a data-driven decision per repo. (report #41/e6)
- **High-similarity-no-clone warning**: warn when functions share >50% of statements but sit below the threshold — "near-miss, lower -t to N to see it". (report #43)
- **Cross-repo recall survey**: scan Lars's Go repos for loop-skeleton clones to size ADR-0023's win. (report #39)
- **CPU-affinity benchmarks post-ADR-0023**: re-run `taskset` pinned comparisons; token streams grew ~10-15%, which shifts the L3-latency tradeoff. (report #47)
- **README before/after example showing loop-skeleton detection**: marketing artifact for the ADR-0023 recall win. (report #45)

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
