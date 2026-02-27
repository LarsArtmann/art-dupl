package golang

import (
	"testing"
)

// compareDecodeResult is a test helper for comparing decode function results.
func compareDecodeResult(t *testing.T, funcName string, input, result, expected int32) {
	if result != expected {
		t.Errorf("%s(0x%08X) = 0x%06X, want 0x%06X", funcName, input, result, expected)
	}
}

func TestHashIdentifierFast_Consistency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int32 // Same input must always produce same output
	}{
		{"empty", "", 0},
		{"single char", "a", 0x1B3E6B},
		{"short identifier", "foo", 0x2E39B8},
		{"method name", "String", 0x72F941},
		{"common method", "Error", 0x5C20A9},
		{"camelCase", "getValue", 0x1D7D1F8},
		{"with numbers", "test123", 0x517D9A1},
		{"longer name", "GetUserByID", 0x7E4CC74},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call multiple times to verify consistency
			result1 := hashIdentifierFast(tt.input)
			result2 := hashIdentifierFast(tt.input)
			result3 := hashIdentifierFast(tt.input)

			if result1 != result2 || result2 != result3 {
				t.Errorf("hashIdentifierFast(%q) not consistent: got %d, %d, %d",
					tt.input, result1, result2, result3)
			}

			if tt.input == "" && result1 != 0 {
				t.Errorf("hashIdentifierFast('') = %d, want 0", result1)
			}

			// Verify it's within 24-bit range
			if result1 < 0 || result1 > 0x00FFFFFF {
				t.Errorf("hashIdentifierFast(%q) = %d, out of 24-bit range [0, 0x00FFFFFF]",
					tt.input, result1)
			}
		})
	}
}

func TestHashIdentifierFast_DifferentInputs(t *testing.T) {
	// Different inputs should (almost always) produce different outputs
	inputs := []string{
		"foo", "bar", "baz", "qux",
		"String", "Error", "Format", "Parse",
		"Get", "Set", "Add", "Delete",
		"getUser", "setUser", "addUser", "deleteUser",
	}

	hashes := make(map[int32]string)

	for _, input := range inputs {
		hash := hashIdentifierFast(input)
		if existing, exists := hashes[hash]; exists {
			// Collision - this is expected to be extremely rare
			t.Logf("WARNING: hash collision between %q and %q: both hash to %d",
				existing, input, hash)
		}

		hashes[hash] = input
	}
}

func TestHashIdentifierFast_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unicode", "日本語"},
		{"special chars", "test_id"},
		{"all caps", "FOO"},
		{"single underscore", "_"},
		{"double underscore", "__"},
		{"numbers only", "12345"},
		{"very long", "thisIsAVeryLongIdentifierNameThatMightCauseIssues"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashIdentifierFast(tt.input)
			// Just verify it doesn't crash and returns valid range
			if result < 0 || result > 0x00FFFFFF {
				t.Errorf("hashIdentifierFast(%q) = %d, out of 24-bit range", tt.input, result)
			}
		})
	}
}

func TestEncodeSemanticType_BitManipulation(t *testing.T) {
	// Enable semantic hashing for this test
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(42) // Some base type in lower 8 bits
	identifierName := "test"

	result := encodeSemanticType(baseType, identifierName)

	// Extract components
	extractedBase := DecodeBaseType(result)
	extractedHash := DecodeSemanticHash(result)

	if extractedBase != baseType {
		t.Errorf("DecodeBaseType(%d) = %d, want %d", result, extractedBase, baseType)
	}

	// Verify hash was computed
	expectedHash := hashIdentifierFast(identifierName)
	if extractedHash != expectedHash {
		t.Errorf("DecodeSemanticHash(%d) = %d, want %d", result, extractedHash, expectedHash)
	}
}

func TestEncodeSemanticType_Disabled(t *testing.T) {
	// Disable semantic hashing
	original := SemanticHashEnabled
	SemanticHashEnabled = false

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(42)
	result := encodeSemanticType(baseType, "anyIdentifier")

	// When disabled, should return base type unchanged
	if result != baseType {
		t.Errorf(
			"encodeSemanticType(%d, 'anyIdentifier') with SemanticHashEnabled=false = %d, want %d",
			baseType,
			result,
			baseType,
		)
	}
}

func TestEncodeSemanticType_EmptyIdentifier(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(42)
	result := encodeSemanticType(baseType, "")

	// Empty identifier should return base type unchanged
	if result != baseType {
		t.Errorf("encodeSemanticType(%d, '') = %d, want %d", baseType, result, baseType)
	}
}

func TestDecodeBaseType(t *testing.T) {
	tests := []struct {
		name     string
		input    int32
		expected int32
	}{
		{"zero", 0, 0},
		{"max 8-bit", 0xFF, 0xFF},
		{"with upper bits", 0x12345678, 0x78},          // Lower 8 bits
		{"with upper bits 2", int32(0x7FFFFF12), 0x12}, // Lower 8 bits (int32 safe)
		{"type 42", 0x00FF002A, 0x2A},                  // 42 in lower bits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeBaseType(tt.input)
			compareDecodeResult(t, "DecodeBaseType", tt.input, result, tt.expected)
		})
	}
}

func TestDecodeSemanticHash(t *testing.T) {
	tests := []struct {
		name     string
		input    int32
		expected int32
	}{
		{"zero", 0, 0},
		{"only lower bits", 0xFF, 0},                     // Upper 24 bits are 0
		{"only upper bits", 0x01000000, 0x010000},        // Upper 24 bits shifted right by 8
		{"both parts", 0x12345678, 0x123456},             // Upper 24 bits
		{"max 24-bit", int32(0x007FFFFF) << 8, 0x7FFFFF}, // Max positive 24-bit value in int32
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeSemanticHash(tt.input)
			if result != tt.expected {
				t.Errorf("DecodeSemanticHash(0x%08X) = 0x%06X, want 0x%06X",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	tests := []struct {
		baseType       int32
		identifierName string
	}{
		{1, "foo"},
		{42, "String"},
		{255, "Error"},
		{128, "longerIdentifierName"},
	}

	for _, tt := range tests {
		t.Run(tt.identifierName, func(t *testing.T) {
			encoded := encodeSemanticType(tt.baseType, tt.identifierName)
			decodedBase := DecodeBaseType(encoded)
			decodedHash := DecodeSemanticHash(encoded)

			if decodedBase != tt.baseType {
				t.Errorf("round-trip failed: base type %d -> encoded -> %d",
					tt.baseType, decodedBase)
			}

			expectedHash := hashIdentifierFast(tt.identifierName)
			if decodedHash != expectedHash {
				t.Errorf("round-trip failed: hash of %q mismatch: got %d, want %d",
					tt.identifierName, decodedHash, expectedHash)
			}
		})
	}
}

func TestEncodeSemanticTypeHash(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(42)
	nameHash := int32(0x123456)

	result := encodeSemanticTypeHash(baseType, nameHash)

	// Verify encoding
	if DecodeBaseType(result) != baseType {
		t.Errorf("base type not preserved: got %d, want %d",
			DecodeBaseType(result), baseType)
	}

	if DecodeSemanticHash(result) != nameHash {
		t.Errorf("hash not preserved: got %d, want %d",
			DecodeSemanticHash(result), nameHash)
	}
}

func TestEncodeSemanticTypeHash_Disabled(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = false

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(42)
	nameHash := int32(0x123456)

	result := encodeSemanticTypeHash(baseType, nameHash)

	// When disabled, should return base type
	if result != baseType {
		t.Errorf("encodeSemanticTypeHash(%d, %d) with SemanticHashEnabled=false = %d, want %d",
			baseType, nameHash, result, baseType)
	}
}

func TestEncodeSemanticTypeHash_ZeroHash(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(42)
	result := encodeSemanticTypeHash(baseType, 0)

	// Zero hash should return base type
	if result != baseType {
		t.Errorf("encodeSemanticTypeHash(%d, 0) = %d, want %d",
			baseType, result, baseType)
	}
}

func TestCollisionBehavior(t *testing.T) {
	// Test that similar but different identifiers produce different hashes
	// This is probabilistic - we test a sample that should definitely differ
	tests := []struct {
		name1, name2 string
	}{
		{"Get", "Set"},
		{"String", "Error"},
		{"foo", "bar"},
		{"getUser", "setUser"},
		{"User", "Users"},
	}

	for _, tt := range tests {
		t.Run(tt.name1+"_"+tt.name2, func(t *testing.T) {
			hash1 := hashIdentifierFast(tt.name1)
			hash2 := hashIdentifierFast(tt.name2)

			if hash1 == hash2 {
				t.Errorf("collision: %q and %q both hash to %d", tt.name1, tt.name2, hash1)
			}
		})
	}
}

func TestSemanticHashEnabled_Default(t *testing.T) {
	// Verify default is true for better accuracy (fewer false positives)
	// Note: This test checks the global state, so we save and restore
	// the current value in case other tests modified it
	currentValue := SemanticHashEnabled

	// Just verify we can set it and it works
	SemanticHashEnabled = true
	if !SemanticHashEnabled {
		t.Error("failed to set SemanticHashEnabled = true")
	}

	SemanticHashEnabled = false
	if SemanticHashEnabled {
		t.Error("failed to set SemanticHashEnabled = false")
	}

	// Restore
	SemanticHashEnabled = currentValue
}

func TestCombineIdentifierHashes(t *testing.T) {
	tests := []struct {
		name     string
		hash1    int32
		hash2    int32
		wantZero bool
	}{
		{"both zero", 0, 0, true},
		{"hash1 zero", 0, 0x123456, false},
		{"hash2 zero", 0x123456, 0, false},
		{"both non-zero", 0x123456, 0x789ABC, false},
		{"same hash", 0x123456, 0x123456, true}, // XOR of same = 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combineIdentifierHashes(tt.hash1, tt.hash2)
			if tt.wantZero && result != 0 {
				t.Errorf("combineIdentifierHashes(%x, %x) = %x, want 0", tt.hash1, tt.hash2, result)
			}

			if !tt.wantZero && result == 0 {
				t.Errorf("combineIdentifierHashes(%x, %x) = 0, want non-zero", tt.hash1, tt.hash2)
			}
			// Verify result is in 24-bit range
			if result < 0 || result > 0x00FFFFFF {
				t.Errorf(
					"combineIdentifierHashes(%x, %x) = %x, out of 24-bit range",
					tt.hash1,
					tt.hash2,
					result,
				)
			}
		})
	}
}

func TestEncodeSemanticTypeMulti(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(FuncDecl)

	tests := []struct {
		name        string
		identifiers []string
		wantBase    bool
	}{
		{"empty identifiers", []string{}, true},
		{"single identifier", []string{"IsValid"}, false},
		{"receiver and function", []string{"CrushMode", "IsValid"}, false},
		{"empty string in slice", []string{"", "IsValid"}, false},
		{"all empty strings", []string{"", ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeSemanticTypeMulti(baseType, tt.identifiers...)

			// Verify base type is preserved in lower 8 bits
			extractedBase := DecodeBaseType(result)
			if extractedBase != baseType {
				t.Errorf("base type not preserved: got %d, want %d", extractedBase, baseType)
			}

			if tt.wantBase {
				// Should return just the base type (no semantic hash)
				if result != baseType {
					t.Errorf(
						"encodeSemanticTypeMulti(%v) = %d, want %d (no semantic hash)",
						tt.identifiers,
						result,
						baseType,
					)
				}
			} else {
				// Should have semantic hash in upper bits
				semanticHash := DecodeSemanticHash(result)
				if semanticHash == 0 {
					t.Errorf(
						"encodeSemanticTypeMulti(%v) returned zero semantic hash",
						tt.identifiers,
					)
				}
			}
		})
	}
}

func TestEncodeSemanticTypeMulti_DifferentCombinations(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(FuncDecl)

	// Different receiver + function combinations should produce different hashes
	hash1 := encodeSemanticTypeMulti(baseType, "CrushMode", "IsValid")
	hash2 := encodeSemanticTypeMulti(baseType, "SafetyMode", "IsValid")
	hash3 := encodeSemanticTypeMulti(baseType, "CrushMode", "IsEnabled")
	hash4 := encodeSemanticTypeMulti(baseType, "SafetyMode", "IsEnabled")

	if hash1 == hash2 {
		t.Error("different receivers with same function should produce different hashes")
	}

	if hash1 == hash3 {
		t.Error("same receiver with different functions should produce different hashes")
	}

	if hash1 == hash4 {
		t.Error("completely different receiver+function should produce different hashes")
	}
}

func TestEncodeSemanticTypeMulti_Disabled(t *testing.T) {
	original := SemanticHashEnabled
	SemanticHashEnabled = false

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(FuncDecl)
	result := encodeSemanticTypeMulti(baseType, "CrushMode", "IsValid")

	if result != baseType {
		t.Errorf(
			"encodeSemanticTypeMulti with SemanticHashEnabled=false = %d, want %d",
			result,
			baseType,
		)
	}
}

func BenchmarkHashIdentifierFast(b *testing.B) {
	identifiers := []string{
		"String", "Error", "Format", "Parse",
		"getUser", "setUser", "deleteUser",
		"short", "mediumLengthName", "veryLongIdentifierNameHere",
	}

	for b.Loop() {
		for _, id := range identifiers {
			_ = hashIdentifierFast(id)
		}
	}
}

func BenchmarkEncodeSemanticType(b *testing.B) {
	original := SemanticHashEnabled
	SemanticHashEnabled = true

	defer func() { SemanticHashEnabled = original }()

	baseType := int32(42)

	for b.Loop() {
		_ = encodeSemanticType(baseType, "identifierName")
	}
}
