package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

type boolTestCase struct {
	name     string
	seqs     [][]*syntax.Node
	expected bool
}

func runBoolTests(t *testing.T, fn func([][]*syntax.Node) bool, cases []boolTestCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			result := fn(tc.seqs)
			if result != tc.expected {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestIsTestDataFilePair(t *testing.T) {
	runBoolTests(t, isTestDataFilePair, []boolTestCase{
		{
			name: "both files in testdata directory",
			seqs: [][]*syntax.Node{
				{mustNodeWithFilename("pkg/testdata/input.go")},
				{mustNodeWithFilename("pkg/testdata/golden.go")},
			},
			expected: true,
		},
		{
			name: "nested testdata directory",
			seqs: [][]*syntax.Node{
				{mustNodeWithFilename("internal/rule/testdata/errors_as/golden.go")},
				{mustNodeWithFilename("internal/rule/testdata/errors_as/input.go")},
			},
			expected: true,
		},
		{
			name: "one file not in testdata",
			seqs: [][]*syntax.Node{
				{mustNodeWithFilename("pkg/testdata/input.go")},
				{mustNodeWithFilename("pkg/rule.go")},
			},
			expected: false,
		},
		{
			name: "production files only",
			seqs: [][]*syntax.Node{
				{mustNodeWithFilename("pkg/handler.go")},
				{mustNodeWithFilename("pkg/service.go")},
			},
			expected: false,
		},
		{name: "empty sequence", seqs: [][]*syntax.Node{}, expected: false},
		{
			name: "file named testdata.go but not in testdata dir",
			seqs: [][]*syntax.Node{
				{mustNodeWithFilename("pkg/testdata.go")},
				{mustNodeWithFilename("pkg/testdata_test.go")},
			},
			expected: false,
		},
		{
			name: "empty node in sequence",
			seqs: [][]*syntax.Node{
				{},
				{mustNodeWithFilename("pkg/testdata/input.go")},
			},
			expected: false,
		},
	})
}

func TestIsTableDrivenTestBody(t *testing.T) {
	runBoolTests(t, isTableDrivenTestBody, []boolTestCase{
		{
			name: "RangeStmt with t.Run in test file",
			seqs: [][]*syntax.Node{
				{mustRangeStmtWithTRun("resolve_test.go")},
				{mustRangeStmtWithTRun("filter_test.go")},
			},
			expected: true,
		},
		{
			name: "RangeStmt without t.Run in test file",
			seqs: [][]*syntax.Node{
				{mustRangeStmtNoTRun("resolve_test.go")},
				{mustRangeStmtNoTRun("filter_test.go")},
			},
			expected: false,
		},
		{
			name: "RangeStmt with t.Run but not test file",
			seqs: [][]*syntax.Node{
				{mustRangeStmtWithTRun("resolve.go")},
				{mustRangeStmtWithTRun("filter.go")},
			},
			expected: false,
		},
		{
			name: "non-RangeStmt in test file",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Filename: "resolve_test.go"}},
				{{Type: golang.FuncDecl, Filename: "filter_test.go"}},
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*syntax.Node{}, expected: false},
		{
			name: "mixed: one RangeStmt one not",
			seqs: [][]*syntax.Node{
				{mustRangeStmtWithTRun("resolve_test.go")},
				{{Type: golang.FuncDecl, Filename: "filter_test.go"}},
			},
			expected: false,
		},
	})
}

func TestIsTestScaffolding(t *testing.T) {
	runBoolTests(t, isTestScaffolding, []boolTestCase{
		{
			name: "Ginkgo TempDir + WriteFile + Expect pattern",
			seqs: [][]*syntax.Node{
				{mustTestScaffolding("benchmark_naming_rule_test.go")},
				{mustTestScaffolding("testdata_directory_rule_test.go")},
			},
			expected: true,
		},
		{
			name: "TempDir without assertions",
			seqs: [][]*syntax.Node{
				{mustTempDirOnly("rule_test.go")},
				{mustTempDirOnly("handler_test.go")},
			},
			expected: false,
		},
		{
			name: "Assertions without TempDir",
			seqs: [][]*syntax.Node{
				{mustAssertionsOnly("rule_test.go")},
				{mustAssertionsOnly("handler_test.go")},
			},
			expected: false,
		},
		{
			name: "3+ distinct assertions without TempDir is scaffolding",
			seqs: [][]*syntax.Node{
				{mustThreeDistinctAssertions("rule_test.go")},
				{mustThreeDistinctAssertions("handler_test.go")},
			},
			expected: true,
		},
		{
			name: "production files with similar pattern",
			seqs: [][]*syntax.Node{
				{mustTestScaffolding("handler.go")},
				{mustTestScaffolding("service.go")},
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*syntax.Node{}, expected: false},
	})
}

func TestIsDataDominated(t *testing.T) {
	runBoolTests(t, isDataDominated, []boolTestCase{
		{
			name: "struct init with many BasicLit values",
			seqs: [][]*syntax.Node{
				mustDataDominatedSequence(),
				mustDataDominatedSequence(),
			},
			expected: true,
		},
		{
			name: "logic-heavy sequence",
			seqs: [][]*syntax.Node{
				mustLogicHeavySequence(),
				mustLogicHeavySequence(),
			},
			expected: false,
		},
		{
			name: "mixed data and logic",
			seqs: [][]*syntax.Node{
				mustMixedDataLogicSequence(),
				mustMixedDataLogicSequence(),
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*syntax.Node{}, expected: false},
		{
			name: "empty node list in sequence",
			seqs: [][]*syntax.Node{
				{},
				{},
			},
			expected: false,
		},
	})
}

func TestEvaluateActionabilityWithLabel(t *testing.T) {
	tests := []struct {
		name           string
		seqs           [][]*syntax.Node
		expectedLabel  PatternLabel
		expectedAction domain.CloneActionability
	}{
		{
			name: "testdata pair gets correct label",
			seqs: [][]*syntax.Node{
				{mustNodeWithFilename("pkg/testdata/input.go")},
				{mustNodeWithFilename("pkg/testdata/golden.go")},
			},
			expectedLabel:  PatternTestData,
			expectedAction: domain.NonActionable,
		},
		{
			name: "table-driven test gets correct label",
			seqs: [][]*syntax.Node{
				{mustRangeStmtWithTRun("resolve_test.go")},
				{mustRangeStmtWithTRun("filter_test.go")},
			},
			expectedLabel:  PatternTableDrivenTest,
			expectedAction: domain.NonActionable,
		},
		{
			name: "test scaffolding gets correct label",
			seqs: [][]*syntax.Node{
				{mustTestScaffolding("rule_test.go")},
				{mustTestScaffolding("handler_test.go")},
			},
			expectedLabel:  PatternTestScaffolding,
			expectedAction: domain.NonActionable,
		},
		{
			name: "RAII defer gets correct label",
			seqs: [][]*syntax.Node{
				{mustDeferSelectorCall("mu", cleanupMethodName)},
				{mustDeferSelectorCall("mu", cleanupMethodName)},
			},
			expectedLabel:  PatternRAIIDefer,
			expectedAction: domain.NonActionable,
		},
		{
			name: "error propagation gets correct label",
			seqs: [][]*syntax.Node{
				{mustIfErrReturnNil()},
				{mustIfErrReturnNil()},
			},
			expectedLabel:  PatternErrorPropagation,
			expectedAction: domain.NonActionable,
		},
		{
			name: "signature-only gets correct label",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 1}},
				{{Type: golang.FuncDecl, Owns: 1}},
			},
			expectedLabel:  PatternSignatureOnly,
			expectedAction: domain.NonActionable,
		},
		{
			name: "interface implementation gets correct label",
			seqs: [][]*syntax.Node{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/json.go")},
				{mustFuncTypeNode("printer/html.go")},
			},
			expectedLabel:  PatternInterfaceImpl,
			expectedAction: domain.NonActionable,
		},
		{
			name: "actionable code gets no label",
			seqs: [][]*syntax.Node{
				mustFuncDeclWithBody(),
				mustFuncDeclWithBody(),
			},
			expectedLabel:  PatternNone,
			expectedAction: domain.Actionable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			label, action := EvaluateActionabilityWithLabel(tc.seqs)
			if label != tc.expectedLabel {
				t.Errorf("label = %q, want %q", label, tc.expectedLabel)
			}

			if action != tc.expectedAction {
				t.Errorf("action = %q, want %q", action, tc.expectedAction)
			}
		})
	}
}

func TestApplyPatternLabel(t *testing.T) {
	tests := []struct {
		name         string
		input        domain.CloneClassification
		label        PatternLabel
		wantCategory domain.CloneCategory
		wantPriority domain.ClonePriority
	}{
		{
			name:         "testdata label upgrades to test-fixture category",
			input:        domain.CloneClassification{Category: domain.CategoryFunction},
			label:        PatternTestData,
			wantCategory: domain.CategoryTestFixture,
			wantPriority: domain.PriorityLow,
		},
		{
			name:         "table-driven label upgrades to test-boilerplate category",
			input:        domain.CloneClassification{Category: domain.CategoryLoop},
			label:        PatternTableDrivenTest,
			wantCategory: domain.CategoryTestBoilerplate,
			wantPriority: domain.PriorityLow,
		},
		{
			name:         "scaffolding label upgrades to test-boilerplate category",
			input:        domain.CloneClassification{Category: domain.CategoryFunction},
			label:        PatternTestScaffolding,
			wantCategory: domain.CategoryTestBoilerplate,
			wantPriority: domain.PriorityLow,
		},
		{
			name:         "data-dominated keeps category but lowers priority",
			input:        domain.CloneClassification{Category: domain.CategoryFunction, Priority: domain.PriorityHigh},
			label:        PatternDataDominated,
			wantCategory: domain.CategoryFunction,
			wantPriority: domain.PriorityLow,
		},
		{
			name:         "no label keeps original classification",
			input:        domain.CloneClassification{Category: domain.CategoryFunction, Priority: domain.PriorityHigh},
			label:        PatternNone,
			wantCategory: domain.CategoryFunction,
			wantPriority: domain.PriorityHigh,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := applyPatternLabel(tc.input, tc.label)
			if result.Category != tc.wantCategory {
				t.Errorf("category = %q, want %q", result.Category, tc.wantCategory)
			}

			if result.Priority != tc.wantPriority {
				t.Errorf("priority = %q, want %q", result.Priority, tc.wantPriority)
			}
		})
	}
}

func mustNodeWithFilename(filename string) *syntax.Node {
	return &syntax.Node{
		Type:     golang.FuncDecl,
		Owns:     5,
		Filename: filename,
		Children: []*syntax.Node{
			{Type: golang.BlockStmt, Children: []*syntax.Node{
				{Type: golang.ExprStmt},
			}},
		},
	}
}

func mustRangeStmtWithTRun(filename string) *syntax.Node {
	return &syntax.Node{
		Type:     golang.RangeStmt,
		Filename: filename,
		Children: []*syntax.Node{
			{
				Type: golang.BlockStmt,
				Children: []*syntax.Node{
					{
						Type: golang.ExprStmt,
						Children: []*syntax.Node{
							{
								Type: golang.CallExpr,
								Children: []*syntax.Node{
									{
										Type: golang.SelectorExpr,
										Name: "Run",
										Children: []*syntax.Node{
											{Type: golang.Ident, Name: "t"},
											{Type: golang.Ident, Name: "Run"},
										},
									},
									{
										Type: golang.BasicLit,
									},
									{
										Type: golang.FuncLit,
										Children: []*syntax.Node{
											{Type: golang.FuncType},
											{
												Type: golang.BlockStmt,
												Children: []*syntax.Node{
													{Type: golang.ExprStmt},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func mustRangeStmtNoTRun(filename string) *syntax.Node {
	return &syntax.Node{
		Type:     golang.RangeStmt,
		Filename: filename,
		Children: []*syntax.Node{
			{
				Type: golang.BlockStmt,
				Children: []*syntax.Node{
					mustSelectorExprStmt("", "Process"),
				},
			},
		},
	}
}

func mustTestScaffolding(filename string) *syntax.Node {
	return &syntax.Node{
		Type:     golang.ExprStmt,
		Filename: filename,
		Children: []*syntax.Node{
			{
				Type: golang.AssignStmt,
				Children: []*syntax.Node{
					{
						Type: golang.CallExpr,
						Children: []*syntax.Node{
							{
								Type: golang.SelectorExpr,
								Name: "TempDir",
								Children: []*syntax.Node{
									{Type: golang.CallExpr, Children: []*syntax.Node{
										{Type: golang.SelectorExpr, Name: "GinkgoT"},
									}},
								},
							},
						},
					},
				},
			},
			mustSelectorExprStmt("", "WriteFile"),
			mustSelectorExprStmt("", "NotTo"),
			mustSelectorExprStmt("", "Equal"),
		},
	}
}

func mustSelectorExprStmt(filename, selectorName string) *syntax.Node {
	return &syntax.Node{
		Type:     golang.ExprStmt,
		Filename: filename,
		Children: []*syntax.Node{
			{
				Type: golang.CallExpr,
				Children: []*syntax.Node{
					{
						Type: golang.SelectorExpr,
						Name: selectorName,
					},
				},
			},
		},
	}
}

func mustTempDirOnly(filename string) *syntax.Node {
	return mustSelectorExprStmt(filename, "TempDir")
}

func mustAssertionsOnly(filename string) *syntax.Node {
	return mustSelectorExprStmt(filename, "Equal")
}

func mustKeyValueExprFields(names ...string) []*syntax.Node {
	nodes := make([]*syntax.Node, len(names))
	for i, name := range names {
		nodes[i] = &syntax.Node{
			Type: golang.KeyValueExpr,
			Children: []*syntax.Node{
				{Type: golang.Ident, Name: name},
				{Type: golang.BasicLit},
			},
		}
	}

	return nodes
}

func mustDataDominatedSequence() []*syntax.Node {
	return []*syntax.Node{
		{
			Type:     golang.CompositeLit,
			Filename: "config_test.go",
			Children: mustKeyValueExprFields("Name", "Reason", "Severity", "Version", "Category", "Action"),
		},
	}
}

func mustLogicHeavySequence() []*syntax.Node {
	return []*syntax.Node{
		{
			Type:     golang.FuncDecl,
			Filename: "service.go",
			Children: []*syntax.Node{
				{Type: golang.FuncType},
				{
					Type: golang.BlockStmt,
					Children: []*syntax.Node{
						{Type: golang.IfStmt, Children: []*syntax.Node{
							{Type: golang.BinaryExpr},
							{Type: golang.BlockStmt, Children: []*syntax.Node{
								{Type: golang.ReturnStmt},
							}},
						}},
						{Type: golang.AssignStmt},
						{Type: golang.ReturnStmt},
					},
				},
			},
		},
	}
}

func mustMixedDataLogicSequence() []*syntax.Node {
	return []*syntax.Node{
		{
			Type:     golang.FuncDecl,
			Filename: "handler.go",
			Children: []*syntax.Node{
				{Type: golang.FuncType},
				{
					Type: golang.BlockStmt,
					Children: []*syntax.Node{
						{Type: golang.AssignStmt},
						{Type: golang.BasicLit},
						{Type: golang.IfStmt, Children: []*syntax.Node{
							{Type: golang.BinaryExpr},
						}},
						{Type: golang.BasicLit},
					},
				},
			},
		},
	}
}

func TestIsInterfaceImplementation(t *testing.T) {
	runBoolTests(t, isInterfaceImplementation, []boolTestCase{
		{
			name: "3+ FuncType fragments from different files",
			seqs: [][]*syntax.Node{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/json.go")},
				{mustFuncTypeNode("printer/html.go")},
			},
			expected: true,
		},
		{
			name: "6 FuncType fragments from different files",
			seqs: [][]*syntax.Node{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/json.go")},
				{mustFuncTypeNode("printer/html.go")},
				{mustFuncTypeNode("printer/plumbing.go")},
				{mustFuncTypeNode("printer/sarif.go")},
				{mustFuncTypeNode("printer/stats.go")},
			},
			expected: true,
		},
		{
			name: "only 2 fragments returns false",
			seqs: [][]*syntax.Node{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/json.go")},
			},
			expected: false,
		},
		{
			name: "3 fragments but same file",
			seqs: [][]*syntax.Node{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/text.go")},
			},
			expected: false,
		},
		{
			name: "non-FuncType root returns false",
			seqs: [][]*syntax.Node{
				{mustNodeWithFilename("printer/text.go")},
				{mustNodeWithFilename("printer/json.go")},
				{mustNodeWithFilename("printer/html.go")},
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*syntax.Node{}, expected: false},
		{
			name: "empty fragment returns false",
			seqs: [][]*syntax.Node{
				{mustFuncTypeNode("printer/text.go")},
				{},
				{mustFuncTypeNode("printer/html.go")},
			},
			expected: false,
		},
	})
}

func mustFuncTypeNode(filename string) *syntax.Node {
	return &syntax.Node{
		Type:     golang.FuncType,
		Filename: filename,
		Children: []*syntax.Node{
			{Type: golang.FieldList},
			{Type: golang.FieldList},
		},
	}
}

func mustThreeDistinctAssertions(filename string) *syntax.Node {
	return &syntax.Node{
		Type:     golang.ExprStmt,
		Filename: filename,
		Children: []*syntax.Node{
			mustSelectorExprStmt("", "Expect"),
			mustSelectorExprStmt("", "NotTo"),
			mustSelectorExprStmt("", "Equal"),
		},
	}
}
