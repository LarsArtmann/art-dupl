package filtertest

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/gogenfilter/v3"
)

// toFSPath converts an absolute path to a path usable with os.DirFS("/").
// Go's fs.ReadFile rejects paths starting with /, so we strip the leading slash.
func toFSPath(path string) string {
	return strings.TrimPrefix(path, "/")
}

// runFilter calls Filter on the given path and fails the test on error.
// Returns whether the path should be filtered.
func runFilter(t *testing.T, f *gogenfilter.Filter, path string) bool {
	t.Helper()

	filtered, err := f.Filter(toFSPath(path))
	if err != nil {
		t.Fatalf("Filter(%q) error: %v", path, err)
	}

	return filtered
}

// AssertFileShouldNotBeFiltered asserts that a file should not be filtered.
func AssertFileShouldNotBeFiltered(t *testing.T, f *gogenfilter.Filter, filepath string) {
	t.Helper()

	if runFilter(t, f, filepath) {
		t.Errorf("%s should not be filtered", filepath)
	}
}

// AssertFileShouldBeFiltered asserts that a file should be filtered.
func AssertFileShouldBeFiltered(t *testing.T, f *gogenfilter.Filter, path string) {
	t.Helper()

	if !runFilter(t, f, path) {
		t.Errorf("%s should be filtered", path)
	}
}

// AssertFilesShouldNotBeFiltered asserts that multiple files should not be filtered.
func AssertFilesShouldNotBeFiltered(t *testing.T, f *gogenfilter.Filter, filepaths []string) {
	t.Helper()

	for _, filepath := range filepaths {
		AssertFileShouldNotBeFiltered(t, f, filepath)
	}
}

// AssertFilesShouldBeFiltered asserts that multiple files should be filtered.
func AssertFilesShouldBeFiltered(t *testing.T, f *gogenfilter.Filter, paths []string) {
	t.Helper()

	for _, p := range paths {
		AssertFileShouldBeFiltered(t, f, p)
	}
}

// AssertFilesFiltered asserts that files should (or should not) be filtered based on shouldFilter.
func AssertFilesFiltered(
	t *testing.T,
	f *gogenfilter.Filter,
	filepaths []string,
	shouldFilter bool,
) {
	t.Helper()

	for _, filepath := range filepaths {
		if shouldFilter {
			AssertFileShouldBeFiltered(t, f, filepath)
		} else {
			AssertFileShouldNotBeFiltered(t, f, filepath)
		}
	}
}
