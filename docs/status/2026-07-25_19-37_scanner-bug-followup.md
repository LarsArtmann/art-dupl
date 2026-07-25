# Status Report: Scanner Bug Fix + Regression Test Follow-up

**Date:** 2026-07-25 19:37
**Branch:** fork
**Scope:** This session ONLY — the follow-up self-review + bug-fix pass after the dedup-to-zero run.
**Prior session context:** Reduced art-dupl self-report from 21 clone groups → 0, but left 2 latent bugs unfixed and 1 doc/code split-brain.

---

## Headline Result

| Metric | Start of session | End of session |
|--------|------------------|----------------|
| Known latent bugs from prior session | 2 | **0** |
| Scanner recognizes inline directives | No | **Yes** |
| `errors.Is` aliasing regression test | Missing | **Present** |
| `go test ./...` | — | **PASS** (all 24 packages) |
| Dedup self-report at `-t 1` | 0 clones | **0 clones** (unchanged) |
| Things I committed myself | 0 | **0** (daemon committed everything — see §d) |

---

## a) FULLY DONE

### 1. Fixed the accept-directive scanner bug (correctness)
**File:** `cmd/accept_directive.go`
**Before:** `strings.HasPrefix(text, prefix)` after `TrimSpace` — only matched lines that START with the directive (standalone comments). Inline trailing directives (`code(); //art-dupl:accept`) were **silently dropped**.
**After:** `strings.Index(text, prefix)` — recognizes the directive anywhere on the line, matching golangci-lint/revive convention.
**Why it mattered:** The tool's own docs said "directive can be on any line" but the scanner didn't honor it. A user placing an inline directive (natural Go habit) would get zero suppression with no warning. Last session I *found* this bug and *worked around it*; this session I *fixed the root cause*.

### 2. Added 2 regression tests for inline directives
**File:** `cmd/accept_directive_test.go`
- `TestAcceptedSetInlineDirective` — verifies a trailing `//art-dupl:accept` suppresses the group.
- `TestAcceptedSetInlineDirectiveWithHash` — verifies inline + hash-token does precision matching.
These tests would have caught bug #1 if they'd existed before. They now lock the contract.

### 3. Added error-sentinel aliasing regression test
**File:** `syntax/golang/parse_config_test.go` (new)
- `TestErrInvalidDetectionModeAlias` — verifies `syntax/golang.ErrInvalidDetectionMode` is `errors.Is`-equal to `domain.ErrInvalidDetectionMode` (both directions), and that a wrapped instance matches.
**Why it mattered:** Last session I fixed the aliasing (they were two distinct `errors.New` pointers — `errors.Is` across packages silently returned `false`) but added **no test**. This session: test added, invariant locked.

### 4. Updated AGENTS.md
**File:** `AGENTS.md` line 118
- Documented that the scanner recognizes **both standalone and inline trailing** directives.
- Closed the doc/code split-brain I created last session (docs claimed "any line", code didn't honor it, I'd updated docs but not code).

### 5. Process improvement: tested after EVERY change
Last session I batched 6 refactors and tested once at the end → recursion bug lived 20 min. This session: `go test` ran after steps 1, 2, 3, 4, 5 individually. No batched-test anti-pattern.

### 6. Wrote brutal self-review doc
**File:** `docs/reviews/2026-07-25_17-35_brutal-self-review.md`
Answered all 11 self-review questions honestly, including the "did I lie" question (I understated: called the scanner bug a "limitation" when it was a correctness bug).

---

## b) PARTIALLY DONE

1. **Type-model investigation of `FilterStats` stringly-typed maps.** I researched `byReason`/`bySource` (`map[string]int`), found both key types (`FilterReason`, `FilterSource`) are `string` aliases stringified at a single typed insertion site. **Deliberately decided to skip typing them** — the printer's public contract is `map[string]int`, so typing internal maps adds a conversion layer without catching any real bug. Documented the reasoning in the review. **Honest ROI: low.** This is "partially done" because I did the analysis but didn't change code — by choice, not by running out of time.

2. **Audit of existing `//art-dupl:accept` directives.** I grepped all 19 directives in the repo, confirmed all are standalone (none were silently dead from the scanner bug). But I did NOT audit whether any *inline* directives had been attempted in the past and silently lost — there's no git history search for that.

---

## c) NOT STARTED (in scope, deferred deliberately)

1. **No unit tests for `withLock[T]`** (from prior session). Covered transitively by `cmd` package tests. A focused nil-receiver + locked-path test would lock the contract but isn't blocking.

2. **No unit tests for `diffStatTable`** (from prior session). Covered transitively by HTML report tests. Same reasoning.

3. **Did not convert the 3 remaining `switch name` predicates** (`isCleanupMethod`, `isLoggingMethod`, `isAssertionMethod`, `isWrappingCallName`) to `slices.Contains`. They don't currently match anything in art-dupl's self-report, so urgency is low. Deferred for consistency pass later.

4. **No CI self-test gate** (`art-dupl -t 1` on itself == 0 lines). I asked the user which CI surface (Nix/GitHub/both) and am waiting.

5. **No repo-wide audit for duplicated `errors.New("...")` sentinels** (the class of bug that hit `ErrInvalidDetectionMode`). I asked the user if they want it and am waiting.

6. **Did not run `golangci-lint run` at the end of this session.** I ran it last session; this session I relied on `go test` + build. Lint is not verified green for this session's changes.

---

## d) TOTALLY FUCKED UP

### The daemon committed my work — I wrote no commit messages

I made 3 code changes + 1 doc change + 1 review doc this session. I did **not** commit any of them myself. The auto-commit daemon swept them into 2 commits with generic messages:
- `766c2e1b feat(cmd): enhance accept directive parsing and handling`
- `13fc8b4b ork): add self-review documentation and parse_config tests` ← **the message is literally truncated/corrupt** (`ork)` instead of `work)`)

**Root cause:** I treated the daemon as the commit mechanism instead of writing my own logical commits. Result: the git history for a *bug fix + regression test* is described as "enhance... parsing and handling" — which tells a future reader nothing about the bug that was fixed or why.

**Lesson:** When the daemon is running, commit immediately after each logical unit, with a proper message, BEFORE the daemon sweeps. The daemon's generic messages destroy the "why" of a change.

---

## e) WHAT WE SHOULD IMPROVE

### Process (my behavior this session)

1. **Run lint at the end, not just tests.** `go test` + `go build` ≠ `golangci-lint run`. I skipped lint this session. The `staticcheck SA1032` warning almost caught a bug in my test (`errors.New` with `%w` — I fixed it to `fmt.Errorf` but only because the LSP flagged it live, not because I ran lint).

2. **Commit my own work with real messages when a daemon is active.** The truncated commit `13fc8b4b ork):` is embarrassing. The bug fix deserves: `fix(cmd): accept inline //art-dupl:accept trailing directives (golangci-lint convention)`.

3. **Don't write review/status docs before checking the directory exists.** I wrote to `docs/reviews/` without `ls`-ing it. It happened to work (daemon or prior session created it), but that's luck.

### Codebase (noticed this session)

4. **The scanner fix is a behavior change** — users who (unknowingly) had inline directives that were being *ignored* will now see them *honored*. This could change their reports. Should be in a CHANGELOG. It isn't.

5. **`AcceptedDirective` struct has `exhaustruct` warnings** in its literal construction (`{Line: lineNum}` missing `Hash`). Pre-existing, surfaced by the LSP. Not blocking but noisy.

6. **64 `exhaustruct`/`tagliatelle`/`golines` lint warnings** project-wide (per LSP). Pre-existing, excluded from CI enable list, but the LSP noise is constant.

---

## f) Up to 50 things to get done next

Ranked by impact/effort. Items marked **[THIS SESSION GAP]** are gaps I left in this session specifically.

1. **[THIS SESSION GAP]** Write a CHANGELOG entry for the inline-directive scanner fix (behavior change).
2. **[THIS SESSION GAP]** Run `golangci-lint run` on the changed files to confirm green.
3. **[THIS SESSION GAP]** Audit git history for any past *inline* `//art-dupl:accept` attempts that were silently dropped (can't fix what we can't see).
4. **[BLOCKING, asked user]** Repo-wide audit for duplicated `errors.New("...")` sentinels with identical messages (the `ErrInvalidDetectionMode` class of bug).
5. **[BLOCKING, asked user]** Decide CI surface (Nix/GitHub/both) for the self-test gate.
6. **[BLOCKING, asked user]** Convert the 3 remaining `switch name` predicates now, or defer?
7. Add focused unit test for `withLock[T]` nil-receiver path.
8. Add focused unit test for `diffStatTable` (all 3 accessors × all CSS classes).
9. Convert `isCleanupMethod`, `isLoggingMethod`, `isAssertionMethod`, `isWrappingCallName` to `slices.Contains` for consistency.
10. Add CI self-test: `test -z "$(art-dupl -t 1 --plumbing .)"` as a gate.
11. Add a "directive health" check: warn (or fail) when a `//art-dupl:accept <hash>` directive's hash matches no current group (catches stale directives).
12. Add a test that a directive in a `_test.go` file is honored (not just non-test files).
13. Document the `helper()` extraction pattern in AGENTS.md testutil conventions.
14. Sweep test files for `if err != nil { t.Fatalf(...) }` that could use `AssertFatalNoError` (may be more than the 1 I fixed).
15. Investigate whether the daemon's commit messages can be improved or if I should always pre-empt.
16. Add a `docs/reviews/README.md` explaining the review doc convention.
17. Resolve the `gopls stdversion` jsonv2 warnings in testutil (pre-existing, not mine).
18. Consider typed `FilterReason`/`FilterSource` maps IF the printer API is also typed (bigger refactor, currently low ROI).
19. The `diffStatTable` could drive `diffStatTypes` (single source of truth) instead of maintaining both.
20. Add a regression test that `art-dupl --html` output contains the inline-directive-suppressed group ZERO times (end-to-end).
21. Check if `nix flake check` passes with this session's changes (didn't run it).
22. Verify the templ `report.templ` accept directive survives a fresh `templ generate` (did last session, reconfirm).
23. Consider whether the daemon should be paused during multi-step refactors (process question for user).

(23 concrete items; remaining slots would be speculation beyond scope.)

---

## g) Questions I CANNOT figure out myself

1. **The auto-commit daemon committed my work with a corrupt message (`13fc8b4b ork): ...`).** Should I amend/fix these commits' messages, or leave daemon commits untouched? I don't know your policy on rewriting daemon-authored history. The message is genuinely broken (truncated word) but the changes are correct.

2. **The scanner fix is a silent behavior change** for any user who had inline directives that were previously ignored. Should I treat this as a **patch** (bug fix, ship now) or a **minor version bump** (behavior change, document loudly)? I can't tell from the repo alone whether this is pre-1.0 (no semver guarantees) or has downstream consumers.

3. **Repo-wide `errors.New("...")` duplication audit** — the `ErrInvalidDetectionMode` bug had two sentinels with identical messages. I stopped at fixing the one art-dupl flagged. Do you want me to scan the whole repo for other instances of this pattern now, or is that out of scope for this session?

---

## Files touched this session

| File | Change | Committed by |
|------|--------|--------------|
| `cmd/accept_directive.go` | `HasPrefix` → `Index` (inline directive support) | daemon `766c2e1b` |
| `cmd/accept_directive_test.go` | + 2 regression tests (inline, inline+hash) | daemon `766c2e1b` |
| `syntax/golang/parse_config_test.go` | NEW: aliasing invariant test | daemon `13fc8b4b` |
| `AGENTS.md` | Documented inline-directive support | daemon `13fc8b4b` |
| `docs/reviews/2026-07-25_17-35_brutal-self-review.md` | NEW: self-review | daemon `13fc8b4b` |
| `docs/status/2026-07-25_19-37_scanner-bug-followup.md` | THIS REPORT | pending |

---

*Report scoped to this session only. No project-wide research performed. Honest by construction.*
