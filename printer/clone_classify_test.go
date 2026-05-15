package printer

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
			tokens:         40,
			lines:          10,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     false,
			wantPriority:   domain.PriorityHigh, // tokens > 25
			wantSuggestion: "Extract to shared utility function",
		},
		{
			name:           "function in test file",
			filename:       "handler_test.go",
			nodeType:       golang.FuncDecl,
			tokens:         45,
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
			tokens:         95,
			lines:          20,
			wantCategory:   domain.CategoryMethod,
			wantIsTest:     false,
			wantPriority:   domain.PriorityCritical, // tokens > 50
			wantSuggestion: "Extract to shared utility function",
		},
		{
			name:           "struct type in production - small",
			filename:       "types.go",
			nodeType:       golang.StructType,
			tokens:         29,
			lines:          8,
			wantCategory:   domain.CategoryStruct,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens < 30
			wantSuggestion: "Consider composition or shared base struct",
		},
		{
			name:           "interface type - large",
			filename:       "interfaces.go",
			nodeType:       golang.InterfaceType,
			tokens:         39,
			lines:          12,
			wantCategory:   domain.CategoryInterface,
			wantIsTest:     false,
			wantPriority:   domain.PriorityHigh, // tokens > 30
			wantSuggestion: "Extract common interface definition",
		},
		{
			name:           "for loop - large",
			filename:       "processor.go",
			nodeType:       golang.ForStmt,
			tokens:         40,
			lines:          15,
			wantCategory:   domain.CategoryLoop,
			wantIsTest:     false,
			wantPriority:   domain.PriorityHigh, // tokens > 30
			wantSuggestion: "Extract loop body to helper function",
		},
		{
			name:           "range loop - small",
			filename:       "iterator.go",
			nodeType:       golang.RangeStmt,
			tokens:         25,
			lines:          8,
			wantCategory:   domain.CategoryLoop,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens < 30
			wantSuggestion: "Extract loop body to helper function",
		},
		{
			name:           "if statement - small",
			filename:       "validator.go",
			nodeType:       golang.IfStmt,
			tokens:         25,
			lines:          10,
			wantCategory:   domain.CategoryConditional,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens < 30
			wantSuggestion: "Consider strategy pattern or early returns",
		},
		{
			name:           "switch statement - large",
			filename:       "router.go",
			nodeType:       golang.SwitchStmt,
			tokens:         50,
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
			tokens:         9,
			lines:          3,
			wantCategory:   domain.CategoryAssignment,
			wantIsTest:     false,
			wantPriority:   domain.PriorityLow, // tokens < 25
			wantSuggestion: "Review and extract common logic",
		},
		{
			name:           "call expression - unmapped node type",
			filename:       "caller.go",
			nodeType:       golang.CallExpr,
			tokens:         15,
			lines:          5,
			wantCategory:   domain.CategoryUnknown, // CallExpr not in nodeTypeToCategory switch
			wantIsTest:     false,
			wantPriority:   domain.PriorityLow, // tokens < 25
			wantSuggestion: "Review and extract common logic",
		},
		{
			name:           "unknown node type",
			filename:       "broken.go",
			nodeType:       golang.BadNode,
			tokens:         20,
			lines:          6,
			wantCategory:   domain.CategoryUnknown,
			wantIsTest:     false,
			wantPriority:   domain.PriorityLow, // tokens < 25
			wantSuggestion: "Review and extract common logic",
		},
		{
			name:           "handler in production - large",
			filename:       "api.go",
			nodeType:       golang.FuncDecl,
			tokens:         120,
			lines:          40,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     false,
			wantPriority:   domain.PriorityCritical,
			wantSuggestion: "Extract to shared utility function",
		},
		{
			name:           "test file with large duplication",
			filename:       "large_test.go",
			nodeType:       golang.FuncDecl,
			tokens:         120,
			lines:          35,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     true,
			wantPriority:   domain.PriorityMedium, // test file with tokens > 100
			wantSuggestion: "Extract test helper function or use table-driven tests",
		},
		{
			name:           "function in production - small",
			filename:       "utils.go",
			nodeType:       golang.FuncDecl,
			tokens:         20,
			lines:          8,
			wantCategory:   domain.CategoryFunction,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens < 25
			wantSuggestion: "Extract to shared utility function",
		},
		{
			name:           "gen decl maps to expression",
			filename:       "consts.go",
			nodeType:       golang.GenDecl,
			tokens:         30,
			lines:          10,
			wantCategory:   domain.CategoryExpression,
			wantIsTest:     false,
			wantPriority:   domain.PriorityMedium, // tokens 25-50
			wantSuggestion: "Review and extract common logic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyClone(tt.filename, tt.nodeType, tt.tokens, tt.lines)

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
		{domain.CategoryHandler, "🎯"},
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

func TestClonePriorityEmoji(t *testing.T) {
	testClonePriority(
		t,
		"GetPriorityEmoji",
		func(p ClonePriority) string { return p.GetPriorityEmoji() },
		map[ClonePriority]string{
			domain.PriorityCritical: "🔴",
			domain.PriorityHigh:     "🟠",
			domain.PriorityMedium:   "🟡",
			domain.PriorityLow:      "🟢",
		},
	)
}

func TestClonePriorityColor(t *testing.T) {
	testClonePriority(
		t,
		"GetPriorityColor",
		func(p ClonePriority) string { return p.GetPriorityColor() },
		map[ClonePriority]string{
			domain.PriorityCritical: "var(--error)",
			domain.PriorityHigh:     "var(--warning)",
			domain.PriorityMedium:   "var(--accent)",
			domain.PriorityLow:      "var(--success)",
		},
	)
}

func TestPriorityHigher(t *testing.T) {
	tests := []struct {
		a, b   ClonePriority
		result bool
	}{
		{domain.PriorityCritical, domain.PriorityHigh, true},
		{domain.PriorityCritical, domain.PriorityCritical, false},
		{domain.PriorityHigh, domain.PriorityMedium, true},
		{domain.PriorityMedium, domain.PriorityLow, true},
		{domain.PriorityLow, domain.PriorityLow, false},
		{domain.PriorityLow, domain.PriorityCritical, false},
	}

	for _, tt := range tests {
		name := string(tt.a) + "_vs_" + string(tt.b)
		t.Run(name, func(t *testing.T) {
			if got := priorityHigher(tt.a, tt.b); got != tt.result {
				t.Errorf("priorityHigher(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.result)
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"handler_test.go", true},
		{"handler.go", false},
		{"test_utils.go", false},
		{"pkg_test.go", true},
		{"_test.go", true},
		{"main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := isTestFile(tt.filename); got != tt.want {
				t.Errorf("isTestFile(%v) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
