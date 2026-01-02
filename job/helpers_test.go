package job

import (
	"os"
	"path/filepath"
	"testing"
)

// Create test files for job package testing.
func setupTestFiles(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create a simple Go file
	goFile := filepath.Join(tmpDir, "test.go")
	content := []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
}

func helper() {
	fmt.Println("helper")
}`)

	err := os.WriteFile(goFile, content, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	return goFile
}

// Test helpers for job package tests.
func setupMultipleTestFiles(t *testing.T) []string {
	tmpDir := t.TempDir()

	files := make([]string, 2)

	// First file
	file1 := filepath.Join(tmpDir, "file1.go")
	content1 := []byte(`package main

func function1() {
	println("test1")
}`)
	err := os.WriteFile(file1, content1, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	files[0] = file1

	// Second file
	file2 := filepath.Join(tmpDir, "file2.go")
	content2 := []byte(`package main

func function1() {
	println("test2")
}`)
	err = os.WriteFile(file2, content2, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}
	files[1] = file2

	return files
}
