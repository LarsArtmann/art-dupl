package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

type boolTestCase struct {
	name     string
	seqs     [][]*domain.CloneNode
	expected bool
}

func runBoolTests(t *testing.T, fn func([][]*domain.CloneNode) bool, cases []boolTestCase) {
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
			seqs: [][]*domain.CloneNode{
				{mustNodeWithFilename("pkg/testdata/input.go")},
				{mustNodeWithFilename("pkg/testdata/golden.go")},
			},
			expected: true,
		},
		{
			name: "nested testdata directory",
			seqs: [][]*domain.CloneNode{
				{mustNodeWithFilename("internal/rule/testdata/errors_as/golden.go")},
				{mustNodeWithFilename("internal/rule/testdata/errors_as/input.go")},
			},
			expected: true,
		},
		{
			name: "one file not in testdata",
			seqs: [][]*domain.CloneNode{
				{mustNodeWithFilename("pkg/testdata/input.go")},
				{mustNodeWithFilename("pkg/rule.go")},
			},
			expected: false,
		},
		{
			name: "production files only",
			seqs: [][]*domain.CloneNode{
				{mustNodeWithFilename("pkg/handler.go")},
				{mustNodeWithFilename("pkg/service.go")},
			},
			expected: false,
		},
		{name: "empty sequence", seqs: [][]*domain.CloneNode{}, expected: false},
		{
			name: "file named testdata.go but not in testdata dir",
			seqs: [][]*domain.CloneNode{
				{mustNodeWithFilename("pkg/testdata.go")},
				{mustNodeWithFilename("pkg/testdata_test.go")},
			},
			expected: false,
		},
		{
			name: "empty node in sequence",
			seqs: [][]*domain.CloneNode{
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
			seqs: [][]*domain.CloneNode{
				{mustRangeStmtWithTRun("resolve_test.go")},
				{mustRangeStmtWithTRun("filter_test.go")},
			},
			expected: true,
		},
		{
			name: "RangeStmt without t.Run in test file",
			seqs: [][]*domain.CloneNode{
				{mustRangeStmtNoTRun("resolve_test.go")},
				{mustRangeStmtNoTRun("filter_test.go")},
			},
			expected: false,
		},
		{
			name: "RangeStmt with t.Run but not test file",
			seqs: [][]*domain.CloneNode{
				{mustRangeStmtWithTRun("resolve.go")},
				{mustRangeStmtWithTRun("filter.go")},
			},
			expected: false,
		},
		{
			name: "non-RangeStmt in test file",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.FuncDecl, Filename: "resolve_test.go"}},
				{{BaseType: golang.FuncDecl, Filename: "filter_test.go"}},
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*domain.CloneNode{}, expected: false},
		{
			name: "mixed: one RangeStmt one not",
			seqs: [][]*domain.CloneNode{
				{mustRangeStmtWithTRun("resolve_test.go")},
				{{BaseType: golang.FuncDecl, Filename: "filter_test.go"}},
			},
			expected: false,
		},
	})
}

func TestIsTestScaffolding(t *testing.T) {
	runBoolTests(t, isTestScaffolding, []boolTestCase{
		{
			name: "Ginkgo TempDir + WriteFile + Expect pattern",
			seqs: [][]*domain.CloneNode{
				{mustTestScaffolding("benchmark_naming_rule_test.go")},
				{mustTestScaffolding("testdata_directory_rule_test.go")},
			},
			expected: true,
		},
		{
			name: "TempDir without assertions",
			seqs: [][]*domain.CloneNode{
				{mustTempDirOnly("rule_test.go")},
				{mustTempDirOnly("handler_test.go")},
			},
			expected: false,
		},
		{
			name: "Assertions without TempDir",
			seqs: [][]*domain.CloneNode{
				{mustAssertionsOnly("rule_test.go")},
				{mustAssertionsOnly("handler_test.go")},
			},
			expected: false,
		},
		{
			name: "3+ distinct assertions without TempDir is scaffolding",
			seqs: [][]*domain.CloneNode{
				{mustThreeDistinctAssertions("rule_test.go")},
				{mustThreeDistinctAssertions("handler_test.go")},
			},
			expected: true,
		},
		{
			name: "production files with similar pattern",
			seqs: [][]*domain.CloneNode{
				{mustTestScaffolding("handler.go")},
				{mustTestScaffolding("service.go")},
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*domain.CloneNode{}, expected: false},
	})
}

func TestIsDataDominated(t *testing.T) {
	runBoolTests(t, isDataDominated, []boolTestCase{
		{
			name: "struct init with many BasicLit values",
			seqs: [][]*domain.CloneNode{
				mustDataDominatedSequence(),
				mustDataDominatedSequence(),
			},
			expected: true,
		},
		{
			name: "logic-heavy sequence",
			seqs: [][]*domain.CloneNode{
				mustLogicHeavySequence(),
				mustLogicHeavySequence(),
			},
			expected: false,
		},
		{
			name: "mixed data and logic",
			seqs: [][]*domain.CloneNode{
				mustMixedDataLogicSequence(),
				mustMixedDataLogicSequence(),
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*domain.CloneNode{}, expected: false},
		{
			name: "empty node list in sequence",
			seqs: [][]*domain.CloneNode{
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
		seqs           [][]*domain.CloneNode
		expectedLabel  PatternLabel
		expectedAction domain.CloneActionability
	}{
		{
			name: "testdata pair gets correct label",
			seqs: [][]*domain.CloneNode{
				{mustNodeWithFilename("pkg/testdata/input.go")},
				{mustNodeWithFilename("pkg/testdata/golden.go")},
			},
			expectedLabel:  PatternTestData,
			expectedAction: domain.NonActionable,
		},
		{
			name: "table-driven test gets correct label",
			seqs: [][]*domain.CloneNode{
				{mustRangeStmtWithTRun("resolve_test.go")},
				{mustRangeStmtWithTRun("filter_test.go")},
			},
			expectedLabel:  PatternTableDrivenTest,
			expectedAction: domain.NonActionable,
		},
		{
			name: "test scaffolding gets correct label",
			seqs: [][]*domain.CloneNode{
				{mustTestScaffolding("rule_test.go")},
				{mustTestScaffolding("handler_test.go")},
			},
			expectedLabel:  PatternTestScaffolding,
			expectedAction: domain.NonActionable,
		},
		{
			name: "RAII defer gets correct label",
			seqs: [][]*domain.CloneNode{
				{mustDeferSelectorCall("mu", cleanupMethodName)},
				{mustDeferSelectorCall("mu", cleanupMethodName)},
			},
			expectedLabel:  PatternRAIIDefer,
			expectedAction: domain.NonActionable,
		},
		{
			name: "error propagation gets correct label",
			seqs: [][]*domain.CloneNode{
				{mustIfErrReturnNil()},
				{mustIfErrReturnNil()},
			},
			expectedLabel:  PatternErrorPropagation,
			expectedAction: domain.NonActionable,
		},
		{
			name: "signature-only gets correct label",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.FuncDecl}},
				{{BaseType: golang.FuncDecl}},
			},
			expectedLabel:  PatternSignatureOnly,
			expectedAction: domain.NonActionable,
		},
		{
			name: "interface implementation gets correct label",
			seqs: [][]*domain.CloneNode{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/json.go")},
				{mustFuncTypeNode("printer/html.go")},
			},
			expectedLabel:  PatternInterfaceImpl,
			expectedAction: domain.NonActionable,
		},
		{
			name: "actionable code gets no label",
			seqs: [][]*domain.CloneNode{
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
			result := ApplyPatternLabel(tc.input, tc.label)
			if result.Category != tc.wantCategory {
				t.Errorf("category = %q, want %q", result.Category, tc.wantCategory)
			}

			if result.Priority != tc.wantPriority {
				t.Errorf("priority = %q, want %q", result.Priority, tc.wantPriority)
			}
		})
	}
}

func mustNodeWithFilename(filename string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.FuncDecl,
		Filename: filename,
		Children: []*domain.CloneNode{
			{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
				{BaseType: golang.ExprStmt},
			}},
		},
	}
}

func mustRangeStmtWithTRun(filename string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.RangeStmt,
		Filename: filename,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ExprStmt,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{
										BaseType: golang.SelectorExpr,
										Name:     "Run",
										Children: []*domain.CloneNode{
											{BaseType: golang.Ident, Name: "t"},
											{BaseType: golang.Ident, Name: "Run"},
										},
									},
									{
										BaseType: golang.BasicLit,
									},
									{
										BaseType: golang.FuncLit,
										Children: []*domain.CloneNode{
											{BaseType: golang.FuncType},
											{
												BaseType: golang.BlockStmt,
												Children: []*domain.CloneNode{
													{BaseType: golang.ExprStmt},
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

func mustRangeStmtNoTRun(filename string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.RangeStmt,
		Filename: filename,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					mustSelectorExprStmt("", "Process"),
				},
			},
		},
	}
}

func mustTestScaffolding(filename string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.ExprStmt,
		Filename: filename,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.AssignStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.CallExpr,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.SelectorExpr,
								Name:     "TempDir",
								Children: []*domain.CloneNode{
									{BaseType: golang.CallExpr, Children: []*domain.CloneNode{
										{BaseType: golang.SelectorExpr, Name: "GinkgoT"},
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

func mustSelectorExprStmt(filename, selectorName string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.ExprStmt,
		Filename: filename,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.SelectorExpr,
						Name:     selectorName,
					},
				},
			},
		},
	}
}

func mustTempDirOnly(filename string) *domain.CloneNode {
	return mustSelectorExprStmt(filename, "TempDir")
}

func mustAssertionsOnly(filename string) *domain.CloneNode {
	return mustSelectorExprStmt(filename, "Equal")
}

func mustKeyValueExprFields(names ...string) []*domain.CloneNode {
	nodes := make([]*domain.CloneNode, 0, len(names))
	for _, name := range names {
		nodes = append(nodes, &domain.CloneNode{
			BaseType: golang.KeyValueExpr,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: name},
				{BaseType: golang.BasicLit},
			},
		})
	}

	return nodes
}

func mustDataDominatedSequence() []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType: golang.CompositeLit,
			Filename: "config_test.go",
			Children: mustKeyValueExprFields("Name", "Reason", "Severity", "Version", "Category", "Action"),
		},
	}
}

func mustLogicHeavySequence() []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType: golang.FuncDecl,
			Filename: "service.go",
			Children: []*domain.CloneNode{
				{BaseType: golang.FuncType},
				{
					BaseType: golang.BlockStmt,
					Children: []*domain.CloneNode{
						{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
							{BaseType: golang.BinaryExpr},
							{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
								{BaseType: golang.ReturnStmt},
							}},
						}},
						{BaseType: golang.AssignStmt},
						{BaseType: golang.ReturnStmt},
					},
				},
			},
		},
	}
}

func mustMixedDataLogicSequence() []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType: golang.FuncDecl,
			Filename: "handler.go",
			Children: []*domain.CloneNode{
				{BaseType: golang.FuncType},
				{
					BaseType: golang.BlockStmt,
					Children: []*domain.CloneNode{
						{BaseType: golang.AssignStmt},
						{BaseType: golang.BasicLit},
						{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
							{BaseType: golang.BinaryExpr},
						}},
						{BaseType: golang.BasicLit},
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
			seqs: [][]*domain.CloneNode{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/json.go")},
				{mustFuncTypeNode("printer/html.go")},
			},
			expected: true,
		},
		{
			name: "6 FuncType fragments from different files",
			seqs: [][]*domain.CloneNode{
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
			seqs: [][]*domain.CloneNode{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/json.go")},
			},
			expected: false,
		},
		{
			name: "3 fragments but same file",
			seqs: [][]*domain.CloneNode{
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/text.go")},
				{mustFuncTypeNode("printer/text.go")},
			},
			expected: false,
		},
		{
			name: "non-FuncType root returns false",
			seqs: [][]*domain.CloneNode{
				{mustNodeWithFilename("printer/text.go")},
				{mustNodeWithFilename("printer/json.go")},
				{mustNodeWithFilename("printer/html.go")},
			},
			expected: false,
		},
		{name: "empty sequences", seqs: [][]*domain.CloneNode{}, expected: false},
		{
			name: "empty fragment returns false",
			seqs: [][]*domain.CloneNode{
				{mustFuncTypeNode("printer/text.go")},
				{},
				{mustFuncTypeNode("printer/html.go")},
			},
			expected: false,
		},
	})
}

func mustFuncTypeNode(filename string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.FuncType,
		Filename: filename,
		Children: []*domain.CloneNode{
			{BaseType: golang.FieldList},
			{BaseType: golang.FieldList},
		},
	}
}

func mustThreeDistinctAssertions(filename string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.ExprStmt,
		Filename: filename,
		Children: []*domain.CloneNode{
			mustSelectorExprStmt("", "Expect"),
			mustSelectorExprStmt("", "NotTo"),
			mustSelectorExprStmt("", "Equal"),
		},
	}
}

func TestIsGuardClause(t *testing.T) {
	runBoolTests(t, isGuardClause, []boolTestCase{
		{
			name: "if cond { return }",
			seqs: [][]*domain.CloneNode{{
				mustGuardClauseIf(golang.UnaryExpr, mustReturnNil()),
			}},
			expected: true,
		},
		{
			name: "if cond { return value }",
			seqs: [][]*domain.CloneNode{{
				mustGuardClauseIf(golang.BinaryExpr, &domain.CloneNode{
					BaseType: golang.ReturnStmt,
					Children: []*domain.CloneNode{{BaseType: golang.Ident, Name: "err"}},
				}),
			}},
			expected: true,
		},
		{
			name: "if cond { return a, b } (2 returns)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
					{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
						{BaseType: golang.ReturnStmt},
						{BaseType: golang.ReturnStmt},
					}},
				}},
			}},
			expected: true,
		},
		{
			name: "if cond { log(); return } — not guard (non-return in body)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
					{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
						{BaseType: golang.ExprStmt},
						{BaseType: golang.ReturnStmt},
					}},
				}},
			}},
			expected: false,
		},
		{
			name: "if cond { } else { return } — has else",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
					{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
						{BaseType: golang.ReturnStmt},
					}},
					{BaseType: golang.BlockStmt},
				}},
			}},
			expected: false,
		},
		{
			name: "if cond { return } else if { return } — else-if chain",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
					{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
						{BaseType: golang.ReturnStmt},
					}},
					{BaseType: golang.IfStmt},
				}},
			}},
			expected: false,
		},
		{
			name: "if cond { 3+ returns } — too many",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
					{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
						{BaseType: golang.ReturnStmt},
						{BaseType: golang.ReturnStmt},
						{BaseType: golang.ReturnStmt},
					}},
				}},
			}},
			expected: false,
		},
		{
			name: "if cond { } — empty body",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
					{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{}},
				}},
			}},
			expected: false,
		},
		{
			name: "not an IfStmt (ReturnStmt)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ReturnStmt},
			}},
			expected: false,
		},
		{
			name: "two statements (not single)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt},
				{BaseType: golang.IfStmt},
			}},
			expected: false,
		},
		{name: "empty (vacuous true)", seqs: [][]*domain.CloneNode{}, expected: true},
	})
}

func mustGuardClauseIf(_ int32, ret *domain.CloneNode) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.IfStmt,
		Children: []*domain.CloneNode{
			{BaseType: golang.BinaryExpr, Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "nil"},
			}},
			{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{ret}},
		},
	}
}

func mustReturnNil() *domain.CloneNode {
	return &domain.CloneNode{BaseType: golang.ReturnStmt}
}
