package job

import (
	"context"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIncrementalTypeAware_EnclosingReturnArity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		source   string
		arity    int32
	}{
		{
			name:     "void",
			filename: "handler.go",
			source: `package main

func handler() {
	if true {
		return
	}
}
`,
			arity: 0,
		},
		{
			name:     "non-void",
			filename: "values.go",
			source: `package main

func compute() (int, error) {
	return 42, nil
}
`,
			arity: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			setup := testutil.NewTestFileSetup(t)
			cacheDir := setup.TmpDir + "/cache"

			if err := setup.CreateTestFile(tt.filename, tt.source); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
			ctx := context.Background()

			fchan := make(chan string, 1)
			fchan <- setup.GetFilePath(tt.filename)
			close(fchan)

			schan, statsChan := parser.ParseIncremental(ctx, fchan)

			found := false
			for seq := range schan {
				for _, n := range seq {
					if n.EnclosingReturnArity == tt.arity {
						found = true
					}
				}
			}
			<-statsChan

			if !found {
				t.Errorf("expected EnclosingReturnArity == %d", tt.arity)
			}
		})
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
