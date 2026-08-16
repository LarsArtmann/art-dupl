# Printer Package Decoupling — Pareto Execution Plan

> **Created:** 2026-07-26 16:31
> **Status:** DONE — executed 2026-07-26
> **Scope:** Split the 6,600-LOC `printer/` god-package into cohesive leaves without breaking behavior.

---

## 0. Diagnosis — Reframing the Premise

**The "circular dependency" framing from the prior session is wrong.** There is no import cycle
today. `cmd → printer` is strictly one-directional; `printer` imports only
`domain / config / syntax / syntax/golang / baseline / errors`. What actually exists is
**cohesion debt**: one god-package holding four genuinely distinct concerns, which makes every
change risk a blast-radius across unrelated output formats.

This reframes the goal: **carve clean leaves, don't unblock a cycle.** That is mechanical,
behavior-preserving, and low-risk — not an architecture rescue.

### The four natural clusters (prod LOC, excluding tests + generated)

| #  | Cluster                                                          | Files | Prod LOC | Inbound coupling                                                                                                                               |
| -- | ---------------------------------------------------------------- | ----- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| C1 | **actionability** (pattern suppression + classification)         | 8     | 1,469    | Depends ONLY on `domain`+`syntax`+`syntax/golang`. **Zero** calls into formats/core. Misclassified: it's detection/classification, not output. |
| C2 | **stats** (statistics collection + formatting)                   | 8     | 1,396    | Self-contained; `stats` struct implements `StatsPrinter`. `StatsView` referenced only by the interface method in `printer.go`.                 |
| C3 | **format printers** (text/json/html/sarif/plumbing)              | 8     | 2,271    | Depend on clone-core; implement `Printer`. Thin layers over core.                                                                              |
| C4 | **core + diff** (clone_processor, sorting, groups, issuer, diff) | 11    | 1,456    | The shared backbone + the `Printer`/`StatsPrinter` interfaces.                                                                                 |

**Total: ~6,592 prod LOC + ~7,100 test/generated LOC = ~13.7k lines in one package.**

### The verified seam facts (from research)

1. **Actionability is the cleanest seam.** `EvaluateActionability*`, `AllActionabilityPatterns`,
   `ListActionabilityPatterns`, `PatternLabel`, `ToCloneNodeSeqs` — all depend only on
   `domain`+`syntax`. Actionability test files have **ZERO** references to format/core symbols.
2. **`clone_classify.go` moves wholesale** with actionability. `ClassifyClone` and
   `applyPatternLabel` are called ONLY from `clone_processor.go` (the bridge), and
   `applyPatternLabel` references `PatternLabel`. No file split needed — move it all.
3. **Stats is the second-cleanest seam.** `StatsView` is referenced only by
   `printer.go`'s `GetStatsView() *StatsView`. Resolution: root imports the sub-package
   (root→stats is legal; stats does NOT import root).
4. **The interfaces stay in root.** `Printer`/`StatsPrinter`/`ReadFile`/`HashSetter`/
   `RichTextSetter`/`ExplainSetter`/`StatsConfig` are consumed by `cmd/` and implemented by
   formats. Moving them would touch every `cmd/` import site for zero benefit — they reference
   `config.SortCriteria`, so they can't live in `domain` anyway.

---

## 1. Pareto Breakdown

### The 1% that delivers 51% → **Extract `printer/actionability/`**

Single highest-impact, lowest-risk move. Removes the most LOC (1,469) from the god-package and
fixes a genuine **misclassification**: pattern suppression is detection/classification logic,
not output formatting. Independently shippable. Zero behavior change.

### The 4% that delivers 64% → **Actionability + Stats extraction**

Add the 1,396-LOC stats cluster. Two clean leaves, ~2,865 LOC removed from root.

### The 20% that delivers 80% → **Actionability + Stats + arch-lint rules + docs**

Wire the new boundaries into `.go-arch-lint.yml` so they can't silently re-couple. Update
`AGENTS.md` architecture map + module tree. Update `CHANGELOG.md`. This is where the split
becomes _durable_ rather than cosmetic.

### The other 20% (to reach 100%) → **Format printer extraction (Phase 3, DEFERRED)**

text/json/html/sarif/plumbing → sub-packages. Higher churn (every format touches
`clone_processor`/`sorter`/`toJSONClone`), modest cohesion gain. **Recommendation: SKIP unless
adding new formats or the root still feels unwieldy after Phase 1+2.** Not part of this sprint.

---

## 2. Comprehensive Plan — Medium Granularity (30–100 min tasks)

Sorted by impact/effort/customer-value. "Customer" = developer navigating the codebase.

| ID      | Task                                                                                                                                                      | Impact      | Effort | Customer Value                                 | Depends     |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | ------ | ---------------------------------------------- | ----------- |
| **M01** | **Phase 1a:** Create `printer/actionability/` package, move 8 prod files (`actionability*.go`, `clone_classify.go`), fix `package` decls                  | 🔴 Critical | 45m    | Biggest LOC reduction; fixes misclassification | —           |
| **M02** | **Phase 1b:** Move 11 actionability test files, fix `package` decls + imports                                                                             | 🔴 High     | 30m    | Tests stay with code they verify               | M01         |
| **M03** | **Phase 1c:** Rewire consumers: export needed symbols, update `clone_processor.go`, `cmd/run_output.go`, `cmd/run_flags.go`, `cmd/diff_report.go` imports | 🔴 Critical | 40m    | Compilation succeeds                           | M01         |
| **M04** | **Phase 1d:** Run `templ generate && GOEXPERIMENT=jsonv2 go test -count=1 ./... && nix flake check` — verify green                                        | 🔴 Critical | 30m    | Quality gate holds                             | M02,M03     |
| **M05** | **Phase 1e:** Commit Phase 1 with detailed message                                                                                                        | 🟡 Medium   | 10m    | Atomic, reviewable history                     | M04         |
| **M06** | **Phase 2a:** Resolve `StatsView` ownership — update `printer.go` interface to return `*stats.View` (root imports sub-package)                            | 🔴 High     | 25m    | Unblocks stats extraction                      | M05         |
| **M07** | **Phase 2b:** Create `printer/stats/` package, move 8 prod `stats*.go` files, fix `package` decls                                                         | 🔴 High     | 40m    | Second clean leaf                              | M06         |
| **M08** | **Phase 2c:** Move 5 stats test files, fix imports                                                                                                        | 🟡 Medium   | 20m    | Tests stay with code                           | M07         |
| **M09** | **Phase 2d:** Rewire `stats` struct to implement `printer.StatsPrinter`; update `NewStats` return; fix `cmd/` consumers                                   | 🔴 High     | 35m    | Compilation succeeds                           | M07         |
| **M10** | **Phase 2e:** Verify green (`go test -count=1 ./... && nix flake check`) + commit Phase 2                                                                 | 🔴 Critical | 30m    | Quality gate holds; atomic history             | M08,M09     |
| **M11** | **Phase 3a:** Add `actionability` + `stats` components to `.go-arch-lint.yml` with allowed deps                                                           | 🔴 High     | 20m    | Boundary becomes enforceable                   | M10         |
| **M12** | **Phase 3b:** Add arch-lint rule: `actionability` must NOT depend on `printer`/`stats`/formats; `stats` must NOT depend on `actionability`                | 🔴 High     | 20m    | Prevents silent re-coupling                    | M11         |
| **M13** | **Phase 3c:** Run `nix flake check` (arch-lint check) + fix violations                                                                                    | 🟡 Medium   | 20m    | Boundary enforced in CI                        | M12         |
| **M14** | **Phase 3d:** Update `AGENTS.md` architecture map + module tree (add `printer/actionability/` and `printer/stats/` lines)                                 | 🟡 Medium   | 15m    | Docs match reality                             | M10         |
| **M15** | **Phase 3e:** Update `CHANGELOG.md` Changed section (package split, behavior-preserving)                                                                  | 🟢 Low      | 10m    | Change history                                 | M10         |
| **M16** | **Phase 3f:** Final verify + commit Phase 3 (arch rules + docs)                                                                                           | 🔴 High     | 15m    | Sprint complete                                | M13,M14,M15 |
| **M17** | **Guardrail:** Check `.golangci.yml` for daemon-re-added `exhaustruct`/`tagliatelle` before every `nix flake check`; remove if present                    | 🔴 Critical | 5m/run | Quality gate isn't sabotaged                   | ongoing     |
| **M18** | **Watch:** Verify generated `report_templ.go` still regenerates cleanly (`templ generate`) — HTML printer stays in root                                   | 🟡 Medium   | 10m    | No templ breakage                              | M04         |
| **M19** | **Sanity:** Confirm `pkg/artdupl` SDK still has ZERO imports of `printer/` (it should — it never imported it)                                             | 🟢 Low      | 5m     | SDK isolation intact                           | M10         |
| **M20** | **Decision gate (DEFER):** Evaluate whether Phase 4 (format printer extraction) is warranted. Recommend: **NO** unless root still feels large.            | 🟢 Low      | 10m    | Avoid over-engineering                         | M16         |

**Total estimated effort: ~7.5 hours.** The 1% (M01–M05) = ~2.5h. The 4% (add M06–M10) = +~2.5h.
The 20% (add M11–M16) = +~1.75h. The remaining 80% effort (format extraction) is **deferred**.

---

## 3. Detailed Breakdown — Fine Granularity (≤12 min tasks)

Sorted by execution order within each phase. IDs prefixed by phase.

### Phase 1 — Extract `printer/actionability/` (the 1%)

| ID  | Task                                                                                                                                                                                                                                                                    | Est | Depends |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --- | ------- |
| F01 | `mkdir printer/actionability`                                                                                                                                                                                                                                           | 1m  | —       |
| F02 | `git mv printer/actionability.go printer/actionability/`                                                                                                                                                                                                                | 1m  | F01     |
| F03 | `git mv` the 7 remaining prod files: `actionability_boilerplate.go`, `actionability_control_flow.go`, `actionability_data.go`, `actionability_interface_method.go`, `actionability_patterns_expanded.go`, `clone_classify.go`                                           | 3m  | F01     |
| F04 | Edit each moved prod file: `package printer` → `package actionability` (8 files)                                                                                                                                                                                        | 6m  | F02,F03 |
| F05 | `git mv` the 11 test files (`actionability*_test.go`, `clone_classify_test.go`, `actionability_test_patterns.go`)                                                                                                                                                       | 3m  | F01     |
| F06 | Edit each moved test file: `package printer` → `package actionability` (11 files)                                                                                                                                                                                       | 8m  | F05     |
| F07 | Export symbols that consumers need: confirm `EvaluateActionability`, `EvaluateActionabilityWithDisabled`, `EvaluateActionabilityWithLabel`, `AllActionabilityPatterns`, `ListActionabilityPatterns`, `PatternLabel`, `ClassifyClone` are already capitalized (they are) | 3m  | F04     |
| F08 | Update `printer/clone_processor.go`: add `import "github.com/LarsArtmann/art-dupl/printer/actionability"`, qualify calls (`actionability.EvaluateActionability...`, `actionability.ClassifyClone`, `actionability.PatternLabel`)                                        | 8m  | F04,F07 |
| F09 | Update `cmd/run_output.go`: add actionability import, qualify `actionability.AllActionabilityPatterns` / `ListActionabilityPatterns` calls                                                                                                                              | 5m  | F07     |
| F10 | Update `cmd/run_flags.go`: qualify actionability calls (if any)                                                                                                                                                                                                         | 5m  | F07     |
| F11 | Update `cmd/diff_report.go`: qualify actionability calls (if any)                                                                                                                                                                                                       | 5m  | F07     |
| F12 | `GOEXPERIMENT=jsonv2 go build ./...` — fix any remaining qualification errors                                                                                                                                                                                           | 8m  | F08-F11 |
| F13 | `templ generate` (ensure report_templ.go unaffected)                                                                                                                                                                                                                    | 2m  | F12     |
| F14 | `GOEXPERIMENT=jsonv2 go test -count=1 ./printer/actionability/...` — sub-package tests pass                                                                                                                                                                             | 5m  | F12     |
| F15 | `GOEXPERIMENT=jsonv2 go test -count=1 ./...` — whole repo passes                                                                                                                                                                                                        | 8m  | F14     |
| F16 | Check `.golangci.yml` for daemon-re-added `exhaustruct`/`tagliatelle`; remove if present (M17 guardrail)                                                                                                                                                                | 3m  | F15     |
| F17 | `nix flake check` — all 12+ checks pass                                                                                                                                                                                                                                 | 10m | F16     |
| F18 | `git add -A && git commit` Phase 1 with detailed message                                                                                                                                                                                                                | 5m  | F17     |

**Phase 1 subtotal: ~84 min (1.4h).** This is the 1% that delivers 51%.

### Phase 2 — Extract `printer/stats/` (completes the 4%)

| ID  | Task                                                                                                                                                                                                                                                                                                                                                           | Est | Depends |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --- | ------- |
| F19 | Decide `StatsView` direction: rename to `stats.View` in sub-package; `printer.go` interface returns `*stats.View`; root imports sub-package                                                                                                                                                                                                                    | 2m  | F18     |
| F20 | `mkdir printer/stats`                                                                                                                                                                                                                                                                                                                                          | 1m  | F18     |
| F21 | `git mv` 8 prod `stats*.go` files into `printer/stats/`                                                                                                                                                                                                                                                                                                        | 3m  | F20     |
| F22 | Edit each: `package printer` → `package stats` (8 files)                                                                                                                                                                                                                                                                                                       | 6m  | F21     |
| F23 | Rename `StatsView` → `View`, `TopCloneGroup` → keep or `stats.TopCloneGroup` (decide); update all internal refs in stats files                                                                                                                                                                                                                                 | 8m  | F22     |
| F24 | `git mv` 5 stats test files into `printer/stats/`                                                                                                                                                                                                                                                                                                              | 2m  | F20     |
| F25 | Edit each test: `package printer` → `package stats` (5 files)                                                                                                                                                                                                                                                                                                  | 4m  | F24     |
| F26 | Update `printer/printer.go`: `import "github.com/LarsArtmann/art-dupl/printer/stats"`; change `GetStatsView() *StatsView` → `*stats.View`                                                                                                                                                                                                                      | 5m  | F23     |
| F27 | Update `printer/stats/stats.go`: `NewStats` returns `printer.Printer`; add `import ".."` alias for root package — **RISK**: stats importing root printer. **Resolution:** `NewStats` returns concrete `*stats` struct (it already implements the interface structurally); caller in `cmd/` type-asserts or assigns to `printer.StatsPrinter`. Check signature. | 8m  | F26     |
| F28 | Update `cmd/stats.go` and other `cmd/` consumers: qualify `stats.NewStats`, `stats.View` refs                                                                                                                                                                                                                                                                  | 6m  | F27     |
| F29 | `GOEXPERIMENT=jsonv2 go build ./...` — fix qualification errors                                                                                                                                                                                                                                                                                                | 8m  | F28     |
| F30 | `GOEXPERIMENT=jsonv2 go test -count=1 ./printer/stats/...`                                                                                                                                                                                                                                                                                                     | 5m  | F29     |
| F31 | `GOEXPERIMENT=jsonv2 go test -count=1 ./...`                                                                                                                                                                                                                                                                                                                   | 8m  | F30     |
| F32 | Check `.golangci.yml` for daemon sabotage; remove if present                                                                                                                                                                                                                                                                                                   | 3m  | F31     |
| F33 | `nix flake check`                                                                                                                                                                                                                                                                                                                                              | 10m | F32     |
| F34 | `git add -A && git commit` Phase 2                                                                                                                                                                                                                                                                                                                             | 5m  | F33     |

**Phase 2 subtotal: ~88 min (1.5h).**

### Phase 3 — Make boundaries durable (completes the 20%)

| ID  | Task                                                                                                                                                                         | Est | Depends |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --- | ------- |
| F35 | Edit `.go-arch-lint.yml`: add `actionability` component (`in: printer/actionability/**`) with deps `domain, syntax, syntax-golang, errors, pkg-utils`                        | 5m  | F34     |
| F36 | Edit `.go-arch-lint.yml`: add `stats` component (`in: printer/stats/**`) with deps `domain, config, syntax, errors, pkg-utils` + root `printer` (for interface satisfaction) | 5m  | F34     |
| F37 | Edit `.go-arch-lint.yml`: update `printer` component deps to ADD `actionability, stats` (root now imports both sub-packages)                                                 | 3m  | F35,F36 |
| F38 | Edit `.go-arch-lint.yml`: `actionability` must NOT list `printer, stats, cli-commands` (enforces it's a leaf)                                                                | 2m  | F35     |
| F39 | Run arch-lint check via `nix flake check`; fix any violation                                                                                                                 | 8m  | F37,F38 |
| F40 | Update `AGENTS.md` architecture tree: add `printer/actionability/` and `printer/stats/` lines under `printer/`                                                               | 5m  | F34     |
| F41 | Update `AGENTS.md` Critical Conventions if any actionability/stats convention moved                                                                                          | 5m  | F40     |
| F42 | Update `CHANGELOG.md` Changed section: "Split `printer/` into `printer/actionability/` and `printer/stats/` sub-packages (behavior-preserving)"                              | 5m  | F34     |
| F43 | Final `nix flake check`                                                                                                                                                      | 10m | F39,F42 |
| F44 | `git add -A && git commit` Phase 3                                                                                                                                           | 5m  | F43     |
| F45 | Verify `pkg/artdupl` still imports nothing from `printer/` (`rg 'art-dupl/printer' pkg/`)                                                                                    | 2m  | F44     |
| F46 | **Decision gate:** document Phase 4 (format extraction) as DEFERRED in `ROADMAP.md` or `TODO_LIST.md`                                                                        | 5m  | F44     |

**Phase 3 subtotal: ~78 min (1.3h).**

---

## 4. Execution Graph

```mermaid
flowchart TD
    classDef phase1 fill:#fee,stroke:#c00,stroke-width:2px,color:#900
    classDef phase2 fill:#eef,stroke:#00c,stroke-width:2px,color:#009
    classDef phase3 fill:#efe,stroke:#0a0,stroke-width:2px,color:#060
    classDef gate fill:#ffd,stroke:#aa0,stroke-width:3px,color:#660
    classDef deferred fill:#eee,stroke:#999,stroke-dasharray: 5 5,color:#666

    START([Start: green quality gate]):::gate

    subgraph P1["Phase 1 — Extract printer/actionability/ (the 1% → 51%)"]
        F01[F01 mkdir] --> F02[F02 git mv prod files]
        F02 --> F03[F03 git mv remaining prod]
        F03 --> F04[F04 fix package decls: prod]
        F02 --> F05[F05 git mv test files]
        F05 --> F06[F06 fix package decls: test]
        F04 --> F08[F08 rewire clone_processor.go]
        F06 --> F08
        F04 --> F09[F09 rewire cmd/run_output]
        F04 --> F10[F10 rewire cmd/run_flags]
        F04 --> F11[F11 rewire cmd/diff_report]
        F08 --> F12[F12 go build ./...]
        F09 --> F12
        F10 --> F12
        F11 --> F12
        F12 --> F13[F13 templ generate]
        F13 --> F14[F14 test actionability/]
        F14 --> F15[F15 go test ./...]
        F15 --> F16[F16 check daemon linters]
        F16 --> F17[F17 nix flake check]
        F17 --> F18[F18 COMMIT Phase 1]:::gate
    end

    START --> F01

    subgraph P2["Phase 2 — Extract printer/stats/ (completes the 4% → 64%)"]
        F19[F19 decide StatsView ownership] --> F20[F20 mkdir]
        F20 --> F21[F21 git mv stats prod]
        F21 --> F22[F22 fix package: prod]
        F22 --> F23[F23 rename StatsView → View]
        F20 --> F24[F24 git mv stats tests]
        F24 --> F25[F25 fix package: test]
        F23 --> F26[F26 update printer.go interface]
        F26 --> F27[F27 update NewStats signature]
        F27 --> F28[F28 rewire cmd/ consumers]
        F28 --> F29[F29 go build ./...]
        F29 --> F30[F30 test stats/]
        F30 --> F31[F31 go test ./...]
        F31 --> F32[F32 check daemon linters]
        F32 --> F33[F33 nix flake check]
        F33 --> F34[F34 COMMIT Phase 2]:::gate
    end

    F18 --> F19

    subgraph P3["Phase 3 — Make boundaries durable (completes the 20% → 80%)"]
        F35[F35 arch-lint: actionability component] --> F38[F38 arch-lint: leaf constraint]
        F34 --> F36[F36 arch-lint: stats component]
        F36 --> F37[F37 arch-lint: printer deps add sub-pkgs]
        F37 --> F39[F39 nix flake check arch-lint]
        F34 --> F40[F40 update AGENTS.md tree]
        F40 --> F41[F41 update AGENTS.md conventions]
        F34 --> F42[F42 update CHANGELOG.md]
        F39 --> F43[F43 final nix flake check]
        F42 --> F43
        F41 --> F43
        F43 --> F44[F44 COMMIT Phase 3]:::gate
        F44 --> F45[F45 verify SDK isolation]
        F44 --> F46[F46 document Phase 4 as DEFERRED]
    end

    F34 --> F35
    F34 --> F40
    F34 --> F42

    subgraph P4["Phase 4 — Format printers (DEFERRED — the other 20%)"]
        DEFER[/"text/json/html/sarif/plumbing → sub-packages.
Higher churn, modest gain.
SKIP unless root still feels large
or new formats are being added."/]:::deferred
    end

    F46 -.->|evaluate later| DEFER

    DONE([Done: ~6,600 → ~3,700 LOC in root,
2 clean enforceable leaves,
zero behavior change]):::gate
    F45 --> DONE
    F46 --> DONE

    classDef p1 fill:#fee; classDef p2 fill:#eef; classDef p3 fill:#efe
```

---

## 5. Risks & Mitigations

| Risk                                                                                                  | Likelihood                       | Mitigation                                                                                                                                                                        |
| ----------------------------------------------------------------------------------------------------- | -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Stats ↔ root import cycle** (Phase 2): `NewStats` returns `printer.Printer`, but stats imports root | Medium                           | Resolution in F27: return concrete `*stats` struct; callers assign to interface. Root imports stats (for `stats.View`), stats does NOT import root. Verify with `go build`.       |
| **Daemon re-adds `exhaustruct`/`tagliatelle`** mid-flake-check                                        | High (happened 2× prior session) | Guardrail task M17 / F16 / F32: check `.golangci.yml` before EVERY `nix flake check`. Remove if present.                                                                          |
| **`clone_classify.go` references `PatternLabel`** which moves                                         | Low                              | Confirmed: move `clone_classify.go` wholesale with actionability. Both its callers are in `clone_processor.go` (the bridge), which imports the new package. No file split needed. |
| **Test files cross-reference non-actionability symbols**                                              | Very Low                         | Verified: actionability test files have ZERO references to `toJSONClone`/`CloneGroup`/`JSONClone`/`StatsView`/`PrintClones`/`sortGroups`/`ProcessClones`/`classifyCloneType`.     |
| **`.go-arch-lint.yml` component globs** don't match new layout                                        | Medium                           | F35–F38 add explicit components. Test files are excluded from arch-lint already (`excludeFiles` regex).                                                                           |
| **`report_templ.go` (generated)** breaks if HTML printer moves                                        | Low                              | HTML printer stays in root for this plan. `templ generate` runs in F13 as a smoke test. Phase 4 would own this risk.                                                              |
| **Verschlimmbesserung**: split adds import churn without real benefit                                 | Medium                           | Phases 1+2 only move code that is genuinely separable (verified zero cross-deps). Phase 4 (the marginal part) is explicitly DEFERRED to avoid churn-for-churn's-sake.             |

---

## 6. Non-Goals (What This Plan Does NOT Do)

- ❌ Move the `Printer`/`StatsPrinter` interfaces out of `printer/` root (premature; touches all `cmd/` import sites for no benefit).
- ❌ Split format printers (text/json/html/sarif/plumbing) — deferred to Phase 4, evaluate later.
- ❌ Change any behavior — this is a pure refactor. Tests are the safety net; no logic changes.
- ❌ Rename public CLI flags or output formats — zero user-visible change.
- ❌ Address the daemon regression at its source (separate concern; tracked in prior status report).

---

## 7. Success Criteria

- [x] `printer/actionability/` exists as an independent package importing only `domain`/`syntax`/`syntax/golang`/`errors`.
- [x] `printer/stats/` exists as an independent package.
- [x] `printer/` root is reduced from ~6,600 → ~3,830 prod LOC.
- [x] `.go-arch-lint.yml` enforces the new boundaries (actionability is a leaf; stats doesn't import actionability).
- [x] `GOEXPERIMENT=jsonv2 go test -count=1 ./...` passes.
- [x] `nix flake check` passes (all checks, including arch-lint + self-test).
- [x] `AGENTS.md` + `CHANGELOG.md` reflect the new structure.
- [x] Zero behavior change (golden tests unchanged, BDD tests unchanged).

---

## 8. Effort Summary

| Phase                   | Tasks                   | Est. Time | Pareto Tier   |
| ----------------------- | ----------------------- | --------- | ------------- |
| Phase 1 (actionability) | M01–M05 / F01–F18       | ~1.4h     | **1% → 51%**  |
| Phase 2 (stats)         | M06–M10 / F19–F34       | ~1.5h     | **4% → 64%**  |
| Phase 3 (durability)    | M11–M16 / F35–F46       | ~1.3h     | **20% → 80%** |
| Phase 4 (formats)       | DEFERRED                | —         | other 20%     |
| **Total (executed)**    | **46 fine / 20 medium** | **~4.2h** | **80% value** |

---

_Generated 2026-07-26. Behavior-preserving refactor — the tests and golden files are the contract._
