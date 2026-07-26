package templ

import (
	"testing"
)

func TestNormalizeExprValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "single var before dot",
			expr:     "user.Name",
			expected: "v0.Name",
		},
		{
			name:     "two vars canonicalized independently",
			expr:     "Component(user.Name, count)",
			expected: "Component(v0.Name, count)",
		},
		{
			name:     "same var reused gets same canonical",
			expr:     "user.Name + user.ID",
			expected: "v0.Name + v0.ID",
		},
		{
			name:     "no dot access - unchanged",
			expr:     "Component(foo, bar)",
			expected: "Component(foo, bar)",
		},
		{
			name:     "uppercase callee preserved",
			expr:     "render(user.Email)",
			expected: "render(v0.Email)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			symbols := make(map[string]string)
			result := normalizeExprValue(tc.expr, symbols)
			if result != tc.expected {
				t.Errorf("normalizeExprValue(%q) = %q, want %q", tc.expr, result, tc.expected)
			}
		})
	}
}

func TestNormalizeExprValue_ReservedWords(t *testing.T) {
	t.Parallel()

	symbols := make(map[string]string)

	result := normalizeExprValue("len(items)", symbols)
	if result != "len(items)" {
		t.Errorf("reserved word 'len' should not be normalized: got %q", result)
	}

	if len(symbols) != 0 {
		t.Errorf("symbols map should be empty for reserved-only expr, got %v", symbols)
	}
}

func TestNormalizeExprValue_NilSymbols(t *testing.T) {
	t.Parallel()

	result := normalizeExprValue("user.Name", nil)
	if result != "user.Name" {
		t.Errorf("normalizeExprValue with nil symbols should return unchanged: got %q", result)
	}
}

func TestNormalizeExprValue_ConsistentCanonicalization(t *testing.T) {
	t.Parallel()

	symbols := make(map[string]string)

	// First call registers user as v0
	r1 := normalizeExprValue("user.Name", symbols)
	// Second call should reuse v0 for user
	r2 := normalizeExprValue("user.Email", symbols)

	if r1 != "v0.Name" {
		t.Errorf("first call: got %q, want v0.Name", r1)
	}

	if r2 != "v0.Email" {
		t.Errorf("second call: got %q, want v0.Email", r2)
	}

	if len(symbols) != 1 {
		t.Errorf("expected 1 symbol, got %d (%v)", len(symbols), symbols)
	}
}

func TestIsReservedWord(t *testing.T) {
	t.Parallel()

	reserved := []string{"len", "cap", "copy", "new", "make", "append", "delete", "panic", "min", "max"}
	for _, word := range reserved {
		if !isReservedWord(word) {
			t.Errorf("isReservedWord(%q) = false, want true", word)
		}
	}

	nonReserved := []string{"user", "items", "foo", "Bar", ""}
	for _, word := range nonReserved {
		if isReservedWord(word) {
			t.Errorf("isReservedWord(%q) = true, want false", word)
		}
	}
}
