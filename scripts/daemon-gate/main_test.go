package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func getenvFunc(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()

	for name, content := range files {
		path := filepath.Join(dir, name)

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestGatePassesNonGoDirectory(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer

	code := run(&stdout, &stderr, getenvFunc(nil), exec.LookPath)

	if code != exitPass {
		t.Fatalf("exit = %d, want %d (stderr: %s)", code, exitPass, stderr.String())
	}
}

func TestGateKillSwitchBeatsEverything(t *testing.T) {
	var stdout, stderr bytes.Buffer

	lookPath := func(string) (string, error) { return "", os.ErrNotExist }
	code := run(&stdout, &stderr, getenvFunc(map[string]string{envDisable: "1"}), lookPath)

	if code != exitPass {
		t.Fatalf("exit = %d, want %d", code, exitPass)
	}

	if !strings.Contains(stderr.String(), "disabled") {
		t.Fatalf("stderr must name the kill switch, got: %s", stderr.String())
	}
}

func TestGateRefusesBrokenGoModule(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain on PATH")
	}

	t.Chdir(writeTree(t, map[string]string{
		"go.mod":  "module broken\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() { undefinedSymbol() }\n",
	}))

	var stdout, stderr bytes.Buffer

	code := run(&stdout, &stderr, getenvFunc(nil), exec.LookPath)

	if code != exitRed {
		t.Fatalf("exit = %d, want %d (stderr: %s)", code, exitRed, stderr.String())
	}

	if !strings.Contains(stderr.String(), "REFUSED") {
		t.Fatalf("stderr must say REFUSED, got: %s", stderr.String())
	}

	if !strings.Contains(stderr.String(), "undefinedSymbol") {
		t.Fatalf("stderr must carry the compiler evidence, got: %s", stderr.String())
	}
}

func TestGatePassesGreenGoModule(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain on PATH")
	}

	t.Chdir(writeTree(t, map[string]string{
		"go.mod":  "module green\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	}))

	var stdout, stderr bytes.Buffer

	code := run(&stdout, &stderr, getenvFunc(nil), exec.LookPath)

	if code != exitPass {
		t.Fatalf("exit = %d, want %d (stderr: %s)", code, exitPass, stderr.String())
	}
}

func TestPlanChecksJSONGateOnlyWhenDirExists(t *testing.T) {
	plain := t.TempDir()

	if got := len(planChecks(plain)); got != 1 {
		t.Fatalf("plain module checks = %d, want 1 (build only)", got)
	}

	withGate := writeTree(t, map[string]string{
		"internal/jsonv2gate/keeper.go": "package jsonv2gate\n",
	})

	checks := planChecks(withGate)

	if len(checks) != 2 {
		t.Fatalf("art-dupl-shaped module checks = %d, want 2 (build + jsonv2gate)", len(checks))
	}

	if checks[1].name != "jsonv2gate" || !strings.Contains(strings.Join(checks[1].args, " "), "-count=1") {
		t.Fatalf("jsonv2gate check misconfigured: %+v", checks[1])
	}
}

func TestResolveGoPinnedBeatsPath(t *testing.T) {
	pinned := filepath.Join(t.TempDir(), "go")

	if err := os.WriteFile(pinned, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	lookPath := func(string) (string, error) { return "/should/not/be/used", nil }

	got, err := resolveGo(getenvFunc(map[string]string{envGoBin: pinned}), lookPath)
	if err != nil || got != pinned {
		t.Fatalf("resolveGo = %q, %v; want pinned %q to win", got, err, pinned)
	}

	missing := filepath.Join(t.TempDir(), "missing")

	_, err = resolveGo(getenvFunc(map[string]string{envGoBin: missing}), lookPath)
	if err == nil {
		t.Fatal("resolveGo must reject a pinned path that does not exist")
	}

	if !strings.Contains(err.Error(), envGoBin) {
		t.Fatalf("error must name %s for an actionable message, got: %v", envGoBin, err)
	}
}

func TestResolveGoMissingToolchainIsRecoverable(t *testing.T) {
	lookPath := func(string) (string, error) { return "", exec.ErrNotFound }

	_, err := resolveGo(getenvFunc(nil), lookPath)
	if err == nil {
		t.Fatal("expected an error when no toolchain exists")
	}

	if !errors.Is(err, errNoToolchain) {
		t.Fatalf("error must wrap the static errNoToolchain sentinel, got: %v", err)
	}
}

func TestGateEnvAddsGOCACHEFallback(t *testing.T) {
	cacheBase := filepath.Join(t.TempDir(), "cache")

	t.Setenv("XDG_CACHE_HOME", cacheBase)

	env := gateEnv([]string{"PATH=/usr/bin"})

	found := false

	for _, kv := range env {
		if strings.HasPrefix(kv, "GOCACHE=") {
			found = true
		}
	}

	if !found {
		t.Fatal("gateEnv must add a GOCACHE fallback when the environment lacks one")
	}
}

func TestGateEnvKeepsExistingGOCACHE(t *testing.T) {
	kept := "GOCACHE=/existing/cache"

	env := gateEnv([]string{kept})

	count := 0

	for _, kv := range env {
		if strings.HasPrefix(kv, "GOCACHE=") {
			count++
		}
	}

	if count != 1 || env[0] != kept {
		t.Fatalf("gateEnv must not duplicate or override an existing GOCACHE, got %v", env)
	}
}
