package domain

import (
	stderrors "errors"
	"testing"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
)

// TestCloneID_NewCloneID tests the NewCloneID constructor.
func TestCloneID_NewCloneID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    CloneID
		wantError bool
	}{
		{
			name:      "valid clone ID",
			input:     "clone-123",
			wantID:    CloneID("clone-123"),
			wantError: false,
		},
		{
			name:      "empty string should error",
			input:     "",
			wantID:    "",
			wantError: true,
		},
		{
			name:      "ID with special characters",
			input:     "clone-123_abc",
			wantID:    CloneID("clone-123_abc"),
			wantError: false,
		},
		{
			name:      "ID with spaces",
			input:     "clone 123",
			wantID:    CloneID("clone 123"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotErr := NewCloneID(tt.input)

			if tt.wantError {
				if gotErr == nil {
					t.Errorf("NewCloneID() expected error, got nil")
					return
				}
				// Verify error type
				var validationErr *duplerrors.DuplError
				if !stderrors.As(gotErr, &validationErr) {
					t.Errorf("NewCloneID() expected ValidationError, got %T", gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("NewCloneID() unexpected error: %v", gotErr)
					return
				}
				if gotID != tt.wantID {
					t.Errorf("NewCloneID() = %v, want %v", gotID, tt.wantID)
				}
			}
		})
	}
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
