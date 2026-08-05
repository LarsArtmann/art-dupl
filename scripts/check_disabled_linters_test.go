package scripts_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// scriptPath resolves the location of check-disabled-linters.sh relative to
// this test file, which lives in the same scripts/ directory.
func scriptPath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}

	p := filepath.Join(filepath.Dir(file), "check-disabled-linters.sh")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("script not found: %v", err)
	}

	return p
}

// runScript executes the guard against the given config file and returns the
// combined exit error (nil on exit 0), the resulting file contents, and stderr.
func runScript(t *testing.T, script, configFile string) (string, string, error) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), "bash", script, configFile)

	var out strings.Builder

	cmd.Stdout = &out
	cmd.Stderr = &out

	runErr := cmd.Run()

	b, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config after run: %v", err)
	}

	return string(b), out.String(), runErr
}

// writeConfig writes content to a config file in a temp dir and returns its path.
func writeConfig(t *testing.T, content string) string {
	t.Helper()

	cfg := filepath.Join(t.TempDir(), ".golangci.yml")
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	return cfg
}

// configWithBannedLinters is a realistic snippet with both banned linters
// enabled plus valid ones that must survive auto-fix.
const configWithBannedLinters = `linters:
  enable:
    - errcheck
    - exhaustruct
    - gosimple
    - tagliatelle
    - govet
`

// TestAutoFixRemovesBannedLinters is the core TODO item: when the config file
// is writable, the script must remove disabled linters via sed -i and exit 0.
func TestAutoFixRemovesBannedLinters(t *testing.T) {
	script := scriptPath(t)
	cfg := writeConfig(t, configWithBannedLinters)

	result, stderr, exitErr := runScript(t, script, cfg)
	if exitErr != nil {
		t.Fatalf("expected exit 0 after auto-fix, got %v (stderr: %s)", exitErr, stderr)
	}

	if strings.Contains(result, "exhaustruct") {
		t.Errorf("exhaustruct not removed after auto-fix:\n%s", result)
	}

	if strings.Contains(result, "tagliatelle") {
		t.Errorf("tagliatelle not removed after auto-fix:\n%s", result)
	}

	for _, want := range []string{"errcheck", "gosimple", "govet"} {
		if !strings.Contains(result, want) {
			t.Errorf("valid linter %q removed by auto-fix:\n%s", want, result)
		}
	}

	if !strings.Contains(stderr, "auto-removed") {
		t.Errorf("expected auto-removal warning on stderr, got: %s", stderr)
	}
}

// TestAutoFixRemovesSettingsBlock verifies that an orphaned "exhaustruct:"
// settings key is also stripped when the file is writable.
func TestAutoFixRemovesSettingsBlock(t *testing.T) {
	script := scriptPath(t)
	cfg := writeConfig(t, `linters:
  enable:
    - errcheck
  settings:
    exhaustruct:
      include:
        - .*
    gosimple:
      min-len: 0
`)

	result, _, exitErr := runScript(t, script, cfg)
	if exitErr != nil {
		t.Fatalf("expected exit 0 after auto-fix, got %v", exitErr)
	}

	if strings.Contains(result, "exhaustruct:") {
		t.Errorf("exhaustruct settings key not removed:\n%s", result)
	}
}

// TestReadOnlyFailsWithBannedLinter verifies the guard fails (exit 1) when a
// banned linter is present but the file cannot be auto-fixed.
func TestReadOnlyFailsWithBannedLinter(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permission bits; cannot test read-only guard")
	}

	script := scriptPath(t)
	cfg := filepath.Join(t.TempDir(), ".golangci.yml")

	if err := os.WriteFile(cfg, []byte("linters:\n  enable:\n    - errcheck\n    - tagliatelle\n"), 0o444); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(context.Background(), "bash", script, cfg)

	var out strings.Builder

	cmd.Stderr = &out

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for read-only file with banned linter")
	}

	if !strings.Contains(out.String(), "read-only") {
		t.Errorf("expected read-only failure message, got: %s", out.String())
	}

	after, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(after), "tagliatelle") {
		t.Error("read-only file was modified despite being unfixable")
	}
}

// TestCleanFilePasses verifies a config without banned linters passes cleanly.
func TestCleanFilePasses(t *testing.T) {
	script := scriptPath(t)
	original := "linters:\n  enable:\n    - errcheck\n    - gosimple\n    - govet\n"
	cfg := writeConfig(t, original)

	result, stderr, exitErr := runScript(t, script, cfg)
	if exitErr != nil {
		t.Fatalf("expected exit 0 for clean config, got %v (stderr: %s)", exitErr, stderr)
	}

	if result != original {
		t.Errorf("clean file was modified:\nbefore:\n%s\nafter:\n%s", original, result)
	}

	if !strings.Contains(stderr, "OK") {
		t.Errorf("expected OK message, got: %s", stderr)
	}
}

// TestCommentsMentioningBannedLintersAllowed ensures explanatory comments that
// mention a disabled linter name are NOT flagged or removed.
func TestCommentsMentioningBannedLintersAllowed(t *testing.T) {
	script := scriptPath(t)
	cfg := writeConfig(t, `# exhaustruct and tagliatelle are intentionally disabled.
# Do NOT re-enable exhaustruct (impractical for zero-value init).
linters:
  enable:
    - errcheck
`)

	result, stderr, exitErr := runScript(t, script, cfg)
	if exitErr != nil {
		t.Fatalf("expected exit 0 when only comments mention linters, got %v (stderr: %s)", exitErr, stderr)
	}

	for _, want := range []string{"exhaustruct", "tagliatelle"} {
		if !strings.Contains(result, want) {
			t.Errorf("comment mentioning %q was stripped:\n%s", want, result)
		}
	}
}

// TestMissingConfigFails verifies the script reports an error for a missing file.
func TestMissingConfigFails(t *testing.T) {
	script := scriptPath(t)
	cmd := exec.CommandContext(context.Background(), "bash", script, filepath.Join(t.TempDir(), "nope.yml"))

	if err := cmd.Run(); err == nil {
		t.Fatal("expected non-zero exit for missing config file")
	}
}
