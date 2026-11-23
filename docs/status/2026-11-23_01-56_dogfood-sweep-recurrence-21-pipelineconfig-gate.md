# Status Report — Dogfood Sweep (-t 1 --suggest-generics --type-aware) + Recurrence #21 Mid-Session

**Date:** 2026-11-23 01:56 CET
**Session window:** ~01:20 – 01:56 CET
**Branch:** fork

## Scope — what this report VERIFIED vs NOT

**VERIFIED (fresh command output this session):**
- Dogfood scan run #1: `go run ./cmd/art-dupl --sort total-tokens -t 1 --suggest-generics --type-aware --timing .` → **1360 detected / 6 shown**, total 22.3s (type-load 19.9s), captured in `/tmp/dogfood-scan.out`.
- Re-scan after all edits: **1362 detected / 0 shown** (1057 non-actionable, 305 filtered suppressed), 23.7s total; `-t 5 .` self-scan → 62 detected / 0 shown.
- jsonv2gate red mid-session naming the six poisoned files; green after restore (`go test -count=1 ./internal/jsonv2gate/`).
- Forensics via `git log --diff-filter=D -- .buildflow.yml` → destroyer commits `8ce90a47` AND `57782482` named; per-file diffs vs restore sources verified pure before restoring.
- New `internal/pipelineconfig` gate canary-verified green → red (file removed) → green.
- Boundary gates: fast profile OK (27s), `--full` race suite OK (13s warm), including the new package.
- `buildflow --dry-run` output: both `go-auto-upgrade` and `golangci-lint-auto-configure` "skipped via skip_steps config".
- Pins after restore: `go 1.27.1` (go.mod), unquoted `go: 1.27.1` (.golangci.yml) — and `cmd/go_line_pin_test.go` exists and passed in the full suite.

**NOT verified (and why):**
- **WHO deletes `.buildflow.yml`** — git only names the packaging daemon commits; the deleting agent is not identifiable from inside the repo (owner question).
- **Skip behavior in a REAL pipeline** — verified via `--dry-run` only; the second live pipeline (pid 2140725) was still running at report time and its full behavior was not observed end-to-end.
- **`nix flake check`** — not run; `flake.nix` vendorHash was mid-repair during the session (BuildFlow's own FOD repair; it self-completed to the correct hash `sha256-AFkwR+…` and the tree is clean now).
- **Arch-lint machine run** — not run; the new `internal/pipelineconfig` package imports stdlib only, so no boundary risk was *reasoned*, but `go-arch-lint` was not executed.
- **Detected-count drift attribution (1360 → 1362)** — expected (new helper fragments re-match the 4-fixture variant/fixture helpers at sub-threshold windows) but not verified per-family with `--show-suppressed`; same eyeball-level caveat as the 2026-10-24 session's drift note.
- **Fate of the helper↔395 shared 3-statement prefix** — predicted as a possible residue group; the re-scan shows 0 shown, so it is suppressed or subsumed, but I did not inspect WHICH mechanism.
- **BDD suite explicitly** — covered only via the full race suite pass, not run isolated.

---

## a) FULLY DONE

1. **Monthly dogfood sweep executed** per `scripts/self-scan.sh` cadence: exact user command run on art-dupl itself; all 6 shown groups read and judged (file-by-file).
2. **Extraction G1**: `recordTwoFamilyBaseline(t)` helper in `cmd/check_cmd_test.go`; 11 byte-identical 5-line test setups collapsed to one line each; the 4-fixture variant (double-count test) deliberately kept inline.
3. **Extraction G2**: `detectionMethodsPolicyKey` now delegates to `detectionMethodsToStringSlice` (`cmd/baseline_identity.go` → `cmd/util.go`). Key insight recorded: the `--suggest-generics` flag was a false alarm for extraction — `[]config.DetectionMethod` is assignable to named `config.DetectionMethods`, so no generic is needed; assignability already unifies the two sites.
4. **Extraction G6**: `setVersionFooter(p printer.Printer)` helper next to the existing `setJSONPrinterFilesCount`; three assertion sites replaced (`run_flags.go`, `run_all_modes.go`, `baseline_cmd.go`); `run_all_modes.go` dropped its now-unused `printer` import. Note: the scan grouped only 2 of the 3 sites — the third was found by grep, not by art-dupl.
5. **Accept directives (3 groups, 5 insertions)**: `os.Stat` guard ×2 (SDK boundary forbids `pkg/artdupl` → `cmd` sharing), `dedupLinesForRange` twin ×1 (documented arch-lint cycle), empty-element guard ×2. All proven covering by the 0-shown re-scan.
6. **Recurrence #21 caught, forensicked, and restored**: jsonv2gate failed mid-session with the Duration refusal → six v1 files + go-line pin re-poisoned by a live `buildflow --fix` pipeline because `57782482` had deleted `.buildflow.yml` (SECOND destroyer strike). All nine files restored diff-verified-pure: six from `fb34125f`, pins from `0c486ae9`, config from `c00f3222`; content verified by grep in the same command chain.
7. **Structural fix: `internal/pipelineconfig` gate** — fails the suite when `.buildflow.yml` is missing or loses either skip step, naming the restore ritual; canary-verified before being trusted (house canary rule).
8. **Ledger updated**: `docs/SELF_CLEAN_LEDGER.md` gained two dated sections (the sweep decision table + the #21 forensic record), including the replace_all self-cannibalization process gotcha.
9. **AGENTS.md updated**: rolling "Last Updated" header rewritten for #21 (prior header content preserved where it lives elsewhere — signature-fragment detail stays in the actionability-patterns bullet; #20 keeps its numbered entry), plus a full numbered **Recurrence #21** entry.
10. **All gates green at close**: jsonv2gate, pipelineconfig, boundary fast (build+vet+race on internal/config/cmd + golangci-lint), full race suite, `-t 5` self-scan 0 shown, go-line pin tests, `buildflow --dry-run` skips live.

## b) PARTIALLY DONE

1. **Destroyer investigation**: two deletion commits identified (`8ce90a47`, `57782482`) and the deletion pattern characterized (rides daemon auto-commit packaging, "N changed file(s)" messages don't list the deletion) — but the ACTOR is unknown. Detection is now automated; prevention is not.
2. **Go-line pin stability**: restored and test-verified, but the pin has now flipped in three separate commits across two days (#19, #20-era `a9d1a0f8`, #21) despite `respect_patch_floor` supposedly ending the flip-flop war. Something still rewrites go.mod; the rewriter is not identified. Also discovered `fb34125f` (the #20 remediation commit itself) shipped the flipped pin — "pin tests green" was true only because no full suite ran between flip and restore, NOT because the landed content was correct.
3. **Skip verification**: `--dry-run` proves the config is live, but no real post-fix pipeline has been observed running WITH the restored config yet (the second pipeline pre-dates parts of the session and its steps were not tracked to completion).
4. **Session-start hygiene convention**: I checked `pgrep buildflow` at session start per the #16 lesson but did NOT check `.buildflow.yml` existence — the #20 "skip verified live" claim was trusted instead of re-verified. The new gate converts this into an enforced invariant, but the *convention* (check the file, not the claim) is only now documented, not yet ritualized in any session-start checklist.
5. **Docs loop**: ledger + AGENTS.md updated, but section (f) of this report has not yet been harvested into `TODO_LIST.md`/`ROADMAP.md` (docs-health HARVEST ritual — explicitly left open per the status-report skill).

## c) NOT STARTED

1. **Upstream BuildFlow exclusion mechanism for `go-auto-upgrade`** (owner Q1 from #17/#18) — the skip is local; the fleet-facing policy question is untouched.
2. **Fleet-wide rollout of the pipelineconfig gate** — other repos with ADR-0024-style v1 invariants (or any `.buildflow.yml` overrides) have no equivalent config-existence gate.
3. **Cross-repo destroyer audit** — if the destroyer is a fleet daemon behavior, other repos' `.buildflow.yml`/config files may already be silently stripped; nobody has looked.
4. **December self-scan** — the monthly cadence's next run (this session discharged November, twice over).
5. **`docs-health` cadence pass** — next due 2026-12-14 per AGENTS.md; this report is input to it.

## d) TOTALLY FUCKED UP

1. **The destroyer has now struck TWICE with the same fingerprint** (deletion of `.buildflow.yml` inside a daemon auto-commit whose message doesn't list the file: `8ce90a47` "8 changed files", `57782482` "13 changed files"). Each strike converts the very next pipeline into an unmitigated re-poisoner. The repo spent ~20 minutes tonight in a poisoned state *during active work* — caught only because the sweep ran a config-touching test. Without that coincidence, the poison would have been committed and possibly pushed.
2. **The remediation pipeline itself shipped a flipped pin** (`fb34125f`): the #20 close-out claimed "pin tests green" but the LANDED commit content carried `go 1.27` + quoted `.golangci.yml` line. The claim was about a working-tree state, not the commit that captured it. This is the same "ledger entry committed by the daemon in the same commit as a re-migration is not evidence" class from #9 — now confirmed for remediation commits too.
3. **My replace_all self-cannibalization**: I inserted the G1 helper BEFORE running the block replacement, so `replace_all` rewrote the helper's own body into a self-call — infinite recursion, caught by the suite as a stack overflow. Mechanical error, zero production impact, one wasted test cycle. Root cause: ran the replacement before re-counting occurrences post-insertion. Now documented in the ledger.

## e) WHAT WE SHOULD IMPROVE (brutal self-review: what I forgot / could do better)

1. **Trust nothing point-in-time — verify existence, not claims.** I opened the session with `pgrep buildflow` + jsonv2gate but skipped a 1-command `ls .buildflow.yml`. The #20 header said "skip verified live"; 20 minutes later the file was gone. The lesson is now a gate, but *I* should have caught it at session start.
2. **Order of edit operations**: new code that CONTAINS the block you're about to `replace_all` must be inserted AFTER the replacement (or the replacement scoped away from it). The suite caught my mistake; my review of the edit plan should have.
3. **Residue verification depth**: after the G1 extraction I accepted "0 shown" as proof of no residue. I should have run `--show-suppressed` once to confirm the helper↔395 3-statement prefix fate (suppressed-by-pattern vs subsumed) — the difference matters if the pattern layer ever changes.
4. **Detected-count drift attribution**: I repeated the exact shortcut the 2026-10-24 ledger entry already flagged ("attribution eyeballed, not verified per-family"). Same shortcut, second occurrence — the corpus-diff ritual (`scripts/corpus-diff.py`, pinned worktrees) exists for exactly this and went unused because the corpus here is the repo itself.
5. **Verify the remediation commit, not the remediation.** I restored the pins and ran tests on the WORKING TREE; had I diffed `fb34125f`'s landed content against `0c486ae9` at #20 time (one command), the flipped pin would have been caught eight hours earlier. The #16 per-file lesson ("diff EVERY file against the restore source") should extend to: diff the RESULTING COMMIT after the daemon packages it.
6. **Pre-commit hook coverage gap**: `.githooks/pre-commit` runs only jsonv2gate. The pin flip class has its own cheap test (`cmd/go_line_pin_test.go`) — tonight proved a flipped pin can sit in a committed state indefinitely when only the narrow hook runs. One-line fix, not yet made.
7. **Scratch artifacts near the daemon**: `.buildflow-fod-journal.json` existed untracked in the tree for several minutes mid-repair. Had the daemon swept at that moment, it would have committed buildflow scratch as repo content (the `57782482` packaging class). A gitignore entry is cheap insurance; not yet added.
8. **Second pipeline blind spot**: a second `buildflow --fix` started mid-session (pid 2140725) and I neither tracked its steps nor re-verified the tree after it finished — I relied on the gates run BEFORE it completed. A post-pipeline re-run of jsonv2gate + pin tests would close that window; not yet done.

## f) TOP 30 THINGS TO GET DONE NEXT

*User asked for up to 50; 30 honest items beat 50 padded ones. Sorted by impact. Items 1–8 are the session's direct follow-ups; 9–30 are adjacent work this session surfaced. Most of 12–30 are ROADMAP fuel per the docs-health HARVEST routing rules.*

1. **HARVEST this report's section into `TODO_LIST.md`/`ROADMAP.md`** (docs-health ritual; this list is the input).
2. **Extend `.githooks/pre-commit` to also run `go test -count=1 ./cmd/ -run TestGoLinePin`** (or the whole pin test package) — the flipped-pin-in-a-commit class needs the same mechanical gate jsonv2gate has. Cheap, tonight-justified.
3. **Root-cause the `.buildflow.yml` destroyer**: instrument or audit what packages daemon commits (BuildFlow daemon config? A cleanup heuristic that treats non-default files as debris?). Two strikes with identical fingerprints = a mechanism, not an accident.
4. **Resolve owner Q1**: upstream BuildFlow exclusion mechanism for `go-auto-upgrade`, or declare the skip permanent fleet-wide (see questions below).
5. **Investigate the residual go-line rewriter**: `respect_patch_floor` was supposed to end flips; `fb34125f`+`a9d1a0f8` say otherwise. Find which step rewrites `go.mod` despite the option (BuildFlow `go-mod-normalize`? gvac? something else?).
6. **Run one REAL `buildflow --fix` pipeline under observation** and verify end-to-end that the skip holds and the six files survive (dry-run proof is necessary, not sufficient).
7. ~~**Add `.buildflow-fod-journal.json` (and any `.*-journal.json` buildflow scratch) to `.gitignore`** — prevent the daemon committing mid-repair scratch.~~ done — .gitignore now carries .buildflow-fod-journal.json + *-journal.json (added 2026-11-23 docs-health pass)
8. **Post-pipeline re-verification ritual**: after any live `buildflow --fix` exits, re-run jsonv2gate + pin tests + `ls .buildflow.yml` before trusting the tree (one command chain; document it next to the pgrep lesson in AGENTS.md).
9. **Fleet rollout**: port `internal/pipelineconfig` to other repos with non-default `.buildflow.yml` overrides (the deletion class is almost certainly not art-dupl-specific).
10. **Cross-repo destroyer audit NOW**: check every fleet repo's `.buildflow.yml` presence + skip content and the daemon logs for deletion commits in the same window — determine blast radius before assuming art-dupl is unique.
11. **Clarify the "language mismatch" line in `buildflow --dry-run`** — both auto-configurer steps reported language mismatch AND skip; if mismatch alone blocks them, the skip is redundant (and the destroyer less dangerous); if not, the dry-run output is misleading. One upstream question or a source read settles it.
12. **Confirm the G1 residue fate** with one `--show-suppressed` run (helper↔395 shared prefix: which layer suppresses it?); record in the ledger if it's pattern-layer (a pattern change could resurface it).
13. **Per-family attribution ritual for detected-count drift**: when detected moves on a self-scan, run `scripts/corpus-diff.py --before X --after X`-style comparison on the repo itself instead of eyeballing (second occurrence of the shortcut; make the ritual explicit).
14. **Session-start battery formalization**: `pgrep buildflow` + `ls .buildflow.yml` + fresh jsonv2gate + pin test — write it as a 4-line checklist in AGENTS.md so it's executable, not folklore (the gate covers the config, the ritual covers the rest).
15. **Watch the stat-guard pair (G3)**: if a third `os.Stat`-skip instance appears, that's the trigger point for a denylist pattern (2 instances = over-suppression risk; 3 = corpus).
16. **Re-run the dogfood scan after tonight's pipelines fully exit** to certify 0-shown on the truly-final tree (content unchanged for `.go` files, but the certification should be against the settled state).
17. **Annotate/archive the sibling report** `docs/status/2026-11-23_01-18_follow-up-close-out-recurrence-20-verification-battery.md` per docs-health ANNOTATE: its "skip verified live" verification battery is now superseded by #21's events — inline-correction, never rewrite.
18. **AGENTS.md size watch**: the recurrence ledger grows every strike; the buildflow preflight AGENTS.md-size gate went red once before. Consider moving recurrences #1–#12 (fully historical, mechanism solved) to `docs/status/` archive with a pointer line.
19. **Evaluate bare-accept discoverability**: the 5 new directives have no dead-directive warning (only hash directives do). A future refactor that moves a guarded clone out of range silently uncovers the group — which resurfaces at the next scan (acceptable), but a one-line note in `docs/ACTIONABILITY_PATTERNS.md` or the ledger about this deliberate behavior would prevent a future session re-litigating it.
20. **Unit-test the G2 delegation edge**: pin `detectionMethodsPolicyKey`'s order-independence and empty-set behavior (`""` for empty) explicitly, so the delegation can't silently change policy-key semantics across cache-version spaces.
21. **Baseline compatibility audit**: the sweep changed source in `cmd/`, `syntax/`, `config/`, `pkg/artdupl/` — group hashes for those files shifted, so any consumer baselines (BuildFlow provider runs over art-dupl itself) will read old groups as resolved + new as introduced. Verify `.art-dupl-baseline.json` self-check still passes (the Clone Detection CI lane).
22. **Provider golden check**: `pkg/provider` findings may include the touched files — run `go test ./pkg/provider/ -run TestProviderFindingsGolden -update` check (positions changed ⇒ golden regen decision).
23. **Decide `internal/pipelineconfig` placement** against go-structure-linter advisories (new internal package; stdlib-only tests). If the linter flags it, either rename/re-home or document the advisory as accepted.
24. **Run `go-arch-lint` once explicitly** this week to machine-verify the new package's boundary-neutrality (reasoned safe, not yet machine-verified).
25. **Document the replace_all-after-insert gotcha beyond the ledger** — it's a cross-project agent-editing hazard; add to `references/lessons.md` in crush-config by commit (global lessons go there, never in-session writes).
26. **Budget enforcement question**: pipeline 1165938 outlived its `--max-time 15m` budget flag (started 01:38:58, still listed well past 01:53:58). Check whether `--budget`/`--max-time` enforce or merely log; a runaway fixer pipeline is the poisoner's best friend.
27. **Kill-switch runbook**: document the safe way to stop a live poisoning pipeline (which pid patterns, whether killing mid-step is safe for the tree) so the next session isn't forced to wait out a 15-minute window while the rewriter works.
28. **Sweep cadence bookkeeping**: AGENTS.md's self-scan cadence note says "once a month"; November got discharged 2026-11-23. Add the actual last-run date to the cadence bullet so the next session knows December is due without reading the ledger.
29. **Consider a `--json` diff mode for self-scans**: judging sweeps would be sharper if the scan emitted machine-comparable group sets (hash + members) so ledger tables could be generated, not hand-written — ROADMAP fuel, ties into existing plumbing formats.
30. **The G6 residual question**: art-dupl's own scan grouped only 2 of the 3 `VersionFooterSetter` sites (grep found the third). Worth a look at WHY the third site didn't group (context divergence? statement-embedding?) — if optional-setter assertions regularly escape grouping, the family undercounts and a future pattern (`optional-setter-assertion`) might deserve the corpus test.

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Who is the destroyer?** Two daemon auto-commits (`8ce90a47`, `57782482`) deleted `.buildflow.yml` without listing it in their "N changed file(s)" messages. From inside the repo I can see WHAT happened but not WHICH process decided to delete a tracked config file mid-flight. Do you have a daemon-side log, a cleanup heuristic config, or a cron that prunes files it considers stale? (This is the difference between fixing the symptom for the third time and ending the war.)
2. **Is the `go-auto-upgrade` skip permanent policy, or do you want the upstream exclusion-mechanism request filed** (owner Q1 from #17/#18, still open)? If permanent: should I roll the pipelineconfig gate + skip to the other fleet repos with ADR-0024-style v1 invariants in the same pass?
3. **The `vendorHash` FOD repair ran unattended tonight** (fake-hash swap, build, self-restore, committed). Should BuildFlow's repair journal/scratch artifacts be gitignored fleet-wide to keep them out of daemon sweeps, or do you consider the journal deliberate provenance worth tracking when it does land?

---

*Point-in-time snapshot — goes stale. Section (f) items 1 and 9–13 are docs-health HARVEST inputs; the rest are routing candidates. WAITING FOR INSTRUCTIONS.*
