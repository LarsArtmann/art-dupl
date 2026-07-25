package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isAssignWithErrorCheck reports whether every clone is a 2-statement sequence
// of an assignment followed by an error-check IfStmt. This is the most common
// Go boilerplate idiom:
//
//	err := someFunc()
//	if err != nil {
//	    return ...
//	}
//
// Every Go program contains dozens of these. They are not actionable
// duplication — extracting them would obscure the control flow without
// reducing complexity.
func isAssignWithErrorCheck(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 2 {
			return false
		}

		if seq[0].BaseType != golang.AssignStmt {
			return false
		}

		if seq[1].BaseType != golang.IfStmt {
			return false
		}

		return isErrorOnlyIf(seq[1])
	})
}

// isSingleCallExpression reports whether every clone is exactly one CallExpr
// node, either as a bare CallExpr (non-statement match) or as an ExprStmt
// wrapping a single CallExpr (statement-level match). Single function calls
// with different arguments are not actionable duplication — they are just
// using the same API with different data.
// Examples: errors.New("foo"), fmt.Println("bar"), http.Get(url), t.Parallel().
func isSingleCallExpression(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		return isLoneCallExpr(seq[0])
	})
}

// isLoneCallExpr reports whether a single CloneNode is a bare CallExpr or an
// ExprStmt wrapping exactly one CallExpr. In Go's AST, a standalone function
// call as a statement (e.g. t.Parallel()) is wrapped in an ExprStmt. When
// statement-level tokenization is active, the matched node has BaseType
// ExprStmt, not CallExpr, so we must look inside the ExprStmt to find the call.
func isLoneCallExpr(n *domain.CloneNode) bool {
	if n.BaseType == golang.CallExpr {
		return true
	}

	if n.BaseType == golang.ExprStmt &&
		len(n.Children) == 1 &&
		n.Children[0].BaseType == golang.CallExpr {
		return true
	}

	return false
}
