# Constant CSS Property Values: Corpus Sweep (2026-11-06)

**Change:** `syntax/templ/transform.go::transformCSSProperty` now hashes a
constant CSS property's VALUE alongside its name (`color: red` no longer
matches `color: blue`, and neither matches `background: red`). CacheVersion
8 → 9.

**Why:** the value was hash-invisible. Any two rules with the same property
name but different values collapsed into one clone group, which is the same
"markup that renders differently is not a clone" class already fixed for
attributes (#6, CacheVersion 7) and text (#14, CacheVersion 8). The value was
the last markup identity hole in the templ transformer.

## Command

```bash
go build -o /tmp/art-dupl-new ./cmd/art-dupl
scripts/corpus-diff.py --before ./result/bin/art-dupl --after /tmp/art-dupl-new \
  --args "-t <n>" <corpus>
```

`./result/bin/art-dupl` = the pre-change build (CacheVersion 8); the new binary
is the fix (CacheVersion 9). Corpora are pinned worktrees under
`~/projects/.corpus-pins/`; `webphone` is the live consumer.

## Results (groups / clones)

| Corpus                    | `-t 1`          | `-t 2`         | `-t 5`      |
| ------------------------- | --------------- | -------------- | ----------- |
| nsfw-classifier@a7cec08a  | 129/343 → same  | 86/228 → same  | 9/18 → same |
| DiscordSync@4c0502c3      | 93/231 → same   | 52/140 → same  | 0/0 → same  |
| go-cqrs-lite@10c9f987     | 468/1486 → same | 322/779 → same | 1/2 → same  |
| go-sse@67a16fa            | 3/7 → same      | 2/5 → same     | 0/0 → same  |
| webphone (live, 36f0c6f4) | 43/113 → same   | 24/73 → same   | 0/0 → same  |

`scripts/corpus-diff.py` reported "clone counts by category: unchanged" for
every corpus/threshold pair — no group was removed, added, or rekeyed.

## Why zero corpus movement is the expected result

The bug only fires when two clone instances share a property NAME and differ
only in its VALUE, inside a `css` block. The corpora do have templ files
(DiscordSync 42, nsfw-classifier 23), but their `css` blocks are either absent,
single-instance, or already distinguished by other tokens. The defect is real
and reproducible, just rare in these trees.

## Minimal reproduction (before → after)

Two files identical except for the CSS value:

```templ
package main

css panel() {
	color: red;   /* vs blue */
	margin: 4px;
}

templ Card_X() {
	<div class="panel">
		<span>Hello</span>
	</div>
}
```

- before: 2 groups (`card_blue.templ:3-6` + `card_red.templ:3-6` reported as a
  clone of the CSS block, plus the genuinely shared markup at 9-12)
- after: 1 group (only the shared markup)

## Guards

- `syntax/templ/behavior_test.go`: `TestConstantCSSPropertyValuePreventsClone`
  (semantic AND exact), `TestConstantCSSPropertyNamePreventsClone`,
  `TestConstantCSSPropertySameValueStillClones`,
  `TestConstantCSSPropertyWhitespaceCollapses` (semantic collapses whitespace
  runs, never token-sorts: CSS shorthand order is semantic), and
  `TestConstantCSSPropertyValueInvisibleInStructuralMode`.
  The first test was canary-verified: reverting the transform to the old
  name-only encoding makes it fail.
- `cache.CacheVersion` bumped to 9, so cached v8 entries cannot replay the
  value-blind stream.

## Consumer impact

CacheVersion 8 → 9 invalidates on-disk caches and every recorded baseline's
hash space (see ADR-0029). Re-record baselines after upgrading:
`art-dupl baseline`.
