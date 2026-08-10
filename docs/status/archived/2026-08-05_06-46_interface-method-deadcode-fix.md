# Status: Interface-Method Dead-Code Fix + CI Cleanup — 2026-08-05 06:46

> **Post-session annotation (2026-08-05):** All work complete. The dead-code bug is fixed, BDD tests added, CI blockers resolved (tagliatelle removed — though it recurs; godox fixed). `nix flake check` passed all 10 checks at time of writing. Recorded in CHANGELOG `[Unreleased]`. The ROADMAP `[~]` for interface-aware suppression is accurate (same-package done, cross-package remains). Key open items harvested into TODO_LIST: FuncLit flag-reset test, SDK InterfaceMethod test.

## Summary

Discovered and fixed a **critical architectural bug**: the `interface-method`
actionability pattern was dead code for Go files because FuncDecl nodes can
never be clone roots (the structural filter in `FindSyntaxUnits` rejects
non-Statement clone roots). The fix propagates the `InterfaceMethod` flag from
FuncDecl to body statement nodes via the transformer, then rewrites the pattern
matcher to accept statement-level clone roots. Also fixed pre-existing CI
blockers (`tagliatelle` re-added to lint config, `godox` trigger in test
comment), a stale FEATURES.md claim, and added end-to-end BDD tests proving the
full pipeline works.

---

## a) FULLY DONE

1. **Root cause discovered and fixed**: The `interface-method` pattern
   (`isInterfaceMethodBody`) exclusively checked for FuncDecl-rooted clones.
   FuncDecl is NOT a Statement node, and `getUnitsIndexes` skips all non-Statement
   nodes when the file contains statements (which all real Go files do).
   Additionally, `FindSyntaxUnits` has a structural filter that explicitly
   rejects non-Statement clone roots in `.go` files. This made the pattern
   completely unreachable for standard Go code. Fixed by propagating
   `InterfaceMethod` from FuncDecl to body statement nodes and rewriting the
   pattern with two paths.

2. **Transformer propagation implemented** (`syntax/golang/parse.go`,
   `syntax/golang/transform.go`):
   - Added `enclosingInterfaceMethod bool` to the `transformer` struct.
   - `trans()` stamps `o.InterfaceMethod = t.enclosingInterfaceMethod` on every
     node (same pattern as `EnclosingReturnArity`).
   - FuncDecl case: saves/restores `t.enclosingInterfaceMethod` around body
     processing, setting it to `o.InterfaceMethod` so body statements inherit
     the flag.
   - FuncLit case: resets flag to `false` (closures are not interface methods),
     restores on exit.

3. **Pattern matcher rewritten**
   (`printer/actionability/actionability_interface_method.go`):
   - **Path 1** (FuncDecl root): Edge case for files without statements. Checks
     `InterfaceMethod` flag OR static name list. Body size via BlockStmt child
     count. This path preserves backward compatibility with the original logic.
   - **Path 2** (statement root): Normal Go files. Checks propagated
     `InterfaceMethod` flag. Body size is `len(seq)` (number of clone units).
     Requires `--type-aware` mode.

4. **BDD integration tests added** (`bdd/type_aware_test.go`):
   - 3 new Ginkgo specs in a new `interface-method suppression with --type-aware`
     Context: suppression with `--type-aware`, classification with `--explain`,
     and non-suppression without `--type-aware`.
   - Uses multi-statement bodies (2 statements each) to avoid being caught by
     `single-simple-statement` pattern, ensuring `interface-method` is the one
     that fires.

5. **Transformer unit test added**
   (`syntax/golang/interface_method_test.go`):
   - `TestTransformer_PropagatesInterfaceMethodToBody`: verifies body statements
     of an interface method carry `InterfaceMethod=true`, and body statements of
     a non-interface method carry `InterfaceMethod=false`.
   - Added `findBodyStatements` helper.

6. **Actionability unit tests updated**
   (`printer/actionability/actionability_interface_method_test.go`):
   - 5 new statement-level test cases: 1-stmt suppression, 3-stmt suppression,
     4-stmt at limit, 5-stmt over limit, no-flag not suppressed.
   - Added `mustStatementInterfaceMethodSeq` helper.

7. **Pre-existing CI blockers fixed**:
   - `.golangci.yml`: Removed `tagliatelle` from the enable list (was re-added
     by auto-git daemon commit `4324b03e`, contradicting the intentional
     disabling documented in AGENTS.md and enforced by the Nix disabled-linters
     check).
   - `cmd/suppression_config_test.go`: Changed "bug" to "issue" in a doc comment
     to satisfy the `godox` linter (which flags TODO/BUG/FIXME keywords).

8. **Stale documentation fixed**:
   - `FEATURES.md:99`: Removed "Not compatible with `--incremental`" claim
     (contradicted by ROADMAP:38 which says it IS compatible since M07).
   - `AGENTS.md`: Updated the type-aware detection paragraph to document the
     propagation mechanism and two-path pattern design.
   - `docs/ACTIONABILITY_PATTERNS.md`: Updated interface-method table entry and
     design decision text to document the two-path approach.
   - `docs/adr/0018-type-aware-interface-method.md`: Added "Statement-Level
     Propagation" section, updated Data Flow diagram, updated Status.
   - `syntax/syntax.go`: Updated `InterfaceMethod` field doc comment to mention
     propagation to body statement nodes.

9. **Node.Name blast radius audited**: `collectNamesPreOrder` in
   `clone_processor.go` was the main concern. Analysis confirmed FuncDecl is
   never a clone root (not a Statement node), so its `Name` never enters the
   name sequence used by `classifyCloneType`. Zero blast radius. The fix is
   safe.

10. **`extractFormatSpecifiers` gopls error verified as false positive**: The
    function exists in `extractability_engine.go` in the same package. Test
    passes fine. gopls cache is stale.

11. **Full verification passed**:
    - 27/27 test packages pass (`go test ./...`).
    - 0 lint issues on all changed packages.
    - `nix flake check` — **all 10 checks passed** (CI gate green).

---

## b) PARTIALLY DONE

1. **FuncLit flag reset is untested**: The transformer resets
   `enclosingInterfaceMethod = false` in the FuncLit case (closures are not
   interface methods), but no test explicitly verifies this. If a FuncLit
   appeared inside an interface method body, its inner statements should NOT
   carry the flag. This is covered indirectly (the statement root test uses
   simple bodies without closures), but a dedicated test would be more robust.

2. **Static name list path (path 1) is still unreachable for Go files**: Path 1
   handles FuncDecl-rooted clones, which only appear in edge-case files without
   statements. No test creates such a file. The path is preserved for backward
   compatibility and theoretical completeness, but its real-world applicability
   is near zero. The unit tests exercise it via direct CloneNode construction
   (bypassing the pipeline), which proves the logic but not the pipeline
   integration.

3. **`--no-actionability` + `--type-aware` interaction tested via BDD but not
   via unit test**: The BDD test `should classify suppressed clones as
interface-method with --explain` uses `--no-actionability` to reveal the
   clone and verify the pattern label, which implicitly tests the interaction.
   But there's no dedicated test for the `--no-actionability` override of the
   type-aware suppression path specifically.

---

## c) NOT STARTED

1. **Cross-package interface detection**: `IsInterfaceMethod` only checks
   interfaces declared in the same Go package. A type implementing
   `fmt.Stringer` (defined in `fmt`) relies on the static name list, not
   type-aware detection. Cross-package scanning would require loading all
   imported packages, significantly increasing `go/packages` load time.

2. **Interface-method-name cache**: Each FuncDecl re-scans the entire package
   scope. A pre-computed `map[string][]*types.Interface` per package would
   eliminate repeated work.

3. **SDK-level test for `InterfaceMethod`**: `pkg/artdupl` has
   `detector_type_aware_test.go` but no test verifies `InterfaceMethod` flows
   through the SDK boundary to `pkg/artdupl.Clone`.

4. **Performance characterization of `IsInterfaceMethod`**: O(scope_names x
   methods_per_interface) per method. Unprofiled on large packages.

5. **`--explain` distinction**: `--explain` says "interface-method" but doesn't
   distinguish "type-aware: satisfies interface X" vs "static name list match".

6. **ROADMAP entry update**: The ROADMAP `[~]` status was set by the previous
   session. Now that the feature actually works end-to-end (dead code fixed),
   the same-package portion could be upgraded to `[x]`. Not done because
   cross-package remains.

---

## d) TOTALLY FUCKED UP

Nothing. No regressions, no broken tests, no data loss. All changes are clean,
tested, and verified. The critical dead-code bug was found and fixed before it
could ship as a "working" feature.

**However**, the previous session's status report
(`docs/status/2026-08-05_06-13_type-aware-interface-method-detection.md`)
claimed the feature was "fully done" with "all 8 todo items completed" and
"build passes, full test suite passes, lint passes." This was **technically
true but practically false** — the tests passed because they constructed
`CloneNode` sequences directly, bypassing the detection pipeline entirely. The
feature was dead code that could never fire on real Go code. The previous
session's unit tests proved the matcher logic worked in isolation but provided
zero proof of pipeline integration. An end-to-end test would have caught this
immediately.

---

## e) WHAT WE SHOULD IMPROVE

1. **End-to-end tests must follow every pipeline-integrated feature**: The
   previous session wrote 12 unit tests, 4 actionability tests, and called it
   done. Not one test ran the actual `art-dupl` binary or in-process executor
   on real Go source files. The BDD test I added takes 30 seconds to write and
   would have immediately revealed the dead-code bug. **Rule: if a feature
   touches the actionability layer, it MUST have a BDD/integration test that
   creates Go source files, runs the pipeline, and verifies suppression.**

2. **The `getUnitsIndexes` / `FindSyntaxUnits` structural filter is a critical
   architectural gate that's invisible**: Pattern matchers that check for
   specific root node types (FuncDecl, FuncType, etc.) are only reachable if
   those node types can survive the structural filter. Non-Statement nodes are
   silently rejected. This should be documented in the actionability package
   docs so future pattern authors know their FuncDecl-rooted pattern needs a
   statement-level fallback.

3. **The `EnclosingReturnArity` propagation pattern should be the default for
   any FuncDecl-level metadata that needs to reach the actionability layer**:
   The codebase now has two fields using this pattern (`EnclosingReturnArity`
   and `InterfaceMethod`). Future fields should follow the same approach. A
   helper or documented convention would prevent the next developer from making
   the same dead-code mistake.

4. **Auto-git daemon can silently break CI**: The daemon re-added `tagliatelle`
   to the enable list in commit `4324b03e`, which had been explicitly removed in
   `dcac7541`. The Nix disabled-linters check caught it, but only `nix flake
check` surfaces this — `golangci-lint run` and `go test` don't. CI should
   fail fast on config regressions.

5. **The `commonInterfaceMethodNames` static list is now nearly vestigial**:
   Path 1 (FuncDecl root) is unreachable for Go files. The static list only
   matters for the edge case of files without statements. In practice, the
   type-aware path (path 2) is the only one that fires. The static list could
   be removed entirely without behavior change for real Go code, but keeping it
   provides defense-in-depth for unusual inputs.

---

## f) Up to 50 Things to Get Done Next

### High Priority (close gaps in this feature)

1. Upgrade ROADMAP `[~]` to `[x]` for same-package interface-method detection (dead code now fixed)
2. Add dedicated FuncLit flag-reset test (verify closures don't inherit InterfaceMethod)
3. Add SDK-level test verifying `InterfaceMethod` flows through `pkg/artdupl.Clone`
4. Run `art-dupl --type-aware --explain` on itself (the art-dupl codebase) to find real-world issues
5. Test `--type-aware` + `--incremental` + interface-method suppression together
6. Add test for generic type receivers (`func (s[T]) Method()`)
7. Add test for embedded struct receivers where the embedded type implements the interface
8. Add test for empty interface (no methods) — should not match anything
9. Add test for very large body (>4 statements) with InterfaceMethod=true — should NOT suppress

### Medium Priority (improve the feature)

10. Pre-compute interface-to-method-name map per package to avoid repeated scope scanning
11. Benchmark `IsInterfaceMethod` on a large package (100+ types, 10+ interfaces)
12. Add debug logging when type info is available but interface check fails
13. Consider cross-package interface scanning (load all packages in module)
14. Extend `commonInterfaceMethodNames` with community-reported missing entries
15. Add `--explain` output distinguishing "type-aware: satisfies interface X" vs "static name list"
16. Add a race test for `IsInterfaceMethod` (multiple goroutines parsing same package)
17. Add fuzz test for `IsInterfaceMethod` with malformed AST nodes
18. Consider whether path 1 (FuncDecl root) should be removed entirely as dead code

### CI / Build

19. Add a CI check that prevents re-adding disabled linters (beyond the Nix check — maybe a pre-commit hook)
20. Consider adding `godox` exception for test files (test doc comments often mention "bug" in regression context)
21. Run `golangci-lint run --timeout 5m ./...` on the FULL project (not just changed packages)
22. Consider adding a BDD test that specifically verifies pattern unreachability for common node types

### Documentation

23. Add a "Pattern Authoring Guide" section to ACTIONABILITY_PATTERNS.md explaining the structural filter
24. Document the `EnclosingReturnArity`/`InterfaceMethod` propagation pattern as a convention
25. Update HOW_TO_USE.md with `--type-aware` interface detection examples
26. Add before/after example showing custom interface suppression
27. Consider adding ADR for the statement-level propagation pattern
28. Update the previous status report (`2026-08-05_06-13`) to note the dead-code bug was found and fixed

### Code Quality

29. Consider extracting interface scanning into a reusable `PackageInterfaceIndex` type
30. Consider whether `InterfaceMethod` should be `InterfaceMethodSource` (enum: none/stdlib/type-aware)
31. Verify `.go-arch-lint.yml` constraints aren't violated by new transformer field
32. Check if `InterfaceMethod` needs to be in the SDK's `CloneNode` equivalent
33. The `findBodyStatements` test helper could be promoted to a shared test utility
34. Consider whether the `EnclosingReturnArity` / `InterfaceMethod` save/restore should use defer instead of manual restore

### Testing Improvements

35. Add test for `--type-aware` on a package with compilation errors (graceful fallback)
36. Add test for pointer receiver interface methods at the statement level
37. Add test for value receiver interface methods at the statement level
38. Add test for interface with embedded methods (e.g., `io.ReadWriter`)
39. Add property-based test: for any Go file with an interface and implementations, the pattern fires correctly
40. Test that the pattern does NOT fire when two functions have the same body but only one implements an interface

### Architecture / Future

41. Consider a general "type-aware actionability" framework that any pattern can opt into
42. The property-based engine (ADR-0017) could incorporate interface satisfaction as a property
43. Consider caching `go/types` results keyed by content hash (ROADMAP: "Caching for type-checking results")
44. Incremental type checking — only re-check changed packages (ROADMAP item)
45. The `type narrowing for interface-typed variables` ROADMAP item could interact with this work
46. Consider whether `interface-implementation` pattern (#2) also needs statement-level propagation
47. Audit all other actionability patterns for the same FuncDecl-root dead-code bug
48. Consider a linter rule that warns when a pattern matcher checks for a non-Statement root type
49. The `isSignatureOnlyMatch` pattern also checks FuncDecl/FuncType roots — verify it's not dead code too
50. Add a comprehensive integration test that runs ALL actionability patterns through the real pipeline

---

## g) Questions

1. **Should path 1 (FuncDecl-root matching) be removed entirely?** It's
   unreachable for real Go files and only exercisable via direct CloneNode
   construction in unit tests. Removing it simplifies the code but loses
   defense-in-depth for unusual inputs (e.g., generated files, templ edge
   cases). Keep, simplify, or remove?

2. **Should the ROADMAP entry be upgraded to `[x]`?** Same-package interface
   detection now works end-to-end (dead code fixed). The remaining gap is
   cross-package scanning, which could be a separate `[ ]` entry. Or should
   `[~]` stay until cross-package is also done?

3. **Should I audit `isSignatureOnlyMatch` and other FuncDecl-root patterns for
   the same dead-code bug?** The structural filter in `FindSyntaxUnits` affects
   ALL patterns that check for non-Statement root types. If `signature-only` or
   `interface-implementation` also require FuncDecl/FuncType roots, they may be
   equally dead code. This could be a significant finding across the entire
   actionability system.

---

_Auto-generated 2026-08-05 06:46 from session fixing the interface-method dead-code bug._

---

## Resolution (2026-08-10)

**Core work shipped.** The dead-code fix (InterfaceMethod flag propagation) is in CHANGELOG `[Unreleased]` → Fixed. Section b items (FuncLit test, static name list path) — FuncLit flag-reset test shipped. Section c items (cross-package detection, interface cache) → ROADMAP.
