package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestClassifyClone(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		nodeType       int32
		tokens         int
		lines          int
		wantCategory   CloneCategory
		wantIsTest     bool
		wantPriority   ClonePriority
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
			wantSuggestion: suggestReviewExtract,
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
			wantSuggestion: suggestReviewExtract,
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
			wantSuggestion: suggestReviewExtract,
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

func TestCloneCategoryEmoji(t *testing.T) {
	tests := []struct {
		category CloneCategory
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
	getValue func(ClonePriority) string,
	expected map[ClonePriority]string,
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
		getValue func(ClonePriority) string
		expected map[ClonePriority]string
	}{
		{
			name:     "GetPriorityEmoji",
			getValue: func(p ClonePriority) string { return p.GetPriorityEmoji() },
			expected: map[ClonePriority]string{
				domain.PriorityCritical: "🔴",
				domain.PriorityHigh:     "🟠",
				domain.PriorityMedium:   "🟡",
				domain.PriorityLow:      "🟢",
			},
		},
		{
			name:     "GetPriorityColor",
			getValue: func(p ClonePriority) string { return p.GetPriorityColor() },
			expected: map[ClonePriority]string{
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
