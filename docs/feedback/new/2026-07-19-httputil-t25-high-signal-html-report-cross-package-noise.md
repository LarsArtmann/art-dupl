# Feedback: Excellent test-code dedup signal at `-t 25`; HTML report shine; cross-package helper duplication flagged as noise

**Date:** 2026-07-19
**Project:** httputil — Go HTTP middleware library (`github.com/larsartmann/httputil`, ~3 700 production LOC, ~6 000 test LOC, 2 packages)
**Command:** `art-dupl --semantic --sort total-tokens -t 25 --html`
**Goal:** Drive duplication to zero; extract every harmful clone.

> Verdict: **High-signal session, low-noise report.** 11 clone groups → 9 actionable extractions → 2 accepted intentional groups. The semantic engine and the HTML report both performed well; the suggestions below are about edges, not fundamentals.

## Session Summary

Ran art-dupl at `-t 25` on a library that already had a prior dedup pass (so the baseline was not messy). The report surfaced **11 clone groups (42 tokens), 100% in test code, 0 production clones** — exactly the right signal: no production noise, all the actionable duplicates were test-fixture boilerplate.

After full iteration, the final report is **2 clone groups (2 tokens)**, both intentional and documented in the project's `AGENTS.md`:

1. `mw1` / `mw2` middleware factories in `stack_test.go` — the labels are the test's subject matter
2. `newTypedBodyHandler` defined once per package (`httputil` + `httpspec`) — dependency direction forbids sharing

**Net result:** 11 files changed, **-86 lines** removed, 8 new helpers extracted, all tests pass (`-race`), `golangci-lint` 0 issues.

## What Worked Exceptionally Well

### 1. `-t 25` is the right default for libraries with heavy test fixtures

This is the strongest signal of the session. The skill recommends `-t 5`, and prior feedback notes note that `-t 5` is where noise drops to near-zero. But on this codebase, `-t 5` reports **321 groups** — virtually all single-line idioms (`t.Parallel()`, `return n, nil`, `if !ok { return }`). `-t 25` cut straight to the 11 groups that were actually worth thinking about, and 9 of those yielded real extractions. **For library-style projects with many small test handlers, `-t 25` is a better default than `-t 5`.**

The session validates raising the default floor for projects where test fixtures dominate LOC. The art-dupl author may want to consider a heuristic: if >60% of LOC is test code, suggest `-t 25`.

### 2. Semantic mode correctly grouped Type-2 clones

The 10 occurrences of:

```go
http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusXXX)  // StatusOK, StatusTeapot, etc.
})
```

were correctly unified into a single clone group despite **different status codes**. This is exactly the Type-2 clone detection I wanted — the structure is identical, only the literal differs. After extracting `newStatusOnlyHandler(status int)`, all 10 sites collapsed to one-liner call sites.

Likewise, the CORS test bodies that differed only in input origin + expected value were correctly unified:

```go
// 4 sites, same shape, only cfg.Origins / req origin / expectation differ
cfg := CORSConfig{ AllowedOrigins: ..., AllowedMethods: ... }
middleware := CORS(cfg)
inner := newNoOpHandler()
req := newTestRequest(http.MethodGet, "/", requestOrigin)
rec := newRecorder()
middleware(inner).ServeHTTP(rec, req)
assertAllowOrigin(t, rec, wantOrigin)
```

These all collapsed into `assertCORSForOrigin(t, allowed, requestOrigin, wantOrigin)`.

### 3. The HTML report is excellent for triage

The collapsible clone groups, the per-group category/priority badges, the "test" vs "production" filter buttons, and the "10 occurrences · 10 tokens" summary line made it trivial to triage 11 groups in under a minute. The 📋 Copy button on each occurrence was unexpectedly handy for building edit operations — I copy-pasted exact source into `old_string`/`new_string` pairs without having to re-read each file.

The filter buttons (`🧪 Test` / `📦 Production`) were particularly important: confirming **0 production clones** at a glance let me focus entirely on the test-fixture extractions without worrying about behavioral risk.

### 4. Iterative re-runs converge quickly

After each batch of extractions, re-running art-dupl was fast (~1s on a ~10k LOC project) and the report shrank monotonically: 11 → 7 → 2. No "whack-a-mole" where fixing one clone introduced two new ones. The only post-extraction re-detection was expected: my newly-extracted `newTypedBodyHandler` helper is structurally identical in `httputil/testutil_test.go` and `httpspec/handlers_test.go` — but that's a **correct observation**, not a false positive.

## What Surprised Me (Mildly Negative)

### 1. Cross-package helper duplication is structurally unresolvable — but still reported

**Group:** `newTypedBodyHandler` defined in both `testutil_test.go` (package `httputil`) and `httpspec/handlers_test.go` (package `httpspec`).

These are byte-identical bodies because they do the same thing — but they CANNOT be shared:

- `httputil` is the parent package
- `httpspec` is a subpackage of `httputil`
- Go's dependency direction means `httputil` CANNOT import from `httputil/httpspec`
- The only way to share would be to create a third internal `testutil` subpackage — significant overhead for two 6-line helpers

art-dupl correctly identifies this as a clone. It IS a clone. But it's **structurally unactionable** in Go. The report offers no way to express "I have reviewed this and it is intentional." The user has to write a `// accepted duplication: ...` comment elsewhere (in my case, in `AGENTS.md`) that the tool does not read.

**Suggestion:** Support a `//art-dupl:accept` or `//nolint:art-dupl` directive at the clone site. When the tool re-encounters the same clone, it could read the directive and omit the group from the report (or show it in a separate "Accepted (N)" section). This would let the user converge to a literally zero-line report and make CI gating possible.

The directive format could be cheap to implement: a comment on the line of the first occurrence, optionally with a reason:

```go
//art-dupl:accept cross-package helper duplication is structural in Go
func newTypedBodyHandler(contentType, body string) http.Handler { ... }
```

### 2. "Integer-labeled" test fixtures are reported as clones

**Group:** `mw1` and `mw2` in `stack_test.go`:

```go
mw1 := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        order = append(order, "mw1-before")
        next.ServeHTTP(w, r)
        order = append(order, "mw1-after")
    })
}

mw2 := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        order = append(order, "mw2-before")
        next.ServeHTTP(w, r)
        order = append(order, "mw2-after")
    })
}
```

art-dupl correctly identifies these as clones — they ARE structurally identical after alpha-normalization of the label literals. But the **label is the entire point of the test**: this test asserts that `mw1-before, mw2-before, handler, mw2-after, mw1-after` is the observed execution order. Parameterizing the helper (`newOrderRecordingMiddleware(name string)`) would obscure that the labels `mw1` and `mw2` are intrinsic to the assertion.

This is a generalizable pattern: when a clone group is **inside a single function** and the differing identifiers are the subject of an assertion later in the same function, extraction usually harms readability. The tool cannot know this without data-flow analysis, but the **single-function** heuristic is cheap and would let the report down-rank such cases.

**Suggestion:** Add a "same-function" category/priority. Clone groups where all occurrences live in the same function (start..end line ranges overlap a single `ast.FuncDecl`) could be tagged `🟡 same-function` instead of `🟢 low`, signaling "probably intentional — review carefully before extracting."

### 3. The `--html` mode writes the report to stdout, not to a file

The skill says "View the HTML output directly — do not save it as a file." In practice, I had to redirect to a file (`art-dupl --html > /tmp/report.html`) and then open it. For a CLI tool, a flag like `--html --out=report.html` (or auto-writing to `art-dupl-report.html` in the cwd) would be more ergonomic than piping.

Minor nit; not a real problem once you know the trick.

### 4. The `code-N-M` element IDs in the HTML are not anchored

When I wanted to send a colleague a link to clone group #6 specifically, I had no URL fragment to use (`#clone-group-6` or similar). Each `<div class="clone-group">` would benefit from an `id="clone-N"` attribute so deep-linking works.

Very minor.

## Metrics

| Metric                             | Value                             |
| ---------------------------------- | --------------------------------- |
| Clone groups at `-t 25` (initial)  | 11 (42 tokens)                    |
| Clone groups at `-t 25` (final)    | 2 (2 tokens, both accepted)       |
| Clone groups at `-t 5` (reference) | 321 (overwhelmingly single-line)  |
| Production clones                  | 0                                 |
| Test clones                        | 11 → 2                            |
| Real duplications found            | 9                                 |
| Extracted helpers                  | 8 (6 in httpspec, 2 in httputil)  |
| Net LOC change                     | -86 lines                         |
| False positives at `-t 25`         | 2 (both intentional, documented)  |
| Estimated FP rate at `-t 25`       | 18% (2/11)                        |
| Estimated FP rate at `-t 5`        | >95% (idiom noise)                |
| Time to triage 11 groups           | ~1 minute (HTML filters + badges) |
| Time to full iteration to zero     | ~10 minutes (extract + verify)    |

## Suggestions (Prioritized)

### 1. Inline `//art-dupl:accept` directive (HIGH IMPACT, LOW RISK)

Allow users to mark intentional clones at the site. When re-running, omit accepted groups from the active report (or list them under a separate "Accepted (N)" section). This is the single biggest unlock for CI gating and "drive to zero" workflows — without it, the terminal state of a dedup session is always "N clones remain, trust me they're fine."

Implementation sketch:

- On each clone group, compute a stable hash from the AST fingerprint (you already have this).
- Before emitting, scan each occurrence's preceding comment lines for `//art-dupl:accept[:<hash-prefix>] [<reason>]`.
- If found, suppress the group (or move to "Accepted" section).
- The hash-suffix is optional but prevents stale accepts after the code changes.

### 2. Consider raising the documented default for test-heavy libraries (MEDIUM IMPACT, LOW RISK)

The skill currently recommends `-t 5` as a sensible default. For projects where >60% of LOC is test code, `-t 25` produces dramatically better signal-to-noise (18% FP rate vs >95%). Consider:

- Adding a note to the skill: "For library projects with heavy test fixtures, try `-t 25` first."
- Or auto-detecting: if the proportion of `_test.go` (or equivalent) LOC exceeds 50%, emit a hint in the report summary: "ℹ️ Test-heavy codebase (62% test LOC). Consider `-t 25` for cleaner signal."

### 3. "Same-function" category for clone groups (MEDIUM IMPACT, LOW RISK)

When every occurrence of a clone group lives inside the same `ast.FuncDecl`, tag it as `🟡 same-function` rather than `🟢 low`. These cases are very often intentional — the duplicated fragments implement a parallel structure (e.g., mw1 vs mw2 in an ordering test, or two columns of a comparison matrix) where extraction would obscure intent.

### 4. `--out=<path>` flag for HTML mode (LOW IMPACT, LOW RISK)

Ergonomic improvement. Auto-detecting a sensible default (`art-dupl-report.html` in cwd when `--html` is passed and stdout is a TTY) would remove the redirect dance.

### 5. Stable `id` attributes on HTML clone groups (LOW IMPACT, LOW RISK)

Add `id="clone-group-N"` to each `<div class="clone-group">` so deep links (`file://.../report.html#clone-group-6`) work for sharing specific findings.

### 6. Heuristic: warn on "would-take-more-params-than-lines" extractions (LOW IMPACT, MEDIUM RISK)

A common false-positive class is a 4-line pattern where extraction would require a 3-parameter helper. The skill itself calls this out ("An abstraction would take more parameters than the duplicated code has lines"). The tool could compute a simple `extraction-cost` score: `num_parameters / num_lines_in_clone`. When this exceeds 1.0, the clone is unlikely to benefit from extraction and could be down-ranked or annotated `🟡 low-value-extraction`.

This is a heuristic, so there's a risk of false negatives — but it would have saved me from briefly considering the `mw1`/`mw2` extraction.

## Conclusion

This was a clean, productive session. At `-t 25`, art-dupl on a test-heavy library gave **9 real findings and only 2 accepted duplicates** — an 82% actionable rate, which is excellent. The HTML report made triage fast, and the semantic engine correctly grouped Type-2 clones with varying literals.

The two remaining clones are **structurally unactionable** in Go (cross-package test-helper duplication) and **intentionally unactionable** (label-parameterized test middleware). Both are correctly detected; what's missing is a way to tell the tool "I've reviewed these, don't report them again." That single feature — an inline `//art-dupl:accept` directive — would unlock CI gating and let "zero clones" become a literally achievable state rather than an aspirational one.

For library authors with heavy test suites, the unstated best practice is `-t 25`, not `-t 5`. The skill documentation would benefit from saying so explicitly.

---

**Author:** Crush on behalf of the `httputil` deduplication session.
**Session hash:** 11 → 2 clone groups in one pass.
