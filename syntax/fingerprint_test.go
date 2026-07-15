package syntax

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
)

// Node type constants mirrored from syntax/golang/nodetypes.go to avoid
// an import cycle (syntax/golang imports syntax).
const (
	testIfStmt     int32 = 31
	testBinaryExpr int32 = 9
	testBlockStmt  int32 = 6
	testReturnStmt int32 = 41
	testFuncDecl   int32 = 24
)

// TestFingerprintPreservesBaseType verifies that serial() stores the composite
// fingerprint in Fingerprint (not Type), so the original base AST type is
// preserved for actionability analysis.
func TestFingerprintPreservesBaseType(t *testing.T) {
	t.Parallel()

	// Build a simple tree: FuncDecl → BlockStmt → [IfStmt(Statement), ReturnStmt(Statement)]
	ifStmt := &Node{
		Type:      testIfStmt,
		Statement: true,
		Children: []*Node{
			{Type: testBinaryExpr, Children: []*Node{
				{Type: 30, Name: "err"}, // Ident
				{Type: 30, Name: "nil"}, // Ident
			}},
			{Type: testBlockStmt, Children: []*Node{
				{Type: testReturnStmt},
			}},
		},
	}
	returnStmt := &Node{
		Type:      testReturnStmt,
		Statement: true,
	}
	block := &Node{
		Type:     testBlockStmt,
		Children: []*Node{ifStmt, returnStmt},
	}
	root := &Node{
		Type:     testFuncDecl,
		Children: []*Node{block},
	}

	stream := Serialize(root)

	// Verify that statement nodes have their original Type preserved
	for _, n := range stream {
		if n.Statement {
			// Type should still be the original base type, NOT the fingerprint
			// (DecodeBaseType extracts lower 8 bits, which should give back the original)
			baseType := n.Type & 0xFF
			if baseType == 0 {
				t.Errorf("Statement node has corrupted base type: Type=%d", n.Type)
			}

			if n.Fingerprint == 0 {
				t.Errorf("Statement node has zero Fingerprint: Type=%d", n.Type)
			}

			if n.Val() != suffixtree.TokenValue(n.Fingerprint) {
				t.Errorf("Val() for statement should return Fingerprint (%d), got %d",
					n.Fingerprint, n.Val())
			}
		}
	}
}

// TestSerializeIdempotency verifies that serializing a tree twice produces
// the same result (no mutation of the original tree).
func TestSerializeIdempotency(t *testing.T) {
	t.Parallel()

	originalType := testIfStmt
	ifStmt := &Node{
		Type:      originalType,
		Statement: true,
		Children: []*Node{
			{Type: testBinaryExpr},
			{Type: testBlockStmt},
		},
	}
	root := &Node{
		Type:     testFuncDecl,
		Children: []*Node{{Type: testBlockStmt, Children: []*Node{ifStmt}}},
	}

	stream1 := Serialize(root)
	stream2 := Serialize(root)

	if len(stream1) != len(stream2) {
		t.Fatalf("Stream lengths differ: %d vs %d", len(stream1), len(stream2))
	}

	for i := range stream1 {
		if stream1[i].Type != stream2[i].Type {
			t.Errorf("Type mismatch at %d: %d vs %d", i, stream1[i].Type, stream2[i].Type)
		}
		if stream1[i].Fingerprint != stream2[i].Fingerprint {
			t.Errorf("Fingerprint mismatch at %d: %d vs %d", i,
				stream1[i].Fingerprint, stream2[i].Fingerprint)
		}
	}

	if ifStmt.Type != originalType {
		t.Errorf("Original IfStmt Type mutated: expected %d, got %d", originalType, ifStmt.Type)
	}
}

// TestNonStatementNodesUseTypeForVal verifies that non-statement nodes
// return their Type from Val(), not Fingerprint.
func TestNonStatementNodesUseTypeForVal(t *testing.T) {
	t.Parallel()

	node := &Node{
		Type:      testBlockStmt,
		Statement: false,
	}

	if node.Val() != suffixtree.TokenValue(node.Type) {
		t.Errorf("Non-statement Val() should return Type (%d), got %d",
			node.Type, node.Val())
	}
}

// TestClonePreservesFingerprint verifies that Node.Clone() copies Fingerprint.
func TestClonePreservesFingerprint(t *testing.T) {
	t.Parallel()

	original := &Node{
		Type:        testIfStmt,
		Statement:   true,
		Fingerprint: -123456,
	}

	clone := original.Clone()

	if clone.Fingerprint != original.Fingerprint {
		t.Errorf("Clone Fingerprint mismatch: %d vs %d",
			clone.Fingerprint, original.Fingerprint)
	}

	if clone.Type != original.Type {
		t.Errorf("Clone Type mismatch: %d vs %d", clone.Type, original.Type)
	}

	if clone.Statement != original.Statement {
		t.Errorf("Clone Statement mismatch: %v vs %v", clone.Statement, original.Statement)
	}
}
