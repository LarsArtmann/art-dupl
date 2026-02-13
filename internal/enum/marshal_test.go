package enum

import (
	"encoding/json"
	"strings"
	"testing"
)

type testEnum StringEnum

const (
	testEnumA testEnum = "alpha"
	testEnumB testEnum = "beta"
	testEnumC testEnum = "gamma"
)

func (e testEnum) String() string {
	return string(e)
}

func TestStringEnum_String(t *testing.T) {
	tests := []struct {
		input    testEnum
		expected string
	}{
		{testEnumA, "alpha"},
		{testEnumB, "beta"},
		{testEnumC, "gamma"},
		{testEnum("custom"), "custom"},
		{testEnum(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.input.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue testEnum
		validValues  []testEnum
		expected     testEnum
		expectError  bool
	}{
		{
			name:         "valid value",
			input:        `"alpha"`,
			defaultValue: testEnumC,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumA,
			expectError:  false,
		},
		{
			name:         "unquoted valid value",
			input:        `beta`,
			defaultValue: testEnumC,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumB,
			expectError:  false,
		},
		{
			name:         "empty string uses default",
			input:        `""`,
			defaultValue: testEnumB,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumB,
			expectError:  false,
		},
		{
			name:         "invalid value returns error and default",
			input:        `"invalid"`,
			defaultValue: testEnumC,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumC,
			expectError:  true,
		},
		{
			name:         "whitespace trimmed",
			input:        ` "alpha" `,
			defaultValue: testEnumC,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumA,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result testEnum
			err := UnmarshalJSON(&result, []byte(tt.input), "testEnum", tt.defaultValue, tt.validValues...)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("result = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		value       testEnum
		validValues []testEnum
		expected    string
		expectError bool
	}{
		{
			name:        "valid value",
			value:       testEnumA,
			validValues: []testEnum{testEnumA, testEnumB, testEnumC},
			expected:    `"alpha"`,
			expectError: false,
		},
		{
			name:        "invalid value returns error",
			value:       testEnum("invalid"),
			validValues: []testEnum{testEnumA, testEnumB, testEnumC},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalJSON(tt.value, tt.validValues...)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && string(result) != tt.expected {
				t.Errorf("result = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestUnmarshalJSONFromStrings(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue testEnum
		validStrings []string
		expected     testEnum
		expectError  bool
	}{
		{
			name:         "valid string",
			input:        `"alpha"`,
			defaultValue: testEnumC,
			validStrings: []string{"alpha", "beta", "gamma"},
			expected:     testEnumA,
			expectError:  false,
		},
		{
			name:         "empty uses default",
			input:        `""`,
			defaultValue: testEnumB,
			validStrings: []string{"alpha", "beta", "gamma"},
			expected:     testEnumB,
			expectError:  false,
		},
		{
			name:         "invalid returns error and default",
			input:        `"invalid"`,
			defaultValue: testEnumC,
			validStrings: []string{"alpha", "beta", "gamma"},
			expected:     testEnumC,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result testEnum
			err := UnmarshalJSONFromStrings(&result, []byte(tt.input), "testEnum", tt.defaultValue, tt.validStrings)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("result = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseEnum(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue testEnum
		validValues  []testEnum
		expected     testEnum
	}{
		{
			name:         "valid value",
			input:        "alpha",
			defaultValue: testEnumC,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumA,
		},
		{
			name:         "empty returns default",
			input:        "",
			defaultValue: testEnumB,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumB,
		},
		{
			name:         "whitespace trimmed",
			input:        "  beta  ",
			defaultValue: testEnumC,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumB,
		},
		{
			name:         "invalid returns default",
			input:        "invalid",
			defaultValue: testEnumC,
			validValues:  []testEnum{testEnumA, testEnumB, testEnumC},
			expected:     testEnumC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEnum(tt.input, tt.defaultValue, tt.validValues...)
			if result != tt.expected {
				t.Errorf("ParseEnum() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestEnumToStringSlice(t *testing.T) {
	enums := []testEnum{testEnumA, testEnumB, testEnumC}
	result := EnumToStringSlice(enums...)

	expected := []string{"alpha", "beta", "gamma"}
	if len(result) != len(expected) {
		t.Fatalf("len(result) = %d, want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestValidateEnum(t *testing.T) {
	validValues := []testEnum{testEnumA, testEnumB, testEnumC}

	tests := []struct {
		value    testEnum
		expected bool
	}{
		{testEnumA, true},
		{testEnumB, true},
		{testEnumC, true},
		{testEnum("invalid"), false},
		{testEnum(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.value), func(t *testing.T) {
			result := ValidateEnum(tt.value, validValues...)
			if result != tt.expected {
				t.Errorf("ValidateEnum() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnumNames(t *testing.T) {
	enums := []testEnum{testEnumA, testEnumB, testEnumC}
	result := EnumNames(enums...)

	expected := []string{"alpha", "beta", "gamma"}
	if len(result) != len(expected) {
		t.Fatalf("len(result) = %d, want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

// validatableTestEnum for testing EnumType interface
type validatableTestEnum StringEnum

const (
	validA validatableTestEnum = "valid_a"
	validB validatableTestEnum = "valid_b"
)

func (e validatableTestEnum) IsValid() bool {
	return e == validA || e == validB
}

func (e validatableTestEnum) String() string {
	return string(e)
}

func TestMarshalJSONForInterface(t *testing.T) {
	tests := []struct {
		name     string
		value    validatableTestEnum
		expected string
	}{
		{
			name:     "valid value",
			value:    validA,
			expected: `"valid_a"`,
		},
		{
			name:     "invalid value returns null",
			value:    validatableTestEnum("invalid"),
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalJSONForInterface(tt.value, "validatableTestEnum")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("result = %s, want %s", result, tt.expected)
			}
		})
	}
}

// Test round-trip JSON marshaling/unmarshalling
func TestJSONRoundTrip(t *testing.T) {
	type config struct {
		Mode testEnum `json:"mode"`
	}

	tests := []struct {
		name  string
		input config
	}{
		{"alpha", config{Mode: testEnumA}},
		{"beta", config{Mode: testEnumB}},
		{"gamma", config{Mode: testEnumC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var result config
			decoder := json.NewDecoder(strings.NewReader(string(data)))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&result); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if result.Mode != tt.input.Mode {
				t.Errorf("round-trip failed: got %q, want %q", result.Mode, tt.input.Mode)
			}
		})
	}
}
