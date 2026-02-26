package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupGitRepo creates a temporary git repository for testing.
// Returns the temp directory path and a cleanup function.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// Initialize git repo with explicit branch name
	cmd := exec.Command("git", "init", "-b", "main")

	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init git repo: %v, output: %s", err, string(output))
	}

	// Configure git user (required for commits) - use --local to ensure it's set for this repo
	cmd = exec.Command("git", "config", "--local", "user.email", "test@test.com")

	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git email: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "--local", "user.name", "Test")

	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git name: %v, output: %s", err, string(output))
	}

	// Disable GPG signing for test repos (user may have it enabled globally)
	cmd = exec.Command("git", "config", "--local", "commit.gpgsign", "false")

	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to disable gpgsign: %v, output: %s", err, string(output))
	}

	return tempDir
}

// createAndCommitFile creates a file and commits it.
func createAndCommitFile(t *testing.T, repoDir, filename, content string) {
	t.Helper()

	writeFile(t, repoDir, filename, content)

	cmd := exec.Command("git", "add", filename)

	cmd.Dir = repoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to git add: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "commit", "-m", "Add "+filename)

	cmd.Dir = repoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to git commit: %v, output: %s", err, string(output))
	}
}

// writeFile writes content to a file for testing.
func writeFile(t *testing.T, repoDir, filename, content string) {
	t.Helper()
	if err := os.WriteFile(
		filepath.Join(repoDir, filename),
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatalf("Failed to write %s: %v", filename, err)
	}
}

// assertNoChanges asserts that GetChangedFiles returns 0 changes for the given since value.
func assertNoChanges(t *testing.T, repoDir, since string) {
	t.Helper()

	detector := NewChangeDetector(repoDir)

	changes, err := detector.GetChangedFiles(since)
	if err != nil {
		t.Fatalf("GetChangedFiles failed: %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("Expected 0 changes, got %d", len(changes))
	}
}

// TestNewChangeDetector tests ChangeDetector creation.
func TestNewChangeDetector(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		detector := NewChangeDetector()
		if detector == nil {
			t.Fatal("Expected non-nil detector")
		}

		if detector.workingDir != "" {
			t.Errorf("Expected empty workingDir, got %q", detector.workingDir)
		}
	})

	t.Run("empty argument", func(t *testing.T) {
		detector := NewChangeDetector("")
		if detector == nil {
			t.Fatal("Expected non-nil detector")
		}

		if detector.workingDir != "" {
			t.Errorf("Expected empty workingDir, got %q", detector.workingDir)
		}
	})

	t.Run("custom directory", func(t *testing.T) {
		customDir := "/some/path"

		detector := NewChangeDetector(customDir)
		if detector == nil {
			t.Fatal("Expected non-nil detector")
		}

		if detector.workingDir != customDir {
			t.Errorf("Expected workingDir=%q, got %q", customDir, detector.workingDir)
		}
	})
}

// TestIsGitRepo tests git repository detection.
func TestChangeDetector_isGitRepo(t *testing.T) {
	t.Run("valid git repo", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		detector := NewChangeDetector(repoDir)

		if !detector.isGitRepo() {
			t.Error("Expected isGitRepo to return true for git repo")
		}
	})

	t.Run("non-git directory", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		if detector.isGitRepo() {
			t.Error("Expected isGitRepo to return false for non-git directory")
		}
	})
}

// testNoChanges is a helper function that tests that no changes are detected
func testNoChanges(t *testing.T, since string) {
	repoDir := setupGitRepo(t)
	createAndCommitFile(t, repoDir, "initial.go", "package main")
	assertNoChanges(t, repoDir, since)
}

// TestGetChangedFiles tests file change detection.
func TestChangeDetector_GetChangedFiles(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetChangedFiles("HEAD")
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})

	t.Run("no changes", func(t *testing.T) {
		testNoChanges(t, "HEAD")
	})

	t.Run("with modified files", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		createAndCommitFile(t, repoDir, "file.go", "package main\n")

		// Modify the file
		writeFile(t, repoDir, "file.go", "package main\n\nfunc main() {}\n")

		detector := NewChangeDetector(repoDir)

		changes, err := detector.GetChangedFiles("HEAD")
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}

		if len(changes) != 1 {
			t.Errorf("Expected 1 change, got %d", len(changes))
		}

		if changes[0].Status != "M" {
			t.Errorf("Expected status M, got %q", changes[0].Status)
		}
	})

	t.Run("empty since defaults to HEAD", func(t *testing.T) {
		testNoChanges(t, "")
	})
}

// TestGetChangedGoFiles tests Go file filtering.
func TestChangeDetector_GetChangedGoFiles(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetChangedGoFiles("HEAD")
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})

	t.Run("filters only go files", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		createAndCommitFile(t, repoDir, "file.go", "package main\n")
		createAndCommitFile(t, repoDir, "file.txt", "text content\n")

		// Modify both files
		writeFile(t, repoDir, "file.go", "package main\n\nfunc main() {}\n")

		writeFile(t, repoDir, "file.txt", "modified text\n")

		detector := NewChangeDetector(repoDir)

		changes, err := detector.GetChangedGoFiles("HEAD")
		if err != nil {
			t.Fatalf("GetChangedGoFiles failed: %v", err)
		}

		if len(changes) != 1 {
			t.Errorf("Expected 1 Go file change, got %d", len(changes))
		}

		if changes[0].Path != "file.go" {
			t.Errorf("Expected file.go, got %q", changes[0].Path)
		}
	})
}

// TestGetStagedFiles tests staged file detection.
func TestChangeDetector_GetStagedFiles(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetStagedFiles()
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})

	t.Run("with staged files", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		createAndCommitFile(t, repoDir, "initial.go", "package main")

		// Create and stage a new file
		writeFile(t, repoDir, "new.go", "package main\n")

		cmd := exec.Command("git", "add", "new.go")

		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to stage file: %v", err)
		}

		detector := NewChangeDetector(repoDir)

		changes, err := detector.GetStagedFiles()
		if err != nil {
			t.Fatalf("GetStagedFiles failed: %v", err)
		}

		if len(changes) != 1 {
			t.Errorf("Expected 1 staged file, got %d", len(changes))
		}
	})
}

// TestGetUnstagedFiles tests unstaged file detection.
func TestChangeDetector_GetUnstagedFiles(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetUnstagedFiles()
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})

	t.Run("with unstaged modifications", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		createAndCommitFile(t, repoDir, "file.go", "package main\n")

		// Modify without staging
		writeFile(t, repoDir, "file.go", "package main\n\nfunc main() {}\n")

		detector := NewChangeDetector(repoDir)

		changes, err := detector.GetUnstagedFiles()
		if err != nil {
			t.Fatalf("GetUnstagedFiles failed: %v", err)
		}

		if len(changes) != 1 {
			t.Errorf("Expected 1 unstaged file, got %d", len(changes))
		}
	})
}

// TestGetUntrackedFiles tests untracked file detection.
func TestChangeDetector_GetUntrackedFiles(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetUntrackedFiles()
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})

	t.Run("with untracked files", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		createAndCommitFile(t, repoDir, "initial.go", "package main")

		// Create untracked file
		writeFile(t, repoDir, "untracked.go", "package main\n")

		detector := NewChangeDetector(repoDir)

		changes, err := detector.GetUntrackedFiles()
		if err != nil {
			t.Fatalf("GetUntrackedFiles failed: %v", err)
		}

		if len(changes) != 1 {
			t.Errorf("Expected 1 untracked file, got %d", len(changes))
		}

		if changes[0].Status != "A" {
			t.Errorf("Expected status A for untracked file, got %q", changes[0].Status)
		}
	})
}

// TestGetAllChanges tests combined change detection.
func TestChangeDetector_GetAllChanges(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetAllChanges()
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})

	t.Run("combines all changes", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		createAndCommitFile(t, repoDir, "initial.go", "package main")

		// Staged file
		writeFile(t, repoDir, "staged.go", "package main\n")

		cmd := exec.Command("git", "add", "staged.go")
		cmd.Dir = repoDir
		_ = cmd.Run()

		// Untracked file
		writeFile(t, repoDir, "untracked.go", "package main\n")

		detector := NewChangeDetector(repoDir)

		changes, err := detector.GetAllChanges()
		if err != nil {
			t.Fatalf("GetAllChanges failed: %v", err)
		}
		// Should have at least the staged and untracked files
		if len(changes) < 2 {
			t.Errorf("Expected at least 2 changes, got %d", len(changes))
		}
	})
}

// TestGetCurrentBranch tests branch name detection.
func TestChangeDetector_GetCurrentBranch(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetCurrentBranch()
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})

	t.Run("in git repo with commits", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		createAndCommitFile(t, repoDir, "initial.go", "package main")

		detector := NewChangeDetector(repoDir)

		branch, err := detector.GetCurrentBranch()
		if err != nil {
			t.Fatalf("GetCurrentBranch failed: %v", err)
		}
		// Should return "main" (we init with -b main)
		if branch != "main" {
			t.Errorf("Expected branch 'main', got %q", branch)
		}
	})
}

// TestGetMergeBase tests merge base detection.
func TestChangeDetector_GetMergeBase(t *testing.T) {
	t.Run("not a git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		detector := NewChangeDetector(tempDir)

		_, err := detector.GetMergeBase("main")
		if !errors.Is(err, ErrNotGitRepo) {
			t.Errorf("Expected ErrNotGitRepo, got %v", err)
		}
	})
}

// TestDeduplicateChanges tests duplicate removal.
func TestDeduplicateChanges(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		result := deduplicateChanges(nil)
		if len(result) != 0 {
			t.Errorf("Expected 0 results, got %d", len(result))
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		changes := []ChangeInfo{
			{Path: "file1.go", Status: "M"},
			{Path: "file2.go", Status: "A"},
		}

		result := deduplicateChanges(changes)
		if len(result) != 2 {
			t.Errorf("Expected 2 results, got %d", len(result))
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		changes := []ChangeInfo{
			{Path: "file1.go", Status: "M"},
			{Path: "file1.go", Status: "A"}, // Duplicate
			{Path: "file2.go", Status: "A"},
		}

		result := deduplicateChanges(changes)
		if len(result) != 2 {
			t.Errorf("Expected 2 results (deduplicated), got %d", len(result))
		}
	})

	t.Run("preserves first occurrence", func(t *testing.T) {
		changes := []ChangeInfo{
			{Path: "file.go", Status: "M"},
			{Path: "file.go", Status: "A"},
		}

		result := deduplicateChanges(changes)
		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		if result[0].Status != "M" {
			t.Errorf("Expected status M from first occurrence, got %q", result[0].Status)
		}
	})
}

// TestParseDiffOutput tests git diff output parsing.
func TestChangeDetector_parseDiffOutput(t *testing.T) {
	detector := &ChangeDetector{}

	t.Run("empty output", func(t *testing.T) {
		result := detector.parseDiffOutput([]byte{})
		if len(result) != 0 {
			t.Errorf("Expected 0 results, got %d", len(result))
		}
	})

	t.Run("single modified file", func(t *testing.T) {
		output := []byte("M\tfile.go\n")

		result := detector.parseDiffOutput(output)
		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		if result[0].Status != "M" {
			t.Errorf("Expected status M, got %q", result[0].Status)
		}

		if result[0].Path != "file.go" {
			t.Errorf("Expected path file.go, got %q", result[0].Path)
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		output := []byte("M\tfile1.go\nA\tfile2.go\nD\tfile3.go\n")

		result := detector.parseDiffOutput(output)
		if len(result) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(result))
		}

		if result[0].Status != "M" || result[1].Status != "A" || result[2].Status != "D" {
			t.Error("Status parsing failed")
		}
	})

	t.Run("renamed file", func(t *testing.T) {
		output := []byte("R100\told.go\tnew.go\n")

		result := detector.parseDiffOutput(output)
		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}
		// Should use the new path (second path)
		if result[0].Path != "new.go" {
			t.Errorf("Expected path new.go, got %q", result[0].Path)
		}
	})

	t.Run("invalid lines ignored", func(t *testing.T) {
		output := []byte("M\tfile.go\ninvalid\nA\tfile2.go\n")

		result := detector.parseDiffOutput(output)
		if len(result) != 2 {
			t.Errorf("Expected 2 results (invalid line ignored), got %d", len(result))
		}
	})
}

// TestIsGitRepo tests the package-level IsGitRepo function.
func TestIsGitRepo(t *testing.T) {
	t.Run("git repo", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		if !IsGitRepo(repoDir) {
			t.Error("Expected IsGitRepo to return true for git repo")
		}
	})

	t.Run("non-git directory", func(t *testing.T) {
		tempDir := t.TempDir()
		if IsGitRepo(tempDir) {
			t.Error("Expected IsGitRepo to return false for non-git directory")
		}
	})
}

// TestFindGitRoot tests finding git repository root.
func TestFindGitRoot(t *testing.T) {
	t.Run("git repo root", func(t *testing.T) {
		repoDir := setupGitRepo(t)

		root := FindGitRoot(repoDir)
		if root != repoDir {
			t.Errorf("Expected root %q, got %q", repoDir, root)
		}
	})

	t.Run("subdirectory of git repo", func(t *testing.T) {
		repoDir := setupGitRepo(t)

		subDir := filepath.Join(repoDir, "subdir")
		err := os.Mkdir(subDir, 0o750)
		if err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}

		root := FindGitRoot(subDir)
		if root != repoDir {
			t.Errorf("Expected root %q, got %q", repoDir, root)
		}
	})

	t.Run("non-git directory", func(t *testing.T) {
		tempDir := t.TempDir()

		root := FindGitRoot(tempDir)
		if root != "" {
			t.Errorf("Expected empty string for non-git dir, got %q", root)
		}
	})
}

// TestGitError tests GitError type.
func TestGitError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &GitError{Op: "test", Err: "test error"}

		expected := "git error: test: test error"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error matching", func(t *testing.T) {
		err1 := &GitError{Op: "check", Err: "not a git repository"}
		err2 := &GitError{Op: "check", Err: "different message"}
		err3 := &GitError{Op: "different", Err: "not a git repository"}

		if !errors.Is(err1, err2) {
			t.Error("Expected err1 to match err2 (same Op)")
		}

		if !errors.Is(err1, err3) {
			t.Error("Expected err1 to match err3 (same Err)")
		}
	})

	t.Run("predefined errors", func(t *testing.T) {
		if !errors.Is(ErrNotGitRepo, &GitError{Op: "check"}) {
			t.Error("Expected ErrNotGitRepo to match GitError with Op=check")
		}

		if !errors.Is(ErrGitNotAvailable, &GitError{Op: "exec"}) {
			t.Error("Expected ErrGitNotAvailable to match GitError with Op=exec")
		}

		if !errors.Is(ErrGitCommand, &GitError{Op: "command"}) {
			t.Error("Expected ErrGitCommand to match GitError with Op=command")
		}
	})
}

// TestChangeInfo tests ChangeInfo struct.
func TestChangeInfo(t *testing.T) {
	change := ChangeInfo{
		Path:   "test.go",
		Status: "M",
	}

	if change.Path != "test.go" {
		t.Errorf("Expected Path test.go, got %q", change.Path)
	}

	if change.Status != "M" {
		t.Errorf("Expected Status M, got %q", change.Status)
	}
}
