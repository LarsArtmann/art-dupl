# Status Report: art-dupl — 2026-05-06 09:40

## Executive Summary

**Date:** 2026-05-06 09:40 AM CEST\
**Branch:** `fork`\
**Last Commit:** `a098597` — fix: SARIF output now uses real clone hashes instead of position strings

---

## Current Work Status

| Component                        | Status                | Notes                                                    |
| -------------------------------- | --------------------- | -------------------------------------------------------- |
| **SARIF Fix (this session)**     | ✅ **FULLY DONE**     | Real clone hashes, proper fingerprints, version config   |
| **Semantic Interface Detection** | 🔄 **PARTIALLY DONE** | `transform.go` and `parse.go` modified but not committed |
| **BDD Test Updates**             | 🔄 **PARTIALLY DONE** | Test updates for templ filtering behavior changes        |
| **Whitespace Cleanup**           | 🔄 **PARTIALLY DONE** | `config/config.go` alignment change                      |

---

## What Was Done This Session

### SARIF Output Fix ✅

**Problem:** SARIF `cloneHash` field contained position strings like `detection/detection_test.go:4705:4782` instead of actual clone hashes.

**Solution:**

1. SARIF printer now implements `HashSetter` interface to receive real clone hashes
2. Added `SARIFConfig` struct with `Version` field
3. Version now pulled from `cmd.GetVersion()` instead of hardcoded `"1.0.0"`
4. Renamed `cloneHash` to `contentFingerprint` per SARIF spec
5. Added `partialFingerprint` (first 8 chars) for approximate matching
6. Added comprehensive spec compliance tests

**Files Changed:**

- `printer/sarif.go` — Core fixes
- `printer/sarif_test.go` — New compliance tests
- `cmd/run_printer.go` — Version propagation
- `cmd/run_flags.go`, `cmd/run_all_modes.go`, `cmd/cmd_utils_test.go` — Support code

**Committed:** `a098597`

---

## What Is Partially Done / Not Committed

### 1. Semantic Interface Detection 🔄

**Files Modified:**

- `syntax/golang/transform.go` — Added `inInterface` flag to track interface context
- `syntax/golang/parse.go` — Added `inInterface bool` field to transformer

**Purpose:** When `--semantic` mode is enabled, function types in interfaces should be marked with semantic type encoding (e.g., `~interface~`).

**Changes:**

```go
// transform.go - In FuncType case:
case *ast.FuncType:
    if t.config.Mode.IsSemantic() && t.inInterface {
        o.Type = encodeSemanticType(FuncType, "~interface~", true)
    } else {
        o.Type = FuncType
    }

// In InterfaceType case:
case *ast.InterfaceType:
    o.Type = InterfaceType
    prev := t.inInterface
    t.inInterface = true
    o.AddChildren(t.trans(n.Methods))
    t.inInterface = prev
```

**Status:** Code compiles, tests pass, but changes are unstaged.

---

### 2. BDD Test Updates 🔄

**Files Modified:**

- `bdd/configuration_file_test.go`
- `bdd/default_filtering_test.go`
- `bdd/filter_features_test.go`
- `bdd/output_formats_and_filters_test.go`
- `bdd/plumbing_output_test.go`
- `bdd/stats_command_test.go`
- `bdd/stats_subcommand_test.go`
- `bdd/templ_clone_detection_test.go`

**Changes:** Tests updated to reflect that `*_templ.go` files are now filtered by default (per commit `45da310`).

**Status:** Tests pass, but changes are unstaged.

---

### 3. Test Data 🔄

**Path:** `testdata/interface_semantic/test.go`

**Content:** Test file for semantic interface detection feature.

**Status:** Untracked, not committed.

---

### 4. Whitespace Fix 🔄

**File:** `config/config.go`

**Change:** Alignment fix for `IncludeTempl` field (added space to align with other fields).

**Status:** Unstaged.

---

## What Is NOT Started

| Item                              | Priority | Notes                                                   |
| --------------------------------- | -------- | ------------------------------------------------------- |
| Column info in SARIF regions      | Low      | Would improve precision, requires `syntax.Node` changes |
| BDD test commits                  | High     | Changes ready to commit                                 |
| Semantic interface feature commit | High     | Code complete, needs review                             |
| testdata/ directory handling      | Medium   | Decide whether to commit or remove                      |

---

## What Is TOTALLY FUCKED UP

Nothing is "totally fucked up" — all tests pass and the codebase is in a working state.

**However, there is technical debt:**

- 8 BDD test files with unstaged changes
- Semantic interface feature partially implemented
- `testdata/` directory created but not decided on

---

## What We Should Improve

1. **Commit BDD test updates** — These are ready and tested
2. **Review and commit semantic interface feature** — Or decide to revert
3. **Decide on `testdata/` directory** — Commit or remove
4. **Add interface tracking** — Consider if `inInterface` flag needs cleanup
5. **Improve error messages** — Some error messages could be more helpful
6. **Consider SARIF validation** — Add SARIF schema validation tests
7. **Add more fingerprint tests** — Edge cases for partial fingerprints
8. **Documentation** — Update HOW_TO_USE.md with SARIF examples
9. **CI/CD improvements** — Add SARIF output validation in tests
10. **Consider GitHub Action integration** — Test SARIF uploads

---

## Top #25 Things to Get Done Next

1. **Commit BDD test changes** — Quick win, tests already pass
2. **Review semantic interface feature** — Test the feature works as expected
3. **Remove or commit `testdata/`** — Clean up untracked files
4. **Fix config/config.go whitespace** — Quick alignment fix
5. **Add `--semantic` interface test** — Verify the new feature works
6. **Update SARIF documentation** — Document the new fingerprint fields
7. **Add SARIF schema validation** — Use official SARIF JSON schema
8. **Consider `contentFingerprint` hash format** — Should it be hex, base64, etc.?
9. **Add partialFingerprint algorithm documentation** — Document why first 8 chars
10. **Run full BDD test suite** — Ensure no regressions
11. **Add GitHub SARIF upload test** — Validate SARIF works with GitHub
12. **Review clone classification** — Is it still accurate with semantic mode?
13. **Consider performance** — Does semantic mode add overhead?
14. **Add CLI completion for `--semantic`** — If not already there
15. **Document version behavior** — Explain "dev" vs actual version
16. **Add integration test for SARIF** — End-to-end SARIF test
17. **Review error handling** — Are there unhandled error paths?
18. **Check cross-platform builds** — Does SARIF work on Windows?
19. **Add more output format tests** — JSON, HTML, etc.
20. **Review threshold behavior** — Is it consistent across formats?
21. **Add profiling markers** — For performance analysis
22. **Consider caching** — Can we cache parsed ASTs?
23. **Review naming conventions** — Are field names consistent?
24. **Add more benchmark tests** — Measure semantic vs non-semantic
25. **Write migration guide** — If SARIF format changes in future

---

## My Top #1 Question I Cannot Figure Out

**Question:** Should the `inInterface` flag for semantic interface detection be part of the `transformer` struct, or should it be passed through the AST traversal differently?

**Context:**

- Currently, `inInterface` is a field on the `transformer` struct
- When entering an interface, we set `inInterface = true`
- When leaving, we restore the previous value
- This works, but it modifies state during traversal

**Alternatives considered:**

1. Pass context through function parameters
2. Use a stack for nested interfaces
3. Pre-process to mark interfaces before semantic encoding

**Why I can't decide:** The current approach is simple and works, but it violates the principle of pure functions. However, the AST traversal inherently has state, so this might be acceptable. The question is whether there's a cleaner pattern for this kind of context-sensitive encoding.

---

## Git Status

```
On branch fork
Your branch is up to date with 'origin/fork'.

Changes not staged for commit:
  bdd/configuration_file_test.go
  bdd/default_filtering_test.go
  bdd/filter_features_test.go
  bdd/output_formats_and_filters_test.go
  bdd/plumbing_output_test.go
  bdd/stats_command_test.go
  bdd/stats_subcommand_test.go
  bdd/templ_clone_detection_test.go
  config/config.go
  syntax/golang/parse.go
  syntax/golang/transform.go

Untracked files:
  testdata/
```

---

## Test Status

All tests pass:

```
ok  github.com/LarsArtmann/art-dupl/bdd         2.748s
ok  github.com/LarsArtmann/art-dupl/printer    (cached)
ok  github.com/LarsArtmann/art-dupl/syntax/golang  0.030s
... (all other tests pass)
```

---

## Recommendations

1. **Immediate:** Commit BDD test updates (low risk, ready to ship)
2. **This week:** Review and commit/revert semantic interface feature
3. **This week:** Clean up `testdata/` directory
4. **Next:** Add SARIF schema validation
5. **Next:** Update documentation

---

**Report Generated:** 2026-05-06 09:40 AM CEST\
**By:** Crush (AI Assistant)
