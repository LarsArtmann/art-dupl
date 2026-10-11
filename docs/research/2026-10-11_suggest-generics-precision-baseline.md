# Suggest-Generics Precision Baseline

**Date:** 2026-10-11
**Task:** T1 of the generics-precision cluster (`docs/planning/2026-10-11_05-06_generics-precision-from-branching-flow-lessons.md`)
**Question:** What is the current TRUE/FALSE precision of `--suggest-generics` on the validation corpus, and WHY are the false ones false?

## Method

Binary: `go build ./cmd/art-dupl` at fork HEAD `21206db0` (v-dev). Runs: `--suggest-generics --no-actionability -t {5,3,2,1} --json` over the 6-repo corpus (go-sse, go-cqrs-lite, go-paperless, DiscordSync, art-dupl self, branching-flow HEAD). Every candidate group labeled TRUE/FALSE by reading the source at each instance location. Two supplementary probes: `--suggest-generics-min-lines 0` (line gate off) and a temporary `MinDivergentPositions 2→1` build (divergence gate off, source restored after measurement).

## Raw numbers

| Repo            | groups @t5 | @t2 | @t1  | candidates (any t)        |
| --------------- | ---------- | --- | ---- | ------------------------- |
| go-sse          | 0          | 10  | 74   | **0**                     |
| go-cqrs-lite    | 13         | 647 | 2926 | **0**                     |
| go-paperless    | 1          | 19  | 74   | **0**                     |
| DiscordSync     | 15         | 209 | 1115 | **0**                     |
| art-dupl (self) | 0          | 51  | 408  | **9 groups / 4 families** |
| branching-flow  | 1          | 31  | 414  | **0**                     |

Candidates survive in art-dupl self at `-t 2` and `-t 1` identically (9 groups, 4 distinct content hashes). No other repo emits a single candidate at ANY threshold with ANY gate disabled.

## Labels (art-dupl self, the only labelable population)

| Family hash | Instances                                                                                       | Verdict  | Reason                                                                                    |
| ----------- | ----------------------------------------------------------------------------------------------- | -------- | ----------------------------------------------------------------------------------------- |
| `17b5f676`  | `syntax/golang/snippet.go:188-195` + `:197-204` (`caseExprsAsNodes`/`statementsAsNodes`)        | **TRUE** | Genuine `[]T → []ast.Node` conversion pair; a generic `asNodes[T ast.Node]` eliminates it |
| `0cfc073b`  | `cmd/type_aware.go:49-52`, `job/incremental_parallel_test.go:53-56`, `syntax/syntax.go:532-536` | FALSE    | Trivial append-collector tails (`x = append(x, y)` in a loop) in otherwise-unrelated code |
| `c4ea6cba`  | `cmd/diff_report.go:107-110`, `pkg/artdupl/detector_conversion.go:32-35`                        | FALSE    | Guard + append tail idiom, not an algorithm                                               |
| `e12d4e29`  | `internal/testutil/file.go:81-84`, `syntax/syntax_match.go:167-171`                             | FALSE    | Append tail idiom again                                                                   |

**Precision: 1/4 families = 25%** (2/9 groups = 22%). The ROADMAP's "12.5%" figure is not reproducible on the current corpus at any threshold — treat 25% @ n=4 as the working baseline.

## False-reason taxonomy (observed)

1. **trivial-append-collector-tail (NEW, 3/4 of FPs):** the matched fragment is a ubiquitous Go idiom tail, not a duplicated algorithm. A generics suggestion here is noise. Candidate fixes: an actionability-style pattern (`append-collector-tail`) or a fragment-shape gate in candidacy.
2. **alias-string / instantiation-args-only / already-generic / min-lines noise:** the plan's anticipated categories — NONE observed in this population. They remain hypothesis-stage until a corpus exhibits them (T4's alias fix is still correct on first principles; it just has zero measured field occurrences today).

## The recall finding (dominant over precision)

Field recall is ~zero: 5/6 repos produce 0 candidates even with `--suggest-generics-min-lines 0`, `-t 1`, and (probed on go-cqrs-lite + DiscordSync) `MinDivergentPositions=1`. The eligibility pool is NOT the blocker — go-cqrs-lite at `-t 2` has 907 pattern-free clone instances (no `non_actionable_pattern`). Two live explanations, not yet distinguished:

- (a) The corpora genuinely lack same-shape-different-types duplication (all five are actively deduplicated with art-dupl; the textbook fixture fires correctly through the full CLI, so the pipeline works when the shape exists).
- (b) VarType population is too sparse: only LOCAL-ident use sites carry VarType (`transform.go::encodeTypeIfAware`, gated on `norm.isLocal`), so type divergence that manifests at non-local positions (composite-literal keys, selectors, instantiation args) is invisible. T7/T8 directly enlarge this surface.

## Doc correction discovered during measurement

`AGENTS.md` and `--help`-adjacent docs say "Actionability still applies (use `--no-actionability` to see all candidates)". That is misleading: `--no-actionability` only un-filters OUTPUT (`cmd/run_output.go:148`); pattern evaluation always runs and `printer/clone_processor.go:167` gates candidacy on `label == PatternNone` regardless of the flag. A pattern-matched group NEVER becomes a generics candidate, even under `--no-actionability`.

## Consequences for the cluster (sizing adjustments)

- T4 (`types.Identical`): still correct, but its measured field impact is 0 candidates today; value is future-proofing the comparison for when T7/T8 widen the candidate surface.
- T5 (reuse-existing-generic): unaffected — it targets suggestion quality.
- T7/T8 (`Instances` wiring, type-args on erased pass) are the recall levers: they are what can make the feature fire on real code. Their priority relative to T4 rises.
- T12/T13 (actionability waves): consider adding `append-collector-tail` — it would have suppressed 3 of 4 false families (75% of observed FPs). Slot it into wave 2 alongside nil-guard.
- Precision re-measurement should repeat THIS method after T7/T8 land; n=4 is too small for confidence, and the population will change.

## Ledger

| Date       | Task              | Outcome                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| ---------- | ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 2026-10-11 | T4 (alias FP fix) | Shipped as population-time `types.Unalias` in `identTypeString` (not the planned printer-side comparator — single seam, same precision). Alias vs direct now produce identical VarType; import-alias spelling was ALREADY path-canonical (pinned by test). Corpus candidates unchanged (0/0/0/9) as predicted; static-fixture output byte-stable; CacheVersion 5→6 (ta/sg streams changed). Corpus repos drift under their own daemons — future before/after comparisons must pin corpus commits or use static fixtures. |
