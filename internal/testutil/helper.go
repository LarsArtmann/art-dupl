package testutil

import (
	"slices"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// AssertIncrementalStats asserts that the incremental parser stats have expected values.
// Use this to verify the results of a first run (filesCount=1, cacheMisses=1) or subsequent runs.
func AssertIncrementalStats(
	t *testing.T,
	actualFilesCount, actualCacheMisses, expectedFilesCount, expectedCacheMisses int,
) {
	t.Helper()

	if actualFilesCount != expectedFilesCount {
		t.Errorf("Expected FilesCount=%d, got %d", expectedFilesCount, actualFilesCount)
	}

	if actualCacheMisses != expectedCacheMisses {
		t.Errorf("Expected CacheMisses=%d, got %d", expectedCacheMisses, actualCacheMisses)
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
