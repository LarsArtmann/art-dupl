# Dependency Graph — art-dupl

**Date:** 2026-05-14
**Source:** `go mod graph` + import analysis of all .go files

---

## 1. Current State (Monolith)

### Layered View (bottom → top)

```
Layer 0 — Leaf packages (no internal deps):
┌──────────────────────┐
│ errors               │  Foundation error types
│ internal/simd        │  SIMD optimizations
│ internal/testhelpers │  Suffix tree test helpers
│ pkg/format           │  Output format utilities
│ pkg/logger           │  Logging (charm.land/log/v2)
│ pkg/position         │  Position types
└──────────────────────┘
         │
         ▼
Layer 1 — Domain & language parsers:
┌──────────────────────┐
│ domain → errors      │  Value objects (Filepath, LineNumber, CloneSeverity)
│ internal/utils → errors │ String/slice utilities
│ syntax/golang → syntax  │ Go AST parser (go/ast, go/parser, go/token)
│ syntax/templ → syntax   │ Templ parser (a-h/templ)
└──────────────────────┘
         │
         ▼
Layer 2 — Core algorithms:
┌──────────────────────┐
│ suffixtree → errors     │  Suffix tree data structure
│ syntax → suffixtree,    │  Unified AST (Node, Match, Serialize)
│          internal/simd, │
│          pkg/format     │
└──────────────────────┘
         │
         ▼
Layer 3 — Secondary algorithms:
┌──────────────────────┐
│ cache → errors, syntax        │  File-level AST caching (gob)
│ hash → syntax, pkg/format,    │  Rolling hash detection (xxh3)
│        pkg/logger             │
│ internal/testutil → syntax,   │  BDD test helpers (golden, ginkgo)
│                     syntax/   │
│                     golang,   │
│                     internal/ │
│                     utils     │
└──────────────────────┘
         │
         ▼
Layer 4 — Orchestration:
┌──────────────────────┐
│ job → syntax, cache,            │  Parse + BuildTree pipeline
│       pkg/logger, suffixtree,   │
│       syntax/golang,            │
│       syntax/templ              │
│ detection → syntax, domain,     │  Multi-method detection
│            config, hash,        │
│            pkg/logger,          │
│            suffixtree           │
└──────────────────────┘
         │
         ▼
Layer 5 — Output & SDK:
┌──────────────────────┐
│ printer → config, errors,        │  6 output formats + stats
│           syntax, syntax/golang, │  (sergi/go-diff)
│           pkg/position           │
│ pkg/artdupl → config, errors,   │  Public SDK
│              pkg/logger,         │
│              detection, job,     │
│              suffixtree, syntax, │
│              pkg/position        │
└──────────────────────┘
         │
         ▼
Layer 6 — CLI:
┌──────────────────────┐
│ cmd → config, errors, printer,  │  Cobra CLI + Fang framework
│        job, syntax, suffixtree, │  (spf13/cobra, charmbracelet/fang,
│        detection, hash,         │   charm.land/lipgloss,
│        internal/utils           │   LarsArtmann/gogenfilter)
└──────────────────────┘
         │
         ▼
Layer 7 — Entry points:
┌──────────────────────┐
│ cmd/art-dupl → cmd       │  Binary entry point
│ examples → pkg/artdupl,  │  Usage demos
│             pkg/logger   │
│ bdd → internal/testutil  │  BDD integration tests (ginkgo/gomega)
└──────────────────────┘
```

### Coupling Severity Matrix

| Coupling Point                         | Severity  | Packages Coupled                                               | Key Type                                        |
| -------------------------------------- | --------- | -------------------------------------------------------------- | ----------------------------------------------- |
| `syntax.Node` as pipeline data carrier | 🔴 TIGHT  | suffixtree, detection, job, printer, hash, cache, artdupl, cmd | `*syntax.Node`                                  |
| `syntax` → `suffixtree.Match`          | 🔴 TIGHT  | syntax ↔ suffixtree                                            | `suffixtree.Match` param in `FindSyntaxUnits()` |
| `printer` → `syntax/golang`            | 🔴 TIGHT  | printer → syntax/golang                                        | 50+ hardcoded node type constants               |
| `MultiDetector` switch/case            | 🟠 MEDIUM | detection → hash, suffixtree                                   | Unused `MethodDetector` interface               |
| `cmd` god orchestrator                 | 🟠 MEDIUM | cmd → 7+ packages                                              | Direct wiring, type assertions                  |
| `printer.Printer` interface            | 🟡 LOOSE  | 6 implementations                                              | Clean abstraction                               |
| `suffixtree.Token` interface           | 🟡 LOOSE  | suffixtree ← syntax.Node                                       | `Val() TokenValue`                              |
| `domain` value objects                 | 🟡 LOOSE  | detection, pkg/artdupl                                         | `Filepath`, `LineNumber`, `CloneSeverity`       |
| `pkg/artdupl` SDK API                  | 🟡 LOOSE  | external consumers                                             | `Detector` interface                            |

---

## 2. Proposed State (5 Modules + go.work)

### Module Dependency DAG

```
                ┌─────────────────────┐
                │   art-dupl-core     │
                │   (root module)     │
                │                     │
                │ domain, errors,     │
                │ suffixtree, syntax, │
                │ config, cache,      │
                │ hash, job,          │
                │ pkg/format,         │
                │ pkg/logger,         │
                │ pkg/position,       │
                │ testutil,           │
                │ testhelpers,        │
                │ bdd, examples       │
                └────────┬────────────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
          ▼              ▼              ▼
┌─────────────┐ ┌──────────────┐ ┌──────────────┐
│ detection   │ │   printer    │ │  pkg/artdupl │
│ (./detect.) │ │ (./printer/) │ │ (./pkg/..)   │
│             │ │              │ │              │
│ MultiDetect.│ │ 6 formats    │ │ SDK Detector │
│ MethodDetect│ │ stats, diff  │ │ Result,Clone │
│ TodoDetect. │ │ sort, class. │ │ Options      │
│ LegacyDetect│ │              │ │              │
└──────┬──────┘ └──────┬───────┘ └──────┬───────┘
       │               │                │
       │    ┌──────────┘                │
       │    │                           │
       ▼    ▼                           │
┌───────────────────────────────────────┐
│          cmd (./cmd/)                 │◀┘
│                                       │
│ CLI orchestration                     │
│ Cobra + Fang + gogenfilter            │
│ cmd/art-dupl binary entry point       │
└───────────────────────────────────────┘
```

### Module Detail Table

| Module        | Path             | Internal Deps            | External Deps                                                    | Packages                                 |
| ------------- | ---------------- | ------------------------ | ---------------------------------------------------------------- | ---------------------------------------- |
| **core**      | `./`             | None (root)              | `xxh3`, `a-h/templ`, `charm.land/log/v2`, `ginkgo/gomega` (test) | 17 production + 4 test + 2 entry         |
| **detection** | `./detection/`   | core                     | None                                                             | 1 package (5 prod files, 2 test files)   |
| **printer**   | `./printer/`     | core                     | `sergi/go-diff`                                                  | 1 package (21 prod files, 10 test files) |
| **sdk**       | `./pkg/artdupl/` | core, detection          | None                                                             | 1 package (8 prod files, 4 test files)   |
| **cli**       | `./cmd/`         | core, detection, printer | `cobra`, `fang`, `lipgloss/v2`, `gogenfilter`                    | 2 packages (15 prod files, 4 test files) |

### Import Flow After Modularization

```
External consumer:
  import "github.com/LarsArtmann/art-dupl/pkg/artdupl"
  → Only pulls: sdk → detection → core
  → Does NOT pull: printer, cli, cobra, fang, lipgloss, gogenfilter

CLI user:
  go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest
  → Pulls everything (expected for CLI)
```

---

## 3. Pre-work: Internal Package Promotions

These must happen BEFORE module creation due to Go's `internal/` visibility rules.

### 3.1 Production code promotions

| Package          | Current          | Target      | Files Affected                                  |
| ---------------- | ---------------- | ----------- | ----------------------------------------------- |
| `internal/utils` | `internal/utils` | `pkg/utils` | `cmd/run_flags.go`, `cmd/stats.go` (production) |

### 3.2 Test code promotions

| Package                | Current                | Target        | Files Affected                                                                     |
| ---------------------- | ---------------------- | ------------- | ---------------------------------------------------------------------------------- |
| `internal/testutil`    | `internal/testutil`    | `testutil`    | 13 test files across detection, printer, sdk, cmd, suffixtree, config, cache, hash |
| `internal/testhelpers` | `internal/testhelpers` | `testhelpers` | 1 test file: `suffixtree/suffixtree_test.go`                                       |

### 3.3 Packages staying internal

| Package               | Reason                                                        |
| --------------------- | ------------------------------------------------------------- |
| `internal/simd`       | Only used by `syntax/hash_simd.go` — both stay in root module |
| `internal/configtest` | Dead code — only used by config tests in root                 |
| `internal/filtertest` | Dead code — only used by filter tests in root                 |

---

## 4. External Dependency Distribution

| External Dep                 | Version   | Used By       | Module  | Production/Test |
| ---------------------------- | --------- | ------------- | ------- | --------------- |
| `charm.land/lipgloss/v2`     | v2.0.3    | cmd           | cli     | Production      |
| `charm.land/log/v2`          | v2.0.0    | pkg/logger    | core    | Production      |
| `LarsArtmann/gogenfilter`    | v0.2.1    | cmd           | cli     | Production      |
| `a-h/templ`                  | v0.3.1001 | syntax/templ  | core    | Production      |
| `charmbracelet/fang`         | v1.0.0    | cmd/art-dupl  | cli     | Production      |
| `charmbracelet/x/exp/golden` | v0.0.0    | bdd, testutil | core    | Test            |
| `onsi/ginkgo/v2`             | v2.28.3   | bdd, testutil | core    | Test            |
| `onsi/gomega`                | v1.40.0   | bdd, testutil | core    | Test            |
| `sergi/go-diff`              | v1.4.0    | printer       | printer | Production      |
| `spf13/cobra`                | v1.10.2   | cmd           | cli     | Production      |
| `zeebo/xxh3`                 | v1.1.0    | syntax, hash  | core    | Production      |
