package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// RAII cleanup method names that indicate non-actionable patterns.
const cleanupMethodName = "Unlock"

// acquireMethodNames are method calls that typically pair with a defer cleanup.
// A sequence like `m.Lock(); defer m.Unlock()` is idiomatic Go that cannot be
// usefully extracted — the defer must remain in the caller's scope.
var acquireMethodNames = map[string]bool{
	"Lock":    true,
	"RLock":   true,
	"Acquire": true,
	"Reserve": true,
	"Obtain":  true,
	"Claim":   true,
	"Take":    true,
	"Begin":   true,
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
					acquireMethodNames[callChild.Name] {
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

// isReturnOrWrappedReturn checks if a BlockStmt only contains a ReturnStmt.
func isReturnOrWrappedReturn(node *domain.CloneNode) bool {
	if len(node.Children) == 0 {
		return false
	}

	// Allow single return or return with CallExpr (e.g., return fmt.Errorf("...")).
	if len(node.Children) == 1 {
		bt := node.Children[0].BaseType

		return bt == golang.ReturnStmt || bt == golang.CallExpr
	}

	return false
}
