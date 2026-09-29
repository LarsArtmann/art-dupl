package cmd

import (
	"bytes"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const acceptFixtureDir = "./testdata/accept_fixture"

// TestStatsHonorsAcceptDirectives is the primary regression test for the bug
// where art-dupl stats ignored //art-dupl:accept directives because
// SuppressionConfig was constructed with a truncated struct literal missing
// the AcceptDirectives field.
//
// The fixture (testdata/accept_fixture/) contains two files with identical
// function bodies; a.go carries a //art-dupl:accept directive on the
// duplicated block. --no-actionability is used to isolate accept-directive
// behavior from the actionability engine (the small 2-file clone is otherwise
// classified as non-actionable, which would mask the directive check).
//
// NOTE: This test is intentionally NOT t.Parallel(): executeTestCommand ->
// CaptureStdoutStderr mutates the process-global os.Stdout/os.Stderr, which
// races with parallel sibling tests that read those globals directly (e.g.
// fmt.Printf in PrintVersion, fmt.Fprintln(os.Stderr) in printBuildingStatus).
func TestStatsHonorsAcceptDirectives(t *testing.T) {
	// Without accept directives: the clone group must appear.
	output, err := executeTestCommand(t, []string{
		binaryName, statsSubCommand,
		"-t", "1",
		"--no-actionability",
		"--no-accept-directives",
		acceptFixtureDir,
	})
	if err != nil {
		t.Fatalf("stats --no-accept-directives failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Clone Groups: 1") {
		t.Errorf("expected 'Clone Groups: 1' without directives, got:\n%s", output)
	}

	// With accept directives: the clone group must be suppressed.
	output, err = executeTestCommand(t, []string{
		binaryName, statsSubCommand,
		"-t", "1",
		"--no-actionability",
		acceptFixtureDir,
	})
	if err != nil {
		t.Fatalf("stats with accept directives failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Clone Groups: 0") {
		t.Errorf("expected 'Clone Groups: 0' with accept directives, got:\n%s", output)
	}
}

// TestBaselineRecordBypassesAcceptDirectives verifies that the baseline record
// subcommand captures ALL groups regardless of //art-dupl:accept directives.
// This prevents the split-brain issue where directive-suppressed groups are
// absent from the baseline, then appear as "new" when directives are later
// removed. The baseline is the complete snapshot; directives are checked
// separately at analysis time.
//
// NOTE: This test is intentionally NOT t.Parallel(): executeTestCommand ->
// CaptureStdoutStderr mutates the process-global os.Stdout/os.Stderr, which
// races with parallel sibling tests that read those globals directly.
func TestBaselineRecordBypassesAcceptDirectives(t *testing.T) {
	// Without accept directives: baseline should record the clone group.
	baselineNoAccept := filepath.Join(t.TempDir(), "baseline-no-accept.json")

	output, err := executeTestCommand(t, []string{
		binaryName, "baseline",
		"--baseline-path", baselineNoAccept,
		"-t", "1",
		"--no-actionability",
		"--no-accept-directives",
		acceptFixtureDir,
	})
	if err != nil {
		t.Fatalf("baseline --no-accept-directives failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Recorded 1 clone group") {
		t.Errorf("expected 'Recorded 1 clone group' without directives, got:\n%s", output)
	}

	// With accept directives: baseline should STILL record 1 group (bypassed).
	baselineWithAccept := filepath.Join(t.TempDir(), "baseline-with-accept.json")

	output, err = executeTestCommand(t, []string{
		binaryName, "baseline",
		"--baseline-path", baselineWithAccept,
		"-t", "1",
		"--no-actionability",
		acceptFixtureDir,
	})
	if err != nil {
		t.Fatalf("baseline with accept directives failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Recorded 1 clone group") {
		t.Errorf("expected 'Recorded 1 clone group' with directives (bypassed), got:\n%s", output)
	}
}

// TestDeadDirectiveWarningEndToEnd verifies the stale-directive detector over
// the real CLI pipeline: a //art-dupl:accept <hash> directive whose hash does
// not match any current group must produce a stderr warning, and rewriting it
// with the live group hash must silence the warning AND suppress the group.
func TestDeadDirectiveWarningEndToEnd(t *testing.T) {
	// executeTestCommand mutates process-global os.Stdout/os.Stderr — serial only.
	dir := t.TempDir()

	dupCode := "package dup\n\n" +
		"import \"fmt\"\n\n" +
		"func fa(x int) int {\n" +
		"\t// art-dupl:accept deadbeefdeadbeef\n" +
		"\ta := x + 1\n" +
		"\tb := a * 2\n" +
		"\tc := b - 3\n" +
		"\td := c + 4\n" +
		"\te := d * 5\n" +
		"\tf := e - 6\n" +
		"\treturn f\n" +
		"}\n"

	for _, name := range []string{"one.go", "two.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(dupCode), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	output, err := executeTestCommand(t, []string{binaryName, "-t", "1", "--no-actionability", dir})
	if err != nil {
		t.Fatalf("run with stale directive failed: %v\nOutput: %s", err, output)
	}

	staleWarning := "warning: stale //art-dupl:accept deadbeefdeadbeef"
	if !strings.Contains(string(output), staleWarning) {
		t.Errorf("expected stale-directive warning on stderr, got:\n%s", output)
	}

	if !strings.Contains(string(output), "one.go") {
		t.Errorf("stale directive must NOT suppress the group (clones should show), got:\n%s", output)
	}

	// Recover the live group hash from the JSON channel and rewrite the
	// directive with it.
	jsonOutput, err := executeTestCommand(t, []string{binaryName, "-t", "1", "--no-actionability", "--json", dir})
	if err != nil {
		t.Fatalf("json run failed: %v\nOutput: %s", err, jsonOutput)
	}

	var parsed struct {
		CloneGroups []struct {
			Hash string `json:"hash"`
		} `json:"clone_groups"`
	}
	// The combined output carries the stale-directive warning (stderr) after
	// the JSON document; Decode stops after the first top-level value.
	if err := json.UnmarshalRead(bytes.NewReader(jsonOutput), &parsed); err != nil {
		t.Fatalf("decode json output: %v\nOutput: %s", err, jsonOutput)
	}

	if len(parsed.CloneGroups) == 0 {
		t.Fatalf("expected 1 clone group in json output, got:\n%s", jsonOutput)
	}

	liveCode := strings.Replace(dupCode, "deadbeefdeadbeef", parsed.CloneGroups[0].Hash, 1)
	for _, name := range []string{"one.go", "two.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(liveCode), 0o600); err != nil {
			t.Fatalf("rewrite directive in %s: %v", name, err)
		}
	}

	output, err = executeTestCommand(t, []string{binaryName, "-t", "1", "--no-actionability", dir})
	if err != nil {
		t.Fatalf("run with live directive failed: %v\nOutput: %s", err, output)
	}

	if strings.Contains(string(output), "warning: stale") {
		t.Errorf("live directive must not warn, got:\n%s", output)
	}

	if strings.Contains(string(output), "one.go") {
		t.Errorf("live directive must suppress the group, got:\n%s", output)
	}
}
