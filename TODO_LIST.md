# TODO List

**Last Updated:** 2026-08-10

Actionable items for the next 2-4 weeks. Completed work lives in `CHANGELOG.md`.
This file is OPEN work only — no completed, rejected, or resolved items.

---

## HIGH Priority

### `--suggest-generics` precision filtering
**Why:** 12.5% precision on real-world validation (3 true / 24 surfaced on DiscordSync). Type-difference detection is necessary but not sufficient — error-handling boilerplate, nil-guards, and collection idioms have type differences but are NOT generics candidates.
**Source:** `docs/status/2026-08-10_03-37_suggest-generics-e2e-validation-discordsync.md` §D1
**Evidence:** `printer/generics_candidate.go` — `ClassifyGenericsCandidate` only checks for type divergence, no precision heuristics.
- [ ] Add minimum-line-count gate (default 4) — most noise is 1-2 statement clones; the 3 real candidates are 4-8 line blocks (`--suggest-generics-min-lines` flag)
- [ ] Cross-reference generics candidates against actionability patterns — exclude groups matching `error-propagation`, `bool-guard`, `guard-clause`, `single-call-expression`, etc.
- [ ] Require N+ type-difference positions at non-trivial depth (filter shallow 1-position differences)
- [ ] Re-run against DiscordSync — target: surface 3 real candidates with <5 false positives

### `--min-tokens` test coverage
**Why:** Flag is wired and committed (`fd3b0d38`) but has ZERO tests — unit or BDD.
**Source:** `docs/status/2026-08-10_05-29_feedback-review-implementation-sprint.md` §e.3
**Evidence:** `grep -rl "MinTokens\|minCloneTokenCount" *_test.go` returns nothing.
- [ ] Add unit tests for `minCloneTokenCount` suppression logic (mirror `TestShouldSuppressGroup_MinLines`)
- [ ] Add BDD test for `--min-tokens` flag

### Tagliatelle lint contradiction
**Why:** `.golangci.yml` has `tagliatelle` enabled but AGENTS.md says it is NOT. 50 `json(camel)` violations exist.
**Source:** `docs/status/2026-08-07_22-05_post-lint-cleanup-pipeline-tests-status.md` §e.4
**Evidence:** `grep tagliatelle .golangci.yml` → `- tagliatelle` (line ~108)
- [ ] Remove `- tagliatelle` from `.golangci.yml` enable list (matching AGENTS.md and ADR-0016 mixed JSON convention)

### `--suggest-generics` output quality
**Why:** `generics_hint` uses fully-qualified type paths (375 chars on one line, unreadable in terminal).
**Source:** `docs/status/2026-08-10_03-37_suggest-generics-e2e-validation-discordsync.md` §D2
**Evidence:** `printer/generics_candidate.go:68` `formatGenericsHint` uses raw `go/types` string output.
- [ ] Strip module path prefix from type strings (use `types.Qualifier` or string manipulation to show `db.GuildID` instead of `github.com/.../db.GuildID`)

---

## MEDIUM Priority

### `--suggest-generics` completeness gaps
- [ ] **Config validation**: `--type-aware --suggest-generics` silently runs suggest-generics with no warning. Add validation or warning. (`config/config_validate.go` — no `SuggestGenerics` check exists)
- [ ] **SARIF support**: SARIF output drops all generics info. Add `generics_candidate` + `generics_hint` to SARIF `properties` bag. (`printer/sarif.go` — no generics fields)
- [ ] **Full-pipeline integration test**: Only unit tests exist. Need a test that writes Go files → `LoadTypeAwareData(eraseHash=true)` → detection → classification → asserts generics candidate. (`printer/actionability/pipeline_integration_test.go` doesn't support type-aware loading)
- [ ] **BDD test**: No `--suggest-generics` scenario in `bdd/`.
- [ ] **ADR-0020**: Document the EraseHash design decision in `docs/adr/`.

### HTML output improvements
**Source:** `docs/status/2026-08-10_05-29_feedback-review-implementation-sprint.md` §b, multiple feedback files
- [ ] **"Detected vs Actionable" in HTML summary**: `SuppressionStatsSetter` is implemented by `TextPrinter` but NOT by the HTML printer. HTML still shows old single-total summary. (`printer/printer.go` — `SuppressionStatsSetter` interface)
- [ ] **TTY-aware HTML output**: Auto-write to `art-dupl-report.html` when TTY detected (currently requires `--html-out` explicitly).
- [ ] **Stable display IDs**: `GroupNum` is not stable across runs (content hash IDs ARE stable). Add stable `id` attributes for deep-linking.

### Cache improvements
**Source:** `docs/status/2026-08-10_04-39_post-audit-hardening-self-critique.md` §f
- [ ] **Bump `CacheVersion` to 3**: The `KeyWithParams` change orphaned all existing on-disk caches (old key format never looked up again). Not a correctness bug, but inelegant. (`cache/file_cache.go` — `CacheVersion` still `2`)
- [ ] **In-memory LRU layer**: Add in-process cache on top of `FileCache` to avoid redundant gob deserialization on hot paths. Biggest perf win for cache-heavy workflows.
- [ ] **Hysteresis pruning**: Current `Prune` sorts ALL entries on every call (O(n log n) per miss). Prune at 110%, evict to 90%.
- [ ] **`cacheKey()` unit test**: Test `ip.cacheKey()` in `job/incremental_test.go` to verify params string format (`"mode:maxChildren:typeAwareTag"`).

### Feedback-driven actionability patterns
**Source:** `docs/status/2026-08-10_05-29_feedback-review-implementation-sprint.md` §c, feedback files in `docs/feedback/new/`
Zero-effort, high-precision patterns identified from real-world feedback:
- [ ] **`//go:embed` directive pattern**: Detect `//go:embed` + `embed.FS` + `fs.Sub` as compiler-bound (go-sse feedback)
- [ ] **`go-http-error-guard`**: `if err != nil { writeError(w,r,err); return }` forced by `http.HandlerFunc` void signature (discordsync — 12 groups)
- [ ] **`go-error-wrap-idiom`**: Unique-string error wrappers (discordsync — 27 groups)
- [ ] **`TestMain` boilerplate**: Go requires one `TestMain` per package (go-cqrs-lite)
- [ ] **`defer-cleanup-of-arbitrary-resource`**: `defer rows.Close()` / `defer tx.Rollback()` (discordsync — 12 groups)

### Infrastructure
- [ ] **`global.out.css` in `.gitignore`**: Untracked generated file appeared during sessions. (`grep global.out.css .gitignore` returns nothing)
- [ ] **`--exclude-pattern` UX**: Warn when pattern matches zero files; document glob vs regex (licenseforge feedback)
- [ ] **Coverage baseline**: Run `go test -cover` across all packages and commit a coverage baseline. We have benchmark baselines but no coverage baseline.

---

## DEFERRED: Architecturally Constrained

Blocked by fundamental design constraints. Cannot be resolved without significant architectural changes.

- [ ] **Branded `NodeType int32`**: Per-package `NodeType` types would prevent cross-package constant collision. **HIGH RISK**: touches gob cache format. Current 8-bit shared encoding is intentional (ADR-0008).
- [ ] **Hide `syntax/golang` behind facade**: **BLOCKED** by import cycle (`syntax/golang` imports `syntax` for Node type; `printer/actionability*.go` imports `syntax/golang` for AST constants).
- [ ] **Hybrid slice/map transition storage**: Map already O(1); slice optimization deferred as low-value.
- [ ] **Restructure `TypeAwareData` so `EraseHash` is collection-level**: Currently per-entry on `PreloadedAST`, validated at runtime. A collection-level type (`type TypeAwareData struct { EraseHash bool; Entries map[string]*PreloadedAST }`) would enforce the invariant at compile time. Breaking change to `syntax/golang/typeinfo.go` with large blast radius.
