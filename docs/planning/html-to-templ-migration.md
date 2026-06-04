# Plan: Convert HTML Printer from Go fmt.Fprintf to .templ

## Context

The HTML printer (`printer/html*.go`) generates HTML via `fmt.Fprintf` with inline string templates. This is ~1,500 lines of fragile, hard-to-maintain Go code mixing presentation and logic. Converting to `.templ` gives us:

- **Type-safe HTML** — compile-time element/attribute checking
- **Auto-escaping** — templ handles HTML escaping by default
- **Separation of concerns** — HTML structure in `.templ`, data prep in `.go`
- **Composability** — small reusable components instead of monolithic `fmt.Fprintf` chains
- **Better DX** — IDE support, formatting, linting for templates

## Current State (files to convert)

| File | Lines | Role |
|------|-------|------|
| `printer/html_template.go` | 523 | CSS/HTML shell template (const string) |
| `printer/html.go` | 317 | Core struct, PrintHeader, PrintClones, writeCloneGroupHeader, writeCloneOccurrences, writeCloneGroupFooter |
| `printer/html_diff.go` | 342 | Diff view rendering (writeDiffView, writeDiffViewToggle, writeDiffSelector, writeDiffComparison, renderDiffLines) |
| `printer/html_summary.go` | 304 | Summary section, PrintFooter, OutputHTML, JavaScript |

**Total: ~1,486 lines** of `fmt.Fprintf`-based HTML generation.

## Challenges

1. **Streaming architecture**: The current `htmlprinter` streams HTML incrementally via `PrintHeader()` → `PrintClones()` → `PrintFooter()`. Templ components render atomically to an `io.Writer`. We need to reconcile these patterns.
2. **Interleaved state**: `htmlprinter` tracks `iota` (group counter), `classificationStats`, and collected clones across calls. This state management stays in Go; templ gets pure data.
3. **JavaScript**: The footer contains ~80 lines of inline JS. Templ's `<script>` tags handle this natively.
4. **CSS**: 500+ lines of CSS in a Go string constant. Templ supports `<style>` tags natively — can move to a `.templ` file or a separate CSS file.
5. **Dynamic attributes**: `data-category`, `data-priority`, `onclick` handlers with dynamic IDs — all natively supported by templ.
6. **Raw HTML injection**: `buildSummarySection()` returns raw HTML string. Convert to a proper templ component that receives structured data.
7. **Build pipeline**: Need `templ generate` integrated into `flake.nix` and `justfile`.
8. **Golden tests**: 2 golden test files assert exact byte output. After conversion, output may have minor whitespace differences — golden files may need updating.

## Step-by-Step Plan

### Phase 1: Build Infrastructure (Prerequisite)

**Step 1.1**: Add `templ` CLI to `flake.nix` nativeBuildInputs
- Add `templ` to `nativeBuildInputs` in the Go package derivation
- Add `preBuild = "templ generate"` before Go compilation

**Step 1.2**: Add `justfile` recipe for `templ generate`
- Add `generate` recipe: `templ generate`
- Add `generate` as dependency of `build` recipe

**Step 1.3**: Add `//go:generate templ generate` directive
- Create `printer/generate.go` with `//go:generate templ generate` in the printer package

### Phase 2: Define View Models (Data Types)

**Step 2.1**: Create `printer/html_views.go` — pure data types for templates
- `HeaderViewData{Threshold int, Metadata ReportMetadata}`
- `CloneGroupViewData{GroupNum int, Category string, Priority string, HasTest bool, Occurrences int, TotalTokens int, BadgesHTML string, Suggestion string, Clones []CloneViewData}`
- `CloneViewData{VSCodeLink string, Filename string, LineStart int, CodeID string, Fragment string}`
- `DiffViewData{GroupNum int, Base CloneViewData, AggregateStats DiffStatsView, Comparisons []DiffComparisonView}`
- `DiffComparisonView{Index int, Active bool, Filename string, LineStart int, VSCodeLink string, Stats DiffStatsView, BaseLines []DiffLineView, ComparedLines []DiffLineView}`
- `DiffLineView{Type string, LineNumber int, Content string, IsModified bool, WordDiffBase string, WordDiffCompared string}`
- `SummaryViewData{TotalClones int, TotalTokens int, ProdCount int, TestCount int, Categories []CategoryView, Priorities []PriorityView}`
- `FooterViewData{Summary SummaryViewData}`

**Step 2.2**: Create conversion functions
- `toHeaderViewData(p *htmlprinter) HeaderViewData`
- `toCloneGroupViewData(groupNum int, clones []ProcessedClone) CloneGroupViewData`
- `toDiffViewData(groupNum int, diff CloneGroupDiff) DiffViewData`
- `toSummaryViewData(stats classificationStats) SummaryViewData`

### Phase 3: Create Templ Components (Incremental)

**Step 3.1**: Create `printer/report.templ` — page shell
- `reportPage(data HeaderViewData)` — DOCTYPE, head, style, body opening, header, stats grid, metadata section
- Move CSS from `html_template.go` constant into templ `<style>` block
- Move `<script>` JS from `html_summary.go` into templ `<script>` block

**Step 3.2**: Create `printer/clone_group.templ` — clone group rendering
- `cloneGroup(data CloneGroupViewData)` — group container, header with badges, body
- `cloneOccurrence(data CloneViewData, groupNum int, index int)` — single occurrence
- `suggestion(text string)` — suggestion box

**Step 3.3**: Create `printer/diff_view.templ` — diff visualization
- `diffView(data DiffViewData)` — diff mode container
- `diffViewToggle(groupNum int)` — side-by-side / inline toggle
- `diffSelector(data DiffViewData)` — dropdown selector
- `diffComparison(data DiffComparisonView, groupNum int)` — comparison panel
- `diffLine(data DiffLineView)` — single diff line
- `diffLegend()` — color legend

**Step 3.4**: Create `printer/summary.templ` — summary section
- `summarySection(data SummaryViewData)` — stats grid, category breakdown, priority breakdown, filter buttons

**Step 3.5**: Create `printer/scripts.templ` — JavaScript
- `reportScripts()` — copyCode, showDiff, toggleDiffView, filterClones, DOMContentLoaded init

### Phase 4: Refactor htmlprinter to Use Templ Components

**Step 4.1**: Refactor `PrintHeader()`
- Prepare `HeaderViewData` from htmlprinter state
- Call `reportPage(data).Render(ctx, p.w)` (or split into header-only render)
- **Design decision**: Since the current printer is streaming, we may need to keep the header/footer split. Options:
  - **Option A**: Keep streaming — `PrintHeader` renders `<!DOCTYPE>...<body><div class="container">`, `PrintFooter` renders `</div></body></html>`. Templ components for the inner parts only.
  - **Option B**: Buffer all clones, render entire page atomically in `PrintFooter`. This changes the interface semantics but is cleaner.
  - **Recommendation**: **Option A** — use `templ.Component` for reusable HTML fragments, but the orchestration stays in Go. Each `write*` method prepares data then renders a templ component.

**Step 4.2**: Refactor `writeCloneGroupHeader` / `writeCloneOccurrences` / `writeCloneGroupFooter`
- Replace `fmt.Fprintf` calls with `cloneGroup(data).Render(ctx, p.w)`
- Keep the data preparation in Go methods

**Step 4.3**: Refactor `writeDiffView` and related methods
- Replace `fmt.Fprintf` chains with templ diff components
- Keep `ComputeCloneGroupDiff` and diff algorithm logic in Go

**Step 4.4**: Refactor `buildSummarySection` / `PrintFooter`
- Replace `strings.Builder` HTML construction with `summarySection(data).Render(ctx, p.w)`
- Replace inline JS with `reportScripts().Render(ctx, p.w)`

**Step 4.5**: Refactor `OutputHTML`
- May simplify if buffering approach is viable

### Phase 5: Remove Legacy Code

**Step 5.1**: Delete `printer/html_template.go`
- The `htmlTemplate` const is now in `report.templ`

**Step 5.2**: Clean up `html.go`, `html_diff.go`, `html_summary.go`
- Remove all `fmt.Fprintf` HTML generation
- Keep only: struct definition, constructors, state management, data preparation functions
- Target: each file shrinks by 60-80%

**Step 5.3**: Remove `html` and `strconv` imports from cleaned files

### Phase 6: Update Tests

**Step 6.1**: Run existing tests — verify they pass
- Golden tests may need updating (whitespace differences)
- All 46 test functions in `html_test.go` should continue to work

**Step 6.2**: Update golden files if needed
- `TestHTMLOutputGolden.golden`
- `TestHTMLOutputGoldenNoDiff.golden`
- Regenerate with `just test -update` (or however golden files update)

**Step 6.3**: Verify BDD tests pass
- BDD tests in `bdd/` that use HTML output should continue passing

**Step 6.4**: Verify nix build passes
- `nix build` and `nix flake check` must pass with `templ generate` in the pipeline

### Phase 7: Documentation & Cleanup

**Step 7.1**: Update `AGENTS.md`
- Add templ generation to development commands
- Note the new `.templ` files in package organization

**Step 7.2**: Update `README.md` / `HOW_TO_USE.md` if needed
- Mention `templ generate` in development workflow

## Execution Order (Priority)

1. **Phase 1** (Build infra) — must be first, enables everything else
2. **Phase 2** (View models) — must be second, defines data contracts
3. **Phase 3.1** (report.templ page shell) — first template, proves the pattern
4. **Phase 3.2** (clone_group.templ) — most impactful conversion
5. **Phase 4.1-4.2** (Refactor header + clone rendering) — main conversion
6. **Phase 6** (Test) — verify after each sub-step
7. **Phase 3.3** (diff_view.templ) + **Phase 4.3** — second wave
8. **Phase 3.4-3.5** (summary.templ + scripts.templ) + **Phase 4.4** — final wave
9. **Phase 5** (Remove legacy) — cleanup
10. **Phase 7** (Documentation) — last

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Streaming vs atomic rendering mismatch | Keep streaming orchestration in Go; templ for fragments only |
| Golden test output changes | Accept minor whitespace diffs; regenerate golden files |
| `templ generate` in Nix sandbox | Verify templ CLI works in fixed-output derivation |
| Performance regression | Benchmark before/after; templ generates efficient Go code |
| 500+ line CSS in templ file | Consider extracting to separate `.css` file and embedding |
| Large number of tests to update (46) | Incremental conversion; run tests after each step |

## Estimated Effort

- Phase 1 (Build infra): ~30 min
- Phase 2 (View models): ~1 hour
- Phase 3 (Templ components): ~2-3 hours
- Phase 4 (Refactor): ~2-3 hours
- Phase 5 (Cleanup): ~30 min
- Phase 6 (Tests): ~1 hour
- Phase 7 (Docs): ~15 min

**Total: ~7-9 hours of focused work**

## Open Questions

1. **CSS extraction**: Keep CSS inline in `.templ` or extract to `.css` with `//go:embed`?
   - Recommendation: Keep inline for now (single-file output), extract later if needed.
2. **One `.templ` file or multiple?**
   - Recommendation: Multiple files by concern (report, clone_group, diff_view, summary, scripts) for maintainability.
3. **`context.Context` for rendering**: Need to pass `context.Background()` or `context.TODO()` since printer doesn't have a request context.
   - Recommendation: Use `context.Background()` — this is a CLI tool, not an HTTP server.
