# Feedback: `-t 1` threshold-1 "duplicates" can hide real split-brain bugs that require second-pass judgment

**Date:** 2026-08-07
**Project:** `storbi`, an Inventory Management System (Go + chi + CQRS via `go-cqrs-lite`, templ frontend, LibSQL/Turso)
**Command:** `art-dupl --type-aware --sort total-tokens -t 1`
**Goal:** Drive the report to zero harmful duplication, with the user's mandate "GET IT DOWN TO ZERO".

> **Verdict:** The report found **6 clone groups / 12 occurrences** at threshold 1 and **0 clone groups** at threshold 5. Five of the six groups were genuine Go idioms that art-dupl correctly identified as clones but a human must judge as acceptable. The sixth group — two `type contextKey string` declarations in different packages — initially looked like the same idiom, but in context was a **real split-brain bug**: the two packages also each declared a `RequestIDKey` constant with **different string values** (`"requestID"` vs `"request_id"`), and both were **dead code** because nothing in the codebase actually set those keys. A naive "accept everything as idiom" pass would have missed the bug.

> **Status:** ADDRESSED (2026-08-07) — Removed the dead `RequestIDKey` / `WithRequestID` / `RequestIDFromContext` / `WithContext` / `GetRequestID` pair across `pkg/logger/context.go`, `pkg/logger/logger.go`, and `internal/middleware/middleware.go`. The remaining request-ID read path now uses `chi/middleware.GetReqID(ctx)` as the single source of truth (chi's `RequestID` middleware is what actually populates the context). After the fix, re-running art-dupl at `-t 1` reports **5 clone groups / 10 occurrences** — the surviving five are irreducible idioms.

---

## Results

| Group                      | Locations                                                                | Initial decision | Actual category                                                                               | Final decision |
| -------------------------- | ------------------------------------------------------------------------ | ---------------- | --------------------------------------------------------------------------------------------- | -------------- |
| `type contextKey string`   | `internal/middleware/middleware.go:164`, `pkg/logger/context.go:10`      | Accept (idiom)   | **Split-brain: each package also defines a `RequestIDKey` with a different value, both dead** | **Eliminate**  |
| `t.Helper()` (3 groups)    | `internal/testutil/{assertions,helpers,database,response_assertions}.go` | Accept (idiom)   | Test helper marker                                                                            | Accept         |
| `ctx := request.Context()` | `internal/ui/handler.go:36-37`, `internal/ui/handler.go:71-72`           | Accept (idiom)   | HTTP handler context extraction                                                               | Accept         |
| `query := ``               | `internal/repository/quantity_repository.go:55-63`, `:93-101`            | Accept (idiom)   | First line of different SQL queries (`ReserveQuantity` vs `CommitReservedQuantity`)           | Accept         |

**Files modified:**

- `pkg/logger/context.go` — removed `RequestIDKey`, `WithRequestID`, `RequestIDFromContext`; kept only `LoggerKey`, `FromContext`, `ToContext`.
- `pkg/logger/logger.go` — removed `WithContext(ctx)` from the `Logger` interface and the `slogger` implementation (it read the dead key and was never called anywhere).
- `internal/middleware/middleware.go` — removed the local `type contextKey`, the local `RequestIDKey = "requestID"`, and the unsafe `r.Context().Value(RequestIDKey).(string)`; `GetRequestID` now delegates to `chi/middleware.GetReqID(r.Context())`.

---

## Why the first pass would have missed the bug

art-dupl flagged `internal/middleware/middleware.go:164` and `pkg/logger/context.go:10` as the same `type contextKey string` clone. Both are private, single-line type declarations. On the surface they look exactly like the documented "context key idiom" — Go programmers regularly introduce one of these per package precisely to prevent collisions, so the standard advice is "leave it, sharing would be wrong."

A surface-level review accepts both. That would have been wrong here.

### What the duplicate was hiding

Each `contextKey` declaration was paired with a `RequestIDKey` constant. The constants had **different values**:

```go
// internal/middleware/middleware.go (BEFORE)
type contextKey string
const RequestIDKey contextKey = "requestID"

// pkg/logger/context.go (BEFORE)
type contextKey string
const RequestIDKey contextKey = "request_id"
```

Both keys were **never written to anywhere in the codebase**. Verified by:

```bash
$ grep -rn "WithRequestID\|RequestIDKey =\|context.WithValue.*RequestID" --include="*.go" \
    internal pkg cmd main.go
# (no matches outside the constant declarations themselves)
```

The only code path that ever populated a request ID was `chi/middleware.RequestID`, set up in `middleware.NewRouter()` at `internal/middleware/middleware.go:94`:

```go
router.Use(middleware.RequestID)  // chi's middleware, NOT ours
```

`chi/middleware.RequestID` writes to **`chi/middleware.RequestIDKey`** (a `ctxKeyRequestID int` constant with value `0`, see `vendor/github.com/go-chi/chi/v5/middleware/request_id.go:21`). It does not write to either of our local keys.

**Result:** every reader of either local key — `middleware.GetRequestID(r)`, `logger.WithContext(ctx)` — silently returned the empty string, because the value at that key was never set. The `logger.WithContext` method did `ctx.Value(RequestIDKey).(string)`, which is the exact "unsafe type assertion that panics if key is missing or holds wrong type" failure mode documented as fixed in `docs/status/2026-05-26_07-11_SDK-UPGRADE-AND-ARCHITECTURE-IMPROVEMENTS.md:120` (the "🔴 Unsafe Type Assertion in `GetRequestID` (FIXED in `13ca5b8`)" entry). That fix used a checked assertion `if id, ok := ...; ok`, but the underlying invariant — that the key gets set somewhere — was never actually satisfied. The fix masked the bug; it didn't eliminate it.

This is a class of issue art-dupl cannot detect directly: the clone is two lines of structurally identical code, but the **semantic** problem is in lines art-dupl does not match (the constant values, and the absence of writers).

---

## Finding 1: Split-brain `RequestIDKey` (ELIMINATED)

### Before

```go
// internal/middleware/middleware.go:163-176
// RequestIDKey is the context key for request ID.
type contextKey string

// RequestIDKey is the context key for tracking request IDs.
const RequestIDKey contextKey = "requestID"

// GetRequestID extracts the request ID from context.
func GetRequestID(r *http.Request) string {
    if id, ok := r.Context().Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}
```

```go
// pkg/logger/context.go:9-58
// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const LoggerKey contextKey = "logger"
const RequestIDKey contextKey = "request_id"   // <-- different value

func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, RequestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}
```

```go
// pkg/logger/logger.go:209-228
func (l *slogger) WithContext(ctx context.Context) Logger {
    if ctx == nil {
        return l
    }
    attrs := []slog.Attr{}
    if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
        attrs = append(attrs, slog.String("request_id", requestID))
    }
    if len(attrs) == 0 {
        return l
    }
    return l.With(attrs...)
}
```

### Why it was harmful

1. **Dead code on the read side.** No call site ever sets either `RequestIDKey`. The only `context.WithValue(...RequestIDKey...)` writer anywhere is `chi/middleware`, and it writes to a **different, third** key (`chi/middleware.RequestIDKey`). Every read silently returned `""`.
2. **Split-brain on the value.** If a future contributor did start setting one of these, the value `"requestID"` and the value `"request_id"` would never agree — a request-ID-aware logger would see no ID when middleware did set one, or vice versa, depending on which package they used.
3. **Two `type contextKey string` declarations** masquerading as the safe "one per package" idiom, when in reality there is only one context value (`LoggerKey`) that needs the protection. The `RequestIDKey` pair was a phantom.
4. **`logger.WithContext` was never called.** Verified with `grep -rn "WithContext" --include="*.go"` — only the definition and the interface declaration match. Removing it required deleting the method from the `Logger` interface in `pkg/logger/logger.go:29`.

### After

```go
// internal/middleware/middleware.go (GetRequestID now delegates to chi)
func GetRequestID(r *http.Request) string {
    return middleware.GetReqID(r.Context())
}
```

`chi/middleware.GetReqID(ctx)` reads the key that chi's `RequestID` middleware actually sets (`chi/middleware.RequestIDKey`, an `int` constant). Single source of truth, single read path, behavior matches what the middleware stack already does.

`pkg/logger/context.go` is now:

```go
package logger

import "context"

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const LoggerKey contextKey = "logger"

func FromContext(ctx context.Context) Logger { /* unchanged */ }
func ToContext(ctx context.Context, l Logger) context.Context { /* unchanged */ }
```

`logger.WithContext` and the interface method are deleted. Callers that want request-ID logging now extract the ID via `middleware.GetReqID(ctx)` and pass it as a `slog.Attr` to `logger.With(...)`. This keeps `pkg/logger` chi-free (the package never imports chi; it stays at the bottom of the dependency graph for `internal/middleware` to depend on).

### Verification

```bash
$ art-dupl --type-aware --sort total-tokens -t 1
# 5 clone groups / 10 occurrences (was 6 / 12)

$ art-dupl --type-aware --sort total-tokens -t 5
Found total 0 clone groups.

$ GOEXPERIMENT=jsonv2 go build ./...
# (clean)

$ GOEXPERIMENT=jsonv2 go test ./...
ok  internal/domain/item     0.004s
ok  internal/handlers        0.027s
ok  internal/repository      0.009s
ok  internal/services        0.003s
ok  internal/services/impl   0.003s
ok  pkg/errors               0.003s
FAIL pkg/types (TestItemID_String — pre-existing, unrelated to this work)
```

---

## Finding 2: `t.Helper()` in test helpers (ACCEPTED)

### What art-dupl matched

Three groups of one-line `t.Helper()` (and one variant `td.t.Helper()`) markers inside `internal/testutil/{assertions,helpers,database,response_assertions}.go`. Each occurrence is a single statement at the top of a test-helper function, e.g.:

```go
// internal/testutil/assertions.go:184-193
func AssertError(t *testing.T, err error, msg string) error {
    t.Helper()
    if err == nil {
        t.Fatalf("%s: expected error, got nil", msg)
    }
    return fmt.Errorf("%s: %w", msg, err)
}
```

```go
// internal/testutil/helpers.go:70-77
func MustError(t *testing.T, err error, msg string) {
    t.Helper()
    if err == nil {
        t.Fatalf("%s: expected error but got nil", msg)
    }
}
```

### Why it is not harmful

`t.Helper()` is a single-statement marker with no logic. There is no "shared setup" to extract — each helper is doing genuinely different work (different assertion, different error message, different domain). Wrapping the `t.Helper()` line in a shared utility would be net-negative:

- It would add an indirection with no parameterization
- It would diverge from Go's documented testing convention (every helper starts with `t.Helper()` as a flat statement)
- golangci-lint's `thelper` rule explicitly enforces this style

This is the canonical "go-cqrs-lite t2/t3 baseline" pattern referenced in `2026-07-29_go-cqrs-lite_t2_t3_baseline-split-brain-and-test-boilerplate-noise.md` — test boilerplate that the test framework requires and that no extraction can simplify.

---

## Finding 3: `ctx := request.Context()` in HTTP handlers (ACCEPTED)

### What art-dupl matched

`internal/ui/handler.go:36-37` (in `Dashboard`) and `internal/ui/handler.go:71-72` (in `Inventory`). Both methods open with:

```go
func (h *Handler) Dashboard(writer http.ResponseWriter, request *http.Request) {
    ctx := request.Context()
    // ... different body ...
}

func (h *Handler) Inventory(writer http.ResponseWriter, request *http.Request) {
    ctx := request.Context()
    // ... different body, conditionally sets filter.LowStock ...
}
```

### Why it is not harmful

This is the standard idiom for HTTP handlers in Go: extract the request context into a named local so the body can pass it to services, transaction wrappers, loggers, etc. The only variation between the two methods is what they do with `ctx` afterwards — one calls `ListItems`, the other calls `ListItems` after a query-string parse, the rest of the file has six more handlers (`CreateItem`, `ItemDetail`, `StockModal`, `CloseModal`) each with the same pattern.

Extracting `ctx := request.Context()` into a helper would not just be net-negative; it would be **anti-idiomatic** — `request.Context()` is the documented public entry point on `*http.Request` and there is no shorter spelling to call it through. Wrapping it in `h.ctx(request)` would add a frame and a name without removing any code.

Notably, the project's own `internal/middleware/middleware.go` follows the same convention in `GetRequestID`:

```go
func GetRequestID(r *http.Request) string {
    return middleware.GetReqID(r.Context())
}
```

---

## Finding 4: `query := `` first line of different SQL updates (ACCEPTED)

### What art-dupl matched

`internal/repository/quantity_repository.go:55-63` (`ReserveQuantity`) and `:93-101` (`CommitReservedQuantity`). Both start with a multi-line raw string assigned to `query`:

```go
// ReserveQuantity
query := `
    UPDATE items
    SET quantity = CASE WHEN (quantity - reserved_quantity) >= ? THEN quantity ELSE quantity END,
        reserved_quantity = CASE WHEN (quantity - reserved_quantity) >= ? THEN reserved_quantity + ? ELSE reserved_quantity END,
        updated_at = ?
    WHERE id = ?
`

// CommitReservedQuantity
query := `
    UPDATE items
    SET reserved_quantity = CASE WHEN reserved_quantity >= ? THEN reserved_quantity - ? ELSE 0 END,
        quantity = CASE WHEN reserved_quantity >= ? THEN quantity - ? ELSE quantity END,
        updated_at = ?
    WHERE id = ?
`
```

### Why it is not harmful

The match is on the literal byte sequence `query :=` `` (eight characters: backtick-newline). art-dupl is correctly identifying that the same opening token appears in both functions, but the SQL bodies are completely different:

- `ReserveQuantity` touches both `quantity` and `reserved_quantity`, branches on `(quantity - reserved_quantity) >= ?`, and uses three placeholders plus the standard two-suffix pair.
- `CommitReservedQuantity` decrements `reserved_quantity` and conditionally decrements `quantity`, branches on `reserved_quantity >= ?`, and uses four placeholders plus the standard two-suffix pair.

Both call `r.execQuantityUpdate(ctx, query, itemID, quantity, true, ...)` with different argument shapes. The `query := `` opening is not duplicated logic; it is the language's syntax for declaring a string literal. art-dupl's type-aware mode treats multi-line raw strings as opaque AST nodes, so it cannot tell the SQL bodies apart — at threshold 1, the matching opening token is enough to flag a "clone." A human reviewing the diff sees two unrelated queries.

This is the same class of false positive as the boolean-initializer case (`2026-08-05_go-humanize-linter_t1_bool-initializer-false-positive.md`): art-dupl correctly detects surface similarity; the correct response is **Accept** with rationale, not extraction.

---

## Lessons for `-t 1` reviews

1. **The threshold-1 report is a checklist, not a verdict.** It catches true positives (`t.Helper()` clusters, dead type aliases) and false positives (boolean initializers, query-string openers, `ctx := r.Context()`) in roughly equal measure. Every group requires human judgment.

2. **Idiom clones can hide real bugs when the duplicated surface sits next to non-duplicated content.** `type contextKey string` is the textbook idiomatic private type. In `storbi`, it sat next to a `RequestIDKey` constant whose value disagreed across the two copies, and neither copy was ever written to. art-dupl cannot see the constant value or the absence of writers — the human must look past the matched lines.

3. **Two unrelated idioms are not the same idiom.** `type contextKey string` per package is correct _when the package actually owns a context value_. When the package doesn't (the case in `internal/middleware` once `RequestIDKey` was eliminated), keeping the type declaration is cargo-culted duplication: there is nothing left for it to protect.

4. **Check for writers before accepting a read-only helper.** `WithRequestID`, `WithContext`, `GetRequestID`, `RequestIDFromContext` were all readers with no matching writers. The grep for `context.WithValue.*RequestIDKey` returning zero hits was the single decisive check that turned "Accept" into "Eliminate." This should be the first question on every "leave it as idiom" decision: does anything write the value these helpers read?

5. **Status-report "fixed" entries can paper over live bugs.** `docs/status/2026-05-26_07-11_SDK-UPGRADE-AND-ARCHITECTURE-IMPROVEMENTS.md:120` celebrated a fix for "Unsafe Type Assertion in `GetRequestID`" by replacing the panic-prone assertion with a checked one. That fix was correct as far as it went, but the deeper invariant — _that the key gets set_ — was already broken. Status reports are point-in-time snapshots; the duplicate detected today is the truth that report was missing.

---

## Process notes

- Iterated from `t 5` (0 groups) down to `t 1` (6 groups), accepting five and eliminating one.
- The five accepted groups are documented with one-line rationales above; the eliminated group is a split-brain bug that the threshold-1 report surfaced only as a single-line `type contextKey string` match.
- All edits verified with `GOEXPERIMENT=jsonv2 go build ./...` and `GOEXPERIMENT=jsonv2 go test ./...`. The one failing test (`pkg/types.TestItemID_String`) is pre-existing and unrelated — see `internal/ui/handler.go` and `pkg/types/ids_test.go:43` for the divergence (output `"Item:test-123"` vs expected `"test-123"`), caused by the upstream `go-branded-id` library format change.
- No `trash`, `git reset`, `git checkout`, or `git restore` was used. All edits were surgical `edit`/`write` calls preceded by `view` reads of the affected regions.

---

## Resolution (2026-08-10)

**PARTIALLY ADDRESSED.** Several new actionability patterns shipped (defer-call, empty-default, state-flag-mutation). Remaining idiom patterns → TODO_LIST.
