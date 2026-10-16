# go-finding Integration Evaluation — Summary + Pointer

> **This is a summary, not a copy.** The canonical evaluation lives upstream:
> [`LarsArtmann/go-finding/docs/feedback/2026-06-05_art-dupl-integration-evaluation.md`](https://github.com/LarsArtmann/go-finding/blob/main/docs/feedback/2026-06-05_art-dupl-integration-evaluation.md)
> (verified present 2026-10-16). It is kept there because go-finding owns the
> interchange type under evaluation. This page exists so a reader of THIS repo
> can reach the verdict without cloning go-finding (issue #4's dead-path fix).

## The verdict

**Adopt go-finding as art-dupl's output adapter — consumer-side only.**
go-finding is the interchange/output layer art-dupl emits TO (via SARIF
import, `FindingsFromSARIF`); it is never an internal representation of
clones inside art-dupl's detection pipeline.

The evaluation found `Finding` covered ~80% of art-dupl's needs; the
remaining 20% were nine numbered gaps (GAP-1..9). As of the 2026-09-07
status table: seven implemented (related-range spans, `GroupID` +
`Report.GroupFindings()` + the `go-finding/groupId` SARIF property, LSP
diagnostic tags, SARIF snippets, related-info ranges, tag round-trips,
`iter.Seq` on `Report.All()`), one resolved differently
(`Category.IsStandard()` vs `IsValid()` split), one deferred as recommended
(per-relationship metadata via `Finding.Metadata`).

## Where the adoption landed in this repo

- `printer/finding/` — the adapter library (`ToFindings`, `ToReport`,
  `GroupIDOf`). Surface decision: **library-only**, no CLI format —
  see [ADR-0028](../adr/0028-go-finding-adapter-library-only.md).
- `printer/sarif.go` — the production wire emitter; every result of a clone
  group carries `go-finding/groupId` = the group's content hash.
- `bdd/sarif_gofinding_integration_test.go` — pins the contract end-to-end:
  real CLI `--sarif` bytes through `FindingsFromSARIF` reconstruct the exact
  group partition (identical to the `--json` `hash` set).
- `pkg/provider/` — the BuildFlow toolsdk detector (LIVE since 2026-09-25).

The evaluation itself was declared fully superseded by its own 2026-09-23
resolution banner (all implementable gaps shipped); follow-up integration
work was tracked as art-dupl issues #2/#3/#4 — all closed by the 2026-10-16
session that wrote this page.
