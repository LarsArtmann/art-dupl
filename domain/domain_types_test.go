package domain

import (
	stderrors "errors"
	"testing"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
)

// testUintType is a helper for testing uint-based types with New*, Uint(), and RoundTrip methods.
type testUintType[T comparable] struct {
	newFunc       func(uint) T
	uintFunc      func(T) uint
	jsonMarshal   func(T) ([]byte, error)
	jsonUnmarshal func(*T, []byte) error
}

// testUintTypeSuite runs the standard test suite for uint-based types.
func testUintTypeSuite[T comparable](t *testing.T, typeName string, tt testUintType[T]) {
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
				if got != tc.want {
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
		if err := tt.jsonUnmarshal(&result, data); err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}
		if result != original {
			t.Errorf("Round trip failed: %v != %v", result, original)
		}
	})
}

// registerUintTypeTest creates and runs the standard test suite for uint-based types.
// This is a convenience wrapper around testUintTypeSuite that takes individual function parameters.
func registerUintTypeTest[T comparable](t *testing.T, typeName string, newFunc func(uint) T, uintFunc func(T) uint, jsonMarshal func(T) ([]byte, error), jsonUnmarshal func(*T, []byte) error) {
	t.Helper()
	testUintTypeSuite(t, typeName, testUintType[T]{
		newFunc:       newFunc,
		uintFunc:      uintFunc,
		jsonMarshal:   jsonMarshal,
		jsonUnmarshal: jsonUnmarshal,
	})
}

// testJSONRoundTrip is a helper for testing JSON marshaling and unmarshaling.
func testJSONRoundTrip[T comparable](t *testing.T, original T, marshal func(T) ([]byte, error), unmarshal func(*T, []byte) error) {
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
type constructorTest[T comparable] struct {
	name      string
	input     any
	want      T
	wantError bool
}

// runConstructorTests runs a series of constructor tests with error checking.
func runConstructorTests[T comparable](t *testing.T, constructorName string, tests []constructorTest[T], newFunc func(any) (T, error)) {
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
				if got != tt.want {
					t.Errorf("%s() = %v, want %v", constructorName, got, tt.want)
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

// runJSONTests runs JSON marshal/unmarshal tests.
func runJSONTests[T comparable](t *testing.T, marshal func(T) ([]byte, error), unmarshal func(*T, []byte) error, tests []jsonTest[T]) {
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
func runJSONUnmarshalTests[T comparable](t *testing.T, unmarshal func(*T, []byte) error, tests []struct {
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

// createUintTypeTestRegistration creates a test registration function for a uint-based type.
// This helper eliminates repetitive closure boilerplate in test tables.
func createUintTypeTestRegistration[T comparable](
	typeName string,
	newFunc func(uint) T,
	uintFunc func(T) uint,
	jsonMarshal func(T) ([]byte, error),
	jsonUnmarshal func(*T, []byte) error,
) func(*testing.T) {
	return func(t *testing.T) {
		registerUintTypeTest(t, typeName, newFunc, uintFunc, jsonMarshal, jsonUnmarshal)
	}
}

// createUintTestCase creates a complete test case struct for a uint-based type.
// This helper eliminates the repetitive boilerplate of manually structuring test case entries.
func createUintTestCase[T comparable](
	typeName string,
	newFunc func(uint) T,
	uintFunc func(T) uint,
	jsonMarshal func(T) ([]byte, error),
	jsonUnmarshal func(*T, []byte) error,
) struct {
	name string
	test func(*testing.T)
} {
	return struct {
		name string
		test func(*testing.T)
	}{
		name: typeName,
		test: createUintTypeTestRegistration(typeName, newFunc, uintFunc, jsonMarshal, jsonUnmarshal),
	}
}

// TestUintTypes consolidates tests for all simple uint wrapper types.
// This approach eliminates duplicate test functions by using the registerUintTypeTest helper.
func TestUintTypes(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		createUintTestCase("BytePosition",
			func(u uint) BytePosition { return BytePosition(u) },
			func(bp BytePosition) uint { return bp.Uint() },
			func(bp BytePosition) ([]byte, error) { return bp.MarshalJSON() },
			func(bp *BytePosition, data []byte) error { return bp.UnmarshalJSON(data) }),
		createUintTestCase("TokenCount",
			func(u uint) TokenCount { return TokenCount(u) },
			func(tc TokenCount) uint { return tc.Uint() },
			func(tc TokenCount) ([]byte, error) { return tc.MarshalJSON() },
			func(tc *TokenCount, data []byte) error { return tc.UnmarshalJSON(data) }),
		createUintTestCase("ComplexityScore",
			func(u uint) ComplexityScore { return ComplexityScore(u) },
			func(cs ComplexityScore) uint { return cs.Uint() },
			func(cs ComplexityScore) ([]byte, error) { return cs.MarshalJSON() },
			func(cs *ComplexityScore, data []byte) error { return cs.UnmarshalJSON(data) }),
		createUintTestCase("FileCount",
			func(u uint) FileCount { return FileCount(u) },
			func(fc FileCount) uint { return fc.Uint() },
			func(fc FileCount) ([]byte, error) { return fc.MarshalJSON() },
			func(fc *FileCount, data []byte) error { return fc.UnmarshalJSON(data) }),
		createUintTestCase("CloneCount",
			func(u uint) CloneCount { return CloneCount(u) },
			func(cc CloneCount) uint { return cc.Uint() },
			func(cc CloneCount) ([]byte, error) { return cc.MarshalJSON() },
			func(cc *CloneCount, data []byte) error { return cc.UnmarshalJSON(data) }),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// TestCloneID_NewCloneID tests the NewCloneID constructor.
func TestCloneID_NewCloneID(t *testing.T) {
	tests := []constructorTest[CloneID]{
		{
			name:      "valid clone ID",
			input:     "clone-123",
			want:      CloneID("clone-123"),
			wantError: false,
		},
		{
			name:      "empty string should error",
			input:     "",
			want:      "",
			wantError: true,
		},
		{
			name:      "ID with special characters",
			input:     "clone-123_abc",
			want:      CloneID("clone-123_abc"),
			wantError: false,
		},
		{
			name:      "ID with spaces",
			input:     "clone 123",
			want:      CloneID("clone 123"),
			wantError: false,
		},
	}
	runConstructorTests(t, "NewCloneID", tests, func(input any) (CloneID, error) {
		return NewCloneID(input.(string))
	})
}

// TestCloneID_String tests the String method.
func TestCloneID_String(t *testing.T) {
	id := CloneID("test-id")
	if got := id.String(); got != "test-id" {
		t.Errorf("String() = %v, want %v", got, "test-id")
	}
}

// registerJSONTestSuite creates and runs a complete test suite for JSON marshaling/unmarshaling.
// This helper reduces boilerplate when creating tests for types with JSON support.
func registerJSONTestSuite[T comparable](t *testing.T, typeName string, marshalTests []jsonTest[T], unmarshalTests []struct {
	name      string
	input     string
	want      T
	wantError bool
}, roundTripValue T, marshalFunc func(T) ([]byte, error), unmarshalFunc func(*T, []byte) error) {
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
func registerTypeTestSuite[T comparable](t *testing.T, typeName string, testFuncs []func(*testing.T), marshalTests []jsonTest[T], unmarshalTests []struct {
	name      string
	input     string
	want      T
	wantError bool
}, roundTripValue T, marshalFunc func(T) ([]byte, error), unmarshalFunc func(*T, []byte) error) {
	t.Helper()
	registerStandardTypeTest(t, typeName, testFuncs...)
	registerJSONTestSuite(t, typeName, marshalTests, unmarshalTests, roundTripValue, marshalFunc, unmarshalFunc)
}

// createStringTypeTestSuite creates a complete test registration for string-based types.
// This helper eliminates repetitive test boilerplate by consolidating:
// - Standard type tests (constructor, methods)
// - JSON marshaling tests
// - JSON unmarshaling tests
// - Round-trip tests
func createStringTypeTestSuite[T comparable](
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
		registerTypeTestSuite(t, typeName, testFuncs, marshalTests, unmarshalTests, roundTripValue, marshalFunc, unmarshalFunc)
	}
}

// createUintTypeTestSuite creates a complete test registration for uint-based types.
// This helper eliminates repetitive test boilerplate for numeric types.
func createUintTypeTestSuite[T comparable](
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
		registerTypeTestSuite(t, typeName, testFuncs, marshalTests, unmarshalTests, roundTripValue, marshalFunc, unmarshalFunc)
	}
}

// createStandardUintJSONTests generates standard JSON test cases for uint-based types.
// This helper reduces boilerplate by providing common test patterns for numeric types.
// ValidValue is used for marshal and unmarshal valid tests, RoundTripValue is used for round-trip testing.
func createStandardUintJSONTests[T comparable](
	validValue T,
	validJSON string,
) (marshalTests []jsonTest[T], unmarshalTests []struct {
	name      string
	input     string
	want      T
	wantError bool
}, roundTripValue T) {
	var zero T
	marshalTests = []jsonTest[T]{
		{name: "valid value", input: validValue, want: validJSON, wantErr: false},
		{name: "zero should error", input: zero, want: "", wantErr: true},
	}
	unmarshalTests = []struct {
		name      string
		input     string
		want      T
		wantError bool
	}{
		{name: "valid JSON", input: validJSON, want: validValue, wantError: false},
		{name: "zero should error", input: `0`, want: zero, wantError: true},
		{name: "invalid JSON", input: `not-json`, want: zero, wantError: true},
	}
	roundTripValue = validValue
	return
}

// createFloatTypeTestSuite creates a complete test registration for float-based types.
// This helper eliminates repetitive test boilerplate for floating-point types.
func createFloatTypeTestSuite[T comparable](
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
		registerTypeTestSuite(t, typeName, testFuncs, marshalTests, unmarshalTests, roundTripValue, marshalFunc, unmarshalFunc)
	}
}

// TestCloneID tests CloneID type.
func TestCloneID(t *testing.T) {
	createStringTypeTestSuite("CloneID",
		[]func(*testing.T){TestCloneID_NewCloneID, TestCloneID_String},
		[]jsonTest[CloneID]{
			{name: "valid clone ID", input: CloneID("clone-123"), want: `"clone-123"`, wantErr: false},
			{name: "empty ID should error", input: CloneID(""), want: "", wantErr: true},
		},
		[]struct {
			name      string
			input     string
			want      CloneID
			wantError bool
		}{
			{name: "valid JSON", input: `"clone-123"`, want: CloneID("clone-123"), wantError: false},
			{name: "empty JSON string should error", input: `""`, want: CloneID(""), wantError: true},
			{name: "invalid JSON", input: `not-json`, want: CloneID(""), wantError: true},
		},
		CloneID("clone-456"),
		func(id CloneID) ([]byte, error) { return id.MarshalJSON() },
		func(id *CloneID, data []byte) error { return id.UnmarshalJSON(data) },
	)(t)
}

// TestLineNumber_NewLineNumber tests the NewLineNumber constructor.
func TestLineNumber_NewLineNumber(t *testing.T) {
	tests := []constructorTest[LineNumber]{
		{
			name:      "valid line number",
			input:     uint(1),
			want:      LineNumber(1),
			wantError: false,
		},
		{
			name:      "another valid line number",
			input:     uint(42),
			want:      LineNumber(42),
			wantError: false,
		},
		{
			name:      "zero should error",
			input:     uint(0),
			want:      0,
			wantError: true,
		},
	}
	runConstructorTests(t, "NewLineNumber", tests, func(input any) (LineNumber, error) {
		return NewLineNumber(input.(uint))
	})
}

// TestLineNumber_Uint tests the Uint method.
func TestLineNumber_Uint(t *testing.T) {
	line := LineNumber(42)
	if got := line.Uint(); got != 42 {
		t.Errorf("Uint() = %v, want %v", got, 42)
	}
}

// TestLineNumber tests LineNumber type.
func TestLineNumber(t *testing.T) {
	marshalTests, unmarshalTests, roundTripValue := createStandardUintJSONTests(LineNumber(10), "10")
	createUintTypeTestSuite("LineNumber",
		[]func(*testing.T){TestLineNumber_NewLineNumber, TestLineNumber_Uint},
		marshalTests,
		unmarshalTests,
		roundTripValue,
		func(line LineNumber) ([]byte, error) { return line.MarshalJSON() },
		func(line *LineNumber, data []byte) error { return line.UnmarshalJSON(data) },
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
		t.Errorf("Float64() = %v, want %v", got, 0.85)
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
			if got := tt.conf.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConfidence tests Confidence type.
func TestConfidence(t *testing.T) {
	createFloatTypeTestSuite("Confidence",
		[]func(*testing.T){TestConfidence_NewConfidence, TestConfidence_Float64, TestConfidence_String},
		[]jsonTest[Confidence]{
			{name: "valid confidence 0.5", input: Confidence(0.5), want: `0.5`, wantErr: false},
			{name: "valid confidence 1.0", input: Confidence(1.0), want: `1`, wantErr: false},
			{name: "negative confidence should error", input: Confidence(-0.1), want: "", wantErr: true},
			{name: "confidence > 1.0 should error", input: Confidence(1.5), want: "", wantErr: true},
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
		func(c Confidence) ([]byte, error) { return c.MarshalJSON() },
		func(c *Confidence, data []byte) error { return c.UnmarshalJSON(data) },
	)(t)
}

// TestProcessingTime_NewProcessingTime tests the NewProcessingTime constructor.
func TestProcessingTime_NewProcessingTime(t *testing.T) {
	tests := []constructorTest[ProcessingTime]{
		{name: "valid processing time", input: uint(500), want: ProcessingTime(500), wantError: false},
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
		name string
		pt   ProcessingTime
		want string
	}{
		{
			name: "milliseconds (< 1000ms)",
			pt:   ProcessingTime(500),
			want: "500ms",
		},
		{
			name: "seconds (< 60s)",
			pt:   ProcessingTime(5000),
			want: "5s",
		},
		{
			name: "minutes (< 60m)",
			pt:   ProcessingTime(180000), // 3 minutes
			want: "3m",
		},
		{
			name: "hours",
			pt:   ProcessingTime(7200000), // 2 hours
			want: "2h",
		},
		{
			name: "edge case: 999ms",
			pt:   ProcessingTime(999),
			want: "999ms",
		},
		{
			name: "edge case: 1000ms (1 second)",
			pt:   ProcessingTime(1000),
			want: "1s",
		},
		{
			name: "edge case: 59 seconds",
			pt:   ProcessingTime(59000),
			want: "59s",
		},
		{
			name: "edge case: 60 seconds (1 minute)",
			pt:   ProcessingTime(60000),
			want: "1m",
		},
		{
			name: "edge case: 59 minutes",
			pt:   ProcessingTime(3540000),
			want: "59m",
		},
		{
			name: "edge case: 60 minutes (1 hour)",
			pt:   ProcessingTime(3600000),
			want: "1h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pt.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestProcessingTime tests ProcessingTime type.
func TestProcessingTime(t *testing.T) {
	marshalTests, unmarshalTests, roundTripValue := createStandardUintJSONTests(ProcessingTime(500), "500")
	createUintTypeTestSuite("ProcessingTime",
		[]func(*testing.T){TestProcessingTime_NewProcessingTime, TestProcessingTime_Uint, TestProcessingTime_String},
		marshalTests,
		unmarshalTests,
		roundTripValue,
		func(pt ProcessingTime) ([]byte, error) { return pt.MarshalJSON() },
		func(pt *ProcessingTime, data []byte) error { return pt.UnmarshalJSON(data) },
	)(t)
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

// registerStringConstructorTest creates and runs tests for a string-based constructor.
// This helper reduces boilerplate when creating tests for types constructed
// from strings with validation logic.
func registerStringConstructorTest[T comparable](t *testing.T, constructorName string, tests []constructorTest[T], constructorFunc func(string) (T, error)) {
	t.Helper()
	runConstructorTests(t, constructorName, tests, func(input any) (T, error) {
		return constructorFunc(input.(string))
	})
}

// registerBasicStringConstructorTest creates and runs basic tests for a string-based constructor.
// It automatically includes a valid test case and an empty string error test.
// Additional custom tests can be provided via the extraTests parameter.
// This helper further reduces boilerplate for simple constructors with standard validation.
func registerBasicStringConstructorTest[T comparable](t *testing.T, constructorName string, sampleValue string, expectedValue T, constructorFunc func(string) (T, error), extraTests ...constructorTest[T]) {
	t.Helper()
	tests := []constructorTest[T]{
		{name: "valid " + constructorName, input: sampleValue, want: expectedValue, wantError: false},
		emptyStringErrorTest[T](),
	}
	tests = append(tests, extraTests...)
	registerStringConstructorTest(t, constructorName, tests, constructorFunc)
}

// TestCloneGroupID_NewCloneGroupID tests the NewCloneGroupID constructor.
func TestCloneGroupID_NewCloneGroupID(t *testing.T) {
	registerBasicStringConstructorTest(t, "NewCloneGroupID", "group-123", CloneGroupID("group-123"), NewCloneGroupID)
}

// TestAnalysisID_NewAnalysisID tests the NewAnalysisID constructor.
func TestAnalysisID_NewAnalysisID(t *testing.T) {
	registerBasicStringConstructorTest(t, "NewAnalysisID", "analysis-456", AnalysisID("analysis-456"), NewAnalysisID)
}

// TestFilepath_NewFilepath tests the NewFilepath constructor.
func TestFilepath_NewFilepath(t *testing.T) {
	registerStringConstructorTest(t, "NewFilepath", []constructorTest[Filepath]{
		{name: "valid filepath", input: "/path/to/file.go", want: Filepath("/path/to/file.go"), wantError: false},
		{name: "relative path", input: "./file.go", want: Filepath("./file.go"), wantError: false},
		emptyStringErrorTest[Filepath](),
	}, NewFilepath)
}

// TestHash_NewHash tests the NewHash constructor.
func TestHash_NewHash(t *testing.T) {
	registerStringConstructorTest(t, "NewHash", []constructorTest[Hash]{
		{name: "valid SHA256 hash", input: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e", want: Hash("a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"), wantError: false},
		{name: "short hash", input: "abc123", want: Hash("abc123"), wantError: false},
		emptyStringErrorTest[Hash](),
	}, NewHash)
}

// TestThreshold_NewThreshold tests the NewThreshold constructor.
func TestThreshold_NewThreshold(t *testing.T) {
	tests := []constructorTest[Threshold]{
		{name: "valid threshold", input: uint(15), want: Threshold(15), wantError: false},
		{name: "zero should error", input: uint(0), want: 0, wantError: true},
	}
	runConstructorTests(t, "NewThreshold", tests, func(input any) (Threshold, error) {
		return NewThreshold(input.(uint))
	})
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
	marshalTests, unmarshalTests, roundTripValue := createStandardUintJSONTests(Threshold(15), "15")
	createUintTypeTestSuite("Threshold",
		[]func(*testing.T){TestThreshold_NewThreshold, TestThreshold_Uint},
		marshalTests,
		unmarshalTests,
		roundTripValue,
		func(t Threshold) ([]byte, error) { return t.MarshalJSON() },
		func(t *Threshold, data []byte) error { return t.UnmarshalJSON(data) },
	)(t)
}
