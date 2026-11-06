# Status: GitHub Issues #9–#14 Resolution Session (2026-11-06)

**Scope:** Review and fix all open GitHub issues on LarsArtmann/art-dupl — six found
(#9 templ suppression, #10 duplicate groups, #11 mixed-parser mega-groups, #12 check
counts, #13 baseline identity, #14 templ markup flood) — plus repairing the tree,
which was red at session start.
**Baseline:** `0cfc9966` (tree broken: `go build ./...` failed). Head at close: `ab043dbf`.
**Diff:** 35 files, +1400/−46.
**Corpora:** `webphone` 36f0c6f4, pinned `nsfw-classifier@a7cec08a`, `go-cqrs-lite@10c9f987`,
`DiscordSync@4c0502c3`, `branching-flow@dbdeb0f1`, `go-paperless@1909797`, art-dupl self.
**Binaries:** before = `v0.9.1-0.20261104060928-0cfc9966` (post repo-repair, pre-changes);
after = `v0.9.1-0.20261106063509-1ce6a7b3f64e`.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                  | Evidence                                                                                                                                                                                                                                                                                                       |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **Tree repair — dependency break**: `go get -u` sweeps bumped indirect `github.com/a-h/parse` while downgrading `templ` to v0.3.1070; templ's parsers no longer satisfied `parse.Parser` (11 compile errors inside the dep) | pinned parse back to templ's declared requirement (`v0.0.0-20250122154542-74294addb73e`); `templ` v0.3.1100 checked and requires the SAME pin, so the downgrade was the right fix rather than a templ bump; `go build ./...` green                                                                           |
| 2  | **Tree repair — jsonv2 recurrence #14**: all three usual files (`internal/jsonutil/jsonutil.go`, `config/config_migrate.go`, `pkg/enum/enum.go`) sat re-migrated to `encoding/json/v2` (jsonutil even had the empty indent-block regression) | restored verbatim from `91419533`, diff-verified; `go test ./internal/jsonv2gate/` green; ledger + dependency rule added to AGENTS.md                                                                                                                                                                         |
| 3  | **#14 fixed (detection-level)**: static markup text now participates in the hash — `TextContent` node type 102 (`syntax/templ/types.go`), emitted by `transformText`, encoded by `applyTextContentType`; semantic mode collapses whitespace runs and coalesces templ's line-break-split text nodes (`appendTextChild`), exact mode hashes verbatim, structural ignores values; text nodes are NOT statement-marked (they ride the enclosing element fingerprint) | `syntax/templ/transform.go`, `transform_node.go`; CacheVersion 7→8; 5 token-level tests (`TestTextContent*`) pin the four mode behaviors + identical-markup guard; 2 BDD specs (`bdd/templ_markup_shape_test.go`); the issue's own fixture (settings/networks sections) now reports **0 groups** (was 1) |
| 4  | **#14 measured on real corpora** (detected/shown at `-t 1`): DiscordSync 3753/112 → 3729/93; nsfw-classifier 2263/147 → 2239/129; webphone 1589/44 → 1587/43; Go-only corpora byte-identical at `-t 1/2/5` (art-dupl self 1379/10, cqrs 8697/468, branching-flow 1042/6, go-paperless 220/9) | `docs/benchmarks/templ-text-precision-2026-11-06.md` + index row; every removed DiscordSync group inspected by member list (e.g. `<th>File</th>`/`<th>Author</th>`, `<span>—</span>`/`<span>:</span>`, "No storage…"/"No attachments…") |
| 5  | **#11 residual fixed (classification leak)**: Go source embedded in `.templ` files carries `Node.GoSnippet`, so classification dispatches on `isMarkup = .templ && !GoSnippet` and Go code gets the Go category table + Go fix wording | `syntax/syntax.go` (field, `Clone()`, `serial()`), `syntax/templ/snippet.go` (`markGoSnippet`, opaque fallback), `domain.ClassificationInput.GoSnippet`, `printer/actionability/clone_classify.go`, `printer/clone_processor.go`; layout test extended; pins: `TestGoSnippetNodesCarryMarker`, `TestClassifyCloneGoSnippetInTemplFile`; on the pin the 15-member `boardQuery` group went from `conditional` + "extract the shared markup block to a child templ component" to `unknown` + "Review and extract common logic" (exactly how the same fragment classifies in a `.go` file) |
| 6  | **#11 verified end-to-end on the pin**: combined mode 46 shown (issue reported 84); **0 duplicate member sets**; 0 mixed *param-fragment* groups (the 52–165-member family is gone); `--only go` drops to 41 shown. One mixed group remains: two const declarations (`excerptRunes = 600` vs `maxRunes = 80`), both statement-level units — legitimate shape, recorded as a pattern-gap observation | python member-set/mixed-parser analysis over `--explain` output; `TestFindSyntaxUnitsMixedParser*`, `trivial-fragment`, `TestMergeCombinedMatchesDropsSubMinLinesGenericsFamilies` pre-existed and hold |
| 7  | **#12 implemented**: `check` footer is `Found total N clone groups (B in baseline, W new[, S suppressed]).` | `printer.BaselineSplit` + `BaselineSplitSetter`, `TextPrinter.writeBaselineSummary`, `baselineFilterPrinter.Split()` + `SetSuppressionStats` forwarding (stats were previously dropped entirely in check mode), version footer wired into check; 3 cmd tests incl. the mixed case; e2e verified: `(2 in baseline, 0 new)`, `(1 in baseline, 1 new)`, `(0 in baseline, 0 new, 2 suppressed)` |
| 8  | **#12 arithmetic bug caught by self-review**: `Split()` computed `DetectedTotal - Shown - InBaseline`, but `Shown` already includes baseline-matched groups → suppressed was double-subtracted; new test `TestCheckSummarySuppressedCountDoesNotDoubleCountBaseline` was verified FAILING against the old formula (canary) and green after the fix; AGENTS.md formula corrected | canary run: 1 FAIL (old) → PASS (new)                                                                                                                                                                                                                                                                         |
| 9  | **#13 implemented**: baselines record `analyzer{version, cacheVersion, policy}`; `policy` = xxh3 over cache version + detection mode + type-aware + suggest-generics + max-children + detection-method SET (order-independent); `check --baseline-version-must-match` refuses cross-policy compares naming both identities | `baseline/baseline.go` (`AnalyzerIdentity`, `String`, `CompatibleWith`, `File.Analyzer`, `SetAnalyzer`), `cmd/baseline_identity.go`; 9 tests (`baseline` round-trip/legacy/compat table + cmd accept/refuse-mismatch/refuse-legacy/legacy-without-flag/policy-sensitivity/method-order); e2e messages verified (`was recorded by … (algorithm revision 8, hashing policy …)`) |
| 10 | **Design decision documented**: compatibility is the POLICY (algorithm revision + hash-affecting settings), not the version string — a same-algorithm rebuild must not invalidate a baseline; the flag stays opt-in so existing CI keeps working; legacy baselines without identity are refused only under the flag | doc comments + `--help` text; website flag table row (gate `TestWebsiteFlagsMatchBinary` green); HOW_TO_USE "baseline split + identity" section; TODO_LIST entry for a default-visible warning |
| 11 | **#9 verified at source level**: line-leading `// art-dupl:accept` is a templ `GoComment` — ran `templ generate` (templ 0.3.1020) on a scratch module and grepped the generated Go: **no marker**; the `<!-- art-dupl:accept -->` form (no nested `//`) suppresses; both already pinned in `bdd/templ_accept_directive_test.go` | scratch module `/tmp/t9gen`, generated `a_templ.go` marker-free; live CLI fixture: directive → `0 shown (1 filtered suppressed)`, control → group shown; docs section "Directives in Templ Markup" already accurate |
| 12 | **#10 verified on the pin**: 0 duplicate member sets across 46 combined-mode groups; no group at `history_page.templ` 405–425 at all (the issue's 411/421 pair) | combined-mode JSON analysis                                                                                                                                                                                                                                                                                    |
| 13 | **Docs/memory updated**: AGENTS.md (CacheVersion 8, templ text paragraph, GoSnippet paragraph, check-split bullet incl. the corrected formula, baseline-identity bullet, recurrence #14 + dependency rule), FEATURES.md (templ row, baseline row, counts intact), CHANGELOG.md (4 entries), HOW_TO_USE.md, website `cli-flags.mdx`, TODO_LIST (3 new open items + stamp), `docs/benchmarks/README.md` + new benchmark doc, doc stamps (FEATURES/HOW_TO_USE/AGENTS) | docs-freshness gate now warns only on README (pre-existing); docs count gates + website flag gate green                                                                                                                                                                                                        |
| 14 | **Verification battery**: full `go test ./... -count=1` green (multiple runs), fast boundary gate green ×3 (121s cold, 116s, 27s warm), full boundary gate green (race `./...`), `gofmt -l` clean, `go mod tidy` clean (only 2 stale parse sums removed), self-scan `-t 5` unchanged (62 detected / 0 shown), `-t 2` unchanged (5 shown) | gate transcripts; binaries rebuilt from HEAD for e2e                                                                                                                                                                                                                                                           |
| 15 | **All six issues closed with evidence** and voice-checked comments (checker: 0 FAIL, 0 WARN on each); drafts kept at `docs/drafts/2026-11-06_issue-closures.md`; `gh issue list` now empty | `gh issue close --reason completed` output for #9–#14                                                                                                                                                                                                                                                         |

**Net new tests: 21** (6 templ behavior, 2 BDD, 1 classification, 3 baseline package, 9 cmd).

---

## b) PARTIALLY DONE

1. **#14 ask #3 (a `.templ` similarity floor above `-t 1`)** — deliberately not implemented. Measurement showed the remaining markup groups are either byte-identical markup or Type-2 clones differing only in alpha-normalized locals, so a floor would suppress signal. Not evidenced by a written per-group adjudication of the residual shown set on webphone (only sampled).
2. **`check --json` does not carry the split** — the split lives on the text footer only; machine consumers still recompute in-baseline vs new. Queued in TODO_LIST.
3. **Identity mismatch is silent by default** — only `--baseline-version-must-match` surfaces it. A stderr note (no exit-code change) is queued; requires the same text/HTML-only format gate the dead-directive warning uses.
4. **`--diff-report <baseline>` has no identity guard** — it loads a baseline and compares hashes like `check` does, but the new guard is wired only into `check`.
5. **Removed groups on nsfw-classifier (18) and webphone (1) were not individually inspected** — only DiscordSync's 19 removals were enumerated (the issue's consumer). Sampling discipline, not proof of purity.
6. **Whitespace-sensitive markup** (`<pre>`, `<textarea>`) — semantic mode collapses whitespace runs, so `a  b` vs `a b` inside `<pre>` now match. Not verified against real content; deliberate tolerance in the doc comment, unmeasured trade-off.
7. **HTML/plumbing/CSV printers** — `BaselineSplit` was implemented for the text printer only; the HTML report has no baseline split, and the pre-existing group-level classification heuristic (`representativeClassification`) was not re-audited for HTML.
8. **`docs/DOMAIN_LANGUAGE.md`** — no entries for the new vocabulary (markup identity / `TextContent`, Go snippet, hashing policy, analyzer identity).
9. **`baseline.go` package doc** — the file-format example comment still shows `version/threshold/recordedAt/entries` without the new `analyzer` block.
10. **`nix flake check`** (arch-lint, vendor-hash, alloc-gate, format, SARIF golden) not run this session; relied on `go test` + boundary gates. Arch-lint impact should be nil (no config change, no new component edges), vendorHash unaffected (no new deps), but unverified.
11. **Release** — no tag, no release notes; CacheVersion 8 invalidates every consumer cache and every baseline's hash space, so consumers must re-record.
12. **Consumer-side validation** — templ-components (mid-merge, dirty) and webphone CI gates were not run with the new binary.

---

## c) NOT STARTED

1. `nix flake check` (CI mirror) and the `self-test`/SARIF-golden lanes.
2. README.md freshness pass (`Last Updated: 2026-09-28`, 39 days) — stamp-only bumps are lies; content needs verification.
3. docs-health monthly pass (due 2026-10-28, now 9 days overdue).
4. `check --json` split fields.
5. Default-visible identity mismatch warning.
6. Identity guard for `--diff-report`.
7. ADR for the baseline identity contract (policy vs version, legacy handling).
8. `SDK_DESIGN.md` update — SDK templ results changed silently with text hashing.
9. Provider (`pkg/provider`) test on a text-divergent templ fixture.
10. Allocation/timing measurement of the extra `TextContent` nodes (parse cost, alloc budgets).
11. `TextContent`/`GoSnippet` documentation in the website's templ section.
12. Fuzz seed for the text-coalescing path.
13. HTML baseline split.
14. Re-baseline guidance for consumers (CacheVersion 8 ⇒ every baseline is stale).
15. `--ignore-markup-text` opt-out (if wanted) for shape-only templ matching.
16. Pattern gap: mixed/const-declaration group (`excerptRunes` vs `maxRunes`) — `data-dominated`/`single-declaration` do not catch it.
17. Baseline file-format version bump (`1.0` → `1.1`?) now that `analyzer` exists (currently additive/omitempty).
18. Threshold mismatch handling — `check -t N` against a baseline recorded at a different threshold silently changes the compared population; not in the policy, no warning.
19. Test for `GoSnippet` surviving the gob cache round-trip.
20. Test for top-level text in a component body (non-statement text unit semantics).
21. Cleanup decision for `docs/drafts/2026-11-06_issue-closures.md` (posted comments; archive or keep).
22. Monthly `scripts/self-scan.sh` cadence + `docs/SELF_CLEAN_LEDGER.md` entry.
23. Reusable corpus before/after differ script (I wrote ad-hoc python three times).

---

## d) TOTALLY FUCKED UP (caught in-session, zero residue)

1. **Split arithmetic double-counted baseline groups** (`DetectedTotal - Shown - InBaseline`). It was masked because every test I had either had zero baseline groups or suppressed everything; found only on self-review of the diff. Fixed, canary-verified, and the AGENTS.md sentence carrying the same wrong formula was corrected.
2. **Two edit-tool slip-ups mangled function signatures** (`transformElement`, then `transformText`) by joining the signature line into the anchor — gopls diagnostics caught both within the same step. Lesson: always include the trailing `{` line in the anchor.
3. **Wrong first assumption on text nodes**: I wrote the reflow-tolerance test assuming a single `Text` node per element; templ splits text at line breaks, the test failed, and only then did I implement coalescing. Fixture-first order saved it, but the mental model was wrong on the first pass.
4. **`go get github.com/a-h/parse@<templ's pin>`** was applied before checking whether a newer templ existed that accepted the bumped parse; a follow-up `go mod download templ@v0.3.1100` showed the same pin requirement. Right answer, wrong order.
5. **Inherited damage**: the tree arrived with both the dep breakage and the jsonv2 recurrence; 11 auto-commits landed mid-session, so "which commit introduced what" is opaque (all my work is in generic daemon commits).
6. **Benchmark index staleness noticed and only partially fixed**: I added my row to `docs/benchmarks/README.md` but left the pre-existing missing rows (`sarif-snippet-share-2026-10-11`, `crawl-overhead-2026-09-29`, `sarif-alloc-2026-10-11`) undocumented.

---

## e) WHAT WE SHOULD IMPROVE

1. **A dependency sweep must be followed by `go build ./...` before the commit lands.** The `go get -u` bot broke the build and nobody noticed; `go build` is already in CI, but the daemon committed a broken tree and the next session paid for it. Consider a sweep-time build check in the bot itself.
2. **The jsonv2 recurrence class has now fired 14 times** and still has no structural fix (daemon exclusion list vs a pre-commit gate that actually runs on daemon paths). This is the single highest-frequency failure in the repo's history.
3. **Corpus verification is manual and ad-hoc.** I wrote the same member-set parser three times; a `scripts/corpus-diff.py` (before/after binary, group-set + category diff) would make precision claims cheap and reviewable.
4. **Tests must cover combination states, not extremes.** The split bug survived because tests only used 0-baseline or all-suppressed fixtures. Any new accounting feature needs a full 2×2 (baseline yes/no × suppressed yes/no) matrix.
5. **Markup identity is now a policy with no opt-out.** If a consumer needs shape-only matching, there is no knob; decide deliberately rather than by default.
6. **Observability gap: CSS constant VALUES are still unhashed.** `transformCSSProperty` hashes only the property NAME for `ConstantCSSProperty`, while the parser exposes `Value` — so `color: red` and `color: blue` collapse inside CSS blocks, exactly the class issue #6 fixed for HTML attributes. Should be filed/fixed as a follow-up issue.
7. **The `--explain` line for Go-in-templ fragments now says `unknown`.** Honest, but a `function`-ish category for signature/param units would read better and would let `signature-only` fire more precisely. Needs a category-table decision, not a quick patch.
8. **Release discipline**: CacheVersion 8 is a consumer-visible break for caches AND baselines; without a release note and a migration line in the README, consumers will see "everything is new" once (the exact failure #13 was filed about).

---

## f) NEXT ACTIONS (prioritized)

1. Run `nix flake check` on `ab043dbf` and fix anything it surfaces.
2. Cut a release (minor) with the CacheVersion bump; note the mandatory consumer baseline re-record.
3. Add `baseline`/`new`/`suppressed` to `check --json` output.
4. Default-on stderr notice when the baseline identity differs (no exit-code change).
5. Wire the identity guard into `--diff-report`.
6. Write the ADR for baseline identity (policy scope, legacy handling, version-string decision).
7. Include (or explicitly exclude with rationale) `threshold` and test-inclusion in the policy.
8. Verify `GoSnippet` survives gob cache round-trip; add the test.
9. Fix CSS constant-value hashing (`ConstantCSSProperty.Value`) — new issue + CacheVersion 9.
10. Audit HTML printer for the group-level classification heuristic and the baseline split.
11. Update `SDK_DESIGN.md` for templ text hashing; add an SDK templ text test.
12. Provider test: templ fixture whose members differ only in text → zero findings.
13. Measure parse/alloc cost of `TextContent` nodes; record in `docs/benchmarks/`.
14. Add `TextContent` + `GoSnippet` to `docs/DOMAIN_LANGUAGE.md`.
15. Update `baseline.go` package-doc format example with the `analyzer` block.
16. Decide baseline file `version` bump (1.0 → 1.1) and whether `Load` should accept both.
17. Write `scripts/corpus-diff.py` and document its use for precision claims.
18. Re-run the consumer gates (templ-components once its merge completes, webphone) with the new binary and record the delta.
19. Add a BDD spec for `check --baseline-version-must-match` (cmd-level tests exist; BDD would pin user-facing text).
20. Add a BDD spec for the `check` split footer (currently cmd-level only).
21. Add a `printer` unit test for `writeBaselineSummary` variants (0 vs >0 suppressed).
22. Verify footer precedence when both `baselineSplit` and suppression stats are set; document it.
23. Add an explicit test for `split.Total()` = in-baseline + new + suppressed invariance.
24. Revisit the docs count gate for templ node types if the CSS-value fix changes dispatch.
25. Investigate the `excerptRunes`/`maxRunes` const-declaration group (pattern gap, not detection).
26. Consider a `signature` category for `FieldList`/`FuncType`/`Field` units (Go table currently `unknown`).
27. Decide whether `TextContent` should be statement-marked for component-top-level text; add a test either way.
28. Add fuzz seeds for text coalescing (mixed whitespace, multi-line, entities).
29. Update the website templ section to describe markup identity hashing.
30. Fix the remaining `docs/benchmarks/README.md` missing rows (sarif-snippet-share, crawl-overhead, sarif-alloc).
31. Run the overdue docs-health pass (due 2026-10-28) and harvest this session's TODO entries.
32. Refresh README.md content and stamp (or record why it is not refreshed).
33. Add the analysis to `docs/SELF_CLEAN_LEDGER.md` after the next `scripts/self-scan.sh` run.
34. Check `--quiet`/`--json` interactions with the new split line (text-only by construction, but pin it).
35. Add a small unit test for `analyzerPolicyParts` output shape (golden string) so future fields are intentional.
36. Verify `analyzerIdentity` does not call into config lazily (no behavior) and costs nothing measurable.
37. Decide whether `baseline` should warn when overwriting a baseline recorded under a different policy.
38. Add the identity guard to any future baseline consumer (LSP/diff-report) via a shared helper — currently one call site.
39. Re-run the monthly self-scan cadence and record decisions.
40. Consider exporting the algorithm revision in `version --json` (useful for CI pin debugging).
41. Review whether `check` should print the analyzer identity on the *green* path for auditability.
42. Add a TODO entry for `<pre>`/whitespace-sensitive text tolerance and measure real cases.
43. Check `bdd/discordsync_regression_test.go` expectations against the new DiscordSync counts (suite green, but the fixture corpus may be synthetic).
44. Re-verify `--show-suppressed` surfaces text-divergent groups correctly (they no longer exist as groups, so nothing to show — confirm no regression in the suppression accounting).
45. Confirm the `trivial-fragment` label still fires for identical single-element runs after text hashing.
46. Document the "0 groups" fixture (issue #14 repro) somewhere durable — currently only in the BDD spec.
47. Clean up `docs/drafts/2026-11-06_issue-closures.md` (archive under docs/drafts/README or delete).
48. Add a CHANGELOG bullet for the split-arithmetic fix (currently only the feature bullet).
49. Re-check `.buildflow.yml` skip list — `golangci-lint-auto-configure` is still skipped; decide if it can be restored now that lint is green.
50. Re-read `HOW_TO_USE.md` check section against the actual `--help` output for drift.

---

## g) QUESTIONS I CANNOT ANSWER MYSELF (max 3)

1. **What belongs in the baseline policy?** I scoped it to hash-affecting settings (algorithm revision, mode, type-aware, suggest-generics, max-children, detector set). The compared *population* also changes with `--threshold` and `--ignore-tests`: a baseline recorded at `-t 1` and checked at `-t 3` can be silently green, and the reverse produces false reds. Should those be policy members (refuse under the flag) or stay out with a separate warning?
2. **Should `--baseline-version-must-match` be default-on for new baselines** (with an opt-out) instead of opt-in? Opt-in keeps every existing CI green; default-on makes the guard self-enforcing, which is what the consumer gate actually wanted.
3. **Markup identity vs shape-only matching:** do you want an explicit opt-out (`--ignore-markup-text`) for consumers who use `.templ` detection as a *shape* finder (design-system templates), or is "different text ⇒ not a clone" the settled contract? This also decides how loudly the CacheVersion 8 release note must tell consumers to re-record baselines.

---

*Session ended clean: working tree clean at `ab043dbf`, full boundary gate green
(build + vet + lint + race `./...`), all six issues closed with evidence.*
