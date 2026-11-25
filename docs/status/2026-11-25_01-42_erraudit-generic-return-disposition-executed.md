# Status Report: erraudit `generic_return` Disposition — Executed

**Date:** 2026-11-25 01:42
**Session scope:** Execution of the 12-task plan from `docs/planning/2026-11-24_20_26-erraudit-generic-return-disposition-plan.html` (planning session 2026-11-24 20:26). This report covers THIS session's run only.
**Branch:** `fork` @ `e327504d` (daemon auto-commit) + uncommitted `AGENTS.md`/`TODO_LIST.md`

---

## a) FULLY DONE

1. **Session-start gate**: `go test -count=1 ./internal/jsonv2gate/` green; `pgrep -af buildflow` clean before any edit (standing recurrence lesson honored).
2. **Strict-posture verification**: `erraudit lint ./... --max --disable generic_return` → **exit 0**, verified live pre- and post-fix. This is the adopted audit command.
3. **Genuine fix #1** (`internal/testutil/capture.go:46-57`): both `os.Pipe()` failure returns now coded via `duplerrors.WrapIO(pipeErr, "stdout"/"stderr", "create capture pipe")`; import added. Arch-lint permits (`testutil: anyProjectDeps`).
4. **Genuine fix #2** (`printer/text.go`): `writeFileDupeHeader` + `writeCloneHeader` `fmt.Errorf`+`%w` sites → `duplerrors.Wrapf(err, InternalError, ...)`, matching sibling style in `PrintClones`.
5. **Scoped verify**: `go build ./...` + `go vet` + `-count=1` tests on `internal/testutil/` and `printer/` — green.
6. **TODO_LIST.md**: erraudit row struck DONE with full disposition evidence; new parked row for the two owner decisions (erraudit `extension` recognizer; BuildFlow-side wiring trigger).
7. **AGENTS.md**: new `erraudit posture` bullet landed directly after the Errors convention bullet.
8. **DOMAIN_LANGUAGE.md**: correctly skipped — erraudit is fleet tooling, not clone-detection domain language. No new terms emerged.
9. **Boundary gates**: fast profile 38s green; `--full` profile green (complete race suite).
10. **Self-scan**: `-t 5` over all first-party packages = 61 detected / 0 shown, exit 0 — matches the documented clone-free invariant exactly.
11. **Plan artifact ledger**: dated `EXECUTED 2026-11-24` completion callout added above the footer (snapshot policy: inline ledger, not rewrite); footer's "pending task 2" corrected.
12. **Parked tasks 11/12**: recorded in TODO_LIST (not silently dropped).

## b) PARTIALLY DONE

Nothing. All 10 active tasks completed; 2 parked-by-design items recorded for the owner.

## c) NOT STARTED

1. **Parked owner decision (a)**: teaching `generic_return` the duplerrors constructor set via erraudit's `extension` subsystem — cross-repo work in the erraudit repo, deliberately not started (owner-gated).
2. **Parked owner decision (b)**: BuildFlow-side erraudit wiring — deliberately deferred; this repo's `.buildflow.yml` doesn't run erraudit, so there is nothing to configure today.

## d) TOTALLY FUCKED UP

Nothing catastrophic. One honest execution surprise worth recording:

- **The fix-values subclass ROSE 15→17 after the fixes.** Pre-execution, the plan expected fixing the 2 genuine sites would shrink the flagged subclass. Reality: erraudit recognized the old `%w` wraps (in `writeFileDupeHeader`/`writeCloneHeader`) as "coded", and cannot see `duplerrors.Wrapf` — so the two newly-coded helpers flipped INTO the "anonymous" bucket. `CaptureStdoutStderr` also remains flagged (its `errors.Join` is a stdlib constructor). This is not a regression in the repo's own terms (both edits are convention-correct improvements), but the plan's prediction was wrong, and the surprise is the single strongest piece of evidence for the disable posture. Recorded in TODO_LIST + AGENTS.md + the plan ledger. Lesson: **when a checker is blind to your conventions, "fixing values" per its advice can make its count worse — predict checker-visible vs checker-invisible constructs separately.**

## e) WHAT WE SHOULD IMPROVE

1. **Plans should classify error-wrapping styles by checker visibility.** The triage read 15 function bodies but never asked "which of these constructs can erraudit SEE?" — that one question would have predicted the 15→17 flip.
2. **`errors.Join` is an uncoded seam in `capture.go`** (fn error + raw `io.Copy` errors from `CopyToBuffer` goroutines). Out of scope by plan, but it's the remaining uncoded path in the file we touched.
3. **The 267→79 finding-count reconciliation is inferred, not proven** — TODO_LIST now says "different sweep (count is flag/version-dependent)" but I never bisected which flag/version produced 267. Cheap to verify, not done.
4. **HTML plan artifacts could carry a "prediction vs actual" section** — the flip would read as a lesson instead of a correction to the footer.
5. **Doc updates (AGENTS.md, TODO_LIST.md) sat uncommitted at close-out** — fine under the auto-commit daemon contract, but worth knowing the daemon's heuristic sometimes takes a sweep or two.

## f) Up to 50 things we should get done next

**erraudit / error-handling follow-ups**

1. Owner decision: sponsor the erraudit `extension` recognizer for duplerrors constructors (parked row, TODO_LIST).
2. Owner decision: BuildFlow-side erraudit wiring policy if the fleet adopts it as a gate step (parked row).
3. Code the `errors.Join` seam in `capture.go` (wrap `CopyToBuffer` goroutine errors as IOError before joining).
4. Verify which erraudit version/flags produced the historical "267 findings" (bisect against TODO_LIST history).
5. Consider a tiny unit test asserting `CaptureStdoutStderr` pipe-failure errors are `IOError`-typed (wrap-path coverage; currently only the happy path is exercised).
6. Sweep the repo for any other bare `fmt.Errorf` in production paths (same class as the two text.go sites) — erraudit's advisory list is a ready-made worklist.

**Standing repo hygiene (from TODO_LIST, untouched this session)**
7. Upgrade fang v1 → v2 (`charm.land/fang/v2`).
8. Add statix + govet lanes for `flake.nix`.
9. `cmd/run_suppression.go` carve-out (suppression/output assembly out of the analysis command file).
10. Sweep machines/repos for multi-copy art-dupl installs (PATH-shadow silent staleness).
11. Install the corrected art-dupl binary into the nix system profile (owner-gated; PATH still resolves poisoned `0.9.0-8ce90a4`).
12. Decide buildflow state-DB retention (1.1 GB live data, VACUUM reclaims nothing).
13. Document-or-decide the test flake check sandbox-exclusion list.
14. docs-health check-rows leveling pass over the 160 wave-archived files.
15. Harden the docs-health batch tools (anchor-uniqueness abort, `--dry-run`, self-test fixture).
16. docs-gates upstream + regime-formula decisions (owner-gated).
17. docs-health annotator upstream fixes (per-cell row strikes, ambiguous-key warning).
18. bandit lane: exclude INFO chatter.
19. sqlite resume-DB policy (single-instance discipline or busy_timeout upstream).
20. file-size-check policy decision (82 findings; worst `config/config_enum_test.go` 1,269 lines).
21. Archive the ~29 undated legacy reports + decide `2026-10-05_pass-verdicts/` fate.
22. empty-snapshot guard C in `scripts/check-docs-freshness.sh`.
23. Duplication-trends dashboard proposal owner decisions (Q1-Q4) + cross-repo cite verification.
24. Next monthly docs-health pass due **2026-12-14**.
25. Next monthly self-scan (`scripts/self-scan.sh -t 1 --type-aware`, ledger judgments) — due by mid-December.

## g) Questions I cannot answer myself

1. **erraudit `extension` sponsorship**: do you want an issue filed in the erraudit repo proposing a duplerrors-family (or generic custom-constructor) recognizer for `generic_return`, or is the `--disable generic_return` posture the end state for this fleet?
2. **`capture.go` scope**: should the `errors.Join` seam (fn error + raw `io.Copy` goroutine errors) also be coded as IOError, or is test-helper error coding below the bar you care about?
3. **AGENTS.md posture wording**: I recorded `--disable generic_return` as THE strict-audit command for this repo. If the fleet later standardizes a different invocation (e.g., BuildFlow step with its own config), should this repo's AGENTS.md defer to the fleet layer instead of owning the command?

---

**Bottom line:** all 12 planned tasks executed and verified green (2 code fixes, 1 posture adopted+verified, 3 doc records, 4 gate runs, 2 parked items recorded). The one prediction miss (15→17) was converted into recorded evidence strengthening the disposition.
