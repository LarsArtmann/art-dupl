package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// EvaluateActionability analyzes a clone group and determines whether it
// represents actionable duplication or idiomatic boilerplate noise.
//
// A group is non-actionable when every clone matches the same simple
// boilerplate pattern:
//
//   - Signature-only FuncDecl: interface method implementations that must
//     have identical signatures. Deduplication is impossible without breaking
//     the interface contract.
//   - Pure DeferStmt: defer mu.Unlock() / defer f.Close() patterns that are
//     standard Go idioms. Extracting them into helpers breaks RAII semantics.
//   - Pure IfStmt: if err != nil { return err } patterns. Go has no mechanism
//     to extract error propagation without worsening the code.
func EvaluateActionability(nodeSeqs [][]*syntax.Node) domain.CloneActionability {
	if len(nodeSeqs) == 0 {
		return domain.Actionable
	}

	if isSignatureOnlyMatch(nodeSeqs) {
		return domain.NonActionable
	}

	if isPureDeferPattern(nodeSeqs) {
		return domain.NonActionable
	}

	if isPureErrorPropagation(nodeSeqs) {
		return domain.NonActionable
	}

	return domain.Actionable
}

// isSignatureOnlyMatch reports whether every clone is a single FuncDecl node.
// When the entire match is just one FuncDecl across multiple files, it is
// almost certainly an interface method signature implementation.
func isSignatureOnlyMatch(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		if seq[0].Type != golang.FuncDecl {
			return false
		}
	}

	return true
}

// isPureDeferPattern reports whether every clone is a DeferStmt.
func isPureDeferPattern(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		if seq[0].Type != golang.DeferStmt {
			return false
		}
	}

	return true
}

// isPureErrorPropagation reports whether every clone is an IfStmt.
func isPureErrorPropagation(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		if seq[0].Type != golang.IfStmt {
			return false
		}
	}

	return true
}
