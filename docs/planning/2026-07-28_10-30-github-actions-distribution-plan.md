# Execution Plan — art-dupl GitHub Action

**Created:** 2026-07-28
**Goal:** Make art-dupl available in GitHub Actions as a one-line, fast, pinned, verified `uses: LarsArtmann/art-dupl@v1`.
**Companion:** `docs/research/github-actions-distribution-options.md` (option comparison + rationale).
**Recommendation:** Composite action (download prebuilt binary) + reusable workflow + `go install@<tag>` fallback.

---

## Outcome & success criteria

After this plan, a user can add this to any repo (Go or not) and get clone-gated CI:

```yaml
- uses: LarsArtmann/art-dupl@v1
  with:
    threshold: '15'
```

**Done when:**
- [ ] `action.yml` exists at repo root and passes `actionlint`.
- [ ] The action installs the pinned binary (checksum-verified) on ubuntu/macos × x64/arm64 runners in <5 s warm, <10 s cold.
- [ ] Fallback to `go install @<tag>` works when a release has no assets (loud warning, same pinned version — no drift).
- [ ] `art-dupl check` runs in `working-directory`, creates a baseline on first run, fails the job on new clones (toggleable), and uploads a report artifact.
- [ ] Recent releases publish real binary assets (goreleaser owns tag releases).
- [ ] A moving `v1` tag tracks the latest stable.
- [ ] README "CI/CD Integration" section leads with `uses:`, and a reusable workflow `art-dupl.yml` exists.
- [ ] The repo's own `art-dupl-check.yml` dogfoods the action (replaces `go install @latest`).

---

## Phases (Pareto-ordered)

| Phase | What | Impact | Effort | Dependency |
| --- | --- | --- | --- | --- |
| 0 | Unblock releases (publish assets again) | 🔴 Blocker for everything binary-based | S (30 min) | None |
| 1 | Ship the composite action (`action.yml` + install script) | 🟢 90% of the value | M (~3–4 h) | Phase 0 for full speed |
| 2 | Validate in CI (`actionlint` + smoke test matrix) | 🟢 trust | S | Phase 1 |
| 3 | Reusable workflow + README + Marketplace metadata | 🟡 reach/UX | S | Phase 1 |
| 4 | Dogfood in own repo; deprecate `go install @latest` templates | 🟢 consistency | S | Phase 1 |

---

## Phase 0 — Unblock releases (prerequisite)

**Problem:** `RELEASE.md` step 6 (line 51) cuts releases with `gh release create … --generate-notes`, which creates an **empty** release. The tag then exists, so `release.yml`'s `on: push: tags: v*` *may* still fire goreleaser — but if goreleaser's `release:` step finds the release already exists, behavior is version-dependent (often it errors or skips asset upload). Net result: `v0.4.0`/`v0.5.0`/`v0.5.1` have `{"assets":[]}`.

**Fix (choose one — Option A recommended):**

**A. Let goreleaser own tag releases (canonical).**
- Edit `RELEASE.md` step 6: replace `gh release create …` with just **pushing the signed tag** (`git push --follow-tags`). goreleaser (triggered by `release.yml`) creates the release + uploads all assets automatically.
- Verify the `release.yml` job has `permissions: contents: write` (it does — confirmed).

**B. Keep manual releases but upload assets.**
- After `gh release create …`, also run:
  ```bash
  goreleaser release --clean --skip=docker --skip=validate   # or full, if secrets are set
  # or, minimal: build + upload archives + checksums.txt
  gh release upload vX.Y.Z dist/*.tar.gz dist/*.zip dist/checksums.txt --clobber
  ```

**Steps:**
1. [ ] Decide A vs B; update `RELEASE.md` accordingly.
2. [ ] Cut a patch release (e.g. `v0.5.2`) using the new flow.
3. [ ] **Verify:** `gh release view v0.5.2 --json assets --jq '.assets[].name'` lists `art-dupl_0.5.2_Linux_x86_64.tar.gz`, `…Darwin_arm64.tar.gz`, `…Windows_x86_64.zip`, `checksums.txt`, etc.
4. [ ] (optional) Confirm ghcr images pushed: `gh api /users/LarsArtmann/packages/container/art-dupl/versions` (needs `read:packages`).
5. [ ] Create moving tag: `git tag -s v1 v0.5.2 && git push origin v1` (document re-pointing on every minor).

**Exit gate:** `gh release view <latest> --json assets` shows ≥ the 3 archives + `checksums.txt`.

---

## Phase 1 — Ship the composite action

**Files to create:** `action.yml` (repo root), `scripts/install-art-dupl.sh`.

### 1.1 `action.yml`

```yaml
name: art-dupl
description: Fast, type-aware code clone detector for Go & templ — runs `art-dupl check` against a committed baseline and fails on new clones.
author: Lars Artmann
branding:
  icon: copy
  color: blue

inputs:
  version:
    description: art-dupl version to install. 'latest' resolves to the newest GitHub release. Pin to a tag (e.g. v0.5.2) for full reproducibility.
    default: latest
  command:
    description: art-dupl subcommand to run. Use '' to pass raw `args` only. Ignored when install-only=true.
    default: check
  threshold:
    description: Clone threshold (-t). Empty = art-dupl default (5).
    default: ''
  args:
    description: Extra args appended to the command.
    default: ''
  working-directory:
    description: Directory to analyze.
    default: .
  install-only:
    description: If 'true', only install art-dupl onto PATH and skip running it.
    default: 'false'
  create-baseline-if-missing:
    description: If 'true' and no baseline exists, record one (commit .art-dupl-baseline.json).
    default: 'true'
  fail-on-new-clones:
    description: If 'true', a non-zero `check` exit fails the step.
    default: 'true'
  upload-report:
    description: Upload reports/ as an artifact (always, for debugging).
    default: 'true'
  verify-signature:
    description: Cosign-verify the archive against its .sig+.pem (requires cosign on PATH; opt-in).
    default: 'false'

outputs:
  version:
    description: Resolved art-dupl version (e.g. 0.5.2).
    value: ${{ steps.install.outputs.version }}
  path:
    description: Directory the binary was installed to.
    value: ${{ steps.install.outputs.path }}

runs:
  using: composite
  steps:
    - name: Install art-dupl
      id: install
      shell: bash
      env:
        INPUT_VERSION: ${{ inputs.version }}
        INPUT_VERIFY_SIGNATURE: ${{ inputs.verify-signature }}
      run: ${{ github.action_path }}/scripts/install-art-dupl.sh

    - name: Run art-dupl
      if: ${{ inputs.install-only != 'true' }}
      shell: bash
      working-directory: ${{ inputs.working-directory }}
      env:
        THRESHOLD: ${{ inputs.threshold }}
        ARGS: ${{ inputs.args }}
        CMD: ${{ inputs.command }}
        CREATE_BASELINE: ${{ inputs.create-baseline-if-missing }}
        FAIL_ON_NEW: ${{ inputs.fail-on-new-clones }}
      run: |
        set -uo pipefail
        extra=()
        [ -n "$THRESHOLD" ] && extra+=( -t "$THRESHOLD" )
        # shellcheck disable=SC2086
        case "$CMD" in
          check)
            if [ "$CREATE_BASELINE" = "true" ] && [ ! -f .art-dupl-baseline.json ]; then
              echo "::notice::No .art-dupl-baseline.json found — recording initial baseline. Commit it to enable gating."
              art-dupl baseline . "${extra[@]}" || exit 1
              exit 0
            fi
            art-dupl check . "${extra[@]}" $ARGS
            code=$?
            ;;
          "")
            art-dupl "${extra[@]}" $ARGS; code=$?
            ;;
          *)
            art-dupl "$CMD" . "${extra[@]}" $ARGS; code=$?
            ;;
        esac
        if [ "$code" -ne 0 ] && [ "$FAIL_ON_NEW" != "true" ]; then
          echo "::warning::art-dupl exited $code (fail-on-new-clones=false)."
          exit 0
        fi
        exit "$code"

    - name: Upload report
      if: ${{ inputs.install-only != 'true' && inputs.upload-report == 'true' && always() }}
      uses: actions/upload-artifact@v4
      with:
        name: art-dupl-report
        path: ${{ inputs.working-directory }}/reports/
        if-no-files-found: ignore
```

### 1.2 `scripts/install-art-dupl.sh`

Resolves version → maps runner → downloads `checksums.txt` → finds exact asset name (robust to goreleaser naming quirks) → sha256-verifies → extracts to tool cache → prepends PATH. Falls back to `go install @<tag>` if assets are absent.

```bash
#!/usr/bin/env bash
# Installs art-dupl into the GitHub Actions tool cache and prepends it to PATH.
# Fallback: go install @<tag> (same pinned version; needs Go on PATH).
set -euo pipefail

VERSION_INPUT="${INPUT_VERSION:-latest}"
REPO="LarsArtmann/art-dupl"
MODULE="github.com/LarsArtmann/art-dupl/cmd/art-dupl"

log() { printf '::group::%s\n%s\n::endgroup::\n' "$1" "$2" >&2 || true; }
die() { printf '::error::%s\n' "$*" >&2; exit 1; }

# --- resolve version -> tag + bare version -------------------------------
resolve_version() {
  local input="$1"
  if [ "$input" = "latest" ]; then
    # GitHub redirects /releases/latest to the newest non-prerelease tag.
    curl -fsSL -o /dev/null -w '%{url_effective}' \
      "https://github.com/$REPO/releases/latest" \
      | sed 's#.*/tag/##'
  else
    # strip optional leading 'v'
    printf 'v%s' "${input#v}"
  fi
}
TAG="$(resolve_version "$VERSION_INPUT")"
VER="${TAG#v}"

# --- map runner OS/arch -> goreleaser tokens -----------------------------
case "${RUNNER_OS:-Linux}" in
  Linux)   OS=Linux;   EXT=tar.gz ;;
  macOS)   OS=Darwin;  EXT=tar.gz ;;
  Windows) OS=Windows; EXT=zip    ;;
  *) die "Unsupported RUNNER_OS: ${RUNNER_OS:-<unset>}" ;;
esac
case "${RUNNER_ARCH:-X64}" in
  X64|amd64) ARCH=x86_64 ;;
  ARM64|arm64) ARCH=arm64 ;;
  *) die "Unsupported RUNNER_ARCH: ${RUNNER_ARCH:-<unset>}" ;;
esac

# --- install dir (tool cache => cached across runs) ----------------------
TOOLCACHE="${RUNNER_TOOL_CACHE:-$HOME/.cache/actions-tool-cache}"
DEST="$TOOLCACHE/art-dupl/$VER/$ARCH"
if [ -x "$DEST/art-dupl" ] || [ -x "$DEST/art-dupl.exe" ]; then
  log "art-dupl cached" "$DEST — reusing"
else
  mkdir -p "$DEST"
  BASE="https://github.com/$REPO/releases/download/$TAG"

  # 1. Try the prebuilt binary (checksum-verified).
  if curl -fsSL "$BASE/checksums.txt" -o /tmp/checksums.txt 2>/dev/null; then
    # Use the Releases API to get the EXACT browser_download_url — never
    # construct URLs from the version string. (v0.1.0's filename contains a
    # space in checksums.txt but a dot in the asset URL, so name-reconstruction
    # is unreliable. jq ships preinstalled on all GitHub-hosted runners.)
    SUFFIX="_${OS}_${ARCH}.${EXT}"
    curl -fsSL "https://api.github.com/repos/$REPO/releases/tags/$TAG" -o /tmp/release.json
    ASSET_URL=$(jq -r --arg s "$SUFFIX" \
      '.assets[] | select(.name|endswith($s)) | .browser_download_url' /tmp/release.json | head -1)
    [ -n "$ASSET_URL" ] || die "No release asset ending in '$SUFFIX' for $TAG"

    ASSET_FILE="$(basename "$ASSET_URL")"
    curl -fsSL "$ASSET_URL" -o "/tmp/$ASSET_FILE"

    # Verify by SUFFIX match: the checksum filename may differ from the asset
    # name (space-vs-dot), so match the invariant suffix and take field $1 (hash).
    expected=$(grep -E "_${OS}_${ARCH}\\.${EXT}\$" /tmp/checksums.txt | awk '{print $1}' | head -1)
    [ -n "$expected" ] || die "No checksum line matching *_${OS}_${ARCH}.${EXT}"
    actual=$(sha256sum "/tmp/$ASSET_FILE" | awk '{print $1}')
    [ "$expected" = "$actual" ] || die "checksum mismatch (want $expected got $actual)"

    if [ "$EXT" = "zip" ]; then unzip -o "/tmp/$ASSET_FILE" -d "$DEST"; else tar -xzf "/tmp/$ASSET_FILE" -C "$DEST"; fi

    if [ "${INPUT_VERIFY_SIGNATURE:-false}" = "true" ] && command -v cosign >/dev/null 2>&1; then
      curl -fsSL "$ASSET_URL.sig" -o "/tmp/$ASSET_FILE.sig"
      curl -fsSL "$ASSET_URL.pem" -o "/tmp/$ASSET_FILE.pem"
      cosign verify-blob --signature "/tmp/$ASSET_FILE.sig" --certificate "/tmp/$ASSET_FILE.pem" \
        --certificate-identity-regexp "https://github.com/$REPO/.+" \
        --certificate-oidc-issuer "https://token.actions.githubusercontent.com" "/tmp/$ASSET_FILE" \
        || die "cosign verification failed for $ASSET_FILE"
    fi

  # 2. Fallback: build from source at the SAME pinned tag (no version drift).
  else
    printf '::warning::%s\n' "Release $TAG has no checksums.txt/assets — falling back to 'go install $MODULE@$TAG' (slower, needs Go). See docs/planning/…github-actions-action-plan Phase 0." >&2
    go install "$MODULE@$TAG"
    mkdir -p "$DEST"
    bin="$(go env GOPATH)/bin"
    cp "$bin"/art-dupl* "$DEST"/ 2>/dev/null || ln -s "$bin" "$DEST"
  fi
fi

printf '%s\n' "$DEST" >> "$GITHUB_PATH"
printf 'version=%s\n' "$VER" >> "$GITHUB_OUTPUT"
printf 'path=%s\n' "$DEST" >> "$GITHUB_OUTPUT"
"$DEST/art-dupl" version || "$DEST/art-dupl.exe" version || true
log "art-dupl ready" "v$VER at $DEST"
```

**Verification steps (Phase 1):**
1. [ ] `chmod +x scripts/install-art-dupl.sh`
2. [ ] `bash -n scripts/install-art-dupl.sh` (syntax check)
3. [ ] ShellCheck: `shellcheck -S warning scripts/install-art-dupl.sh` → clean (or `# shellcheck disable` with reason)
4. [ ] Local dry-run (without GHA env): `RUNNER_OS=Linux RUNNER_ARCH=X64 INPUT_VERSION=v0.1.0 bash scripts/install-art-dupl.sh` (v0.1.0 has real assets — should download+verify+extract).

---

## Phase 2 — Validate in CI

Add `.github/workflows/action-validation.yml`:

```yaml
name: Action Validation
on:
  push:
    paths: ['action.yml', 'scripts/install-art-dupl.sh', '.github/workflows/action-validation.yml']
  pull_request:
    paths: ['action.yml', 'scripts/install-art-dupl.sh']
jobs:
  actionlint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          curl -sSL https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash | bash
          ./actionlint -color
  shellcheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ludeeus/action-shellcheck@2.0.0
        env:
          SHELLCHECK_OPTS: -S warning
  smoke:
    needs: [actionlint, shellcheck]
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, macos-13, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: ./                # exercises the action itself
        with:
          version: v0.1.0       # known-good assets
          install-only: 'true'
      - run: art-dupl version
```

**Exit gate:** actionlint + shellcheck clean; `art-dupl version` prints on all 4 runner matrices.

---

## Phase 3 — Reusable workflow + docs + Marketplace

**3.1 Reusable workflow** `.github/workflows/art-dupl.yml` (batteries-included, callable via `uses: LarsArtmann/art-dupl/.github/workflows/art-dupl.yml@v1`):

```yaml
name: art-dupl
on:
  workflow_call:
    inputs:
      threshold: { required: false, type: string, default: '15' }
      version:   { required: false, type: string, default: 'latest' }
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: LarsArtmann/art-dupl@v1
        with:
          version: ${{ inputs.version }}
          threshold: ${{ inputs.threshold }}
```

**3.2 README** — rewrite `## CI/CD Integration` (`README.md:91`) to lead with the action:
```markdown
## CI/CD Integration

**GitHub Action (recommended):**
```yaml
- uses: LarsArtmann/art-dupl@v1
  with: { threshold: '15' }
```
**Reusable workflow · Pre-commit · SARIF · SDK** — see [docs](docs/research/github-actions-distribution-options.md).
```
Keep the existing baseline/SARIF prose below it.

**3.3 Marketplace** (optional follow-up): `gh release create` is not needed; publish via the Marketplace "Publish action" UI once `v1` tag exists. Tag categories: `go`, `code-quality`.

---

## Phase 4 — Dogfood + deprecate `go install @latest`

1. [ ] Rewrite `.github/workflows/art-dupl-check.yml` to `uses: ./` (or `LarsArtmann/art-dupl@v1`) instead of `setup-go` + `go install @latest`.
2. [ ] Update `templates/github-actions-duplicate-check.yml` to either (a) point at the action, or (b) mark it "legacy — prefer `uses:`". Add a `## Legacy (go install)` note.
3. [ ] `templates/pre-commit-hook.yaml` uses `language: golang` (compiles) — consider adding a system-hook variant that calls the prebuilt binary; document both.

**Exit gate:** repo's own CI runs through the action; `grep -R "go install.*art-dupl" .github templates` returns only intentional fallback paths.

---

## Risks & mitigations

| Risk | Mitigation |
| --- | --- |
| Recent releases still have no assets when action ships | Fallback to `go install @<tag>` keeps it functional; warning tells users to upgrade the release process (Phase 0) |
| goreleaser renames assets | Script parses `checksums.txt` for the real name — never hardcodes |
| macOS/Windows runner arch mismatch | `RUNNER_ARCH` map covers X64/ARM64; smoke matrix proves it |
| Composite-action `working-directory` + upload path | Upload uses `${{ inputs.working-directory }}/reports/` |
| Cosign not present | `verify-signature` is opt-in; auto-skipped if cosign absent |
| Moving `v1` tag points at a broken release | Re-point policy documented; `release.yml` gates on build/tests before tag |

---

## Out of scope (deliberate)

- JavaScript action (no benefit over composite + bash).
- Auto-bumping the `v1` tag via bot — manual, deliberate re-point on each stable release.
- Replacing Homebrew/Nix/Scoop/deb/rpm distribution.
- Marketplace listing before `v1` is battle-tested in the repo's own CI.
