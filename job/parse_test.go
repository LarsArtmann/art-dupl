package job

import (
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestParse(t *testing.T) {
	ctx := t.Context()
	setup := testutil.NewTestFileSetup(t)

	testContent := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}

func helper() {
	fmt.Println("helper")
}`

	goFile := setup.GetFilePath("test.go")

	err := setup.CreateTestFile("test.go", testContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fchan := make(chan string, 2)
	fchan <- goFile

	close(fchan)

	schan, _ := Parse(ctx, fchan, golang.DetectionModeSemantic, 0, nil)

	waitForParsedNodes(
		t,
		schan,
		"Parse timed out - possible deadlock",
		"Expected some parsed nodes, got empty sequence",
	)
}

func TestParseErrorHandling(t *testing.T) {
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- "nonexistent_file.go"

	close(fchan)

	schan, _ := Parse(ctx, fchan, golang.DetectionModeSemantic, 0, nil)

	select {
	case seq := <-schan:
		// Should receive something (even if empty) for non-existent file
		_ = seq
	case <-time.After(5 * time.Second):
		t.Error("Parse should handle errors gracefully, not block")
	}
}

func TestParseMultipleFiles(t *testing.T) {
	ctx := t.Context()
	setup := testutil.NewTestFileSetup(t)

	files := map[string]string{
		"file1.go": `package main

func function1() {
	println("test1")
}`,
		"file2.go": `package main

func function1() {
	println("test2")
}`,
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	filePaths := []string{
		setup.GetFilePath("file1.go"),
		setup.GetFilePath("file2.go"),
	}

	fchan := make(chan string, 3)
	for _, file := range filePaths {
		fchan <- file
	}

	close(fchan)

	schan, _ := Parse(ctx, fchan, golang.DetectionModeSemantic, 0, nil)

	// Should receive sequences for both files
	count := 0
	for count < 2 {
		select {
		case seq := <-schan:
			if len(seq) == 0 {
				t.Error("Expected parsed nodes for valid Go file")
			}

			count++
		case <-time.After(5 * time.Second):
			t.Error("Parse timed out waiting for multiple files")

			return
		}
	}
}
