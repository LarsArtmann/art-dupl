# Roadmap

**No timeline** — Aspirational items for future consideration.

---

## Language Support

- [ ] Add TypeScript/JavaScript support
- [ ] Add Python support

## IDE & Tooling

- [ ] Implement watch mode for continuous monitoring and incremental detection
- [x] ~~Create GitHub Actions workflow templates~~ — DONE: `templates/github-actions-duplicate-check.yml`
- [x] ~~Create pre-commit hooks~~ — DONE: `templates/pre-commit-hook.yaml`

## Quality & Documentation

- [x] ~~Create Architecture Decision Records (ADRs) for major design decisions~~ — DONE: 14 ADRs in `docs/adr/` (0001-0014: map-based transitions, semantic default, reflection config merge, actionability patterns, split-brain type unification, non-destructive serial, detection mode enum, semantic encoding layout, default threshold change, JSON v2 migration, min-lines minimum, dump-tokens iowriter, exit codes, suppression config)
- [ ] Continue adding ADRs for future major design decisions

## Performance

- [x] ~~Create performance baseline benchmarks~~ — DONE: `syntax/syntax_bench_test.go`
- [x] ~~Create regression test suite for performance~~ — DONE: `syntax/perf_regression_test.go` + `checks.bench` in `flake.nix`

---

_These items are aspirational and have no committed timeline. They represent potential future directions based on user needs and project evolution._
