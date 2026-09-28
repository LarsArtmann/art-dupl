# Proposal — Empty-Snapshot Prevention: Stop 0-Byte Status Reports at the Source

**Date:** 2026-09-28
**Origin:** the 2026-09-28 docs-health audit found **10 committed zero-byte (or near-zero) status snapshots** — daemon auto-commits swept empty or truncated report files into history, and each one needed a hand-written "honest one-line resolution note" during archiving. This page proposes the guard; it is NOT implemented here.

## The failure mode

1. A session creates `docs/status/<timestamp>_<name>.md` intending to fill it.
2. The session dies (user stop, crash, context switch) before content lands.
3. The daemon's heuristic sweep commits the empty file with `chore: auto-commit N file(s)`.
4. Weeks later a docs-health pass finds a title-less 0-byte file in `docs/status/` and cannot tell abandoned-skeleton from deleted-content.

Ten instances predating 2026-09-28. Each converts a "what happened here?" into archaeology.

## Proposed guards (in preference order)

### A. Daemon-side (fixes the root cause)

In the auto-commit daemon's pre-commit check: refuse to include any `docs/status/*.md` (and `docs/planning/*.md`) file whose size is < 120 bytes or whose content lacks a markdown H1 — UNLESS the commit message explicitly marks it (e.g. `[skeleton]`). The daemon already classifies files; this is one predicate.

- Cost: one predicate in the daemon sweep.
- Risk: a legitimately tiny file (a one-line resolution note) gets held — mitigate with the `[skeleton]` escape or a 120-byte floor that one-line notes clear.

### B. Session-side (defense in depth, no daemon change)

A pre-write checklist rule for agents: a status report is written in ONE `write` call with full content — never `write` a stub "to be filled later". If a session must stop early, it writes `## Status: session ended before content; see <successor or none>` instead of leaving the file bare. This is a convention, teachable via AGENTS.md — but conventions without gates rot (this proposal exists because ten stubs shipped).

### C. Gate-side (detects, doesn't prevent)

Extend `scripts/check-docs-freshness.sh` (or a sibling `check-snapshots.sh`): every `docs/status/*.md` must contain an H1 and ≥ 1 non-boilerplate line; violations listed warn-only initially, strict after one clean cycle. This catches future stubs at the next pass but cannot stop the commit.

## Recommendation

Implement **A** in the daemon config (the daemon is the only component that can prevent the commit), plus **C** as the safety net for whatever slips through. **B** costs one AGENTS.md bullet and is free — adopt it immediately regardless.

## Decision needed

A requires a change in the daemon's configuration (outside this repo). C is in-repo and can land with the next docs-health pass. B is an AGENTS.md convention edit.
