# Homebrew Formula for art-dupl

This directory contains the Homebrew formula for installing art-dupl.

## Installation

### From Tap (Recommended)

```bash
brew tap LarsArtmann/art-dupl
brew install art-dupl
```

### From Source (Development)

If you want to install the latest development version:

```bash
brew install --build-from-source LarsArtmann/art-dupl/art-dupl
```

## Updating the Formula

When a new release is created:

1. Download the release archives for each platform
2. Calculate SHA256 checksums:
   ```bash
   sha256sum art-dupl_1.0.0_darwin_arm64.tar.gz
   sha256sum art-dupl_1.0.0_darwin_amd64.tar.gz
   sha256sum art-dupl_1.0.0_linux_arm64.tar.gz
   sha256sum art-dupl_1.0.0_linux_amd64.tar.gz
   ```
3. Update the formula with new version and checksums
4. Test locally:
   ```bash
   brew audit --strict --online art-dupl
   brew install --build-from-source ./art-dupl.rb
   brew test art-dupl
   ```

## Shell Completions

The formula automatically installs shell completions for:
- Bash
- Zsh
- Fish

## Verification

After installation, verify with:

```bash
art-dupl --version
art-dupl stats --help
```
