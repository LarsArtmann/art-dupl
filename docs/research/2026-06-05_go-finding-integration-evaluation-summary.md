# go-finding Integration Evaluation — Summary

**Upstream canonical doc:** `go-finding/docs/feedback/2026-06-05_art-dupl-integration-evaluation.md`
(local checkout: `~/projects/go-finding/docs/feedback/2026-06-05_art-dupl-integration-evaluation.md`, 375 lines)

This file exists because the upstream path above was cited by art-dupl issue #1 and
status reports without a live in-repo anchor (dead-path finding, reported three times:
2026-09-22_22-40, 2026-09-22_23-01, 2026-09-23_01-34). The upstream doc stays the
source of truth; this summary is the in-repo entry point.

## Verdict (2026-06-05, re-confirmed by implementation)

**Adopt go-finding as art-dupl's output adapter — a consumer-side output layer,
NEVER an internal representation.**

The rejected alternative (making go-finding's `Finding` the internal clone type)
fails on a domain mismatch: a `Finding` is a single issue at a single position,
while an art-dupl clone group is an N-location relationship. Flattening groups
into per-position findings internally would lose the group identity that
deduplication, baseline matching, and HTML rendering all rely on.

## Gap inventory status (upstream table, updated 2026-09-07)

All implementable gaps shipped; GAP-3 (per-relationship metadata) deliberately
deferred — `Finding.Metadata` is the escape hatch. The integration now runs both
ways: art-dupl's `printer/finding` adapter converts `domain.ProcessedCloneGroup`s
into findings, and go-finding's dedup passes consume art-dupl reports directly.

| Gap                                   | Status                   | art-dupl touchpoint                                                                                                            |
| ------------------------------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| GAP-1 `RelatedRef.Range`              | Implemented              | LSP round-trip                                                                                                                 |
| GAP-2 `GroupID`                       | Implemented (2026-09-07) | `finding.GroupIDOf` = clone-group content hash; SARIF property `go-finding/groupId`; JSON `clone_groups[].hash` is the same id |
| GAP-3 per-relationship metadata       | Deferred                 | `Finding.Metadata` under the `art-dupl/` namespace (`MetadataKey*` constants)                                                  |
| GAP-4 `LSPDiagnosticTag`              | Implemented              | `ToLSP`/`FromLSP` tag preservation                                                                                             |
| GAP-5 snippet in SARIF                | Implemented              | `printer/finding.ToSARIF`                                                                                                      |
| GAP-6 `ToLSP` uses `RelatedRef.Range` | Implemented              | real spans for related info                                                                                                    |
| GAP-7 strict `Category.IsValid()`     | Resolved differently     | `IsStandard()` vs `IsValid()` split, no global registry                                                                        |
| GAP-8 `FromLSP` preserves tags        | Implemented              | `Metadata[LSPDiagnosticTagsKey]`                                                                                               |
| GAP-9 `iter.Seq` on `Report.All()`    | Implemented              | `report_query.go`                                                                                                              |

## What art-dupl ships on top (see `printer/finding/` + ADR-0025)

- `printer/finding.ToFindings` converts `domain.ProcessedCloneGroup`s; classification
  metadata travels in `Metadata` under the `art-dupl/` namespace, and the SDK provider
  (`pkg/provider`, BuildFlow core lane) is classification-free per ADR-0025.
- Severity mapping mirrors the SARIF level ladder from `Options.Threshold`
  (0 = `config.DefaultThreshold`); the provider caps advisory severity at warning.
- Three interchange guarantees are pinned by tests: SARIF round-trip preserves
  `GroupID`, `Report.GroupFindings()` reconstructs the exact input partition,
  and `ToLSP`/`FromLSP` round-trips `GroupID`.

## Follow-ups

- Issue #2 (committed SARIF integration test feeding real CLI bytes through
  `FindingsFromSARIF`) — tracked in `TODO_LIST.md`.
- GAP-2 GroupID consumer-intent issue in go-finding — was unfiled at the
  2026-09-23 docs-health pass; check upstream before filing a duplicate.
