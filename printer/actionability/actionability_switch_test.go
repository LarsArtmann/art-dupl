package actionability

import (
	"slices"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIsLoggingMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Print", "Print", true},
		{"Printf", "Printf", true},
		{"Println", "Println", true},
		{"Error", "Error", true},
		{"log-Errorf method", "Errorf", true},
		{"Warnf", "Warnf", true},
		{"Info", "Info", true},
		{"Infof", "Infof", true},
		{"Debug", "Debug", true},
		{"Debugf", "Debugf", true},
		{"Fatal", "Fatal", true},
		{"Fatalf", "Fatalf", true},
		{"Panic", "Panic", true},
		{"Panicf", "Panicf", true},
		{"non-matching method", "Process", false},
		{"empty string", "", false},
		{"case sensitive", "print", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := slices.Contains(loggingMethodNames, tc.input)
			if result != tc.expected {
				t.Errorf("slices.Contains(loggingMethodNames, %q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsAssertionMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Expect", "Expect", true},
		{"Assert", "Assert", true},
		{"Require", "Require", true},
		{"Should", "Should", true},
		{"Must", "Must", true},
		{"So", "So", true},
		{"non-matching method", "Process", false},
		{"empty string", "", false},
		{"case sensitive", "expect", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isAssertionMethod(tc.input)
			if result != tc.expected {
				t.Errorf("isAssertionMethod(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCallReferencesIdent(t *testing.T) {
	t.Parallel()

	// Build a CallExpr subtree: queryError(err, "msg")
	// queryError is Ident, err is Ident, "msg" is BasicLit
	callExpr := &domain.CloneNode{
		BaseType: golang.CallExpr,
		Name:     "",
		Children: []*domain.CloneNode{
			{BaseType: golang.Ident, Name: "queryError"},
			{BaseType: golang.Ident, Name: "err"},
			{BaseType: golang.BasicLit, Name: `"failed to query"`},
		},
	}

	tests := []struct {
		name     string
		node     *domain.CloneNode
		target   string
		expected bool
	}{
		{"call references err", callExpr, "err", true},
		{"call references queryError", callExpr, "queryError", true},
		{"call does not reference foo", callExpr, "foo", false},
		{"empty target", callExpr, "", false},
		{"nil-safe on bare ident", &domain.CloneNode{BaseType: golang.Ident, Name: "err"}, "err", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := callReferencesIdent(tc.node, tc.target)
			if result != tc.expected {
				t.Errorf("callReferencesIdent = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestErrorWrappingStructural(t *testing.T) {
	t.Parallel()

	// if err != nil { return nil, queryError(err, "unique msg") }
	queryErrorIf := &domain.CloneNode{
		BaseType: golang.IfStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BinaryExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "err"},
					{BaseType: golang.Ident, Name: "nil"},
				},
			},
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ReturnStmt,
						Children: []*domain.CloneNode{
							{BaseType: golang.Ident, Name: "nil"},
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{BaseType: golang.Ident, Name: "queryError"},
									{BaseType: golang.Ident, Name: "err"},
									{BaseType: golang.BasicLit, Name: `"failed to query"`},
								},
							},
						},
					},
				},
			},
		},
	}

	// Regression: fmt.Errorf("wrap: %w", err) should still match
	fmtErrorfIf := &domain.CloneNode{
		BaseType: golang.IfStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BinaryExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "err"},
					{BaseType: golang.Ident, Name: "nil"},
				},
			},
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ReturnStmt,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{
										BaseType: golang.SelectorExpr,
										Name:     "Errorf",
										Children: []*domain.CloneNode{
											{BaseType: golang.Ident, Name: "fmt"},
										},
									},
									{BaseType: golang.BasicLit, Name: `"wrap: %w"`},
									{BaseType: golang.Ident, Name: "err"},
								},
							},
						},
					},
				},
			},
		},
	}

	// Negative: call that does NOT reference err should not match
	noErrRefIf := &domain.CloneNode{
		BaseType: golang.IfStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BinaryExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "err"},
					{BaseType: golang.Ident, Name: "nil"},
				},
			},
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ReturnStmt,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{BaseType: golang.Ident, Name: "doSomething"},
									{BaseType: golang.BasicLit, Name: `"unrelated"`},
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		node     *domain.CloneNode
		expected bool
	}{
		{"queryError wrapping err", queryErrorIf, true},
		{"fmt.Errorf wrapping err", fmtErrorfIf, true},
		{"call not referencing err", noErrRefIf, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isErrorWrappingBody(tc.node)
			if result != tc.expected {
				t.Errorf("isErrorWrappingBody = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestIsTestingVarName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"t", "t", true},
		{"tt", "tt", true},
		{"tc", "tc", true},
		{"test", "test", true},
		{"ts", "ts", true},
		{"tb", "tb", true},
		{"testing", "testing", true},
		{"t0", "t0", true},
		{"non-testing var", "svc", false},
		{"empty string", "", false},
		{"case sensitive", "T", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := slices.Contains(testingVarNames, tc.input)
			if result != tc.expected {
				t.Errorf("slices.Contains(testingVarNames, %q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}
