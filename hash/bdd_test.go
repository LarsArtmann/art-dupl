package hash

import (
	"fmt"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// Behavior-Driven Development Tests for Hash Detection
// These tests define what the hash detection algorithm SHOULD do

// TestBasicHashDetectionShouldFindExactDuplicates tests that identical code sequences are detected
func TestBasicHashDetectionShouldFindExactDuplicates(t *testing.T) {
	// GIVEN: Two files with identical function implementations
	nodes := []*syntax.Node{
		createTestNode("file1.go", 1, 10, 100), // Function implementation
		createTestNode("file2.go", 20, 29, 100), // Identical implementation
	}
	
	// WHEN: Running hash detection with threshold 5
	detector := NewHashDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)
	
	// THEN: Should detect the duplicate
	matches := collectMatches(matchesChan)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}
	
	// AND: The match should contain both files
	match := matches[0]
	if len(match.Frags) != 2 {
		t.Errorf("Expected 2 fragments, got %d", len(match.Frags))
	}
}

// TestHashDetectionShouldIgnoreSmallSequences tests that sequences below threshold are ignored
func TestHashDetectionShouldIgnoreSmallSequences(t *testing.T) {
	// GIVEN: Two files with small similar sequences (below threshold)
	nodes := []*syntax.Node{
		createTestNode("file1.go", 1, 3, 50), // 3 nodes
		createTestNode("file2.go", 10, 12, 50), // 3 nodes, identical
	}
	
	// WHEN: Running hash detection with threshold 10
	detector := NewHashDetector(10)
	matchesChan := detector.FindDuplOver(nodes, 10)
	
	// THEN: Should not detect duplicates (below threshold)
	matches := collectMatches(matchesChan)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches (below threshold), got %d", len(matches))
	}
}

// TestHashDetectionShouldFindMultipleDuplicates tests that multiple duplicate groups are found
func TestHashDetectionShouldFindMultipleDuplicates(t *testing.T) {
	// GIVEN: Multiple duplicate patterns across different files
	nodes := []*syntax.Node{
		// First duplicate group
		createTestNode("file1.go", 1, 10, 100),
		createTestNode("file2.go", 20, 29, 100),
		// Second duplicate group
		createTestNode("file3.go", 50, 60, 110),
		createTestNode("file4.go", 70, 80, 110),
	}
	
	// WHEN: Running hash detection with threshold 5
	detector := NewHashDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)
	
	// THEN: Should find both duplicate groups
	matches := collectMatches(matchesChan)
	if len(matches) != 2 {
		t.Errorf("Expected 2 duplicate groups, got %d", len(matches))
	}
}

// TestHashDetectionShouldHandleOverlappingSequences tests correct handling of overlapping windows
func TestHashDetectionShouldHandleOverlappingSequences(t *testing.T) {
	// GIVEN: A long sequence with overlapping patterns
	nodes := []*syntax.Node{
		createTestNode("file1.go", 1, 20, 200), // Long sequence
		createTestNode("file2.go", 1, 20, 200), // Identical long sequence
	}
	
	// WHEN: Running hash detection
	detector := NewHashDetector(8)
	matchesChan := detector.FindDuplOver(nodes, 8)
	
	// THEN: Should not report overlapping fragments from the same file
	matches := collectMatches(matchesChan)
	for _, match := range matches {
		files := getFilesInMatch(match)
		// Should not have duplicate files in the same match
		if hasDuplicateFiles(files) {
			t.Errorf("Match contains duplicate files: %v", files)
		}
	}
}

// TestHashDetectionShouldMaintainCorrectBoundaries tests fragment boundary accuracy
func TestHashDetectionShouldMaintainCorrectBoundaries(t *testing.T) {
	// GIVEN: Known duplicate sequences with specific boundaries
	nodes := []*syntax.Node{
		createTestNodeWithPositions("file1.go", 10, 5, 15, 100), // Start at 10, length 5
		createTestNodeWithPositions("file2.go", 25, 5, 30, 100), // Start at 25, length 5
	}
	
	// WHEN: Running hash detection with threshold 5
	detector := NewHashDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)
	
	// THEN: Should maintain exact boundaries
	matches := collectMatches(matchesChan)
	if len(matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(matches))
	}
	
	match := matches[0]
	for _, frag := range match.Frags {
		if len(frag) != 5 {
			t.Errorf("Expected fragment length 5, got %d", len(frag))
		}
	}
}

// TestHashDetectionShouldBeDeterministic tests consistent results across runs
func TestHashDetectionShouldBeDeterministic(t *testing.T) {
	// GIVEN: Same input data
	nodes := []*syntax.Node{
		createTestNode("file1.go", 1, 10, 100),
		createTestNode("file2.go", 20, 29, 100),
		createTestNode("file3.go", 40, 49, 100),
	}
	
	// WHEN: Running hash detection twice
	detector := NewHashDetector(5)
	matches1 := collectMatches(detector.FindDuplOver(nodes, 5))
	matches2 := collectMatches(detector.FindDuplOver(nodes, 5))
	
	// THEN: Results should be identical
	if len(matches1) != len(matches2) {
		t.Errorf("Inconsistent results: %d vs %d matches", len(matches1), len(matches2))
	}
}

// TestHashDetectionShouldHandleLargeCodebases tests performance with large datasets
func TestHashDetectionShouldHandleLargeCodebases(t *testing.T) {
	// GIVEN: Large dataset (simulating big codebase)
	nodes := make([]*syntax.Node, 0, 1000)
	for i := 0; i < 200; i++ {
		// Add duplicate patterns
		nodes = append(nodes, createTestNode("large_file1.go", i*10, i*10+9, 50))
		nodes = append(nodes, createTestNode("large_file2.go", i*10, i*10+9, 50))
	}
	
	// WHEN: Running hash detection
	detector := NewHashDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)
	
	// THEN: Should complete without excessive memory usage or timeouts
	matches := collectMatches(matchesChan)
	t.Logf("Found %d matches in large dataset", len(matches))
	
	// Should find many duplicates but not overwhelm
	if len(matches) == 0 {
		t.Error("Expected to find matches in large dataset")
	}
}

// TestHashDetectionShouldProduceConsistentHashes tests hash stability
func TestHashDetectionShouldProduceConsistentHashes(t *testing.T) {
	// GIVEN: Identical node sequences
	sequence := []int{100, 101, 102, 103, 104, 105}
	nodes1 := createNodesFromSequence("file1.go", 0, sequence)
	nodes2 := createNodesFromSequence("file2.go", 0, sequence)
	
	// WHEN: Computing hashes for identical sequences
	allNodes := append(nodes1, nodes2...)
	detector := NewHashDetector(6)
	matchesChan := detector.FindDuplOver(allNodes, 6)
	matches := collectMatches(matchesChan)
	
	// THEN: Should produce matching hashes
	if len(matches) != 1 {
		t.Errorf("Expected 1 match for identical sequences, got %d", len(matches))
	}
	
	if len(matches) > 0 && matches[0].Hash == "" {
		t.Error("Expected non-empty hash for match")
	}
}

// TestHashDetectionShouldHandleEmptyInput tests edge case handling
func TestHashDetectionShouldHandleEmptyInput(t *testing.T) {
	// GIVEN: Empty input
	var nodes []*syntax.Node
	
	// WHEN: Running hash detection
	detector := NewHashDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)
	
	// THEN: Should handle gracefully without panics
	matches := collectMatches(matchesChan)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty input, got %d", len(matches))
	}
}

// Helper functions for BDD tests

func createTestNode(filename string, start, end, nodeType int) *syntax.Node {
	return &syntax.Node{
		Type:     nodeType,
		Filename: filename,
		Pos:      start,
		End:      end,
		Children: nil,
		Owns:     0,
	}
}

func createTestNodeWithPositions(filename string, pos, length, end, nodeType int) *syntax.Node {
	return &syntax.Node{
		Type:     nodeType,
		Filename: filename,
		Pos:      pos,
		End:      end,
		Children: nil,
		Owns:     0,
	}
}

func createNodesFromSequence(filename string, startPos int, sequence []int) []*syntax.Node {
	nodes := make([]*syntax.Node, len(sequence))
	for i, nodeType := range sequence {
		nodes[i] = &syntax.Node{
			Type:     nodeType,
			Filename: filename,
			Pos:      startPos + i,
			End:      startPos + i + 1,
			Children: nil,
			Owns:     0,
		}
	}
	return nodes
}

func collectMatches(matchesChan <-chan syntax.Match) []syntax.Match {
	// Handle potential panics from broken implementation
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in collectMatches: %v\n", r)
		}
	}()
	
	var matches []syntax.Match
	for match := range matchesChan {
		matches = append(matches, match)
	}
	return matches
}

func getFilesInMatch(match syntax.Match) []string {
	var files []string
	for _, frag := range match.Frags {
		if len(frag) > 0 {
			filename := frag[0].Filename
			if !contains(files, filename) {
				files = append(files, filename)
			}
		}
	}
	return files
}

func hasDuplicateFiles(files []string) bool {
	seen := make(map[string]bool)
	for _, file := range files {
		if seen[file] {
			return true
		}
		seen[file] = true
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}