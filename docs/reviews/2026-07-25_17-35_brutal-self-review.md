# Brutal Self-Review: Dedup Session Follow-up

**Date:** 2026-07-25 17:35
**Scope:** Follow-up to the dedup-to-zero session — fixing the bugs I left behind.
**Reviewer:** The agent, on itself.

---

## The 11 Questions (honest answers)

### 1. What did I forget?
I forgot to run tests after **each** change. I batched 6 refactors across files and tested only at the end. The infinite-recursion bug in `helper()` lived for ~20 minutes and was caught by the BDD suite, not by me. The critical rule "After each change: run tests" was violated repeatedly.

### 2. What's stupid that we do anyway?
The `//art-dupl:accept` scanner used `strings.HasPrefix` after `TrimSpace` — it **silently dropped every inline trailing directive** (`code(); //art-dupl:accept ...`). Go's entire linter ecosystem (golangci-lint, revive, eslint) supports inline directives. art-dupl documented "directive can be on any line" but the scanner only honored standalone comment lines. I **found** this bug last session and **worked around it** instead of fixing it. That was the stupidest decision in the session.

**FIXED THIS SESSION:** Scanner now uses `strings.Index` to find the directive anywhere on the line. Two regression tests added (`TestAcceptedSetInlineDirective`, `TestAcceptedSetInlineDirectiveWithHash`).

### 3. What could I have done better?
Never use `replace_all` on a code pattern immediately after introducing new code whose body matches that pattern. After adding `helper()`, its body `if s.T != nil { s.T.Helper() }` became a match target, and `replace_all` rewrote it into `s.helper()` → infinite recursion. The safe move: re-read the file after the structural addition, then replace external call sites with a more specific pattern (or individual edits).

### 4. What could I still improve?
- **Type models:** The `FilterStats.byReason`/`bySource` maps are `map[string]int` but the keys are really `FilterReason`/`FilterSource` (both string aliases). I investigated typing them and **decided against it**: the keys are stringified at a single typed insertion site (`s.bySource[string(source)]` where `source` is already `FilterSource`), so typing the map catches nothing new. The external printer API (`map[string]int`) would need a conversion layer that adds noise without caller-side safety. **Honest ROI: low. Skipped deliberately.**
- **Remaining `switch name` predicates:** `isCleanupMethod`, `isLoggingMethod`, `isAssertionMethod`, `isWrappingCallName` still use the switch form. Only the two art-dupl flagged were converted. Low urgency (they don't match anything currently), but consistency argues for converting all four.

### 5. Did I lie to you?
No. But I understated: I called the scanner bug a "limitation" in the status report when it is a **correctness bug contradicting the tool's own documentation**. "Limitation" implies intentional design; this was an accidental restriction.

### 6. How can we be less stupid?
- **Fix root causes on sight.** I found the scanner bug and worked around it. The opposite of the principle "fix issues on sight" and "fix root cause not symptom." This session: fixed.
- **A regression test for every bug fixed.** Last session I fixed `ErrInvalidDetectionMode` aliasing with no test. This session: added `TestErrInvalidDetectionModeAlias` locking in `errors.Is` cross-package equality.
- **Don't trust build == correct.** Infinite recursion compiles. Test the changed package immediately.

### 7. Is everything correctly integrated or are we building ghost systems?
No ghost systems. All 19 `//art-dupl:accept` directives verified live (plumbing output is empty at `-t 1`). The scanner fix was verified end-to-end: dedup still reports zero, all 24 packages pass.

### 8. Are we focusing on the scope creep trap?
No. Stayed in the dedup/correctness scope. The type-model investigation was requested and I reported the honest ROI rather than doing busywork.

### 9. Did we remove something that was actually useful?
No.

### 10. Did we create ANY split brains?
One **small** one, now closed: AGENTS.md described the accept-directive semantics without mentioning inline support. The scanner now supports inline, and AGENTS.md is updated to say so. Doc/code alignment restored.

### 11. How are we doing on tests? What can we do better?
- **Added this session:** 3 new tests (2 for scanner inline support, 1 for error-sentinel aliasing).
- **Still missing:** targeted tests for `withLock[T]` (nil-receiver safety) and `diffStatTable` (all accessors). Both are covered transitively by existing tests, but a focused unit test would lock the contract. Deferred — not blocking, and the transitive coverage is real.
- **Process improvement applied:** tested after EVERY change this session, not at the end.

---

## What I executed this session (verified)

| Step | Change | Test after? | Result |
|------|--------|-------------|--------|
| 1 | Scanner: `HasPrefix` → `Index` (recognize inline directives) | ✅ `TestAcceptedSet*` | PASS |
| 2 | + `TestAcceptedSetInlineDirective` regression test | ✅ | PASS |
| 3 | + `TestAcceptedSetInlineDirectiveWithHash` regression test | ✅ | PASS |
| 4 | `TestErrInvalidDetectionModeAlias` (cross-package `errors.Is`) | ✅ | PASS |
| 5 | AGENTS.md: document inline-directive support | ✅ build | PASS |
| 6 | Full suite `go test ./...` | ✅ all 24 pkgs | PASS |
| 7 | Dedup verification `-t 1` | ✅ empty output | 0 clones |

---

## Architectural notes

### Type model (deliberate non-change)
The `FilterStats` stringly-typed maps were investigated and **intentionally left as-is**. Both key types (`FilterReason`, `FilterSource`) are `string` aliases, and stringification happens at a single typed insertion point. Typing the internal maps would add a conversion layer at the printer boundary (`map[string]int` is the public contract) without catching any real bug. This is documented here so the next reviewer doesn't re-investigate the same tradeoff.

### Library consideration
The name-set predicates (`isAcquireMethod`, `isTestingVarName`) were converted to `slices.Contains` last session — the correct idiomatic Go choice (stdlib, no dependency). No third-party library would improve this.

---

## Remaining work (honest priority order)

1. **[LOW EFFORT, MEDIUM IMPACT]** Convert remaining 3 `switch name` predicates to `slices.Contains` for consistency.
2. **[LOW EFFORT, LOW IMPACT]** Add focused unit tests for `withLock[T]` and `diffStatTable` (transitive coverage exists).
3. **[MEDIUM EFFORT, MEDIUM IMPACT]** Add a CI self-test gate: `art-dupl -t 1 --plumbing` on its own repo must emit zero lines.
4. **[MEDIUM EFFORT, MEDIUM IMPACT]** Audit repo-wide for other duplicated `errors.New("...")` sentinels (same message, different pointers) — the class of bug fixed in `ErrInvalidDetectionMode`.

---

*Point-in-time review. Honest by construction.*
