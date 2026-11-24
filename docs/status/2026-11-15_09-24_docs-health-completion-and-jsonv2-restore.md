# Status Report: Docs-Health Wave Completed + jsonv2 #17 Tree Restored — Session Report

**Date:** 2026-11-15 09:24 CET
**Scope:** Second half of the 2026-11-15 docs-health wave (part 1: `docs/status/2026-11-15_08-18_docs-health-cadence-wave-and-2026-0x-docs-triage.md`). This pass: executed the full planning/review/feedback triage (17 files), archived 14 docs across 4 directories with per-row verdicts / ledger-marked accounting / resolution banners, and — once the live `buildflow --fix` pipeline exited — restored the 6 jsonv2-poisoned files (recurrence #17), returning the whole tree to green.
**Tree state at report time:** GREEN. jsonv2gate green, full suite 34 pkgs green (`-count=1`), docs-gates 5/5 PASS, docs count gates 5/5 PASS, boundary fast gate OK (10s). The re-poisoning vector returns the next time a `buildflow --fix` runs (Q1 open).

---

## a) FULLY DONE

1. **jsonv2 recurrence #17 state RESTORED, verified, landed** — after `pgrep -af buildflow` showed the pipeline gone: per-file source discipline (`91419533` for `internal/jsonutil/jsonutil.go`, `config/config_migrate.go`, `pkg/enum/enum.go`, `pkg/artdupl/types.go`; `3cf34a23` for `baseline/baseline.go`, `config/enum_helpers.go`), every diff first verified as a PURE v2-import swap (zero feature work at risk — checked, not assumed), restored via `git restore --source=… --worktree --staged`, content-verified by same-chain greps (the one `encoding/json/v2` grep hit in jsonutil.go is the v1-only rationale COMMENT, byte-identical in the source), worktree-vs-source diffs = 0 on all six. Gates: `go test -count=1 ./internal/jsonv2gate/` green, full suite 34 pkgs green, boundary fast profile OK. AGENTS.md stamp + #17 entry updated (restore landed; Q1 still open).
2. **All 8 planning docs dispositioned and archived** to `docs/planning/archived/` (new dir) with `MANIFEST-2026-11-15.md`: two empty 2026-03-20 stubs (own resolution banners); actionability-phase-2 (3 Pareto-tier rows struck: phases 1-2 shipped, HTML badges/plumbing annotation won't-implemented — superseded by the 42-label classification surface + plumbing's deliberate machine-minimalism); ZERO-FP-FN extractability engine (50 M-rows struck per-row — M22 superseded by the shipped D1 threshold knob, M24 parked behind the v1.0 cut; 34 accounting tables ledger-marked); core-lane BuildFlow inversion (28 tables ledger-marked + execution banner: lane live 2026-09-25, D1 2026-10-05, D2 on BuildFlow master; residuals routed to TODO_LIST/BuildFlow tracker); both SUPERB plans (44 M-rows struck per-row; 52 accounting/record tables ledger-marked; SUPERB-v2's M05 superseded — v0.8.0 never tagged, v0.9.0 shipped 2026-10-06).
3. **`v0.8.0-release-park.md` defused and archived** — it still read "awaiting go" for a release that never happened (an executable dead sequence); now carries a SUPERSEDED banner + both execution rows struck as superseded, archived as the standing release-recipe reference.
4. **Both brutal self-reviews archived** to `reviews/archived/` — their 2026-10-05 resolution banners already existed; the three named-open rows struck with verdicts (clone-type consolidation → ROADMAP parked; Data→View rename → NOT-DO by design, the parallel printer types are legitimate format DTOs per SPLIT-BRAIN; `withLock`/`diffStatTable` focused tests → parked, transitive coverage real). The July review's accomplishments table ledger-marked.
5. **All three `feedback/new/` files resolved into `feedback/done/`** — every tool suggestion verdict-ed inline against shipped code: accept-directive (shipped 2026-07-25), TTY-aware HTML (shipped 2026-10-24 `resolveHTMLOutput`), `--out` flag (shipped as `--html-out`), stable HTML ids (shipped as `anchor_id`), extraction-cost heuristic (shipped as the ADR-0017 property engine), test-heavy default (shipped as `--test-threshold`); stale "Open" status cells in the t50 suggestion ledger corrected to their shipped dates; two genuinely-open asks routed (same-function tag + unknown-category fallback → ROADMAP rows; DescribeTable recipe → TODO_LIST owner row). Session metric tables ledger-marked as records.
6. **boolblind analysis archived** to `analysis/archived/` (settled DO-NOT-IMPLEMENT decision record; its findings/tradeoff tables ledger-marked) and **go-output-t3 feedback moved to done/** (its 2 record tables ledger-marked — moved before gating, then gated clean in place).
7. **Two live docs dispositioned without archiving, honestly**: `decision-sheet` (v0.8.0 question CLOSED with an inline OUTCOME note — de-facto hold, superseded by v0.9.0; options table ledger-marked; g1/g2/g3/binfmt remain open owner questions so the file stays) and `modularization/2026-06-20_PROPOSAL.html` (NOT-EXECUTED annotation banner: single go.mod, no `internal/engine`, premises overtaken by the toolsdk provider; stays in its proposal workspace).
8. **Routing rows landed**: ROADMAP +2 (same-function clone tag; unknown→AST-type category fallback — both from feedback residuals); TODO_LIST +5 (empty-snapshot guard C, decision-sheet owner answers, DescribeTable recipe P2 #25, bandit INFO lane, docs-health annotator upstream fixes).
9. **My own 08:18 report's freshness debt cleared** — its §f list instantly became the newest report (the self-perpetuating report loop); the 3 unharvested items (bandit lane, annotator upstream ×2) harvested as TODO_LIST rows; the 08:18 report's §b1 corrected in place to record same-day execution.
10. **Three older reports' audit-all residuals annotated** (2026-06-17 go.mod bump → superseded by the settled 1.27.1 pin policy; 2026-09-29_03-01 v0.8.0 build verify → superseded by v0.9.0; 2026-10-05_21-46 docs-health calendar → done 2026-11-14/15) and **two path citations repaired** in `2026-09-28_23-50_superb-execution-m00-m11-landed.md` that my file moves had broken (the annotate gate's path checker caught them — the moved-file breaks-cited-paths class is now known).
11. **Final gate state, all mechanized**: docs-gates 5/5 PASS (freshness, audit-all 0 unharvested across 550+ reports, annotate 0 findings, routing, cadence 1 unarchived dated report = today's, healthy), docs count gates 5/5 PASS, full suite 34 pkgs green, boundary fast gate OK.

## b) PARTIALLY DONE

1. **Formatter-stability of today's ~120 per-cell strikes is designed-for but unproven**: the canon's "annotate → check-rows → format → check-rows AGAIN" leg was skipped because running dprint mid-session against the then-live pipeline was exactly the race this session avoided. All table strikes use the verified formatter-stable per-cell form, but no format pass has run over them.
2. **The jsonv2 defense layering**: tree restored and green, but `go-auto-upgrade` is NOT skipped — the next `buildflow --fix` re-poisons all six files. Restore-only remediation continues until Q1 lands.
3. **The undated legacy reports (~29) + `2026-10-05_pass-verdicts/` fate**: unchanged (gate-neutral, TODO row updated with today's context).
4. **check-rows leveling pass over the 160 pass-3-wave archived files**: not touched (today's 12 + 14 moves were gated strict; the older 160 were not).

## c) NOT STARTED

Everything carried in the 08:18 report §c plus the 07:39 report's P0/P1 items that remain owner-gated or pipeline-gated: the `go-auto-upgrade` skip decision (Q1), pre-commit hook-bypass root-cause, dprint 211-file decision (Q2), cue-vet fleet violations, `arch-lint.yml` go-version pin, quiet-machine full-pipeline verification, release of the markup-identity work, go-paperless v0.4.3, branch protection, G8/G9/G11/D3, gogenfilter fleet sweep, T07b upstream prototype. None of these were touched by this pass (per scope discipline).

## d) TOTALLY FUCKED UP

1. **My ledger-marking script sprayed 141+163+164 markers across 3 plan files** — the detector condition was loose enough to fire above nearly every table ROW, not just table headers. Caught immediately by sanity-checking the marker counts against table counts; all three files reverted from HEAD (`git restore`, my own changes only) and redone with a real table detector (header line followed by a `| ---` separator). Lesson recorded: verify insert-count vs structure-count BEFORE writing, and prefer revert-and-redo over incremental cleanup when the shape is globally wrong.
2. **An infinite loop in the same script family** (insert-during-`enumerate` over a growing list) hung the first SUPERB run — killed; the file was untouched on disk (the script writes only after all passes), which is the only reason the revert was trivial. Write-at-end atomicity saved me from myself.
3. **One annotate key silently struck the WRONG line** (§f item 1 of the 08:16 report): the `1@marker-log line` key matched the §e prose item containing the same words instead of the table row — both share the number space and the anchor. Caught by check-rows reporting 44/45 rows struck; repaired by giving the §e item its own honest verdict and striking the intended table row. This is a tool gap (`N@anchor` keys don't warn on multi-line matches), now a TODO upstream row.
4. **`| tail -1` masked a gate exit code once** and the archive `git mv` ran before the SUPERB plan's gates were actually green (pipe eats the exit status). Caught when the follow-up gate failed on the already-moved file; fixed in its archived position. The gate-then-move discipline needs rc-checked invocations, not piped tails.
5. **A path-citation break caused by my own moves**: annotating + archiving files invalidated `path:` citations in an older archived report's verdicts (the annotate gate flagged both). Fixed by updating the citations to the new paths with "(moved 2026-11-15)". Lesson: bulk moves must grep the docs tree for citations to the moved paths.
6. **go-output feedback was moved to `feedback/done/` BEFORE its gate ran** (the earlier gate failure was from the pre-move path) — caught in the same command block and gated clean in its new location. Ordering slip, self-caught, no damage.

## e) WHAT WE SHOULD IMPROVE

1. **`status-items` must strike table rows per-cell** (upstream docs-gates): the tool emits whole-row spans (`~~| 2 | … |~~ verdict`) — the exact dprint-mangle class the 2026-11-05 canon bans; 101 rows needed post-hoc mechanical conversion this session. TODO row exists.
2. **Warn on ambiguous `N@anchor` keys** (same upstream): when a key matches multiple candidate lines (prose vs table sharing a number), refuse or list candidates instead of silently striking the first.
3. **Table detection = header + separator line**: the reliable signature (used in the fixed script) is `|` line followed by a `| ---` line; any looser heuristic will misfire at corpus scale. Worth encoding in the shared batch tooling.
4. **Bulk moves must run a citation sweep** (`rg '<old-path>' docs/`) before `git mv` — the annotate gate's path checker caught this today, but only because the moved files' citations pointed INTO gated verdicts; pointers in ungated prose would rot silently.
5. **The self-perpetuating report loop**: every status report with a §f list creates instant freshness debt against itself. The fleet convention handles it (harvest the newest §f immediately), but a `--pre-flight` check BEFORE writing the report ("would these §f lines flip historical items? what's unharvested?") would let a session write the report with dispositions already in place.
6. **Annotator `--verify` should print the full post-image** per key (matched line + resulting line). Both wrong-line and wrong-shape mistakes today would have been visible before any write.
7. **Revert-early is cheap**: the 3-file marker spray cost one `git restore` + a corrected script (~2 minutes) because the files were reverted instead of patched. The instinct to "fix forward" on mechanical damage is wrong at corpus scale when the damage is uniform.
8. **Daemon interplay**: the auto-commit daemon committed mid-pass twice (my git archaeology for the pre-edit line had to go back TWO commits; ROADMAP needed a re-read after an external mtime bump). Sessions doing precise line-math against files should re-read anchors after any daemon commit boundary.

## f) Up to 50 Things To Get Done Next (prioritized)

**P0 — this session's direct residues**

1. Answer Q1 (go-auto-upgrade: skip / upstream exclusion / migrate) — the tree is green NOW but the next `buildflow --fix` re-poisons; every day of delay is a future restore.
2. Add `go-auto-upgrade` to `.buildflow.yml` `skip_steps` per the Q1 answer (precedent: the `golangci-lint-auto-configure` skip).
3. Prove formatter-stability of this wave's ~120 per-cell strikes: run one dprint/treefmt pass over the 14 newly archived files + 12 status reports, then `check-rows` them again (the skipped canon leg).
4. Answer Q2 items from the 08-18 report: decision-sheet g1/g2/g3/binfmt (TODO rows now carry them) — unblocks M17's GO-AWAY gate and the decision-sheet's own archival.
5. Decide the DescribeTable skill recipe (P2 #25, TODO row) — closes the last open row of the t50 feedback ledger.
6. Root-cause the pre-commit hook bypass (`2a9291ba` + #15 are two pass-throughs; reproduce the daemon's exact git env; then the G5 daemon-gate deployment decision can be re-priced).
7. Add a jsonv2gate step inside the BuildFlow pipeline (defense layer independent of hooks).

**P1 — docs-health next wave**
8. Archive the ~29 undated legacy status reports (the subagent-extraction recipe worked well twice now).
9. check-rows leveling pass over the 160 pass-3-wave archived files.
10. Decide `2026-10-05_pass-verdicts/` fate (companion to #8).
11. AGENTS.md size reduction (~375 → ≤220; buildflow finding persists).
12. Run `docs-gates purge` when the struck-TODO threshold trips (PURGE-HEALTHY today; this wave added ~12 struck rows).
13. Upstream the annotator fixes (per-cell table strikes + ambiguous-key warning + `--verify` post-image) — TODO row.
14. Cross-project lesson by commit to crush-config `references/lessons.md`: "file-rewriting repair tools beat commit hooks; exclude per-repo and pgrep for live pipelines before forensics" (the 07:39 report's §f#44, still unrecorded).
15. Add a citation-sweep step to the archive ritual (grep for moved paths before `git mv`).

**P2 — carried release/infra work (owner-gated where marked)**
16. Release the markup-identity + check-json work (release note must say "re-run `art-dupl baseline`" for CacheVersion 9) — owner-gated.
17. Decide threshold + test-inclusion in the analyzer policy — owner question.
18. dprint 211-file drift decision (Q2 of the 07:39 report).
19. cue-vet: align `ci.yml` `runs-on` schema; add `concurrency:` to `release.yml`.
20. Drop the patch component from `go-version` in `.github/workflows/arch-lint.yml`.
21. shellcheck SC2034: remove/export `CGONOTE` in `scripts/check-boundary.sh`.
22. bandit lane: exclude INFO chatter (TODO row).
23. pytest lane: point at real tests or skip (collects 0 items).
24. sqlite resume-DB policy: single-instance doc or busy_timeout upstream.
25. `nix flake check --all-systems` (aarch64 lanes unchecked on this host).
26. One full quiet-machine `buildflow --build-mode=full` run to confirm the test-race/nix-build chronic reds are dead post-restore.
27. Rebuild + reinstall the BuildFlow binary (binary 1d474a0 vs HEAD 57e1f83 was flagged in the 07:39 run).
28. `/mnt/buildcache/npm` "no space left on device" (recreate or repoint).

**P3 — carried product work**
29. go-paperless v0.4.3 tag + release (#20, memo exists).
30. Branch protection + failure notifications (#18).
31. Windows exe-start fix: one green windows-lane `TestExitCodes_Process` confirmation (#4).
32. G8: release go-finding `pipeline` Unreleased upstream.
33. G9: retire the windows-lane GOEXPERIMENT env split after one clean verification.
34. G11 (optional, upstream): toolsdk per-value validation hook.
35. D3 (post-soak): jscpd-vs-art-dupl lane-overlap measurement.
36. Fleet rollout of the core-lane pattern to other BuildFlow consumers.
37. Sweep gogenfilter consumers to v3.6.1 (9 repos).
38. T07b: go-finding coverage-evidence channel (upstream prototype first).
39. `emit-suppressed-accepted` default-flip decision (owner answer tracked).
40. M24: v1.0 ZERO-FP/FN readiness audit (parked behind the v1.0 cut, verdict recorded today).
41. Same-function clone tag (ROADMAP row, routed today).
42. `unknown` category → AST-type fallback (ROADMAP row, routed today).
43. empty-snapshot guard C: extend `check-docs-freshness.sh` (TODO row).
44. `--type-aware-load-timeout` flag (ROADMAP-class idea; parked).
45. Real-signal SIGINT spawn test for the type-load phase (parked since 2026-10-24; unit-level ctx test exists).
46. `restore-from.sh` playbook script (parked twice; per-file restore discipline is documented in AGENTS #16).
47. Monthly self-scan cadence: next `scripts/self-scan.sh` window (ledger last entry 2026-10-11; AGENTS tracks the schedule).
48. File-size-check policy (82 findings, worst 1,269 lines — owner call).
49. Dockerfile: adopt multi-stage or delete (unused).
50. Fix the non-UTF-8 bytes in whichever status report crashes byte-exact tooling (found today via UnicodeDecodeError in the ledger pass).

## g) Questions I Cannot Answer Myself

**Q1 — go-auto-upgrade disposition (now the ONLY thing standing between green and the next re-poisoning):** `skip_steps` for this repo, an upstream exclusion mechanism, or is the fleet end-state "everything migrates to encoding/json/v2" (meaning ADR-0024 gets overturned and art-dupl plans a real migration)? The tree is restored and green; without the answer, the next `buildflow --fix` reverts that.

**Q2 — archival depth for proposal-class docs:** today's planning archives used per-row strikes + ledger-marked accounting tables + resolution banners (regime A treatment). For future waves: is that the wanted depth for PLANNING docs too (vs the lighter A-lite grouped banner), and should `docs/planning/archived/` + `docs/analysis/archived/` get their own README regime pages like `docs/status/archived/README.md` and `docs/reviews/archived/README.md` have?

**Q3 — v0.9.1 release intent (carried):** FEATURES now says "v0.9.0 (latest release; the 2026-11-06 work sits in Unreleased)". Do you want a v0.9.1 cut soon (its note must say "re-run `art-dupl baseline`" — CacheVersion 9 invalidates every consumer cache and baseline hash space), or keep riding Unreleased?

---

_Session artifacts (this pass): 6 Go files restored + gates green; 14 docs archived across planning/reviews/feedback/analysis with 118 M-rows struck per-cell, ~115 tables ledger-marked, 7 resolution/superseded banners; 12 suggestion verdicts inline; 2 stale ledger cells corrected; 2 ROADMAP + 5 TODO_LIST routing rows; 3 old-report residuals annotated; 2 broken citations repaired; 1 report §b1 corrected. Gates at close: docs-gates 5/5 PASS, count gates 5/5 PASS, full suite 34 pkgs green, boundary fast OK. Part 1 of this wave: `docs/status/2026-11-15_08-18_docs-health-cadence-wave-and-2026-0x-docs-triage.md`._
