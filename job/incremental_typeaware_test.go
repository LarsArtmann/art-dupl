package job

import (
	"context"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func parseIncrementalTypeAwareNodes(t *testing.T, filename, content string) []*syntax.Node {
	t.Helper()

	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	err := setup.CreateTestFile(filename, content)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath(filename)

	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	var allNodes []*syntax.Node

	for seq := range schan {
		allNodes = append(allNodes, seq...)
	}

	<-statsChan

	return allNodes
}

func TestIncrementalTypeAware_EnclosingReturnArityVoid(t *testing.T) {
	t.Parallel()

	nodes := parseIncrementalTypeAwareNodes(t, "handler.go", `package main

func handler() {
	if true {
		return
	}
}
`)

	for _, n := range nodes {
		if n.EnclosingReturnArity == 0 {
			return
		}
	}

	t.Error("expected EnclosingReturnArity == 0 for void function body nodes")
}

func TestIncrementalTypeAware_EnclosingReturnArityNonVoid(t *testing.T) {
	t.Parallel()

	nodes := parseIncrementalTypeAwareNodes(t, "values.go", `package main

func compute() (int, error) {
	return 42, nil
}
`)

	for _, n := range nodes {
		if n.EnclosingReturnArity == 2 {
			return
		}
	}

	t.Error("expected EnclosingReturnArity == 2 for (int, error) function body nodes")
}

func TestIncrementalTypeAware_SetTypeAwareData(t *testing.T) {
	t.Parallel()

	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	err := setup.CreateTestFile("test.go", `package main

func main() {}
`)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)

	// SetTypeAwareData with nil should not panic and should be a no-op
	parser.SetTypeAwareData(nil)

	// SetTypeAwareData with empty map should also be safe
	parser.SetTypeAwareData(golang.TypeAwareData{})
}
