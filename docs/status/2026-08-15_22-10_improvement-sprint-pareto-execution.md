# Status Report: Pareto Improvement Sprint — Execution Phase

**Date:** 2026-08-15 22:10
**Branch:** fork
**Session goal:** "How can we improve this project?" — full READ → UNDERSTAND → RESEARCH → REFLECT → PLAN → EXECUTE → VERIFY cycle with brutal honesty.

---

## FULLY DONE (a)

All items below are implemented, tested, and verified: `go build ./...` clean, `go test ./...` all 30 packages pass, `golangci-lint run ./...` **0 issues** (down from 52 at session start).

### 1. Lint hygiene (52 → 0 issues)

- **Removed banned `tagliatelle` from `.golangci.yml`** (50 violations). The auto-commit daemon had re-added it; `scripts/check-disabled-linters.sh` auto-removed it. Root cause: the guard only runs in Nix checks/CI, not locally pre-commit.
- **Fixed `varnamelen`**: renamed `td` → `typeAwareData` in `job/incremental.go::SetTypeAwareData`.
- **Fixed `nestif` (complexity 11)**: extracted `TextPrinter.writeSuppressionSummary` from `PrintFooter` (`printer/text.go`).

### 2. `--suggest-generics` precision filtering (HIGH #1 — the 1%→51% Pareto item)

Three precision gates added, attacking the verified 12.5% precision problem (3 true / 24 surfaced on DiscordSync):

1. **Minimum divergent positions** (`printer/generics_candidate.go`): `MinDivergentPositions = 2` — single-position type differences are shallow idiom noise.
2. **Minimum line gate** (`printer/clone_processor.go`): groups where ANY instance spans fewer than `--suggest-generics-min-lines` lines (default 4, `0` disables) are never candidates. Mirrors `--min-lines` weakest-instance semantics. New `ProcessOption` API: `ProcessClones(fread, dups, WithGenericsMinLines(n))` — backward compatible, all existing callers unaffected.
3. **Actionability cross-reference**: any group matching a boilerplate pattern (error-propagation, guard-clause, single-call-expression, ...) is disqualified from generics candidacy. Idioms are not generics candidates.

Also fixed the `formatGenericsHint` reversed-pair dedup bug (`A vs B` / `B vs A` collapsed via canonical `divergenceKey`).

**Wiring:** `config.SuggestGenericsMinLines` (+ `DefaultSuggestGenericsMinLines = 4`, validation), `--suggest-generics-min-lines` flag, `SuppressionConfig.GenericsMinLines`, threaded to both `ProcessClones` call sites (`cmd/run_output.go`, `cmd/diff_report.go`). Sync test `TestGenericsMinLinesDefaultsInSync` guards the mirrored config/printer constants (same pattern as `DefaultThreshold`).

**Tests:** all existing generics tests updated to precision semantics; new tests for single-position rejection, line gate (default/zero/custom), pattern disqualification, pair canonicalization.

### 3. `--min-tokens` test coverage (HIGH #2 — shipped flag had ZERO tests)

- `TestShouldSuppressGroup_MinTokens` (6 table cases, mirrors the min-lines test)
- `TestMinCloneTokenCount` (3 cases)
- `bdd/min_tokens_test.go`: 4 Ginkgo scenarios (report without flag, suppress at high value + "0 shown/filtered" footer, pass at low value, reject negative)

### 4. `--type-aware` + `--suggest-generics` warning (MEDIUM)

`warnTypeAwareSuggestGenerics` in `cmd/config_builder.go` — warns that suggest-generics erases types from the hash and takes precedence. Verified end-to-end.

### 5. SARIF generics fields (MEDIUM)

`printer/sarif.go` result properties now carry `generics_candidate` and `generics_hint` (absent for non-candidates). 2 tests.

### 6. NEW actionability pattern: `error-guard-fallthrough` (feedback-driven, HIGH impact)

Covers **both** DiscordSync HIGH-priority feedback findings in ONE precise pattern:

- `if err != nil { writeError(w, r, err, ""); return }` + tail (12 groups — forced by `http.HandlerFunc` void signature)
- `if err != nil { return nil, queryError(err, "unique op") }; return x, nil` (27 groups — the wrapper IS the dedup, unique strings are parameters)

Shape: seq[0] is an error guard (nil-compare + report/return body), remaining statements are ONLY assignments/returns (trivial fallthrough). Rich tails (calls, branches, loops) stay actionable — guarded by test. Multi-statement clones previously escaped ALL single-statement error patterns (`len(seq) != 1`).

Registered after `error-wrapping` in the priority table; label `PatternErrorGuardFallthrough`. Pattern count: 29 → 30. Verified with real Go files end-to-end (both variants suppressed, 0 shown).

### 7. HTML "Detected vs Actionable" summary (MEDIUM)

`htmlprinter` now implements `SuppressionStatsSetter` (previously only `TextPrinter`). Summary section renders "Detected Groups / Actionable (Shown) / Suppressed" cards when groups were suppressed. Also fixes empty-report case: summary renders when only suppressed groups exist. 2 tests.

### 8. Cache follow-ups (MEDIUM — subset, see (b))

- `syntax.CloneNodes` — canonical deep-clone helper; eliminated duplicated `cache.cloneNodes` AND `job.deepCloneNodes`
- `Stats.MemHits` + `lru.hits()` — in-memory hit count surfaced
- Configurable LRU capacity: `cache.NewFileCacheWithMemoryEntries`, `DefaultMemoryEntries` exported, `config.MemoryCacheEntries` (+ `DefaultMemoryCacheEntries = 512`), `--memory-cache-entries` flag, threaded through `NewIncrementalParser`
- `printCacheStats` (verbose) — cache hits/misses/mem-hits now actually reported to users
- Lock-ordering comments on `FileCache` and `lru`; `Set` ownership contract documented
- Tests: `cache/lru_test.go` (MemHits propagation, capacity-1 eviction)

---

## PARTIALLY DONE (b)

### Cache follow-ups — remaining items

- [ ] **LRU benchmark** (`cache/file_cache_test.go`): LRU hit vs empty-LRU `Get` comparison to quantify deserialization-avoidance win
- [ ] **`GetShared` for singleflight**: `Get` deep-clones, then singleflight callers clone again (double clone). Add `GetShared` returning canonical pointer for callers that clone anyway (`job/incremental.go::parseFile`)

### Docs — not yet updated for this session's changes

- [ ] `docs/ACTIONABILITY_PATTERNS.md`: add `error-guard-fallthrough` (pattern count 29→30)
- [ ] `AGENTS.md`: new pattern, `--suggest-generics-min-lines`, `--memory-cache-entries`, `syntax.CloneNodes`, cache Stats.MemHits
- [ ] `HOW_TO_USE.md` / `FEATURES.md`: new flags + pattern
- [ ] `TODO_LIST.md` refresh (see (e) — stale entries found)

### NOT STARTED (from session plan)

- [ ] ADR for EraseHash suggest-generics design (NOTE: ADR-0020 is taken by "algorithmic-alternatives-analysis" — the TODO_LIST's "ADR-0020" ask means a new ADR-0021)
- [ ] BDD scenario for `--suggest-generics` (min-tokens BDD was done instead)
- [ ] Root doc-sprawl cleanup (see (d))
- [ ] Self-review HTML report (`docs/reviews/`) — this Markdown report covers content; HTML rendering pending

---

## TODO NEXT (c)

1. **Verify suggest-generics precision against DiscordSync** (blocked: repo not available locally). Target: 3 real candidates, <5 false positives. The three gates are theory-validated against the feedback doc's examples; real-world re-measurement still required.
2. **Update docs** (ACTIONABILITY_PATTERNS.md, AGENTS.md, HOW_TO_USE.md, FEATURES.md, TODO_LIST.md) — mechanical, ~30min.
3. **Dead code decision**: `printer.Issuer`/`MakeIssues` and `printer.NodesToGroup` have zero production callers. Delete or wire. Report-only here — deletion needs owner intent.
4. **Doc sprawl triage** (see (d)).
5. Remaining cache follow-ups (benchmark, `GetShared`).
6. ADR-0021 (EraseHash design) + `--suggest-generics` BDD.
7. Remaining TODO_LIST MEDIUM items not in this session's scope: TTY-aware HTML output, stable display IDs, `//go:embed` + `TestMain` patterns, coverage baseline, `--exclude-pattern` zero-match warning.

---

## TOTALLY FUCKED UP (d) — mistakes, honestly

1. **The `NewIncrementalParser` signature migration was a mess.** After adding the 6th param, I migrated test callers with a greedy sed regex that DOUBLE-APPLIED (7-arg calls), then "fixed" it with a python regex whose second pattern re-matched the corrected 6-arg forms, re-breaking them. Took 4 rounds across 3 files. Correct approach: `lsp_rename`-style single pass, or one python script asserting final arg-count per call BEFORE writing.
2. **Edit-gutted a test function.** A `multiedit` on `cmd/run_output_test.go` deleted `TestMinCloneLineCount`'s opening lines (my old/new strings were inverted in intent). Caught immediately by viewing + restoring, but it should never have happened: I was appending via replace instead of appending at EOF.
3. **Appended to a nonexistent file.** `cat >> cache/lru_test.go` created the file WITHOUT a package header → package build broke. `ls` after the fact misleadingly showed it. Lesson: verify existence before `>>`.
4. **Batched edits without reading files first** (flags.go, config_validate.go, config.go) → 3 "you must read the file" round-trip failures.
5. **A multiedit accidentally deleted a doc comment** on `classifyCloneType` (replaced comment+signature with bare signature). Restored, but shows replace-based edits near doc comments are risky.

---

## QUESTIONS & DECISIONS FOR LARS (e)

1. **Dead code**: delete `printer.Issuer` + `printer.NodesToGroup`? (zero production callers; possibly kept for SDK consumers?)
2. **Doc sprawl**: `USAGE.md` documents the OLD `dupl` tool (pre-fork name) — split brain with `HOW_TO_USE.md`. Also candidates to archive into `docs/`: `PARTS.md`, `branching-flow-analysis.md`, `branching-flow-findings-table.md`, `MIGRATION_QUICK_START.md`, `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md`, `BENCHMARK_COMPARISON.md`, `PERFORMANCE_OPTIMIZATION.md`, `BDD_TESTS_REVIEW.md`, `WHAT_THIS_PROJECT_IS_NOT.md`. Archive, delete, or keep?
3. **TODO_LIST.md stale entries found**: `shortenTypeString` (done pre-session), `global.out.css` gitignore (done), "ADR-0020" numbering (0020 taken). Confirm I should refresh TODO_LIST.md in the docs pass.
4. **`--suggest-generics-min-lines` default**: I chose 4 (flag default + both Default constants). JSON-config users who don't set the field get 4 via DefaultConfig; CLI default is 4; `0` disables. OK?

---

## WTF & BAD IDEAS FOUND (f)

1. **Ghost system (now half-fixed)**: cache stats (Hits/Misses/MemHits) were tracked since the LRU sprint but NEVER surfaced — `GetCacheStats()` had zero production callers. Now printed in verbose mode. The deeper WTF: we built the whole stats infrastructure without a consumer.
2. **Guard script gap**: `check-disabled-linters.sh` exists precisely because "the auto-committer has re-added exhaustruct and tagliatelle multiple times" (its own comment) — yet it doesn't run locally, so the daemon re-adds and local lint noise accumulates until someone notices. Recommend wiring into the Nix devShell shellHook or a pre-commit hook.
3. **Doc drift**: 21 tracked .md files at repo root; at least one (`USAGE.md`) describes the pre-fork `dupl` CLI. The AGENTS.md "Project Documentation Files" table defines the right homes — root ignores it.
4. **TODO_LIST truth decay**: 2 of 3 HIGH items were partially stale. Re-verified against code before planning; recurring cost every session.

---

## SESSION SUMMARY (g)

- **Scope:** Pareto plan (1%→51%, 4%→64%, 20%→80%) built from TODO_LIST verification + fresh research; executed tiers 1 and 2 fully, tier 3 partially.
- **Delivered:** suggest-generics precision gates (flag + config + API), `error-guard-fallthrough` pattern (both DiscordSync HIGH findings), min-tokens test coverage (unit + BDD), SARIF generics fields, type-aware warning, HTML suppression summary, cache improvements (MemHits, configurable LRU, clone dedup, ghost stats surfaced, ownership/lock docs), lint 52→0.
- **State:** build clean, all tests pass, lint 0 issues, dogfood self-invariant holds (52 groups, 0 shown).
- **Honest gaps:** docs not yet updated, DiscordSync re-validation blocked, report is Markdown not the styled HTML the brutal-self-review skill prescribes, 3 execution fumbles (d) — all self-caught and repaired, none shipped broken.

_Every claim above verified by the commands shown in the session transcript; nothing trophy-cased without a test or e2e run._
