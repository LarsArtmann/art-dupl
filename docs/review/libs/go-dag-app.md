# Library adoption audit — go-dag-app

| Field | Value |
| --- | --- |
| Library | `github.com/larsartmann/go-dag-app` |
| Version reviewed | checkout HEAD `c2b6c99aeed4f5eabbef01888b4194145e2e52cf` = `git describe --tags` `v0.1.1-1-gc2b6c99` (v0.1.1 tagged 2026-11-02 at `d9934c8`; HEAD is one commit past it, dated 2026-11-20) |
| Consumer | `/home/lars/projects/art-dupl` @ HEAD `02b42ff3c6768d23d9d9d257885d13e2fc666a77`, single Go module `github.com/LarsArtmann/art-dupl` (root go.mod only) |
| Date | 2026-11-23 |
| Auditor | group C fleet run (read-only on code; this file is the only write) |
| Verdict | **NOT-ADOPTED** — score 0/100 (unrealized), see rubric note on the zero denominator |

## Module map

| Repo | Modules | Tags |
| --- | --- | --- |
| go-dag-app | single module `github.com/larsartmann/go-dag-app` (go.mod:1, `go 1.27`, deps: `samber/do/v2 v2.1.0`) | `v0.1.0` (2026-07-18), `v0.1.1` (2026-11-02) |
| art-dupl | single module `github.com/LarsArtmann/art-dupl` (root go.mod only) | not applicable — art-dupl is the consumer under audit |

go-dag-app is tiny by design: two production files, `go-dag-app/container.go` (103 lines) and `go-dag-app/container_test.go` (142 lines). Its [Unreleased] CHANGELOG section (go-dag-app/CHANGELOG.md:7) lists docs/tooling plus the samber/do v2.1.0 bump; the HEAD commit `c2b6c99` adds `NewContainerWithOpts` and is **not yet in the CHANGELOG or a tag** — the API surface is two constructors, one env gate, one printer.

## Capability inventory

| # | Capability | Library evidence | art-dupl evidence | Status |
| --- | --- | --- | --- | --- |
| 1 | `NewContainer(appName)` — samber/do injector + debug-aware cleanup | go-dag-app/container.go:75 | No DI container anywhere: `grep -rn "samber/do" go.mod cmd pkg internal` in art-dupl returns zero hits; go.mod has no DI dependency | N/A — art-dupl has no injector to construct |
| 2 | `NewContainerWithOpts(appName, *do.InjectorOpts)` (HEAD, unreleased) | go-dag-app/container.go:91 | Same as above; no `do.InjectorOpts` surface exists in art-dupl | N/A — requires samber/do |
| 3 | `IsDebugDIEnabled(appName)` — `<APP>_DEBUG_DI` env gate | go-dag-app/container.go:34 (suffix const at go-dag-app/container.go:29) | art-dupl's debug verbosity is CLI flags + `ARTDUPL_NO_PROGRESS` (AGENTS.md progress section; cmd/progress.go suppresses via `--quiet`/env), not a DI-tree gate | N/A — different debug channel, no DI tree to print |
| 4 | `PrintDependencyTree(injector)` — `do.ExplainInjector` wrapper | go-dag-app/container.go:49 | No dependency graph of injected services; the "graph" art-dupl manages is the AST suffix tree (suffixtree/) | N/A — no injector to explain |
| 5 | Shutdown sequence: print tree if debug, then `injector.Shutdown()` | go-dag-app/container.go:96–103 (cleanup closure inside `NewContainerWithOpts`) | art-dupl's teardown is `detector.Close()` + context cancellation (`pkg/provider/provider.go:168–172` defer; `select … ctx.Done()` send discipline per AGENTS.md context-propagation convention; job/sendctx.go centralizes it) | N/A — different lifecycle shape (context+closes, not injector shutdown) |
| 6 | Env-var naming convention for per-app debug (`BUILDFLOW_DEBUG_DI`, `LIBRARY_POLICY_DEBUG_DI`) | go-dag-app/README.md env table | art-dupl's env convention is `ARTDUPL_NO_PROGRESS=1` only; adopting `<APP>_DEBUG_DI` would buy a gate for a tree art-dupl does not have | N/A |
| 7 | "Bootstrap skeleton only" scope discipline (deep DAG features explicitly out of scope) | go-dag-app/README.md "What is NOT shared"; go-dag-app/container.go:8–13 | art-dupl's orchestration is a hand-rolled channel pipeline (`job/parse.go` → `job/buildtree.go`, parallel variants, worker routing in job/parse.go per AGENTS.md workers-routing convention) — no DAG framework underneath | N/A — the library explicitly does not sell orchestration |
| 8 | Consumer posture: BuildFlow + library-policy registered consumers | go-dag-app/README.md Consumers section | art-dupl is neither a DAG app nor a `samber/do` consumer; it *is* a BuildFlow tool **provider** (pkg/provider/provider.go:87 self-registration) but that integration is go-finding/toolsdk, not DI | N/A |

Score: 8 rows, 8 N/A rows (one-line reasons above) → A = 0 applicable rows. Σ = 0. With a zero denominator the rubric is undefined; by convention this audit records **0/100, band unrealized**. The honest reading is "no applicable surface at all", not "applicable surface wasted".

## Fit analysis

### Does art-dupl have any DAG/orchestration/DI need?

No, on all three counts:

- **DI**: art-dupl wires its pipeline with plain constructors and interfaces (`detection.MultiDetector` dispatches to `MethodDetector` implementations per AGENTS.md architecture table; cmd/run_analysis.go assembles config → crawl → parse → tree → search by direct calls). There is no service graph that would benefit from `samber/do` lazy resolution, and no debug-tree need — the AGENTS.md io.Writer-injection convention (all cmd output takes injectable writers, no `os.Stderr` globals) is a stricter discipline than go-dag-app's `fmt.Println`-based `PrintDependencyTree` (go-dag-app/container.go:49–61 carries a `//nolint:forbidigo` for exactly that).
- **Orchestration/DAG**: art-dupl's pipeline is a linear staged flow (parse → serialize → build tree → search, threaded with `DetectionMode` and `SearchWorkers`), with fan-out via bounded worker channels (job/parse.go, job/incremental.go) and the `sendCtx[T]` cancellation pattern (job/sendctx.go). A DAG framework (go-workflow, as used by BuildFlow) would add nothing: there are no conditional stages, retries, or resume requirements. go-dag-app itself refuses this scope (go-dag-app/container.go:11–13).
- **Consumer fit**: go-dag-app's two existing consumers (BuildFlow, library-policy) are long-running multi-service apps. art-dupl is a batch CLI + an embedded toolsdk detector; neither shape matches.

### What art-dupl would actually have to change to adopt

Hypothetically: add `samber/do/v2`, wrap cmd startup in `dagapp.NewContainer("ARTDUPL")`, register the detector/services as `do.Provider` instances. Every step is negative-value: a new dependency for a pipeline that has exactly three services, plus a debug env var duplicating existing flag/env plumbing. No integration sketch is warranted.

## Version currency

| Item | Finding |
| --- | --- |
| Local tag | `v0.1.1` (2026-11-02, `d9934c8`); HEAD `c2b6c99` is `v0.1.1-1` |
| Cadence | Slow, deliberate: v0.1.0 (2026-07-18) → v0.1.1 (2026-11-02), ~3.5 months |
| [Unreleased] intent | go-dag-app/CHANGELOG.md:7 — docs/tooling tier plus samber/do v2.1.0 bump; **HEAD's `NewContainerWithOpts` is not yet in the CHANGELOG** (doc lag behind the 2026-11-20 commit) |
| Tag ETA | go-dag-app/TODO_LIST.md §3: "Tag the next release … After the current doc/test work lands, cut v0.1.1 so library-policy can pin a tag" — stale text (v0.1.1 already exists); the next cut (containing `NewContainerWithOpts`) is unqueued |

## Fleet argument

The fleet argument **for** go-dag-app is real but narrow: BuildFlow and library-policy genuinely hand-copied the same ~30 lines of DI bootstrap, and centralizing that kills per-repo drift. art-dupl is not part of that duplication class — it never wrote the scaffold, so there is nothing to deduplicate by adopting. For art-dupl, adopting would mean *new* coupling (a DI framework) to gain a shared utility for machinery it does not have. The local channel-pipeline conventions (AGENTS.md context-propagation, io.Writer injection, workers routing) are more precise for this codebase than a generic container bootstrap.

## Ranked opportunities

None. There is no impact×ease cell with non-zero impact. If art-dupl ever grows a long-running service component (e.g. a daemon watching for clone regressions), the RECONSIDER-WHEN trigger below applies.

## RECONSIDER-WHEN (flip triggers)

1. art-dupl ships any long-running, multi-service component (watch daemon, LSP server, HTTP dashboard) that would naturally use `samber/do` DI — then evaluate `NewContainerWithOpts` for the bootstrap + `<APP>_DEBUG_DI` gate.
2. art-dupl becomes a BuildFlow **host** consuming sibling tools in-process with a service graph (today it is a provider, the other side of that seam — pkg/provider/provider.go:87–124).
3. go-dag-app's scope grows beyond the bootstrap skeleton into shared DAG orchestration that matches art-dupl's job pipeline shape.

## How to verify

```bash
cd /home/lars/projects/art-dupl && git rev-parse HEAD   # 02b42ff3c6768d23d9d9d257885d13e2fc666a77 at audit time
cd /home/lars/projects/go-dag-app && git rev-parse HEAD && git describe --tags
   # c2b6c99aeed4f5eabbef01888b4194145e2e52cf / v0.1.1-1-gc2b6c99
git -C /home/lars/projects/go-dag-app log -1 --format='%h %ci' v0.1.1   # d9934c8 2026-11-02
grep -rn "samber/do" /home/lars/projects/art-dupl/go.mod /home/lars/projects/art-dupl/cmd /home/lars/projects/art-dupl/pkg /home/lars/projects/art-dupl/internal   # zero hits
grep -rn "samber\|do.Injector\|go-workflow" /home/lars/projects/art-dupl --include='*.go' | grep -v _test   # zero production hits
sed -n '1,103p' /home/lars/projects/go-dag-app/container.go   # full API surface: 4 exported symbols
```
