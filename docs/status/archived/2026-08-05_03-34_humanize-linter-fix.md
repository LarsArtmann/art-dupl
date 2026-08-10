# Status Report: go-humanize-linter H001 Fix

> **Post-session annotation (2026-08-05):** All work complete and committed. The `KB`→`KiB` output change was accepted as correct (IEC labels for binary divisors). The `go-humanize` dependency is now a direct dep. Recorded in CHANGELOG `[Unreleased]`. Section C/N items are out-of-scope notes; section F is a brainstorm — key items harvested into TODO_LIST/ROADMAP.

**Date:** 2026-08-05 03:34
**Session scope:** Single finding (`H001` manual byte formatting) in `printer/text.go:364`, fixed end-to-end.
**Branch:** `fork` (3 commits ahead of `origin/fork`)

---

## Executive Summary

One linter finding was resolved properly: `printer/text.go::formatBytes` (18-line manual `KB/MB/GB` switch on `1024` divisors) replaced with `humanize.IBytes` (3 lines, correct `KiB/MiB/GiB` labels). `go.mod`/`go.sum`/`flake.nix::vendorHash` updated. Build, tests, race tests, `nix build`, `nix flake check fmt` and `bench` all green. Linter reports **0 findings**.

---

## a) FULLY DONE

| #   | Item                                                               | Evidence                                                     |
| --- | ------------------------------------------------------------------ | ------------------------------------------------------------ |
| 1   | `printer/text.go` `formatBytes` rewritten using `humanize.IBytes`  | `printer/text.go:364-371` (3 lines vs 18)                    |
| 2   | Doc comment added explaining unit choice (IEC, `du -h` convention) | `printer/text.go:364-368`                                    |
| 3   | Negative-byte safety preserved (`int`→`uint64` trap avoided)       | `printer/text.go:370`                                        |
| 4   | Test expectations updated for new labels (`KiB/MiB/GiB`)           | `printer/text_utils_test.go:104-108`                         |
| 5   | `github.com/dustin/go-humanize v1.0.1` added as direct dependency  | `go.mod`, `go.sum`                                           |
| 6   | `flake.nix` `vendorHash` recomputed (`sha256-xaky…→sha256-j7xK…`)  | `flake.nix:53`; `nix build` succeeds                         |
| 7   | Linter reports zero findings                                       | `/tmp/go-humanize-linter .` → `0 findings`                   |
| 8   | `go test ./...` passes (all 28 packages)                           | full suite green                                             |
| 9   | `go test -race ./printer/...` passes                               | race detector clean                                          |
| 10  | `nix build` succeeds                                               | `/nix/store/…-art-dupl-0.6.1` produced                       |
| 11  | `nix flake check` `fmt` and `bench` checks pass                    | both derivations built                                       |
| 12  | All changes committed via auto-commit daemon                       | `cd14e44a` (refactor), `133e1391` (vendorHash + templ regen) |

---

## b) PARTIALLY DONE

| #   | Item                         | Status                                                                                     |
| --- | ---------------------------- | ------------------------------------------------------------------------------------------ |
| P1  | `nix flake check` full suite | `fmt` ✓, `bench` ✓, `disabled-linters` ✗ (pre-existing, NOT mine), `lint` ✗ (pre-existing) |

`disabled-linters` and `lint` failures are **pre-existing on `cd14e44a^`** (verified by stashing my changes — baseline fails the same way). They are NOT caused by this session.

---

## c) NOT STARTED (Out of Scope)

| #   | Item                                                                                                          |
| --- | ------------------------------------------------------------------------------------------------------------- |
| N1  | Fixing pre-existing `tagliatelle` re-enabling in `.golangci.yml` (50 issues)                                  |
| N2  | Fixing pre-existing `godox` + `nlreturn` lint issues                                                          |
| N3  | Investigating/auditing other rules (H002-H009) for additional findings in the repo                            |
| N4  | Updating AGENTS.md / TODO_LIST.md / FEATURES.md / ROADMAP.md / CHANGELOG.md to record the new dep             |
| N5  | Adding a `TestFormatBytes_NegativeBytes` table case if we ever lift the negative-handling into the public SDK |
| N6  | Updating `.golangci.yml` to recognize `humanize.IBytes` style or marking the file exempt                      |

---

## d) TOTALLY FUCKED UP

Nothing. No regressions introduced. All tests that passed before still pass. The user-visible output change (`KB` → `KiB`) is intentional and correct (the old output was technically wrong: SI labels with binary divisors is a common misuse).

**One near-miss caught**: when I first added `go mod tidy` after `go get`, it removed the unused dep. I added the source edit before re-tidying — correct ordering saved a debugging round-trip.

---

## e) WHAT WE SHOULD IMPROVE

1. **The user output `KB`→`KiB` change is a behavior change**, not just a linter suppression. The text-printer header for file duplicates now shows IEC units. This is technically correct but may surprise users with grep-able output. **Decision**: correct is better than compatible. Document in CHANGELOG.md.

2. **`go-humanize` is a transitive dep through nothing** — it's now a new direct dep. The project's dep list is curated and minimal. **Trade-off accepted**: one new dep saves ~15 LOC of bespoke formatting. Worth it.

3. **The `makezero` linter** could now flag the new `uint64(bytes)` cast if `bytes < 0` makes it zero. It's a 3-line guard, fine.

4. **The linter doesn't enforce `humanize` as the only valid choice** — it just flags manual implementations. A future contributor could legitimately write a 1-liner `fmt.Sprintf` with `"%d B"` for a single-byte case and not get flagged. The rule has surface area for false negatives.

5. **I asked the user about the fix strategy** — good (genuine trade-off: dep cost vs suppression). Could have been faster by defaulting to `//nolint:all` since it's a single isolated helper. But the proper fix is better.

6. **I didn't check whether `nix flake check` was passing before** — should have run it once on baseline to establish ground truth before touching anything. I did this mid-session via `git stash`, which works but is noisy.

7. **No test was added for the negative-input branch** in `formatBytes` — YAGNI since `FileSize = len(content)` is always ≥ 0. But the branch exists; coverage is incomplete.

8. **The `nix` flake check failed at first** because the old `vendorHash` was already stale (the auto-commit daemon committed a flake.nix update before I touched it). I had to set `vendorHash=""`, build to get the new hash, then update. This is standard but fragile — a hash mismatch is a common Nix pain point.

9. **The `buildflow pre-commit` hook failed** when I tried to commit the flake.nix change manually. The auto-commit daemon handled it before I could. The hook is undocumented in AGENTS.md — discovered by accident. Worth documenting.

10. **No golden-file/BDD test was updated** because none of them reference the byte-formatting strings. Verified via `grep`. Good — but I should document why I checked.

---

## f) Up to 50 Next Things to Get Done

Priority order — Pareto (1% → 51% impact first):

| #   | Task                                                                                                | Impact | Effort |
| --- | --------------------------------------------------------------------------------------------------- | ------ | ------ |
| 1   | Run the same linter against the WHOLE repo with all H rules enabled — find what else is hand-rolled | High   | Low    |
| 2   | Audit `.golangci.yml` for the `tagliatelle` re-enabling (50 errors reported by `nix flake check`)   | High   | Low    |
| 3   | Fix the `godox` + `nlreturn` issues in `nix flake check lint`                                       | Medium | Low    |
| 4   | Update CHANGELOG.md with the `KB`→`KiB` user-visible output change                                  | Medium | Low    |
| 5   | Update FEATURES.md to mention `go-humanize` as a new direct dependency                              | Low    | Low    |
| 6   | Audit `printer/` for any other hand-rolled formatting (relative time, percentages, pluralization)   | Medium | Medium |
| 7   | Audit `cmd/` for manual progress/timing formatting (often a hand-rolled `humanize` candidate)       | Medium | Medium |
| 8   | Consider exposing `formatBytes` (or similar helpers) as `pkg/format` exports for SDK consumers      | Medium | Medium |
| 9   | Add `golangci-lint` exclusion for the `templ-rendering-idiom` false-positive on `.templ` files      | Medium | Medium |
| 10  | Profile startup time impact of the new `go-humanize` import (negligible but worth measuring)        | Low    | Low    |
| 11  | Add `gocyclo`/`gocognit` exclusions for the regex-heavy acceptance-directive scanner                | Low    | Low    |
| 12  | Consider `go.mod` `toolchain` directive pinning (currently just `go 1.26.5`)                        | Medium | Low    |
| 13  | Migrate `flake.nix` from `proxyVendor = true` to direct vendor tracking if any CI flakiness         | Low    | Medium |
| 14  | Add an integration test that the text-printer header is byte-correct vs `len(content)`              | Low    | Low    |
| 15  | Audit `errors/` package for any hand-rolled error-message formatting that could use `humanize`      | Low    | Medium |
| 16  | Document the `buildflow pre-commit` hook in AGENTS.md (discovered this session)                     | Low    | Low    |
| 17  | Add a CONTRIBUTING.md note that H001-H009 are enforced by a pre-commit-grade linter                 | Low    | Low    |
| 18  | Check if `go-humanize` has a stable v2 (it doesn't as of 2026-08) — revisit annually                | Low    | Low    |
| 19  | Verify `nix flake check` from a clean cache (the eval-cache SQLite busy errors suggest dirty state) | Low    | Low    |
| 20  | Consider vendoring `go-humanize` directly (skip the proxy) — already covered by `proxyVendor`       | Low    | Low    |
| 21  | Run `golangci-lint run --timeout 5m ./...` with the project-default config and diff vs my changes   | Low    | Low    |
| 22  | Re-run `/tmp/go-humanize-linter` against `examples/` and `bdd/` test fixtures                       | Low    | Low    |
| 23  | Profile `nix build` time delta from the new dep (likely <1s)                                        | Low    | Low    |
| 24  | Verify `nix develop` (devShell) still works with the new `vendorHash`                               | Low    | Low    |
| 25  | Consider replacing any other hand-rolled `fmt.Sprintf("%.1f X", n/Y)` patterns with `humanize.SI`   | Medium | Medium |
| 26  | Check if `templ` has its own byte/duration formatting that supersedes `humanize` for `.templ` files | Low    | Low    |
| 27  | Verify the `--explain` flag (mentioned in AGENTS.md) still works after the label change             | Low    | Low    |
| 28  | Run `go test -bench=. ./...` to confirm no benchmark regression                                     | Low    | Low    |
| 29  | Add a `// Code generated by … DO NOT EDIT` guard to `printer/report_templ.go` if missing            | Low    | Low    |
| 30  | Confirm the `auto-git-commit` daemon's two-commit split (refactor + chore) was correct semantically | Low    | Low    |

(Stopping at 30 — the remaining 20 would be nits, refactors downstream, or speculative.)

---

## g) Questions I CANNOT Figure Out Myself

1. **`KB`→`KiB` is a user-facing output change.** Is this acceptable for the next minor release (`v0.6.2`), or should I revert to `KB/MB/GB` labels by writing a small `IECtoSI` wrapper that strips the `i`? The latter keeps the dep but preserves old output. **I cannot decide this** because it depends on whether you treat the text-printer header output as a stable contract (some users may grep for `KB` in CI logs).

2. **The `nix flake check` `disabled-linters` failure is pre-existing but blocks CI.** Should I fix it in this session (one-line `.golangci.yml` edit to remove `tagliatelle` from `enable:`), or leave it for a dedicated lint-cleanup session? **I cannot decide** because the fix is trivial but it expands scope beyond "linter finding".

3. **Should `go-humanize` be added to `pkg/format`** as a public SDK helper (e.g., `pkg/format/Bytes()`), or kept private inside `printer/`? **I cannot decide** because it depends on whether the SDK should expose formatting utilities or remain a thin wrapper around the detector. The AGENTS.md split (SDK vs internal) suggests no, but downstream consumers might want it.

---

## Resolution (2026-08-10)

**Main work shipped.** The `humanize.IBytes` fix is in CHANGELOG `[Unreleased]` → Added/Changed. The `KB`→`KiB` label change (Q1) was accepted — it's the correct IEC terminology.

**Section f brainstorm routing:** Item 2 (tagliatelle) → TODO_LIST HIGH priority. Items 3 (godox) → done. Item 4 (CHANGELOG) → done. Remaining items are nits or obsolete.

**Questions Q1-Q3:** Q1 answered (KiB labels accepted). Q2 → TODO_LIST (tagliatelle fix). Q3 → ROADMAP (SDK formatting helpers, not pursued).
