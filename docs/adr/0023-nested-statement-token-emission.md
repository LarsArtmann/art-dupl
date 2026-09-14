# ADR-0023: Nested-Statement Token Emission

- **Status:** Accepted (2026-09-14)
- **Deciders:** Lars Artmann
- **Supersedes:** the statement-granularity scope of the T1 tokenization report
  (`docs/status/2026-06-20_22-10_t1-statement-tokenization-in-progress.md`)
- **Motivated by:** `docs/status/2026-09-13_15-45_gopaperless-false-negative-investigation.md`

## Context

Statement-level tokenization (T1) fingerprints every statement subtree into
ONE composite token. Two composite statements (for/if/switch/select/else-if)
that share leading statements but diverge deeper inside produce different
composite fingerprints and therefore ZERO matching tokens: the clone is
invisible at every threshold, every detection mode, and every filter
combination. A 26-line duplicated pagination skeleton in go-paperless
(September 2026 investigation) demonstrated the class on real code; the tool
reported "0 clones, Health A" for that repository.

Synthetic fixture proof (report experiments A-F, re-verified and extended on
2026-09-14): the masking applies to `for`/`range` loops, `if` bodies,
`switch`/`select` case bodies, `else if` chain links, and arbitrarily nested
combinations. It is the most common real-world clone shape ("shared loop
skeleton, divergent accumulator"), so a detection tool without a fix here
silently issues false-clean verdicts.

## Decision

1. **Composite statements additionally emit their nested statements as
   individual tokens.** For a statement atom S, `serial()` emits S's composite
   fingerprint token first (unchanged), then descends:
   - direct statement-atom children of S (e.g. `else if` chain links, which
     the Go transformer flags as statement atoms), except GenDecl children;
   - statement children of unflagged `BlockStmt` container children
     (function/for/if/switch/select bodies, else blocks), each recursing
     through the same rule.

2. **GenDecl specs stay composite-only.** A single-spec `var x T = v` would
   contribute two tokens (GenDecl composite + ValueSpec) for ONE source
   statement — pure double-counting with no recall gain. Decl sharing
   recall (partially-shared `var (...)` blocks) is deliberately deferred to a
   future spec-level ADR.

3. **Subsumed units are trimmed at the `FindSyntaxUnits` boundary.** A maximal
   match may contain a composite token AND tokens from inside its subtree
   (identical guard clones: `[if][return]`). Nested tokens exist for matching
   recall, not as extra source statements, so units whose `[Pos, End)` range
   lies strictly inside another unit's range are dropped before threshold
   gating, hashing, size metrics, and actionability classification. This
   keeps group shapes, sizes, hashes, and threshold semantics IDENTICAL to
   the pre-change behavior whenever the matched statements are fully
   identical, and changes them only for the newly-recallable divergent
   interiors (where no composite is part of the match).

4. **Default-on, not flag-gated.** The blind spot is a correctness bug in the
   tool's core promise; a flag would keep the default behavior wrong.
   Precedent for default-changing releases exists (ADR-0009 changed the
   default threshold). Consumers who depend on old counts must pin a version;
   the CHANGELOG carries a migration note and CacheVersion was bumped (v4)
   so stale cached token streams cannot silently reproduce old results.

5. **All block-structured statements at once (loops, if, switch, select,
   else-if), not loops-only.** The fixture proofs show the identical
   mechanism everywhere; a loops-only fix would leave the most common Go
   nesting (`if` inside `for`) broken and force a second behavioral release.

## Alternatives Considered

- **(B) Secondary sub-statement pass over large composite statements** —
  rejected: a second pass duplicates the matching pipeline, needs its own
  overlap elimination, and still misses small composite statements.
- **(C) Document as a known limitation** — rejected: "0 shown ≠ clean" on the
  most common clone shape is not a limitation, it is a defect.
- **Flag-gated (`--nested-statements`)** — rejected per Decision 4; the
  corpus evidence (go-cqrs-lite cross-engine clones, go-sse parser clones)
  showed additions are real recall, not noise requiring an escape hatch.
- **Loops-only first PR** — rejected per Decision 5; the marginal risk of
  covering all block statements is the same mechanism exercised more often,
  and is bounded by actionability suppression plus the default threshold.

## Consequences

**Recall (the point of the change)**

- go-paperless: the find-by-name family (`FindCustomField`/`FindStoragePath`
  query+decode skeletons) now surfaces; the pagination pair from the report
  was already extracted at source (`fetchAllPages[T]`) before the fix landed.
- go-cqrs-lite at `-t 2` (default actionability): 206 → 380 shown groups —
  spot-checked additions are genuine cross-engine clones (30+ line
  aggregation/parity/explain blocks inside if/for bodies).
- go-sse at `-t 1`: 1 → 16 shown groups — 5 genuine Go parser/statement
  clones plus templ HTML sibling structure (see below).
- art-dupl self-analysis and go-paperless at the default `-t 5`: shown
  counts unchanged (0 and 1 respectively). The default-threshold UX is
  stable; growth appears at diagnostic thresholds.

**Costs and mitigations**

- Token streams grow (~10-15% on real Go code), suffix trees grow
  accordingly. The alloc-budget gate stays green: serialization still
  allocates exactly 2 (arena + stream); all suffixtree budgets within
  tolerance. Benchmark baseline re-captured in
  `docs/benchmarks/baseline-2026-09-14-nested-tokens.txt`.
- Fully-identical composite matches are trimmed back to the composite unit,
  so sizes/hashes/baselines for previously-detectable clones are unchanged.
- templ files get nested-element emission for free (same mechanism, same
  masking existed). At `-t 1` this surfaces HTML sibling repeats in
  templ-heavy repos (go-sse examples). Accepted: structural HTML repeats are
  real, templ matching is structural by design, and the default threshold
  filters them. If they become a support topic, the fix is an
  html-sibling-boilerplate actionability pattern, not a revert.
- Fully-identical loops now report their interior statements too when the
  surrounding run extends the repeat — no user-visible change after
  subsumption trimming (same span, same unit count).
- Corpus gates in AGENTS.md updated (go-sse `-t 1`: 1 → 16 shown;
  go-cqrs-lite `-t 2`: 206 → 380 shown).

**Guard rails added**

- `syntax/nested_statement_serial_test.go`: emission structure for loops,
  else-if, switch/case, GenDecl exclusion, arena-count mirror invariant, the
  divergent-tail detection regression, and the guard-clone trim regression.
- `syntax/golang/nodetypes_pin_test.go`: pins the node-type literals the
  syntax package relies on (`BlockStmt`=6, `GenDecl`=24).
- `bdd/nested_block_clone_test.go`: end-to-end DescribeTable over the five
  masking classes plus default-threshold and guard-suppression behavior.
- `BenchmarkMemoryUsage` now uses a fixed RNG seed: its time-seeded input
  made allocs/op vary by ~±25 (4780-4827 measured on unchanged code), which
  previously let the ±1 gate fail spuriously.

## Verification

- Full test suite green (`go test ./...`), `-race` on syntax/printer/job/cache.
- `scripts/check-alloc-regression.sh` green.
- Synthetic fixture (`/tmp/fntest`, 5 masking classes): baseline detected
  only trivial top-level statements; fixed binary detects all five classes
  at `-t 2` and default filtering.
- Real-world: go-paperless, go-cqrs-lite, go-sse, art-dupl self-analysis
  (numbers above).
