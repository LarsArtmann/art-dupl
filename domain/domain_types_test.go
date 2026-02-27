package domain

import (
	stderrors "errors"
	"fmt"
	"testing"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// testUintType is a helper for testing uint-based types with New*, Uint(), and RoundTrip methods.
type testUintType[T any] struct {
	newFunc       func(uint) T
	uintFunc      func(T) uint
	jsonMarshal   func(T) ([]byte, error)
	jsonUnmarshal func(*T, []byte) error
}

// testUintTypeSuite runs the standard test suite for uint-based types.
func testUintTypeSuite[T any](t *testing.T, typeName string, tt testUintType[T]) {
	t.Helper()
	t.Run("New"+typeName, func(t *testing.T) {
		tests := []struct {
			name  string
			input uint
			want  T
		}{
			{"zero", 0, tt.newFunc(0)},
			{"positive", 100, tt.newFunc(100)},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got := tt.newFunc(tc.input)
				// For types implementing UintWrapper, compare using Uint() to get value comparison
				// rather than pointer comparison
				if uintWrapper, ok := any(got).(interface{ Uint() uint }); ok {
					if wantWrapper, ok := any(tc.want).(interface{ Uint() uint }); ok {
						if uintWrapper.Uint() != wantWrapper.Uint() {
							t.Errorf("New%s() = %v, want %v", typeName, got, tc.want)
						}

						return
					}
				}
				// Fallback for non-UintWrapper types
				// Use reflection or type assertion to compare
				gotValue := fmt.Sprintf("%v", got)

				wantValue := fmt.Sprintf("%v", tc.want)
				if gotValue != wantValue {
					t.Errorf("New%s() = %v, want %v", typeName, got, tc.want)
				}
			})
		}
	})

	t.Run("Uint", func(t *testing.T) {
		val := tt.newFunc(42)
		if got := tt.uintFunc(val); got != 42 {
			t.Errorf("Uint() = %v, want %v", got, 42)
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		original := tt.newFunc(123)
		data, _ := tt.jsonMarshal(original)

		var result T

		err := tt.jsonUnmarshal(&result, data)
		if err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}
		// Compare using the uint function to get value comparison
		// rather than pointer comparison (for pointer types)
		if tt.uintFunc(result) != tt.uintFunc(original) {
			t.Errorf("Round trip failed: %v != %v", result, original)
		}
	})
}

// testJSONRoundTrip is a helper for testing JSON marshaling and unmarshaling.
func testJSONRoundTrip[T comparable](
	t *testing.T,
	original T,
	marshal func(T) ([]byte, error),
	unmarshal func(*T, []byte) error,
) {
	t.Helper()

	data, err := marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	var result T
	if err := unmarshal(&result, data); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if result != original {
		t.Errorf("Round trip failed: %v != %v", result, original)
	}
}

// constructorTest is a helper for testing constructors that may return errors.
type constructorTest[T any] struct {
	name      string
	input     any
	want      T
	wantError bool
}

// runConstructorTests runs a series of constructor tests with error checking.
func runConstructorTests[T any](
	t *testing.T,
	constructorName string,
	tests []constructorTest[T],
	newFunc func(any) (T, error),
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := newFunc(tt.input)

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("%s() expected error, got nil", constructorName)

					return
				}

				var validationErr *duplerrors.DuplError
				if !stderrors.As(gotErr, &validationErr) {
					t.Errorf("%s() expected ValidationError, got %T", constructorName, gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("%s() unexpected error: %v", constructorName, gotErr)

					return
				}
				// Special handling for uint-based types
				switch v := any(got).(type) {
				case interface{ Uint() uint }:
					if wantWrapper, ok := any(tt.want).(interface{ Uint() uint }); ok {
						if v.Uint() != wantWrapper.Uint() {
							t.Errorf("%s() = %v, want %v", constructorName, got, tt.want)
						}

						return
					}
				default:
					// Use string comparison for incomparable types
					gotValue := fmt.Sprintf("%v", got)

					wantValue := fmt.Sprintf("%v", tt.want)
					if gotValue != wantValue {
						t.Errorf("%s() = %v, want %v", constructorName, got, tt.want)
					}
				}
			}
		})
	}
}

// jsonTest is a helper for testing JSON marshaling/unmarshaling.
type jsonTest[T comparable] struct {
	name    string
	input   T
	want    string
	wantErr bool
}

// jsonUnmarshalTest is a helper for testing JSON unmarshaling.
type jsonUnmarshalTest[T comparable] struct {
	name      string
	input     string
	want      T
	wantError bool
}

// runJSONTests runs JSON marshal/unmarshal tests.
func runJSONTests[T comparable](
	t *testing.T,
	marshal func(T) ([]byte, error),
	unmarshal func(*T, []byte) error,
	tests []jsonTest[T],
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := marshal(tt.input)

			if tt.wantErr {
				if gotErr == nil {
					t.Errorf("MarshalJSON() expected error, got nil")

					return
				}
			} else {
				if gotErr != nil {
					t.Errorf("MarshalJSON() unexpected error: %v", gotErr)

					return
				}

				if string(got) != tt.want {
					t.Errorf("MarshalJSON() = %v, want %v", string(got), tt.want)
				}
			}
		})
	}
}

// runJSONUnmarshalTests runs JSON unmarshal tests.
func runJSONUnmarshalTests[T comparable](
	t *testing.T,
	unmarshal func(*T, []byte) error,
	tests []struct {
		name      string
		input     string
		want      T
		wantError bool
	},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got T

			gotErr := unmarshal(&got, []byte(tt.input))

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("UnmarshalJSON() expected error, got nil")

					return
				}
			} else {
				if gotErr != nil {
					t.Errorf("UnmarshalJSON() unexpected error: %v", gotErr)

					return
				}

				if got != tt.want {
					t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// uintTypeNames defines all uint-based type names for testing.
//

var uintTypeNames = []string{
	"BytePosition",
	"TokenCount",
	"ComplexityScore",
	"FileCount",
	"CloneCount",
}

// TestCase represents a test case with a name and test function.
// This struct eliminates duplicate anonymous struct definitions across test helper functions.
type TestCase struct {
	name string
	test func(*testing.T)
}

// createUintTypeTest generates a test case for a uint-based type.
// This helper eliminates repetitive boilerplate by auto-generating the standard test functions:
// - Constructor (Type(uint))
// - Uint() method
// - MarshalJSON() method
// - UnmarshalJSON() method
//
// Parameters:
//   - typeName: Name of the type being tested (e.g., "BytePosition")
//   - newFunc: Constructor function that creates the type from uint
//   - uintFunc: Method that extracts uint from the type
//   - marshalFunc: Method that marshals the type to JSON
//   - unmarshalFunc: Method that unmarshals JSON to the type
func createUintTypeTest[T comparable](
	typeName string,
	newFunc func(uint) T,
	uintFunc func(T) uint,
	marshalFunc func(T) ([]byte, error),
	unmarshalFunc func(*T, []byte) error,
) TestCase {
	return TestCase{
		name: typeName,
		test: func(t *testing.T) {
			t.Helper()
			testUintTypeSuite(t, typeName, testUintType[T]{
				newFunc:       newFunc,
				uintFunc:      uintFunc,
				jsonMarshal:   marshalFunc,
				jsonUnmarshal: unmarshalFunc,
			})
		},
	}
}

// registerTestForType is a generic helper that eliminates code duplication by handling the common pattern.
// This function encapsulates the repetitive logic for testing uint-based types.
// It delegates to createUintTypeTest to maintain consistency.
func registerTestForType[T comparable](t *testing.T, typeName string, constructor func(uint) T,
	getUint func(T) uint, marshal func(T) ([]byte, error), unmarshal func(*T, []byte) error,
) {
	t.Helper()

	testCase := createUintTypeTest(typeName, constructor, getUint, marshal, unmarshal)
	testCase.test(t)
}

// UintWrapper defines the common interface for uint-based types.
// This interface allows us to write generic code that works with all uint wrapper types.
type UintWrapper interface {
	Uint() uint
}

// UintConstructor defines a constructor function for uint-based types.
type UintConstructor[T interface{ Uint() uint }] func(uint) T

// registerUintTypeTestGeneric is a generic helper that eliminates code duplication
// by handling the common pattern for uint-based types that implement UintWrapper.
// This function encapsulates the repetitive logic for testing uint wrapper types.
func registerUintTypeTestGeneric[T interface {
	Uint() uint
	comparable
}](t *testing.T, typeName string, constructor func(uint) T) {
	t.Helper()
	registerTestForType(t, typeName,
		constructor,
		func(v T) uint { return v.Uint() },
		func(v T) ([]byte, error) {
			// MarshalJSON is defined on the value receiver
			if marshaler, ok := any(v).(interface{ MarshalJSON() ([]byte, error) }); ok {
				return marshaler.MarshalJSON()
			}

			return nil, fmt.Errorf("type %T does not implement MarshalJSON", v)
		},
		jsonUnmarshalFuncPtr[T]())
}

// registerUintTypeByName registers a test for a uint-based type by name.
// This helper eliminates the repetitive switch statement by using the generic helper.
func registerUintTypeByName(t *testing.T, typeName string) {
	t.Helper()

	switch typeName {
	case "BytePosition":
		// Adapter to convert func(uint32) BytePosition to func(uint) BytePosition
		adapter := func(n uint) BytePosition {
			return NewBytePosition(uint32(n))
		}
		registerUintTypeTestGeneric[BytePosition](t, typeName, adapter)
	case "TokenCount":
		registerUintTypeTestGeneric[TokenCount](t, typeName, NewTokenCount)
	case "ComplexityScore":
		// Adapter to convert func(uint16) ComplexityScore to func(uint) ComplexityScore
		adapter := func(n uint) ComplexityScore {
			return NewComplexityScore(uint16(n))
		}
		registerUintTypeTestGeneric[ComplexityScore](t, typeName, adapter)
	case "FileCount":
		registerUintTypeTestGeneric[FileCount](t, typeName, NewFileCount)
	case "CloneCount":
		registerUintTypeTestGeneric[CloneCount](t, typeName, NewCloneCount)
	default:
		t.Fatalf("Unknown uint type: %s", typeName)
	}
}

// createUintTypeTestFromName generates a test case for a uint-based type by type name.
// This helper further reduces boilerplate by using reflection to create the necessary functions.
func createUintTypeTestFromName(typeName string) TestCase {
	return TestCase{
		name: typeName,
		test: func(t *testing.T) {
			t.Helper()
			registerUintTypeByName(t, typeName)
		},
	}
}

// TestUintTypes consolidates tests for all simple uint wrapper types.
// Uses the createUintTypeTest helper to reduce boilerplate and eliminate code duplication.
func TestUintTypes(t *testing.T) {
	// Define test case structure
	var tests []TestCase
	for _, typeName := range uintTypeNames {
		tests = append(tests, createUintTypeTestFromName(typeName))
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// TestLineNumber_NewLineNumber tests the NewLineNumber constructor.
// This helper reduces boilerplate when creating tests for types with JSON support.
func registerJSONTestSuite[T comparable](
	t *testing.T,
	typeName string,
	marshalTests []jsonTest[T],
	unmarshalTests []jsonUnmarshalTest[T],
	roundTripValue T,
	marshalFunc func(T) ([]byte, error),
	unmarshalFunc func(*T, []byte) error,
) {
	t.Helper()

	t.Run(typeName+"_MarshalJSON", func(t *testing.T) {
		runJSONTests(t, marshalFunc, unmarshalFunc, marshalTests)
	})

	t.Run(typeName+"_UnmarshalJSON", func(t *testing.T) {
		runJSONUnmarshalTests(t, unmarshalFunc, unmarshalTests)
	})

	t.Run(typeName+"_RoundTrip", func(t *testing.T) {
		testJSONRoundTrip(t, roundTripValue, marshalFunc, unmarshalFunc)
	})
}

// registerStandardTypeTest creates and runs a standard test suite for types.
// This helper reduces boilerplate by running standard test functions and JSON test suite.
func registerStandardTypeTest(t *testing.T, typeName string, testFuncs ...func(*testing.T)) {
	t.Helper()

	for _, tf := range testFuncs {
		tf(t)
	}
}

// registerTypeTestSuite creates and runs a complete test suite for types with JSON support.
// This helper combines standard type tests with JSON marshaling/unmarshaling tests.
// It reduces boilerplate by consolidating registerStandardTypeTest and registerJSONTestSuite calls.
func registerTypeTestSuite[T comparable](
	t *testing.T,
	typeName string,
	testFuncs []func(*testing.T),
	marshalTests []jsonTest[T],
	unmarshalTests []jsonUnmarshalTest[T],
	roundTripValue T,
	marshalFunc func(T) ([]byte, error),
	unmarshalFunc func(*T, []byte) error,
) {
	t.Helper()
	registerStandardTypeTest(t, typeName, testFuncs...)
	registerJSONTestSuite(
		t,
		typeName,
		marshalTests,
		unmarshalTests,
		roundTripValue,
		marshalFunc,
		unmarshalFunc,
	)
}

// createTypeTestSuite creates a complete test registration for types with JSON support.
// This helper eliminates repetitive test boilerplate by consolidating:
// - Standard type tests (constructor, methods)
// - JSON marshaling tests
// - JSON unmarshaling tests
// - Round-trip tests.
func createTypeTestSuite[T comparable](
	typeName string,
	testFuncs []func(*testing.T),
	marshalTests []jsonTest[T],
	unmarshalTests []struct {
		name      string
		input     string
		want      T
		wantError bool
	},
	roundTripValue T,
	marshalFunc func(T) ([]byte, error),
	unmarshalFunc func(*T, []byte) error,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		registerTypeTestSuite(
			t,
			typeName,
			testFuncs,
			marshalTests,
			unmarshalTests,
			roundTripValue,
			marshalFunc,
			unmarshalFunc,
		)
	}
}

// createStandardUintJSONTests generates standard JSON test cases for uint-based types.
// This helper reduces boilerplate by providing common test patterns for numeric types.
// ValidValue is used for marshal and unmarshal valid tests, RoundTripValue is used for round-trip testing.
func createStandardUintJSONTests[T comparable](
	validValue T,
	validJSON string,
) (marshalTests []jsonTest[T], unmarshalTests []jsonUnmarshalTest[T], roundTripValue T,
) {
	var zero T

	marshalTests = []jsonTest[T]{
		{name: "valid value", input: validValue, want: validJSON, wantErr: false},
		{name: "zero should error", input: zero, want: "", wantErr: true},
	}
	unmarshalTests = []jsonUnmarshalTest[T]{
		{name: "valid JSON", input: validJSON, want: validValue, wantError: false},
		{name: "zero should error", input: `0`, want: zero, wantError: true},
		{name: "invalid JSON", input: `not-json`, want: zero, wantError: true},
	}
	roundTripValue = validValue

	return marshalTests, unmarshalTests, roundTripValue
}

// TestLineNumber_NewLineNumber tests the NewLineNumber constructor.
func TestLineNumber_NewLineNumber(t *testing.T) {
	// Adapter to convert func(uint16) (LineNumber, error) to func(uint) (LineNumber, error)
	adapter := func(n uint) (LineNumber, error) {
		return NewLineNumber(uint16(n))
	}
	registerBasicUintConstructorTest(t, "NewLineNumber",
		uint(1), uint(42),
		LineNumber(1), LineNumber(42),
		adapter)
}

// TestLineNumber_Uint tests the Uint method.
func TestLineNumber_Uint(t *testing.T) {
	line := LineNumber(42)
	if got := line.Uint(); got != 42 {
		t.Errorf("Uint() = %v, want %v", got, 42)
	}
}

// createStandardUintTypeTest creates a complete test suite for a uint-based type.
// This helper eliminates the remaining boilerplate by consolidating:
// - Standard test functions (constructor, Uint method)
// - JSON test generation via createStandardUintJSONTests
// - Type test suite registration
//
// Parameters:
//   - typeName: Name of the type being tested (e.g., "LineNumber")
//   - testFuncs: Slice of test functions to include (e.g., TestLineNumber_NewLineNumber, TestLineNumber_Uint)
//   - validValue: A valid value of the type for JSON tests
//   - validJSON: The JSON representation of validValue (e.g., "10")
//   - marshalFunc: Function to marshal the type to JSON
//   - unmarshalFunc: Function to unmarshal JSON to the type
func createStandardUintTypeTest[T comparable](
	typeName string,
	testFuncs []func(*testing.T),
	validValue T,
	validJSON string,
	marshalFunc func(T) ([]byte, error),
	unmarshalFunc func(*T, []byte) error,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		marshalTests, unmarshalTests, roundTripValue := createStandardUintJSONTests(
			validValue,
			validJSON,
		)
		createTypeTestSuite(
			typeName,
			testFuncs,
			marshalTests,
			unmarshalTests,
			roundTripValue,
			marshalFunc,
			unmarshalFunc,
		)(t)
	}
}

// TestLineNumber tests LineNumber type.
func TestLineNumber(t *testing.T) {
	createStandardUintTypeTest(
		"LineNumber",
		[]func(*testing.T){TestLineNumber_NewLineNumber, TestLineNumber_Uint},
		LineNumber(10),
		"10",
		jsonMarshalFunc[LineNumber](LineNumber(0)),
		jsonUnmarshalFuncPtr[LineNumber](),
	)(t)
}

// TestConfidence_NewConfidence tests the NewConfidence constructor.
func TestConfidence_NewConfidence(t *testing.T) {
	tests := []constructorTest[Confidence]{
		{
			name:      "valid confidence 0.0",
			input:     0.0,
			want:      Confidence(0.0),
			wantError: false,
		},
		{
			name:      "valid confidence 0.5",
			input:     0.5,
			want:      Confidence(0.5),
			wantError: false,
		},
		{
			name:      "valid confidence 1.0",
			input:     1.0,
			want:      Confidence(1.0),
			wantError: false,
		},
		{
			name:      "negative confidence should error",
			input:     -0.1,
			want:      0,
			wantError: true,
		},
		{
			name:      "confidence > 1.0 should error",
			input:     1.5,
			want:      0,
			wantError: true,
		},
	}
	runConstructorTests(t, "NewConfidence", tests, func(input any) (Confidence, error) {
		return NewConfidence(input.(float64))
	})
}

// TestConfidence_Float64 tests the Float64 method.
func TestConfidence_Float64(t *testing.T) {
	conf := Confidence(0.85)
	if got := conf.Float64(); got != 0.85 {
		t.Errorf("Confidence.Float64() = %v, want %v", got, 0.85)
	}
}

// TestConfidence_String tests the String method.
func TestConfidence_String(t *testing.T) {
	tests := []struct {
		name string
		conf Confidence
		want string
	}{
		{
			name: "0.0 confidence",
			conf: Confidence(0.0),
			want: "0.0%",
		},
		{
			name: "0.5 confidence",
			conf: Confidence(0.5),
			want: "50.0%",
		},
		{
			name: "1.0 confidence",
			conf: Confidence(1.0),
			want: "100.0%",
		},
		{
			name: "0.85 confidence",
			conf: Confidence(0.85),
			want: "85.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.AssertEqual(t, tt.conf.String(), tt.want, "Confidence.String()")
		})
	}
}

// TestConfidence tests Confidence type.
func TestConfidence(t *testing.T) {
	createTypeTestSuite(
		"Confidence",
		[]func(*testing.T){
			TestConfidence_NewConfidence,
			TestConfidence_Float64,
			TestConfidence_String,
		},
		[]jsonTest[Confidence]{
			{name: "valid confidence 0.5", input: Confidence(0.5), want: `0.5`, wantErr: false},
			{name: "valid confidence 1.0", input: Confidence(1.0), want: `1`, wantErr: false},
			{
				name:    "negative confidence should error",
				input:   Confidence(-0.1),
				want:    "",
				wantErr: true,
			},
			{
				name:    "confidence > 1.0 should error",
				input:   Confidence(1.5),
				want:    "",
				wantErr: true,
			},
		},
		[]struct {
			name      string
			input     string
			want      Confidence
			wantError bool
		}{
			{name: "valid JSON 0.5", input: `0.5`, want: Confidence(0.5), wantError: false},
			{name: "valid JSON 1.0", input: `1.0`, want: Confidence(1.0), wantError: false},
			{name: "negative should error", input: `-0.1`, want: Confidence(0), wantError: true},
			{name: "> 1.0 should error", input: `1.5`, want: Confidence(0), wantError: true},
			{name: "invalid JSON", input: `not-json`, want: Confidence(0), wantError: true},
		},
		Confidence(0.75),
		func(v Confidence) ([]byte, error) { return v.MarshalJSON() },
		jsonUnmarshalFuncPtr[Confidence](),
	)(
		t,
	)
}

// TestProcessingTime_NewProcessingTime tests the NewProcessingTime constructor.
func TestProcessingTime_NewProcessingTime(t *testing.T) {
	tests := []constructorTest[ProcessingTime]{
		{
			name:      "valid processing time",
			input:     uint(500),
			want:      ProcessingTime(500),
			wantError: false,
		},
		{name: "one millisecond", input: uint(1), want: ProcessingTime(1), wantError: false},
		{name: "zero should error", input: uint(0), want: ProcessingTime(0), wantError: true},
	}
	runConstructorTests(t, "NewProcessingTime", tests, func(input any) (ProcessingTime, error) {
		return NewProcessingTime(input.(uint))
	})
}

// TestProcessingTime_Uint tests the Uint method.
func TestProcessingTime_Uint(t *testing.T) {
	pt := ProcessingTime(5000)
	if got := pt.Uint(); got != 5000 {
		t.Errorf("Uint() = %v, want %v", got, 5000)
	}
}

// TestProcessingTime_String tests the String method.
func TestProcessingTime_String(t *testing.T) {
	tests := []struct {
		name  string
		value ProcessingTime
		want  string
	}{
		{
			name:  "milliseconds (< 1000ms)",
			value: ProcessingTime(500),
			want:  "500ms",
		},
		{
			name:  "seconds (< 60s)",
			value: ProcessingTime(5000),
			want:  "5s",
		},
		{
			name:  "minutes (< 60m)",
			value: ProcessingTime(180000), // 3 minutes
			want:  "3m",
		},
		{
			name:  "hours",
			value: ProcessingTime(7200000), // 2 hours
			want:  "2h",
		},
		{
			name:  "edge case: 999ms",
			value: ProcessingTime(999),
			want:  "999ms",
		},
		{
			name:  "edge case: 1000ms (1 second)",
			value: ProcessingTime(1000),
			want:  "1s",
		},
		{
			name:  "edge case: 59 seconds",
			value: ProcessingTime(59000),
			want:  "59s",
		},
		{
			name:  "edge case: 60 seconds (1 minute)",
			value: ProcessingTime(60000),
			want:  "1m",
		},
		{
			name:  "edge case: 59 minutes",
			value: ProcessingTime(3540000),
			want:  "59m",
		},
		{
			name:  "edge case: 60 minutes (1 hour)",
			value: ProcessingTime(3600000),
			want:  "1h",
		},
	}

	runStringMethodTests(t, tests, func(pt ProcessingTime) string { return pt.String() })
}

// TestProcessingTime tests ProcessingTime type.
func TestProcessingTime(t *testing.T) {
	createStandardUintTypeTest(
		"ProcessingTime",
		[]func(*testing.T){
			TestProcessingTime_NewProcessingTime,
			TestProcessingTime_Uint,
			TestProcessingTime_String,
		},
		ProcessingTime(500),
		"500",
		jsonMarshalFunc(ProcessingTime(0)),
		jsonUnmarshalFuncPtr[ProcessingTime](),
	)(
		t,
	)
}

// jsonMarshalFunc returns a marshal function for types with MarshalJSON method.
// This helper reduces boilerplate in test functions.
func jsonMarshalFunc[T interface{ MarshalJSON() ([]byte, error) }](_ T) func(T) ([]byte, error) {
	return func(v T) ([]byte, error) { return v.MarshalJSON() }
}

// jsonUnmarshalFuncPtr returns an unmarshal function for types with pointer receiver UnmarshalJSON method.
// This helper reduces boilerplate in test functions.
func jsonUnmarshalFuncPtr[T any]() func(*T, []byte) error {
	return func(v *T, data []byte) error {
		if unmarshaler, ok := any(v).(interface{ UnmarshalJSON(data []byte) error }); ok {
			return unmarshaler.UnmarshalJSON(data)
		}

		return fmt.Errorf("type %T does not implement UnmarshalJSON", v)
	}
}

// emptyStringErrorTest returns a test case that verifies empty string input causes an error.
func emptyStringErrorTest[T comparable]() constructorTest[T] {
	var zero T

	return constructorTest[T]{
		name:      "empty string should error",
		input:     "",
		want:      zero,
		wantError: true,
	}
}

// runStringMethodTests runs table-driven tests for a String method.
// This helper eliminates boilerplate for testing String() methods with multiple cases.
func runStringMethodTests[T any](t *testing.T, tests []struct {
	name  string
	value T
	want  string
}, stringFunc func(T) string,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringFunc(tt.value); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// registerBasicUintConstructorTest creates and runs basic tests for a uint-based constructor.
// It automatically includes valid test cases and a zero error test.
// Additional custom tests can be provided via the extraTests parameter.
// This helper reduces boilerplate for simple uint constructors with standard validation.
func registerBasicUintConstructorTest[T any](
	t *testing.T,
	constructorName string,
	validValue1, validValue2 uint,
	expectedValue1, expectedValue2 T,
	constructorFunc func(uint) (T, error),
	extraTests ...constructorTest[T],
) {
	t.Helper()

	tests := []constructorTest[T]{
		{
			name:      "valid " + constructorName,
			input:     validValue1,
			want:      expectedValue1,
			wantError: false,
		},
		{
			name:      "another valid " + constructorName,
			input:     validValue2,
			want:      expectedValue2,
			wantError: false,
		},
		{name: "zero should error", input: uint(0), want: *new(T), wantError: true},
	}
	tests = append(tests, extraTests...)
	runConstructorTests(t, constructorName, tests, func(input any) (T, error) {
		return constructorFunc(input.(uint))
	})
}

// registerStringConstructorTest creates and runs tests for a string-based constructor.
// This helper reduces boilerplate when creating tests for types constructed
// from strings with validation logic.
func registerStringConstructorTest[T comparable](
	t *testing.T,
	constructorName string,
	tests []constructorTest[T],
	constructorFunc func(string) (T, error),
) {
	t.Helper()
	runConstructorTests(t, constructorName, tests, func(input any) (T, error) {
		return constructorFunc(input.(string))
	})
}

// registerBasicStringConstructorTest creates and runs basic tests for a string-based constructor.
// It automatically includes a valid test case and an empty string error test.
// Additional custom tests can be provided via the extraTests parameter.
// This helper further reduces boilerplate for simple constructors with standard validation.
func registerBasicStringConstructorTest[T comparable](
	t *testing.T,
	constructorName, sampleValue string,
	expectedValue T,
	constructorFunc func(string) (T, error),
	extraTests ...constructorTest[T],
) {
	t.Helper()

	tests := []constructorTest[T]{
		{
			name:      "valid " + constructorName,
			input:     sampleValue,
			want:      expectedValue,
			wantError: false,
		},
		emptyStringErrorTest[T](),
	}
	tests = append(tests, extraTests...)
	registerStringConstructorTest(t, constructorName, tests, constructorFunc)
}

// TestCloneGroupID_NewCloneGroupID tests the NewCloneGroupID constructor.
func TestCloneGroupID_NewCloneGroupID(t *testing.T) {
	registerBasicStringConstructorTest(
		t,
		"NewCloneGroupID",
		"group-123",
		CloneGroupID("group-123"),
		NewCloneGroupID,
	)
}

// TestAnalysisID_NewAnalysisID tests the NewAnalysisID constructor.
func TestAnalysisID_NewAnalysisID(t *testing.T) {
	registerBasicStringConstructorTest(
		t,
		"NewAnalysisID",
		"analysis-456",
		AnalysisID("analysis-456"),
		NewAnalysisID,
	)
}

// TestFilepath_NewFilepath tests the NewFilepath constructor.
func TestFilepath_NewFilepath(t *testing.T) {
	registerBasicStringConstructorTest(
		t,
		"NewFilepath",
		"/path/to/file.go",
		Filepath("/path/to/file.go"),
		NewFilepath,
		constructorTest[Filepath]{
			name:      "relative path",
			input:     "./file.go",
			want:      Filepath("./file.go"),
			wantError: false,
		},
	)
}

// TestHash_NewHash tests the NewHash constructor.
func TestHash_NewHash(t *testing.T) {
	registerBasicStringConstructorTest(
		t,
		"NewHash",
		"a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		Hash("a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"),
		NewHash,
		constructorTest[Hash]{
			name:      "short hash",
			input:     "abc123",
			want:      Hash("abc123"),
			wantError: false,
		},
	)
}

// TestThreshold_NewThreshold tests the NewThreshold constructor.
func TestThreshold_NewThreshold(t *testing.T) {
	registerBasicUintConstructorTest(t, "NewThreshold",
		uint(15), uint(30),
		Threshold(15), Threshold(30),
		NewThreshold)
}

// TestThreshold_Uint tests the Uint method.
func TestThreshold_Uint(t *testing.T) {
	th := Threshold(30)
	if got := th.Uint(); got != 30 {
		t.Errorf("Uint() = %v, want %v", got, 30)
	}
}

// TestThreshold tests Threshold type.
func TestThreshold(t *testing.T) {
	createStandardUintTypeTest(
		"Threshold",
		[]func(*testing.T){TestThreshold_NewThreshold, TestThreshold_Uint},
		Threshold(15),
		"15",
		jsonMarshalFunc(Threshold(0)),
		jsonUnmarshalFuncPtr[Threshold](),
	)(t)
}
