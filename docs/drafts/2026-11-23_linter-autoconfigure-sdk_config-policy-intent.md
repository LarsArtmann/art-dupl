> [!NOTE]
> This filing was drafted by GLM-5.3 via Crush from an AI-run investigation (fleet library review across art-dupl/go-finding/linter-autoconfigure-sdk consumers), not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[placeholder]`

**TL;DR:** art-dupl bans two linters from the fleet defaults and therefore skips `golangci-lint-auto-configure` entirely — the all-or-nothing skip means the repo also loses every OTHER auto-fix the step would apply. Ask: should the SDK grow a consumer-side config-policy intent (per-repo carve-outs the auto-configurer honors), or is the skip-and-guard pattern the intended end state?

## Problem

art-dupl deliberately bans `exhaustruct` and `tagliatelle` (AGENTS.md "Lint config"; `scripts/check-disabled-linters.sh`). The auto-configurer re-adds both from fleet defaults on every run, so the repo:

1. skips the whole step via `.buildflow.yml` `skip_steps: [golangci-lint-auto-configure]`,
2. keeps a guard script that `sed`-removes the linters if a re-add ever lands,
3. fails red in the nix sandbox (`disabled-linters` check) when both defenses are bypassed.

This has been the stable arrangement since 2026-09-19, but it trades away the step's remaining value: go-line alignment, new-default adoption, and dead-reference cleanup all stop for the one repo that needs the most customization.

## What I checked

- The SDK surface today (autoconfigure.go, README §API): `DiffBlobs` for overrides-style block compares exists, but there is no per-repo policy input that would tell the auto-configurer "honor these bans / keep these blocks".
- The skip precedent: BuildFlow's `skip_steps` + the repo-side guard is the documented workaround class.

## The question

Is a consumer-side intent mechanism in scope for this SDK — e.g. a declared per-repo policy the auto-configurer reads (protected keys, banned-linter carve-outs) so the step can stay ON with carve-outs instead of being skipped wholesale? If the answer is "skip-and-guard is the intended contract", knowing that explicitly would let art-dupl delete its flip-trigger TODO and stop re-evaluating every train.

💘 Generated with Crush
