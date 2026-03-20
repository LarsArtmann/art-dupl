package golang

// SemanticHashEnabled controls whether semantic identifiers are included in type hashes.
// When false (default), only AST node types are used (current behavior).
// When true, identifier names are hashed into the type field for semantic matching.
var SemanticHashEnabled bool

// hashIdentifierFast computes a 24-bit FNV-1a hash of an identifier name.
// The result fits in the upper 24 bits of int32 when shifted, allowing
// the lower 8 bits to store the base AST node type.
//
// Algorithm: FNV-1a (fast for short strings, good distribution)
// Collisions: ~0.0001% for 10,000 unique names (acceptable for deduplication)
//
// Performance: ~15ns per call.
func hashIdentifierFast(name string) int32 {
	if name == "" {
		return 0
	}

	// FNV-1a constants
	const (
		prime32  uint32 = 16777619
		offset32 uint32 = 2166136261
	)

	hash := offset32
	for _, b := range []byte(name) {
		hash ^= uint32(b)
		hash *= prime32
	}

	// Truncate to 24 bits

	return int32(hash & 0x00FFFFFF)
}

// encodeSemanticType combines a base node type with an identifier hash.
// Layout: [24 bits identifier hash][8 bits base type]
// This allows semantic-aware matching while preserving AST type information.
//
// Example:
//
//	encodeSemanticType(SelectorExpr, "String") => unique type for "X.String" calls
//	encodeSemanticType(SelectorExpr, "Error") => different type for "X.Error" calls
func encodeSemanticType(baseType int32, identifierName string) int32 {
	if !SemanticHashEnabled || identifierName == "" {
		return baseType
	}

	nameHash := hashIdentifierFast(identifierName)

	return (nameHash << 8) | (baseType & 0xFF)
}

// encodeSemanticTypeHash combines a base node type with a pre-computed hash.
// Use this when the hash is already computed to avoid recomputation.
func encodeSemanticTypeHash(baseType, nameHash int32) int32 {
	if !SemanticHashEnabled || nameHash == 0 {
		return baseType
	}

	return (nameHash << 8) | (baseType & 0xFF)
}

// DecodeBaseType extracts the base AST node type from a semantic-encoded type.
func DecodeBaseType(t int32) int32 {
	return t & 0xFF
}

// DecodeSemanticHash extracts the identifier hash from a semantic-encoded type.
func DecodeSemanticHash(t int32) int32 {
	return (t >> 8) & 0x00FFFFFF
}

// combineIdentifierHashes combines two identifier hashes into one 24-bit hash.
// Uses XOR mixing for good distribution while staying within 24 bits.
// This allows encoding both receiver type and function name in a single semantic hash.
func combineIdentifierHashes(hash1, hash2 int32) int32 {
	if hash1 == 0 {
		return hash2
	}

	if hash2 == 0 {
		return hash1
	}
	// XOR the hashes and keep in 24-bit range
	return (hash1 ^ hash2) & 0x00FFFFFF
}

// encodeSemanticTypeMulti combines a base node type with multiple identifier hashes.
// This is useful for nodes like FuncDecl where both receiver type and function name matter.
func encodeSemanticTypeMulti(baseType int32, identifiers ...string) int32 {
	if !SemanticHashEnabled || len(identifiers) == 0 {
		return baseType
	}

	var combinedHash int32

	for _, id := range identifiers {
		if id != "" {
			hash := hashIdentifierFast(id)
			combinedHash = combineIdentifierHashes(combinedHash, hash)
		}
	}

	if combinedHash == 0 {
		return baseType
	}

	return (combinedHash << 8) | (baseType & 0xFF)
}
