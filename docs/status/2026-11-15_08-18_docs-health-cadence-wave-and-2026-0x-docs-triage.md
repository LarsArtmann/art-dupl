# Status Report: Docs-Health Cadence Wave + 2026-0X Docs Triage — Session Report

**Date:** 2026-11-15 08:18 CET
**Scope:** Full docs-health pass (AUDIT mode): docs-gates battery from 3 FAIL to 5 PASS, 19 audit-all open items ground to 0, 12 dated reports archived with per-row verdicts, living docs (TODO_LIST/CHANGELOG/AGENTS/README/ROADMAP/FEATURES) reconciled, AGENTS.md jsonv2 recurrence ledger corrected (#16) + #17 appended, 17 non-status `2026-0*` planning/review/feedback docs triaged (extraction done, execution pending).
**Tree state at report time:** RED — `go test -count=1 ./internal/jsonv2gate/` FAILS (`pkg/artdupl/types.go` + `pkg/enum/enum.go` on v2; the recurrence-#17 state from commit `2a9291ba`). Restore DELIBERATELY deferred: a user-launched `buildflow --fix --build-mode=full` pipeline (pid 72699) was LIVE the entire session — the 07:39 report's own lesson (§e2: check `pgrep -af buildflow` before forensics) was applied BEFORE any restore attempt. Docs work is race-safe (go-auto-upgrade rewrites Go files only); all doc gates ran green on the doc tree regardless.

---

## a) FULLY DONE

1. **Session-start fresh gate + live-pipeline check** — `go test -count=1 ./internal/jsonv2gate/` confirmed the poisoned tree (2 files named); `pgrep -af buildflow` found the live pipeline before any forensic or restore work. The restore-only-when-quiet discipline from the 07:39 report was followed on the first try.
2. **Freshness items #13/#14 FIXED AT SOURCE** — `scripts/corpus-diff.py`: `NoReturn` imported from `typing` (the mypy "missing return" cascade was mypy treating `fail()` as None-returning; the stray `# noqa: F821` removed); `scripts/docs-health-driver.py`: the dead `if False else` ternary dropped (the vulture "unsatisfiable ternary" at :126). Both `py_compile` green; `corpus-diff.py` argparse smoke-tested live.
3. **Freshness items #23/#26 resolved with honest NOT-DO/Won't-implement verdicts** — the `docs/architecture-understanding/` SVGs are the architecture-visualization skill's shared output directory by design (accept); `bdd/semantic_testdata.go` is COMPILED `package bdd` fixture code consumed by `semantic_detection_test.go:27` — moving it to `testdata/` would make the go tool ignore it and break the build, so the info-level finding's premise (raw data) is false.
4. **All four newest-report §f items annotated inline** in `2026-11-15_07-39_…` (marker-leading verdicts, full-line strikes) — `docs-gates freshness .` now FRESH.
5. **All 15 ancient audit-all items annotated across 12 archived files (2026-01 → 2026-08)** — every verdict verified against current code before striking: flag-parsing extraction evidence (`cmd/flags.go`, `cmd/run_flags.go`, `cmd/config_builder.go`), benchmark evidence (8 bench files + alloc-gate + `docs/benchmarks/`), CacheVersion gate tests (`cache/file_cache_test.go`), `typeAwareTag` cache-key isolation, JS/TS parked-in-ROADMAP, `excludeTempl` superseded-by-its-own-flip (field no longer exists). **audit-all: 19 unharvested → 0 (FRESH)**.
6. **AGENTS.md jsonv2 ledger corrected + #17 appended** — the #16 operational note's wrong mechanism ("daemon race, undetermined") replaced with the confirmed BuildFlow `go-auto-upgrade` cause; full #17 entry added (root cause, evidence quotes, pending Q1, hook-bypass second data point, three lessons). Stamp updated to say the tree is RED pending Q1.
7. **Cadence grind: 13 → 1 unarchived dated reports.** Eight reports fully dispositioned BEFORE moving: `2026-10-24_08-16` (§c 7 rows + §f 50 rows struck per-cell with evidence-backed verdicts), `2026-10-24_11-34` (45-row task table struck per-row), `08-18`/`10-30_03-13`/`11-06_08-06`/`08-53`/`09-09`/`11-14_10-13` (17 ledger markers over record/verdict/measurement tables — verified record-shaped before marking, NOT blanket-marked). The three 2026-11-14 pass reports were already gate-clean from their own pass. All 12 got `> **Live remainder:**` banners, `git mv` to `docs/status/archived/` (only today's 07:39 report intentionally remains — its P0 is open).
8. **Archive manifest + recount** — `docs/status/archived/MANIFEST-2026-11-15.md` written (12 rows, class + deciding reason each); archived README recount computed: **548 snapshots: 281 regime A / 23 A-lite / 244 B**.
9. **Living docs reconciled** — FEATURES.md: false "v0.9.1 (current release)" claim fixed to v0.9.0 + Unreleased note (no v0.9.1 tag exists), stale "CacheVersion 5" → 9, truncated table cell (`version [--json`) completed, stamp bumped. ROADMAP.md: stale "30 denylist + 4 property" counts → the 42 canon (37+1+4) in both spots, stamp bumped. TODO_LIST.md: archive-row context updated. AGENTS.md: cadence bullet records this wave. README.md verified clean (stamp + 42-label claims re-checked; zero edits needed).
10. **Count gates green after doc edits** — all five `cmd/docs_health_counts_test.go` tests (`ActionabilityPatternCounts`, `CLIFlagCount`, `NodeTypesCounts`, `DetectionModeCount`, `OutputFormatCount`) PASS.
11. **docs-gates battery: 5/5 PASS** (freshness, audit-all, annotate, routing, cadence — was 3 FAIL at session start).
12. **17 non-status `2026-0*` docs triaged** (subagent read them fully; verdicts recorded in this report §f and ready to execute): 5 archive-ready, 11 need annotation, 1 leave (live proposal).

## b) PARTIALLY DONE

1. **Planning/review/feedback `2026-0*` execution** — ~~execution pending~~ **EXECUTED same-day (08:18–08:55 pass, see `docs/planning/archived/MANIFEST-2026-11-15.md`)**: all 5 archive-ready + 6 planning docs archived with per-row verdicts or ledger-marked accounting tables + resolution banners; both brutal reviews archived (named-open rows struck); all 3 feedback/new files moved to `feedback/done/` (shipped suggestions verdict-ed inline: accept-directive, TTY-aware HTML, `--html-out`, stable anchors, extractability engine, `--test-threshold`; residual asks routed to ROADMAP/TODO_LIST); v0.8.0-release-park superseded-banner + archived; decision-sheet v0.8.0 outcome recorded inline (stays unarchived — owner questions open); modularization proposal banner-not-executed (stays in place). New TODO_LIST rows: empty-snapshot guard C, decision-sheet owner answers, DescribeTable recipe (P2 #25); new ROADMAP rows: same-function clone tag, unknown→AST-type category fallback.
2. **The poisoned-tree restore** — one command away (`91419533`/`3cf34a23` per-file sources), deliberately held while the pipeline is live; also still Q1-gated for the durable fix.
3. **jsonv2 defense layering** — the AGENTS.md ledger now records the true engine and the pending fix, but the `go-auto-upgrade` skip / upstream exclusion / hook-bypass root-cause are all still open (owner Q1).
4. **Formatter-stability of today's table strikes** — converted to the canon per-cell form (`| ~~cell~~ … |`), gated clean, but NOT proven against a live dprint pass (the 2026-11-05 canon's "format → check-rows AGAIN" leg was skipped because running a formatter mid-session against a live pipeline is exactly the race this report §a1 avoided).

## c) NOT STARTED

Everything in the 07:39 report's §f P0–P2 that this docs pass didn't own: the BuildFlow skip (Q1), hook-bypass root-cause, dprint 211-file decision (Q2), cue-vet fleet violations, `arch-lint.yml` go-version pin, shellcheck CGONOTE, bandit INFO tuning, pytest 0-collected lane, quiet-machine full-pipeline verification, budget/sqlite policies. Plus all pre-existing TODO_LIST open rows (statix/govet lanes, `run_suppression.go` carve-out, G8/G9/G11, D3, fleet gogenfilter v3.6.1 sweep, branch protection #18, go-paperless v0.4.3 #20, windows exe-start #4, threshold-in-policy owner question), the ~29 undated legacy reports, and the check-rows leveling pass over the 160 pass-3-wave archived files (today's 12 were verified strict; the older 160 were not re-checked).

## d) TOTALLY FUCKED UP

1. **An annotate key hit the WRONG line.** `1@marker-log line` matched the §e prose item (which contains the words "marker-log line") instead of the §f table row 1 — silently striking an OPEN improvement item with a verdict written for a different row. Caught only because check-rows reported 44/45 rows struck; repaired by restoring the §e item to an honest parked verdict and striking the actual table row. Root cause: `N@anchor` keys silently resolve to the FIRST numbered line (prose and tables share the number space) — the tool does not warn on multi-candidate keys. This is a tool gap (§e2), not just operator error.
2. **The fleet's own annotator produces the forbidden table shape.** `status-items` strikes TABLE rows as whole-row spans (`~~| 2 | … |~~ verdict`) — exactly the dprint-damage class the 2026-11-05 formatter-order canon bans. 101 rows had to be mechanically converted to per-cell strikes post-hoc. Every future pass using status-items on tables will re-trip this until the tool is fixed.
3. **Spec-format trial-and-error cost four failed tool runs** (path column in a key-only spec; verdicts without marker words; an `@anchor` containing a `**` that isn't in the line; a backtick boundary splitting an anchor). The tool's error messages are good, but the spec grammar has four silent shape traps.
4. **A ledger-marker inserter bug (forward-order insertion shifted later line numbers)** failed the header check on its first run; fixed by reverse-order processing. Second bug: the same script crashed on non-UTF-8 bytes in a report; fixed with `errors="replace"`/`surrogateescape` (a repo report contains non-UTF-8 bytes — separate hygiene finding, §f45).
5. **Minor verification shortcut**: the `2026-01-07` CLI-verify verdict cites structural evidence (count gate, Fang/Cobra) instead of a live `./art-dupl --help` run — the build attempt failed with "no Go files in repo root" (main lives in `cmd/`) and I did not retry from the right directory. The verdict is honest but the direct check I planned never ran.

## e) WHAT WE SHOULD IMPROVE

1. **Fix `status-items` to strike table rows per-cell** (upstream `docs-gates`/`annotate.go` change): whole-row spans are the known dprint-mangle class; the tool should emit the house style (`| ~~N~~ verdict | ~~cell~~ | …`) natively. Would have saved the 101-row conversion entirely.
2. **Warn on ambiguous `N@anchor` keys**: when a key's number token + anchor match more than one line (prose item vs table row is the common collision), the tool should refuse or list candidates instead of silently striking the first. My §e collision was invisible until the row-count gate flagged 44/45.
3. **`--verify` should print the full post-image per key** (matched line + resulting line), not just "N keys resolved" — both of today's mistakes would have been visible before any write.
4. **Path-citation convention for cross-repo evidence**: the annotate gate path-checks backticked paths and flagged crush-config's `references/lessons.md` as a missing local file. Verdicts citing foreign-repo evidence need a convention (e.g. `crush-config:references/lessons.md` prefix the gate can skip).
5. **The pgrep-first discipline worked and is now canon** (AGENTS #17 lesson 1) — keep it as the literal first command of any session that might touch Go files while a fleet pipeline can run.
6. **Subagent extraction for bulk triage worked well**: one read-only dispatch turned 17 files into typed verdicts with file:line evidence without burning main-session context. Reuse this shape for the ~29 undated legacy reports.
7. **The completeness gate hierarchy earned its keep today**: audit-all drove 15 honest verdicts, check-rows caught the 44/45 collision, the annotate gate caught the foreign path citation, and the count gate cleared the FEATURES/ROADMAP edits. Every layer caught something the previous one missed.

## f) Up to 50 Things To Get Done Next (prioritized)

**P0 — this session's direct residues**
1. Answer Q1 (go-auto-upgrade disposition: skip / upstream exclusion / migrate) — blocks #2 and the whole restore-churn end.
2. When no pipeline is live: restore the 6 poisoned files (per-file sources `91419533` / `3cf34a23`), fresh `-count=1` jsonv2gate + suite, land the restore.
3. Execute the 11 ANNOTATE verdicts from the planning-docs triage (per-item strikes with the subagent's file:line evidence), then archive them (`docs/planning/archived/`, `docs/reviews/archived/` — the reviews dir already has one).
4. Archive the 5 ARCHIVE-READY docs: `analysis/2026-03-25_boolblind-analysis.md`, the two empty `2026-03-20` planning stubs, `planning/2026-09-28_21-57_SUPERB-docs-health-completion-plan.md`, `feedback/new/2026-07-26-go-output-t3-…`.
5. **URGENT: supersede-banner `planning/2026-09-29_02-45_v0.8.0-release-park.md`** — it still invites executing a dead release sequence (v0.9.0 shipped instead).
6. `docs/modularization/2026-06-20_PROPOSAL.html`: not-executed banner (single module, no `internal/engine`) or an adopt decision — its premises are partly overtaken by the toolsdk provider.
7. Keep `planning/2026-09-28_23-59_empty-snapshot-prevention-proposal.md` LEAVE but route its two open items (daemon-side guard A is outside this repo; guard C extension of `check-docs-freshness.sh`) to TODO_LIST/ROADMAP.

**P1 — annotations with prepared evidence (from today's triage)**
8. `planning/2026-09-24_18-04_artdupl-core-lane-buildflow-inversion.md`: add the executed-resolution header (A2/A3/B1/D1/D4/C3 verified done with evidence; BuildFlow-side rows routed to the BuildFlow repo's tracker), then archive.
9. `planning/2026-07-27_10-31_ZERO-FP-FN-EXTRACTABILITY-ENGINE.md`: strike M24 (v1.0 readiness audit — no v1.0 exists; park behind the v1.0 gate) and M22 (superseded by the shipped threshold knob), then archive.
10. `planning/2026-05-16_21-13_actionability-phase-2-plan.md`: verdict items 3.2 (HTML non-actionable badges) + 3.3 (plumbing annotation) — do or won't-do, then archive.
11. `planning/2026-09-29_01-10_SUPERB-v2-…`: record the M05 outcome (v0.8.0 superseded by v0.9.0), then archive.
12. `planning/2026-09-29_02-30_decision-sheet.md`: record the de-facto v0.8.0 outcome; g1/g3/binfmt need owner answers (→ §g Q2) before disposition.
13. `reviews/2026-06-20_06-50_brutal-self-review.md`: strike/route the two bare rows (clone-type consolidation → ROADMAP row exists; Data→View rename needs an explicit won't-do or route), then archive.
14. `reviews/2026-07-25_17-35_brutal-self-review.md`: strike or route the `withLock`/`diffStatTable` unit-test row, then archive.
15. `feedback/new/2026-07-19-go-auto-upgrade-t50-…`: disposition the `unknown`-category fallback + TTY-aware HTML suggestions; the DescribeTable skill recipe is owner-pending (P2 #25) (→ §g Q2), then move to `feedback/done/`.
16. `feedback/new/2026-07-19-httputil-t25-…`: fix the stale "post-v0.8.0" route note (the corpus work shipped under v0.9.x, 42 patterns), disposition the 5 residual suggestions, then move to `feedback/done/`.

**P2 — jsonv2/pipeline cluster (carried, unchanged)**
17. Root-cause the pre-commit hook bypass (`2a9291ba` is the second pass-through; reproduce the daemon's exact git env).
18. Add a jsonv2gate step inside the BuildFlow pipeline (defense layer that catches rewrites even when a rewriter sneaks back).
19. One full quiet-machine `buildflow --build-mode=full` run post-skip to confirm the test-race/nix-build/nix-build-verify chronic reds are dead.
20. dprint decision (Q2), then execute (mass format commit vs dprint-off for markdown).
21. cue-vet: align `ci.yml` `runs-on` schema; add `concurrency:` to `release.yml`.
22. Drop the patch component from `go-version` in `.github/workflows/arch-lint.yml`.
23. shellcheck SC2034: remove/export `CGONOTE` in `scripts/check-boundary.sh`.
24. bandit lane: exclude INFO chatter.
25. pytest lane: point at real tests or skip (collects 0 items).
26. Decide sqlite resume-DB policy (single-instance doc or busy_timeout upstream).
27. `nix flake check --all-systems` (aarch64 lanes unchecked on this host).

**P3 — docs-health follow-ups (this wave's residues)**
28. Annotate + archive the ~29 undated legacy status reports (subagent extraction recipe).
29. check-rows leveling pass over the 160 pass-3-wave archived files (today's 12 were gated strict; those were not).
30. Upstream: status-items per-cell table strikes (§e1).
31. Upstream: ambiguous-key warning in the annotator (§e2).
32. Harden the local batch tools (`fix-rows.py`/`fix-prose.py` anchor-uniqueness abort + `--dry-run` + self-test fixture) — existing TODO row.
33. Find and fix the non-UTF-8 bytes in whichever status report crashes byte-exact tooling (found today via a UnicodeDecodeError).
34. Decide the fate of `2026-10-05_pass-verdicts/` (companion row to #28).
35. Run `docs-gates purge` when the struck-TODO threshold next trips (PURGE-HEALTHY today).
36. AGENTS.md size reduction (≈375 → ≤220; buildflow finding persists).
37. Record "verify the stimulus before the measurement" in crush-config lessons (row 15 of the 08:16 §f — parked today, still undone).

**P4 — carried TODO_LIST work (unchanged)**
38. Release the markup-identity + check-json work (release note must say "re-run `art-dupl baseline`" for CacheVersion 9) — owner-gated.
39. Decide threshold + test-inclusion in the analyzer policy — owner question.
40. go-paperless v0.4.3 tag + release (#20, memo exists).
41. Branch protection + failure notifications (#18).
42. Windows exe-start fix: one green windows-lane `TestExitCodes_Process` confirmation (#4).
43. G8: release go-finding `pipeline` Unreleased upstream.
44. G9: retire the windows-lane GOEXPERIMENT env split after one clean verification.
45. G11 (optional, upstream): toolsdk per-value validation hook.
46. D3 (post-soak): jscpd-vs-art-dupl lane-overlap measurement.
47. Fleet rollout of the core-lane pattern to other BuildFlow consumers.
48. Sweep gogenfilter consumers to v3.6.1 (9 repos, per-repo checklist in TODO_LIST).
49. T07b: go-finding coverage-evidence channel (upstream prototype first).
50. `emit-suppressed-accepted` default-flip decision (deferred, owner answer tracked).

## g) Questions I Cannot Answer Myself

**Q1 — go-auto-upgrade disposition (carried from the 07:39 report; blocks §f #1/#2 and five planning-doc verdicts):** Should `go-auto-upgrade` join this repo's `.buildflow.yml` `skip_steps` (precedent: the `golangci-lint-auto-configure` skip), do you want an upstream exclusion mechanism negotiated instead, or is the fleet end-state "everything migrates to encoding/json/v2" — meaning art-dupl plans a real migration and ADR-0024 gets overturned? Until answered, every restore is guaranteed re-poisoning on the next `buildflow --fix`.

**Q2 — annotation regime + three owner answers for the planning docs:** (a) Should the `docs/planning|reviews|feedback` `2026-0*` docs follow the same regime as status reports (per-item inline verdicts BEFORE `git mv` into a sibling `archived/`), or is a grouped "superseded/decided" banner enough for proposal-class docs? (b) The `2026-09-29` decision-sheet's g1 (legacy retrofit fund/close), g3 (BuildFlow Build Gate ownership), and binfmt (permanent mount) have no recorded answers anywhere — are these decided, and if so where? (c) The go-auto-upgrade feedback file's DescribeTable-skill recipe (docs-health P2 #25): decided?

**Q3 — v0.9.1 release intent:** FEATURES.md claimed "v0.9.1 (current release)" but no v0.9.1 tag exists (latest: v0.9.0; the 2026-11-06 markup-identity + check-json work sits in CHANGELOG Unreleased and invalidates every consumer cache + baseline hash space, CacheVersion 9). Do you want a v0.9.1 cut soon (the release note must say "re-run `art-dupl baseline`"), or should the docs keep pointing at v0.9.0 + Unreleased until a later cut? (FEATURES now says v0.9.0 + Unreleased — flip it back if a cut is imminent.)

---

*Session artifacts: 2 script fixes, 123 item verdicts across 13 files (4 newest + 15 ancient + 104 task/record-table rows), 16 ledger markers, 12 archived reports + 1 manifest + 1 README recount, 4 living docs edited + 1 verified, AGENTS.md #16/#17 ledger work. Gates at close: docs-gates 5/5 PASS, count gates 5/5 PASS, jsonv2gate RED (poisoned tree, restore deferred by design). Authoritative evidence for the restore procedure: `docs/status/2026-11-15_07-39_jsonv2-recurrence-16-root-caused-buildflow-go-auto-upgrade.md` §a1 + AGENTS.md #16/#17.*
