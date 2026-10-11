// Command daemon-gate is the auto-commit daemon's pre-commit gate: a single
// static Go binary (no interpreter — the daemon service PATH has no bash) that
// refuses to let a broken tree enter git history.
//
// Semantics (Q1 decision doc: docs/status/2026-10-11_*daemon-gate-q1*):
//   - fail-OPEN on infrastructure errors (no Go toolchain, no GOCACHE): the
//     gate must never brick the daemon — a hook that fails for non-verdict
//     reasons is the exact retry-forever class that got hooks disabled
//     (PMA ce0f638). Infrastructure failures print an actionable stderr
//     message and exit 0.
//   - fail-CLOSED on verdicts: `go build ./...` red, or the jsonv2gate AST
//     scanner red, exits 1 so `git commit` refuses and the daemon's failure
//     cooldown absorbs the red window.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	exitPass = 0
	exitRed  = 1

	// envDisable is the kill switch: PMA_GATE_DISABLE=1 turns the gate into
	// a no-op for emergencies without redeploying the hooks directory.
	envDisable = "PMA_GATE_DISABLE"

	// envGoBin pins the Go toolchain for environments whose PATH does not
	// carry go (the daemon service PATH is pma, git, coreutils only).
	envGoBin = "PMA_GATE_GO"

	// jsonv2gatePackage is art-dupl's production-wire AST gate. It runs only
	// when the directory exists, so the same binary is fleet-generic: every
	// Go module gets the build gate, art-dupl additionally gets the
	// encoding/json/v2 wire-regression scanner.
	jsonv2gateDir = "internal/jsonv2gate"
	jsonv2gatePkg = "./internal/jsonv2gate/"

	checkTimeout = 10 * time.Minute
)

// Infrastructure failures (fail-open) and verdict failures (fail-closed) are
// distinct static errors so callers and logs can tell them apart.
var (
	errNoToolchain  = errors.New("no Go toolchain found")
	errPinnedGoBad  = errors.New("pinned Go binary is not an executable file")
	errCheckTimeout = errors.New("gate check exceeded its time budget and was killed")
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Getenv, exec.LookPath))
}

// run is the testable entry point; it returns the process exit code.
func run(stdout, stderr io.Writer, getenv func(string) string, lookPath func(string) (string, error)) int {
	if getenv(envDisable) == "1" {
		fmt.Fprintln(stderr, "daemon-gate: disabled via "+envDisable+"=1, passing")

		return exitPass
	}

	goBin, err := resolveGo(getenv, lookPath)
	if err != nil {
		fmt.Fprintf(stderr, "daemon-gate: PASS (no gate): %v\n", err)

		return exitPass
	}

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "daemon-gate: PASS (no gate): cannot determine working directory: %v\n", err)

		return exitPass
	}

	if _, err := os.Stat(filepath.Join(workdir, "go.mod")); err != nil {
		return exitPass
	}

	checks := planChecks(workdir)
	start := time.Now()

	for _, check := range checks {
		if err := runCheck(goBin, check, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "daemon-gate: REFUSED (red %s after %s): %v\n",
				check.name, time.Since(start).Round(time.Second), err)
			fmt.Fprintln(stderr, "daemon-gate: the commit was blocked because the tree does not pass its gate;")
			fmt.Fprintln(stderr, "daemon-gate: fix or stash the red state; bypass one commit with "+envDisable+"=1")

			return exitRed
		}
	}

	elapsed := time.Since(start).Round(time.Millisecond)
	fmt.Fprintf(stderr, "daemon-gate: green (%d check(s), %s)\n", len(checks), elapsed)

	return exitPass
}

// check is one gate verdict: a name for messages and the go subcommand args.
type check struct {
	name string
	args []string
}

// planChecks derives the check list from the repo shape: every Go module gets
// `go build ./...`; art-dupl-shaped modules additionally get the jsonv2gate
// AST scanner (the encoding/json/v2 Duration/wire regression class).
func planChecks(wd string) []check {
	checks := []check{{name: "build", args: []string{"build", "./..."}}}

	if info, err := os.Stat(filepath.Join(wd, jsonv2gateDir)); err == nil && info.IsDir() {
		checks = append(checks, check{name: "jsonv2gate", args: []string{"test", "-count=1", jsonv2gatePkg}})
	}

	return checks
}

// resolveGo finds the Go toolchain without assuming PATH carries it.
func resolveGo(getenv func(string) string, lookPath func(string) (string, error)) (string, error) {
	if pinned := getenv(envGoBin); pinned != "" {
		info, statErr := os.Stat(pinned)
		if statErr != nil || info.IsDir() {
			return "", fmt.Errorf("%w: %s=%q (SystemNix must pin the nix store path)", errPinnedGoBad, envGoBin, pinned)
		}

		return pinned, nil
	}

	found, err := lookPath("go")
	if err != nil {
		return "", fmt.Errorf(
			"%w: %s is unset and %q is not on PATH — the gate cannot judge this repo, so it passes open; fix by pinning %s in the service environment",
			errNoToolchain,
			envGoBin,
			"go",
			envGoBin,
		)
	}

	return found, nil
}

// runCheck executes one gate check and turns tool failures into verdicts.
// Output streams live so the daemon journal carries the compiler's evidence.
func runCheck(goBin string, check check, stdout, stderr io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	args := strings.Join(check.args, " ")

	cmd := exec.CommandContext(ctx, goBin, check.args...) // #nosec G204 -- goBin comes from the pinned env var or LookPath, args are static
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = gateEnv(os.Environ())

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start go %s: %w", args, err)
	}

	if err := cmd.Wait(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("%w: go %s exceeded %s — a gate that hangs must not hang the daemon", errCheckTimeout, args, checkTimeout)
		}

		return fmt.Errorf("go %s failed (exit): %w", args, err)
	}

	return nil
}

// gateEnv builds the child environment: inherit everything, but make sure
// GOCACHE exists so go does not die with "GOCACHE is not defined" in a
// stripped service environment.
func gateEnv(env []string) []string {
	if !envHas(env, "GOCACHE") {
		if cache, err := os.UserCacheDir(); err == nil {
			env = append(env, "GOCACHE="+filepath.Join(cache, "go-build"))
		}
	}

	return env
}

// envHas reports whether the environment carries the given key.
func envHas(env []string, key string) bool {
	prefix := key + "="

	for _, kv := range env {
		if strings.HasPrefix(kv, prefix) {
			return true
		}
	}

	return false
}
