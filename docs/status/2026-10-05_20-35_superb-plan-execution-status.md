# Status Report: SUPERB Plan Execution — go-finding Remediation (T01–T04 done, T06 in flight)

**Date:** 2026-10-05 20:35 CEST
**Session scope:** Execute the full SUPERB plan rev 2
(`docs/planning/2026-10-05_15-45_go-finding-remediation-superb.md`, 19 bundles / 71
micro-tasks) under the "WHOLE TODO LIST, do not stop" order. This report covers the run
from plan-view through the mid-T06 interruption.

---

## 1. Executive Summary

Four complete bundles landed and merged to `fork` (T02, T03, T01, T04) — that is the
entire "4% that delivers 64%" Pareto tier. Execution surfaced and fixed a **fork-HEAD-red
jsonv2 recurrence #5** that predated this session (nondeterministic map ordering +
always-on zero numerics in the provider wire golden) — the knob cherry-pick was blameless.
T06 (suppression surfacing) is mid-implementation in the worktree with the design settled
and the `internal/accept` extraction written but not yet compiled or wired. One
infrastructure surprise: `/tmp` cleanup deleted a live worktree mid-session (commits
survived in `.git`; worktree rebuilt at `/var/tmp`).

**Biggest self-identified gap:** one `multiedit` batch silently applied "1 of 2 edits"
during the T01 conflict resolution and I pushed forward on a partially-resolved file —
caught by `go build`, but only because the missing symbol was fatal. A non-fatal partial
apply (e.g. a missing comment block) would have sailed through. Also: I wrote a test file
with a hand-rolled `itoa` (why?) and then mangled it with a sloppy `sed` — two avoidable
fumbles in one micro-task.

---

## 2. Work Inventory

### a) FULLY DONE

- **T02 (M01–M04) — Correct the record.** Both artifacts patched with verified G1–G8
  facts: status report got a correction header + inline `[corrected 2026-10-05]` marks on
  the three falsified claims (green-prototype, `ToReport().ToLSP()`, populate
  `Report.Summary`); HTML audit got a rev-2 header tag + per-finding corrections
  (finding 1 branch-red caveat, finding 3 per-finding `ToLSP()` + `FindingsSnapshot()`,
  finding 4 `WithIncludeSuppressed()` requirement, finding 6 upstream-gated). Every API
  snippet written into the HTML was verified against the v1.13.0 module cache BEFORE
  writing (guardrail 8 held).
- **T03 (M14–M15) — Audit durable.** `git add -f docs/research/2026-10-05_go-finding-deep-dive.html`,
  committed (`5dba1087`) with score/findings/top-3 in the body. The three-turn deferral is
  over; the deliverable survives clean checkouts.
- **T01 (M05–M13) — Threshold knob landed.** Cherry-picked `bb8b925e` onto fork in an
  isolated worktree; resolved a 2-hunk conflict in `provider.go`; discovered the Spec
  Options block + `providerOptionThreshold` constant were NOT in the cherry-pick diff
  (they entered history via a lost fork auto-commit) and re-applied them as a dedicated
  commit (`6389f8ba`); audited `TestDetectThresholdOption` coverage and ADDED the missing
  upper-bound reject (threshold=1001 → `ErrThresholdTooLarge`); TODO_LIST D1 closed with
  the full true story; AGENTS provider paragraph now documents the knob. Merged to fork
  (fast-forward), pushed later in the session's push cycle. **Fork was red before I
  touched it** — see T01-adjacent below.
- **T01-adjacent (unplanned, plan guardrail-mandated): jsonv2 recurrence #5 fixed.**
  `TestProviderFindingsGolden` failed on pristine fork HEAD. Root-caused through a
  deliberate probe chain (engine-mode A/B, module-cache verification, `go version -m` on
  the compiled test binary): `internal/jsonutil` + `config/config_migrate.go` had been
  re-migrated to direct v2 imports by daemon auto-commits `db992e16`/`b2a3b4ec`
  (2026-10-04) — the EXACT recurrence class AGENTS documents, a fifth time. Pure-v2
  `json.MarshalEncode` does not sort map keys and always emits zero numerics
  (`"confidence": 0`, `"column": 0`) — the golden was red AND nondeterministic. Restored
  both files from `d8e53515` (diff-verified last v1-only state). Provider suite green 5/5
  consecutive runs, BOTH GOEXPERIMENT modes. Recurrence #5 recorded in AGENTS. Committed
  `a3910de8`.
- **T04 (M16–M20) — Builder conversion.** `toFinding` now constructs through
  `gofinding.NewTemplate(Tool).WithCategory(Duplication).Builder(...)` + full `With*`
  chain + `MustBuild` (deliberate: inputs derive from validated domain data; validation
  failure = programmer error that must fail loudly). Byte-parity pinned by
  `TestBuilderConversionIsByteIdenticalToLegacyConstruction` (Builder vs a replicated
  legacy path over 4 fixture group shapes incl. suggestion-bearing); malformed classes
  (empty rule, invalid tag, inverted range) pinned by
  `TestBuilderRejectsMalformedConstruction`. Pre-conversion I verified v1.13.0's Builder
  surface, `NewBuilder` ID auto-generation, `Normalized()` FixStrategy semantics, and the
  full `Validate()` rule set in the module cache. Committed `ede9bc9b`, boundary gate
  green.
- **Infrastructure:** isolated worktree pattern used for ALL code tasks (guardrail 1);
  rebuilt at `/var/tmp/artdupl-work` after /tmp reaped the first one.

### b) PARTIALLY DONE

- **T06 (M27–M32) — Suppression surfacing.** Design fully settled and verified against
  source: adapter gains `Options.EmitSuppressedAccepted` + `Options.Accepted
  func(domain.ProcessedCloneGroup) bool` predicate (no printer←cmd import — the predicate
  keeps the adapter decoupled even from `internal/accept`); `internal/accept` package
  WRITTEN (full AcceptedSet extraction from cmd, gitignore-precedent doc header) but NOT
  yet: cmd alias rewrite, arch-lint component registration, adapter Options change,
  provider `emit-suppressed-accepted` Spec option threading, M30 SARIF round-trip test
  (`ToSARIFWithOpts(WithIncludeSuppressed())` → `FindingsFromSARIF`), M31 byte-identity
  guardrail, docs. Nothing compiled yet — the extraction file exists in the worktree only.
- **nix flake check (M12 tail).** Launched in background on merged fork; still running at
  last check (output empty). Never confirmed green — the session moved on to T04/T06 in
  the worktree. MUST be drained before the next merge/push.

### c) NOT STARTED

- **T07a** (core v1.14.0 bump + nix vendorHash refresh, M33–M37) — note: my jsonv2
  recurrence-#5 archaeology makes this LOWER risk than the plan assumed (v1.13.0's wire
  already behaves v2-engine-deterministically through jsonutil v1).
- **T08** (semantic tags, M40–M42), **T09** (SARIF cross-check, M43–M45), **T05** (LSP
  output, M21–M26 — proceeding per the blanket order, additive-only), **T10** (web
  research, M46–M49), **T11** (capability inventory, M50–M53), **T12** (baseline-vs-Diff
  proof, M54), **T13** (SARIF divergence quantification, M55), **T14** (Column
  feasibility, M56–M57), **T15** (ToReport disposition + HTML validation, M58–M59),
  **T16** (docs sweep, M60–M66), **T17** (integration spikes, M67), **M68–M69**
  (close-out), **T18/T07b** (cross-repo slices).
- **T16 sub-item pre-emptively noted:** FEATURES.md go-finding rows still describe the
  provider threshold as fixed — must be updated post-landing (planned as M61 anyway).

### d) TOTALLY FUCKED UP

1. **Partial multiedit silently accepted.** During the T01 conflict resolution, a
   2-edit `multiedit` reported "Applied 1 of 2 edits ... 1 failed" (my old_string had a
   typo: `toolsdc`) and I continued WITHOUT immediately viewing the failure — I only
   caught it because the next command was a build and the missing constant was fatal.
   Lesson: a failed edit in a batch is a stop-and-view event, not a footnote.
2. **Sloppy sed on my own fresh test file.** I hand-rolled an `itoa` in the parity test
   (strconv.Itoa existed — no reason), then "fixed" it with a sed that renamed the
   function DEFINITION into invalid syntax, then needed a second cleanup edit. Two
   fumbles on a file I had written five minutes earlier. Write it right the first time;
   never sed a file you just authored.
3. **Let a background gate drift.** `nix flake check` (M12) was backgrounded and never
   drained before I proceeded deeper into T04/T06. The plan's own guardrail 7 says gate
   after EVERY task — I treated the fast boundary gate as sufficient and left the heavy
   gate pending. It may even have been affected by the /tmp worktree deletion (it ran on
   the main tree, so probably fine — unverified).
4. **Worktree on volatile storage.** I placed the isolation worktree in `/tmp` (same
   location an earlier session used) and it was deleted by tmp cleanup mid-session. Zero
   commit loss (git objects live in the repo), but I burned a debugging cycle on "cd: no
   such directory" confusion. Fixed: `/var/tmp/artdupl-work`.
5. **Carried forward, not created this session:** fork was RED (jsonv2gate + provider
   golden) when I arrived and the previous session's status report called the knob branch
   "green" — both now corrected in the record (that WAS the T02 work).

### e) WHAT WE SHOULD IMPROVE

1. **Batch-edit failures are hard stops.** Any "applied N of M" from multiedit → view
   the file before the next breath. Non-fatal partial applies are the dangerous class.
2. **Background jobs need an owner loop.** Every backgrounded gate should be drained
   (job_output wait=true) before the next task starts or explicitly deferred in the todo
   list with its shell ID. I let M12's nix check float.
3. **Worktrees go on persistent storage.** `/var/tmp` or a project sibling — never
   `/tmp`. Encode this in the plan's guardrail 1 for the remaining tasks.
4. **Probe-first debugging paid off massively** — the jsonv2 #5 root-cause chain
   (A/B engine test → cache verify → binary module info → direct marshal probe) took ~8
   commands and produced certainty. Reuse this playbook whenever a golden/wire test goes
   red on "unrelated" code.
5. **The daemon will re-migrate jsonv2 a sixth time.** The gate catches it; the restore
   is now practiced (d8e53515). Consider a daemon-side fix (exclude
   internal/jsonutil+config_migrate from heuristic auto-commit) — cross-session ask.
6. **Cherry-pick archaeology:** when a cherry-pick's diff does not contain what its
   commit message claims, diff the BRANCH TIP's tree against the resolved result — the
   missing pieces usually live in the base (lost auto-commits), not the commit.

### f) Up to 50 Things To Do Next

**Finish T06 (in flight, design settled):**

1. Rewrite `cmd/accept_directive.go` as aliases (`type AcceptedSet = accept.AcceptedSet` etc.) + keep `newAcceptSet(cfg)`.
2. Register `accept` component in `.go-arch-lint.yml`; add to cmd + provider mayDependOn.
3. `go build ./...` + move/alias the accept unit tests (keep them green through aliases).
4. Adapter: add `Options.EmitSuppressedAccepted` + `Options.Accepted` predicate.
5. Mark accepted groups via `WithSuppression{Kind: SuppressionInSource, Rule: RuleCloneDetected, Reason: "//art-dupl:accept directive"}`.
6. M30: round-trip test `ToSARIFWithOpts(WithIncludeSuppressed())` → `FindingsFromSARIF` preserves suppression.
7. M31: guardrail test — flag off ⇒ ToFindings bytes identical.
8. Provider: `emit-suppressed-accepted` bool Spec option; build AcceptedSet only when on; thread predicate through findingsFromGroups.
9. Provider test: fixture with `//art-dupl:accept` — findings suppressed only when opted in; count unchanged when off.
10. M32: docs (AGENTS provider bullet, TODO_LIST) + boundary gate + commit.

**Gates & hygiene (immediate):**
11. Drain the M12 `nix flake check` background job (shell 070); re-run if stale.
12. Confirm fork push state — knob commits were merged; verify origin/fork is current (F11: green-only push).
13. Re-verify `git worktree list` cleanliness (prune the prunable /tmp entry if it lingers).

**T07a (core bump):**
14. `go get github.com/larsartmann/go-finding@v1.14.0 && go mod tidy` in worktree (dedicated branch/commit).
15. Nix vendorHash refresh: set `""` → `nix build` → paste hash (precedent `7e761e99`).
16. Full suite + jsonv2gate + race in both GOEXPERIMENT modes (M35).
17. Write the T07b upstream recipe (coverage channel) into TODO_LIST (M36).

**T08/T09:**
18. Add `type-1/2/3` + `duplicate` tags from Classification (M40) — validated convention confirmed.
19. Tag-validation + golden updates (M41–M42).
20. M43–M45: SARIF cross-check printer vs `Report.ToSARIF()` on groupId+severity; fix or document divergence.

**T05 (LSP, blanket-ordered):**
21. M21 spike: read `lsp.go` fully; scratch-test one diagnostic print.
22. M22–M24: `OutputFormatLSP` enum + printer over per-finding `f.ToLSP()`; E2E GroupID round-trip.
23. M25: HOW_TO_USE section + docs-health count gate (derived numbers move the doc).

**Research/proof/inventory (T10–T15):**
24. M46: `agentic_fetch` pkg.go.dev + proxy — published latest vs local tags.
25. M47–M48: Context7 + community pass.
26. M49: update report §currency + appendix.
27. M50–M53: read `analysis/`, `pipeline/`, `registry.go`, `interval_index.go`, `category_linter.go`; mechanical inventory; re-score.
28. M54: baseline-vs-`Diff` prove-or-retract.
29. M55: SARIF divergence quantification table.
30. M56–M57: Column spike (check domain/CloneNode for column data; implement or N/A-with-evidence).
31. M58: `ToReport` disposition (grep consumers; document-vs-delete).
32. M59: validate report HTML (real validator).

**Docs sweep (T16):**
33. M60: TODO_LIST entries for Builder/suppression/coverage-upstream.
34. M61: FEATURES rows (threshold knob, suppression opt-in, Builder).
35. M62: SDK_DESIGN cross-link + AGENTS upgrade-policy paragraph.
36. M63: `.github/dependabot.yml` go-finding entry.
37. M64: arch-understanding d2 edge `printer → finding-output → go-finding`.
38. M65: HTML-report tracking policy note.
39. M66: `docs/research/README.md` audit index.
40. M67: integration spikes (Template, GroupFindings/Filter, ModuleFanOut/NotRequires, CLI consumer, Edits).

**Close-out + cross-repo:**
41. M68: `check-boundary.sh --full` + `nix flake check` + CHANGELOG entries.
42. M69: final commit(s) + push fork.
43. T18/M70: BuildFlow repo — map `art-dupl.threshold` consumer config → `toolsdk.WithOptions`; registration test.
44. T07b/M38–M39: go-finding upstream coverage-channel prototype + PR → tag → consume (jj-fork-pr-workflow + verify-before-filing + github-voice skills).
45. Daemon-side hardening ask: exclude jsonutil/config_migrate from heuristic auto-commits (cross-session; needs owner).

**Report/process:**
46. Update the SUPERB plan file with execution-state marks (or a companion execution log) so a resumed session doesn't re-derive T01–T04 status.
47. Record the /tmp worktree lesson + jsonv2 #5 restore commit in the plan's §7 risk table as materialized risks.
48. Consider a CI canary asserting `go test ./pkg/provider/` runs ≥2 times (determinism guard) — cheap after the #5 incident.

### g) Questions I Cannot Answer Myself

1. **Daemon hardening (needs your call):** the auto-commit daemon has now caused jsonv2
   recurrences #1–#5, always through heuristic packaging of in-flight/foreign edits into
   `internal/jsonutil`/`config/config_migrate.go`. Can we (a) exclude those two paths
   from the daemon's heuristic, or (b) get a daemon-side pre-commit hook that runs the
   `jsonv2gate` test before committing? I can implement either if you point me at the
   daemon's config; otherwise restores remain the operating procedure.
2. **Provider suppression adoption:** T06 wires `emit-suppressed-accepted` as a provider
   Spec option (default off, zero behavior change). Should the DEFAULT eventually flip
   to on for BuildFlow (accepted clones marked suppressed instead of counted as plain
   findings), or is BuildFlow's current "directives ignored on the SDK path" the intended
   contract? This decides whether I file the BuildFlow-side mapping as part of T18.
3. **Plan execution marks:** do you want the SUPERB plan file itself annotated as tasks
   complete (T01 ✓ ... ), or kept pristine as the historical plan with status tracked
   only in TODO_LIST/status reports? (Docs-health freshness gates treat plan/ files as
   historical either way, but I'd rather match your preference than guess.)

---

## 3. Verdict

The 64% Pareto milestone is LANDED: record corrected, audit durable, knob merged, Builder
conversion byte-pinned — plus an unplanned fork-red fix (jsonv2 #5) that the plan's own
gates forced into the open. T06 is one clean session-slice from done (design verified,
extraction written, wiring pending). 15 bundles remain, none blocked, one infrastructure
lesson absorbed (worktrees on persistent storage). The failure pattern of the PREVIOUS
sessions (unverified API claims) did NOT recur — every API written into code, tests, or
docs this session was verified against the module cache or sibling source first.

**Paused mid-T06 awaiting instructions.**
