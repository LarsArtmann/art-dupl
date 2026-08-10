# Status: 2026-08-05 06:03 — Bool Guard & Format Specifier Hardening

> **Post-session annotation (2026-08-05):** Both TODO items fully completed and removed from TODO_LIST. Recorded in CHANGELOG `[Unreleased]`. Remaining section C items (calibrate confidence, property engine labels) are in TODO_LIST. The AGENTS.md/ACTIONABILITY_PATTERNS.md bool-guard description drift (section E items 6-7) is a known minor gap. Section F is a brainstorm — key items harvested into TODO_LIST.

## Context

Two TODO items from `TODO_LIST.md` (Property Engine Follow-up section):

1. Broaden `findOkVarName` beyond the literal `"ok"`.
2. Fix `isFormatSpecifierDifference` edge cases (`%%`, width specifiers, arg indices).

Both functions live in `printer/actionability/` and feed the extractability/actionability engine.

---

## a) FULLY DONE

### 1. `findOkVarName` broadened (`actionability_boilerplate.go`)

- **Before**: Hardcoded `child.Name == "ok"` — only matched the literal `"ok"`.
- **After**: New `isBoolGuardVarName(name string) bool` helper using a type switch over `ok`, `found`, `exists`, `success`, `present`.
- `findOkVarName` delegates to `isBoolGuardVarName` — no behavior change for existing `"ok"` clones, just accepts additional conventional bool-guard names.
- Chosen as a **function** (not a global `var`) to satisfy `gochecknoglobals` lint rule.

### 2. `isFormatSpecifierDifference` rewritten (`extractability_engine.go`)

- **Before**: Naive char-after-`%` comparison — broke on `%%` (treated second `%` as a verb), width/precision differences (`%5d` vs `%3d` collapsed to `%d`), and argument indices (`%[1]d`).
- **After**: Full Go fmt verb parser:
  - `extractFormatSpecifiers(s)` — scans a string, returns ordered list of specifier tokens.
  - `scanFormatSpecifier(s, start)` — parses one verb, skips `%%`.
  - `scanFmtBody(s, i)` → delegates to `scanArgIndex`, `scanDigits` — keeps gocyclo under 15.
  - `isFmtFlag(c byte)` — flag set check.
  - `isFormatSpecifierDifference(a, b)` now compares extracted specifier **lists** (count + element-wise), so different verb counts, widths, precision, flags, and indices all detected.
- Complexity distributed across 5 small functions to satisfy gocyclo and readability.

### 3. Unit tests added

- `actionability_boolguard_test.go` — `findOkVarName` (8 cases), `isAssignWithBoolGuard` integration tests (all 5 names + 1 rejection).
- `extractability_format_test.go` — `isFormatSpecifierDifference` (20 cases: `%%`, width, precision, flags, arg indices, mixed), `extractFormatSpecifiers` (9 parser cases).

### 4. Verification

- `go test ./...` — all packages pass.
- `go test -race ./printer/actionability/...` — passes.
- `golangci-lint run ./printer/actionability/...` — 0 issues.
- Lint on full project: 51 pre-existing `tagliatelle`/`godox` issues in files NOT touched (documented as disabled in AGENTS.md).

### 5. Docs

- `TODO_LIST.md`: Both items removed, last-updated date bumped to 2026-08-05.

---

## b) PARTIALLY DONE

Nothing — both items were small and fully completed.

---

## c) NOT STARTED

From the same TODO section (Property Engine Follow-up), still open:

- Calibrate confidence values against real-world data.
- Add property engine labels to `--list-patterns`.

---

## d) TOTALLY FUCKED UP

Nothing. All changes compile, lint clean, tests green (including race detector).

---

## e) WHAT WE SHOULD IMPROVE

### Self-critique of this session's work

1. **`isBoolGuardVarName` is a denylist, not a type-based heuristic.** It matches by name only. A variable named `found` that holds a `string` (not a bool) would still be treated as a bool guard. The `isBoolGuardIf` function has no type information to verify the variable is actually boolean. The `--type-aware` feature (`Node.VarType`) could be leveraged here when available to validate the guard variable is actually `bool`.

2. **`boolGuardVarNames` set is arbitrary.** I picked `ok`, `found`, `exists`, `success`, `present` based on "common Go conventions" — but I didn't measure against real codebases. Other common names: `hasAccess`, `isSet`, `isValid`, `canProceed`. The list is opinionated and incomplete by nature. A more robust approach: match any variable whose name starts with a bool-suggesting prefix (`is`, `has`, `can`) OR is in the explicit list.

3. **Format specifier parser is hand-rolled, not using `go/printer` or `fmt` internals.** It's a correct subset of the Go fmt grammar, but edge cases like `%[2]*.[1]*f` (reordered width/precision via index) are not handled. The parser could also validate verb characters (only `[vTbtcdoqxXgGeEfFsUcp]` are valid) instead of accepting any trailing byte.

4. **No test exercises `isFormatSpecifierDifference` through the FULL extractability pipeline.** The unit tests call the function directly. An integration test that builds a `CloneNode` tree with `BasicLit` children, runs `EvaluateExtractability`, and checks the `Parameterizable`/`Reason` fields would catch wiring regressions.

5. **`hasFormatSpecifierDifferences` caller still has a subtle limitation.** It compares `literals[0][pos]` against `literals[i][pos]` — it uses clone-0 as the reference. If clone-1 and clone-2 differ in specifiers but both differ from clone-0, this still works (returns true). But the comparison is pairwise against index 0 only, not all pairs. For the current 2-clone minimum this is fine; for 3+ clones it could miss some patterns (though the "any difference" semantics makes it unlikely).

6. **AGENTS.md actionability pattern count says "23 patterns"** — the bool-guard pattern description lists specific names. Now that `findOkVarName` accepts more names, the AGENTS.md description should be updated to reflect the broadened set. I did NOT update AGENTS.md.

7. **`docs/ACTIONABILITY_PATTERNS.md` was not updated** — it likely documents the bool-guard pattern with the old `"ok"`-only description. This is a documentation drift risk.

### Process observations

8. **First lint pass had 5 issues** (global var, gocyclo 18, wsl whitespace, nonamedreturns). I fixed them reactively instead of anticipating them. Knowing the lint config (gochecknoglobals, gocyclo ≤15, wsl_v5, nonamedreturns), I should write lint-clean code on the first pass.

---

## f) Up to 50 things to do next

### Property Engine Follow-up (HIGH)

1. Calibrate confidence values against 3-5 real OSS Go projects — measure FP/FN rates.
2. Add property engine labels to `--list-patterns` output or document as internal-only.
3. Integration test: `EvaluateExtractability` with format-specifier-differing `CloneNode` trees.
4. Integration test: `isAssignWithBoolGuard` with real DiscordSync-style AST from `syntax.Serialize`.

### Actionability Pattern Hardening

5. Update AGENTS.md bool-guard description to list all accepted names.
6. Update `docs/ACTIONABILITY_PATTERNS.md` bool-guard section with broadened name set.
7. Consider type-aware bool-guard validation (leverage `Node.VarType` from `--type-aware`).
8. Consider prefix-based bool name matching (`is*`, `has*`, `can*`) instead of fixed denylist.
9. Add valid-verb-character validation to `scanFormatSpecifier` (only `vTbtcdoqxXgGeEfFsUcp`).
10. Add test for reordered width/precision via arg index (`%[2]*.[1]*f`).
11. Test `hasFormatSpecifierDifferences` with 3+ clones to verify all-pairs semantics.

### Missing Test Coverage (actionability package)

12. `isBoolGuardIf` — no direct unit test (only tested via `isAssignWithBoolGuard`).
13. `isNotOkExpr` — no unit test.
14. `isReturnOnlyBody` — no unit test.
15. `literalsDifferOnlyInValues` — no unit test.
16. `sameLiteralCount` — no unit test.
17. `collectStringLiterals` — no unit test.
18. `hasFormatSpecifierDifferences` — no unit test (only the inner function is tested now).

### Extractability Engine Gaps

19. `confidenceHigh`/`confidenceMedium`/`confidenceLower` constants are hardcoded — should be configurable or at least documented with rationale.
20. `helperDominanceRatio = 0.6` — undocumented magic number, needs empirical validation.
21. `EvaluateExtractability` has no golden-file / snapshot tests for complex clone groups.

### General Code Quality

22. Pre-existing `tagliatelle` lint (50 issues) — documented as disabled, but json tags are inconsistent (snake_case in some structs, camelCase in others). ADR-0016 documents this but it's unresolved.
23. Pre-existing `godox` lint (1 issue) — a TODO/FIXME in production code somewhere.
24. `extractability_bench_test.go` has 2 `b.N` → `b.Loop()` modernization warnings (pre-existing, not mine).
25. Run `nix flake check` to verify reproducible CI still passes (didn't run this session).

### Documentation

26. `CHANGELOG.md` `[Unreleased]` section — add entries for the two completed fixes.
27. `docs/adr/0017-property-based-classification.md` — was modified at conversation start (git status), verify consistency with code changes.
28. `FEATURES.md` — verify bool-guard / format-specifier features are accurately described.

---

## g) Questions I CANNOT figure out myself

1. **Should the bool-guard name set be configurable (CLI flag / config) or stay hardcoded?** Some teams may use project-specific naming conventions (e.g., `hasPerm`, `isReady`) that shouldn't require a code change. A `--bool-guard-names` flag or `Config.BoolGuardNames` would solve this, but it adds config surface area. What's the preference?

2. **Should `isFormatSpecifierDifference` validate verb characters?** Currently `%Z` (invalid verb) is treated as a valid specifier and compared. Adding validation would make it stricter but could reject real-world strings that happen to contain `%` followed by non-verb chars (e.g., `"50% off"` — though that's technically malformed Go format usage). Strict validation or keep it permissive?

3. **Should I update AGENTS.md and `docs/ACTIONABILITY_PATTERNS.md` now?** These describe the bool-guard pattern with the old `"ok"`-only description. I noticed the drift but didn't fix it since you said "DO NOT RESEARCH OTHER STUFF UNRELATED." Is updating these docs in scope for a follow-up, or do you want me to do it now?

---

## Resolution (2026-08-10)

**Core work shipped.** Bool-guard name broadening and format specifier parser rewrite are in CHANGELOG `[Unreleased]` → Changed. Section c items (calibrate confidence, property engine labels) — labels are in `--list-patterns`; calibration needs real-world data → ROADMAP.
