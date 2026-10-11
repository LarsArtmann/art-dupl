# Status Report — SARIF / GitHub Visibility Session

**Date:** 2026-10-11 11:27 CEST
**Repo:** art-dupl (branch `fork`, working tree clean, HEAD `b0048a6b`)
**Session type:** Q&A / investigation only — **zero code changes, zero commits by this session**

## Session Scope

Three questions were asked and answered, all in the territory "make art-dupl's SARIF
output visible on GitHub":

1. How do SARIF outputs become visible on GitHub?
2. Does `github.com/larsartmann/go-finding/toolsdk` support that?
3. Confirm: Findings → `Report.ToSARIF` and toolsdk → Findings?

Everything below is scoped to this session's run and what it noticed along the way.
No unrelated research was done.

---

## a) FULLY DONE

Verifiable, complete during this session (evidence cited):

| #  | Done                                                                                                                                                                                                                                          | Evidence                                                                                                                                             |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| A1 | Answered "SARIF → GitHub" with a concrete, repo-fitting recipe: `--sarif` to stdout → file → `github/codeql-action/upload-sarif@v3` with `permissions.security-events: write` + `category: art-dupl`; alerts land in Security → Code scanning | Chat recipe; verified SARIF 2.1.0 output at `printer/sarif.go:26`, stdout-only emission, level ladder `error/warning/note` at `printer/sarif.go:336` |
| A2 | Verified art-dupl exits **0 even when clones are found** (`ExitCodeForError`), so no `continue-on-error` is needed in an upload workflow                                                                                                      | `cmd/exit_codes.go:22-39`                                                                                                                            |
| A3 | Verified toolsdk v1.15.0 has **no SARIF/upload/GitHub code at all** — it is purely the BuildFlow self-registration contract                                                                                                                   | Module cache: `toolsdk@v1.15.0/doc.go`, `spec.go:71`; grep for sarif/upload across toolsdk = 0 hits                                                  |
| A4 | Verified go-finding core v1.14.0 **can emit SARIF**: `Report.ToSARIF()` / `WriteSARIFWithOpts` with `WithIncludeSuppressed` / `WithMinSeverity`                                                                                               | Module cache: `go-finding@v1.14.0/sarif_export.go:64/70/92/98/22/27`                                                                                 |
| A5 | Confirmed the full chain end-to-end: `Spec.Detect (finding.Detector)` → `([]Finding, error)` → `NewReportFromFindings` / `AddFindings` → `ToSARIF()` — and the nuance that toolsdk itself never calls ToSARIF                                 | `detector.go:10/18`, `report.go:97/125`, `spec.go:71`                                                                                                |
| A6 | Mapped the existing CI surface before answering: no SARIF upload exists in any of the 7 workflows; the Clone Detection lane uploads only a downloadable **artifact**, not Code Scanning SARIF                                                 | `grep` over `.github/workflows/` (0 SARIF hits); `.github/workflows/art-dupl-check.yml:39-45`                                                        |

All five answers were verified against the module cache and repo source, not from
memory — no fabricated API surface.

## b) PARTIALLY DONE

| #  | Item                               | What works                                                | What remains                                                                                                                                                                                | Effort                  |
| -- | ---------------------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| B1 | SARIF-on-GitHub enablement         | Complete investigation + working recipe delivered in chat | **Nothing implemented**: no workflow change, no README/docs update, no commit. The recipe exists only in this conversation                                                                  | S (workflow) / M (docs) |
| B2 | Durable memory of session findings | Findings are written down here                            | AGENTS.md was **not** updated at discovery time (aggressive-update-protocol miss — see D2) with: exit-0-with-findings contract, go-finding SARIF-export API map, toolsdk boundary statement | S                       |

## c) NOT STARTED

Planned/identified during this session, zero code written:

| #  | Item                                                                                                                                                                                     | Why not started                                          | Still wanted?                                          |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- | ------------------------------------------------------ |
| C1 | Add a SARIF upload lane (job/step) to `.github/workflows/`                                                                                                                               | Session was a question, not an implementation order      | Yes — user is visibly driving toward GitHub visibility |
| C2 | Publish a consumer-facing recipe (HOW_TO_USE.md / website): "Show clone alerts on GitHub" including `--no-actionability` and category caveats                                            | Same                                                     | Presumably; needs decision (see G1)                    |
| C3 | Emit `//art-dupl:accept`-accepted groups as go-finding SARIF **suppressions** (go-finding supports in-source suppression emission; `emit-suppressed-accepted` exists provider-side only) | Not asked; spotted the parity gap while verifying        | Candidate                                              |
| C4 | Upstream go-finding: a GitHub Code Scanning upload/exporter package                                                                                                                      | Out of scope; libraries deliberately stop at SARIF bytes | ROADMAP fuel                                           |
| C5 | TODO_LIST harvest of this report's section (f)                                                                                                                                           | User said report-then-wait                               | Required by the skill contract once work resumes       |

## d) TOTALLY FUCKED UP

Radical honesty. Nothing broke this session (no code was touched), but two real
defects were **noticed** and one process failure was **committed by me**:

| #  | What is broken                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | Severity                                             | Root cause                                                               | Mitigation                                                |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------------------ | --------------------------------------------------------- |
| D1 | **The "Clone Detection" gate cannot fail.** `.github/workflows/art-dupl-check.yml:37` runs `art-dupl check . -t 15 --json --output-dir reports \|\| true`. The `\|\| true` swallows the exit code, so the job is green even when new clones appear. This contradicts the workflow's own header comment ("Fails when new code clones are introduced") and the AGENTS.md CI-gate table, which lists "Clone Detection → Check for new clones" as a gate enforcing the baseline. If this job is a required status check, branch protection is enforcing a check that structurally cannot go red | High (false green on the repo's own product dogfood) | `                                                                        |                                                           |
| D2 | **Session produced zero durable artifacts.** I discovered D1 and a verified go-finding/toolsdk API map, and wrote none of it into AGENTS.md / TODO_LIST at discovery time — it only exists in chat until this report. Per the aggressive-update-protocol ("Update at the moment of discovery"), that is a miss. If this session had ended before the report, the D1 finding would have evaporated                                                                                                                                                                                           | Medium (process, not code)                           | Answer-mode momentum: treating Q&A sessions as exempt from memory duties | This report; harvest items F2/F3 into TODO_LIST on resume |
| D3 | **README oversells the CI reality** (docs drift, pre-existing, noticed not caused): README.md:85/111-112 presents SARIF + `upload-sarif` as the integration path, but no workflow in this repo does it. art-dupl's own repo does not dogfood its headline GitHub integration                                                                                                                                                                                                                                                                                                                | Low-Medium (credibility, not function)               | The upload lane was never built                                          | F1/F2                                                     |

## e) WHAT WE SHOULD IMPROVE

Process and design (not bugs — those are in d):

1. **Answer-sessions should still leave receipts.** When a Q&A session verifies
   facts against the module cache (A3–A5), those facts should land in AGENTS.md or
   a research ledger immediately; otherwise the next session re-pays the research
   cost. Concretely: record the toolsdk ≠ output-layer boundary and the
   Findings→Report→ToSARIF chain where SARIF is documented.
2. **Gate lanes must prove they can fail.** D1 is the same failure class as the
   canary rule already used for the docs-health count gate (verify the gate fails
   before trusting it). A one-time "introduce a fake clone, watch CI go red"
   drill would have caught the swallowed exit code months ago. Consider a
   quarterly gate-drill habit for every lane listed as a gate in AGENTS.md.
3. **Decide the SARIF severity contract once.** CLI SARIF emits `error` for large
   clones (`determineLevel`), while the BuildFlow provider caps advisory severity
   at warning (`maxAdvisorySeverity`). Once an upload lane exists, `error`-level
   alerts interact with branch-protection policies. One documented decision
   (ADR-length note is enough) avoids surprise red Security tabs.
4. **Suppression parity CLI ↔ SARIF.** The CLI filters actionability-suppressed
   and `//art-dupl:accept`-accepted groups out of output; go-finding has native
   SARIF suppression emission. Making SARIF carry suppressions (instead of
   dropping the groups) would let GitHub show _why_ something is not alerting —
   the same "consumers decide" philosophy already chosen for
   `emit-suppressed-accepted` in the provider.
5. **Recipe provenance in docs.** README code fences that reference CI mechanics
   (`upload-sarif`) should link or mirror an actual workflow file, so docs drift
   (D3) is grep-detectable. A tiny docs-lint could assert: every CI-mechanics
   fence cites a file that exists.

## f) NEXT TASKS (brainstorm, up to 50 — most are ROADMAP fuel, not commitments)

Ranked by impact. I = Impact (Critical/High/Medium/Low), E = Effort (S <30min /
M 30min–2h / L >2h), C = Category.

| #   | Task                                                                                                                                                                                 | I      | E | C             |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | - | ------------- |
| F1  | Add SARIF upload lane: `art-dupl --sarif` step + `github/codeql-action/upload-sarif@v3` (SHA-pinned per repo convention), `category: art-dupl`, `permissions.security-events: write` | High   | S | Feature       |
| F2  | Fix the swallowed gate: remove `\|\| true` from `art-dupl-check.yml:37` (or fail explicitly via step summary); then prove the lane can fail with a one-time canary drill             | High   | S | Bug           |
| F3  | Record in TODO_LIST: the D1 finding and the D2 process miss, so they survive this report                                                                                             | High   | S | Cleanup       |
| F4  | Update AGENTS.md with session-verified facts: exit-0-with-findings contract (`cmd/exit_codes.go:22`), go-finding SARIF-export API map, toolsdk boundary statement                    | Medium | S | Documentation |
| F5  | Decide + document the CLI-SARIF severity contract (`error` level vs branch protection; provider caps at warning — align or justify the difference)                                   | Medium | S | Documentation |
| F6  | Emit `//art-dupl:accept`-accepted groups as SARIF suppressions (reuse go-finding `WithIncludeSuppressed` path; provider already has the predicate pattern)                           | High   | M | Feature       |
| F7  | Add `--no-actionability` guidance + suppression caveat to any published SARIF recipe                                                                                                 | Medium | S | Documentation |
| F8  | Consumer recipe doc: "Show clone alerts on GitHub" in HOW_TO_USE.md and/or website (pending G1)                                                                                      | High   | M | Documentation |
| F9  | Emit `automationDetails.id` in CLI SARIF matching the upload `category`, so multiple art-dupl lanes on one repo don't clobber alerts                                                 | Low    | S | Feature       |
| F10 | Size-guard: warn (stderr) when SARIF output exceeds GitHub's 10 MB / 25k-result ingestion limits, with a "raise `-t`" hint                                                           | Low    | M | Feature       |
| F11 | Extend the flake `sarif-validate` check to round-trip CLI SARIF through go-finding `FindingsFromSARIF` (GroupID preservation is already pinned in `printer/finding` — reuse it)      | Medium | M | Quality       |
| F12 | Verify BDD/integration coverage for `--sarif` on a fixture repo (add if missing)                                                                                                     | Medium | M | Quality       |
| F13 | Security-tab badge in README once F1 exists                                                                                                                                          | Low    | S | Documentation |
| F14 | Decide whether `art-dupl check` should exit non-zero on new clones (flag or documented contract), unblocking the F2 fix semantics                                                    | Medium | S | Feature       |
| F15 | PR-annotation alternative output (check annotations) for consumers without Advanced Security                                                                                         | Medium | M | Feature       |
| F16 | Per-module SARIF runs with distinct categories for multi-module repos                                                                                                                | Low    | M | Feature       |
| F17 | Upstream go-finding: GitHub Code Scanning upload/exporter package (REST)                                                                                                             | Low    | L | Feature       |
| F18 | Upstream toolsdk: one-line doc note that it owns discovery only, never output/upload                                                                                                 | Low    | S | Documentation |
| F19 | Docs-lint: every CI-mechanics code fence in README must cite an existing workflow/doc file                                                                                           | Low    | M | Quality       |
| F20 | Rename `clone-detection-report` artifact to include the SARIF file when F1 lands                                                                                                     | Low    | S | Cleanup       |
| F21 | Document the threshold→level ladder in FEATURES.md/HOW_TO_USE.md                                                                                                                     | Low    | S | Documentation |
| F22 | October self-scan: confirm the monthly `-t 1 --type-aware` self-scan + ledger append happened; run if not (AGENTS.md cadence, 2026-09-22)                                            | Medium | M | Quality       |
| F23 | Docs-health monthly pass due 2026-10-28 (HARVEST this report's section f as entry point)                                                                                             | Medium | M | Documentation |
| F24 | Close the A-lite→A re-keying follow-up (tracked docs-health follow-up)                                                                                                               | Low    | M | Cleanup       |
| F25 | Daemon-hardening decision (jsonv2gate as pre-commit hook vs exclusion list) — flagged since report §g Q1 as the only structural fix for the 6× jsonutil regression class             | High   | M | Quality       |
| F26 | Decide fork→main sync: session sits on `fork`; auto-commits and docs are piling on the fork branch                                                                                   | Medium | S | Process       |
| F27 | Add `art-dupl check` CI-usage docs (exit codes, `\|\| true` pitfalls) so consumers don't copy the D1 pattern                                                                         | Medium | S | Documentation |
| F28 | Consider a `--sarif-file PATH` flag (today: stdout redirect only) to de-awkward CI wiring                                                                                            | Low    | S | Feature       |
| F29 | Verify SARIF golden output survives a Go/GOEXPERIMENT churn (belt-and-suspenders with the jsonv2gate; the wire format is v1-pinned)                                                  | Low    | S | Quality       |
| F30 | Record this session's go-finding/toolsdk verification in `docs/research/README.md` index (one line, per the HTML-report tracking policy)                                             | Low    | S | Documentation |

HARVEST routing note: F1–F5, F14, F25 are TODO_LIST material; the rest are
ROADMAP candidates pending prioritization.

## g) QUESTIONS ONLY YOU CAN ANSWER

1. **Who is the SARIF lane for — art-dupl itself, or art-dupl's users?** If it is
   self-dogfooding, I implement F1 in `.github/workflows/` next. If it is a
   published recipe for consumers (docs/website), the deliverable is F8 and F1
   becomes an example. Both is likely, but the order decides what "done" means —
   and I cannot infer intent from the repo (README hints at consumers, but the
   session's origin was your own visibility).
2. **Is the `\|\| true` in the Clone Detection lane deliberate?** The workflow
   header says it fails on new clones; the code cannot fail. If deliberate
   (report-only lane), the header and the AGENTS.md gate table are lying and I
   would fix the docs instead of the workflow. If not deliberate, F2 is a bug
   fix. I cannot determine intent from code.
3. **Should accepted clones alert on GitHub?** When SARIF is uploaded: should
   `//art-dupl:accept`-accepted groups (a) be dropped (today's CLI behavior),
   (b) appear as SARIF suppressions so reviewers see _why_ there is no alert, or
   (c) be a flag? This is a product/UX call about alert noise vs transparency;
   go-finding supports (b) natively, so it is a wiring decision, not a library
   limit.

---

## Wait State

Per instructions: report written, **waiting for further instructions**. Section (f)
is HARVEST-ready (`docs-health` → HARVEST) but was **not** executed, per the
"report only, then wait" scope.

### Format note

Skill default is a styled HTML dashboard; this report is `.md` because the
requesting prompt explicitly demanded `docs/status/<...>.md`. One-off override,
not propagated anywhere.
