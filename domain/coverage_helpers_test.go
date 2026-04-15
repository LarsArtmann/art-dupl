package domain

import (
	"strings"
	"testing"
)

func assertErrorContains(t *testing.T, err error, substrs ...string) {
	t.Helper()

	for _, substr := range substrs {
		if !strings.Contains(err.Error(), substr) {
			t.Errorf("error = %v, want error containing %q", err, substr)
		}
	}
}

// TestValidationHelpers tests validation helper functions.
func TestValidateRules(t *testing.T) {
	t.Run("all rules pass", func(t *testing.T) {
		rules := []validationRule{
			{valid: true, msg: "rule1"},
			{valid: true, msg: "rule2"},
		}

		err := validateRules(rules)
		if err != nil {
			t.Errorf("validateRules() error = %v, want nil", err)
		}
	})

	t.Run("first rule fails", func(t *testing.T) {
		rules := []validationRule{
			{valid: false, msg: "rule1 failed"},
			{valid: true, msg: "rule2"},
		}

		err := validateRules(rules)
		if err == nil {
			t.Fatal("validateRules() error = nil, want error")
		}

		assertErrorContains(t, err, "validation failed", "rule1 failed")
	})

	t.Run("middle rule fails", func(t *testing.T) {
		rules := []validationRule{
			{valid: true, msg: "rule1"},
			{valid: false, msg: "rule2 failed"},
			{valid: true, msg: "rule3"},
		}

		err := validateRules(rules)
		if err == nil {
			t.Fatal("validateRules() error = nil, want error")
		}

		assertErrorContains(t, err, "validation failed", "rule2 failed")
	})
}

func TestValidateFields(t *testing.T) {
	t.Run("all fields valid", func(t *testing.T) {
		err := validateFields(
			validationRule{true, "field1"},
			validationRule{true, "field2"},
		)
		if err != nil {
			t.Errorf("validateFields() error = %v, want nil", err)
		}
	})

	t.Run("one field invalid", func(t *testing.T) {
		err := validateFields(
			validationRule{true, "field1"},
			validationRule{false, "field2 is invalid"},
		)
		if err == nil {
			t.Error("validateFields() error = nil, want error")
		}
	})
}

// TestMarshalStringID tests the marshalStringID helper function.
func TestMarshalStringID(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		got, err := marshalStringID("test-id", "ID cannot be empty")
		if err != nil {
			t.Errorf("marshalStringID() error = %v", err)
		}

		if string(got) != `"test-id"` {
			t.Errorf("marshalStringID() = %v, want `\"test-id\"`", string(got))
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := marshalStringID("", "ID cannot be empty")
		if err == nil {
			t.Error("marshalStringID() error = nil, want error for empty string")
		}
	})
}

// assertUnmarshalStringIDError tests that unmarshalStringID returns an error for the given input.
func assertUnmarshalStringIDError(t *testing.T, input, wantErrContains string) {
	t.Helper()

	var result string

	err := unmarshalStringID([]byte(input), "TestType", "TestType cannot be empty", func(s string) {
		result = s
	})
	_ = result

	if err == nil {
		t.Errorf("unmarshalStringID() error = nil, want error %q", wantErrContains)
	}
}

func TestUnmarshalStringID(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		var result string

		err := unmarshalStringID(
			[]byte(`"test-id"`),
			"TestType",
			"TestType cannot be empty",
			func(s string) {
				result = s
			},
		)
		if err != nil {
			t.Errorf("unmarshalStringID() error = %v", err)
		}

		if result != "test-id" {
			t.Errorf("unmarshalStringID() result = %v, want 'test-id'", result)
		}
	})

	tests := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: `""`},
		{name: "invalid JSON", input: `invalid`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertUnmarshalStringIDError(t, tt.input, tt.name)
		})
	}
}

func TestMarshalUint(t *testing.T) {
	got, err := marshalUint(42)
	if err != nil {
		t.Errorf("marshalUint() error = %v", err)
	}

	if string(got) != "42" {
		t.Errorf("marshalUint() = %v, want '42'", string(got))
	}
}

func TestUnmarshalUint(t *testing.T) {
	t.Run("valid uint", func(t *testing.T) {
		var result uint

		err := unmarshalUint([]byte("42"), "TestType", func(n uint) {
			result = n
		})
		if err != nil {
			t.Errorf("unmarshalUint() error = %v", err)
		}

		if result != 42 {
			t.Errorf("unmarshalUint() result = %v, want 42", result)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var result uint

		err := unmarshalUint([]byte("invalid"), "TestType", func(n uint) {
			result = n
		})
		_ = result

		if err == nil {
			t.Error("unmarshalUint() error = nil, want error for invalid JSON")
		}
	})
}

func TestUnmarshalUintNonZero(t *testing.T) {
	t.Run("valid non-zero", func(t *testing.T) {
		testUnmarshalUintNonZero(t, []byte("42"), false, 42)
	})

	t.Run("zero value", func(t *testing.T) {
		testUnmarshalUintNonZero(t, []byte("0"), true, 0)
	})
}

func testUnmarshalUintNonZero(t *testing.T, input []byte, expectError bool, expectedValue uint) {
	t.Helper()

	var result uint

	err := unmarshalUintNonZero(
		input,
		"TestType",
		"TestType cannot be zero",
		func(n uint) {
			result = n
		},
	)

	if expectError {
		_ = result

		if err == nil {
			t.Error("unmarshalUintNonZero() error = nil, want error for zero value")
		}

		return
	}

	if err != nil {
		t.Errorf("unmarshalUintNonZero() error = %v", err)
	}

	if result != expectedValue {
		t.Errorf("unmarshalUintNonZero() result = %v, want %d", result, expectedValue)
	}
}

func TestUnmarshalUintGeneric(t *testing.T) {
	for _, tc := range []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{name: "uint16", data: []byte("42"), expectError: false},
		{name: "uint32", data: []byte("42"), expectError: false},
		{name: "invalid JSON", data: []byte("invalid"), expectError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result uint16

			err := unmarshalUintGeneric[uint16](tc.data, "TestType", func(n uint16) {
				result = n
			})

			if tc.expectError {
				if err == nil {
					t.Error("expected error for invalid JSON")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if result != 42 {
					t.Errorf("result = %v, want 42", result)
				}
			}
		})
	}
}
