# GitHub Actions Distribution Options — Comparison Report

**Created:** 2026-07-28
**Author:** Parakletos (research)
**Question:** *How can we make art-dupl easily and superbly available in GitHub Actions?*
**Status:** Research complete — recommendation locked (see bottom).

---

## TL;DR

art-dupl is **already distributed everywhere** (goreleaser binaries, Docker multi-arch images, Homebrew, Nix, Scoop, deb/rpm/apk, pre-commit, Go SDK) — **but not as a first-class GitHub Action.** Users today must copy a template workflow that runs `go install @latest`, which is slow (~30–60 s compile), requires a Go toolchain, and **drifts in version**. The single highest-leverage change is shipping a **composite action** (`action.yml`) that downloads the pinned, checksum-verified binary and runs `art-dupl check` in one line.

**One blocker first:** the last three releases (`v0.4.0`, `v0.5.0`, `v0.5.1`) have **zero release assets** — `RELEASE.md` step 6 creates releases manually via `gh release create --generate-notes`, bypassing the tag-triggered `release.yml` goreleaser workflow. Any binary-download design depends on goreleaser actually publishing assets again. (Confirmed: `gh release view v0.5.1 --json assets` → `{"assets":[]}`; only `v0.1.0` was fully published.)

---

## Current state (verified facts)

| Artifact | Where | Status |
| --- | --- | --- |
| Goreleaser config | `.goreleaser.yaml` | ✅ Full: binaries (linux/darwin/windows × amd64/arm64), archives, `checksums.txt`, cosign sigs, SBOMs, Docker manifests, deb/rpm/apk, Nix, Brew, Scoop |
| Tag release workflow | `.github/workflows/release.yml` | ✅ Exists, triggers on `v*` tags, runs goreleaser |
| Releases cut | GitHub Releases | ⚠️ **Broken** — v0.4.0/v0.5.0/v0.5.1 created manually with **no assets**; only v0.1.0 has binaries |
| Docker image | `ghcr.io/larsartmann/art-dupl` | ⚠️ Likely not pushed for recent tags (goreleaser didn't run). Distroless nonroot (`Dockerfile`) |
| GA workflow template | `templates/github-actions-duplicate-check.yml` | ⚠️ Uses `go install @latest` (slow + version drift) |
| Self-check workflow | `.github/workflows/art-dupl-check.yml` | ⚠️ Same `go install @latest` problem |
| **Reusable `action.yml`** | — | ❌ **Does not exist** — the gap |
| README CI section | `README.md:91` | Points only at the template, not at a `uses:` action |

**Asset naming** (from v0.1.0): `art-dupl_<version>_<OsTitle>_<arch>.tar.gz` where `OsTitle ∈ {Linux,Darwin,Windows}`, `arch ∈ {x86_64,arm64}` (Windows uses `.zip`). *Caveat (verified):* the v0.1.0 archive is named `art-dupl_.0.1.0_Linux_x86_64.tar.gz` in the release, but the **same file appears in `checksums.txt` as `art-dupl_ 0.1.0_Linux_x86_64.tar.gz`** — a **space** before the version in `checksums.txt` vs a **dot** in the asset URL. Name reconstruction from the version string is therefore unreliable; the action must **discover the exact asset via the Releases API and match the checksum by suffix**.

---

## Options compared

Legend: 🟢 strong · 🟡 acceptable · 🔴 weak.

| # | Option | One-line UX | Cold-start speed | Version pin / reproducible | No Go toolchain needed | Cross-runner (linux/darwin × x64/arm64) | Supply-chain (checksum/cosign) | Auto-tracks releases | Maintenance effort | Verdict |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **1** | **Composite action — download prebuilt binary** | 🟢 `uses: LarsArtmann/art-dupl@v1` | 🟢 ~2–4 s | 🟢 pin by tag + sha256 | 🟢 yes | 🟢 all GH-hosted runners | 🟢 verify vs `checksums.txt` (cosign opt-in) | 🟢 via tags | 🟡 one script + `action.yml` | **✅ RECOMMENDED** |
| 2 | Docker container action | 🟢 `uses:` | 🟡 image pull ~10–30 s | 🟢 image digest | 🟢 yes | 🔴 Linux only (container actions can't run on macOS/Windows runners) | 🟢 signed image | 🟡 rebuild manifests | 🟡 low | Good alt if you only need Linux; lose macOS/Windows |
| 3 | `go install` in composite/template (status quo) | 🟡 copy template | 🔴 30–60 s compile | 🔴 `@latest` drifts (fixable with `@<tag>`) | 🔴 needs Go | 🟢 yes | 🔴 none | 🟡 manual | 🟢 none | What we have today. Slow + brittle |
| 4 | Reusable workflow (`workflow_call`) | 🟡 `uses: …/art-dupl/.github/workflows/…yml@v1` | 🟢 (if it uses #1) | 🟢 | 🟢 | 🟢 | 🟢 | 🟢 | 🟡 low | **Pair with #1** for "batteries-included" job; less flexible than an action |
| 5 | Homebrew install step | 🔴 multi-line, brew overhead | 🔴 ~30–60 s | 🟡 tap pin | 🟢 | 🟡 macOS native; Linuxbrew on Linux | 🟡 formula checksum | 🔴 manual formula bump | 🟡 med | Overkill for CI; great for local dev (already shipped) |
| 6 | Nix install step (`cachix/install-nix-action` + flake) | 🟡 multi-line | 🔴 Nix setup heavy | 🟢 flake lock | 🟢 | 🟢 | 🟢 reproducible/hermetic | 🟢 flake follows repo | 🟡 med | Superb for Nix shops; too heavy for general CI |
| 7 | Raw `curl` binary in user's own workflow | 🔴 DIY every repo | 🟢 fast | 🟢 by URL | 🟢 | 🟢 | 🟡 manual checksum | 🔴 DIY | n/a | What #1 automates — don't ask users to do this |
| 8 | Go SDK (`pkg/artdupl`) in a Go-based action/job | 🟡 import + write code | 🔴 build cost | 🟢 go.mod pin | 🔴 Go project | 🟢 | 🟢 | 🟡 | 🔴 high | Already exists for programmatic use; not a CI UX |

---

## Detailed assessment of the top 3

### ✅ Option 1 — Composite action (download prebuilt binary) — **RECOMMENDED**

**Shape:** `action.yml` at repo root + `scripts/install-art-dupl.sh`. The script maps `RUNNER_OS` × `RUNNER_ARCH` → archive, downloads `checksums.txt`, finds the exact asset name by parsing the checksum file (robust to goreleaser naming quirks), verifies sha256, extracts into `$RUNNER_TOOL_CACHE/art-dupl/<version>/<arch>` (free caching across runs), prepends to `$GITHUB_PATH`. Graceful fallback to `go install …@<tag>` if the release has no assets (keeps it working today, with a loud warning).

**User experience (the goal):**
```yaml
- uses: LarsArtmann/art-dupl@v1
  with:
    threshold: '15'
    # command: check (default) · working-directory: . (default)
```

**Pros**
- One line, discoverable, Marketplace-listable, version-pinned by git ref + checksum-verified.
- Runs on **all** GitHub-hosted runners (ubuntu/macos × x64/arm64, windows).
- No Go toolchain in the consumer repo → usable by non-Go projects (art-dupl supports `.go` + `.templ`).
- Tool-cache caching → near-zero cost on warm runs.
- Reuses the existing goreleaser output (no new build pipeline).

**Cons / risks**
- Requires recent releases to actually publish assets (the blocker above).
- Composite actions can't run post-step cleanup as elegantly as JS/Docker actions (fine here).
- Needs `actionlint` in CI to keep `action.yml` valid.

**Build cost:** ~1 `action.yml` + ~1 install script + `actionlint` CI step. **Low.**

### Option 2 — Docker container action

**Shape:** `runs.using: docker`, `image: 'docker://ghcr.io/larsartmann/art-dupl:v1'` (or `Dockerfile`).

**Pros:** trivial `action.yml`, hermetic, image is already built & signed by goreleaser.
**Cons:** **Linux runners only** (container actions don't run on macOS/Windows runners) — defeats cross-platform. Workspace mounts at `/github/workspace` (path friction). Slower pull than a 3 MB binary.

**Verdict:** Strong only if you scope CI to Linux. Loses too much vs Option 1.

### Option 3 — Status quo (`go install`)

Already shipped in `templates/github-actions-duplicate-check.yml` and `.github/workflows/art-dupl-check.yml`. **Keep as the fallback path inside Option 1**, but stop offering it as the primary UX.

---

## The recommended combination

> **Option 1 (composite action) as the flagship**, **Option 4 (reusable workflow) as a convenience layer**, and **Option 3 (`go install@<tag>`) only as the action's internal fallback** when a release lacks binaries. Option 2 (Docker) and 6 (Nix) stay available for users who want them — no extra work, they already exist.

This gives every audience the best path:
- **Most users:** `uses: LarsArtmann/art-dupl@v1` → fast, pinned, verified.
- **"Just give me a job" users:** `uses: …/art-dupl/.github/workflows/art-dupl.yml@v1` (reusable).
- **Nix shops:** flake (existing).
- **Pre-commit users:** `.pre-commit-hooks.yaml` (existing).
- **Programmatic:** `pkg/artdupl` SDK (existing).

---

## Hard prerequisites (must fix regardless of option)

1. **Restore goreleaser-published releases.** Either let `release.yml` own tag releases (stop the manual `gh release create` in `RELEASE.md` step 6 (line 51)), or add a `gh release upload`/`goreleaser release` step to the manual flow. **Options 1, 2, 4, 7 all depend on this.**
2. **Pin a moving tag (`v1`, `v1.2`).** goreleaser already emits `v{{ .Major }}` manifests for Docker; mirror this with a git moving tag (`v1`) so users can `uses: …@v1` and get patch updates safely.
3. **Verify ghcr images are actually being pushed** for recent tags (couldn't confirm — `read:packages` scope denied; user to check `ghcr.io/larsartmann/art-dupl`).

---

## Non-goals / what NOT to build

- A **JavaScript action** (needs Node build + `@actions/*` deps) — overkill; composite + bash is enough and ships with zero npm.
- A **Marketplace listing** right away — optional follow-up; `uses:` works without it.
- Replacing Homebrew/Nix/Scoop/deb/rpm — those serve other surfaces; leave them.
