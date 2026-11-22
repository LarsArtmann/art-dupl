# Library Review: linter-autoconfigure-sdk

| Field             | Value                                                              |
| ----------------- | ------------------------------------------------------------------ |
| Library           | `github.com/larsartmann/linter-autoconfigure-sdk`                  |
| Version reviewed  | `v0.8.1-7-gbfe96d7` (checkout `bfe96d7f921fa7c58f59650ecadee6c6f25f67b3`, branch `feat/audit-subpackage`) |
| Consumer          | `github.com/LarsArtmann/art-dupl` @ `cee67b4a47a8e759f5af05669cad97dca8d1f22d` (branch fork, clean at audit start) |
| Date              | 2026-11-22                                                         |
| Auditor           | Group B fleet run                                                  |
| Verdict           | **NOT-ADOPTED** — score **0/100** (0 of 10 applicable capabilities adopted; band: unrealized) |

## What the library is

A shared foundation for linter auto-configuration tools: config-file round-trip (typed read, deterministic marshal, crash-durable atomic write via `go-atomic-write`), a config diff engine (`DiffMaps`/`DiffSets`/`DiffBlobs`), finding emission for config issues (`FindingFromIssue` → go-finding `Finding`), a bridge to go-finding's canonical BuildFlow `toolsdk.Spec` (`ProviderFromSpec`), a generic generate-if-missing bootstrap lifecycle (`BootstrapProviderFromSpec[T]`, `linter-autoconfigure-sdk/bootstrap.go:128`), a go/analysis determinism analyzer (`determinism.NewAnalyzer` + `cmd/jsondeterminism` vettool), and — new on the audited branch, still `[Unreleased]` — an `audit` subpackage that detects dead linter references (exclusion rules, settings blocks, `//nolint` directives naming a linter that is neither enabled, disabled, nor default-enabled) in any golangci-lint-consuming repo (`linter-autoconfigure-sdk/audit/audit.go:1-31`, `linter-autoconfigure-sdk/CHANGELOG.md` `[Unreleased]` Added). Its README frames it precisely: "Adding a third auto-configurer becomes a config-schema exercise, not a from-scratch build" (`linter-autoconfigure-sdk/README.md`, "Why?" table). It deliberately does NOT parse YAML and does NOT model a `ProjectType` enum — those stay app-side (`linter-autoconfigure-sdk/README.md`, Design notes).

## Module map

- Single module `github.com/larsartmann/linter-autoconfigure-sdk` (root `go.mod`, `go 1.27`), packages: `autoconfigure` (root: `autoconfigure.go`, `diff.go`, `bootstrap.go`), `audit`, `determinism`, `cmd/audit`, `cmd/jsondeterminism`. No build tags; a testdata fixture go.mod exists for hermetic tests.
- Hard coupling: go-finding v1.15.1 + `go-finding/toolsdk` v1.16.0 (`linter-autoconfigure-sdk/go.mod:5-8`) — branded types (`finding.FilePath`, `finding.RuleName`) and the toolsdk contract are part of the public API, so a consumer imports the whole go-finding family transitively.
- art-dupl is a single module (`github.com/LarsArtmann/art-dupl`, root `go.mod` only). A dependency would pin as a standard `require github.com/larsartmann/linter-autoconfigure-sdk vX.Y.Z` in `art-dupl/go.mod` plus the nix `vendorHash` refresh dance documented in `art-dupl/AGENTS.md:241-239` (public-dep path: `go get` → set `vendorHash=""` → `nix build` → paste `got:`; or `buildflow -s nix-hash-fix --fix`). Because the module is public and tagged, no gogenfilter-style two-phase dummy/replace is needed (`art-dupl/AGENTS.md:223-229`).

## Capability inventory

A = 10 applicable rows. Score = round(100 × Σ/A) with every applicable row at 0 (nothing adopted) → **0**. N/A rows excluded with reasons.

| # | Capability | Would apply to art-dupl | Evidence (library) | Evidence (art-dupl hand-roll / absence) |
|---|-----------|-------------------------|--------------------|------------------------------------------|
| 1 | Typed config I/O: `ReadConfig`, `LoadJSON[T]`, `ParseJSON[T]` | Partial (art-dupl loads YAML/JSON config itself; different domain) | `linter-autoconfigure-sdk/autoconfigure.go:99,109,128` | `art-dupl/config/config_io.go` (YAML load/save for its own config; no autoconfigure-shaped byte-level layer) |
| 2 | Deterministic canonical marshal + idempotent atomic save (`MarshalJSONIndented`, `SaveJSON` write-if-changed) | Partial — philosophically adjacent to `internal/jsonutil`'s byte-stability contract, but art-dupl's contract is v1-API wire fidelity, not v2 determinism | `linter-autoconfigure-sdk/autoconfigure.go:155,213` | `art-dupl/internal/jsonutil/jsonutil.go` (v1-only marshal: no HTML escaping, no trailing newline); gate `art-dupl/internal/jsonv2gate/` |
| 3 | Config diff engine (`DiffMaps`/`DiffSets`/`DiffBlobs`, `FormatDiff`) | Low — art-dupl has no config-drift diff surface worth a dependency | `linter-autoconfigure-sdk/diff.go:43,71,103,147` | No equivalent (absence is fine; nothing consumes a diff) |
| 4 | Multi-candidate config discovery (`ProviderSpec.ConfigFiles`, `FirstExisting`) | No — art-dupl reads fixed paths (`.art-dupl-baseline.json`, its own config), no candidate-list discovery problem | `linter-autoconfigure-sdk/autoconfigure.go:500`; `bootstrap.go:297` | `art-dupl/baseline/baseline.go` (fixed path contract) |
| 5 | Finding emission for config issues (`FindingFromIssue`, `FindingsFromIssues`) | Yes-shaped, already satisfied — art-dupl already speaks go-finding natively for its own findings; a *lint-config* finding emitter would only matter if art-dupl audited lint configs via findings, which its guard scripts do better | `linter-autoconfigure-sdk/autoconfigure.go:288,336` | `art-dupl/pkg/provider/` + `art-dupl/printer/finding/` (native go-finding emission for clones, not config issues) |
| 6 | BuildFlow provider bridge (`ProviderFromSpec` → `toolsdk.Spec`) | No — art-dupl already registers its own toolsdk provider directly and correctly | `linter-autoconfigure-sdk/autoconfigure.go:437` | `art-dupl/pkg/provider/provider.go` (`toolsdk.Register` via blank-import pattern, AGENTS.md "BuildFlow toolsdk provider" bullet at `art-dupl/AGENTS.md:324`) |
| 7 | Bootstrap generate-if-missing lifecycle (`BootstrapProviderFromSpec[T]`) | No — art-dupl ships no config-generating tool; its invariant is the opposite ("the linter set is hand-curated; config sync for it is unwanted", `art-dupl/.buildflow.yml` last-known content, deleted by daemon commit `8ce90a47`) | `linter-autoconfigure-sdk/bootstrap.go:128,171,182,222,255` | `art-dupl/scripts/check-disabled-linters.sh:1-56` (repair = remove foreign additions, never generate) |
| 8 | Determinism analyzer (`determinism.NewAnalyzer`, `cmd/jsondeterminism`) | Adjacent but mis-aimed — it polices v2 `Marshal` without `json.Deterministic`; art-dupl's invariant is *no production v2 at all*, enforced by `jsonv2gate` | `linter-autoconfigure-sdk/determinism/determinism.go` | `art-dupl/internal/jsonv2gate/jsonv2gate_test.go` (AST gate; allowlist of exactly two deliberate test-support v2 files, `art-dupl/AGENTS.md:99-107`) |
| 9 | `audit` subpackage: dead linter references (exclusions / settings / `//nolint`) | Complementary, not substituting — checks the *opposite direction* of art-dupl's guard: dead refs vs banned refs. Genuinely useful as an ungated `go run` tool | `linter-autoconfigure-sdk/audit/audit.go:79-96`; rules `linter-autoconfigure-sdk/audit/exclusions.go:14`, `audit/settings.go:16`, `audit/nolint.go:18`; default set from live binary `audit/defaults.go:30` | `art-dupl/scripts/check-disabled-linters.sh:22` (banned-list guard, reverse semantics); rename-drift class visible at `art-dupl/.golangci.yml:359` (`exhaustruct_v5` exclusion entry — exactly the rename-orphan shape the audit catches) |
| 10 | Typed error model (`ConfigError` with `Op` enum, `Is`/`As`/`Unwrap`) | No — art-dupl has its own typed error package and sentinel-alias convention | `linter-autoconfigure-sdk/autoconfigure.go:73-98` | `art-dupl/errors/types.go` + sentinel alias convention (`art-dupl/AGENTS.md:211`) |

N/A rows excluded: none — all ten were assessed as applicable-or-partial above; the two fully N/A concerns (YAML parsing, `ProjectType` detection) are capabilities the library deliberately does not offer (`linter-autoconfigure-sdk/README.md` Design notes), so they are not inventory rows.

## Gap analysis: what art-dupl hand-rolls vs what the lib offers

**1. Banned-linter guard vs dead-reference audit — two halves of one problem, neither replaces the other.**
art-dupl guards *policy* (a curated linter list that foreign rewriters keep violating):

```bash
# art-dupl/scripts/check-disabled-linters.sh:22,32,38
DISABLED_LINTERS="exhaustruct tagliatelle"
if grep -qE "^[[:space:]]*-[[:space:]]+${linter}([[:space:]]|$)|..." "$CONFIG"; then
    grep -vE "..." "$CONFIG" >"$CONFIG.tmp" ...   # auto-removes the re-added linter
```

The SDK's audit checks *drift* (references to linters that no longer exist, e.g. after an upstream rename `exhaustruct → exhaustruct_v5`), deriving the valid set from the live binary so the audit cannot rot (`linter-autoconfigure-sdk/audit/audit.go:21-27`). art-dupl's config currently carries exactly the rename-drift shape the audit targets — `- exhaustruct_v5` inside an exclusions rule at `art-dupl/.golangci.yml:359` — while `- tagliatelle` sits in the enable list at `art-dupl/.golangci.yml:106` in direct contradiction of the ban recorded at `art-dupl/AGENTS.md:307` and `art-dupl/scripts/check-disabled-linters.sh:22`. That contradiction at the audited commit is live evidence for both halves of the problem: policy violation (guard's domain) *and* possible rename drift (audit's domain). **Adopting the SDK would not replace the guard**: the audit has no concept of a policy-banned-but-valid linter (tagliatelle IS valid to golangci-lint; art-dupl bans it for wire-format reasons). The two checks are orthogonal.

**2. Atomic config writes vs art-dupl's sed-removal repair.** `SaveJSON` (`linter-autoconfigure-sdk/autoconfigure.go:213`) is idempotent, crash-durable, and race-safe; the guard's `grep -v >tmp && mv` (`art-dupl/scripts/check-disabled-linters.sh:38-39`) is none of those. But the guard repairs a 449-line YAML file by line filtering — a domain the SDK explicitly does not enter (no YAML support, `linter-autoconfigure-sdk/README.md` Design notes). No leverage.

**3. The recurring BuildFlow rewriter war — the decisive question.** art-dupl's chronic operational pain is BuildFlow-side steps violating local config invariants: `golangci-lint-auto-configure` re-adding banned linters (skipped via `skip_steps`, `art-dupl/.buildflow.yml` last-known lines 18-30), and `go-auto-upgrade` re-poisoning six v1-API files to `encoding/json/v2` across recurrences #16-#19 (`art-dupl/AGENTS.md:348-395`), defended only by `skip_steps` + the `jsonv2gate` AST gate + restore-from-parent discipline. The SDK sits *inside that machinery*: it is the plumbing layer of the very tools doing the poisoning (golangci-lint-auto-configure is a named consumer, `linter-autoconfigure-sdk/README.md` Consumers). **The SDK offers no consumer-side mechanism to declare "this config region is policy-immutable"** — no intent manifest, no protected-ranges API, no honor-the-pin contract. `respect_patch_floor` (the fix that ended the go-line flip-flop, `art-dupl/AGENTS.md:116-123`) landed in go-version-auto-configure itself, not in anything art-dupl could import. Adopting the SDK in art-dupl would add a dependency to enforce invariants art-dupl already enforces more precisely and more cheaply.

**Which side the value lands on: HOST-SIDE, unambiguously.** The SDK's leverage accrues to BuildFlow and the `*-auto-configure` family (better diffs, atomic writes, bootstrap safety). art-dupl's role in that relationship is *victim of policy collisions*, and the remedies that worked (`respect_patch_floor`, `skip_steps`, upstream plumbing fixes) all live host-side. If anything, the correct fleet move is the inverse of adoption: push the *policy* half (a consumer-declared config-intent manifest that rewriters must honor) into BuildFlow/the auto-configurers — ideally modeled ON this SDK, since that is where the rewriter code lives. art-dupl-side, the honest cheap win is `go run github.com/larsartmann/linter-autoconfigure-sdk/cmd/audit` as an optional CI lane: it adds the dead-reference check with zero go.mod impact (not counted as adoption in the score above, which measures module adoption).

## Verdict + derived score

**NOT-ADOPTED.** Inputs: 10 applicable capability rows, 0 adopted (Σ=0) → **0/100**, band "unrealized" (0-24). The score states unrealized potential honestly: several capabilities are *good* (bootstrap lifecycle, diff engine, atomic writes), but none maps to a job art-dupl actually does or suffers from. The library is already a success in the fleet — six sibling repos pin it (see Fleet argument) — art-dupl is simply not in its audience.

## Version currency

- Local describe: `v0.8.1-7-gbfe96d7`; branch `feat/audit-subpackage`, last commit 2026-11-20 19:36 (+0100). Seven commits ahead of `v0.8.1`.
- Tag cadence: `v0.1.0` 2026-09-10 → `v0.2.0` 2026-09-11 → v0.3.x/0.4.x/0.5/0.6/0.7 all 2026-09-23 → `v0.8.0` 2026-10-03 → `v0.8.1` (unpublished-in-CHANGELOG patch; the audited branch's `[Unreleased]` carries the audit subpackage). Cadence: bursty but active — 8+ tags in 6 weeks, then a 6-week quiet stretch now ending with the audit-subpackage feature.
- `[Unreleased]` intent (`linter-autoconfigure-sdk/CHANGELOG.md`): the audit subpackage + first non-larsartmann dependency (`go-faster/yaml`). This is the feature with the most art-dupl relevance, and it is not yet tagged. Realistic ETA: next minor `v0.9.0`; given the 2026-10-03 precedent (0.8.0 landed ~10 days after feature commits), a tag within 1-2 weeks of the branch merging is plausible. The fleet-audit TODO item (T21 BuildFlow CI, T22 toolsdk declarative pass-through) shows the host-side direction continues.

## Fleet argument

Lars's fleet already voted: `~/projects` go.mod files show `linter-autoconfigure-sdk v0.8.1` required by `golangci-lint-auto-configure`, `oxlint-auto-configure`, `go-version-auto-configure`, `dependabot-auto-configure`, `clean-wizard`, and indirectly by `BuildFlow` — plus `library-policy` consuming the sibling DSL. The collective-improvement argument (every repo reimplementing config plumbing multiplies bugs) is *valid* and *already satisfied*: the plumbing IS centralized. art-dupl is the outlier because its relationship to this machinery is adversarial, not collaborative — it is the repo whose hand-curated configs get rewritten by the fleet tools. Centralizing more code into the SDK does not reduce art-dupl's pain; what propagates fleet-wide is *policy*, and policy currently has no home in the SDK at all. The strongest fleet play is therefore host-side: extend the SDK (or toolsdk) with a consumer-declared config-intent/policy manifest that `skip_steps` today approximates crudely — then art-dupl's `scripts/check-disabled-linters.sh` sed-guard and the `go-auto-upgrade` skip both become expressible declaratively, and every fleet repo inherits the fix. That work belongs in the SDK's repos, not in art-dupl's go.mod.

## Ranked opportunities (impact × ease)

1. **Run `cmd/audit` as an ungated CI lane** (`go run .../cmd/audit` — no dependency, no go.mod change). Impact: medium (catches rename-drift like `exhaustruct_v5` at `art-dupl/.golangci.yml:359` and orphaned `//nolint`s). Ease: trivial once `v0.9.0` tags.
2. **Host-side: propose a config-policy manifest in toolsdk** (consumer declares banned/protected config keys; rewriters consult before write). Impact: high — kills the entire recurrence ledger class (#16-#19) fleet-wide. Ease: hard, upstream, owner-gated (this is AGENTS.md's pending Q1, `art-dupl/AGENTS.md:371-373`).
3. **Not recommended:** importing the SDK module into art-dupl for config I/O or provider bridging. Impact: nil (art-dupl already has both concerns, natively). Ease: easy but pointless — dependency weight (go-finding family) for zero capability delta.

## RECONSIDER-WHEN

Flip to ADOPT-CANDIDATE when **any** of:
- The SDK (or toolsdk) ships a **consumer-side declarative config-intent API** — banned-linter/protected-key manifests that auto-configurer rewriters honor at repair time. That would make `art-dupl/scripts/check-disabled-linters.sh` and the `.buildflow.yml` `skip_steps` block expressible as data instead of shell+sed, and would justify adoption for the policy layer alone.
- art-dupl grows a **config-generating or config-auditing tool of its own** (e.g. a `art-dupl init` that bootstraps `.art-dupl.yml`, or first-class lint-config findings) — then `BootstrapProviderFromSpec`/`FindingFromIssue` are direct replacements for hand-rolled equivalents.
- The **audit subpackage tags** and art-dupl's lint-config guard evolves beyond two banned names — then import the module (or stay on `go run`) for the drift check.
- art-dupl ever adopts **json/v2-style marshaling with determinism requirements** — `cmd/jsondeterminism` slots into CI exactly like `jsonv2gate` does today.

## How to verify (appendix — exact commands)

```bash
# versions / state
git -C /home/lars/projects/art-dupl rev-parse HEAD                      # cee67b4a... (worktree live; audit pinned to this commit)
git -C /home/lars/projects/linter-autoconfigure-sdk rev-parse HEAD      # bfe96d7f...
git -C /home/lars/projects/linter-autoconfigure-sdk describe --tags     # v0.8.1-7-gbfe96d7
git -C /home/lars/projects/art-dupl show 8ce90a47^:.buildflow.yml       # last-known skip_steps (file deleted by daemon commit 8ce90a47)
git -C /home/lars/projects/go-policy-dsl status --short                 # (for report 2)

# consumer evidence
grep -n "linter-autoconfigure-sdk" ~/projects/{golangci-lint-auto-configure,oxlint-auto-configure,go-version-auto-configure,dependabot-auto-configure,clean-wizard,BuildFlow}/go.mod
sed -n '22p;32p;38p' /home/lars/projects/art-dupl/scripts/check-disabled-linters.sh
grep -n "tagliatelle" /home/lars/projects/art-dupl/.golangci.yml         # :106 — ban violated at audited commit
grep -n "exhaustruct_v5" /home/lars/projects/art-dupl/.golangci.yml      # :359 — rename-drift shape
grep -n "exhaustruct\|tagliatelle" /home/lars/projects/art-dupl/AGENTS.md  # :155, :307 ban record

# library capability evidence
grep -n "^func " /home/lars/projects/linter-autoconfigure-sdk/autoconfigure.go
sed -n '1,31p' /home/lars/projects/linter-autoconfigure-sdk/audit/audit.go
cat /home/lars/projects/linter-autoconfigure-sdk/CHANGELOG.md           # full read incl. [Unreleased]
cat /home/lars/projects/linter-autoconfigure-sdk/README.md /home/lars/projects/linter-autoconfigure-sdk/TODO_LIST.md

# deliberately NOT run: art-dupl/scripts/check-disabled-linters.sh (auto-fixes = mutates .golangci.yml; read-only mandate)
```
