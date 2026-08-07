package actionability

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// pipelineResult holds the output of a full-pipeline run for actionability
// assertion. Each entry is one clone group's CloneNode sequences.
type pipelineResult struct {
	groups [][][]*domain.CloneNode
}

// runPipeline parses Go source files, builds a suffix tree, finds duplicate
// sequences, and converts them to CloneNode sequences ready for actionability
// evaluation. This replicates the production pipeline: parse → serialize →
// suffix tree → FindSyntaxUnits → CloneNode conversion.
//
// The cross-layer bugs fixed in this session (serial() dropping IsAlias,
// BasicLit Name not populated) were invisible to unit tests that construct
// CloneNodes directly. This function exercises the exact code path where
// those bugs occurred.
func runPipeline(t *testing.T, files map[string]string, threshold int) *pipelineResult {
	t.Helper()

	tree := suffixtree.New()

	var data []*syntax.Node

	sentinelFP := int32(math.MinInt32 / 2)
	tmpDir := t.TempDir()

	for name, src := range files {
		path := filepath.Join(tmpDir, name)

		if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}

		root, err := golang.Parse(path)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}

		stream := syntax.Serialize(root)

		for _, node := range stream {
			data = append(data, node)

			if err := tree.Update(node); err != nil {
				t.Fatalf("suffix tree update: %v", err)
			}
		}

		sentinel := &syntax.Node{
			Type:        -1, // BadNodeType — never matches real nodes
			Statement:   true,
			Fingerprint: sentinelFP,
		}
		sentinelFP++

		data = append(data, sentinel)

		if err := tree.Update(sentinel); err != nil {
			t.Fatalf("suffix tree update sentinel: %v", err)
		}
	}

	ctx := context.Background()
	result := &pipelineResult{}

	for m := range tree.FindDuplOver(ctx, threshold) {
		sm := syntax.FindSyntaxUnits(data, m, threshold)
		if len(sm.Frags) < 2 {
			continue
		}

		nodeSeqs := make([][]*domain.CloneNode, len(sm.Frags))
		for i, frag := range sm.Frags {
			nodeSeqs[i] = make([]*domain.CloneNode, len(frag))
			for j, n := range frag {
				nodeSeqs[i][j] = cloneNodeFromSyntax(n)
			}
		}

		result.groups = append(result.groups, nodeSeqs)
	}

	return result
}

// TestPipeline_TypeAliasReExportNotActionable is the definitive cross-layer
// regression test for the serial() IsAlias bug (commit 13e3dcac).
//
// It writes two Go files containing identical type alias declarations,
// parses them through the FULL pipeline (parse → serialize → suffix tree →
// FindSyntaxUnits → CloneNode conversion → actionability), and verifies that:
//
// 1. The suffix tree finds a match (the aliases are structurally identical)
// 2. The matched CloneNode has IsAlias=true (proving serial() preserved it)
// 3. The actionability verdict is NonActionable
//
// Before the fix, serial() did not copy IsAlias into the shallow copy, so the
// token stream always had IsAlias=false, causing isAliasTypeSpec() to fail
// and the clone to be reported as actionable.
func TestPipeline_TypeAliasReExportNotActionable(t *testing.T) {
	t.Parallel()

	aliasSrc := `package main

type ErrorType = string
`

	result := runPipeline(t, map[string]string{
		"aliases1.go": aliasSrc,
		"aliases2.go": aliasSrc,
	}, 1)

	if len(result.groups) == 0 {
		t.Fatal("expected at least one clone group from two identical type alias files")
	}

	foundAliasMatch := false

	for _, group := range result.groups {
		for _, seq := range group {
			for _, node := range seq {
				if node.BaseType == golang.TypeSpec && node.IsAlias {
					foundAliasMatch = true

					_, verdict := EvaluateActionabilityWithLabel(group)
					if verdict != domain.NonActionable {
						t.Errorf(
							"alias TypeSpec classified as %s, want NonActionable",
							verdict,
						)
					}
				}
			}
		}
	}

	if !foundAliasMatch {
		t.Error("no TypeSpec with IsAlias=true found in any clone group — serial() may be dropping IsAlias again")
	}
}

// TestPipeline_BasicLitNamePopulated is the definitive cross-layer regression
// test for the BasicLit Name bug (commit 71e12c96).
//
// It writes two identical Go files containing string literals in control-flow
// statements, runs them through the full pipeline, and verifies that the
// CloneNode tree contains BasicLit-derived nodes with non-empty Name
// (proving transform.go populates BasicLit Name AND serial preserves it).
//
// Before the fix, BasicLit nodes never had their Value stored in Name, so
// collectStringLiterals always returned empty slices, and the parameterizability
// engine was completely non-functional.
func TestPipeline_BasicLitNamePopulated(t *testing.T) {
	t.Parallel()

	funcSrc := `package main

import "fmt"

func helper(val int) error {
	if val == 0 {
		return fmt.Errorf("value is zero")
	}
	if val < 0 {
		return fmt.Errorf("value is negative")
	}
	return nil
}
`

	result := runPipeline(t, map[string]string{
		"helper1.go": funcSrc,
		"helper2.go": funcSrc,
	}, 1)

	if len(result.groups) == 0 {
		t.Fatal("expected at least one clone group from two identical function files")
	}

	foundBasicLitName := false

	for _, group := range result.groups {
		for _, seq := range group {
			for _, node := range seq {
				if hasBasicLitWithName(node) {
					foundBasicLitName = true
				}
			}
		}
	}

	if !foundBasicLitName {
		t.Error(
			"no BasicLit node with non-empty Name found — " +
				"transform.go may not be populating BasicLit Name, " +
				"or serial() may be dropping it",
		)
	}
}

// hasBasicLitWithName recursively checks whether any node in the subtree is a
// BasicLit with a non-empty Name field. This proves the BasicLit Name fix
// (transform.go:43) survived serialization (serial shallow copy).
func hasBasicLitWithName(n *domain.CloneNode) bool {
	if n == nil {
		return false
	}

	if n.BaseType == golang.BasicLit && n.Name != "" {
		return true
	}

	return slices.ContainsFunc(n.Children, hasBasicLitWithName)
}
