# Actionability Phase 2 — Execution Plan

**Date:** 2026-05-16 21:13  
**Context:** Previous session added `CloneActionability` type, `EvaluateActionability` analyzer, `--rich-text` flag, and wired semantic mode filtering. Critical bugs discovered in post-session review.

---

## Pareto Breakdown

| Tier | % Work | % Result | What |
|---|---|---|---|
| **1%** | 3 fixes | **51%** | Populate Actionability field + fix text notation + fix gci |
| **4%** | 3 fixes | **64%** | Real error propagation + real defer detection + JSON serialization |
| **20%** | 8 tasks | **80%** | HTML badges + BDD tests + priority adjustment + docs |

---

## Phase 1: Critical Bugs (1% → 51%)

### 1.1 Populate Actionability in `ProcessClones`

**File:** `printer/clone_processor.go`  
**Bug:** `CloneClassification.Actionability` is never set — it's an empty string.

**Fix:** Call `EvaluateActionability(dups)` before the loop and set on every `ProcessedClone.Classification`.

**Why this matters:** Every printer that wants to display or filter by actionability needs this field populated. Currently it's dead code.

**Decision:** Keep `EvaluateActionability` as a group-level computation. `ClassifyClone` computes instance-level properties (category, priority); actionability is a group property.

### 1.2 Keep semantic filtering in `cmd` (performance)

**File:** `cmd/run_output.go`  
**Decision:** Do NOT remove semantic filtering from `cmd`. Filtering at `cmd` layer skips file I/O for non-actionable groups (performance win). Filtering at printer layer would require reading files first.

**Architecture:**
- `ProcessClones` → ALWAYS populates `Actionability` (data completeness)
- `cmd/printCloneGroups` → Filters when `semantic == true` (performance)
- `TextPrinter` → Can show `[non-actionable]` badge when `richText == true` AND `Actionability == NonActionable`

### 1.3 Fix text output line range notation

**File:** `printer/text.go`  
**Bug:** `writeCloneLines` uses `%d,%d` format (`store.go:97,103`). User explicitly called this "confusing" — they read it as "tokens 97-103" when it means "lines 97-103".

**Fix:** Change format to `%d-%d` (`store.go:97-103`). Also check `OutputText` for `97,103` patterns.

### 1.4 Fix gci import formatting

**File:** `printer/actionability_test.go`  
**Bug:** `gci` linter warning. Fix import grouping.

---

## Phase 2: Real Pattern Detection (4% → 64%)

### 2.1 Implement `isPureErrorPropagation`

**File:** `printer/actionability.go`  
**Current:** `len(seq) == 1 && seq[0].Type == golang.IfStmt`
**Real:** An error propagation match spans multiple nodes:
```
IfStmt
├── BinaryExpr (err != nil)
│   ├── Ident (err)
│   └── Ident (nil)
└── BlockStmt
    └── ReturnStmt
        └── Ident (err)
```
**Fix:** Inspect `seq[0].Children` for `BinaryExpr` (with `Ident/err` and `Ident/nil`) AND `ReturnStmt` child.

### 2.2 Implement `isPureDeferPattern`

**File:** `printer/actionability.go`  
**Current:** `len(seq) == 1 && seq[0].Type == golang.DeferStmt`
**Real:** Need to distinguish:
- `defer mu.Unlock()` → NonActionable (RAII)
- `defer expensiveCleanup()` → Actionable (could be extracted)
**Fix:** Inspect `DeferStmt.Children` for `CallExpr` → `SelectorExpr` with method names like `Unlock`, `RLock`, `Close`, `UnlockMutex`.

### 2.3 Add `Actionability` to JSON output

**File:** `printer/json.go`  
**Current:** `JSONClone` struct only has `Filename`, `LineStart`, `LineEnd`, `Fragment`.
**Fix:** Add `Actionability string` field.

---

## Phase 3: Output Enrichment (20% → 80%)

### 3.1 Add Actionability to text output in structural mode

**File:** `printer/text.go`  
When `--structural` + `--rich-text`: show `[HIGH] method [non-actionable]` badge. This answers the priority/actionability relationship question — keep orthogonal but make both visible.

### 3.2 Add Actionability to HTML report

**File:** `printer/html.go`, `html_summary.go`  
Add `[non-actionable]` badges to clone group headers and filter button.

### 3.3 Add Actionability to plumbing output

**File:** `printer/plumbing.go`  
Add comment line `# non-actionable` or similar when rich-text enabled.

### 3.4 BDD test for `--semantic` suppression

**File:** `bdd/` (new file)  
Test that `--semantic` suppresses interface signature clones while `--structural` shows them.

### 3.5 BDD test for `--rich-text`

**File:** `bdd/` (new file)  
Test that `--rich-text` produces enhanced output with badges.

### 3.6 Update AGENTS.md

Document actionability patterns and `--rich-text` flag.

### 3.7 Stats subcommand actionability

**Decision:** StatsPrinter already flows through `printCloneGroups` with semantic filtering. No extra wiring needed. But verify counts are correct (non-actionable groups are excluded from stats).

---

## Type Architecture Decisions

1. **Keep `Actionability` orthogonal to `Priority`** — Both fields on `CloneClassification`. Printers can display both.
2. **Group-level computation** — `EvaluateActionability` computes once per group. `ClassifyClone` computes per instance.
3. **Policy at `cmd`, data at `printer`** — `cmd` decides filtering based on `semantic` flag. `printer` provides the data.

---

## Execution Order

1. Fix gci (`actionability_test.go`) — unblock lint
2. Populate Actionability in `ProcessClones` — critical data fix
3. Fix text `,` → `-` notation — UX fix
4. Implement real error propagation detection
5. Implement real defer pattern detection
6. Add Actionability to JSON output
7. Add `[non-actionable]` badge to text output
8. BDD tests
9. HTML report update (deferred — large change)
10. Update AGENTS.md
11. Commit & push

---

## Verification Checklist

- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] `just check` / lint passes
- [ ] New tests cover actionability edge cases
- [ ] Manual: `./art-dupl --semantic --rich-text .` shows enhanced output
- [ ] Manual: `./art-dupl --structural --rich-text .` shows `[non-actionable]` badges
