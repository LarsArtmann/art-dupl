# TODO List

**Last Updated:** 2026-08-05

Actionable items planned for the next 2-4 weeks. Completed work is in `CHANGELOG.md` (`[Unreleased]`).
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### Property Engine Follow-up

- [ ] **Calibrate confidence values against larger codebases**: Initial calibration on 8 projects showed 0 false positives with current thresholds, but sample size was too small (only 1 project produced meaningful clone data). Run on 3+ projects with 500+ Go files, manually label 50+ clone groups as actionable/non-actionable, compute precision/recall metrics, and tune thresholds empirically. The calibration harness (`scripts/calibrate-confidence.sh`) and initial report (`docs/calibration/confidence-thresholds-2026-08-05.md`) exist as starting points.

---

## MEDIUM Priority

### CI and Infrastructure

- [ ] **Integrate go-arch-lint into CI**: The `arch-lint` derivation was removed from `nix flake check` because `go-arch-lint` panics in the pure Nix sandbox (it uses `go/packages` which needs a Go toolchain + module cache at runtime). Options: (1) run it as a GitHub Actions step, (2) use a Nix devShell alias, or (3) add it to a pre-commit hook. The `.go-arch-lint.yml` config is complete and `go-arch-lint check` passes locally with 0 violations.

### Testing

- [ ] **Add test for the auto-fix guard script**: Verify `scripts/check-disabled-linters.sh` actually removes banned linters via `sed -i` when the file is writable.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
