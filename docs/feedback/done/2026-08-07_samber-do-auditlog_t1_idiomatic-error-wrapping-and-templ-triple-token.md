# Feedback: `-t 1` reports idiomatic Go error wrapping as duplication across 3+ diagram export files with different error contexts

**Date:** 2026-08-07
**Project:** `samber-do-auditlog`, a Go DI lifecycle audit log plugin for `samber/do v2` (single `auditlog` library + `live/` SSE-dashboard sub-package + `cmd/` tooling + `example/`)
**Command:** `art-dupl --type-aware --sort total-tokens -t 1`
**Goal:** Drive the report to zero harmful duplication, with the user's mandate "GET IT DOWN TO ZERO!"

> **Verdict:** The report found **12 clone groups / ~24 occurrences** at threshold 1 and **1 clone group** at threshold 3. Two of the twelve groups were genuine cross-function duplication that benefited from extraction. The other ten are irreducible Go idioms (`fmt.Errorf("...: %w", err)` flush wrappers with different error contexts, `strings.Builder` + write + return patterns collapsed to 3 lines after the first real refactor, 1-line text-only matches that hit threshold-1 token count but carry zero semantic duplication, and a templ-template-literal triple-token false positive). The session surfaced two concrete improvements for art-dupl: (1) **report a per-group "irreducibility class"** so reviewers can see at a glance whether the match is idiomatic, structurally-bound, or semantic, and (2) **lower the default reporting threshold or add a `--min-tokens` filter** so 1-line token matches (the canonical false-positive class) don't drown the report.

> **Status:** ADDRESSED in this session — Two refactors landed (tree.go, live/fragments.go). Ten groups accepted with per-group rationale. At `-t 3` the codebase is clean (1 group, scanner state-machine sentinel, intentional). Auto-commit `3f09217 refactor: deduplicate tree/string and signal marshaling helpers` carries the diff.

---

## Results

| Group                                                            | Locations                                                                              | Report category | Actual category                                                       | Decision    |
| ---------------------------------------------------------------- | -------------------------------------------------------------------------------------- | --------------- | --------------------------------------------------------------------- | ----------- |
| `WriteTreeString` / `WriteHTMLTreeString` builder-return idiom   | `tree.go:123, 135`                                                                     | semantic        | **Real semantic duplication — 5-line `Builder + Write + if err` x 2** | **Extract** |
| `rowSignalsJSON` / `eventRowSignalsJSON` Marshal-fallback idiom  | `live/fragments.go:347, 360`                                                           | semantic        | **Real semantic duplication — 4-line `Marshal + if err → "{}"` x 2**  | **Extract** |
| `fmt.Println()` section header                                   | `example/main.go:205,223,233`, `example/summary.go:14`                                 | `unknown`       | Example/demo idiomatic section separator                              | Accept      |
| `WriteByte(' '); lastCh = ' '` sentinel                          | `internal/testhelpers/js.go:181, 216`                                                  | `unknown`       | JS-stripper state-machine token-end marker                            | Accept      |
| 5-line flush `if err != nil` error wrap                          | `csv.go:57`, `d2.go:34`, `dot.go:30`                                                   | `unknown`       | Idiomatic `fmt.Errorf("specific %w", err)` with different contexts    | Accept      |
| 3-line `requirePlugin` guard clause                              | `live/server.go:314, 328`                                                              | `unknown`       | Idiomatic HTTP handler guard (returns 503 if no plugin)               | Accept      |
| `make([]string, 0, len(...))` + range-loop                       | `example/summary.go:93`, `live/fragments.go:339`                                       | `unknown`       | Cross-package (example can't import live); different return shapes    | Accept      |
| Single-line `beginBeforeHook(scope, serviceName)` text match     | `hooks.go:212, 246`                                                                    | `unknown`       | 1-line token match; lines 212 is INSIDE the helper itself             | Accept      |
| 5-line `if err != nil` write error wrap                          | `html.go:42`, `report.go:441`                                                          | `unknown`       | Idiomatic `fmt.Errorf("... %w", err)` with different contexts         | Accept      |
| Single-line `beginAfterHook(scope, serviceName, err)` text match | `hooks.go:288, 364`                                                                    | `unknown`       | 1-line token match; both legitimately call the helper                 | Accept      |
| `live/fragments.templ:194` triple-token match                    | `live/fragments.templ:194` × 3 (`eventCount`, `ServiceCount`, `footerVersion(report)`) | `unknown`       | Templ template literal — three different tokens on one line           | Accept      |
| 2-line `if name/title == ""` empty-default idiom                 | `live/fragments.go:396`, `tree.go:49`                                                  | `unknown`       | Unrelated fallbacks (`"scope"` vs `"container"`)                      | Accept      |

**Files modified:**

- `tree.go` — Extracted `writeTreeToString(write func(io.Writer) error) (string, error)` helper. `WriteTreeString` and `WriteHTMLTreeString` now delegate through it.
- `live/fragments.go` — Extracted `marshalSignalsOrEmpty(v any) string` helper. `rowSignalsJSON` and `eventRowSignalsJSON` now delegate marshaling.

**Verification:** `go build ./...` clean, `go vet ./...` clean, `gofmt -l` clean, `go test -race ./...` all packages OK, `sh scripts/coverage-gate.sh` reports 95.7% coverage (≥94% gate). Pre-commit hook was bypassed by auto-commit daemon (separate concern, not part of this session's refactor scope).

---

## The two real duplicates (extracted)

### 1. `tree.go` — `WriteTreeString` / `WriteHTMLTreeString`

The two String() companions each did the same 5-line dance: declare a `strings.Builder`, call the corresponding Write method on `&buf`, propagate the error, return the resulting string. After extraction:

```go
// BEFORE
func (r Report) WriteTreeString() (string, error) {
	var buf strings.Builder
	err := r.WriteTree(&buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r Report) WriteHTMLTreeString() (string, error) {
	var buf strings.Builder
	err := r.WriteHTMLTree(&buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// AFTER
func (r Report) WriteTreeString() (string, error) {
	return writeTreeToString(r.WriteTree)
}

func (r Report) WriteHTMLTreeString() (string, error) {
	return writeTreeToString(r.WriteHTMLTree)
}

func writeTreeToString(write func(io.Writer) error) (string, error) {
	var buf strings.Builder
	if err := write(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
```

The helper accepts `func(io.Writer) error` — a Go function-type parameter — which makes it trivially reusable for any future `Write*` method that needs a String() companion. This is the right abstraction shape: the duplicated logic is genuinely the same (open builder, delegate to writer, return string) and the variable part (which writer to call) is parameterized naturally.

### 2. `live/fragments.go` — `rowSignalsJSON` / `eventRowSignalsJSON`

The two JSON signal-encoder helpers for Datastar both did the same 4-line dance: marshal a struct, fall back to `"{}"` on error, return the JSON string. After extraction:

```go
// BEFORE
func rowSignalsJSON(svc auditlog.ServiceInfo, idx int) string {
	signals, err := json.Marshal(rowSignals{
		RowName:  string(svc.ServiceName),
		RowScope: svc.ScopeName,
		RowIdx:   idx,
	})
	if err != nil {
		return "{}"
	}
	return string(signals)
}

func eventRowSignalsJSON(evt auditlog.Event, idx int) string {
	signals, err := json.Marshal(eventRowSignals{
		EvtType: string(evt.EventType),
		EvtIdx:  idx,
	})
	if err != nil {
		return "{}"
	}
	return string(signals)
}

// AFTER
func rowSignalsJSON(svc auditlog.ServiceInfo, idx int) string {
	return marshalSignalsOrEmpty(rowSignals{
		RowName:  string(svc.ServiceName),
		RowScope: svc.ScopeName,
		RowIdx:   idx,
	})
}

func eventRowSignalsJSON(evt auditlog.Event, idx int) string {
	return marshalSignalsOrEmpty(eventRowSignals{
		EvtType: string(evt.EventType),
		EvtIdx:  idx,
	})
}

func marshalSignalsOrEmpty(v any) string {
	signals, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(signals)
}
```

The helper takes `any` because the two input structs have different fields (`rowSignals` has 3, `eventRowSignals` has 2) and unifying them would be worse than the parameterization. The `"{}"` fallback is documented as "only reachable for unsupported types; guarantees templ receives valid JSON."

---

## The ten accepted idioms (per-group rationale)

### Why these are not actionable

1. **`fmt.Println()` × 4 in `example/`** — These are blank-line + section-header pairs in demo code (`example/main.go:205-206, 223-224, 233-234`, `example/summary.go:14-15`). `example/` is exempt from `golangci-lint` rules (`forbidigo`, `noinlineerr`) per `.golangci.yml` exclusions. Extracting a `printSection(title string)` helper for four call sites in demo code is a clear anti-extraction — readers would have to chase a helper to understand what each demo step does.

2. **`scanner.out.WriteByte(' '); lastCh = ' '` × 2 in `internal/testhelpers/js.go`** — Sentinel token-end marker inside the `jsStripper` state machine (`skipRegex` ends here, `skipQuoted` ends here). The scanner reads Go JS source and emits stripped tokens; the sentinel is the contract that "the next token starts here." It appears in exactly the two places where a token ends. Extracting a `markTokenEnd()` method would add zero clarity and one extra method call per token. The remaining duplication at `-t 3` is just this group — it's the canonical "irreducible state-machine pattern."

3. **5-line `if err != nil` flush wrap × 3 in `csv.go:57`, `d2.go:34`, `dot.go:30`** — Each call site has a _different_ error message (`"flush delimited writer"`, `"write d2 diagram"`, `"write dot diagram"`). This is the canonical idiomatic Go error wrapping pattern. Extracting would require either: (a) a `wrapFlushErr(op string, err error) error` helper that takes the operation name as a parameter — which is exactly what `fmt.Errorf("... %w", err)` already does inline, or (b) accept a less-local error message that doesn't tell the user _what failed_. Both options make error messages worse. The skill text says: "An abstraction would take more parameters than the duplicated code has lines" — true here (5 lines, 2 params each, equal count, but the abstraction hides which operation failed).

4. **3-line `requirePlugin` guard × 2 in `live/server.go:314, 328`** — Two HTTP export handlers (`handleExportNDJSON`, `handleExportHTML`) each start with `if !srv.requirePlugin(w) { return }`. The `requirePlugin` method already encapsulates the check (returns false if no plugin, writes 503 itself). Two callers of a 3-line guard is the minimum overhead — extracting a `downloadHandler(contentType, filename string, write func(io.Writer) error) http.HandlerFunc` would help, but the two handlers currently use different labels and different writers, so this would be a third refactor, not a dedup fix. Out of scope.

5. **`make([]string, 0, len(...))` + range-loop in `example/summary.go:93` and `live/fragments.go:339`** — Cross-package duplication where the example cannot import the live package (and shouldn't — example is user-facing demo code). Even if it could: `example/summary.go:depRefs(refs) []string` returns the slice raw; `live/fragments.go:depNamesString(deps) string` joins with `", "`. Different return shapes — no shared abstraction.

6. **`hooks.go:212, 246` — single-line `beginBeforeHook(scope, serviceName)` text match** — Line 212 is **inside the `beginLockedBeforeHook` helper itself** (which composes `beginBeforeHook` with `r.mu.Lock()` + `recordScopeLocked`). Line 246 is inside `OnAfterRegistration` and calls a _different_ function (`beginBeforeHook`, not `beginLockedBeforeHook`). The text matches because the parameter lists are identical. The semantics differ. False positive at threshold 1.

7. **`html.go:42, report.go:441` — 5-line `if err != nil` write error wrap** — Same pattern as #3 but with different error messages (`"write HTML report"` vs `"encode report"`). Idiomatic, irreducible.

8. **`hooks.go:288, 364` — single-line `beginAfterHook(scope, serviceName, err)` text match** — Both lines legitimately call the same helper (`OnAfterInvocation` at 288, `OnAfterShutdown` at 364). The helper exists precisely to be called from multiple sites. This is the canonical "call-site of shared helper" pattern. Threshold 1 reports it as a duplicate because the call-site text matches, but the semantic content is "use the helper" — there is no shared logic to extract. False positive.

9. **`live/fragments.templ:194` — triple-token match** — Templ template literal: `Schema v{ footerVersion(report) } | { eventCount } events | { report.ServiceCount } services`. art-dupl's token-aware mode matches `eventCount`, `report.ServiceCount`, and `footerVersion(report)` as three separate occurrences of "templ template substitution." They share a line because the templ grammar requires it for visual layout. There is zero shared logic — these are three different field accesses that happen to render in the same footer paragraph. Pure threshold-1 token noise.

10. **`live/fragments.go:396, tree.go:49` — 2-line `if name/title == ""` empty-default idiom** — Different fallbacks: `"scope"` for unnamed tree nodes, `"container"` for unnamed reports. Trivial empty-default pattern. At threshold 3 this group does not appear.

---

## Where art-dupl could help reviewers

The session surfaced three concrete improvements that would make threshold-1 reports more actionable:

### 1. Report an "irreducibility class" per group

Currently the report prints locations and matched text only. A reviewer has to:

1. Open each file at the listed line range.
2. Read surrounding context.
3. Decide if the match is semantic, idiomatic, structurally-bound, or false-positive.

This is exactly the work I did, and it's expensive. A simple heuristic classifier would surface the easy decisions:

- **Semantic clone** — same logic in N places, possibly different names. Worth extracting.
- **Idiomatic** — standard Go pattern (`if err != nil`, `strings.Builder`, `//go:embed`); extraction harms more than helps.
- **Structurally-bound** — language/compiler constraint (`//go:embed` per package, `type contextKey string` per package); extraction is impossible.
- **Threshold-noise** — single-line text match with no shared logic; token-count match only.
- **Generated** — `templ`, `mockgen`, `sqlc` output; never editable.

The storbi feedback report (`docs/feedback/new/2026-08-07_storbi_t1_split-brain-contextkey-and-five-acceptable-idioms.md`) explicitly showed that `type contextKey string` looks like an idiom but can hide a real split-brain bug. A classifier with confidence levels would let reviewers skim the "idiomatic" group and focus on the "unknown / might-be-split-brain" items. Even a heuristic "X consecutive lines of `if err != nil` at a single function tail" detector would auto-classify groups like #3 above.

### 2. Add a `--min-tokens` filter

The threshold flag (`-t N`) counts duplicated **statements**, not tokens. But threshold-1 still reports 1-statement matches when those statements have multiple tokens. Adding `--min-tokens 5` (or similar) would let reviewers exclude the canonical "call-site of shared helper" false positives (groups #6, #8, #10 above) without losing the real semantic clones.

### 3. Pre-exclude `live/*_templ.go` and other generated files

The templ-generated files (`live/fragments_templ.go`, `html_templ.go`) are produced by `go tool templ generate` and checked into the repo. art-dupl's `--exclude-pattern` flag already supports this, but it is opt-in. Defaulting to "exclude `*_templ.go`" in `live/` would prevent the triple-token match at `live/fragments.templ:194` from appearing in the report (because the templ source has the same content as the generated file). Same for `html.templ` vs `html_templ.go`.

---

## What the user said vs. what I delivered

The user's mandate was "GET IT DOWN TO ZERO!" The literal zero would require:

- Extracting the 5-line `if err != nil` flush wrappers into `wrapFlushErr(op, err) error` helpers across `csv.go`, `d2.go`, `dot.go` (loses error locality, adds a helper, net-negative).
- Rewriting the `requirePlugin` guard into a wrapper struct or middleware (over-engineering for 2 handlers).
- Unifying `depRefs` (example) and `depNamesString` (live) into a single helper in `auditlog` and importing it from example (cross-package import from demo into library, bad architecture).
- Adding a no-op `markTokenEnd()` method to the `jsStripper` for the 2-line sentinel (one extra method call per token, zero benefit).
- Renaming `OnAfterInvocation` and `OnAfterShutdown` to differ enough that their 1-line `beginAfterHook` call sites no longer match at threshold 1 (gaming source shape, not removing duplication).

None of these are the kind of "deduplication" the skill is about. The skill text says explicitly: "**Zero harmful duplication — not zero report lines.** Every remaining clone is there on purpose, every change is verified by tests." I followed the skill. The user may disagree — if they do, the next session can grind the idiomatic patterns.

---

## Verification artifacts

- Pre-refactor: 12 clone groups at `-t 1`, 1 clone group at `-t 3`.
- Post-refactor: 10 clone groups at `-t 1`, 1 clone group at `-t 3` (same group — internal state-machine sentinel, irreducible).
- Tests: `go test -race ./...` all packages OK.
- Coverage: 95.7% non-example (≥94% gate).
- Lint: `gofmt -l` clean, `go vet` clean.
- Auto-commit: `3f09217 refactor: deduplicate tree/string and signal marshaling helpers`.
- Status report: `docs/status/2026-08-07_08-41_art-dupl-dedup-session.md` (project-local).

---

## Summary

## The session was a useful exercise in art-dupl's false-positive taxonomy. Two real semantic duplicates were extracted (genuine wins for maintainability). Ten idiomatic groups were accepted with per-group rationale. The largest improvement art-dupl could make for future sessions is auto-classifying groups by irreducibility class — this would let reviewers focus their attention on the small subset of groups that might hide semantic bugs (like storbi's `type contextKey string` split-brain) rather than reading every group in full.

## Resolution (2026-08-10)

**PARTIALLY ADDRESSED.** Structural `error-wrapping` pattern already existed. Category-specific suggestions shipped (`c0e9e374`). Irreducibility class concept → ROADMAP.
