package actionability

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isErrorGuardFallthrough reports whether every clone is an error guard
// (an if err != nil block that returns or reports-and-returns) followed only
// by trivial fallthrough statements (assignments and returns). Two real-world
// shapes fall under this pattern (DiscordSync feedback, 2026-07-27):
//
//  1. HTTP handler guard: the void signature of http.HandlerFunc forces
//     `if err != nil { writeError(w, r, err, ""); return }` — extraction is
//     impossible without a callback wrapper that still needs the bare return.
//  2. Wrapper-call guard with trailing return:
//     `if err != nil { return nil, queryError(err, "unique op") }; return x, nil`
//     — the wrapper call IS the already-extracted helper; the unique strings
//     are parameters, not duplication.
//
// Multi-statement clones like these escape the single-statement error
// patterns (error-propagation, error-wrapping) because those require
// len(seq) == 1.
func isErrorGuardFallthrough(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) < 2 {
			return false
		}

		guard := seq[0]
		if guard.BaseType != golang.IfStmt {
			return false
		}

		return isErrorGuard(guard) && isTrivialFallthroughTail(seq[1:])
	})
}

// isErrorGuard reports whether an IfStmt is an error guard: the condition
// compares an identifier against nil, and the body is only error handling
// (a return, or a report call followed by a return).
func isErrorGuard(node *domain.CloneNode) bool {
	if errorVarNameFromComparison(node) == "" {
		return false
	}

	hasNilCompare := false
	hasBody := false

	for _, child := range node.Children {
		switch child.BaseType {
		case golang.BinaryExpr:
			if containsNilIdentifier(child) {
				hasNilCompare = true
			}
		case golang.BlockStmt:
			if isReturnOrWrappedReturn(child) {
				hasBody = true
			}
		case golang.IfStmt:
			return false
		}
	}

	return hasNilCompare && hasBody
}

// isTrivialFallthroughTail reports whether every trailing statement is a
// simple assignment or return. Anything richer (calls, branches, loops)
// carries extractable logic and must not be suppressed by this pattern.
func isTrivialFallthroughTail(seq []*domain.CloneNode) bool {
	for _, node := range seq {
		if node.BaseType != golang.AssignStmt && node.BaseType != golang.ReturnStmt {
			return false
		}
	}

	return true
}
