package domain

import (
	"encoding/json"
	"math"
	"testing"
)

func TestNewTokenValue(t *testing.T) {
	tests := []struct {
		name     string
		value    int32
		expected TokenValue
	}{
		{"zero", 0, 0},
		{"positive", 42, 42},
		{"negative", -1, -1},
		{"max int32", math.MaxInt32, TokenValue(math.MaxInt32)},
		{"min int32", math.MinInt32, TokenValue(math.MinInt32)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTokenValue(tt.value)
			if result != tt.expected {
				t.Errorf("NewTokenValue(%d) = %d, want %d", tt.value, result, tt.expected)
			}
		})
	}
}

func TestNewTokenValueFromInt(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"zero", 0, false},
		{"positive", 42, false},
		{"negative", -1, false},
		{"max int32", math.MaxInt32, false},
		{"min int32", math.MinInt32, false},
		{"overflow positive", math.MaxInt32 + 1, true},
		{"overflow negative", math.MinInt32 - 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewTokenValueFromInt(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTokenValueFromInt(%d) expected error, got nil", tt.value)
				}
				return
			}
			if err != nil {
				t.Errorf("NewTokenValueFromInt(%d) unexpected error: %v", tt.value, err)
				return
			}
			if int(result.Int()) != tt.value {
				t.Errorf("NewTokenValueFromInt(%d) = %d, want %d", tt.value, result, tt.value)
			}
		})
	}
}

func TestTokenValue_Int32(t *testing.T) {
	tv := NewTokenValue(42)
	if got := tv.Int32(); got != 42 {
		t.Errorf("Int32() = %d, want 42", got)
	}
}

func TestTokenValue_Int(t *testing.T) {
	tv := NewTokenValue(42)
	if got := tv.Int(); got != 42 {
		t.Errorf("Int() = %d, want 42", got)
	}
}

func TestTokenValue_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		value    TokenValue
		wantErr  bool
	}{
		{"zero", 0, false},
		{"positive", 42, false},
		{"negative", -1, false},
		{"max", MaxTokenValue, false},
		{"min", MinTokenValue, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.IsValid()
			if tt.wantErr && err == nil {
				t.Errorf("IsValid() expected error for %d", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("IsValid() unexpected error for %d: %v", tt.value, err)
			}
		})
	}
}

func TestTokenValue_IsZero(t *testing.T) {
	tests := []struct {
		value TokenValue
		want  bool
	}{
		{0, true},
		{1, false},
		{-1, false},
	}

	for _, tt := range tests {
		if got := tt.value.IsZero(); got != tt.want {
			t.Errorf("IsZero() for %d = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestTokenValue_IsNegative(t *testing.T) {
	tests := []struct {
		value TokenValue
		want  bool
	}{
		{-1, true},
		{0, false},
		{1, false},
	}

	for _, tt := range tests {
		if got := tt.value.IsNegative(); got != tt.want {
			t.Errorf("IsNegative() for %d = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestTokenValue_IsPositive(t *testing.T) {
	tests := []struct {
		value TokenValue
		want  bool
	}{
		{1, true},
		{0, false},
		{-1, false},
	}

	for _, tt := range tests {
		if got := tt.value.IsPositive(); got != tt.want {
			t.Errorf("IsPositive() for %d = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestTokenValue_String(t *testing.T) {
	tv := NewTokenValue(42)
	want := "TokenValue(42)"
	if got := tv.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestTokenValue_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		value    TokenValue
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 42, "42"},
		{"negative", -1, "-1"},
		{"max", MaxTokenValue, "2147483647"},
		{"min", MinTokenValue, "-2147483648"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.MarshalJSON()
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestTokenValue_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected TokenValue
		wantErr  bool
	}{
		{"zero", "0", 0, false},
		{"positive", "42", 42, false},
		{"negative", "-1", -1, false},
		{"max", "2147483647", MaxTokenValue, false},
		{"min", "-2147483648", MinTokenValue, false},
		{"invalid string", `"invalid"`, 0, true},
		{"invalid float", "3.14", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tv TokenValue
			err := tv.UnmarshalJSON([]byte(tt.json))
			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON(%s) expected error", tt.json)
				}
				return
			}
			if err != nil {
				t.Errorf("UnmarshalJSON(%s) unexpected error: %v", tt.json, err)
				return
			}
			if tv != tt.expected {
				t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.json, tv, tt.expected)
			}
		})
	}
}

func TestTokenValue_JSONRoundTrip(t *testing.T) {
	tests := []TokenValue{0, 42, -1, MaxTokenValue, MinTokenValue}

	for _, tv := range tests {
		data, err := json.Marshal(tv)
		if err != nil {
			t.Errorf("Marshal(%d) error: %v", tv, err)
			continue
		}

		var result TokenValue
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("Unmarshal(%d) error: %v", tv, err)
			continue
		}

		if result != tv {
			t.Errorf("Round-trip failed: %d -> %s -> %d", tv, data, result)
		}
	}
}

func TestTokenValue_Equal(t *testing.T) {
	tests := []struct {
		a    TokenValue
		b    TokenValue
		want bool
	}{
		{42, 42, true},
		{42, 43, false},
		{0, 0, true},
		{-1, -1, true},
		{-1, 1, false},
	}

	for _, tt := range tests {
		if got := tt.a.Equal(tt.b); got != tt.want {
			t.Errorf("Equal(%d, %d) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTokenValue_Less(t *testing.T) {
	tests := []struct {
		a    TokenValue
		b    TokenValue
		want bool
	}{
		{1, 2, true},
		{2, 1, false},
		{1, 1, false},
		{-1, 0, true},
		{0, -1, false},
	}

	for _, tt := range tests {
		if got := tt.a.Less(tt.b); got != tt.want {
			t.Errorf("Less(%d, %d) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTokenValue_Max(t *testing.T) {
	tests := []struct {
		a    TokenValue
		b    TokenValue
		want TokenValue
	}{
		{1, 2, 2},
		{2, 1, 2},
		{1, 1, 1},
		{-1, 0, 0},
	}

	for _, tt := range tests {
		if got := tt.a.Max(tt.b); got != tt.want {
			t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTokenValue_Min(t *testing.T) {
	tests := []struct {
		a    TokenValue
		b    TokenValue
		want TokenValue
	}{
		{1, 2, 1},
		{2, 1, 1},
		{1, 1, 1},
		{-1, 0, -1},
	}

	for _, tt := range tests {
		if got := tt.a.Min(tt.b); got != tt.want {
			t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTokenValue_Abs(t *testing.T) {
	tests := []struct {
		value TokenValue
		want  TokenValue
	}{
		{42, 42},
		{-42, 42},
		{0, 0},
		{MinTokenValue, MinTokenValue}, // Special case: MinInt32 overflows
	}

	for _, tt := range tests {
		if got := tt.value.Abs(); got != tt.want {
			t.Errorf("Abs(%d) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestTokenValue_Add(t *testing.T) {
	tests := []struct {
		a    TokenValue
		b    TokenValue
		want TokenValue
	}{
		{1, 2, 3},
		{-1, 1, 0},
		{0, 0, 0},
	}

	for _, tt := range tests {
		if got := tt.a.Add(tt.b); got != tt.want {
			t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTokenValue_AddChecked(t *testing.T) {
	tests := []struct {
		name    string
		a       TokenValue
		b       TokenValue
		want    TokenValue
		wantErr bool
	}{
		{"simple", 1, 2, 3, false},
		{"negative", -1, 1, 0, false},
		{"zero", 0, 0, 0, false},
		{"overflow positive", MaxTokenValue, 1, 0, true},
		{"overflow negative", MinTokenValue, -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.AddChecked(tt.b)
			if tt.wantErr {
				if err == nil {
					t.Errorf("AddChecked(%d, %d) expected error", tt.a, tt.b)
				}
				return
			}
			if err != nil {
				t.Errorf("AddChecked(%d, %d) unexpected error: %v", tt.a, tt.b, err)
				return
			}
			if got != tt.want {
				t.Errorf("AddChecked(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestTokenValueFromMapKey(t *testing.T) {
	m := map[TokenValue]string{
		42: "forty-two",
		0:  "zero",
	}

	tests := []struct {
		key       TokenValue
		wantValue string
		wantOK    bool
	}{
		{42, "forty-two", true},
		{0, "zero", true},
		{999, "", false},
	}

	for _, tt := range tests {
		got, ok := TokenValueFromMapKey(m, tt.key)
		if ok != tt.wantOK {
			t.Errorf("TokenValueFromMapKey(%d) ok = %v, want %v", tt.key, ok, tt.wantOK)
			continue
		}
		if got != tt.wantValue {
			t.Errorf("TokenValueFromMapKey(%d) = %q, want %q", tt.key, got, tt.wantValue)
		}
	}
}

// Benchmarks

func BenchmarkTokenValue_Int32(b *testing.B) {
	tv := NewTokenValue(42)
	for i := 0; i < b.N; i++ {
		_ = tv.Int32()
	}
}

func BenchmarkTokenValue_MarshalJSON(b *testing.B) {
	tv := NewTokenValue(42)
	for i := 0; i < b.N; i++ {
		_, _ = tv.MarshalJSON()
	}
}

func BenchmarkTokenValue_UnmarshalJSON(b *testing.B) {
	data := []byte("42")
	for i := 0; i < b.N; i++ {
		var tv TokenValue
		_ = tv.UnmarshalJSON(data)
	}
}
