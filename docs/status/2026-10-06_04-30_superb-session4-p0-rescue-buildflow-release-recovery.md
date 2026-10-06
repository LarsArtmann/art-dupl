# Status Report — SUPERB Session 4: Both Handoff Reports Executed (P0 Rescue, BuildFlow Shipped, v0.9.0 Recovered, Consumer Sweep, Recurrence #6)

**Date:** 2026-10-06 04:30 CEST
**Scope:** the two reports the owner pointed at — `2026-10-06_01-41_superb-session3-t05-t18-v0.9.0-released.md` (resume point: BuildFlow step 1 + NEXT list) and `2026-10-06_01-42_docs-health-pass-self-review.md` (P0 list). Standing blanket order; the three session-3 §g questions remain UNANSWERED (checked: report file, TODO_LIST, commits, and BuildFlow's D1–D16 dossier is a different session's board — it does not answer them).
**Repos touched:** art-dupl (fork @ `2e3d3cd7`+, pushed), BuildFlow (master @ `6b3ccc667`, pushed), DiscordSync (`4b1418bb`, local), InboxClean (daemon `e82614b`), code-duplicate-analyzer (`0535f79`), auto-deduplicate (`35e5eca`).

---

## a) FULLY DONE (each verified before moving on)

1. **Docs-health P0 #1 — the /tmp rescue, before anything else.** `/tmp/dh` still existed; driver → `scripts/docs-health-driver.py`, verdict corpus (4,244 lines, ~2,500 per-item verdicts) → `docs/status/2026-10-05_pass-verdicts/` + provenance README. The vanish-class the 09-28 pass swore off is now impossible for this pass.
2. **Docs-health P0 #2 — the two bare review files annotated the DOCUMENTED way** (`--emit-keys` → dry-run on a copy → `--verify` → apply; the key-grammar lesson applied, zero atomic aborts). Verdicts re-verified against code first: `2026-07-25_17-35` (3 of 4 items struck — predicates gone, Self-Analysis CI gate, `TestNoDuplicateErrorNewMessages` at domain/cross_package_alias_test.go:76; withLock/diffStatTable tests stay bare = open) and `2026-06-20_06-50` (Fragment unification + coverage struck done; Clone consolidation + Data→View stay bare; decoupling acceptances struck WONT). Both got dated Resolution banners; check-rows green.
3. **Docs-health P0 #4/#5/#6/#8/#10.** Regime **A-lite** named in `docs/status/archived/README.md` with honesty statement + mechanically-computed recount (**376 snapshots = 57 A / 23 A-lite / 296 B**; method documented so counts stay derived). The 05-21_19-44 FIX-MISS headline fixed (bold-placement mismatch was the cause). All FIVE empty-snapshot "adjacent same-day reports" claims verified per-file with concrete file names appended. Non-md artifact policy decided + written (renders follow their .md; solo HTML reports need the same ANNOTATE-then-move protocol; research HTML governed by the tracking policy). AGENTS.md key-grammar bullet + `docs/SELF_CLEAN_LEDGER.md` lesson entry added. Bonus fix on sight: the stale half-merged `isAssertionMethod` comment the predicate conversion left in actionability_patterns_expanded.go.
4. **BuildFlow T18 completion — the clean resume point, fully cleared.** Root blocker root-caused: `GOWORK=off` leaks from art-dupl's devShell into every shell (plus a stale mid-bump vendor: root/execution still required art-dupl v0.7.2 vs tools' v0.9.0 — the real source of `-mod=readonly`). Fixed: aligned all three go.mods to v0.9.0, `go work vendor`, two `update-vendor-hash` rounds. **All 33 workspace modules green** (`nix run .#test`), root + tools module tests green, `nix build .` green. One REAL gate finding fixed en route: the T18 code had introduced 3 em dashes in .go comments (`TestNoNewEmDashes` — the ratchet gate works).
5. **BuildFlow docs + push.** `docs/configuration.md` `tool_options` row + example; AGENTS gotcha #147 key count 27→28 with semantics; CHANGELOG Unreleased entry. Proper commit `6b3ccc667`, **pushed** — exactly up to my gate-verified commit, leaving the concurrent session's in-flight `d29bbb566` local (theirs to push).
6. **BuildFlow lane soak — tool_options verified end-to-end with the real binary.** Fixture (6+ flat shared statements per gotcha #191): default run = **1 finding** (`alpha.go:6:2`, "16 tokens in 2 instances" — the v0.9.0 column precision visible live), `threshold: 20` = **0 findings**, typo'd `threshhold` = **loud startup failure naming tool + option** (rendered as a clean DI resolve error, better than the expected panic).
7. **v0.9.0 release RECOVERED — it had silently failed.** The goreleaser Release workflow failed on tag push (cosign-installer's floating default jumped v2.5.2→v3.0.6; v3 rejects the config's `--output-signature` flag set: "cannot specify service URLs and use signing config"). No release object existed. Fixed: pinned `cosign-release: v2.5.2` + added a `workflow_dispatch` tag input (retry path for tags whose own workflow is broken), dispatched from fork → **release now live: 52 signed assets** (all platforms, SBOMs, signatures, checksums) + **curated release body** (headlines, 0.8.0-subsumption note, the finding-ID shape change callout). pkg.go.dev verified rendering v0.9.0.
8. **Flake version drift found + structurally fixed.** `nix build github:…/v0.9.0` printed **art-dupl-0.7.0** — the hardcoded `version = "0.7.0"` (flake.nix:33) had drifted through THREE releases despite RELEASE.md's own 2026-09-29 stale-bump warning. Now **derived from CHANGELOG.md's newest `## [X.Y.Z]` header** (assert-guarded; POSIX-ERE bracket-class regex — Nix rejects `\[` escapes and `builtins.split` interleaves capture lists: two real eval quirks navigated). `nix eval` prints `"0.9.0"`. The frozen tag keeps its cosmetic 0.7.0 (tags don't move); noted in the release body.
9. **Consumer sweep to v0.9.0 (go-ecosystem-upgrade skill loaded FIRST this time).** Enumerated with no-ignore: 8 consumers. **DiscordSync**: dedup gate dry-run against v0.9.0 BEFORE flipping (0 new clones vs the 88-group baseline; group hashes unaffected by the ID column change) → tag pin `refs/tags/v0.9.0`, lock updated, devShell builds, PIN-TRAP ritual retired in the comment. **InboxClean**: same flip, sandboxed `dupl` check derivation builds green. **code-duplicate-analyzer**: go.mod v0.9.0, build green (its `internal/errors` test-compile failure is pre-existing, no art-dupl import). **auto-deduplicate**: go.mod v0.9.0, resolves+tidies (its `internal/adapters` build failure recorded as pre-existing BEFORE the bump). **BuildFlow**: already done (a).
10. **jsonv2 recurrence #6 caught at the boundary and restored.** The 33-file daemon sweep `b729d8a7` (session start) had re-migrated BOTH files AGAIN — my closing boundary gate caught it (fresh `-count=1`), restored verbatim from parent (`2e3d3cd7`, pushed), AGENTS.md recurrence #6 entry added. **This is the sixth occurrence; the §g Q1 daemon-hardening question is now six data points old.**
11. **art-dupl loose ends:** TODO_LIST T18 row updated (BuildFlow side done, soak evidence); byte-vs-UTF-16 column semantics documented in HOW_TO_USE LSP section; research README ledger row carries v0.9.0 + consumer status; stale `feat/provider-threshold-knob` branch deleted (content lives in the v2 cherry-pick, verified); README provider paragraph updated (threshold is a declared option, not fixed).

## b) PARTIALLY DONE / DELIBERATELY HELD

1. **CV stays on v0.7.2 (held, with evidence).** Its baseline-ratchet gate FAILS under v0.9.0: the expression-aware templ detection finds NEW clone groups in CV's `.templ` files (e.g. `admin_page.templ:564`, `coaching_page.templ:57`). Accept-vs-extract on those groups is CV's session's/owner's judgment, not a pin-flip call. The v0.7.2 tag pin is stable (no branch race) — no urgency.
2. **SystemNix skipped**: rev-pinned (stable) AND its redeploy is staged for the owner at the console (BuildFlow dossier D15) — a flake-input change now would collide.
3. **DiscordSync/InboxClean/analyzer/auto-dedup commits are LOCAL** (their daemons/sessions push per their own cadence; InboxClean's already went through its daemon with a heuristic message after its pre-commit gate failed on PRE-EXISTING templ breakage — `templ.KeyValue[string, bool]` vs `attributeValue` in generated `chat_templ.go:357`, unrelated to the pin).
4. **Docs-health P0 #3/#7 (A-lite→A re-keying + mechanical banner counts)** — a focused session's work, verdicts now safely in-repo to feed it; owner §g Q1/Q2 of the docs-health report still gate the depth decision.

## c) NOT STARTED (unchanged, ranked)

- **T07b (M38–M39)** upstream go-finding coverage-channel PR (+ ecosystem.md staleness + S4 CLI-discovery gap) — next big block, skill-triplet first.
- Owner answers: daemon hardening (Q1, 6th recurrence now), emit-suppressed-accepted default (Q2), retro-tag v0.8.0 (Q3).
- art-dupl TODO_LIST fleet items (gogenfilter sweep, jsonv2 fleet audit, branch protection #18 — 5th reminder).

## d) WHAT I FUCKED UP (honest ledger)

1. **First CV gate dry-run read the pipe's exit code, not the binary's** (`| tail; echo $?`) — reported RC=0 on a failing gate. Caught on suspicion within one step (usage output in tail) and re-run capturing the real RC=2. The sweep protocol's "verify" step is only as good as the exit code you actually capture.
2. **Two Nix-eval missteps on the version derivation** before it worked: `builtins.split` capture-list interleaving, then POSIX-ERE rejecting `\[` escapes. Both caught by eval errors, fixed with `lib.splitString` + `[[]]` character classes — verified against live eval before committing.
3. **sed possessive-mangling** (`tool's own` → `tool own`) on the BuildFlow comment fix — caught by re-view, fixed with the edit tool. sed on prose remains a footgun.
4. **The release-failure blindness was inherited, not mine, but I verified late**: session 3's report said "goreleaser Release workflow ran" — it ran and FAILED, and nothing checked the release object existed until this session's `gh release view` came back empty. Post-release checklists must include "the release OBJECT exists with assets", not just "the workflow triggered".
5. Recurrence #6 rode a push I made (b729d8a7 lineage went out with my docs push before the boundary gate ran) — the gate caught it minutes later, but the ordering (push-then-gate instead of gate-then-push) is exactly what the task-boundary rule exists to prevent; the daemon racing manual commits makes gate-before-push mandatory even for "docs-only" pushes.

## e) JUDGMENT CALLS worth review

- **Pushing BuildFlow only up to `6b3ccc667`** (not the concurrent session's tip): green-only applied to MY verified lineage; their in-flight `d29bbb566` stays theirs.
- **InboxClean commit landed via its daemon** after the pre-commit gate blocked mine (pre-existing templ breakage). Amending in a foreign repo against its daemon = churn for a message; left as-is with the failure cause documented here.
- **CV held rather than baseline-regenerated**: regenerating a foreign repo's dedup baseline to make a bump pass would be accepting clones nobody judged.

## NEXT (ordered)

1. Owner answers to the three §g questions (Q1 now has recurrence #6 as its sixth justification).
2. T07b upstream PR session (skills first: jj-fork-pr-workflow + verify-before-filing + github-voice).
3. CV v0.9.0 bump after its new templ clone groups are judged (accept into baseline or extract).
4. Docs-health P0 #3: A-lite→A re-keying session (verdicts are in-repo now).
5. self-scan post-column changes; benchstat vs baseline for the O(line) column scan.
6. CLI SARIF relatedLocations/snippet + empty-`category` decision (ledger candidates).
7. Watch: fork CI on `2e3d3cd7`+ (running at report time); the goreleaser Homebrew tap + Docker image legs of the recovered release (formula `skip_upload: true` — tap intentionally untouched).

---

_Point-in-time snapshot. Evidence: commits listed inline; gates: boundary fast green post-restore, 33-module BuildFlow gate green, `nix build .` green both repos touched, soak outputs quoted, CI green on art-dupl @ `2e3d3cd7`-lineage (final tip run in flight at write time)._

_Arte in Aeternum_
