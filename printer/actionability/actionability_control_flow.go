package actionability

import (
	"slices"
	"strings"

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
// Handles three forms:
//  1. Direct method call: defer x.Close()
//  2. Direct function call: defer cancel() (bare Ident callee)
//  3. FuncLit wrapping: defer func() { _ = rows.Close() }()
func isRAIIDeferCall(node *domain.CloneNode) bool {
	for _, child := range node.Children {
		if child.BaseType == golang.CallExpr {
			if callIsRAIICleanup(child) {
				return true
			}

			if hasCleanupInFuncLit(child) {
				return true
			}
		}
	}

	return false
}

// callIsRAIICleanup checks if a CallExpr is a direct cleanup call:// either x.Close() (SelectorExpr) or cancel() (bare Ident).
func callIsRAIICleanup(call *domain.CloneNode) bool {
	for _, child := range call.Children {
		if child.BaseType == golang.SelectorExpr && slices.Contains(cleanupMethodNames, child.Name) {
			return true
		}

		if child.BaseType == golang.Ident && isCleanupIdentName(child.Name) {
			return true
		}
	}

	return false
}

// isCleanupIdentName checks if a bare ident name (like "cancel") matches
// a cleanup method name case-insensitively. This catches context.CancelFunc
// variables named cancel, done, etc.
func isCleanupIdentName(name string) bool {
	lname := strings.ToLower(name)
	for _, cn := range cleanupMethodNames {
		if strings.ToLower(cn) == lname {
			return true
		}
	}

	return false
}

// hasCleanupInFuncLit checks if a CallExpr wraps a FuncLit whose body
// contains a cleanup call. Handles: defer func() { _ = rows.Close() }().
func hasCleanupInFuncLit(call *domain.CloneNode) bool {
	for _, child := range call.Children {
		if child.BaseType == golang.FuncLit {
			if subtreeHasCleanupCall(child) {
				return true
			}
		}
	}

	return false
}

// subtreeHasCleanupCall recursively searches a subtree for any cleanup CallExpr.
func subtreeHasCleanupCall(node *domain.CloneNode) bool {
	if node.BaseType == golang.CallExpr && callIsRAIICleanup(node) {
		return true
	}

	return slices.ContainsFunc(node.Children, subtreeHasCleanupCall)
}

// cleanupMethodNames are method names that release resources in RAII-style
// defer patterns (defer x.Unlock(), defer x.Close(), etc.).
var cleanupMethodNames = []string{ //nolint:gochecknoglobals // static name set
	cleanupMethodName, "RUnlock",
	"Close", "Done", "Cancel", "Release", "Finish", "Disconnect", "Free",
	"Stop", "Shutdown", "Cleanup", "Reset", "Put", "Drop", "Abort", "Teardown",
	"Rollback",
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

	// 2-statement pattern: any call + return.
	// e.g., writeError(w, r, err, ""); return
	// e.g., log.Print(err); return err
	// e.g., slog.Error("msg", err); return
	// In the context of if err != nil, any call followed by a bare return
	// is the error-handling idiom (HTTP handler guard, logging+return, etc.).
	if len(node.Children) == 2 {
		return isExprStmtCallExpr(node.Children[0]) && node.Children[1].BaseType == golang.ReturnStmt
	}

	return false
}

// isExprStmtCallExpr checks if a node is an ExprStmt wrapping a CallExpr.
// This accepts ANY function call as a statement (writeError, queryError,
// slog.Error, custom helpers, etc.).
func isExprStmtCallExpr(node *domain.CloneNode) bool {
	return node.BaseType == golang.ExprStmt &&
		len(node.Children) > 0 &&
		node.Children[0].BaseType == golang.CallExpr
}

// loggingMethodNames are method names for logging/print functions.
var loggingMethodNames = []string{ //nolint:gochecknoglobals,goconst // static name set, not a domain constant
	"Print", "Printf", "Println",
	"Error", calleeErrorf, "Warn", "Warnf", "Info", "Infof", "Debug", "Debugf",
	"Fatal", "Fatalf", "Panic", "Panicf",
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
