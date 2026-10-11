package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func getenvFunc(t *testing.T, values map[string]string) func(string) string {
	t.Helper()
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

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestGatePassesNonGoDirectory(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, getenvFunc(t, nil), exec.LookPath)

	if code != exitPass {
		t.Fatalf("exit = %d, want %d (stderr: %s)", code, exitPass, stderr.String())
	}
}

func TestGateKillSwitchBeatsEverything(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, getenvFunc(t, map[string]string{envDisable: "1"}), func(string) (string, error) {
		return "", os.ErrNotExist
	})

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
	dir := writeTree(t, map[string]string{
		"go.mod":  "module broken\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() { undefinedSymbol() }\n",
	})
	chdir(t, dir)

	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, getenvFunc(t, nil), exec.LookPath)

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
	dir := writeTree(t, map[string]string{
		"go.mod":  "module green\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})
	chdir(t, dir)

	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, getenvFunc(t, nil), exec.LookPath)

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

	got, err := resolveGo(getenvFunc(t, map[string]string{envGoBin: pinned}), func(string) (string, error) {
		return "/should/not/be/used", nil
	})
	if err != nil || got != pinned {
		t.Fatalf("resolveGo = %q, %v; want pinned %q to win", got, err, pinned)
	}

	_, err = resolveGo(getenvFunc(t, map[string]string{envGoBin: filepath.Join(t.TempDir(), "missing")}), func(string) (string, error) {
		return "/unused", nil
	})
	if err == nil {
		t.Fatal("resolveGo must reject a pinned path that does not exist")
	}
	if !strings.Contains(err.Error(), envGoBin) {
		t.Fatalf("error must name %s for an actionable message, got: %v", envGoBin, err)
	}
}

func TestResolveGoMissingToolchainIsRecoverable(t *testing.T) {
	_, err := resolveGo(getenvFunc(t, nil), func(string) (string, error) {
		return "", exec.ErrNotFound
	})
	if err == nil {
		t.Fatal("expected an error when no toolchain exists")
	}
	if !strings.Contains(err.Error(), envGoBin) {
		t.Fatalf("error must tell the operator how to fix it (pin %s), got: %v", envGoBin, err)
	}
}

func TestGateEnvAddsGOCACHEFallback(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	os.Unsetenv("GOCACHE")
	env := gateEnv()
	found := false
	for _, kv := range env {
		if strings.HasPrefix(kv, "GOCACHE=") {
			found = true
		}
	}
	if !found {
		t.Fatal("gateEnv must add a GOCACHE fallback when the environment lacks one")
	}

	t.Setenv("GOCACHE", filepath.Join(t.TempDir(), "kept"))
	for _, kv := range gateEnv() {
		if strings.HasPrefix(kv, "GOCACHE=") && !strings.HasSuffix(kv, "kept") {
			t.Fatalf("gateEnv must not override an existing GOCACHE, got %q", kv)
		}
	}
	_ = runtime.GOOS
}
