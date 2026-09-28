package cmd

import (
	"regexp"
	"strings"
	"testing"
)

// Go-line pin gate: go.mod deliberately pins the PATCH version (go 1.27.1)
// to match .golangci.yml run.go, flake.nix (pkgs.go_1_27) and the CI matrix.
// Tooling has flipped this line down to "go 1.27" five times (2026-09-19/23,
// again 2026-09-28 via an auto-commit) — the workspace/go-line-flipflop class
// BuildFlow's own preflight warns about. This test turns the next silent flip
// into a hard CI failure instead of a stale-shell mystery.
//
// If the pin must move (e.g. a deliberate Go upgrade), change it here, in
// .golangci.yml, flake.nix and the CI matrix TOGETHER, and update
// AGENTS.md + .buildflow.yml skip_steps rationale.

const expectedGoLine = "1.27.1"

func TestGoModPinsPatchVersion(t *testing.T) {
	goMod := readRepoFile(t, "go.mod")
	goLine := extractGoDirective(t, goMod)
	if goLine != expectedGoLine {
		t.Errorf(`go.mod go line drifted: got "go %s", want "go %s".

The patch pin is DELIBERATE (AGENTS.md "Go toolchain"): it aligns with
.golangci.yml run.go, flake.nix pkgs.go_1_27 and the CI matrix, and the
go-version-auto-configure vs go-mod-normalize dispositions keep fighting
over it (BuildFlow preflight workspace/go-line-flipflop). Restore the pin:
  sed -i 's/^go 1\.27$/go %s/' go.mod
and investigate which tool rewrote it before assuming this is noise.`,
			goLine, expectedGoLine, expectedGoLine)
	}
}

func TestGolangciGoVersionMatchesPin(t *testing.T) {
	config := readRepoFile(t, ".golangci.yml")
	// run: block, go: <version>
	re := regexp.MustCompile(`(?m)^\s{2}go:\s*(\S+)`)
	match := re.FindStringSubmatch(config)
	if match == nil {
		t.Fatal(".golangci.yml has no `go:` entry under run:; alignment check is blind")
	}
	if match[1] != expectedGoLine {
		t.Errorf(".golangci.yml run.go is %q but go.mod pins %q — the two must match",
			match[1], expectedGoLine)
	}
}

func extractGoDirective(t *testing.T, goMod string) string {
	t.Helper()
	for line := range strings.SplitSeq(goMod, "\n") {
		trimmed := strings.TrimSpace(line)
		if version, ok := strings.CutPrefix(trimmed, "go "); ok && !strings.Contains(trimmed, "toolchain") {
			return strings.TrimSpace(version)
		}
	}
	return ""
}
