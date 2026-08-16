# ADR-0021: Suggest-generics type-erased hashing and precision gates

## Status

Accepted (2026-08-15)

## Context

`--suggest-generics` aims to surface the class of duplication that Go generics
can eliminate: clones where the **algorithm is identical but local variable
types differ** across instances (e.g. `SumFirst(f, g First)` vs
`SumSecond(s, t Second)`).

This is exactly the class that `--type-aware` **suppresses**: type-aware
hashing appends each local's static type to its identifier hash, so two
structurally-identical functions on different types produce different hashes
and never match. The two flags are duals:

| Flag                 | Hash includes types? | Purpose                                         |
| -------------------- | -------------------- | ----------------------------------------------- |
| `--type-aware`       | Yes                  | Suppress same-shape-different-type FPs          |
| `--suggest-generics` | No (erased)          | Surface same-shape-different-type as candidates |

An early design ran suggest-generics as a **filter** (only candidates shown).
The 2026-08-10 redesign made it an **enhancer**: all groups are shown and
candidates are annotated (see `docs/status/2026-08-10_14-49_suggest-generics-enhancer-redesign.md`).

Initial real-world validation on DiscordSync surfaced a precision problem:
12.5% precision (3 true / 24 surfaced). Type divergence alone is necessary but
not sufficient — error-handling boilerplate and nil-guards also exhibit type
differences but are NOT generics candidates.

## Decision

### 1. Type-erased hashing via `EraseHash`

`LoadTypeAwareData(files, eraseHash=true)` reuses the full `--type-aware`
pipeline (`go/packages` type checking) but sets
`PreloadedAST.EraseHash = true`. The transformer's `encodeTypeIfAware()`
populates `Node.VarType` but does **not** append the type to the identifier
hash. Structurally-identical functions on different types therefore match,
while `VarType` is still available for post-detection classification.

Cache isolation: the incremental cache key carries a `typeAwareTag`
(`"ta"` = type-aware, `"sg"` = suggest-generics) so erased-hash ASTs are never
reused for type-aware runs and vice versa (`job/incremental.go::cacheKey`).

When both flags are passed, suggest-generics wins (eraseHash=true overrides
type-aware hashing); the CLI prints a warning
(`warnTypeAwareSuggestGenerics` in `cmd/config_builder.go`).

### 2. Three precision gates on candidacy

A group is a generics candidate only when ALL of the following hold:

1. **≥ `MinDivergentPositions` (2) distinct structural positions** with
   differing non-empty `VarType` values across instances
   (`printer/generics_candidate.go`). Single-position divergence is dominated
   by shallow idiom noise (one differently-typed call result).
2. **Min-lines gate**: every clone instance reaches
   `--suggest-generics-min-lines` (default 4, `0` disables). Mirrors
   `--min-lines` semantics: the weakest instance decides. Most noise was 1-2
   statement clones; the real candidates were 4-8 line blocks.
3. **No actionability-pattern match**: candidacy requires
   `label == actionability.PatternNone`. Boilerplate (error guards, single
   calls, guards) disqualifies even when types diverge.

Hints are deduplicated via `divergenceKey()` canonicalization (A vs B ==
B vs A) and package paths are stripped by `shortenTypeString()`.

### 3. Output wiring

Text: `generics:` hint line. JSON: `generics_candidate` + `generics_hint`
fields. SARIF: same fields as result properties (only present for
candidates).

## Consequences

- DiscordSync validation after the gates: **296 detected → 2 shown, both
  genuine, 0 false positives** (previously 82 groups at ~97% FP under
  `--min-lines 6`). The remaining projection-pair survivor
  (`emojis_relational` vs `stickers_relational`) is itself a textbook generics
  candidate.
- Cost: same ~10-100x slowdown as `--type-aware` (full type checking is
  required). Both modes are opt-in.
- The `EraseHash` flag lives per-entry on `PreloadedAST`; `SetTypeAwareData`
  validates that all entries share one value (warns on mixed maps) since the
  cache key is run-global. A collection-level type would enforce this at
  compile time but is a breaking change (see TODO_LIST, deferred).
- Precision gates are heuristic constants. `MinDivergentPositions = 2` and
  the default min-lines of 4 are tunable via flag (min-lines) or constant;
  recalibrate against real-world corpora if precision regresses.

## References

- Implementation: `printer/generics_candidate.go`, `printer/clone_processor.go`
  (`passesGenericsLineGate`, `WithGenericsMinLines`), `syntax/golang/typeinfo.go`
- Flags: `--suggest-generics`, `--suggest-generics-min-lines`,
  `--type-aware` (warning interaction)
- Tests: `printer/generics_candidate_test.go`,
  `printer/generics_integration_test.go`, `bdd/suggest_generics_test.go`
- Validation: `docs/feedback/new/2026-07-27-discordsync-t1-82-groups-97pct-false-positives.md`
- Related: ADR-0018 (type-aware detection), ADR-0017 (property engine)
