package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templpkg "github.com/LarsArtmann/art-dupl/syntax/templ"
)

func TestClassifyClone(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		nodeType       int32
		tokens         int
		lines          int
		wantCategory   domain.CloneCategory
		wantIsTest     bool
		wantPriority   domain.ClonePriority
		wantSuggestion string
	}{
		{
			name:           "function declaration in production code - medium",
			filename:       "handler.go",
			nodeType:       golang.FuncDecl,
			tokens:         12,
			lines:          10,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     false,
			wantPriority:   domain.PriorityHigh, // tokens > 8
			wantSuggestion: suggestExtractUtility,
		},
		{
			name:           "function in test file",
			filename:       "handler_test.go",
			nodeType:       golang.FuncDecl,
			tokens:         12,
			lines:          12,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     true,
			wantPriority:   domain.PriorityLow,
			wantSuggestion: "Consider extracting to shared test utility",
		},
		{
			name:           "method (function literal) in production - large",
			filename:       "service.go",
			nodeType:       golang.FuncLit,
			tokens:         20,
			lines:          20,
			wantCategory:   domain.CategoryMethod,
			wantIsTest:     false,
			wantPriority:   domain.PriorityCritical, // tokens > 15
			wantSuggestion: suggestExtractUtility,
		},
		{
			name:           "struct type in production - small",
			filename:       "types.go",
			nodeType:       golang.StructType,
			tokens:         8,
			lines:          8,
			wantCategory:   domain.CategoryStruct,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens <= 10
			wantSuggestion: "Consider composition or shared base struct",
		},
		{
			name:           "interface type - large",
			filename:       "interfaces.go",
			nodeType:       golang.InterfaceType,
			tokens:         12,
			lines:          12,
			wantCategory:   domain.CategoryInterface,
			wantIsTest:     false,
			wantPriority:   domain.PriorityHigh, // tokens > 10
			wantSuggestion: "Extract common interface definition",
		},
		{
			name:           "for loop - large",
			filename:       "processor.go",
			nodeType:       golang.ForStmt,
			tokens:         12,
			lines:          15,
			wantCategory:   domain.CategoryLoop,
			wantIsTest:     false,
			wantPriority:   domain.PriorityHigh, // tokens > 10
			wantSuggestion: "Extract loop body to helper function",
		},
		{
			name:           "range loop - small",
			filename:       "iterator.go",
			nodeType:       golang.RangeStmt,
			tokens:         8,
			lines:          8,
			wantCategory:   domain.CategoryLoop,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens <= 10
			wantSuggestion: "Extract loop body to helper function",
		},
		{
			name:           "if statement - small",
			filename:       "validator.go",
			nodeType:       golang.IfStmt,
			tokens:         8,
			lines:          10,
			wantCategory:   domain.CategoryConditional,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens <= 10
			wantSuggestion: "Consider strategy pattern or early returns",
		},
		{
			name:           "switch statement - large",
			filename:       "router.go",
			nodeType:       golang.SwitchStmt,
			tokens:         12,
			lines:          15,
			wantCategory:   domain.CategoryConditional,
			wantIsTest:     false,
			wantPriority:   domain.PriorityHigh, // tokens > 30
			wantSuggestion: "Consider strategy pattern or early returns",
		},
		{
			name:           "assignment statement - small",
			filename:       "main.go",
			nodeType:       golang.AssignStmt,
			tokens:         5,
			lines:          3,
			wantCategory:   domain.CategoryAssignment,
			wantIsTest:     false,
			wantPriority:   domain.PriorityLow, // tokens <= 8
			wantSuggestion: suggestAssignment,
		},
		{
			name:           "call expression - mapped to call category",
			filename:       "caller.go",
			nodeType:       golang.CallExpr,
			tokens:         5,
			lines:          5,
			wantCategory:   domain.CategoryCall,
			wantIsTest:     false,
			wantPriority:   domain.PriorityLow, // tokens <= 8
			wantSuggestion: suggestCall,
		},
		{
			name:           "unknown node type",
			filename:       "broken.go",
			nodeType:       golang.BadNode,
			tokens:         5,
			lines:          6,
			wantCategory:   domain.CategoryUnknown,
			wantIsTest:     false,
			wantPriority:   domain.PriorityLow, // tokens < 25
			wantSuggestion: suggestReviewExtract,
		},
		{
			name:           "handler in production - large",
			filename:       "api.go",
			nodeType:       golang.FuncDecl,
			tokens:         20,
			lines:          40,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     false,
			wantPriority:   domain.PriorityCritical,
			wantSuggestion: suggestExtractUtility,
		},
		{
			name:           "test file with large duplication",
			filename:       "large_test.go",
			nodeType:       golang.FuncDecl,
			tokens:         35,
			lines:          35,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     true,
			wantPriority:   domain.PriorityMedium, // test file with tokens > 30
			wantSuggestion: "Extract test helper function or use table-driven tests",
		},
		{
			name:           "function in production - small",
			filename:       "utils.go",
			nodeType:       golang.FuncDecl,
			tokens:         5,
			lines:          8,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens <= 8
			wantSuggestion: suggestExtractUtility,
		},
		{
			name:           "gen decl maps to expression",
			filename:       "consts.go",
			nodeType:       golang.GenDecl,
			tokens:         10,
			lines:          10,
			wantCategory:   domain.CategoryExpression,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens 8-15
			wantSuggestion: suggestExpression,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyClone(domain.ClassificationInput{
				Filename: tt.filename,
				NodeType: tt.nodeType,
				Tokens:   tt.tokens,
				Lines:    tt.lines,
			})

			if got.Category != tt.wantCategory {
				t.Errorf("ClassifyClone().Category = %v, want %v", got.Category, tt.wantCategory)
			}

			if got.IsTest != tt.wantIsTest {
				t.Errorf("ClassifyClone().IsTest = %v, want %v", got.IsTest, tt.wantIsTest)
			}

			if got.Priority != tt.wantPriority {
				t.Errorf("ClassifyClone().Priority = %v, want %v", got.Priority, tt.wantPriority)
			}

			if got.Suggestion != tt.wantSuggestion {
				t.Errorf(
					"ClassifyClone().Suggestion = %v, want %v",
					got.Suggestion,
					tt.wantSuggestion,
				)
			}
		})
	}
}

func TestTemplNodeTypeToCategory(t *testing.T) {
	tests := []struct {
		name         string
		nodeType     int32
		wantCategory domain.CloneCategory
	}{
		// Declarations are functions.
		{"component declaration", templpkg.ComponentDeclaration, domain.CategoryFunction},
		{"css declaration", templpkg.CSSDeclaration, domain.CategoryFunction},
		{"script declaration", templpkg.ScriptDeclaration, domain.CategoryFunction},
		// Flow control mirrors the Go conditional category.
		{"if statement", templpkg.ComponentIfStatement, domain.CategoryConditional},
		{"switch statement", templpkg.ComponentSwitchStatement, domain.CategoryConditional},
		{"switch expression case", templpkg.ComponentSwitchExpressionCase, domain.CategoryConditional},
		{"switch default case", templpkg.ComponentSwitchDefaultCase, domain.CategoryConditional},
		{"fallthrough", templpkg.ComponentFallthroughStatement, domain.CategoryConditional},
		{"else container", templpkg.ComponentElseStatement, domain.CategoryConditional},
		{"conditional attribute", templpkg.ConditionalAttributeIfStatement, domain.CategoryConditional},
		{"for statement", templpkg.ComponentForStatement, domain.CategoryLoop},
		// Component calls are calls.
		{"component render", templpkg.ComponentRender, domain.CategoryCall},
		// Expressions.
		{"attribute", templpkg.Attribute, domain.CategoryExpression},
		{"spread attributes", templpkg.SpreadAttributes, domain.CategoryExpression},
		{"expression", templpkg.Expression, domain.CategoryExpression},
		{"children expression", templpkg.ComponentChildrenExpression, domain.CategoryExpression},
		{"raw go block", templpkg.RawGoBlock, domain.CategoryExpression},
		{"component import", templpkg.ComponentImport, domain.CategoryExpression},
		// Markup blocks.
		{"element", templpkg.Element, domain.CategoryBlock},
		{"style element", templpkg.StyleElement, domain.CategoryBlock},
		{"script element", templpkg.ScriptElement, domain.CategoryBlock},
		{"component block", templpkg.ComponentBlock, domain.CategoryBlock},
		{"doctype", templpkg.Doctype, domain.CategoryBlock},
		{"tag start", templpkg.TagStart, domain.CategoryBlock},
		{"tag end", templpkg.TagEnd, domain.CategoryBlock},
		{"self-closing tag", templpkg.SelfClosingTag, domain.CategoryBlock},
		// Unknowns.
		{"bad node", templpkg.BadNode, domain.CategoryUnknown},
		{"file root", templpkg.File, domain.CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templNodeTypeToCategory(tt.nodeType)
			if got != tt.wantCategory {
				t.Errorf("templNodeTypeToCategory(%d) = %v, want %v", tt.nodeType, got, tt.wantCategory)
			}
		})
	}
}

// TestNodeTypeToCategoryFileKindCollision pins the hazard that forced the
// table split: the golang and templ enums overlap numerically (both are
// parallel iota sequences starting at 0), so the SAME raw value must classify
// differently depending on the file kind. golang.BasicLit and templ.Element
// both equal 4; a merged switch could only ever serve one of them.
func TestNodeTypeToCategoryFileKindCollision(t *testing.T) {
	if golang.BasicLit != templpkg.Element {
		t.Fatalf("pin stale: golang.BasicLit (%d) != templ.Element (%d)",
			golang.BasicLit, templpkg.Element)
	}

	goGot := goNodeTypeToCategory(golang.BasicLit)
	if goGot != domain.CategoryExpression {
		t.Errorf("goNodeTypeToCategory(golang.BasicLit) = %v, want %v", goGot, domain.CategoryExpression)
	}

	templGot := nodeTypeToCategory(templpkg.Element, true)
	if templGot != domain.CategoryBlock {
		t.Errorf("nodeTypeToCategory(templ.Element, templ) = %v, want %v", templGot, domain.CategoryBlock)
	}

	sameValueViaGoTable := nodeTypeToCategory(golang.BasicLit, false)
	if sameValueViaGoTable != domain.CategoryExpression {
		t.Errorf("nodeTypeToCategory(golang.BasicLit, go) = %v, want %v",
			sameValueViaGoTable, domain.CategoryExpression)
	}
}

func TestTemplSuggestion(t *testing.T) {
	goFallback := "Review and extract common logic"

	tests := []struct {
		name     string
		category domain.CloneCategory
		want     string
	}{
		{
			name:     "component declaration",
			category: domain.CategoryFunction,
			want:     "Extract to a shared templ component",
		},
		{
			name:     "component render",
			category: domain.CategoryCall,
			want:     "Parameterize the component or extract shared markup to a child component",
		},
		{
			name:     "if statement",
			category: domain.CategoryConditional,
			want:     "Extract the shared markup block to a child templ component",
		},
		{
			name:     "for statement",
			category: domain.CategoryLoop,
			want:     "Extract the shared markup block to a child templ component",
		},
		{
			name:     "element",
			category: domain.CategoryBlock,
			want:     "Extract the shared markup to a child templ component",
		},
		{
			name:     "expression keeps go wording",
			category: domain.CategoryExpression,
			want:     goFallback,
		},
		{
			name:     "unknown keeps go wording",
			category: domain.CategoryUnknown,
			want:     goFallback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templSuggestion(tt.category, goFallback)
			if got != tt.want {
				t.Errorf("templSuggestion(%v) = %q, want %q", tt.category, got, tt.want)
			}
		})
	}
}

func TestCloneCategoryEmoji(t *testing.T) {
	tests := []struct {
		category domain.CloneCategory
		want     string
	}{
		{domain.CategoryFunction, "⚡"},
		{domain.CategoryMethod, "🔧"},
		{domain.CategoryTest, "🧪"},
		{domain.CategoryStruct, "📦"},
		{domain.CategoryInterface, "🔌"},
		{domain.CategoryHandler, "🎨"},
		{domain.CategoryLoop, "🔄"},
		{domain.CategoryConditional, "🔀"},
		{domain.CategoryAssignment, "📝"},
		{domain.CategoryExpression, "📊"},
		{domain.CategoryUnknown, "📄"},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			if got := tt.category.GetCategoryEmoji(); got != tt.want {
				t.Errorf("GetCategoryEmoji() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testClonePriority(
	t *testing.T,
	name string,
	getValue func(domain.ClonePriority) string,
	expected map[domain.ClonePriority]string,
) {
	t.Helper()

	for priority, want := range expected {
		t.Run(name+"_"+string(priority), func(t *testing.T) {
			if got := getValue(priority); got != want {
				t.Errorf("%s() = %v, want %v", name, got, want)
			}
		})
	}
}

func TestClonePriorityAccessors(t *testing.T) {
	tests := []struct {
		name     string
		getValue func(domain.ClonePriority) string
		expected map[domain.ClonePriority]string
	}{
		{
			name:     "GetPriorityEmoji",
			getValue: func(p domain.ClonePriority) string { return p.GetPriorityEmoji() },
			expected: map[domain.ClonePriority]string{
				domain.PriorityCritical: "🔴",
				domain.PriorityHigh:     "🟠",
				domain.PriorityMedium:   "🟡",
				domain.PriorityLow:      "🟢",
			},
		},
		{
			name:     "GetPriorityColor",
			getValue: func(p domain.ClonePriority) string { return p.GetPriorityColor() },
			expected: map[domain.ClonePriority]string{
				domain.PriorityCritical: "var(--error)",
				domain.PriorityHigh:     "var(--warning)",
				domain.PriorityMedium:   "var(--accent)",
				domain.PriorityLow:      "var(--success)",
			},
		},
	}

	for _, tt := range tests {
		testClonePriority(t, tt.name, tt.getValue, tt.expected)
	}
}
