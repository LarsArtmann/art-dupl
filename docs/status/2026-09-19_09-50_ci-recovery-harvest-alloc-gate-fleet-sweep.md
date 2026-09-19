# Status Report: CI Recovery to Green, Harvest Execution, Alloc-Gate Root Cause, and gogenfilter Fleet Sweep

**Date:** 2026-09-19 09:50 CEST
**Session window:** 2026-09-19 ~08:03 → ~09:50 CEST (resumed after the 06:41 report)
**Branch:** `fork` (art-dupl), `master` (gogenfilter + 14 bumped consumers)
**Head at report time:** art-dupl `fa806995` (last authored push: `20938fe6`), gogenfilter `325502f`

**Scope note:** per instruction, this report covers ONLY this session's work — the continuation
session that executed the harvest of `2026-09-19_06-41_v0.7.0-release-go1.27-coherence-ci-recovery.md`
and the work that fell out of it. A concurrent session (go-finding GAP-2 adapter, issue #1) was
live in this repo throughout; its work is NOT reported here except where it collided with mine.

---

## Executive Summary

The session started from the 06:41 handoff with three stated next steps (pkg.go.dev verify,
HARVEST, alloc variance) and immediately found the repo in a state the handoff did not know
about: **CI was red again** — the windows Test lane had failed at 21:25Z and 06:04Z/06:10Z UTC —
and **a concurrent session was actively committing** (go-finding adapter, 9 daemon commits
07:34–08:40). The session therefore split into: (1) bring CI back to genuinely green, (2) execute
the docs harvest, (3) root-cause the alloc-gate variance, (4) sweep the gogenfilter fleet.

**Net state:** all 5 CI workflows green on every push I authored (3 consecutive windows-green
runs after the fix); harvest complete with 8 items executed-then-dropped; the alloc-gate
"cross-machine variance" theory **disproven and replaced** with a real root cause; 14 fleet
repos moved to gogenfilter v3.6.1 with 9 pre-existing-red repos documented as debt. One new
discovery class: the fleet-wide red-baseline debt (9 repos) is bigger news than the sweep itself.

---

## a) FULLY DONE

Each item is committed and verified (test, gate, or CI evidence cited).

### CI recovery: two distinct windows failure classes, both closed

1. **Recurring stdin-cancel flake root-caused and fixed.** `TestStdinFeed_CancelUnblocksScanner`
   failed at 23:03Z and 06:04Z. Root cause: the test depends on `os.File.Close` unblocking a
   pending pipe `Read`, which Windows `os.Pipe` does not guarantee — whether `Scan()` has
   already entered `Read` when `Close()` fires is a scheduling race, so the outcome is flaky,
   not deterministic (it passed at 23:17Z on luck). Fixed by applying the same documented skip
   the `StderrSuppressedOnCancel` sibling already carries, to BOTH `CancelUnblocksScanner` and
   `TimeoutUnblocksScanner` (same dependency), with the why written into the test comments.
   Verified locally with `CGO_ENABLED=1 go test -race`. Evidence: commit in `feda5fe0`.
2. **Fatal GC-crash class scoped out of windows.** The 21:25Z failure crashed with
   `fatal error: found pointer to free object` + `missing stackmap` across bdd/cmd/pkg-artdupl/
   syntax-golang, and the 06:10Z re-run crashed the same class in bdd/cmd/job/actionability —
   while the 21:17Z run PASSED on binary-identical inputs (the only delta was a CHANGELOG.md
   edit, and no test reads the CHANGELOG). Evidence chain: nondeterministic, windows/amd64-only,
   **zero `unsafe.` in production code** (grep-verified), no 1.27.2 exists to upgrade to
   (go.dev/dl checked). Mitigation: matrix `include` sets `goexperiment: "false"` for
   windows-latest so the Test lane runs without `GOEXPERIMENT=jsonv2`; linux/macOS keep it.
   Safety of the mitigation verified locally: full build + broad test suite green WITHOUT the
   experiment (ADR-0024 already made output bytes experiment-independent by construction).
   Comment in `.github/workflows/ci.yml` carries the full evidence chain.
   **Result: CI green on `ea05eb1e`, `9160e7fc`, `20938fe6` — three consecutive windows-green runs.**
3. **AGENTS.md stale invariant corrected.** "`GOEXPERIMENT=jsonv2` is required" → "optional
   since ADR-0024", with the windows-lane exception and its rationale recorded where future
   sessions read first (AGENTS.md build section).

### Handoff next-steps closed

4. **pkg.go.dev verification (handoff item a.1, report #1):** `@v0.7.0` renders fully — README,
   Valid-go.mod/Stable/Tagged badges, complete directory tree including `internal/jsonutil`.
   The "not in the latest version of its module" banner is a pkg.go.dev UI stale fragment
   (it appears on the canonical `/` page too, which also shows v0.7.0 + Latest badge). Proxy
   cross-check: `go list -m -versions` → v0.7.0 is the newest. Item CLOSED.
5. **HARVEST of section (f) (report #3):** all 50 items verified against code/state, then routed:
   **16 → TODO_LIST** (MEDIUM, with `(#N)` citations back to the report), **7 → ROADMAP** (new
   "Platform and Ecosystem" section), **8 executed-then-dropped** (ADR-0024, FEATURES, AGENTS,
   HOW_TO_USE, jsonutil tests, gogenfilter CHANGELOG, performance.yml pins, the `-t 1` recount),
   **3 dropped as already-done/no-op** (#17 `.gitattributes` already had `* text=auto eol=lf`
   which strictly covers the ask; #47's triggers already live in the PARKED table; #49 verified
   a no-op), 1 re-verified (#34 Deploy Site content current), 1 done cross-repo (#45 push,
   #21 CHANGELOG). No 50-item dump — the anti-pattern the harvest guide warns about.
6. **ADR-0024 written (#6):** `docs/adr/0024-json-v1-api-behind-canonical-marshaler.md` —
   supersedes ADR-0010's API-surface half, keeps the engine half; records the two pinned wire
   behaviors, the `omitempty`-stays-deleted decision, the jsonutil-only rule, and the Go 1.28
   watch item. The arch-lint claim inside it was grep-verified before shipping (jsonutil
   component, `.go-arch-lint.yml:64`).
7. **FEATURES.md brought to v0.7.0 (#7):** header version/date; CacheVersion 3→4 (stale, fixed
   on sight); pattern count 30→37 (**verified against the binary**: `art-dupl --list-patterns |
   wc -l` = 37 = 33 denylist + 4 property-engine); JSON anchor_id row; include/exclude
   zero-match-warning row; dump-tokens position column row.
8. **AGENTS.md Critical Conventions additions (#8, minus the already-present anchor_id):**
   include-pattern zero-match tracking (unconditional-warning semantics), dump-tokens source
   positions (with the per-subtest-table race note), the crawl slash-normalization Windows
   invariant, and the jsonutil canonical-marshaler rule.
9. **HOW_TO_USE.md additions (#9 + the #25 gap):** anchor_id cross-format contract (and its
   deliberate absence from --simple-json/--plumbing), `--include-tests` flag doc (behavior was
   verified matching docs; the flag itself was undocumented), and the suggest-generics ×
   --exact/--structural validation error — wording pulled verbatim from
   `cmd/config_builder.go:285-301`, not paraphrased.
10. **jsonutil direct unit tests (#26):** 5 tests pinning the contract (no-HTML-escaping incl.
    `\u003c`/`\u003e`/`\u0026` negative checks, no-trailing-newline for both entry points,
    indentation semantics, `Marshal ≡ MarshalIndent(v,"","")`, error wrapping). All green;
    `go vet` clean.
11. **performance.yml toolchain pins (#41):** both jobs `GO_VERSION: stable` → `"1.27.1"`
    (explicit pin per the 09-14 arch-lint lesson); release.yml already correct via
    `go-version-file: go.mod`. Verified in CI: Performance Tests green on every subsequent push.
12. **gogenfilter CHANGELOG v3.6.1 entry (#21):** written in the repo's Keep-a-Changelog format
    with the real fix description and test name (pulled from the v3.6.1 diff, not memory);
    committed and pushed (`325502f`).
13. **`-t 1` self-scan true count (#13):** measured `40` groups (grep-counted `^found .* clones:`,
    identical with and without `--explain`), replacing the report's file-count-heuristic "44".
    Recorded in TODO_LIST's ledger item so the future ledger starts from the real number.
14. **#34 Deploy Site verified:** art-dupl.lars.software renders the current feature set (3 modes,
    7 formats, generated-code filtering, baseline gating, SDK) with no stale version claims.
15. **#45 go-paperless CI debt closed:** the findByName consolidation commits from 09-18
    (`7d7d9e7`..`04c32dc`) had NEVER been pushed (origin at 09-17); pushed, CI now covers them.
    Their CI shows green on main.
16. **Alloc-gate variance ROOT-CAUSED (#11) — the report's premise was wrong.** Measured
    `seq/tokens_10000` three times on the SAME machine, SAME go1.27.1: **30745, 30742, 30743** —
    a ±3 run-to-run band on identical inputs. GOMAXPROCS=1 does not stabilize it (retested).
    So the 2026-09-18 "dev box 30742 vs CI 30744 cross-machine variance" note was a
    misattribution: it is goroutine/runtime-internal allocation noise. Fixed the gate properly:
    `scripts/check-alloc-regression.sh` now runs each suite with `-count=3` and compares the
    **minimum** (converges to the deterministic floor — three draws of exactly 30742 in
    verification), budgets re-anchored at the floor (30744→30742), header + budget file document
    the corrected attribution, and the script now sets `GOTOOLCHAIN=auto` so it runs from any
    shell. Verified: gate EXIT=0 (25 ok, 3 "improved" ratchet infos), and green in CI.
17. **gogenfilter v3.6.1 fleet sweep (#2) executed** — see (a) items 18–20 and the tally below.
18. **Fleet enumeration:** 23 consumer repos found under ~/projects (8 direct require, 15
    indirect; art-dupl and file-and-image-renamer already on v3.6.1).
19. **14 repos bumped + on their remotes:** go-auto-upgrade, go-humanize-linter,
    go-structure-linter, golangci-lint-auto-configure, dynamic-markdown-site, index,
    project-dependency-graph, project-discovery-sdk, projects-management-automation,
    prompt-crusher-exec, prompt-crusher, template-AUTHORS, template-SECURITY (all baseline→bump→
    `go test ./...` green→committed→pushed), plus **oxlint-auto-configure** which needed an
    extra `go mod vendor` — the sweep's post-bump build caught the stale `vendor/modules.txt`
    (failure-mode F13 exactly); vendor sync committed and pushed. go-humanize-linter,
    go-structure-linter, BuildFlow-side `go.work` files handled per the skill's F2 rename rule
    where present.
20. **9 pre-existing-red repos correctly SKIPPED and recorded** (F11 discipline — no bump on a
    red baseline): branching-flow (dep build fail: go-design-smells), erraudit (3
    TestRunner_OopsFix failures), go-filewatcher (`TestFilterGeneratedCode_SingleFilters/SQLC`
    — possibly a stale gogenfilter-behavior assumption, flagged as the first one to revisit),
    hierarchical-errors, BuildFlow (baseline build fails), Cyberdom, auto-deduplicate (build,
    still on v3.3.2), overview, project-discovery-daemon. All listed in TODO_LIST as fleet debt.

---

## b) PARTIALLY DONE

1. **Windows support (report #4/#5).** Two of the three failure classes are closed (stdin flakes
   skipped with rationale; GC-crash lane mitigated), but `TestExitCodes_Process` remains
   Windows-skipped (exe-start `ProcessState nil` — needs an actual Windows runner to
   root-cause; nothing more can be done from this machine), and the GC-crash fix is a
   **mitigation, not a root fix** — if crashes recur without the experiment, the next step is an
   upstream Go issue, which needs a Windows repro capability I don't have.
2. **Fleet sweep completeness.** 14/23 moved; the 9 skipped repos still carry ≤v3.6.0 (and two
   carry v3.3.x). The sweep itself is done; the debt is the red baselines, which are per-repo
   projects, not sweep work.
3. **CHANGELOG entries for this session's changes.** The alloc-gate fix, windows CI mitigation,
   and fleet sweep are committed but not yet in CHANGELOG (Unreleased). Deferred deliberately:
   the concurrent session was editing CHANGELOG (go-finding entry) and interleaving would have
   created churn; they should ship in the next release cut's notes.
4. **Global AGENTS.md lesson (report #42).** The stale-lock + foreign-generator-binary lesson
   could not be written — `~/.config/crush/AGENTS.md` sits on a **read-only filesystem** in this
   session (write attempt failed with EROFS). The lesson text exists in this report instead.
5. **CI confirmation on two bumped repos.** go-humanize-linter CI green on its bump commit;
   go-auto-upgrade only showed the Dependency Graph workflow — its CI either doesn't run on
   push or hadn't triggered; not verified further.
6. **The harvest itself.** Routing is complete, but the 16 TODO_LIST items are by definition not
   executed; only the 8 S-effort ones were done inline this session.

---

## c) NOT STARTED

Carried in TODO_LIST/ROADMAP with citations (report #N refs); listed here for completeness.

1. **HIGH-priority benchmark evidence (T2.2, T23, FindTranBoundary benchstat)** — still parked;
   load was 16–29 all session, entry criterion (<4 sustained) never met. Correctly parked.
2. **Fleet audit: `encoding/json/v2` imports / `format:` tags on Go 1.27** (#14) — the other
   half of this session's breakage class, unstarted.
3. **Fleet audit: `filepath.Separator` + `strings.Split(_, ":")` path parsing** (#15).
4. **`scripts/pre-release-check.sh`** (#16) — tag-collision + CI-green + toolchain-pin gate.
5. **Branch protection with required checks + failure notifications** (#18) — owner action.
6. **Auto-tag workflow investigation** (#19).
7. **go-paperless tag + release** (#20) — needs a proper go-release session (45 commits since
   v0.3.1; version decision is not a drive-by).
8. **Corpus re-baseline post-v0.7.0** (#22).
9. **Self-clean decision ledger** (#12) — true baseline now known: 40 groups.
10. **Stale-shell graceful error** (#24) — hit twice in practice this session (LSP noise, alloc
    gate first run); the script-level fix landed, the product-level UX did not.
11. **`--dump-tokens` positions on a templ file** (#48).
12. **`nix flake check` arch-lint coverage** (#40).
13. **docs-health VERIFY over 2026-08-* reports** (#35).
14. **Small quality batch** (#29 shouldSkipPath separator tables, #30 exhaustruct noise, #31
    monthly -t 1 routine, #32 errors.AsType, #38 t.Chdir, #39 tagalign/nestif confirm).
15. **Windows real cancellation / windows-required lane** (ROADMAP) — gated on the platform
    question from the 06:41 report, still unanswered.

---

## d) TOTALLY FUCKED UP

Radical honesty — what I broke, got wrong, or mishandled this session.

1. **I typed a banned command into the fleet-sweep script.** The failure-recovery path used
   `git checkout -- go.mod go.sum`; the global AGENTS.md bans `git checkout` outright in favor
   of `git restore`. I caught it by re-reading the script before launch and fixed it pre-run —
   but the fact it was in the first draft at all means the rule isn't reflexive yet. A
   mechanical rules-checklist pass over any git-mutating script should be ritual, not luck.
2. **My first jsonutil test could never pass.** The "combined" HTML case contained raw `"`
   characters, so the substring assertion failed (JSON escapes quotes as `\"`). The test run
   caught it immediately and the case was trivially fixed — but it means I wrote a wire-format
   assertion without first simulating the exact output bytes. For a test whose entire purpose
   is pinning byte-level behavior, that's sloppy.
3. **I harvested an item and invalidated it within the same session.** The alloc-variance item
   (#11) went into TODO_LIST as open at ~08:45 and was root-caused by ~09:05, forcing an
   immediate TODO_LIST correction. Harvest-then-execute in one session creates churn; the
   S-effort items should have been executed BEFORE being written down as open work.
4. **The alloc-gate script failed on its first run in a stale shell** (`go.mod requires go >=
   1.27.1; GOTOOLCHAIN=local`) — the exact failure class I had harvested from the report an
   hour earlier (#24), with the environment caveat sitting in my own handoff notes. I fixed the
   script on the second pass, but the first-pass miss was avoidable.
5. **`go test -race` failed once for missing cgo** before I added `CGO_ENABLED=1` — the caveat
   was explicitly in the handoff summary. Read the environment cautions, then use them on
   command one; not on command two.
6. **The sweep's dynamic-markdown-site push failure cost three diagnostic cycles** because my
   script logged only "push failed/uptodate". Reality: a pre-push hook ran `go test -race`
   (needs CGO), lost a race with the daemon's own push, and the remote already had my commit.
   A push-failure log that captured hook-vs-remote-vs-tracking would have resolved it in one
   look. (The `! [remote rejected] ... is at a3f1a26 but expected 5fad036` line — where the
   "expected" hash was BEHIND the "is at" hash — was the tell that the remote was already
   updated.)
7. **I never closed the templ-binary forensics.** The `report_templ.go` drift was first
   attributed to "a newer templ binary", then the withGo127 wrapper turned out to be
   version-identical, and the drift style actually matches goimports-style reformatting from
   the other actor's tooling. I resolved it operationally — HEAD is canonical because CI/nix
   build with the flake's pinned `pkgs.templ`, and my regeneration reproduced HEAD byte-for-
   byte — but I never identified WHICH binary produced the drift. Correct outcome, incomplete
   root cause; two forensic cycles was enough and I stopped, which was right, but I should
   have said so out loud at the time instead of leaving the hypothesis wrong in my notes.
8. **The sweep runner script lived in /tmp**, against the go-ecosystem-upgrade skill's explicit
   "persist runner scripts and results in the repo" rule. Justified it to myself as a one-off;
   the rule exists precisely so re-runs are possible, and this report's summary is now the only
   durable record of the script's logic.
9. **I watched the concurrent session bump go.mod (fa806995, go-finding dependency) and moved
   on without a sanity check** that the new dependency version resolves and builds. Not my lane
   and their session verified their own work — but a one-line `go build` on the merged HEAD
   would have cost nothing and covered the seam between two sessions' work.

---

## e) WHAT WE SHOULD IMPROVE

1. **Rules-checklist pass on git-mutating scripts, before first run** (banned commands, daemon
   race, path scoping). Catching the `git checkout` pre-launch was luck layered on habit; make
   it procedure.
2. **Write the expected bytes next to wire-format assertions.** For golden/contract tests,
   draft the literal expected output first, then the assertion — prevents untestable-by-
   construction cases like the jsonutil quote bug.
3. **Execute S-items before harvesting them.** Same-session harvest→execute→de-harvest is
   churn; the harvest pass should carry a "verify-then-execute-if-S" step so only genuine
   backlog lands in TODO_LIST.
4. **Environment cautions are first-command material.** `CGO_ENABLED=1` for -race and
   `GOTOOLCHAIN=auto` for every go invocation should be applied from the first test run, not
   discovered per-failure. (Session-level env setup belongs at resume, not mid-stream.)
5. **Sweep scripts should classify push failures** (pre-push hook / remote rejection /
   stale tracking ref / network) — the dynamic-markdown-site ambiguity cost three cycles.
6. **Budget captures record the band, not a draw.** Now codified in the gate (min-of-3, floor
   budgets); the broader habit: any "deterministic" metric that can drift by ±3 needs its noise
   band measured at capture time.
7. **Cross-session collisions want a claim mechanism.** Both sessions edited AGENTS.md,
   CHANGELOG, FEATURES this session; daemons serialized the commits, but semantics could have
   collided. Even a `.session-claim` file with timestamp+scope would help.
8. **Time-box forensics, state the residue.** The templ hunt got two cycles and stopped with an
   operational answer plus an open question — right call, but the open question should have
   been written down at stop time, not reconstructed later.
9. **The 06:41 report's section (f) should have carried "verify-then-execute-if-S"** — feed this
   back into the status-report skill's next-tasks template so future harvests front-load
   execution.

---

## f) TOP THINGS TO GET DONE NEXT

Ranked by impact. Effort: S (<30m) / M (30m–2h) / L (>2h). Items already in TODO_LIST keep
their `(#N)` citations; new discoveries from this session are marked **[new]**.

| #  | Task                                                                                                                                                                                                          | Impact | Effort | Source    |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | --------- |
| 1  | Fix the 9 red-baseline fleet repos so the gogenfilter bump can land fleet-wide — start with go-filewatcher's `TestFilterGeneratedCode_SingleFilters/SQLC` (may be a real behavior assumption, not rot)          | High   | M–L    | [new]     |
| 2  | Watch the windows Test lane for 1–2 weeks; if the GC-crash class recurs WITHOUT the experiment, prepare the upstream Go issue (repro needs Windows access — see question 2)                                    | High   | S–M    | [new]     |
| 3  | Fleet audit: `encoding/json/v2` / `format:` tags on Go 1.27 repos (#14) — same breakage class as the v0.7.0 blocker                                                                                            | High   | M      | TODO_LIST |
| 4  | `scripts/pre-release-check.sh`: ls-remote tag-collision + CI-green + toolchain-pin gate (#16)                                                                                                                  | High   | M      | TODO_LIST |
| 5  | Branch protection with required checks + failure notifications (#18) — owner action; red CI sat 4 days last week and ~10h again this week                                                                      | High   | S      | TODO_LIST |
| 6  | go-paperless: tag + release the findByName consolidation via go-release (#20) — pushed but unreleased; also confirm its CI green on `04c32dc`                                                                   | Medium | M      | TODO_LIST |
| 7  | erraudit TestRunner_OopsFix failures (3 pre-existing) — my own adjacent tooling, likely quick                                                                                                                   | Medium | S–M    | [new]     |
| 8  | Corpus re-baseline post-v0.7.0 + AGENTS.md number refresh (#22)                                                                                                                                                | Medium | M      | TODO_LIST |
| 9  | Self-clean decision ledger from the true 40-group baseline (#12)                                                                                                                                               | Medium | S–M    | TODO_LIST |
| 10 | CHANGELOG entries for this session's infra work (alloc gate min-of-3, windows experiment scoping, fleet sweep) at the next release cut                                                                          | Medium | S      | [new]     |
| 11 | Root-cause Windows exe-start `ProcessState nil`; un-skip TestExitCodes_Process (#4) — requires a Windows runner session                                                                                         | Medium | M      | TODO_LIST |
| 12 | Stale-shell graceful error in the product (#24) — hit twice in practice this session                                                                                                                           | Medium | S      | TODO_LIST |
| 13 | Investigate dynamic-markdown-site's pre-push hook: it runs `-race` tests without guaranteeing CGO — hook bug or env assumption?                                                                                 | Medium | S      | [new]     |
| 14 | templ `--dump-tokens` position verification (#48)                                                                                                                                                              | Medium | S      | TODO_LIST |
| 15 | `nix flake check` arch-lint coverage (#40)                                                                                                                                                                     | Medium | S      | TODO_LIST |
| 16 | Fleet audit: `filepath.Separator` + `Split(_,":")` path parsing (#15)                                                                                                                                          | Medium | M      | TODO_LIST |
| 17 | Auto-tag workflow vs manual release tags (#19)                                                                                                                                                                 | Medium | S      | TODO_LIST |
| 18 | docs-health VERIFY over 2026-08-* reports (#35)                                                                                                                                                                | Medium | M      | TODO_LIST |
| 19 | Persist a generalized fleet-sweep runner (parameterized lib+version) in a tools repo — this session's script logic survives only in this report                                                                | Low    | S      | [new]     |
| 20 | Verify go-auto-upgrade CI actually ran on the bump commit (only Dependency Graph observed)                                                                                                                     | Low    | S      | [new]     |
| 21 | BuildFlow baseline build failure (go.mod/go.work interplay?) — unblocks a bumped consumer later                                                                                                                | Medium | M      | [new]     |
| 22 | auto-deduplicate baseline build failure + v3.3.2 pin — oldest consumer, likely needs two hops                                                                                                                  | Low    | M      | [new]     |
| 23 | Small quality batch (#29 separator tables, #30 exhaustruct noise, #31 monthly -t 1 routine, #32 errors.AsType, #38 t.Chdir, #39 tagalign confirm)                                                               | Low    | S each | TODO_LIST |
| 24 | ROADMAP capture: report section-(f) template should embed "verify-then-execute-if-S" (status-report skill feedback, e9/#36)                                                                                     | Low    | S      | [new]     |
| 25 | Go 1.28 watch: `GOEXPERIMENT=jsonv2` retirement behavior (ADR-0024 consequence; re-verify `nix build` on first 1.28 beta) (#50)                                                                                 | Medium | S      | ROADMAP   |
| 26 | Keep the PARKED benchmark items parked (T2.2/T23) — entry criterion (load <4) was never met this session either; revisit only when the machine is idle                                                          | —      | —      | TODO_LIST |

(Stopping at 26: items 27–50 of the 06-41 report remain routed in TODO_LIST/ROADMAP and are not
duplicated here; the ranking above is the delta this session added on top of them.)

---

## g) TOP 3 QUESTIONS I CANNOT ANSWER MYSELF

1. **The windows fatal GC crashes: mitigation accepted, or chase the root cause upstream?** The
   `GOEXPERIMENT` scoping has held for 3 consecutive windows runs, but I cannot *prove* the
   experiment was the trigger — binary-identical runs crashed with it ON, and I have no Windows
   environment to reproduce against. If you have a Windows machine/VM I can drive, I will build
   a minimal repro and file the Go issue; otherwise, is skip-and-mitigate the accepted resting
   state for a windows/amd64 Go 1.27.1 toolchain bug?

2. **Is the 9-repo red-baseline fleet debt worth budgeting now?** The sweep exposed that 9 of 23
   gogenfilter consumers have pre-existing red builds/tests (some stale for weeks, e.g.
   auto-deduplicate on v3.3.2 with a failing build). That is a bigger health signal than the
   bump itself. Do you want a dedicated fleet-hygiene session (fix baselines repo by repo,
   starting with go-filewatcher's suspicious sqlc test), or do those repos only get touched
   when you next work on them?

3. **How should concurrent sessions on this repo coordinate — and who owns the next release?**
   This session overlapped live with the go-finding GAP-2 session; the daemons serialized our
   commits, but we both edited AGENTS.md/CHANGELOG/FEATURES and one of my daemon-commit reads
   included their in-progress files. Do you want a claim/lock convention for multi-agent work
   on art-dupl — and when the next release cut happens (bundling GAP-2 + the alloc-gate/CI/
   sweep work), should it be one coordinated release session rather than whichever session
   gets there first?

---

## Handoff

CI is genuinely green (3 consecutive windows-green runs), the 06-41 harvest is fully routed with
8 items executed, the alloc gate is now evidence-based instead of folklore-based, and the
gogenfilter fleet is as bumped as it can be without fixing 9 unrelated red baselines. The single
most valuable next session is probably **fleet hygiene (f1/f7/f21)** — the red-baseline debt is
now measured, named, and waiting. Waiting for instructions.
