# Performance Optimization Guide

art-dupl is designed to be fast on large Go codebases. This guide covers tuning
options for different codebase sizes and use cases.

## Quick Reference

| Codebase Size   | Recommended Flags                                     | Expected Time                      |
| --------------- | ----------------------------------------------------- | ---------------------------------- |
| < 100 files     | (defaults)                                            | < 1s                               |
| 100–1000 files  | `--workers 0`                                         | 1–5s                               |
| 1000–5000 files | `--workers 0 --incremental`                           | 3–15s (first run), < 5s (cached)   |
| 5000+ files     | `--workers 0 --incremental --max-cache-entries 10000` | 10–60s (first run), < 10s (cached) |

## Workers (`--workers`)

Controls how many goroutines parse files in parallel.

| Value         | Behavior                                     | When to use                                            |
| ------------- | -------------------------------------------- | ------------------------------------------------------ |
| `0` (default) | Auto-detect: `runtime.GOMAXPROCS(0)` workers | **Recommended for almost all cases**                   |
| `1`           | Sequential parsing                           | Debugging, single-threaded environments, CI with 1 CPU |
| `>1`          | N parallel workers                           | Override auto-detection (rarely needed)                |

**Gotcha:** Values > 1 use a worker pool that dispatches to parallel parsing.
Value `1` uses sequential parsing. Value `0` auto-detects based on available CPUs.

```bash
art-dupl --workers 0 .           # auto-detect (default)
art-dupl --workers 4 .           # 4 parallel workers
art-dupl --workers 1 .           # sequential (for debugging)
```

## Incremental Analysis (`--incremental`)

Enables AST caching to skip re-parsing unchanged files.

```bash
art-dupl --incremental .                        # uses .cache/art-dupl/
art-dupl --incremental --cache-dir /tmp/cache . # custom cache location
art-dupl --incremental --clear-cache .          # force full re-parse
art-dupl --incremental --max-cache-entries 5000 . # limit cache size
```

### Cache Semantics

- **Cache key:** SHA-256 of file content (content-addressed, not filename-based)
- **Invalidation:** Automatic — any content change invalidates that file's cache entry
- **Cache version:** Bumped when the AST serialization format changes; old caches are ignored
- **Eviction:** `--max-cache-entries N` evicts oldest entries by modification time (0 = unlimited)
- **Concurrency-safe:** `singleflight.Group` deduplicates concurrent parses of identical files

### When NOT to use `--incremental`

- **CI pipelines with fresh checkout** — no cache to reuse, adds overhead
- **One-off scans** — cache write overhead with no future benefit
- **Debugging** — simpler to reason about without caching

## Type-Aware Mode (`--type-aware`)

Uses `go/types` for higher-precision detection. **10–100x slower** than syntax-only.

```bash
art-dupl --type-aware .    # type-aware mode (requires --semantic, the default)
```

**When to use:** When you see false positives from same-method-name-different-receiver-type
patterns (e.g., `a.String()` where `a` is `time.Time` vs `*big.Int`).

**When NOT to use:** Normal development. The default semantic mode catches the vast majority
of real clones. Reserve `--type-aware` for final cleanup passes or CI verification.

**Limitation:** Not compatible with `--incremental` (the incremental parser doesn't support
preloaded type data).

## Threshold Tuning (`-t`)

| Threshold     | Effect                                       | Use case               |
| ------------- | -------------------------------------------- | ---------------------- |
| `1–3`         | Maximum sensitivity; catches small clones    | Audit, cleanup sprints |
| `5` (default) | Balanced; catches meaningful cloning         | Day-to-day development |
| `10–15`       | Low sensitivity; only large clones           | CI gate, reduce noise  |
| `20+`         | Very low sensitivity; major duplication only | Legacy codebase triage |

```bash
art-dupl -t 3 .        # high sensitivity
art-dupl -t 5 .        # default (recommended)
art-dupl -t 15 .       # CI gate
```

## Detection Modes

| Mode               | Flag                    | What it detects                            | Speed   |
| ------------------ | ----------------------- | ------------------------------------------ | ------- |
| Semantic (default) | `--semantic` or no flag | Type 2 clones (renamed variables/literals) | Fast    |
| Exact              | `--exact`               | Type 1 clones (identical code)             | Fastest |
| Structural         | `--structural`          | AST shape only (ignores identifiers)       | Fast    |

Semantic mode is recommended for most use cases. It normalizes local variable
names and literal values, catching the most common form of copy-paste duplication.

## Output Format Performance

| Format         | Flag            | Relative Speed           | Use case                 |
| -------------- | --------------- | ------------------------ | ------------------------ |
| Text (default) | (none)          | Fast                     | Interactive use          |
| Plumbing       | `--plumbing`    | Fastest                  | Scripts, CI              |
| JSON           | `--json`        | Moderate                 | Tooling integration      |
| SARIF          | `--sarif`       | Moderate                 | GitHub Advanced Security |
| HTML           | `--html`        | Slow (renders fragments) | Human review             |
| Simple JSON    | `--simple-json` | Moderate                 | Impact scoring           |

## Benchmarking

```bash
go test -run='^TestPerfRegression' -v ./syntax/...    # Run perf regression tests
go test -bench=. ./...                                 # Run all benchmarks
```

The `bench` Nix check runs performance regression tests in CI:

```bash
nix build .#checks.x86_64-linux.bench
```
