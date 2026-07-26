package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// commandSelectorName is used in test tree construction for cobra.Command/fang.Command.
const commandSelectorName = "Command"

func TestHasCommandReceiver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sel      *domain.CloneNode
		expected bool
	}{
		{
			name: "cobra receiver",
			sel: &domain.CloneNode{
				BaseType: golang.SelectorExpr,
				Name:     commandSelectorName,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "cobra"},
					{BaseType: golang.Ident, Name: commandSelectorName},
				},
			},
			expected: true,
		},
		{
			name: "fang receiver",
			sel: &domain.CloneNode{
				BaseType: golang.SelectorExpr,
				Name:     commandSelectorName,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "fang"},
					{BaseType: golang.Ident, Name: commandSelectorName},
				},
			},
			expected: true,
		},
		{
			name: "non-cli receiver",
			sel: &domain.CloneNode{
				BaseType: golang.SelectorExpr,
				Name:     commandSelectorName,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "myapp"},
					{BaseType: golang.Ident, Name: commandSelectorName},
				},
			},
			expected: false,
		},
		{
			name: "no children",
			sel: &domain.CloneNode{
				BaseType: golang.SelectorExpr,
				Name:     commandSelectorName,
			},
			expected: false,
		},
		{
			name: "children but no Ident",
			sel: &domain.CloneNode{
				BaseType: golang.SelectorExpr,
				Name:     commandSelectorName,
				Children: []*domain.CloneNode{
					{BaseType: golang.BasicLit, Name: "cobra"},
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := hasCommandReceiver(tc.sel)
			if result != tc.expected {
				t.Errorf("hasCommandReceiver() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestHasTestingReceiver(t *testing.T) {
	t.Parallel()

	mustSelWithReceiver := func(recv string) *domain.CloneNode {
		return &domain.CloneNode{
			BaseType: golang.SelectorExpr,
			Name:     "Run",
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: recv},
				{BaseType: golang.Ident, Name: "Run"},
			},
		}
	}

	tests := []struct {
		name     string
		sel      *domain.CloneNode
		expected bool
	}{
		{"t receiver", mustSelWithReceiver("t"), true},
		{"tt receiver", mustSelWithReceiver("tt"), true},
		{"tc receiver", mustSelWithReceiver("tc"), true},
		{"test receiver", mustSelWithReceiver("test"), true},
		{"tb receiver", mustSelWithReceiver("tb"), true},
		{"non-testing receiver", mustSelWithReceiver("svc"), false},
		{"no children", &domain.CloneNode{BaseType: golang.SelectorExpr, Name: "Run"}, false},
		{
			"children but no Ident",
			&domain.CloneNode{
				BaseType: golang.SelectorExpr,
				Name:     "Run",
				Children: []*domain.CloneNode{
					{BaseType: golang.BasicLit, Name: "t"},
				},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := hasTestingReceiver(tc.sel)
			if result != tc.expected {
				t.Errorf("hasTestingReceiver() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestIsChainOfCallsWithDifferentReceivers(t *testing.T) {
	t.Parallel()

	mustCallExprWithSelector := func(selectorName string) *domain.CloneNode {
		return &domain.CloneNode{
			BaseType: golang.CallExpr,
			Children: []*domain.CloneNode{
				{BaseType: golang.SelectorExpr, Name: selectorName},
			},
		}
	}

	tests := []struct {
		name     string
		seq      []*domain.CloneNode
		expected bool
	}{
		{
			name: "3 calls with 2 distinct receivers",
			seq: []*domain.CloneNode{
				mustCallExprWithSelector("Builder"),
				mustCallExprWithSelector("WithX"),
				mustCallExprWithSelector("Build"),
			},
			expected: true,
		},
		{
			name: "3 calls with same receiver",
			seq: []*domain.CloneNode{
				mustCallExprWithSelector("Builder"),
				mustCallExprWithSelector("Builder"),
				mustCallExprWithSelector("Builder"),
			},
			expected: false,
		},
		{
			name: "2 calls with different receivers",
			seq: []*domain.CloneNode{
				mustCallExprWithSelector("Builder"),
				mustCallExprWithSelector("Build"),
			},
			expected: false,
		},
		{
			name: "single CallExpr",
			seq: []*domain.CloneNode{
				mustCallExprWithSelector("Builder"),
			},
			expected: false,
		},
		{
			name:     "empty sequence",
			seq:      []*domain.CloneNode{},
			expected: false,
		},
		{
			name: "non-CallExpr nodes only",
			seq: []*domain.CloneNode{
				{BaseType: golang.AssignStmt},
				{BaseType: golang.ReturnStmt},
			},
			expected: false,
		},
		{
			name: "4 calls with 3 distinct receivers",
			seq: []*domain.CloneNode{
				mustCallExprWithSelector("A"),
				mustCallExprWithSelector("B"),
				mustCallExprWithSelector("C"),
				mustCallExprWithSelector("D"),
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isChainOfCallsWithDifferentReceivers(tc.seq)
			if result != tc.expected {
				t.Errorf("isChainOfCallsWithDifferentReceivers() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestIsReturnOrWrappedReturn_TwoStmt(t *testing.T) {
	t.Parallel()

	mustLogPrintExprStmt := func(methodName string) *domain.CloneNode {
		return &domain.CloneNode{
			BaseType: golang.ExprStmt,
			Children: []*domain.CloneNode{
				{
					BaseType: golang.CallExpr,
					Children: []*domain.CloneNode{
						{
							BaseType: golang.SelectorExpr,
							Name:     methodName,
						},
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		block    *domain.CloneNode
		expected bool
	}{
		{
			name: "log.Print + return (2-stmt)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					mustLogPrintExprStmt("Print"),
					{BaseType: golang.ReturnStmt},
				},
			},
			expected: true,
		},
		{
			name: "fmt.Warnf + return (2-stmt)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					mustLogPrintExprStmt("Warnf"),
					{BaseType: golang.ReturnStmt},
				},
			},
			expected: true,
		},
		{
			name: "slog.Error + return (2-stmt)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					mustLogPrintExprStmt("Error"),
					{BaseType: golang.ReturnStmt},
				},
			},
			expected: true,
		},
		{
			name: "non-log call + return (2-stmt, should fail)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					mustLogPrintExprStmt("Process"),
					{BaseType: golang.ReturnStmt},
				},
			},
			expected: false,
		},
		{
			name: "single return (1-stmt)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{BaseType: golang.ReturnStmt},
				},
			},
			expected: true,
		},
		{
			name: "single CallExpr (1-stmt)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
				},
			},
			expected: true,
		},
		{
			name: "empty block",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
			},
			expected: false,
		},
		{
			name: "3 children (too many)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					mustLogPrintExprStmt("Print"),
					{BaseType: golang.AssignStmt},
					{BaseType: golang.ReturnStmt},
				},
			},
			expected: false,
		},
		{
			name: "log + non-return (2-stmt, should fail)",
			block: &domain.CloneNode{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					mustLogPrintExprStmt("Print"),
					{BaseType: golang.AssignStmt},
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isReturnOrWrappedReturn(tc.block)
			if result != tc.expected {
				t.Errorf("isReturnOrWrappedReturn() = %v, want %v", result, tc.expected)
			}
		})
	}
}
