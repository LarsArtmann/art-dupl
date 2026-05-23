package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

const testTempl = "templ"

// mockReadFileContent provides a mock file content for testing.
func mockReadFileContent() []byte {
	// Return minimal Go code for testing
	return []byte(`package main

func main() {
	println("hello")
}`)
}

// assertFileContains checks if output contains expected filename and value.
func assertFileContains(t *testing.T, output, filename, value string) {
	t.Helper()

	testutil.AssertStringContains(t, output, filename, "Output should contain "+filename)
	testutil.AssertStringContains(t, output, value, "Output should contain "+value)
}

// assertMapFloat64Equal asserts that a float64 value in a map equals the expected value.
func assertMapFloat64Equal(t *testing.T, m map[string]any, key string, expected float64) {
	t.Helper()

	if m[key] != expected {
		t.Errorf("%s = %v, want %v", key, m[key], expected)
	}
}

// newTestStatsPrinter creates a stats printer with default test configuration.
func newTestStatsPrinter() (*stats, *bytes.Buffer) {
	var buf bytes.Buffer

	sp := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)

	return sp, &buf
}

// createNodeSlice creates a slice of syntax.Node with sequential positions and types.
// startPos is the starting position (inclusive), endPos is the ending position (inclusive).
func createNodeSlice(filename string, startPos, endPos int) []*syntax.Node {
	var nodes []*syntax.Node

	for i := 0; i <= endPos-startPos; i++ {
		pos := startPos + i
		end := pos + 1
		typ := i + 1
		nodes = append(
			nodes,
			testutil.CreateNodeWithPos(int32(typ), filename, int32(pos), int32(end)),
		)
	}

	return nodes
}

// createTestCloneGroups creates a standard set of clone groups for testing.
func createTestCloneGroups() [][]*syntax.Node {
	cloneGroup := func(filename string) []*syntax.Node {
		return makeDupNodePair(filename, 2, 3)
	}

	return [][]*syntax.Node{
		cloneGroup("file1.go"),
		cloneGroup("file2.go"),
	}
}

// makeDupNodePair creates a pair of duplicate nodes at the same position.
func makeDupNodePair(filename string, pos, end int32) []*syntax.Node {
	return testutil.CreateNodeSlice([]struct {
		Type     int32
		Filename string
		Pos      int32
		End      int32
	}{
		{Type: 0, Filename: filename, Pos: pos, End: end},
		{Type: 0, Filename: filename, Pos: pos, End: end},
	})
}

// printFooterAndGetData is a helper function to call PrintFooter and return stats data.
func printFooterAndGetData(t *testing.T, statsPrinter *stats) *StatsData {
	t.Helper()

	err := statsPrinter.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	data := statsPrinter.GetStatsData()

	return data
}

func TestStatsDataAggregation(t *testing.T) {
	tests := []struct {
		name       string
		duplicates [][][]*syntax.Node
		filesCount int
		threshold  int
		checkStats func(t *testing.T, stats *StatsData)
	}{
		{
			name:       "empty clones",
			duplicates: [][][]*syntax.Node{},
			filesCount: 0,
			threshold:  15,
			checkStats: func(t *testing.T, stats *StatsData) {
				t.Helper()

				testutil.AssertEqual(t, stats.TotalCloneGroups, 0, "TotalCloneGroups")
				testutil.AssertEqual(t, stats.TotalClones, 0, "TotalClones")
			},
		},
		{
			name: "single clone group with two instances",
			duplicates: [][][]*syntax.Node{
				{
					{
						testutil.CreateNodeWithPos(1, "file1.go", 1, 3),
						testutil.CreateNodeWithPos(2, "file1.go", 2, 4),
					},
					{
						testutil.CreateNodeWithPos(1, "file2.go", 10, 12),
						testutil.CreateNodeWithPos(2, "file2.go", 11, 13),
					},
				},
			},
			filesCount: 2,
			threshold:  15,
			checkStats: func(t *testing.T, stats *StatsData) {
				t.Helper()

				testutil.AssertEqual(t, stats.TotalCloneGroups, 1, "TotalCloneGroups")
				testutil.AssertEqual(t, stats.TotalClones, 2, "TotalClones")
				// TotalDuplicateLines depends on actual file content, just verify it's positive
				if stats.TotalDuplicateLines <= 0 {
					t.Errorf("TotalDuplicateLines = %d, want > 0", stats.TotalDuplicateLines)
				}

				testutil.AssertEqual(
					t,
					stats.TotalTokens,
					4,
					"TotalTokens",
				) // 2 nodes per clone × 2
			},
		},
		{
			name: "multiple clone groups",
			duplicates: [][][]*syntax.Node{
				{
					{
						testutil.CreateNodeWithPos(1, "file1.go", 1, 2),
					},
				},
				{
					{
						testutil.CreateNodeWithPos(2, "file2.go", 5, 6),
					},
				},
			},
			filesCount: 2,
			threshold:  15,
			checkStats: func(t *testing.T, stats *StatsData) {
				t.Helper()

				testutil.AssertEqual(t, stats.TotalCloneGroups, 2, "TotalCloneGroups")
				testutil.AssertEqual(t, stats.TotalClones, 2, "TotalClones")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), tt.threshold).(*stats)
			statsPrinter.SetFilesCount(tt.filesCount)

			for _, dupGroup := range tt.duplicates {
				err := statsPrinter.PrintClones(
					processTestNodes(mockReadFile(string(mockReadFileContent())), "test", dupGroup),
				)
				if err != nil {
					t.Fatalf("PrintClones failed: %v", err)
				}
			}

			err := statsPrinter.PrintFooter()
			if err != nil {
				t.Fatalf("PrintFooter failed: %v", err)
			}

			tt.checkStats(t, statsPrinter.GetStatsData())
		})
	}
}
