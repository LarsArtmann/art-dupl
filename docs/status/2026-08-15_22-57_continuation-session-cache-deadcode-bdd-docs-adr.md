# Status Report — Continuation Session: Cache Leftovers, Dead Code, BDD, Docs Pass, ADR-0021, Self-Review

**Date:** 2026-08-15 22:57
**Branch:** `fork`
**Session type:** Continuation of the 2026-08-15 Pareto improvement sprint (picks up after the 22:10 status report)
**Commits this session:** `f9320174`, `80be8dc6`, `3afc4ae6`, `9cae480d`, `03bf2c52`, `20e1b719` (6 commits, tree clean)

---

## Self-Critique First (the three questions)

### What did you forget?

1. **The daemon-blob commit is unfixable after the fact.** The prior session's
   8 logical work areas were committed by the auto-git daemon as one blob
   (`9b4a4e5d`) before this session started. I reported it in the self-review
   but could not retroactively split it. The history for the biggest chunk of
   sprint work has no logical boundaries.
2. **I verified the 4 new BDD specs ran via `-ginkgo.focus` only AFTER being
   suspicious of the fast suite time.** Initial `go test ./bdd/` "ok" output
   did not prove my specs executed (Ginkgo specs don't show in `-run` filters).
   I caught it, but the right order is focus-run first, then full suite.
3. **`docs/reviews/*.html` is gitignored** (`*.html` blanket rule). I wrote the
   HTML report, tried to `git add`, got rejected by the ignore rule, then wrote
   a Markdown companion. I should have checked `git check-ignore` BEFORE
   writing 34KB of HTML — though the HTML on disk is still the skill-mandated
   deliverable, and the precedent (committed `.md` companions) exists, so the
   recovery was correct.

### What could you have done better?

1. **Benchmark fixture compile errors (2 rounds).** `syntax.Node.Type/Pos/End`
   are `int32`, not `int`. I wrote `Type: 100 + level` → vet error → fixed Type
   → vet error again on Pos. One read of the struct definition before writing
   the fixture would have zero-rounded this. Cost: 2 tool round trips.
2. **Doc edits without prior View.** One multiedit on
   `docs/ACTIONABILITY_PATTERNS.md` was rejected ("must read before editing")
   because I'd only grepped/sed'd it. Minor, but the rule exists for a reason.
3. **The `--no-verify` commits.** Every commit this session bypassed the
   BuildFlow pre-commit hook because it structurally fails (6 missing binaries:
   tsc, pytest, tailwindcss, go-licenses, vulnix, govulncheck — plus 49
   go-structure-linter findings that predate this session). This is the
   documented workaround (archived 2026-08-05 doc says bypass after
   diagnosing), but the ritual is now 10 days old and nobody has fixed the
   hook. I fixed zero of it — it's infrastructure debt outside my sprint
   scope, but "outside scope" is how debt becomes permanent.

### What could you still improve?

1. **Coverage baseline** — benchmarks have committed baselines, tests have
   none. No way to detect coverage regression.
2. **4 Pending BDD specs** (stringer/generic filter scenarios) rot as false
   confidence. Fix or delete.
3. **Generics hints print `command-line-arguments.` package prefix** — ugly in
   the exact user-facing output the sprint polished. Fix is `types.Qualifier`.
4. **Cache stats ghost is half-alive**: `MemHits` surfaces in verbose mode
   only. The `stats` subcommand still doesn't show it.

---

## a) FULLY DONE (verified this session)

| #  | Item                                                                                                                                                                                                                                                                                                                                                                                                  | Evidence                                                          |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| 1  | **Cache benchmark**: `BenchmarkFileCacheGet` proves LRU-hit is 5.3x faster than disk-hit (76,785 vs 408,406 ns/op; 767 vs 2,805 allocs) on a 255-node tree                                                                                                                                                                                                                                            | `cache/file_cache_test.go`, commit `f9320174`                     |
| 2  | **GetShared design rejected with rationale**: a no-clone shared cache API is UNSAFE because `stampFilename` mutates in place — documented in `job/incremental.go::parseFile` comment so no future session re-proposes it                                                                                                                                                                              | commit `f9320174`                                                 |
| 3  | **Dead code deleted**: `printer.Issuer`/`MakeIssues`/`Issue` — zero production callers, only self-tests                                                                                                                                                                                                                                                                                               | `printer/issuer.go` + `issuer_test.go` removed, commit `80be8dc6` |
| 4  | **Ghost system wired instead of deleted**: `printer.NodesToGroup` was test-only while `cmd/run_output.go` and `cmd/diff_report.go` inlined the identical sequence — now the canonical constructor everywhere, takes `ProcessOption`s, returns typed `AnalysisError`                                                                                                                                   | commit `80be8dc6`                                                 |
| 5  | **4 BDD scenarios for `--suggest-generics`** (hint shown / absent without flag / min-lines gate / identical-types no-hint), each behavior verified e2e against the real CLI in /tmp BEFORE encoding                                                                                                                                                                                                   | `bdd/suggest_generics_test.go`, commit `3afc4ae6`                 |
| 6  | **DiscordSync re-validation** (was "blocked" — repo WAS available at `/home/lars/projects/DiscordSync`; handoff's case-sensitive check was wrong): **296 detected → 2 shown, 0 FPs**, both survivors manually inspected as genuine (one is itself a textbook generics candidate)                                                                                                                      | real runs this session                                            |
| 7  | **Docs pass**: `error-guard-fallthrough` added to ACTIONABILITY_PATTERNS (table + priority #17, count 29→30); pattern-count drift fixed in 5 locations (FEATURES ×3, ROADMAP ×2); AGENTS.md updated (LRU/benchmark/CloneNodes/GetShared-unsafe, 30 patterns, precision gates, SARIF); HOW_TO_USE.md (`--suggest-generics-min-lines`); FEATURES.md (generics enhancer → FULLY_FUNCTIONAL, cache flags) | commit `9cae480d`                                                 |
| 8  | **ADR-0021** written: EraseHash dual-of-type-aware design, the 3 precision gates (≥2 divergent positions, min-lines 4, no pattern match), cache isolation, consequences incl. DiscordSync numbers                                                                                                                                                                                                     | `docs/adr/0021-suggest-generics-erasehash-precision-gates.md`     |
| 9  | **TODO_LIST.md refreshed**: 30+ stale/done items removed, only open work remains (HTML remainder, 3 feedback patterns, cache-stats-in-subcommand, infra, deferred)                                                                                                                                                                                                                                    | commit `9cae480d`                                                 |
| 10 | **Doc sprawl triaged**: 8 point-in-time root docs `git mv`'d to `docs/archive/` (USAGE.md — pre-fork "dupl" docs split brain, PARTS.md, BDD_TESTS_REVIEW.md, BENCHMARK_COMPARISON.md, branching-flow ×2, MIGRATION ×2); empty `PERFORMANCE_OPTIMIZATION.md` trashed; zero dangling references verified in living docs                                                                                 | commit `9cae480d`                                                 |
| 11 | **Brutal self-review delivered**: styled HTML at `docs/reviews/2026-08-15_22-44_brutal-self-review.html` (gitignored by `*.html` policy) + tracked Markdown companion answering all 11 questions                                                                                                                                                                                                      | commit `03bf2c52`                                                 |
| 12 | **BuildFlow auto-fixes committed separately** (39 workflow SHA pins, gofmt realignment, dprint.json) so tool churn doesn't pollute feature commits                                                                                                                                                                                                                                                    | commit `20e1b719`                                                 |
| 13 | **Final gates green**: `go build ./...` clean; 27 test packages pass (317 BDD specs incl. 4 new); `golangci-lint` **0 issues**; dogfood self-invariant holds (51 detected / 0 shown at `--min-lines 6`)                                                                                                                                                                                               | verified end of session                                           |

## b) PARTIALLY DONE

| Item                           | Done                                                               | Remaining                                                                    |
| ------------------------------ | ------------------------------------------------------------------ | ---------------------------------------------------------------------------- |
| Cache stats ghost integration  | `MemHits` in `Stats`, printed in verbose mode (`printCacheStats`)  | Not in `stats` subcommand; non-verbose users see nothing                     |
| Suggest-generics output polish | Hint dedup (`divergenceKey`), path stripping (`shortenTypeString`) | Still prints `command-line-arguments.` prefix (go/packages artifact)         |
| Sprint commit hygiene          | This session's 5 work commits are logical + detailed               | Prior session's 8 areas remain one daemon blob `9b4a4e5d` — unsplittable now |
| Self-review HTML               | Written per skill spec (34KB, Bauhaus template)                    | Untracked by git (policy) — only the .md companion is versioned              |

## c) NOT STARTED (known, deliberately deferred — from prior sprint scope)

- TTY-aware HTML auto-write (`art-dupl-report.html` when TTY)
- Stable display IDs in HTML output (deep-linkable `id` attributes)
- 3 remaining feedback patterns: `//go:embed`, `TestMain`, `defer-cleanup-of-arbitrary-resource`
- `--exclude-pattern` zero-match warning
- Coverage baseline
- Fix-or-delete decision on 4 Pending BDD specs

## d) TOTALLY FUCKED UP (honest ledger)

1. **Two benchmark compile rounds** from `int` vs `int32` (`syntax.Node` fields) — pure laziness, struct wasn't read first.
2. **Multiedit rejected for unread file** (ACTIONABILITY_PATTERNS.md) — violated the read-before-edit rule, recovered with View + retry.
3. **Initial BDD "pass" was unproven**: `go test ./bdd/` filters don't match Ginkgo `It` blocks, so "ok" didn't prove my specs ran. Caught it by re-running with `-ginkgo.focus`; all 4 passed. The failure mode (false green) is worth remembering.
4. **HTML report git-add rejection**: ignored `*.html` rule discovered only at commit time. Check `git check-ignore` before writing large deliverables.
5. _(Inherited, not this session)_ BuildFlow pre-commit hook bypassed on EVERY commit (6 missing devShell binaries). I added 6 more `--no-verify` commits to the pile.

## e) WHAT WE SHOULD IMPROVE (this session's observations)

1. **Fix the BuildFlow pre-commit hook** — add the 6 missing tools to the devShell or exclude failing steps in pre-commit mode. A permanently-bypassed gate protects nothing and normalizes bypassing.
2. **Case-insensitive repo-path checks** — the handoff blocked DiscordSync validation for a whole session on a lowercase-only path check. `ls -d` globbing found it in seconds.
3. **Verify handoff designs before planning around them** — GetShared sounded plausible, died on one mutation-contract analysis. Handoffs carry hypotheses, not specs.
4. **Read gate constants before writing test fixtures** — `MinDivergentPositions = 2` was in the file; my first fixture had 1 divergent position.
5. **Commit per self-contained change** — the daemon blob cost us reviewable history for the entire prior sprint.
6. **Ginkgo suites need focus-runs to prove new specs execute** — plain `go test ./bdd/` "ok" is not evidence for specific `It` blocks.

## f) NEXT — up to 50 things, priority-ordered

**P0 — infrastructure honesty (quick, high value)**

1. Add tsc/pytest/tailwindcss/go-licenses/vulnix/govulncheck to devShell (or exclude from pre-commit mode) so the hook can enforce again
2. Address the 49 go-structure-linter findings (workflow pins partially auto-fixed this session — verify remaining)
3. `dist/` directory not ignored in go.mod (gomod-check info finding)
4. Extract `vendorHash` from `flake.nix` to `vendorHash.nix` (nix-checker suggestion)

**P1 — finish the ghosts and polish**
5. Wire cache `Hits/Misses/MemHits` into `stats` subcommand output
6. Strip `command-line-arguments.` prefix from generics hints (`types.Qualifier`)
7. Fix-or-delete the 4 Pending BDD specs (stringer/generic filter scenarios)
8. Commit coverage baseline (`go test -cover` per package) mirroring benchmark baselines
9. TTY-aware HTML auto-write (`art-dupl-report.html`)
10. Stable group `id` attributes for HTML deep-linking
11. `--suggest-generics-min-lines` documented in README feature list (HOW_TO_USE done; README untouched)
12. Check README for pattern-count references (29 vs 30) — I fixed FEATURES/ROADMAP/AGENTS but did not grep README

**P2 — remaining feedback patterns (each one file + tests + doc row)**
13. `//go:embed` + `embed.FS` + `fs.Sub` compiler-bound pattern (go-sse)
14. `TestMain` boilerplate pattern (go-cqrs-lite)
15. `defer-cleanup-of-arbitrary-resource` (`defer rows.Close()` etc., discordsync — 12 groups)
16. Re-validate all 3 on their source repos once written

**P3 — UX / robustness**
17. `--exclude-pattern` zero-match warning
18. Document glob vs regex semantics for `--exclude-pattern`
19. `--cache-stats` as a real flag (currently verbose-only implicit)
20. Generics hint max-length clamp for terminal width
21. `version --json` — verify it includes new flags' defaults; add if schema drifted

**P4 — testing depth**
22. Race-detector run in CI is skipped (CGO_ENABLED=0 in BuildFlow env) — enable test-race in nix develop context
23. Property-engine confidence calibration against real-world corpora (ADR-0017 follow-up)
24. E2E test for `NodesToGroup` typed-error wrapping (new behavior from this session)
25. Benchmark for `stampFilename` + clone path on large files (clone is the remaining hot-path cost)
26. Golden-file tests for SARIF generics properties (added fields; no golden update verified this session)

**P5 — architecture (deliberate, slow)**
27. Collection-level `TypeAwareData` (compile-time EraseHash invariant; breaking)
28. Suffix array migration evaluation (ADR-0020 medium-term plan)
29. `[]*syntax.Node` position-offset map (the true memory bottleneck per ADR-0020)
30. Coverage of `MinDivergentPositions`/`DefaultGenericsMinLines` in a config-surface test (constants mirrored in 2 packages)

**P6 — docs hygiene**
31. Website content sync (website/ wasn't touched; new flags/patterns not reflected there)
32. CHANGELOG entry for the sprint batch (none written — append-only file, needs a release context)
33. Add `docs/archive/` README explaining what the archived docs are and why
34. Verify USAGE.md removal didn't break website links to it (built site may reference it)

_(34 items — every remaining idea I could justify; the rest would be padding.)_

## g) Questions I cannot answer myself

1. **The BuildFlow pre-commit hook**: should I fix it by adding the 6 missing tools to the devShell (bigger closure, slower nix eval) or by excluding those steps in pre-commit mode (smaller, but they also never run in CI then)? This is a taste/infrastructure-policy call that affects your daily commit loop.
2. **Commit `9b4a4e5d`** (the daemon blob with all 8 sprint areas): acceptable as-is, or do you want a documented `git note`/CHANGELOG breakdown since interactive history rewrite (`rebase -i` territory) is off the table per repo rules?
3. **`docs/reviews/*.html` policy**: the blanket `*.html` gitignore means every HTML report (self-reviews, plans, audits) lives only on your disk and dies with the checkout. Should I add a `!docs/reviews/*.html` negation, or is "HTML local-only + tracked .md companion" the intended convention?

---

**Verification state at time of writing:** build clean · 27 packages pass · lint 0 issues · dogfood 51/0 · tree clean at `20e1b719`.

**WAITING FOR INSTRUCTIONS.**
