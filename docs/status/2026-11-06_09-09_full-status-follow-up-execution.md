# Full Status Update — Follow-Up Queue Execution & Verification

**Date:** 2026-11-06 09:09
**Branch:** `fork` · HEAD `aea30779` · working tree **clean**
**Session span:** `5ad3bd7a..aea30779` — 13 commits, 38 files, **+1522 / −50**
**Predecessor reports:**
[`08-06` (issues #9–#14)](./2026-11-06_08-06_github-issues-9-14-resolution.md) ·
[`08-53` (queue execution)](./2026-11-06_08-53_issue-follow-up-queue.md)

Scope of this report: what I executed, what I got wrong, and the ranked
backlog that falls out of it. No new research beyond the session's own work.

---

## Direct answers first

### What did I forget?

1. **`baseline.File.Threshold` is recorded and never compared.** The baseline
   JSON has carried a `threshold` field since its inception; `check` never looks
   at it. A baseline recorded at `-t 5` and checked at `-t 2` reports a
   whole-tree "new clones" verdict — the exact failure mode issue #13 was
   filed about — and the identity guard I extended this session cannot see it
   (threshold changes _which_ groups exist, not their hashes). I noticed this
   while writing `analyzerPolicyParts`, decided it was an owner question, and
   did **not** even add the cheap version of the fix (compare
   `bf.Threshold != cfg.Threshold` → warn). That is the single biggest miss of
   the session.
2. **`--diff-report` did not get the strict flag it was asked for.** The queue
   said "wire `ensureBaselineIdentityMatches` into `--diff-report` behind the
   same flag". I substituted a text-only warning because
   `--baseline-version-must-match` is registered on the `check` subcommand and
   the root command has no baseline flags at all. Defensible, but it is a
   **substitution, not the requested behaviour** — recorded as PARTIAL below.
3. **Type-aware / combined mode was never swept** after the CSS-value change.
   I swept 5 corpora in default semantic mode only. The bump is CacheVersion
   9, which also invalidates the `ta` and `sg` cache tags; those paths are
   covered by unit tests but not by an end-to-end run.
4. **Doc claims I did not re-read while editing neighbours.** I verified README
   counts and `FEATURES`/`HOW_TO_USE` baseline text, but did not audit their
   `--simple-json` claims before changing that format. (Luckily `FEATURES:36`
   already described the _correct_ shape, so my fix made a false claim true —
   but that was luck, not process.)
5. **No test for `--all`'s `report.simple-json` content**, only file existence
   (`cmd/cmd_integration_test.go:317`). The new
   `TestSimpleJSONFormatEmitsImpactArray` covers the flag, not the batch path.
6. **`scripts/corpus-diff.py` has no test, no CI hook, and is not in `flake.nix`.**
   I validated it by hand (same-binary run + real diff), nothing guards it from
   rotting.
7. **No ADR amendment for the exit-code change.** Changing `check`'s verdict
   from exit 2 to exit 1 is a documented-contract fix, but ADR-0013's table
   says `1 = General error` and I reused exit 1 rather than defining a code for
   the verdict. The docs I wrote acknowledge it; the ADR does not.

### What could I have done better?

| # | Better version                                                                                              | Why it matters                                                                                                                                                                                        |
| - | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Grep for **dead exported functions** at session start (`OutputSimpleJSON` had zero callers since inception) | The `--simple-json` defect sat in plain sight; a `grep -rn "func (p \*JSONPrinter) Output" --include=*.go \| xargs` call-site check finds it in seconds                                               |
| 2 | Run the **"one invocation per documented format"** smoke sweep FIRST, not last                              | Both the JSON-document bug and the simple-json bug are format-level; format smoke testing found them in 10 seconds once I finally ran it, after ~3 hours of work                                      |
| 3 | Add compile-time **decorator interface assertions** instead of only documenting the gotcha                  | `baselineFilterPrinter` silently dropping `SetHash` produced empty hashes in production JSON; a `var _ printer.HashSetter = (*baselineFilterPrinter)(nil)` line turns that class into a build failure |
| 4 | Compare **`bf.Threshold`** while I was already in `runCheck` touching identity                              | One `if` and one warning; the highest-value 15 lines in the backlog                                                                                                                                   |
| 5 | Write the CSS BDD fixture **from the CLI reproduction** instead of inventing a fixture                      | My first fixture (two differently named css rules in one file) was toothless — the canary caught it, but only after I had written a passing test that proved nothing                                  |
| 6 | Treat `nix flake check` as the **first** gate of every session, not the last                                | The repo had been un-buildable via nix since the 2026-11-04 dep sweeps; `go build`/`go test` were green the whole time and hid it                                                                     |
| 7 | Verify `golangci-lint` **version parity** between devshell and sandbox                                      | The sandbox (2.15.0) reported 5 findings the devshell did not; the divergence is still unguarded                                                                                                      |

### What can I still improve?

- **Gate the sandbox toolchain**: a check that compares `golangci-lint --version`
  in the devshell against the flake's, and fails loudly on drift. Right now the
  only warning is prose in `AGENTS.md`.
- **Make "the doc claims are true" mechanical** for output shapes: assert each
  documented format's top-level JSON shape in one table-driven test, so
  `--simple-json`-class dead paths cannot survive.
- **Turn every warning I emit into a tested string**: the identity notice is
  pinned by substring only; the exact block format (`baseline:` / `running:` /
  `effect:` / `fix:`) has no golden.
- **Consolidate the stderr-notice policy**: dead directives, near-miss
  directives, identity mismatch, and (proposed) threshold mismatch each carry
  their own ad-hoc format gate. One helper (`warnIfHumanReadable(format, ...)`)
  would remove three copies of the same reasoning.

---

## a) FULLY DONE

Everything below is verified on this tree; `go test ./... -count=1`,
`scripts/check-boundary.sh --full`, `nix develop -c golangci-lint run` (0
issues), and `nix flake check` (12/12) all pass at `aea30779`.

| #  | Item                                                                                                                                                                                                                         | Evidence                                                                                                                                                                            |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **nix build repaired** — stale `vendorHash` FOD whose cached content predated the dep sweeps; refreshed to `sha256-AFkwR+czRL1litT7pMz0L1ZeUYfq/H23tACxWzgFn8I=`                                                             | `nix build .#art-dupl` builds; `./result/bin/art-dupl version` → `0.9.0`; root cause recorded in `AGENTS.md`                                                                        |
| 2  | **5 sandbox-only lint findings fixed** (`nonamedreturns`, `prealloc`, 2× `wsl_v5`, `unused`)                                                                                                                                 | sandbox lint 2.15.0 → `0 issues`; flake `lint` check green                                                                                                                          |
| 3  | **ADR-0029** — baseline analyzer identity contract (policy scope, version-as-provenance, legacy nil, opt-in gate)                                                                                                            | `docs/adr/0029-baseline-analyzer-identity.md`                                                                                                                                       |
| 4  | **Default-visible identity warning** on `check` without the strict flag and on text `--diff-report`; labelled `baseline/running/effect/fix` block; `--json` silent; exit codes unchanged                                     | 3 pins: `TestCheckWarnsOnIncomparableBaselineWithoutFlag`, `TestCheckWarnsOnLegacyBaselineWithoutFlag`, `TestDiffReportWarnsOnIncomparableBaseline`; live smoke on the built binary |
| 5  | **`check --json` emits a document at all** (it emitted nothing)                                                                                                                                                              | `TestCheckJSONEmitsDocumentAndBaselineSplit`                                                                                                                                        |
| 6  | **`check --json` new groups carry their hash** (`baselineFilterPrinter` now forwards `SetHash`)                                                                                                                              | same test asserts non-empty hash                                                                                                                                                    |
| 7  | **`check --json` carries `baseline: {in_baseline, new, suppressed, total}`** + `files_analyzed`                                                                                                                              | same test (1/0/0/1 unchanged; 1/1/0/2 after adding a family)                                                                                                                        |
| 8  | **`check` exit code matches its documented contract** — new clones now exit 1 via `errors.NewClonesFoundError`; missing baseline still exits 2                                                                               | `TestCheckNewClonesExitCodeIsOne`; live binary: `1` / `2`                                                                                                                           |
| 9  | **Constant CSS property values join the hash** (`setConstantCSSProperty`), CacheVersion 8 → 9; semantic collapses whitespace but never sorts shorthand                                                                       | 5 token tests + 1 BDD spec + 1 SDK test, **all canary-verified** (revert → fail)                                                                                                    |
| 10 | **`--simple-json` emits the documented impact array** (`[{hash, score, instances}]`), sorted; discovered it had **no caller at all**                                                                                         | `TestSimpleJSONFormatEmitsImpactArray`; live output inspected                                                                                                                       |
| 11 | **Gob cache round-trip guard** for every classification field (`GoSnippet`, `VarType`, `EnclosingReturnArity`, `InterfaceMethod`, `IsAlias`, `Statement`, `Fingerprint`, …)                                                  | `cache.TestFileCache_PreservesNodeClassificationFields`, canary-verified by deleting `GoSnippet` from `Clone()`                                                                     |
| 12 | **`scripts/corpus-diff.py`** — two binaries, one corpus, added/removed groups + same-span rekeys + category deltas, `--json` output                                                                                          | validated same-binary (empty) and real diff (2→1 groups on the CSS fixture); 5 corpora × `-t 1/2/5` swept                                                                           |
| 13 | **CSS corpus report**                                                                                                                                                                                                        | `docs/benchmarks/templ-css-value-2026-11-06.md` + index row                                                                                                                         |
| 14 | **`DOMAIN_LANGUAGE.md`** extended (markup identity, text content, Go snippet, cache version, hashing policy, analyzer identity, baseline split) and its stale "33 patterns" row corrected to 42                              | doc reads; count gate green                                                                                                                                                         |
| 15 | **`baseline` package doc** corrected (`recorded_at` → `recordedAt`) and extended with the `analyzer` block + ADR reference                                                                                                   | source read                                                                                                                                                                         |
| 16 | **README drift fixed** — 42 pattern labels (was "33+4"), LSP row added, CSV via `stats`, sorted-slice transitions (was "O(1) map-based"), stamp refreshed **after** verifying claims                                         | `scripts/check-docs-freshness.sh` clean; `TestNodeTypesCountsMatchDocs` / `TestActionabilityPatternCountsMatchDocs` / `TestWebsiteFlagsMatchBinary` green                           |
| 17 | **`AGENTS.md`** — CacheVersion 9, CSS-value policy, `check --json` + decorator-forwarding gotcha, format-routing branch, identity warning + ADR, gob boundary, stale-FOD gotcha, corpus-diff process rule, exit-code verdict | read against code                                                                                                                                                                   |
| 18 | **`CHANGELOG.md`** — 5 entries (simple-json, exit code, check --json, default warning, CSS values)                                                                                                                           | read                                                                                                                                                                                |
| 19 | **`FEATURES.md` / `HOW_TO_USE.md`** — templ row, baseline row, warning sample, JSON `baseline` sample, exit-code table                                                                                                       | read                                                                                                                                                                                |
| 20 | **`TODO_LIST.md`** — 2 items closed, 5 added (release, diff-report JSON guard, decorator assertions, threshold-in-policy owner question, SDK directory gap)                                                                  | read                                                                                                                                                                                |
| 21 | **Status index** — the two November reports added to `docs/status/README.md`                                                                                                                                                 | read                                                                                                                                                                                |

**14 new test entry points** (5 cmd, 1 cmd format pin, 1 cache, 1 BDD spec, 5 templ token tests, 1 SDK).

---

## b) PARTIALLY DONE

| # | Item                                             | What is there                                                                       | What is missing                                                                                                                                                                        |
| - | ------------------------------------------------ | ----------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **`--diff-report` identity guard**               | Text mode warns; JSON stays silent by design                                        | No strict mode. The requested `--baseline-version-must-match`-equivalent for diff-report does not exist (needs a root-level flag + website doc row)                                    |
| 2 | **Threshold/test-inclusion awareness**           | Documented as an owner question; the warning only fires for policy/cache mismatch   | `bf.Threshold` vs `cfg.Threshold` is never compared, so a threshold mismatch is still a silent whole-tree false alarm                                                                  |
| 3 | **Docs-health pass**                             | README stamp + claims re-verified; status index updated; freshness script clean     | The overdue monthly pass (due 2026-10-28) is not executed: no HARVEST of Feb–Oct status reports into TODO, no ANNOTATE, no archive gate run                                            |
| 4 | **`docs/status/README.md`**                      | Two new rows at the top                                                             | The "Overall Project Status" table below them still carries **January 2026** numbers (~181 lint violations, 65–75% coverage, 15–20% duplication) — now visually adjacent to fresh rows |
| 5 | **Combined/type-aware parity of the CSS change** | Semantic/exact/structural token tests; CacheVersion 9 invalidates `ta`/`sg` entries | No end-to-end combined-mode run on a templ corpus after the change                                                                                                                     |
| 6 | **`scripts/corpus-diff.py`**                     | Working, documented, used for 15 measurements                                       | No test, no CI wiring, not referenced from `flake.nix`                                                                                                                                 |
| 7 | **Exit-code change**                             | Behaviour matches `--help`/README; documented in AGENTS/CHANGELOG/HOW_TO_USE        | No ADR-0013 amendment; no decision on whether the verdict deserves its own code                                                                                                        |
| 8 | **Decorator-forwarding defence**                 | Gotcha documented in `AGENTS.md`; bug fixed                                         | No compile-time assertion, no lint rule — the next decorator can repeat it                                                                                                             |
| 9 | **`--all` path coverage**                        | `report.simple-json` is produced                                                    | Content is untested (existence only), so a regression there would pass CI                                                                                                              |

---

## c) NOT STARTED

1. **Release** carrying CacheVersion 9 (attribute 7, text 8, CSS 9) — needs the
   "re-run `art-dupl baseline`" note and the exit-code change called out.
2. **Threshold in the analyzer policy** (or the cheap warning).
3. **Strict identity mode for `--diff-report`** / root-level baseline flags.
4. **Compile-time decorator assertions** (`printer.Forwarding` or per-type `var _`).
5. **Sandbox/devshell toolchain-drift gate** (`golangci-lint`, `templ`, `go`).
6. **Format-shape table test** (every documented output's top-level JSON shape).
7. **Full docs-health pass** (HARVEST/ANNOTATE/archive).
8. **`templ` CLI drift** (generator 0.3.1020 vs library 0.3.1070).
9. **SDK directory-input rejection** (silently returns "no duplicates").
10. **`check --json` summary semantics** (`summary.total_clone_groups` counts
    new-only groups while `baseline.total` counts everything — documented in a
    comment, not in a test).

---

## d) TOTALLY FUCKED UP (fixed this session) and STILL FUCKED UP

### Fixed (was badly broken before this session)

| Defect                                                           | Blast radius                                                                                                            |
| ---------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| **`nix build` broken repo-wide** since 2026-11-04                | Every nix consumer, CI's nix lane, the release path — invisible to `go build`/`go test`                                 |
| **`check --json` emitted zero bytes**                            | A CI job parsing it saw an empty stream and could read it as "no clones" — the worst possible silent failure for a gate |
| **`check --json` new groups had `"hash": ""`**                   | Any consumer grouping/deduping by hash would collapse distinct groups                                                   |
| **`--simple-json` was byte-identical to `--json`**               | A documented format never existed; `OutputSimpleJSON` was dead code since inception                                     |
| **`check` exited 2 for new clones** while docs said 1            | Scripts branching on the documented codes misclassify a review failure as a config failure                              |
| **Sandbox lint red (5 findings)** while devshell said "0 issues" | The repo's lint gate was failing for anyone building in nix                                                             |
| **`color: red` cloned with `color: blue`** (templ css)           | False-positive clone groups on any templ file with constant CSS                                                         |

### Still fucked up (known, not fixed)

1. **`docs/status/README.md`'s "Overall Project Status" table is January 2026.**
   I added fresh rows above a table claiming ~181 lint violations and 65–75%
   coverage. That is my edit making an existing lie louder. Fixing it needs
   real measurements (coverage, duplication), which is the docs-health pass.
2. **Threshold mismatch is silent** (see b/2). Same class as the issue I closed
   yesterday, still live.
3. **`--diff-report --json` has no identity guard of any kind** — no warning, no
   strict mode. A machine consumer of the diff report can compare across hash
   spaces and believe it.
4. **Toolchain divergence is unguarded.** A devshell that reports 0 lint
   findings while the gate is red is a trust-destroying failure mode; today it
   is only described in prose.
5. **The SDK silently reports "no duplicates" for a directory input.** A
   consumer passing `"."` gets a clean bill of health for an unscanned tree.
   (Found while writing the SDK CSS test; TODO'd, not fixed.)
6. **`templ` generator/version skew** warns on every nix build; harmless today,
   a landmine when a generator-format change lands.

---

## e) WHAT WE SHOULD IMPROVE

1. **Session-opening gate order**: `nix flake check` → format smoke → suite.
   This session's two most severe defects (un-buildable tree, silent format
   bugs) were both invisible to `go build`/`go test`.
2. **Every documented output shape gets one executed assertion.** Documentation
   claimed simple-json's shape for years while the code emitted full JSON; only
   a real invocation disproved it.
3. **Canary every new test that guards a detected class.** Three of my new tests
   were toothless on first draft (BDD CSS fixture, SDK fixture, and a pre-sized
   `--all` assertion would have been) — the canary is what made them real.
4. **Prefer a compare-and-warn for every field already in the baseline file.**
   `Threshold` was sitting there unused; the same is true of `Version`.
5. **Make the toolchain part of the artifact.** Pin `golangci-lint`/`templ` so
   devshell and sandbox cannot disagree silently.
6. **Report format-level facts from the binary, not from the printers.** My
   `check --json` fix had to trace two layers (printer creation vs document
   emission) because the split lives in `handleJSONOutput`.
7. **Keep warnings uniform.** Four stderr notices now follow three different
   format-gate conventions.
8. **Write the "why" into the ADR, not only into the changelog.** The exit-code
   verdict change deserves an ADR line; today the reasoning lives in a commit
   message and a code comment.

---

## f) NEXT 50 ITEMS (ranked)

### Correctness / gates (do first)

1. Compare `baseline.File.Threshold` with the running threshold; warn (and
   optionally add it to the hashing policy behind a flag).
2. Add `--baseline-version-must-match` to the root command so `--diff-report`
   can refuse an incomparable baseline.
3. Add compile-time assertions (`var _ printer.HashSetter = ...` or a
   `printer.Forwarding` helper) for every decorator.
4. Add an `--all` content test for `report.simple-json` (array, score > 0).
5. Add a format-shape table test: for each documented format, assert the
   top-level JSON type and the presence of its documented fields.
6. Golden-test the identity warning block (exact `baseline:`/`running:`/
   `effect:`/`fix:` lines).
7. Reject directory paths in the SDK with an actionable error (or crawl them).
8. Add `summary.total_clone_groups`-vs-`baseline.total` semantics to the
   `check --json` test so the documented distinction is enforced.
9. Sweep `--type-aware` and combined mode on 2 templ corpora after CSS hashing.
10. Add a `golangci-lint` version-parity check between devshell and flake.

### Release / packaging

11. Cut the release (CacheVersion 9 + `check` exit code + simple-json) with an
    explicit "re-run `art-dupl baseline`" instruction.
12. Verify `CHANGELOG`'s newest header matches the tag before tagging
    (`scripts/pre-release-check.sh --version`).
13. Confirm `pkg.go.dev` / module proxy pick up the tag.
14. Decide whether the exit-code change requires a minor (not patch) version.
15. Add the pre-fix binary hash + date to the CSS sweep report for provenance.

### Documentation

16. Run the overdue docs-health pass (HARVEST Feb–Oct status reports into TODO).
17. Replace the January-2026 status table in `docs/status/README.md` with
    measured numbers, or delete it.
18. Index the Feb–Oct status reports in `docs/status/README.md`.
19. Add ADR-0013 amendment for the `check` verdict exit code.
20. Document `corpus-diff.py` in `CONTRIBUTING.md` (it is only in
    `docs/benchmarks/README.md`).
21. Add a "markup identity" subsection to `HOW_TO_USE.md` covering text,
    attributes, and CSS values together with the CacheVersion history.
22. Record the CSS-value defect + CacheVersion 9 in the templ limitation list in
    `AGENTS.md`'s Known Limitations (it is currently only in the templ bullet).
23. Refresh `docs/benchmarks/templ-text-precision-2026-11-06.md` with a pointer
    to its CSS sibling so the pair is discoverable.
24. Re-check `FEATURES.md`'s "8 output formats" claim against the enum
    (8 = text/html/json/csv/plumbing/simple-json/sarif/lsp; rich-text and
    LSP-vs-CSV wording needs one canonical sentence).

### Testing

25. Test `scripts/corpus-diff.py` (same-binary empty; fixture diff; JSON skip).
26. Add a BDD spec for `check --json` (currently cmd-level only).
27. Add a test that a `-t` mismatch produces the (item 1) warning.
28. Add a fuzz seed or property test for `normalizeConstantAttributeValue` /
    CSS value normalization.
29. Add a test asserting `OutputSimpleJSON` sorts (two groups, reverse input).
30. Add a race test around `JSONPrinter.SetBaselineSplit` + `OutputJSON`
    (both now mutate printer state).
31. Add an alloc-budget test for the CSS value path (`syntax/alloc_budget_test.go`
    sibling) since it adds a concat per constant property.

### Architecture / cleanup

32. Introduce `printer.warnHuman(format, msg)` and migrate the four stderr
    notices (dead directive, near-miss, identity, threshold).
33. Move `BaselineSplit` JSON shaping into a single `toJSONBaseline` helper
    (mirroring `toJSONClone`).
34. Consider a `baseline.Compatibility` value object that owns
    policy/version/threshold comparison instead of `cmd` free functions.
35. Replace the `format == config.OutputFormatSimpleJSON` branch with a
    `documentFor(format)` map, so a new format cannot silently inherit the full
    document.
36. Extract `runCheck`'s output assembly (printer + split + JSON + footer) into
    one helper that `baseline`/`check`/`--diff-report` share.
37. Audit `--simple-json` consumers inside the repo (`stats` CSV printer is
    reused via `withThreshold`) for the same "wrong document" class.

### Toolchain / infra

38. Upgrade the `templ` CLI (flake) to ≥ 0.3.1070 so the generator warning stops.
39. Add `templ generate` output-diff verification to the boundary gate.
40. Add a `nix flake check` invocation to the pre-push recipe in `AGENTS.md`
    (only the fast gate is documented there today).
41. Consider a flake check that asserts `go.mod`'s `a-h/parse` pin matches
    templ's requirement (the 11-error outage class).
42. Add a CI lane that runs `scripts/corpus-diff.py` self-comparison on the
    self-scan corpus (proves determinism per commit).

### Product / detection quality

43. Re-run the 7-corpus sweep for #11/#12/#13-affecting changes (they shipped
    without a sweep; only #14 and the CSS change have reports).
44. Evaluate whether CSS _expression_ values need the same treatment (they are
    parsed as Go already — verify with a fixture).
45. Check whether templ constant attribute values and text should also be
    sorted/coalesced in `--exact` mode (currently verbatim by design — confirm
    the intent in a test comment).
46. Investigate the residual "1-line templ fragment" groups that survive
    `trivial-fragment` on templ-heavy corpora (the CSS sweep surfaced none, but
    DiscordSync still shows small groups at `-t 1`).
47. Add an `--ignore-markup-text` opt-out **if** the owner wants an escape hatch
    after #14 (owner question).

### Process

48. Adopt "smoke one CLI invocation per documented format" as the first step of
    any output-layer change (add to `AGENTS.md` process rules).
49. Add a session-opening checklist to `AGENTS.md`: `nix flake check`,
    `check-docs-freshness.sh`, format smoke.
50. Add the "dead exported function" grep to the boundary gate or a small
    `unused-exported` check (`unused` only catches unexported).

---

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Release shape and ownership.** Should I cut `v0.9.2` now, or is a minor
   bump required because `check` now exits 1 (was 2) for new clones — and do you
   want me to tag and push, or do releases stay owner-only?
2. **The gate verdict code.** Do you want "new clones found" to stay exit 1
   (documented, general-error bucket) or get its own code (e.g. 4) so CI can
   distinguish "the tool ran and found clones" from "the tool failed"?
3. **Threshold and test-inclusion in the hashing policy.** Folding `--threshold`
   and `--ignore-tests` into the analyzer policy turns a mismatched comparison
   into a hard/gated error, at the cost of invalidating every baseline whose
   threshold a team tunes per lane. Which contract do you want: strict
   (policy includes threshold) or advisory (warn only)?

---

## Verification log (this tree, `aea30779`)

| Gate             | Command                                                                                      | Result                              |
| ---------------- | -------------------------------------------------------------------------------------------- | ----------------------------------- |
| Build            | `go build ./...`                                                                             | PASS                                |
| Suite            | `go test ./... -count=1`                                                                     | PASS (all packages)                 |
| Boundary (fast)  | `scripts/check-boundary.sh`                                                                  | OK, 17–20 s                         |
| Boundary (full)  | `scripts/check-boundary.sh --full`                                                           | OK (48 s cold / 17 s warm)          |
| Sandbox lint     | `nix develop -c golangci-lint run ./...` (2.15.0)                                            | 0 issues                            |
| Nix package      | `nix build .#art-dupl`                                                                       | PASS after vendorHash refresh       |
| Full nix CI      | `nix flake check`                                                                            | all 12 checks passed                |
| Docs freshness   | `scripts/check-docs-freshness.sh`                                                            | clean                               |
| Docs count gates | `TestNodeTypes…`, `TestActionabilityPattern…`, `TestCLIFlag…`, `TestWebsiteFlagsMatchBinary` | PASS                                |
| Corpus drift     | `scripts/corpus-diff.py`, 5 corpora × `-t 1/2/5`                                             | no drift                            |
| Canaries         | CSS token test, BDD CSS spec, SDK CSS test, gob field test                                   | each FAILS when the fix is reverted |
