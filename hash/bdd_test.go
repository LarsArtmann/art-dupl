package hash

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// Behavior-Driven Development Tests for Hash Detection
// These tests verify file-level exact duplicate detection using SHA-256 hashing

// TestBasicHashDetectionShouldFindExactDuplicates tests that identical files are detected.
func TestBasicHashDetectionShouldFindExactDuplicates(t *testing.T) {
	// GIVEN: Two identical files
	tmpDir := t.TempDir()

	// Create identical files
	content1 := `package main

func processUser(name string, age int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if age <= 0 {
		return fmt.Errorf("age must be positive")
	}
	return nil
}`

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	if err := os.WriteFile(file1, []byte(content1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content1), 0o644); err != nil {
		t.Fatal(err)
	}

	// Parse files to get nodes
	node1, err := golang.Parse(file1)
	if err != nil {
		t.Fatalf("Failed to parse file1: %v", err)
	}
	node2, err := golang.Parse(file2)
	if err != nil {
		t.Fatalf("Failed to parse file2: %v", err)
	}

	nodes := []*syntax.Node{node1, node2}

	// WHEN: Running hash detection with threshold 5
	detector := NewHashDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)

	// THEN: Should detect the duplicate
	matches := collectMatches(matchesChan)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
		return
	}

	// AND: The match should contain both files
	match := matches[0]
	if len(match.Frags) != 2 {
		t.Errorf("Expected 2 fragments, got %d", len(match.Frags))
	}

	// AND: Verify files in match
	filenames := getFilesInMatch(match)
	foundFile1 := false
	foundFile2 := false
	for _, fn := range filenames {
		if strings.HasSuffix(fn, "file1.go") {
			foundFile1 = true
		}
		if strings.HasSuffix(fn, "file2.go") {
			foundFile2 = true
		}
	}

	if !foundFile1 || !foundFile2 {
		t.Errorf("Expected to find both files in match, got: %v", filenames)
	}
}

// TestHashDetectionShouldIgnoreSmallFiles tests that small files below threshold are ignored.
func TestHashDetectionShouldIgnoreSmallFiles(t *testing.T) {
	// GIVEN: Two small identical files (below threshold)
	tmpDir := t.TempDir()

	// Create small files
	content := `package main

func hello() {
	println("hi")
}`

	file1 := filepath.Join(tmpDir, "small1.go")
	file2 := filepath.Join(tmpDir, "small2.go")

	if err := os.WriteFile(file1, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Parse files
	node1, err := golang.Parse(file1)
	if err != nil {
		t.Fatalf("Failed to parse file1: %v", err)
	}
	node2, err := golang.Parse(file2)
	if err != nil {
		t.Fatalf("Failed to parse file2: %v", err)
	}

	nodes := []*syntax.Node{node1, node2}

	// WHEN: Running hash detection with high threshold (1000 bytes)
	detector := NewHashDetector(1000)
	matchesChan := detector.FindDuplOver(nodes, 1000)

	// THEN: Should not detect duplicates (files too small)
	matches := collectMatches(matchesChan)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches (files too small), got %d", len(matches))
	}
}

// TestHashDetectionShouldFindMultipleDuplicateGroups tests that multiple duplicate groups are found.
func TestHashDetectionShouldFindMultipleDuplicateGroups(t *testing.T) {
	// GIVEN: Multiple files with duplicate patterns
	tmpDir := t.TempDir()

	// Create files with two distinct patterns
	content1 := `package main

func processUser(name string, age int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if age <= 0 {
		return fmt.Errorf("age must be positive")
	}
	return nil
}`

	content2 := `package main

func processProduct(name string, price int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	return nil
}`

	// Create 2 files with content1 and 2 files with content2
	file1 := filepath.Join(tmpDir, "user1.go")
	file2 := filepath.Join(tmpDir, "user2.go")
	file3 := filepath.Join(tmpDir, "product1.go")
	file4 := filepath.Join(tmpDir, "product2.go")

	if err := os.WriteFile(file1, []byte(content1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file3, []byte(content2), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file4, []byte(content2), 0o644); err != nil {
		t.Fatal(err)
	}

	// Parse all files
	nodes := make([]*syntax.Node, 0, 4)
	files := []string{file1, file2, file3, file4}
	for _, file := range files {
		node, err := golang.Parse(file)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", file, err)
		}
		nodes = append(nodes, node)
	}

	// WHEN: Running hash detection with threshold 5
	detector := NewHashDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)

	// THEN: Should find both duplicate groups
	matches := collectMatches(matchesChan)
	t.Logf("Found %d matches (expected 2)", len(matches))
	if len(matches) != 2 {
		t.Errorf("Expected 2 duplicate groups, got %d", len(matches))
	}
}

// TestHashDetectionShouldHandleEmptyInput tests handling of no input.
func TestHashDetectionShouldHandleEmptyInput(t *testing.T) {
	// GIVEN: No input files
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

func collectMatches(matchesChan <-chan syntax.Match) []syntax.Match {
	//nolint:prealloc // Can't preallocate for channel inputs
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

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
