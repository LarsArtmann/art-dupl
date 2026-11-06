# Issue closure comments (2026-11-06)

Drafted for `gh issue close` on LarsArtmann/art-dupl. Terse by design.

## #9

Both forms work now. The line-leading `// art-dupl:accept` is a templ Go comment and `templ generate` drops it — checked on a scratch module, no marker in the generated Go. `<!-- art-dupl:accept ... -->` without the nested `//` is recognized too. Pinned in `bdd/templ_accept_directive_test.go`.

## #10

`syntax.DeduplicateMemberSets` collapses groups whose displayed member sets are identical, keeping the richer fragment. nsfw-classifier @ a7cec08a combined mode: 0 duplicate member sets across 46 shown groups.

## #11

Three fixes: mixed-parser matches are rejected when any `.go` fragment is a non-statement unit (`TestFindSyntaxUnitsMixedParserGoParamsRejected`), 1-token/1-line groups get the `trivial-fragment` shape label, cross-type families failing the min-lines gate are dropped instead of printed unmarked, and Go snippets in `.templ` files carry `Node.GoSnippet` so they classify through the Go table — no templ fix suggestion on Go code anymore. nsfw-classifier @ a7cec08a combined mode: 46 shown, zero mixed param-fragment groups — the 52-member family is gone.

## #12

Done — `check` prints `Found total N clone groups (B in baseline, W new[, S suppressed])`, and the summary stamps the build version.

## #13

Done — baselines record the analyzer identity (version, algorithm revision, hashing policy over mode/type-aware/suggest-generics/max-children) and `check --baseline-version-must-match` refuses a cross-policy compare, naming both identities. Compatibility is decided by the policy fingerprint, not the version string, so a same-algorithm rebuild keeps the baseline valid; baselines without a recorded identity are refused by the flag.

## #14

Fixed at the source: markup text now participates in the hash (the #6 policy, extended to text), so sections that render different words stop cloning. At `-t 1`: DiscordSync 112→93 shown, nsfw-classifier 147→129, webphone 44→43, Go-only corpora byte-identical. Removed groups verified as text/attr-divergent pairs (`<th>File</th>` vs `<th>Author</th>`, `<span>—</span>` vs `<span>:</span>`). The non-rendering `//` directive form also works now, so per-site suppression replaces baseline growth.
