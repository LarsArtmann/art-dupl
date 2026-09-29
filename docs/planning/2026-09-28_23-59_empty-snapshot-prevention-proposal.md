# Proposal — Daemon Guards: Empty-Snapshot Prevention + Concurrent-Session Tripwire

**Date:** 2026-09-28 (tripwire section added 2026-09-29, SUPERB v2 M04)
**Origin:** the 2026-09-28 docs-health audit found **10 committed zero-byte (or near-zero) status snapshots** — daemon auto-commits swept empty or truncated report files into history, and each one needed a hand-written "honest one-line resolution note" during archiving. The 2026-09-29 jsonutil incident added the second daemon failure mode: a foreign working-tree edit rode a docs sweep into a pushed commit and lived ~90 minutes because nothing named it. This page proposes the guards; the tripwire (section D) is now IMPLEMENTED in-repo.

## The failure modes

**F1 — empty snapshots.**

1. A session creates `docs/status/<timestamp>_<name>.md` intending to fill it.
2. The session dies (user stop, crash, context switch) before content lands.
3. The daemon's heuristic sweep commits the empty file with `chore: auto-commit N file(s)`.
4. Weeks later a docs-health pass finds a title-less 0-byte file in `docs/status/` and cannot tell abandoned-skeleton from deleted-content.

Ten instances predating 2026-09-28. Each converts a "what happened here?" into archaeology.

**F2 — unnamed foreign edits.**

1. Two sessions share the working tree (the norm here).
2. Session A commits its task; the daemon sweep carries Session B's in-flight edits into A's commit under a heuristic message.
3. The jsonutil class: a forbidden `encoding/json/v2` rewrite reached the REMOTE this way and survived ~90 minutes; the go.mod `go 1.27.1` pin has flipped this way repeatedly (caught live by the pin gate on 2026-09-29).

## Proposed guards (in preference order)

### A. Daemon-side (fixes the root cause — F1)

In the auto-commit daemon's pre-commit check: refuse to include any `docs/status/*.md` (and `docs/planning/*.md`) file whose size is < 120 bytes or whose content lacks a markdown H1 — UNLESS the commit message explicitly marks it (e.g. `[skeleton]`). The daemon already classifies files; this is one predicate.

- Cost: one predicate in the daemon sweep.
- Risk: a legitimately tiny file (a one-line resolution note) gets held — mitigate with the `[skeleton]` escape or a 120-byte floor that one-line notes clear.

### B. Session-side (defense in depth, no daemon change — F1)

A pre-write checklist rule for agents: a status report is written in ONE `write` call with full content — never `write` a stub "to be filled later". If a session must stop early, it writes `## Status: session ended before content; see <successor or none>` instead of leaving the file bare. This is a convention, teachable via AGENTS.md — but conventions without gates rot (this proposal exists because ten stubs shipped).

### C. Gate-side (detects, doesn't prevent — F1)

Extend `scripts/check-docs-freshness.sh` (or a sibling `check-snapshots.sh`): every `docs/status/*.md` must contain an H1 and ≥ 1 non-boilerplate line; violations listed warn-only initially, strict after one clean cycle. This catches future stubs at the next pass but cannot stop the commit.

### D. Intent tripwire (IMPLEMENTED 2026-09-29 — F2)

`scripts/check-intent.sh`: a session declares the files it intends to touch in a local `.check-intent` manifest (one glob-or-path per line, gitignored); the script diffs `git status --porcelain` against it and names every UNEXPECTED file. Advisory by default; `--strict` exits 1. Verified by canary: a foreign file is named; a manifest-covered tree is silent; a clean tree is silent. Canary-verified before trusted, per the standing gate rule.

- Cost: one `--strict` call (or habitually, a plain call) before every batch commit.
- Limits: it cannot stop the DAEMON (the daemon doesn't run it) — it makes the HUMAN/session naming step mechanical before a manual commit, and gives the daemon-proposal (A) a concrete pre-commit predicate to reuse: "every staged path must match an intent line".
- Enforcement placement (decision point for Lars): (i) discipline only (current — call it before commits), (ii) a pre-commit hook (blocks manual commits; needs the hook to tolerate the daemon's own path or the daemon deadlocks), (iii) daemon-side integration (strongest — the sweep itself checks intent before including files). Decision documented, not imposed.

## Recommendation

Implement **A** in the daemon config (the daemon is the only component that can prevent the commit), plus **C** as the safety net for whatever slips through. **B** costs one AGENTS.md bullet and is free — adopt it immediately regardless. **D** is live in-repo today as the session-side tripwire; wiring it into the daemon (iii) is the durable version of F2 prevention.

## Decision needed

A requires a change in the daemon's configuration (outside this repo). C is in-repo and can land with the next docs-health pass. B is an AGENTS.md convention edit. D's enforcement placement (discipline vs hook vs daemon) is Lars's call — the discipline mode is live and costs nothing.
