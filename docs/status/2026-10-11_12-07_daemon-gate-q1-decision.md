# Daemon Gate-Safety (Q1) — Prototype Ready, Owner Answer Requested

**Created:** 2026-10-11 12:07 CEST
**Status:** PROTOTYPE COMPLETE — awaiting the Q1 owner answer (asked 2026-10-05, re-asked 2026-10-11 status report §g)
**Question (Q1):** How do we stop the auto-commit daemon (pma) from enshrining broken tree states in git history — a pre-commit gate hook, or a daemon-side exclusion/gating mechanism?

---

## a) The damage class

The daemon commits whatever the working tree holds at debounce time. With concurrent agent sessions live on `fork`, a foreign half-done state gets snapshotted as HEAD. Since 2026-09-23 this fired at least **eight times**:

| Date             | Incident                                                                                                                                    | Cost                                                             |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| 2026-09-23/24    | jsonv2gate recurrences #1–#3 (`internal/jsonutil` + `config` re-migrated to `encoding/json/v2`; Duration refusal broke `config.SaveConfig`) | masked for days by stale test caches before detection            |
| 2026-09-29       | jsonutil regression **reached the remote** (`b679c615`, restored `42ee0a72`)                                                                | ~90 minutes live on `origin/fork`                                |
| 2026-09-29       | third jsonutil clobber (`ce716456`, restored `9a97ec82`)                                                                                    | session time                                                     |
| 2026-10-04       | recurrence #4 (`758646e5`, restored)                                                                                                        | session time                                                     |
| 2026-10-04/05    | recurrence #5 (`db992e16`/`b2a3b4ec`, restored from `d8e53515`)                                                                             | golden tests flipped nondeterministically before detection       |
| 2026-10-06       | recurrence #6 (`b729d8a7`, restored `2e3d3cd7`)                                                                                             | caught by boundary gate before any consumer rebuilt              |
| 2026-10-11 06:15 | **build-broken half-state committed as HEAD** (`8afa9803`, healed `60b3b403`) — foreign generics-session edit with `undefined: golang`      | session time; the class is not JSON-specific, it is ANY red tree |

## b) Evidence reproduced today (2026-10-11)

```
$ git worktree add /tmp/m2-evidence 8afa9803
$ cd /tmp/m2-evidence && go build ./...
# github.com/LarsArtmann/art-dupl/printer
printer/generics_candidate.go:118:12: undefined: golang
printer/generics_candidate.go:119:12: undefined: golang
$ echo $?
1
```

Any consumer that checked out `8afa9803` between 06:15 and the heal could not build. Nothing today prevents the next one.

## c) The prototype (shipped in this repo)

`scripts/daemon-gate/` — a **single static Go binary** usable as a git `pre-commit` hook. Design constraints discovered in research:

- The daemon's systemd environment **deliberately disables repo hooks** (`GIT_CONFIG_VALUE_0=/var/empty` in `SystemNix/modules/nixos/services/projects-management-automation.nix:187-189`): its PATH carries only `pma, git, coreutils, findutils, grep, sed, systemd` — no bash — so every `#!/usr/bin/env bash` hook failed to exec and the daemon **retried forever** (live incident 2026-08-19, PMA `ce0f638`). Therefore the hook must be an interpreter-free binary; a shell script is a non-starter.
- pma shells out to real `git commit` **without** `--no-verify` (`committer/committer.go:90-110` never sets `SkipHooks`), so a hooks-path override re-enables hooks **for the daemon only**, leaving interactive commits untouched.
- pma failures hit a **failure cooldown** (not the old hot-retry) — `failureCooldown = 5 * time.Minute`, capped at `maxFailureCooldown = 4 * time.Hour` (`pma-daemon/service.go:36,43`) — and message generation happens before staging; with our AI provider already failing every cycle (all daemon commits are `(heuristic)`), a refused commit costs the same per cycle as today's normal cycle.

Semantics (encoded in the binary, not just documented):

- **Fail-OPEN on infrastructure errors** (no Go toolchain, no GOCACHE, non-Go directory): exit 0 with an actionable stderr line. A gate that bricks the daemon is the exact `ce0f638` retry-forever class.
- **Fail-CLOSED on verdicts**: `go build ./...` red → exit 1 (git refuses); art-dupl additionally runs the `internal/jsonv2gate` AST scanner (the Duration/wire regression catcher) when that directory exists — fleet-generic by construction, art-dupl-shaped modules get the deeper gate.
- Kill switch: `PMA_GATE_DISABLE=1` bypasses for one commit without redeploying anything.
- Go toolchain discovery via `PMA_GATE_GO` absolute pin (the daemon PATH has no `go`) with `LookPath` fallback; `exec.CommandContext` with a 10-minute per-check timeout (a hung gate must not hang the daemon).

**Verified:** 8 unit tests (broken module refused with compiler evidence carried through, green module passes, non-Go dir passes through, kill switch, GOCACHE fallback, pinned-toolchain precedence, `errors.Is` sentinel) + an end-to-end proof in a scratch repo under the daemon's exact env mechanism (`GIT_CONFIG_COUNT`/`core.hooksPath`): broken state → `git commit` exits 1 with `daemon-gate: REFUSED`; fixed state → commit succeeds with `daemon-gate: green`. On art-dupl itself the gate runs build + jsonv2gate green in ~13 s warm — comfortably inside the daemon's 120 s minimum commit interval. Lint clean under the repo's 106-linter config; `go-arch-lint` component `daemon-gate` registered; flake output `.#daemon-gate` added so SystemNix can consume the binary directly.

## d) Options costed

| Option                         | Mechanism                                                                                                   | Red-tree behavior                                                                                       | Code changed                                     | Fleet coverage                                | Risk                                                                                                                                           |
| ------------------------------ | ----------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- | ------------------------------------------------ | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| **A. Hook binary** (prototype) | SystemNix flips `GIT_CONFIG_VALUE_0` from `/var/empty` to a hooks dir containing the `.#daemon-gate` binary | `git commit` exits 1 → pma failure cooldown absorbs; ~1 failed LLM call + 13 s gate per cycle while red | **Zero** (config-only)                           | Every Go repo the daemon commits, immediately | A red tree lasting hours burns cooldown cycles; hook cannot tell pma "skip quietly"                                                            |
| **B. pma-native gate**         | pma committer runs a per-repo gate command BEFORE staging; red → existing `StatusSkipped` path              | No LLM call, no git attempt, clean skip logged                                                          | pma committer + config schema + SystemNix render | Same, after pma release                       | Touches the fleet-critical daemon; the SystemNix comment already says "root fix belongs upstream (PMA committer should skip hooks explicitly)" |
| **C. Exclusion list**          | Add art-dupl to daemon `excludePaths`                                                                       | No commits at all from art-dupl                                                                         | Config-only                                      | art-dupl only                                 | Loses auto-commit for the most actively developed repo; protects nothing else. **Rejected**                                                    |
| **D. Status quo**              | —                                                                                                           | Broken states keep entering history                                                                     | Zero                                             | None                                          | The class fired twice in the last 5 days. **Rejected**                                                                                         |

## e) Recommendation

**Option A now, Option B as the upstream follow-up.**

A is reversible (flip one env var back), requires no daemon code change during a period when concurrent sessions are actively editing pma, and kills the entire class fleet-wide the moment SystemNix deploys it. Its only real cost — refusal cycles while a tree is legitimately red mid-session — is bounded by the existing cooldown (5-minute backoff escalating to a 4-hour cap, `pma-daemon/service.go:36,43`) and is strictly cheaper than the incidents (90 minutes of a broken remote; days of masked cache staleness). B remains the semantically correct end state (silent skip instead of refused commit) and the SystemNix rationale comment already anticipates it; prototype A's gate logic is reusable verbatim if B execs the same binary.

## f) Rollout checklist (owner)

1. `nix build github:LarsArtmann/art-dupl#daemon-gate` (or from the local checkout).
2. In `SystemNix/modules/nixos/services/projects-management-automation.nix`: point `GIT_CONFIG_VALUE_0` at a hooks dir containing `pre-commit → ${daemonGate}/bin/daemon-gate` (a `runCommand` symlink dir), keep `GIT_CONFIG_COUNT`/`KEY_0` as is.
3. Add to the same `Environment` block: `"PMA_GATE_GO=${lib.getExe pkgs.go_1_27}"` (and `GOCACHE=/var/cache/pma-gate` with a matching `RuntimeDirectory`/`StateDirectory` if the service HOME is not writable).
4. `nixos-rebuild test` → stage a deliberately broken file in a scratch clone → confirm `git commit` refuses with `daemon-gate: REFUSED` in the journal → revert the file.
5. Record the decision in art-dupl `AGENTS.md` (M2.6) and the SystemNix rationale comment (replace the `/var/empty` rationale with the gate story).

## g) Rollback

Set `GIT_CONFIG_VALUE_0=/var/empty` again (one line, one rebuild). `PMA_GATE_DISABLE=1` is the per-commit emergency bypass that does not even need a rebuild.

## h) The question

**Q1 (owner):** Deploy Option A via SystemNix as specified in §f? And should B (pma-native `StatusSkipped` gate) be filed as the upstream pma feature it already half-anticipates?
