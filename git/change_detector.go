// Package git provides git-based file change detection for incremental analysis.
//
// This package enables art-dupl to analyze only changed files since a git reference,
// significantly speeding up duplicate detection in large codebases during iterative
// development.
//
// Usage:
//
//	detector := git.NewChangeDetector()
//	changed, err := detector.GetChangedFiles("HEAD~1")
//	if err != nil {
//	    // Not a git repo or git unavailable - fall back to full scan
//	}
//	// Use changed files for incremental analysis
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ChangeDetector detects file changes using git.
type ChangeDetector struct {
	// workingDir is the directory to run git commands in
	workingDir string
}

// NewChangeDetector creates a new ChangeDetector.
// If workingDir is empty, uses current working directory.
func NewChangeDetector(workingDir ...string) *ChangeDetector {
	dir := ""
	if len(workingDir) > 0 && workingDir[0] != "" {
		dir = workingDir[0]
	}
	return &ChangeDetector{workingDir: dir}
}

// ChangeInfo contains information about a changed file.
type ChangeInfo struct {
	// Path is the relative path to the file
	Path string
	// Status indicates the type of change (A=added, M=modified, D=deleted, R=renamed)
	Status string
}

// GetChangedFiles returns files changed since the given git reference.
//
// The reference can be:
//   - A commit hash (e.g., "abc123")
//   - A branch name (e.g., "main", "feature/foo")
//   - A tag (e.g., "v1.0.0")
//   - A relative reference (e.g., "HEAD~1", "HEAD~5")
//   - Empty string (defaults to "HEAD" - all uncommitted changes)
//
// Returns ErrNotGitRepo if not in a git repository.
// Returns ErrGitNotAvailable if git command fails.
func (d *ChangeDetector) GetChangedFiles(since string) ([]ChangeInfo, error) {
	if !d.isGitRepo() {
		return nil, ErrNotGitRepo
	}

	if since == "" {
		since = "HEAD"
	}

	// Get changed files using git diff
	// --name-status shows status (A/M/D/R)
	// --diff-filter=ACMR excludes deleted files (we can't analyze deleted files)
	cmd := exec.Command("git", "diff", "--name-status", "--diff-filter=ACMR", since)
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		// Check if it's an exit error (invalid reference)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("%w: %s", ErrGitCommand, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	return d.parseDiffOutput(output), nil
}

// GetChangedGoFiles returns only Go files changed since the given reference.
// This is a convenience method that filters the results of GetChangedFiles.
func (d *ChangeDetector) GetChangedGoFiles(since string) ([]ChangeInfo, error) {
	changes, err := d.GetChangedFiles(since)
	if err != nil {
		return nil, err
	}

	var goFiles []ChangeInfo
	for _, change := range changes {
		if strings.HasSuffix(change.Path, ".go") {
			goFiles = append(goFiles, change)
		}
	}
	return goFiles, nil
}

// GetStagedFiles returns files that are staged for commit.
func (d *ChangeDetector) GetStagedFiles() ([]ChangeInfo, error) {
	if !d.isGitRepo() {
		return nil, ErrNotGitRepo
	}

	cmd := exec.Command("git", "diff", "--name-status", "--cached", "--diff-filter=ACMR")
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	return d.parseDiffOutput(output), nil
}

// GetUnstagedFiles returns modified files that are not staged.
func (d *ChangeDetector) GetUnstagedFiles() ([]ChangeInfo, error) {
	if !d.isGitRepo() {
		return nil, ErrNotGitRepo
	}

	cmd := exec.Command("git", "diff", "--name-status", "--diff-filter=ACMR")
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	return d.parseDiffOutput(output), nil
}

// GetAllChanges returns all changes (staged + unstaged + untracked).
func (d *ChangeDetector) GetAllChanges() ([]ChangeInfo, error) {
	if !d.isGitRepo() {
		return nil, ErrNotGitRepo
	}

	var allChanges []ChangeInfo

	// Get staged changes
	staged, err := d.GetStagedFiles()
	if err != nil {
		return nil, err
	}
	allChanges = append(allChanges, staged...)

	// Get unstaged changes
	unstaged, err := d.GetUnstagedFiles()
	if err != nil {
		return nil, err
	}
	allChanges = append(allChanges, unstaged...)

	// Get untracked files
	untracked, err := d.GetUntrackedFiles()
	if err != nil {
		return nil, err
	}
	allChanges = append(allChanges, untracked...)

	return deduplicateChanges(allChanges), nil
}

// GetUntrackedFiles returns untracked (new) files.
func (d *ChangeDetector) GetUntrackedFiles() ([]ChangeInfo, error) {
	if !d.isGitRepo() {
		return nil, ErrNotGitRepo
	}

	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	var changes []ChangeInfo
	lines := bytes.Split(bytes.TrimSpace(output), []byte("\n"))
	for _, line := range lines {
		if len(line) > 0 {
			changes = append(changes, ChangeInfo{
				Path:   string(line),
				Status: "A", // Untracked files are effectively "added"
			})
		}
	}
	return changes, nil
}

// GetMergeBase returns the merge base between current branch and main/master.
// This is useful for finding all changes in a feature branch.
func (d *ChangeDetector) GetMergeBase(mainBranch string) (string, error) {
	if !d.isGitRepo() {
		return "", ErrNotGitRepo
	}

	if mainBranch == "" {
		mainBranch = "main"
	}

	cmd := exec.Command("git", "merge-base", "HEAD", mainBranch)
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		// Try master if main fails
		if mainBranch == "main" {
			return d.GetMergeBase("master")
		}
		return "", fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name.
func (d *ChangeDetector) GetCurrentBranch() (string, error) {
	if !d.isGitRepo() {
		return "", ErrNotGitRepo
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// isGitRepo checks if the working directory is inside a git repository.
func (d *ChangeDetector) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "true"
}

// parseDiffOutput parses git diff --name-status output.
func (d *ChangeDetector) parseDiffOutput(output []byte) []ChangeInfo {
	var changes []ChangeInfo
	lines := bytes.Split(bytes.TrimSpace(output), []byte("\n"))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		parts := bytes.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := string(parts[0])
		path := string(parts[1])

		// Handle rename (R100 old_path new_path)
		if len(parts) >= 3 && strings.HasPrefix(status, "R") {
			path = string(parts[2])
		}

		changes = append(changes, ChangeInfo{
			Path:   path,
			Status: status,
		})
	}

	return changes
}

// deduplicateChanges removes duplicate entries (same path).
func deduplicateChanges(changes []ChangeInfo) []ChangeInfo {
	seen := make(map[string]bool)
	var result []ChangeInfo

	for _, change := range changes {
		if !seen[change.Path] {
			seen[change.Path] = true
			result = append(result, change)
		}
	}

	return result
}

// IsGitRepo checks if a directory is inside a git repository.
// This is a convenience function that doesn't require creating a ChangeDetector.
func IsGitRepo(dir string) bool {
	return NewChangeDetector(dir).isGitRepo()
}

// FindGitRoot finds the root of the git repository containing the given path.
// Returns empty string if not in a git repository.
func FindGitRoot(startPath string) string {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return ""
	}

	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current || parent == "" {
			return ""
		}
		current = parent
	}
}

// Errors returned by the git package.
var (
	// ErrNotGitRepo indicates the directory is not inside a git repository.
	ErrNotGitRepo = &GitError{Op: "check", Err: "not a git repository"}

	// ErrGitNotAvailable indicates git command is not available.
	ErrGitNotAvailable = &GitError{Op: "exec", Err: "git command not available"}

	// ErrGitCommand indicates a git command failed.
	ErrGitCommand = &GitError{Op: "command", Err: "git command failed"}
)

// GitError represents an error from git operations.
type GitError struct {
	Op  string
	Err string
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git error: %s: %s", e.Op, e.Err)
}

func (e *GitError) Is(target error) bool {
	t, ok := target.(*GitError)
	if !ok {
		return false
	}
	return e.Op == t.Op || e.Err == t.Err
}
