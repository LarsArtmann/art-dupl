package domain

import (
	stderrors "errors"
	"testing"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
)

// testUintType is a helper for testing uint-based types with New*, Uint(), and RoundTrip methods.
type testUintType[T comparable] struct {
	newFunc     func(uint) T
	uintFunc    func(T) uint
	jsonMarshal func(T) ([]byte, error)
	jsonUnmarshal func(*T, []byte) error
}

// testUintTypeSuite runs the standard test suite for uint-based types.
func testUintTypeSuite[T comparable](t *testing.T, typeName string, tt testUintType[T]) {
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

// testJSONRoundTrip is a helper for testing JSON marshaling and unmarshaling.
func testJSONRoundTrip[T comparable](t *testing.T, original T, marshal func(T) ([]byte, error), unmarshal func(*T, []byte) error) {
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

// testConstructorWithError is a helper for testing constructors that may return errors.
type constructorTest[T comparable] struct {
	name      string
	input     any
	want      T
	wantError bool
}

// runConstructorTests runs a series of constructor tests with error checking.
func runConstructorTests[T comparable](t *testing.T, constructorName string, tests []constructorTest[T], newFunc func(any) (T, error)) {
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

// testJSONMarshalUnmarshal is a helper for testing JSON marshaling/unmarshaling.
type jsonTest[T comparable] struct {
	name    string
	input   T
	want    string
	wantErr bool
}

// runJSONTests runs JSON marshal/unmarshal tests.
func runJSONTests[T comparable](t *testing.T, marshal func(T) ([]byte, error), unmarshal func(*T, []byte) error, tests []jsonTest[T]) {
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
}) {
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

// TestCloneID_MarshalJSON tests JSON marshaling.
func TestCloneID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		id      CloneID
		want    string
		wantErr bool
	}{
		{
			name:    "valid clone ID",
			id:      CloneID("clone-123"),
			want:    `"clone-123"`,
			wantErr: false,
		},
		{
			name:    "empty ID should error",
			id:      CloneID(""),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.id.MarshalJSON()

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

// TestCloneID_UnmarshalJSON tests JSON unmarshaling.
func TestCloneID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    CloneID
		wantError bool
	}{
		{
			name:      "valid JSON",
			input:     `"clone-123"`,
			wantID:    CloneID("clone-123"),
			wantError: false,
		},
		{
			name:      "empty JSON string should error",
			input:     `""`,
			wantID:    "",
			wantError: true,
		},
		{
			name:      "invalid JSON",
			input:     `not-json`,
			wantID:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID CloneID
			gotErr := gotID.UnmarshalJSON([]byte(tt.input))

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
				if gotID != tt.wantID {
					t.Errorf("UnmarshalJSON() = %v, want %v", gotID, tt.wantID)
				}
			}
		})
	}
}

// TestCloneID_RoundTrip tests JSON marshaling and unmarshaling round trip.
func TestCloneID_RoundTrip(t *testing.T) {
	original := CloneID("clone-456")

	// Marshal
	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	// Unmarshal
	var result CloneID
	if err := result.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	// Verify round trip
	if result != original {
		t.Errorf("Round trip failed: %v != %v", result, original)
	}
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

// TestLineNumber_MarshalJSON tests JSON marshaling.
func TestLineNumber_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		line    LineNumber
		want    string
		wantErr bool
	}{
		{
			name:    "valid line number",
			line:    LineNumber(10),
			want:    `10`,
			wantErr: false,
		},
		{
			name:    "zero line number should error",
			line:    LineNumber(0),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.line.MarshalJSON()

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

// TestLineNumber_UnmarshalJSON tests JSON unmarshaling.
func TestLineNumber_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      LineNumber
		wantError bool
	}{
		{
			name:      "valid JSON",
			input:     `42`,
			want:      LineNumber(42),
			wantError: false,
		},
		{
			name:      "zero should error",
			input:     `0`,
			want:      0,
			wantError: true,
		},
		{
			name:      "invalid JSON",
			input:     `not-json`,
			want:      0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got LineNumber
			gotErr := got.UnmarshalJSON([]byte(tt.input))

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

// TestLineNumber_RoundTrip tests JSON marshaling and unmarshaling round trip.
func TestLineNumber_RoundTrip(t *testing.T) {
	original := LineNumber(123)

	// Marshal
	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	// Unmarshal
	var result LineNumber
	if err := result.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	// Verify round trip
	if result != original {
		t.Errorf("Round trip failed: %v != %v", result, original)
	}
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

// TestConfidence_MarshalJSON tests JSON marshaling.
func TestConfidence_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		conf    Confidence
		want    string
		wantErr bool
	}{
		{
			name:    "valid confidence 0.5",
			conf:    Confidence(0.5),
			want:    `0.5`,
			wantErr: false,
		},
		{
			name:    "valid confidence 1.0",
			conf:    Confidence(1.0),
			want:    `1`,
			wantErr: false,
		},
		{
			name:    "negative confidence should error",
			conf:    Confidence(-0.1),
			want:    "",
			wantErr: true,
		},
		{
			name:    "confidence > 1.0 should error",
			conf:    Confidence(1.5),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.conf.MarshalJSON()

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

// TestConfidence_UnmarshalJSON tests JSON unmarshaling.
func TestConfidence_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Confidence
		wantError bool
	}{
		{
			name:      "valid JSON 0.5",
			input:     `0.5`,
			want:      Confidence(0.5),
			wantError: false,
		},
		{
			name:      "valid JSON 1.0",
			input:     `1.0`,
			want:      Confidence(1.0),
			wantError: false,
		},
		{
			name:      "negative should error",
			input:     `-0.1`,
			want:      0,
			wantError: true,
		},
		{
			name:      "> 1.0 should error",
			input:     `1.5`,
			want:      0,
			wantError: true,
		},
		{
			name:      "invalid JSON",
			input:     `not-json`,
			want:      0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Confidence
			gotErr := got.UnmarshalJSON([]byte(tt.input))

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

// TestConfidence_RoundTrip tests JSON marshaling and unmarshaling round trip.
func TestConfidence_RoundTrip(t *testing.T) {
	original := Confidence(0.75)

	// Marshal
	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	// Unmarshal
	var result Confidence
	if err := result.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	// Verify round trip
	if result != original {
		t.Errorf("Round trip failed: %v != %v", result, original)
	}
}

// TestProcessingTime_NewProcessingTime tests the NewProcessingTime constructor.
func TestProcessingTime_NewProcessingTime(t *testing.T) {
	tests := []struct {
		name      string
		input     uint
		want      ProcessingTime
		wantError bool
	}{
		{
			name:      "valid processing time",
			input:     500,
			want:      ProcessingTime(500),
			wantError: false,
		},
		{
			name:      "one millisecond",
			input:     1,
			want:      ProcessingTime(1),
			wantError: false,
		},
		{
			name:      "zero should error",
			input:     0,
			want:      0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewProcessingTime(tt.input)

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("NewProcessingTime() expected error, got nil")
					return
				}
				var validationErr *duplerrors.DuplError
				if !stderrors.As(gotErr, &validationErr) {
					t.Errorf("NewProcessingTime() expected ValidationError, got %T", gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("NewProcessingTime() unexpected error: %v", gotErr)
					return
				}
				if got != tt.want {
					t.Errorf("NewProcessingTime() = %v, want %v", got, tt.want)
				}
			}
		})
	}
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

// TestProcessingTime_MarshalJSON tests JSON marshaling.
func TestProcessingTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		pt      ProcessingTime
		want    string
		wantErr bool
	}{
		{
			name:    "valid processing time",
			pt:      ProcessingTime(500),
			want:    `500`,
			wantErr: false,
		},
		{
			name:    "zero should error",
			pt:      ProcessingTime(0),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.pt.MarshalJSON()

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

// TestProcessingTime_UnmarshalJSON tests JSON unmarshaling.
func TestProcessingTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ProcessingTime
		wantError bool
	}{
		{
			name:      "valid JSON",
			input:     `500`,
			want:      ProcessingTime(500),
			wantError: false,
		},
		{
			name:      "zero should error",
			input:     `0`,
			want:      0,
			wantError: true,
		},
		{
			name:      "invalid JSON",
			input:     `not-json`,
			want:      0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ProcessingTime
			gotErr := got.UnmarshalJSON([]byte(tt.input))

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

// TestProcessingTime_RoundTrip tests JSON marshaling and unmarshaling round trip.
func TestProcessingTime_RoundTrip(t *testing.T) {
	original := ProcessingTime(5000)

	// Marshal
	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	// Unmarshal
	var result ProcessingTime
	if err := result.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	// Verify round trip
	if result != original {
		t.Errorf("Round trip failed: %v != %v", result, original)
	}
}

// TestCloneGroupID_NewCloneGroupID tests the NewCloneGroupID constructor.
func TestCloneGroupID_NewCloneGroupID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      CloneGroupID
		wantError bool
	}{
		{
			name:      "valid clone group ID",
			input:     "group-123",
			want:      CloneGroupID("group-123"),
			wantError: false,
		},
		{
			name:      "empty string should error",
			input:     "",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewCloneGroupID(tt.input)

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("NewCloneGroupID() expected error, got nil")
					return
				}
				var validationErr *duplerrors.DuplError
				if !stderrors.As(gotErr, &validationErr) {
					t.Errorf("NewCloneGroupID() expected ValidationError, got %T", gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("NewCloneGroupID() unexpected error: %v", gotErr)
					return
				}
				if got != tt.want {
					t.Errorf("NewCloneGroupID() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestAnalysisID_NewAnalysisID tests the NewAnalysisID constructor.
func TestAnalysisID_NewAnalysisID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      AnalysisID
		wantError bool
	}{
		{
			name:      "valid analysis ID",
			input:     "analysis-456",
			want:      AnalysisID("analysis-456"),
			wantError: false,
		},
		{
			name:      "empty string should error",
			input:     "",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewAnalysisID(tt.input)

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("NewAnalysisID() expected error, got nil")
					return
				}
				var validationErr *duplerrors.DuplError
				if !stderrors.As(gotErr, &validationErr) {
					t.Errorf("NewAnalysisID() expected ValidationError, got %T", gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("NewAnalysisID() unexpected error: %v", gotErr)
					return
				}
				if got != tt.want {
					t.Errorf("NewAnalysisID() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestFilepath_NewFilepath tests the NewFilepath constructor.
func TestFilepath_NewFilepath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Filepath
		wantError bool
	}{
		{
			name:      "valid filepath",
			input:     "/path/to/file.go",
			want:      Filepath("/path/to/file.go"),
			wantError: false,
		},
		{
			name:      "relative path",
			input:     "./file.go",
			want:      Filepath("./file.go"),
			wantError: false,
		},
		{
			name:      "empty string should error",
			input:     "",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewFilepath(tt.input)

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("NewFilepath() expected error, got nil")
					return
				}
				var validationErr *duplerrors.DuplError
				if !stderrors.As(gotErr, &validationErr) {
					t.Errorf("NewFilepath() expected ValidationError, got %T", gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("NewFilepath() unexpected error: %v", gotErr)
					return
				}
				if got != tt.want {
					t.Errorf("NewFilepath() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestHash_NewHash tests the NewHash constructor.
func TestHash_NewHash(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Hash
		wantError bool
	}{
		{
			name:      "valid SHA256 hash",
			input:     "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
			want:      Hash("a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"),
			wantError: false,
		},
		{
			name:      "short hash",
			input:     "abc123",
			want:      Hash("abc123"),
			wantError: false,
		},
		{
			name:      "empty string should error",
			input:     "",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewHash(tt.input)

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("NewHash() expected error, got nil")
					return
				}
				var validationErr *duplerrors.DuplError
				if !stderrors.As(gotErr, &validationErr) {
					t.Errorf("NewHash() expected ValidationError, got %T", gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("NewHash() unexpected error: %v", gotErr)
					return
				}
				if got != tt.want {
					t.Errorf("NewHash() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestBytePosition tests BytePosition type.
func TestBytePosition(t *testing.T) {
	testUintTypeSuite(t, "BytePosition", testUintType[BytePosition]{
		newFunc: func(u uint) BytePosition { return BytePosition(u) },
		uintFunc: func(bp BytePosition) uint { return bp.Uint() },
		jsonMarshal: func(bp BytePosition) ([]byte, error) { return bp.MarshalJSON() },
		jsonUnmarshal: func(bp *BytePosition, data []byte) error { return bp.UnmarshalJSON(data) },
	})
}

// TestTokenCount tests TokenCount type.
func TestTokenCount(t *testing.T) {
	testUintTypeSuite(t, "TokenCount", testUintType[TokenCount]{
		newFunc: func(u uint) TokenCount { return TokenCount(u) },
		uintFunc: func(tc TokenCount) uint { return tc.Uint() },
		jsonMarshal: func(tc TokenCount) ([]byte, error) { return tc.MarshalJSON() },
		jsonUnmarshal: func(tc *TokenCount, data []byte) error { return tc.UnmarshalJSON(data) },
	})
}

// TestComplexityScore tests ComplexityScore type.
func TestComplexityScore(t *testing.T) {
	testUintTypeSuite(t, "ComplexityScore", testUintType[ComplexityScore]{
		newFunc: func(u uint) ComplexityScore { return ComplexityScore(u) },
		uintFunc: func(cs ComplexityScore) uint { return cs.Uint() },
		jsonMarshal: func(cs ComplexityScore) ([]byte, error) { return cs.MarshalJSON() },
		jsonUnmarshal: func(cs *ComplexityScore, data []byte) error { return cs.UnmarshalJSON(data) },
	})
}

// TestFileCount tests FileCount type.
func TestFileCount(t *testing.T) {
	testUintTypeSuite(t, "FileCount", testUintType[FileCount]{
		newFunc: func(u uint) FileCount { return FileCount(u) },
		uintFunc: func(fc FileCount) uint { return fc.Uint() },
		jsonMarshal: func(fc FileCount) ([]byte, error) { return fc.MarshalJSON() },
		jsonUnmarshal: func(fc *FileCount, data []byte) error { return fc.UnmarshalJSON(data) },
	})
}

// TestCloneCount tests CloneCount type.
func TestCloneCount(t *testing.T) {
	testUintTypeSuite(t, "CloneCount", testUintType[CloneCount]{
		newFunc: func(u uint) CloneCount { return CloneCount(u) },
		uintFunc: func(cc CloneCount) uint { return cc.Uint() },
		jsonMarshal: func(cc CloneCount) ([]byte, error) { return cc.MarshalJSON() },
		jsonUnmarshal: func(cc *CloneCount, data []byte) error { return cc.UnmarshalJSON(data) },
	})
}

// TestThreshold tests Threshold type.
func TestThreshold(t *testing.T) {
	t.Run("NewThreshold", func(t *testing.T) {
		tests := []struct {
			name      string
			input     uint
			want      Threshold
			wantError bool
		}{
			{
				name:      "valid threshold",
				input:     15,
				want:      Threshold(15),
				wantError: false,
			},
			{
				name:      "zero should error",
				input:     0,
				want:      0,
				wantError: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, gotErr := NewThreshold(tt.input)

				if tt.wantError {
					if gotErr == nil {
						t.Errorf("NewThreshold() expected error, got nil")
						return
					}
					var validationErr *duplerrors.DuplError
					if !stderrors.As(gotErr, &validationErr) {
						t.Errorf("NewThreshold() expected ValidationError, got %T", gotErr)
					}
				} else {
					if gotErr != nil {
						t.Errorf("NewThreshold() unexpected error: %v", gotErr)
						return
					}
					if got != tt.want {
						t.Errorf("NewThreshold() = %v, want %v", got, tt.want)
					}
				}
			})
		}
	})

	t.Run("Uint", func(t *testing.T) {
		th := Threshold(30)
		if got := th.Uint(); got != 30 {
			t.Errorf("Uint() = %v, want %v", got, 30)
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		original := Threshold(50)
		data, err := original.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON() error: %v", err)
		}
		var result Threshold
		if err := result.UnmarshalJSON(data); err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}
		if result != original {
			t.Errorf("Round trip failed: %v != %v", result, original)
		}
	})
}
