package actionability

import (
	"fmt"
	"io"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// actionabilityPatternTable defines the priority-ordered list of patterns
// checked by evaluateActionabilityDetailed. First match wins.
type patternEntry struct {
	check   func([][]*domain.CloneNode) bool
	pattern PatternLabel
}

var actionabilityPatternTable = []patternEntry{ //nolint:gochecknoglobals // static table
	{isSignatureOnlyMatch, PatternSignatureOnly},
	{isInterfaceImplementation, PatternInterfaceImpl},
	{isInterfaceMethodBody, PatternInterfaceMethod},
	{isPureDeferPattern, PatternRAIIDefer},
	{isPureErrorPropagation, PatternErrorPropagation},
	{isGuardClause, PatternGuardClause},
	{isAssignWithErrorCheck, PatternAssignErrorCheck},
	{isAssignWithBoolGuard, PatternBoolGuard},
	{isSingleCallExpression, PatternSingleCallExpr},
	{isSingleSimpleStatement, PatternSingleSimpleStmt},
	{isSingleDeclaration, PatternSingleDeclaration},
	{isTestHelperDelegate, PatternTestHelperDelegate},
	{isErrorWrappingReturn, PatternErrorWrapping},
	{isAssertionChain, PatternAssertionChain},
	{isCobraCommandBoilerplate, PatternCobraBoilerplate},
	{isTestDataFilePair, PatternTestData},
	{isTableDrivenTestBody, PatternTableDrivenTest},
	{isTestScaffolding, PatternTestScaffolding},
	{isDataDominated, PatternDataDominated},
	{isDescribeTablePattern, PatternDescribeTable},
	{isBuilderCallbackPattern, PatternBuilderCallback},
}

// AllActionabilityPatterns returns all registered pattern labels in priority order.
func AllActionabilityPatterns() []PatternLabel {
	labels := make([]PatternLabel, 0, len(actionabilityPatternTable))
	for _, entry := range actionabilityPatternTable {
		labels = append(labels, entry.pattern)
	}

	return labels
}

// ListActionabilityPatterns writes all pattern labels to the writer, one per line.
func ListActionabilityPatterns(w io.Writer) {
	for _, label := range AllActionabilityPatterns() {
		if _, err := fmt.Fprintln(w, label); err != nil {
			return
		}
	}
}

// EvaluateActionability analyzes a clone group and determines whether it
// represents actionable duplication or idiomatic boilerplate noise.
//
// everySequenceMatch returns true if pred returns true for every sequence.
// Used by actionability checks that require ALL clones to match a pattern.
func everySequenceMatch(
	nodeSeqs [][]*domain.CloneNode,
	pred func(seq []*domain.CloneNode) bool,
) bool {
	for _, seq := range nodeSeqs {
		if !pred(seq) {
			return false
		}
	}

	return true
}

// unwrapExprStmt returns the inner CallExpr if n is an ExprStmt wrapping
// exactly one CallExpr. Otherwise returns n unchanged. This handles the Go
// AST convention where standalone function calls as statements are wrapped
// in ExprStmt nodes. Patterns that match CallExpr nodes must unwrap first
// to avoid missing statement-position calls.
func unwrapExprStmt(n *domain.CloneNode) *domain.CloneNode {
	if n.BaseType == golang.ExprStmt && len(n.Children) == 1 &&
		n.Children[0].BaseType == golang.CallExpr {
		return n.Children[0]
	}

	return n
}

// EvaluateActionability analyzes a clone group and determines whether it
// represents actionable duplication or idiomatic boilerplate noise.
//
// A group is actionable when it contains real logic that can be extracted,
// composed, or otherwise refactored. Non-actionable patterns are standard
// Go idioms that cannot be eliminated without breaking semantics.
//
// To be non-actionable, EVERY clone in the group must match the same
// boilerplate pattern. If any clone differs, the group is actionable.
func EvaluateActionability(nodeSeqs [][]*domain.CloneNode) domain.CloneActionability {
	_, a := evaluateActionabilityDetailed(nodeSeqs)

	return a
}

// EvaluateActionabilityWithDisabled is like EvaluateActionability but skips
// patterns whose labels are in the disabled set.
func EvaluateActionabilityWithDisabled(
	nodeSeqs [][]*domain.CloneNode,
	disabled map[PatternLabel]bool,
) domain.CloneActionability {
	_, a := evaluateActionabilityWithDisabled(nodeSeqs, disabled)

	return a
}

// PatternLabel identifies which non-actionable pattern was detected.
// Empty string means actionable.
type PatternLabel string

const (
	PatternNone               PatternLabel = ""
	PatternTestData           PatternLabel = "testdata-pair"
	PatternTableDrivenTest    PatternLabel = "table-driven-test"
	PatternTestScaffolding    PatternLabel = "test-scaffolding"
	PatternDataDominated      PatternLabel = "data-dominated"
	PatternSignatureOnly      PatternLabel = "signature-only"
	PatternRAIIDefer          PatternLabel = "raii-defer"
	PatternErrorPropagation   PatternLabel = "error-propagation"
	PatternErrorWrapping      PatternLabel = "error-wrapping"
	PatternAssertionChain     PatternLabel = "assertion-chain"
	PatternCobraBoilerplate   PatternLabel = "cobra-boilerplate"
	PatternInterfaceImpl      PatternLabel = "interface-implementation"
	PatternDescribeTable      PatternLabel = "describe-table"
	PatternBuilderCallback    PatternLabel = "builder-callback"
	PatternAssignErrorCheck   PatternLabel = "assign-error-check"
	PatternSingleCallExpr     PatternLabel = "single-call-expression"
	PatternSingleSimpleStmt   PatternLabel = "single-simple-statement"
	PatternSingleDeclaration  PatternLabel = "single-declaration"
	PatternGuardClause        PatternLabel = "guard-clause"
	PatternTestHelperDelegate    PatternLabel = "test-helper-delegate"
	PatternInterfaceMethod       PatternLabel = "interface-method"
	PatternBoolGuard             PatternLabel = "bool-guard"
	PatternPropertyEngine        PatternLabel = "property-engine"
	PatternPropertyControlFlow   PatternLabel = "property-control-flow"
	PatternPropertyROI           PatternLabel = "property-roi"
	PatternPropertyParameterizable PatternLabel = "property-parameterizable"
)

// EvaluateActionabilityWithLabel returns both the actionability and the
// pattern label that caused it. This allows downstream code to adjust
// category/suggestion based on which specific pattern was detected.
func EvaluateActionabilityWithLabel(nodeSeqs [][]*domain.CloneNode) (PatternLabel, domain.CloneActionability) {
	return evaluateActionabilityDetailed(nodeSeqs)
}

func evaluateActionabilityDetailed(nodeSeqs [][]*domain.CloneNode) (PatternLabel, domain.CloneActionability) {
	return evaluateActionabilityWithDisabled(nodeSeqs, nil)
}

// evaluateActionabilityWithDisabled checks actionability patterns in priority
// order, skipping any pattern whose label is in the disabled set.
//
// The property-based extractability engine runs FIRST as a pre-filter. If it
// determines the clone is not harmful (any property fails), the clone is
// NonActionable with a property-specific reason. The denylist patterns then
// run as a fallback for cases the property engine doesn't cover.
func evaluateActionabilityWithDisabled(
	nodeSeqs [][]*domain.CloneNode,
	disabled map[PatternLabel]bool,
) (PatternLabel, domain.CloneActionability) {
	if len(nodeSeqs) == 0 {
		return PatternNone, domain.Actionable
	}

	// Pattern table: specific, well-tested denylist patterns run first.
	// This preserves existing behavior and labels.
	for _, p := range actionabilityPatternTable {
		if disabled != nil && disabled[p.pattern] {
			continue
		}

		if p.check(nodeSeqs) {
			return p.pattern, domain.NonActionable
		}
	}

	// Property-based fallback: the extractability engine runs as a second-pass
	// filter for clones the denylist doesn't cover. When type info is available
	// (EnclosingReturnArity > 0 or VarType populated), the engine can suppress
	// clones that are structurally non-harmful (void-function traps, etc.).
	analysis := EvaluateExtractability(nodeSeqs)
	if !domain.IsHarmful(analysis) {
		label := propertyLabelForReason(analysis.Reason)
		if disabled == nil || !disabled[label] {
			return label, domain.NonActionable
		}
	}

	return PatternNone, domain.Actionable
}

// propertyLabelForReason converts an extractability analysis reason to a
// PatternLabel for consistent display in --explain output.
func propertyLabelForReason(reason string) PatternLabel {
	switch {
	case strings.Contains(reason, "void function") || strings.Contains(reason, "break/continue"):
		return PatternPropertyControlFlow
	case strings.Contains(reason, "too small") || strings.Contains(reason, "single call expression"):
		return PatternPropertyROI
	case strings.Contains(reason, "string-literal values"):
		return PatternPropertyParameterizable
	default:
		return PatternPropertyEngine
	}
}

// isSignatureOnlyMatch reports whether every clone is a single FuncDecl node
// WITHOUT a substantial body (>3 child nodes). Interface method signatures,
// empty stubs, and forwarding methods fit this pattern.
//
// A FuncDecl with a real body (BlockStmt with >0 children) is actionable —
// identical bodies can be extracted to shared functions. But if two FuncDecls
// match and the body is empty (or has only boilerplate), deduplication is
// impossible without changing the interface.
func isSignatureOnlyMatch(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.FuncDecl {
			return false
		}

		return !hasRealBody(root)
	})
}

// hasRealBody checks if a FuncDecl contains a body with meaningful logic
// beyond the signature itself.
func hasRealBody(node *domain.CloneNode) bool {
	for _, child := range node.Children {
		if child.BaseType == golang.BlockStmt && len(child.Children) > 0 {
			return true
		}
	}

	return false
}

// isInterfaceImplementation detects when all clone fragments are implementations
// of the same interface method. The signal is: multiple fragments from different
// files that all start with a FuncType node, indicating identical method signatures.
//
// In Go, when 3+ files contain a method with the same FuncType signature and
// matching body prefix, they almost certainly satisfy a common interface. The
// suffix tree already guarantees the token sequences match, so we only need to
// verify the structural pattern and file diversity.
//
// This works because:
//   - The suffix tree matches on Type values, so identical FuncType subtrees
//     (same parameter types, same return types) produce the same token sequence.
//   - 3+ different files with matching FuncType + body prefix is almost never
//     coincidence — it means they all satisfy the same interface contract.
//   - Semantic mode makes this even more precise by encoding identifier/operator
//     names into the Type field.
func isInterfaceImplementation(nodeSeqs [][]*domain.CloneNode) bool {
	if len(nodeSeqs) < 3 {
		return false
	}

	files := make(map[string]bool)

	for _, seq := range nodeSeqs {
		if len(seq) == 0 {
			return false
		}

		root := seq[0]
		// The fragment must be rooted at a FuncType — this is the signature
		// node shared across interface implementations.
		if root.BaseType != golang.FuncType {
			return false
		}

		if root.Filename != "" {
			files[root.Filename] = true
		}
	}

	return len(files) >= 3
}
