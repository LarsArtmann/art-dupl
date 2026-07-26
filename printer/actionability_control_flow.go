package printer

import (
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// RAII cleanup method names that indicate non-actionable patterns.
const cleanupMethodName = "Unlock"

// acquireMethodNames are method names that pair with defer cleanup calls in
// the Lock + Defer Unlock idiom.
var acquireMethodNames = []string{ //nolint:gochecknoglobals // static name set
	"Lock", "RLock", "Acquire", "Reserve", "Obtain", "Claim", "Take", "Begin",
}

// isPureDeferPattern reports whether every clone is a DeferStmt
// wrapping a RAII-style call (Unlock, Close, etc.).
//
// Also matches the 2-statement Lock/Acquire + Defer Unlock/Release pattern:
//
//	m.Lock()
//	defer m.Unlock()
//
// This is idiomatic Go resource management — extracting it into a helper would
// break the defer scope semantics.
func isPureDeferPattern(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) == 1 && seq[0].BaseType == golang.DeferStmt {
			return isRAIIDeferCall(seq[0])
		}

		if len(seq) == 2 &&
			isAcquireCall(seq[0]) &&
			seq[1].BaseType == golang.DeferStmt {
			return isRAIIDeferCall(seq[1])
		}

		return false
	})
}

// isRAIIDeferCall checks if a DeferStmt wraps a known RAII cleanup method.
func isRAIIDeferCall(node *domain.CloneNode) bool {
	for _, child := range node.Children {
		if child.BaseType == golang.CallExpr {
			for _, arg := range child.Children {
				if arg.BaseType == golang.SelectorExpr && isCleanupMethod(arg.Name) {
					return true
				}
			}
		}
	}

	return false
}

// isCleanupMethod reports whether a method name is a known RAII cleanup.
func isCleanupMethod(name string) bool {
	switch name {
	case cleanupMethodName, "RUnlock",
		"Close", "Done", "Cancel", "Release", "Finish", "Disconnect", "Free",
		"Stop", "Shutdown", "Cleanup", "Reset", "Put", "Drop", "Abort", "Teardown":
		return true
	default:
		return false
	}
}

// isAcquireCall reports whether a statement is a call to a known acquire
// method (Lock, RLock, Acquire, etc.). This identifies the first half of the
// Lock + Defer Unlock idiom.
func isAcquireCall(node *domain.CloneNode) bool {
	if node.BaseType != golang.ExprStmt {
		return false
	}

	for _, child := range node.Children {
		if child.BaseType == golang.CallExpr {
			for _, callChild := range child.Children {
				if callChild.BaseType == golang.SelectorExpr &&
					slices.Contains(acquireMethodNames, callChild.Name) {
					return true
				}
			}
		}
	}

	return false
}

// isPureErrorPropagation reports whether every clone is an IfStmt
// that only contains error propagation: if err != nil { return err }.
func isPureErrorPropagation(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.IfStmt {
			return false
		}

		return isErrorOnlyIf(root)
	})
}

// isErrorOnlyIf checks if an IfStmt is a pure error propagation pattern.
// It must have:
//   - A condition containing a comparison against nil (BinaryExpr with nil)
//   - A body containing only a ReturnStmt (or CallExpr wrapping error)
//   - No Else branch
func isErrorOnlyIf(node *domain.CloneNode) bool {
	var (
		hasNilCompare bool
		hasReturnErr  bool
		hasElse       bool
	)

	for _, child := range node.Children {
		switch child.BaseType {
		case golang.BinaryExpr:
			if containsNilIdentifier(child) {
				hasNilCompare = true
			}
		case golang.BlockStmt:
			if isReturnOrWrappedReturn(child) {
				hasReturnErr = true
			}
		case golang.IfStmt:
			return false
		default:
			if child.BaseType != golang.AssignStmt &&
				child.BaseType != golang.DeclStmt {
				hasElse = true
			}
		}
	}

	return hasNilCompare && hasReturnErr && !hasElse
}

// containsNilIdentifier checks if a BinaryExpr compares against nil.
// Uses the Name field to verify an identifier named "nil" is present,
// rather than just checking for any Ident node.
func containsNilIdentifier(node *domain.CloneNode) bool {
	for _, child := range node.Children {
		if child.BaseType == golang.Ident && child.Name == "nil" {
			return true
		}
	}

	return false
}

// isReturnOrWrappedReturn checks if a BlockStmt contains only error handling:
// either a single return, or a log/print call followed by a return.
func isReturnOrWrappedReturn(node *domain.CloneNode) bool {
	if len(node.Children) == 0 {
		return false
	}

	// Single return or return with CallExpr (e.g., return fmt.Errorf("...")).
	if len(node.Children) == 1 {
		bt := node.Children[0].BaseType

		return bt == golang.ReturnStmt || bt == golang.CallExpr
	}

	// 2-statement pattern: log/print error, then return.
	// e.g., log.Print(err); return err
	if len(node.Children) == 2 {
		return isLogOrPrintStmt(node.Children[0]) && node.Children[1].BaseType == golang.ReturnStmt
	}

	return false
}

// isLogOrPrintStmt checks if a statement is a logging or print call
// (log.Print/Errorf/Warnf, fmt.Println/Printf, slog.Error/Warn/Info).
func isLogOrPrintStmt(node *domain.CloneNode) bool {
	if node.BaseType != golang.ExprStmt {
		return false
	}

	for _, child := range node.Children {
		if child.BaseType == golang.CallExpr {
			for _, callChild := range child.Children {
				if callChild.BaseType == golang.SelectorExpr && isLoggingMethod(callChild.Name) {
					return true
				}
			}
		}
	}

	return false
}

// isLoggingMethod reports whether a method name is a known logging/print function.
func isLoggingMethod(name string) bool {
	switch name {
	case "Print", "Printf", "Println",
		"Error", calleeErrorf, "Warn", "Warnf", "Info", "Infof", "Debug", "Debugf",
		"Fatal", "Fatalf", "Panic", "Panicf":
		return true
	default:
		return false
	}
}

// isGuardClause reports whether every clone is a single IfStmt used as a guard
// clause: a condition check followed by an immediate return, with no else
// branch and a body containing only return statements (max 2).
//
// Guard clauses are the most common control-flow idiom in Go after error
// checks. Examples:
//
//	if !enabled { return }
//	if ctx.Err() != nil { return ctx.Err() }
//	if len(items) == 0 { return nil, ErrEmpty }
//
// They are not actionable duplication — extracting a guard clause into a helper
// would require passing in the return type, obscuring the control flow.
//
// Error-specific guards (if err != nil { return err }) are caught earlier by
// isPureErrorPropagation. This pattern catches the remaining boolean/nil/value
// guards.
func isGuardClause(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.IfStmt {
			return false
		}

		return isGuardClauseBody(root)
	})
}

// isGuardClauseBody checks if an IfStmt has the guard-clause shape:
//   - A BlockStmt body containing only ReturnStmt(s), max 2
//   - No else branch (no IfStmt or extra BlockStmt sibling)
func isGuardClauseBody(node *domain.CloneNode) bool {
	var (
		hasBody bool
		hasElse bool
	)

	for _, child := range node.Children {
		switch child.BaseType {
		case golang.BlockStmt:
			if !isReturnOnlyBody(child) {
				return false
			}

			hasBody = true

		case golang.IfStmt:
			// else-if branch → not a simple guard
			return false

		case golang.BinaryExpr, golang.UnaryExpr, golang.CallExpr,
			golang.Ident, golang.SelectorExpr, golang.BasicLit,
			golang.ParenExpr, golang.IndexExpr:
			// Condition expression nodes — expected, skip

		default:
			// Any other child type (AssignStmt, DeclStmt, etc.) suggests
			// this is an init-statement or complex if, not a simple guard.
			if child.BaseType == golang.AssignStmt || child.BaseType == golang.DeclStmt {
				// Init statement (if x := f(); x != nil) — still could be a guard
				continue
			}

			hasElse = true
		}
	}

	return hasBody && !hasElse
}

// isReturnOnlyBody checks if a BlockStmt body contains only ReturnStmt(s),
// with at most 2 statements.
func isReturnOnlyBody(block *domain.CloneNode) bool {
	if len(block.Children) == 0 || len(block.Children) > 2 {
		return false
	}

	for _, stmt := range block.Children {
		if stmt.BaseType != golang.ReturnStmt {
			return false
		}
	}

	return true
}
