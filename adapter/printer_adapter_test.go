package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// testCloneGroup creates a CloneGroup with clones for the given filenames.
func testCloneGroup(filenames ...string) domain.CloneGroup {
	clones := make([]domain.Clone, len(filenames))
	for i, f := range filenames {
		clones[i] = domain.Clone{Filename: domain.GlobalPool().Intern(f)}
	}

	return domain.CloneGroup{Clones: clones}
}

// createTestCloneGroup creates a standardized CloneGroup with 3 clones for testing
func createTestCloneGroup(id string, size int, severity domain.CloneSeverity) domain.CloneGroup {
	return domain.CloneGroup{
		ID:       domain.CloneGroupID(id),
		Clones:   []domain.Clone{{}, {}, {}}, // 3 clones = 2 duplicates
		Size:     uint(size),
		Severity: severity,
	}
}

// testNode creates a sample syntax node for testing.
func testNode() *syntax.Node {
	return &syntax.Node{
		Type:     1,
		Pos:      0,
		End:      10,
		Owns:     5,
		Filename: "test.go",
	}
}

// testNodes creates multiple nodes for testing.
func testNodes() [][]*syntax.Node {
	return [][]*syntax.Node{
		{
			{Type: 1, Pos: 0, End: 20, Filename: "file1.go"},
			{Type: 2, Pos: 5, End: 15, Filename: "file1.go"},
		},
		{
			{Type: 3, Pos: 100, End: 150, Filename: "file2.go"},
		},
	}
}

// TestNodeToDomainClone tests node to domain clone conversion.
func TestNodeToDomainClone(t *testing.T) {
	t.Run("basic conversion without file", func(t *testing.T) {
		node := testNode()
		clone := NodeToDomainClone(node, "test.go")

		if clone.Status != domain.FileProcessingStateCompleted {
			t.Errorf(
				"Expected status %s, got %s",
				domain.FileProcessingStateCompleted,
				clone.Status,
			)
		}

		if clone.FilenameString() != "test.go" {
			t.Errorf("Expected filename 'test.go', got %q", clone.FilenameString())
		}
	})

	t.Run("with existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "existing.go")

		content := "package main\n\nfunc main() {}\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		node := &syntax.Node{
			Type:     1,
			Pos:      0,
			End:      int32(len(content)),
			Filename: testFile,
		}

		clone := NodeToDomainClone(node, testFile)

		if clone.FilenameString() != testFile {
			t.Errorf("Expected filename %q, got %q", testFile, clone.FilenameString())
		}
		// Fragment should contain the file content
		if clone.FragmentString() != content {
			t.Errorf("Expected fragment to be file content, got %q", clone.FragmentString())
		}
	})

	t.Run("edge case filenames", func(t *testing.T) {
		tests := []struct {
			name     string
			filename string
		}{
			{"non-existent file", "/nonexistent/file.go"},
			{"empty filename", ""},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := testNode()
				clone := NodeToDomainClone(node, tt.filename)

				if clone.Status != domain.FileProcessingStateCompleted {
					t.Errorf(
						"Expected status %s, got %s",
						domain.FileProcessingStateCompleted,
						clone.Status,
					)
				}
			})
		}
	})

	t.Run("with children", func(t *testing.T) {
		node := &syntax.Node{
			Type:     1,
			Pos:      0,
			End:      50,
			Filename: "parent.go",
			Children: []*syntax.Node{
				{Type: 2, Pos: 10, End: 30, Filename: "parent.go"},
			},
		}

		clone := NodeToDomainClone(node, "parent.go")

		// Should handle nodes with children
		if clone.Complexity == 0 {
			t.Error("Expected non-zero complexity for node with children")
		}
	})
}

// TestCloneGroupFromNodes tests clone group creation from nodes.
func TestCloneGroupFromNodes(t *testing.T) {
	t.Run("basic group creation", func(t *testing.T) {
		nodes := testNodes()
		group := CloneGroupFromNodes("group-1", nodes)

		if group.ID != "group-1" {
			t.Errorf("Expected ID 'group-1', got %q", group.ID)
		}

		if len(group.Clones) == 0 {
			t.Error("Expected at least one clone in group")
		}

		if group.Hash != "group-hash" {
			t.Errorf("Expected hash 'group-hash', got %q", group.Hash)
		}

		if group.Status != domain.FileProcessingStateCompleted {
			t.Errorf(
				"Expected status %s, got %s",
				domain.FileProcessingStateCompleted,
				group.Status,
			)
		}
	})

	t.Run("size calculation", func(t *testing.T) {
		nodes := [][]*syntax.Node{
			{
				{Type: 1, Pos: 0, End: 100, Filename: "test.go"},
			},
		}

		group := CloneGroupFromNodes("group-1", nodes)

		// Size should be End - Pos = 100
		if group.Size != 100 {
			t.Errorf("Expected size 100, got %d", group.Size)
		}
	})

	t.Run("empty node groups", func(t *testing.T) {
		nodes := [][]*syntax.Node{
			{}, // Empty group
			{}, // Another empty group
		}

		group := CloneGroupFromNodes("group-1", nodes)

		// Should handle empty groups gracefully
		if len(group.Clones) != 0 {
			t.Errorf("Expected 0 clones, got %d", len(group.Clones))
		}
	})

	t.Run("mixed empty and filled groups", func(t *testing.T) {
		nodes := [][]*syntax.Node{
			{},
			{{Type: 1, Pos: 0, End: 50, Filename: "test.go"}},
			{},
		}

		group := CloneGroupFromNodes("group-1", nodes)

		// Should only include non-empty groups
		if len(group.Clones) != 1 {
			t.Errorf("Expected 1 clone, got %d", len(group.Clones))
		}
	})

	t.Run("severity calculation", func(t *testing.T) {
		// Large size should result in critical severity
		nodes := [][]*syntax.Node{
			{
				{Type: 1, Pos: 0, End: 300, Filename: "large.go"}, // size = 300 > 200 = critical
			},
		}

		group := CloneGroupFromNodes("group-1", nodes)

		if group.Severity != domain.CloneSeverityCritical {
			t.Errorf("Expected severity %s, got %s", domain.CloneSeverityCritical, group.Severity)
		}
	})

	t.Run("multiple nodes in one group", func(t *testing.T) {
		nodes := [][]*syntax.Node{
			{
				{Type: 1, Pos: 0, End: 50, Filename: "file1.go"},
				{Type: 2, Pos: 60, End: 100, Filename: "file1.go"},
			},
		}

		group := CloneGroupFromNodes("group-1", nodes)

		// Should have 2 clones
		if len(group.Clones) != 2 {
			t.Errorf("Expected 2 clones, got %d", len(group.Clones))
		}
	})
}

// TestCreateAnalysisFromClones tests analysis creation from clone data.
func TestCreateAnalysisFromClones(t *testing.T) {
	t.Run("empty clone groups", func(t *testing.T) {
		analysis := CreateAnalysisFromClones(nil, 15)

		if analysis.ID != "analysis-id" {
			t.Errorf("Expected ID 'analysis-id', got %q", analysis.ID)
		}

		if analysis.State != domain.DetectionStateCompleted {
			t.Errorf("Expected state %s, got %s", domain.DetectionStateCompleted, analysis.State)
		}

		if analysis.Mode != domain.AnalysisModeFull {
			t.Errorf("Expected mode %s, got %s", domain.AnalysisModeFull, analysis.Mode)
		}

		if analysis.Threshold != 15 {
			t.Errorf("Expected threshold 15, got %d", analysis.Threshold)
		}

		if len(analysis.CloneGroups) != 0 {
			t.Errorf("Expected 0 clone groups, got %d", len(analysis.CloneGroups))
		}
	})

	t.Run("with clone groups", func(t *testing.T) {
		groups := []domain.CloneGroup{
			{
				ID:     "group-1",
				Clones: []domain.Clone{{}}, // One clone
				Size:   100,
			},
			{
				ID:     "group-2",
				Clones: []domain.Clone{{}, {}}, // Two clones
				Size:   200,
			},
		}

		analysis := CreateAnalysisFromClones(groups, 20)

		if analysis.Threshold != 20 {
			t.Errorf("Expected threshold 20, got %d", analysis.Threshold)
		}

		if len(analysis.CloneGroups) != 2 {
			t.Errorf("Expected 2 clone groups, got %d", len(analysis.CloneGroups))
		}

		if analysis.Stats.TotalClones != 3 {
			t.Errorf("Expected 3 total clones, got %d", analysis.Stats.TotalClones)
		}

		if analysis.Stats.TotalTokenSize != 300 {
			t.Errorf("Expected 300 total token size, got %d", analysis.Stats.TotalTokenSize)
		}
	})

	t.Run("stats calculation", func(t *testing.T) {
		groups := []domain.CloneGroup{
			createTestCloneGroup("group-1", 100, domain.CloneSeverityMedium),
		}

		analysis := CreateAnalysisFromClones(groups, 15)

		if analysis.Stats.TotalClones != 3 {
			t.Errorf("Expected 3 total clones, got %d", analysis.Stats.TotalClones)
		}

		if analysis.Stats.FilesAnalyzed == 0 {
			t.Error("Expected files analyzed to be calculated")
		}
	})

	t.Run("duplication ratio", func(t *testing.T) {
		// 3 clones in one group = 2 duplicates
		groups := []domain.CloneGroup{
			createTestCloneGroup("group-1", 100, domain.CloneSeverityMedium),
		}

		analysis := CreateAnalysisFromClones(groups, 15)

		// 2 duplicates / 3 total = 0.666...
		expectedRatio := 2.0 / 3.0
		if analysis.Stats.DuplicationRatio < expectedRatio-0.01 ||
			analysis.Stats.DuplicationRatio > expectedRatio+0.01 {
			t.Errorf(
				"Expected duplication ratio ~%.2f, got %.2f",
				expectedRatio,
				analysis.Stats.DuplicationRatio,
			)
		}
	})

	t.Run("complexity score", func(t *testing.T) {
		groups := []domain.CloneGroup{
			{ID: "group-1", Clones: []domain.Clone{{}}, Size: 100},
			{ID: "group-2", Clones: []domain.Clone{{}}, Size: 200},
		}

		analysis := CreateAnalysisFromClones(groups, 15)

		// Complexity = totalSize / (numGroups + 1) = 300 / 3 = 100
		expectedScore := 100.0
		if analysis.Stats.ComplexityScore != expectedScore {
			t.Errorf(
				"Expected complexity score %.2f, got %.2f",
				expectedScore,
				analysis.Stats.ComplexityScore,
			)
		}
	})
}

// TestCalculateDuplicationRatio tests duplication ratio calculation.
func TestCalculateDuplicationRatio(t *testing.T) {
	tests := []struct {
		name     string
		groups   []domain.CloneGroup
		expected float64
	}{
		{
			name:     "nil groups",
			groups:   nil,
			expected: 0.0,
		},
		{
			name:     "empty groups",
			groups:   []domain.CloneGroup{},
			expected: 0.0,
		},
		{
			name: "single clone - no duplication",
			groups: []domain.CloneGroup{
				{Clones: []domain.Clone{{}}},
			},
			expected: 0.0,
		},
		{
			name: "two clones - 50% duplication",
			groups: []domain.CloneGroup{
				{Clones: []domain.Clone{{}, {}}},
			},
			expected: 0.5,
		},
		{
			name: "three clones - 66.6% duplication",
			groups: []domain.CloneGroup{
				{Clones: []domain.Clone{{}, {}, {}}},
			},
			expected: 2.0 / 3.0,
		},
		{
			name: "multiple groups",
			groups: []domain.CloneGroup{
				{Clones: []domain.Clone{{}, {}}},     // 1 duplicate
				{Clones: []domain.Clone{{}, {}, {}}}, // 2 duplicates
			},
			expected: 3.0 / 5.0, // 3 duplicates / 5 total
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDuplicationRatio(tt.groups)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("Expected ratio %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

// TestCountUniqueFiles tests unique file counting.
func TestCountUniqueFiles(t *testing.T) {
	t.Run("empty groups", func(t *testing.T) {
		count := countUniqueFiles(nil)
		if count != 0 {
			t.Errorf("Expected 0 unique files, got %d", count)
		}
	})

	t.Run("single file", func(t *testing.T) {
		groups := []domain.CloneGroup{
			{
				Clones: []domain.Clone{
					{}, // Empty filename
				},
			},
		}

		count := countUniqueFiles(groups)
		if count != 1 {
			t.Errorf("Expected 1 unique file, got %d", count)
		}
	})

	t.Run("multiple unique files", func(t *testing.T) {
		groups := []domain.CloneGroup{
			testCloneGroup("file1.go", "file2.go"),
			testCloneGroup("file1.go", "file3.go"), // file1.go is duplicate
		}

		count := countUniqueFiles(groups)
		if count != 3 {
			t.Errorf("Expected 3 unique files, got %d", count)
		}
	})
}

// TestPrinterAdapter tests the PrinterAdapter struct.
func TestPrinterAdapter(t *testing.T) {
	adapter := PrinterAdapter{}
	// Just verify it exists
	if adapter != (PrinterAdapter{}) {
		t.Error("Expected empty PrinterAdapter")
	}
}
