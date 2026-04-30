// Package hash_test tests the hash-based duplicate detection.
package hash

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// Behavior-Driven Development Tests for Hash Detection
// These tests verify file-level exact duplicate detection using XXH3 hashing

// TestBasicHashDetectionShouldFindExactDuplicates tests that identical files are detected.
func TestBasicHashDetectionShouldFindExactDuplicates(
	t *testing.T,
) { // BDD-style test with multiple test scenarios
	// GIVEN: Two identical files
	setup := testutil.NewTestFileSetup(t)

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

	err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, content1)
	if err != nil {
		t.Fatal(err)
	}

	// Parse files to get nodes
	file1 := setup.GetFilePath("file1.go")
	file2 := setup.GetFilePath("file2.go")
	node1 := testutil.ParseFile(t, file1)
	node2 := testutil.ParseFile(t, file2)

	nodes := []*syntax.Node{node1, node2}

	// WHEN: Running hash detection with threshold 5
	detector := NewFileDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)

	// THEN: Should detect duplicate
	matches := testutil.CollectMatches(matchesChan)
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
	filenames := testutil.GetFilesInMatch(match)
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
	setup := testutil.NewTestFileSetup(t)

	// Create small files
	content := `package main

func hello() {
	println("hi")
}`

	err := setup.CreateDuplicateFiles([]string{"small1.go", "small2.go"}, content)
	if err != nil {
		t.Fatal(err)
	}

	// Parse files
	file1 := setup.GetFilePath("small1.go")
	file2 := setup.GetFilePath("small2.go")
	node1 := testutil.ParseFile(t, file1)
	node2 := testutil.ParseFile(t, file2)

	nodes := []*syntax.Node{node1, node2}

	// WHEN: Running hash detection with high threshold (1000 bytes)
	detector := NewFileDetector(1000)
	matchesChan := detector.FindDuplOver(nodes, 1000)

	// THEN: Should not detect duplicates (files too small)
	matches := testutil.CollectMatches(matchesChan)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches (files too small), got %d", len(matches))
	}
}

// TestHashDetectionShouldFindMultipleDuplicateGroups tests that multiple duplicate groups are found.
func TestHashDetectionShouldFindMultipleDuplicateGroups(t *testing.T) {
	// GIVEN: Multiple files with duplicate patterns
	setup := testutil.NewTestFileSetup(t)

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
	files := map[string]string{
		"user1.go":    content1,
		"user2.go":    content1,
		"product1.go": content2,
		"product2.go": content2,
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatal(err)
	}

	// Parse all files
	filePaths := []string{
		setup.GetFilePath("user1.go"),
		setup.GetFilePath("user2.go"),
		setup.GetFilePath("product1.go"),
		setup.GetFilePath("product2.go"),
	}
	nodes := testutil.ParseFiles(t, filePaths)

	// WHEN: Running hash detection with threshold 5
	detector := NewFileDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)

	// THEN: Should find both duplicate groups
	matches := testutil.CollectMatches(matchesChan)
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
	detector := NewFileDetector(5)
	matchesChan := detector.FindDuplOver(nodes, 5)

	// THEN: Should handle gracefully without panics
	matches := testutil.CollectMatches(matchesChan)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty input, got %d", len(matches))
	}
}
