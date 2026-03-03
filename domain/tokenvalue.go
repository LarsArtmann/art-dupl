// Package domain provides core domain types for clone detection.
package domain

import (
	"fmt"
	"math"

	"github.com/LarsArtmann/art-dupl/errors"
)

// TokenValue represents a unique token identifier in the suffix tree.
//
// This type provides type safety for token values, preventing accidental
// mixing with other int types (like positions or counts). It uses int32
// as the underlying type to match Pos and minimize memory usage.
//
// Valid TokenValue ranges:
//   - Minimum: math.MinInt32 (for special/error tokens)
//   - Maximum: math.MaxInt32 (for regular tokens)
//
// Usage:
//
//	value := domain.NewTokenValue(42)
//	if err := value.IsValid(); err != nil {
//	    return err
//	}
//	key := value.Int32() // Use as map key
//
// Type Safety Benefits:
//   - Cannot accidentally use a position (Pos) as a token value
//   - Cannot accidentally use a count (TokenCount) as a token value
//   - Compiler catches type mismatches at build time
//   - Self-documenting code - intent is clear
//
// Performance:
//   - Same memory usage as int32 (4 bytes)
//   - No allocation overhead
//   - Direct map key usage (no conversion needed)
type TokenValue int32

// MinTokenValue is the minimum valid token value.
const MinTokenValue TokenValue = TokenValue(math.MinInt32)

// MaxTokenValue is the maximum valid token value.
const MaxTokenValue TokenValue = TokenValue(math.MaxInt32)

// NewTokenValue creates a TokenValue from an int32.
//
// This is a simple constructor that performs no validation.
// Use IsValid() to check if the value is within acceptable bounds.
//
// Example:
//
//	value := domain.NewTokenValue(42)
//	if err := value.IsValid(); err != nil {
//	    return err
//	}
func NewTokenValue(v int32) TokenValue {
	return TokenValue(v)
}

// NewTokenValueFromInt creates a TokenValue from an int with validation.
//
// Returns an error if the int value cannot be represented as an int32.
// This is useful when converting from int (which may be 64-bit on some platforms).
//
// Example:
//
//	value, err := domain.NewTokenValueFromInt(someInt)
//	if err != nil {
//	    return err
//	}
func NewTokenValueFromInt(v int) (TokenValue, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, errors.NewValidationError(
			fmt.Sprintf("token value %d out of int32 range [%d, %d]", v, math.MinInt32, math.MaxInt32),
			nil,
		)
	}

	return TokenValue(v), nil
}

// Int32 returns the underlying int32 value.
//
// Use this when you need the primitive value, such as for:
//   - Map keys
//   - Array indices
//   - Protocol buffers
//   - Database storage
func (tv TokenValue) Int32() int32 {
	return int32(tv)
}

// Int returns the value as an int.
//
// This is a convenience method for interoperability with APIs that expect int.
// Note: On 32-bit platforms, this is a direct conversion. On 64-bit platforms,
// the value is sign-extended.
func (tv TokenValue) Int() int {
	return int(tv)
}

// IsValid checks if the TokenValue is within valid bounds.
//
// Currently, all int32 values are considered valid, but this method
// provides a hook for future validation (e.g., reserved value ranges).
//
// Returns nil if valid, error if invalid.
func (tv TokenValue) IsValid() error {
	// All int32 values are currently valid
	// Future: could reserve specific ranges for special tokens
	return nil
}

// IsZero returns true if the TokenValue is zero.
//
// Zero is often used as a default or "unset" value.
func (tv TokenValue) IsZero() bool {
	return tv == 0
}

// IsNegative returns true if the TokenValue is negative.
//
// Negative values may be used for special/error tokens.
func (tv TokenValue) IsNegative() bool {
	return tv < 0
}

// IsPositive returns true if the TokenValue is positive.
//
// Positive values are typically used for regular tokens.
func (tv TokenValue) IsPositive() bool {
	return tv > 0
}

// String returns a string representation of the TokenValue.
//
// This implements the fmt.Stringer interface for debugging and logging.
func (tv TokenValue) String() string {
	return fmt.Sprintf("TokenValue(%d)", tv)
}

// MarshalJSON implements json.Marshaler for TokenValue.
//
// Serializes as a JSON number for compact representation.
func (tv TokenValue) MarshalJSON() ([]byte, error) {
	return marshalInt32(int32(tv))
}

// UnmarshalJSON implements json.Unmarshaler for TokenValue.
//
// Deserializes from a JSON number with validation.
func (tv *TokenValue) UnmarshalJSON(data []byte) error {
	return unmarshalInt32(data, "TokenValue", func(n int32) {
		*tv = TokenValue(n)
	})
}

// Equal compares two TokenValues for equality.
//
// This is provided for consistency with other domain types, though
// direct comparison (==) works as well.
func (tv TokenValue) Equal(other TokenValue) bool {
	return tv == other
}

// Less compares two TokenValues (for sorting).
//
// Returns true if tv < other.
func (tv TokenValue) Less(other TokenValue) bool {
	return tv < other
}

// Max returns the maximum of two TokenValues.
func (tv TokenValue) Max(other TokenValue) TokenValue {
	if tv > other {
		return tv
	}
	return other
}

// Min returns the minimum of two TokenValues.
func (tv TokenValue) Min(other TokenValue) TokenValue {
	if tv < other {
		return tv
	}
	return other
}

// Abs returns the absolute value of a TokenValue.
//
// Note: math.MinInt32 will return math.MinInt32 (negative) due to overflow.
func (tv TokenValue) Abs() TokenValue {
	if tv < 0 {
		return -tv
	}
	return tv
}

// Add returns a new TokenValue with the sum of tv and other.
//
// Note: This does not check for overflow. Use AddChecked for safe addition.
func (tv TokenValue) Add(other TokenValue) TokenValue {
	return tv + other
}

// AddChecked returns the sum of tv and other with overflow checking.
//
// Returns an error if the result would overflow int32.
func (tv TokenValue) AddChecked(other TokenValue) (TokenValue, error) {
	result := int64(tv) + int64(other)
	if result < math.MinInt32 || result > math.MaxInt32 {
		return 0, errors.NewValidationError(
			fmt.Sprintf("token value addition overflow: %d + %d", tv, other),
			nil,
		)
	}
	return TokenValue(result), nil
}

// TokenValueFromMapKey retrieves a TokenValue from a map.
//
// This is a helper for the common pattern of looking up transitions
// in the suffix tree. Returns the value and true if found.
//
// Example:
//
//	transitions := map[TokenValue]*tran{...}
//	if tr, ok := domain.TokenValueFromMapKey(transitions, key); ok {
//	    // use tr
//	}
func TokenValueFromMapKey[V any](m map[TokenValue]V, key TokenValue) (V, bool) {
	v, ok := m[key]
	return v, ok
}
