# Roadmap

**No timeline** — Aspirational items for future consideration. Items graduate to `TODO_LIST.md` when they become actionable.

---

## Type-Aware Detection

- [ ] **Interface-aware suppression** — Detect method signatures that implement an interface contract and suppress them as structural duplication, not actionable cloning. Needs call-graph analysis or `go/types`. _(The `--type-aware` opt-in mode itself is tracked in TODO_LIST as HIGH priority.)_

## Language Support

- [ ] **TypeScript/JavaScript support** — Would require a TS AST parser (Babel, swc, or tree-sitter). The detection pipeline (suffix tree, hash, actionability) is language-agnostic; only the AST-to-Node transformer needs per-language implementation.
- [ ] **Python support** — Same architecture as TypeScript: a Python AST-to-Node transformer. The `syntax/` package is designed for multi-language extension.

## IDE & Tooling

- [ ] **LSP server mode** — Run art-dupl as a Language Server, surfacing clones as diagnostics in VS Code / GoLand. Would enable real-time clone detection during development.
- [ ] **Watch mode** — Continuous monitoring with incremental detection. Re-run only on changed files. Would pair with `--incremental` caching.
- [x] ~~GitHub Actions workflow templates~~ — DONE: `templates/github-actions-duplicate-check.yml`
- [x] ~~Pre-commit hooks~~ — DONE: `templates/pre-commit-hook.yaml`

## Architecture

- [ ] **Plugin architecture for detection methods** — Allow third-party detectors beyond suffix-tree and hash. The `detection.MethodDetector` interface supports this but is not documented as a public extension point.
- [ ] **WASM target** — Compile art-dupl to WebAssembly for in-browser clone detection. Would enable a web UI dashboard.
- [ ] **Parallel suffix tree construction** — Current Ukkonen's algorithm is single-threaded. Parallel construction could improve throughput on large codebases (10000+ files).

_(Clone type consolidation and printer/ package split are tracked in TODO_LIST as actionable work.)_

## Quality & Intelligence

- [ ] **ML-based actionability classification** — Train a model on labeled clone data to predict whether a clone is actionable, replacing the rule-based actionability patterns. Would handle edge cases the 15 current patterns miss.
- [ ] **Fixability score** — Replace binary Actionable/NonActionable with a score reflecting extraction cost (params needed, lines saved, complexity). Feedback: httputil session suggested "would-take-more-params-than-lines" heuristic.
- [ ] **Nested-scope shadowing in alpha-normalization** — Current symbol table is flat (no nested-scope shadowing). Proper lexical scoping would improve Type-2 clone accuracy in deeply nested code.

_(`--diff-report` baseline mode is tracked in TODO_LIST as actionable work.)_

## Documentation & Adoption

- [x] ~~Architecture Decision Records (ADRs)~~ — DONE: 14 ADRs in `docs/adr/` (0001-0014)
- [x] ~~Astro + Starlight documentation website~~ — DONE: deployed to `art-dupl.lars.software`
- [ ] **SARIF output validation** — Validate emitted SARIF JSON against GitHub's official schema validator library in CI.
- [ ] **Performance optimization guide** — Document `--workers`, `--incremental`, `--cache-dir` tuning for different codebase sizes.
- [ ] **awesome-go submission** — Submit to awesome-go list once stable v1.0 is tagged.

---

_These items are aspirational and have no committed timeline. They represent potential future directions based on user needs, feedback sessions, and project evolution._
