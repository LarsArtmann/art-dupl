# Library utilization audit — go-finding (core module)

| Field | Value |
| --- | --- |
| Library | `github.com/larsartmann/go-finding` (+ workspace sub-modules) |
| Module paths covered here | core `github.com/larsartmann/go-finding` (primary); `analysis/`, `pipeline/`, `cmd/go-finding/` as adoption candidates (toolsdk has its own report: `docs/review/libs/go-finding-toolsdk.md`) |
| Library checkout reviewed | `/home/lars/projects/go-finding` @ `5f568e058a15221ac198d378a1d907231451edf0` == root tag `v1.16.0` exactly (annotated, SSH-signed, tagged 2026-11-22) |
| Consumer checkout reviewed | `/home/lars/projects/art-dupl` @ `cee67b4a47a8e759f5af05669cad97dca8d1f22d` (fork, clean) |
| Consumer pins | core `v1.15.0` (go.mod:90), toolsdk `v1.16.0` (go.mod:91) — mixed: core lags, toolsdk current |
| Date | 2026-11-22 |
| Auditor | group A fleet run (read-only audit; this file is the only write) |

## Prior coverage (superseded by this report)

- `docs/research/2026-10-05_go-finding-deep-dive.html` — §01 findings, §04 capability appendix
  (pre-remediation snapshot), §05 remediation ledger (2026-10-06 re-review, incl. §05.2 verdicts
  and §05.5 re-scored table), §06 currency re-review (2026-10-11, incl. §06.4 wire correction).
  **This report supersedes §04/§05.5/§06 as the current adoption truth** for the core module.
  The deep-dive remains canonical for *why* each deliberate decline was verdict-ed; per
  art-dupl AGENTS.md (upgrade-policy note, line 241) those verdicts were re-read before judging
  and are not re-litigated here without new evidence.
- Related art-dupl TODO rows audited: **G8** (pipeline release) and **T07b** (coverage-evidence
  channel) — verdicts in [Gap analysis](#gap-analysis) below.
- Prior integration evaluation: `docs/research/2026-06-05_go-finding-integration-evaluation-summary.md`
  (go-finding stays a consumer-side output layer, never an internal representation — still honored).

## Module map

go-finding is a five-module Go workspace. Tags are per-module and independent (AGENTS.md:241
warns about exactly this); latest local tags, semver-sorted, re-verified for this audit:

| Module | Import path | Latest tag | Tag date | art-dupl consumes? |
| --- | --- | --- | --- | --- |
| core (root) | `github.com/larsartmann/go-finding` | `v1.16.0` (== HEAD) | 2026-11-22 | **Yes** — pinned `v1.15.0` |
| toolsdk | `github.com/larsartmann/go-finding/toolsdk` | `toolsdk/v1.16.0` (tag at `9c675c7`, ≠ HEAD) | 2026-10-11 | **Yes** — pinned `v1.16.0` (current) |
| analysis | `github.com/larsartmann/go-finding/analysis` | `analysis/v1.15.1` | 2026-11-08 | No |
| pipeline | `github.com/larsartmann/go-finding/pipeline` | `pipeline/v1.15.1` | 2026-11-08 | No |
| cmd/go-finding | `github.com/larsartmann/go-finding/cmd/go-finding` | `cmd/go-finding/v1.15.1` | 2026-11-08 | No |

art-dupl is a **single Go module** (`github.com/LarsArtmann/art-dupl`, root go.mod only, no
vendor dir) and consumes **core + toolsdk only** — confirmed via `go list -m all | grep go-finding`.

## Current usage map

All citations are `path:line` in the art-dupl checkout unless prefixed `go-finding/` (library
checkout at HEAD = v1.16.0).

| Consumer site | What it uses | Library symbols |
| --- | --- | --- |
| `printer/finding/finding.go` (adapter, LIBRARY-ONLY per ADR-0028) | Full finding construction | `NewTemplate`, `Builder`, `MustBuild`, `WithGroupID/Tags/Confidence/Range/Snippet/Metadata/Related/Suppression/FixStrategy/Suggestion`, `GenerateID`, `GroupID`, `Tag`, `Severity`, `CategoryDuplication`, `Confidence().Clamp()`, `Suppression{SuppressionInSource}`, `FixStrategySuggest`, `RelationCloneOf`, `RelatedRef`, `Position`, `Range`, `FilePath`, `ToolName`, `RuleName`, `NewReportFromFindings`, `ToolInfo` (finding.go:92, 100, 125, 131, 139, 152–199, 204–257, 259–312) |
| `printer/sarif.go` (CLI product wire, hand-rolled) | Severity/level ladder source, group-id property key, relation name | `gofinding.RelationCloneOf` (printer/sarif.go:231), `finding.SARIFPropGroupID` (printer/sarif.go:241) |
| `printer/lsp.go` (CLI `--lsp` format) | Per-finding LSP conversion | `gofinding.LSPDiagnostic`, `Finding.ToLSP()` (lsp.go:22, 66–69); envelope struct `LSPDocument` is local (lsp.go:20–23) |
| `pkg/provider/provider.go` (BuildFlow toolsdk provider) | Detector contract, context, severity cap, tags | `Detector`, `WorkingDirFromContext`, `SeverityWarning/SeverityError/SeverityCritical`, `Tag` (pkg/provider/provider.go:59, 131–181, 274–289) |
| `bdd/sarif_gofinding_integration_test.go` | SARIF import round-trip of real CLI bytes | `FindingsFromSARIF`, `NewReportFromFindings`, `ToolInfo` (lines 95, 99, 129, 133, 184, 188) |
| `bdd/lsp_output_test.go` | LSP import round-trip | `FromLSP`, `FilePath` (lines 51–54) |
| Contract tests | Interchange guarantees | `printer/finding/finding_test.go` (GroupID determinism, SARIF/LSP/JSON round-trips, `Report.GroupFindings()` reconstruction), `printer/sarif_finding_crosscheck_test.go:22` `TestSARIFCrossCheckCLIvsGoFinding` (CLI printer vs `Report.ToSARIF()` parity, plan T09) |

Mechanical footprint: ~72 go-finding identifier references across
`printer/finding/finding.go` (45), `pkg/provider/provider.go` (23), `printer/lsp.go` (3),
`printer/sarif.go` (1).

## Capability inventory (core)

Status rubric: **Fully Leveraged / Partially Used / Missed Opportunity / Misused / Not Applicable**
(every N/A carries a one-line reason). 26 rows.

| # | Capability | Status | Evidence | Impact |
| --- | --- | --- | --- | --- |
| 1 | Finding value model + branded types (`ID`, `RuleName`, `ToolName`, `FilePath`, `GroupID`, `Tag`) | Fully Leveraged | printer/finding/finding.go:92, 139–149, 238–257 | Foundation of both output lanes; compile-time safety at no runtime cost |
| 2 | Builder + Template validated construction | Fully Leveraged | printer/finding/finding.go:101–102, 168–199 (`MustBuild`); pinned by `TestBuilderConversionIsByteIdenticalToLegacyConstruction`, `TestBuilderRejectsMalformedConstruction` | Malformed findings cannot enter the wire (plan T04, 2026-10-06) |
| 3 | GroupID grouping key (JSON/SARIF/LSP round-trip) | Fully Leveraged | printer/finding/finding.go:92–94; printer/sarif.go:241 (`go-finding/groupId` property); `TestSARIFRoundTripPreservesGroupID`, bdd:95–100 | Consumers diff reports without re-deriving grouping |
| 4 | Related links (`RelationCloneOf`) | Fully Leveraged | printer/finding/finding.go:288–312 (precomputed IDs, no quadratic regen); printer/sarif.go:229–232, 260–267 | Sibling topology on every wire |
| 5 | Suppression model (`Suppression{InSource}`) | Fully Leveraged | printer/finding/finding.go:187–193; provider option `emit-suppressed-accepted` (pkg/provider/provider.go:108–116, 158–160); SARIF round-trip via `WithIncludeSuppressed` pinned | `//art-dupl:accept` is visible on the wire, not an absence (plan T06) |
| 6 | Severity ladder + SARIF level agreement | Fully Leveraged | `severityFor` printer/finding/finding.go:204–213 vs `determineLevel` printer/sarif.go:413–422; `TestToFindingsSeverityMirrorsSARIFLadder`; `TestSARIFCrossCheckCLIvsGoFinding` (cross-check contract, plan T09) | Both lanes agree per group; desync fails CI |
| 7 | Tags + category-in-tags consistency | Fully Leveraged | `cloneTags` printer/finding/finding.go:221–228; `TestToFindingsTagsCarryCategoryAndCloneType` (plan T08) | Satisfies `Finding.Validate`'s tag rule trivially |
| 8 | Confidence + `Clamp()` | Fully Leveraged | printer/finding/finding.go:230–236 | Analysis confidence normalized into the model |
| 9 | Position/Range (lines, columns, offsets) | Fully Leveraged | `positionOf`/`rangeOf` printer/finding/finding.go:238–257 (columns since 2026-10-06, §05.1 T13) | Full fidelity regions in SARIF/LSP |
| 10 | FixStrategy (suggest) | Fully Leveraged | printer/finding/finding.go:182–185 (`Suggestion` → `FixStrategySuggest`); `TestToFindingsSuggestionSetsFixStrategy` | Classification suggestions survive interchange |
| 11 | `Finding.Metadata` interchange | Fully Leveraged | `metadataFor` printer/finding/finding.go:259–286; `art-dupl/*` namespace, reserved `go-finding/` respected | Clone-specific data travels without model surgery |
| 12 | Detector contract (v1) + WorkingDir context | Fully Leveraged | pkg/provider/provider.go:135–141 (`Detect`, `WorkingDirFromContext`) | Provider integrates with BuildFlow lanes |
| 13 | **Coverage-evidence channel** (`DetectResult`, `DetectorV2`, `RunDetect`, `Report.AddDetectResult` → `Summary.FilesScanned`/`SkippedModules`) | **Missed Opportunity** | pkg/provider/provider.go:135 implements v1 `Detect` only; zero production references to `DetectV2`/`RunDetect`/`FilesScanned`/`SkippedModules` in art-dupl (rg-verified). API **shipped at the current pin**: core v1.15.0 carries `detect_result.go` (go-finding/detect_result.go:15, 39, 53, 72) and `Summary.FilesScanned`/`SkippedModules` (go-finding/report.go:71, 78) | "0 findings" vs "not analyzed" unanswerable from the provider lane; TODO T07b consumption step is unblocked TODAY |
| 14 | SARIF export (`ToSARIF`, `ToSARIFWithOpts`, `WithIncludeSuppressed`) | Fully Leveraged | Hand-rolled printer (printer/sarif.go:27–161, 269–286) is a deliberate superset (fingerprints, invocations, rule descriptors — §05.3); agreement with `ToSARIF()` is a pinned contract (`TestSARIFCrossCheckCLIvsGoFinding`, `TestSARIFCrossCheckRelatedLocations`) | Best of both: product wire + library parity guarantee |
| 15 | SARIF import (`FindingsFromSARIF`, `FindingsFromReader`) | Fully Leveraged | bdd/sarif_gofinding_integration_test.go:95, 129, 184 (real CLI `--sarif` bytes reconstruct the exact group partition) | The consumer path is proven, not assumed |
| 16 | **Run-level SARIF propertyBag** (`SetRunProperty`, `RunProperties`, `RunPropertiesFromSARIF`) — v1.16.0 | **Missed Opportunity** | Not at pin (core v1.15.0). Exists at v1.16.0: go-finding/report.go:140, :155; go-finding/sarif_import.go:29; fetchability proven below in Version currency | Per-run facts (threshold posture, semantic mode, tool version) still copied per-result or absent; `runs[].properties` is the shipped home for them |
| 17 | LSP diagnostics (`ToLSP`/`FromLSP`) | Fully Leveraged | printer/lsp.go:66–69 (`f.ToLSP()`), bdd/lsp_output_test.go:51–54 (`FromLSP`); `TestToLSPRoundTripPreservesGroupID`. Local `LSPDocument` envelope (lsp.go:20–23) is a justified shim — core ships no Report-level LSP export | Editor consumers get diagnostics directly (plan T05) |
| 18 | **`Relativize`** (root-relative, checkout-independent IDs) — v1.15.1 | **Missed Opportunity** | Not at pin. Exists at v1.15.1+/v1.16.0: go-finding/relativize.go:30. art-dupl's provider lane derives IDs from `cl.Filename` as handed in (printer/finding/finding.go:139–150, 238–244) | SDK/provider lane run under BuildFlow's absolute working dirs produce IDs that embed host paths; CLI (relative) and provider (absolute) lanes disagree on IDs for the same source fact — exactly what `Relativize` was released to fix (upstream issue #47) |
| 19 | JSON wire (`Report.JSON`/`MarshalJSON`, `FindingsFromJSON`) | Partially Used | Round-trip pinned (`TestFindingJSONCarriesGroupID`); no production emitter — ADR-0028 verdict: consumers interchange via SARIF, no `--output finding` CLI format | Low: second wire would split the contract; current usage is contract-pinning only |
| 20 | Report query/filter/group API (`GroupFindings`, `BySeverity`, `Filter`, `GroupBy`, `SortBy*`) | Not Applicable | CLI/SDK stream groups for constant memory; Report queries are consumer-side — §05.2 "keep streaming" verdict, re-confirmed §06.2 |
| 21 | Merge / Diff / Correlate report algebra | Not Applicable | `baseline/` is a human-curated group-granularity gate, verdict-ed a "justified parallel" (§05.2 retraction); revisit only if a machine consumer asks for a finding-level diff |
| 22 | `DetectorRegistry` (named-constructor catalog) | Not Applicable | Serves config-file-driven CLI resolution (cmd/go-finding); art-dupl registers via toolsdk's process-global `Register` — separate mechanism by design |
| 23 | `ToolAdapter[O]` (external JSON-tool wrapper) | Not Applicable | art-dupl is a native Go detector; there is no external binary whose JSON output needs adapting |
| 24 | `TextEdit`/`Finding.Edits`/`ApplySimpleFixes` | Not Applicable | Detector-only tool; which copy of a clone survives is a human decision (§05.2 verdict; ADR-0025 classification freedom) |
| 25 | Severity aliases (`RegisterSeverityAlias`, `SeverityFromLevel`) | Not Applicable | They parse external level strings; art-dupl needs the inverse direction (Severity → SARIF level string, unexported upstream `severityToSARIFLevel`) — the dual ladder is instead pinned by the T09 cross-check |
| 26 | `gotoken` position helpers (`LineColToOffset`, `NodeByteRange`) | Not Applicable | go/token `FileSet`-based interop; art-dupl's `pkg/position` (pkg/position/lines.go:12, 64) works on raw content bytes and is language-agnostic (Go + templ — templ files have no `token.File`) |

### Adoption candidates not consumed (module-level)

- **`analysis/` (v1.15.1)** — go/analysis interop: `FromDiagnostic(WithSource)`, `ToDiagnostic`,
  `AnalyzerDetector` (runs any analyzer as a `finding.Detector`), `NodePosition`/`NodeRange`
  (go-finding/analysis/analysis.go:37, 184, :141/:146; adapter.go:16, 39). **Not Applicable**:
  art-dupl is not a go/analysis analyzer (suffix tree over its own unified AST); nothing here
  replaces local code. The v1.14.0 multi-edit losslessness fix (issue #36) is irrelevant for the
  same reason.
- **`pipeline/` (v1.15.1)** — detect→triage→apply→verify orchestration: `FixEngine` with provider
  chain (byte-offset → line/col → substring, `EditListProvider` first — go-finding/pipeline/fix_provider.go:61),
  conflict detection, file backups, flight recorder, config-file resolution; since v1.15.0 every
  detect path dispatches through core's `RunDetect` so `DetectorV2` coverage flows end-to-end.
  **Not Applicable**: BuildFlow owns orchestration for art-dupl's findings (§05.2/§06.2 verdict);
  art-dupl produces no fixes for the fix engine to apply.
- **`cmd/go-finding` (v1.15.1)** — the CLI resolves detectors through `DetectorRegistry` +
  config file, not toolsdk discovery, so a self-registering toolsdk provider (art-dupl's mode)
  is invisible to it (§05.2 verdict, unchanged at these tags). **Not Applicable** as an adoption
  target for art-dupl; noted as an upstream visibility gap, not a consumer gap.

## Gap analysis

Each Partially/Missed/Misused row: what the library offers, what art-dupl does instead, impact,
concrete fix.

### G1. Coverage-evidence channel unused though pinned (row 13 — the T07b finding)

- **Library offers**: `DetectResult{Findings, FilesScanned, SkippedModules}` behind optional
  `DetectorV2` (go-finding/detect_result.go:15, 39); host-side `RunDetect(ctx, d)` (v2 when
  implemented, v1 fallback with coverage zeroed); `Report.AddDetectResult` accumulates into
  `Summary.FilesScanned` and unions `SkippedModules` (go-finding/detect_result.go:53–85).
  Released in **core v1.15.0** — commit `a275b57` "feat: DetectResult + DetectorV2
  coverage-evidence channel (T07b)", which IS in the v1.15.0 tag (verified:
  `git show v1.15.0:detect_result.go` exists). Runnable examples exist at HEAD
  (`go-finding/examples/detectorv2/main.go`, `go-finding/pipeline/examples/detectorv2/`).
- **art-dupl does instead**: `cloneDetector` implements only v1 `Detect` returning
  `[]gofinding.Finding` (pkg/provider/provider.go:135–181). Its crawl already knows the file count
  (`collectSourceFiles`, pkg/provider/crawl.go:32) and discards it.
- **Impact**: BuildFlow reports built from this provider render `filesScanned: 0` —
  indistinguishable from a real zero. The TODO T07b row's premise ("upstream-gated … next
  concrete step is the upstream prototype session") is **stale on the upstream half**: the
  channel shipped 2026-10-12 at art-dupl's own pin. The remaining work is purely art-dupl-side.
- **Fix** (drop-in, no wire change):

```go
// pkg/provider/provider.go — add alongside Detect
func (d cloneDetector) DetectV2(ctx context.Context) (gofinding.DetectResult, error) {
    // ... existing Detect body, but return:
    return gofinding.DetectResult{
        Findings:      findings,
        FilesScanned:  len(files), // the crawl already collected them
        // SkippedModules: leave empty until the crawl starts skipping nested go.mod trees
    }, nil
}
```

  (Hosts that only know `Detector` keep working — source-compatible per upstream design.)

### G2. Run-level SARIF propertyBag missed (row 16 — the v1.15.0→v1.16.0 delta)

- **Library offers**: `Report.SetRunProperty(key, value)` records per-run metadata ONCE into
  `runs[].properties` (SARIF 2.1.0 §3.10.1); `RunPropertiesFromSARIF` reads the bag back
  (go-finding/report.go:135–160, go-finding/sarif_import.go:29). `omitempty` when unset —
  zero output diff for producers that never call the setter.
- **art-dupl does instead**: nothing at the pin (impossible — v1.15.0); per-run facts either
  absent or implicit. The CLI SARIF wire has no run-level properties field at all
  (printer/sarif.go:34–38 `SARIFRun`), and `ToReport` (printer/finding/finding.go:125–135)
  cannot record posture.
- **Impact**: threshold posture / semantic-mode / version are the exact "enforcement posture,
  scan root, tool version tag" examples the feature cites (upstream issue #53). Consumers
  auditing whether a SARIF was produced at threshold 1 vs 15 have to guess.
- **Fix**: bump core to v1.16.0, then in the adapter lane:

```go
report := finding.ToReport(groups, opts)
report.SetRunProperty("art-dupl/threshold", threshold)
report.SetRunProperty("art-dupl/mode", "semantic")
```

  CLI-wire adoption (adding `Properties` to `SARIFRun`) is a separate, deliberate wire decision
  — AGENTS.md:323 requires regenerating `printer/testdata/TestSARIFOutputGolden.golden` only
  for INTENTIONAL wire changes. The SDK lane adoption changes no art-dupl wire bytes.

### G3. `Relativize` missed (row 18)

- **Library offers**: `finding.Relativize(findings, root)` rewrites absolute `Position.File`
  under root to root-relative slash paths and regenerates IDs — "the CLI lane (relative target)
  and the SDK/orchestrator lane (absolute working directory) now produce the same ID for the
  same source fact" (go-finding/relativize.go:10–28, released v1.15.1, upstream issue #47).
- **art-dupl does instead**: nothing; `findingIDs` hashes positions as handed in
  (printer/finding/finding.go:139–150), and the provider's `dir` comes from
  `WorkingDirFromContext` (pkg/provider/provider.go:138–141), which BuildFlow supplies as absolute paths.
- **Impact**: provider-lane finding IDs embed host paths (`/home/...`) and differ from CLI-lane
  IDs for the identical source fact — breaks cross-lane diffing/baselining and leaks host paths.
- **Fix**: after the bump, at the end of `Detect`: `return gofinding.Relativize(findings, dir), nil`.

### G4. JSON wire partially used (row 19)

- **Library offers**: stable `Report`/`Finding` JSON wire (`FindingsFromJSON`, `ReportFromJSON`,
  go-finding/json.go:139–190).
- **art-dupl does instead**: uses it only as a pinned round-trip contract
  (`TestFindingJSONCarriesGroupID`); production consumers are pointed at SARIF (ADR-0028).
- **Impact**: minimal — this is a verdict-ed product decision, kept as Partially (not N/A)
  only because the wire is genuinely exercised. No action recommended beyond the standing
  ADR-0028 review.

### Stale/incorrect TODO rows (audit findings, no code change)

- **G8 is STALE**: "release the pipeline Unreleased section (EditListProvider et al.) as
  `pipeline/v1.14.0`" — `pipeline/v1.14.0` was tagged **2026-10-11** with exactly that content
  (`EditListProvider` go-finding/pipeline/fix_provider.go:61, `ResolveFlightRecorderConfig`
  go-finding/pipeline/config_file.go:162), followed by `pipeline/v1.15.0` (2026-10-12,
  coverage channel) and `pipeline/v1.15.1` (2026-11-08). pipeline's current [Unreleased]
  reads "Nothing yet." The row should be closed.
- **T07b is half-stale**: upstream half DONE (core v1.15.0 shipped `DetectResult`/`DetectorV2`;
  pipeline v1.15.0 dispatches `RunDetect` end-to-end); remaining half is the art-dupl-side
  consumption in G1 above. The row's "Blocked on: nothing art-dupl-side" conclusion now inverts:
  everything remaining IS art-dupl-side.

## Adoption score (derived)

Rubric: rows N, minus N/A rows, = applicable A; Fully = 1, Partially = 0.5, Missed/Misused = 0;
score = round(100 × Σ / A).

- N = 26 inventory rows; N/A = 7 (rows 20–26); **A = 19 applicable**
- Fully Leveraged = 15 · Partially = 1 (× 0.5) · Missed = 3 (× 0)
- Σ = 15.5 → score = round(100 × 15.5 / 19) = round(81.6) = **82 / 100**
- Inputs to publish: **15.5 of 19 applicable leveraged (15 Fully, 1 Partial, 3 Missed)**
- Band: **high (70–84)** — the three Missed rows are all recent-release deltas/one unblocked TODO,
  not architectural rejection.

## Version currency

| Module | art-dupl pin | Latest tag | Gap | Delta content |
| --- | --- | --- | --- | --- |
| core | `v1.15.0` | `v1.16.0` (2026-11-22) | **1 minor + 1 patch behind** | v1.15.1 (2026-11-04): `finding.Relativize` (issue #47), dep refresh. v1.16.0 (2026-11-22): run-level SARIF propertyBag (`SetRunProperty`/`RunProperties`/`RunPropertiesFromSARIF`, issue #53), CI lint stack, release-workflow nixpkgs-404 fix, sub-module go.sum completion |
| toolsdk | `v1.16.0` | `toolsdk/v1.16.0` (2026-10-11) | current | — (see toolsdk report) |

**Mixed-pin finding**: core lags while toolsdk is current. History explains it: toolsdk v1.16.0
(2026-10-11) predates core v1.15.0 (2026-10-12), art-dupl took the toolsdk train immediately and
the core T07a bump separately; core v1.16.0 landed only 2026-11-22. The published
`toolsdk/v1.16.0` go.mod requires only core v1.14.0, so MVS keeps art-dupl's explicit v1.15.0 —
the mixed state is stable but strands two releases of core features.

**v1.16.0 tag-content verification (stamp-defect audit)**: the root tag `v1.16.0` points at HEAD
`5f568e0` ("feat: run-level SARIF propertyBag"), and `git show v1.16.0 --stat` shows exactly the
propertyBag diff (`report.go`, `sarif_runproperties_test.go`, `version.go`, `CHANGELOG.md`,
`.config/metadata.yaml`). `version.go` at the tag correctly stamps `VersionMinor = 16`. The API
**is fetchable**: `git show v1.16.0:report.go` contains `func (r *Report) SetRunProperty` (line 140)
and `git show v1.16.0:sarif_import.go` contains `RunPropertiesFromSARIF` (line 29) — so
`go get github.com/larsartmann/go-finding@v1.16.0` delivers the feature. The defect is
**documentation-only**: the core CHANGELOG still heads that content `[Unreleased]` (no `[1.16.0]`
section; `grep -n "1.16" CHANGELOG.md` → no match) — the recurring stamp lesson, now gated
upstream by the pre-tag stamp and consumer-proof gates added in the v1.15.1..v1.16.0 range
(`c753540`, `6ccdc23`). The `[Unreleased]` "Fixed" entries (release workflow, go.sum completion)
are also part of the tagged v1.16.0.

**Recommendation**: bump core `v1.15.0 → v1.16.0`. Procedure per AGENTS.md:241 — `go get`, refresh
nix `vendorHash`, `go test -count=1 ./...`, regenerate the provider golden
(`go test ./pkg/provider/ -run TestProviderFindingsGolden -update`) only if positions/ids changed
(neither `Relativize` nor `SetRunProperty` alters any wire byte unless deliberately called, so no
golden churn is expected). Then adopt G1 (works at either pin), G2/G3 (require the bump). Keep
toolsdk at v1.16.0 (current).

## Fleet argument

Lars runs many projects through one findings interchange (go-finding + BuildFlow, ~25 internal
consumers). Centralizing capability in go-finding beats per-repo reimplementation wherever the
capability is *interchange-shaped* — and for art-dupl that argues hardest on exactly the three
Missed rows:

1. **Coverage evidence (G1)** is only valuable if EVERY fleet detector emits it the same way;
   a per-repo "I'll emit my own filesScanned metadata key" answer fragments the signal
   BuildFlow needs to render once. art-dupl crawling without reporting is free-riding on a
   channel its own TODO designed (T07b's recipe names the exact shape that shipped).
2. **Run-level propertyBag (G2)** exists because the fleet asked (issue #53): posture belongs on
   the run, once. If art-dupl re-encodes threshold per result property instead, every consumer
   (including future art-dupl dashboards) re-derives the de-duplication.
3. **`Relativize` (G3)** is a fleet-wide lane-stability primitive (CLI vs orchestrator IDs);
   go-finding fixed it centrally precisely so each detector would not ship its own path
   normalization.

Where the argument does NOT apply, art-dupl is right to keep local code: the CLI text/HTML/JSON
product formats, the streaming architecture (constant memory), the SARIF superset wire
(fingerprints, rule descriptors), and the actionability engine are *product*, not interchange —
the deep-dive verdicts there remain sound and are honored by this audit's N/A rows.

## Ranked opportunities (impact × ease)

| # | Opportunity | Impact (1–5) | Ease (1–5) | Priority | Gate |
| --- | --- | --- | --- | --- | --- |
| 1 | Implement `DetectorV2` in the provider; populate `FilesScanned` from the crawl (G1; closes T07b consumption) | 5 | 5 | 25 | none — API already at pin |
| 2 | Bump core `v1.15.0 → v1.16.0` (unblocks G2+G3, picks up release-infra fixes) | 4 | 5 | 20 | nix vendorHash refresh |
| 3 | Adopt `SetRunProperty` in the `ToReport` lane (threshold, mode) (G2) | 3 | 4 | 12 | after #2 |
| 4 | Apply `Relativize` in provider `Detect` (G3; stable cross-lane IDs) | 3 | 4 | 12 | after #2; provider golden regen (ids change — intended) |
| 5 | Close stale TODO rows G8 (done upstream) and rewrite T07b's remaining scope as art-dupl-side only | 2 | 5 | 10 | none |
| 6 | (Optional, deliberate) add `runs[].properties` to the CLI SARIF wire + golden regen | 2 | 3 | 6 | AGENTS.md:323 wire-change intent |

## RECONSIDER-WHEN

Re-open this audit's N/A verdicts when: art-dupl grows an auto-fix path (→ Edits/pipeline/
SimpleFixes become applicable); a machine consumer requests finding-level diffs across runs
(→ Merge/Diff/Correlate); `cmd/go-finding` gains toolsdk discovery (→ provider becomes
CLI-runnable upstream); a second rule/category forces the Report query lane over streaming;
or upstream exports a public Severity→SARIF-level mapping (→ row 25). Re-check version currency
on every go-finding train (modules tag independently — AGENTS.md:241).

## Appendix: how to verify

Commands run for this audit (replay literally; paths absolute):

```bash
# Checkouts & tags
git -C /home/lars/projects/art-dupl rev-parse HEAD          # cee67b4a47a8e759f5af05669cad97dca8d1f22d
git -C /home/lars/projects/go-finding rev-parse HEAD        # 5f568e058a15221ac198d378a1d907231451edf0
git -C /home/lars/projects/go-finding describe --tags --exact-match HEAD   # v1.16.0
git -C /home/lars/projects/go-finding tag --sort=-v:refname | head -5      # v1.16.0 v1.15.1 v1.15.0 v1.14.0 v1.13.0
git -C /home/lars/projects/go-finding tag --sort=-v:refname | grep -E '^(toolsdk|analysis|pipeline|cmd)/' | head -8
git -C /home/lars/projects/go-finding log -1 --format=%ci v1.15.0 v1.15.1 v1.16.0   # per-tag dates

# v1.16.0 stamp-defect / fetchability
git -C /home/lars/projects/go-finding log --oneline v1.15.1..v1.16.0 | head -3
git -C /home/lars/projects/go-finding show v1.16.0 --stat --oneline | head -12
git -C /home/lars/projects/go-finding show v1.16.0:report.go | grep -n "func (r \*Report) SetRunProperty"        # :140
git -C /home/lars/projects/go-finding show v1.16.0:sarif_import.go | grep -n "func RunPropertiesFromSARIF"       # :29
git -C /home/lars/projects/go-finding show v1.16.0:relativize.go | grep -n "func Relativize"                      # :30
git -C /home/lars/projects/go-finding show v1.16.0:version.go | grep -n "VersionMinor"                            # = 16
grep -n "1.16" /home/lars/projects/go-finding/CHANGELOG.md   # (no match — [Unreleased] header defect)

# Coverage channel release state (T07b)
git -C /home/lars/projects/go-finding show v1.15.0:detect_result.go | head -5              # exists at art-dupl's pin
git -C /home/lars/projects/go-finding log --oneline --follow -- detect_result.go | tail -2 # a275b57 feat: DetectResult + DetectorV2 (T07b)
git -C /home/lars/projects/go-finding show v1.15.0:report.go | grep -n "FilesScanned\|SkippedModules"  # :71/:78
git -C /home/lars/projects/go-finding show v1.15.0:relativize.go | head -3                 # fatal: not in v1.15.0

# Consumer pins & usage
cd /home/lars/projects/art-dupl && go list -m all | grep go-finding    # core v1.15.0, toolsdk v1.16.0
grep -n "larsartmann/go-finding" /home/lars/projects/art-dupl/go.mod   # :90, :91
rg -n "DetectV2|DetectorV2|RunDetect|FilesScanned|SkippedModules|SetRunProperty|Relativize" \
   /home/lars/projects/art-dupl --glob '!docs/**' -g '!*.md'           # no production hits (printer/stats* hits are art-dupl's own text/HTML stats)
rg -n "larsartmann/go-finding" /home/lars/projects/art-dupl -l --glob '!go.mod' --glob '!go.sum'
rg -c "gofinding\.|toolsdk\." /home/lars/projects/art-dupl/printer/finding/finding.go \
   /home/lars/projects/art-dupl/pkg/provider/provider.go \
   /home/lars/projects/art-dupl/printer/lsp.go /home/lars/projects/art-dupl/printer/sarif.go
rg -n "func thresholdFromContext|func boolOptionFromContext|func (cloneDetector) Detect" /home/lars/projects/art-dupl/pkg/provider/provider.go

# Prior-report context
sed -n '960,1804p' /home/lars/projects/art-dupl/docs/research/2026-10-05_go-finding-deep-dive.html | sed 's/<[^>]*>//g' | grep -v '^\s*$'
grep -n "G8\|T07b" /home/lars/projects/art-dupl/TODO_LIST.md
```

Line numbers stamped in this report were read at the cited commits and are reproducible with
`sed -n '<start>,<end>p'` / `rg -n` on the same checkouts.
