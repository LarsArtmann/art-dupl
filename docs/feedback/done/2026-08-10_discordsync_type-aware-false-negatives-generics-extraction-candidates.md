# Feedback: `--type-aware` correctly suppresses 28/31 clone groups but hides 3 real duplicates (same-algorithm-different-element-type loops + cross-package logic twin)

**Date:** 2026-08-10
**Project:** `DiscordSync`, a Go 1.26 Discord backup/archiving daemon (CQRS + SQLite/Turso + templ dashboard, ~878 Go files)
**Commands compared:**
- `art-dupl --sort total-tokens -t 1` (no type filter) → **31 clone groups**
- `art-dupl --type-aware --sort total-tokens -t 1` → **0 clone groups** (all 31 suppressed)

**Goal:** Determine whether `--type-aware` is producing false-negatives (over-suppression) on a large production codebase, since switching to `--type-aware` by default would hide every group the plain scan surfaced.

---

## TL;DR

Ran a full judgment walk of all **31 clone groups** from the non-type-aware scan against the `DiscordSync` codebase. The `--type-aware` filter is **correct for 28 of 31 groups (90%)** — those are genuine false-positives where identical Go syntax is operating on different types for different semantic purposes (error-handling boilerplate, entity-load guards, env-var parsing idioms, etc.).

However, `--type-aware` hides **3 groups that represent real, eliminable duplication** (10% false-negative rate):

1. **`maxXxxCount` / sum loops (highest value)** — 4-6 near-identical "find max int64 field over slice" / "sum int64 field over slice" functions across 3 files, differing only in the element type and field accessor. Go generics (`maxValue[T]` / `sumValue[T]`) collapse them to 2 functions + call sites. `--type-aware` suppresses them because each function takes a different concrete slice type.

2. **Kind derivation logic (cross-package logic twin)** — Identical 8-line `if kind == "" { derive from IsBot }` block duplicated across `internal/db` and `internal/projection` operating on different struct types (`db.User` vs `events.UserPayload`). The algorithm is type-agnostic but the struct types differ, so `--type-aware` suppresses the match.

The core issue: **`--type-aware` treats different concrete types as "different semantics" even when the types are structurally identical (same field name, same field type) and the algorithm is type-agnostic.** This is correct for the majority of cases (preventing false merges) but systematically misses the class of "same algorithm over same-shaped data, different named type" — which is exactly what Go generics were designed to eliminate.

**Scorecard:** 28 correctly suppressed (true false-positives), 3 incorrectly suppressed (real duplicates hidden). 0 false-positives in the remaining `--type-aware` output (because it showed nothing).

---

## Detection Tally

### Full scan (no type filter)

```
art-dupl --sort total-tokens -t 1
→ 31 clone groups, 82 clones total
```

### Type-aware scan

```
art-dupl --type-aware --sort total-tokens -t 1
→ 0 clone groups
```

### Group-by-group verdict

| #   | Clone shape                                                | Locations (abbreviated)                                | Verdict               | `--type-aware` correct? |
| --- | ---------------------------------------------------------- | ------------------------------------------------------ | --------------------- | ----------------------- |
| 1   | `if err != nil { queryRowErr(...) }` boilerplate × 7      | `internal/db/*.go`                                     | **Correctly filtered** | Yes                     |
| 2   | Same `if err != nil { queryRowErr }` × 5 in one file      | `internal/db/query_phase2_detail.go`                   | **Correctly filtered** | Yes                     |
| 3   | `var highest int64; for ... { if X > highest }` × 3       | `internal/web/activity_helpers.go`, `handlers_extras.go` | **FALSE NEGATIVE**     | **No — real dup hidden** |
| 4   | `if err != nil \|\| ch == nil { return nil }` × 3         | `internal/web/types.go`                                | **Correctly filtered** | Yes                     |
| 5   | `if err != nil { return nil }` detail-author guard × 3    | `internal/web/handlers_phase2_detail.go`               | **Correctly filtered** | Yes                     |
| 6   | `context.WithTimeout(ctx, 30*time.Second)` × 2            | `cmd/gallery/capture.go`, `visual_regression_test.go`  | **Correctly filtered** | Yes                     |
| 7   | `if result.GuildID != "" { ... paginateOverfetch }` × 5   | `internal/web/handlers_phase2.go`                      | **Correctly filtered** | Yes                     |
| 8   | `if X.CreatedAt.IsZero() { X.CreatedAt = time.Now() }` × 3 | `internal/db/entities.go`                             | **Correctly filtered** | Yes                     |
| 9   | `var total int64; for ... { total += X.Count }` × 2       | `internal/web/activity_helpers.go`, `attachment_analytics_helpers.go` | **FALSE NEGATIVE** | **No — real dup hidden** |
| 10  | `kind := user.Kind; if kind == "" { if IsBot ... }` × 2   | `internal/db/entities.go`, `internal/projection/users.go` | **FALSE NEGATIVE**     | **No — real dup hidden** |
| 11  | `if avatar == "" { return "" }` vs `if icon == ""` × 2    | `internal/provider/discord.go`                         | **Correctly filtered** | Yes                     |
| 12  | `resolveMentionUsers`/`resolveMentionChannels` map→slice × 2 | `internal/web/mentions.go`                          | **Correctly filtered** | Yes                     |
| 13  | `restFetchUser`/`restFetchGuild` error-wrapping tail × 2  | `internal/bot/bot_fetchers.go`                         | **Correctly filtered** | Yes                     |
| 14  | `envDuration`/`envInt` env-var-with-default × 2           | `internal/db/turso_sync.go`                            | **Correctly filtered** | Yes                     |
| 15  | `withDefaultDuration`/`withDefaultInt` × 2                | `internal/config/config_file.go`                       | **Correctly filtered** | Yes                     |
| 16  | `if X != "" { query += " WHERE ..." }` × 2                | `internal/api/handlers_dlq.go`, `internal/db/query_phase2_counts.go` | **Correctly filtered** | Yes |
| 17  | `http.NewServeMux()` × 2                                  | `internal/e2e/smoke_test.go`, `internal/gallery/server.go` | **Correctly filtered** | Yes                  |
| 18  | `for _, k := range kinds { ... k.Count ... }` × 2         | `internal/web/activity_helpers.go`, `helpers.go`       | **Correctly filtered** | Yes                     |
| 19  | `if ch != nil && ch.Name != ""` vs `if g != nil && g.Name` × 2 | `internal/web/types.go`                           | **Correctly filtered** | Yes                     |
| 20  | `if err != nil { queryRowErr }` embed_media vs embeds × 2 | `internal/db/embed_media.go`, `embeds.go`              | **Correctly filtered** | Yes                     |
| 21  | `if err != nil { queryRowErr }` member detail × 2         | `internal/db/query_member_detail.go`                   | **Correctly filtered** | Yes                     |
| 22  | `downloadUserAvatar`/`downloadGuildIcon` refresh tail × 2 | `internal/bot/avatars.go`, `icons.go`                  | **Correctly filtered** | Yes (also `//nolint:dupl`) |
| 23  | `if err != nil { return defaultValue }` env parse × 2     | `internal/db/turso_sync.go`                            | **Correctly filtered** | Yes                     |
| 24  | `args := make([]any, 0, len(X)+1)` × 2                    | `internal/db/query_phase2.go`, `projection/emojis_relational.go` | **Correctly filtered** | Yes          |
| 25  | Error-chain-walk `for unwrapped := err; ...` (prod vs test) × 2 | `internal/storage/encryption.go`, `encryption_metrics_test.go` | **Correctly filtered** | Yes      |
| 26  | `if pct < 1 { pct = 1 }` vs `if offset < 0 { offset = 0 }` × 2 | `internal/web/activity_helpers.go`, `handler_helpers.go` | **Correctly filtered** | Yes   |
| 27  | `if err != nil \|\| user == nil` vs `if err != nil \|\| guild == nil` × 2 | `internal/web/handlers_attachments.go` | **Correctly filtered** | Yes               |
| 28  | `if ch.ParentID() != nil` vs `if thread.ParentID() != nil` × 2 | `internal/bot/discordadapter/misc.go`            | **Correctly filtered** | Yes                     |
| 29  | `for id := range ids { keys = append(keys, id) }` × 2     | `internal/web/mentions.go`                             | **Correctly filtered** | Yes                     |
| 30  | `if att.URL == ""` vs `if media.URL == ""` repair skip × 2 | `internal/bot/integrity_repair.go`                    | **Correctly filtered** | Yes                     |
| 31  | `if !content.IsAttachmentGoneError(err)` avatar vs icon × 2 | `internal/bot/avatars.go`, `icons.go`                | **Correctly filtered** | Yes                     |

---

## Finding 1 (FALSE NEGATIVE): `maxXxxCount` / sum loops — same algorithm, different element type

### What art-dupl matched (without `--type-aware`)

Three "find the highest int64" functions, plus a 4th using `.TotalSize` instead of `.Count` that wasn't even matched:

```go
// internal/web/activity_helpers.go:141
func maxAuthorKindCount(kinds []db.AuthorKindActivity) int64 {
	var highest int64
	for _, k := range kinds {
		if k.Count > highest {
			highest = k.Count
		}
	}
	return highest
}

// internal/web/activity_helpers.go:167
func maxMemberGrowth(growth []db.MemberGrowthPoint) int64 {
	var highest int64
	for _, g := range growth {
		if g.Count > highest {
			highest = g.Count
		}
	}
	return highest
}

// internal/web/activity_helpers.go:179
func maxStorageGrowth(growth []db.StorageGrowthPoint) int64 {
	var highest int64
	for _, g := range growth {
		if g.TotalSize > highest {
			highest = g.TotalSize
		}
	}
	return highest
}

// internal/web/handlers_extras.go:305
func maxReactionCount(reactions []db.TopReaction) int64 {
	var highest int64
	for _, r := range reactions {
		if r.Count > highest {
			highest = r.Count
		}
	}
	return highest
}
```

### Why `--type-aware` suppresses it

Each function iterates over a **different concrete slice type**: `[]db.AuthorKindActivity`, `[]db.MemberGrowthPoint`, `[]db.StorageGrowthPoint`, `[]db.TopReaction`. The type-aware filter sees 4 different type signatures and treats them as semantically distinct.

### Why this is a false-negative

The **algorithm is identical** and the **accessed fields are structurally identical** (all are `int64` fields named `.Count` or `.TotalSize`). This is exactly the use case Go generics solve:

```go
func maxValue[T any](items []T, get func(T) int64) int64 {
	var highest int64
	for _, item := range items {
		if v := get(item); v > highest {
			highest = v
		}
	}
	return highest
}
```

This collapses 4 × 8-line functions (32 lines) into 1 × 9-line generic + 4 one-line call sites (~13 lines). The duplication is real, harmful (a 5th variant was already being copy-pasted for `maxStorageGrowth`), and Go-idiomatic to eliminate.

### What `--type-aware` would need to detect this

The filter would need to recognize that the **range loop body is type-parametric** — i.e., the only type-dependent operation is a single field access whose result type is identical (`int64`) across all variants. Two possible heuristics:

1. **Structural field-access normalization:** when the loop body is `if <expr>.<field> > <accumulator>`, normalize `<expr>.<field>` to `SELECTOR(int64)` and compare the normalized AST. All 4 variants normalize to the same shape.
2. **"Same algorithm, different container" detection:** if N functions share the same control-flow shape (`var acc; for _, x := range slice { acc <op> x.field }; return acc`) and the only difference is the slice element type, flag as a potential generics-extraction candidate.

---

## Finding 2 (FALSE NEGATIVE): Sum loops — the summation twin of Finding 1

### What art-dupl matched (without `--type-aware`)

```go
// internal/web/activity_helpers.go:191
func totalAuthorKindCount(kinds []db.AuthorKindActivity) int64 {
	var total int64
	for _, k := range kinds {
		total += k.Count
	}
	return total
}

// internal/web/attachment_analytics_helpers.go:167 (inside computeDonutSegments)
var total int64
for _, c := range cats {
	total += c.Count
}
```

### Why `--type-aware` suppresses it

Different element types: `[]db.AuthorKindActivity` vs `[]db.AttachmentCategoryStat`.

### Why this is a false-negative

Same algorithm class as Finding 1 (accumulate an `int64` field over a slice), just with `+=` instead of `>` comparison. The generic fix:

```go
func sumValue[T any](items []T, get func(T) int64) int64 {
	var total int64
	for _, item := range items {
		total += get(item)
	}
	return total
}
```

### Relationship to Finding 1

Findings 1 and 2 are the **same root cause**: a family of reduce-style loops over typed slices where only the element type and field accessor differ. If `--type-aware` could detect either, it would likely detect both, since they share the "same control flow, different container element type" shape.

---

## Finding 3 (FALSE NEGATIVE): Cross-package kind-derivation logic twin

### What art-dupl matched (without `--type-aware`)

```go
// internal/db/entities.go:131 (inside UpsertUser)
kind := user.Kind
if kind == "" {
	if user.IsBot {
		kind = domain.UserKindBot
	} else {
		kind = domain.UserKindHuman
	}
}

// internal/projection/users.go:23 (inside ensureUserWithKind)
kind := user.Kind
if kind == "" {
	if user.IsBot {
		kind = domain.UserKindBot
	} else {
		kind = domain.UserKindHuman
	}
}
```

### Why `--type-aware` suppresses it

`entities.go` operates on `*db.User`; `users.go` operates on `events.UserPayload`. Different struct types, so the type-aware filter sees them as semantically distinct.

### Why this is a false-negative

The **logic is byte-for-byte identical** and **completely type-agnostic** — it only reads two fields (`.Kind string` and `.IsBot bool`) that exist on both types with identical names and types. The branch conditions, the constant lookups (`domain.UserKindBot`, `domain.UserKindHuman`), and the assignment are all identical. This is not "similar code on different types" — it is **the same 8 lines copy-pasted** across a package boundary.

The fix extracts to the `domain` package (where the constants live):

```go
func ResolveUserKind(kind string, isBot bool) string {
	if kind != "" {
		return kind
	}
	if isBot {
		return UserKindBot
	}
	return UserKindHuman
}
```

Both call sites become `kind := domain.ResolveUserKind(user.Kind, user.IsBot)`.

### What `--type-aware` would need to detect this

The filter would need to recognize that the cloned block's **field accesses resolve to the same underlying types** across both variants — `.Kind` is `string` on both `db.User` and `events.UserPayload`; `.IsBot` is `bool` on both. When the accessed fields are structurally identical (same name → same type), different container types should NOT suppress the clone. This is a "duck-typed field-access equivalence" check.

---

## What `--type-aware` got right (28 groups correctly suppressed)

The 28 correctly-suppressed groups fall into these categories:

### Error-handling boilerplate (groups 1, 2, 12, 13, 20, 21, 23)

```go
if err != nil {
    return nil, queryRowErr(err, "failed to get X", "id", string(id))
}
```

`queryRowErr` IS the shared helper. The `if err != nil` wrapper is irreducible Go. Each site calls a different `Scan()` target with different context keys. Extracting a generic `scanAndQueryRow[T]` would add indirection for zero readability gain — Go's error handling is verbosely typed by design.

### Entity-load-then-nil-guard (groups 4, 5, 27)

```go
ch, err := database.GetChannel(ctx, channelID)
if err != nil || ch == nil {
    return nil
}
return ch
```

Different DB methods (`GetChannel`, `GetThread`, `GetGuild`, `GetUser`) returning different types (`*db.Channel`, `*db.Thread`, etc.). A generic `resolveEntity[T]` wouldn't simplify because the `db.Database` interface has 183 named methods, not a generic loader. The 3-line guard is the irreducible form.

### Avatar/icon mirror pairs (groups 22, 31)

Already acknowledged with `//nolint:dupl` — structurally similar but semantically distinct (user avatar pipeline vs guild icon pipeline). Different Discord API entities, different event types, different DB tables.

### Phase2 handler pagination guard (group 7)

```go
if result.GuildID != "" {
    data.Pagination = paginateOverfetch(result.Items, result.Offset, phase2Page)
}
```

Each of the 5 occurrences constructs a **different ViewModel type** (`PresencesViewModel`, `InteractionsViewModel`, etc.) and calls a **different row renderer** (`presencesRows`, `interactionsRows`, etc.). The 2-line guard is the only shared part.

### Coincidental short idioms (groups 14, 15, 17, 24, 26, 28, 29)

- `os.Getenv` + parse + fallback default (duration/int variants)
- `http.NewServeMux()` (2-line setup in different packages)
- `make([]any, 0, len(x)+1)` (SQL arg slice allocation)
- `if val < N { val = N }` (different fields, different thresholds)
- `ParentID() != nil` (channel vs thread, different disgo types)
- `for id := range ids { keys = append(keys, id) }` (map→slice, different key types)

Each is a 2-3 line Go idiom where extraction would add complexity, not reduce it. `--type-aware` correctly identifies that the types differ.

### Cross-file WHERE-clause coincidence (group 16)

`api/handlers_dlq.go` builds a SQL filter on `projection_name`; `db/query_phase2_counts.go` builds one on `guild_id`. Different tables, different packages, different filter columns. The 2-line `if X != "" { query += " WHERE ..." }` pattern is coincidental.

### Test mirrors production (groups 6, 25)

One clone is a test that copies production logic to verify it. Different intent, correct separation.

### CreatedAt zero-check (group 8)

```go
if guild.CreatedAt.IsZero() {
    guild.CreatedAt = time.Now()
}
```

3-line idiom on 3 different entity types (`Guild`, `Channel`, `User`). Extraction to a generic `setDefaultTime[T]` would be over-engineering for a 2-statement guard.

---

## Root-cause analysis: what pattern does `--type-aware` systematically miss?

The 3 false-negatives share a single root cause:

> **`--type-aware` treats different concrete container types as "different semantics" even when the algorithm is type-agnostic and the accessed fields are structurally identical.**

This is correct 90% of the time — most type-different clones ARE semantically different (error handling on different operations, guards on different entities, idioms on different types). But it systematically misses the "same algorithm, same-shaped data, different named type" class, which is precisely what Go generics exist to eliminate.

### Proposed detection improvement

A **"structural field-access equivalence"** pass could reduce false-negatives without increasing false-positives:

1. After the type-aware filter suppresses a clone group, perform a secondary check on the suppressed group.
2. For each field access in the cloned block (`x.Field`), resolve the field's type from both variants' type information.
3. If all corresponding field accesses resolve to the **same underlying type** (e.g., `.Count` is `int64` on both `AuthorKindActivity` and `MemberGrowthPoint`), the clone is a **generics-extraction candidate** — re-surface it with a hint like `"same algorithm over structurally-identical fields; consider generics extraction"`.

This would catch Findings 1, 2, and 3 without re-surfacing the 28 correctly-suppressed groups (where the field types or algorithms genuinely differ).

### Alternative: a `--suggest-generics` flag

Rather than changing `--type-aware`'s default behavior (which is well-calibrated for false-positive reduction), a separate `--suggest-generics` flag could run the structural-equivalence pass as an opt-in mode. This lets users who are specifically looking for generics-extraction opportunities find them, without surprising users who just want "show me the real duplicates."

---

## Summary

| Metric | Value |
| ------ | ----- |
| Clone groups (no type filter) | 31 |
| Clone groups (`--type-aware`) | 0 |
| Correctly suppressed by `--type-aware` | 28 (90%) |
| Incorrectly suppressed (false-negatives) | 3 (10%) |
| False-positives in `--type-aware` output | 0 (vacuously — it showed nothing) |
| Highest-value missed fix | `maxValue[T]` / `sumValue[T]` generics (6 functions → 2) |
| Root cause | Type-aware filter treats different container types as different semantics, even when fields are structurally identical |

**Bottom line:** `--type-aware` is well-calibrated for false-positive reduction but has a systematic blind spot for generics-extraction candidates. The blind spot is small (10% of real duplicates in this codebase) but high-value (the missed groups are the ones most amenable to clean elimination). A structural field-access equivalence pass — either integrated or as an opt-in `--suggest-generics` flag — would close the gap.

---

## Update (2026-08-10): `--suggest-generics` implemented and validated E2E

The `--suggest-generics` flag proposed above was implemented in art-dupl (commit `fdbaccce`, "feat(syntax/golang): add EraseHash mode for generics-extraction detection"). Running it against DiscordSync validates the original analysis:

### Before any fixes

```
art-dupl --type-aware --suggest-generics -t 1 .
→ 24 clone groups
```

All 3 false-negatives from the original analysis are surfaced:

| Original finding | art-dupl `--suggest-generics` output | Clones |
| --- | --- | --- |
| Finding 1: `maxXxxCount` loops | `generics: same algorithm, different types: AuthorKindActivity vs TopReaction vs MemberGrowthPoint` at `activity_helpers.go:142`, `:168`, `handlers_extras.go:289` | 3 |
| Finding 2: `sumValue` loops | `generics: same algorithm, different types: AttachmentCategoryStat vs AuthorKindActivity` at `activity_helpers.go:192`, `attachment_analytics_helpers.go:167` | 2 |
| Finding 3: kind-derivation logic twin | `generics: same algorithm, different types: events.UserPayload vs *db.User` at `entities.go:131`, `projection/users.go:23` | 2 |

The remaining 19 groups are all the correctly-identified generics-extraction candidates for irreducible boilerplate (error-handling guards, named-method calls, intentional mirror pairs) — these are technically "same algorithm, different types" but not worth extracting.

### Fixes applied to DiscordSync

All 3 false-negatives were eliminated:

1. **`maxValue[T any]` generic** (`internal/web/activity_helpers.go`) — replaces `maxAuthorKindCount`, `maxMemberGrowth`, `maxStorageGrowth`, `maxReactionCount` (4 functions → 1 generic + 4 one-line call sites).

2. **`sumValue[T any]` generic** (`internal/web/activity_helpers.go`) — replaces `totalAuthorKindCount` and the inline sum in `computeDonutSegments` (2 loops → 1 generic + 2 one-line call sites).

3. **`domain.ResolveUserKind(kind UserKind, isBot bool) UserKind`** (`internal/domain/user_kind.go`) — replaces the kind-derivation logic in both `internal/db/entities.go:UpsertUser` and `internal/projection/users.go:ensureUserWithKind`. Also consolidated `events.DeriveUserKindFromBot` (a third copy of the same logic in `internal/events/payloads.go`) and its 3 upcaster call sites in `internal/eventschema/upcasters.go` — eliminating a pre-existing split brain the original analysis didn't catch.

### After fixes

```
art-dupl --type-aware --suggest-generics -t 1 .
→ 20 clone groups (down from 24)
```

The 4 eliminated groups correspond exactly to the 3 false-negatives (Finding 1 produced 2 clone groups in `--suggest-generics` output because art-dupl detected the `maxAuthorKindCount`/`maxMemberGrowth`/`maxReactionCount` group and the `maxStorageGrowth`/`totalAuthorKindCount` loop-overlap separately). The remaining 20 groups are all correctly-classified non-actionable boilerplate.

### Validation

- **Quality gate:** `nix run .#quality` → 0 issues (lint + fmt + file-size clean)
- **Tests:** `nix run .#test` → all 25 packages pass, 0 failures
- **`--suggest-generics` precision:** 24 groups reported, 3 actionable (12.5% precision), 21 correctly identified as non-actionable boilerplate. The 3 actionable groups were all real generics-extraction candidates.

### What `--suggest-generics` got right

1. **All 3 false-negatives surfaced** — zero false-negatives in the `--suggest-generics` output.
2. **Type-difference metadata** is excellent for triage: `"AuthorKindActivity vs TopReaction vs MemberGrowthPoint (8 type differences total)"` immediately tells you whether the clone is worth extracting (few structural types, same algorithm) or not (many type differences, different semantics).
3. **The 21 non-actionable groups** are all genuinely "same algorithm, different types" — they're just not worth extracting. This is the correct trade-off: better to over-report and let the human filter than to miss the 3 real candidates.

### What could improve

The 21 non-actionable groups are dominated by two patterns that could be suppressed:

1. **Error-handling guards** (`if err != nil { return ... }`) — 7+ groups are `queryRowErr`-style boilerplate where the algorithm is "check error, wrap, return." These are structurally identical across Go codebases but extracting a generic doesn't help because Go's error handling is verbosely typed by design.
2. **2-line entity nil-guards** (`if err != nil || x == nil { return nil }`) — the guard itself is the irreducible form; the type differences (`*Channel` vs `*Guild` vs `*Thread`) are real and don't benefit from generics.

Suppressing these two patterns (e.g., by recognizing that the clone body is a single `if err != nil` branch or a single nil-guard) would reduce the noise-to-signal ratio from 21:3 to ~6:3, making the output much more actionable.

---

## Resolution (2026-08-10)

**ADDRESSED.** `--suggest-generics` feature shipped (CHANGELOG `[Unreleased]` → Added). All 3 false negatives recovered. Precision filtering (12.5% → target >50%) is in TODO_LIST HIGH priority.
