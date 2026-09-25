# Self-Clean Decision Ledger

Durable per-group record for the `-t 1` self-scan, so future sessions do not
re-litigate accepted clone groups. Routine: `scripts/self-scan.sh` (first run
of each month) — judge every shown group, append a dated section here.

Convention per entry: where the clone lives, why it is accepted (or what was
extracted), and the date of the judgment. Entries never get deleted; a group
that reappears in a later sweep with a _changed shape_ gets a fresh entry.

---

## 2026-09-19 — baseline `-t 1` self-scan

**Shown groups: 40** (re-measured 2026-09-19 with the current tree; the
"44" quoted in the 2026-08-16 status report was a file-count heuristic, not a
group count). All 40 judged as accepted boilerplate/intentional; detailed
per-group rationale lives in the 2026-08 self-scan sessions (see
`docs/status/2026-08-16_*` reports and CHANGELOG).

## 2026-09-22 — full `-t 1 --type-aware` sweep

**Detected: 1061 / shown: 30** at `-t 1 --type-aware`. Every shown group
judged (details: `docs/status/2026-09-22_21-47_dedup-sprint-self-scan-t4-t1.md`).

### Extracted (harmful)

| Clone                                                                                   | Action                                                                                                                                                                            | Evidence          |
| --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `isAssignFromMRun` / `isOSExitCall` (`printer/actionability/actionability_preamble.go`) | Extracted shared `hasMethodCall(call, recv, method)` walker; both are now thin wrappers. `hasCommandReceiver` deliberately NOT merged (set-of-receivers is a different contract). | commit `27438816` |

### Accepted (intentional, with rationale)

| Pattern                                                                    | Count   | Rationale                                                                                                                                                  |
| -------------------------------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Enum `IsValid()` return-true tails                                         | 14      | Documented enum convention: per-type explicit valid sets beat a shared table for diff-ability.                                                             |
| `close(schan)` pairs                                                       | 3       | Documented ctx-propagation lifecycle (sender closes; mirrored close sites).                                                                                |
| Nil-guard + lock pairs in `cmd/filter_stats.go`                            | several | Idiomatic guard clauses; `withLock[T]` already centralizes the general case.                                                                               |
| Flag-setter pairs (`cmd/`)                                                 | few     | Cobra flag wiring is positional by nature; a setter table would obscure help text order.                                                                   |
| Single `default:` clauses                                                  | few     | Switch-shape boilerplate, extraction adds indirection for 2 lines.                                                                                         |
| `case golang.IfStmt:` rows in the actionability denylist table             | several | Documented first-match-wins pattern table; rows are data, not logic.                                                                                       |
| Ukkonen `if oldr != t.root` pair (`suffixtree/`)                           | 1       | Hot construction path, alloc-gate protected; branch differs semantically at each site.                                                                     |
| `countSerializedNodes` / `serial` mirror (`syntax/`)                       | 1       | Documented must-stay-in-lockstep invariant (arena under/overflow guard).                                                                                   |
| `statsChan <- stats` finalize pairs (`cmd/`)                               | 1       | Verified buffered(1) + single terminal send cannot block.                                                                                                  |
| Min-vs-max loops (`cmd/run_output.go`, `printer/extractability_engine.go`) | 1       | Opposite semantics, cross-package; stdlib `slices` already covers the concept where it fits.                                                               |
| Templ view markup siblings (`printer/`)                                    | few     | Structural similarity of templ components is the format's grain, not duplication.                                                                          |
| BDD helper guards (`bdd/`)                                                 | few     | Ginkgo setup boilerplate; readability beats DRY in specs.                                                                                                  |
| filter_stats warning loops (exclude/include symmetry)                      | 1       | Deliberate exclude/include symmetry; four `//art-dupl:accept deliberate exclude/include symmetry` directives re-anchored to the loops (commit `3873decf`). |

**Result of the sweep:** 1059 detected / 28 shown after the extraction; no new
groups introduced. Sort-order churn between runs is expected (parallel search
output order is nondeterministic — documented in ADR-0019).

## 2026-09-25 sweep (C3, post-B1 gitignore work)

Scan: `scripts/self-scan.sh` (`-t 1 --type-aware`) → 1078 detected / 28 shown
(after fixing the one group this change-set introduced; the raw scan first
reported 29 shown).

**Extracted this sweep:**
- `internal/gitignore/gitignore.go` — the duplicated `if len(rules) == 0 {
  return nil }` tail across `LoadGitignore` and the new `LoadTree` collapsed
  into `matcherFromRules` (the new group this scan caught in our own B1 work;
  found by our own tool the same day it was written).

**Accepted (existing idiom classes, unchanged from the 2026-09-22 table):**
- `case golang.IfStmt:` rows, single `default:` clauses (actionability
  pattern-table rows are data, not logic).
- `child.Statement = true` pairs across `syntax/golang` / `syntax/templ`
  (transformer shape mirrors the two ASTs; a shared helper would need the
  two node types to unify, which is the ROADMAP syntax-facade cut).
- `statsChan <- stats` finalize pairs, `if i > maxChildren` arena guards,
  `if split` match pairs, BDD `if s.T != nil` guards (test-support
  boilerplate), filter_stats guard pairs, min/max cross-package loops — all
  carry their prior rationales in the table above.

## B7 — SDK-path self-scan (provider crawl on art-dupl itself)

Ran the toolsdk provider path (`toolsdk.All()` → `spec.Detect` with the
art-dupl repo as working dir) against our own tree:

- SDK path: **179 findings / 72 distinct clone groups** (GroupID-stable).
- CLI at `-t 5`: 60 detected groups, 0 shown (23 non-actionable, 37
  filtered suppressed).
- The 60→72 delta is crawl policy, not detection drift: the CLI's default
  config ignores `*_test.go`; the provider's crawl deliberately does NOT
  (BuildFlow parity — jscpd also scans test files). Consumers comparing the
  two channels must expect this difference.
- Cross-checked GroupID determinism in
  `TestFindingsFromGroups_GroupIDStableAcrossRuns` (identical GroupIDs
  across independent runs).
