package actionability

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

const (
	boolLiteralFalse = "false"
	boolLiteralTrue  = "true"
)

// isBoolAccumulatorInitializer reports whether every clone is a sequence of two
// or more simple boolean variable initializations such as:
//
//	hasFloatFormat := false
//	hasSeparatorLoop := false
//
// or the equivalent var-declaration form:
//
//	var hasFloatFormat bool = false
//	var hasSeparatorLoop bool = false
//
// These are common in AST-traversal accumulators, feature-flag parsers, and
// other code that independently tracks a few boolean observations. They are
// not duplicated logic: the variables are local to each detector and carry
// unrelated meaning, so extracting them would couple unrelated code without
// saving meaningful lines. Suppressing them at low thresholds avoids the
// noise that comes from matching the `name := false` shape in isolation.
func isBoolAccumulatorInitializer(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) < 2 {
			return false
		}

		for _, node := range seq {
			if !isBoolVariableInitialization(node) {
				return false
			}
		}

		return true
	})
}

// isBoolVariableInitialization reports whether a single node is a boolean
// variable initialization that assigns a boolean literal to one or more
// identifiers, with no other expression structure. Handles all three Go
// syntactic forms:
//
//	hasX := false               (AssignStmt)
//	hasX, hasY := false, true   (AssignStmt, multi-value)
//	var hasX bool = false       (DeclStmt -> GenDecl -> ValueSpec)
//	var hasX = false            (DeclStmt -> GenDecl -> ValueSpec, untyped)
//
// Package-level declarations appear as bare ValueSpec nodes (no DeclStmt
// wrapping), so that form is handled directly.
//
// The check is intentionally structural: it does not try to prove the
// assignment is a *declaration*, because the clone is already matched by the
// suffix tree and the only actionable distinction left is whether the RHS is
// a plain boolean literal. Extracting even a repeated `x = true` pair into a
// helper would be worse than the duplication.
func isBoolVariableInitialization(node *domain.CloneNode) bool {
	idents, ok := boolInitIdents(node)
	if !ok {
		return false
	}

	if len(idents) < 2 {
		return false
	}

	var (
		hasBoolLiteral bool
		hasIdentifier  bool
	)

	for _, child := range idents {
		if child.BaseType != golang.Ident {
			return false
		}

		if isBoolLiteralName(child.Name) {
			hasBoolLiteral = true
		} else {
			hasIdentifier = true
		}
	}

	return hasBoolLiteral && hasIdentifier
}

// boolInitIdents extracts the identifier-level children from a node that
// represents a boolean variable initialization, regardless of the wrapping
// syntactic form. Returns the children and true if the node is a recognized
// bool-initialization form; returns nil and false otherwise.
func boolInitIdents(node *domain.CloneNode) ([]*domain.CloneNode, bool) {
	switch node.BaseType {
	case golang.AssignStmt, golang.ValueSpec:
		return node.Children, true

	case golang.DeclStmt:
		// DeclStmt -> GenDecl -> ValueSpec (local var/const declaration).
		// Only single-spec declarations are handled; grouped declarations
		// (var ( a = false; b = true )) have multiple ValueSpec children
		// and are left to other patterns.
		if len(node.Children) != 1 {
			return nil, false
		}

		genDecl := node.Children[0]
		if genDecl.BaseType != golang.GenDecl || len(genDecl.Children) != 1 {
			return nil, false
		}

		spec := genDecl.Children[0]
		if spec.BaseType != golang.ValueSpec {
			return nil, false
		}

		return spec.Children, true

	default:
		return nil, false
	}
}

// isBoolLiteralName reports whether a name is one of the Go boolean literals.
func isBoolLiteralName(name string) bool {
	return name == boolLiteralTrue || name == boolLiteralFalse
}
