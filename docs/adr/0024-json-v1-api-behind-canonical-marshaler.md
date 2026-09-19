# ADR-0024: JSON v1 API Behind Canonical Marshaler

**Date:** 2026-09-19
**Status:** Accepted (supersedes the API-surface part of ADR-0010; the v2-engine part remains in force)

## Context

ADR-0010 migrated all JSON serialization to the `encoding/json/v2` API behind
`GOEXPERIMENT=jsonv2` (v2 API + v2 engine). Go 1.27 removed the v2 struct-tag
grammar (`format:` tags, [go.dev/issue/71631](https://go.dev/issue/71631)), so
every `format:nano` tag and all 36 files importing `encoding/json/v2` /
`jsontext` stopped compiling on the 1.27.1 toolchain. The breakage class is
structural: the v2 **API** is experimental and can change under any release,
while the project's JSON output is a compatibility surface (golden files,
SARIF consumers, baselines, downstream tooling all pin its bytes).

Two v2-era wire behaviors had been adopted and are depended upon:

1. **No HTML escaping** of `<`, `>`, `&` in string values.
2. **No trailing newline** from `json.Marshal`-style calls.

The stable v1 API differs visibly in exactly these two ways (v1 escapes HTML
by default; `Encoder.Encode` appends `\n`), and under `GOEXPERIMENT=jsonv2`
the v1 API runs on the v2 engine — so the engine stays, the surface must not.

## Decision

1. **All production marshaling goes through `internal/jsonutil`**
   (`MarshalIndent` / `Marshal`). It pins `SetEscapeHTML(false)` and trims the
   encoder's trailing newline, so output bytes are identical to the v2-era
   wire format **and** independent of whether `GOEXPERIMENT=jsonv2` is set.
   No production file imports `encoding/json/v2` or `jsontext`; v1
   `encoding/json` imports appear only in `internal/jsonutil` and tests.
2. **`GOEXPERIMENT=jsonv2` stays set** in flake.nix / CI / devShell. It swaps
   the engine under the stable v1 API (duration handling keeps v1 integer-
   nanosecond semantics either way; see AGENTS.md conventions).
3. **Struct-tag conventions return to v1 grammar.** `omitempty` on
   scalar output fields is deliberately NOT re-added: the v2-era output always
   emitted zero values, and restoring `omitempty` would silently drop keys
   consumers learned to expect. `omitzero` is not available via the v1 tag
   grammar on this toolchain.

## Consequences

- Wire formats are frozen across toolchain upgrades; golden files stay green.
- New output code must use `jsonutil`, never `json.Marshal` directly — a
  direct v1 `json.Marshal` would re-introduce HTML escaping and (via
  `Encoder`) trailing newlines. Enforced socially via AGENTS.md conventions
  and the `jsonutil` component in `.go-arch-lint.yml`.
- The v1 tag grammar is the compatibility contract; if a future Go makes the
  v2 API stable, the migration is a one-package change inside `jsonutil`
  (callers keep their signatures), with golden files proving byte-equality.
- Watch item: `GOEXPERIMENT=jsonv2` retirement behavior in Go 1.28 — because
  `jsonutil` output is experiment-independent, removal of the flag changes
  performance/semantics underneath but not our emitted bytes.
