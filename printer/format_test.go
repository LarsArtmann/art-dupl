package printer

import (
	"errors"
	"strings"
	"testing"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

func TestFormatIsValid(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		expected bool
	}{
		{"text is valid", FormatText, true},
		{"json is valid", FormatJSON, true},
		{"csv is valid", FormatCSV, true},
		{"empty is invalid", Format(""), false},
		{"invalid format", Format("xml"), false},
		{"wrong case", Format("TEXT"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.AssertEqual(t, tt.format.IsValid(), tt.expected, "Format.IsValid()")
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Format
		wantError bool
	}{
		{"parse text", "text", FormatText, false},
		{"parse json", "json", FormatJSON, false},
		{"parse csv", "csv", FormatCSV, false},
		{"invalid format", "xml", "", true},
		{"empty string", "", "", true},
		{"wrong case", "TEXT", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)

			expectErr := tt.wantError
			if (err != nil) != expectErr {
				t.Errorf("ParseFormat() error = %v, wantError %v", err, expectErr)

				return
			}

			if got != tt.expected {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseFormatErrorMessage(t *testing.T) {
	_, err := ParseFormat("invalid")
	if err == nil {
		t.Fatal("ParseFormat() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("ParseFormat() error = %q, want error containing 'invalid'", err.Error())
	}

	var enumErr *duplerrors.EnumValidationError
	if !errors.As(err, &enumErr) {
		t.Errorf("ParseFormat() error = %q, want EnumValidationError", err.Error())
	}
}

func TestFormatString(t *testing.T) {
	if got := FormatText.String(); got != "text" {
		t.Errorf("FormatText.String() = %q, want %q", got, "text")
	}

	if got := FormatJSON.String(); got != "json" {
		t.Errorf("FormatJSON.String() = %q, want %q", got, "json")
	}

	if got := FormatCSV.String(); got != "csv" {
		t.Errorf("FormatCSV.String() = %q, want %q", got, "csv")
	}
}
