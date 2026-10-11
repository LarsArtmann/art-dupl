# Precision Re-measure Data (2026-10-11)

Measurement artifacts for the `--suggest-generics` precision work — the compact
per-family candidate tables backing the ledger rows in
[`docs/research/2026-10-11_suggest-generics-precision-baseline.md`](../2026-10-11_suggest-generics-precision-baseline.md)
and the pattern-validation deltas in the CHANGELOG.

## What is here

- `candidates-<repo>.tsv` — one row per generics-candidate **family** (unique
  group hash) in the PRE-pattern population (`-t 2`, `--no-actionability`).
  This is the 37.7% (23/61 TRUE) measurement basis.
- `survivors-<repo>-fixed.tsv` — families still flagged after the fixed
  `append-collector-tail` + `option-apply-loop` patterns (post 2026-10-11
  child-order fix), with re-derived TRUE/FALSE labels.

Raw JSON outputs are deliberately NOT committed (multi-MB, re-derivable with
one command per repo — see below). The TSVs carry the durable part: which
families existed, where, and what survived.

## Regeneration

```sh
art-dupl <repo-path> --suggest-generics --no-actionability -t 2 --format json --quiet \
  | jq -r '[.clone_groups[] | select(.files[]?.generics_candidate == true)]
           | unique_by(.hash) | .[]
           | .hash + "\t" + ([.files[] | .filename + ":" + (.line_start|tostring) + "-" + (.line_end|tostring)] | unique | join(", "))'
```

## Corpus pins (pre-pattern population, 2026-10-11 ~12:13 CEST)

| Repo            | HEAD        | Note                                     |
| --------------- | ----------- | ---------------------------------------- |
| go-cqrs-lite    | `80bf5a18a` | the only repo session 2 pinned           |
| go-sse          | `67a16fa`   | recorded post-hoc (repo drifts under its own daemon) |
| go-paperless    | `1909797`   | recorded post-hoc                        |
| DiscordSync     | `4c0502c3`  | recorded post-hoc                        |
| branching-flow  | `0d11b13a`  | recorded post-hoc                        |
| art-dupl        | fork HEAD at measurement time |                           |

Post-fix survivor tables were measured at the 2026-10-11 18:5x HEADs
(recorded in each TSV header) because the corpus drifted between
measurements; family-level deltas are still valid because group hashes are
content-derived, but absolute family sets can differ where the corpus
changed.
