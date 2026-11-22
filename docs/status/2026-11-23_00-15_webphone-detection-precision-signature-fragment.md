# Status: webphone detection precision — signature-fragment pattern, PATH-shadow warning, recurrence #19 restore

**Date:** 2026-11-23 (session started 2026-11-22 late evening)
**Trigger:** User reported art-dupl junk output on `~/projects/webphone` (`-t 3 --suggest-generics --type-aware`): a 36-member group of one-line param fragments (`dir string`, `raw string`, `lang Lang`, `contacts CRMContacts`) plus dozens of one-line templ signature/type fragments drowning the genuine findings. User: "feels like we have some kind of bug in the way we detect duplicates (THAT ARE deduplicatable in Golang v1.27+!!!) … or .templ".

## Root causes (both confirmed)

1. **Stale binary.** The user's `art-dupl` resolved to `~/go/bin/art-dupl` = `v0.9.1-0.20261017225421-93432fd910fb` (commit 2026-10-18, **144 commits behind** HEAD). The `dir string` mega-group and the mixed `.go`+`.templ` param groups are the junk class already killed by the 2026-10-30 precision pass (issues #10/#11: mixed-parser guard + `trivial-fragment` shape label). At current HEAD the same runs show 14 (not 35) groups at `-t 3`, 5 (not 19) at `-t 4`, all genuine at `-t 4`.
2. **Real residual bug (fixed this session).** At `-t 3`, current HEAD still showed 2 junk groups (11 members): templ **Go-snippet signature tails** — `lang Lang) string` ×9 and ×2, the shared parameter/result tail of `func` helpers embedded in `.templ` files. The suffix tree matches the longest common token substring; two helpers whose signatures merely END alike surface the shared tail as its own fragment. Identical signature parts are never extractable (bodies differ; a signature cannot be "extracted"). `trivial-fragment` misses them (2 tokens, not 1); the mixed-parser guard exempts `.templ` files by design.
3. **Environment cause of #1 (found late, mid-report):** the machine carries MULTIPLE art-dupl installs. At session end: `~/go/bin/art-dupl` and `~/.local/bin/art-dupl` were removed by the user mid-session; PATH now resolves to the nix system-path copy = **`0.9.0-8ce90a4` — built from the recurrence-#19 POISONED daemon commit, still pre-fix**. The user asked for a warn-on-multiple-PATH-copies feature (implemented, below).

## a) FULLY DONE

1. **Session-start gates caught recurrence #19**: jsonv2gate red — daemon commit `8ce90a47` ("8 changed files") re-poisoned the SIX v2 files (jsonutil, config_migrate, config/enum_helpers, pkg/enum, baseline.go, pkg/artdupl/types.go). Last known-v1: `d9e66e58`; diff verified pure v2 import swaps, zero feature work → single-source restore, gate green in the same command chain. Ledgered as #19 in AGENTS.md (with the note that the `.buildflow.yml` `go-auto-upgrade` skip did NOT prevent it — sweep came through a path that ignores skip_steps; mechanism UNVERIFIED, marked as suspect).
2. **Go-line pin restore** (same daemon commit): go.mod `go 1.27` → `1.27.1`, `.golangci.yml` `run.go` `"1.27"` (quoted) → `1.27.1` (unquoted, canonical format). Pin tests green.
3. **Stale-binary diagnosis proven via git history**: `93432fd910fb` predates the 2026-10-30 precision pass; precision-pass code (`trivial-fragment`, `anyFragUnitIsNonStatementGo`) confirmed present at HEAD.
4. **Corpus pin created**: `~/projects/.corpus-pins/webphone@10327ed8` (worktree in the USER's repo — see questions).
5. **Reproduced all three user runs on the pin with HEAD** (`/tmp/art-dupl-head`): `-t 3` combined 35→14 shown; `-t 4` combined 19→5; `-t 4` plain 19→5. Full triage: at `-t 3` 12/14 genuine, 2 junk; at `-t 4` 5/5 genuine in every mode (incl. both `generics:` candidates — `*app.ShutdownError` vs `*webphone.bootJoinError`, `readmodels.CallReviewView` vs `CallNoteView`).
6. **`signature-fragment` denylist pattern implemented** (`printer/actionability/actionability_signature_fragment.go`): fires when EVERY instance's root is a GoSnippet node with base type ∈ {FuncType, Field, FieldList}. Gated on the new `domain.CloneNode.GoSnippet` field (bridge: `printer/clone_processor.go::syntaxToCloneNode`) because the golang/templ enums overlap numerically (ADR-0027 file-kind scoping). Narrow by construction: statement fragments untouched, markup untouched, plain `.go` files untouched (mixed-parser guard already handles those).
7. **Tests**: real-templ-source unit tests (`actionability_signature_fragment_test.go`: pattern fires on parsed `lang Lang` fields, markup roots don't match, plain-Go Field doesn't match, mixed group doesn't match, disable semantics) + BDD e2e (`bdd/templ_signature_fragment_test.go`: suppressed at `-t 3`, surfaced+labelled by `--show-suppressed --explain`, restored only with BOTH `--disable-pattern signature-fragment` and `trivial-fragment` — defense in depth, the 1-token shape is also caught by trivial-fragment).
8. **Docs**: ACTIONABILITY_PATTERNS.md row; AGENTS.md pattern list + counts + recurrence-#19 ledger + header date; FEATURES.md counts + corpus line (added webphone; corrected to 5,600+ Go / 65+ templ — first draft wrongly included nsfw-classifier, fixed); DOMAIN_LANGUAGE.md count; docs-health count gate ratchet 42→43 (test pins total; FEATURES claims updated to match derived numbers).
9. **SARIF golden regenerated** (intentional wire change: new `art-dupl/pattern/signature-fragment` rule descriptor) per the documented `-update` procedure.
10. **Verification battery**: full suite green (`-count=1`, 34 pkgs); BDD 373 specs green; boundary fast gate green (build+vet+scoped race+golangci-lint); jsonv2gate green; self-scan `-t 5` 0 shown; **5 pinned corpora byte-identical before/after** at default threshold (corpus-diff.py); DiscordSync identical even at `-t 3` (zero signature-fragment suppressions — pattern is genuinely narrow); webphone member-set diff: exactly the 11 junk members removed, 0 additions.
11. **PATH-shadow warning implemented** (user request, second message): `cmd/path_shadow.go` — plain `art-dupl version` (NOT `--json`/`--short`, machine formats stay clean) walks PATH, stats `art-dupl[.exe]`, dedupes by resolved symlink realpath (nix system-path link + store target count once), probes each foreign copy's own `version --short` (2s timeout, stubbed seam in tests), and prints a stderr warning naming every copy with version + "running now" / "first in PATH, wins" markers. 4 unit tests green; live-demoed (it exposed the nix copy being `0.9.0-8ce90a4`); FEATURES.md version row updated; boundary gate green over the final tree.

## b) PARTIALLY DONE

- **Binary refresh on the user's machine**: NOT done. `go install ./cmd/art-dupl` is permission-blocked in this session (dangerous-command class). Tested binary sits at `/tmp/art-dupl-warn` (and `/tmp/art-dupl-fix`, behaviorally identical). The nix system copy (`0.9.0-8ce90a4`, poisoned-commit build) is what `art-dupl` currently resolves to — **the user is still effectively on a stale+poisoned build** until they rebuild/install.
- **Recurrence #19 forensics**: restore done, ledger done, but HOW `8ce90a47` bypassed the `.buildflow.yml` skip was NOT investigated (hypothesis recorded as hypothesis). The structural fix remains owner-gated Q1.

## c) NOT STARTED

- `nix flake check` on the feature tree (format/race/alloc-gate/arch-lint nix lanes) — only the Go-side boundary gate ran.
- `scripts/check-boundary.sh --full` (the full race-suite close-out gate; fast profile ran instead).
- webphone generics-candidate RECALL audit (precision validated; never checked for MISSED real generics-extractable clones).
- Determinism re-check (parallel vs `--search-workers 1`) on webphone output.
- Registering the webphone pin in `docs/benchmarks/README.md` (pin convention documented there; no registry table updated).
- Root-causing two anomalies noticed in passing (see d/e).

## d) TOTALLY FUCKED UP (honest ledger)

- **First bdd fixture silently tested nothing**: my initial one-line-body fixtures (`return "a"` vs `return "b"`) produced **0 detected groups** — semantic literal-kind normalization makes the bodies identical, changing the fragment shape entirely. All 3 specs "failed" for the wrong reason (group never formed, not suppressed). Caught because I ran the suite; fixed with divergent bodies mirroring the real webphone shape. Lesson: fixture bodies must DIVERGE past normalization when the match is meant to stop at the signature.
- **First live demo of the PATH warning didn't fire** (binary named `art-dupl-warn` in /tmp — no `art-dupl` entry). Sloppy staging; fixed with a proper second copy.
- **FEATURES corpus counts first draft wrong** (wrote 6,000+/85+ by summing ALL pins including nsfw-classifier, which is not part of the documented 5-repo corpus; corrected to 5,600+/65+ the same session).
- **`t.Setenv` inside `t.Parallel()` test** panicked on first run (Setenv forbids parallel) — fixed by injecting the PATH string instead of mutating env (which is also better design).
- **Walked past a real anomaly**: two templ files with byte-identical Go-snippet funcs reporting **0 detected groups at `-t 3`** (no `--no-actionability` either). I noticed, shrugged, moved on. Identical files SHOULD report line-scoped groups (the phantom bdd pins prove that for templ components). Possible masking bug, uninvestigated.
- **Self-scan drift 61→62 detected** hand-waved as "tree drift" without proving the +1 (likely my own new test fixtures; CI lanes for it are advisory so nothing breaks, but the claim was unverified).

## e) WHAT WE SHOULD IMPROVE (systemic, from this session)

1. **corpus-diff.py is blind to suppression changes** — it compares `--json` DETECTED groups; my suppression-only change was invisible to it (had to hand-run `-t 3` shown-count diffs on DiscordSync). Add a shown/suppressed-count comparison mode.
2. **Low-threshold regime is under-validated**: every precision pass was validated at DEFAULT threshold only; both junk classes the user hit live at `-t 3`/`-t 4`. Make `-t 3` part of the corpus validation cadence (or a dedicated lane).
3. **Binary distribution is the #1 support cost**: three installs, three versions, none current. The PATH-shadow warning helps; a `--upgrade-hint` or single canonical install path (nix profile) would help more.
4. **The jsonv2 recurrence (19x) and go-line flip-flop both arrived in ONE daemon commit** (`8ce90a47`) through a path that ignores skip_steps — the skip-based defense has a hole; Q1 (upstream exclusion) is overdue.
5. **Session-start gate discipline works** — recurrence #19 and the go-line flip were both caught before any feature work; keep the fresh `-count=1` jsonv2gate run mandatory at every session start.

## f) NEXT (up to 50, ranked)

1. **USER: refresh the installed binary** — `cd ~/projects/art-dupl && go install ./cmd/art-dupl` (permission-blocked for the agent), or `nix build .#art-dupl` and repoint the system-path/profile copy; current resolution is the poisoned `0.9.0-8ce90a4`.
2. Run `nix flake check` on the feature tree (format/race/alloc/arch-lint lanes).
3. Run `scripts/check-boundary.sh --full` (close-out gate).
4. Root-cause the "identical templ Go-snippet files → 0 detected groups" anomaly (possible masking bug; fixture in this session's /tmp/sigfrag history).
5. Prove the self-scan 61→62 detected delta source (expect: the two new test fixtures).
6. Determinism check on webphone (parallel vs `--search-workers 1`, sorted group sets byte-identical).
7. Owner Q1: investigate how `8ce90a47` bypassed `go-auto-upgrade` skip_steps (git timestamps + BuildFlow logs).
8. Owner Q1: upstream BuildFlow exclusion mechanism for ADR-0024 repos (end the restore-forever loop).
9. Extend corpus-diff.py with shown/suppressed count comparison (suppression-visible mode).
10. Add `-t 3` corpus validation to the precision-pass protocol (documented cadence entry).
11. Audit webphone for MISSED generics candidates (recall, not precision).
12. Register webphone@10327ed8 in docs/benchmarks/README.md corpus notes.
13. Consider warning on analysis runs too (not just `version`) when PATH shadows — maybe only when footer version != newest copy seen.
14. Add a `-t 3` self-maintaining bdd fixture for the `.go`-test-file param class (`dir string` mega-group) pinning the mixed-parser guard end-to-end.
15. Review whether `signature-fragment` should also cover statement-free `.go` files (currently GoSnippet-gated; identical struct-field groups in pure-type files stay visible — deliberate, but undocumented as a decision).
16. Multi-line signature fragments: no line gate by design — add a corpus probe for multi-line Field/FieldList-rooted groups to confirm zero over-suppression in the wild.
17. Reconcile the resurfaced bloop warnings (`extractability_bench_test.go` b.N) with AGENTS.md's "CLOSED" claim.
18. go-cqrs-lite corpus pin is stale vs its documented baseline (repo drifted under its own daemon) — re-pin and re-baseline counts.
19. CHANGELOG: check whether the repo keeps one and whether the pattern + warning warrant entries.
20. Consider `art-dupl doctor` subcommand aggregating: PATH-shadow check, go/toolchain pin check, cache-version vs binary, stale-baseline detection.
21. Sweep other LarsArtmann machines/repos for multi-copy art-dupl installs (the warning only helps when `version` is run).
22. Add the PATH-shadow warning output to a bdd spec (currently unit-level only).
23. Document the defense-in-depth relationship (signature-fragment ∪ trivial-fragment) in ACTIONABILITY_PATTERNS.md priority notes.
24. Investigate why `--include-pattern` scoping run showed different totals (23 groups vs 261 — expected, but worth a sanity pin).
25. Templ `templ List(lang Lang)` component params: verify they classify as GoSnippet Fields too (fixture suggested they might not — the pattern test found ≥2, did not assert exact count).
26. Re-run the 2026-10-30 precision-pass corpus matrix with the new binary to refresh all documented numbers in one wave.
27. Check whether `printer/finding` provider output should expose the new pattern label (currently CLI-only classification — deliberate, but re-confirm after provider version bump).
28. The user's `-t 3` habit: if intentional, consider a documented "low-threshold profile" (paired flags) instead of raw `-t 3`.
29. Remove/keep `/tmp/art-dupl-fix`, `/tmp/art-dupl-head`, `/tmp/art-dupl-warn` scratch binaries after user installs.
30. Consider nix flake `apps` output for a canonical `nix run` path so the system copy stops drifting.

## g) QUESTIONS FOR THE OWNER (cannot be answered from the repo)

1. **Binary install**: `go install` is agent-blocked here — do you want to run `go install ./cmd/art-dupl` yourself, or should the nix system-path copy be rebuilt/repointed (and if nix, is `nix build .#art-dupl && nix profile install` the flow you want)?
2. **Keep the webphone corpus pin?** `~/projects/.corpus-pins/webphone@10327ed8` is a worktree attached to your webphone repo (`.git/worktrees` metadata). Keep it as a standing corpus entry, or remove it?
3. **Is `-t 3` your working threshold for webphone?** If yes, I'll treat the low-threshold regime as a first-class validated profile (corpus cadence + more fixtures); if it was a one-off probe, default-threshold validation stays the only cadence.

## Verification state at close

- Full suite `-count=1`: **green** · BDD 373 specs: **green** · boundary fast gate: **green** · jsonv2gate: **green** · pin tests: **green** · docs gates: **green** · self-scan `-t 5`: 0 shown · corpora: 5/5 identical before/after · webphone `-t 3` combined: 12 shown, all genuine.
- Working tree: uncommitted (daemon will auto-commit); nothing pushed anywhere.
