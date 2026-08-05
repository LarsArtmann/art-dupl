# TODO List

**Last Updated:** 2026-08-05

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md` (`[Unreleased]`).
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### Actionability Engine (from calibration)

- [ ] **Exclude demo/example directories by default**: Large-scale calibration (`docs/calibration/confidence-calibration-large-2026-08-05.md`) found 4/12 false positives are throwaway code in `examples/`, `demo/`, `test_plugin/`, `test_temp/` dirs. Add these to the default exclusion list (alongside vendor), with `--include-examples` to override. Eliminates 4 FPs (86.2% → 91% precision).
- [ ] **Extend single-declaration pattern to type-alias blocks**: 3/12 FPs are `type X = pkg.Y` re-export shims (multiple aliases in one `type(...)` group). The existing `single-declaration` pattern catches lone aliases but not blocks. Extend it. Eliminates 3 FPs.
- [ ] **Add interface-assertion + widen test-helper-delegate patterns**: 2/12 FPs are compile-time `_ I = (*T)(nil)` assertions and `b.Helper()` (testing.B) delegates. Add interface-assertion pattern; widen test-helper-delegate receiver check from `testing.T` to include `testing.B`.
- [ ] **Surface suppressed groups for recall measurement**: Add `--show-suppressed` flag (or include suppressed groups in `--explain` with a "suppressed" label) so calibration can measure false negatives, not just precision. Currently recall is unmeasurable.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
