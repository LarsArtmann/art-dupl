package printer

import (
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// RAII cleanup method names that indicate non-actionable patterns.
const cleanupMethodName = "Unlock"

// processOrderMethodName is a non-RAII method name for testing.
const processOrderMethodName = "processOrder"

// baseTypeOf extracts the base AST node type from a syntax.Node,
// decoding any semantic encoding (identifier hash, operator hash) that
// may be baked into the upper bits of Type. This ensures type comparisons
// work correctly in both semantic and structural detection modes.
func baseTypeOf(n *syntax.Node) int32 {
	return golang.DecodeBaseType(n.Type)
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
func EvaluateActionability(nodeSeqs [][]*syntax.Node) domain.CloneActionability {
	_, a := evaluateActionabilityDetailed(nodeSeqs)

	return a
}

// PatternLabel identifies which non-actionable pattern was detected.
// Empty string means actionable.
type PatternLabel string

const (
	PatternNone             PatternLabel = ""
	PatternTestData         PatternLabel = "testdata-pair"
	PatternTableDrivenTest  PatternLabel = "table-driven-test"
	PatternTestScaffolding  PatternLabel = "test-scaffolding"
	PatternDataDominated    PatternLabel = "data-dominated"
	PatternSignatureOnly    PatternLabel = "signature-only"
	PatternRAIIDefer        PatternLabel = "raii-defer"
	PatternErrorPropagation PatternLabel = "error-propagation"
)

// EvaluateActionabilityWithLabel returns both the actionability and the
// pattern label that caused it. This allows downstream code to adjust
// category/suggestion based on which specific pattern was detected.
func EvaluateActionabilityWithLabel(nodeSeqs [][]*syntax.Node) (PatternLabel, domain.CloneActionability) {
	return evaluateActionabilityDetailed(nodeSeqs)
}

func evaluateActionabilityDetailed(nodeSeqs [][]*syntax.Node) (PatternLabel, domain.CloneActionability) {
	if len(nodeSeqs) == 0 {
		return PatternNone, domain.Actionable
	}

	if isSignatureOnlyMatch(nodeSeqs) {
		return PatternSignatureOnly, domain.NonActionable
	}

	if isPureDeferPattern(nodeSeqs) {
		return PatternRAIIDefer, domain.NonActionable
	}

	if isPureErrorPropagation(nodeSeqs) {
		return PatternErrorPropagation, domain.NonActionable
	}

	if isTestDataFilePair(nodeSeqs) {
		return PatternTestData, domain.NonActionable
	}

	if isTableDrivenTestBody(nodeSeqs) {
		return PatternTableDrivenTest, domain.NonActionable
	}

	if isTestScaffolding(nodeSeqs) {
		return PatternTestScaffolding, domain.NonActionable
	}

	if isDataDominated(nodeSeqs) {
		return PatternDataDominated, domain.NonActionable
	}

	return PatternNone, domain.Actionable
}

// isSignatureOnlyMatch reports whether every clone is a single FuncDecl node
// WITHOUT a substantial body (>3 child nodes). Interface method signatures,
// empty stubs, and forwarding methods fit this pattern.
//
// A FuncDecl with a real body (BlockStmt with >0 children) is actionable —
// identical bodies can be extracted to shared functions. But if two FuncDecls
// match and the body is empty (or has only boilerplate), deduplication is
// impossible without changing the interface.
func isSignatureOnlyMatch(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if baseTypeOf(root) != golang.FuncDecl {
			return false
		}

		// A FuncDecl with a real body has a BlockStmt child that has children.
		// If the body is empty (no children), or if there are very few children
		// (just receiver + name + type), it's a signature-only implementation.
		if hasRealBody(root) {
			return false
		}
	}

	return true
}

// hasRealBody checks if a FuncDecl contains a body with meaningful logic
// beyond the signature itself.
func hasRealBody(node *syntax.Node) bool {
	for _, child := range node.Children {
		if child.Type == golang.BlockStmt && len(child.Children) > 0 {
			return true
		}
	}

	return false
}

// isPureDeferPattern reports whether every clone is a DeferStmt
// wrapping a RAII-style call (Unlock, Close, etc.).
//
// Using the Name field on child nodes, we can now distinguish
// `defer mu.Unlock()` from `defer processOrder()` — the former is
// idiomatic RAII cleanup (non-actionable), the latter is real duplication.
func isPureDeferPattern(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		if baseTypeOf(seq[0]) != golang.DeferStmt {
			return false
		}

		if !isRAIIDeferCall(seq[0]) {
			return false
		}
	}

	return true
}

// isRAIIDeferCall checks if a DeferStmt wraps a known RAII cleanup method.
func isRAIIDeferCall(node *syntax.Node) bool {
	for _, child := range node.Children {
		if baseTypeOf(child) == golang.CallExpr {
			for _, arg := range child.Children {
				if baseTypeOf(arg) == golang.SelectorExpr && isCleanupMethod(arg.Name) {
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
	case cleanupMethodName, "Close", "Done", "Cancel", "Release", "Finish", "Disconnect", "Free":
		return true
	default:
		return false
	}
}

// isPureErrorPropagation reports whether every clone is an IfStmt
// that only contains error propagation: if err != nil { return err }.
func isPureErrorPropagation(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if baseTypeOf(root) != golang.IfStmt {
			return false
		}

		if !isErrorOnlyIf(root) {
			return false
		}
	}

	return true
}

// isErrorOnlyIf checks if an IfStmt is a pure error propagation pattern.
// It must have:
//   - A condition containing a comparison against nil (BinaryExpr with nil)
//   - A body containing only a ReturnStmt (or CallExpr wrapping error)
//   - No Else branch
func isErrorOnlyIf(node *syntax.Node) bool {
	var (
		hasNilCompare bool
		hasReturnErr  bool
		hasElse       bool
	)

	for _, child := range node.Children {
		switch baseTypeOf(child) {
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
			if baseTypeOf(child) != golang.AssignStmt &&
				baseTypeOf(child) != golang.DeclStmt {
				hasElse = true
			}
		}
	}

	return hasNilCompare && hasReturnErr && !hasElse
}

// containsNilIdentifier checks if a BinaryExpr compares against nil.
// Uses the Name field to verify an identifier named "nil" is present,
// rather than just checking for any Ident node.
func containsNilIdentifier(node *syntax.Node) bool {
	for _, child := range node.Children {
		if baseTypeOf(child) == golang.Ident && child.Name == "nil" {
			return true
		}
	}

	return false
}

// isReturnOrWrappedReturn checks if a BlockStmt only contains a ReturnStmt.
func isReturnOrWrappedReturn(node *syntax.Node) bool {
	if len(node.Children) == 0 {
		return false
	}

	// Allow single return statement.
	if len(node.Children) == 1 && baseTypeOf(node.Children[0]) == golang.ReturnStmt {
		return true
	}

	// Allow return with a CallExpr (e.g., return fmt.Errorf("...")).
	if len(node.Children) == 1 {
		child := node.Children[0]
		if baseTypeOf(child) == golang.ReturnStmt || baseTypeOf(child) == golang.CallExpr {
			return true
		}
	}

	return false
}

// isTestDataFilePair reports whether all clones originate from files inside
// a testdata/ directory. Golden/input file pairs are conventional Go test
// fixtures that are expected to be structurally similar — they represent
// before/after snapshots of code transformations.
func isTestDataFilePair(nodeSeqs [][]*syntax.Node) bool {
	if len(nodeSeqs) == 0 {
		return false
	}

	for _, seq := range nodeSeqs {
		if len(seq) == 0 {
			return false
		}

		if !isInTestDataDir(seq[0].Filename) {
			return false
		}
	}

	return true
}

// isInTestDataDir checks if a file path contains a testdata/ directory
// component, following Go's standard testing convention.
func isInTestDataDir(filename string) bool {
	parts := strings.Split(filepath.ToSlash(filename), "/")
	for _, part := range parts {
		if part == "testdata" {
			return true
		}
	}

	return false
}

// isTableDrivenTestBody reports whether every clone is a RangeStmt wrapping
// a t.Run or t.Parallel call — the standard Go table-driven test pattern.
// The loop body is structurally identical across tests because it's the
// testing framework pattern, not duplicated business logic.
func isTableDrivenTestBody(nodeSeqs [][]*syntax.Node) bool {
	if len(nodeSeqs) == 0 {
		return false
	}

	for _, seq := range nodeSeqs {
		if len(seq) == 0 {
			return false
		}

		root := seq[0]
		if baseTypeOf(root) != golang.RangeStmt {
			return false
		}

		if !containsTRunCall(root) {
			return false
		}

		if !allFromTestFile(seq) {
			return false
		}
	}

	return true
}

// containsTRunCall checks if a node tree contains a CallExpr where the
// function is a SelectorExpr with method name "Run" — matching t.Run().
func containsTRunCall(node *syntax.Node) bool {
	if baseTypeOf(node) == golang.CallExpr {
		for _, child := range node.Children {
			if baseTypeOf(child) == golang.SelectorExpr && child.Name == "Run" {
				return true
			}
		}
	}

	for _, child := range node.Children {
		if containsTRunCall(child) {
			return true
		}
	}

	return false
}

// allFromTestFile checks if all nodes in a sequence come from _test.go files.
func allFromTestFile(seq []*syntax.Node) bool {
	for _, n := range seq {
		if !isTestFile(n.Filename) {
			return false
		}
	}

	return true
}

// isTestScaffolding reports whether every clone matches the common test
// setup/teardown pattern: create temp dir, write file, run check, assert.
// This covers Ginkgo When/It blocks and standard test helpers that follow
// the same structural pattern with only data differences.
//
// Detection strategy: look for assertion calls within _test.go files where
// the clone also contains either temp dir/file creation OR file I/O (WriteFile,
// ReadFile, MkdirAll). The combination of assertions + file setup is a strong
// signal of test scaffolding. Falls back to detecting 3+ distinct assertion
// methods as sufficient evidence on its own.
func isTestScaffolding(nodeSeqs [][]*syntax.Node) bool {
	if len(nodeSeqs) == 0 {
		return false
	}

	for _, seq := range nodeSeqs {
		if !isTestScaffoldingSequence(seq) {
			return false
		}
	}

	return true
}

// isTestScaffoldingSequence checks a single clone sequence for the
// test scaffolding pattern.
func isTestScaffoldingSequence(seq []*syntax.Node) bool {
	if len(seq) == 0 {
		return false
	}

	if !allFromTestFile(seq) {
		return false
	}

	var (
		hasFileIO      bool
		assertionNames map[string]bool
	)

	for _, node := range seq {
		walkForTestScaffoldingSignals(node, &hasFileIO, &assertionNames)
	}

	if hasFileIO && len(assertionNames) >= 1 {
		return true
	}

	return len(assertionNames) >= 3
}

// walkForTestScaffoldingSignals walks a node tree looking for test
// scaffolding indicators: file I/O operations and assertion calls.
func walkForTestScaffoldingSignals(node *syntax.Node, hasFileIO *bool, assertionNames *map[string]bool) {
	if *assertionNames == nil {
		*assertionNames = make(map[string]bool)
	}

	if baseTypeOf(node) == golang.CallExpr {
		for _, child := range node.Children {
			bt := baseTypeOf(child)
			name := child.Name

			if bt == golang.SelectorExpr {
				switch name {
				case "TempDir", "TempFile":
					*hasFileIO = true
				case "WriteFile", "ReadFile", "MkdirAll", "MkdirTemp":
					*hasFileIO = true
				case "Expect", "Should", "So", "Assert", "Check", "Require":
					(*assertionNames)["expect-family"] = true
				case "NotTo", "To", "Not", "ToNot":
					(*assertionNames)["matcher-chain"] = true
				case "Equal", "HaveLen", "BeEmpty", "BeNil", "BeTrue", "BeFalse",
					"BeZero", "ContainElement", "ContainSubstring", "MatchRegexp",
					"ConsistOf", "HaveCap", "HaveKey", "HaveValue", "OccurOnlyOnce",
					"(HaveOccurred", "ShouldNot":
					(*assertionNames)["assertion"] = true
				case "Fatalf", "Errorf", "Skipf", "Logf", "FailNow":
					(*assertionNames)["testing-t"] = true
				}
			}
		}
	}

	for _, child := range node.Children {
		walkForTestScaffoldingSignals(child, hasFileIO, assertionNames)
	}
}

// isDataDominated reports whether every clone is dominated by data nodes
// (BasicLit and KeyValueExpr) rather than logic. When >70% of leaf nodes
// are data, the clone represents struct initialization, config fixtures,
// or test data arrays — not duplicated business logic.
func isDataDominated(nodeSeqs [][]*syntax.Node) bool {
	if len(nodeSeqs) == 0 {
		return false
	}

	for _, seq := range nodeSeqs {
		if !isSequenceDataDominated(seq) {
			return false
		}
	}

	return true
}

const dataDominanceRatio = 0.6

// isSequenceDataDominated checks if a single clone sequence is dominated
// by data nodes (BasicLit, KeyValueExpr) rather than logic nodes.
func isSequenceDataDominated(seq []*syntax.Node) bool {
	total := 0
	data := 0

	for _, node := range seq {
		countDataNodes(node, &total, &data)
	}

	if total == 0 {
		return false
	}

	return float64(data)/float64(total) >= dataDominanceRatio
}

// countDataNodes walks a node tree counting all nodes and data-type nodes.
// Data nodes are BasicLit (string/number literals) and KeyValueExpr
// (struct field initializers). A high ratio of data nodes indicates
// struct initialization or config fixtures, not duplicated logic.
func countDataNodes(node *syntax.Node, total, data *int) {
	*total++

	bt := baseTypeOf(node)
	if bt == golang.BasicLit || bt == golang.KeyValueExpr {
		*data++
	}

	for _, child := range node.Children {
		countDataNodes(child, total, data)
	}
}
