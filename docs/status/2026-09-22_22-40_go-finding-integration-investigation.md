# Status Report: go-finding Integration Investigation Session

**Date:** 2026-09-22 22:40 CEST
**Session scope:** Two turns, read-only. Turn 1: answer "why doesn't art-dupl use go-finding and its ecosystem?". Turn 2: this report. Zero production code touched; zero repo mutations by this session (auto-commit daemon may have picked up nothing).

---

## Self-Critique (what I forgot / could have done better)

1. **Assertion-before-verification (process lapse, claims were correct).** In turn 1 I asserted two facts sourced from `docs/status/2026-09-19_08-33_go-finding-gap2-adapter-status.md` without opening code: (a) "no test pipes art-dupl's real SARIF bytes through `FindingsFromSARIF`", (b) "three interchange guarantees pinned by tests". I only verified both in turn 2 (`printer/finding/finding_test.go:384` — `TestSARIFRoundTripPreservesGroupID` does `report.ToSARIF()` → `FindingsFromSARIF`, i.e. go-finding's own serializer, NOT art-dupl's hand-rolled CLI SARIF). Both claims true, but the verification debt was paid one turn late. This is exactly the `verify-external-claims` anti-pattern (inbound flavor: status reports are external claims too).
2. **Never read the adapter source.** All adapter behavior claims (metadata keys, 254 lines, GroupID contract) came from the 2026-09-19 status report and AGENTS.md. I never opened `printer/finding/finding.go`. Correct outcome, lucky process.
3. **Never ran the tests.** I cited "pinned by tests" without executing `go test ./printer/finding/...` once this session.
4. **TODO_LIST.md unchecked.** I reported "three known gaps" without confirming whether they are already tracked as TODO_LIST tasks — risk of re-surfacing queued work as if new.
5. **Dead-end call:** `mcp_qmd_get` with relative path `../go-finding/...` failed (QMD collections don't cover the sibling repo) — wasted call; `view` worked immediately after.
6. **Noisy shell:** first combined bash command exited 1 (final grep in a chain found no matches → non-zero exit, indistinguishable from failure at a glance). Re-ran split. Same again on the ADR-dir grep.
7. **Soft close:** ended turn 1 with an open offer ("Want me to close any of those three?") instead of a decisive recommended next action. Acceptable while waiting for instructions, but less useful than a ranked recommendation.

---

## a) FULLY DONE (this session)

| # | Item | Evidence |
|---|------|----------|
| 1 | Root-cause answer delivered: art-dupl DOES use go-finding — as output/interchange layer at the boundary, by deliberate decision, not oversight | go.mod:88 (`go-finding v1.12.0` direct dep), `printer/finding/finding.go` adapter |
| 2 | Located and fully read the governing decision doc: `go-finding/docs/feedback/2026-06-05_art-dupl-integration-evaluation.md` (373 lines) — verdict: adopt as consumer-side output layer, NEVER internal representation; appendix records the rejected alternative (domain mismatch: findings = single issue/position, art-dupl = clone groups = N-location relationships) | Read lines 1-373 |
| 3 | Verified production wiring state: adapter is wired at **constant level only** — `printer/sarif.go` imports it for `SARIFPropGroupID`; `ToFindings`/`ToReport`/`GroupIDOf` have **zero production callers** (grep across `printer/`, `cmd/`, tests excluded) | grep, 2026-09-22 |
| 4 | Confirmed NO ADR exists for the go-finding adoption decision (`docs/adr/` stops at 0024-json-v1-api-behind-canonical-marshaler) — the status report's item 23 remains open | `ls docs/adr/` |
| 5 | Verified test-suite shape (turn 2): 17 test functions in `finding_test.go`, incl. the three interchange guarantees (SARIF round-trip :384, LSP round-trip :435, Report.GroupFindings reconstruction :336) and 10 ToFindings behavior tests | grep + view |
| 6 | Answer included the three known gaps with an offer of follow-up (user then requested this report instead) | turn-1 reply |

## b) PARTIALLY DONE (project-level work surfaced by this session — none of it authored this session)

| # | Item | State |
|---|------|-------|
| 1 | go-finding adoption overall | Adapter shipped + tested (2026-09-19, issue #1, GAP-2), but library-only surface: no production code path calls `ToFindings`/`ToReport`. Consumers reach findings via art-dupl's SARIF output + go-finding's `FindingsFromSARIF`. |
| 2 | GroupID identity | One concept, three names: SARIF `go-finding/groupId`, JSON `clone_groups[].hash`, go-finding `Finding.GroupID`. Documented in AGENTS.md, pinned by tests; drift risk lives in docs, not code (status report's known-gap b.2). |
| 3 | Evaluation-doc linkage | The verdict doc lives only in the go-finding repo; art-dupl readers following its path hit a dead end (status item 24). Not fixed this session. |
| 4 | Adoption decision formalization | Carried in AGENTS.md convention entry + status report; no ADR, no TODO_LIST tracking verified. |

## c) NOT STARTED (untouched this session; known-open from the 2026-09-19 report)

1. **ADR for the go-finding adoption + GroupID=hash contract** (former item 23; ADR-0024 slot was taken by the JSON marshaler).
2. **Integration test piping art-dupl's REAL `--sarif` CLI bytes through `FindingsFromSARIF`** (former item 10) — the existing round-trip test exercises go-finding's own `ToSARIF`, so art-dupl's hand-rolled SARIF emitter is import-verified only by a one-off live CLI run, not a committed test.
3. **Local copy or summary of the evaluation doc** inside art-dupl (former item 24).
4. **CLI `--output finding` format** — not started **by decision** (rejected as YAGNI; SARIF is the wire format). Listed because "not started" ≠ "forgotten": it remains the most plausible full-adoption completion if product intent changes.
5. Everything else in the broader project backlog (out of scope per user instruction; not researched this session).

## d) TOTALLY FUCKED UP

**Nothing.** This session was read-only: no production code, no docs, no git state touched. No breakage possible, none found.

The honest worst-candidate is the self-critique item 1: I stated two unverified claims in turn 1. No damage resulted (both claims verified true in turn 2), but the discipline slipped. That is a process lapse, not wreckage — recording it here so the pattern is visible.

## e) WHAT WE SHOULD IMPROVE

1. **Verify at assertion time, not next turn.** Any claim sourced from a status report, AGENTS.md, or feedback doc gets a 10-second code grep before it is spoken. Status reports are external claims.
2. **Run the tests you cite.** "Pinned by tests" requires having executed them (or at minimum reading the test body — which I only did in turn 2).
3. **Read the source, not the report about the source**, when the source is 254 lines and one `view` away.
4. **Check TODO_LIST.md before declaring gaps untracked** — avoids re-reporting queued work.
5. **Avoid grep-at-the-end-of-chained-bash** — a no-match exit code reads as tool failure. Separate find-greps from verification-greps.
6. **Close with a ranked recommendation, not an open menu** — even when waiting for instructions, "I would do X first" costs one line.
7. **Skip QMD for sibling-repo paths** — collections here don't index `~/projects/go-finding`; go straight to `view`.

## f) Next Tasks (up to 50 — brainstorm, ROADMAP fuel; prioritized by impact)

**P0 — closes real verification/wiring gaps**
1. Committed integration test: run the actual CLI with `--sarif` on a synthetic duplicate, pipe the emitted bytes through `FindingsFromSARIF`, assert GroupIDs restore (closes the import-side gap; converts the 2026-09-19 one-off live proof into a permanent gate).
2. Decision + ADR: is the adapter's library-only surface final, or does a production path (SDK `Detector.Report()` / CLI flag) come? Record either way as an ADR (frees the stale "item 23").
3. Run `go test ./printer/finding/...` this session's successor — confirm all 17 green on current HEAD (never executed this session).
4. Copy or summarize the integration-evaluation doc into art-dupl docs; fix the dead path cited by issue #1.
5. Verify the three known gaps exist in TODO_LIST.md; add any that are missing (HARVEST from this report).

**P1 — reduces drift / duplication**
6. Three-names-one-concept: add a cross-reference table (SARIF property ↔ JSON field ↔ Finding field) to AGENTS.md or FEATURES.md.
7. Evaluate go-finding's `Template` builder API (`WithGroupID`) to replace hand-construction in `toFinding` (former item 19).
8. Consider go-finding `Tags` (`duplicate`, `type-1`/`type-2`) instead of/in addition to metadata keys (former item 47).
9. Wire `Report.Summary` (`FilesScanned`, `DurationMs`) if/when any CLI wiring exists — currently unused (former item 38).
10. Drop local `SARIFPropGroupID` if go-finding upstream exports the SARIF property key (former item 25).
11. Add `printer → finding-output → go-finding` edge to the architecture-understanding component diagram (former item 45).
12. Add `go-finding` to `.github/dependabot.yml` (former item 32; BuildFlow has a step).
13. Record the "latest stable, core-API-diff-checked" go-finding upgrade policy in AGENTS.md (former item 33).

**P2 — hygiene surfaced by this session**
14. Verify FEATURES.md reflects that the adapter is library-only (update if the P0-2 decision says so).
15. Check adapter test coverage of `NonActionablePattern`/`--explain` metadata keys (are they exercised in `finding_test.go`?).
16. Diagnose the LSP cross-repo lying (`golangci_lint_ls` reported "no required module provides package go-finding" while in go.mod; gopls likely outside devShell env) — former known issue, root cause undiagnosed.
17. Reconcile the concurrent session's TODO_LIST edits against this backlog (former item 22).
18. Watch-item: upstream decision D8 — revisit `RelatedRef.Metadata` when a SECOND consumer needs per-relationship metadata.
19. Watch-item: go-finding `GroupID` validation evolution (128-byte/machine-safe rules) — our 16-hex hashes are far inside; pin a test only if upstream tightens.

**P3 — ROADMAP fuel (deliberately not scheduled)**
20. Native `--output finding` CLI format (rejected as YAGNI 2026-09-19; revisit only with a named consumer).
21. SDK streaming findings surface (`Options` flag emitting `iter.Seq[finding.Finding]`).
22. HTML report perm-links carrying GroupID anchors (today: `groupAnchorID` from hash; could unify with finding GroupID).
23. Baseline/check subcommands emitting findings (they currently bypass the adapter entirely).
24. Evaluate `go-error-family` (transitive via go-finding) for a direct-use policy or stay transitive-only.
25. LSP diagnostic server fed by `ToLSP()` — evaluated and rejected in the evaluation doc; revisit only if IDE demand materializes.

(Stopped at 25 substantive items; padding to 50 would fabricate work. The status-report skill notes: larger N is a brainstorm, not a commitment list.)

## g) Questions I cannot figure out from the repo

1. **Product intent:** Is a native go-finding output surface (CLI flag or SDK method) planned, or is "adapter as public library API, SARIF as the wire" the final architecture? This single answer decides P0-2, item 20, and item 21.
2. **Downstream consumer:** Is another of your tools actually waiting to consume art-dupl findings right now? If yes, P0-1 + P0 wiring jump the queue; if no, the library-only ratification path is cheaper.
3. **Doc strategy for the dead-path evaluation doc:** copy it into art-dupl `docs/` (duplicates content, fork risk) or write a summary + link (rot risk if upstream moves it)? Both are defensible; I need your preference.

---

**Verdict:** Session goal (answer the go-finding question) fully achieved with evidence. Zero mutations. Two process lapses (late verification, unexecuted cited tests) — both self-caught in this report's preparation. Awaiting instructions.
