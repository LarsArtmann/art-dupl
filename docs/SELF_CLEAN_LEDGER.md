# Self-Clean Decision Ledger

Durable per-group record for the `-t 1` self-scan, so future sessions do not
re-litigate accepted clone groups. Routine: `scripts/self-scan.sh` (first run
of each month) — judge every shown group, append a dated section here.

Convention per entry: where the clone lives, why it is accepted (or what was
extracted), and the date of the judgment. Entries never get deleted; a group
that reappears in a later sweep with a _changed shape_ gets a fresh entry.

---

## 2026-09-19 — baseline `-t 1` self-scan

**Shown groups: 40** (re-measured 2026-09-19 with the current tree; the
"44" quoted in the 2026-08-16 status report was a file-count heuristic, not a
group count). All 40 judged as accepted boilerplate/intentional; detailed
per-group rationale lives in the 2026-08 self-scan sessions (see
`docs/status/2026-08-16_*` reports and CHANGELOG).

## 2026-09-22 — full `-t 1 --type-aware` sweep

**Detected: 1061 / shown: 30** at `-t 1 --type-aware`. Every shown group
judged (details: `docs/status/2026-09-22_21-47_dedup-sprint-self-scan-t4-t1.md`).

### Extracted (harmful)

| Clone                                                                                   | Action                                                                                                                                                                            | Evidence          |
| --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `isAssignFromMRun` / `isOSExitCall` (`printer/actionability/actionability_preamble.go`) | Extracted shared `hasMethodCall(call, recv, method)` walker; both are now thin wrappers. `hasCommandReceiver` deliberately NOT merged (set-of-receivers is a different contract). | commit `27438816` |

### Accepted (intentional, with rationale)

| Pattern                                                                    | Count   | Rationale                                                                                                                                                  |
| -------------------------------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Enum `IsValid()` return-true tails                                         | 14      | Documented enum convention: per-type explicit valid sets beat a shared table for diff-ability.                                                             |
| `close(schan)` pairs                                                       | 3       | Documented ctx-propagation lifecycle (sender closes; mirrored close sites).                                                                                |
| Nil-guard + lock pairs in `cmd/filter_stats.go`                            | several | Idiomatic guard clauses; `withLock[T]` already centralizes the general case.                                                                               |
| Flag-setter pairs (`cmd/`)                                                 | few     | Cobra flag wiring is positional by nature; a setter table would obscure help text order.                                                                   |
| Single `default:` clauses                                                  | few     | Switch-shape boilerplate, extraction adds indirection for 2 lines.                                                                                         |
| `case golang.IfStmt:` rows in the actionability denylist table             | several | Documented first-match-wins pattern table; rows are data, not logic.                                                                                       |
| Ukkonen `if oldr != t.root` pair (`suffixtree/`)                           | 1       | Hot construction path, alloc-gate protected; branch differs semantically at each site.                                                                     |
| `countSerializedNodes` / `serial` mirror (`syntax/`)                       | 1       | Documented must-stay-in-lockstep invariant (arena under/overflow guard).                                                                                   |
| `statsChan <- stats` finalize pairs (`cmd/`)                               | 1       | Verified buffered(1) + single terminal send cannot block.                                                                                                  |
| Min-vs-max loops (`cmd/run_output.go`, `printer/extractability_engine.go`) | 1       | Opposite semantics, cross-package; stdlib `slices` already covers the concept where it fits.                                                               |
| Templ view markup siblings (`printer/`)                                    | few     | Structural similarity of templ components is the format's grain, not duplication.                                                                          |
| BDD helper guards (`bdd/`)                                                 | few     | Ginkgo setup boilerplate; readability beats DRY in specs.                                                                                                  |
| filter_stats warning loops (exclude/include symmetry)                      | 1       | Deliberate exclude/include symmetry; four `//art-dupl:accept deliberate exclude/include symmetry` directives re-anchored to the loops (commit `3873decf`). |

**Result of the sweep:** 1059 detected / 28 shown after the extraction; no new
groups introduced. Sort-order churn between runs is expected (parallel search
output order is nondeterministic — documented in ADR-0019).

## 2026-09-25 sweep (C3, post-B1 gitignore work)

Scan: `scripts/self-scan.sh` (`-t 1 --type-aware`) → 1078 detected / 28 shown
(after fixing the one group this change-set introduced; the raw scan first
reported 29 shown).

**Extracted this sweep:**

- `internal/gitignore/gitignore.go` — the duplicated `if len(rules) == 0 {
  return nil }` tail across `LoadGitignore` and the new `LoadTree` collapsed
  into `matcherFromRules` (the new group this scan caught in our own B1 work;
  found by our own tool the same day it was written).

**Accepted (existing idiom classes, unchanged from the 2026-09-22 table):**

- `case golang.IfStmt:` rows, single `default:` clauses (actionability
  pattern-table rows are data, not logic).
- `child.Statement = true` pairs across `syntax/golang` / `syntax/templ`
  (transformer shape mirrors the two ASTs; a shared helper would need the
  two node types to unify, which is the ROADMAP syntax-facade cut).
- `statsChan <- stats` finalize pairs, `if i > maxChildren` arena guards,
  `if split` match pairs, BDD `if s.T != nil` guards (test-support
  boilerplate), filter_stats guard pairs, min/max cross-package loops — all
  carry their prior rationales in the table above.

## B7 — SDK-path self-scan (provider crawl on art-dupl itself)

Ran the toolsdk provider path (`toolsdk.All()` → `spec.Detect` with the
art-dupl repo as working dir) against our own tree:

- SDK path: **179 findings / 72 distinct clone groups** (GroupID-stable).
- CLI at `-t 5`: 60 detected groups, 0 shown (23 non-actionable, 37
  filtered suppressed).
- The 60→72 delta is crawl policy, not detection drift: the CLI's default
  config ignores `*_test.go`; the provider's crawl deliberately does NOT
  (BuildFlow parity — jscpd also scans test files). Consumers comparing the
  two channels must expect this difference.
- Cross-checked GroupID determinism in
  `TestFindingsFromGroups_GroupIDStableAcrossRuns` (identical GroupIDs
  across independent runs).

## 2026-09-28/29 — SUPERB docs-health completion pass (M00–M20)

Doc-debt decisions from the plan in
`docs/planning/2026-09-28_21-57_SUPERB-docs-health-completion-plan.md` — recorded so the
monthly self-scan does not re-litigate them:

- **Master plan closure**: the "24/26 done" claim resolved as 24 executed + 2
  measured no-gos (T7 defer-cleanup pattern, T14 stream-slice pool). Nothing open.
  Evidence map lives in the master-plan header and TODO_LIST.
- **GitHub-Action distribution plan**: formally won't-implement — superseded by the
  BuildFlow provider lane; revisit trigger documented on the plan file itself.
- **Legacy-archive retrofit (272→294 pre-regime files)**: not scheduled; policy and
  gates documented in `docs/status/archived/README.md`. DO NOT strike legacy files
  without per-item verification (fabricated evidence is worse than silence).
- **Count regime**: living-doc numbers are CI-enforced by
  `cmd/docs_health_counts_test.go` (patterns 37, user-visible flags 56 of 64
  registered — the 8 hidden are 6 deprecated `--include-*` aliases plus
  `profile`/`timeout` — Go node types 49, templ 29). Update the DOC to the
  derived number, never the test to the doc.
- **Accepted classes** (unchanged from prior sweeps): tolerated `_test.go` v2
  importers (gate prevents new ones), format-specific clone DTOs (ADR-0025),
  single-module layout (arch-lint DAG), table-driven skip for transform.go.
- **New standing gates**: docs count gate, SDK/CLI DefaultThreshold drift test,
  docs-freshness script (warn-only until one clean cycle), archive regime README.
- **Template flags fixed**: consumer templates referenced non-existent flags
  (`--output`, `--baseline`); corrected to `--plumbing` / `--baseline-path` and
  verified against the built binary end-to-end.

## 2026-09-29 — SUPERB v2 harden-harness pass (M00–M22)

Decisions from `docs/planning/2026-09-29_01-10_SUPERB-v2-harden-harness-unblock-ship.md`,
recorded so the monthly self-scan does not re-litigate them:

- **Count-gate 49→51 repair**: the concurrent session's Go 1.27 SelectorExpr-key
  dispatch raised transformer node types 49→51; the docs count gate demanded the doc
  change (README/FEATURES updated, gate test re-derived, canary verified failing
  first). Regime held: doc follows derived number.
- **go.mod pin flip, caught twice**: the boundary gate's first live catch (M02,
  restored in 9eb2cfd6) and again after the 03:41 daemon clobber (this pass). The
  `cmd/go_line_pin_test.go` gate + boundary gate both fire; keep both.
- **jsonutil/config_migrate v2 re-clobber (3rd recurrence)**: daemon commit ce716456
  re-migrated both files to `encoding/json/v2` — broke Duration marshaling
  (SaveConfig) and empty-indent output, exactly the 2026-09-23/24 class. Boundary
  gate caught all three files in one run; restored from canonical commits (42ee0a72,
  3b6327d5); the go-error-family v0.11.0 bump in the same daemon commit was kept.
  The gate's fast profile is now the proven tripwire for this class.
- **Daemon-swept commits (ba2d5ce5, 9a97ec82 class)**: real-message commits were
  twice pre-empted by the auto-daemon sweeping staged files under a heuristic
  message (once mixing my count fix with its lint cleanups). Disposition: never
  fight the daemon; verify content via `git log -- <file>` and move on.
- **SDK_DESIGN ↔ ADR cross-refs (F081)**: forward refs 0 broken; 13 ADRs uncited
  from SDK_DESIGN but their features are documented in it = not a defect, no action.
- **Windows exe-start probe (M18)**: skip now env-gated
  (`ARTDUPL_WINDOWS_PROCESS_TEST=1`) + continue-on-error CI probe after
  `Set-MpPreference -ExclusionPath $env:RUNNER_TEMP`. Next pass reads the probe log:
  green ⇒ delete the skip; red ⇒ exclusion hypothesis has evidence against it.
  **RESOLVED 2026-09-29**: probe log RED (run 36509027090) but the red is INVALID
  as hypothesis evidence — the probe excluded `RUNNER_TEMP` (D:\a\_temp) while the
  test binary lives under `%TMP%` (C:\Users\...\AppData\Local\Temp), and the real
  cause was found in stdlib source: `go build -o art-dupl-test` writes an
  extensionless file on windows (cmd/go appends `.exe` only when `-o` is empty,
  work/build.go:477), and os/exec `findExecutable` refuses extension-less absolute
  paths (tries PATHEXT suffixes, none exist → ErrNotFound) — the discarded
  `_ = cmd.Run()` error hid "file not found" behind "ProcessState is nil". Not AV,
  not the runner. Fix: `-o art-dupl-test.exe` + start-error surfaced via
  `errors.AsType[*exec.ExitError]`; windows skip deleted; Defender probe job
  deleted; pending windows-CI confirmation.
- **Skills upstream home (M14)**: `skills/docs-health` resolves via symlink chain
  crush-config → `~/.agents/skills` → `~/projects/SKILLS`; upstream edits go to the
  SKILLS repo (committed 2514a55 there), NOT crush-config. Annotator hardening
  (`--verify`, `--emit-keys`, dup-key refusal) lives there; push gated on Lars.
- **v0.8.0 + go-paperless stay GATED on Lars**: release execution parked in
  `docs/planning/2026-09-29_02-45_v0.8.0-release-park.md` and the go-paperless memo
  (v0.4.3 recipe); TODO_LIST #18 branch-protection recipe written for owner action.
- **Linter Pareto (F105)**: 106 enabled, 0 findings; keep all — noise is carried by
  ~6 linters with bespoke settings, verdict recorded in AGENTS.md Lint config bullet.
- **Docs-health 2026-10-05 early pass (key-grammar fall)**: per-item strikes
  landed on ZERO of ~185 files because spec keys were transcribed from agent
  verdicts (`b@T19`, `M26:yaml`, `1-15` ranges) instead of generated with
  `--emit-keys`; the annotator's atomicity refused every file (correctly).
  Fallback: dated banners (regime A-lite, now defined in
  `docs/status/archived/README.md`). Verdict corpus + driver rescued from
  /tmp into `docs/status/2026-10-05_pass-verdicts/` + `scripts/docs-health-driver.py`
  (2026-10-06) — the /tmp-vanish class the 09-28 pass already swore off.
  Rule: keys are MECHANICAL (`--emit-keys <file> <lineno...>`) — verdict
  agents quote evidence, the driver resolves keys against the file.

## 2026-10-16 — accept-directive self-run triage (`-t 3 --suggest-generics`)

First self-run after the near-miss detector (plan T15) surfaced three
warnings, and the deterministic accounting fix exposed a fourth latent one.
Two real defects, two fixture/prose classes:

- **IsAccepted accounting bug (fixed)**: `IsAccepted` returned at the FIRST
  matching clone, so directives covering a group's LATER clones were never
  recorded as matched. `printer/diff.go:227` (inline directive on the second
  clone of the accepted LCS-backtrack group {214-216, 227-229}) was falsely
  reported as a near-miss against the neighboring group; the same shape
  could produce false stale-hash reports, and warning visibility was
  clone-order and run-scope dependent (the integration-test fixture warned
  in ./cmd-scoped runs but not full runs). Fix: record every covering
  directive before returning. Pinned by
  `TestAcceptedSetAccountsDirectivesOnEveryClone` (cmd/accept_directive_test.go).
- **Prose directives in production sources (fixed)**: the scanner
  regex-matches any source line, so every doc-comment mention of the full
  marker syntax was a live bare directive — the one on `DeadDirectives`'s
  doc (internal/accept/accept.go:269) sat 6 lines above its own function's
  nil-guard clone group and warned. All prose mentions in production files
  were rewritten without the `//` prefix; runtime strings use the
  concatenated `accept.MarkerText` const (assembled so no art-dupl source
  line contains the contiguous marker). printer/finding uses an inline
  split (arch-lint: may not import internal/accept).
- **Test fixtures in scanned sources (fixed where active)**: `_test.go`
  files are analyzed by default (config `IgnoreFiles` default is empty), so
  directive fixtures inside the accept tests are scanned as real
  directives. The hash fixture at cmd/accept_directive_test.go:147 (dead
  warning), the integration fixture at :122 (near-miss in scoped runs), and
  the gofmt-style fixtures at :458/:462 (near-miss + silent accept of the
  adjacent readFile group) are now assembled by concatenation — same
  discipline as the fuzz-seed gate's fixture rule. Remaining bare fixtures
  (:187, :254, :498, …) are accepted residuals: they only double-suppress
  test-file groups that test-context patterns already suppress, and
  splicing every one would wreck fixture readability.
- **Deferred**: a quote-aware scanner (skip marker matches inside string
  literals) would fix the fixture class generally — rejected for now as
  hand-rolled-lexer risk (escapes, backtick blocks, .templ) for noise the
  warnings already surface deterministically.
- **diff.go cleanup**: the doubled inline directives at 214/219 (appended
  twice by an earlier session) were deduplicated; all four LCS directives
  (214/219/227/233) are live acceptors under the accounting fix.
- Cosmetic: `loadTypeAwareData`'s progress label now says
  "Suggest-generics mode" when only `--suggest-generics` is active
  (previously always "Type-aware mode").
- Post-fix full self-run: ZERO warnings, exit 0, 5 shown / 205 detected
  (was 207 — the two dropped groups are fixture-adjacent test-file clones
  whose shapes the splice restructured; scoped ./cmd count 24→22 for the
  same reason), `-t 5` self-analysis invariant still 0 shown.
- **jsonv2 recurrence #8 (restored same session)**: auto-commit `7348b4ba`
  (2026-10-11 23:15) re-migrated `internal/jsonutil/jsonutil.go`,
  `config/config_migrate.go`, AND `pkg/enum/enum.go` to direct v2 imports;
  caught by a fresh `-count=1` test run, restored verbatim from `d3c326b7`
  (the recurrence-#7 v1-only state).

## 2026-10-16 — October monthly self-scan (`-t 1 --type-aware`) + follow-up pass

Routine self-scan plus the triage follow-ups, one session. Start state:
44 shown. End state: **0 shown at -t 1, -t 3 --suggest-generics, and -t 5,
zero warnings on all three.**

- **Prose-hygiene correction (same day)**: the earlier "write
  `art-dupl:accept` without `//`" rule was INSUFFICIENT for comment lines —
  the scanner tolerates whitespace between `//` and the marker, so a marker
  leading a comment line (the grammar block in internal/accept/accept.go,
  printer/finding/finding.go:81, actionability_preamble.go:155) still
  matched. Fixed by backtick-wrapping the marker (a quote char breaks
  `//\s*`). Correct rule, now in AGENTS.md: in comments the marker must not
  directly follow `//` + whitespace.
- **Case/comm-label pattern extension**: `isTerminalStatement` now treats
  CaseClause/CommClause fragments with at most a trivial body (one simple
  statement; a comm clause allows its receive-statement key plus one) as
  terminal, so `single-simple-statement` suppresses the bare-label clone
  class (~8 of 44 groups were pure label clones: `default:`, `case <-ctx.Done():`).
  Fat case bodies stay actionable. Real-source regressions:
  `TestCaseLabelTerminalStatement`, `TestCommClauseLabelTerminalStatement`.
  CommClause keys are STATEMENTS (unlike CaseClause expression keys) — the
  initial ≤1 rule missed them; caught by the multidetector/combined.go
  `case <-ctx.Done():` group surviving the first cut. The extension ALSO
  tripped the documented file-kind collision trap on its first cut: templ
  `ComponentIfStatement` shares a raw value with a golang clause type, so
  the BDD "duplicate if statements in templates" spec went red (the group
  became non-actionable) — fixed by suffix-scoping the branch to non-.templ
  files (same rule as clone_classify), pinned by the templ-collision case
  in `TestCaseLabelTerminalStatement`.
- **Extraction**: `pkg/artdupl/detector.go` — the duplicated
  buildResult/reportProgress/ErrNoDuplicatesFound tail of `findClones` and
  `findClonesCombined` extracted to `finalizeResult` (this was the user's
  -t 3 run group #2).
- **Accept directives (32 + 1 templ)**: one bare directive per remaining
  group on its first occurrence (acceptance is group-level — one covered
  occurrence accepts the whole group). Classes: terminal-return fragments
  across distinct matchers (actionability helpers), channel-send/close
  cleanup idioms (job/parse, job/incremental), flag/accumulator assignment
  pairs, guard idioms (nil-guard + mutex prologues incl. the user's -t 3
  groups #1/#4), the 16-clone enum `return true` family (domain IsValid
  bodies; one directive on domain/detection_method.go), the SDK/CLI
  `.go`-filter mirror (pkg/artdupl/combined.go + detector_pipeline.go,
  deliberate ADR-0026 mirror), Ukkonen internals (suffixtree root checks),
  and transformer state save/restore prologues (transform.go FuncDecl/
  FuncLit symmetry — visible only under the type-erased pass, hence not in
  the -t 1 type-aware list; accepted at its -t 3 occurrence).
  The templ directive (report.templ wrapper loops delegating to the shared
  cloneOccurrence component) is an HTML comment — valid templ markup that
  still regex-matches.
- **Stale directive removed**: pkg/artdupl/types.go:17 (`architectural
  alias` above `type DetectionMethod`) matched nothing since its alias-decl
  family dissolved; the deterministic accounting surfaced it as a near-miss
  against the const group at :23 (which has its own directive at :22).
- **Judged-not-extracted**: pkg/position/lines.go defensive clamp pair
  (guards differ in shape; extraction would obscure), baseline_cmd
  constructor flag calls, bdd_helpers test guards.
- **Deferred pattern candidate (data point for the next mining pass)**:
  "≤2 terminal statements" would kill most of the 2-token residue
  (`return false` tails, `hasX = true` pairs) without per-site directives;
  needs the corpus precision ritual before adoption.
- **Follow-up fixes riding the same session**: restored the four docs
  gutted by auto-refresh commit `8a4ba024` (2026-07-26 — MIGRATION_GUIDE,
  DEDUPLICATION_FINAL_REPORT, EXECUTION_SUMMARY, IMPROVEMENTS_REPORT were
  0-byte husks; 1261 lines recovered from `8a4ba024^`); created
  `lychee.toml` excluding `docs/brainstorming/archived` (buildflow
  preflight remedy); added the `type-load` timing phase (job.PhaseTypeLoad)
  recorded around every go/packages load — the previously invisible ~15-30s
  inside `ingest` is now its own row; `--suggest-generics`-only runs now
  label the loading message "Suggest-generics mode" instead of
  "Type-aware mode".

## 2026-10-24 — off-cadence `-t 1 --suggest-generics --type-aware` combined sweep (crawl-pruning session)

**Detected: 1282 / shown: 2** (`art-dupl --sort total-tokens -t 1 --suggest-generics
--type-aware --html` on art-dupl itself; the user's run showed the same 2 groups
at 1283 detected). Both judged; zero harmful duplication found.

### Accepted (intentional)

| Clone | Rationale | Suppression |
| ----- | --------- | ----------- |
| `setAttributeKey` / `setConstantAttribute` two-statement prologue (`o.Type = Attribute; o.Name = key`, `syntax/templ/transform_node.go:98,150`) | Shared 2-line node-identity init ahead of deliberately divergent encodings (key-only vs key+value). An extraction takes as many parameters as the duplicated code has lines and hides the encode step it would sit next to. | Bare `//art-dupl:accept` on both doc comments (one covered occurrence accepts the group; two for edit-robustness). |
| `shortRevision` (cmd, truncate at 7) / `truncHash` (printer, truncate at 12) | Truncate-to-N guard idiom with different lengths and different domains (VCS short-hash convention vs group-hash display). No shared change pressure; a cross-package helper would out-param its 4-token body. | Bare `//art-dupl:accept` on both doc comments. No new denylist pattern: a truncation-idiom pattern from a 2-instance corpus is over-suppression risk with zero measured benefit. |

### Follow-up fixes riding the same session

- **Crawl dir pruning (the run's real anomaly)**: the timing report showed
  `crawl 31.837s` for 440 files. Warm reproduction was 132ms — the user's
  number was cold-cache inflation of a REAL defect: the CLI walk filtered
  excluded entries per-stat but never pruned, so `.git` (1.1k files),
  `website/node_modules` (16k files), `website/dist`, and every gitignored
  dir were fully traversed. Fixed three-layer: `shouldPruneDir`
  (dir-level form of `shouldSkipPath` via trailing-slash probe),
  `handleWalkEntry` returns `filepath.SkipDir` for excluded/ignored dirs, and
  a new `gitignore.GitignoreMatcher.IsIgnoredDir` (the matcher's dirOnly
  logic only matched paths UNDER a dir, and `filepath.Rel` strips trailing
  slashes, so directory-only exclusions could never prune the dir entry —
  the same latent gap silently disabled the provider crawl's gitignore
  pruning; fixed there too). File-selection semantics unchanged for
  structural skips; gitignore pruning now matches git + provider semantics
  (a negation cannot re-include beneath an excluded dir — pinned by
  `TestCrawlPrunesGitignoredDirs`). Warm crawl 132ms → 74ms; cold crawls
  lose the ~20x ignored-entry stat multiplier. Detected-group counts
  byte-identical before/after.
- **Not defects, verified**: `type-load` (17-36s) is the documented single
  shared go/packages load of combined mode (ADR-0026); HTML-to-stdout without
  a notice is the piped-stdout branch of `resolveHTMLOutput` (TTY branch
  auto-writes + notices, covered by `cmd/html_pty_test.go`).
