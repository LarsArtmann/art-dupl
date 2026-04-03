// Package testutil provides testing utilities for art-dupl.
package testutil

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/onsi/ginkgo/v2"
)

// RequireGolden compares output against a golden file, printing a diff if they don't match.
// The golden file is stored in testdata/{testName}.golden relative to the calling file.
//
// To update golden files, run:
//
//	go test -update ./...
func RequireGolden(tb testing.TB, output []byte) {
	tb.Helper()
	golden.RequireEqual(tb, output)
}

// RequireGoldenString is like RequireGolden but takes a string.
func RequireGoldenString(tb testing.TB, output string) {
	tb.Helper()
	golden.RequireEqual(tb, output)
}

// ginkgoTB extracts a testing.TB from Ginkgo's test context.
// Panics if not called from within a Ginkgo test.
func ginkgoTB() testing.TB {
	ginkgoT := ginkgo.GinkgoT()
	if tb, ok := ginkgoT.(testing.TB); ok {
		return tb
	}

	panic("must be called from within a Ginkgo test")
}

// GinkgoRequireGolden is like RequireGolden but uses Ginkgo's test context automatically.
func GinkgoRequireGolden(output []byte) {
	golden.RequireEqual(ginkgoTB(), output)
}

// GinkgoRequireGoldenString is like GinkgoRequireGolden but takes a string.
func GinkgoRequireGoldenString(output string) {
	golden.RequireEqual(ginkgoTB(), output)
}
