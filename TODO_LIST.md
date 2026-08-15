# TODO List

**Last Updated:** 2026-08-15

Actionable items for the next 2-4 weeks. Completed work lives in `CHANGELOG.md`.
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### HTML output improvements (remainder)
**Source:** `docs/status/2026-08-10_05-29_feedback-review-implementation-sprint.md` §b
"Detected vs Actionable" summary is DONE (both text and HTML printers). Remaining:
- [ ] **TTY-aware HTML output**: Auto-write to `art-dupl-report.html` when TTY detected (currently requires `--html-out` explicitly).
- [ ] **Stable display IDs**: `GroupNum` is not stable across runs (content hash IDs ARE stable). Add stable `id` attributes for deep-linking.

---

## MEDIUM Priority

### Feedback-driven actionability patterns (remainder)
**Source:** feedback files in `docs/feedback/new/`
`go-http-error-guard` and `go-error-wrap-idiom` are DONE — unified into the single
`error-guard-fallthrough` pattern (validated on DiscordSync: 82 groups → 2 shown, 0 FPs).
Remaining candidates:
- [ ] **`//go:embed` directive pattern**: Detect `//go:embed` + `embed.FS` + `fs.Sub` as compiler-bound (go-sse feedback)
- [ ] **`TestMain` boilerplate**: Go requires one `TestMain` per package (go-cqrs-lite)
- [ ] **`defer-cleanup-of-arbitrary-resource`**: `defer rows.Close()` / `defer tx.Rollback()` (discordsync — 12 groups)

### Cache stats visibility
- [ ] **Surface cache stats in `stats` subcommand**: `Stats.MemHits`/`Hits`/`Misses` are printed in verbose mode only (`cmd/run_analysis.go::printCacheStats`). Wire into `stats` output so non-verbose users see cache effectiveness.

### Infrastructure
- [ ] **`--exclude-pattern` UX**: Warn when pattern matches zero files; document glob vs regex (licenseforge feedback)
- [ ] **Coverage baseline**: Run `go test -cover` across all packages and commit a coverage baseline. We have benchmark baselines but no coverage baseline.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
- [ ] **Restructure `TypeAwareData` so `EraseHash` is collection-level**: Currently per-entry on `PreloadedAST`, validated at runtime with a warning (`job/incremental.go::SetTypeAwareData`). A collection-level type would enforce the invariant at compile time. Breaking change to `syntax/golang/typeinfo.go` with large blast radius.
