# Status Report: Templ Follow-Up — Verification, Gates, and the Queue Close-Out

- **Date:** 2026-10-04 10:51
- **Session scope:** the follow-up queue from `2026-10-04_08-28_templ-expression-aware-detection.md`
  (report tasks 1–9, 12, 18 of 22) — verification gates, missed tests, ADR, benchmark,
  BDD guard, docs. No new feature work.
- **End state:** full suite green (33 pkgs), `golangci-lint` 0 issues, `go-arch-lint` OK,
  boundary gate OK, `nix flake check` **all checks passed** (final run, settled tree).

---

## a) FULLY DONE

1. **Re-verification after the classify split** (report tasks 1–2). The full suite and
   full lint had never run after the `nodeTypeToCategory` split. Lint caught a REAL
   `exhaustive` finding in `templSuggestion` (missing cases in the switch over
   `domain.CloneCategory`) — fixed by listing the 12 remaining categories explicitly,
   matching the `calculateProductionPriority` convention (`printer/actionability/
   clone_classify.go:84`).
2. **Classify unit tests** (report task 3): `TestTemplNodeTypeToCategory` covers every
   templ constant → category (29 entries incl. unknowns); `TestTemplSuggestion` covers
   every templ-specific branch + Go-wording passthrough;
   `TestNodeTypeToCategoryFileKindCollision` pins the hazard that forced the table
   split. **Correction:** my first pin guess (`golang.BasicLit == templ.TagStart == 5`)
   failed; probing showed the real collision is `golang.BasicLit == templ.Element == 4`.
   The test now asserts the collision exists (fails loudly if the enums drift apart).
3. **`nix flake check` closed** (report task 4) — and it caught a REAL sandbox-only
   failure: `TestFuzzTargetsHaveSeedCorpora` counted the fuzz targets of
   `gogenfilter-real/fuzz_test.go` (upstream source the flake materializes at repo root
   during `preBuild`) as ours. `go test ./...` was green locally because that directory
   only exists inside the nix sandbox. Fix: `collectFuzzTargets` skips nested Go modules
   (`isNestedGoModule` = directory with its own `go.mod`), `bareFuzzTargets` extracted
   with root threading, regression test `TestCollectFuzzTargetsSkipsNestedModules`
   added (`cmd/fuzz_seed_gate_test.go`). Durable rule recorded in AGENTS.md: any
   test-time gate that walks the tree must skip nested `go.mod` directories.
4. **ADR-0027** (`docs/adr/0027-expression-aware-templ-detection.md`): the eight
   diagnosed defect classes, the `golang.ParseSnippet` bridge contract, the enum ≥100
   policy, the file-kind-scoped switch rule, fallback semantics, alternatives, corpus
   consequences, and verification gates.
5. **`BenchmarkTemplParse` + committed baseline** (report task 5): three-mode benchmark
   over a representative component file (declarations, if/else, for, switch with
   default, conditional attributes, CSS template, `{{ goCode }}`, spread-style attr
   expressions). Baseline: `docs/benchmarks/templ-parse-2026-10-04.txt` (6 runs,
   benchstat format; ~250–365 µs/op, ~5.7k allocs/op).
6. **BDD false-positive guard** (report task 7): `bdd/templ_clone_detection_test.go`
   new context with two specs — differing conditions must NOT produce the if-rooted
   clone, renamed locals must still clone. See (e)/d) for how the assertion evolved.
7. **Spread-attribute verification** (report task 9): confirmed `{ user.Attrs... }`
   expression children participate in matching — div composite fingerprints provably
   differ for divergent spread sources (in-process: `-4056ae9d` vs `-13dcee9d`); added
   `TestSpreadAttributeExpressionDifferencesPreventClone` +
   `TestRenamedSpreadReceiverStillClones` to `syntax/templ/behavior_test.go`.
8. **Docs updated:** AGENTS.md (clone-categories bullet rewritten for the file-kind
   split, new fuzz seed-gate bullet with the nested-module rule, File+Decl artifact
   documented, templ bullet extended), HOW_TO_USE.md (templ-specific `--explain`
   suggestions note), FEATURES.md verified current (no stale claims found).
9. **Report tasks 12 and 18 swept:** TODO_LIST.md has no stale templ entries;
   FEATURES "Known Limitations" wording is still accurate. Nothing to change.

## b) PARTIALLY DONE

1. **Templ benchmark is informational, not gated.** The baseline is committed, but
   `BenchmarkTemplParse` is NOT in `scripts/alloc-budgets.txt` / the alloc-gate —
   it will not catch a regression until wired in. Deliberate: budgets should be set
   after a few more baseline samples on this thermally noisy machine.
2. **BDD templ coverage remains thin** (report task 11): the new guard covers the
   condition false-positive class only. CSS-duplicate-output, conditional attributes,
   and exact/structural modes have token-level tests but no end-to-end specs.
3. **The `-t 1` File+Decl artifact is documented, not fixed** — see (e)1.
4. **Corpus numbers not re-documented** (report task 8): go-sse 17 vs baseline 16,
   go-cqrs-lite 347 vs baseline 411 remain undocumented as a dated re-baseline —
   blocked on your re-baseline-policy answer (question g2).

## c) NOT STARTED (from the 08-28 report's queue, still open)

| Report task | Item                                                                                     |
| ----------- | ---------------------------------------------------------------------------------------- |
| 8           | Corpus re-baseline docs (blocked on g2)                                                  |
| 10          | `SnippetGoFile` position-clamping edge test (top-level Go in the first bytes)            |
| 11          | TESTING.md: decide the owning layer for templ behavior tests                             |
| 13          | jsonv2 recurrence escalation (blocked on g3)                                             |
| 14          | `--dump-tokens`: decoded names for templ statement composites (currently bare `11`/`12`) |
| 15          | Templ-tuned priority evaluation (sample 20 groups from templ-components)                 |
| 16          | Mixed Go+templ group classification rule: define + test + document                       |
| 17          | `ScriptTemplate.Parameters` expression parsing                                           |
| 19          | Self-scan ledger for the new templ code (`scripts/self-scan.sh`, due 2026-10-22)         |
| 20          | Evaluate caching `ParseSnippet` per (content-hash, kind) — only with bench data          |
| 21          | Website templ feature page: mention expression-aware detection                           |
| 22          | Upstream a-h/templ `ConstantCSSProperty` Range check; bump if fixed                      |

## d) TOTALLY FUCKED UP (process cost this session — nothing shipped is broken)

1. **Concurrent nix build vs. tree editing.** A scratch probe file with a compile
   error sat in the tree for ~2 minutes while a background `nix flake check` was
   running; the bench derivation snapshotted the broken file and failed red. The
   failure was pure artifact — the fourth run on the settled tree passed. Lesson:
   NEVER edit the working tree while a nix build from that tree is in flight.
2. **The gate I was fixing caught my own test.** The fuzz-gate regression test wrote
   literal `func FuzzOurs(f *testing.F)` strings; the gate's scanner reads ALL
   `_test.go` sources and counted them as bare targets. Fixed with a `fuzzDecl`
   concatenation helper — and the gotcha is now in AGENTS.md.
3. **Three fixture iterations for the benchmark:** wrong templ syntax (`{% if %}`
   blocks don't exist — templ uses bare keywords), then a missing `</div>`, plus a
   broken (unbalanced) bisect along the way. Should have read `behavior_test.go`
   fixtures FIRST.
4. **Wrong initial collision pin** (TagStart/5 instead of Element/4) — I counted iota
   values by hand instead of probing. One wasted test cycle.
5. **BDD assertion needed three attempts:** file-absence was wrong (the shared body
   is a legitimate clone), `NotTo("if user.")` was weak (subsumption hides if-rooted
   groups even in true positives). Only the line-range pin (body `6-7` present, div
   `4-9` absent) actually discriminates pre-fix from post-fix. Both premises were
   verified against real CLI output before landing.
6. **The "spread-attribute gap" false alarm:** ~30 minutes of debugging a
   "fingerprint identical" conclusion that came from misreading `--dump-tokens` (it
   prints the encoded type hash, NOT the composite fingerprint). The in-process
   comparison immediately disproved it. Lesson: dump-tokens is not a fingerprint
   oracle; compare `Val()`/`Fingerprint` directly.

## e) WHAT WE SHOULD IMPROVE

1. **Drop the File token from the templ token stream** — it carries zero content and
   can only create phantom `[File, ComponentDecl]` whole-file prefix matches at low
   thresholds (newly documented artifact; length-2, so invisible at default `-t 5`).
   Cost: CacheVersion bump + baseline churn. Needs your call (g1).
2. **Never overlap tree edits with nix builds** (d1). A `nix flake check` should be
   the LAST thing after the tree settles, not run concurrently with fixes.
3. **Read existing fixtures before writing new ones** — the templ syntax stumbles
   (d3) were avoidable.
4. **The templ benchmark should join the alloc-gate** once the baseline has enough
   samples; right now a templ parse regression would be invisible to CI.
5. **Fuzz corpus is starved:** `FuzzParseBytes` has exactly ONE committed seed. The
   seed gate requires ≥1, but one seed is nominal coverage for an expression-aware
   parser.
6. **golangci-lint's own cache still points at the corrupted `/mnt/buildcache`**
   (warning spam every run). Cosmetic — lint results were correct — but the warning
   noise invites glossing over real warnings.
7. **The BDD layer should own user-visible templ behavior** end-to-end; the
   token-stream layer keeps implementation pins. Undecided (report task 11).
8. **`--type-aware`, `--suggest-generics`, and combined mode are UNVERIFIED against
   the new templ expression children** — they thread through the same transformer,
   but no test exercises templ files under them.

## f) NEXT 50 (prioritized; ★ = needs your decision, ✓ = this session's additions)

**Correctness-adjacent / quick wins**

1. ★ Decide + implement File-token drop from the templ stream (kills the `-t 1`
   phantom; CacheVersion 5 → 6).
2. `SnippetGoFile` position-clamp edge test (report task 10).
3. Verify `--type-aware` on templ files end-to-end (snippet expressions + VarType).
4. Verify `--suggest-generics` on templ files (does candidacy make sense?).
5. Verify combined mode (ADR-0026) merge semantics with templ inputs.
6. `FuzzParseBytes`: add ~10 more committed seeds (spread, cond-attr, GoCode,
   nested components, CSS, switch/fallthrough, else-if chains, hybrid GoCode).
7. Audit OTHER self-scanning test gates for the literal-in-test hazard class (d2).
8. GoCode inside attribute values — probe whether templ tolerates it and what we emit.
9. Nested component calls `@Card(@Inner())` token behavior test.
10. else-if chain in templ: full-chain structure test (Go-parity already pinned for
    2 links; verify 3+).
11. `ChildrenExpression` (`{ children... }`) test coverage.
12. `RawElement` (style/script blocks) expression handling verification.
13. CSS class usage `class={ panelStyle().Class() }` handling test.
14. Incremental cache: manually verify v4 entries are actually dropped by
    CacheVersion 6... 5 (one-run check, not assumed).

**Coverage & verification**
15. Wire `BenchmarkTemplParse` into the alloc-gate budgets (after more samples).
16. BDD: CSS duplicate-output regression spec (the old "printed twice" bug class).
17. BDD: conditional-attributes spec (cond + both branches end-to-end).
18. BDD: exact/structural mode specs for templ end-to-end.
19. Mixed Go+templ group classification: define the rule, test it, document (task 16).
20. `ScriptTemplate.Parameters` parsing (task 17).
21. `--dump-tokens` decoded names for templ composites (task 14).
22. Templ priority evaluation: sample 20 real groups from templ-components (task 15).
23. SARIF output spot-check for templ groups (decoded names, levels).
24. HTML report visual check for templ clone groups (anchors, suggestions).
25. go-finding adapter: pin templ GroupID stability.
26. Provider (BuildFlow SDK) downstream check with expression-aware templ output.
27. Windows CI lane: confirm the templ changes pass without GOEXPERIMENT.
28. Baseline/check commands with templ files end-to-end.
29. Self-scan ledger entry for the new templ code — calendar 2026-10-22 (task 19).
30. Monthly self-scan: judge the new templ/ and golang/snippet code groups.

**Performance**
31. Benchstat harness alert for templ parse regressions (baseline exists now).
32. Evaluate `ParseSnippet` caching per (content-hash, kind) (task 20) — only if
benchmarks show templ parse cost matters in real runs.
33. CPU-pinning A/B for the templ benchmark on this machine (stability, per
CPU_TOPOLOGY.md — pinning is for CI variance, not speed).
34. Measure real-run templ overhead: templ-components corpus wall-time before/after.

**Docs & release**
35. ★ Corpus re-baseline now vs. 2026-10-28 docs-health pass (g2, task 8).
36. ★ jsonv2 escalation: catch-and-restore forever, or BuildFlow upstream hook (g3).
37. ★ `_sources/` vendored-copy pairs (templ-components): default-exclude, flag, or
surface? (carried from the 08-28 report's question 2).
38. TESTING.md: owning layer for templ behavior tests (task 11).
39. README known-limitations: mention the `-t 1` File+Decl artifact until fixed.
40. CHANGELOG entry for expression-aware templ detection (user-facing behavior change).
41. ★ Version/release decision: this is a behavior change — v0.8.0 via go-release flow?
42. Website templ feature page: expression-aware as differentiator (task 21).
43. Upstream a-h/templ: check `ConstantCSSProperty` Range fix; bump if available (task 22).

**Hygiene**
44. Pre-commit "tests green at commit time" hook or daemon exclusion list for
invariant files (the jsonv2 root cause, report e1).
45. Redirect golangci-lint cache away from the corrupted `/mnt/buildcache`.
46. `fuzzSeedExemptions` placeholder comment: replace with a real doc link or remove.
47. Share BDD/stream test helpers via `internal/testutil` if templ BDD grows.
48. `--min-tokens` docs: clarify composite tokens count as ONE in templ.
49. Evaluate `CategoryHandler` mapping for templ `ScriptDeclaration` (currently
`function`; script handlers might deserve `handler` priority).
50. Gopls stale-diagnostic lag after edits (exhaustive warnings persisted minutes
after the fix) — investigate or restart client automatically in workflow.

## g) THREE QUESTIONS ONLY YOU CAN ANSWER

1. **File-token phantom:** should I drop the File token from the templ token stream
   (strictly hygiene — it only creates whole-file phantom matches at `-t 1..2`; costs
   a CacheVersion bump and baseline churn), or document it as acceptable threshold-1
   noise and leave the stream alone?
2. **Corpus re-baseline timing:** re-baseline go-sse / go-cqrs-lite / templ-components
   NOW with a dated note separating the templ-change delta from corpus drift, or wait
   for the scheduled 2026-10-28 docs-health pass? The stale numbers will confuse the
   next status report either way. (Carried from the 08-28 report.)
3. **jsonv2 recurrence escalation:** after 4 identical daemon-packaged regressions, is
   jsonv2gate + catch-and-restore the permanent control, or do you want the root cause
   (auto-commit daemon packaging in-flight edits) taken to BuildFlow upstream — e.g.
   a pre-commit test hook? (Carried from the 08-28 report.)

---

_All session changes are auto-committed on `fork` by the daemon; working tree clean
except `syntax/templ/behavior_test.go` at report time (latest spread tests)._
