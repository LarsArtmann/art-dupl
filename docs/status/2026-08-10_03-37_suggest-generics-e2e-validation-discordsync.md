# Status Report: `--suggest-generics` E2E Validation on DiscordSync

**Date:** 2026-08-10 03:37
**Session scope:** Ran the new local `--suggest-generics` build against `/home/lars/projects/DiscordSync/` and compared against the globally-installed `art-dupl 0.6.1-c170f4d`. This report covers ONLY what was observed and measured in this session.

---

## A. FULLY DONE

### E2E comparison completed (4 scan modes)

| Scan Mode | Old (`0.6.1-c170f4d`) | New (local `dev`) | Match? |
| --- | --- | --- | --- |
| Plain (`-t 1 --sort total-tokens`) | 31 groups | 31 groups | **Identical hashes** (zero regression) |
| `--type-aware` | 0 groups | 0 groups | Identical |
| `--suggest-generics` | _(flag doesn't exist)_ | **24 groups** | New feature |
| `--suggest-generics --no-actionability` | N/A | **135 groups** | New feature |

### All 3 known false-negatives recovered (100% recall)

The feedback doc identified 3 groups that `--type-aware` incorrectly suppressed. All 3 are recovered by `--suggest-generics`:

| Feedback # | Shape | Files | Type diff in hint |
| --- | --- | --- | --- |
| #3 | `var highest int64; for...{ if X > highest }` | activity_helpers:142,168 + handlers_extras:289 | `AuthorKindActivity` vs `TopReaction` vs `MemberGrowthPoint` |
| #9 | `var total int64; for...{ total += X }` | activity_helpers:192 + attachment_analytics:167 | `AttachmentCategoryStat` vs `AuthorKindActivity` |
| #10 | `kind := user.Kind; if kind == "" {...}` | db/entities:131 + projection/users:23 | `events.UserPayload` vs `*db.User` |

### Output format verification

| Format | Generics fields present? | Details |
| --- | --- | --- |
| Text | **Yes** | `generics: same algorithm, different types: ...` line before each group |
| `--explain` (text) | **Yes** | `generics:` line shown alongside `explain:` line |
| JSON | **Yes** | `generics_candidate: true` + `generics_hint: "..."` on every clone in every group (59 clones across 24 groups) |
| SARIF | **NO** | 59 results, zero generics fields anywhere |
| Plumbing | **NO** | File:line ranges only, no generics info |

### No-regression verification

Old and new plain scans produce byte-identical clone group hashes (31/31 match via `diff` on sorted hashes). The `--suggest-generics` feature is purely additive.

### Conflict behavior confirmed

`--type-aware --suggest-generics` together produces **24 groups** (identical to `--suggest-generics` alone). Suggest-generics silently takes precedence. No validation error, no warning to the user.

### Performance measured

| Mode | Time | Notes |
| --- | --- | --- |
| Plain (old binary) | 0.55s | Parse-only |
| `--suggest-generics` (new binary) | 57.7s | go/packages type checking |

**~105x slower**, confirming the AGENTS.md claim of "10-100x slower than parsing alone."

---

## B. PARTIALLY DONE

### Feature is functionally complete but lacks polish

The core mechanism works: type-erased hashing causes structural matches, and `ClassifyGenericsCandidate` correctly identifies type differences. But output quality, precision, and format coverage are incomplete (see sections C and E).

---

## C. NOT STARTED

1. **No user docs**: `HOW_TO_USE.md`, `FEATURES.md`, `TODO_LIST.md`, `CHANGELOG.md` not updated.
2. **No ADR**: EraseHash design decision undocumented in `docs/adr/`.
3. **No SARIF support**: SARIF output contains zero generics fields.
4. **No plumbing support**: Plumbing format has no generics info.
5. **No SDK E2E test**: `Options.SuggestGenerics` only has unit-test coverage, no integration test.
6. **No full-pipeline integration test**: Only unit tests exist; no test that exercises parse -> type-load -> detect -> classify -> output end-to-end on real Go files.
7. **No BDD/Ginkgo test**: No `--suggest-generics` scenario in `bdd/`.
8. **Incremental cache bug not fixed**: Cache key in `job/incremental.go` doesn't distinguish `type-aware` from `suggest-generics` mode. A cached AST from a `--type-aware` run would be incorrectly reused by a `--suggest-generics` run (wrong hash encoding).
9. **Config validation not added**: `--type-aware --suggest-generics` silently runs suggest-generics with no warning or error.
10. **Feedback doc not annotated**: Still in `docs/feedback/new/`, not moved to `docs/feedback/done/`.

---

## D. TOTALLY FUCKED UP

### D1. CRITICAL: 12.5% precision — 21 of 24 surfaced groups are noise

This is the single biggest problem with the feature, and I did not catch it until asked "what did you forget."

The feedback doc manually analyzed all 31 clone groups from the plain scan. It judged **3 as real generics-extraction candidates** (false-negatives from `--type-aware`) and **28 as correctly-suppressed false-positives**.

`--suggest-generics` surfaces **24 groups** (all 24 have type differences). But only **3 of those 24** are real generics candidates per the feedback doc's manual analysis. The other **21** are boilerplate that the feedback doc explicitly marked "correctly filtered":

| Surfaced group | Feedback verdict | Why it's noise |
| --- | --- | --- |
| `if err != nil { queryRowErr(...) }` x7 | Correctly filtered (#1,#2) | Already a shared helper; different entity types don't make this a generics candidate |
| `if err != nil || ch == nil` x3 | Correctly filtered (#4) | Nil-guard boilerplate, not algorithmic duplication |
| `if err != nil` detail-author x3 | Correctly filtered (#5) | Same pattern |
| `if result.GuildID != ""` x5 | Correctly filtered (#7) | Pagination guard, different view-model types |
| `envDuration`/`envInt` x2 | Correctly filtered (#14) | Already a helper; `time.Duration` vs `int` doesn't make it genericizable |
| `avatars/icons` refresh x2 | Correctly filtered (#22) | Already has `//nolint:dupl`; different download types |
| `if ch != nil && ch.Name` x2 | Correctly filtered (#19) | Nil-check boilerplate |
| `resolveMention` map->slice x2 | Correctly filtered (#12) | Different collection types, already extracted |
| `if avatar == "" vs icon == ""` x2 | Correctly filtered (#11) | Guard clause |
| `withDefaultDuration/Int` x2 | Correctly filtered (#15) | Same as #14 |
| `args := make([]any, ...)` x2 | Correctly filtered (#24) | SQL args boilerplate |
| `if projectionName != "" vs guildID != ""` x2 | Correctly filtered (#16) | SQL WHERE guard |
| `restFetchUser/Guild` x2 | Correctly filtered (#13) | Error-wrapping tail |
| `for _, k := range kinds` x2 | Correctly filtered (#18) | Different element types, but iteration boilerplate |
| `if pct < 1 vs offset < 0` x2 | Correctly filtered (#26) | Clamp guard |
| `if err != nil || user/guild == nil` x2 | Correctly filtered (#27) | Nil-guard |
| `IsAttachmentGoneError` x2 | Correctly filtered (#31) | Error-check tail |
| `ParentID() != nil` x2 | Correctly filtered (#28) | Nil-guard |
| `for id := range ids` x2 | Correctly filtered (#29) | Keys extraction idiom |
| `if att.URL == "" vs media.URL == ""` x2 | Correctly filtered (#30) | Skip guard |

**Scorecard: 3 true positives, 21 false positives. 12.5% precision, 100% recall.**

The root cause: type-difference detection is **necessary but not sufficient** for identifying generics-extraction candidates. Having different types at corresponding positions doesn't mean the code would benefit from generics. Error-handling boilerplate, nil-guards, and collection idioms all have type differences but are NOT algorithmic duplication that generics would eliminate.

### D2. Hint verbosity — fully-qualified type names are unreadable

Every `generics_hint` uses the full Go package path:

```
same algorithm, different types: github.com/larsartmann/DiscordSync/internal/db.GuildID vs github.com/larsartmann/DiscordSync/internal/db.ChannelID; github.com/larsartmann/DiscordSync/internal/db.Guild vs github.com/larsartmann/DiscordSync/internal/db.Thread; github.com/larsartmann/DiscordSync/internal/db.Guild vs github.com/larsartmann/DiscordSync/internal/db.Channel (11 type differences total)
```

This is **375 characters** on a single line. In a terminal, it wraps and becomes unreadable. Should be:

```
same algorithm, different types: db.GuildID vs db.ChannelID; db.Guild vs db.Thread (11 type differences total)
```

The fix is to strip the module path prefix from `go/types` string output (or use `types.TypeString` with a `types.Qualifier` that returns short names).

### D3. Built binary from wrong path on first attempt

I ran `go build -o /tmp/art-dupl-new ./cmd/` which produced an ar archive instead of an executable (because `./cmd/` contains subdirectories, not a main package). The correct path is `./cmd/art-dupl/`. This wasted a round trip. The `file` command caught it (`current ar archive`).

### D4. Binary disappeared from /tmp mid-session

After the initial comparison runs, `/tmp/art-dupl-new` vanished (possibly systemd-tmpfiles cleanup). Had to rebuild. Should have placed it in the project directory or `/home/lars/` instead.

---

## E. WHAT WE SHOULD IMPROVE

### E1. Add precision filtering (highest priority)

The feature needs additional heuristics to distinguish "same algorithm over different types" (real generics candidate) from "same boilerplate happening to use different types" (false positive). Possible approaches:

- **Exclude single-statement groups**: Most noise (nil-guards, error-checks, clamp guards) are 1-2 statement clones. The 3 real candidates are 4-8 line algorithmic blocks. A minimum-line-count gate for `--suggest-generics` (e.g., `MinLines: 4`) would eliminate most noise.
- **Exclude already-extracted helpers**: Groups where the duplication is already inside a shared function call (e.g., `queryRowErr(...)`) are not generics candidates.
- **Require multiple type-difference positions**: The 3 real candidates have type differences at multiple structural positions (loop variable, field accessor, accumulator). Many false positives have only 1 type difference at a shallow position.
- **Pattern-aware filtering**: Reuse the actionability pattern system. Groups matching `error-propagation`, `bool-guard`, `guard-clause`, `single-call-expression`, etc. should be excluded from generics suggestions even when they have type differences.

### E2. Shorten type names in hints

Use `types.RelativeTo(pkg)` or string manipulation to strip module paths. Display `db.GuildID` instead of `github.com/larsartmann/DiscordSync/internal/db.GuildID`.

### E3. Add SARIF support

SARIF output currently drops all generics info. Add `generics_candidate` and `generics_hint` to the `properties` bag of each SARIF result.

### E4. Add config validation or warning for flag conflict

`--type-aware --suggest-generics` silently runs suggest-generics. Either reject with an error, or print a warning: "suggest-generics takes precedence over type-aware."

### E5. Fix incremental cache key

Cache key must include the `eraseHash` mode to prevent stale-cache correctness bugs when switching between `--type-aware` and `--suggest-generics`.

### E6. Add full-pipeline integration test

The current tests are unit-level (construct `CloneNode` trees manually, test `ClassifyGenericsCandidate` in isolation). Need a test that: writes Go files -> runs `LoadTypeAwareData(eraseHash=true)` -> runs detection -> runs classification -> asserts generics candidate is flagged.

---

## F. NEXT TASKS (up to 50)

### Precision & Quality (highest impact)

1. Add minimum-line-count gate (`MinLines >= 4`) for `--suggest-generics` groups to filter single-statement noise
2. Cross-reference generics candidates against actionability patterns — exclude groups matching boilerplate patterns
3. Require N+ type-difference positions at non-trivial depth (filter shallow 1-position differences)
4. Evaluate excluding groups where the clone body is dominated by a single `CallExpr` to a shared helper
5. Re-run against DiscordSync after precision filtering — target: surface the 3 real candidates with <5 false positives
6. Add a `--suggest-generics-min-lines` flag (default 4) to let users tune precision vs recall
7. Consider a "confidence score" for generics candidates (based on statement count, type-diff depth, pattern match)

### Type Hint Quality

8. Strip module path from type strings in `formatGenericsHint` (use package-relative names)
9. Truncate hint to N unique type pairs (currently 3, but each pair is very long)
10. Consider multi-line hint format for text output (one type pair per line)
11. Add short-form hint for JSON (e.g., `"type_diffs": 11` count alongside the verbose hint)

### Output Formats

12. Add `generics_candidate` + `generics_hint` to SARIF `properties` bag
13. Add generics info to plumbing output format
14. Verify `--explain` text mentions "generics-extraction candidate" explicitly (currently shows `generics:` line separately, may be unclear)

### Config & Validation

15. Add validation or warning for `--type-aware --suggest-generics` conflict
16. Fix incremental cache key in `job/incremental.go` to include `eraseHash` mode
17. Add `SuggestGenerics` to `ValidateConfig()` for early conflict detection
18. Consider making `--suggest-generics` imply `--semantic` mode (currently relies on it being the default)

### Tests

19. Write full-pipeline integration test (Go files -> type load -> detect -> classify -> assert)
20. Add SDK unit test for `Options.SuggestGenerics: true` in `pkg/artdupl/`
21. Add BDD/Ginkgo scenario for `--suggest-generics` in `bdd/`
22. Add test that `--suggest-generics` output survives JSON round-trip (marshal -> unmarshal -> verify fields)
23. Add test for conflict behavior (`--type-aware --suggest-generics` produces expected result/warning)
24. Add benchmark test for `ClassifyGenericsCandidate` on large clone groups
25. Add test that `generics_hint` uses short type names (after fix #8)
26. Add fuzz test for `ClassifyGenericsCandidate` with malformed node trees

### Documentation

27. Write ADR-0020 documenting the EraseHash design decision
28. Update `HOW_TO_USE.md` with `--suggest-generics` section and usage examples
29. Update `FEATURES.md` — mark `--suggest-generics` as DONE
30. Update `TODO_LIST.md` — mark feedback item as done
31. Update `CHANGELOG.md` with the new feature
32. Annotate/move feedback doc from `docs/feedback/new/` to `docs/feedback/done/`
33. Add `--suggest-generics` to the CLI help text examples
34. Document the 100x performance cost in user-facing docs

### SDK

35. Add `SuggestGenerics` to SDK README/examples
36. Verify SDK `Options.SuggestGenerics` produces same results as CLI flag
37. Add SDK integration test that exercises the full detector pipeline with `SuggestGenerics: true`

### Actionability Integration

38. Audit which actionability patterns should auto-exclude generics candidates
39. Test `--suggest-generics --no-actionability` behavior is documented and intentional
40. Consider whether generics candidates should bypass actionability by default (current: they don't)

### UX Polish

41. Add progress indicator for type-checking phase (57s with no feedback is bad UX)
42. Consider streaming output for large codebases instead of blocking on full type-check
43. Add `--suggest-generics-threshold` to control minimum type-difference count
44. Color-code generics hint in text output (green for high-confidence, yellow for low)
45. Add `--suggest-generics-format` option (verbose vs compact hint)

### Architecture

46. Consider extracting type-name shortening into a shared `syntax/golang/typeformat.go` utility
47. Evaluate whether `ClassifyGenericsCandidate` should live in `printer/` or `domain/` (currently in printer, but it's classification logic not formatting)
48. Consider adding `GenericsCandidate` to `pkg/artdupl.Clone` (SDK type) for parity with `domain.ProcessedClone`
49. Audit whether `VarType` should be carried through to `pkg/artdupl.CloneNode` (SDK type) for SDK consumers
50. Evaluate whether the `EraseHash` boolean should be a `DetectionMode` value instead (cleaner type-safety, more invasive)

---

## G. QUESTIONS (cannot determine without user input)

### G1. Precision target: what false-positive rate is acceptable?

The feature currently has 12.5% precision (3 true / 24 surfaced). I can add heuristics (min-lines, pattern exclusion, multi-position requirement) to improve this, but each filter risks missing real candidates. What precision/recall tradeoff do you want? Options:
- **Conservative** (surface only high-confidence, may miss some real candidates) — target <5 groups on DiscordSync
- **Balanced** (moderate filtering, some noise acceptable) — target ~10 groups
- **Permissive** (current behavior, let users filter manually) — keep 24 groups

### G2. Should `--type-aware --suggest-generics` be a hard error or a warning?

Currently it silently runs suggest-generics (takes precedence). The implementation handoff listed this as an open question. I cannot determine your preference — it depends on whether you want users to ever combine these flags intentionally (I see no valid use case for running both simultaneously, but you may disagree).

### G3. Should generics candidates bypass actionability filtering by default?

Currently, `--suggest-generics` WITH actionability produces 24 groups; WITHOUT produces 135. The 3 real candidates are in both sets. The question is whether generics candidates should be shown even when they match boilerplate patterns (since the whole point is finding duplication that looks like boilerplate but is actually extractable). I cannot determine this without your product judgment — it changes the feature's default user experience significantly.
