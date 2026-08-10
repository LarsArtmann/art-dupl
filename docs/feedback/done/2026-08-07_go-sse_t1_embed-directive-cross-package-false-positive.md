# Feedback: `-t 1` reports `//go:embed`-anchored cross-package duplication that is structurally impossible to eliminate

**Date:** 2026-08-07
**Project:** `go-sse`, a Server-Sent Events transport library for Go (single `sse` package plus three self-contained `example/` packages)
**Command:** `art-dupl --type-aware --sort total-tokens -t 1`
**Goal:** Drive harmful duplication to zero across the library and examples without introducing detector-driven abstractions.

> **Verdict:** The report found **1 clone group / 2 occurrences / 7 tokens**, but the match is a **structural false positive caused by a Go compiler constraint, not a style choice.** The duplicated region is anchored on a `//go:embed` directive, which is lexically bound to the declaring package and *cannot be shared across packages by any refactoring*. This is a stronger class of false positive than the previously-documented "unwise to extract" idioms (mutex scopes, `strings.Builder` openers, functional-option contracts): here, extraction is **impossible**, not merely undesirable. The codebase has zero actionable duplication; literal zero would require either violating Go's embed semantics or gaming source shape.

> **Status (at time of writing):** Not yet addressed. No `//go:embed`-aware classification exists in the detector.

---

## Results

| Group                              | Locations                                    | Report category | Actual category                                      | Decision |
| ---------------------------------- | -------------------------------------------- | --------------- | ---------------------------------------------------- | -------- |
| `staticFiles embed.FS` + sub-FS open | `example/datastar/main.go:56-62`, `example/htmx/main.go:32-38` | `unknown`       | Go compiler-directive-bound declaration (`//go:embed`) | Accept   |

No source files were modified. `go test ./... -race`, `go vet ./...` clean. The single reported group is a language-mandated pattern, not a refactoring opportunity.

---

## Finding: `//go:embed all:static` + `embed.FS` declaration + `fs.Sub` open

### What the report matched

Both occurrences are the static-file bootstrap block at the top of two independent `package main` example programs:

```go
// example/datastar/main.go:55-62
//go:embed all:static
var staticFiles embed.FS

func main() {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("static sub FS: %v", err)
	}
	// ... DataStar-specific server setup ...
}
```

```go
// example/htmx/main.go:31-38
//go:embed all:static
var staticFiles embed.FS

func main() {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("static sub FS: %v", err)
	}
	// ... HTMX-specific server setup ...
}
```

The matched region (`56-62` / `32-38`) is the `var staticFiles embed.FS` declaration plus the `fs.Sub` error-handling stanza inside `main`. The `//go:embed` comment itself sits one line above the reported range but is the structural cause of the match: the `var` declaration it annotates is what makes the two regions textually and semantically identical.

### The Go language constraint (why this is not a style choice)

`//go:embed` is a **compiler directive**, not a statement or an import. Per the Go `embed` package contract:

1. **The directive may only appear immediately before a `var` declaration in the same file.**
2. **The `var` it populates must be in the same package as the directive.**
3. **There is no mechanism — no import, no re-export, no alias — to share an embedded `embed.FS` across packages without re-declaring the directive in each package.**

This is enforced by the compiler. Unlike mutex scopes (where extraction is *unwise* because it hides lock ordering) or `strings.Builder` openers (where extraction *saves no logic*), an `embed.FS` declaration **cannot** be moved to a shared package: the directive would no longer resolve, and the build would fail.

Every Go program that serves embedded static assets from more than one entry point pays this cost identically and unavoidably.

### Why extraction is impossible (and the partial extraction is worse)

The complete duplication cannot be removed. The only *partial* extraction would move the `fs.Sub` + `log.Fatalf` three lines into a shared helper:

```go
// example/internal/staticfs/staticfs.go
package staticfs

func Sub(root embed.FS) fs.FS {
	sub, err := fs.Sub(root, "static")
	if err != nil {
		log.Fatalf("static sub FS: %v", err)
	}
	return sub
}
```

But each `main.go` would still be required to contain:

```go
//go:embed all:static
var staticFiles embed.FS

func main() {
	staticFS := staticfs.Sub(staticFiles)
	// ...
}
```

Net effect of the "fix":

1. **The clone group survives.** The `//go:embed` directive + `var staticFiles embed.FS` declaration — the actual root of the match — is still duplicated verbatim in every package, because the compiler demands it.
2. **A new shared package is introduced** (`example/internal/staticfs`), with the coupling and import machinery that implies.
3. **The examples are no longer self-contained.** The project's AGENTS.md explicitly states: *"Three runnable examples, each an independent `package main`."* Each example is meant to be readable and runnable in isolation. Adding a shared internal dependency breaks that pedagogical boundary.
4. **Three lines saved per call site, four lines added in the helper.** Negative value even before counting the import.
5. **The hard-coded `"static"` directory name** gets baked into the helper, removing a degree of freedom that each example currently owns independently.

This satisfies the detector's own "Accept" criterion verbatim: *"An abstraction would take more parameters than the duplicated code has lines."* It also satisfies *"Different business rules or domain contexts behind a similar shape"* — the two examples are deliberately distinct demos (DataStar reactive signals vs. HTMX HTML-fragment swaps) that happen to share an asset-serving bootstrap.

### Additional context: these are example packages, not production code

The `example/` directory contains throwaway demonstration programs. Their purpose is to be read top-to-bottom by a developer learning the library. Duplicating a 5-line bootstrap across two examples is **more readable** than indirecting through a shared helper, because the reader of `example/htmx/main.go` never needs to open another file to understand how static files are served. This is the same reason Go's standard-library examples happily repeat setup boilerplate.

---

## Tool feedback

### The core signal: detect `//go:embed` as a duplication anchor

The cheapest, highest-signal improvement is to teach the detector that a clone occurrence immediately preceded by (or containing) a `//go:embed` comment is **structurally uneliminable**. The directive is a comment in the AST (`*ast.Comment` with `Text` starting with `//go:embed`), so detection is trivial and does not require type information.

Suggested classification:

- **Pattern:** `go-embed-directive` (or more broadly: `compiler-bound-declaration`)
- **Priority:** informational / suppressed at all thresholds
- **Suggestion:** *"Declaration is bound to a `//go:embed` compiler directive and cannot be shared across packages; duplication is mandated by Go's embed semantics."*
- **Actionability test:**
  - occurrence contains or is immediately preceded by a comment matching `//go:embed`;
  - the annotated declaration is a package-level `var` of type `embed.FS`, `[]byte`, or `string`;
  - occurrences are in different packages.
- **Action:** auto-accept; do not count toward the actionable-clone total.

### Why this matters beyond embed

`//go:embed` is the most common offender, but the same "compiler-bound declaration" reasoning applies to a small family of Go compiler directives that force package-local declarations:

| Directive            | Forced local declaration                          | Shared concept                                            |
| -------------------- | ------------------------------------------------- | --------------------------------------------------------- |
| `//go:embed`         | package-level `var` of `embed.FS`/`[]byte`/`string` | Embedded assets, repeated per serving entry point         |
| `//go:linkname`      | local `var`/func symbol paired with a target      | Runtime linking, inherently per-package                   |
| `//go:generate`      | (directive itself, often repeated per package)    | Code-generation triggers, per-package by design           |
| `func init()` bodies | one per package                                   | Per-package initialization, no sharing mechanism exists   |

A unified `compiler-bound-declaration` category would cover all of these. The unifying property: **Go offers no language mechanism to share these declarations across packages, so any cross-package clone anchored on them is not a refactoring target.**

### Threshold sensitivity

At `-t 1` this finding appears because the matched statement count (the `var` + `fs.Sub` + error check) clears the minimum. At `-t 5` (the documented default) it would still appear — the block is 5-7 statements depending on how `func main() {` is counted. So this is **not a threshold artifact** like the `bool-accumulator-initializer` case; it would survive at the default threshold and is worth classifying on its own merits rather than relying on threshold tuning.

---

## What works well

### Accurate structural detection

The detector correctly identified the two regions as semantically and textually identical. Type-aware mode did not produce a spurious mismatch from the fact that both variables are named `staticFiles` of type `embed.FS` — that is exactly the shared shape. The detection is correct; only the *actionability classification* is wrong.

### Clean, minimal output at `-t 1`

For a 25-file Go project, `-t 1` surfaced exactly one group. That is a high signal-to-noise ratio at the most aggressive threshold and speaks well of the `*_templ.go` and generated-code exclusion (the project has checked-in `*_templ.go` files that were correctly ignored). The single false positive is a language-level artifact, not noise from test boilerplate or table-driven patterns.

### Honest reporting

The text output (`found 2 clones: ... | staticFiles embed.FS`) named the shared symbol clearly, making triage fast. The HTML report (inspected but not relied upon here) showed accurate file/line links.

---

## Recommended improvements

### 1. Recognize `//go:embed` as a non-actionable anchor (highest value, lowest effort)

Detect `//go:embed` comments adjacent to a clone occurrence and classify the group as `go-embed-directive`. This is a near-zero-effort heuristic (comment text match) with very high precision in Go codebases. The classification should:

- suppress the group from the actionable total,
- still report it (informationally) so the user sees the detector did its job,
- emit the domain-specific suggestion above instead of "Review and extract common logic."

### 2. Introduce a `compiler-bound-declaration` umbrella category

Generalize beyond embed to cover `//go:linkname`, repeated `func init()`, and other Go constructs whose duplication is mandated by the language rather than chosen by the author. The defining property is: *no Go refactoring can reduce this duplication; it exists because the language offers no sharing mechanism for this construct across package boundaries.*

This is conceptually distinct from the existing idiom categories (`sync-rwmutex-read`, `strings-builder-opener`, `functional-option-contract`), which are about extraction being **unwise**. `compiler-bound-declaration` is about extraction being **impossible**. That distinction is worth surfacing to the user, because "impossible" forecloses even the consideration of refactoring.

### 3. Separate "detected" from "actionable" totals (reiterated)

Echoing prior feedback (`go-output`, `go-humanize-linter`): the summary should report both:

```text
Detected clone groups:    1
Actionable clone groups:  0
Compiler-bound / idioms:  1
```

This preserves detector honesty (the clone *was* found) without pressuring the user — or an automated agent following a "drive to zero" instruction — into a harmful or impossible refactoring. The pressure to hit literal zero is acute when an agent is told "GET IT DOWN TO ZERO," and a compiler-bound group with no escape hatch invites either source-shape gaming or a genuinely damaging shared-package extraction.

### 4. Tailor the suggestion to the classification

Replace the blanket "Review and extract common logic" with pattern-aware guidance for compiler-bound groups:

```text
Compiler-bound: declaration is anchored on a //go:embed directive and cannot
be shared across packages. Duplication is mandated by Go's embed semantics;
no refactoring can reduce it.
```

### 5. Document a `//art-dupl:accept` escape hatch (lower priority)

For cases where the detector cannot yet classify a group but the user has made a deliberate accept decision, an inline marker (`//art-dupl:accept reason: ...`) that suppresses future reports of that group would let teams record their rationale next to the code. This complements automatic classification and is especially valuable for example/demo code where self-containment is a deliberate design property. (This was hinted at in prior feedback; the compiler-directive case is another strong argument for it.)

---

## Acceptance criteria for a fix

Run the same command against `go-sse`:

```bash
art-dupl --type-aware --sort total-tokens -t 1
```

A successful improvement should produce one of these outcomes:

1. **Zero actionable groups**, with the `staticFiles embed.FS` finding suppressed or reclassified as `go-embed-directive` / `compiler-bound-declaration` and excluded from the actionable count; **or**
2. **One informational group**, accurately classified, carrying the embed-semantics suggestion instead of "Review and extract common logic," with an actionable count of zero.

The detector should continue to report cases where an `//go:embed`-anchored declaration is followed by genuinely duplicated *domain* logic (not just the unavoidable `fs.Sub` open). The goal is not blanket suppression of `embed.FS`, but correct actionability classification for the compiler-mandated minimum.

---

## Conclusion

`art-dupl` correctly detected a real structural clone, but at `-t 1` it could not distinguish *duplication the author chose* from *duplication the compiler forces*. For Go's `//go:embed` directive — and the broader family of compiler-bound declarations — the duplication is not a maintenance burden and has no elimination path. The correct result for this codebase is **zero actionable duplication with one accepted compiler-bound group**, and the report should be able to express that directly rather than recommending an extraction that would either fail to compile or violate the project's self-contained-examples design.

This finding extends the idiom-suppression work already landed for `bool-accumulator-initializer` and proposed for mutex/builder/functional-option patterns. The new contribution is the recognition that some duplication is not merely *unwise to extract* but *impossible to extract*, and that distinction deserves its own classification.

---

## Resolution (2026-08-10)

**IDENTIFIED.** `//go:embed` directive pattern identified as zero-effort, high-precision. → TODO_LIST (feedback-driven actionability patterns).
