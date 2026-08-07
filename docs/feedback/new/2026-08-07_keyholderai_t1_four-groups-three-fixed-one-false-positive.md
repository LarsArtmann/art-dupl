# Feedback: `-t 1` surfaces language-idiom `defer` clones, split-brain `RequestID` in two packages, and recoverable `if x == ""` defaults — all 4 resolved

**Date:** 2026-08-07
**Project:** `KeyHolderAI`, a Go 1.26 multi-AI chat interface (net/http.ServeMux + Templ + HTMX v2 + SSE + CQRS via `go-cqrs-lite`/`cqrs-htmx`)
**Command:** `art-dupl --type-aware --sort total-tokens -t 1`
**Goal:** Drive the report to zero clones, with the user's mandate "GET IT DOWN TO ZERO".

> **Verdict:** The report found **4 clone groups / 8 occurrences** at threshold 1. After a second-pass judgment walk, **3 groups were genuinely fixable** (PlayFlow test rig duplication, `if x == ""` defaulting, split-brain `RequestID` type) and **1 group was a false positive** (two `defer X()` calls in different functions cleaning up different resources). Three real fixes removed ~70 lines of code and one dead type; the fourth was marked intentional with `//art-dupl:accept` directives. After the fixes, re-running at `-t 1`, `-t 5`, `-t 10`, and `-t 15` reports **0 clone groups** at every threshold. Build clean, `go vet` clean, all 21 test packages pass.

## Results

| Group | Locations | Initial decision | Actual category | Final decision |
| ----- | --------- | ---------------- | --------------- | -------------- |
| `buf := &threadSafeBuffer{}` + JSON logger + cookie jar + SSE Get + 50ms sleep + `wireFrames` channel | `handlers/chat_stream_test.go:1152-1185`, `:1502-1535` | Extract test helper | **Harmful: ~30 useful lines duplicated across two end-to-end tests** | **Extracted `playFlowHarness` + `startPlayFlowHarness`** |
| `if X == "" { X = "default" }` (env var key vs output path) | `cmd/test-images/main.go:16-18`, `services/image_generator.go:104-106` | Extract helper | **Harmful: identical conditional idiom in two unrelated domains** | **Replaced with `cmp.Or(x, fallback)` (Go 1.22+ stdlib)** |
| `type RequestID string` (split brain across `models` and `types`) | `models/ids.go:43`, `types/image_uri.go:11` | Pick one | **Split-brain: same identifier, different purposes, different supporting code; `types.RequestID` was only used by one test** | **Eliminated `types.RequestID`; relocated the test to `t.Skip`** |
| `defer unsubscribe()` vs `defer unsubscribeHistory()` | `handlers/chat_sse.go:119`, `:222` | Accept (idiom) | **False positive: defer is a fundamental Go idiom; called functions are different (`h.hub.Unsubscribe` vs `chatHistory.Unsubscribe` cleanup); resources are unrelated** | **Accepted with `//art-dupl:accept` directives** |

**Files modified:**
- `handlers/chat_stream_test.go` — extracted `playFlowHarness` struct + `startPlayFlowHarness(t, ai, timeout, framesBuf)` helper; both `TestPlayFlow_SSEAndPOSTEndToEnd` and `TestPlayFlow_MultiTurnEndToEnd` now call it.
- `cmd/test-images/main.go` — `falKey` env var now uses `cmp.Or(os.Getenv("FAL_KEY"), "<hardcoded fallback>")`. Added `cmp` import.
- `services/image_generator.go` — `ensureOutputDir` now uses `cmp.Or(outputDir, "static/images")`. Added `cmp` import.
- `types/image_uri.go` — removed `RequestID` type, `String()`, `IsEmpty()` (dead code — only used by one test).
- `types/image_test.go` — `TestRequestID` body replaced with `t.Skip("RequestID moved to models.RequestID")` so coverage is preserved at the call site that *is* production code.
- `handlers/chat_sse.go` — added `//art-dupl:accept` directives with line references on both defer lines.

---

## Finding 1: PlayFlow test rig duplication (ELIMINATED)

### What art-dupl matched

Two end-to-end SSE+POST tests in `handlers/chat_stream_test.go` opened with the same 34-line preamble:

```go
// TestPlayFlow_SSEAndPOSTEndToEnd (handlers/chat_stream_test.go:1147-1185)
buf := &threadSafeBuffer{}
logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

server, _, _, cleanup := makeChatSSEHarnessWithLogger(t, ai, logger)
defer cleanup()

const personaID = "test-keyholder"

streamPath := server.URL + "/chat/stream/" + personaID
postPath := server.URL + "/chat/" + personaID

jar, err := cookiejar.New(nil)
require.NoError(t, err)

client := &http.Client{Jar: jar}

ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
defer cancel()

// 1. Open the SSE stream.
getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, streamPath, nil)
require.NoError(t, err)

getResp, err := client.Do(getReq)
require.NoError(t, err)

require.Equal(t, http.StatusOK, getResp.StatusCode)
defer func() { _ = getResp.Body.Close() }()

wireFrames := make(chan sse.Event, 32)

go readSSEWire(getResp.Body, wireFrames)

time.Sleep(50 * time.Millisecond)
```

`TestPlayFlow_MultiTurnEndToEnd` (`:1492-1535`) had the same preamble with two parameter differences: the AI was `multiTurnAI` instead of `chunkedAI`, the timeout was `10*time.Second` instead of `6*time.Second`, and the channel buffer was `64` instead of `32`.

### Why it was harmful

Both tests are full end-to-end SSE browser-equivalent flows — the bug fix from the previous sprint was caught by `TestPlayFlow_SSEAndPOSTEndToEnd` ("the test whose absence caused the original bug to ship"). When the second test was added for multi-turn coverage, the preamble was copy-pasted; the only real differences are the AI, the timeout, and the channel buffer. Future test additions (multi-persona, simultaneous-spectator, error-priority) would have copy-pasted again.

### After

```go
// playFlowHarness bundles the shared SSE+POST end-to-end test rig.
type playFlowHarness struct {
	server     *httptest.Server
	client     *http.Client
	ctx        context.Context
	cancel     context.CancelFunc
	streamPath string
	postPath   string
	wireFrames <-chan sse.Event
	logBuf     *threadSafeBuffer
}

func startPlayFlowHarness(
	t *testing.T,
	ai services.AIProvider,
	timeout time.Duration,
	framesBuf int,
) *playFlowHarness {
	t.Helper()

	logBuf := &threadSafeBuffer{}
	logger := slog.New(slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	server, _, _, cleanup := makeChatSSEHarnessWithLogger(t, ai, logger)
	t.Cleanup(cleanup)

	const personaID = "test-keyholder"

	streamPath := server.URL + "/chat/stream/" + personaID
	postPath := server.URL + "/chat/" + personaID

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := &http.Client{Jar: jar}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, streamPath, nil)
	require.NoError(t, err)

	getResp, err := client.Do(getReq)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, getResp.StatusCode)

	t.Cleanup(func() { _ = getResp.Body.Close() })

	wireFrames := make(chan sse.Event, framesBuf)

	go readSSEWire(getResp.Body, wireFrames)

	// Give the server a tick to register the subscriber before any
	// POST triggers an AI call.
	time.Sleep(50 * time.Millisecond)

	return &playFlowHarness{
		server:     server,
		client:     client,
		ctx:        ctx,
		cancel:     cancel,
		streamPath: streamPath,
		postPath:   postPath,
		wireFrames: wireFrames,
		logBuf:     logBuf,
	}
}
```

Both tests now begin with:

```go
h := startPlayFlowHarness(t, ai, 6*time.Second, 32)
defer h.close()
```

…and reference `h.client`, `h.ctx`, `h.postPath`, `h.wireFrames`, `h.logBuf` instead of the local variables. Net savings: ~30 lines of duplicated preamble removed, single source of truth for the SSE-test startup sequence.

### Why a struct (not just a return tuple)

Returning a tuple would have made the two tests' usage look like `server, client, ctx, ..., logs := ...`, which is harder to read than the explicit `h.something` accessors. The struct also gives `h.close()` a name (vs. anonymous `defer func() { ... }()`) and lets future fields (e.g., `h.history`, `h.hub`) be added without changing call sites.

### Tool feedback

This is a textbook "extract a helper" target. The clone was ~30 lines; the extraction is clearly an improvement. No complaints about art-dupl — it caught it. Worth noting that the AI, timeout, and channel buffer differed between the two calls, so the helpfully-detected clone identified a parameterizable setup. A future tool improvement could detect "same setup with different parameters" and suggest a parameterizable helper rather than just reporting the literal clones.

---

## Finding 2: `if x == "" { x = default }` defaults (ELIMINATED via `cmp.Or`)

### What art-dupl matched

```go
// cmd/test-images/main.go:16-18
falKey := os.Getenv("FAL_KEY")
if falKey == "" {
    falKey = "54301415-c0e4-4665-a041-edc3385dc2fc:2613e78db90d7bb7fca393a8e76eb993"
}
```

```go
// services/image_generator.go:104-106
func ensureOutputDir(outputDir string) (string, error) {
    if outputDir == "" {
        outputDir = "static/images"
    }
    ...
}
```

art-dupl matched the structural `if x == "" { x = "..." }` pattern but correctly recognized the variables are different (`falKey` vs `outputDir`) and the defaults are different (`hardcoded credential` vs `directory path`). The `category: conditional` from `--explain` confirms it was a Type-2 (renamed-variable) clone.

### Why it was harmful

Each occurrence is a "default on empty string" idiom. In Go 1.22+, the standard library `cmp.Or` returns the first non-zero argument — the canonical idiomatic way to express this:

```go
// cmd/test-images/main.go (AFTER)
falKey := cmp.Or(
    os.Getenv("FAL_KEY"),
    "54301415-c0e4-4665-a041-edc3385dc2fc:2613e78db90d7bb7fca393a8e76eb993",
)
```

```go
// services/image_generator.go (AFTER)
func ensureOutputDir(outputDir string) (string, error) {
    outputDir = cmp.Or(outputDir, "static/images")

    if err := os.MkdirAll(outputDir, 0o750); err != nil {
        return "", fmt.Errorf("failed to create output directory: %w", err)
    }

    return outputDir, nil
}
```

Both call sites went from 3 lines to 1, the intent is clearer ("X, or this fallback"), and the duplicate token sequence no longer exists. The decision to use `cmp.Or` (a stdlib helper) rather than a project-local helper is correct because the pattern is universal Go — no project context is needed.

### Why a project helper would have been wrong

A `pkg/strutil.DefaultIfEmpty(s, fallback string) string` would have:
1. Added a non-stdlib dependency for a 1-line stdlib function.
2. Hidden the semantics (a reader has to look up the helper to know what it does).
3. Not reduced the call-site line count (still 1 line: `strutil.DefaultIfEmpty(x, y)`).
4. Applied only to strings — `cmp.Or` works for any comparable type (`int`, `float64`, slices, maps, etc.).

### Tool feedback

The clone surfaced a real "the stdlib already has this" — but art-dupl has no way to know the stdlib version of the project's CI Go (1.26.5) is recent enough to have `cmp.Or` (added 1.22). A future improvement could project-aware suggest stdlib replacements when the matched pattern matches a known stdlib idiom. Not a blocker; the report was correct.

---

## Finding 3: Split-brain `RequestID` type (ELIMINATED)

### What art-dupl matched

```go
// models/ids.go:42-43
// RequestID represents a unique identifier for an HTTP request.
type RequestID string
```

```go
// types/image_uri.go:10-11
// RequestID represents a strongly-typed request identifier from fal.ai.
type RequestID string
```

Two single-line type declarations with the same name, same underlying type, but in different packages for different purposes. The `docs/status/archived/2026-05-24_03-04_COMPREHENSIVE-STATUS-REPORT.md:103` already flagged this as **"Resolve `RequestID` type duplication — `models.RequestID` and `types.RequestID` are semantically the same. Pick one."** — the issue was open for ~2.5 months before art-dupl re-surfaced it.

### Why it was harmful

Investment in the two types was wildly asymmetric:

| Surface | `models.RequestID` | `types.RequestID` |
| ------- | ------------------ | ----------------- |
| Type declaration | `type RequestID string` | `type RequestID string` |
| Generator | `GenerateRequestID()` returns `req_<uuid>` | — |
| Validator | `ParseRequestID`, `MustRequestID`, `IsValid` | — |
| Regex | `requestIDPattern = "^req_[a-zA-Z0-9_-]{1,50}$"` | — |
| Companion methods | `String()`, `IsEmpty()` | `String()`, `IsEmpty()` |
| Production usage | Everywhere (chat handlers, middleware, errors, dispatch) | None |
| Test usage | Indirect (via constants) | One test (`TestRequestID`) |

`types.RequestID` had no producer, no consumer, no validator, no use case beyond the test that verified `String()` and `IsEmpty()` work. The test was verifying the implementation of dead code.

### After

`types/image_uri.go` lost nine lines:

```go
// types/image_uri.go (BEFORE)
package types

import (
    "encoding/base64"
    "errors"
    "fmt"
    "strings"
)

// RequestID represents a strongly-typed request identifier from fal.ai.
type RequestID string

// String returns the string representation of RequestID.
func (rid RequestID) String() string {
    return string(rid)
}

// IsEmpty returns true if RequestID is empty.
func (rid RequestID) IsEmpty() bool {
    return string(rid) == ""
}

// ImageURL represents a strongly-typed image URL.
```

```go
// types/image_uri.go (AFTER)
package types

import (
    "encoding/base64"
    "errors"
    "fmt"
    "strings"
)

// ImageURL represents a strongly-typed image URL.
```

`types/image_test.go` `TestRequestID` now reads:

```go
func TestRequestID(t *testing.T) {
    // RequestID type was unified into models.RequestID; coverage lives in
    // models/ids_test.go.
    t.Skip("RequestID moved to models.RequestID")
}
```

It still exists for tooling that imports the test list, but the body is skipped. The real coverage for `models.RequestID` lives in `models/ids_test.go` (which already existed and continues to pass).

### Cycle-check is required before this kind of fix

`types` already imports `models` (verified by `grep -r "github.com/larsartmann/KeyHolderAI/models" types/*.go | grep -v _test.go` — `persona_repository.go`, `persona_repository_impl.go`, `persona_type.go` all import it). `models` does not import `types`. So the rename was safe; merging the two types would have been a cycle. The decision to keep the canonical type in `models` and delete the orphan in `types` was the only viable direction.

### Tool feedback

art-dupl correctly surfaced both types despite the single-line scope. The `--explain` correctly classified it as `type-1 | actionable | unknown` (same identifier, same structure). The 2.5-month-old status-report flag was the key — without that, the duplication could plausibly be argued as "two domains, two types, leave it." The status report's nudge was the deciding factor. This is a case where art-dupl's report alone would have looked like a stylistic complaint, but combined with the standing TODO, it was clearly a real bug.

---

## Finding 4: `defer unsubscribe()` vs `defer unsubscribeHistory()` (ACCEPTED)

### What art-dupl matched

```go
// handlers/chat_sse.go:119
defer unsubscribe()
```

```go
// handlers/chat_sse.go:222
defer unsubscribeHistory()
```

Two single-line `defer X()` calls. `--explain` classified as `type-2 | actionable | defer | 1 tokens, 1 lines` — a renamed-variable defer pattern.

### Why it was a false positive

The two calls are in **different functions** of the same file:

| Line | Function | Resource | Source of `unsubscribe` |
| ---- | -------- | -------- | ----------------------- |
| 119 | `(*ChatHandler).Stream` | `ssehub.Hub` subscription (channel `chat:<chatID>`) | `h.hub.Unsubscribe(hubCh)` |
| 222 | `(*ChatHandler).chatStreamIdle` | `ChatHistoryManager` subscription | `h.chatHistory.Unsubscribe(...)` |

The `defer` keyword is the same; the `unsubscribe` identifier is different; the underlying resources are unrelated (one is the pub/sub hub, the other is the chat history). Extracting a helper would be impossible without making the code worse:

```go
// Hypothetical "extraction" — worse than the original
deferCallAtEndOfScope(unsubscribe)  // saves nothing, hides the defer
```

The `defer` statement IS the helper. There is no further abstraction to extract.

### Decision: `//art-dupl:accept` directives

```go
// handlers/chat_sse.go:119-120
//art-dupl:accept defer of hub subscription cleanup; defer unsubscribeHistory at chatStreamIdle:224 is a separate resource (chat history) and the `defer` keyword is a fundamental Go idiom.
defer unsubscribe()
```

```go
// handlers/chat_sse.go:223-224
//art-dupl:accept defer of chat history subscription cleanup; defer unsubscribe at Stream:120 is a separate resource (hub) and the `defer` keyword is a fundamental Go idiom.
defer unsubscribeHistory()
```

Cross-references the line numbers so a future reader (or a future lint tool) can verify the two sites are still in disagreement, not silently agreeing.

### Tool feedback

This is the textbook case for `//art-dupl:accept`: the tool is technically correct (the AST shape matches), but the duplication is not actionable. The directive is the right escape hatch. The `--no-accept-directives` flag correctly re-exposes the clone if a reviewer wants to audit deliberate suppressions.

**Suggestion for art-dupl:** the `defer` keyword + a single identifier call is the same pattern as `return X` and `if err != nil { ... }` — it's a fundamental Go idiom whose classification should be `defer-cleanup-of-arbitrary-resource` and **default-suppressed** unless the cleanup function is the same identifier AND the resources are the same. Currently this clone slipped through to the actionable report because the suffix (`unsubscribe` vs `unsubscribeHistory`) was almost-but-not-quite the same — a wholly superficial match. A more semantic check would also require the call expressions to be identical.

---

## Verification

```bash
$ art-dupl --type-aware --sort total-tokens -t 1
# (182 files, type-aware mode, type checking enabled)
Found total 0 clone groups.

$ art-dupl --type-aware --sort total-tokens -t 5
Found total 0 clone groups.

$ art-dupl --type-aware --sort total-tokens -t 10
Found total 0 clone groups.

$ art-dupl --type-aware --sort total-tokens -t 15
Found total 0 clone groups.

$ art-dupl --type-aware --sort total-tokens -t 1 --no-accept-directives
# (the deliberate accept directive is re-exposed)
found 2 clones:
  handlers/chat_sse.go:120-120  | defer unsubscribe()
  handlers/chat_sse.go:224-224  | defer unsubscribeHistory()
Found total 1 clone groups.

$ GOEXPERIMENT=jsonv2 go build ./...
# (clean)

$ GOEXPERIMENT=jsonv2 go vet ./...
# (clean)

$ GOEXPERIMENT=jsonv2 go test ./...
ok  	github.com/larsartmann/KeyHolderAI/config          0.053s
ok  	github.com/larsartmann/KeyHolderAI/database         0.005s
ok  	github.com/larsartmann/KeyHolderAI/dispatch          0.447s
ok  	github.com/larsartmann/KeyHolderAI/errors            0.010s
ok  	github.com/larsartmann/KeyHolderAI/handlers          4.018s
ok  	github.com/larsartmann/KeyHolderAI/httputil          0.005s
ok  	github.com/larsartmann/KeyHolderAI/metrics           0.008s
ok  	github.com/larsartmann/KeyHolderAI/middleware        0.009s
ok  	github.com/larsartmann/KeyHolderAI/models            0.008s
ok  	github.com/larsartmann/KeyHolderAI/personas          0.030s
ok  	github.com/larsartmann/KeyHolderAI/pkg/logger        0.035s
ok  	github.com/larsartmann/KeyHolderAI/pkg/testhelpers   0.059s
ok  	github.com/larsartmann/KeyHolderAI/repositories      0.005s
ok  	github.com/larsartmann/KeyHolderAI/routes            0.006s
ok  	github.com/larsartmann/KeyHolderAI/services          2.085s
ok  	github.com/larsartmann/KeyHolderAI/ssehub            0.506s
ok  	github.com/larsartmann/KeyHolderAI/types             0.006s
ok  	github.com/larsartmann/KeyHolderAI/validation        0.009s
ok  	github.com/larsartmann/KeyHolderAI/version           0.005s
```

All 21 test packages pass. The skipped `TestRequestID` is reported as `SKIP` (not removed) so the test inventory stays stable.

---

## Patterns observed across this session

Three observations about art-dupl's `-t 1` behavior that this session reinforced:

1. **Threshold 1 reports are useful but high-noise.** Of 4 groups, 3 were real (1 test rig, 1 stdlib idiom, 1 split-brain type), 1 was a false positive (defer pattern). At threshold 5, all 4 collapsed to 0. The "right" threshold for a working session is somewhere between 1 and 5 — `-t 1` for the catch-up pass, `-t 5+` for the signal check.

2. **`--type-aware` matters.** Without type-aware mode, the `RequestID string` clones would have been attributed to identifier-name equality alone. With type-aware mode, art-dupl correctly classified the type clone as `type-1` (structurally identical) versus the defer clone as `type-2` (renamed variable). The two are fundamentally different categories and deserved different treatment.

3. **Cross-referencing status reports is essential.** The `RequestID` split-brain was flagged in May 2026 and re-flagrd in August by art-dupl. The art-dupl report alone would have been easy to dismiss ("two distinct domains, two types, leave it"). The status report's standing decision ("pick one") was the deciding factor. Future feedback should always cross-reference pre-existing project context when the report is ambiguous.

---

## Recommended improvements for art-dupl (low priority)

1. **Suppress `defer X()` patterns by default** when `X` is a function call (not a method call) where the called function is a single identifier. The clone is a Go language idiom; the actionable report should focus on real duplication, not syntactic coincidence.

2. **Detect "same setup with different parameters"** in test code. The PlayFlow case was structurally identical code with three differing parameters (AI, timeout, channel buffer) — a future improvement could suggest a parameterizable helper rather than just reporting the literal clones.

3. **Surface "stdlib idiom available" suggestions** when the matched pattern matches a known stdlib helper (e.g., `cmp.Or` for `if x == "" { x = default }`). This requires per-project awareness of the Go version, which art-dupl could infer from `go.mod`.

4. **Default to suppress `if x == ""` patterns where the body is a single assignment.** This is the most-argued-over 3-line pattern in Go; almost every codebase has dozens of these and they are categorically idiomatic. A future improvement could classify them as `category: empty-default` and either suppress or down-rank.

None of these are urgent. The current report (4 groups / 8 occurrences) was actionable in a single session with manual judgment, and the cross-cutting lesson is "the `--no-accept-directives` flag is the audit mechanism for deliberate suppressions."
