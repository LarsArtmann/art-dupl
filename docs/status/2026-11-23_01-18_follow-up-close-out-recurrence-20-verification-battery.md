# Follow-up close-out: recurrence #20, verification battery, sandbox gate fix

**Date:** 2026-11-23 01:18 CET · **Session:** follow-up to `2026-11-23_00-15_webphone-detection-precision-signature-fragment.md` (same work item; this session executed that report's next-step list)
**Scope line — VERIFIED:** recurrence #20 restore (grep + fresh jsonv2gate + pin tests, all green), `.buildflow.yml` skip honored (`buildflow --dry-run` live output), determinism (two full webphone runs, canonical group-set equality), self-scan delta (two binaries on one pinned tree), anomaly mechanism (token dumps + threshold-gate code + variant matrix), nix flake check (initial FAIL reproduced in log, post-fix **all checks passed**), boundary gate `--full` + fast (both OK). **NOT verified:** which specific group added the self-scan +1 (no October baseline JSON exists to diff against); detected/suppressed totals across the determinism pair (`--quiet` suppressed the footers; shown-group identity is the documented check); the identity of the `.buildflow.yml` deleter (evidence ends at "removed during the 2026-11-15 12:37 pipeline window").

## a) FULLY DONE

1. **Session-start gates** — caught a LIVE `buildflow --fix` pipeline (pid 1171191) and recurrence #20 in progress; `pgrep`-before-forensics discipline held (no restore raced the pipeline).
2. **#20 timeline reconstructed** — file mtimes (00:22:41) prove the poison came from a ~00:21:35 pipeline (minutes after the prior session closed), committed as `cee67b4a` 00:23:31; the 00:32:15 pipeline was innocent (mtimes unchanged under it). Pickaxe confirmed `e16e45f2` as the last v1 source for the six files.
3. **Restore #20, all diff-verified pure** — six v2 files from `e16e45f2`; `go.mod` + `.golangci.yml` pin from `0c486ae9` (pin tests green); `.buildflow.yml` recreated from `c00f3222`. Fresh jsonv2gate green in the same command chain as the restore.
4. **Q1 ROOT CAUSE (the #17–#19 "skip failure")** — `8ce90a47` (#19) **deleted `.buildflow.yml` in its entirety** (41 lines: both skip_steps entries + env + tool_options); the commit message doesn't even list the file. Every pipeline since 2026-11-15 12:37 ran with zero project overrides. The #19 ledger's "skip held" hypothesis was disproven and corrected in place; recurrence #20 entry added to AGENTS.md.
5. **Skip re-armed and verified LIVE** — `buildflow --dry-run`: `go-auto-upgrade: skipped via skip_steps config`.
6. **Self-scan 61→62 delta PROVEN BENIGN** — pre-feature (`2db9fc78`) and post-feature binaries both report `Detected 62` on the same pinned tree; zero groups touch the session's new files. Delta = three weeks of ordinary tree drift.
7. **Webphone determinism PROVEN** — `-t 3 --suggest-generics --type-aware`, parallel vs `--search-workers 1`: 12 shown groups each, canonical (hash, member-set) collections identical.
8. **"Identical Go-snippet → 0 detected" anomaly CLOSED: not a bug** — mechanism fully traced: `trimSubsumedIndexes` makes if+trailing-return = 2 effective statements; `FindSyntaxUnits` (syntax/syntax.go:493) requires `threshold` statements for statement windows; and an identical body extends the window past the signature so non-statement tokens are discarded as noise (masking the signature match that divergent bodies surface). Identical behavior `.go` vs `.templ`. Pinned by `bdd/short_body_threshold_test.go` (2 specs green); documented in AGENTS.md Known Limitations.
9. **corpus-diff.py: suppression-summary dimension** — text-footer parse per binary (detected/shown/non-actionable/filtered) as a fourth diff dimension; self-run empty on branching-flow, head-vs-fix control unchanged (zero collateral confirmed at `-t 3`).
10. **Corpus pin registry** — webphone pin + full registered-pins list + `-t 5` AND `-t 3` dual-threshold validation cadence in docs/benchmarks/README.md.
11. **nix flake check: RED → FIXED → ALL CHECKS PASSED** — initial failure was sandbox-only: jsonv2gate tripped on `gogenfilter-real/cmd/gendocs/main.go` (upstream v2 import in the flake-materialized dependency source). Applied the documented 2026-10-04 rule (tree-walking test gates must skip nested `go.mod` dirs) to jsonv2gate AND preventively to `domain`'s duplicate-errors.New scanner; verified locally with a scratch nested module; full flake battery green (test, race, arch-lint, alloc-gate, fmt, lint, sarif, disabled-linters).
12. **Gates** — boundary `--full` OK (34 packages, incl. lint — also proves the gopls "unused" warnings on the new pattern file are stale); final fast boundary OK.
13. **Prior report updated** — close-out addendum written into `2026-11-23_00-15_*.md`; scratch self-scan worktree removed; daemon committed everything (HEAD `44e5bf39` at report time).

## b) PARTIALLY DONE

- **Owner Q1 (structural fix)**: the skip is re-armed and verified, but the fleet-facing decision (keep skip / upstream BuildFlow exclusion mechanism / migrate the policy) is still owner-gated, and the **deleter of `.buildflow.yml` is unidentified** — a re-deletion would silently re-open the hole (nothing currently fails when the file disappears).
- **Determinism**: shown-group identity verified; detected/suppressed totals were not compared (both runs used `--quiet`; the footer with those numbers was suppressed).
- **Self-scan attribution**: binary-equality proven, but the specific +1 group was not identified (no October JSON snapshot to diff; CI lanes are advisory so nothing gates on it).
- **PATH-shadow warning**: ships in `version` only (prior session); not on analysis runs, no bdd spec (prior report items 13/22 — untouched this session).

## c) NOT STARTED (open items inherited from the prior report's list)

Binary reinstall (user action; PATH copy still resolves to the poisoned-commit build) · `/tmp/art-dupl-{head,fix,warn}` cleanup (blocked on install) · TODO_LIST.md harvest of the two reports' next-lists · `-t 3` bdd fixture for the `.go` test-file param class · multi-line signature-fragment corpus probe · `signature-fragment` scope decision for statement-free `.go` files · bloop/AGENTS "CLOSED" reconciliation · go-cqrs-lite pin re-baseline · CHANGELOG check · `art-dupl doctor` subcommand · PATH sweep on other machines · ACTIONABILITY_PATTERNS defense-in-depth note · `--include-pattern` totals sanity pin · templ component-param GoSnippet count assertion · corpus matrix refresh wave · provider label exposure re-check · low-threshold profile decision · nix `apps` output for canonical `nix run` · reviewer-eye pass on the 8 prior-session TODO rows not restated here.

## d) TOTALLY FUCKED UP (honest ledger of this session's mistakes)

1. **The first #20 restore was INCOMPLETE.** I restored the six v2 files + `.buildflow.yml` but MISSED `go.mod` + `.golangci.yml` — both also poisoned by `cee67b4a`, both explicitly listed in the prior session's restore set. I noticed only ~15 minutes later, BY ACCIDENT, via a buildflow preflight warning (`go-line-flipflop`) I was reading to verify the skip. Root cause: I restored from the handoff's summary instead of diffing the poison commit's FULL file list first (`git show --stat cee67b4a`). The pin tests existed and would have caught it in the boundary gate — but I nearly declared "restore complete, gates green" with a red pin tree. Corrected: checklist is now "poison-commit file list FIRST, per-file diff, then restore".
2. **Self-scan pin created before verifying cleanliness** — the `cee67b4a` worktree was poisoned; I checked only after creating it. No harm (both binaries scanned identical bytes), but the rigor order was wrong.
3. **Gate canary in one direction only** — for the nested-module skip I demonstrated green-with-fixture (skip works), never red-without-fix locally. Mitigation: the sandbox failure WAS the observed red and the fix turned the real `nix flake check` green — but the repo's canary rule ("observed failing before trusted") was satisfied by accident of sequencing, not by design.
4. **My first status claim this session was wrong for ~15 minutes** ("restore verified green" while the go pin was broken) — same class the ledger keeps punishing: a claim of green without the full gate battery behind it.
5. Minor: the foreign untracked file `docs/review/libs/go-linter-sdk.md` (another session's output) was noted mid-session but omitted from my final summary to the user; and I re-ran tests to disprove stale gopls "unused" warnings instead of following the documented LSP-hygiene fix (restart the client) — the boundary gate settled it, but procedurally sloppy.

**Did I lie?** No intentional lies. One inherited claim ("the skip held", from the #19 ledger) was disproven by evidence and corrected in place; one of my own claims (complete restore) was wrong for a quarter hour and self-corrected. Both are documented above and in AGENTS.md.

## e) WHAT WE SHOULD IMPROVE

- **Restore playbook must start from `git show --stat <poison-commit>`**, never from a remembered file list. Per-file diff against candidate sources BEFORE restoring (the #16 lesson, applied late this time).
- **Mechanical tripwire for `.buildflow.yml`**: a test asserting the file exists and contains both skips — a silent deletion currently re-opens the recurrence class with zero signal.
- **Deduplicate the nested-module skip**: the same logic now lives in THREE test files (fuzz gate, jsonv2gate, domain scanner) — a shared `internal/testutil` walker helper would stop the next tree-walking gate from forgetting it (small split brain, created by patches across sessions including mine).
- **Determinism check should compare detected/suppressed totals** too (drop `--quiet`, parse both footers).
- **Canary gates in BOTH directions** (red without fix, green with fix) as a written rule, not an accident.
- **Proactive TODO_LIST harvest**: two reports now carry ~30 open items in timestamped files; docs-health HARVEST is monthly — the top items should land in TODO_LIST.md now, not on the cadence.
- **buildflow hygiene surfaced but not acted on**: state DB at 1.07 GB (VACUUM suggested), binary-freshness warn (built 769b3a4 vs HEAD 40fec90), io-history-ratchet warn — all BuildFlow-repo-side, flagged to owner below.

## f) Up to 50 things to get done next

1. **USER: install the fixed binary** — `cd ~/projects/art-dupl && go install ./cmd/art-dupl` (or rebuild the nix copy); current PATH copy = poisoned-commit `0.9.0-8ce90a4`, pre-fix.
2. **USER decisions**: keep/remove webphone pin; confirm `-t 3` as standing threshold.
3. Add `.buildflow.yml` existence + skip-content tripwire test (kills the silent-deletion class).
4. Hunt the `.buildflow.yml` deleter: buildflow run history around 2026-11-15 12:37 (`buildflow history`), check whether any step deletes unknown-keyed config files.
5. Upstream (BuildFlow repo): exclusion mechanism so go-auto-upgrade respects per-repo v1-API invariants without skip_steps (owner Q1); investigate the config-deletion behavior there too.
6. Refresh the stale fleet buildflow binary (769b3a4 → HEAD 40fec90) so preflight warnings reflect current code.
7. VACUUM the 1.07 GB buildflow state DB (`sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM`). **→ routed to TODO_LIST MEDIUM Priority Decide buildflow state-DB retention (VACUUM reclaims nothing) row** — VACUUM executed 2026-11-23 with the daemon idle: still 1.1 GB, `PRAGMA integrity_check` ok — the space is live data, not bloat.
8. Extract shared nested-module-aware walker into `internal/testutil`; refactor fuzz gate, jsonv2gate, domain scanner onto it.
9. Canary the nested-module skip in the failing direction (temporarily revert the skip in a scratch run, observe red).
10. TODO_LIST.md HARVEST from this report + the 00-15 report (top ~10 rows).
11. Determinism follow-up: rerun the webphone pair without `--quiet`, compare detected/suppressed totals.
12. bdd spec for the PATH-shadow warning (currently unit-level only).
13. PATH-shadow warning on analysis runs when the footer binary isn't the newest copy seen.
14. `art-dupl doctor` subcommand (PATH shadows, toolchain pin, cache-version vs binary, stale baseline).
15. nix `apps` output so `nix run` is the canonical invocation and the system copy stops drifting.
16. `-t 3` bdd fixture for the `.go` test-file param mega-group class (mixed-parser guard end-to-end).
17. Multi-line signature-fragment corpus probe (zero over-suppression check).
18. Decision + doc: should `signature-fragment` cover statement-free `.go` files (currently GoSnippet-gated)?
19. Reconcile resurfaced bloop warnings with AGENTS "CLOSED" claim (prior item 17).
20. Re-pin go-cqrs-lite corpus (drifted vs documented baseline) and re-baseline counts.
21. ~~CHANGELOG check for pattern + warning entries.~~ done — CHANGELOG Unreleased section now carries the signature-fragment pattern + PATH-shadow warning entries (added 2026-11-23 docs-health pass)
22. Sweep other machines/repos for multi-copy art-dupl installs. **→ routed to TODO_LIST MEDIUM Priority Sweep machines/repos for multi-copy art-dupl installs row**
23. ACTIONABILITY_PATTERNS.md: document the signature-fragment ∪ trivial-fragment defense-in-depth in priority notes.
24. `--include-pattern` totals sanity pin (23 vs 261 observation).
25. Assert exact GoSnippet Field counts for templ component params in the pattern test.
26. Re-run the 2026-10-30 corpus matrix with the new binary; refresh all documented numbers in one wave.
27. Re-confirm provider output label exposure after the next provider version bump.
28. Decide the "low-threshold profile" (paired flags) if `-t 3` is confirmed standing.
29. Trash `/tmp/art-dupl-{head,fix,warn}` + `/tmp/sigfrag*` after the user installs. **→ routed to TODO_LIST MEDIUM Priority Install-or-trash the /tmp scratch binaries row**
30. One-line ledger note: `config/config_test.go` v2 import tolerance (test file, gate-exempt — confirm deliberate).
31. Consider asserting `.golangci.yml`/go.mod pin equality from ONE source (the pin test reads both; a third writer could re-split them).
32. Self-scan +1 group identification next time drift matters (snapshot today's JSON as the new baseline for that comparison).
33. Review the foreign untracked `docs/review/libs/go-linter-sdk.md` (another session's output — commit or remove, owner call).
34. Prior-session rows still open: items 10–12, 14–16, 20–21, 24–26 of the 00-15 report's list (not restated individually here).

## g) Questions for the owner (cannot be figured out from the repo)

1. **Binary install**: will you run `go install ./cmd/art-dupl` yourself, or should the nix system-path copy be rebuilt/repointed — and if nix, is `nix build .#art-dupl && nix profile install` the flow you want? (The PATH copy is still the poisoned-commit build; agent-side `go install` is permission-blocked.)
2. **The `.buildflow.yml` deleter**: do you recall what ran around 2026-11-15 12:37 (manual `buildflow --fix`? another agent session?), and did you ever deliberately delete `.buildflow.yml`? If not deliberate, I'll treat its deletion as a BuildFlow-step bug and file it upstream.
3. **Webphone regime**: keep `~/projects/.corpus-pins/webphone@10327ed8` as a standing corpus pin, and is `-t 3` your intended working threshold (drives whether I build out the low-threshold validation profile)?

---

**Verification state at close:** jsonv2gate fresh green · pin tests green · domain/actionability/bdd suites green · boundary `--full` + fast green · `nix flake check` all checks passed · daemon-committed through `44e5bf39` · nothing pushed anywhere.

_Written, waiting for instructions._
