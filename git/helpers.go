// Package git provides git-based file change detection for incremental analysis.
package git

import (
	"os"
	"path/filepath"
)

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
