package domain

import (
	"testing"
)

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
			t.Error("validateRules() error = nil, want error")
		}

		if err.Error() != "validation failed: rule1 failed" {
			t.Errorf("validateRules() error = %v, want 'validation failed: rule1 failed'", err)
		}
	})

	t.Run("middle rule fails", func(t *testing.T) {
		rules := []validationRule{
			{valid: true, msg: "rule1"},
			{valid: false, msg: "rule2 failed"},
			{valid: true, msg: "rule3"},
		}

		err := validateRules(rules)
		if err == nil {
			t.Error("validateRules() error = nil, want error")
		}

		if err.Error() != "validation failed: rule2 failed" {
			t.Errorf("validateRules() error = %v, want 'validation failed: rule2 failed'", err)
		}
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

	t.Run("empty string", func(t *testing.T) {
		assertUnmarshalStringIDError(t, `""`, "empty string")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		assertUnmarshalStringIDError(t, `invalid`, "invalid JSON")
	})
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
	t.Run("uint16", func(t *testing.T) {
		var result uint16

		err := unmarshalUintGeneric[uint16]([]byte("42"), "TestType", func(n uint16) {
			result = n
		})
		if err != nil {
			t.Errorf("unmarshalUintGeneric() error = %v", err)
		}

		if result != 42 {
			t.Errorf("unmarshalUintGeneric() result = %v, want 42", result)
		}
	})

	t.Run("uint32", func(t *testing.T) {
		var result uint32

		err := unmarshalUintGeneric[uint32]([]byte("42"), "TestType", func(n uint32) {
			result = n
		})
		if err != nil {
			t.Errorf("unmarshalUintGeneric() error = %v", err)
		}

		if result != 42 {
			t.Errorf("unmarshalUintGeneric() result = %v, want 42", result)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var result uint16

		err := unmarshalUintGeneric[uint16]([]byte("invalid"), "TestType", func(n uint16) {
			result = n
		})
		_ = result

		if err == nil {
			t.Error("unmarshalUintGeneric() error = nil, want error for invalid JSON")
		}
	})
}
