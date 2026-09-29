package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// Go 1.27 legalized methods with their own type parameters (generic
// methods, #77273). These fixtures pin that the pipeline — parse,
// serialize, normalize — handles them: the receiver, the METHOD's type
// parameters, the parameters, and body locals must all canonicalize, and
// serialization must stay idempotent.

const genericMethodSrc = `package test

type Store[T any] struct{ items []T }

func (s *Store[T]) Load() []T {
	out := []T{}
	for _, it := range s.items {
		out = append(out, it)
	}
	return out
}
`

func TestGenericMethod_SerializeIdempotent(t *testing.T) {
	t.Parallel()

	// ADR-0006: serialization must never mutate the tree — serialize the
	// SAME parsed generic method twice and require identical streams.
	node, err := ParseWithConfig(parseAndWriteTempSrc(t, genericMethodSrc), ParseConfig{Mode: DetectionModeSemantic})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	stream1 := syntax.Serialize(node)
	stream2 := syntax.Serialize(node)

	if len(stream1) != len(stream2) {
		t.Fatalf("stream length drift after re-serialize: %d vs %d", len(stream1), len(stream2))
	}

	for i := range stream1 {
		if stream1[i].Val() != stream2[i].Val() || stream1[i].Statement != stream2[i].Statement {
			t.Fatalf("token %d drifted after re-serialize: %v vs %v", i, stream1[i].Val(), stream2[i].Val())
		}
	}
}

// parseAndWriteTempSrc writes src to a temp file and returns its path.
func parseAndWriteTempSrc(t *testing.T, src string) string {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "generic_method.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	return tmpFile
}

func TestGenericMethod_Type2RenamedMatches(t *testing.T) {
	t.Parallel()

	// Scope: alpha-normalization canonicalizes FUNCTION-scoped names
	// (receiver variable, method type params, locals). The generic TYPE
	// declaration (Store[T]) is package-level API surface by design and
	// stays in the hash — so the renamed fixture keeps the same type and
	// method name and renames only the receiver variable and locals.
	renamed := `package test

type Store[T any] struct{ items []T }

func (r *Store[T]) Load() []T {
	result := []T{}
	for _, item := range r.items {
		result = append(result, item)
	}
	return result
}
`

	tokensA := parseAndSerializeT(t, genericMethodSrc, DetectionModeSemantic)
	tokensB := parseAndSerializeT(t, renamed, DetectionModeSemantic)

	if !tokensEqualKV(tokensA, tokensB) {
		t.Error("renamed generic method (receiver, type param, locals) must match as Type-2 clone in semantic mode")
	}
}

func TestGenericMethod_DifferentBodyDiverges(t *testing.T) {
	t.Parallel()

	reversed := `package test

type Store[T any] struct{ items []T }

func (s *Store[T]) Load() []T {
	out := []T{}
	for i := len(s.items) - 1; i >= 0; i-- {
		out = append(out, s.items[i])
	}
	return out
}
`

	tokensA := parseAndSerializeT(t, genericMethodSrc, DetectionModeSemantic)
	tokensB := parseAndSerializeT(t, reversed, DetectionModeSemantic)

	if tokensEqualKV(tokensA, tokensB) {
		t.Error("generic methods with different loop bodies must NOT match")
	}
}
