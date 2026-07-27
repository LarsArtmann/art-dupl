package domain

import (
	"encoding/json/v2"
	"errors"
	"testing"
)

func TestClonePriority_Rank(t *testing.T) {
	cases := []struct {
		priority ClonePriority
		want     int
	}{
		{PriorityCritical, 4},
		{PriorityHigh, 3},
		{PriorityMedium, 2},
		{PriorityLow, 1},
		{ClonePriority("invalid"), 0},
	}

	for _, tc := range cases {
		got := tc.priority.Rank()
		if got != tc.want {
			t.Errorf("%q.Rank() = %d, want %d", tc.priority, got, tc.want)
		}
	}
}

func TestClonePriority_Rank_Ordering(t *testing.T) {
	// Higher-priority clones must rank above lower-priority ones.
	if PriorityCritical.Rank() <= PriorityHigh.Rank() {
		t.Error("critical must outrank high")
	}

	if PriorityHigh.Rank() <= PriorityLow.Rank() {
		t.Error("high must outrank low")
	}
}

func TestCloneCategory_IsValid(t *testing.T) {
	valid := []CloneCategory{
		CategoryFunction, CategoryMethod, CategoryTest, CategoryStruct,
		CategoryInterface, CategoryHandler, CategoryLoop, CategoryConditional,
		CategoryTestBoilerplate, CategoryTestFixture, CategoryAssignment,
		CategoryExpression, CategoryUnknown,
	}

	for _, c := range valid {
		if !c.IsValid() {
			t.Errorf("%q should be valid", c)
		}
	}

	if CloneCategory("bogus").IsValid() {
		t.Error("'bogus' should be invalid")
	}
}

func TestCloneCategory_String(t *testing.T) {
	if CategoryFunction.String() != "function" {
		t.Errorf("CategoryFunction.String() = %q, want \"function\"", CategoryFunction.String())
	}

	if CategoryMethod.String() != string(CategoryMethod) {
		t.Error("String() should match underlying string value")
	}
}

func TestCloneCategory_MarshalJSON(t *testing.T) {
	data, err := json.Marshal(CategoryLoop)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != `"loop"` {
		t.Errorf("MarshalJSON() = %s, want \"loop\"", data)
	}
}

func TestCloneCategory_MarshalJSON_Invalid(t *testing.T) {
	_, err := json.Marshal(CloneCategory("nope"))
	if err == nil {
		t.Error("MarshalJSON('nope') expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCloneCategory) {
		t.Errorf("expected ErrInvalidCloneCategory, got %v", err)
	}
}

func TestCloneCategory_UnmarshalJSON(t *testing.T) {
	var c CloneCategory

	err := json.Unmarshal([]byte(`"interface"`), &c)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if c != CategoryInterface {
		t.Errorf("UnmarshalJSON() = %q, want interface", c)
	}
}

func TestCloneCategory_UnmarshalJSON_Invalid(t *testing.T) {
	var c CloneCategory

	err := json.Unmarshal([]byte(`"bogus"`), &c)
	if err == nil {
		t.Error("UnmarshalJSON('bogus') expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCloneCategory) {
		t.Errorf("expected ErrInvalidCloneCategory, got %v", err)
	}
}

func TestCloneActionability_IsValid(t *testing.T) {
	if !Actionable.IsValid() {
		t.Error("Actionable should be valid")
	}

	if !NonActionable.IsValid() {
		t.Error("NonActionable should be valid")
	}

	if !LowConfidence.IsValid() {
		t.Error("LowConfidence should be valid")
	}

	if CloneActionability("maybe").IsValid() {
		t.Error("'maybe' should be invalid")
	}
}

func TestCloneActionability_String(t *testing.T) {
	if Actionable.String() != "actionable" {
		t.Errorf("Actionable.String() = %q", Actionable.String())
	}

	if NonActionable.String() != "non-actionable" {
		t.Errorf("NonActionable.String() = %q", NonActionable.String())
	}

	if LowConfidence.String() != "low-confidence" {
		t.Errorf("LowConfidence.String() = %q", LowConfidence.String())
	}
}

func TestCloneActionability_MarshalJSON(t *testing.T) {
	data, err := json.Marshal(Actionable)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != `"actionable"` {
		t.Errorf("MarshalJSON() = %s, want \"actionable\"", data)
	}
}

func TestCloneActionability_MarshalJSON_Invalid(t *testing.T) {
	_, err := json.Marshal(CloneActionability("perhaps"))
	if err == nil {
		t.Error("MarshalJSON('perhaps') expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCloneActionability) {
		t.Errorf("expected ErrInvalidCloneActionability, got %v", err)
	}
}

func TestCloneActionability_UnmarshalJSON(t *testing.T) {
	var a CloneActionability

	err := json.Unmarshal([]byte(`"non-actionable"`), &a)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if a != NonActionable {
		t.Errorf("UnmarshalJSON() = %q, want non-actionable", a)
	}
}

func TestCloneActionability_UnmarshalJSON_Invalid(t *testing.T) {
	var a CloneActionability

	err := json.Unmarshal([]byte(`"sometimes"`), &a)
	if err == nil {
		t.Error("UnmarshalJSON('sometimes') expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCloneActionability) {
		t.Errorf("expected ErrInvalidCloneActionability, got %v", err)
	}
}

func TestCloneType_IsValid(t *testing.T) {
	for _, ct := range []CloneType{CloneType1, CloneType2, CloneType3} {
		if !ct.IsValid() {
			t.Errorf("%q should be valid", ct)
		}
	}

	if CloneType("type-9").IsValid() {
		t.Error("'type-9' should be invalid")
	}
}

func TestCloneType_String(t *testing.T) {
	cases := []struct {
		ct   CloneType
		want string
	}{
		{CloneType1, "type-1"},
		{CloneType2, "type-2"},
		{CloneType3, "type-3"},
	}

	for _, tc := range cases {
		if tc.ct.String() != tc.want {
			t.Errorf("%q.String() = %q, want %q", tc.ct, tc.ct.String(), tc.want)
		}
	}
}

func TestCloneType_MarshalJSON(t *testing.T) {
	data, err := json.Marshal(CloneType2)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != `"type-2"` {
		t.Errorf("MarshalJSON() = %s, want \"type-2\"", data)
	}
}

func TestCloneType_MarshalJSON_Invalid(t *testing.T) {
	_, err := json.Marshal(CloneType("type-99"))
	if err == nil {
		t.Error("MarshalJSON('type-99') expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCloneType) {
		t.Errorf("expected ErrInvalidCloneType, got %v", err)
	}
}

func TestCloneType_UnmarshalJSON(t *testing.T) {
	var ct CloneType

	err := json.Unmarshal([]byte(`"type-3"`), &ct)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if ct != CloneType3 {
		t.Errorf("UnmarshalJSON() = %q, want type-3", ct)
	}
}

func TestCloneType_UnmarshalJSON_Invalid(t *testing.T) {
	var ct CloneType

	err := json.Unmarshal([]byte(`"type-0"`), &ct)
	if err == nil {
		t.Error("UnmarshalJSON('type-0') expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCloneType) {
		t.Errorf("expected ErrInvalidCloneType, got %v", err)
	}
}

// TestCloneType_RoundTrip verifies marshal → unmarshal preserves the value,
// which is the property CI integrations rely on for baseline serialization.
func TestCloneType_RoundTrip(t *testing.T) {
	for _, original := range []CloneType{CloneType1, CloneType2, CloneType3} {
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal %q: %v", original, err)
		}

		var restored CloneType
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("unmarshal %q: %v", original, err)
		}

		if restored != original {
			t.Errorf("round-trip mismatch: %q → %q", original, restored)
		}
	}
}
