# Status Report + Brutal Self-Review (Round 2): Storage Corrections + C Re-Evaluation

**Date:** 2026-11-24 19:01 (Tuesday)
**Session span covered:** ~11:30-19:01 continuation (this report supersedes the 11:27 report's
recommendation sections; that report stands as a point-in-time snapshot and is not rewritten)
**Scope line:** Covers ONLY this session's continued work: Option C re-evaluation (owner design
input), the three-part correction round (BuildFlow counts-only cause / go-cqrs-lite storage /
no-artifacts-in-repos directive), and this self-review. VERIFIED at report time: the proposal
artifact passes all mechanical checks (parser clean, 22 anchors resolve, matrix 9x5 consistent);
boundary fast gate GREEN after every burst; the NEW cites this round were verified FIRSTHAND:
go-finding `report_query.go:100-125` (Report.Findings, GroupFindingsSorted), BuildFlow
`migrations.go` v13 + `pipeline_phases_2.go:226-233`, daemon `server.go:211-223` +
`background_refresh.go:240-250`, art-dupl `--incremental` (cmd/flags.go:78,
job/incremental.go:207). NOT VERIFIED: the older overview/go-compliance/most-BuildFlow cites
remain explorer-sourced (debt carried from round 1); no code was written this session
(docs-only, so no full-suite run — fast gate only); go-cqrs-lite was assessed from its SKILL.md
and skill references, NOT from reading its source.
**Pinned commit:** `7607a9e2`

---

## a) FULLY DONE

1. **Option C re-evaluated** after owner design input (incremental review + async enrichment):
   daemon refresh machinery verified firsthand (bounded `refreshSem`, per-path single-flight
   `refreshInFlight`, stale-while-revalidate, enricher-config-preserving background refresh,
   RPC-triggered async goroutine). C upgraded from "later runner upgrade" to credible primary
   runner via two named design choices: exec-not-embed (kills private-dep pinning objection) and
   C-produces-A's-datapoints (daemon = trigger, store = store). Honest caveat kept: parse is
   incremental, detection is not (suffix tree rebuilt per run, O(repo) with small constant).
2. **BuildFlow counts-only cause established firsthand**: toolsdk is NOT the bottleneck —
   go-finding delivers full fidelity in-memory (Report.Findings; GroupFindingsSorted() built for
   serialization); BuildFlow's own step_outputs schema drops it by size design (counts + capped
   rule_histogram; migration v13's own description says counts "cannot answer which rules").
   Written into E's costs cell + appendix sources row with the two fix paths (digest column vs
   Phase 1 store).
3. **Storage story redesigned per owner directive**: no measurement artifacts committed into
   measured repos (recorded as DECIDED in the proposal); central trends store via go-cqrs-lite
   `system.New` + `SnapshotRecorded` events + `metaengine.Query[Q,R]` read models (sqlite
   default, pebble optional) + `ServeSSE`; NDJSON demoted to producer-to-store interchange, never
   the store. Fleet I/O rule (all durable data I/O via metaengine) now cited as the substrate
   rationale.
4. **Option D reframed**: "git-native" storage premise removed (rejected by directive);
   D is now "static report, zero server" — a renderer over A's store or E's log, with the
   read-only baseline-history backfill exception (reading existing committed
   `.art-dupl-baseline.json` revisions adds nothing to repos).
5. **Proposal consistency sweep**: all JSONL/committed/XDG/git-native mentions replaced or
   converted to directive statements; TOC, matrix (History-store row, D header), roadmap Phases
   1-3, gap box, and Q1 (now "Store placement + engine") all updated; two fresh inconsistencies
   found during THIS self-review and fixed on sight (scorecard "0 new services" scoped to Phase 0;
   matrix A "New moving parts" now includes the store host).
6. **go-cqrs-lite skill loaded** before designing against system/metaengine (skill-rule compliance).
7. **Foreign-edit conflict handled correctly**: TODO_LIST changed under me mid-edit (another
   session struck the /tmp-binaries row OBSOLETE) — re-read, preserved their edit, re-applied mine.
8. **TODO_LIST row updated** to amendment x3 state with the DECIDED constraint and rewritten Q1.

## b) PARTIALLY DONE

1. **First-hand verification debt** (carried from round 1): overview, go-compliance, and most
   BuildFlow cites are still explorer-sourced. This round ADDED five firsthand verifications but
   did not clear the old debt.
2. **The appendix "verified against source" wording**: improved (methodology paragraph now names
   what was verified firsthand per amendment), but still not per-row labeled.
3. **Q1 (store placement)**: framed with a recommendation (tiny dedicated host) but genuinely
   open — and my recommendation adds a service to the fleet, which is owner-taste territory.
4. **Proposal estimate honesty**: Phase 1 "~2 days" now includes a system.New composition root,
   one event type, read models, a host, and systemscenario tests — optimistic. Flagged here, not
   yet softened in the doc.

## c) NOT STARTED

1. All implementation (Phase 0 reader, Phase 1 snapshot + store, Phase 2 panel) — unchanged,
   gated on the owner's architecture decision.
2. Q1-Q4 decisions.
3. The proposal index/discoverability item (carried from round 1, still open).
4. DOMAIN_LANGUAGE entries for Snapshot/trends concepts (still correctly deferred).

## d) TOTALLY FUCKED UP

1. **I recommended committing measurement artifacts into the measured repos.** The owner had to
   veto it. This is not a preference miss; it violates principles I already hold: the observer
   must not mutate the observed, and art-dupl's own corpus discipline (pinned worktrees at
   `~/projects/.corpus-pins/` exist precisely to isolate measurement from repos). Worse: the
   same proposal's Phase 0/E celebrated reading BuildFlow's log BECAUSE it required no repo
   changes — and one section later I proposed writing files into every repo. Self-contradictory
   within one document.
2. **JSONL-as-store contradicted the fleet I/O rule** (owner directive 2026-10-24: all durable
   data I/O through metaengine). The rule lives in the go-cqrs-lite skill I did not load before
   making a storage recommendation — the moment Option A needed a history STORE, the stack-policy
   question should have fired. I reached for "dumb simple files" without consulting the fleet's
   storage substrate. Both d1 and d2 have the same root: I designed storage from local
   simplicity instead of from fleet policy.
3. **The E section implied a fidelity gap without naming the cause** ("full findings exist only
   inside a live run") — the user's "is toolsdk not good enough?" question exposed that I never
   asked WHERE the drop happens. One grep (report_query.go) answered it. Sloppy: I characterized
   a limitation without locating it.
4. **Fresh inconsistencies shipped in the storage round**: matrix A still said "One subcommand +
   one panel" while A gained a store host; scorecard still said "0 new services in recommended
   path" while the recommended path's Phase 1 adds a store host. Caught only in THIS self-review,
   not in the round's own verification pass — my per-round check list verifies structure
   (anchors/cells) but not cross-section semantic consistency.
5. Carried from round 1 (still true, still mine): the AND/OR and BuildFlow blind-spot misses that
   forced three owner correction rounds total.

## e) WHAT WE SHOULD IMPROVE

1. **Storage decisions go through fleet policy, not local simplicity.** Checklist addition: any
   proposal involving durable state loads the go-cqrs-lite skill (I/O rule) and the stack policy
   BEFORE picking a substrate. The metaengine answer was one skill-read away.
2. **Observer-effect rule for measurement tools**: a measuring tool never writes into the
   measured subject. Write it into art-dupl's conventions when Phase 1 starts (AGENTS.md /
   HOW_TO_USE) so the next session does not re-derive it.
3. **Locate limitations before characterizing them**: "X is not possible" claims must name the
   exact layer where the drop happens (contract vs consumer vs schema). One grep each.
4. **Per-round semantic consistency sweep**: after each amendment round, grep the artifact for
   terms tied to the OLD recommendation (this round: "JSONL", "committed", "one panel") — the
   final sweep caught leftovers; doing it per-round would have caught the matrix/scorecard cells
   before the user's self-review request did.
5. **Soften estimates when scope grows in review**: "~1 day" became "~2 days + a store host"
   silently across amendments; estimates should be revisited explicitly.

## f) Up to 50 things to get done next

**Owner-gated (blocking everything below):**

1. Architecture decision: E-now-then-A (as recommended) or another mix
2. Q1: store placement — dedicated tiny host (recommended) vs overview vs discovery daemon; engine sqlite vs pebble
3. Q2: go-compliance hub endgame yes/no (flip trigger #3)
4. Q3: analysis depth for trends (proposal: semantic, never type-aware)
5. Q4: BuildFlow state-DB retention (pre-existing item, load-bearing for E)

**Verification debts (cheap, before trusting cross-repo rows):**
6. Verify overview cites firsthand (enrichment_engine.go:44-81, compliance_source.go:18-44, hotspot_source.go:37,148-184)
7. Verify go-compliance cites firsthand (cache.go:34-46, queries.go:60, fold.go:113, reporter.go:62, v1.go:19-32)
~~8. Verify remaining BuildFlow cites (history_cmd.go, compliance_1.go:83-97, timings_setup.go:20)~~ decided 2026-11-24 — verification debts harvested into the open TODO_LIST row "Verify the duplication-trends proposal's cross-repo cites" (with the 11-27 session's debts 6-15).
9. Query the real DB once: step_outputs shape for step='art-dupl', one repo, 90-day row count
10. Cached-run behavior: does a cache hit re-record the count?
11. Label appendix rows firsthand-vs-explorer (one-line amendment)
12. Soften the Phase 1 "~2 days" estimate to honest range in the doc

**Phase 0: BuildFlow reader (if approved):**
13. Probe modules/dbstore import (module path/version) vs raw SQL
14. Prototype one repo's 90-day count curve
15. Dogfood art-dupl's own repo first
16. Choose host: overview panel vs static page vs `buildflow history` extension
17. If overview: dupl_source.go per hotspot recipe + engine union + route/nav/templ
18. Policy segmentation labels at count level (cacheVersion breaks when known)
19. Sparse-cadence rendering (gaps, not zero-fills)
20. Cross-machine reachability check for the DB (ties to Q-topology from round 1, still unanswered)

**Phase 1: snapshot + central store (if approved):**
21. Domain types first: Snapshot event, TrendsMetrics, with invariants written down
22. NDJSON wire spec (stdout + file) — interchange only
23. system.New composition root: DomainConfig (SnapshotRecorded) + DeploymentConfig (sqlite)
24. metaengine Query[Q,R] read models: per-project series, cohort sets, fleet rollup
25. Decide event vs command ingest surface (ingest command with idempotency middleware?)
26. systemscenario BDD: Given snapshots When query Then series (frozen ManualClock)
~~27. metaengine.ServeSSE dashlet for live latest-values~~ decided 2026-11-24 — proposal Phase-2+ surface, rides the open decision row "Decide the duplication-trends dashboard architecture" (TODO_LIST); no standalone dispatch.
28. Engine A/B: sqlite vs pebble via cqrs-bench if write volume matters
29. `art-dupl snapshot` subcommand (reuses stats pipeline; emits NDJSON)
30. AnalyzerIdentity + git head captured per datapoint; policy-break marker semantics
31. Cohort diff engine (new/resolved/surviving) as a read model, not CLI logic
32. Retention/pruning policy for the store (does 90d apply? owner taste)
33. Module layout decision: store inside art-dupl repo (single go.mod gains system+metaengine deps) vs separate module — owner question (see g3)
34. Mixed-fidelity decision: do BuildFlow count-datapoints ingest into the same store? (see g2)
35. HOW_TO_USE + AGENTS.md conventions: observer-effect rule, snapshot flags
36. Baseline-history backfill importer (read-only, from committed .art-dupl-baseline.json revisions)
37. Threshold/mode standard recorded per datapoint (self-describing curves)

**Phase 2: overview panel (if approved):**
38. Trends-store query client (overview stays database-free — queries only)
~~39. Time-axis/sparkline component (first in repo)~~ decided 2026-11-24 — proposal Phase-2 rendering concern, rides the open decision row "Decide the duplication-trends dashboard architecture" (TODO_LIST); no standalone dispatch.
40. Fleet heatmap + per-project rows + cohort-churn table
41. Family-age metric (oldest surviving GroupID)
42. TTL/SSE refresh wiring via enrichmentEngine[T]

**Later / upstream:**
43. Phase 3 daemon trigger: async analyzer lane (timeout, min-interval, budget) exec'ing snapshot
44. BuildFlow digest-column upstream option (v13 pattern, ~1KB group-id list) if E ceiling matters
45. `buildflow history` exposing finding counts in RunSummary (upstream)
46. Phase 4 hub graduation watch (tag + series API)
47. JSONL→compliance.Record adapter only if Phase 4 happens

**Docs/hygiene:**
48. Index the proposal (research README row or proposals/README)
49. Screenshot when a browser exists (carried debt)
50. Next docs-health cadence (2026-12-14): harvest BOTH round reports + the proposal row with routing rigor (most f-items stay contingent on the architecture decision)

## g) Questions I cannot answer myself

1. **Store host placement (Q1, now THE blocking design decision):** dedicated tiny service
   (my recommendation — keeps art-dupl SDK dep-clean, overview database-free, daemon
   single-mission) vs overview hosting it vs the discovery daemon vs something BuildFlow-adjacent?
   Each adds supervision somewhere; only you know where you want another long-lived process.
2. **Mixed fidelity in one store:** should BuildFlow's count-only datapoints (Phase 0/E) be
   ingested into the same trends store as full-fidelity snapshots (mixed-fidelity series,
   simpler pipeline), or does the store accept only full-fidelity datapoints with E remaining a
   separate read-only view? Design taste with real curve-semantics consequences.
3. **Dependency weight in art-dupl's module:** Phase 1 as designed pulls `system` + `metaengine`
   into art-dupl's single go.mod (today a lean CLI/SDK). Accept that, or split the trends store
   into its own module (or repo) so the detector stays dependency-minimal? This is fleet
   architecture taste, not something I can derive.

---

_Report per status-report skill (user format wins). Round 2 of the session; the 11:27 report
stands as history. Not committed manually; the auto-commit daemon picks this up. Section (f)
items remain contingent on the owner's architecture decision; HARVEST routing rigor applies at
the 2026-12-14 cadence._
