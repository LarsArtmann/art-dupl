> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation (fleet library review of go-finding consumers), not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[placeholder]`

**TL;DR:** v1.16.0 was tagged at `5f568e0` while its release notes still sit under `## [Unreleased]` — CHANGELOG.md has no `[1.16.0]` section at all, so the CHANGELOG says the run-level propertyBag is unreleased when the tag ships it. Ask: stamp the shipped notes as `[1.16.0] - 2026-11-22` and start a fresh `[Unreleased]`.

## Problem

`git show v1.16.0:CHANGELOG.md | grep -n '^## '` returns:

```
13:## [Unreleased]
56:## [1.15.1] - 2026-11-04
```

The tag `v1.16.0` (commit `5f568e058`, "feat: run-level SARIF propertyBag") carries no `[1.16.0]` heading — everything between lines 13 and 55 (the propertyBag `SetRunProperty`/`RunPropertiesFromSARIF` feature from #53, the golangci-lint v2.15.0 bump) is documented as unreleased.

## Why it matters

A consumer bumping to core v1.16.0 cannot confirm from the CHANGELOG what the release contains or whether it is safe to pin. art-dupl bumped `github.com/larsartmann/go-finding` v1.15.0 → v1.16.0 today (with toolsdk already at v1.16.0): verifying that `SetRunProperty` actually shipped in the tag required reading the tag tree (`git show v1.16.0:report.go`) because the CHANGELOG still listed the feature under `[Unreleased]`. Proxy users without a checkout get no signal at all.

## Suggested fix

Rename the current `[Unreleased]` heading to `[1.16.0] - 2026-11-22` (tag date) and open a new empty `[Unreleased]` above it. One-heading change, no content edits.

💘 Generated with Crush
