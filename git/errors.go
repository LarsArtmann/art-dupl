// Package git provides git-based file change detection for incremental analysis.
package git

import "fmt"

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
