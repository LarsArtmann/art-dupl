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
	"context"
	"errors"
	"fmt"
	"os/exec"
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
	cmd := exec.CommandContext(context.Background(), "git", "diff", "--name-status", "--diff-filter=ACMR", since)
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

	cmd := exec.CommandContext(context.Background(), "git", "diff", "--name-status", "--cached", "--diff-filter=ACMR")
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

	cmd := exec.CommandContext(context.Background(), "git", "diff", "--name-status", "--diff-filter=ACMR")
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

	cmd := exec.CommandContext(context.Background(), "git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	var changes []ChangeInfo
	lines := bytes.SplitSeq(bytes.TrimSpace(output), []byte("\n"))
	for line := range lines {
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

	cmd := exec.CommandContext(context.Background(), "git", "merge-base", "HEAD", mainBranch)
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

	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGitCommand, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// isGitRepo checks if the working directory is inside a git repository.
func (d *ChangeDetector) isGitRepo() bool {
	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--is-inside-work-tree")
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
	lines := bytes.SplitSeq(bytes.TrimSpace(output), []byte("\n"))

	for line := range lines {
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
