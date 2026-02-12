package testutil

import "testing"

// TableTestCase represents a single test case in a table-driven test.
type TableTestCase struct {
	Name string
}

// RunTableTest executes a table-driven test with the given test cases and assertion function.
// This helper reduces boilerplate in standard table-driven test patterns.
//
// Example usage:
//
//	tests := []struct {
//		testutil.TableTestCase
//		format   Format
//		expected bool
//	}{
//		{TableTestCase: testutil.TableTestCase{Name: "text is valid"}, format: FormatText, expected: true},
//	}
//	testutil.RunTableTest(t, tests, func(t *testing.T, tt struct {
//		testutil.TableTestCase
//		format   Format
//		expected bool
//	}) {
//		if got := tt.format.IsValid(); got != tt.expected {
//			t.Errorf("Format.IsValid() = %v, want %v", got, tt.expected)
//		}
//	})
func RunTableTest[T any](t *testing.T, tests []T, assertion func(t *testing.T, tt T)) {
	t.Helper()
	for _, tt := range tests {
		// Extract name using type assertion if possible
		var name string
		if tc, ok := any(tt).(interface{ GetName() string }); ok {
			name = tc.GetName()
		}
		t.Run(name, func(t *testing.T) {
			assertion(t, tt)
		})
	}
}

// RunTableTestWithName executes a table-driven test with a custom name extractor function.
// This is useful when test structs have a lowercase 'name' field or need custom naming logic.
//
// Example usage:
//
//	tests := []struct {
//		name     string
//		format   Format
//		expected bool
//	}{
//		{name: "text is valid", format: FormatText, expected: true},
//	}
//	testutil.RunTableTestWithName(t, tests, func(tt struct {
//		name     string
//		format   Format
//		expected bool
//	}) string {
//		return tt.name
//	}, func(t *testing.T, tt struct {
//		name     string
//		format   Format
//		expected bool
//	}) {
//		testutil.AssertEqual(t, tt.format.IsValid(), tt.expected, "Format.IsValid()")
//	})
func RunTableTestWithName[T any](t *testing.T, tests []T, getName func(T) string, assertion func(t *testing.T, tt T)) {
	t.Helper()
	for _, tt := range tests {
		t.Run(getName(tt), func(t *testing.T) {
			assertion(t, tt)
		})
	}
}

// RunNamedTest executes a subtest with the given name and test function.
// This is a thin wrapper around t.Run for consistency.
func RunNamedTest(t *testing.T, name string, testFunc func(t *testing.T)) {
	t.Helper()
	t.Run(name, testFunc)
}

// AssertEqual checks if got equals want and reports an error if not.
// The messagePrefix is used in the error message (e.g., "Format.IsValid()").
func AssertEqual[T comparable](t *testing.T, got, want T, messagePrefix string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", messagePrefix, got, want)
	}
}
