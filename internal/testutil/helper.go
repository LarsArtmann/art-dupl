package testutil

import (
	"slices"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// PanicRecovery returns a defer recover function that reports a panic with the given input.
// Use in fuzz tests to catch panics and report them with the input that caused the panic.
func PanicRecovery(t *testing.T, input string) func() {
	t.Helper()
	return func() {
		if r := recover(); r != nil {
			t.Errorf("Panicked with input %q: %v", input, r)
		}
	}
}

// CollectMatches collects all matches from a channel into a slice.
func CollectMatches(matchesChan <-chan syntax.Match) []syntax.Match {
	var matches []syntax.Match
	for match := range matchesChan {
		matches = append(matches, match)
	}

	return matches
}

// GetFilesInMatch extracts unique filenames from a match.
func GetFilesInMatch(match syntax.Match) []string {
	var files []string

	for _, frag := range match.Frags {
		if len(frag) > 0 {
			filename := frag[0].Filename
			if !ContainsString(files, filename) {
				files = append(files, filename)
			}
		}
	}

	return files
}

// ContainsString checks if a string slice contains a specific string.
func ContainsString(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// ContainsSubstring checks if a string contains any of the given substrings.
func ContainsSubstring(str string, substrings []string) bool {
	for _, sub := range substrings {
		if strings.Contains(str, sub) {
			return true
		}
	}

	return false
}
