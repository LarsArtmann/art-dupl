package cmd

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/gogenfilter"
)

// detectionMethodsToString converts detection methods to a comma-separated string.
func detectionMethodsToString(methods config.DetectionMethods) string {
	if len(methods) == 0 {
		return ""
	}

	result := make([]string, len(methods))
	for i, dm := range methods {
		result[i] = dm.String()
	}

	return strings.Join(result, ",")
}

// detectionMethodsToStringSlice converts detection methods to a slice of strings.
func detectionMethodsToStringSlice(methods config.DetectionMethods) []string {
	if len(methods) == 0 {
		return nil
	}

	result := make([]string, len(methods))
	for i, dm := range methods {
		result[i] = dm.String()
	}

	return result
}

// shouldIncludeFile returns true if the file should be included (not filtered).
func shouldIncludeFile(f *gogenfilter.Filter, path string) bool {
	return f == nil || !f.ShouldFilter(path)
}
