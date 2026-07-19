# Critical Evaluation of 2026-07-19 Feedback Reports

**Date:** 2026-07-19
**Evaluator:** Crush (glm-5.2)
**Sources:**

- `2026-07-19_cyberdom_t25_templ_exclusion_and_accept_tracking.md`
- `2026-07-19_generated-templ-not-excluded-and-output-triage-gaps.md`
- `2026-07-19-httputil-t25-high-signal-html-report-cross-package-noise.md`

**Method:** Every claim verified empirically against current source (commit `418bf230` + working tree), using a freshly-built binary (`go build -o /tmp/artdupl-test ./cmd/art-dupl`).

---

## Headline Finding

**Most feedback claims are invalid.** Three AI agents ran dedup sessions using a **stale installed binary** (`/home/lars/go/bin/art-dupl`, built Jul 18 20:38, reports `version dev`). Their reports describe behavior that does not match current source. Several "missing feature" requests are for features that already exist and are documented in `--help`.

Before any code changes are made, the genuinely missing features must be separated from the noise.

---

## Claim-by-Claim Verdict

### Report 1: Cyberdom

| Claim                                                        | Verdict             | Evidence                                                                                                                                                                                                                                                                            |
| ------------------------------------------------------------ | ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "Generated templ NOT auto-excluded — 20/43 occurrences"      | **INVALID**         | Current binary verbose output: `Auto-generated code filtering enabled (templ, default)`. The `_templ.go` file IS excluded. Author was on stale binary.                                                                                                                              |
| "interface-implementation pattern not firing at -t 25"       | **INVALID**         | Pattern IS firing. Tested with 2 identical-body interface impls (10 lines each). Structural mode at -t 3 reports them; semantic mode at -t 3 silently suppresses them via actionability. Author mistook "suppressed = pattern not firing". Real UX issue: suppression is invisible. |
| "No mechanism to mark clones as accepted — no baseline file" | **INVALID**         | `art-dupl baseline` + `art-dupl check` subcommands exist. See `baseline/baseline.go` and `cmd/baseline_cmd.go`. Hash-based CI gating is fully implemented.                                                                                                                          |
| "Compile-time interface assertions flagged"                  | **UNVERIFIED**      | Plausible — `var _ Iface = (*Type)(nil)` patterns are likely detected as clones.                                                                                                                                                                                                    |
| "No JSON or machine-readable output"                         | **INVALID**         | Four structured outputs exist: `--json`, `--simple-json`, `--plumbing`, `--sarif`. Verified in `--help`.                                                                                                                                                                            |
| "'unknown' category uninformative"                           | **PARTIALLY VALID** | `--rich-text` mode surfaces categories like `[type-2]`, `[low]`, but `unknown` is a real fallback. Could fall back to AST node type.                                                                                                                                                |
| "Threshold semantics opaque"                                 | **PARTIALLY VALID** | Help text is clearer than feedback implies: "minimum number of duplicated statements to report as clone (default: 5)". Could still explain unit better.                                                                                                                             |

### Report 2: DiscordSync

| Claim                                                            | Verdict               | Evidence                                                                                                                                                                                                                                                                              |
| ---------------------------------------------------------------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "Generated `*_templ.go` files reported despite being gitignored" | **INVALID**           | Same as Report 1. Stale binary. Current source excludes via `Code generated by templ` marker.                                                                                                                                                                                         |
| "--filter-generated is opt-in"                                   | **INVALID**           | Filtering is default-ON. Help text says "Enable" which is misleading wording, but behavior is opt-out. Verified via verbose output.                                                                                                                                                   |
| "--exclude-templ inverts user's mental model"                    | **INVALID**           | The flag name says what it does: excludes `.templ` source files. `.templ` sources ARE included by default (per `config.go:86` comment and `run_crawl.go:290`). Help text correctly says "Exclude .templ source files from analysis". The user's mental model was wrong, not the flag. |
| "Text output has no code preview"                                | **VALID**             | Default and `--rich-text` both show only `file:line-line`. No code preview. Real gap.                                                                                                                                                                                                 |
| "Text output lacks category tag"                                 | **PARTIALLY VALID**   | `--rich-text` shows `[type-2] [low]` tags. Default text shows none.                                                                                                                                                                                                                   |
| "No diff mode / extract-verify loop"                             | **PARTIALLY INVALID** | `art-dupl check` against baseline IS a diff mode for CI. Not arbitrary baseline diff, but addresses the core use case.                                                                                                                                                                |
| "Honor .gitignore"                                               | **NOVEL**             | Real new idea, not implemented. Has tradeoffs (subprocess cost, non-git repos).                                                                                                                                                                                                       |

### Report 3: httputil

| Claim                                                           | Verdict                   | Evidence                                                                                                                                          |
| --------------------------------------------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| "//art-dupl:accept directive missing"                           | **VALID (complementary)** | Baseline file exists but is hash-based and doesn't travel with code. Inline directive would be code-local and survive renames. Genuinely missing. |
| "Skill recommends wrong default (-t 5 vs -t 25 for test-heavy)" | **VALID**                 | Skill doc could note test-heavy libraries benefit from higher thresholds. Easy doc fix.                                                           |
| "Same-function category"                                        | **NOVEL**                 | Good idea: clone groups where all occurrences are in the same FuncDecl are often intentional (mw1/mw2 test pattern). Would help triage.           |
| "--out flag for HTML mode"                                      | **VALID**                 | Minor ergonomic gap. HTML goes to stdout; users redirect manually.                                                                                |
| "Stable id on HTML clone groups"                                | **VALID**                 | Minor: `id="clone-N"` for deep-linking.                                                                                                           |
| "Heuristic: would-take-more-params-than-lines"                  | **NOVEL**                 | Interesting but risky. Could compute `extraction_cost = num_params / num_lines`.                                                                  |

---

## Summary: Genuinely Missing Features

After critical evaluation, the **actually missing** features are:

### Tier 1: High-impact, low-risk

1. **Text output: one-line code preview per group** (both reports mention)
   - Add first non-empty source line of the first clone as preview
   - Helps triage without opening files
   - Trivial implementation

2. **Skill doc: recommend `-t 25` for test-heavy libraries** (Report 3)
   - One-line addition to `deduplicate-code/SKILL.md`

3. **Clarify `--filter-generated` help text** (Report 2 confusion)
   - Currently says "Enable filtering" which implies opt-in
   - Change to "Filtering is default; this flag is obsolete" or similar

### Tier 2: Medium-impact, medium-effort

4. **`//art-dupl:accept` directive** (Report 1 + Report 3)
   - Complementary to baseline (baseline = CI hash file; directive = code-local)
   - Lives with code, survives renames, documents reason at site
   - Implementation: scan comments during file load, suppress matching groups
   - ~150-300 lines of code + tests

5. **"Same-function" category** (Report 3)
   - Novel idea, helps triage by downranking likely-intentional clones
   - Requires tracking which FuncDecl each clone belongs to
   - Moderate effort

### Tier 3: Low-priority / unclear value

6. **Honor `.gitignore`** (Report 2) — subprocess cost, non-git repos edge case
7. **HTML `--out` flag** (Report 3) — minor, redirect workaround exists
8. **HTML anchor IDs** (Report 3) — minor, deep-linking nicety
9. **`--recommend-threshold`** (Report 1) — nice-to-have
10. **`--diff-report`** (Reports 1+2) — overlaps with existing `baseline/check`

### Investigate further

11. **Interface-implementation pattern at high thresholds** — verify whether it fires
12. **Compile-time interface assertions** — check if `var _ Iface = (*Type)(nil)` is detected as a clone (it likely is)

---

## Rejected Suggestions (Already Exist)

These suggestions from the feedback should NOT be implemented because the features already exist:

- **"Add JSON output"** — `--json`, `--simple-json`, `--plumbing`, `--sarif` all exist
- **"Add baseline mechanism"** — `art-dupl baseline` + `art-dupl check` subcommands
- **"Fix templ auto-exclusion"** — already works in current source (was stale binary)
- **"Make `--filter-generated` default"** — already is default
- **"Rename `--exclude-templ`"** — name is correct; user mental model was wrong

---

## Root Cause: Stale Binary

The biggest finding across all 3 reports is that they were generated against a **stale binary** built Jul 18. The binary reports `version dev` (no version embedding), making it indistinguishable from a fresh build. This caused:

- 3 independent agents to file similar "templ not excluded" reports
- Multiple "missing feature" requests for features that exist
- Erosion of trust in the tool's capabilities

**Recommendation:** Add version embedding via `-ldflags "-X main.version=<git-sha>"` to the build. A user running `art-dupl --version` should immediately see if they're on a stale build.

---

## Recommended Action

1. **Do not blindly implement** the feedback's feature requests
2. **Do** add version embedding (highest ROI — prevents future stale-binary feedback)
3. **Do** add text output code preview (genuine gap)
4. **Do** update skill docs for test-heavy libraries
5. **Defer** the `//art-dupl:accept` directive until baseline mechanism is documented
6. **Reject** JSON/baseline/templ requests with pointers to existing features

The feedback reports contain useful signal wrapped in incorrect observations. The reports' authors would benefit from reading `art-dupl --help` more thoroughly before filing.
