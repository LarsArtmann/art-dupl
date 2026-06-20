package enum

import (
	"errors"
	"testing"
)

type color string

const (
	red   color = "red"
	green color = "green"
	blue  color = "blue"
)

var errInvalidColor = errors.New("invalid color")

func (c color) isValid() bool {
	switch c {
	case red, green, blue:
		return true
	default:
		return false
	}
}

func TestMarshalJSON_Valid(t *testing.T) {
	data, err := MarshalJSON(red, color.isValid, errInvalidColor)
	if err != nil {
		t.Fatalf("MarshalJSON valid color: unexpected error: %v", err)
	}

	if string(data) != `"red"` {
		t.Errorf("MarshalJSON valid color: got %s, want \"red\"", data)
	}
}

func TestMarshalJSON_Invalid(t *testing.T) {
	_, err := MarshalJSON(color("purple"), func(c color) bool { return false }, errInvalidColor)
	if err == nil {
		t.Fatal("MarshalJSON invalid color: expected error, got nil")
	}

	if !errors.Is(err, errInvalidColor) {
		t.Errorf("MarshalJSON invalid color: error should wrap errInvalidColor, got: %v", err)
	}
}

func TestUnmarshalJSON_Valid(t *testing.T) {
	parsed, err := UnmarshalJSON([]byte(`"green"`), color.isValid, errInvalidColor)
	if err != nil {
		t.Fatalf("UnmarshalJSON valid: unexpected error: %v", err)
	}

	if parsed != green {
		t.Errorf("UnmarshalJSON valid: got %q, want %q", parsed, green)
	}
}

func TestUnmarshalJSON_Invalid(t *testing.T) {
	_, err := UnmarshalJSON([]byte(`"purple"`), color.isValid, errInvalidColor)
	if err == nil {
		t.Fatal("UnmarshalJSON invalid: expected error, got nil")
	}

	if !errors.Is(err, errInvalidColor) {
		t.Errorf("UnmarshalJSON invalid: error should wrap errInvalidColor, got: %v", err)
	}
}

func TestUnmarshalJSON_MalformedJSON(t *testing.T) {
	_, err := UnmarshalJSON([]byte(`not json`), color.isValid, errInvalidColor)
	if err == nil {
		t.Fatal("UnmarshalJSON malformed: expected error, got nil")
	}

	if !errors.Is(err, errInvalidColor) {
		t.Errorf("UnmarshalJSON malformed: error should wrap errInvalidColor, got: %v", err)
	}
}

func TestParse_Valid(t *testing.T) {
	parsed, err := Parse("blue", color.isValid, errInvalidColor)
	if err != nil {
		t.Fatalf("Parse valid: unexpected error: %v", err)
	}

	if parsed != blue {
		t.Errorf("Parse valid: got %q, want %q", parsed, blue)
	}
}

func TestParse_Invalid(t *testing.T) {
	_, err := Parse("purple", color.isValid, errInvalidColor)
	if err == nil {
		t.Fatal("Parse invalid: expected error, got nil")
	}

	if !errors.Is(err, errInvalidColor) {
		t.Errorf("Parse invalid: error should wrap errInvalidColor, got: %v", err)
	}
}
