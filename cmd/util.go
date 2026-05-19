package cmd

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/gogenfilter/v3"
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

// newReportMetadata creates ReportMetadata from a Config and sort criteria string.
func newReportMetadata(cfg *config.Config, sortBy string) printer.ReportMetadata {
	return printer.ReportMetadata{
		Semantic:         cfg.Semantic,
		DetectionMethods: detectionMethodsToStringSlice(cfg.DetectionMethods),
		SortBy:           sortBy,
		IncludeSQLC:      cfg.IncludeSQLC,
		IncludeTempl:     cfg.IncludeTempl,
	}
}

// shouldIncludeFile returns true if the file should be included (not filtered).
// When stats is non-nil, the filter result is recorded for aggregation.
func shouldIncludeFile(f *gogenfilter.Filter, path string, stats *FilterStats) bool {
	if f == nil {
		return true
	}

	result, err := f.FilterDetailed(path)
	if err != nil {
		return true
	}

	stats.Record(result)

	return !result.Filtered
}
