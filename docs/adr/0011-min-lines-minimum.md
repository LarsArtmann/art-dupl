# ADR-0011: --min-lines filters on minimum clone LineCount across all clones

## Status

Accepted

## Date

2026-07-16

## Context

The `--min-lines` flag suppresses clone groups whose clones span fewer than N source lines. The initial implementation in `shouldSuppressGroup` compared only `group.Clones[0].LineCount()` against the threshold.

This was a bug: clones within the same group can have different line spans due to differences in surrounding context, nested function bodies, or whitespace. If the first clone happened to be the longest, shorter clones would bypass the filter entirely. A group with a 20-line clone and a 2-line clone would pass `--min-lines 10`, even though the 2-line clone is trivial noise.

Three options were considered:

1. **Minimum** (current): Suppress the group if ANY clone is below the threshold.
2. **Maximum**: Suppress only if ALL clones are below the threshold.
3. **Average**: Suppress if the mean LineCount is below the threshold.

## Decision

Use **minimum** across all clones in the group.

`minCloneLineCount(group)` walks all clones and returns the smallest `LineCount`. If that minimum is below `--min-lines N`, the entire group is suppressed.

## Rationale

- **Most conservative**: Suppresses the most noise. A group where any clone is trivially short likely represents low-value duplication.
- **Prevents false positives**: Users who set `--min-lines 5` expect to never see 2-line clones in the output. Minimum enforcement guarantees this.
- **Alternative rejected (maximum)**: Would allow a 2-line clone to appear in output as long as some other clone in the group is long enough. This defeats the purpose of the filter.
- **Alternative rejected (average)**: Hides the distribution. A group with clones of 1 and 20 lines averages 10.5, which would pass `--min-lines 10` despite containing a 1-line clone.

## Consequences

- A group is suppressed more aggressively than before the fix.
- Users who relied on the old (buggy) behavior may see fewer groups reported.
- The `SuppressionConfig` struct now bundles `MinLines` with `SuppressTestLow` and `TestThreshold`, making the filtering parameters explicit and type-safe.
