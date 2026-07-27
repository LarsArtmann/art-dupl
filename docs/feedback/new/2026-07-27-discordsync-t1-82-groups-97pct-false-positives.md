# Feedback: Threshold 1 surfaces 82 groups — 2 harmful, 80 language-idiom false positives

**Date:** 2026-07-27
**Project:** `DiscordSync`, a ~573-file Go application (Discord archiving bot with CQRS event sourcing, SQLite projections, web dashboard)
**Command:** `art-dupl --type-aware --sort total-tokens -t 1 --html`
**Goal:** Deduplicate until ZERO harmful duplication remains.

> **Verdict:** The report found **82 clone groups** at threshold 1. After walking every group, only **2 contained genuinely harmful duplication** (extracted: `invokeService[T]`, `addIfPositive64`). The remaining **80 groups (97.5%) are language-idiom false positives** — standard Go control flow (`if err != nil`, `defer cancel()`, `defer rows.Close()`), calls to already-extracted helpers, or AST-shape coincidences across unrelated domains. At `-t 1` the signal-to-noise ratio is extremely low. The tool would benefit enormously from classifying and suppressing/down-ranking these built-in Go idioms rather than reporting them as actionable clones.

## Results

| Category | Groups | Decision | Tool feedback |
| --- | --- | --- | --- |
| `if err != nil` + `writeError` + bare `return` (HTTP handler) | 12 | Accept | **Suppress: forced by `http.HandlerFunc` void signature** |
| `if err != nil` + `return X, queryError(err, uniqueOp)` (DB layer) | 27 | Accept | **Down-rank: unique operation strings are parameters** |
| `defer cancel()` / `defer rows.Close()` / `defer tx.Rollback()` | 12 | Accept | **Suppress: universal Go resource cleanup** |
| `t.Parallel()` (test boilerplate) | 8 | Accept | Already documented; confirm suppression works at `-t 1` |
| `ctx, cancel := helper()` (calling extracted helper) | 3 | Accept | **Down-rank: the call IS the extraction** |
| `if bool { return STR_A }; return STR_B` (different domains) | 3 | Accept | **False positive: same AST, different return values** |
| `fmt.Sprintf("#%06x", color)` vs `fmt.Sprintf("#%06X", color)` | 1 | Accept | **False positive: different format specifiers** |
| Calling shared helpers (`requireQueryParam`, `loadGuildsAndParams`) | 5 | Accept | **Down-rank: helper invocation, not duplication** |
| `if id == nil { return "" }; return id.String()` (3-line cross-package utils) | 1 | Accept | Borderline — too small to extract across packages |
| Single-site `if err != nil { log/slog }` one-liners | several | Accept | **Suppress: unique messages, single-line logging** |
| `.templ` render blocks (empty-state vs populated loop) | 4 | Accept | **Exclude: templ-generated AST noise** |
| **Genuinely harmful (extracted)** | **2** | **Extracted** | — |

---

## Finding 1 (HIGH PRIORITY): `if err != nil { writeError(w,r,err,msg); return }` cannot be extracted — forced by `http.HandlerFunc`

**12 groups** matched this shape across `internal/web/` and `internal/api/`:

```go
// internal/web/handlers.go:125
bundle, err := v.database.GetMessageBundle(r.Context(), messageID)
if err != nil {
    errorpage.WriteError(w, r, err, "")

    return
}
```

```go
// internal/web/handlers_extras.go:122
guilds, err := v.database.GetGuilds(r.Context())
if err != nil {
    errorpage.WriteError(w, r, err, "")

    return
}
```

### Why extraction is impossible

The bare `return` is **forced by Go's `http.HandlerFunc` signature** (`func(http.ResponseWriter, *http.Request)` — void return). A helper cannot issue `return` on behalf of its caller. The only "extraction" would be a callback wrapper:

```go
func withErrorHandler(w http.ResponseWriter, r *http.Request, err error, msg string, onFail func()) {
    if err != nil {
        errorpage.WriteError(w, r, err, msg)
        onFail() // still can't return from the caller!
        return
    }
}
```

…which still requires the caller to write `return` afterward, so it saves nothing and obscures control flow. This is not a design flaw in the code — it is a property of Go's HTTP handler contract.

### Tool feedback

**Suppress this pattern entirely** when the AST matches:

```
IfStmt {
    Cond: BinaryExpr(err != nil)
    Body: [
        ExprStmt(CallExpr{writeError | s.writeAppError | errorpage.WriteError, [w, r, err, ...]})
        ReturnStmt{}   // <-- EMPTY (no return values)
    ]
}
```

Classification: `go-http-error-guard`. Priority: **suppressed by default** at all thresholds, or informational-only. Suggestion: *"Forced by http.HandlerFunc void signature — cannot be extracted without a callback wrapper that still requires a bare return."*

---

## Finding 2 (HIGH PRIORITY): `if err != nil { return X, wrap(err, uniqueOp, uniqueCtx...) }` — unique strings are parameters, not duplication

**27 groups** matched in the DB layer (`internal/db/`):

```go
// internal/db/backfill_progress_stale.go:40
if err != nil {
    return nil, queryError(err, "failed to query stale backfill progress")
}
return progress, nil
```

```go
// internal/db/query.go:363
if err != nil {
    return nil, queryError(err, "failed to query backfill progress")
}
return progress, nil
```

### Why this is not harmful

Per the dedup skill's own guidance: *"Unique values are parameters, not duplication."* Every site calls a **different** operation with a **unique** error message (`"failed to query stale backfill progress"` vs `"failed to query backfill progress"` vs `"failed to get guild member"`...). The `queryError` / `errkit.Transient` wrapper **IS** the already-extracted helper. The return arity (`return nil, err` vs `return nil, nil, err` vs `return 0, err`) and not-found semantics (`isErrNoRows` guard present or absent) also differ per site.

### Tool feedback

**Down-rank** clones where the only variation is in string-literal arguments to a shared wrapper function. Detection heuristic: if the cloned block is `if err != nil { return <expr list containing call to known wrapper(err, stringLit, ...)> }`, classify as `go-error-wrap-idiom`, priority **informational**. The wrapper call itself is the deduplication; the unique strings are the convention.

---

## Finding 3 (HIGH PRIORITY): `defer cancel()` / `defer rows.Close()` / `defer tx.Rollback()` — universal Go resource cleanup

**12 groups** matched:

```go
// internal/bot/handlers_events.go:57 (and 3 more sites in same file)
ctx, cancel := newHandlerContext()
defer cancel()
```

```go
// internal/db/attachment_downloads.go:244 (and 3+ more sites)
defer func() { _ = rows.Close() }()
```

```go
// internal/db/messages.go:41
defer func() { _ = transaction.Rollback() }()
```

### Why extraction is harmful

`defer cancel()` after `ctx, cancel := ...` is the **most idiomatic pattern in Go** (every context acquisition). Extracting a `deferCancel(cancel)` helper would hide the defer from readers — the defer scope is a correctness property, not boilerplate. Same for `rows.Close()` and `tx.Rollback()`.

### Tool feedback

**Suppress three patterns by default**:

1. `defer cancel()` (or `defer cancelFunc()`) — classify as `go-context-cancel`, suppressed.
2. `defer func() { _ = rows.Close() }()` — classify as `go-rows-close`, suppressed.
3. `defer func() { _ = tx.Rollback() }()` — classify as `go-tx-rollback`, suppressed.

These three alone would eliminate ~12 groups from this report.

---

## Finding 4 (MEDIUM): `if bool { return STR_A }; return STR_B` — same AST shape, completely different domains

**3 groups** matched three unrelated functions that happen to share the AST shape `if bool { return stringLit }; return stringLit`:

```go
// internal/web/handlers_phase2_detail.go:10
func fmtBool(b bool) string {
    if b { return "Yes" }
    return "No"
}

// internal/provider/discord.go:111
func emojiExtension(animated bool) string {
    if animated { return ".gif" }
    return ".png"
}

// internal/web/helpers_phase2.go:256
func banActiveSelectValue(activeOnly bool) string {
    if activeOnly { return "1" }
    return ""
}
```

### Why this is a false positive

Three completely different domains (UI yes/no formatting, Discord CDN file extension, HTML select value) with different return values (`"Yes"/"No"` vs `".gif"/".png"` vs `"1"/""`). The shared AST shape is a coincidence of Go's `if-else` expressiveness. Extracting a `boolToStr(b bool, ifTrue, ifFalse string) string` helper would be more lines than the duplicated code and would obscure each call site's intent.

### Tool feedback

**Type-aware mode should flag this as a false positive** when:
- The cloned block is a complete function (not a fragment),
- The function body is `if <bool param> { return <string literal A> }; return <string literal B>`,
- The string literals differ between clone instances.

Classification: `go-bool-to-string-func`. Priority: **suppressed** — this is a 3-line idiom, and the values are the domain logic.

---

## Finding 5 (MEDIUM): `fmt.Sprintf("#%06x", ...)` vs `fmt.Sprintf("#%06X", ...)` — different format specifiers produce different output

```go
// internal/web/handlers_embeddings.go:220 (embeds — lowercase hex)
return fmt.Sprintf("#%06x", color)

// internal/web/types.go:51 (roles — uppercase hex)
return fmt.Sprintf("#%06X", color)
```

### Why this is a false positive

`%06x` (lowercase) and `%06X` (uppercase) are **different format specifiers** producing different output (`#5b4bf7` vs `#5B4BF7`). CSS hex is case-insensitive, but the two functions live in unrelated domains (embed border color vs role dot color). The tool matched them because the rest of the function (`if color == 0 { return "" }; return fmt.Sprintf(...)`) is identical — but the format specifier is the meaningful part.

### Tool feedback

**Type-aware/semantic mode should detect differing string-literal format specifiers** in `fmt.Sprintf` calls and treat them as semantically distinct. Heuristic: if two `fmt.Sprintf` calls share a format string template but differ in case (`%06x` vs `%06X`) or verb (`%d` vs `%s`), they are **not** clones.

---

## Finding 6 (MEDIUM): Calling an already-extracted helper is not duplication

**5+ groups** are just calls to shared helpers:

```go
// 5 sites, all identical:
guildID, ok := s.requireQueryParam(w, r, "guild_id")
if !ok { return }
```

```go
// 3 sites:
guildID, offset, guilds, ok := v.loadGuildsAndParams(w, r)
if !ok { return }
```

### Why this is not harmful

`requireQueryParam` and `loadGuildsAndParams` **ARE** the extractions. Calling a shared helper N times is the *goal* of deduplication, not duplication itself. The `"guild_id"` string is intentionally repeated — every guild endpoint needs it, and it's the parameter name from the HTTP query string.

### Tool feedback

**Down-rank** clones that consist solely of a call to a shared function + a trivial guard (`if !ok { return }`). These are "helper invocations," not duplications. Heuristic: if the cloned fragment is `X, ok := <sharedHelper>(...); if !ok { return }`, classify as `go-helper-invocation`, priority **informational**.

---

## Finding 7 (LOW): 3-line cross-package utility — too small to warrant coupling

```go
// internal/bot/handlers_core.go:132 — package bot
func webhookIDStr(id *snowflake.ID) string {
    if id == nil { return "" }
    return id.String()
}

// internal/bot/discordadapter/message_extras.go:79 — package discordadapter
func optSnowflakeToString(id *snowflake.ID) string {
    if id == nil { return "" }
    return id.String()
}

// internal/bot/discordadapter/ids.go:34 — same package, different return type
func OptSnowflakeToChannelID(id *snowflake.ID) domain.ChannelID {
    if id == nil { return "" }
    // ...
}
```

### Why accepted

3-line nil-guard utilities in different packages. Extracting to a shared package would create a cross-package dependency for 3 lines. The third site returns a branded `domain.ChannelID` type, not `string` — so even the return type differs.

### Tool feedback

This is a **borderline true positive** — the first two ARE semantic clones. But at 3 lines with a nil-guard pattern that is itself an idiom, the tool could note "cross-package 3-line utility" and let the user decide. No suppression needed; just acknowledge in the report that sub-5-line functions sharing only a nil-guard are low-value.

---

## Finding 8 (LOW): `.templ` render blocks — generated AST noise

**4 groups** matched `.templ` source blocks like:

```templ
if len(channels) == 0 {
    No channels with messages.
} else {
    for _, ch := range channels {
        @barChartRow(ch.ChannelName, ch.Count, maxCount)
    }
}
```

### Why accepted

These are `.templ` source files (not generated `*_templ.go`). The "clones" are templ rendering blocks with different entities, different inner calls, and different empty-state messages. The AST shape (`if empty { text } else { loop }`) is a templ/HTML convention.

### Tool feedback

**Exclude `.templ` source files** from analysis by default (the same way generated `*_templ.go` files are excluded). The `.templ` format produces repetitive control-flow shapes that are rendering conventions, not logic duplication. Alternatively, if `.templ` analysis is valuable, develop templ-specific heuristics that recognize empty-state-vs-loop as a rendering idiom.

---

## Finding 9 (LOW): Single-site `if err != nil { slog.X(...) }` logging — unique messages

Several groups matched single-site error-logging:

```go
// internal/db/maintenance.go:24
if err != nil {
    slog.Error("failed ...", "error", err)
}

// internal/bot/attachments.go:77
if err != nil {
    logError("failed to mark attachment as failed", err, "attachment_id", attachmentID)
}
```

### Why accepted

Each site has a unique message, unique context keys, and unique severity (`Debug` vs `Error` vs `Warn`). A helper would need more parameters than the duplicated lines.

### Tool feedback

**Suppress** single-statement `if err != nil { <log call with unique msg> }` blocks when the log message string literal is unique per site. Classification: `go-error-log-one-liner`, suppressed.

---

## Summary: what suppressing these patterns would achieve

| Suppression rule | Groups eliminated | New report size |
| --- | --- | --- |
| (baseline) | 0 | 82 |
| Finding 1: http-error-guard | -12 | 70 |
| Finding 2: error-wrap-idiom (down-rank) | -27 (to informational) | 43 actionable |
| Finding 3: defer cleanup (3 rules) | -12 | 31 |
| Finding 4: bool-to-string-func | -3 | 28 |
| Finding 6: helper-invocation (down-rank) | -5 (to informational) | 23 |
| Finding 8: .templ source exclude | -4 | 19 |
| Finding 9: error-log-one-liner | -4 | 15 |
| **Total** | **~67 suppressed/down-ranked** | **~15 actionable** |

Suppressing these built-in Go idioms would take this report from **82 groups (2 harmful, 97.5% noise)** to **~15 actionable groups (2 harmful, 13 worth reviewing)** — a **5x improvement in signal-to-noise ratio** at `-t 1`.

The two genuinely harmful clones (`invokeService[T]` and `addIfPositive64`) would still be reported and extracted. No true positive would be lost.

---

## What worked well

- **Type-aware mode** correctly distinguished the `%06x` vs `%06X` case as worth flagging (even though it's a false positive, the format-specifier difference is the right thing to notice).
- **HTML report format** is excellent for triage — the clone-group cards with file:line links and syntax-highlighted snippets made walking 82 groups fast.
- **`--sort total-tokens`** correctly surfaced the highest-value clones first.
- **Generated `*_templ.go` exclusion** works — no generated files appeared in the report.
- **Threshold granularity** (`-t 1` through `-t N`) gives useful control over noise level.

## Environment

- **art-dupl version:** (latest, as of 2026-07-27)
- **Project size:** 573 Go files
- **Go version:** 1.26 with `GOEXPERIMENT=jsonv2`
- **Clone detection mode:** `--type-aware`
