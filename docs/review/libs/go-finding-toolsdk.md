# Library utilization audit — go-finding/toolsdk

| Field                      | Value                                                                                                                                                                                                                                            |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Library                    | `github.com/larsartmann/go-finding/toolsdk` (sub-module of the go-finding workspace)                                                                                                                                                             |
| Module path covered here   | `github.com/larsartmann/go-finding/toolsdk` (core module covered in `docs/review/libs/go-finding.md`)                                                                                                                                            |
| Library checkout reviewed  | `/home/lars/projects/go-finding` @ `5f568e058a15221ac198d378a1d907231451edf0` (= root tag `v1.16.0`); the toolsdk tag `toolsdk/v1.16.0` points at `9c675c7` (2026-10-11) — the module's content at both commits matches for every API cited here |
| Consumer checkout reviewed | `/home/lars/projects/art-dupl` @ `cee67b4a47a8e759f5af05669cad97dca8d1f22d` (fork, clean)                                                                                                                                                        |
| Consumer pin               | `github.com/larsartmann/go-finding/toolsdk v1.16.0` (go.mod:91) — **current** (latest tag)                                                                                                                                                       |
| Date                       | 2026-11-22                                                                                                                                                                                                                                       |
| Auditor                    | group A fleet run (read-only audit; this file is the only write)                                                                                                                                                                                 |

## Prior coverage (superseded by this report)

- `docs/research/2026-10-05_go-finding-deep-dive.html` — §01 top finding (options channel
  unused, provider threshold contract frozen), §05.1 ledger (T01 threshold options channel
  SHIPPED 2026-10-06: provider declares `threshold` int 1–1000 + `emit-suppressed-accepted`
  bool), §05.5 re-scored table (options channel → Fully leveraged), §06 currency re-review
  (2026-10-11) incl. §06.2 verdicts (`Option.Validate` checks declarations only — value-range
  enforcement at Detect IS the D1 contract; SwitchCases/Repairer/ModuleFanOut N/A re-confirmed;
  SnapshotForTest not needed). **This report supersedes those rows as the current toolsdk
  adoption truth**; the deep-dive verdicts remain the canonical rationale and were re-read
  (per AGENTS.md:241) before judging. Not re-litigated without new evidence.
- BuildFlow provider context: AGENTS.md:324 (provider package doc, options D1, suppression
  option T06, crawl policy, LIVE in BuildFlow since 2026-09-25).

## Module map

go-finding is a five-module workspace; toolsdk is the host-integration SDK (declarative tool
specs + process-global registry + context channels). Tags are per-module and independent:

| Module         | Import path                                        | Latest tag                       | Tag date   | art-dupl consumes?                                |
| -------------- | -------------------------------------------------- | -------------------------------- | ---------- | ------------------------------------------------- |
| toolsdk        | `github.com/larsartmann/go-finding/toolsdk`        | `toolsdk/v1.16.0` (at `9c675c7`) | 2026-10-11 | **Yes — pinned `v1.16.0` (current)**              |
| core (root)    | `github.com/larsartmann/go-finding`                | `v1.16.0` (== HEAD)              | 2026-11-22 | Yes — pinned `v1.15.0` (lagging; see core report) |
| analysis       | `github.com/larsartmann/go-finding/analysis`       | `analysis/v1.15.1`               | 2026-11-08 | No                                                |
| pipeline       | `github.com/larsartmann/go-finding/pipeline`       | `pipeline/v1.15.1`               | 2026-11-08 | No                                                |
| cmd/go-finding | `github.com/larsartmann/go-finding/cmd/go-finding` | `cmd/go-finding/v1.15.1`         | 2026-11-08 | No                                                |

art-dupl is a single Go module and is itself a toolsdk **provider** (the consumer of the SDK),
not a host: `pkg/provider` self-registers art-dupl as a BuildFlow tool; BuildFlow is the host
that calls `toolsdk.All()` / `WithOptions` / `EnsureContext` on the other side.

Note on the published tag: `git show toolsdk/v1.16.0:toolsdk/go.mod` requires core `v1.14.0`
(the workspace `replace => ../` is inert when fetched as a module), so art-dupl's explicit core
`v1.15.0` wins MVS — the mixed pin is stable and toolsdk v1.16.0 compiles against it.

## Current usage map

All citations are `path:line` in art-dupl unless prefixed `go-finding/` (library checkout).
The provider is the sole toolsdk consumer in art-dupl (tests aside):

| Consumer site                      | What it uses                   | toolsdk symbols                                                                                                                                                                                                                                                                                 |
| ---------------------------------- | ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/provider/provider.go:87–124`  | Self-registration spec         | `toolsdk.Register`, `toolsdk.Spec` (Name:88, Description:89–94, Trigger:99 `OnFiles("go", "**/*.go", "**/*.templ")`, Inputs:100, Options:101–117, Detect:118, HealthCheck:123)                                                                                                                  |
| `pkg/provider/provider.go:135–136` | Detect entry + context hygiene | `cloneDetector` implements `gofinding.Detector`; `toolsdk.EnsureContext(ctx)`                                                                                                                                                                                                                   |
| `pkg/provider/provider.go:152–160` | Per-run option reads           | `thresholdFromContext` (local), `boolOptionFromContext` (local) — built ON `toolsdk.OptionsFromContext` (pkg/provider/provider.go:203, 223)                                                                                                                                                     |
| `pkg/provider/provider_test.go`    | Contract pins                  | `TestProviderRegistersValidSpec` (:82), `TestDetectThresholdOption` (range-reject 0/1001 per D1), `TestDetectEmitSuppressedAcceptedOption`, `TestEmitSuppressedAcceptedOffIsByteIdentical`, `TestProviderDetect_ConcurrentSameWorkingDir` (:409), `TestDetect_MultiModuleWorkspaceShape` (:609) |
| BuildFlow side (out of repo)       | Host integration               | blank import of `pkg/provider`; `toolsdk.All()` discovery; `toolsdk.WithOptions` per-run values (AGENTS.md:324)                                                                                                                                                                                 |

Identifier footprint: 23 `toolsdk.` references in `pkg/provider/provider.go`, plus
`provider_test.go`.

## Capability inventory

Status rubric: **Fully Leveraged / Partially Used / Missed Opportunity / Misused / Not
Applicable** (every N/A carries a one-line reason). 15 rows.

| #  | Capability                                                                                                               | Status                                                                              | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Impact                                                                                                              |
| -- | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------- |
| 1  | Spec + process-global registry (`Register`, `All`)                                                                       | Fully Leveraged                                                                     | pkg/provider/provider.go:87 ~`toolsdk.Register` (`var Provider = toolsdk.Register(...)` — package-level var = the entire BuildFlow wiring step via blank import)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Zero host-side glue; live in BuildFlow since 2026-09-25                                                             |
| 2  | Trigger declaration (`OnFiles` language + patterns; vs `OnGoFiles`/`OnGoModule`)                                         | Fully Leveraged                                                                     | pkg/provider/provider.go:99 — `OnFiles("go", "**/*.go", "**/*.templ")` is the correct form: `OnGoFiles`/`OnGoModule` would drop `.templ` coverage (§06.2 verdict); empty `Requires` deliberate (works on non-module trees, M59 §f#24, pkg/provider/provider.go:95–98)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Fires on exactly the files it can parse                                                                             |
| 3  | Trigger validation at `Register` (v1.16.0, issue #44; `ErrInvalidTrigger`)                                               | Fully Leveraged (passive)                                                           | art-dupl's Trigger is validation-eligible and valid: language `"go"` ∈ `ValidLanguages()`, patterns non-empty (go-finding/toolsdk/spec.go:97–178); `TestProviderRegistersValidSpec` pins the registration                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Malformed-trigger startup failures now impossible by construction; nothing for art-dupl to do                       |
| 4  | Options channel — declaration + kind contract (`Spec.Options`, `Option`, `OptionKind*`, consumer-side `ValidateOptions`) | Fully Leveraged                                                                     | pkg/provider/provider.go:101–117: `threshold` (int, default `artdupl.DefaultThreshold`, 1–1000 range enforced at Detect per the D1 contract — §06.2 dissolved verdict) and `emit-suppressed-accepted` (bool, default false load-bearing); `TestDetectThresholdOption`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Consumers tune per run without rebuilds                                                                             |
| 5  | **Typed option accessors `StringFromContext`/`IntFromContext`/`BoolFromContext` (v1.16.0, issue #43)**                   | ~~Missed Opportunity~~ **ADOPTED 2026-11-23 (errata, closed by execution plan T4)** | ERRATA 2026-11-23: the provider now reads options via `toolsdk.IntFromContext` (pkg/provider/provider.go:186 ~`IntFromContext`) and `toolsdk.BoolFromContext` (pkg/provider/provider.go:193 ~`BoolFromContext`); the hand-rolled `thresholdFromContext`/`boolOptionFromContext` helpers cited below are DELETED, and a wrong-kind value that bypassed `ValidateOptions` now fails Detect loudly with fleet-semantics `ErrOptionKindMismatch` (pinned by `TestDetectWrongKindOptionBypassFailsLoudly`) instead of silently defaulting. Original finding, for history: pkg/provider/provider.go:201–219 (`thresholdFromContext`) and pkg/provider/provider.go:221–241 (`boolOptionFromContext`) hand-rolled the lookup/assert/fallback plumbing the accessors ship — go-finding/toolsdk/options.go:142, 149, 156, **present in the pinned tag** (`git show toolsdk/v1.16.0:toolsdk/options.go`). Upstream shipped them explicitly "so tools stop hand-rolling the lookup/assert/fallback plumbing" | Adopted state: ~40 lines of duplicated option plumbing deleted; wrong-kind handling now matches the fleet read path |
| 6  | `EnsureContext`                                                                                                          | Fully Leveraged                                                                     | pkg/provider/provider.go:163 ~`EnsureContext` (nil-context hardening before any read, inside the shared `detectWithCoverage` run core)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Correct contract at the Detect boundary                                                                             |
| 7  | `HealthCheck`                                                                                                            | Fully Leveraged                                                                     | pkg/provider/provider.go:123 ~`HealthCheck` — self-documenting no-op (pure-Go detector, nothing to probe; declared so absence isn't mistaken for unverified)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Honest spec surface                                                                                                 |
| 8  | `Inputs` declaration (read contract)                                                                                     | Fully Leveraged                                                                     | pkg/provider/provider.go:100 (`**/*.go`, `**/*.templ`) — gates the tool on file presence and feeds host data-flow edges                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | Host ordering correctness                                                                                           |
| 9  | `Spec.Outputs` write contract (v1.16.0)                                                                                  | Not Applicable                                                                      | Read-only detector; empty `Outputs` IS the documented correct declaration for detect-only tools ("empty = the tool writes nothing hosts must order around", go-finding/toolsdk/spec.go:71–77)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |                                                                                                                     |
| 10 | `DependsOn`                                                                                                              | Not Applicable                                                                      | Single self-contained detector; no intra-run tool ordering to declare beyond what Inputs-derived edges already give the host                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |                                                                                                                     |
| 11 | `ModuleFanOut`                                                                                                           | Not Applicable                                                                      | Cross-module clone detection is art-dupl's value — a clone duplicated between workspace modules is exactly the finding BuildFlow wants; per-module fan-out would blind it (§05.2 verdict, re-confirmed §06.2)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |                                                                                                                     |
| 12 | `SwitchCases`                                                                                                            | Not Applicable                                                                      | Conditional repair-path topology keyed on diagnose rule IDs; art-dupl declares no Repairer and has a single rule                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |                                                                                                                     |
| 13 | `Repairer`/`RepairResult`/`Fact` incl. v1.16.0 `ModifiedFiles`/`Facts` provenance (issue #42)                            | Not Applicable                                                                      | Detector-only tool (ADR-0025 classification freedom; §05.2 Edits verdict — deleting a clone needs a human decision about which copy survives)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |                                                                                                                     |
| 14 | Dry-run channel (`WithDryRun`/`DryRunFromContext`)                                                                       | Not Applicable                                                                      | Gates repair semantics (hosts dry-run Repair, not Detect); art-dupl has no repair path to dry-run                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |                                                                                                                     |
| 15 | Test registry helpers (`SnapshotForTest`/`RestoreForTest`/`ResetForTest`)                                                | Not Applicable                                                                      | Registration is a single package-level `var Provider`; tests never mutate the global registry, so there is no isolation problem to solve (§06.2 verdict)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |                                                                                                                     |

## Gap analysis

### G1. Typed option accessors unused — hand-rolled equivalents at the same pin (row 5)

- **Library offers** (toolsdk v1.16.0 — **the version art-dupl already pins**):
  `IntFromContext(ctx, name, fallback)` and `BoolFromContext(ctx, name, fallback)`
  (go-finding/toolsdk/options.go:149, 156): any miss (unset, no values, nil context, nil
  value) yields the fallback; a present value of the wrong Go type yields an error matching
  `ErrOptionKindMismatch`. Shipped per upstream changelog issue #43 precisely so "tools stop
  hand-rolling the lookup/assert/fallback plumbing."
- **art-dupl does instead**: two local helpers, `thresholdFromContext`
  (pkg/provider/provider.go:201–219: `OptionsFromContext` → map lookup → `.(int)` assert → `(int, bool)`)
  and `boolOptionFromContext` (pkg/provider/provider.go:221–241: same shape for bool).
- **Impact**: functional duplication with a semantic divergence: the local helpers silently
  treat a wrong-kind value as unset (documented as "only reachable by callers that bypassed
  ValidateOptions"), while the library accessors surface `ErrOptionKindMismatch`. Maintenance
  cost is small but the pattern is exactly what the fleet centralizes — and the replacement is
  already in the module cache.
- **Fix** (drop-in; behavior for validated callers identical — fallback
  `artdupl.DefaultThreshold` equals the existing "unset → don't override `DefaultOptions()`"
  path, and `false` equals the existing bool default):

```go
// pkg/provider/provider.go — replaces both helpers
threshold, err := toolsdk.IntFromContext(ctx, providerOptionThreshold, artdupl.DefaultThreshold)
if err != nil {
    return nil, fmt.Errorf("read threshold option: %w", err) // ErrOptionKindMismatch path
}

if emit, err := toolsdk.BoolFromContext(ctx, providerOptionEmitSuppressedAccepted, false); err == nil && emit {
    acceptedSet = accept.NewAcceptedSet(os.ReadFile)
}
```

`TestDetectThresholdOption` and `TestEmitSuppressedAcceptedOffIsByteIdentical` already pin
the behavior and should pass unchanged (add a wrong-kind case asserting the new error path).

### G2. Everything else on the v1.16.0 menu is verdict-ed N/A (rows 9–15)

`Outputs`, `DependsOn`, `ModuleFanOut`, `SwitchCases`, `Repairer`/`RepairResult` provenance,
dry-run, and the test-registry helpers were each checked against this provider and either
don't apply to a read-only single-rule detector or carry a standing §05.2/§06.2 verdict.
No action; the N/A reasons above are the audit's justification, one line each as required.

## Adoption score (derived)

Rubric: rows N, minus N/A rows, = applicable A; Fully = 1, Partially = 0.5, Missed/Misused = 0;
score = round(100 × Σ / A).

- N = 15 inventory rows; N/A = 7 (rows 9–15); **A = 8 applicable**
- Fully Leveraged = 7 · Missed = 1 (× 0)
- Σ = 7 → score = round(100 × 7 / 8) = round(87.5) = **88 / 100**
- Inputs to publish: **7 of 8 applicable leveraged (7 Fully, 0 Partial, 1 Missed)**
- Band: **very high (≥85)** — the single Missed row is a ~40-line plumbing swap inside one file,
  available at the current pin.

## Version currency

| Module  | art-dupl pin | Latest tag                     | Gap         | Delta content |
| ------- | ------------ | ------------------------------ | ----------- | ------------- |
| toolsdk | `v1.16.0`    | `toolsdk/v1.16.0` (2026-10-11) | **current** | —             |

toolsdk v1.16.0 content (per go-finding/toolsdk/CHANGELOG.md [1.16.0], re-verified in the tag):
Trigger validation at `Register` (issue #44) — passively protecting art-dupl's registration;
typed option accessors (issue #43) — **the one v1.16.0 feature art-dupl fails to use** (G1);
`RepairResult` provenance `ModifiedFiles`/`Facts` (issue #42) — N/A, no repairer;
`Spec.Outputs` write contract — N/A, read-only. toolsdk's own [Unreleased] is empty.

**Recommendation**: no version change needed on this module (current). The mixed-pin issue
lives on the core side (core `v1.15.0` vs latest `v1.16.0`) — see
`docs/review/libs/go-finding.md` § Version currency for the core bump procedure
(AGENTS.md:241: `go get`, nix `vendorHash` refresh, `go test -count=1 ./...`, provider golden
regen only if positions/ids changed).

## Fleet argument

Lars has many projects, and toolsdk is the fleet's single declaration point for what a tool
is, when it fires, and what it reads. For art-dupl the argument concentrates on **one spot**:
option plumbing. Every per-repo hand-rolled `OptionsFromContext` + type-assert + fallback
helper is a small drift factory — wrong-kind semantics diverge repo by repo (art-dupl's
silent-unset vs the library's `ErrOptionKindMismatch` is a live example). The typed accessors
were built for the fleet and adopted here costs ~10 lines; declining them means the next
option added to this provider will grow a third copy of the pattern. Conversely, the fleet
argument does NOT argue for adopting the repair-topology surface (SwitchCases, Outputs,
RepairResult, dry-run): art-dupl's detector-only stance is the correct specialization, and
forcing those fields in would misdeclare the tool to every host. The provider's restraint —
one spec, two options, valid trigger, explicit no-op HealthCheck — is what a good fleet
citizen looks like; only the accessor swap remains.

## Ranked opportunities (impact × ease)

| # | Opportunity                                                                                                                      | Impact (1–5) | Ease (1–5) | Priority | Gate                           |
| - | -------------------------------------------------------------------------------------------------------------------------------- | ------------ | ---------- | -------- | ------------------------------ |
| 1 | Replace `thresholdFromContext`/`boolOptionFromContext` with `IntFromContext`/`BoolFromContext`; add a wrong-kind error-path test | 2            | 5          | 10       | none — API already at pin      |
| 2 | (Adjacent, core-side) implement `DetectorV2` coverage in the same provider — see core report G1                                  | 5            | 5          | 25       | none — core API already at pin |
| 3 | (Adjacent, core-side) bump core to v1.16.0 for `Relativize` + run propertyBag                                                    | 4            | 5          | 20       | nix vendorHash refresh         |

## RECONSIDER-WHEN

Re-open the N/A verdicts when: art-dupl grows a repair path (→ `Repairer`, `RepairResult`
provenance, `SwitchCases`, `Outputs`, dry-run all become live); BuildFlow ever fans art-dupl
out per workspace module (→ `ModuleFanOut` — though the cross-module-clone verdict argues
never); a fourth option kind is added upstream (→ revisit whether a generic accessor shape
would serve); or tests start mutating the global registry (→ `SnapshotForTest`). Re-check
toolsdk currency on every go-finding train — modules tag independently (AGENTS.md:241).

## Appendix: how to verify

Commands run for this audit (replay literally; paths absolute):

```bash
# Checkouts & tags
git -C /home/lars/projects/art-dupl rev-parse HEAD            # cee67b4a47a8e759f5af05669cad97dca8d1f22d
git -C /home/lars/projects/go-finding rev-parse HEAD          # 5f568e058a15221ac198d378a1d907231451edf0
git -C /home/lars/projects/go-finding rev-list -n1 toolsdk/v1.16.0   # 9c675c77bc37a1c8fdd4adc2172ba3af4f712686
git -C /home/lars/projects/go-finding log -1 --format=%ci toolsdk/v1.16.0   # 2026-10-11

# toolsdk v1.16.0 surface (tag + HEAD)
git -C /home/lars/projects/go-finding show toolsdk/v1.16.0:toolsdk/options.go | grep -n "func StringFromContext\|func IntFromContext\|func BoolFromContext"  # :142/:149/:156
rg -n "^func (Register|All|EnsureContext)|^type (Spec|Trigger|SwitchCase|Repairer|RepairResult|Fact) " /home/lars/projects/go-finding/toolsdk/*.go
rg -n "Outputs \[\]string" /home/lars/projects/go-finding/toolsdk/spec.go          # :71
git -C /home/lars/projects/go-finding show toolsdk/v1.16.0:toolsdk/go.mod          # requires core v1.14.0 (published)
head -60 /home/lars/projects/go-finding/toolsdk/CHANGELOG.md                       # [1.16.0] - 2026-10-11 content

# Consumer pin & usage
cd /home/lars/projects/art-dupl && go list -m all | grep go-finding   # core v1.15.0, toolsdk v1.16.0
grep -n "toolsdk" /home/lars/projects/art-dupl/go.mod                  # :91 v1.16.0
rg -n "toolsdk\." /home/lars/projects/art-dupl/pkg/provider/provider.go
rg -n "func thresholdFromContext|func boolOptionFromContext|func \(cloneDetector\) Detect" \
   /home/lars/projects/art-dupl/pkg/provider/provider.go               # :201, :221, :135
rg -n "IntFromContext|BoolFromContext|StringFromContext" /home/lars/projects/art-dupl --glob '!docs/**'   # no hits (the gap)
rg -n "func Test" /home/lars/projects/art-dupl/pkg/provider/provider_test.go | head -12

# Prior-report context
sed -n '960,1804p' /home/lars/projects/art-dupl/docs/research/2026-10-05_go-finding-deep-dive.html | sed 's/<[^>]*>//g' | grep -v '^\s*$'
grep -n "go-finding upgrade policy" /home/lars/projects/art-dupl/AGENTS.md         # :241
```

Line numbers stamped in this report were read at the cited commits and are reproducible with
`sed -n '<start>,<end>p'` / `rg -n` on the same checkouts.
