# Status Report: Actionability & Semantic Detection Improvements

> **Live remainder:** archived 2026-11-14 (cadence wave); every item in this report is struck or routed, and open work is owned by TODO_LIST.md and ROADMAP.md.

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 12 items verified resolved, 5 open (37 patterns shipped; HTML badges still absent).

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 12 items verified resolved, 5 open (37 patterns shipped; HTML badges still absent).

**Date:** 2026-05-16\
**Branch:** fork\
**Author:** Parakletos (Crush AI)\
**Topic:** Reduce false positives in `--semantic` mode; add `--rich-text` enhanced output

---

## Motivation

External review of art-dupl on a Go CQRS monorepo (~10K LOC, 9 modules) found with `--semantic --sort total-tokens -t 25`:

- **80 clone groups** detected
- **~60-70% false positives** that are not actionable code smell (interface signatures, defer unlock, error propagation)

Core insight: _"We only want to report things that can ACTUALLY be deduplicated"_. Adding 6 separate `--ignore-*` flags is flag explosion. Instead: teach `--semantic` mode to understand actionability.

---

## Work Done

### a) FULLY DONE

#### 1. Domain Type: `CloneActionability` (`domain/processed_clone.go`)

```go
type CloneActionability string

const (
    Actionable    CloneActionability = "actionable"
    NonActionable CloneActionability = "non-actionable"
)
```

- Non-actionable = idiomatic Go patterns that **cannot** be deduplicated without breaking semantics
- Added `Actionability` field to `CloneClassification`
- Exhaustive: covers all 3 major false-positive patterns from the review

#### 2. Actionability Analyzer (`printer/actionability.go`)

`EvaluateActionability(nodeSeqs [][]*syntax.Node) CloneActionability`

| Pattern Detected                                                                   | Why Non-Actionable                               | Example                                    |
| ---------------------------------------------------------------------------------- | ------------------------------------------------ | ------------------------------------------ |
| ~~Single `FuncDecl` across files~~ done — verified — design rationale, historical  | ~~Interface method **must** match signature~~    | ~~`func (s *Store) Save(ctx, ...) error`~~ |
| ~~Single `DeferStmt` across files~~ done — verified — design rationale, historical | ~~RAII pattern, extraction breaks semantics~~    | ~~`defer mu.Unlock()`~~                    |
| ~~Single `IfStmt` across files~~ done — verified — design rationale, historical    | ~~Go has no macro system for error propagation~~ | ~~`if err != nil { return err }`~~         |

Logic: **skip entire group ONLY if ALL clones match the same boilerplate pattern**. Conservative — one clone with real body means the whole group stays.

#### 3. Semantic Mode Now Filters (`cmd/run_output.go`)

```go
if semantic {
    if printer.EvaluateActionability(uniq) == domain.NonActionable {
        continue  // skip this clone group
    }
}
```

- `--semantic` → suppresses non-actionable clones (lower noise)
- ~~`--structural` (default) → reports everything (raw power, unchanged)~~ **resolved — semantic is the default since ADR-0007; structural is opt-in.**

#### 4. Rich Text Output (`--rich-text` flag)

New CLI flag: `--rich-text`

Enhances text output from:

```
found 3 clones:
  store.go:97,103
  repo.go:38,44
```

To:

```
found 3 clones: [HIGH] method (45 tokens, 12 lines) suggestion: Extract to shared utility
  store.go:97-103
  repo.go:38-44
```

- Added `RichText` to `config.Config`
- Added `RichTextSetter` interface in `printer`
- `TextPrinter.SetRichText(bool)` implemented

#### 5. Full Wire Integration

Files modified:

- `cmd/run_output.go` — `printDupls` now accepts `semantic bool`; `printCloneGroups` filters
- `cmd/run_flags.go` — passes `semantic` and enables `RichTextSetter` on printer
- `cmd/run_all_modes.go` — passes `semantic` through
- `cmd/stats.go` — passes `semantic` through
- `cmd/flags.go` — added `--rich-text` flag definition
- `cmd/config_builder.go` — extracts `rich-text` flag into `Config`
- `cmd/cmd_test.go` — updated test callers for new signatures
- `config/config.go` — added `RichText bool` field
- `printer/printer.go` — added `RichTextSetter` interface
- `printer/text.go` — `SetRichText` method + `writeRichGroupHeader`

#### 6. Tests (`printer/actionability_test.go`)

8 table-driven test cases:

- Empty sequences → actionable
- Single `FuncDecl` → non-actionable
- Single `DeferStmt` → non-actionable
- Single `IfStmt` → non-actionable
- FuncDecl **with body** → actionable (conservative)
- `ForStmt` loop → actionable
- Mixed types → actionable
- Single sequence only → non-actionable

All pass: `go test ./printer/... -run TestEvaluateActionability -v`

#### 7. Build & Test Verification

- `go build ./...` — ✅ clean
- `go test ./...` — ✅ all 31 packages pass

---

### b) PARTIALLY DONE

None. (This was a focused, single-purpose improvement session.)

---

### c) NOT STARTED

1. ~~**Enhance actionability detection** — currently only handles 3 simple patterns. Could expand to:~~ done — expanded to 42 patterns incl. property engine (ADR-0017, docs/ACTIONABILITY_PATTERNS.md)
   - `go func(...) { ... }()` goroutine patterns
   - `context.WithCancel(ctx)` boilerplate
   - Constructor patterns: `return &Type{Field: val}`
   - Package-level variable declarations

2. ~~**JSON output enrichment** — `JSONPrinter` doesn't show `Actionability` field yet; should include it in JSON payload for CI pipelines~~ done — Actionability in JSON (printer/json.go)

3. ~~**Deeper pattern detection** — Currently only checks `len(seq) == 1`. Could check "if sequence 80% DeferStmt + IfStmt + ReturnStmt" as "pure boilerplate function wrapper"~~ done — property-based extractability engine (ADR-0017)

4. ~~**Performance impact** — `EvaluateActionability` runs per clone group. With 1000+ groups on large repos, could add ~ms overhead. Not benchmarked yet.~~ done — BenchmarkEvaluateActionability_SyntaxOnly (printer/actionability/extractability_bench_test.go)

---

### d) TOTALLY FUCKED UP

Nothing. All tests pass, all builds clean, all changes are minimal and focused.

---

### e) WHAT WE SHOULD IMPROVE

1. ~~**Actionability detection is too simple** — only checks `len(seq) == 1` for FuncDecl/DeferStmt/IfStmt. In reality, a `FuncDecl` match might include just the signature **plus one line of body** (e.g., `defer` + `if err != nil`). We need a token-weighted approach: count actual logic tokens vs boilerplate tokens.~~ done — property-based extractability engine (ADR-0017, 4 properties)

2. ~~**Text output line format inconsistency** — Plain text uses `:` separator (`97:103` meaning tokens); with `--rich-text` it becomes `97-103` (line range). The comma notation is confusing. The report author specifically called this out. Should standardize on `97-103` (line range) everywhere.~~ done — text output uses file:lineStart-lineEnd (printer/text.go)

3. ~~**No `--rich-text` for plumbing/JSON/SARIF** — The flag is silently ignored for non-Text printers. This is undefined behavior from user perspective. Either error, or extend to all formats.~~ **Won't implement — dropped — --rich-text stays text-only; JSON/SARIF carry actionability fields.**

4. ~~**Missing `Actionability` in HTML output** — HTML already has priority badges and category filters. Actionability should be a filterable dimension too.~~ **Won't implement — dropped — HTML shows classification badges, no actionability filter.**

5. ~~**`isPureErrorPropagation` is a stub** — Currently just checks `IfStmt`. Real error propagation detection should inspect children for `err != nil` binary expression + `return` statement.~~ done — error-wrapping detection inspects IfStmt bodies (printer/actionability, T9 patterns)

6. ~~**`isPureDeferPattern` is a stub** — Currently just checks `DeferStmt`. Should inspect the call to distinguish `defer mu.Unlock()` (non-actionable) from `defer expensiveCleanup()` (maybe actionable).~~ done — RAII cleanup list distinguishes defer targets (printer/actionability)

7. ~~**Stats subcommand doesn't benefit** — Would a `art-dupl stats --semantic` suppress non-actionable from statistics too? Currently filtering only happens in `printCloneGroups`, not in `StatsPrinter`. The stats counts will include non-actionable clones, making numbers misleading.~~ done — stats counts Actionable/NonActionable groups (printer/stats/stats.go)

---

### f) Top #25 Things To Do Next

| #                                                                                                | Task                                                                        | Priority | Effort     | Impact                                                 |
| ------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------- | -------- | ---------- | ------------------------------------------------------ |
| ~~1~~ done — property-based engine (ADR-0017)                                                    | ~~Weighted token analysis for actionability (not just single-node)~~        | ~~P0~~   | ~~Medium~~ | ~~Fixes real false positives on partial-body matches~~ |
| ~~2~~ done — line-range notation in printer/text.go                                              | ~~Standardize text output line notation `:` → `-`~~                         | ~~P1~~   | ~~Low~~    | ~~UX fix user explicitly complained about~~            |
| ~~3~~ done — Actionability in JSON (printer/json.go)                                             | ~~Add `Actionability` to JSON output~~                                      | ~~P1~~   | ~~Low~~    | ~~CI pipelines need structured actionability data~~    |
| ~~4~~ **Won't implement — dropped — HTML shows classification badges, no actionability filter.** | ~~Extend `--rich-text` to HTML (filter badges)~~                            | ~~P1~~   | ~~Medium~~ | ~~HTML already rich, should show actionability~~       |
| ~~5~~ done — real error-propagation inspection (T9 patterns)                                     | ~~Implement real `isPureErrorPropagation` (inspect IfStmt body)~~           | ~~P2~~   | ~~Medium~~ | ~~Current is a heuristic placeholder~~                 |
| ~~6~~ done — defer-target inspection via RAII list                                               | ~~Implement real `isPureDeferPattern` (inspect call target)~~               | ~~P2~~   | ~~Medium~~ | ~~Distinguish `Unlock` from business defer~~           |
| ~~7~~ done — stats counts actionability (printer/stats/stats.go)                                 | ~~Actionability filtering for `stats` subcommand~~                          | ~~P2~~   | ~~Low~~    | ~~Consistent behavior across commands~~                |
| ~~8~~ done — BenchmarkEvaluateActionability_SyntaxOnly                                           | ~~Benchmark `EvaluateActionability` overhead on large repos~~               | ~~P2~~   | ~~Low~~    | ~~Performance confidence~~                             |
| ~~9~~ done — HOW_TO_USE.md documents --rich-text                                                 | ~~Add `--rich-text` behavior docs / `--help` examples~~                     | ~~P2~~   | ~~Low~~    | ~~User discoverability~~                               |
| ~~10~~ **Won't implement — dropped — goroutine pattern never added.**                            | ~~Extend actionability to goroutine patterns~~                              | ~~P3~~   | ~~Medium~~ | ~~`go func() { ... }()` is common boilerplate~~        |
| ~~11~~ **Won't implement — dropped — WithCancel pattern never added.**                           | ~~Extend actionability to context.WithCancel patterns~~                     | ~~P3~~   | ~~Low~~    | ~~Standard Go pattern~~                                |
| ~~12~~ done — PatternInterfaceImpl with PriorityLow (printer/actionability/clone_classify.go)    | ~~Add `CategoryInterfaceImpl` to `CloneCategory` enum~~                     | ~~P3~~   | ~~Low~~    | ~~Precise classification vs generic `method`~~         |
| ~~13~~ **Won't implement — dropped — no actionability filter in HTML.**                          | ~~Add actionability filter buttons to HTML report~~                         | ~~P3~~   | ~~Medium~~ | ~~Parity with Production/Test/Category filters~~       |
| ~~14~~ done — interface-impl suggestion string (clone_classify.go)                               | ~~Actionability should set `Suggestion` to "Required by interface"~~        | ~~P3~~   | ~~Low~~    | ~~Better UX on non-actionable reports~~                |
| ~~15~~ done — bdd/actionability_test.go suppresses identical interface method signatures         | ~~Add BDD test for `--semantic` suppressing interface signatures~~          | ~~P3~~   | ~~Medium~~ | ~~Integration test for core feature~~                  |
| ~~16~~ done — bdd/actionability_test.go --rich-text spec                                         | ~~Add BDD test for `--rich-text` flag~~                                     | ~~P3~~   | ~~Low~~    | ~~CLI behavior coverage~~                              |
| ~~17~~ **Won't implement — dropped — no ActionabilityThreshold config.**                         | ~~Consider `ActionabilityThreshold` config (weighted score cutoff)~~        | ~~P4~~   | ~~Medium~~ | ~~Tunable strictness~~                                 |
| ~~18~~ done — docs/ACTIONABILITY_PATTERNS.md + count gate                                        | ~~Document actionability patterns in FEATURES.md / AGENTS.md~~              | ~~P4~~   | ~~Low~~    | ~~Keep documentation current~~                         |
| ~~19~~ **Won't implement — dropped — no --only-actionable flag.**                                | ~~Add `--only-actionable` as alias (if user debate continues)~~             | ~~P4~~   | ~~Low~~    | ~~Future-proofing~~                                    |
| ~~20~~ **Won't implement — dropped — no recorded follow-up.**                                    | ~~Check `RunWithSetAnalysisData` in text output for rich-text~~             | ~~P5~~   | ~~Low~~    | ~~Text output has two paths~~                          |
| ~~21~~ **Won't implement — dropped — errors.Is/As pattern never added.**                         | ~~Error handling pattern detection: inspect `errors.Is` / `As`~~            | ~~P5~~   | ~~Medium~~ | ~~More Go idioms~~                                     |
| ~~22~~ **Won't implement — dropped — nil-check pattern never added.**                            | ~~`nil` check patterns: `if x == nil { return nil, err }`~~                 | ~~P5~~   | ~~Low~~    | ~~Common boilerplate~~                                 |
| ~~23~~ **Won't implement — dropped — type-assertion pattern never added.**                       | ~~Type assertion boilerplate: `if v, ok := x.(T); ok { ... }`~~             | ~~P5~~   | ~~Low~~    | ~~Go idiom~~                                           |
| ~~24~~ done — verified — vendor exclusion already handled                                        | ~~Vendor directory exclusion in actionability (already handled by filter)~~ | ~~P5~~   | ~~Low~~    | ~~Already works~~                                      |
| ~~25~~ done — decided — semantic is the default (ADR-0002)                                       | ~~Investigate if semantic mode should become default~~                      | ~~P5~~   | ~~High~~   | ~~Big UX shift, needs user validation~~                |

---

### g) Top #1 Question I Cannot Figure Out Myself

**What is the correct balance between `EvaluateActionability` and `CloneClassification.calculatePriority`?**

Current design has TWO overlapping systems:

~~1. **Actionability** (this PR) — is this clone group _possible_ to dedup?~~ done — moot — narrative definition, not an open task; actionability shipped (printer/actionability/, 42 patterns per AGENTS.md)
2. **Priority** (existing) — how _important_ is this clone to address?

A `FuncDecl` interface signature gets `Priority: Medium` (from existing logic) but `Actionability: NonActionable` (from new logic). This is correct — it's "medium importance to know about" but "impossible to fix."

But: should `CalculatePriority` be updated to auto-downgrade `NonActionable` clones to `PriorityLow`? Currently `--structural` mode (default) shows them with their original priority. This means a user running without `--semantic` sees `[MEDIUM] method` for interface signatures — misleading, since they can't act on it.

**Should non-actionable always map to PriorityLow, regardless of category? Or should actionability be orthogonal to priority?**

I lean toward orthogonal (current behavior) because:

- `--semantic` already hides non-actionable entirely
- `--structural` is the "show me everything" mode — user explicitly wants to see raw matches
- Downgrading priority for non-actionable in structural mode would be confusing

**But** the classification system's whole purpose is to help users triage. If a structural-mode user sees 50 `[MEDIUM] method` clones that are all interface signatures, that's still noise.

**Proposed answer:** Keep orthogonal, but enhance text output to show `[MEDIUM] method [non-actionable]` when `--structural` is used. This gives full transparency without lying about importance.

Seeking confirmation on this design decision before implementing.

---

## Files Changed This Session

### New Files

- `printer/actionability.go` — Actionability analyzer
- `printer/actionability_test.go` — 8 test cases for analyzer

### Modified Files

- `domain/processed_clone.go` — Added `CloneActionability` type + field
- `config/config.go` — Added `RichText bool` field
- `cmd/flags.go` — Added `--rich-text` CLI flag
- `cmd/config_builder.go` — Extracts `rich-text` flag, applies to config
- `cmd/run_output.go` — `printDupls` + `printCloneGroups` filter on semantic
- `cmd/run_flags.go` — Enables `RichTextSetter` on printer, passes `semantic`
- `cmd/run_all_modes.go` — Passes `semantic` through
- `cmd/stats.go` — Passes `semantic` through
- `cmd/cmd_test.go` — Updated test callers for new signatures
- `printer/printer.go` — Added `RichTextSetter` interface
- `printer/text.go` — `SetRichText` + `writeRichGroupHeader`

---

## Verification

```bash
$ go build ./...          # ✅ PASS
$ go test ./...            # ✅ PASS (31 packages)
$ ./art-dupl --help        # ✅ --semantic, --structural, --rich-text visible
$ ./art-dupl --semantic .  # ✅ runs, suppresses boilerplate (manual validation needed)
```

---

## Next Steps Recommended

1. Address Top #1 question (priority/actionability relationship)
2. Implement #1 (weighted token analysis) for production robustness
3. Run `--semantic` against `go-cqrs-lite` repo to validate ~60-70% reduction claim
