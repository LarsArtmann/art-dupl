# Status Report: docs-gates Clearance Pass 1 (freshness + routing GREEN, annotate 689→289, fleet orchestration lessons)

> **Generated:** 2026-11-14 12:50 CET
> **Branch:** `fork`
> **Session:** one docs-health AUDIT/ANNOTATE/ARCHIVE pass over the whole status corpus, triggered by `docs-gates run` showing all five gates red (freshness/audit-all/annotate/routing/cadence — the overdue pass the 2026-11-14 10:13 report flagged as §f item 21/22).

---

## a) FULLY DONE

1. **Rules decoded before touching anything.** Read the docs-health SKILL.md in full, then read the docs-gates Go source (`~/projects/docs-gates/internal/docsannotate/*`, `docsfreshness/reports.go`) to extract the EXACT gate contracts: marker words (`done|resolved|superseded|fixed|adopted|decided|verified|won't implement|not-do|duplicate|struck|moot|no longer`), B/hash (`done at <sha>` must resolve via `git cat-file -e`), C/path (backticked verdict paths with doc extensions must exist on disk), F/spacing (`[A-Za-z0-9]` may never touch a backtick), G/uniform (per-cell tilde parity in table rows), and the freshness parser's resolution rule (**an item is resolved only when its text contains `~~`** — verdict words alone do not close an item).
2. **Wave-1 mechanical repairs (annotate findings 689 → 289's precursor):** F/spacing ×21 fixed across AGENTS/FEATURES/README/SDK_DESIGN/ADR-0022/0025/0028/planning/research/status (the whole class was a plural s glued directly after a code span's closing backtick, plus three glued-space lines); B/hash fixed (archived/README.md's doc-example hash replaced with the real, resolving `91419533`); C/path ×3 fixed (two citations re-pointed at files that moved into `archived/`, one NOT-DO-absorbed cache-ADR item, one resolved-with-archived-pointer).
3. **Routing gate GREEN:** the two unrouted coverage rows in `2026-06-20_15-11_comprehensive-status.md` annotated (`superseded — 2026-06 snapshot; coverage is tracked by the CI Coverage job now`). Gate now reports `0 unrouted open row(s)`.
4. **Freshness gate GREEN:** five candidate living-doc lines were pre-flighted (`--pre-flight` — first draft was UNSAFE: a double "health" token and a "store" token would have genericized two historical items into DRIFT; reworded until SAFE) and landed byte-exact: four `CHANGELOG.md` Fixed entries (docs-gates clearance, ruff findings, shellcheck SC2207, go.mod dist ignore) + one `TODO_LIST.md` open row (govulncheck/vulnix rides the next nixpkgs/Go bump). The newest report's §f items 22/25/26/28/32 are now harvested; `docs-gates freshness .` reports **FRESH, 0 unharvested**.
5. **Three real tooling fixes landed with the CHANGELOG entries to prove them** (newest report §f items 25/26/28): all six ruff findings in `scripts/` fixed (PLW1510 explicit `check=False` ×2, F841 unused vars ×2, SIM102 flattened, EXE001 `chmod +x remove_testify.py`) — `ruff check scripts/` clean; shellcheck SC2207 in `scripts/bench-realworld.sh` fixed via `mapfile -t`; `go.mod` gained an `ignore (dist)` block. `go build ./...` green after the go.mod edit.
6. **Six heavy files taken solo and driven to `check-rows --strict` + `bare-rows` CLEAN:**
   - `docs/status/2026-06-20_15-11_comprehensive-status.md` — 45 table rows + 2 prose items verdict-annotated, duplicate 2026-10-05 banner deduped, three measurement tables ledger-exempted, coverage table uniformly struck.
   - `docs/status/archived/2026-07-16_03-38_templ-call-expression-fp-fix.md` — 35 rows (mangled triple-duplicated verdict text hand-reconstructed to single honest verdicts, per-cell).
   - `docs/status/archived/2026-01-07_14-45_comprehensive-codebase-status-update.md` ("the monster") — 53 numbered items + all instances of ambiguous duplicates + 35 measurement rows; five `routed` verdicts converted to struck `superseded` after gate testing showed routed items stay open.
   - `docs/status/2026-09-19_06-41_v0.7.0-release-go1.27-coherence-ci-recovery.md` — 47 table rows with per-row evidence (ADR-0024 path corrected to its real filename, `.gitattributes`/`go-env-doctor.sh`/`pre-release-check.sh` verified to exist before citing).
   - `docs/status/2026-06-20_23-08_statement-level-tokenization-complete.md` + `docs/status/2026-06-23_08-34_multi-skill-sprint-status.md` — verdict tables done, accomplishment/metrics tables ledger-exempted, and five verdicts that a stale line-number spec had crossed onto wrong rows were caught and re-fixed content-anchored.
   - `docs/status/archived/2026-06-22_19-44_split-brain-resolution-complete.md` — SB-1..8 issues, §c/§d/§f/§g all verdict-annotated (including correcting a wrong NOT-DO on the CloneRef row — it landed).
7. **Nine solo HTML reports archived:** six via `git mv`, three via plain `mv` (untracked files — no history to preserve). Each was verified before moving (open-marker scan + obligations traced to landed work or TODO_LIST/ROADMAP routing, per the archived/README non-md-artifacts protocol). `docs/status/archived/MANIFEST-2026-11-14.md` created with a per-file deciding reason.
8. **Batch-07 (22 files) completed by subagent cleo-annot-07** — every file `check-rows --strict` + `bare-rows` clean, shas verified via `git cat-file -e`, paths verified via `test -f`, including a mid-run table corruption it caught and rebuilt from the git original. ~900 verdicts landed across the two of us this session (1,582 → 4,137 struck items).
9. **CHANGELOG F/spacing fixed** (the last of the 22 F findings), landed after the pre-flight so the additions-only model stayed intact.

## b) PARTIALLY DONE

1. **Annotate gate: 689 → 289 findings** (58% cleared). The remaining findings concentrate in files owned by the not-yet-resumed batches (00/01/02/04/05/06/08/09/10) plus a long tail of G/uniform rows in archived files.
2. **Audit-all: 694 → 495 UNHARVESTED items.** The oldest era (2026-01/02/04/05 reports) is mostly untouched; my solo files and cleo's batch are fully cleared.
3. **Cadence: archive phase not started for .md reports.** The 9 HTML files are archived, but the gate still counts **168 unarchived reports (threshold 15), oldest 197 days (threshold 21)** — the git-mv wave with `> **Live remainder:**` banners + regime-A verification + README recount is entirely pending.
4. **Living-docs polish:** stamps not yet bumped (AGENTS/FEATURES/README/HOW_TO_USE sit at 2026-11-06, ROADMAP at 2026-10-11); the ROADMAP `config.Config god object` row is drafted but NOT yet pre-flighted/landed (only the README-stamp strike was pre-flighted SAFE — landing it is safe but not yet done); cleo's 11 HARVEST-CANDIDATEs are not yet routed into TODO_LIST/ROADMAP.
5. **AGENTS.md memory updates pending:** the routed-kind trap, the annotate.go repair/spec workflow, the provider-capacity lesson, and the docs-health cadence line ("next due 2026-10-28") all need writing; AGENTS.md stamp bump rides that edit.
6. **Fleet state:** cleo-07 delivered its full batch; geraldo-03 (batch-03, 82+ calls in) and cleo-07 (re-resumed) were re-dispatched after rate-limit kills but had not reported again by report time; cindy-01 died at 62 calls on synthetic quota; batches 00/02/04/05/06/08/09/10 have no live lane.

## c) NOT STARTED

1. Archive ritual proper: per-file `check-rows --strict` + `bare-rows` → `> **Live remainder:**` banner → `git mv` for ~150 dated reports → MANIFEST extension → archived/README recount (computed, never asserted).
2. Undated legacy files in `docs/status/` (29 files; 10 carry untouched tables — regime-B class, gate-neutral): curation pass (true proposals belong in top-level `docs/archive/` per the two-dirs convention) — deliberately deferred as gate-neutral.
3. The second pre-flight round (ROADMAP god-object row + README-stamp strike + harvest candidates from cleo's report: `--explain` structured JSON object, transform.go table-driven split, run_analysis split, `--min-files`, interface-impl threshold, v1.0 release notes, threshold recommendation, nix govet check, crush-daily doc pattern).
4. AGENTS.md updates (§e/§b above) and the final battery re-run + `go test ./cmd/ -run Docs` count gates + boundary script.
5. Fleet wave 2: resuming the eight dead batches one-at-a-time as zai slots free.

## d) TOTALLY FUCKED UP

1. **Fleet fan-out without a capacity probe.** I dispatched 11 subagents simultaneously; 7 died instantly on the zai rate limit. Then I burned another round discovering that gemini, huggingface and hyper all have NO credits (payment-required) and synthetic is subscription-capped — so 8 subagent restarts were doomed before they started. A single-probe dispatch would have revealed the whole constraint in one step.
2. **Findings-extraction bug:** my per-file UNHARVESTED manifests appended a spurious `.txt` to every path (I derived paths from finding-file names). Every fleet prompt carried broken file references; the subagents silently coped (stripping the suffix), but I found it only when re-ranking batches myself. Generating paths without asserting `test -f` at generation time was sloppy.
3. **Line-numbered specs on files I was mutating mid-pass.** Ledger-marker insertions shifted line numbers; my subsequent repair specs then crossed verdicts onto wrong rows (e.g. `MethodDetector.Name()` shipped → landed on the hash-detector row; a wrong NOT-DO briefly claimed CloneRef never landed). All caught by re-reading and fixed content-anchored, but the correct discipline from the start is: content-anchored keys for anything touched in the same session, line numbers only for frozen files.
4. **The routed-kind trap.** I used `r` (routed, no strike) verdicts before testing what closes a freshness item; the parser resolves on `~~` presence, so routed items stayed open. Caught by gate testing on the monster file, fixed there, and broadcast to all live lanes — but a 2-minute gate test BEFORE writing fleet instructions would have avoided three corrective broadcasts.
5. **Sequencing churn on the pre-flight model.** I landed the five freshness lines before the fleet finished, so harvest candidates arriving later require a second pre-flight round. Process-legal, but the single-batch plan I stated earlier dissolved into rounds; the discipline that matters (never write living docs before their pre-flight) held.
6. **Cost:** cleo-07 alone burned $7.31 across 222 tool calls (it was thorough and its output is excellent — but per-file cost control was never set; a per-batch token/timeout budget would have bounded the tail risk).

## e) WHAT WE SHOULD IMPROVE

1. **Probe-then-fan-out:** one subagent dispatch to validate provider capacity and prompt quality, then scale. The probe should include ONE file end-to-end so the prompt's assumptions (paths, tool flags, gate rules) are validated by execution, not review.
2. **Content-anchored specs everywhere** (`N@substring` keys / substring-matched specs); reserve line numbers for files nothing else touches during the pass.
3. **Gate-truth before fleet instructions:** run a 3-line scratch test per verdict KIND against the actual binary and paste the observed shapes into the fleet prompt (my prompts asserted the tool's documented behavior, not the gate's parsing behavior — those diverged on `routed`).
4. **Manifest generation asserts existence:** every path/hashes citation the tooling emits should be `test -f` / `git cat-file -e`-verified at generation time, not just at use time.
5. **Ledger-marker insertions belong at the END of a file's processing** (they shift every later line number).
6. **Fleet reporting contract:** require per-file gate exit codes in the report (cleo did this spontaneously and it made review trivial — make it mandatory).
7. **The `.md.txt` class of bug is caught by one `ls` assertion** — add a smoke check to any generated work manifest.

## f) UP TO 50 THINGS TO GET DONE NEXT

**Gate closure (this pass, in order):**

1. Finish the annotate gate: 289 findings → 0 (fleet lanes or solo; the G/uniform long tail in archived 2026-03/04 files is the biggest block).
2. Finish audit-all: 495 UNHARVESTED → 0 (same files, same pass).
3. Round-2 pre-flight + land: ROADMAP `config.Config god object` row; TODO README-stamp strike (already pre-flighted SAFE); cleo's 11 harvest candidates (curate — see §g Q2).
4. Archive ritual wave 1: the ~17 files older than 2026-10-24 that are already fully annotated (my six + cleo's 22 overlap archived/) — check-rows --strict + bare-rows → banner → `git mv`.
5. Archive ritual wave 2: the remaining dated reports as their batches land; recount `archived/README.md` (computed snippet) + extend MANIFEST-2026-11-14.md.
6. Cadence GREEN: ≤15 unarchived, oldest ≤21 days.
7. Undated legacy curation (10 files with untouched tables → `docs/archive/` or annotate; 19 clean ones move freely).
8. Full battery re-run until the matrix is green; `--json` face for the record.
9. `go test ./cmd/ -run Docs` count gates + `scripts/check-boundary.sh` fast profile (docs-only change, but the canary discipline applies).
10. AGENTS.md: stamp bump, cadence line ("executed 2026-11-14; next due ~2026-12-14"), new lessons (routed-kind trap, content-anchored specs, probe-then-fan-out, provider-capacity map).
11. FEATURES/README/HOW_TO_USE stamp bumps (claims already count-gated).
12. ROADMAP stamp bump + the god-object row.
13. TODO_LIST: land the curated harvest rows; bump stamp.
14. Session hand-off: harvest THIS report's §f into TODO_LIST/ROADMAP (§f item 8 of the standing convention).

**Leftover tooling findings from the pasted runs (carried from the 10:13 report, still open):**

15. `nix vendorHash` refresh (fakeHash → `got:` → rebuild) — flake.lock swept 2026-11-14, BuildFlow says 13/13 red.
16. `nix flake check`/`nix build .#art-dupl` before the next push (recurrence #10 rule).
17. `buildflow --fix` re-run to confirm the restored tree is green end-to-end.
18. `ErrNotARegularFile` wrapped message says "is a directory" unconditionally — fifo/socket would get a lying message (+ fifo test).
19. SDK zero-files-parsed-but-inputs>0 false-clean sibling (§e 5 of the 10:13 report).
20. Determine the `5412d55d` hook-bypass mechanism (§g Q1 of the 10:13 report; owner decision).
21. G5 daemon-gate deployment (SystemNix hooksPath flip) — owner answer pending since 2026-10-05.
22. Periodic jsonv2gate probe (daemon-tick/cron) for outside-the-hook red states.
23. Extract the jsonv2 recurrence ledger (#1–#15) from AGENTS.md into `docs/` + pointer stub (AGENTS.md 363/220 lines).
24. v0.10.0 release (markup-identity + check-json; CacheVersion 9; release note must say "re-run `art-dupl baseline`").
25. `--diff-report --json` identity guard (strict mode or ADR-0029 contract statement).
26. Owner decisions: threshold + `--ignore-tests` in analyzer policy; `--baseline-version-must-match` default-on; emit-suppressed-accepted default flip.
27. `go-auto-upgrade` 81 "file left on v1" warnings — decline/silence disposition.
28. cue-vet violations (`ci.yml` runs-on vs fleet policy; `release.yml` missing `concurrency`).
29. Lychee dead links (MIGRATION_GUIDE `file:///` refs, ADR-0017, +11).
30. Website pnpm-audit highs (js-yaml, undici, fast-uri, devalue).
31. arch-lint.yml go-version pin vs flake nixpkgs drift.
32. `buildflow format` re-verify under the drifted toolchain identity.
33. HOW_TO_USE.md stale SDK input docs check (same class as the README/SDK_DESIGN fix).
34. deploy-site.yml `html-validate`/`pnpm audit` `continue-on-error` policy decision.
35. A-lite → regime-A re-keying of the 2026-10-05 archive (tracked follow-up).

**Product backlog (re-confirmed relevant, unchanged):**

36. G9 windows lane GOEXPERIMENT split retirement; 37. windows exe-start confirm (#4); 38. G10 provider `Spec.Trigger.Requires` doc; 39. G8 go-finding pipeline release; 40. G11 toolsdk per-value validation; 41. T07b coverage-evidence upstream; 42. gogenfilter v3.6.1 consumer sweep (9 repos); 43. `filepath.Separator` fleet sweep (#15); 44. branch protection + notifications (#18, owner); 45. go-paperless v0.4.3 tag (#20, memo exists); 46. D3 lane-overlap post-soak; 47. core-lane fleet rollout; 48. monthly self-scan (check due date vs 2026-09-22); 49. `--ignore-markup-text` opt-out decision; 50. benchstat sweep after the next detection-touching change.

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Provider budget for the remaining corpus:** every external lane is dead (gemini/huggingface/hyper = no credits; synthetic = subscription-capped; zai = ~2-3 concurrent before 429s). ~289 annotate findings + ~495 UNHARVESTED items remain across ~150 files. Do you want me to grind it serially with one zai lane (hours, zero cost outside this session), or top up one provider so 4-6 lanes can run and finish in ~1-2 hours?
2. **Harvest curation for cleo's 11 candidates:** land ALL of them as TODO/ROADMAP rows (`--explain` structured JSON object, `--min-files`, interface-impl threshold 3→2, transform.go table-driven split, run_analysis split, v1.0 release notes, threshold-recommendation feature, nix govet check, crush-daily doc pattern, + the two already-tracked owner-gated rows), or curate to the ~6 that are real product ideas and NOT-DO the rest? My instinct is curate — TODO_LIST is for bounded work, not a museum.
3. **Archive threshold confirmation:** cadence demands ≤15 unarchived reports and oldest ≤21 days — that means `git mv` of EVERYTHING dated before 2026-10-24 (~150 files) into `docs/status/archived/` with regime-A banners. This is the largest single mutation of the pass and restructures how the corpus reads. Confirm the threshold (or tell me to seed a `.docs-gates.json` with different cadence knobs — the binary supports per-repo config, and I will NOT silently re-threshold a fleet gate without your call).

---

## ADDENDUM (12:55 CET, written before hand-off)

10. **Batch-03 (24 files) completed by subagent geraldo-annot-03** — every file `check-rows --strict` + `bare-rows` clean, ~1,000+ verdicts, all shas verified with `git cat-file -e` and paths with `test -f`; 5 new HARVEST-CANDIDATEs reported (untouched `docs/planning/` files — next pass; BRS reflection debt — parked in TODO_LIST/ROADMAP; ADR-0007 transport appendix; `docs/archive/` README; fuzz-seed corpus for `isAbsLike` — upstream-gated).
11. **RECURRENCE #16 caught and restored (jsonv2 class).** geraldo's report flagged a pre-existing `TestSARIFOutputGolden` failure; verification showed the **jsonv2gate itself red** — daemon commit `21f8ba8c` (which was already HEAD **before this session started**) re-migrated all three files (`internal/jsonutil/jsonutil.go`, `config/config_migrate.go`, `pkg/enum/enum.go`) to `encoding/json/v2`, 22 lines vs the recurrence-#15 restore source `91419533`. Restored verbatim via `git restore --source=91419533 -- <3 files>` (byte-exact verified by diff against the source); jsonv2gate green, SARIF golden green (it was the v2 marshaler's shadow failure), build green, fresh `-count=1` on `internal/jsonutil/`, `config/`, `pkg/enum/`, `printer/` green. The daemon commits the restore per the established pattern.
12. **Self-indictment recorded:** the #15 lesson says the session-start fresh `-count=1` jsonv2gate run is what catches a red tree arriving from OUTSIDE the hook — this session skipped it and worked ~2.5 hours on a red tree that the golden finally exposed via a subagent's side observation. The session-start gate is command #1 of every session in this repo, before `go build` (§e 1 of the 10:13 report said exactly this; I re-learned it at triple cost).

---

**Verdict:** two of five gates are GREEN (freshness, routing), annotate is 58% cleared with the repair pipeline proven and scripted, the archive ritual is fully specified but not yet executed, and the fleet orchestration cost real money to learn lessons that are now written down. The fastest path to all-green is: resume the annotation grind (Q1), land the round-2 living-doc lines, then run the archive ritual. Waiting for instructions.
