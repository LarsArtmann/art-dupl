# Modularization Proposal — art-dupl

**Date:** 2026-05-14 (updated 2026-06-15)
**Status:** Proposal — Reviewed and Updated
**Module:** `github.com/LarsArtmann/art-dupl`

---

## 1. Executive Summary

art-dupl is a Go code duplication detection tool currently structured as a single-module monolith with 26 packages, ~16,300 production LOC, and ~29,500 test LOC. While the internal package structure is well-layered (proper DAG, no import cycles), the single go.mod allows unchecked coupling — every package can reach every other package.

**Why modularize:**

1. **Enforce architectural boundaries** — The `printer` package (41 files) directly imports `syntax/golang` constants, making the output layer language-specific. Module boundaries would make this a compile-time error.
2. **Enable independent versioning** — The SDK (`pkg/artdupl`) and CLI are separate consumers with different stability guarantees. The SDK should be independently versionable.
3. **Faster CI** — Changes to `printer` should not require re-testing `suffixtree`. Module boundaries enable targeted CI.
4. **Clearer ownership** — Each module has a defined public API, making it clear what's internal vs. external.

**What changes:** Split the monolith into 5 sub-modules coordinated by a `go.work` file. Requires promoting 3 `internal/` packages to public packages first (pre-work), then creating new `go.mod` files at existing package roots.

**Expected benefits:**

- Compile-time enforcement of dependency direction
- Independent build/test cycles per module
- Clearer public API surfaces (each module's exports are its contract)
- SDK can be consumed without pulling CLI dependencies

---

## 2. Current State Analysis

### 2.1 Module Landscape

| Property                 | Value           |
| ------------------------ | --------------- |
| go.mod files             | 1 (root only)   |
| go.work                  | None            |
| Go version               | 1.26.2          |
| Total packages           | 26              |
| Production LOC           | ~16,300         |
| Test LOC                 | ~29,500         |
| Direct external deps     | 10              |
| Transitive external deps | 42              |
| Import cycles            | None (verified) |

### 2.2 Current Dependency Graph (Layered)

```
Layer 0 (leaf nodes, no internal deps):
  errors, internal/simd, internal/testhelpers, pkg/format, pkg/logger, pkg/position

Layer 1:
  domain → errors
  internal/utils → errors
  syntax/golang → syntax
  syntax/templ → syntax

Layer 2:
  syntax → suffixtree, internal/simd, pkg/format
  suffixtree → errors

Layer 3:
  cache → errors, syntax
  hash → pkg/format, pkg/logger, syntax
  internal/testutil → syntax, syntax/golang, internal/utils

Layer 4:
  job → syntax, cache, pkg/logger, suffixtree, syntax/golang, syntax/templ
  detection → syntax, domain, config, hash, pkg/logger, suffixtree

Layer 5:
  printer → config, errors, syntax, syntax/golang, pkg/position
  pkg/artdupl → config, errors, pkg/logger, detection, job, suffixtree, syntax, pkg/position

Layer 6:
  cmd → config, errors, printer, job, syntax, suffixtree, detection, hash, internal/utils

Layer 7 (entry points):
  cmd/art-dupl → cmd
  examples → pkg/artdupl, pkg/logger
  bdd → internal/testutil
```

### 2.3 Coupling Hotspots

| Hotspot                                 | Severity   | Description                                                                                      |
| --------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------ |
| `syntax.Node` as universal data carrier | **TIGHT**  | Flows through every pipeline stage: suffixtree, detection, job, printer, hash, cache, artdupl    |
| `syntax` → `suffixtree.Match` type      | **TIGHT**  | `FindSyntaxUnits()` takes `suffixtree.Match` as parameter — conceptual circular dependency       |
| `printer` → `syntax/golang` constants   | **TIGHT**  | 50+ hardcoded Go AST node type constants in `clone_classify.go`                                  |
| `MultiDetector` concrete wiring         | **MEDIUM** | `MethodDetector` interface defined but unused — switch/case on method strings                    |
| `cmd` as god orchestrator               | **MEDIUM** | Imports 7+ packages, does type assertions to concrete printer types                              |
| `pkg/artdupl` SDK internal coupling     | **MEDIUM** | Exposes clean external API but directly constructs `detection.MultiDetector` and `job.BuildTree` |

### 2.4 God-Package Analysis

| Package           | Files  | Concerns                                                                                            | Verdict                                         |
| ----------------- | ------ | --------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| **printer**       | **41** | 6 output formats, stats collection/formatting/visualization, sorting, diffing, clone classification | **God package** — needs splitting within module |
| **cmd**           | **18** | CLI orchestration, flags, config building, crawling, analysis, output                               | Large but coherent                              |
| **bdd**           | **22** | BDD integration tests                                                                               | Large but coherent (test-only)                  |
| **syntax/golang** | **15** | Go AST parsing, transform, node types, identifier hashing                                           | Borderline                                      |

---

## 3. Proposed Module Structure

### 3.1 Module Definitions

#### Module 1: `art-dupl-core` (path: `./`)

**Purpose:** Foundation — domain types, error types, suffix tree algorithm, AST processing, and configuration.

This is the root module, keeping the existing module path for backward compatibility.

| Field                          | Content                                                                                                                                                                                                                                                    |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Module path**                | `github.com/LarsArtmann/art-dupl`                                                                                                                                                                                                                          |
| **Contains packages**          | `domain/`, `errors/`, `suffixtree/`, `syntax/`, `syntax/golang/`, `syntax/templ/`, `config/`, `internal/simd/`, `internal/utils/`, `internal/testhelpers/`, `pkg/format/`, `pkg/logger/`, `pkg/position/`, `cache/`, `hash/`, `job/`, `internal/testutil/` |
| **Production deps (internal)** | None (root module)                                                                                                                                                                                                                                         |
| **Production deps (external)** | `zeebo/xxh3`, `a-h/templ`, `go/ast`, `go/parser`, `go/token`                                                                                                                                                                                               |
| **Test deps (external)**       | `onsi/ginkgo/v2`, `onsi/gomega`, `charmbracelet/x/exp/golden`                                                                                                                                                                                              |
| **Public API**                 | `suffixtree.STree`, `suffixtree.Match`, `syntax.Node`, `syntax.Match`, `syntax.Serialize()`, `syntax.FindSyntaxUnits()`, `config.Config`, `domain.*`, `errors.*`, `job.Parse()`, `job.BuildTree()`, `hash.*`                                               |

**Rationale:** These packages form a coherent "engine" — they handle parsing, tree building, and detection primitives. They share `syntax.Node` as the core data type and cannot be meaningfully separated without introducing a complex interface layer for `Node` (which would add overhead to every pipeline stage). Keeping them together avoids premature abstraction while still establishing boundaries against higher-level modules.

#### Module 2: `art-dupl-detection` (path: `./detection/`)

**Purpose:** Multi-method detection coordination — orchestrates art-dupl, hash, TODO, and legacy detectors.

| Field                          | Content                                                                                                   |
| ------------------------------ | --------------------------------------------------------------------------------------------------------- |
| **Module path**                | `github.com/LarsArtmann/art-dupl/detection`                                                               |
| **Contains packages**          | `detection/` (single package)                                                                             |
| **Production deps (internal)** | `art-dupl-core` (syntax, domain, config, hash, suffixtree, pkg/logger)                                    |
| **Production deps (external)** | None beyond what core provides                                                                            |
| **Public API**                 | `MultiDetector`, `MethodDetector` interface, `TodoDetector`, `LegacyDetector`, `TodoIssue`, `LegacyIssue` |

**Rationale:** Detection is a separate concern from parsing or output. It consumes the suffix tree and produces matches. Isolating it means detection methods can be added/modified without touching the core engine or output layer.

#### Module 3: `art-dupl-printer` (path: `./printer/`)

**Purpose:** All output formatting — text, HTML, JSON, SARIF, plumbing, statistics, sorting, and diff visualization.

| Field                          | Content                                                                                                                                       |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| **Module path**                | `github.com/LarsArtmann/art-dupl/printer`                                                                                                     |
| **Contains packages**          | `printer/` (single package)                                                                                                                   |
| **Production deps (internal)** | `art-dupl-core` (config, errors, syntax, syntax/golang, pkg/position)                                                                         |
| **Production deps (external)** | `sergi/go-diff`                                                                                                                               |
| **Public API**                 | `Printer` interface, `StatsPrinter` interface, `BuildCloneGroups()`, `SortCloneGroups()`, `ClassifyClone()`, all format-specific constructors |

**Rationale:** The printer is the god-package and the worst coupling offender (imports `syntax/golang` constants). Making it a separate module forces its public API to be explicit. Future work can split `syntax/golang` constant mapping into a registry pattern to break the language-specific coupling.

#### Module 4: `art-dupl-sdk` (path: `./pkg/artdupl/`)

**Purpose:** Public SDK for programmatic use — clean API surface wrapping internal pipeline.

| Field                          | Content                                                                                                              |
| ------------------------------ | -------------------------------------------------------------------------------------------------------------------- |
| **Module path**                | `github.com/LarsArtmann/art-dupl/pkg/artdupl`                                                                        |
| **Contains packages**          | `pkg/artdupl/` (single package)                                                                                      |
| **Production deps (internal)** | `art-dupl-core` (config, errors, pkg/logger, suffixtree, syntax, pkg/position), `art-dupl-detection`                 |
| **Test deps (internal)**       | `art-dupl-core` (internal/testutil)                                                                                  |
| **Production deps (external)** | None beyond what core provides                                                                                       |
| **Public API**                 | `Detector` interface, `NewDetector()`, `Result`, `CloneGroup`, `Clone`, `Options`, `Summary`, `Metadata`, `Progress` |

**Rationale:** The SDK already has a clean external API. Making it a module means consumers import only the SDK without pulling CLI, printer, or BDD dependencies. This is the most impactful split for external users.

#### Module 5: `art-dupl-cli` (path: `./cmd/`)

**Purpose:** CLI application — Cobra commands, flag parsing, and analysis orchestration.

| Field                          | Content                                                                                                                   |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| **Module path**                | `github.com/LarsArtmann/art-dupl/cmd`                                                                                     |
| **Contains packages**          | `cmd/`, `cmd/art-dupl/`                                                                                                   |
| **Production deps (internal)** | `art-dupl-core` (config, errors, syntax, suffixtree, hash, job, internal/utils), `art-dupl-detection`, `art-dupl-printer` |
| **Test deps (internal)**       | `art-dupl-core` (internal/testutil, config, job, printer, syntax, suffixtree)                                             |
| **Production deps (external)** | `spf13/cobra`, `charmbracelet/fang`, `charm.land/lipgloss/v2`, `LarsArtmann/gogenfilter`                                  |
| **Public API**                 | None (entry point only)                                                                                                   |

**Rationale:** The CLI is the top-level orchestrator. It's the only module that depends on everything else. Isolating it means CLI deps (cobra, fang, lipgloss) don't leak into the SDK or core.

### 3.2 Additional Directories (not modules)

| Directory              | Treatment                                                                 |
| ---------------------- | ------------------------------------------------------------------------- |
| `bdd/`                 | Stays in root module — tests cross all modules, needs `internal/testutil` |
| `examples/`            | Stays in root module — demo code, not a library                           |
| `internal/configtest/` | Stays in root module — test helper                                        |
| `internal/filtertest/` | Stays in root module — test helper                                        |

### 3.3 Dependency DAG

```
                    art-dupl-core (root)
                   /         |          \
                  v           v           v
     art-dupl-detection  art-dupl-printer  art-dupl-sdk
                  \           |           /
                   v          v          v
                      art-dupl-cli
```

**Cycle verification:**

- `core` → no internal deps ✅
- `detection` → `core` only ✅
- `printer` → `core` only ✅
- `sdk` → `core` + `detection` only ✅
- `cli` → `core` + `detection` + `printer` + `sdk` (optional) ✅
- No cycles detected ✅

---

## 4. Self-Review Findings

### 4.1 What did we forget?

1. **`internal/` package visibility is the #1 blocker.** Go enforces `internal` visibility at the **module boundary**, not the repo or workspace boundary. A sub-module with its own `go.mod` cannot import the root module's `internal/` packages — even in `_test.go` files, even with `go.work`.

2. **13 test files across all 5 proposed sub-modules import `internal/testutil`.** This is the most widespread issue.

3. **`internal/simd` is imported by `syntax/hash_simd.go` in production code.** Since syntax stays in the root module, this is actually fine — `internal/simd` stays internal to root. No promotion needed.

4. **`internal/utils` is imported by `cmd/run_flags.go` and `cmd/stats.go` in production code.** Since cmd becomes a separate module, `internal/utils` MUST be promoted before cmd can be a module.

### 4.2 Corrected pre-work requirements

| Package                | Current                | Target                | Affected Files                  | Why                                                                   |
| ---------------------- | ---------------------- | --------------------- | ------------------------------- | --------------------------------------------------------------------- |
| `internal/testutil`    | `internal/testutil`    | `testutil`            | 13 test files across 5 packages | Cross-module test visibility                                          |
| `internal/utils`       | `internal/utils`       | `pkg/utils`           | 2 production files in cmd       | Cross-module production visibility                                    |
| `internal/testhelpers` | `internal/testhelpers` | `testhelpers`         | 1 test file in suffixtree       | Consistency (root module, technically fine, but promotes consistency) |
| `internal/simd`        | `internal/simd`        | `internal/simd`       | 1 production file in syntax     | **NO CHANGE NEEDED** — syntax stays in root                           |
| `internal/configtest`  | `internal/configtest`  | `internal/configtest` | 0 imports                       | **NO CHANGE** — dead code, stays in root                              |
| `internal/filtertest`  | `internal/filtertest`  | `internal/filtertest` | 0 imports                       | **NO CHANGE** — dead code, stays in root                              |

### 4.3 Is the granularity right?

**Concern: Are 5 modules too many for a ~16K LOC project?**

Consideration: The alternative is fewer modules (e.g., 3: core, sdk, cli) with detection and printer staying in core. This reduces the `internal/` promotion scope but provides weaker boundary enforcement.

**Decision: Keep 5 modules.** The `internal/testutil` promotion is needed regardless of how many modules we create (even 3 would need it), and the 5-module split provides the clearest separation of concerns.

### 4.4 `internal/testhelpers` — keep or promote?

`suffixtree_test.go` imports `internal/testhelpers`. Since suffixtree stays in the root module, this import would still work after modularization. However, for consistency and to avoid confusion (developers expecting all test helpers to be public), promote it alongside `testutil`.

**Decision: Promote to `testhelpers/` for consistency.**

### 4.5 Existing code reuse

The `pkg/artdupl` SDK already provides the clean API boundary we want for the SDK module. No new interfaces needed — just a `go.mod`.

The `printer.Printer` interface already defines the contract. No new extraction needed.

### 4.6 Versioning reality check

The SDK (`pkg/artdupl`) has its own `Clone`, `CloneGroup`, `Result` types that are separate from `printer.Clone`, `printer.CloneGroup`. This is already a split brain documented in AGENTS.md. Modularization doesn't fix this — it's deferred to the Printer DTO refactor.

---

## 5. DAG Verification

Formal proof of acyclicity:

1. **core** (Layer 0): has zero internal module dependencies → cannot be part of a cycle
2. **detection** (Layer 1): depends only on core → cannot create a cycle (core has no upward deps)
3. **printer** (Layer 1): depends only on core → same argument
4. **sdk** (Layer 2): depends on core + detection → detection is Layer 1, core is Layer 0 → no upward path exists
5. **cli** (Layer 3): depends on core + detection + printer + sdk → all are Layer 0-2 → no upward path exists

Since every edge points from a higher layer to a strictly lower layer, the graph is a DAG.

---

## 6. Replace / Workspace Strategy

**Chosen: `go.work` at repo root**

| Factor                   | Decision                                                                   |
| ------------------------ | -------------------------------------------------------------------------- |
| All modules in same repo | Yes — `go.work` is ideal                                                   |
| Published to proxy       | No (internal tool) — no need for versioned imports                         |
| Module count             | 5 — above the threshold where `go.work` is cleaner than per-module replace |

### go.work file

```go
go 1.26.2

use (
    .
    ./detection
    ./printer
    ./pkg/artdupl
    ./cmd
)
```

**Rules:**

- No `replace` directives in any go.mod — `go.work` handles everything
- `go.work` is committed to the repo (NOT in .gitignore) since all modules live here
- Verify `go mod tidy` works both with and without workspace for the SDK module (external consumers)

### Per-module go.mod external dependencies

| Module        | External deps                                                                                                                      |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `core` (root) | `zeebo/xxh3`, `a-h/templ`, `charm.land/log/v2`, `charmbracelet/x/exp/golden` (test), `onsi/ginkgo/v2` (test), `onsi/gomega` (test) |
| `detection`   | None beyond core                                                                                                                   |
| `printer`     | `sergi/go-diff`                                                                                                                    |
| `sdk`         | None beyond core + detection                                                                                                       |
| `cli`         | `spf13/cobra`, `charmbracelet/fang`, `charm.land/lipgloss/v2`, `LarsArtmann/gogenfilter`                                           |

---

## 7. Test Dependency Isolation

| Module      | Production Deps                | Test-Only Deps                                             |
| ----------- | ------------------------------ | ---------------------------------------------------------- |
| `core`      | None (root)                    | All internal testutil/testhelpers                          |
| `detection` | `core`                         | `core` (testutil for integration tests)                    |
| `printer`   | `core`                         | `core` (testutil), `core` (syntax/golang for golden tests) |
| `sdk`       | `core`, `detection`            | `core` (testutil)                                          |
| `cli`       | `core`, `detection`, `printer` | `core` (testutil), `detection`, `printer`                  |

**Test helpers strategy:**

- `internal/testutil` stays in the root module (core) — it depends on `syntax`, `syntax/golang`, `internal/utils`
- `internal/testhelpers` stays in root — leaf package with no deps
- `internal/configtest` stays in root — test-only, depends on config
- `internal/filtertest` stays in root — test-only, depends on testutil
- BDD tests stay in root — they need access to all modules

Cross-module test dependencies are acceptable in `_test.go` files (imported via go.work). No production code may depend on test-only modules.

---

## 8. Interface Extraction Plan

### 7.1 Current state

| Module boundary            | Current interface                            | Needed                                         |
| -------------------------- | -------------------------------------------- | ---------------------------------------------- |
| `printer` → `syntax`       | Direct struct access (`*syntax.Node`)        | **No change yet** — too many call sites (111+) |
| `detection` → `suffixtree` | `suffixtree.Token` interface (already clean) | Already good                                   |
| `sdk` → `detection`        | `detection.MultiDetector` concrete type      | Acceptable for now                             |
| `cli` → `printer`          | `printer.Printer` interface                  | Already good                                   |

### 7.2 Future extraction (post-modularization)

**Priority 1: `syntax.ProcessedClone` DTO**

Replace `printer.PrintClones(dups [][]*syntax.Node)` with `PrintClones(clones []ProcessedCloneGroup)`. This eliminates the printer → syntax.Node dependency and unlocks:

- Printer module independent of AST internals
- SDK can provide its own Clone types without conversion
- Language-specific node type mapping moves to a registry in core

**This is deferred** — it touches 111 test call sites and requires careful coordination. The modularization itself creates the boundary; the DTO extraction is a separate improvement.

### 7.3 No new interfaces needed for initial split

The existing public API surfaces are sufficient. Each module's `go.mod` enforces that packages don't reach into other modules' internals.

---

## 9. Versioning Strategy

**Chosen: Root-only versioning**

| Factor        | Decision                                     |
| ------------- | -------------------------------------------- |
| Internal tool | Yes — no external consumers publish to proxy |
| Single team   | Yes — no need for independent versioning     |
| go.work usage | Yes — `go.work` + root tags is the simplest  |

**Strategy:**

- Only the root module gets git tags (`v1.2.3`)
- Sub-modules use `go.work` for local development
- No per-module semver tags — all modules bump together
- If the SDK (`pkg/artdupl`) is ever published as a standalone library, it can be extracted with its own tags at that point

**Tag format:** `v<major>.<minor>.<patch>` on root

---

## 10. Migration Strategy

### Phase 0: Pre-work — Promote internal packages (Pareto 1% → 51%)

1. Promote `internal/testutil/` → `testutil/` (update 13+ test file imports)
2. Promote `internal/testhelpers/` → `testhelpers/` (update 1 test file import)
3. Promote `internal/utils/` → `pkg/utils/` (update 2 production file imports + test imports)
4. Verify `go build ./...` and `go test ./...` pass

### Phase 1: Foundation (Pareto 1% → 51%)

5. Create `go.work` file at repo root
6. Create `detection/go.mod` — most self-contained module
7. Verify `go work sync` and `go build ./...`
8. Create `printer/go.mod`
9. Verify build

### Phase 2: SDK and CLI (Pareto 4% → 64%)

10. Create `pkg/artdupl/go.mod`
11. Verify SDK builds independently
12. Create `cmd/go.mod`
13. Verify CLI builds
14. Run full test suite

### Phase 3: CI and Cleanup (Pareto 20% → 80%)

15. Update `flake.nix` for multi-module build
16. Update CI workflows for per-module testing
17. Update documentation (README, AGENTS.md, HOW_TO_USE)
18. Final verification: `go work sync`, `go mod tidy` per module, full test suite

### Ordering principle

Each step leaves the project in a buildable, testable state. Each step is a single commit.

---

## 11. Risk Assessment

| Risk                                          | Likelihood | Impact   | Mitigation                                                                                            |
| --------------------------------------------- | ---------- | -------- | ----------------------------------------------------------------------------------------------------- |
| Import path breakage                          | Medium     | High     | Use `go.work` for development; verify `go mod tidy` per module                                        |
| Circular dependency discovered during split   | Low        | High     | DAG verified above; if found, merge modules                                                           |
| `internal/` package visibility                | Medium     | Medium   | `internal/` packages only visible within their module — move `internal/testutil` tests to root module |
| `flake.nix` complexity increases              | Medium     | Medium   | Keep root build aggregating sub-modules                                                               |
| External consumers of SDK broken              | Low        | Critical | SDK import path unchanged (`github.com/LarsArtmann/art-dupl/pkg/artdupl`)                             |
| Test dependencies leak into production go.mod | Low        | Medium   | Audit each module's go.mod after creation                                                             |
| BDD tests can't import across modules         | Medium     | Medium   | BDD stays in root module which uses go.work — all imports resolve                                     |

### Critical constraint: `internal/` package scoping

In Go, `internal/` packages are only visible to the module they belong to. Currently `detection`, `printer`, `cmd`, `pkg/artdupl` all import `internal/testutil` and `internal/utils`. If these packages move to separate modules, they **cannot** import the root module's `internal/` packages.

**Resolution:** Test-only cross-module imports use `go.work` resolution but the testutil package stays in the root module. For production code:

- `internal/utils` is imported only by `cmd` → it can be moved to `cmd/internal/utils` or promoted to a shared package
- `internal/simd` is imported only by `syntax` → stays in root (no issue)
- `internal/testutil` and `internal/testhelpers` → used only in `_test.go` files across modules; since test files in sub-modules can import the root module via `go.work`, this works. However, `internal/` visibility rules mean a sub-module's test files **cannot** import the root's `internal/` packages.

**Actual resolution needed (updated after self-review):**

1. `internal/utils` → promote to `pkg/utils` (used by cmd production code across module boundary)
2. `internal/testutil` → promote to `testutil/` (used by 13 test files across all proposed sub-modules)
3. `internal/testhelpers` → promote to `testhelpers/` (used by suffixtree tests; technically fine in root, but promoted for consistency)
4. `internal/simd` → stays `internal/simd` in root (only used by syntax, which stays in root)
5. `internal/configtest` → stays `internal/configtest` (dead code, only used by config tests in root)
6. `internal/filtertest` → stays `internal/filtertest` (dead code, only used by filter tests in root)

**This is the most critical pre-work before module creation.** Phase 0 in the migration plan covers this.

---

## 12. Build System Impact

### flake.nix changes

The current `flake.nix` uses `buildGoModule` for the root module. With multi-module:

1. Root `buildGoModule` continues to build the CLI binary
2. `vendorHash` must account for all sub-module dependencies
3. Consider `buildGoApplication` or per-module builds if needed
4. `nix flake check` should run `go work sync` + `go test ./...`

### Justfile changes

- `just build` → must work with `go.work`
- `just test` → `go work sync && go test ./...`
- `just check` → lint per module or root-level
- Add `just work-sync` convenience command

### CI changes

- GitHub Actions should run `go work sync` before build
- Per-module test jobs for parallel CI (optional, post-modularization)
- Coverage aggregation across modules

---

## Appendix A: Package-to-Module Mapping

| Package                 | Module                        | Rationale                                           |
| ----------------------- | ----------------------------- | --------------------------------------------------- |
| `domain/`               | core                          | Domain types, zero deps beyond errors               |
| `errors/`               | core                          | Foundation error types                              |
| `suffixtree/`           | core                          | Core algorithm, depends only on errors              |
| `syntax/`               | core                          | AST processing, depends on suffixtree               |
| `syntax/golang/`        | core                          | Go language parser                                  |
| `syntax/templ/`         | core                          | Templ language parser                               |
| `config/`               | core                          | Configuration used by all layers                    |
| `internal/simd/`        | core                          | SIMD optimizations for syntax                       |
| `internal/utils/`       | **promote to `pkg/utils/`**   | Used by cmd across module boundary                  |
| `internal/testhelpers/` | **promote to `testhelpers/`** | Used by suffixtree tests                            |
| `internal/testutil/`    | **promote to `testutil/`**    | Used by tests across modules                        |
| `internal/configtest/`  | core                          | Only used by config tests                           |
| `internal/filtertest/`  | core                          | Only used by filter tests                           |
| `pkg/format/`           | core                          | Utility package                                     |
| `pkg/logger/`           | core                          | Utility package                                     |
| `pkg/position/`         | core                          | Utility package                                     |
| `cache/`                | core                          | File caching for job                                |
| `hash/`                 | core                          | Hash-based detection (consumed by detection module) |
| `job/`                  | core                          | Parse + build orchestration                         |
| `detection/`            | **detection**                 | Multi-method detection coordination                 |
| `printer/`              | **printer**                   | Output formatting                                   |
| `pkg/artdupl/`          | **sdk**                       | Public SDK                                          |
| `cmd/`                  | **cli**                       | CLI application                                     |
| `cmd/art-dupl/`         | **cli**                       | CLI entry point                                     |
| `bdd/`                  | core                          | Cross-module BDD tests                              |
| `examples/`             | core                          | Demo code                                           |

## Appendix B: External Dependency Mapping

| External Dep                 | Used By                | Module  |
| ---------------------------- | ---------------------- | ------- |
| `charm.land/lipgloss/v2`     | cmd                    | cli     |
| `charm.land/log/v2`          | pkg/logger             | core    |
| `LarsArtmann/gogenfilter`    | cmd                    | cli     |
| `a-h/templ`                  | syntax/templ           | core    |
| `charmbracelet/fang`         | cmd/art-dupl           | cli     |
| `charmbracelet/x/exp/golden` | bdd, internal/testutil | core    |
| `onsi/ginkgo/v2`             | bdd, internal/testutil | core    |
| `onsi/gomega`                | bdd, internal/testutil | core    |
| `sergi/go-diff`              | printer                | printer |
| `spf13/cobra`                | cmd                    | cli     |
| `zeebo/xxh3`                 | syntax, hash           | core    |

---

## 10. Update Log — 2026-06-15

### Changes Since Original Proposal

- **`internal/simd/` deleted** — dead code removed (was listed in pre-work table, now N/A)
- **Go version updated** to 1.26.3 (was 1.26.2)
- **Package count**: 25 (was 26 — simd removed)
- **`.go-arch-lint.yml` fixed**: `detection → domain` dependency now declared; `internal/utils` mapped to `pkg-utils`
- **`MethodDetector` interface still unused** — hard-coded dispatch in `MultiDetector.FindDuplOver` remains the #1 composability defect
- **`printer/` grew** from 41 to 50 files, confirming it as the primary scalability concern

### Updated Recommendation

The original 5-module proposal remains sound. However, a **pragmatic alternative** worth considering:

- **Phase 1**: Extract only `pkg/artdupl` → `sdk/` module (minimal risk, immediate SDK consumer benefit)
- **Phase 2**: Full 5-module split once `MethodDetector` seam is activated and printer is split

This phased approach delivers value incrementally without the full `internal/` promotion scope upfront.
