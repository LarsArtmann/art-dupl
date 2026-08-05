package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIsAssignWithErrorCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name:     "assign + error check if",
			seqs:     [][]*domain.CloneNode{assignWithErrorCheckSeq()},
			expected: true,
		},
		{
			name: "two error check sequences",
			seqs: [][]*domain.CloneNode{
				assignWithErrorCheckSeq(),
				assignWithErrorCheckSeq(),
			},
			expected: true,
		},
		{
			name: "only assign without error check",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.AssignStmt},
				{BaseType: golang.AssignStmt},
			}},
			expected: false,
		},
		{
			name: "single statement (not 2)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt},
			}},
			expected: false,
		},
		{
			name: "three statements (not 2)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.AssignStmt},
				{BaseType: golang.IfStmt},
				{BaseType: golang.ReturnStmt},
			}},
			expected: false,
		},
		{
			name:     "empty (vacuous true, guarded by caller)",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isAssignWithErrorCheck(tt.seqs)
			if result != tt.expected {
				t.Errorf("isAssignWithErrorCheck() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsSingleCallExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name: "single CallExpr",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.CallExpr},
			}},
			expected: true,
		},
		{
			name: "ExprStmt wrapping CallExpr (statement-level t.Parallel())",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr, Children: []*domain.CloneNode{
						{BaseType: golang.Ident, Name: "t"},
					}},
				}},
			}},
			expected: true,
		},
		{
			name: "multiple ExprStmt(CallExpr) clones across files",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
				}}},
				{{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
				}}},
			},
			expected: true,
		},
		{
			name: "ExprStmt wrapping non-CallExpr (not lone call)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
				}},
			}},
			expected: false,
		},
		{
			name: "ExprStmt with multiple children (not lone call)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
					{BaseType: golang.CallExpr},
				}},
			}},
			expected: false,
		},
		{
			name: "two single CallExpr clones",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.CallExpr}},
				{{BaseType: golang.CallExpr}},
			},
			expected: true,
		},
		{
			name: "single IfStmt (not CallExpr)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt},
			}},
			expected: false,
		},
		{
			name: "two nodes (not single)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.CallExpr},
				{BaseType: golang.CallExpr},
			}},
			expected: false,
		},
		{
			name:     "empty (vacuous true, guarded by caller)",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isSingleCallExpression(tt.seqs)
			if result != tt.expected {
				t.Errorf("isSingleCallExpression() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func assignWithErrorCheckSeq() []*domain.CloneNode {
	return []*domain.CloneNode{
		{BaseType: golang.AssignStmt},
		{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
			{BaseType: golang.BinaryExpr, Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "nil"},
			}},
			{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
				{BaseType: golang.ReturnStmt},
			}},
		}},
	}
}

func TestIsSingleSimpleStatement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name:     "single ReturnStmt",
			seqs:     [][]*domain.CloneNode{{{BaseType: golang.ReturnStmt}}},
			expected: true,
		},
		{
			name: "single ReturnStmt with expression",
			seqs: [][]*domain.CloneNode{{
				{
					BaseType: golang.ReturnStmt,
					Children: []*domain.CloneNode{
						{BaseType: golang.Ident, Name: "nil"},
					},
				},
			}},
			expected: true,
		},
		{
			name:     "single AssignStmt",
			seqs:     [][]*domain.CloneNode{{{BaseType: golang.AssignStmt}}},
			expected: true,
		},
		{
			name:     "single IncDecStmt",
			seqs:     [][]*domain.CloneNode{{{BaseType: golang.IncDecStmt}}},
			expected: true,
		},
		{
			name:     "single BranchStmt (break)",
			seqs:     [][]*domain.CloneNode{{{BaseType: golang.BranchStmt}}},
			expected: true,
		},
		{
			name:     "single SendStmt",
			seqs:     [][]*domain.CloneNode{{{BaseType: golang.SendStmt}}},
			expected: true,
		},
		{
			name: "DeclStmt wrapping ValueSpec (var x int)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.DeclStmt, Children: []*domain.CloneNode{
					{BaseType: golang.GenDecl, Children: []*domain.CloneNode{
						{BaseType: golang.ValueSpec},
					}},
				}},
			}},
			expected: true,
		},
		{
			name: "DeclStmt wrapping TypeSpec (type Foo struct{}) — NOT filtered",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.DeclStmt, Children: []*domain.CloneNode{
					{BaseType: golang.GenDecl, Children: []*domain.CloneNode{
						{BaseType: golang.TypeSpec, Children: []*domain.CloneNode{
							{BaseType: golang.StructType},
						}},
					}},
				}},
			}},
			expected: false,
		},
		{
			name: "single IfStmt — not a terminal statement",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt},
			}},
			expected: false,
		},
		{
			name: "single ForStmt — not a terminal statement",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ForStmt},
			}},
			expected: false,
		},
		{
			name: "single CallExpr — not handled here (isSingleCallExpression)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.CallExpr},
			}},
			expected: false,
		},
		{
			name: "two statements — not single",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ReturnStmt},
				{BaseType: golang.ReturnStmt},
			}},
			expected: false,
		},
		{
			name:     "empty (vacuous true, guarded by caller)",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isSingleSimpleStatement(tt.seqs)
			if result != tt.expected {
				t.Errorf("isSingleSimpleStatement() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubtreeContainsTypeSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		node     *domain.CloneNode
		expected bool
	}{
		{
			name:     "direct TypeSpec",
			node:     &domain.CloneNode{BaseType: golang.TypeSpec},
			expected: true,
		},
		{
			name: "TypeSpec nested in GenDecl",
			node: &domain.CloneNode{BaseType: golang.DeclStmt, Children: []*domain.CloneNode{
				{BaseType: golang.GenDecl, Children: []*domain.CloneNode{
					{BaseType: golang.TypeSpec},
				}},
			}},
			expected: true,
		},
		{
			name: "no TypeSpec (ValueSpec only)",
			node: &domain.CloneNode{BaseType: golang.DeclStmt, Children: []*domain.CloneNode{
				{BaseType: golang.GenDecl, Children: []*domain.CloneNode{
					{BaseType: golang.ValueSpec},
				}},
			}},
			expected: false,
		},
		{
			name:     "bare ReturnStmt",
			node:     &domain.CloneNode{BaseType: golang.ReturnStmt},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := subtreeContainsTypeSpec(tt.node)
			if result != tt.expected {
				t.Errorf("subtreeContainsTypeSpec() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsSingleDeclaration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name: "ValueSpec const re-export (Foo = pkg.Foo)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.SelectorExpr},
				}},
			}},
			expected: true,
		},
		{
			name: "ValueSpec with BasicLit value (const N = 42)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.BasicLit},
				}},
			}},
			expected: true,
		},
		{
			name: "TypeSpec alias referencing named type (type Mode = domain.Mode)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.TypeSpec, IsAlias: true, Children: []*domain.CloneNode{
					{BaseType: golang.Ident},
					{BaseType: golang.SelectorExpr},
				}},
			}},
			expected: true,
		},
		{
			name: "TypeSpec named type definition (type Mode pkg.Mode) — NOT suppressed",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.TypeSpec, IsAlias: false, Children: []*domain.CloneNode{
					{BaseType: golang.Ident},
					{BaseType: golang.SelectorExpr},
				}},
			}},
			expected: false,
		},
		{
			name: "TypeSpec struct definition — NOT suppressed",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.TypeSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident},
					{BaseType: golang.StructType},
				}},
			}},
			expected: false,
		},
		{
			name: "TypeSpec interface definition — NOT suppressed",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.TypeSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident},
					{BaseType: golang.InterfaceType},
				}},
			}},
			expected: false,
		},
		{
			name: "single AssignStmt — not a declaration node",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.AssignStmt},
			}},
			expected: false,
		},
		{
			name: "two ValueSpecs — not single",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ValueSpec},
				{BaseType: golang.ValueSpec},
			}},
			expected: false,
		},
		{
			name: "bare GenDecl — not the spec itself",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.GenDecl},
			}},
			expected: false,
		},
		{
			name:     "empty (vacuous true, guarded by caller)",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isSingleDeclaration(tt.seqs)
			if result != tt.expected {
				t.Errorf("isSingleDeclaration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsInterfaceAssertion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name: "var _ I = (*T)(nil) — blank ident ValueSpec",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "_"},
					{BaseType: golang.Ident, Name: "Reader"},
					{BaseType: golang.CallExpr, Children: []*domain.CloneNode{
						{BaseType: golang.StarExpr, Children: []*domain.CloneNode{
							{BaseType: golang.Ident, Name: "MyReader"},
						}},
						{BaseType: golang.Ident, Name: "nil"},
					}},
				}},
			}},
			expected: true,
		},
		{
			name: "two blank-ident assertions (different types) — still match",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "_"},
				}}},
				{{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "_"},
				}}},
			},
			expected: true,
		},
		{
			name: "named ValueSpec — NOT suppressed",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "foo"},
				}},
			}},
			expected: false,
		},
		{
			name: "single AssignStmt — not a ValueSpec",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.AssignStmt},
			}},
			expected: false,
		},
		{
			name: "multi-node sequence — not single",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "_"},
				}},
				{BaseType: golang.ValueSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "_"},
				}},
			}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isInterfaceAssertion(tt.seqs)
			if result != tt.expected {
				t.Errorf("isInterfaceAssertion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsTypeAliasBlock(t *testing.T) {
	t.Parallel()

	pkgAlias := func(name, pkg, typ string) *domain.CloneNode {
		return &domain.CloneNode{
			BaseType: golang.TypeSpec,
			IsAlias:  true,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: name},
				{BaseType: golang.SelectorExpr, Name: pkg + "." + typ, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: pkg},
					{BaseType: golang.Ident, Name: typ},
				}},
			},
		}
	}

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name: "2-alias block (type ToolCall/ToolResult = pkg.ToolCall/ToolResult)",
			seqs: [][]*domain.CloneNode{{
				pkgAlias("ToolCall", "protocoltypes", "ToolCall"),
				pkgAlias("ToolResult", "protocoltypes", "ToolResult"),
			}},
			expected: true,
		},
		{
			name: "3-alias block — still matches",
			seqs: [][]*domain.CloneNode{{
				pkgAlias("A", "pkg", "A"),
				pkgAlias("B", "pkg", "B"),
				pkgAlias("C", "pkg", "C"),
			}},
			expected: true,
		},
		{
			name: "single alias — NOT matched (handled by single-declaration)",
			seqs: [][]*domain.CloneNode{{
				pkgAlias("Mode", "domain", "Mode"),
			}},
			expected: false,
		},
		{
			name: "block with one composite type — NOT suppressed",
			seqs: [][]*domain.CloneNode{{
				pkgAlias("A", "pkg", "A"),
				{BaseType: golang.TypeSpec, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "Config"},
					{BaseType: golang.StructType},
				}},
			}},
			expected: false,
		},
		{
			name: "block with non-alias TypeSpec (type X string) — NOT suppressed",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.TypeSpec, IsAlias: false, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "Severity"},
					{BaseType: golang.Ident, Name: "string"},
				}},
				{BaseType: golang.TypeSpec, IsAlias: false, Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "Priority"},
					{BaseType: golang.Ident, Name: "string"},
				}},
			}},
			expected: false,
		},
		{
			name:     "empty (vacuous true, guarded by caller)",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isTypeAliasBlock(tt.seqs)
			if result != tt.expected {
				t.Errorf("isTypeAliasBlock() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsTestHelperDelegate(t *testing.T) {
	t.Parallel()

	helperDelegate := func(receiver string) []*domain.CloneNode {
		return []*domain.CloneNode{
			{
				BaseType: golang.ExprStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.CallExpr,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.SelectorExpr,
								Name:     "Helper",
								Children: []*domain.CloneNode{
									{BaseType: golang.Ident, Name: receiver},
									{BaseType: golang.Ident, Name: "Helper"},
								},
							},
						},
					},
				},
			},
			{
				BaseType: golang.ExprStmt,
				Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
				},
			},
		}
	}

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name:     "t.Helper() + delegate (testing.T)",
			seqs:     [][]*domain.CloneNode{helperDelegate("t")},
			expected: true,
		},
		{
			name:     "b.Helper() + delegate (testing.B)",
			seqs:     [][]*domain.CloneNode{helperDelegate("b")},
			expected: true,
		},
		{
			name:     "tb.Helper() + delegate (testing.TB)",
			seqs:     [][]*domain.CloneNode{helperDelegate("tb")},
			expected: true,
		},
		{
			name: "not a helper call — first stmt is not .Helper()",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr, Children: []*domain.CloneNode{
						{BaseType: golang.SelectorExpr, Name: "Setup", Children: []*domain.CloneNode{
							{BaseType: golang.Ident, Name: "t"},
							{BaseType: golang.Ident, Name: "Setup"},
						}},
					}},
				}},
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
				}},
			}},
			expected: false,
		},
		{
			name: "3-statement body — too many statements",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt},
				{BaseType: golang.AssignStmt},
				{BaseType: golang.ReturnStmt},
			}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isTestHelperDelegate(tt.seqs)
			if result != tt.expected {
				t.Errorf("isTestHelperDelegate() = %v, want %v", result, tt.expected)
			}
		})
	}
}
