package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templpkg "github.com/LarsArtmann/art-dupl/syntax/templ"
)

// Real-source fixture mirroring the webphone finding (2026-11-22): .templ
// helper functions whose signatures merely END alike — `..., lang Lang)
// string` across many view helpers — surface the shared tail as its own
// clone fragment at low thresholds. Nothing is extractable from identical
// signature parts, so the group is noise.
const signatureFragmentTemplSrc = `package main

type Lang string

func metaA(count int, lang Lang) string {
	return "a"
}

func metaB(size int, lang Lang) string {
	return "b"
}

templ List(lang Lang) {
	<div>{ "static" }</div>
}
`

// signatureFragmentFields returns the Go-snippet Field roots of the shared
// `lang Lang` parameter, the exact node the suffix tree matches in the real
// pipeline.
func signatureFragmentFields(t *testing.T) []*syntax.Node {
	t.Helper()

	root, _, err := templpkg.ParseBytes("test.templ", []byte(signatureFragmentTemplSrc))
	if err != nil {
		t.Fatalf("parse templ: %v", err)
	}

	var fields []*syntax.Node

	var walk func(n *syntax.Node)

	walk = func(n *syntax.Node) {
		if golang.DecodeBaseType(n.Type) == golang.Field && n.GoSnippet &&
			len(n.Children) > 0 && n.Children[0].Name == "lang" {
			fields = append(fields, n)
		}

		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)

	return fields
}

// TestIsSignatureFragmentMatch_RealTemplSource is the definitive test: it
// parses ACTUAL .templ source and verifies the pattern fires on the real
// (non-hand-built) Go-snippet Field nodes through the evaluation pipeline.
func TestIsSignatureFragmentMatch_RealTemplSource(t *testing.T) {
	t.Parallel()

	fields := signatureFragmentFields(t)
	if len(fields) < 2 {
		t.Fatalf("expected ≥2 Go-snippet `lang` Fields (metaA, metaB), got %d", len(fields))
	}

	for _, field := range fields {
		if !field.GoSnippet {
			t.Fatalf("field %q not marked GoSnippet — markGoSnippet must cover FuncDecl params", field.Name)
		}
	}

	seqs := make([][]*domain.CloneNode, len(fields))
	for i, field := range fields {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(field)}
	}

	if !isSignatureFragmentMatch(seqs) {
		t.Error("isSignatureFragmentMatch = false for real `lang Lang` snippet fields")
	}

	label, action := EvaluateActionabilityWithLabel(seqs)
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternSignatureFragment {
		t.Errorf("label = %q, want %q", label, PatternSignatureFragment)
	}
}

// TestIsSignatureFragmentMatch_MarkupRootsDoNotMatch pins the enum-collision
// guard: templ markup nodes are a parallel iota and must never be read
// through the golang table. Element-rooted fragments stay visible.
func TestIsSignatureFragmentMatch_MarkupRootsDoNotMatch(t *testing.T) {
	t.Parallel()

	root, _, err := templpkg.ParseBytes("test.templ", []byte(signatureFragmentTemplSrc))
	if err != nil {
		t.Fatalf("parse templ: %v", err)
	}

	elements := findNodesByBaseType(root, templpkg.Element)
	if len(elements) == 0 {
		t.Fatal("expected Element nodes in fixture")
	}

	seqs := make([][]*domain.CloneNode, len(elements))
	for i, element := range elements {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(element)}
	}

	if isSignatureFragmentMatch(seqs) {
		t.Error("isSignatureFragmentMatch = true for markup Element roots")
	}
}

// TestIsSignatureFragmentMatch_PlainGoFieldDoesNotMatch pins the GoSnippet
// gate: a Field root from a plain .go file must not be suppressed by this
// pattern (identical struct fields in pure-Go files may be worth merging,
// and statement-bearing .go files already reject non-statement roots in
// FindSyntaxUnits). Hand-built nodes are safe here: the check reads two
// root scalars only, no child-order sensitivity (cf. the 2026-10-11
// real-source lesson, which was about child ordering).
func TestIsSignatureFragmentMatch_PlainGoFieldDoesNotMatch(t *testing.T) {
	t.Parallel()

	plain := &domain.CloneNode{BaseType: golang.Field, Filename: "a.go"}
	snippet := &domain.CloneNode{BaseType: golang.Field, Filename: "a.templ", GoSnippet: true}

	if isSignatureFragmentMatch([][]*domain.CloneNode{{plain}, {plain}}) {
		t.Error("plain .go Field roots must not match (GoSnippet gate)")
	}

	if isSignatureFragmentMatch([][]*domain.CloneNode{{snippet}, {plain}}) {
		t.Error("mixed snippet/plain group must not match (every-instance rule)")
	}

	if !isSignatureFragmentMatch([][]*domain.CloneNode{{snippet}, {snippet}}) {
		t.Error("snippet Field roots must match")
	}
}

// TestIsSignatureFragmentMatch_Disabled guards --disable-pattern semantics:
// with the label disabled, the group is not suppressed through this pattern.
func TestIsSignatureFragmentMatch_Disabled(t *testing.T) {
	t.Parallel()

	fields := signatureFragmentFields(t)
	if len(fields) < 2 {
		t.Fatalf("expected ≥2 Go-snippet `lang` Fields, got %d", len(fields))
	}

	seqs := make([][]*domain.CloneNode, len(fields))
	for i, field := range fields {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(field)}
	}

	label, _ := EvaluateActionabilityWithLabelAndDisabled(
		seqs,
		map[PatternLabel]bool{PatternSignatureFragment: true},
	)

	if label == PatternSignatureFragment {
		t.Error("label = signature-fragment despite --disable-pattern signature-fragment")
	}
}
