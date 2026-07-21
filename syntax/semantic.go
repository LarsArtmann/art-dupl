package syntax

// FNV-1a constants for identifier hashing.
const (
	fnvOffset32 uint32 = 2166136261
	fnvPrime32  uint32 = 16777619
)

// HashIdentifier computes a 24-bit FNV-1a hash of an identifier name.
// The result fits in the upper 24 bits of int32 when shifted, allowing
// the lower 8 bits to store the base node type.
func HashIdentifier(name string) int32 {
	if name == "" {
		return 0
	}

	hash := fnvOffset32

	for _, b := range []byte(name) {
		hash ^= uint32(b)
		hash *= fnvPrime32
	}

	return int32(hash & 0x00FFFFFF)
}

// EncodeSemanticType combines a base node type with an identifier hash.
// Layout: [24 bits identifier hash][8 bits base type].
// When semanticEnabled is false, returns baseType unchanged.
func EncodeSemanticType(baseType int32, identifierName string, semanticEnabled bool) int32 {
	if !semanticEnabled || identifierName == "" {
		return baseType
	}

	return (HashIdentifier(identifierName) << 8) | (baseType & 0xFF)
}

// DecodeBaseType extracts the base AST node type from a semantic-encoded type.
func DecodeBaseType(t int32) int32 {
	return t & 0xFF
}
