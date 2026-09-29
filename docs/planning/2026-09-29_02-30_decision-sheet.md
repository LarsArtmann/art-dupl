# Decision Sheet — five answers that unblock ~15 gated items

**Date:** 2026-09-29 02:30 CEST
**Prepared by:** SUPERB v2 pass, M01. Each decision: options, effort math, recommendation. Reply with the five letters (e.g. `g1b g2a g3a binfmt-go v080-go`) or override any line.

---

## g1 — Legacy archive retrofit: fund or close?

**The ask** (audit f27): convert the 272 pre-regime archived files (pre-2026-08-15) to inline-annotated regime-A form, verifying per item.

**Options**
- **(a) Fund it.** ~2–3 focused sessions (≈6–9h) of the same batch workflow that struck ~2,000 items in the last pass. Verdicts are mostly derivable from CHANGELOG eras, but every item needs a check — the June batch (30 files) alone was a full M-task.
- **(b) Close it permanently.** The honesty gap is already covered: `docs/status/archived/README.md` documents the two regimes and the retrofit policy (line 28: consolidation commit `32f12cad` deliberately did NOT retrofit). Nothing false is claimed — the files are simply unannotated.

**Effort math:** ~272 files ÷ ~90 files/session (observed June-batch velocity) ≈ 3 sessions ≈ 9h. Reader value: near-zero — all open items those reports carried are already routed to TODO_LIST/ROADMAP; the pass verified this during harvest.

**Recommendation: (b) close.** Spend the 9h on v0.8.1+ product work instead. Revisit trigger: if a historical incident ever requires proving when a decision was made, annotate that one file on demand.

**Unblocks if answered:** v2 M17 retrofit waves (the only gated docs item).

## g2 — GitHub-Action distribution: confirm won't-implement?

**The ask:** the 0/20 GH-Action plan carries a unilateral WON'T-IMPLEMENT verdict (2026-09-28): superseded by the BuildFlow toolsdk provider lane (LIVE since 2026-09-25, core Go+templ detector).

**Options**
- **(a) Confirm won't-implement.** The plan file keeps its verdict header; the topic dies. Revisit trigger: a real consumer asks for marketplace distribution that BuildFlow cannot serve.
- **(b) Keep it as a ROADMAP maybe.** Costs a TODO_LIST/ROADMAP row and re-litigating every pass; the provider lane already covers the only known consumer (BuildFlow).

**Recommendation: (a) confirm.** The provider lane IS the distribution story; a marketplace action would be a second integration surface with zero known consumers.

**Unblocks if answered:** permanent closure of `docs/planning/2026-07-28_10-30-github-actions-distribution-plan.md` (currently decision-documented, awaiting your confirm).

## g3 — BuildFlow Build Gate red on master: whose fix?

**The ask:** BuildFlow's `Build Gate` workflow on master is RED (fails in ~3s, config-shaped — not a code failure). art-dupl's provider lane lives in that repo.

**Options**
- **(a) Yours (owner action).** You open the failing run and fix the config (likely a `.buildflow.yml` or workflow-input change only you can validate).
- **(b) Delegate to a session with BuildFlow checkout.** Any session can read the 3s log and propose the patch as a PR; you still merge.

**Recommendation: (b) delegate the diagnosis, you merge.** A 3s config failure is cheap to diagnose; the merged fix is the only part that needs owner judgment.

**Unblocks if answered:** BuildFlow-side D1 (toolsdk Spec.Timeout mapping), D2 (warnings_budget `art-dupl` key), D3 (lane-overlap query) — all parked on BuildFlow-side work.

## binfmt — permanent SystemNix fix + vendorHash cycle

**The ask:** local `nix build`/`nix flake check` on this host still depends on the `/run/binfmt` stopgap (re-applied manually after the 2026-09-28 gap); the pending gogenfilter vendorHash cycle needs one real `nix build` to land the hash.

**Options**
- **(a) Do it now (needs root once).** Add the binfmt mount to the SystemNix/NixOS configuration so reboots keep it; then run `nix build` once to settle the vendorHash (≈10 min build + hash copy).
- **(b) Leave as CI-only.** Heavier gates already run in CI (`nix flake check` green in CI lanes); local nix stays best-effort.

**Recommendation: (a).** The stopgap has already rotted once (the 09-28 gap); making it permanent is a one-line config change plus one build, and it re-enables the flake's `-race` + alloc-gate checks locally.

**Unblocks if answered:** local `nix build`/`nix flake check` for every future pass (currently: verify in CI only).

## v0.8.0 — go / no-go for the release

**The facts:** the B1–B6 hardening + this week's additions are finished, curated under `## [0.8.0] - 2026-09-29` in CHANGELOG.md (link defs already updated), with ~15 user-facing entries incl. two real bug fixes (jsonv2 Duration crash class, Go 1.27 selector-key false positives). RELEASE.md exists; the remaining prep (rehearsal, tag draft) is this pass's M05 and stops one command short of execution.

**Options**
- **(a) GO after close-out.** When v2 M22 is green (full `-race` + CI on pushed HEAD), run the prepared tag + `git push --follow-tags`. The release reaches every BuildFlow fleet consumer.
- **(b) Hold.** Everything stays on `fork`, unreleased; the [0.8.0] section keeps aging.

**Recommendation: (a) GO, gated on M22 green.** The section is complete; holding it means the fleet keeps pre-hardening providers.

**Unblocks if answered:** M05 execution (tag + push), and the post-v0.8.0 chore list (reviews-render retirement, renders refresh).

---

## Gated-item map — what unlocks per answer

| Your answer | Items unlocked |
| --- | --- |
| g1 = (a) fund | M17 legacy retrofit waves (~3 sessions; scheduled after v2 pass) |
| g1 = (b) close | M17 collapses to a one-line ledger note; topic permanently closed |
| g2 = (a) confirm | GH-Action plan closed permanently; no revisit in future passes |
| g2 = (b) keep | Plan routed to ROADMAP; one ROADMAP row added |
| g3 = (a) yours | D1/D2/D3 wait for your BuildFlow session |
| g3 = (b) delegate | Diagnosis memo prepared next pass; you merge the fix |
| binfmt = (a) | Local `nix build` + vendorHash cycle; alloc-gate + race checks run locally |
| v0.8.0 = (a) | M05 execution: tag + push on M22 green |
| v0.8.0 = (b) | [0.8.0] stays parked; M05 prep archived as the standing release recipe |
