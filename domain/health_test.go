package domain

import (
	"encoding/json/v2"
	"errors"
	"testing"
)

func TestHealthScore_IsValid(t *testing.T) {
	valid := []HealthScore{HealthScoreA, HealthScoreB, HealthScoreC, HealthScoreD, HealthScoreF}
	for _, score := range valid {
		if !score.IsValid() {
			t.Errorf("HealthScore %q should be valid", score)
		}
	}

	if HealthScore("Z").IsValid() {
		t.Error("HealthScore 'Z' should be invalid")
	}
}

func TestHealthScore_String(t *testing.T) {
	if HealthScoreA.String() != "A" {
		t.Errorf("HealthScoreA.String() = %q, want 'A'", HealthScoreA.String())
	}
}

func TestHealthScore_MarshalJSON(t *testing.T) {
	data, err := HealthScoreA.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: unexpected error: %v", err)
	}

	if string(data) != `"A"` {
		t.Errorf("MarshalJSON: got %s, want \"A\"", data)
	}
}

func TestHealthScore_MarshalJSON_Invalid(t *testing.T) {
	_, err := HealthScore("Z").MarshalJSON()
	if err == nil {
		t.Fatal("MarshalJSON invalid: expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidHealthScore) {
		t.Errorf("MarshalJSON invalid: error should wrap ErrInvalidHealthScore, got: %v", err)
	}
}

func TestHealthScore_UnmarshalJSON(t *testing.T) {
	var h HealthScore

	err := json.Unmarshal([]byte(`"B"`), &h)
	if err != nil {
		t.Fatalf("UnmarshalJSON: unexpected error: %v", err)
	}

	if h != HealthScoreB {
		t.Errorf("UnmarshalJSON: got %q, want %q", h, HealthScoreB)
	}
}

func TestHealthScore_UnmarshalJSON_Invalid(t *testing.T) {
	var h HealthScore

	err := json.Unmarshal([]byte(`"Z"`), &h)
	if err == nil {
		t.Fatal("UnmarshalJSON invalid: expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidHealthScore) {
		t.Errorf("UnmarshalJSON invalid: error should wrap ErrInvalidHealthScore, got: %v", err)
	}
}
