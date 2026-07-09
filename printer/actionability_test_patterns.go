package printer

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isTestDataFilePair reports whether all clones originate from files inside
// a testdata/ directory. Golden/input file pairs are conventional Go test
// fixtures that are expected to be structurally similar — they represent
// before/after snapshots of code transformations.
func isTestDataFilePair(nodeSeqs [][]*domain.CloneNode) bool {
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

	return slices.Contains(parts, "testdata")
}

// isTableDrivenTestBody reports whether every clone is a RangeStmt wrapping
// a t.Run or t.Parallel call — the standard Go table-driven test pattern.
// The loop body is structurally identical across tests because it's the
// testing framework pattern, not duplicated business logic.
func isTableDrivenTestBody(nodeSeqs [][]*domain.CloneNode) bool {
	if len(nodeSeqs) == 0 {
		return false
	}

	for _, seq := range nodeSeqs {
		if len(seq) == 0 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.RangeStmt {
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
func containsTRunCall(node *domain.CloneNode) bool {
	if node.BaseType == golang.CallExpr {
		for _, child := range node.Children {
			if child.BaseType == golang.SelectorExpr && child.Name == "Run" {
				return true
			}
		}
	}

	return slices.ContainsFunc(node.Children, containsTRunCall)
}

// allFromTestFile checks if all nodes in a sequence come from _test.go files.
func allFromTestFile(seq []*domain.CloneNode) bool {
	for _, n := range seq {
		if !strings.HasSuffix(n.Filename, "_test.go") {
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
func isTestScaffolding(nodeSeqs [][]*domain.CloneNode) bool {
	if len(nodeSeqs) == 0 {
		return false
	}

	return everySequenceMatch(nodeSeqs, isTestScaffoldingSequence)
}

// isTestScaffoldingSequence checks a single clone sequence for the
// test scaffolding pattern.
func isTestScaffoldingSequence(seq []*domain.CloneNode) bool {
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
func walkForTestScaffoldingSignals(node *domain.CloneNode, hasFileIO *bool, assertionNames *map[string]bool) {
	if *assertionNames == nil {
		*assertionNames = make(map[string]bool)
	}

	if node.BaseType == golang.CallExpr {
		for _, child := range node.Children {
			bt := child.BaseType
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
					"HaveOccurred", "ShouldNot":
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
