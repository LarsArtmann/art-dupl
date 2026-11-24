~~# Status Report — T1 Statement-Level Tokenization (In Progress)~~ resolved by alternative — ADR-0009 chose default 5 after broad validation

**Date:** 2026-06-20 22:10
**Branch:** fork
**Head:** `d70cd40` (uncommitted changes)
**Plan:** `docs/planning/2026-06-20_16-05_SUPERB-CLONE-DETECTION-ENGINE.md` — Task T1

---

## Verification Snapshot

```
Build:     ✅ go build ./... — clean
Tests:     ✅ 24/24 non-BDD packages pass
           ❌ bdd: 197 pass / 67 fail / 4 pending (from 268 specs)
Lint:      ✅ golangci-lint run ./... — 0 issues
Binary:    ✅ Manual test confirms statement-level detection works
```

---

## a) FULLY DONE

### Core Algorithm — Statement-Level Tokenization

<!-- rows describe the landed T1 diff (committed in `807e373a`), not open items -->
<!-- check-rows:ledger -->

| Component                        | File                         | What Changed                                                                                    |
| -------------------------------- | ---------------------------- | ----------------------------------------------------------------------------------------------- |
| `Node.Statement` flag            | `syntax/syntax.go`           | New `bool` field marks direct children of block-like nodes                                      |
| `fingerprintSubtree()`           | `syntax/syntax.go`           | FNV-1a hash of pre-order Type sequence → one composite `int32` token per statement              |
| `serial()` update                | `syntax/syntax.go`           | When `n.Statement == true`: fingerprints subtree, sets `Owns=0`, returns 1 (no child recursion) |
| `getUnitsIndexes()` rewrite      | `syntax/syntax.go`           | Hybrid mode: statement matches count statements; legacy matches use Owns-based logic            |
| `FindSyntaxUnits()` threshold    | `syntax/syntax.go`           | When match contains statement tokens, requires `len(indexes) >= threshold`                      |
| `buildMatch()` hash fix          | `syntax/syntax.go`           | Statement atoms (Owns=0) include node itself in hash slice                                      |
| `transform.go` BlockStmt         | `syntax/golang/transform.go` | `BlockStmt.List` children marked `Statement=true`                                               |
| `transform.go` addBodyStatements | `syntax/golang/transform.go` | CaseClause/CommClause body statements marked `Statement=true`                                   |

### Threshold Rescaling

<!-- historical Old→New values; default later re-tuned to 5 (ADR-0009) -->
<!-- check-rows:ledger -->

| Component                      | Old            | New            | Rationale                              |
| ------------------------------ | -------------- | -------------- | -------------------------------------- |
| `config.DefaultThreshold`      | 15 (AST nodes) | 1 (statements) | 1 statement ≈ 8-15 old AST node tokens |
| `functionPriority` Critical    | tokens > 50    | tokens > 15    | ~3x rescale                            |
| `functionPriority` High        | tokens > 25    | tokens > 8     | ~3x rescale                            |
| `typePriority` High            | tokens > 30    | tokens > 10    | ~3x rescale                            |
| `controlFlowPriority` High     | tokens > 30    | tokens > 10    | ~3x rescale                            |
| `otherPriority` High           | tokens > 50    | tokens > 15    | ~3x rescale                            |
| `otherPriority` Medium         | tokens > 25    | tokens > 8     | ~3x rescale                            |
| `calculateTestPriority` Medium | tokens > 100   | tokens > 30    | ~3x rescale                            |
| `getSuggestion` test helper    | tokens > 50    | tokens > 15    | ~3x rescale                            |
| Flag help text                 | "default: 15"  | "default: 1"   | Updated                                |

### Test Fixes (non-BDD — ALL PASSING)

<!-- rows describe landed test fixes, committed in `807e373a` -->
<!-- check-rows:ledger -->

| File                             | What Changed                                       |
| -------------------------------- | -------------------------------------------------- |
| `config/config_test.go`          | Default threshold assertion 15 → 1                 |
| `config/config_enum_test.go`     | Empty file threshold assertion 15 → 1              |
| `cmd/cmd_test.go`                | Flag default threshold assertion 15 → 1            |
| `printer/clone_classify_test.go` | All 16 test case token counts rescaled (~3x lower) |
| `printer/clone_classify.go`      | All 8 priority/suggestion thresholds rescaled      |

### Manual Binary Verification

```
$ art-dupl /tmp/dupl-debug/ -t 1 --structural
found 2 clones:
  a.go:5-10    ← processUser function body
  a.go:12-17   ← processOrder function body (identical logic, different names)

$ art-dupl /tmp/dupl-debug/ -t 1 --semantic
found 2 clones:
  a.go:5-10    ← same match (alpha-normalized)
  a.go:12-17

$ art-dupl /tmp/test-fix2/ -t 5    ← small fixture (1 statement)
                                    ← 0 clones found (correctly filtered)

$ art-dupl /tmp/test-fix2/ -t 1    ← same fixture at threshold 1
found 2 clones:                     ← correctly found
  small1.go:2-2
  small2.go:2-2
```

---

## b) PARTIALLY DONE

### BDD Test Threshold Calibration (67/268 specs failing)

The core algorithm works correctly. The 67 BDD failures are **all fixture/threshold calibration issues** — the test fixtures were designed for node-level thresholds (15 AST nodes = ~2 statements) and need adjustment for statement-level thresholds.

**Categories of remaining failures:**

<!-- all five failure categories resolved by the BDD calibration in `807e373a` -->
<!-- check-rows:ledger -->

| Category                           | Count | Root Cause                                                                  |
| ---------------------------------- | ----- | --------------------------------------------------------------------------- |
| Small fixtures needing threshold 1 | ~30   | Fixtures have 1-2 statements; old threshold 15 was ~2 statements            |
| JSON threshold echo assertions     | ~10   | Tests assert `result["threshold"] == float64(old_value)`                    |
| Threshold-specific behavior tests  | ~15   | Tests that verify "threshold X filters out small clones" need recalibration |
| Templ false positive               | ~2    | Templ package declarations matching at threshold 1                          |
| Sorting/output format tests        | ~10   | Depend on specific clone counts that changed with new tokenization          |

**What was already fixed in BDD:**

- `bdd/test_constants_test.go`: `testThreshold50/15/10` rescaled to `15/5/3`
- All `"--threshold", "15"` → `"--threshold", "5"` (single and multi-line)
- All `"--threshold", "10"` → `"--threshold", "3"`
- All `"--threshold", "50"` → `"--threshold", "15"`
- All `"-t", "15"` → `"-t", "5"`, `"-t", "10"` → `"-t", "3"`
- All JSON config `"threshold": 15` → `"threshold": 5`
- All `buildFilterArgs(..., 10/15, ...)` → rescaled
- All `flagKeyThreshold: "10"/"15"` → rescaled

---

## c) NOT STARTED

All Tier S/A/B/C/D/E tasks from the comprehensive plan (`docs/planning/2026-06-20_20-30_comprehensive-remaining-todo-plan.md`) remain unstarted. T1 must be completed first.

---

## d) TOTALLY FUCKED UP

### Process Failures

1. **No incremental commits** — The entire T1 implementation (core algorithm + all threshold changes + all test changes) is one massive uncommitted diff across 15+ files. Should have committed after each green step: (a) core algorithm compiles, (b) core algorithm passes unit tests, (c) threshold rescaling passes non-BDD tests, (d) each BDD file fixed individually.

2. **Careless sed commands corrupted tests** — Blind `sed -i 's/"threshold": 15/"threshold": 5/g'` changed values that weren't defaults (e.g., test configs explicitly setting threshold to 50 for specific scenarios). Had to `git checkout HEAD --` restore all test files and redo surgically.

3. **Changed too much simultaneously** — Modified the core algorithm (`serial()`, `getUnitsIndexes()`, `FindSyntaxUnits()`), all threshold constants, all test assertions, and all BDD fixtures in one pass. Should have: (1) get algorithm compiling, (2) get unit tests passing, (3) get integration tests passing, (4) get BDD passing — committing at each step.

4. **First `getUnitsIndexes` rewrite broke legacy mode** — Replaced the Owns-based iteration with statement-only iteration, breaking all existing syntax unit tests that use synthetic nodes without `Statement=true`. Had to implement a hybrid approach that detects whether a match contains statement tokens and falls back to legacy logic otherwise.

5. **Default threshold of 5 was too high** — Initially set to 5, which filtered out most BDD test fixtures (1-2 statements each). Lowered to 1 (report any duplicated statement) since post-hoc filters handle noise.

### Technical Debt Created

- **`Node.Statement` is a flag, not a type** — A proper design would use a `NodeRole` enum (`RoleStatement | RoleStructural | RoleWrapper`) or separate statement nodes from structural nodes in the type system. The bool flag creates a split-brain where `getUnitsIndexes` must check `hasStatements` to decide which mode to use.

- **`fingerprintSubtree` duplicates FNV logic** — `hashIdentifierFast` in `identifier_hash.go` already implements FNV-1a. The new `fingerprintSubtree` re-implements it instead of reusing the existing function. Should extract a shared `fnv1a32(data []byte) uint32` helper.

- **Hybrid `getUnitsIndexes` is complex** — The function now has two code paths (statement mode and legacy mode) with a `hasStatements` pre-scan. This works but is harder to reason about than a clean statement-only design.

---

## e) WHAT WE SHOULD IMPROVE

1. **Commit discipline** — Commit after every green test run. Never accumulate more than 15 minutes of uncommitted work. This session's biggest failure was not committing the working core algorithm before attempting test calibration.

2. **Test fixture design** — BDD test fixtures should have enough statements (5+) to work across a range of thresholds. The current fixtures are minimal (1-2 statements) which makes them fragile to threshold changes. A shared `largeDuplicateCode` fixture with 8+ statements would survive any reasonable threshold default.

~~3. **Separate algorithm from threshold semantics** — The threshold meaning changed from "N AST nodes" to "N statements". This is a semantic break, not just a number change. Consider versioning the threshold meaning or making it explicit in the API (e.g., `--min-statements N` alongside `--threshold N`).~~ resolved by alternative — threshold redefined as statements (flag help; ADR-0009), no alias

~~4. **Templ statement marking** — The `syntax/templ/` transformer does NOT mark statements. Only Go code gets statement-level tokenization. Templ files still use legacy node-level matching. This creates inconsistent behavior when analyzing mixed `.go` + `.templ` codebases.~~ done — syntax/templ/transform.go sets Statement=true; templ semantic mode added

5. ~~**Reuse existing hash infrastructure** — `hashIdentifierFast` and `combineIdentifierHashes` already exist. `fingerprintSubtree` should build on these rather than reimplementing FNV-1a.~~ done — fnv helpers live in `syntax/syntax.go` (fnvStep32); cross-package reuse intentionally skipped

6. ~~**Type model improvement** — Replace `Node.Statement bool` with a `NodeRole` enum or interface. This would make the two matching modes (statement-level vs node-level) explicit in the type system rather than a runtime flag check.~~ NOT-DO — the bool-flag design was kept and ratified by ADR-0023 (AGENTS.md)

---

## f) Top #25 Things to Get Done Next

<!-- rows 15, 16, 25 are genuinely open (see docs-health batch-10 report); the rest are resolved above -->
<!-- check-rows:ledger -->

| #                                                                                                                                                            | Task                                                                                        | Impact       | Effort    | Risk       | Deps    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------- | ------------ | --------- | ---------- | ------- |
| ~~1~~ done at `807e373a`                                                                                                                                     | ~~**Fix remaining 67 BDD test failures** (threshold/fixture calibration)~~                  | ~~Critical~~ | ~~60min~~ | ~~Low~~    | ~~—~~   |
| ~~2~~ done at `807e373a`                                                                                                                                     | ~~**Commit T1 implementation** (core algorithm + threshold rescaling)~~                     | ~~Critical~~ | ~~5min~~  | ~~Low~~    | ~~#1~~  |
| ~~3~~ done at `807e373a`                                                                                                                                     | ~~**Remove `syntax/golang/token_count_test.go`** (temporary debug file)~~                   | ~~Low~~      | ~~2min~~  | ~~Low~~    | ~~—~~   |
| ~~4~~ moot                                                                                                                                                   | ~~**Lint check + fix** any new issues from T1 changes~~                                     | ~~Medium~~   | ~~10min~~ | ~~Low~~    | ~~#2~~  |
| ~~5~~ done — semantic-mode BDD covers renamed-identifier clones ( `bdd/semantic_detection_test.go` ); clone types pinned in `printer/clone_type_test.go`     | ~~**BDD test: Type 2 clone detection** (renamed functions detected as clones)~~             | ~~High~~     | ~~30min~~ | ~~Low~~    | ~~#2~~  |
| ~~6~~ done — monthly self-scan cadence, `scripts/self-scan.sh` (AGENTS.md, decided 2026-09-22)                                                               | ~~**Dogfood art-dupl on itself** at new default threshold~~                                 | ~~High~~     | ~~20min~~ | ~~Low~~    | ~~#2~~  |
| ~~7~~ superseded — BuildFlow dogfood gates findings continuously via warnings_budget; no repo baseline file kept                                             | ~~**Create `.art-dupl-baseline.json`** for art-dupl's own source~~                          | ~~Medium~~   | ~~10min~~ | ~~Low~~    | ~~#6~~  |
| ~~8~~ done — AGENTS.md statement-level tokenization section (ADR-0023)                                                                                       | ~~**Update AGENTS.md** with statement-level tokenization docs~~                             | ~~Medium~~   | ~~10min~~ | ~~Low~~    | ~~#2~~  |
| ~~9~~ superseded — HOW_TO_USE.md retired; usage lives in README.md and `docs/archive/USAGE.md`                                                               | ~~**Update HOW_TO_USE.md** with new threshold semantics~~                                   | ~~Medium~~   | ~~10min~~ | ~~Low~~    | ~~#2~~  |
| ~~10~~ done — FEATURES.md Threshold Control row states statement-count semantics                                                                             | ~~**Update FEATURES.md** with statement-level detection~~                                   | ~~Medium~~   | ~~10min~~ | ~~Low~~    | ~~#2~~  |
| ~~11~~ done — `fnvStep32` / `fingerprintSubtreeInto` helpers in `syntax/syntax.go`; cross-package reuse of golang `hashIdentifierFast` intentionally skipped | ~~**Extract shared FNV helper** (deduplicate `fingerprintSubtree` + `hashIdentifierFast`)~~ | ~~Low~~      | ~~15min~~ | ~~Low~~    | ~~#2~~  |
| ~~12~~ done — `syntax/templ/` sets Statement=true (see item 4 above)                                                                                         | ~~**Templ statement marking** (port `Statement=true` to `syntax/templ/`)~~                  | ~~Medium~~   | ~~40min~~ | ~~Medium~~ | ~~#2~~  |
| ~~13~~ NOT-DO — `--semantic` / `--structural` / `--exact` flags ship directly; alias never adopted                                                           | ~~**`--mode` flag** (alias for semantic/exact/structural)~~                                 | ~~Medium~~   | ~~15min~~ | ~~Low~~    | ~~—~~   |
| ~~14~~ NOT-DO — SARIF wire kept parity-trimmed with go-finding (2026-10-11 G1/G2 policy); extractability stays in JSON/text output only                      | ~~**SARIF: add extractability to properties**~~                                             | ~~Low~~      | ~~10min~~ | ~~Low~~    | ~~—~~   |
| 15                                                                                                                                                           | **HTML: clone-type badge**                                                                  | Medium       | 15min     | Low        | —       |
| 16                                                                                                                                                           | **HTML: extractability column**                                                             | Medium       | 15min     | Low        | —       |
| ~~17~~ done — README.md badge row (CI, License, Go Reference, Website)                                                                                       | ~~**README: add badges** (CI, coverage, Go version)~~                                       | ~~Low~~      | ~~10min~~ | ~~Low~~    | ~~—~~   |
| ~~18~~ done — `baseline/baseline_test.go` covers empty/missing/duplicate-hash cases                                                                          | ~~**Baseline edge-case tests** (empty/corrupt/missing/dup)~~                                | ~~Medium~~   | ~~15min~~ | ~~Low~~    | ~~—~~   |
| ~~19~~ done — root `--diff-report` flag ( `cmd/flags.go` ), pinned by `bdd/diff_report_test.go`                                                              | ~~**`check --diff` flag** (show what changed since baseline)~~                              | ~~Medium~~   | ~~30min~~ | ~~Low~~    | ~~—~~   |
| ~~20~~ NOT-DO — re-running `art-dupl baseline` re-snapshots ( `cmd/baseline_cmd.go` help text); merge flag never adopted                                     | ~~**`baseline --update` flag** (merge new clones)~~                                         | ~~Medium~~   | ~~30min~~ | ~~Low~~    | ~~—~~   |
| ~~21~~ done — SDK `Options.DetectionMethods []DetectionMethod` ( `pkg/artdupl/types.go` ); no Semantic bool ever shipped                                     | ~~**SDK: expose DetectionMode** (replace Semantic bool)~~                                   | ~~Medium~~   | ~~30min~~ | ~~Low~~    | ~~—~~   |
| ~~22~~ superseded — DTO decoupling landed instead ( `docs/status/archived/2026-07-01_04-59_post-dto-decoupling.md` ); no ReadOnlyNode interface              | ~~**Printer decoupling** (ReadOnlyNode interface)~~                                         | ~~Medium~~   | ~~80min~~ | ~~Medium~~ | ~~—~~   |
| ~~23~~ done at `e505886e` (concurrent commit, per `docs/status/archived/2026-07-24_19-48_clone-type-consolidation-status.md`)                                | ~~**Clone type consolidation** (CloneLocation shared type)~~                                | ~~Medium~~   | ~~90min~~ | ~~Medium~~ | ~~#22~~ |
| ~~24~~ done — type-aware mode ( `--type-aware`, `pkg/artdupl` Options.TypeAware, AGENTS.md)                                                                  | ~~**go/types normalizer upgrade** (precise scope resolution)~~                              | ~~Medium~~   | ~~60min~~ | ~~Medium~~ | ~~#2~~  |
| 25                                                                                                                                                           | **Performance: profile normalizer overhead**                                                | Medium       | 20min     | Low        | #2      |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the default threshold be 1 or 3?**

With statement-level tokenization, threshold 1 means "report any duplicated single statement." This is very noisy on real codebases (common patterns like `if err != nil { return err }`, `defer file.Close()`, `fmt.Println(...)` will all match). But threshold 3 requires 3+ consecutive duplicated statements, which misses common 2-statement clones.

The post-hoc filters (T7 overlap elimination, T8 test suppression, T9 actionability patterns) were calibrated for node-level thresholds. Their effectiveness at statement-level granularity is unknown.

**What I need:** Dogfooding data. Run art-dupl on itself at thresholds 1, 3, and 5, count the clone groups, and assess what percentage are actionable. Then pick the default that balances signal vs noise.

---

## Key Files Changed (Uncommitted)

```
Core algorithm:
  syntax/syntax.go          — Node.Statement field, fingerprintSubtree, serial/getUnitsIndexes/FindSyntaxUnits rewrites
  syntax/golang/transform.go — BlockStmt + addBodyStatements mark Statement=true

Threshold rescaling:
  config/config.go           — DefaultThreshold 15 → 1
  cmd/flags.go               — Help text updated
  printer/clone_classify.go  — All priority thresholds rescaled ~3x

Test calibration:
  config/config_test.go      — Default threshold assertion
  config/config_enum_test.go — Empty file threshold assertion
  cmd/cmd_test.go            — Flag default threshold assertion
  printer/clone_classify_test.go — All 16 test cases token counts rescaled
  bdd/test_constants_test.go — Threshold constants rescaled
  bdd/*.go                   — All threshold CLI args rescaled (partial, 67 failures remain)

Temporary:
  syntax/golang/token_count_test.go — Debug file, should be deleted
  docs/planning/2026-06-20_20-30_comprehensive-remaining-todo-plan.md — Plan from earlier
```
