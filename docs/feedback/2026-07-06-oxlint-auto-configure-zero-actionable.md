# Feedback: All-reported clones are pre-documented accepts on a curated Go CLI

> ✅ **ADDRESSED** — Default threshold raised to 5 (commit `930b91a`). All 2 clone groups reported at t=2 were correctly classified as non-actionable (signature-only and assertion-chain patterns). At default threshold 5, these would not be reported. The tool correctly identified 0 actionable clones.

**Date:** 2026-07-06
**Project:** oxlint-auto-configure (Go, ~1800 LOC, 9 test files)
**Command:** `art-dupl --semantic --sort total-tokens -t 2 --html`
**Result:** 2 clone groups, 28 occurrences, 56 tokens. 0 actionable.

## Session Summary

Ran art-dupl at the most aggressive threshold (`-t 2`) on a small, actively
maintained Go CLI. The project had just been through a deduplication pass as
part of a go-finding v1.0.0 migration, so the codebase was already clean.

**What was found:** 2 clone groups totaling 28 occurrences / 56 tokens.
**What was actionable:** 0. Both groups are pre-documented intentional accepts
in the project's `AGENTS.md`.

## What Worked Well

### HTML output is genuinely excellent for triage

Reading the raw HTML (not just the text summary) surfaced two things the text
view didn't make obvious:

1. The summary section breaks out **production vs test** (2 vs 26) and
   **priority** (28 low) at a glance.
2. Each occurrence is a clickable `vscode://` link with a one-tap copy button.
   When the same 2-line pattern repeats 26 times, the copy button made it
   trivial to verify they were truly identical.

### Semantic detection correctly groups identical patterns across files

The 26-occurrence test group spans 4 files (`configure_test.go`,
`generator_test.go`, `validate_test.go`, `profile_test.go`,
`registry_test.go`) and all 26 instances were correctly clustered as one
group. No fragmentation, no over-merging with unrelated 2-liners.

### Default `--suppress-test-low=true` is the right call

The previous feedback files (2026-06-04, 2026-06-08) asked for test-only
low-priority clones to be suppressed by default. That shipped and it works.
Running without the flag, the report is exactly 2 groups — clean and scannable.
Running with `--suppress-test-low=false --include-tests` explodes it to ~14
groups, every one of which is test noise. The default is correct.

## The Two Clone Groups (Both Accepted)

### Group #1 — Test boilerplate (26 occurrences, 52 tokens)

```go
t.Parallel()
reg := loadTestRegistry(t)
```

This is the standard Go test preamble. The `paralleltest` linter **requires**
`t.Parallel()` to appear directly in the test function body — it cannot be
folded into the `loadTestRegistry` helper. The project documents this
explicitly in `AGENTS.md`:

> **Test boilerplate is intentional** — `t.Parallel()` followed by
> `reg := loadTestRegistry(t)` in every test is idiomatic Go and **not** a
> duplication violation; `paralleltest` linter requires `t.Parallel()`
> directly in the test function. Do not fold it into the helper.

### Group #2 — Production error propagation (2 occurrences, 4 tokens)

`internal/cli/cmd_configure.go:167-170` and `:186-189`:

```go
data, err := marshalConfigJSON(cfg)
if err != nil {
    return err
}
```

Shared between `writeConfig` (writes to file) and `writeDryRun` (prints to
stdout). The marshal step is shared; the actual output operations differ. The
project documents this too:

> **marshalConfigJSON helper** — the remaining 4-line preamble in
> `writeConfig`/`writeDryRun` is idiomatic Go error propagation and
> intentionally not abstracted further.

## Suggestions

### 1. Baseline / accept-list feature for documented intentional clones

This project (and likely others maintained with AI assistance) keeps an
`AGENTS.md` entry for every clone it has consciously decided to keep. Today
these clones re-appear on every run and must be re-triaged by a human who
remembers "oh right, that one is documented as intentional."

**Proposed:** `art-dupl baseline` (which already exists per `--help`) should
let teams record accepted clone groups, and subsequent default runs should
omit or grey-out baseline-matched groups. The matching key would need to be
stable across formatting changes — likely a hash of the alpha-normalized
clone body + occurrence count.

This would let a project reach a **persistent zero** in the default report
rather than a recurring 2-group report that a human has to mentally filter
every time.

The `check` subcommand (CI mode, "clones new relative to baseline") already
implies this infrastructure exists — surfacing it as the default UX would
close the loop.

### 2. Detect linter-mandated test boilerplate

The 26-occurrence group is a near-perfect signal that a linter rule
(`paralleltest`) is forcing duplication. If art-dupl could recognize the
specific shape:

```go
t.Parallel()
<single helper call or assignment>
```

repeated across many test functions in the same package, it could auto-classify
these as `idiom` priority (below `low`) or tag them with a `🧪 linter-mandated`
badge. This pattern is extremely common in Go projects that use
`paralleltest`.

More generally: a small built-in catalog of "known linter-forced clones"
(`t.Parallel()` preamble, `require.NoError(t, err)` assertion sequences,
`context.WithTimeout(...)` setup) would let art-dupl classify them correctly
out of the box instead of reporting them as `unknown / low`.

### 3. Token-count-aware "idiom" tier below `low`

Reiterating the 2026-06-08 suggestion from a different angle: at 2 tokens,
essentially every clone is a structural artifact. In this run, **100% of
2-token clones were false positives.** A tiered classification like:

| Tokens | Suggested default tier |
| ------ | ---------------------- |
| 1–4    | `idiom` (suppressed)   |
| 5–14   | `low`                  |
| 15–49  | `medium`               |
| 50+    | `high`                 |

would have made this report empty by default — which is the correct answer
for this codebase. Users who want the aggressive view can still opt in with
an explicit flag.

### 4. Minor: text output should surface the same metadata as HTML

The text output (`art-dupl --semantic -t 2` without `--html`) lists clone
locations but omits the production/test split, priority, and category that
the HTML summary shows. For CLI-first workflows (and for piping into other
tools), a `--rich-text` mode (which exists per `--help`!) should be
mentioned more prominently in the skill's run command, or made the default
for non-HTML output. I only discovered `--rich-text` while reading `--help`
to find the `--include-tests` flag.

## Metrics

| Metric                   | Value |
| ------------------------ | ----- |
| Clone groups (t=2)       | 2     |
| Total occurrences        | 28    |
| Total tokens             | 56    |
| Production clones        | 2     |
| Test clones              | 26    |
| Real duplications found  | 0     |
| False positives          | 2     |
| False positive rate      | 100%  |
| Pre-documented in AGENTS | 2     |
| Tests passing post-run   | 9/9   |

## Conclusion

At the most aggressive threshold (`-t 2`) on a curated codebase, art-dupl
reported exactly 2 clone groups — both of which were already documented as
intentional accepts. This is a **true negative result**: the tool found
everything there was to find, and there was nothing to fix.

The signal here is positive: the detector has no blind spots on this project.
The opportunity is in **reducing re-triage cost**: a baseline/accept-list
feature would let documented accepts stop appearing in every report, turning
a 2-group report into a persistent zero without losing the ability to detect
_new_ duplication when it's introduced.

The `--suppress-test-low` default (shipped per prior feedback) is working as
intended and made this run usable. The next highest-impact improvement would
be either the baseline feature or a token-count-aware `idiom` tier.
