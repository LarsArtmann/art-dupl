# Status Report: Combined Type-Aware + Suggest-Generics Mode (ADR-0026)

**Date:** 2026-10-03 03:30 CEST
**Session scope:** Implemented the combined two-pass mode for `--type-aware --suggest-generics` (CLI + SDK), superseding the old "suggest-generics wins + warning" resolution of ADR-0021. This report covers THIS session's work only — what shipped, what I missed, and what to do next.
**Working tree:** clean (auto-commit daemon captured everything: `4d2436a6` and earlier).
**Full race gate:** `scripts/check-boundary.sh --full` — **PASSED** (all 31 packages, `-race`, 93s). Close-out gate closed; no regressions anywhere in the tree.

---

## a) FULLY DONE

1. **Combined mode, CLI** (`cmd/combined_analysis.go`): crawl once, filter once, ONE `go/packages` type-check, two sequential detection passes, merged match stream. Branch wired inside `executeAnalysis` — every subcommand that calls it (`run`, `stats`, `baseline`, `check`, `diff-report`, `--all`) gets combined mode automatically when both flags are set.
2. **Merge rule, single structural gate**: erased-pass families flow through iff ≥ `syntax.MinDivergentPositions` (2) divergent type positions (`syntax.IsGenericsCandidateStructure`). Zero-divergence families dropped (exact pass-1 duplicates); one-position divergence dropped (the receiver-noise class type-aware exists to kill). Presentation gates (actionability, min-lines, accept directives) apply downstream identically for both passes.
3. **`TypeAwareData.WithEraseHash`** (`syntax/golang/typeinfo.go`): shallow copy sharing ASTs + `types.Info` + `FileSet` — both hash dispositions from one expensive load. Unit-tested.
4. **Canonical constant move**: `MinDivergentPositions` now lives in `syntax/type_divergence.go` (canonical); printer re-exports as a const alias. Arch-lint compliant (`syntax` may not import `domain`/`printer` — discovered the hard way, first attempt violated it and was deleted).
5. **SDK combined mode** (`pkg/artdupl/combined.go`): `FindClones` + `FindClonesStreamResult` both branch to two pipelines sharing one type load, same merge rule at the group-map level. `buildAnalysisPipeline` refactored into `buildPipelineWithTypeData` shared by single and combined modes; `detectGroups`/`streamCloneGroups` extracted for reuse.
6. **Warning → note**: `warnTypeAwareSuggestGenerics` replaced by `noteCombinedMode` ("combined mode active — ... See ADR-0026").
7. **Tests** (all green, `-count=1`):
   - 3 BDD scenarios: both outputs in one run (pair group + family with `generics:` hint); receiver-noise full-body group suppressed in combined while `--suggest-generics` alone shows it (line-range-pinned assertions); graceful fallback on broken type info.
   - 3 SDK tests: pair+family groups both present; noise family dropped in combined / kept in sg-only (line-6-divergent-statement check); streaming smoke.
   - `printer/generics_parity_test.go`: pins `ClassifyGenericsCandidate` ↔ `syntax.IsGenericsCandidateStructure` agreement.
   - Unit tests for `CountTypeDivergencePositions`, `IsGenericsCandidateStructure`, `WithEraseHash`, pre-order traversal.
8. **Incremental combined mode**: smoke-tested via real CLI (`--incremental --type-aware --suggest-generics` on a 2-file fixture): both groups output, both passes print status, per-pass cache stats work. Uses the existing `ta`/`sg` cache-key tags — no `CacheVersion` bump needed.
9. **Refactors shipped along the way**: `buildSuffixTreeStandardPass` / `buildSuffixTreeIncrementalPass` extracted from the monolithic builders (shared by single + combined modes); `collectFiles` extracted from `loadTypeAwareData`; `streamCombinedInto` extracted (gocognit fix).
10. **Docs**: ADR-0026 written; ADR-0021 superseded paragraph amended; HOW_TO_USE gained a "Combined Mode" section + updated tradeoffs; FEATURES.md new row (+ formatter re-alignment); SDK_DESIGN.md updated; AGENTS.md conventions bullet rewritten.
11. **Quality gates**: build ✓, vet ✓, targeted `-race` (cmd, sdk, syntax, golang, printer) ✓, golangci-lint via BuildFlow → 0 findings (fixed my own 4: gocognit-45, 2× nonamedreturns, prealloc), boundary fast profile ✓, full BDD suite 328 specs ✓.

---

## b) PARTIALLY DONE

1. **`--timing` in combined mode**: `executeCombinedAnalysis` never calls `job.RecordStage(PhaseIngest)`. PhaseSearch IS recorded (reused `spawnCloneDetection`), so `--timing` output for combined runs is missing the ingest stage. The `--timing` flag is user-visible. Shipped incomplete wiring.
2. **`--profile` in combined mode**: `startProfiling`/`endProfiling` are NOT threaded into the combined path — the flag is silently ignored there. (`profile` is a hidden flag, so blast radius is small, but it's still silently-discard behavior — exactly the smell this feature was supposed to eliminate.)
3. **Incremental combined-mode test coverage**: manual smoke test only. No automated regression test pins `--incremental` + combined (standard path has 3 BDD scenarios + smoke via other tests).
4. **SDK streaming combined test**: smoke-level only (count > 0); no content assertions (pair vs family) like the FindClones test has.
5. **`Options.SuggestGenerics` doc comment** (`pkg/artdupl/types.go:175`): still says "all clones still returned" — true for sg-only, but stale for the combined case (where cross-type non-candidates are filtered). One-line comment update owed.
6. **CHANGELOG.md**: no Unreleased entry for this user-facing behavior change (flag combination semantics changed from "warn + discard" to "combined mode").
7. **TESTING.md / AGENTS.md BDD gotcha**: discovered that go/packages in the BDD sandbox tmp dirs resolves stdlib imports to `invalid type` (worked around with same-package fixture types) — this gotcha is documented NOWHERE. Future sessions will rediscover it the hard way.

---

## c) NOT STARTED

1. Provider adoption: `pkg/provider` runs the SDK in semantic-only mode; combined mode is not exposed to BuildFlow findings (deliberate — needs user decision, see question 1).
2. Combined-mode benchmark entry in `docs/benchmarks/` (overhead of two parses + two trees + two searches vs. one; wall-clock AND peak memory).
3. Baseline/check interplay with combined mode: group hashes come from the per-pass hash functions, so a baseline recorded in single mode will not match a combined-mode run. No docs, no warning, no test.
4. Monthly self-scan (`scripts/self-scan.sh`) with combined mode + ledger decisions.
5. TODO_LIST.md harvest of this report's section (f) — intentionally deferred: user said report, then wait.

---

## d) TOTALLY FUCKED UP (honest ledger)

1. **First attempt at the canonical constant violated arch-lint** — I put `MinDivergentPositions` in `domain/`, but `syntax` (where the helper lives) may not depend on `domain`. Caught on first build, deleted and relocated. Cost: one wasted write cycle. Root cause: I designed the dependency graph in my head instead of checking `.go-arch-lint.yml` first.
2. **Two broken BDD fixture iterations before the third worked**:
   - Iteration 1 asserted on function NAMES in output — but clone previews show the first BODY line, never the signature. The assertion could never pass.
   - Iteration 2's "receiver noise" fixture didn't USE the receiver in the cloned region, so the type-aware pass legitimately matched the common statements and the "suppressed" assertion failed. Also exposed that my mental model ("suppressed in combined") needed refinement: the correct contract is _full-body cross-type group dropped, same-type core still reported_ — the final test pins exactly that with line ranges.
   - Lesson: I wrote assertions against imagined output instead of running the thing first. Two full test cycles burned.
3. **Undiagnosed root cause, worked around blindly**: stdlib imports in BDD tmp-dir fixtures fail go/packages type-checking (`invalid type` per ident, mixed resolution across files of the same adhoc package). I redesigned the fixture to same-package types instead of understanding WHY. It smells like module-less adhoc-package behavior of go/packages `file=` patterns in that execution context. Unit tests in `syntax/golang` with the same imports DO resolve — the difference (two files, same dir, same package, different imports) is unverified.
4. **Shipped with known unwired flags** (see b1/b2): `--timing` loses a stage, `--profile` is a no-op in combined mode. I noticed both only during this self-review, not during implementation. This is the same class of "silently ignore what the user asked for" that motivated the whole feature — embarrassing specifically because of that.

Nothing is broken in the shipped path: all gates green, no regressions. The damage in this section is wasted cycles, an undiagnosed environment quirk, and two niche flags unwired.

---

## e) WHAT WE SHOULD IMPROVE (session-level self-review)

**What did I forget?** CHANGELOG, TESTING.md gotcha, `--timing`/`--profile` threading, the stale `Options.SuggestGenerics` comment, full-race-gate-at-closeout (running now, in background).

**What is stupid that we do anyway?**

- `job.sendCtx` is unexported, so `cmd` re-implements the context-aware send THREE times now (`spawnCloneDetection`, my `sendMatch`, the SDK closures). Export it or move a `cmd` helper.
- The flatten walker now exists twice (`syntax.flattenNodes` for `*syntax.Node`, `printer.flattenCloneNodes` for `*domain.CloneNode`) with a parity test holding them together. A test-enforced duplicate is a split brain with a bodyguard — the guard works, but the duplication is structural debt.

**What could I have done better?**

- Read `.go-arch-lint.yml` BEFORE designing where the constant lives (dependency directions were one `view` away).
- Run the CLI on a scratch fixture BEFORE writing output-format assertions (would have saved both broken BDD iterations).
- Check which optional flags thread through `executeAnalysis` when adding a sibling path — a checklist ("timing? profile? stats? filter warnings?") would have caught b1/b2 in implementation instead of review.

**What can still be improved?**

- Combined-mode peak memory: both suffix trees + both node-slice sets are alive simultaneously (both detectors run). A large repo pays ~2x tree memory. Option: fully drain pass-1 search, free its tree, then build pass 2 (halves peak, adds wall-clock). Needs a benchmark first — no data, no decision.
- Incremental combined creates one `IncrementalParser` per pass (fresh LRU each, cache stats printed per pass). One shared parser with `SetTypeAwareData` per pass (tag re-derives correctly — verified reading `SetTypeAwareData`) would share the LRU and print once.
- The CLI note text references "ADR-0026" — an internal doc id in user-facing output. Users without the repo can't follow it; drop or soften.

**Did I lie to you?** Not knowingly. One claim to downgrade: "cost: one type-check + two CHEAP parse/tree passes" — "cheap" is relative and currently unmeasured; the benchmark (c2) is the honesty check on that sentence.

**Split brains created?** One, contained: the dual flatten walkers + dual constant references (canonical + alias). Pinned by `generics_parity_test.go`. Ghost systems: none — everything written this session is reachable (CLI branch, SDK both entries, tests).

**Tests:** good coverage of the new semantics at three levels (unit, BDD, SDK). Gaps: incremental combined, streaming content, hash-only interplay, output-format fields (JSON/SARIF) under combined mode.

---

## f) NEXT (prioritized, ≤50)

**Immediate fixes from this session's misses:**

1. Thread `--profile` through `executeCombinedAnalysis` (b2).
2. Record `PhaseIngest` (plus per-pass stage tags) in combined mode so `--timing` is complete (b1).
3. Add CHANGELOG.md Unreleased entry for combined mode.
4. Update `Options.SuggestGenerics` comment in `pkg/artdupl/types.go` (combined-mode behavior).
5. Document the BDD-sandbox stdlib-import `invalid type` gotcha in TESTING.md + AGENTS.md.
6. Root-cause the go/packages adhoc-package `invalid type` behavior (two files, same dir, same package, different imports, no go.mod) — write a minimal reproducer outside BDD.
7. Add automated regression test for `--incremental` + combined mode.
8. Upgrade SDK streaming combined test to content assertions (pair + family).
9. Soften "ADR-0026" reference in user-facing CLI note.

**Correctness/consistency hardening:**
10. Decide + implement behavior for combined + hash-only detection methods (currently: hash-only branch wins first, both type flags silently ignored — warn or validate).
11. Baseline interplay: document mode→hash coupling in baseline docs; consider warning when `check`/`record` run with flags whose hash mode differs from the recorded baseline.
12. Test combined + `--json` / `--sarif`: `generics_candidate`/`generics_hint` present on family groups only.
13. Test combined + `--show-suppressed` and `--no-actionability` interplay.
14. Test accept-directives against erased-pass group hashes (hash values differ per pass; document which hash the directive must target).
15. Test combined + `--search-workers N` (parallel search across two trees).
16. Test combined + `--plumbing` output determinism.
17. Nil/empty-frag edge cases for `CountTypeDivergencePositions` (defensive unit tests).
18. `stats` subcommand under combined mode: cache stats are pass-1 only — decide if summed stats are more honest, then document.

**Performance/memory:**
19. Benchmark combined mode vs. two separate runs (wall-clock + peak memory) → `docs/benchmarks/`.
20. Based on 19: optionally serialize pass-2 build after pass-1 search to halve peak tree memory.
21. Share one `IncrementalParser` across combined passes (shared LRU, single cache-stats print, summed stats).
22. Re-verify `crawl-overhead` benchmark note if the provider ever adopts combined mode.

**Structure/debt:**
23. Export `job.SendCtx` (or add a cmd-local helper) and collapse the three hand-rolled context-send sites.
24. Unify the dual flatten walkers behind one implementation (e.g., domain-level walker + exported syntax→CloneNode bridge both layers can use); keep the parity test regardless.
25. Delete the pre-existing no-op `warnTypeAwareIncremental` in `cmd/config_builder.go`.
26. Consider exposing combined mode to `pkg/provider` (BuildFlow) — needs user decision (question 1).
27. Property/fuzz test for the parity invariant (random VarType trees, assert both implementations agree).
28. Self-scan art-dupl with combined mode (`-t 1`) and ledger per-group decisions in `docs/SELF_CLEAN_LEDGER.md`.

**Docs/process:**
29. HARVEST this report's section (f) into TODO_LIST.md (docs-health) — pending user go-ahead per "wait for instructions".
30. Add combined-mode section to TESTING.md inventory.
31. Verify next docs-health monthly pass (due 2026-10-28) picks up this report + ADR-0026.
32. Confirm gopls diagnostics on `pkg/artdupl/combined.go` are clean post-restart (the "unused" warnings were stale all session despite builds proving usage).
33. Run `nix flake check` (full CI mirror) in background at next convenience.
34. Add a "wiring checklist" (timing/profile/stats/filter-warnings) to AGENTS.md for anyone adding a new analysis path sibling to `executeAnalysis`.
35. Consider a README one-liner for SDK combined mode (README currently doesn't mention either flag).

---

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Provider adoption**: Should `pkg/provider` (BuildFlow's Go detector) gain combined mode? It doubles parse+tree work per CI run; findings quality would improve (exact duplicates + generics hints in one gate pass). Cost/benefit on YOUR CI fleet is your call, not mine.
2. **Baseline policy for hash-mode drift**: baselines are only comparable within the same hash disposition (single `ta`, single `sg`, or combined). Options: (a) document only, (b) warn on `check` when the active mode differs from the recorded one, (c) stamp the mode into the baseline file and hard-fail on mismatch. Which strictness do you want?
3. **Peak-memory vs. wall-clock for large repos**: combined mode currently holds both suffix trees simultaneously. I can serialize (halve peak memory, some latency) — but whether your largest repos are memory-bound or latency-bound in CI is something only you know. Benchmark first, or serialize now?

---

_Report written 2026-10-03 03:30 CEST. Format note: user explicitly requested `.md`; status-report skill default is HTML — override honored, flagged here. Waiting for instructions._
