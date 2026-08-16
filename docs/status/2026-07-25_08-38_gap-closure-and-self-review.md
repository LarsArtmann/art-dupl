# Status Report: Gap-Closure Sprint & Brutal Self-Review

> **Date:** 2026-07-25 08:38
> **Session scope:** Followed the "Exact Next Steps" from the prior session's handoff. Closed all 3 known gaps (JSON field, FEATURES.md, writeExplanation test), fixed the `.golangci.yml` daemon regression (4th time), answered the 3 open design questions, then self-critiqued hard.
> **Branch:** `fork`

---

## A) FULLY DONE (verified green)

### A1. `.golangci.yml` cleaned (4th reversion) + ROOT CAUSE FOUND

**What:** Removed all 3 references to `exhaustruct` + `tagliatelle` (enable list x2 + config block x1).

**Root cause identified (Q3 from prior session):** The auto-git daemon commits as `Unknown Author <unknown@example.com>`. I ran `git show 4b33fc05 -- .golangci.yml` and confirmed that **single commit** re-added all 3 references in one diff. The daemon runs a `golangci-lint` config normalization/migration that re-enables **all available linters** and re-indents the file (2-space -> 4-space), then commits **directly to git WITHOUT running the CI guard** (`scripts/check-disabled-linters.sh`). The guard catches it in `nix flake check`, but the daemon's commit path bypasses CI.

**What I did NOT do (and why):** I cannot modify the external daemon. The durable fix requires either (a) a daemon-side exclude list for `.golangci.yml`, or (b) a pre-receive/CI gate that rejects commits touching `.golangci.yml` containing forbidden linters. Neither is actionable from the codebase.

### A2. JSON gap CLOSED — `NonActionablePattern` now serialized

**Files changed:**

- `printer/json.go:32-41` — added `NonActionablePattern string` field to `JSONClone` struct with tag `json:"non_actionable_pattern,omitempty"`.
- `printer/json.go:52-62` — mapped the field in `toJSONClone()` (the single conversion point).
- Deliberately left `simpleJSONClone` unchanged — it is intentionally minimal (CloneRef + token_count only) and should NOT carry classification metadata.

**Verified:** JSON output now includes the field; round-trips via `json.Unmarshal`.

### A3. Unit tests added — JSON field + `writeExplanation` branches

**`printer/json_test.go`:**

- `TestToJSONClone_NonActionablePattern` — verifies the mapping function populates the field.
- `TestJSONPrinter_NonActionablePatternSerialized` — end-to-end: `OutputJSON` -> `json.Unmarshal` -> assert field value.

**`printer/text_test.go`:**

- `TestTextPrinter_writeExplanation` — 6 subtests covering ALL branches:
  1. actionable clone (checks `notWant: non-actionable, extractable`)
  2. non-actionable WITH pattern (`non-actionable (guard-clause)`)
  3. non-actionable WITHOUT pattern (defaults to `(boilerplate)`)
  4. extractable shows savings (`extractable: ~12 lines saved across 3 sites`)
  5. suggestion on actionable uses `fix:` label
  6. suggestion on non-actionable uses `why:` label

All 8 test cases pass (`go test -run` confirmed).

### A4. FEATURES.md updated

Added 3 rows to the Refactoring Advisor section (`Actionability Override`, `Explain Mode`, `Pattern in JSON`) and 2 rows to the Professional CLI section (`Explain Mode`, `Actionability Toggle`). Links the flags to their output behavior.

### A5. Design questions answered (Q1, Q2, Q3)

Documented in `AGENTS.md:111` and appended **Section H (Resolutions)** to this report's predecessor (`docs/status/2026-07-25_08-21_no-actionability-and-explain-flags.md`):

- **Q1 (`NonActionablePattern` typing):** keep `string` in domain. The pattern taxonomy is a `printer/` output concern; moving `PatternLabel` to `domain/` inverts the dependency wrongly.
- **Q2 (subcommand scope):** keep root-only. `stats`/`baseline`/`check` have output semantics where per-clone actionability filtering doesn't map.
- **Q3 (daemon):** root cause found (see A1).

### A6. Full verification suite — ALL GREEN

| Check                  | Command                                                            | Result                 |
| ---------------------- | ------------------------------------------------------------------ | ---------------------- |
| Build                  | `GOEXPERIMENT=jsonv2 go build ./...`                               | PASS                   |
| Tests                  | `go test ./... -count=1`                                           | PASS (all 26 packages) |
| Race                   | `CGO_ENABLED=1 go test -race ./printer/... ./domain/... ./cmd/...` | PASS                   |
| Lint                   | `nix build .#checks.x86_64-linux.lint`                             | PASS (0 issues)        |
| Disabled-linters guard | `nix build .#checks.x86_64-linux.disabled-linters`                 | PASS                   |
| BDD flags              | `go test ./bdd/... --ginkgo.focus="no-actionability                | explain flag"`         |

---

## B) PARTIALLY DONE

### B1. The 3 known gaps are closed, but the status report's "47-item Pareto" is untouched

I only executed the handoff's "Exact Next Steps" (6 items). The prior session's status report (`docs/status/2026-07-25_08-21...`) lists **50 items** across 5 tiers. I did not start ANY of them. The highest-ROI untouched item is **#1: `--diff-report baseline` mode (2h, HIGH)**.

### B2. AGENTS.md "Lint config" note is now slightly stale

The AGENTS.md says: _"`exhaustruct` and `tagliatelle` are NOT in the `.golangci.yml` enable list... A CI guard prevents them from being re-added."_ This is technically true, but the guard only runs in `nix flake check`, NOT on the daemon's commit path. The note should warn that the daemon re-adds them periodically and that `nix build .#disabled-linters` is the authoritative check (not a naive `golangci-lint run` which would just use whatever's in the file).

---

## C) NOT STARTED (from the handoff / Pareto backlog)

The following were listed as "Exact Next Steps" or Pareto items and I did **zero** work on them this session:

1. **`--diff-report baseline` mode** (Pareto Tier 1, #1, 2h, HIGH) — not started.
2. **HTML report improvements** (Pareto Tier 1) — not started.
3. **YAML config support** (Pareto Tier 1) — not started.
4. **SARIF rule metadata for actionability patterns** (TODO #49) — not started. SARIF output still has no rule tags for which pattern matched.
5. **`--explain` for JSON format** (TODO #48) — I added the raw `non_actionable_pattern` STRING field, but there is no structured explanation object in JSON. A JSON consumer still can't get the formatted "why" rationale, only the pattern identifier.
6. **Golden file tests for text printer** (TODO #35) — not started. I added unit tests for `writeExplanation` but no golden-file regression guard for the full text output.
7. **Pre-receive/CI gate for `.golangci.yml`** (infra, TODO #37) — not started. This is the durable fix for A1.
8. **ADR for `PatternLabel` location** (TODO #43) — I made the decision (keep `string` in domain) and documented it, but did not write a formal ADR file in `docs/adr/`.

---

## D) TOTALLY FUCKED UP (nothing — this session was clean)

No regressions introduced. No tests broken. No reverts needed. The daemon re-added the linters but that is external, not my error.

**One honest caveat:** I did NOT run `golangci-lint fmt` (gci/goimports/gofumpt/golines) on my two edited test files (`json_test.go`, `text_test.go`) even though the LSP was emitting `golines`/`gci` warnings on OTHER files. Those warnings are pre-existing (on `filter_bench_test.go`, `filter_stats_test.go`) and not mine, but I should have verified my own files were fmt-clean before declaring done. The `nix build .#lint` check PASSED, which runs the full formatter, so my files are in fact clean — but I didn't verify it explicitly per-file.

---

## E) WHAT WE SHOULD IMPROVE

### E1. The daemon is an unaddressed systemic threat

This is the **4th time** the forbidden linters were removed. The pattern is now fully diagnosed but the fix is out of reach from the codebase. Every session will keep paying this tax until either the daemon is configured with an exclude list OR a pre-receive hook rejects the commit. This should be the #1 infrastructure priority.

### E2. I worked to the handoff, not to the product vision

I treated the "Exact Next Steps" as a checklist and stopped when the checklist was empty. A more ambitious session would have, after closing the gaps, immediately picked up Pareto item #1 (`--diff-report baseline`) and shipped it. I left velocity on the table.

### E3. `NonActionablePattern` as `string` is a known smell I accepted

I rationalized keeping it as `string` to preserve the domain/printer boundary. That's defensible, but it means typos in the pattern string (e.g., a consumer checking `== "guard-clause"`) are uncaught at compile time. A formal ADR would lock in the reasoning and make the tradeoff auditable instead of buried in a status report.

### E4. No integration test for the `--explain` JSON round-trip

I tested `toJSONClone` and `OutputJSON` in isolation. There is no BDD test that runs `art-dupl --explain --json` end-to-end and asserts the structured field appears on a REAL clone. The BDD suite only covers the text `explain:` line.

### E5. I didn't touch the SARIF gap

SARIF is a first-class output format and I left it without the pattern metadata. For GitHub Advanced Security consumers, `non_actionable_pattern` should surface as a rule tag. This is a real feature gap, not just docs.

---

## F) Up to 50 things to do next

Grouped by theme, rough priority order within each group.

### F1. High-impact features (Pareto Tier 1)

1. `--diff-report baseline` mode — diff current run vs baseline, report only NEW clones (2h)
2. HTML report: add `--explain` content as a tooltip/column on clone groups
3. HTML report: show `non_actionable_pattern` badge on suppressed clones
4. YAML config file support (`--config config.yaml`) alongside JSON
5. `--explain` for JSON: structured `explanation` object, not just the pattern string
6. SARIF: emit actionability pattern as `rule.tags` + `rule.id`
7. SARIF: emit clone type as a rule property

### F2. Testing & verification

8. BDD test: `--explain --json` end-to-end asserts `non_actionable_pattern` on a real clone
9. Golden file tests for text printer (`printer/text_golden_test.go` expand) — catch formatting regressions
10. Golden file tests for `--explain` output specifically
11. Fuzz test `writeExplanation` (classification -> string mapping never panics)
12. Property-based test for `classifyCloneType` (TODO #50)
13. Add `gosec` + `govulncheck` to Nix checks (TODO #38)
14. Fuzz test in CI dedicated workflow (TODO #36)
15. Race test the full `bdd/` suite, not just printer/domain/cmd
16. Test that `simpleJSONClone` does NOT have `non_actionable_pattern` (negative guard against drift)

### F3. Infrastructure & daemon defense

17. **Pre-receive hook / CI gate rejecting `.golangci.yml` commits containing forbidden linters** (durable daemon fix)
18. Daemon-side exclude list for `.golangci.yml` (requires daemon config access)
19. Pre-commit hook running `golangci-lint fmt` (TODO #39)
20. Dependabot / nix flake update automation (TODO #40)
21. Release automation via GitHub Actions + `nix build` versioned binaries (TODO #41)
22. ADR for `generatorIncludes` design (TODO #42)
23. ADR for `PatternLabel` location — record the `string`-in-domain decision (TODO #43)
24. Cache format versioning — bump on incompatible `Node` changes (TODO #44)
25. Exit code 4 (clones found) distinct from 1 (error) (TODO #46)
26. Watch mode (`--watch`) re-run on file change (TODO #47)

### F4. Code quality / refactors

27. Split `printer/` into sub-packages (TODO #26-30) — blocked by circular dep on `StatsPrinter`
28. Move `Printer`/`ReadFile`/`StatsPrinter` to `printer/base/`
29. Extract HTML printer to `printer/html/`
30. Extract JSON printer to `printer/json/`
31. Extract text printer to `printer/text/`
32. Extract SARIF printer to `printer/sarif/`
33. Decouple `actionability.go` from `syntax.Node` via `domain.ProcessedClone` (TODO #32)
34. Eliminate `printer.simpleJSONClone` — fold into `JSONClone` (TODO #33)
35. Add `context.Context` to `Printer` interface methods (TODO #34)
36. Populate `NoActionability` in `baseline record` `SuppressionConfig` site (the one valid Q2 use case I deferred)

### F5. Documentation

37. Update AGENTS.md "Lint config" note to warn about the daemon reversion cycle + that `nix build .#disabled-linters` is authoritative
38. HOW_TO_USE.md: add `--explain --json` example showing the `non_actionable_pattern` field
39. `docs/ACTIONABILITY_PATTERNS.md`: cross-reference the JSON field name
40. ADR-0017 (or next number): formalize `NonActionablePattern` as `string` in domain
41. ROADMAP.md: reflect closed gaps + reprioritize

### F6. Detection & UX

42. Type-aware detection: cache `go/packages` results between runs (TODO #23)
43. Structural mode: option to ignore comments (TODO #24)
44. Generics instantiation detection: `Foo[int]` vs `Foo[string]` (TODO #25)
45. Progress output: ETA calculation for large codebases (TODO #45)
46. `--min-tokens` flag (complement to `--min-lines`)
47. Configurable actionability patterns (let users disable specific patterns, e.g. `--disable-pattern guard-clause`)
48. `--explain` on `stats` subcommand: aggregate "top suppressed patterns" summary
49. Streaming JSON output (NDJSON) for large codebases
50. Web playground / WASM build for browser-based demo

---

## G) Questions I CANNOT figure out myself

### G1. Do you want the durable daemon fix to be a **pre-receive git hook** (server-side, requires repo admin) or a **CI workflow** that fails the build on forbidden-linter commits?

I cannot tell which side of the repo you control. A pre-receive hook blocks the commit before it lands; a CI workflow lets it land but fails the check run (and the daemon's commit is already in history). The former is cleaner but needs admin access I don't have. The latter is something I can implement in `.github/workflows/` right now. Which do you want?

### G2. Should I write a formal **ADR** for the `NonActionablePattern`-as-`string` decision now, or is the AGENTS.md + status-report documentation sufficient?

TODO #43 calls for an ADR on `PatternLabel` location. I made the decision and documented the reasoning informally. Writing a formal `docs/adr/0007*-patternlabel-location.md` would make it auditable and linkable, but it's overhead if you consider the AGENTS.md note binding. I cannot tell whether your team treats AGENTS.md as the authoritative decision log or whether ADRs are required for architecture-relevant decisions.

### G3. For the next sprint, do you want me to prioritize **closing the output-format parity gap** (SARIF + JSON `--explain` + HTML badges — items F1.5-7) or **new detection value** (the `--diff-report baseline` mode — item F1.1)?

Both are high-impact but serve different users. The output-format work benefits CI/GitHub-Security consumers (SARIF) and refactoring workflows (structured explanations). The baseline-diff mode benefits teams running `art-dupl check` in CI who want "new clones since last green build." I can't infer which user segment you care about more from the codebase alone.
