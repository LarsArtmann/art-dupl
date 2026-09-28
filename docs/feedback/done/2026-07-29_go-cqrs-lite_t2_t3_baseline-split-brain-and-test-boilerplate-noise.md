# Feedback: Baseline + accept-directive split-brain; test boilerplate dominates at -t 2

**Date:** 2026-07-29
**Project:** `go-cqrs-lite`, a 59-module Go workspace (1358 files) — CQRS/ES library with event sourcing, projections, catalog generation, storage backends
**Command:** `art-dupl --type-aware --sort total-tokens -t 3` then `-t 2`
**Goal:** Drive harmful duplication to zero across a large multi-module monorepo

> **Verdict:** The baseline workflow (`baseline` / `check` / `.art-dupl-baseline.json`) is well-designed and the CI gate correctly reported "0 new clones." However, two real gaps emerged: (1) **split-brain between baselines and `//art-dupl:accept` directives** — I accidentally double-accepted the same clones via both mechanisms with no warning, and (2) **test boilerplate (`t.Parallel()`, `TestMain`) floods the `-t 2` report** with ~190 low-value clones across 9 groups, drowning the 24 genuinely actionable production-code groups. The `--explain` classification and `--no-accept-directives` debug flag are excellent; the type-aware detection is valuable but doesn't prevent structural-identity false positives when the clone body is a standard-library API call.

---

## Results at a Glance

| Threshold | Clone Groups | Truly Actionable | Accepted (baseline) | Test Boilerplate       |
| --------- | ------------ | ---------------- | ------------------- | ---------------------- |
| `-t 3`    | 2            | 0                | 2                   | 0                      |
| `-t 2`    | 48           | ~24              | 15                  | 9 groups (~190 clones) |

The `-t 3` gate (`nix run .#check-duplication`) is clean. The `-t 2` frontier has real work but is 80% noise by volume.

---

## Finding 1 (HIGH PRIORITY): No conflict detection between baseline and `//art-dupl:accept` directives

### What happened

The repo has a committed `.art-dupl-baseline.json` (17 entries) and a `dedup-acceptance.md` documenting why each accepted clone exists. At `-t 3`, the bare report showed 2 clone groups — both already in the baseline.

I then added `//art-dupl:accept` inline directives to 18 files covering those exact same clones. The tool happily applied both acceptance mechanisms simultaneously with zero warning. I also regenerated the baseline (`art-dupl baseline`), which silently dropped from 17 to 15 entries because the directives now suppressed 2 groups before the baseline recorder could see them.

This created a **split-brain**: future sessions can't tell which clones are accepted via baseline vs directives. The baseline regeneration corrupted the acceptance record because it operates on post-suppression output.

### Why this matters

The baseline is the CI gate (`art-dupl check`). The directives are in-source. When both exist:

- The baseline file shrinks (directives hide clones before baseline recording)
- Removing a directive later silently re-surfaces a clone that the baseline no longer covers
- The `dedup-acceptance.md` rationale and the directive rationale can diverge

### Suggested fix

When `art-dupl` detects a clone group that has BOTH a baseline entry AND an `//art-dupl:accept` directive, emit a warning:

```
⚠ Clone group hash=550f4ba1... is accepted via BOTH baseline and inline directive.
  Consider removing the directive (baseline already covers it) or removing the
  baseline entry (directive already covers it) to avoid split-brain acceptance.
```

Additionally, `art-dupl baseline` should either:

- Refuse to run when `//art-dupl:accept` directives exist (require `--force`), OR
- Record directives as a separate `"directive_accepted": true` field so the baseline tracks the full picture.

---

## Finding 2 (MEDIUM PRIORITY): `t.Parallel()` + setup-line groups dominate the -t 2 report

### The problem

At `-t 2`, **9 of 48 clone groups (19%) are `t.Parallel()` followed by a one-line setup** — totaling ~190 individual occurrences. These are the standard Go test idiom:

```go
func TestFoo(t *testing.T) {
    t.Parallel()
    // one setup line: NewTestRegistry(), t.TempDir(), context.Background(), etc.
}
```

`--suppress-test-low` defaults to `true`, yet these still surface. They are structurally clones (3 duplicated statements) but semantically zero-maintenance boilerplate. Each test independently chooses its setup; no extraction is possible or desirable.

### Impact

These 9 groups bury the 24 genuinely actionable production-code groups. The report is ~530 lines long, and ~190 of those lines are `t.Parallel()` occurrences. A user scanning for real duplication has to mentally filter all of them.

### Suggested fix

Add a built-in suppression for the `t.Parallel()` + single-setup-line pattern in test files. Classify it as `test-boilerplate` (like the existing `--suppress-test-low` category) with a lower priority. The pattern is:

```go
func TestX(t *testing.T) {
    t.Parallel()
    <one statement>
```

This is a 3-statement clone by structure, but it's the universal Go test preamble. Suppressing it would cut the `-t 2` report from 48 groups to 39 and remove ~190 low-value lines.

---

## Finding 3 (LOW PRIORITY): `snaps_clean_test.go` TestMain is structurally inextractable — detect and suppress

### The pattern

The go-snaps library requires each test package to define `TestMain`:

```go
func TestMain(m *testing.M) {
    code := m.Run()
    snaps.Clean(m)
    os.Exit(code)
}
```

This appeared as a 16-file clone group (16 separate Go modules). It's structurally identical and completely inextractable — Go's `TestMain` must be defined in-package, one per package. No helper, no shared function, no import can eliminate it.

### Suggested fix

Detect the `TestMain(m *testing.M)` signature + `snaps.Clean` call and classify it as `library-required-boilerplate` rather than `actionable`. This is analogous to how generated code is auto-excluded — it's code the library forces you to write.

---

## What Works Well

### `--explain` output

The classification (`type-1` vs `type-2`, `actionable` badge, `assignment`/`expression`/`unknown` category) is genuinely useful for triage. The extractability hint ("Review and extract common logic" vs "Consider extracting to shared test utility") helped me prioritize quickly.

### `--no-accept-directives` debug flag

This let me verify my directives were actually suppressing clones (run with → 2 groups, run without → 0 groups). Essential for debugging acceptance behavior. Good design.

### `art-dupl check` (CI gate mode)

Clean separation between "report everything" (`art-dupl`) and "report only new clones" (`art-dupl check`). The baseline diff approach is the right CI pattern. The exit code semantics (0 = no new clones) are correct for gating.

### Type-aware detection performance

1358 files with full `go/types` checking completed in ~30 seconds. Acceptable for a monorepo of this size. The cache (`--incremental`) would help for repeated runs.

### `--sort total-tokens`

The right default for triage — highest-impact groups first. This put the 16-file `snaps_clean_test.go` group at the top, which is the correct priority for volume (even if it turned out to be inextractable).

---

## Summary of Suggestions

| Priority | Suggestion                                                                 | Effort |
| -------- | -------------------------------------------------------------------------- | ------ |
| HIGH     | Warn on baseline + directive conflict (split-brain detection)              | Small  |
| HIGH     | `art-dupl baseline` should detect directives and warn/merge                | Medium |
| MEDIUM   | Suppress `t.Parallel()` + single-setup-line as test-boilerplate            | Small  |
| LOW      | Detect and down-rank `TestMain` + library-required `Clean` calls           | Small  |
| LOW      | Consider a `--diff-baseline` mode showing what changed since last baseline | Medium |
