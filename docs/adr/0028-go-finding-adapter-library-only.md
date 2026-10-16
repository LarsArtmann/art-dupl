# ADR-0028: go-finding Adapter Stays Library-Only

**Date:** 2026-10-16
**Status:** Accepted (closes #3)

## Context

The go-finding adapter (`printer/finding/finding.go`, shipped with #1)
converts `domain.ProcessedCloneGroup`s into `gofinding.Finding`s. Source-level
survey (2026-09-22, re-verified 2026-10-16): it has **zero production callers**
— `printer/sarif.go` imports the package only for the `SARIFPropGroupID`
constant name. Consumers reach findings through art-dupl's SARIF output
(`--sarif`) plus go-finding's `FindingsFromSARIF`, pinned end-to-end by
`bdd/sarif_gofinding_integration_test.go` (real CLI bytes → identical group
partition as `--json`).

Two loose ends demanded a recorded decision:

1. Is library-only the FINAL architecture or an unfinished wiring?
   The integration evaluation (upstream: `LarsArtmann/go-finding`
   `docs/feedback/2026-06-05_art-dupl-integration-evaluation.md`, summarized
   at [`docs/feedback/2026-06-05_art-dupl-integration-evaluation.md`](../feedback/2026-06-05_art-dupl-integration-evaluation.md))
   already verdict-ed go-finding as a consumer-side output layer, never an
   internal representation — but nothing in THIS repo recorded ratifying it.
2. `Report.Summary` (`FilesScanned`, `DurationMs`) is unused either way.

A CLI `--output finding` format was rejected as YAGNI on 2026-09-19
(docs/status/2026-09-19 gap-2 report): machine consumers already share the
SARIF channel, and a second wire format would double the compatibility
surface for zero consumers.

## Decision

**Ratify library-only.** The adapter is a library surface for embedders
(in-process SDK/tool consumers constructing `Finding`s or `Report`s) and the
constant home of the GroupID interchange contract. It is NOT wired into the
CLI: there is no `--output finding` format, and none is planned.

**GroupID contract:** `GroupIDOf(group)` = the group's content hash
(16-char lowercase hex, deterministic for identical input). The same value
appears as the `go-finding/groupId` SARIF property on every result of the
group and as `clone_groups[].hash` in JSON output. This is the load-bearing
interchange guarantee — changing the hash algorithm is a wire break for every
consumer.

`Report.Summary` stays unfilled: no production caller exists to compute it,
and populating fields nobody reads is ghost surface (the same reasoning as
ADR-0025's classification-freedom).

## Rationale

1. **One wire format.** SARIF is the shared channel every consumer already
   parses (BuildFlow, GitHub code scanning, go-finding importers). A parallel
   native-format output would fork the golden-tested wire surface.
2. **The evaluation verdict holds.** Re-verified 2026-10-16: go-finding's
   value here is import/interop (`FindingsFromSARIF`, grouping, LSP
   conversion), all consumer-side. Internal representation would couple
   detection core to an external module's release cadence.
3. **Reversal is cheap.** Wiring a CLI format later is additive (a
   `createPrinter` case + a constructor); nothing in this decision removes
   capability.

## Consequences

- `printer/finding` may evolve its library API freely; only the three
  interchange guarantees pinned by tests are frozen: SARIF round-trip
  preserves GroupID, `Report.GroupFindings()` reconstructs the input
  partition, and `ToLSP`/`FromLSP` round-trips GroupID.
- Any future CLI wiring of the adapter requires superseding this ADR.
- FEATURES.md's adapter row points here; the upstream evaluation verdict is
  reachable from this repo via the summary page (issue #4's dead-path fix).
