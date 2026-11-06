#!/usr/bin/env python3
"""Compare two art-dupl binaries' clone-group output on the same corpus.

Before/after corpus measurements are otherwise done by hand, three different
ways, every session (the 2026-11-06 templ-text and CSS-value sessions each
rewrote an ad-hoc analyzer). This script is the durable version: it runs both
binaries with --json, then reports what actually changed.

Reported dimensions:

  groups        clone groups present only before / only after
  member sets   identical displayed spans (`file:start-end`) under a different
                group hash — a granularity or identity change, not a fixed or
                introduced clone
  categories    per-category group-count delta (where the change lands)

Usage:
  scripts/corpus-diff.py --before ./old --after ./new [--args "-t 2"] CORPUS

Exit codes: 0 = comparison produced, 1 = a run failed, 2 = usage error.
"""

from __future__ import annotations

import argparse
import collections
import json
import os
import shlex
import subprocess
import sys
from typing import Any


def fail(message: str) -> NoReturn:  # noqa: F821 - string annotation, py3.8 safe
    print(f"corpus-diff: {message}", file=sys.stderr)
    raise SystemExit(2)


def run_binary(binary: str, corpus: str, extra_args: list[str]) -> dict[str, Any]:
    """Run one binary and return its parsed JSON document."""
    if not os.path.exists(binary):
        fail(f"binary not found: {binary}")

    cmd = [binary, "--json", *extra_args, corpus]
    env = dict(os.environ)
    env["ARTDUPL_NO_PROGRESS"] = "1"

    try:
        completed = subprocess.run(
            cmd, capture_output=True, text=True, env=env, check=False
        )
    except OSError as exc:
        fail(f"cannot execute {binary}: {exc}")
    if completed.returncode != 0 and not completed.stdout.strip():
        print(completed.stderr, file=sys.stderr)
        fail(f"{binary} failed with exit code {completed.returncode}")

    try:
        return json.loads(completed.stdout)
    except json.JSONDecodeError as exc:
        print(completed.stdout[:2000], file=sys.stderr)
        fail(f"{binary} did not emit JSON: {exc}")


def member_set(group: dict[str, Any]) -> tuple[str, ...]:
    """Displayed spans of a group, order-independent."""
    spans = []
    for clone in group.get("files") or []:
        start = clone.get("line_start")
        end = clone.get("line_end")
        spans.append(f"{clone.get('filename')}:{start}-{end}")
    return tuple(sorted(spans))


def summarize(doc: dict[str, Any]) -> dict[str, Any]:
    """Index a JSON document by group hash and by member set."""
    groups = doc.get("clone_groups") or []
    by_hash = {g.get("hash") or "": g for g in groups}
    by_members: dict[tuple[str, ...], str] = {}
    categories: collections.Counter[str] = collections.Counter()

    for group in groups:
        by_members[member_set(group)] = group.get("hash") or ""
        for clone in group.get("files") or []:
            categories[str(clone.get("category") or "unknown")] += 1

    return {
        "groups": by_hash,
        "members": by_members,
        "categories": categories,
        "total": len(groups),
        "clones": sum(len(g.get("files") or []) for g in groups),
    }


def describe(group: dict[str, Any]) -> str:
    spans = member_set(group)
    category = (group.get("files") or [{}])[0].get("category") or "unknown"
    return (
        f"  {group.get('hash') or '?':16} {group.get('size'):>5} tokens "
        f"[{category}] {spans[0] if spans else '?'}"
        + (f" (+{len(spans) - 1} more)" if len(spans) > 1 else "")
    )


def print_samples(title: str, groups: list[dict[str, Any]], limit: int) -> None:
    if not groups:
        return
    print(f"\n{title} ({len(groups)}):")
    for group in groups[:limit]:
        print(describe(group))
    if len(groups) > limit:
        print(f"  ... and {len(groups) - limit} more")


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(
        description="Compare two art-dupl binaries' clone groups on one corpus.",
        epilog='Example: scripts/corpus-diff.py --before ./old --after ./new --args "-t 2" ~/projects/webphone',
    )
    parser.add_argument("--before", required=True, help="reference binary")
    parser.add_argument("--after", required=True, help="new binary")
    parser.add_argument(
        "--args",
        default="",
        help='extra art-dupl flags for BOTH runs (e.g. "-t 2 --type-aware")',
    )
    parser.add_argument(
        "--samples", type=int, default=8, help="groups to print per bucket (default 8)"
    )
    parser.add_argument(
        "--json", dest="json_out", help="write the structured diff to this path"
    )
    parser.add_argument("corpus", help="directory or file to analyze")
    args = parser.parse_args(argv)

    extra_args = shlex.split(args.args)
    before = summarize(run_binary(args.before, args.corpus, extra_args))
    after = summarize(run_binary(args.after, args.corpus, extra_args))

    removed_hashes = [h for h in before["groups"] if h not in after["groups"]]
    added_hashes = [h for h in after["groups"] if h not in before["groups"]]

    removed = [before["groups"][h] for h in removed_hashes]
    added = [after["groups"][h] for h in added_hashes]

    # A member set present on both sides under different hashes is not a
    # fixed/introduced clone: the displayed span is unchanged, the identity is.
    rekeyed = [
        {
            "members": list(members),
            "before": before["members"][members],
            "after": after["members"][members],
        }
        for members in before["members"].keys() & after["members"].keys()
        if before["members"][members] != after["members"][members]
    ]

    print(f"corpus: {args.corpus}")
    print(f"flags : {' '.join(extra_args) or '(defaults)'}")
    print(
        f"before: {before['total']} groups / {before['clones']} clones\n"
        f"after : {after['total']} groups / {after['clones']} clones"
    )

    print_samples("removed groups", sorted(removed, key=describe), args.samples)
    print_samples("added groups", sorted(added, key=describe), args.samples)

    if rekeyed:
        print(f"\nsame spans, different hash ({len(rekeyed)}):")
        for entry in rekeyed[: args.samples]:
            print(
                f"  {entry['before'] or '?'} -> {entry['after'] or '?'} {entry['members'][0]}"
            )

    categories = sorted(before["categories"].keys() | after["categories"].keys())
    changed = [
        (name, before["categories"][name], after["categories"][name])
        for name in categories
        if before["categories"][name] != after["categories"][name]
    ]
    if changed:
        print("\nclone counts by category (before -> after):")
        for name, old, new in changed:
            print(f"  {name:12} {old:>6} -> {new}")
    else:
        print("\nclone counts by category: unchanged")

    if args.json_out:
        payload = {
            "corpus": args.corpus,
            "args": extra_args,
            "before": {"groups": before["total"], "clones": before["clones"]},
            "after": {"groups": after["total"], "clones": after["clones"]},
            "removed": removed_hashes,
            "added": added_hashes,
            "rekeyed": rekeyed,
            "categories": {
                name: {
                    "before": before["categories"][name],
                    "after": after["categories"][name],
                }
                for name in categories
            },
        }
        with open(args.json_out, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, indent=2, sort_keys=True)
            handle.write("\n")
        print(f"\nstructured diff written to {args.json_out}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
