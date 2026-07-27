# Status Report: Templ Rendering Idiom — Verification & Hardening

**Date:** 2026-07-27 21:12
**Session scope:** Verify the `isTemplRenderingIdiom` fix, trace real pipeline behavior, close documentation gaps
**Previous session:** Fixed the compile error (`undefined: isTemplRenderingIdiom`); this session verified it works

---

## a) FULLY DONE (Verified This Session)

### Build & CI

- [x] `go build ./...` — exits 0 (28 packages)
- [x] `go test ./... -count=1` — all 28 packages PASS (no cache), including `bdd/`
- [x] `go test -race ./printer/actionability/` — PASS
- [x] `golangci-lint run ./printer/actionability/` — 0 issues
- [x] `go vet ./printer/actionability/ ./printer/` — clean
- [x] `nix build .#art-dupl` — **PASS (exit 0)** — the original failing buildflow command is resolved
- [x] BDD tests pass (`go test ./bdd/ -count=1`)

### End-to-End Verification

- [x] **Wrote `TestIsTemplRenderingIdiom_RealGoSource`** — parses actual Go source through `golang.Parse()`, converts to `domain.CloneNode`, confirms `isTemplRenderingIdiom` returns true. **The pattern is NOT dead code.** It fires on real `.go` files.
- [x] **Wrote `TestIsTemplRenderingIdiom_TemplSourceDoesNotMatch`** — parses actual `.templ` source through `templ.ParseBytes()`, confirms the pattern does NOT fire on `.templ` files (because templ uses `ComponentIfStatement`=11, not `golang.IfStmt`=31, and drops the condition expression entirely). Documents this as intentional behavior.

### Documentation Fixes

- [x] `docs/ACTIONABILITY_PATTERNS.md` — Added missing table rows for `bool-guard` and `templ-rendering-idiom` (were bare one-liners while other 20 patterns had full rows). Fixed "20 pattern checks" → "22 pattern checks".
- [x] `AGENTS.md` — Clarified the `templ-rendering-idiom` scope: fires on `.go` files (incl. generated `_templ.go`), NOT on `.templ` source, with the technical reason (node type mismatch + dropped condition).

### Test Hardening

- [x] `actionability_patterns_disabled_test.go` — Replaced hardcoded `!= 22` with `len(actionabilityPatternTable)` (prevents future table/function drift). Added `PatternBoolGuard` + `PatternTemplRenderingIdiom` to the required-labels list (were missing from `TestAllActionabilityPatterns_ContainsLabels`).

---

## b) PARTIALLY DONE

### End-to-End Coverage of the Pattern

The e2e test uses hand-written Go source that mimics the templ empty-state idiom. It does NOT test with:

- Actual **generated `_templ.go`** files (the most realistic scenario where the pattern should fire)
- The `--include-generated templ` CLI flag path
- A full `art-dupl` binary run on a directory containing such clones

The unit-level proof is solid (real AST → real CloneNode → pattern matches), but the integration path through the full CLI pipeline is untested.

---

## c) NOT STARTED

1. **No BDD test** for the templ-rendering-idiom pattern (the Ginkgo suite in `bdd/` has no scenario for it)
2. **No `nix flake check`** — only `nix build` was run (the original failing command). `nix flake check` runs additional CI gates including templ generate.
3. **No verification against the real `printer/report.templ`** — this file exists in the repo and contains `if len(diffData.Others) == 0 { for ... } else { ... }` at line 88, but it's a variant (len==0 → iterate; else → diff view) that does NOT match the pattern's shape (len==0 → text; else → range). Did not test whether running art-dupl on its own source with `--include-generated templ` surfaces or suppresses this.

---

## d) TOTALLY FUCKED UP

### Nothing catastrophic, but one honest miss:

**I didn't catch the stale LSP diagnostics early enough.** Throughout the session, the LSP reported `undefined: isTemplRenderingIdiom` and 4x `wsl_v5` warnings even after the build passed and lint was clean. I spent tool calls restarting the LSP and re-verifying instead of recognizing immediately that the LSP cache was stale from the previous session's broken state. This wasted ~2 tool calls. The actual `golangci-lint` CLI was the source of truth and it reported 0 issues from the start.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture / Design Issues

1. **The `templ-rendering-idiom` pattern has a scope gap.** It's named for templ but only fires on `.go` files. The actual `.templ` source uses `ComponentIfStatement` which drops the condition expression entirely. This means:
   - The pattern will NEVER fire on `.templ` clones detected by the templ detector
   - It WILL fire on generated `_templ.go` files IF included via `--include-generated templ`
   - But generated `_templ.go` is filtered by DEFAULT, so the pattern effectively only fires on hand-written Go that coincidentally has this shape
   - **The name is misleading.** `templ-rendering-idiom` implies it handles templ source. It should either be renamed or extended to handle `ComponentIfStatement`.

2. **The templ transform drops too much information.** `transformIfExpression` in `syntax/templ/transform_expressions.go` only adds body children — the condition (`ie.Expression`) is never encoded into the node. This means NO actionability pattern can check templ if-conditions. This is a deeper architectural limitation than just this pattern.

3. **Test helper duplication.** I had to write `cloneNodeFromSyntax` in the test file because `printer.syntaxToCloneNode` would create a circular import (printer imports actionability). This conversion logic is duplicated. It could live in `domain` or a shared internal package.

### Process Improvements

4. **Always run `nix build` FIRST.** It was the original failing command and the last to be verified. Should have been the first thing checked in this session.

5. **LSP diagnostics can be stale.** When the LSP reports errors but `go build` passes, trust the compiler. Restart the LSP once, then ignore.

---

## f) Up to 50 Things We Should Get Done Next

### High Priority (Verifies the pattern actually works in production)

1. Run `nix flake check` (full CI gate, includes templ generate)
2. Run `art-dupl` on its own source with `--include-generated templ --semantic` and check if the `if len(...) == 0` pattern in `report_templ.go` is detected/suppressed
3. Add a BDD test in `bdd/` that creates two `.go` files with the empty-state idiom, runs detection, and asserts suppression with label `templ-rendering-idiom`
4. Add an integration test that runs the full CLI pipeline (not just unit-level `EvaluateActionabilityWithLabel`) on generated `_templ.go` files
5. Verify the pattern fires when `--include-generated templ` is passed (the only realistic path where generated templ Go code reaches the actionability layer)

### Medium Priority (Correctness & Architecture)

6. Extend `syntax/templ/transform_expressions.go::transformIfExpression` to encode the condition expression into the `ComponentIfStatement` node (as a child or via `encodeSemanticType`), so actionability patterns can inspect templ if-conditions
7. OR: Add a templ-specific variant of the pattern that matches `ComponentIfStatement` with a `ComponentForStatement` child (structural, without condition check)
8. Rename the pattern from `templ-rendering-idiom` to `empty-state-rendering` or `len-zero-else-range` (more honest about what it matches)
9. Extract `cloneNodeFromSyntax` to `domain` or `internal/testutil` to avoid duplication between test and production code
10. Check if any OTHER actionability patterns have the same `.templ` scope gap (they all use `golang.*` node types)
11. Audit all 22 patterns for whether they work on templ-detected clones (likely many don't, since templ uses different node types)
12. Add a test that asserts `AllActionabilityPatterns()` length matches the priority-order list in `docs/ACTIONABILITY_PATTERNS.md` (docs/code sync)

### Low Priority (Polish)

13. The `docs/ACTIONABILITY_PATTERNS.md` "Pattern Priority Order" section (lines 44-65) lists 22 items but doesn't note which are `.go`-only vs cross-language
14. Add a column to the patterns table: "Applies to: `.go` / `.templ` / both"
15. The `findNodesByBaseType` test helper could be promoted to a test utility package for reuse
16. Run `nix flake update` to check if the `vendorHash` staleness warning at `flake.nix:53` is resolved by a dependency bump
17. Add `//art-dupl:accept` directive to `printer/report.templ` line 88 if the pattern is extended to templ and causes false positives there
18. Consider whether the pattern should also match the inverse shape: `if len(x) > 0 { for ... }` (without else) — common variant
19. Consider matching `== 0` AND `<= 0` AND `< 1` as empty-check conditions
20. Document in ADR format the decision to keep the pattern `.go`-only (current behavior) vs extending to `.templ` source

### Broader Codebase Health (noticed during this session)

21. The `printer/json.go` has `stdversion` warnings (uses `json.Marshal` from go1.27 while `go.mod` says 1.26) — pre-existing, not my change
22. `extractability_bench_test.go` has `b.N` → `b.Loop()` modernization hints (gopls) — pre-existing
23. The auto-commit daemon committed my changes mid-session (commit `18a575ea`, `bf3f2902`) — this is expected behavior per AGENTS.md but means I can't create a clean single commit for review
24. The previous session's status report (`docs/status/2026-07-27_20-55_templ-rendering-idiom-fix.md`) should be updated with a resolution note now that verification is complete
25. The `TestNoDuplicateErrorNewMessages` self-maintaining AST scanner pattern is excellent — consider a similar self-maintaining test for pattern table/docs sync

---

## g) Questions (Cannot Figure Out Myself)

### 1. Should the pattern be extended to fire on `.templ` source?

Currently `templ-rendering-idiom` only fires on `.go` files. The templ transform (`transformIfExpression`) drops the condition expression, so `ComponentIfStatement` has no `len()` call to check. Options:

- **A:** Leave as-is (`.go`-only). Rename to avoid confusion.
- **B:** Extend `transformIfExpression` to encode conditions (affects ALL templ if-statements, bigger change).
- **C:** Add a second pattern variant matching `ComponentIfStatement + ComponentForStatement` structurally (no condition check, higher false-positive risk).

This is a design decision with real tradeoffs — your call.

### 2. Should I update the previous status report or leave it as a historical artifact?

`docs/status/2026-07-27_20-55_templ-rendering-idiom-fix.md` documents the investigation with open questions. Now that verification is complete, should I:

- **A:** Append a resolution section to that file (non-destructive annotation)
- **B:** Leave it as-is (historical point-in-time snapshot)
- **C:** Mark it as resolved and point readers to this report instead

### 3. Is the `false-positive on non-templ Go code` acceptable?

The pattern matches ANY Go code with `if len(x) == 0 { ... } else { for ... }`, not just templ-generated code. This means legitimate Go code with this shape would be suppressed as "non-actionable." Is that the intended behavior, or should the pattern be gated to only fire when the file is known to be templ-generated (e.g., filename ends in `_templ.go`)?

---

## Test Results Summary

```
go build ./...                          PASS
go test ./... -count=1                  28/28 packages PASS
go test -race ./printer/actionability/  PASS
golangci-lint ./printer/actionability/  0 issues
go vet ./printer/actionability/         clean
nix build .#art-dupl                    PASS (exit 0)
BDD tests (./bdd/)                      PASS
```

## Git State

```
Branch: fork (34 commits ahead of origin)
Working tree: clean (auto-committed)
Key commits this session:
  bf3f2902 chore(printer): update agent documentation and actionability e2e tests
  18a575ea test(actionability): add comprehensive tests and documentation for actionability patterns
```
