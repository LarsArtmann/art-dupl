package cmd

import (
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

// TestBaselineRecordHonorsAcceptDirectives verifies that the baseline record
// subcommand honors //art-dupl:accept directives — another site that had a
// truncated SuppressionConfig before the fix.
//
// NOTE: This test is intentionally NOT t.Parallel(): executeTestCommand ->
// CaptureStdoutStderr mutates the process-global os.Stdout/os.Stderr, which
// races with parallel sibling tests that read those globals directly.
func TestBaselineRecordHonorsAcceptDirectives(t *testing.T) {
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

	// With accept directives: baseline should record 0 groups.
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

	if !strings.Contains(string(output), "Recorded 0 clone group") {
		t.Errorf("expected 'Recorded 0 clone group' with accept directives, got:\n%s", output)
	}
}
