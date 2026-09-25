# ADR-0025: SDK Findings Stay Classification-Free

**Date:** 2026-09-25
**Status:** Accepted

## Context

The toolsdk provider (`pkg/provider`) emits go-finding `Finding`s through the
shared `printer/finding` adapter, so GroupID, severity, positions, snippets,
and `clone-of` links are identical to the CLI's finding output. The
classification metadata keys (`art-dupl/clone-type`, `-category`,
`-priority`, `-actionability`, `-non-actionable-pattern`, `-generics-*`) are
deliberately ABSENT: the SDK pipeline never runs actionability evaluation,
clone-type classification, or the parameterizability engine, and
`stripEmptyMetadata` drops the zero-valued keys.

The open question: should the SDK compute classification (CLI-parity
findings), or stay classification-free with the absence documented?

## Decision

**Path DOCUMENT — the SDK stays classification-free.** The provider package
doc and `stripEmptyMetadata`'s full-key-set test (B4.3) are the durable
documentation.

## Rationale

1. **The logic lives in the output layer by design.** Actionability
   evaluation (`printer/actionability`) walks AST-pattern denylists over
   clone groups; clone-type classification walks subtree identifier names
   (`printer/clone_processor.go`). Both are presentation heuristics that
   answer "should a human look at this group?" — they are not detection
   facts. Moving them under `pkg/artdupl` would invert the architecture the
   `.go-arch-lint.yml` boundaries encode (sdk may not depend on printer).
2. **No consumer needs them yet.** BuildFlow renders findings generically
   (rule, message, positions, snippet, severity) and consumes none of the
   `art-dupl/*` keys. Ghost metadata nobody renders is the gotcha-#147
   disease in miniature.
3. **Severity safety is already solved at the boundary.** The provider caps
   advisory severities at warning (`maxAdvisorySeverity`, the v0.7.2
   `capAdvisorySeverities` change), so detector-only findings can never fail
   BuildFlow's default error gate with or without an actionability label.
4. **Reversal is cheap.** If a consumer later wants classification, the
   change is additive (populate Classification in the SDK pipeline, or run
   the classifier as an optional post-pass) — none of today's contracts
   would break.

## Consequences

- SDK consumers get interchange facts (where, what, how grouped) without
  presentation judgments (is it boilerplate?).
- The CLI remains the only source of actionability-filtered output;
  `--show-suppressed` parity in SDK consumers would require the
  classification pass (tracked as a ROADMAP candidate, not planned work).
- `TestStripEmptyMetadata_DropsFullClassificationKeySet` pins the exact key
  set so a new classification key cannot silently leak into SDK findings.
