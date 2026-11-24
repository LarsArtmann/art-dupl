package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// PATH-shadow warning: art-dupl gets installed several ways on one machine
// (go install → ~/go/bin, nix profile/system-path → /nix/store, manual copies
// → ~/.local/bin), and a stale copy early in PATH silently answers every
// `art-dupl` invocation — the 2026-11-22 webphone session was debugged against
// a five-week-old ~/go/bin copy while two other builds sat unread in PATH.
// The `version` subcommand therefore reports every copy it can find.

// probeArtDuplVersion runs `<path> version --short` to label each PATH
// candidate with its own reported version. It is a var seam so tests can
// stub the child-binary probe.
var probeArtDuplVersion = func(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, path, "version", "--short").Output()
	if err != nil {
		return "unknown"
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		return "unknown"
	}

	return version
}

// artDuplExecutableName returns the binary name PATH lookups must stat,
// including the platform executable suffix.
func artDuplExecutableName() string {
	if runtime.GOOS == "windows" {
		return "art-dupl.exe"
	}

	return "art-dupl"
}

// artDuplPathCandidates returns the distinct realpaths of every executable
// named art-dupl found in the colon-separated pathEnv, in first-seen order
// (the order shell resolution would use). Symlinks are resolved before
// deduplication so a nix system-path link and its store target count once.
func artDuplPathCandidates(pathEnv string) []string {
	seen := make(map[string]bool)
	var candidates []string

	for _, dir := range filepath.SplitList(pathEnv) {
		//art-dupl:accept empty PATH element guard (filepath.SplitList may yield blanks)
		if dir == "" {
			continue
		}

		candidate := filepath.Join(dir, artDuplExecutableName())
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}

		if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
			continue
		}

		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			resolved = candidate
		}

		if !seen[resolved] {
			seen[resolved] = true
			candidates = append(candidates, resolved)
		}
	}

	return candidates
}

// runningRealpath returns the resolved path of the currently executing
// binary, or "" when it cannot be determined.
func runningRealpath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}

	return resolved
}

// warnIfMultipleArtDuplInPath prints a stderr warning naming every distinct
// art-dupl copy reachable through pathEnv when there is more than one, marking
// the copy that is running and the copy shell resolution would pick. Single
// copy (or undeterminable): silent.
func warnIfMultipleArtDuplInPath(stderr io.Writer, pathEnv string) {
	warnIfMultipleArtDuplInPathWithProbe(stderr, pathEnv, probeArtDuplVersion)
}

// warnIfMultipleArtDuplInPathWithProbe is warnIfMultipleArtDuplInPath with the
// version probe injected. The probe travels as a parameter so tests can stub
// it without writing the package-level seam var: parallel tests reassigning a
// global while other parallel tests read it through the version command is a
// data race (-race caught it under -shuffle), and parameter injection removes
// the writer entirely.
func warnIfMultipleArtDuplInPathWithProbe(stderr io.Writer, pathEnv string, probe func(path string) string) {
	candidates := artDuplPathCandidates(pathEnv)
	if len(candidates) < 2 {
		return
	}

	running := runningRealpath()

	fmt.Fprintf(stderr,
		"warning: art-dupl found %d times in PATH; the FIRST entry is what `art-dupl` runs:\n",
		len(candidates))

	for _, candidate := range candidates {
		version := "unknown"
		if candidate == running && running != "" {
			version = GetVersion()
		} else {
			version = probe(candidate)
		}

		role := ""
		if candidate == running {
			role = " — running now"
		}
		if candidate == candidates[0] && len(candidates) > 1 {
			if role != "" {
				role += " and first in PATH"
			} else {
				role = " — first in PATH, wins"
			}
		}

		fmt.Fprintf(stderr, "  %s (%s)%s\n", candidate, version, role)
	}

	fmt.Fprintln(stderr, "  remove stale copies or reorder PATH so the intended build wins")
}
