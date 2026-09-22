# Status Report: Self-Hosting Dedup Sprint (-t 4 → -t 1)

**Date:** 2026-09-22 21:47 CEST
**Scope:** art-dupl self-scan deduplication sessions (3 user turns)
**Working tree:** clean — all session changes auto-committed on `fork` (628a3b10, 27b5c43b, 3873decf, 27438816)
**Verification baseline:** `go test ./...` green throughout; gofmt clean; buildflow golangci-lint clean on touched files (turn 1 + 2; see (b)3 for turn 3 gap)

---

## a) FULLY DONE

### a1. -t 4 type-aware self-scan: test-fixture pipeline deduplicated
- **What:** The reported clone pair (`syntax/golang/generics_erase_test.go:126-158` / `typeinfo_test.go:108-146`) was the visible tip of a 7×-repeated fixture pipeline (write 2 sources → `LoadTypeAwareData` → `LookupPreloaded` → `parsePreloadedTest`) spread over 3 test files, with error handling already drifted (some sites `t.Fatalf`, others silently `_`).
- **Change:** New helpers `typeAwarePair` / `newTypeAwarePair()` / `parseTypeAwareFile()` in `syntax/golang/typeinfo_test.go` (fail-fast); all 7 call sites rewritten (`typeinfo_test.go` ×2, `generics_erase_test.go` ×3 incl. the reload-on-same-files case, `interface_method_test.go` ×2). ~80 lines → ~20.
- **Evidence:** commits 628a3b10 + 27b5c43b; `go test ./...` green; gofmt clean; buildflow lint: 0 findings in the 3 changed files; re-scan at `-t 4 --type-aware`: **1 group → 0 shown**; new helpers introduce no new clones (checked with `--no-actionability --show-suppressed`).

### a2. filter_stats.go "??? deduplicate?" root-caused and resolved
- **What:** The flagged clone (`cmd/filter_stats.go:226-232` / `253-259`, the two `for _, pattern := range unmatched` warning loops) kept appearing **despite** four pre-existing `//art-dupl:accept deliberate exclude/include symmetry` directives.
- **Root cause:** `AcceptedSet.IsAccepted` (cmd/accept_directive.go:85-88) matches a directive only within `[LineStart-5, LineEnd]` of a clone instance. All four directives sat on FuncDecl lines (153, 170, 213, 240) — 8 lines outside both windows. ADR-0023's statement-level tokenization made the inner print loop (not the enclosing function) the clone root, silently orphaning the directives.
- **Fix:** Re-anchored — added one directive directly above each flagged loop. Kept the old FuncDecl-level ones (author's text; harmless documentation). Deliberately did **NOT** extract a helper: the recorded owner judgment is "deliberate exclude/include symmetry" (flag-specific example globs; parallel API), and overriding recorded intent to silence a report is the wrong trade.
- **Evidence:** commit 3873decf; group absent from `-t 1` output while other filter_stats groups still show (proves directive matching actually fires, not a reporting fluke).

### a3. Full -t 1 self-scan swept: all 30 shown groups judged, 1 harmful clone extracted
- **What:** Dumped every group shown at `-t 1 --type-aware` (30 groups, 1061 detected) and judged each per the dedup skill (extract harmful / accept intentional with rationale).
- **Accepted (27 groups) with reasons:** enum `IsValid()` return-true tails ×14 (documented enum convention, per-type valid sets); `close(schan)` ×3 (documented ctx-propagation lifecycle); nil-guard + lock pairs in `filter_stats.go` (idiomatic guard clauses); flag-setter pairs; single `default:` clauses; `case golang.IfStmt:` clauses in the denylist pattern table (documented first-match-wins design); Ukkonen `if oldr != t.root` pair (hot path, alloc-gate protected); `countSerializedNodes`/`serial` mirror (documented must-stay-in-lockstep invariant); `statsChan <- stats` finalize pairs (verified buffered(1) + single terminal send ⇒ cannot block); min-vs-max loops in `run_output.go` / `extractability_engine.go` (opposite semantics, cross-package, stdlib `slices` already covers the concept); templ view markup; BDD helper guards.
- **Extracted (1 harmful semantic clone):** `isAssignFromMRun` / `isOSExitCall` (printer/actionability/actionability_preamble.go) — two 13-line identical walkers (unwrap → CallExpr → SelectorExpr → receiver Ident) differing only in names ("m"/"Run" vs "os"/"Exit"). Extracted `hasMethodCall(call, recv, method)`; both functions are now thin wrappers. Checked `firstChild` / `hasCommandReceiver` first to match existing vocabulary; did NOT merge `hasCommandReceiver` (set-of-receivers is a different contract).
- **Evidence:** commit 27438816; `go test ./...` + actionability package tests green; gofmt clean; re-scan: **1061/30 → 1059/28 detected/shown**; content-level diff of before/after reports confirms exactly the two intended removals (the walker pair, plus its now-restructured `BaseType != CallExpr` guard twin), no new groups, only sort-order churn (parallel-search output order is nondeterministic — documented).

---

## b) PARTIALLY DONE

### b1. AGENTS.md gotcha: accept-directive placement (S)
- **Works:** Diagnosis complete (see a2).
- **Open:** The lesson is not yet recorded in the project's `//art-dupl:accept` AGENTS.md bullet: *directives on an enclosing FuncDecl do NOT cover statement-level clone roots deeper in the body — anchor within 5 lines of the actual clone root (post-ADR-0023 reality)*.
- **Blocker:** None — interrupted by this report request. Next action item (f1).

### b2. Post-extraction regression at -t 4 / default -t 5 (S)
- **Works:** `-t 1` verified after the `hasMethodCall` extraction (28 shown, no new groups).
- **Open:** `-t 4` and `-t 5` (default gate) not re-run after the preamble change. Risk is low (pure helper extraction, tests green) but the threshold-ladder check is the documented discipline (AGENTS.md corpus-validation cadence).

### b3. buildflow lint on turn-3 touched file (S)
- **Works:** `go test` + gofmt verified for `printer/actionability/actionability_preamble.go`.
- **Open:** `buildflow -s golangci-lint` not run after commit 27438816. Turns 1–2 were lint-verified; turn 3 was not.

### b4. Stale installed CLI binary diagnosed, not fixed (S)
- **Works:** Root cause identified — the user's terminal `art-dupl` prints the old summary format ("Found total 1 clone groups.") and reports 1 group at `-t 1` where the current tree reports 1059 detected / 28 shown. The binary predates the ADR-0023-era detection/printer behavior, which explains why the user saw only the filter_stats group while the tree contains 28 shown groups at that threshold.
- **Open:** Rebuild/reinstall not performed (`nix build` / buildflow owns the deliverable; channel choice is an owner decision — see question g3).

---

## c) NOT STARTED

| Item | Why not started | Priority |
|---|---|---|
| Dead-directive detector: report `//art-dupl:accept` directives whose window matched zero groups this run | Feature idea born from a2; needs design (flag vs `--explain` integration) | High — would have caught a2 instantly, self-hosting dogfood |
| Consolidate/remove the 4 inert FuncDecl directives in `filter_stats.go` | Blocked on the detector above (safe identification) or an owner decision | Medium |
| `docs-health` HARVEST of this report's section (f) into `TODO_LIST.md` / `ROADMAP.md` | User instructed "THEN WAIT" — awaiting go-ahead | High (closes the report loop) |
| `-t 2` / `-t 3` self-scan classification sweep | The two unvisited thresholds of this session's ladder | Medium |
| End-to-end test pinning the directive-vs-statement-clone scenario (directive 6+ lines above root ⇒ NOT suppressed; ≤5 ⇒ suppressed) at file level | Unit tests exist in `cmd`; the filter_stats file-level scenario is untested | Medium |
| Update AGENTS.md actionability corpus numbers if go-sse / go-cqrs-lite counts shifted due to this session's helper extraction | Requires re-running the documented corpus commands | Medium |
| `warnUnmatchedPatterns(stderr, flag, examples, patterns)` helper for the filter_stats warn loops | Deliberately NOT done — conflicts with the owner's recorded accept (see g1) | Low / decision-gated |

---

## d) TOTALLY FUCKED UP

**Nothing is broken, red, or data-losing.** All tests green, tree clean, every change committed and verified at its threshold. What follows is the honest list of session-level own goals:

1. **Four shipped `//art-dupl:accept` directives were silently inert since ADR-0023** (2026-09-14). Severity: none at runtime (they only under-suppress, i.e. the report is noisier than intended), but it means the self-hosting discipline never re-scanned `cmd/filter_stats.go` after detection granularity changed — the directives were written pre-ADR-0023 and never re-validated. Mitigation: done (re-anchored, 3873decf); prevention: dead-directive detector (c1) + AGENTS.md rule (b1).
2. **Turn-1 verification stopped at the user's threshold instead of sweeping the ladder.** I verified `-t 4` clean and reported done; the user's very next command was `-t 1`, which surfaced a different class (statement-level micro-clones + the directive bug). The AGENTS.md corpus-validation pattern (check multiple thresholds) was right there and I applied it only after being asked twice. Cost: one extra round-trip.
3. **Two wasted tool calls in turn 1** (`go run .`, `go run ./cmd` — the main package is `./cmd/art-dupl`, which AGENTS.md's architecture section documents). Trivial, but the check was free.
4. **Session-1 self-report said "0 clone groups shown" without qualifying the threshold.** Accurate but misleading in hindsight; thresholds are load-bearing context in this tool and should always be quoted.

---

## e) WHAT WE SHOULD IMPROVE

1. **Anchor accept directives to clone roots, not enclosing declarations.** Pain: silently inert suppression (a2). Fix: AGENTS.md rule + dead-directive detector (c1).
2. **Threshold-ladder regression after any tokenization/actionability change.** Pain: bugs survive weeks (directive staleness). Fix: add `-t 1` (and one mid threshold) self-scan to the corpus-validation checklist in AGENTS.md.
3. **Stable group ordering for diffable reports.** Pain: before/after `-t 1` reports reorder run-to-run (parallel search output order), turning verification diffs into churn. Fix: tie-break equal-sort-key groups by content hash in the text printer. Small, self-hosting-specific UX win.
4. **Lint every touched file before claiming done.** Turns 1–2 did, turn 3 skipped it until flagged here. Fix: keep buildflow on the pre-finish checklist (it already is in the global workflow — this session drifted).
5. **Type-aware self-scan latency.** Observed ~15–17 s per run, dominated by go/packages type loading of ~380 files. Pain: dedup iteration at low thresholds is slow. Fix idea: session-local type-data cache keyed by content hashes (the AST cache already does this for parses).
6. **Quote the threshold in every self-scan claim.** Cheap documentation discipline; prevents the (d)4 class of confusion.

---

## f) NEXT TASKS (ranked; Impact / Effort / Category — HARVEST feed for TODO_LIST.md / ROADMAP.md)

> The user asked for "up to 50". Below are the **25 grounded** items; inventing 25 more from thin air would poison the backlog — items beyond these should come from harvesting `TODO_LIST.md` / `ROADMAP.md`, not from this session's imagination.

| # | Task | Impact | Effort | Category |
|---|---|---|---|---|
| 1 | Record accept-directive placement gotcha in AGENTS.md (anchor to clone root, ≤5 lines) | Critical | S | Documentation |
| 2 | Re-run `-t 4` and `-t 5` self-scan after `hasMethodCall` extraction; confirm 0 shown | Critical | S | Quality |
| 3 | `buildflow -s golangci-lint` on `printer/actionability/` (turn-3 file unlinted) | High | S | Quality |
| 4 | Rebuild/install the CLI binary the user runs (nix/buildflow) so terminal runs match the tree | Critical | S | Bug |
| 5 | Harvest this section into `TODO_LIST.md` + `ROADMAP.md` (docs-health HARVEST) | High | S | Documentation |
| 6 | Dead-directive detector: flag `//art-dupl:accept` directives that matched zero groups this run | High | M | Feature |
| 7 | File-level test: statement-level clone + directive 6 lines above ⇒ shown; ≤5 lines ⇒ suppressed | High | S | Quality |
| 8 | Remove/consolidate the 4 inert FuncDecl directives in `cmd/filter_stats.go` (after #6 confirms) | Medium | S | Cleanup |
| 9 | Stable tie-break (content hash) for equal sort keys in text output — diffable self-scan reports | Medium | S | Feature |
| 10 | Classify `-t 2` / `-t 3` self-scan thresholds like `-t 1` (inventory + judgment) | Medium | S | Quality |
| 11 | Re-run corpus validation (go-sse, go-cqrs-lite) after helper extraction; update AGENTS.md counts if shifted | Medium | M | Quality |
| 12 | Document the `-t 1` acceptance inventory (28 idiomatic groups + reasons) in AGENTS.md or a doc so future sessions start from it | Medium | S | Documentation |
| 13 | Session-local type-aware data cache (content-hash keyed) to kill the ~15 s go/packages load per run | High | L | Feature |
| 14 | Owner decision + optional extraction: `warnUnmatchedPatterns` helper for filter_stats warn loops (see g1) | Low | S | Cleanup |
| 15 | Consider printing "shown at threshold T" in the final summary line (threshold is load-bearing context) | Low | S | Feature |
| 16 | Add `--version` staleness notice: compare built commit vs… (needs decision, see g3) | Medium | S | Feature |
| 17 | Revisit `hasCommandReceiver` vs `hasMethodCall` consolidation IF a third selector+receiver variant appears (YAGNI now) | Low | S | Cleanup |
| 18 | Consider an actionability pattern for terminal `send; close()` channel-finalize pairs if `-t 1` noise annoys (currently accepted) | Low | S | Feature |
| 19 | Pin the observed nondeterministic group ORDER behavior in a test (documents the parallel-search contract) | Low | S | Quality |
| 20 | Sweep AGENTS.md "Known Limitations" for entries this session invalidated (directive examples, counts) | Medium | S | Documentation |
| 21 | Add the `-t 1` threshold-ladder self-scan step to the corpus-validation checklist in AGENTS.md | Medium | S | Documentation |
| 22 | Extract the shared `sendCtx`-style finalization note: confirm consumer-drain invariant for terminal stats sends is documented (verified safe this session: buffered(1)) | Low | S | Documentation |
| 23 | Dogfood: run art-dupl ON the actionability package with `--explain` and record why each accepted pattern exists (feeds #12) | Medium | M | Quality |
| 24 | Review whether `Find total N clone groups`-era printers still exist anywhere (stale binary shows old format; confirm no dead printer paths in tree) | Low | S | Cleanup |
| 25 | After #4: capture a fresh `-t 1` / `-t 4` baseline pair into `docs/benchmarks/`-style reference so future dedup sessions diff against a pinned snapshot | Low | S | Documentation |

---

## g) QUESTIONS (cannot answer myself)

1. **filter_stats warn loops:** Your four directives record "deliberate exclude/include symmetry" — I honored that (re-anchored directives, no extraction). Final answer: keep accepted-by-directive, or do you want the two loops extracted into `warnUnmatchedPatterns(stderr, flag, examples, patterns)` after all? *(Tried: reading the directive history and semantics; owner intent is not derivable from code.)*
2. **Low-threshold operating contract:** Is `-t 1` diagnostic mode where 28 idiomatic micro-groups are *expected* output (my recommendation: yes — default gate stays `-t 5`), or should the tool suppress micro-clones below a statement-count floor by default so even `-t 1` reads clean? This decides whether #18 is a feature or a non-goal.
3. **Binary channel:** Your terminal binary predates ADR-0023 (old summary format, 1 group at `-t 1`). Which install channel is canonical for daily self-hosting — `nix build` flake app, buildflow release, or `go run ./cmd/art-dupl`? This unblocks #4 and the #16 staleness-notice design.

---

*Report generated per user instruction; format overridden from the status-report skill's HTML default to Markdown at explicit user request (docs/status/*.md). Section (f) is the HARVEST feed — do not let it die in this timestamped file.*
