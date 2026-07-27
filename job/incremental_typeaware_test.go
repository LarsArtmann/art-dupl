package job

import (
	"context"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIncrementalTypeAware_EnclosingReturnArityVoid(t *testing.T) {
	t.Parallel()

	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	err := setup.CreateTestFile("handler.go", `package main

func handler() {
	if true {
		return
	}
}
`)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("handler.go")

	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	var foundVoidArity bool

	for seq := range schan {
		for _, n := range seq {
			if n.EnclosingReturnArity == 0 {
				foundVoidArity = true
			}
		}
	}

	<-statsChan

	if !foundVoidArity {
		t.Error("expected EnclosingReturnArity == 0 for void function body nodes")
	}
}

func TestIncrementalTypeAware_EnclosingReturnArityNonVoid(t *testing.T) {
	t.Parallel()

	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	err := setup.CreateTestFile("values.go", `package main

func compute() (int, error) {
	return 42, nil
}
`)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("values.go")

	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	var foundArityTwo bool

	for seq := range schan {
		for _, n := range seq {
			if n.EnclosingReturnArity == 2 {
				foundArityTwo = true
			}
		}
	}

	<-statsChan

	if !foundArityTwo {
		t.Error("expected EnclosingReturnArity == 2 for (int, error) function body nodes")
	}
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

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)

	// SetTypeAwareData with nil should not panic and should be a no-op
	parser.SetTypeAwareData(nil)

	// SetTypeAwareData with empty map should also be safe
	parser.SetTypeAwareData(golang.TypeAwareData{})
}
