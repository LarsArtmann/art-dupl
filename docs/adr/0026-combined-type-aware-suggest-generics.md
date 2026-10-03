# ADR-0026: Combined type-aware + suggest-generics mode

## Status

Accepted (2026-10-03)

Supersedes the mutual-exclusion resolution in ADR-0021 ("suggest-generics
wins, the CLI prints a warning").

## Context

ADR-0021 made `--type-aware` and `--suggest-generics` duals over the same
hash: type-aware encodes local variable types into identifier hashes
(suppressing same-shape-different-type matches), suggest-generics erases them
(surfacing those matches as generics-extraction candidates). When both flags
were passed, suggest-generics silently won and the CLI printed a "drop
--type-aware" warning.

That resolution was a design cop-out. The two flags answer different questions
about the same codebase — "what is EXACTLY duplicated (same types)?" and
"what is duplicated modulo types (generics opportunities)?" — and a user
asking both questions had to run the tool twice, paying the go/packages
type-check (10-100x parse cost) twice.

Key observation: the expensive part is the type-check, and `EraseHash` is only
a per-`PreloadedAST` bool stamped AFTER `packages.Load`
(`syntax/golang/typeinfo.go`). One load can feed both hash dispositions via
`TypeAwareData.WithEraseHash`, a shallow copy sharing the ASTs and
`types.Info`.

## Decision

When both `--type-aware` and `--suggest-generics` are set, run **two detection
passes over one shared type-check** and merge the match streams:

1. **Type-aware pass** (`EraseHash=false`): types in the hash. Produces exact
   type-matched clone groups — identical to `--type-aware` alone.
2. **Type-erased pass** (`EraseHash=true`): types on `Node.VarType` only.
   Produces structural families across different concrete types.

### Merge rule (structural, single gate)

A type-erased family flows into the output **iff** it has at least
`syntax.MinDivergentPositions` (2) distinct divergent type positions
(`syntax.IsGenericsCandidateStructure`):

- **Zero divergence**: the family's instances have identical types at every
  position, so the type-aware pass already reports the same instances (the
  hash dispositions only change token VALUES, not cross-instance equality).
  Keeping it would duplicate every same-type group. Dropped.
- **One divergent position**: the shallow receiver-noise class
  (`a.String()` on `time.Time` vs `*big.Int`) that `--type-aware` exists to
  eliminate. Dropped.
- **≥2 divergent positions**: a generics-extraction candidate structure. Kept
  as a family group spanning all instances; same-type subgroups within the
  family are additionally reported by the type-aware pass as regular groups.

Presentation gates — actionability patterns, `--suggest-generics-min-lines`
hint gate, `--min-lines`, accept directives, `--show-suppressed` — apply
downstream, identically for both passes. A pattern-matching cross-type family
is hidden by actionability exactly as in single modes; the structural gate
never makes presentation decisions.

### Consequences

- A family with a same-type pair plus a divergent member (e.g. `A(int),
  B(int), C(string)`) reports BOTH `{A,B}` (type-aware pass, actionable
  duplicate) and `{A,B,C}` (candidate family) — strictly more informative
  than either single mode.
- If type loading fails, both passes fall back to syntax-only, produce
  identical results, and the filter drops the entire second pass — combined
  mode degrades to plain syntax-aware output.
- Incremental cache: the two passes use the existing `ta`/`sg` key tags; a
  combined run writes both entry classes. No `CacheVersion` change.
- Cost: one go/packages load + two transform/suffix-tree passes. The parse
  and tree passes are the cheap part relative to type-checking.

### Canonical constant

`MinDivergentPositions` moved from `printer/generics_candidate.go` to
`syntax/type_divergence.go` (canonical); printer re-exports it as a const
alias for its classification hint. `syntax.CountTypeDivergencePositions`
mirrors `printer.ClassifyGenericsCandidate`'s traversal (pre-order flatten,
shared-prefix comparison, empty-`VarType` skip) — pinned by
`printer/generics_parity_test.go`. Arch-lint note: `syntax` may not depend on
`domain` or `printer`, hence the constant lives in `syntax`.

## Implementation map

- `syntax/golang/typeinfo.go` — `TypeAwareData.WithEraseHash`.
- `syntax/type_divergence.go` — canonical constant + structural gate.
- `cmd/combined_analysis.go` — CLI two-pass orchestration, match-stream merge
  (`mergeCombinedMatches`), single crawl + single type load
  (`loadCombinedTypeData`).
- `cmd/run_analysis.go` — extracted per-pass runners
  (`buildSuffixTreeStandardPass`, `buildSuffixTreeIncrementalPass`) shared
  with single modes; combined branch in `executeAnalysis`.
- `pkg/artdupl/combined.go` — SDK two-pass mode for `FindClones` and
  `FindClonesStreamResult` (same merge rule at the group-map level).
- `cmd/config_builder.go` — `noteCombinedMode` replaces the old warning.
