package config

import "testing"

// AssertMergedConfig asserts that the merged config has expected values for Verbose, Paths, and IgnoreFiles.
func AssertMergedConfig(
	t *testing.T,
	merged *Config,
	expectedPath, expectedIgnoreFile, messagePrefix string,
) {
	t.Helper()

	prefix := messagePrefix
	if prefix == "" {
		prefix = "merged"
	}

	if !merged.Verbose {
		t.Errorf("Expected %s Verbose true, got %v", prefix, merged.Verbose)
	}

	if len(merged.Paths) != 1 || merged.Paths[0] != expectedPath {
		t.Errorf("Expected %s paths [%q], got %v", prefix, expectedPath, merged.Paths)
	}

	if len(merged.IgnoreFiles) != 1 || merged.IgnoreFiles[0] != expectedIgnoreFile {
		t.Errorf(
			"Expected %s ignoreFiles [%q], got %v",
			prefix,
			expectedIgnoreFile,
			merged.IgnoreFiles,
		)
	}
}
