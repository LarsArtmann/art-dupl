package actionability

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// Confidence values for property analysis results.
const (
	confidenceHigh   = 0.9
	confidenceMedium = 0.85
	confidenceLower  = 0.75
)

// helperDominanceRatio is the threshold for helper-dominance: if >60% of clone
// tokens are inside a single CallExpr, the call IS the extraction.
const helperDominanceRatio = 0.6

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
		Confidence:              confidenceHigh,
		Reason:                  "all properties pass",
	}

	cf := checkControlFlow(nodeSeqs)
	if !cf.ControlFlowExtractable {
		analysis.ControlFlowExtractable = false
		analysis.Confidence = minConfidence(analysis.Confidence, cf.Confidence)
		analysis.Reason = cf.Reason
	}

	roi := checkROI(nodeSeqs)
	if !roi.ROIPositive {
		analysis.ROIPositive = false
		analysis.Confidence = minConfidence(analysis.Confidence, roi.Confidence)
		if analysis.ControlFlowExtractable {
			analysis.Reason = roi.Reason
		}
	}

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
// that are forced by the enclosing function's signature. Only fires for
// error-guard IfStmts (nil comparison + return) in void functions — the exact
// HTTP handler pattern from DiscordSync feedback.
func checkControlFlow(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	for _, seq := range nodeSeqs {
		for _, node := range seq {
			if node.BaseType != golang.IfStmt {
				continue
			}

			if node.EnclosingReturnArity != 0 {
				continue
			}

			if !isErrorGuardInVoidFunc(node) {
				continue
			}

			return domain.ExtractabilityAnalysis{
				ControlFlowExtractable: false,
				Confidence:             confidenceMedium,
				Reason:                  "clone contains return in void function — extraction cannot issue return on behalf of caller",
			}
		}
	}

	return domain.ExtractabilityAnalysis{
		ControlFlowExtractable: true,
		Confidence:             confidenceHigh,
	}
}

// isErrorGuardInVoidFunc checks if an IfStmt is an error guard: has a nil
// comparison condition and a return in the body. This is the HTTP handler
// error guard pattern that cannot be extracted from void functions.
func isErrorGuardInVoidFunc(node *domain.CloneNode) bool {
	hasNilCompare := false
	hasReturn := false

	for _, child := range node.Children {
		if child.BaseType == golang.BinaryExpr && subtreeHasNodeType(child, golang.Ident) {
			for _, c := range child.Children {
				if c.BaseType == golang.Ident && c.Name == "nil" {
					hasNilCompare = true
				}
			}
		}

		if child.BaseType == golang.BlockStmt && subtreeHasNodeType(child, golang.ReturnStmt) {
			hasReturn = true
		}
	}

	return hasNilCompare && hasReturn
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

// checkROI evaluates whether extracting the clone would save tokens.
// Only flags helper-dominance: when >60% of clone tokens are inside a single
// CallExpr, the call IS the extraction. Does NOT flag small clones — that's
// handled by existing patterns (single-call-expression, single-simple-statement).
func checkROI(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	for _, seq := range nodeSeqs {
		if len(seq) < 2 {
			continue
		}

		totalTokens := 0

		for _, node := range seq {
			totalTokens += countNodes(node)
		}

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
				Confidence:  confidenceLower,
				Reason:       "clone dominated by single call expression — the call IS the extraction",
			}
		}
	}

	return domain.ExtractabilityAnalysis{
		ROIPositive: true,
		Confidence:  confidenceMedium,
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
// string-literal values. When the ONLY differences are in domain data,
// the clones are "already parameterized" — the variation IS the logic.
func checkParameterizability(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	if len(nodeSeqs) < 2 {
		return domain.ExtractabilityAnalysis{
			Parameterizable: true,
			Confidence:      confidenceHigh,
		}
	}

	literalsPerClone := make([][]string, len(nodeSeqs))

	for i, seq := range nodeSeqs {
		for _, node := range seq {
			collectStringLiterals(node, &literalsPerClone[i])
		}
	}

	if !sameLiteralCount(literalsPerClone) {
		return domain.ExtractabilityAnalysis{
			Parameterizable: true,
			Confidence:      confidenceHigh,
		}
	}

	if literalsDifferOnlyInValues(literalsPerClone) && !hasFormatSpecifierDifferences(literalsPerClone) {
		return domain.ExtractabilityAnalysis{
			Parameterizable: false,
			Confidence:      confidenceLower,
			Reason:          "clones differ only in string-literal values — already parameterized, not duplicated",
		}
	}

	return domain.ExtractabilityAnalysis{
		Parameterizable: true,
		Confidence:      confidenceHigh,
	}
}

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

	for pos := range literals[0] {
		values := make(map[string]bool)
		for _, clone := range literals {
			if pos < len(clone) {
				values[clone[pos]] = true
			}
		}

		if len(values) > 1 {
			return true
		}
	}

	return false
}

func hasFormatSpecifierDifferences(literals [][]string) bool {
	for pos := range literals[0] {
		for i := 1; i < len(literals); i++ {
			if pos < len(literals[i]) && isFormatSpecifierDifference(literals[0][pos], literals[i][pos]) {
				return true
			}
		}
	}

	return false
}

func isFormatSpecifierDifference(a, b string) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] && i > 0 && a[i-1] == '%' {
			return true
		}
	}

	return false
}
