package actionability

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// EvaluateExtractability runs the 4-property extractability analysis on a clone group.
// Returns the merged analysis from all checkers.
//
// When type info is unavailable (syntax-only mode), properties default to true
// (harmful), preventing false negatives. The denylist patterns still run as fallback.
func EvaluateExtractability(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	if len(nodeSeqs) == 0 {
		return domain.DefaultExtractabilityAnalysis()
	}

	analysis := domain.ExtractabilityAnalysis{
		MechanicallyExtractable: true,
		ControlFlowExtractable:  true,
		ROIPositive:             true,
		Parameterizable:         true,
		Confidence:              0.9,
		Reason:                  "all properties pass",
	}

	// Property 2: Control-flow extractability
	cf := checkControlFlow(nodeSeqs)
	if !cf.ControlFlowExtractable {
		analysis.ControlFlowExtractable = false
		analysis.Confidence = minConfidence(analysis.Confidence, cf.Confidence)
		analysis.Reason = cf.Reason
	}

	// Property 3: ROI + helper-dominance
	roi := checkROI(nodeSeqs)
	if !roi.ROIPositive {
		analysis.ROIPositive = false
		analysis.Confidence = minConfidence(analysis.Confidence, roi.Confidence)
		if analysis.ControlFlowExtractable {
			analysis.Reason = roi.Reason
		}
	}

	// Property 4: Parameterizability
	param := checkParameterizability(nodeSeqs)
	if !param.Parameterizable {
		analysis.Parameterizable = false
		analysis.Confidence = minConfidence(analysis.Confidence, param.Confidence)
		if analysis.ControlFlowExtractable && analysis.ROIPositive {
			analysis.Reason = param.Reason
		}
	}

	return analysis
}

func minConfidence(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}

// --- Property 2: Control-flow extractability ---

// checkControlFlow evaluates whether the clone contains terminating statements
// (return/break/continue) that are forced by the enclosing function's signature.
// When the enclosing function is void (arity 0), bare returns cannot be extracted
// into a helper — the helper cannot issue return on behalf of the caller.
func checkControlFlow(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	for _, seq := range nodeSeqs {
		hasReturn := false
		hasBranch := false
		arity := int32(0)

		for _, node := range seq {
			if node.EnclosingReturnArity > arity {
				arity = node.EnclosingReturnArity
			}

			if subtreeHasNodeType(node, golang.ReturnStmt) {
				hasReturn = true
			}

			if subtreeHasNodeType(node, golang.BranchStmt) {
				hasBranch = true
			}
		}

		// Void function with return statements — forced by signature
		if hasReturn && arity == 0 {
			return domain.ExtractabilityAnalysis{
				ControlFlowExtractable: false,
				Confidence:             0.85,
				Reason:                  "clone contains return in void function — extraction cannot issue return on behalf of caller",
			}
		}

		// break/continue — cannot be extracted (loop control is scope-bound)
		if hasBranch {
			return domain.ExtractabilityAnalysis{
				ControlFlowExtractable: false,
				Confidence:             0.85,
				Reason:                  "clone contains break/continue — extraction cannot preserve loop control flow",
			}
		}
	}

	return domain.ExtractabilityAnalysis{
		ControlFlowExtractable: true,
		Confidence:             0.9,
	}
}

// subtreeHasNodeType recursively checks if any node in the subtree has the given BaseType.
func subtreeHasNodeType(node *domain.CloneNode, nodeType int32) bool {
	if node.BaseType == nodeType {
		return true
	}

	for _, child := range node.Children {
		if subtreeHasNodeType(child, nodeType) {
			return true
		}
	}

	return false
}

// --- Property 3: ROI + helper-dominance ---

// minCloneTokens is the minimum token count for extraction to be worthwhile.
// Single-statement clones are almost always too small to benefit.
const minCloneTokens = 10

// helperDominanceRatio is the threshold for helper-dominance: if >60% of clone
// tokens are inside a single CallExpr, the call IS the extraction.
const helperDominanceRatio = 0.6

// checkROI evaluates whether extracting the clone would save tokens.
// Returns ROIPositive=false when the clone is too small or dominated by a single call.
func checkROI(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	for _, seq := range nodeSeqs {
		totalTokens := 0

		for _, node := range seq {
			totalTokens += countNodes(node)
		}

		// Too small to benefit from extraction
		if totalTokens < minCloneTokens {
			return domain.ExtractabilityAnalysis{
				ROIPositive: false,
				Confidence:  0.8,
				Reason:       "clone too small — extraction overhead exceeds savings",
			}
		}

		// Helper-dominance: check if a single CallExpr dominates the clone
		largestCallSize := 0

		for _, node := range seq {
			callSize := findLargestCallExprSize(node)
			if callSize > largestCallSize {
				largestCallSize = callSize
			}
		}

		if totalTokens > 0 && float64(largestCallSize) > helperDominanceRatio*float64(totalTokens) {
			return domain.ExtractabilityAnalysis{
				ROIPositive: false,
				Confidence:  0.75,
				Reason:       "clone dominated by single call expression — the call IS the extraction",
			}
		}
	}

	return domain.ExtractabilityAnalysis{
		ROIPositive: true,
		Confidence:  0.85,
	}
}

// countNodes recursively counts all nodes in a subtree (proxy for token count).
func countNodes(node *domain.CloneNode) int {
	if node == nil {
		return 0
	}

	count := 1

	for _, child := range node.Children {
		count += countNodes(child)
	}

	return count
}

// findLargestCallExprSize finds the largest CallExpr subtree size in the clone.
// This detects helper-dominance: when most of the clone is inside one call.
func findLargestCallExprSize(node *domain.CloneNode) int {
	if node == nil {
		return 0
	}

	if node.BaseType == golang.CallExpr {
		return countNodes(node)
	}

	largest := 0

	for _, child := range node.Children {
		size := findLargestCallExprSize(child)
		if size > largest {
			largest = size
		}
	}

	return largest
}

// --- Property 4: Parameterizability ---

// checkParameterizability evaluates whether the clones differ only in
// string-literal values. When the ONLY differences are in domain data
// (not format specifiers), the clones are "already parameterized" —
// the variation IS the logic, not duplication.
func checkParameterizability(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	if len(nodeSeqs) < 2 {
		return domain.ExtractabilityAnalysis{
			Parameterizable: true,
			Confidence:      0.9,
		}
	}

	// Collect string literal values from each clone instance
	literalsPerClone := make([][]string, len(nodeSeqs))

	for i, seq := range nodeSeqs {
		for _, node := range seq {
			collectStringLiterals(node, &literalsPerClone[i])
		}
	}

	// Check if all clones have the same number of string literals
	if !sameLiteralCount(literalsPerClone) {
		return domain.ExtractabilityAnalysis{
			Parameterizable: true,
			Confidence:      0.9,
		}
	}

	// Check if the literal VALUES differ but structure is the same
	if literalsDifferOnlyInValues(literalsPerClone) {
		// Check if the differences are in format specifiers (semantically distinct)
		if hasFormatSpecifierDifferences(literalsPerClone) {
			return domain.ExtractabilityAnalysis{
				Parameterizable: true,
				Confidence:      0.7,
			}
		}

		return domain.ExtractabilityAnalysis{
			Parameterizable: false,
			Confidence:      0.75,
			Reason:          "clones differ only in string-literal values — already parameterized, not duplicated",
		}
	}

	return domain.ExtractabilityAnalysis{
		Parameterizable: true,
		Confidence:      0.9,
	}
}

// collectStringLiterals recursively collects string literal values from a subtree.
// In semantic mode, BasicLit nodes have their values normalized to KIND,
// so we check if the node has any children that carry the original value via Name.
func collectStringLiterals(node *domain.CloneNode, literals *[]string) {
	if node == nil {
		return
	}

	if node.BaseType == golang.BasicLit && node.Name != "" {
		*literals = append(*literals, node.Name)
	}

	for _, child := range node.Children {
		collectStringLiterals(child, literals)
	}
}

func sameLiteralCount(literals [][]string) bool {
	if len(literals) < 2 {
		return true
	}

	count := len(literals[0])
	for _, l := range literals[1:] {
		if len(l) != count {
			return false
		}
	}

	return true
}

func literalsDifferOnlyInValues(literals [][]string) bool {
	if len(literals) < 2 || len(literals[0]) == 0 {
		return false
	}

	// Check if each position has different values across clones
	for pos := range literals[0] {
		values := make(map[string]bool)
		for _, clone := range literals {
			if pos < len(clone) {
				values[clone[pos]] = true
			}
		}

		// If all clones have the same value at this position, it's not a difference
		// We need at least one position where values differ to call it "differing"
		if len(values) > 1 {
			return true
		}
	}

	return false
}

// hasFormatSpecifierDifferences checks if the string literal differences
// involve format specifiers (%x vs %X, %d vs %s), which are semantically distinct.
func hasFormatSpecifierDifferences(literals [][]string) bool {
	for pos := range literals[0] {
		for i := 1; i < len(literals); i++ {
			if pos < len(literals[i]) {
				if isFormatSpecifierDifference(literals[0][pos], literals[i][pos]) {
					return true
				}
			}
		}
	}

	return false
}

// isFormatSpecifierDifference checks if two strings differ only in format
// specifier verbs (case or verb type).
func isFormatSpecifierDifference(a, b string) bool {
	// Check for common format specifier patterns
	// e.g., "#%06x" vs "#%06X", "%d" vs "%s"
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			// Check if this is a format specifier position
			if i > 0 && a[i-1] == '%' {
				return true // Difference right after % is a format specifier difference
			}
		}
	}

	return false
}
