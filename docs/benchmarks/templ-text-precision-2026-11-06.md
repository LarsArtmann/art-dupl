# Templ markup-text precision: before/after (2026-11-06, issue #14)

Text content in `.templ` markup used to be dropped from the token stream
("focus on structure"), so parallel-by-design sections collided on element
shape alone: `<th>File</th>` matched `<th>Author</th>`, `<span>—</span>`
matched `<span>:</span>`, and each UI train re-red the `-t 1` new-clones gate
in consumer repos. Issue #14 collected 20+ such ACCEPTed groups across five
trains.

`TextContent` (node type 102) now hashes the text (`syntax/templ/transform_node.go`);
semantic mode collapses whitespace runs and coalesces the parser's
line-break-split text nodes, exact mode hashes verbatim, structural mode
ignores values. CacheVersion 8.

## Measurement

"before" = `v0.9.1-0.20261104060928-0cfc9966`, "after" = same tree plus the
text-content change (`detected / shown` from the summary footer). Corpora:
pinned worktrees under `~/projects/.corpus-pins/` plus the two live consumer
repos at their heads (`webphone` 36f0c6f4, `nsfw-classifier` a7cec08a).

| Corpus                   | `-t`  | detected before → after   | shown before → after  |
| ------------------------ | ----- | ------------------------- | --------------------- |
| art-dupl (self)          | 1     | 1379 → 1379               | 10 → 10               |
| art-dupl (self)          | 2     | 425 → 425                 | 5 → 5                 |
| art-dupl (self)          | 5     | 62 → 62                   | 0 → 0                 |
| webphone (templ-heavy)   | 1     | 1589 → 1587               | 44 → 43               |
| webphone (templ-heavy)   | 2     | 443 → 441                 | 24 → 24               |
| webphone (templ-heavy)   | 5     | 62 → 60                   | 0 → 0                 |
| nsfw-classifier          | 1     | 2263 → 2239               | 147 → 129             |
| nsfw-classifier          | 2     | 725 → 722                 | 87 → 86               |
| nsfw-classifier          | 5     | 103 → 103                 | 9 → 9                 |
| DiscordSync              | 1     | 3753 → 3729               | 112 → 93              |
| DiscordSync              | 2     | 1486 → 1485               | 53 → 52               |
| DiscordSync              | 5     | 348 → 348                 | 0 → 0                 |
| go-cqrs-lite (Go-only)   | 1/2/5 | 8697/3975/993 → identical | 468/322/1 → identical |
| branching-flow (Go-only) | 1/2/5 | 1042/335/51 → identical   | 6/3/0 → identical     |
| go-paperless (Go-only)   | 1/2/5 | 220/76/17 → identical     | 9/8/1 → identical     |

Go-only corpora are byte-identical at every threshold: the change is scoped to
`.templ` markup. Templ-heavy corpora lose groups only where the rendered text
differs.

## Removed shown groups (DiscordSync `-t 1`, 19 of 112)

Every removed group is the issue-#14 class, verified by member inspection:

- `<th class="pb-2 font-medium">File</th>` / `Author` / `Type`
- `<span class="text-gray-400">—</span>` / `:</span>`
- `<p class="text-sm text-gray-500 ...">No storage ...</p>` / `No attachments ...`
- `<caption class="sr-only">Message count heatmap ...</caption>` / `Command audit entries`
- `<h2 class="text-sm font-medium ...">` headings differing only in wording

Groups that survive are genuine duplicates (identical markup and text) or
Type-2 renamed-identifier clones, which stay reportable by design.

## Residual classes (deliberately unchanged)

- Parallel markup whose only difference is an i18n key
  (`{ T(lang, "a.b") }` vs `{ T(lang, "a.c") }`) still matches in semantic
  mode (Go literal values normalize to their kind) but is classified
  non-actionable by the property engine (`property-parameterizable`).
- Whitespace-only text (`<span> </span>`) is skipped like any whitespace.
