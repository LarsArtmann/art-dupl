# Feedback: `--suggest-generics` precision validation — 3 real catches eliminated, 20 remaining groups exhaustively classified (19 non-actionable, 1 borderline)

**Date:** 2026-08-10
**Tool Version:** `art-dupl` v0.6.1 (`445588c3`)
**Tested On:** `DiscordSync` — Go 1.26 Discord backup/archiving daemon, ~879 Go files (841 with type information)
**Detection Mode:** `--type-aware --suggest-generics -t 1`
**Commits this session:** `6e82fea1` (ResolveUserKind extraction), `bb0b0c51` (maxValue/sumValue generics), `374ee64a` (consolidate DeriveUserKindFromBot into ResolveUserKind)

---

## TL;DR

The `--suggest-generics` flag (commit `fdbaccce`, "feat(syntax/golang): add EraseHash mode for generics-extraction detection") was validated end-to-end against DiscordSync. **All 3 real generics-extraction candidates were correctly surfaced and eliminated.** After the fixes, the scan produces **20 clone groups** — this report is the exhaustive per-group verdict on those 20, categorized by why they are not worth extracting.

**The core finding:** `--suggest-generics` has zero false-negatives (every real generics candidate was found) but a **19:1 noise-to-signal ratio on the remaining output**. The noise comes from 5 recurring patterns that are structurally "same algorithm, different types" but semantically irreducible — dominated by Go's `if err != nil` named-method boilerplate. Suppressing these patterns would make the flag dramatically more actionable.

**Scorecard:**

- Real generics-extraction candidates found: **3 of 3** (100% recall)
- False-negatives: **0**
- Remaining groups after fixes: **20** (19 non-actionable, 1 borderline)
- Precision of remaining output: **~5%** (1 of 20 marginally extractable)

---

## Detection Tally

### Phase 1: Initial `--suggest-generics` scan (before fixes)

```
art-dupl --type-aware --suggest-generics -t 1 .
→ 24 clone groups
```

### Phase 2: After fixing the 3 real candidates

```
art-dupl --type-aware --suggest-generics -t 1 .
→ 20 clone groups
```

The 4 eliminated groups (24→20) correspond to the 3 generics candidates documented in the prior feedback file (`2026-08-10_discordsync_type-aware-false-negatives-generics-extraction-candidates.md`):

| Finding                    | Generic introduced       | Sites replaced             | Groups eliminated |
| -------------------------- | ------------------------ | -------------------------- | ----------------- |
| `maxXxxCount` loops        | `maxValue[T any]`        | 4 functions across 2 files | 2                 |
| `sumXxxCount` loops        | `sumValue[T any]`        | 2 loops across 2 files     | 1                 |
| Kind-derivation logic twin | `domain.ResolveUserKind` | 3 sites across 3 packages  | 1                 |

---

## The 20 Remaining Clone Groups: Exhaustive Classification

### Category A: Error-handling named-method boilerplate (8 groups, 21 clones)

These are Go's irreducible `if err != nil` pattern where each clone calls a **different named method** on the `db.Database` interface (which has 180 named methods, not a generic loader). The algorithm is always `call method → check error → wrap/return`, but you can't generic-ify `database.GetGuild(ctx, id)` vs `database.GetThread(ctx, id)` without restructuring the entire interface.

| # | Clones | Locations                                                                                                                                                  | Shape                                                        | Why `--suggest-generics` surfaces it                                                                                     | Why it's NOT extractable                                                                                                                                                                                          |
| - | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | 7      | `audit_log_checkpoint.go:68`, `backfill_progress.go:69`, `entities.go:205`, `query.go:168`, `query.go:322`, `query_channels.go:39`, `stickers_polls.go:51` | `if err != nil { return nil, queryRowErr(...) }`             | Different return types (`*Guild`, `*Thread`, `*AuditLogCheckpoint`, etc.) → flagged as "same algorithm, different types" | `queryRowErr` IS the shared helper. Each calls a different `QueryRowContext().Scan()` with different typed scan targets. A `scanRow[T]` generic would fight `database/sql`'s variadic `rows.Scan(any...)` design. |
| 2 | 3      | `handlers_phase2_detail.go:25,43,59`                                                                                                                       | `if err != nil { return nil, ... }`                          | `*Channel` vs `*Guild` vs `map[UserID]*User`                                                                             | 3 different `GetXxxByID` calls returning 3 different types. The 2-line guard is irreducible.                                                                                                                      |
| 3 | 3      | `types.go:75,87,98`                                                                                                                                        | `if err != nil \|\| ch == nil { return nil }`                | `*Channel` vs `*Thread` vs `*Guild`                                                                                      | `resolveChannel`/`resolveThread`/`resolveGuild` call different DB methods. Nil-guard is 2 lines.                                                                                                                  |
| 4 | 2      | `handlers_attachments.go:285,302`                                                                                                                          | `if err != nil \|\| user == nil` vs `guild == nil`           | `*User` vs `*Guild`                                                                                                      | `GetUser` vs `GetGuild` — different named methods.                                                                                                                                                                |
| 5 | 2      | `bot_fetchers.go:115,130`                                                                                                                                  | `restFetchUser`/`restFetchGuild` error tail                  | `*discord.User` vs `*discord.RestGuild`                                                                                  | Discord REST `GetUser` vs `GetGuild` — different disgo API calls.                                                                                                                                                 |
| 6 | 2      | `avatars.go:73`, `icons.go:72`                                                                                                                             | Download error handling                                      | `UserAvatarDownload` vs `GuildIconDownload`                                                                              | Different DB tables, different Discord entities. Already `//nolint:dupl`.                                                                                                                                         |
| 7 | 2      | `avatars.go:54`, `icons.go:53`                                                                                                                             | `if !content.IsAttachmentGoneError(err)`                     | Same avatar/icon mirror pair, different error path                                                                       | Already `//nolint:dupl`.                                                                                                                                                                                          |
| 8 | 2      | `mentions.go:94,118`                                                                                                                                       | `resolveMentionUsers` vs `resolveMentionChannels` error tail | `map[UserID]*User` vs `map[ChannelID]*Channel`                                                                           | The resolution logic above is completely different; only the 2-line error tail matches.                                                                                                                           |

**Suggested suppression pattern:** `error-guard-named-method` — when the clone body is `if err != nil { return ..., wrapErr(...) }` and the error source is a call to a named method (not a type-parametric function), suppress from `--suggest-generics` output. The existing `error-propagation` and `error-wrapping` actionability patterns (#5 and #16 in the priority table) should already catch simple cases — they may not fire because the clone spans 4+ statements (the `Scan` call + error check) rather than the 2-statement `assign-error-check` shape.

---

### Category B: Env-var-with-default parsers (3 groups, 4 clones)

Two packages (`internal/db/turso_sync.go` and `internal/config/config_file.go`) each have a `GetDuration`/`GetInt` pair. The parse calls differ (`time.ParseDuration` vs `strconv.Atoi`), so a generic can't unify them without a `Parse[T]` constraint abstraction that adds more complexity than it saves.

| #  | Clones | Locations                | Shape                                  | Why NOT extractable                                                                                        |
| -- | ------ | ------------------------ | -------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| 9  | 2      | `turso_sync.go:82,98`    | `envDuration`/`envInt` error handling  | `time.Duration` vs `int` — `ParseDuration` vs `Atoi` are different stdlib calls                            |
| 10 | 2      | `turso_sync.go:76,92`    | `v := os.Getenv(key)` + parse prefix   | Same function pair as #9, different line range of the same function                                        |
| 11 | 2      | `config_file.go:178,187` | `withDefaultDuration`/`withDefaultInt` | `Duration` vs `int` again; could theoretically unify with `constraints.Integer` but the parse calls differ |

**Note:** Groups 9+10 are the same 2 functions detected at different line ranges. This is a **fragmentation** issue — `--suggest-generics` reports overlapping windows of the same function pair as separate groups. A deduplication pass (if two groups share >50% of clone sites, merge them) would reduce 3 groups to 1.

**Suggested suppression pattern:** `env-var-parse-pair` — when the clone body is `os.Getenv` + `ParseXxx` + error + default, suppress. This is canonical Go env-var boilerplate.

---

### Category C: Intentional mirror pairs (3 groups, 6 clones)

These are structurally similar code operating on different domain entities (avatars vs icons, attachments vs embed media, channels vs threads). They're already acknowledged with `//nolint:dupl` where applicable or are too short to warrant suppression directives.

| #  | Clones | Locations                        | Shape                                  | Why NOT extractable                                                                                                       |
| -- | ------ | -------------------------------- | -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| 12 | 2      | `provider/discord.go:50,82`      | `AvatarURL` vs `IconURL` guard         | `GuildID` vs `UserID` branded types, different Discord CDN path construction below the guard                              |
| 13 | 2      | `integrity_repair.go:37,167`     | Attachment vs embed-media repair skip  | `*Attachment` vs `*EmbedMedia`, different DB tables, different download pipelines. Logic diverges below the 4-line guard. |
| 14 | 2      | `discordadapter/misc.go:156,241` | `ch.ParentID()` vs `thread.ParentID()` | `GuildChannel` vs `GuildThread` disgo interfaces — different types, different methods                                     |

---

### Category D: Coincidental 2-line idioms (4 groups, 11 clones)

Different code that happens to share the same shape by coincidence. The types are genuinely different AND the semantics are genuinely different.

| #  | Clones | Locations                                          | Shape                                                      | Why NOT extractable                                                                                                                                                                                       |
| -- | ------ | -------------------------------------------------- | ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 15 | 5      | `handlers_phase2.go:32,77,130,146,162`             | `if result.GuildID != "" { paginate }`                     | Each constructs a **different ViewModel type** and calls a **different row renderer**. The 2-line guard is the only shared part. `phase2ListResult[T]` is already generic; the handler call sites aren't. |
| 16 | 2      | `types.go:106,116`                                 | `if ch != nil && ch.Name != ""` vs `g.Name`                | `channelName()` vs `guildName()` — `.Name` field coincidentally exists on both `Channel` and `Guild`                                                                                                      |
| 17 | 2      | `handlers_dlq.go:156`, `query_phase2_counts.go:19` | `if X != "" { query += " WHERE..." }`                      | `projectionName` (string) vs `guildID` (branded `GuildID`). Different tables, different columns, different packages.                                                                                      |
| 18 | 2      | `activity_helpers.go:27`, `handler_helpers.go:276` | `if pct < 1 { pct = 1 }` vs `if offset < 0 { offset = 0 }` | `float64` vs `int`, CSS percentage clamp vs pagination offset clamp. Completely different semantics.                                                                                                      |

---

### Category E: Borderline extractable (1 group, 2 clones)

| #  | Clones | Locations            | Shape                                             | Assessment                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| -- | ------ | -------------------- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 19 | 2      | `mentions.go:89,113` | `for id := range ids { keys = append(keys, id) }` | `map[UserID]struct{}` vs `map[ChannelID]struct{}`. **This IS extractable** via `collectKeys[K comparable, V any](m map[K]V) []K`. Left unfixed because 2 call sites is below the threshold where a generic pays for itself (3+ is the standard threshold). The generic would be: <br><br>`func collectKeys[K comparable, V any](m map[K]V) []K {`<br>`keys := make([]K, 0, len(m))`<br>`for k := range m { keys = append(keys, k) }`<br>`return keys`<br>`}` |

---

### Category F: Trivial allocation (1 group, 2 clones)

| #  | Clones | Locations                                         | Shape                              | Why NOT extractable                                                                                 |
| -- | ------ | ------------------------------------------------- | ---------------------------------- | --------------------------------------------------------------------------------------------------- |
| 20 | 2      | `query_phase2.go:135`, `emojis_relational.go:121` | `args := make([]any, 0, len(X)+1)` | `[]UserID` vs `[]EmojiSnapshot`. A `sqlArgs(cap int) []any` helper saves 0 characters. Too trivial. |

---

## Improvement Suggestions for `--suggest-generics`

### 1. Suppress error-handling named-method boilerplate (highest impact)

**Problem:** 8 of 20 remaining groups (40%) are `if err != nil` guards on different named DB/interface methods. These are Go's most common clone shape and never benefit from generics extraction.

**Suggested approach:** After the generics-equivalence pass surfaces a candidate, run a secondary actionability check on the clone body. If the body is dominated by:

- An `if err != nil` branch (≥50% of statements), AND
- The error source is a `CallExpr` to a named method (not a type-parametric function)

...then suppress with label `error-guard-named-method` or reuse the existing `error-propagation` / `assign-error-check` patterns.

**Expected impact:** 8 of 20 groups suppressed → 12 remaining → noise-to-signal ratio improves from 19:1 to 11:1.

### 2. Merge fragmented groups from the same function pair

**Problem:** Groups 9+10 are the same `envDuration`/`envInt` function pair detected at different line ranges (error handling vs getenv prefix). They appear as 2 separate groups with 2 clones each, inflating the group count.

**Suggested approach:** Post-aggregation merge — if two groups share ≥50% of their clone sites (same file+function), merge them into one group covering the union of line ranges.

**Expected impact:** 3 env-var groups → 1 group. Total: 20→18.

### 3. Suppress 2-line coincidental idioms

**Problem:** Groups 16, 17, 18, 20 are all 2-line snippets where the shape match is coincidental (different semantics, different packages, different domains). `--suggest-generics` correctly identifies them as "same algorithm, different types" — they ARE the same algorithm — but extracting a generic for a 2-line idiom adds indirection for zero readability gain.

**Suggested approach:** A `min-effective-generics-lines` threshold (analogous to `--min-lines`) that suppresses generics candidates where the clone body is ≤2 statements. At 2 statements, the generic function signature is often longer than the code it replaces.

**Expected impact:** 4 more groups suppressed → ~8 remaining → noise-to-signal ratio ~7:1.

### 4. Surface the `collectKeys[K,V]` pattern as a hint

**Problem:** Group 19 (`map→slice key collection`) is the one remaining actionable candidate, but `--suggest-generics` treats it the same as the 19 non-actionable groups. There's no signal distinguishing "this IS extractable" from "this is coincidental shape similarity."

**Suggested approach:** When the clone body is a `RangeStmt` over a map where the key is appended to a result slice and nothing else happens, emit a hint: `"collectible: map→slice key extraction; consider collectKeys[K comparable, V any]"`. This is one of Go's most common generic patterns.

### 5. Document the `--type-aware --suggest-generics` combination

**Problem:** The `--help` text says `--suggest-generics` is "incompatible with `--type-aware`". But `--type-aware --suggest-generics` combined works correctly — type-aware loads go/types information, and suggest-generics uses the type differences to power its "same algorithm, different types" detection. This is the mode that produces the best results.

**Suggested fix:** Either (a) update the help text to say the combination is supported and recommended, or (b) make `--suggest-generics` implicitly enable `--type-aware` if it's not already set (since type info is required for generics detection).

---

## Summary

| Metric                               | Value                                   |
| ------------------------------------ | --------------------------------------- |
| Clone groups (initial scan)          | 24                                      |
| Real generics candidates             | 3 (found and eliminated)                |
| Clone groups (after fixes)           | 20                                      |
| Non-actionable (error boilerplate)   | 8 groups (40%)                          |
| Non-actionable (env-var parsers)     | 3 groups (15%)                          |
| Non-actionable (intentional mirrors) | 3 groups (15%)                          |
| Non-actionable (coincidental idioms) | 4 groups (20%)                          |
| Borderline extractable               | 1 group (5%)                            |
| Trivial                              | 1 group (5%)                            |
| **False-negatives**                  | **0**                                   |
| **Precision of remaining output**    | **~5%** (1 of 20 actionable)            |
| **Recall**                           | **100%** (3 of 3 real candidates found) |

**Bottom line:** `--suggest-generics` has excellent recall (zero false-negatives) but low precision on remaining output (19:1 noise). The three highest-impact improvements are: (1) suppress error-handling named-method boilerplate, (2) merge fragmented groups from the same function, (3) add a minimum-lines threshold for generics candidates. Together these would reduce 20 groups to ~8, with the 1 actionable candidate (`collectKeys`) clearly surfaced.

---

## Resolution (2026-08-10)

**IDENTIFIED.** Precision filtering plan (min-line-count gate, pattern exclusion, multi-position requirement) → TODO_LIST HIGH priority. Target: surface 3 real candidates with <5 false positives on DiscordSync.
