# Execution Plan — art-dupl Modularization

**Date:** 2026-05-14
**Status:** Draft
**Prerequisite:** PROPOSAL.md and DEPENDENCY_GRAPH.md reviewed and approved

---

## Overview

18 tasks across 4 phases, ordered by Pareto impact. Each task:
- Takes 15–30 minutes
- Leaves the project in a buildable, testable state
- Is independently revertable (single commit)
- Has clear verification steps

---

## Phase 0 — Pre-work: Promote Internal Packages

**Impact: 1% → 51%** — Without this, no sub-modules can be created.

### Task 0.1: Promote `internal/utils` → `pkg/utils`

**What:** Move `internal/utils/` directory to `pkg/utils/`, update all import paths.

**Why:** `cmd/` (which becomes a separate module) imports `internal/utils` in production code (`run_flags.go`, `stats.go`). Go's `internal/` visibility rules prevent cross-module access.

**Steps:**
1. `git mv internal/utils pkg/utils`
2. Update all imports: `github.com/LarsArtmann/art-dupl/internal/utils` → `github.com/LarsArtmann/art-dupl/pkg/utils`
3. Affected files:
   - `cmd/run_flags.go` (production)
   - `cmd/stats.go` (production)
   - `internal/testutil/bdd.go` (test — imports internal/utils)
   - `internal/utils/*_test.go` (3 test files)
   - Any other files found by `grep -r "internal/utils" --include="*.go"`

**Verification:**
```bash
go build ./...
go test ./...
go vet ./...
```

**Rollback:** `git revert HEAD`

**Effort:** 15 min

---

### Task 0.2: Promote `internal/testutil` → `testutil`

**What:** Move `internal/testutil/` directory to `testutil/`, update all import paths.

**Why:** 13 test files across all 5 proposed sub-modules import `internal/testutil`. Cross-module `internal/` visibility prevents this.

**Steps:**
1. `git mv internal/testutil testutil`
2. Update all imports: `github.com/LarsArtmann/art-dupl/internal/testutil` → `github.com/LarsArtmann/art-dupl/testutil`
3. Affected files (test only):
   - `bdd/*.go` (19+ files)
   - `detection/detection_test.go`
   - `printer/sarif_printer_test.go`, `stats_golden_test.go`, `stats_test.go`, `diff_test.go`, `json_printer_test.go`, `issuer_test.go`, `html_golden_test.go`
   - `pkg/artdupl/detector_uncovered_test.go`, `detector_types_test.go`
   - `cmd/cmd_utils_test.go`, `stats_integration_test.go`
   - `config/config_test.go`
   - `cache/cache_test.go`
   - `hash/hash_test.go`
   - Any other files found by `grep -r "internal/testutil" --include="*.go"`

**Verification:**
```bash
go build ./...
go test ./...
go vet ./...
```

**Rollback:** `git revert HEAD`

**Effort:** 20 min

---

### Task 0.3: Promote `internal/testhelpers` → `testhelpers`

**What:** Move `internal/testhelpers/` directory to `testhelpers/`, update all import paths.

**Why:** Consistency with testutil promotion. Technically only used by `suffixtree_test.go` which stays in root, but promoting avoids confusion about which test helpers are internal vs shared.

**Steps:**
1. `git mv internal/testhelpers testhelpers`
2. Update imports: `github.com/LarsArtmann/art-dupl/internal/testhelpers` → `github.com/LarsArtmann/art-dupl/testhelpers`
3. Affected files:
   - `suffixtree/suffixtree_test.go`

**Verification:**
```bash
go build ./...
go test ./...
go vet ./...
```

**Rollback:** `git revert HEAD`

**Effort:** 10 min

---

### Task 0.4: Clean up empty `internal/` directory

**What:** Remove the `internal/` directory if only `simd`, `configtest`, and `filtertest` remain.

**Why:** These 3 packages stay internal (only used by root module code). The directory stays.

**Steps:**
1. Verify `internal/` contains only: `simd/`, `configtest/`, `filtertest/`
2. No action needed — they stay

**Verification:** `ls internal/`

**Effort:** 2 min

---

## Phase 1 — Foundation: Create go.work and First Sub-Module

**Impact: 1% → 51%** — Establishes the multi-module pattern.

### Task 1.1: Create `go.work` file

**What:** Create a `go.work` file at the repo root.

**Why:** Required for multi-module development. Coordinates all sub-modules.

**Steps:**
1. Create `go.work`:
   ```go
   go 1.26.2

   use (
       .
   )
   ```
   (Start with only the root module — add sub-modules as they're created)

**Verification:**
```bash
go work sync
go build ./...
go test ./...
```

**Rollback:** `git rm go.work`

**Effort:** 5 min

---

### Task 1.2: Create `detection/go.mod`

**What:** Create a `go.mod` file for the detection module.

**Why:** Detection is the most self-contained sub-module — it has no production dependencies beyond core packages.

**Steps:**
1. Create `detection/go.mod`:
   ```
   module github.com/LarsArtmann/art-dupl/detection

   go 1.26.2

   require github.com/LarsArtmann/art-dupl v0.0.0
   ```
2. Add `detection` to `go.work` use block
3. Run `go work sync`
4. Run `go mod tidy` in `detection/` to populate dependencies
5. If needed, set `vendorHash` to empty and rebuild for nix

**Dependencies (production):**
- `github.com/LarsArtmann/art-dupl` → syntax, domain, config, hash, pkg/logger, suffixtree

**Dependencies (test):**
- `github.com/LarsArtmann/art-dupl` → testutil

**Verification:**
```bash
cd detection && go build ./... && cd ..
go work sync
go build ./...
go test ./...
```

**Rollback:** Remove `detection/go.mod`, revert `go.work`, `git checkout -- go.mod go.sum`

**Effort:** 20 min

---

### Task 1.3: Create `printer/go.mod`

**What:** Create a `go.mod` file for the printer module.

**Why:** Printer is the god-package that most benefits from module boundary enforcement.

**Steps:**
1. Create `printer/go.mod`:
   ```
   module github.com/LarsArtmann/art-dupl/printer

   go 1.26.2

   require (
       github.com/LarsArtmann/art-dupl v0.0.0
       github.com/sergi/go-diff v1.4.0
   )
   ```
2. Add `printer` to `go.work` use block
3. Run `go work sync`
4. Run `go mod tidy` in `printer/`

**Dependencies (production):**
- `github.com/LarsArtmann/art-dupl` → config, errors, syntax, syntax/golang, pkg/position
- `github.com/sergi/go-diff` → diff visualization

**Dependencies (test):**
- `github.com/LarsArtmann/art-dupl` → testutil

**Verification:**
```bash
cd printer && go build ./... && cd ..
go work sync
go build ./...
go test ./...
```

**Rollback:** Remove `printer/go.mod`, revert `go.work`, `git checkout -- go.mod go.sum`

**Effort:** 25 min

---

## Phase 2 — SDK and CLI Modules

**Impact: 4% → 64%** — Completes the module structure.

### Task 2.1: Create `pkg/artdupl/go.mod`

**What:** Create a `go.mod` file for the SDK module.

**Why:** The SDK is the public API for programmatic use. Making it a separate module means consumers don't pull CLI/printer dependencies.

**Steps:**
1. Create `pkg/artdupl/go.mod`:
   ```
   module github.com/LarsArtmann/art-dupl/pkg/artdupl

   go 1.26.2

   require (
       github.com/LarsArtmann/art-dupl v0.0.0
       github.com/LarsArtmann/art-dupl/detection v0.0.0
   )
   ```
2. Add `pkg/artdupl` to `go.work` use block
3. Run `go work sync`
4. Run `go mod tidy` in `pkg/artdupl/`

**Dependencies (production):**
- `github.com/LarsArtmann/art-dupl` → config, errors, pkg/logger, suffixtree, syntax, pkg/position, job
- `github.com/LarsArtmann/art-dupl/detection` → MultiDetector

**Dependencies (test):**
- `github.com/LarsArtmann/art-dupl` → testutil

**Verification:**
```bash
cd pkg/artdupl && go build ./... && cd ..
go work sync
go build ./...
go test ./...
```

**Rollback:** Remove `pkg/artdupl/go.mod`, revert `go.work`, `git checkout -- go.mod go.sum`

**Effort:** 25 min

---

### Task 2.2: Create `cmd/go.mod`

**What:** Create a `go.mod` file for the CLI module.

**Why:** The CLI is the top-level orchestrator. Isolating it prevents CLI deps (cobra, fang, lipgloss) from leaking into other modules.

**Steps:**
1. Create `cmd/go.mod`:
   ```
   module github.com/LarsArtmann/art-dupl/cmd

   go 1.26.2

   require (
       github.com/LarsArtmann/art-dupl v0.0.0
       github.com/LarsArtmann/art-dupl/detection v0.0.0
       github.com/LarsArtmann/art-dupl/printer v0.0.0
       github.com/LarsArtmann/art-dupl/pkg/artdupl v0.0.0
       github.com/spf13/cobra v1.10.2
       github.com/charmbracelet/fang v1.0.0
       charm.land/lipgloss/v2 v2.0.3
       github.com/LarsArtmann/gogenfilter v0.2.1-0.20260504180622-235fb88077c7
   )
   ```
2. Add `cmd` to `go.work` use block
3. Run `go work sync`
4. Run `go mod tidy` in `cmd/`

**Dependencies (production):**
- `github.com/LarsArtmann/art-dupl` → config, errors, syntax, suffixtree, hash, job, pkg/utils
- `github.com/LarsArtmann/art-dupl/detection` → MultiDetector
- `github.com/LarsArtmann/art-dupl/printer` → Printer, BuildCloneGroups, etc.
- `github.com/LarsArtmann/art-dupl/pkg/artdupl` → (if needed for SDK subcommand)

**Dependencies (test):**
- `github.com/LarsArtmann/art-dupl` → testutil

**Verification:**
```bash
cd cmd && go build ./... && cd ..
go work sync
go build ./...
go test ./...
# Verify CLI binary still works
go run ./cmd/art-dupl --help
```

**Rollback:** Remove `cmd/go.mod`, revert `go.work`, `git checkout -- go.mod go.sum`

**Effort:** 30 min

---

### Task 2.3: Full test suite verification

**What:** Run the complete test suite with all modules active.

**Why:** Ensure no regressions from the modularization.

**Steps:**
1. `go work sync`
2. `go build ./...`
3. `go test -race ./...`
4. `go vet ./...`
5. Verify `go mod tidy` is clean in each module:
   ```bash
   for dir in . detection printer pkg/artdupl cmd; do
     echo "=== $dir ==="
     (cd $dir && go mod tidy && git diff --stat go.mod go.sum)
   done
   ```
6. Run BDD tests: `go test -v ./bdd`
7. Run integration tests: `go test -v ./internal/configtest ./internal/filtertest`

**Verification:** All green.

**Rollback:** Revert to commit before any module creation.

**Effort:** 15 min

---

## Phase 3 — CI and Documentation

**Impact: 20% → 80%** — Makes the modularization sustainable.

### Task 3.1: Update `flake.nix` for multi-module build

**What:** Update the Nix flake to build with `go.work`.

**Why:** The flake currently uses `buildGoModule` with `vendorHash`. Multi-module requires adjusting the build to handle `go.work`.

**Steps:**
1. Update `vendorHash` — set to empty string, run `nix build`, copy correct hash
2. Verify `nix build` still produces a working binary
3. Verify `nix run .#test` still passes
4. Verify `nix flake check` passes

**Verification:**
```bash
nix build
nix run .#test
nix flake check
```

**Rollback:** `git checkout -- flake.nix`

**Effort:** 30 min

---

### Task 3.2: Update justfile for multi-module

**What:** Update justfile commands to work with `go.work`.

**Steps:**
1. Ensure `just build` runs `go work sync` before build
2. Ensure `just test` runs `go work sync` before tests
3. Add `just work-sync` convenience command
4. Verify `just ci` passes

**Verification:**
```bash
just ci
```

**Rollback:** `git checkout -- justfile`

**Effort:** 15 min

---

### Task 3.3: Update CI workflows

**What:** Update GitHub Actions to handle multi-module.

**Steps:**
1. Add `go work sync` step before build/test in all workflows
2. Add per-module build verification (optional)
3. Verify matrix builds still work

**Verification:** Push to branch, verify CI passes.

**Rollback:** `git checkout -- .github/workflows/`

**Effort:** 20 min

---

### Task 3.4: Update README.md

**What:** Document the multi-module structure in README.

**Steps:**
1. Add "Project Structure" section explaining the 5 modules
2. Update build instructions to mention `go.work`
3. Update SDK import path documentation (unchanged, but clarify)
4. Add "Development" section with `go work sync` instructions

**Verification:** Read-through.

**Rollback:** `git checkout -- README.md`

**Effort:** 15 min

---

### Task 3.5: Update AGENTS.md

**What:** Update the project's AGENTS.md with modularization info.

**Steps:**
1. Add "Module Structure" section with the 5 modules
2. Update build/test commands section
3. Add `go.work` instructions
4. Update "Architecture Highlights" section
5. Document the internal → public package promotions

**Verification:** Read-through.

**Rollback:** `git checkout -- AGENTS.md`

**Effort:** 15 min

---

### Task 3.6: Final cleanup and verification

**What:** Final sweep to ensure everything is clean.

**Steps:**
1. `go work sync`
2. `go mod tidy` in each module directory
3. `go mod verify` in each module directory
4. `go build ./...`
5. `go test -race ./...`
6. `go vet ./...`
7. `just ci`
8. `nix build`
9. Verify no stale `go.sum` entries
10. Verify `.gitignore` doesn't exclude `go.work`

**Verification:** All green.

**Effort:** 15 min

---

## Task Dependency Graph

```
Phase 0 (pre-work):
  0.1 promote internal/utils  ──┐
  0.2 promote internal/testutil ─┤
  0.3 promote internal/testhelp.─┤
  0.4 verify internal/ remains  ─┘
                                  │
Phase 1 (foundation):            │
  1.1 create go.work  ←──────────┤
  1.2 create detection/go.mod ←──┤ (depends on 0.2)
  1.3 create printer/go.mod  ←──┤ (depends on 0.2)
                                  │
Phase 2 (modules):               │
  2.1 create pkg/artdupl/go.mod ←┤ (depends on 1.2)
  2.2 create cmd/go.mod  ←──────┤ (depends on 1.2, 1.3, 2.1)
  2.3 full test suite       ←───┘ (depends on all above)
                                  │
Phase 3 (CI + docs):             │
  3.1 update flake.nix            │ (depends on 2.3)
  3.2 update justfile             │ (depends on 2.3)
  3.3 update CI workflows         │ (depends on 2.3)
  3.4 update README.md            │ (depends on 2.3)
  3.5 update AGENTS.md            │ (depends on 2.3)
  3.6 final verification          │ (depends on 3.1-3.5)
```

---

## Estimated Total Effort

| Phase | Tasks | Effort |
|---|---|---|
| Phase 0 | 4 tasks | ~47 min |
| Phase 1 | 3 tasks | ~50 min |
| Phase 2 | 3 tasks | ~70 min |
| Phase 3 | 6 tasks | ~110 min |
| **Total** | **16 tasks** | **~4.5 hours** |

---

## Risk Mitigation Per Task

| Task | Risk | Mitigation |
|---|---|---|
| 0.1-0.3 | Import path miss | Use `grep -r` to find ALL occurrences before editing |
| 1.2 | detection has hidden deps | `go mod tidy` will surface them; add as needed |
| 1.3 | printer → syntax/golang coupling | Accepted — printer go.mod depends on core which includes syntax/golang |
| 2.2 | cmd imports everything | Expected — cmd is the top-level orchestrator |
| 3.1 | flake.nix vendor hash changes | Use dummy hash → build → copy correct hash pattern |
| All | Test failures | Fix immediately or revert — never accumulate broken state |

---

## Commit Messages

| Task | Commit Message |
|---|---|
| 0.1 | `refactor: promote internal/utils to pkg/utils for cross-module access` |
| 0.2 | `refactor: promote internal/testutil to testutil for cross-module test access` |
| 0.3 | `refactor: promote internal/testhelpers to testhelpers for consistency` |
| 1.1 | `build: add go.work for multi-module development` |
| 1.2 | `build: create detection sub-module with own go.mod` |
| 1.3 | `build: create printer sub-module with own go.mod` |
| 2.1 | `build: create pkg/artdupl (SDK) sub-module with own go.mod` |
| 2.2 | `build: create cmd (CLI) sub-module with own go.mod` |
| 2.3 | `test: verify full test suite passes with all modules` |
| 3.1 | `build(nix): update flake.nix for multi-module go.work build` |
| 3.2 | `build: update justfile for multi-module go.work commands` |
| 3.3 | `ci: update GitHub Actions for multi-module go.work` |
| 3.4 | `docs: update README with multi-module structure` |
| 3.5 | `docs: update AGENTS.md with modularization details` |
| 3.6 | `chore: final cleanup and verification of multi-module setup` |
