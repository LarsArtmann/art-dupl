package cmd

import (
	"io"

	"github.com/LarsArtmann/art-dupl/internal/gitignore"
)

// GitignoreMatcher matches file paths against .gitignore patterns collected
// from the directory tree. The implementation lives in internal/gitignore so
// non-cmd consumers (pkg/provider) share the same matcher; this alias keeps
// the cmd-layer call sites and tests stable.
type GitignoreMatcher = gitignore.GitignoreMatcher

// LoadGitignore walks the analyzed paths (and their ancestors) collecting
// .gitignore rules. It delegates to internal/gitignore.LoadGitignore.
func LoadGitignore(paths []string, stderr io.Writer) *GitignoreMatcher {
	return gitignore.LoadGitignore(paths, stderr)
}
