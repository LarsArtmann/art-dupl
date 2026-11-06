# ADR-0029: Baselines Record the Analyzer That Produced Them

**Date:** 2026-11-06
**Status:** Accepted

## Context

A baseline stores clone-group hashes. A group's hash is an XXH3 digest over the
serialized token stream, so it is only meaningful *inside one hash space*: the
same source produces different hashes under a different algorithm revision
(`cache.CacheVersion`), detection mode, type-aware setting, or detector set.

Before this decision, `art-dupl check --baseline-path` compared today's hashes
against whatever was on disk with no compatibility statement. After any change
that moved the token stream (ADR-0023 nested statements, ADR-0027
expression-aware templ, the `CacheVersion` 5→8 revisions), every recorded entry
looked new: a consumer's CI gate reported a whole-tree regression with no code
change, and the only stated remedy was "re-record the baseline" — which was
indistinguishable from "the analyzer thinks you added clones".

Issue #13 asked for the missing guarantee: a `check` run must not silently
compare hashes from two different analyzers.

## Decision

Record an `AnalyzerIdentity` on every baseline:

```json
"analyzer": {
  "version": "0.9.1",
  "cacheVersion": 8,
  "policy": "5f3c1e2a9b7d4c60"
}
```

1. **`policy` is a fingerprint of the hash-affecting settings**, not a list of
   them: `cache=<CacheVersion>;mode=<DetectionMode>;typeAware=<bool>;`
   `suggestGenerics=<bool>;maxChildren=<n>;methods=<sorted detector names>`,
   hashed with XXH3-64 and rendered as 16 hex chars.
2. **`version` is provenance only.** It appears in error messages and nowhere
   in the compatibility test.
3. **Compatibility = same `policy` AND same `cacheVersion`.** Both are checked
   because `policy` embeds the cache version; the explicit field keeps a
   human-readable reason in the message and survives a hashing change to the
   policy string itself.
4. **A baseline without an `analyzer` block is unverifiable**, not compatible.
   It was written by a build that could not state its hash space.
5. **The refusal is opt-in:** `check --baseline-version-must-match` turns an
   incompatible baseline into a validation error (exit code 2) that names both
   identities. Without the flag, behavior is unchanged.

## Rationale

1. **Two identities, two purposes.** Policy answers "can these hashes be
   compared?"; version answers "who do I talk to about this number?". Folding
   the version into compatibility would invalidate every consumer baseline on
   every release that touches an unrelated printer or doc, which trains people
   to record flag state instead of reading it. Excluding it is the load-bearing
   choice of this ADR.
2. **Hashed policy, not a field list.** A structured policy would either be
   compared field-by-field (a table that rots as `Config` grows) or embedded
   verbatim (a `Config` dump that leaks settings like `Timeout` that cannot
   affect hashes). One function (`analyzerPolicyParts`) owns the list, one hash
   owns the comparison, and adding a hash-affecting setting is a one-line edit
   in one place.
3. **Order-independent detector set.** `-m hash,art-dupl` and
   `-m art-dupl,hash` run the same detectors and must compare equal; the
   names are sorted before folding.
4. **Opt-in, because the alternative is a breaking default.** Consumers whose
   baselines predate identity recording would fail their next CI run under a
   default-on flag, with a message telling them to re-record — the exact
   silent-regression experience the ADR exists to remove. The flag lets a team
   migrate on its own commit.
5. **Legacy `nil` is refused under the flag.** The caller asked for a
   guarantee; "probably compatible" is not a guarantee. `CompatibleWith`
   returns false for an empty policy on either side, and the error message
   names both re-recording and dropping the flag.

## Consequences

- `baseline.File.Analyzer` is additive and `omitempty`: old baselines still
  load, and old readers ignore the new key.
- A `CacheVersion` bump now has a *stated* consumer cost: every baseline must
  be re-recorded. Release notes must say so (see the CacheVersion 7→8 note in
  `CHANGELOG.md`).
- `ensureBaselineIdentityMatches` is the single guard; `check` calls it after
  config resolution so the policy reflects the flags actually in force.
- Tests pin the contract instead of the bytes: round-trip, legacy-nil, and a
  compatibility table (`baseline/baseline_test.go`).
- Follow-ups tracked in `TODO_LIST.md`: surface the mismatch as a default-on
  stderr notice (no exit-code change), emit the split in `check --json`, and
  apply the same guard to `--diff-report`.
