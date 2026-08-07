package actionability

import (
	"slices"

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
				Reason:                 "clone contains return in void function — extraction cannot issue return on behalf of caller",
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
				Reason:      "clone dominated by single call expression — the call IS the extraction",
			}
		}
	}

	return domain.ExtractabilityAnalysis{
		ROIPositive: true,
		Confidence:  confidenceMedium,
	}
}

// countNodes recursively counts all nodes in a subtree (proxy for token count).
// art-dupl:accept different type than syntax.Node variant; cannot share code without generics overhead
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
//
// This veto only applies to clones WITHOUT control flow. A clone with
// if/switch/for statements and differing literals is still actionable:
// the control-flow pattern is worth extracting into a parameterized helper.
// Bare sequential calls with different arguments are "already parameterized"
// because the calls themselves are the abstraction.
func checkParameterizability(nodeSeqs [][]*domain.CloneNode) domain.ExtractabilityAnalysis {
	if len(nodeSeqs) < 2 {
		return domain.ExtractabilityAnalysis{
			Parameterizable: true,
			Confidence:      confidenceHigh,
		}
	}

	if anySeqHasControlFlow(nodeSeqs) {
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

// anySeqHasControlFlow reports whether any clone instance contains a
// control-flow statement (if, switch, for, range, select). Such clones
// represent structural logic worth extracting even when literals differ.
func anySeqHasControlFlow(nodeSeqs [][]*domain.CloneNode) bool {
	for _, seq := range nodeSeqs {
		if slices.ContainsFunc(seq, subtreeHasControlFlow) {
			return true
		}
	}

	return false
}

func subtreeHasControlFlow(node *domain.CloneNode) bool {
	if node == nil {
		return false
	}

	switch node.BaseType {
	case golang.IfStmt, golang.SwitchStmt, golang.TypeSwitchStmt,
		golang.ForStmt, golang.RangeStmt, golang.SelectStmt:
		return true
	}

	return slices.ContainsFunc(node.Children, subtreeHasControlFlow)
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

// isFormatSpecifierDifference reports whether two string literals differ in
// their Go fmt verbs. Returns true when the format specifiers differ (e.g.
// "%d" vs "%s", "%5d" vs "%3d", "%[1]d" vs "%[2]d"), indicating the clones
// have genuinely different formatting behavior.
//
// "%%" (literal percent) is not a real specifier and is ignored, so two
// strings that differ only around "%%" are not flagged as format differences.
func isFormatSpecifierDifference(a, b string) bool {
	specsA := extractFormatSpecifiers(a)
	specsB := extractFormatSpecifiers(b)

	if len(specsA) != len(specsB) {
		return true
	}

	for i, spec := range specsA {
		if spec != specsB[i] {
			return true
		}
	}

	return false
}

// extractFormatSpecifiers parses Go fmt verbs from a string literal and returns
// each complete specifier (e.g. "%d", "%5.2f", "%[1]s", "%-#7.3v"). "%%" (literal
// percent) is excluded since it does not consume an argument.
func extractFormatSpecifiers(s string) []string {
	var specs []string

	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			continue
		}

		spec, end := scanFormatSpecifier(s, i)
		if spec != "" {
			specs = append(specs, spec)
		}

		i = end
	}

	return specs
}

// scanFormatSpecifier parses a single Go fmt verb starting at s[start] (which
// must be '%'). Returns the specifier substring and the index of the last
// consumed byte. Returns ("", start) for "%%" or malformed input.
func scanFormatSpecifier(s string, start int) (string, int) {
	i := start + 1
	if i >= len(s) {
		return "", start
	}

	// "%%" is a literal percent, not a real specifier.
	if s[i] == '%' {
		return "", i
	}

	// Skip argument index, flags, width, and precision to reach the verb.
	i = scanFmtBody(s, i)
	if i >= len(s) {
		return "", start
	}

	return s[start : i+1], i
}

// scanFmtBody advances past the optional argument index, flags, width, and
// precision components of a format verb, returning the index of the verb char.
func scanFmtBody(s string, i int) int {
	i = scanArgIndex(s, i)

	for i < len(s) && isFmtFlag(s[i]) {
		i++
	}

	i = scanDigits(s, i)

	if i < len(s) && s[i] == '.' {
		i = scanDigits(s, i+1)
	}

	return i
}

// scanArgIndex skips an optional %[n] argument index. Returns the original
// index if there is no argument index or if the bracket is never closed.
func scanArgIndex(s string, i int) int {
	if i >= len(s) || s[i] != '[' {
		return i
	}

	j := i + 1
	for j < len(s) && s[j] != ']' {
		j++
	}

	if j >= len(s) {
		return i // malformed: no closing ]
	}

	return j + 1
}

// scanDigits advances past a run of ASCII digits starting at s[i].
func scanDigits(s string, i int) int {
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}

	return i
}

func isFmtFlag(c byte) bool {
	switch c {
	case '+', '-', '#', '0', ' ':
		return true
	default:
		return false
	}
}
