package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxScannerBufferSize = 64 * 1024 // 64 KiB max line size for .gitignore parsing

const gitignoreScannerInitBufSize = 4096 // 4 KiB initial buffer for scanner

// GitignoreMatcher matches file paths against .gitignore patterns collected
// from the directory tree. It supports the most common gitignore syntax:
// simple names, globs (* and ?), directory-only (trailing /), anchored paths
// (leading /), and negation (!).
type GitignoreMatcher struct {
	// rules are ordered from least specific (root) to most specific (deepest).
	// Each rule knows its base directory so patterns are resolved correctly.
	rules []gitignoreRule
}

type gitignoreRule struct {
	baseDir  string // directory containing the .gitignore file
	patterns []gitignorePattern
}

type gitignorePattern struct {
	raw      string // original pattern text
	negated  bool   // starts with !
	dirOnly  bool   // ends with /
	anchored bool   // starts with /
}

// LoadGitignore walks up from startDir to find .gitignore files, then walks
// down through the analyzed paths to find nested ones. Returns nil if no
// .gitignore files are found.
func LoadGitignore(paths []string) *GitignoreMatcher {
	var rules []gitignoreRule

	seen := make(map[string]bool)

	// Collect .gitignore from each analyzed path and its ancestors
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}

		// Walk up the tree to find .gitignore files
		dir := abs
		for {
			gitignorePath := filepath.Join(dir, ".gitignore")
			if _, err := os.Stat(gitignorePath); err == nil {
				if !seen[dir] {
					seen[dir] = true

					if pats, err := parseGitignoreFile(gitignorePath); err != nil {
						fmt.Fprintf(os.Stderr, "warning: %v\n", err)
					} else if len(pats) > 0 {
						rules = append(rules, gitignoreRule{baseDir: dir, patterns: pats})
					}
				}
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}

			dir = parent
		}
	}

	if len(rules) == 0 {
		return nil
	}

	return &GitignoreMatcher{rules: rules}
}

// IsIgnored reports whether path should be ignored according to the loaded
// .gitignore rules. Rules are evaluated from root to deepest: later rules
// (more specific) can override earlier ones via negation.
func (m *GitignoreMatcher) IsIgnored(path string) bool {
	if m == nil {
		return false
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	ignored := false

	for _, rule := range m.rules {
		rel, err := filepath.Rel(rule.baseDir, abs)
		if err != nil {
			continue
		}

		// If the path is outside this rule's baseDir, skip
		if strings.HasPrefix(rel, "..") {
			continue
		}

		for _, pat := range rule.patterns {
			if matchGitignorePattern(pat, rel) {
				ignored = !pat.negated
			}
		}
	}

	return ignored
}

// matchGitignorePattern checks if relPath matches a single gitignore pattern.
func matchGitignorePattern(pat gitignorePattern, relPath string) bool {
	cleanPath := filepath.ToSlash(relPath)

	pattern := pat.raw

	if pat.dirOnly {
		// Directory-only patterns match any path that contains the directory
		// name as a component
		dirName := strings.TrimSuffix(pattern, "/")
		parts := strings.Split(cleanPath, "/")

		for _, part := range parts[:len(parts)-1] { // exclude the filename itself
			if part == dirName || matchGlob(dirName, part) {
				return true
			}
		}

		// Also match if the full path matches
		return matchGlob(pattern, cleanPath)
	}

	if pat.anchored {
		// Anchored patterns match from the root of the .gitignore directory
		return matchGlob(pattern, cleanPath)
	}

	// Unanchored patterns match at any level
	parts := strings.SplitSeq(cleanPath, "/")

	for part := range parts {
		if matchGlob(pattern, part) {
			return true
		}
	}

	// Also try matching the full path (for multi-segment patterns)
	return matchGlob(pattern, cleanPath)
}

// matchGlob is filepath.Match but always using forward slashes.
func matchGlob(pattern, name string) bool {
	matched, err := filepath.Match(pattern, name)

	return err == nil && matched
}

// parseGitignoreFile reads a .gitignore file and returns its patterns.
func parseGitignoreFile(path string) ([]gitignorePattern, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("open gitignore %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var patterns []gitignorePattern

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, gitignoreScannerInitBufSize), maxScannerBufferSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pat := gitignorePattern{raw: line}

		// Handle negation
		if strings.HasPrefix(line, "!") {
			pat.negated = true
			pat.raw = line[1:]
		}

		// Handle directory-only
		if strings.HasSuffix(pat.raw, "/") {
			pat.dirOnly = true
		}

		// Handle anchored patterns
		if strings.HasPrefix(pat.raw, "/") {
			pat.anchored = true
			pat.raw = strings.TrimPrefix(pat.raw, "/")
		}

		patterns = append(patterns, pat)
	}

	if err := scanner.Err(); err != nil {
		return patterns, fmt.Errorf("parse gitignore %s: %w", path, err)
	}

	return patterns, nil
}
