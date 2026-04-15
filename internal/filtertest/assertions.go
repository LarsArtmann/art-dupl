package filtertest

import (
	"testing"

	"github.com/LarsArtmann/gogenfilter"
)

// AssertFileShouldNotBeFiltered asserts that a file should not be filtered.
func AssertFileShouldNotBeFiltered(t *testing.T, fltr *gogenfilter.Filter, filepath string) {
	t.Helper()

	if fltr.ShouldFilter(filepath) {
		t.Errorf("%s should not be filtered", filepath)
	}
}

// AssertFileShouldBeFiltered asserts that a file should be filtered.
func AssertFileShouldBeFiltered(t *testing.T, fltr *gogenfilter.Filter, filepath string) {
	t.Helper()

	if !fltr.ShouldFilter(filepath) {
		t.Errorf("%s should be filtered", filepath)
	}
}

// AssertFilesShouldNotBeFiltered asserts that multiple files should not be filtered.
func AssertFilesShouldNotBeFiltered(t *testing.T, fltr *gogenfilter.Filter, filepaths []string) {
	t.Helper()

	for _, filepath := range filepaths {
		AssertFileShouldNotBeFiltered(t, fltr, filepath)
	}
}

// AssertFilesShouldBeFiltered asserts that multiple files should be filtered.
func AssertFilesShouldBeFiltered(t *testing.T, fltr *gogenfilter.Filter, filepaths []string) {
	t.Helper()

	for _, filepath := range filepaths {
		AssertFileShouldBeFiltered(t, fltr, filepath)
	}
}
