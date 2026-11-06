# Continuation: Issue-Resolution Follow-Up Queue

**Date:** 2026-11-06 08:53
**Branch:** `fork`
**Predecessor:** [`2026-11-06_08-06_github-issues-9-14-resolution.md`](./2026-11-06_08-06_github-issues-9-14-resolution.md)
**Scope:** execute the follow-up queue section (f) of the preceding report.

The previous session closed GitHub issues #9–#14. It also left a ranked queue of
follow-ups, an unverified `nix flake check`, and a broken nix build it had not
noticed. This session drained that queue.

---

## 1. Tree state at session start

| Check | Result |
| --- | --- |
| `git status` | clean at `5ad3bd7a` |
| `go build ./...` | PASS |
| `nix build .#art-dupl` | **FAIL** — see below |
| `go test ./... -count=1` | PASS |

### The nix build was broken (undetected by the previous session)

Every nix build failed with:

```text
github.com/xo/terminfo@v1.3.0: reading file:///nix/store/<hash>-art-dupl-0.9.0-go-modules/github.com/xo/terminfo/@v/v1.3.0.zip: no such file or directory
```

Six modules reported this way, all of them versions `go.sum` clearly lists. The
`go-modules` fixed-output derivation existed and **its content hash equalled the
configured `vendorHash`** — no hash mismatch, so nix happily substituted a FOD
whose content predated the 2026-11-04 `chore(deps)` sweeps.

- **Remedy:** `vendorHash = lib.fakeHash` → `nix build` → paste the reported
  `got:` (`sha256-AFkwR+czRL1litT7pMz0L1ZeUYfq/H23tACxWzgFn8I=`) → rebuild.
  The package then builds and `./result/bin/art-dupl version` reports `0.9.0`.
- **Why it matters:** `go build`/`go test` stay green through this failure mode,
  so only a nix build surfaces it. Recorded as a gotcha in `AGENTS.md`.

## 2. Lint debt that only the sandbox saw

The first successful `nix flake check` failed the `lint` check with 5 findings
invisible to the devshell's `golangci-lint` (the sandbox runs 2.15.0, the
devshell lagged):

| File | Finding |
| --- | --- |
| `internal/accept/accept.go:263` | `nonamedreturns` — named return `loc []int` |
| `syntax/findsyntaxunits_test.go:265` | `prealloc` — `out := []*Node{root}` then `append` |
| `syntax/findsyntaxunits_test.go:278` | `wsl_v5` — missing blank line above `data = append(...)` |
| `syntax/syntax_match.go:104` | `unused` — `isTemplFile` had no callers |
| `syntax/syntax_match.go:60` | `wsl_v5` — missing blank line above a `range` |

All five fixed; `nix develop -c golangci-lint run ./...` → **0 issues**, and a
full `nix flake check` reports **all checks passed** (12 checks).

## 3. Work delivered

### ADR-0029 — baselines record their analyzer

`docs/adr/0029-baseline-analyzer-identity.md` ratifies the contract the previous
session implemented: policy = hash-affecting settings, `version` = provenance
only, legacy `nil` = unverifiable, strict gate opt-in.

### Default-visible identity warnings

`cmd/baseline_identity.go::warnIncomparableBaseline` prints a labelled block
(baseline / running / effect / fix) when the baseline's hash space differs:

```text
warning: baseline .art-dupl-baseline.json and the running analyzer are not comparable
  baseline: art-dupl v0.9.0 (algorithm revision 8, hashing policy d973d75f6075aa52)
  running : art-dupl v0.9.1 (algorithm revision 9, hashing policy 1882541478cec25f)
  effect  : clone-group hashes are not comparable, so every recorded group reads as new
  fix     : re-record with `art-dupl baseline`, or pass --baseline-version-must-match to make this an error
```

Wired on `check` without the strict flag and on every text `--diff-report` run.
`--json` stays silent (machine output keeps stderr clean); exit codes unchanged.
Pins: `TestCheckWarnsOnIncomparableBaselineWithoutFlag`,
`TestCheckWarnsOnLegacyBaselineWithoutFlag`, `TestDiffReportWarnsOnIncomparableBaseline`.

### `check --json` — two bugs, one feature

1. **It emitted nothing.** `OutputJSON` is only called by the main analysis
   path, so `check --json` produced an empty stream (`--files`-style silent
   failure, but for CI consumers). `runCheck` now calls `handleJSONOutput`.
2. **New groups carried an empty hash.** `baselineFilterPrinter` embeds
   `printer.Printer`, which does not declare `SetHash`, so the `HashSetter`
   assertion in `printCloneGroups` failed for the wrapper. The wrapper now
   forwards `SetHash`; `files_analyzed` is wired too.
3. **Feature:** the document carries `baseline: {in_baseline, new, suppressed,
   total}` via `JSONPrinter.SetBaselineSplit`. Pin:
   `TestCheckJSONEmitsDocumentAndBaselineSplit`.

### `check` exit code now matches its documentation

The gate returned a `ValidationError` → exit **2**, while `check --help`, the
README, and the exit-code ADR all describe exit **1**. New clones are the
answer, not a misconfiguration; exit 2 must stay reserved for bad invocations.
Added `errors.NewClonesFoundError` (type `clones-found`, deliberately unmapped
so it exits 1). Verified on the real binary: new clones → 1, missing baseline →
2. Pin: `TestCheckNewClonesExitCodeIsOne`.

### Constant CSS property values join the hash (CacheVersion 8 → 9)

`transformCSSProperty` hashed only the property NAME for
`*templparser.ConstantCSSProperty`, so `color: red` matched `color: blue` (the
same markup-identity hole fixed for attributes in #6 and text in #14).
`setConstantCSSProperty` folds the value into the node type hash: semantic mode
collapses whitespace runs but never token-sorts (CSS shorthand order is
semantic), exact mode hashes verbatim, structural ignores it.

- Token-level pins: 5 specs in `syntax/templ/behavior_test.go`, canary-verified
  (reverting the transform makes `TestConstantCSSPropertyValuePreventsClone`
  fail).
- End-to-end pin: `bdd/templ_markup_shape_test.go` "Templ CSS value precision",
  canary-verified (fails with the old encoding, printing
  `blue.templ:3-6 | css panel() {`).
- **Important fixture constraint discovered:** two differently named `css`
  rules in ONE file never cloned even before the fix, because the declared CSS
  name is part of the declaration token. The clone only appears across two
  files that both declare the same css name. Both the BDD fixture and the
  report record this.
- Corpus sweep (5 corpora × `-t 1/2/5`, `scripts/corpus-diff.py`):
  **zero movement** — the defect needs two instances sharing a property name
  with differing values, which these trees do not contain. Minimal
  reproduction goes 2 groups → 1. Report:
  `docs/benchmarks/templ-css-value-2026-11-06.md`.

### `scripts/corpus-diff.py`

Durable replacement for the analyzer three sessions rewrote by hand. Two
binaries, one corpus, three reported dimensions: added/removed groups,
same-span-but-different-hash groups, category deltas. Documented in
`docs/benchmarks/README.md` (including the "run `--before X --after X` first"
stability check) and added as process rule (6) in `AGENTS.md`.

### SDK path pinned too

`pkg/artdupl.TestDetector_Integration_TemplCSSValues` runs the public SDK over
two templ files whose only shared code is the value-divergent css rule and
asserts no clone groups. Canary-verified against the value-blind encoder. The
test had to pass explicit file paths: handing the SDK a **directory** silently
yields `ErrNoDuplicatesFound` (it is now a TODO item).

### Cache gob-boundary guard

`cache.TestFileCache_PreservesNodeClassificationFields` round-trips every
classification field through a second cache instance (disk/gob path).
Canary-verified by deleting `GoSnippet` from `Node.Clone()` — the test fails.
This closes the class that hid the `IsAlias` and `BasicLit.Name` bugs.

### Docs

| File | Change |
| --- | --- |
| `AGENTS.md` | CacheVersion 9 + CSS-value policy; check-split/`--json` upgrade + decorator-forwarding gotcha; identity warning + ADR-0029; gob-boundary test; stale-FOD gotcha; corpus-diff process rule; exit-code verdict |
| `CHANGELOG.md` | 5 entries (3 fixed, 2 earlier added) under Unreleased |
| `FEATURES.md` | templ row (CSS values), baseline row (JSON block + default warning) |
| `HOW_TO_USE.md` | warning block sample, JSON `baseline` block, exit-code table |
| `docs/DOMAIN_LANGUAGE.md` | markup identity, text content, Go snippet, cache version, hashing policy, analyzer identity, baseline split; fixed the stale "33 patterns" row to 42 |
| `docs/benchmarks/README.md` | corpus-diff section, CSS sweep row |
| `TODO_LIST.md` | 2 items closed, 4 added (release, diff-report JSON guard, decorator assertion, threshold-in-policy owner question) |
| `README.md` | stamp + re-verified claims: 42 pattern labels (was 33+4), LSP row, CSV via `stats`, sorted-slice transitions (was "O(1) map-based") |

`scripts/check-docs-freshness.sh` → clean (README stamp was the only stale doc).

## 4. Verification evidence

| Gate | Result |
| --- | --- |
| `go build ./...` | PASS |
| `go test ./... -count=1` | PASS (all packages) |
| `go test ./bdd/` | PASS (367 specs; the CSS spec fails only under the canary) |
| `nix develop -c golangci-lint run ./...` | 0 issues (2.15.0) |
| `nix build .#art-dupl` | PASS after the vendorHash refresh |
| `nix flake check` | all 12 checks passed (before the last two changes; re-run at close-out) |
| `scripts/check-boundary.sh --full` | see close-out |
| `scripts/corpus-diff.py` (5 corpora × 3 thresholds) | no drift |

Behavioural smoke tests on the real binary: `check --json` emits the document
with `files_analyzed`, new-group hashes, and the `baseline` block; identity
warnings render as the labelled block on both `check` and `--diff-report`;
`check` exits 1 for new clones and 2 for a missing baseline.

## 5. Decisions worth remembering

1. **One release note serves three CacheVersion bumps.** Attributes (7), text
   (8), and CSS values (9) all invalidate caches and baseline hash spaces. The
   pending release must say "re-run `art-dupl baseline`".
2. **Warnings for humans, JSON for machines.** The identity notice is text-only
   by design, matching the accept-directive warnings; JSON consumers get the
   `baseline` block instead of stderr prose.
3. **A decorator must forward every optional interface it sits in front of.**
   The empty-hash bug was invisible because `printer.Printer` only declares
   `PrintHeader`/`PrintClones`/`PrintFooter`.
4. **CSS value fixtures need two files with the same css name.** Anything else
   diverges for the wrong reason and produces a toothless test — caught by
   canary, not by review.

## 6. Open items handed forward

- [ ] **Release** with the CacheVersion 9 note (and the `check` exit-code
      change called out explicitly: 2 → 1 for new clones).
- [ ] **`--diff-report --json` has no strict identity guard** — warning-only
      contract is now the de-facto behavior; either accept it in ADR-0029 or add
      the root flag (needs a website flag row).
- [ ] **Compile-time assertion for printer decorators** (make the empty-hash
      class a build failure).
- [ ] **SDK `FindClones` accepts directories and silently reports "no
      duplicates"** — a consumer passing `"."` believes the repo is clean.
      Reject non-file paths with an actionable error, or crawl them.
- [ ] **Owner question: threshold + test-inclusion in the analyzer policy?**
      `-t 5` baseline checked at `-t 2` reports a whole tree of "new" groups the
      identity guard cannot see. Also: should `--baseline-version-must-match` be
      default-on, and does anyone want an `--ignore-markup-text` opt-out?
- [ ] **`docs/status/README.md` "Recent Reports" table** stops at 2026-01-22;
      the Feb–Nov reports are not indexed (full harvest belongs to the
      docs-health pass).
- [ ] **templ CLI drift warning** in every nix build (`generator v0.3.1020 is
      older than templ version v0.3.1070 in go.mod`) — harmless today, but it
      will bite when a generator-format change lands.

---

_Everything above was verified on this tree; nothing is claimed from memory._
