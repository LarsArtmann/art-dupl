# TODO List

**Last Updated:** 2026-08-16

Actionable items for the next 2-4 weeks. Completed work lives in `CHANGELOG.md`.
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### HTML output improvements (remainder)
**Source:** `docs/status/2026-08-10_05-29_feedback-review-implementation-sprint.md` §b
"Detected vs Actionable" summary is DONE (both text and HTML printers). Remaining:
- [ ] **TTY-aware HTML output**: Auto-write to `art-dupl-report.html` when TTY detected (currently requires `--html-out` explicitly).
- [ ] **Stable display IDs**: `GroupNum` is not stable across runs (content hash IDs ARE stable). Add stable `id` attributes for deep-linking.

---

## MEDIUM Priority

### Feedback-driven actionability patterns (remainder)
**Source:** feedback files in `docs/feedback/new/`
`go-http-error-guard` and `go-error-wrap-idiom` are DONE — unified into the single
`error-guard-fallthrough` pattern (validated on DiscordSync: 82 groups → 2 shown, 0 FPs).
Remaining candidates:
- [ ] **`//go:embed` directive pattern**: Detect `//go:embed` + `embed.FS` + `fs.Sub` as compiler-bound (go-sse feedback)
- [ ] **`TestMain` boilerplate**: Go requires one `TestMain` per package (go-cqrs-lite)
- [ ] **`defer-cleanup-of-arbitrary-resource`**: `defer rows.Close()` / `defer tx.Rollback()` (discordsync — 12 groups)

### Correctness hardening
**Source:** `docs/status/2026-08-16_04-23_suffixtree-followup-slice-transitions-race-fixes.md` §e/§f — fallout of the cache metadata race fix.
- [ ] **Atomic/mutex-mixing audit**: grep-driven sweep for struct fields touched by `atomic.*` in one method and read plainly in another (the class that caused the cache race). Fix or document each hit.
- [ ] **Cache `Clear()` concurrent-stats regression test**: assert hit/miss counters survive concurrent `Get` during `Clear` under `-race` (locks in this session's fix).
- [ ] **CI `-race` cadence + flake gate**: run `nix flake check`; decide whether full `-race ./...` runs every push or nightly (it is green now for the first time); wire the decision.

### Code hygiene
- [ ] **Suffixtree cleanup trio**: verify `benchmarkFindTranMethod` is dead → delete; modernize `b.N` → `b.Loop()` (3 sites in `parallel_bench_test.go`); verify/fix the stale "Workers routing" bullet in `AGENTS.md`.
- [ ] **Suffixtree doc polish**: refresh package doc type list; name the `benchmarkMemoryUsage` magic numbers; evaluate `maxStackKeys` getAll helper dedup.

### Cache stats visibility
- [ ] **Surface cache stats in `stats` subcommand**: `Stats.MemHits`/`Hits`/`Misses` are printed in verbose mode only (`cmd/run_analysis.go::printCacheStats`). Wire into `stats` output so non-verbose users see cache effectiveness.

### Infrastructure
- [ ] **`--exclude-pattern` UX**: Warn when pattern matches zero files; document glob vs regex (licenseforge feedback)
- [ ] **Coverage baseline**: Run `go test -cover` across all packages and commit a coverage baseline. We have benchmark baselines but no coverage baseline.

### Suffix tree / performance follow-ups
**Source:** `docs/status/2026-08-16_03-34_data-layout-allocation-optimization-sprint.md` §f + ADR-0022.
The core layout work (slice transitions, arena, pool, budgets) is DONE — see CHANGELOG.
- [ ] **`serial()` bulk Node allocation** (`syntax/syntax.go`): pre-allocate `make([]Node, count)` and index instead of `&Node{}` per node — nodes become cache-line adjacent. Identified twice, never attempted.
- [ ] **`sync.Pool` for `[]*Node` stream slices**: `SerializeWithMaxChildren` allocates `make([]*Node, 0, 10)` per call.
- [ ] **CI allocation regression detection**: `suffixtree/alloc_budget_test.go` covers the suffix tree; extend `testing.AllocsPerRun` budgets to `syntax/` serialization, or add a CI benchstat job on allocation columns (timing too noisy for CI).
- [ ] **Real-world benchmark**: benchmark against an actual Go project repo (not synthetic tokens) to measure end-to-end impact of the suffix tree work.
- [ ] **`taskset -c 1` benchmark protocol**: pin benchmarks to one core to cut thermal noise; current timing comparisons stay noisy.
- [ ] **Slice-vs-reference-map property test**: build both transition representations from one token stream; assert identical `findTran` results and transition sets (guards insert-sort/binary-search bugs).
- [ ] **Fuzz high-fanout seeds**: extend `FuzzSuffixTreeUpdate` with many-distinct-token alphabets to stress binary-search `findTran` + insert-sorted `addTran`.
- [ ] **`linearScanMax` boundary micro-benchmark**: exact 8 vs 9 transitions per state, assert the crossover holds.
- [ ] **`perf stat` cache-miss evidence**: hardware-counter proof (or refutation) of ADR-0022's cache-locality claims via git worktree A/B against `23fa1b4f`.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **`int32` arena indices instead of `*state` pointers**: slice-based transitions (ADR-0022) already removed most pointer chasing; the remaining win is small vs the risk of converting every `*state` to index arithmetic. Revisit only with profile evidence. (Supersedes the old "hybrid slice/map" item — implemented as pure sorted slices in ADR-0022.)
- [ ] **`sync.Pool` for contextList `[]Pos` slices**: rejected in ADR-0022 — slices transfer between contextLists via `append` (which may reallocate), so lifetime tracking would out-complex the savings. Revisit only if search allocations become dominant again.
- [ ] **Restructure `TypeAwareData` so `EraseHash` is collection-level**: Currently per-entry on `PreloadedAST`, validated at runtime with a warning (`job/incremental.go::SetTypeAwareData`). A collection-level type would enforce the invariant at compile time. Breaking change to `syntax/golang/typeinfo.go` with large blast radius.
