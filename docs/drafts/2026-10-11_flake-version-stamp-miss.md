## Problem

The flake's `version = "0.7.0"` (`flake.nix:35`) is not bumped on release, so
every nix build of a tagged ref stamps the wrong version: the `-X
...cmd.Version=${version}` ldflag feeds `art-dupl version` from that stale
constant. Confirmed on three tags — v0.7.1, v0.7.2, and v0.9.0 all declare
`version = "0.7.0"` (`git show v0.9.0:flake.nix | grep 'version ='`).

Downstream effect seen in CV: the nix-built detector reports `0.7.0` while
running v0.7.2 code, which is exactly the drifted-binary shape
`art-dupl-ci-gate.sh` version-gates against; CV's `nix/packages.nix`
currently works around it by sed-stamping the version into the FOD
(`ldflags = [ "-X ...cmd.Version=v0.7.2" ]`).

## Ask

Bump `version` in the same change as the tag (one line + the ldflag inherits).
If you want it mechanical, a `scripts/release-check.sh` asserting
`flake.nix`'s version equals the pending tag would close the class.

💘 Generated with Crush

<!-- FILED as https://github.com/LarsArtmann/art-dupl/issues/5 -->
