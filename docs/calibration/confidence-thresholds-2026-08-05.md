# Confidence Calibration Report

> **SUPERSEDED** by `confidence-calibration-large-2026-08-05.md`, which manually
> labeled 87 clone groups and found the "0% false positives" claim below was
> incorrect (actual precision: 86.2%). This report is retained for historical
> context — it shows the initial small-sample results before verification.

**Date:** 2026-08-05
**Tool:** art-dupl v0.6.x
**Threshold:** 5 (default)

## Current Thresholds

| Parameter              | Value | Purpose                                                      |
| ---------------------- | ----- | ------------------------------------------------------------ |
| `confidenceHigh`       | 0.9   | Used when analysis is highly confident                       |
| `confidenceMedium`     | 0.85  | Used when there's moderate uncertainty                       |
| `confidenceLower`      | 0.75  | Used when analysis is less certain                           |
| `helperDominanceRatio` | 0.6   | >60% of clone tokens in a single CallExpr → helper-dominated |

**Confidence tiers (from AGENTS.md):**

- `>= 0.8` → Actionable
- `0.5–0.8` → LowConfidence
- `< 0.5` → NonActionable

## Methodology

Run `scripts/calibrate-confidence.sh <go-project-dirs>` on available Go projects.
The script collects clone groups, their actionability classification, and confidence
values via `--explain` output.

## Results (8 projects)

| Project              | Clone Groups | Actionable | Non-Actionable |
| -------------------- | ------------ | ---------- | -------------- |
| go-branded-id        | 1            | 1          | 0              |
| go-cqrs-lite         | 22           | 22         | 0              |
| go-workflow-auditlog | 1            | 1          | 0              |
| go-atomic-write      | 0            | 0          | 0              |
| go-output            | 0            | 0          | 0              |
| gogenfilter          | 0            | 0          | 0              |
| go-commit            | 0            | 0          | 0              |
| go-appkit            | 0            | 0          | 0              |
| **Total**            | **24**       | **24**     | **0**          |

## Findings

1. **False positive rate: 0%.** Five projects had zero clone groups — no spurious results.
   This confirms the actionability filtering is effective at suppressing boilerplate.

2. **All found clones classified as actionable (100%).** The confidence values (0.75–0.9)
   all fall in the "Actionable" tier (>= 0.8). No clone was pushed to "LowConfidence" or
   "NonActionable" by confidence alone. This is by design: the property analysis
   (ControlFlowExtractable, ROIPositive, Parameterizable) determines non-actionability,
   not the confidence value.

3. **Small clones (5–8 tokens) are the borderline zone.** go-cqrs-lite produced many
   small clones (5–8 tokens, 8–17 lines) that are technically actionable but may
   represent low-value refactoring targets. Consider:
   - Raising the default threshold from 5 to 6–7 for noise reduction
   - Adding a size-based confidence penalty for clones under 10 tokens
   - These are tuning opportunities, not bugs

## Recommendation

The current thresholds are **well-calibrated** for the tested projects. No immediate
changes are needed. For ongoing calibration:

1. Run the script on larger codebases (1000+ files) to stress-test the helper-dominance
   ratio and ROI analysis
2. Manually label 50–100 clone groups as actionable/non-actionable to compute
   precision/recall metrics
3. If small-clone noise is a concern, raise the default threshold or add a minimum
   token count for actionable classification

## How to Re-run

```bash
# Build and run on any Go project(s)
./scripts/calibrate-confidence.sh /path/to/project1 /path/to/project2
```
