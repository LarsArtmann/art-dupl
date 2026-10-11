# SARIF Snippet Share Measurement — relatedLocations bloat (M6)

**Date:** 2026-10-11
**Corpus:** `~/projects/go-cqrs-lite` @ `cd7b50c61`, all Go files (vendor excluded), default actionability on
**Command:** `go run ./cmd/art-dupl -t {1|default} --sarif -q ~/projects/go-cqrs-lite`

## Numbers

| Threshold  | SARIF bytes | Primary snippet bytes | Related snippet bytes | Related snippet share | Results | Related locs (all w/ snippet) |
| ---------- | ----------- | --------------------- | --------------------- | --------------------- | ------- | ----------------------------- |
| `-t 1`     | 13,883,532  | 136,012 (0.98%)       | **1,268,255**         | **9.13%**             | 1,840   | 17,904                        |
| default(5) | 70,911      | 8,269 (11.7%)         | **8,458**             | **11.9%**             | 5       | 8                             |

Every related location duplicated the sibling's full `Fragment` text: an N-instance
group shipped the same fragment N times (once primary + N-1 copies).

## Measurement method

```sh
jq '{primary_snippet_bytes: ([.runs[].results[].locations[].physicalLocation.region.snippet.text // empty] | map(length) | add // 0),
     related_snippet_bytes: ([.runs[].results[].relatedLocations[].physicalLocation.region.snippet.text // empty] | map(length) | add // 0),
     related_total: ([.runs[].results[].relatedLocations[]] | length),
     results: ([.runs[].results[]] | length)}' sarif.json
```

## Policy decision (G2)

**Drop sibling snippets — related regions carry location data only.** Reasons, in order:

1. **Upstream wire parity (M5):** go-finding's `sarifRelatedLocs` (`sarif_export.go:360-392`)
   emits NO snippet on related regions, and its import type (`sarifRelatedLoc`,
   `sarif_types.go:195`) has no snippet field — CLI-emitted sibling snippets are
   dropped by every go-finding consumer. Dead weight for the primary interchange path.
2. **Measured waste:** 9.1% of all SARIF bytes at `-t 1` (1.27 MB on one corpus),
   12% at the default threshold.
3. **Consumer UX unaffected:** GitHub code scanning renders the primary region's
   snippet; related locations render as links to the sibling location, which the
   URI + line range already identify.

Implemented 2026-10-11 in `printer/sarif.go` (`sarifSiblingLocation` — sibling
regions carry uri + line/column range + `clone-of` message, no snippet). The
primary result region keeps its snippet. Pinned by
`TestSARIFOutput_RelatedLocationsAndSnippet` (inverted assertion) and the
cross-check contract comment in `TestSARIFCrossCheckRelatedLocations`.
