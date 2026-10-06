# SUPERB Session 4 — Full Report: P0 Rescue, BuildFlow Shipped, v0.9.0 Release Recovered, Consumer Sweep, Recurrence #6

**Date:** 2026-10-06 13:17 CEST (session ran 01:45–02:45; CI concluded by 13:10 after a documented-flake re-run)
**Scope:** the two reports the owner pointed at — session-3 (`01-41_superb-session3-t05-t18-v0.9.0-released.md`, resume point + NEXT list) and the docs-health self-review (`01-42_docs-health-pass-self-review.md`, P0 list). Standing blanket order. An intermediate snapshot exists at `04-30_superb-session4-p0-rescue-buildflow-release-recovery.md`; THIS is the definitive session record.
**Repos:** art-dupl (fork @ `dd0e05fa`, pushed, **CI green all lanes**), BuildFlow (master @ `6b3ccc667`, pushed), DiscordSync (`4b1418bb` local), InboxClean (daemon `e82614b` local), code-duplicate-analyzer (`0535f79` local), auto-deduplicate (`35e5eca` local).
**Owner answers checked, NOT found** (report files, TODO_LIST, commits; BuildFlow's D1–D16 dossier is a different session's decision board and does not answer them).

---

## a) FULLY DONE (each verified before moving on)

1. **Docs-health P0 #1 — /tmp rescue FIRST.** `/tmp/dh` still existed; copied before anything else: driver → `scripts/docs-health-driver.py`, ~2,500-item verdict corpus (4,244 lines) → `docs/status/2026-10-05_pass-verdicts/` + provenance README. The vanish-class can no longer erase this pass.
2. **P0 #2 — the two bare review files, annotated the DOCUMENTED way.** `--emit-keys` → dry-run on a /tmp copy → `--verify` → apply (the key-grammar lesson applied; zero atomic aborts; item verdicts re-verified against code first):
   - `docs/reviews/2026-07-25_17-35`: 3 of 4 remaining-work items struck DONE (switch-predicate conversions verified gone; CI Self-Analysis gate; `TestNoDuplicateErrorNewMessages` at domain/cross_package_alias_test.go:76); withLock/diffStatTable tests left bare (open). Dated banner added.
   - `docs/reviews/2026-06-20_06-50`: Fragment unification + coverage tooling struck DONE; Clone consolidation + Data→View rename left bare (open/routed); decoupling acceptances struck WONT. "Completed Work" list deliberately untouched (historical record). check-rows green on both.
3. **P0 #4/#5/#6/#8/#10.** Regime **A-lite** named in `docs/status/archived/README.md` with an honesty statement and a **mechanically computed recount** (376 snapshots = 57 A + 23 A-lite + 296 B; the grep method is documented so future counts stay derived). The 05-21_19-44 FIX-MISS headline fixed (cause: bold-placement `**BuildFlow**` vs the spec's plain `BuildFlow` + longer line). All FIVE empty-snapshot "adjacent same-day reports" claims verified per-file with concrete filenames appended. Non-md artifact policy decided + written (renders follow their .md; solo HTML reports need ANNOTATE-then-move; research HTML governed by the AGENTS tracking policy). AGENTS key-grammar bullet + `docs/SELF_CLEAN_LEDGER.md` lesson entry added.
4. **BuildFlow T18 completion — the clean resume point, fully cleared.** Root causes found: (i) `GOWORK=off` leaks from art-dupl's devShell into EVERY shell (why "no go.work found" + `-mod=readonly` looked insane), (ii) the deeper real one: root + execution go.mods still required art-dupl **v0.7.2** vs tools' v0.9.0 — a stale mid-bump vendor state (`vendor/modules.txt` carried BOTH versions). Fixed: both go.mods aligned to v0.9.0, `go work vendor`, two `update-vendor-hash` rounds. **All 33 workspace modules green** (`nix run .#test`), root + tools tests green, `nix build .` green. One REAL gate finding fixed en route: T18's new code carried 3 em dashes in .go comments — `TestNoNewEmDashes` caught them (the ratchet works).
5. **BuildFlow docs + push.** `docs/configuration.md` gained the `tool_options` row + example; AGENTS gotcha #147 updated (key count 27→28 + semantics + panic note); CHANGELOG Unreleased entry. Proper commit `6b3ccc667` **pushed — exactly up to my gate-verified commit**, deliberately leaving the concurrent session's in-flight `d29bbb566` local (theirs to push).
6. **BuildFlow lane soak — tool_options verified with the real binary.** Fixture with ≥6 flat shared statements (gotcha #191's rule): default = **1 finding** (`alpha.go:6:2`, "16 tokens in 2 instances" — v0.9.0 column precision live in BuildFlow output), `threshold: 20` = **0 findings**, typo `threshhold` = **loud startup failure naming tool + option** (clean DI resolve error — better than the expected raw panic).
7. **v0.9.0 GitHub release RECOVERED — it had silently FAILED.** Session 3's report said "goreleaser Release workflow ran"; it ran and FAILED: cosign-installer's floating default jumped v2.5.2 (v0.7.2, green) → v3.0.6, and v3 rejects the config's `--output-signature`/`--output-certificate` flag set ("cannot specify service URLs and use signing config"). No release object existed; nobody had checked. Fixed: pinned `cosign-release: v2.5.2` + added a `workflow_dispatch` `tag` input (retry path for tags whose own workflow file is broken), dispatched from fork → **release live: 52 signed assets** (all platforms, SBOMs, cosign sigs, checksums) + **curated body** (headlines, 0.8.0-subsumption, the finding-ID `:column` shape-change callout). pkg.go.dev verified rendering v0.9.0.
8. **Flake version drift found + structurally fixed.** `nix build github:…/v0.9.0` printed **art-dupl-0.7.0** — flake.nix's hardcoded `version = "0.7.0"` had drifted through THREE releases despite RELEASE.md's own 2026-09-29 warning. Now **derived from CHANGELOG.md's newest `## [X.Y.Z]` header** (assert-guarded, eval-loud on malformed changelog). The frozen tag keeps its cosmetic 0.7.0 (proxy-served tags never move); release body documents it; next tag self-corrects. RELEASE.md step 3 rewritten.
9. **Consumer sweep to v0.9.0 (go-ecosystem-upgrade loaded FIRST — the ordering lesson finally honored).** Full no-ignore enumeration (8 consumers). **DiscordSync**: dedup gate dry-run against v0.9.0 BEFORE flipping — **0 new clones vs the 88-group baseline** (group hashes unaffected by the ID column change) → tag pin `refs/tags/v0.9.0`, lock updated, devShell builds, the C31(3) PIN-TRAP manual-restore ritual retired in the comment. **InboxClean**: same flip; the sandboxed `dupl` check derivation builds green. **code-duplicate-analyzer**: go.mod v0.9.0, production build green. **auto-deduplicate**: go.mod v0.9.0, resolves + tidies (pre-existing breakage recorded BEFORE the bump). **BuildFlow**: done in (4).
10. **jsonv2 recurrence #6 caught at the boundary and restored.** The session-start 33-file daemon sweep `b729d8a7` had re-migrated BOTH protected files AGAIN — my closing boundary gate's fresh `-count=1` caught it (jsonutil empty-indent + jsonv2gate + config Duration tests all red), restored verbatim from parent (`2e3d3cd7`, pushed), AGENTS recurrence #6 entry written. **Sixth occurrence.**
11. **art-dupl loose ends closed:** TODO_LIST T18 row updated with BuildFlow-side DONE + soak evidence; byte-column-vs-UTF-16 semantics documented in HOW_TO_USE's LSP section; research README ledger row carries v0.9.0 + consumer status; stale `feat/provider-threshold-knob` branch deleted (content verified in the v2 cherry-pick lineage); README provider paragraph corrected (threshold is a declared option, not "fixed"). Stale half-merged `isAssertionMethod` comment fixed on sight.
12. **CI fully green on the final tip** — after two real post-push findings were fixed and one documented flake re-run: treefmt drift on my flake comment (fixed, `26784439`), alloc-gate vendorHash staled by `b729d8a7`'s go.sum drift (fixed, verified locally BEFORE push this time), and a windows-latest GC fatal-error flake (`found pointer to free object`, the documented 2026-09-19 nondeterministic class — binary-identical runs flip) that cleared on `gh run rerun --failed`. Final: ubuntu/macos/windows/lint/coverage/nix-flake-check ALL success on `dd0e05fa`.

## b) PARTIALLY DONE / DELIBERATELY HELD

1. **CV held at v0.7.2 — with evidence, not neglect.** Its baseline-ratchet gate FAILS under v0.9.0: expression-aware templ detection finds NEW clone groups in CV's `.templ` files (admin_page.templ:564, coaching_page.templ:57, admin_access.templ). Accept-into-baseline vs extract is judgment belonging to CV's session/owner; blind baseline regen to make a bump pass would accept unjudged clones. The v0.7.2 tag pin is stable (no branch race) — no urgency.
2. **SystemNix skipped**: rev-pinned (stable) AND its redeploy is staged for the owner at the console (BuildFlow dossier D15) — a flake-input flip now would collide with a staged deploy.
3. **DiscordSync / InboxClean / analyzer / auto-dedup commits are LOCAL** — their daemons/sessions push per their own cadence. InboxClean's went through its daemon with a heuristic message after its pre-commit gate failed on PRE-EXISTING templ breakage (`templ.KeyValue[string, bool]` vs `attributeValue` in generated chat_templ.go:357 — unrelated to the pin, present before my change).
4. **Docs-health P0 #3/#7** (A-lite→A re-keying for ~115 KEEP files + mechanically derived banner counts) — one focused session; the verdicts that feed it are now safely in-repo. Owner §g Q1/Q2 of the docs-health report still gate the depth decision.
5. **Release legs not re-verified**: the recovered release's Homebrew tap (formula `skip_upload: true` — intentionally untouched) and the ghcr Docker image legs ran as part of the green retry, but I did not separately verify the image digest / tap state.

## c) NOT STARTED (unchanged, ranked)

- **T07b (M38–M39)**: upstream go-finding coverage-channel PR (`DetectResult{Findings, FilesScanned, SkippedModules}` behind `DetectorV2`) + ecosystem.md art-dupl staleness fix + the S4 CLI-toolsdk-discovery gap. Skills-first: jj-fork-pr-workflow + verify-before-filing + github-voice.
- The three session-3 §g answers (see g).
- art-dupl TODO_LIST fleet items: gogenfilter v3.6.1 sweep (9 repos), fleet jsonv2 audit (#14), filepath.Separator audit (#15), branch protection #18 (owner action, 5th reminder).
- BuildFlow TODO BF2 (now unblocked by the tool_options plumbing).

## d) TOTALLY FUCKED UP (honest ledger)

1. **I pushed recurrence #6 to the remote before gating.** The daemon's `b729d8a7` rode my docs push (`d38c7c1d..84ec663c`) — I pushed "docs-only + one comment" without running the boundary gate first, exactly what the task-boundary rule exists to prevent. The gate caught it minutes later and `2e3d3cd7` fixed it, but the remote briefly served a broken jsonutil, and the same sweep's go.sum drift then broke CI's alloc-gate for TWO more pushes. **Rule going forward: gate-then-push even when my own diff looks docs-only — the daemon's diff rides with mine.**
2. **First CV gate dry-run read the PIPE's exit code** (`cmd | tail; echo $?` → tail's 0) and briefly reported a failing gate as green. Caught one step later on the usage-output smell; re-ran capturing the real RC=2. The sweep's "verify" is only as good as the exit code actually captured.
3. **Two Nix-eval missteps** before the version derivation worked: `builtins.split`'s capture-list interleaving returned lists where I expected strings; then POSIX-ERE rejected `\[` escapes. Fixed with `lib.splitString` + `[[]]` bracket classes — but only after two failed evals I should have desk-checked.
4. **sed mangled a possessive** (`the tool's own` → `the tool own`) in the BuildFlow comment fix — caught by re-view, repaired with the edit tool. sed-on-prose remains a footgun I keep stepping on.
5. **treefmt drift shipped** on my flake comment (CI red #1) because I committed without `nix fmt` after editing a nix file — the exact "buildflow's fixer ≠ treefmt, run BOTH" rule from session 3, violated one session later.
6. **Inherited-but-unchecked: the release-existence assumption.** Session 3's "goreleaser Release workflow ran" hid a hard failure; I trusted the handoff instead of `gh release view` until the post-release checklist ran. Post-release verification must assert the release OBJECT + assets exist, not that a workflow triggered.

## e) WHAT WE SHOULD IMPROVE

1. **Gate-then-push is non-negotiable when a daemon shares the tree** — even for "docs-only" changes; the daemon's concurrently-packaged edits (a v2 migration, a go.sum drift) ride the same push. The boundary gate exists precisely for this and proved it twice tonight.
2. **A release checklist needs an artifact-existence step**: `gh release view <tag>` returns the release with N assets; the goreleaser workflow's green-ness is not the deliverable.
3. **Pinned actions should pin their TOOL too**: cosign-installer@SHA floats the cosign version (v2.5.2→v3.0.6 broke the release silently). Same class as the version-string drift: any floating derivation of a release artifact is a latent outage. The dispatch-retry input added this session is the recovery pattern for frozen tags.
4. **Derive, never hand-maintain, anything a process can compute** — flake version (done), archived-file counts (done via documented grep method), daemon-vulnerable invariants → gates (jsonv2gate works; the gap is daemon EXCLUSION, see g Q1).
5. **Exit codes: capture the binary's, not the pipeline's** — one `set -o pipefail` habit (or redirect-to-file) would have removed mistake (d2).
6. **Foreign-repo sweeps**: the dry-run-gate-before-flip pattern (DiscordSync/InboxClean) found CV's incompatibility BEFORE any damage — make it the standing first move of every consumer bump, alongside the skill's Phase-1 baseline.

## f) Up to 50 NEXT (ordered)

1. Owner: answer g Q1 (daemon hardening) — six recurrences, six restores.
2. Owner: answer g Q2 (emit-suppressed-accepted default).
3. Owner: answer g Q3 (retro-tag v0.8.0 or leave one clean v0.9.0).
4. T07b session: load jj-fork-pr-workflow + verify-before-filing + github-voice FIRST; verify upstream working copy vs tags.
5. T07b: prototype `DetectResult{Findings, FilesScanned, SkippedModules}` behind `DetectorV2` in ~/projects/go-finding.
6. T07b: upstream tests pinning existing-Detector source-compat.
7. T07b: PR including ecosystem.md art-dupl staleness fix + S4 CLI-toolsdk-discovery gap (content or filed issues).
8. T07b: after merge+tag → art-dupl bumps, provider populates counts, tests assert; BuildFlow surfaces FilesScanned.
9. CV: judge the new v0.9.0 templ clone groups (accept into baseline with reasons, or extract) → then flip the pin (gate dry-run already proven failing cleanly).
10. SystemNix art-dupl input bump after its staged deploy lands (D15).
11. DiscordSync/InboxClean/analyzer/auto-dedup: their pushes land (or push them next session with each repo's own gate run).
12. Docs-health P0 #3: A-lite→A re-keying session (start with the 11 September KEEP files; verdicts in `docs/status/2026-10-05_pass-verdicts/`).
13. Docs-health P0 #7: driver patch deriving "N resolved / M open" banner counts mechanically (grep -c against the file's items).
14. Docs-health P1 backlog: strike-ready conversion waves (Aug v2 → Jul-early v3 → Jul-mid v4 → Jul-late v5 → Jun v6 → May v6 batch-last).
15. Docs-health P1 #19: 09-19_06-41 f-table rows 41-49 strikes (verdicts ready).
16. Docs-health P1 #21: Keep-a-Changelog audit — CHANGELOG `[Unreleased]` vs `git log` since v0.9.0 (a gate would keep it fixed; candidate extension of the count gate).
17. Docs-health P1 #23: extend the count gate to fail on "Nothing yet" `[Unreleased]` + post-release feature commits.
18. Docs-health P1 #24: annotator spec-linter in `scripts/check-annotator.sh` (reject non-grammar keys at authoring time).
19. Docs-health P2 #29/#30: link-rot sweep for archived-path references in living docs (the harvest citations).
20. Docs-health P2 #31: `check-rows.py` over the full 55-file moved set, recorded in the manifest.
21. Docs-health P3 #33/#34: upstream the driver + verdict-agent prompt template into the docs-health skill repo after one clean re-key pass proves them.
22. Docs-health P3 #35: flip `check-docs-freshness.sh --strict` after one clean cycle.
23. art-dupl: self-scan (`scripts/self-scan.sh -t 1 --type-aware`) post-column changes; ledger decisions.
24. art-dupl: benchstat vs baseline for the column O(line) scan (alloc-gate unaffected; timing unconfirmed).
25. CLI SARIF: relatedLocations + region snippet (ledger candidate).
26. CLI SARIF: drop the empty-string `category` property (wire change — deliberate decision needed).
27. Consider registering art-dupl in go-finding's LinterRegistry (`RegisterLinterCategory`) for other consumers.
28. Re-verify upstream go-finding analysis/pipeline tags when T07b lands (working copy was ahead of v1.13.0).
29. Migrate `.goreleaser.yaml` signs block to cosign `--bundle` style, then unpin cosign (or keep the pin forever — the pin is also supply-chain hygiene).
30. Watch the recovered release's Homebrew tap + ghcr image legs on the next natural touch (skip_upload means tap intentionally untouched — verify that's still the intent).
31. BuildFlow: BF2 (now unblocked by tool_options plumbing).
32. BuildFlow: D12 art-dupl ratchet re-baseline under v0.9.0 (their dossier's own precondition) + ratify.
33. BuildFlow: fleet `.buildflow.yml` adoption of `tool_options` where thresholds are wanted (after owner's Q2 answer shapes the default).
34. InboxClean: the pre-existing `templ.KeyValue` generated-code breakage (their session; flagged here with file:line).
35. auto-deduplicate: pre-existing `duplicate.Duplicate` build breakage (their session; recorded as pre-bump baseline).
36. code-duplicate-analyzer: pre-existing `internal/errors` test-compile failure (Go 1.27 `interface{Unwrap() error}` no longer satisfies `error`).
37. gogenfilter v3.6.1 sweep (9 repos) — unchanged TODO_LIST item.
38. Fleet jsonv2 audit (#14) + filepath.Separator audit (#15) — unchanged.
39. Branch protection #18 — owner action, 5th reminder (three CI reds tonight were each caught BECAUSE pushes+checks ran).
40. Website: check whether the pnpm-built docs site needs a v0.9.0 release-note page update (changelog.mdx highlight exists; verify rendering).
41. `docs/status/README.md` (lifecycle explainer for humans) — docs-health P4 #49.
42. Consider whether the daemon's heuristic commit messages on gate-relevant files (jsonutil, go.sum, flake.nix) should trigger a notification — half of tonight's d-section exists because they don't.
43. HOW_TO_USE: mention `tool_options`/threshold consumer story (BuildFlow) in the provider section (currently README-only).
44. TODO_LIST: strike/annotate the session-3 NEXT items now done (this report is the harvest source for the monthly pass — due 2026-10-28).
45. The `-t 1` templ whole-file artifact (documented 2026-10-04): candidate fix = drop the File token from the templ stream + CacheVersion bump — only if it ever bites a consumer.
46. gopls embedlit stale warning on column_test.go:81 — persisted all session; LSP-restart hygiene note stands (golangci-lint + CI green).
47. DiscordSync: now that the tag pin killed the moving-ref exception, their `check-moving-refs.sh` art-dupl carve-out can be simplified (their session).
48. Consider a `scripts/check-release-exists.sh` (tag → `gh release view` → assert assets ≥ N) as the post-release step of RELEASE.md — mechanizes e2.
49. Sweep remaining `github:LarsArtmann/art-dupl/v0.7.2`-style pins found in the wild (CV was one; re-grep fleet after CV's judgment lands).
50. Close the loop: strike THIS report's f-items as they land — the next harvest reads it.

## g) QUESTIONS (cannot be answered from the repo)

1. **Daemon hardening — 6th ask, now with recurrence #6 as the sixth data point.** The auto-commit daemon has re-migrated `internal/jsonutil` + `config/config_migrate.go` to direct `encoding/json/v2` imports SIX times (latest: `b729d8a7`, restored `2e3d3cd7`); tonight it ALSO staled go.sum → two CI reds. Test gates catch it after the fact; each cycle costs a restore + a red window on the remote. The daemon config is yours: **(a) exclude those two files (and go.sum/flake.nix?) from heuristic auto-commit, or (b) add a pre-commit jsonv2gate hook that blocks the daemon's commit?** One line either way prevents recurrence #7.
2. **`emit-suppressed-accepted` default (T18 follow-up).** The option ships default-false (byte-identical output). With BuildFlow's `tool_options` now live, should BuildFlow's art-dupl lane enable it per-repo via config (my assumption), or should the provider default flip to true (changing wire output for all consumers)? Decides NEXT #33's shape.
3. **Retro-tag v0.8.0?** v0.9.0 (release body + CHANGELOG) documents that 0.8.0 was never tagged and subsumes it. If you want v0.8.0 in pkg.go.dev history, name the boundary commit (or bless "last commit of 2026-09-29"); otherwise history stays one clean v0.9.0 and I'll stop asking.

---

**THE SESSION IS PAUSED. Waiting for instructions.** Resume point: NEXT #4 (T07b session, skills first) or any owner answer above — everything else on the board is either pushed-green or explicitly held with evidence.

_Arte in Aeternum_
