package job

import (
	"context"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIncrementalTypeAware_EnclosingReturnArityPopulated(t *testing.T) {
	t.Parallel()

	setup := newIncrementalTestSetup(t)
	defer setup.cleanup()

	// Write a file with a void-return function (arity 0)
	err := setup.writeFile("handler.go", `package main

import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	_ = w
}
`)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewIncrementalParser(setup.TmpDir+"/cache", false, golang.DetectionModeSemantic, 0, 0)

	ctx := context.Background()
	fchan := make(chan string, 1)
	fchan <- setup.TmpDir + "/handler.go"
	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)
	nodes := collectNodes(schan)
	<-statsChan

	if len(nodes) == 0 {
		t.Fatal("expected at least one node sequence")
	}

	// Verify EnclosingReturnArity was computed on at least some nodes.
	// In void functions (handler), arity should be 0.
	foundAny := false
	for _, seq := range nodes {
		for _, n := range seq {
			if n.EnclosingReturnArity == 0 {
				foundAny = true
				break
			}
		}
		if foundAny {
			break
		}
	}

	if !foundAny {
		t.Error("expected EnclosingReturnArity to be 0 for void function nodes")
	}
}

func TestIncrementalTypeAware_NonVoidFunctionHasArity(t *testing.T) {
	t.Parallel()

	setup := newIncrementalTestSetup(t)
	defer setup.cleanup()

	err := setup.writeFile("values.go", `package main

func compute() (int, error) {
	return 42, nil
}
`)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewIncrementalParser(setup.TmpDir+"/cache", false, golang.DetectionModeSemantic, 0, 0)

	ctx := context.Background()
	fchan := make(chan string, 1)
	fchan <- setup.TmpDir + "/values.go"
	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)
	nodes := collectNodes(schan)
	<-statsChan

	if len(nodes) == 0 {
		t.Fatal("expected at least one node sequence")
	}

	// Verify that at least some nodes inside compute() have EnclosingReturnArity == 2
	foundArity := false
	for _, seq := range nodes {
		for _, n := range seq {
			if n.EnclosingReturnArity == 2 {
				foundArity = true
				break
			}
		}
		if foundArity {
			break
		}
	}

	if !foundArity {
		t.Error("expected EnclosingReturnArity == 2 for (int, error) function body nodes")
	}
}
