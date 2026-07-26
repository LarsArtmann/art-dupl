package actionability

import (
	"slices"
	"testing"
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

func TestIsWrappingCallName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"wrap-Errorw", "Errorw", true},
		{"Wrap", "Wrap", true},
		{"Wrapf", "Wrapf", true},
		{"Wrapr", "Wrapr", true},
		{"non-matching method", "Process", false},
		{"empty string", "", false},
		{"case sensitive", "errorf", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isWrappingCallName(tc.input)
			if result != tc.expected {
				t.Errorf("isWrappingCallName(%q) = %v, want %v", tc.input, result, tc.expected)
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
