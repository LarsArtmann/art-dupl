# Status: Accept-Directive UX Fix + Test-Helper-Delegate Pattern

**Date:** 2026-07-25 06:28
**Branch:** fork (5 commits ahead of origin, NOT pushed)
**Task:** Execute `docs/planning/2026-07-25_05-14_SUPERB-accept-directive-ux-fix-and-test-helper-pattern.md`
**Outcome:** Functionally complete. Git history is a disaster.

> **Resolution (2026-07-25):** Feature shipped. Commits `95547e7f`, `c30f683d`, `a057928e`
> (HEAD of `fork`). `test-helper-delegate` is pattern #18 in the actionability system.
> CHANGELOG `[Unreleased]`, FEATURES.md, and AGENTS.md updated in the subsequent
> docs-health session. Accept-directive docs (`HOW_TO_USE.md`) remain a known gap.
> Verified: `art-dupl --semantic -t 2` reports 0 groups on the art-dupl codebase.

---

## a) FULLY DONE

### M01 — Accept-directive scanning fixed (2 bugs, not 1)

**Bug 1 (planned): Above-range scanning.** `cmd/accept_directive.go` — directives
now match up to 5 lines above `clone.LineStart`. Uses `max(1, LineStart-5)` to
clamp at file start. Matches every linter convention (golangci-lint, eslint,
revive all place directives on the line above).

**Bug 2 (DISCOVERED, not in plan): Hash-vs-description collision.** The scanner
treated ALL text after `//art-dupl:accept` as a precision hash. So
`//art-dupl:accept idiomatic test helper boilerplate` stored the entire sentence
as `d.Hash`, which never matched any real group hash, silently disabling the
directive. **This was the actual root cause of all 6 clone groups appearing.**
Fix: only treat single-token text (no spaces/tabs) as a hash; multi-word text
is a human-readable description → bare accept (matches any group).

**Tests added** (`cmd/accept_directive_test.go`):

- `TestAcceptedSetIsAccepted/directive_one_line_above_LineStart_suppresses` — above-range
- `TestAcceptedSetIsAccepted/directive_6_lines_above_LineStart_does_not_suppress` — beyond window
- `TestAcceptedSetDescriptionText` — multi-word description acts as bare accept

### M02 — `test-helper-delegate` actionability pattern

**New pattern** (`printer/actionability_boilerplate.go`): `isTestHelperDelegate`
detects 2-statement bodies where:

1. First statement is `t.Helper()` (ExprStmt → CallExpr → SelectorExpr, Name="Helper")
2. Second statement is any lone call expression (the delegate)

Registered as pattern #9 of 18 in `printer/actionability.go`. Constants:
`PatternTestHelperDelegate = "test-helper-delegate"`.

**Tests added** (`printer/actionability_integration_test.go`):

- `TestEvaluateActionabilityWithLabel_TestHelperDelegate` — positive match (bare-call delegate)
- `..._SelectorDelegate` — positive match (selector-call delegate like `assert.Equal`)
- `..._ThreeStmts` — negative (3+ statements = potentially actionable)
- `..._NonHelperFirstStmt` — negative (first call is `t.Parallel()`, not `t.Helper()`)

**Result:** 4/6 clone groups eliminated systemically. Works on ANY Go project,
not just art-dupl. Verified: `--no-accept-directives` still suppresses the
4 `t.Helper()` clones via the pattern alone.

### M03 — BDD test strengthened

`bdd/type_aware_test.go`: The old test asserted `ContainSubstring("other.go")`
(positive presence) but never verified suppression. Worse, the assertion was
**semantically wrong** — accept directives suppress the ENTIRE clone group,
so `other.go` (the duplicate partner) also disappears. Fixed to assert:
`ContainSubstring("Found total 0 clone groups")` +
`NotContainSubstring("accepted.go")` +
`NotContainSubstring("other.go")`.

### M04 — Redundant directives removed

4 `//art-dupl:accept` directives removed from `internal/testutil/assert.go`
(lines 35, 52, 60, 263). The test-helper-delegate pattern handles these
systemically now. Manual directives are no longer needed for `t.Helper()` clones.

### M05 — AGENTS.md updated

- Actionability patterns count: 17 → 18, added test-helper-delegate description
- Accept-directive entry: documented placement (within range OR 5 lines above)
  and hash-vs-description semantics

### M06 — ACTIONABILITY_PATTERNS.md updated

- Added `test-helper-delegate` row to pattern table
- Updated count: 17 → 18
- Updated priority order list

### M07 — Verification at multiple thresholds

| Threshold                     | Clone Groups | Status                                                |
| ----------------------------- | ------------ | ----------------------------------------------------- |
| `-t 1`                        | 19           | Expected noise (single-token matches)                 |
| `-t 2`                        | 0            | All 6 original groups suppressed                      |
| `-t 2 --no-accept-directives` | 2            | Only const-alias + example (pattern handles the rest) |
| `-t 3`                        | 0            | Clean                                                 |
| `-t 5`                        | 0            | Clean                                                 |

### M08 — Test suite + build

- **Build:** `GOEXPERIMENT=jsonv2 go build ./...` — passes
- **Tests:** 24/24 packages pass, 0 failures (including BDD, which had 22
  pre-existing failures in the handoff notes — now ALL pass)
- **Lint:** No NEW issues introduced. 1 `modernize` finding auto-fixed
  (`max()` builtin). Pre-existing exhaustruct/tagliatelle/gci noise unchanged.

---

## b) PARTIALLY DONE

### Git history — TOTALLY FUCKED (see section d)

The code is correct but the commit history is garbage. 9 auto-commits by
BuildFlow with wrong messages and wrong author.

### Documentation consistency

Updated AGENTS.md and ACTIONABILITY_PATTERNS.md, but did NOT update:

- `HOW_TO_USE.md` (lines 324-363 document accept directives — stale)
- `CHANGELOG.md` (has accept-directive entry from initial implementation)
- `FEATURES.md` (says "FULLY_FUNCTIONAL" — was actually BROKEN before this fix)

---

## c) NOT STARTED

- `HOW_TO_USE.md` update for accept-directive placement + description semantics
- `CHANGELOG.md` entry for the bug fixes
- `FEATURES.md` correction (accept directive was NOT fully functional before)
- `pkg/artdupl/types.go` directive check (plan F18 mentioned it; the existing
  directive in `detection/config.go` covers the same clone group, so it may be
  redundant — needs verification)
- Nix flake check (`nix flake check`) — only ran `go build`/`go test`
- Pushing to origin (5 commits ahead, not pushed)

---

## d) TOTALLY FUCKED UP

### Git commit history is GARBAGE

**BuildFlow auto-committed my work as I edited, with catastrophically bad
commit messages.** I didn't notice until near the end. The result:

```
95547e7f Unknown Author: feat(cmd): add accept_directive command
a057928e Unknown Author: docs(agents): add comprehensive agent guidelines and actionability patterns documentation
c30f683d Unknown Author: test(bdd): enhance type-aware testing and assertion utilities
e11586a2 Unknown Author: feat(printer): add actionability feature for print queue management
e2b19a56 Unknown Author: feat(cmd): implement accept_directive command
79ff82fb Unknown Author: docs(status): add postmortem for BDD fixture actionability fix
```

**Every single message is wrong:**

- "implement accept_directive command" — the command ALREADY EXISTED. I was
  FIXING TWO BUGS in it.
- "add actionability feature for print queue management" — I added ONE pattern
  checker to an EXISTING pattern system. There is no "print queue."
- "add accept_directive command" (95547e7f) — this was just the `max()` lint fix.
- Author is "Unknown Author <unknown@example.com>" on all commits.
- The messages violate the project's commit message conventions (no `fix:`
  prefix, no body explaining WHY, generic filler text).

**9 commits for what should have been 2-3 logical commits:**

1. `fix(cmd): accept directives now scan above clone range and handle descriptions`
2. `feat(printer): add test-helper-delegate actionability pattern`
3. `test(bdd): strengthen accept-directive suppression assertion`

**What I should have done:** Disabled BuildFlow auto-commit OR squashed
immediately after noticing. I did neither.

### The plan's root-cause analysis was wrong

The plan (which I wrote in the previous session) identified the line-range
check as Bug 1 (the "1% → 51%" fix). In reality, the hash-vs-description
collision was the dominant root cause — it silently disabled EVERY directive
with descriptive text, which was ALL of them. The line-range fix was real but
secondary. I discovered this during execution, which is good, but the plan
wasted analysis effort on the wrong primary cause.

---

## e) WHAT WE SHOULD IMPROVE

1. **Disable or control BuildFlow auto-commit.** It silently destroys commit
   history quality. At minimum, squash after each logical task.
2. **The accept-directive scanner is still line-based, not AST-based.** A
   `//art-dupl:accept` inside a string literal or block comment would be
   falsely detected. Documented as a known limitation but never fixed.
3. **No integration test for the above-range placement in a real temp dir.**
   I verified via unit tests and the repo itself, but never built a binary and
   tested directive-on-line-above in an isolated temp directory (the plan's F05).
4. **The `test-helper-delegate` pattern requires exactly 2 statements.** A
   helper with `t.Helper()` + blank line + delegate + return would NOT match.
   This is intentionally narrow, but may need widening based on real-world usage.
5. **BDD test fixture design is fragile.** The accept-directive test creates
   two files with identical code, one with a directive. But since the directive
   suppresses the ENTIRE group, both files disappear. The test now correctly
   asserts this, but the test name ("should suppress accepted groups entirely")
   reflects a semantic that wasn't documented before.
6. **The hash detection heuristic (`!strings.ContainsAny(rest, " \t")`) is
   fragile.** A hash with a trailing space (`//art-dupl:accept abc123`) would
   be treated as a description. `strings.TrimSpace` handles this, but a
   multi-line hash or tab-separated hash would behave unexpectedly.
7. **Pattern priority ordering is undocumented in code.** The slice in
   `evaluateActionabilityDetailed` defines priority, but there's no comment
   explaining WHY `test-helper-delegate` is #9 (after single-simple-statement,
   before error-wrapping). The order matters: `single-call-expression` (#7)
   would NOT catch the 2-statement helper, so `test-helper-delegate` must run
   after it but before the generic fallbacks.

---

## f) Up to 50 things to get done next

### Critical (git hygiene)

1. **Squash the 9 garbage commits into 2-3 logical commits with proper messages**
2. **Fix author attribution** (currently "Unknown Author")
3. **Push to origin** (5 commits ahead)
4. **Decide: amend history or leave it?** (irreversible if pushed)

### Documentation

5. **Update `HOW_TO_USE.md`** — accept-directive section (lines 324-363) is stale,
   doesn't mention above-range placement or description semantics
6. **Update `CHANGELOG.md`** — add entry for accept-directive bug fixes + new pattern
7. **Update `FEATURES.md`** — accept directive was NOT "FULLY_FUNCTIONAL" before
   this fix; correct the status or add a note
8. **Add code comment in `evaluateActionabilityDetailed`** explaining pattern
   priority ordering and WHY each pattern is in its position
9. **Verify `pkg/artdupl/types.go` directive** — plan F18 mentioned it; may be
   redundant with `detection/config.go` directive (same clone group, 2 occurrences)

### Testing

10. **Add integration test in temp dir** for above-range directive placement
    (plan F05 — never done, only unit-tested)
11. **Add test for hash with trailing whitespace** (`//art-dupl:accept abc123`)
12. **Add test for directive on LineStart boundary with above-range scan** (edge case)
13. **Add BDD test for `test-helper-delegate` pattern** end-to-end (currently only
    unit-tested at the pattern-checker level)
14. **Add test for `t.Helper()` + `return` (no delegate)** — should NOT match
    (delegate is required)
15. **Add test for `tb.Helper()` and `b.Helper()`** — different receiver names
    (currently only `t.Helper()` is tested, but the pattern checks Name="Helper"
    regardless of receiver)

### Actionability pattern improvements

16. **Consider widening `test-helper-delegate`** to allow `t.Helper()` + blank
    line + delegate (real-world helpers sometimes have a blank line)
17. **Consider `t.Cleanup()` + delegate pattern** — similar irreducible boilerplate
18. **Consider `t.Skip()` + reason pattern** — another single-call test idiom
19. **Add `t.Helper()` detection for `testing.TB` interface** (not just `testing.T`)
20. **Audit remaining -t 1 clone groups** (19 groups at threshold 1) — identify
    which are real noise vs actionable

### Accept-directive improvements

21. **Make directive scanning AST-aware** (not line-based) to avoid false
    positives in string literals / block comments
22. **Support `/* art-dupl:accept */` block comment style** (currently only `//`)
23. **Support range directives** (`//art-dupl:accept-line 10-15`) for multi-line accepts
24. **Add `--list-accepted` flag** to show which groups are accepted and why
25. **Warn on directives that match nothing** (orphaned directives after refactoring)

### Code quality

26. **Run `nix flake check`** — only ran `go build`/`go test`, never the full
    Nix CI pipeline
27. **Fix pre-existing `gci` formatting issue** on `printer/actionability.go:46`
28. **Run `golangci-lint` on the full repo** and triage the 93 pre-existing issues
29. **Add `acceptDirectiveScanAbove` to config** — currently hardcoded to 5;
    should be configurable for edge cases
30. **Consider extracting `firstChild` helper** — it's generic enough to be reused
    by other pattern checkers

### Architecture / design

31. **Decouple accept-directive scanning from file reading** — currently tightly
    coupled via `readFile` func; consider an interface for testability
32. **Cache directive scan results across runs** — currently per-run only; a
    persistent cache would speed up CI
33. **Add metrics on pattern hit rates** — track which patterns fire most often
    to guide future improvements
34. **Consider a `--explain-suppression` flag** — show WHY a group was suppressed
    (which pattern or directive matched)
35. **Profile actionability evaluation** — 18 pattern checks per group could be
    slow on large codebases; consider short-circuit optimizations

### Remaining plan items

36. **M07 verification at -t 1** — done but not analyzed (19 groups, no breakdown)
37. **Investigate WHY all 22 BDD pre-existing failures now pass** — the handoff
    said they were unrelated to this work, but they all pass now. Correlation?
38. **Verify the `max()` builtin works on Go 1.26** (project uses GOEXPERIMENT=jsonv2)
39. **Check if `go vet` passes** (separate from golangci-lint)
40. **Run tests with `-race` flag** on the modified packages

### Future patterns to consider

41. **`context.WithCancel` + `defer cancel()`** — common boilerplate pair
42. **`if !ok { return }` type-assertion guard** — map/index guard idiom
43. **`switch v := x.(type)` dispatch** — type-switch boilerplate
44. **`for _, x := range` + single append** — loop-accumulate idiom
45. **`errgroup.Group` setup** — concurrent error handling boilerplate

### Process improvements

46. **Always check `git log` BEFORE starting work** — I would have noticed
    BuildFlow auto-committing much earlier
47. **Squash after each M-task** — prevents commit-message drift
48. **Verify the plan's root-cause analysis with a quick test BEFORE writing the
    full plan** — would have caught the hash bug earlier
49. **Add a `make verify` or `just verify` target** that runs build + test + lint
    in one command
50. **Consider a pre-commit hook that rejects generic commit messages** —
    prevents "feat(cmd): implement X" when X already exists

---

## g) Questions I cannot figure out myself

### 1. Should I squash the 9 BuildFlow commits into clean logical commits?

The 9 auto-commits have wrong messages and wrong author ("Unknown Author").
Squashing would give a clean history like:

- `fix(cmd): accept directives scan above range and handle descriptions`
- `feat(printer): add test-helper-delegate actionability pattern`
- `test(bdd): strengthen accept-directive suppression assertion`

But the global AGENTS.md says **"NEVER `git reset`"** and **"NEVER force push"**.
Squashing requires `git reset --soft` or `git rebase -i`, both of which violate
the safety rules. **Do you want me to break the rule this once, or leave the
garbage history and just fix going forward?**

### 2. The 22 "pre-existing" BDD failures from the handoff now ALL pass. Why?

The handoff said: _"22 BDD specs fail identically on original and modified code
(confirmed via stash comparison). These are unrelated to dedup work — likely a
`RunArtDupl` subprocess/binary issue."_ Now ALL 283 BDD specs pass. My changes
to `bdd/type_aware_test.go` and `internal/testutil/assert.go` should not have
fixed 22 unrelated specs. **Was the baseline wrong, or did a concurrent
agent/session fix them in commits `84199352`/`61b51837`/`79ff82fb`/`55cba9d3`?**

### 3. Push to origin now, or after squashing?

Branch is 5 commits ahead of `origin/fork`. The global rules say
**"NEVER PUSH TO REMOTE unless explicitly asked."** But the work is functionally
complete and the handoff context implied pushing after each M-task
(_"Commit after each M-task with detailed message, then push."_).
**Do you want me to push as-is, squash first then push, or hold?**
