// Package valueobjects provides domain-specific value types for type safety.
//
// This package defines strongly-typed value objects that replace generic types
// (uint, string, float64) with domain-specific types. This provides:
// - Compile-time type safety (can't accidentally use wrong values)
// - Self-documenting code (intent is explicit)
// - Encapsulation of validation logic
// - Better IDE support and autocomplete
// - Prevention of type errors
//
// All types in this package are value objects: immutable and validated.
//
// Usage Examples:
//
//	// Instead of:
//	id := "clone-123"
//	line := uint(10)
//
//	// Use:
//	id := valueobjects.CloneID("clone-123")
//	line := valueobjects.LineNumber(10)
//
// Benefits:
// - Can't accidentally pass CloneID where Filepath is expected
// - Validation is enforced at construction time
// - Code is self-documenting
// - Refactoring is safer (find all usages easily)

package domain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/LarsArtmann/art-dupl/errors"
)

// marshalStringID is a helper function for marshaling string-based ID types.
// It handles the common pattern of validating non-empty strings and marshaling to JSON.
func marshalStringID(s, typeName, validationMsg string) ([]byte, error) {
	if s == "" {
		return nil, errors.NewValidationError(validationMsg, nil)
	}
	return json.Marshal(s)
}

// unmarshalWithValidation is a generic helper for unmarshaling JSON with custom validation.
// It unmarshals data to type T, validates it using the provided validator function,
// and assigns the result if validation passes.
func unmarshalWithValidation[T any](data []byte, typeName, validationMsg string, validator func(T) bool, assign func(T)) error {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", typeName, err)
	}
	if !validator(value) {
		return errors.NewValidationError(validationMsg, nil)
	}
	assign(value)
	return nil
}

// unmarshalStringID is a helper for unmarshaling string IDs that must not be empty.
func unmarshalStringID(data []byte, typeName, validationMsg string, assign func(string)) error {
	return unmarshalWithValidation(data, typeName, validationMsg, func(s string) bool { return s != "" }, assign)
}

// unmarshalUintNonZero is a helper for unmarshaling uint-based types that must not be zero.
func unmarshalUintNonZero(data []byte, typeName, validationMsg string, assign func(uint)) error {
	return unmarshalWithValidation(data, typeName, validationMsg, func(n uint) bool { return n != 0 }, assign)
}

// unmarshalUint is a helper function for unmarshaling uint-based types
// that allow zero values. It handles the common pattern of unmarshaling JSON to uint.
func unmarshalUint(data []byte, typeName string, assign func(uint)) error {
	return unmarshalWithValidation(data, typeName, "", func(uint) bool { return true }, assign)
}

// marshalUint is a helper function for marshaling uint-based types.
// It handles the common pattern of marshaling uint-wrapped types to JSON.
func marshalUint(n uint) ([]byte, error) {
	return json.Marshal(n)
}

// CloneGroupID represents a unique identifier for a clone group.
type CloneGroupID string

// NewCloneGroupID creates a validated CloneGroupID from a string.
func NewCloneGroupID(id string) (CloneGroupID, error) {
	if id == "" {
		return "", errors.NewValidationError("clone group ID cannot be empty", nil)
	}
	return CloneGroupID(id), nil
}

// String returns the string representation of CloneGroupID.
func (id CloneGroupID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for CloneGroupID.
func (id CloneGroupID) MarshalJSON() ([]byte, error) {
	return marshalStringID(string(id), "CloneGroupID", "clone group ID cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for CloneGroupID.
func (id *CloneGroupID) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "CloneGroupID", "clone group ID cannot be empty", func(s string) {
		*id = CloneGroupID(s)
	})
}

// AnalysisID represents a unique identifier for an analysis.
type AnalysisID string

// NewAnalysisID creates a validated AnalysisID from a string.
func NewAnalysisID(id string) (AnalysisID, error) {
	if id == "" {
		return "", errors.NewValidationError("analysis ID cannot be empty", nil)
	}
	return AnalysisID(id), nil
}

// String returns the string representation of AnalysisID.
func (id AnalysisID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for AnalysisID.
func (id AnalysisID) MarshalJSON() ([]byte, error) {
	return marshalStringID(string(id), "AnalysisID", "analysis ID cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for AnalysisID.
func (id *AnalysisID) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "AnalysisID", "analysis ID cannot be empty", func(s string) {
		*id = AnalysisID(s)
	})
}

// Filepath represents a filesystem path.
type Filepath string

// NewFilepath creates a validated Filepath from a string.
func NewFilepath(path string) (Filepath, error) {
	if path == "" {
		return "", errors.NewValidationError("filepath cannot be empty", nil)
	}
	return Filepath(path), nil
}

// String returns the string representation of Filepath.
func (fp Filepath) String() string {
	return string(fp)
}

// MarshalJSON implements json.Marshaler for Filepath.
func (fp Filepath) MarshalJSON() ([]byte, error) {
	return marshalStringID(string(fp), "Filepath", "filepath cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for Filepath.
func (fp *Filepath) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "Filepath", "filepath cannot be empty", func(s string) {
		*fp = Filepath(s)
	})
}

// LineNumber represents a line number in a source file.
// Line numbers start at 1 (not 0) in most editors.
// Optimized: uint16 provides 0-65,535 range (sufficient for any source file).
type LineNumber uint16

// NewLineNumber creates a validated LineNumber from a uint16.
// Returns error if the line number is 0 (invalid).
func NewLineNumber(n uint16) (LineNumber, error) {
	if n == 0 {
		return 0, errors.NewValidationError("line number cannot be 0", nil)
	}
	return LineNumber(n), nil
}

// Uint16 returns the underlying uint16 value.
func (ln LineNumber) Uint16() uint16 {
	return uint16(ln)
}

// Uint returns the underlying uint value (for backward compatibility).
// Deprecated: Use Uint16() instead for type safety.
func (ln LineNumber) Uint() uint {
	return uint(ln)
}

// MarshalJSON implements json.Marshaler for LineNumber.
func (ln LineNumber) MarshalJSON() ([]byte, error) {
	if ln == 0 {
		return nil, errors.NewValidationError("line number cannot be 0", nil)
	}
	return json.Marshal(uint16(ln))
}

// UnmarshalJSON implements json.Unmarshaler for LineNumber.
func (ln *LineNumber) UnmarshalJSON(data []byte) error {
	var n uint16
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal LineNumber: %w", err)
	}
	if n == 0 {
		return errors.NewValidationError("line number cannot be 0", nil)
	}
	*ln = LineNumber(n)
	return nil
}

// BytePosition represents a byte position in a file.
// Optimized: uint32 provides 0-4GB range (sufficient for file positions).
type BytePosition uint32

// NewBytePosition creates a validated BytePosition from a uint32.
func NewBytePosition(pos uint32) BytePosition {
	return BytePosition(pos)
}

// Uint32 returns the underlying uint32 value.
func (bp BytePosition) Uint32() uint32 {
	return uint32(bp)
}

// Uint returns the underlying uint value (for backward compatibility).
// Deprecated: Use Uint32() instead for type safety.
func (bp BytePosition) Uint() uint {
	return uint(bp)
}

// MarshalJSON implements json.Marshaler for BytePosition.
func (bp BytePosition) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(bp))
}

// UnmarshalJSON implements json.Unmarshaler for BytePosition.
func (bp *BytePosition) UnmarshalJSON(data []byte) error {
	var n uint32
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal BytePosition: %w", err)
	}
	*bp = BytePosition(n)
	return nil
}

// TokenCount represents a count of tokens in code.
type TokenCount uint

// NewTokenCount creates a validated TokenCount from a uint.
func NewTokenCount(count uint) TokenCount {
	return TokenCount(count)
}

// Uint returns the underlying uint value.
func (tc TokenCount) Uint() uint {
	return uint(tc)
}

// MarshalJSON implements json.Marshaler for TokenCount.
func (tc TokenCount) MarshalJSON() ([]byte, error) {
	return marshalUint(uint(tc))
}

// UnmarshalJSON implements json.Unmarshaler for TokenCount.
func (tc *TokenCount) UnmarshalJSON(data []byte) error {
	return unmarshalUint(data, "TokenCount", func(n uint) {
		*tc = TokenCount(n)
	})
}

// Confidence represents a confidence score (0.0 to 1.0).
type Confidence float64

// NewConfidence creates a validated Confidence from a float64.
// Returns error if the confidence is not between 0.0 and 1.0.
func NewConfidence(c float64) (Confidence, error) {
	if c < 0.0 || c > 1.0 {
		return 0, errors.NewValidationError(
			fmt.Sprintf("confidence must be between 0.0 and 1.0, got: %f", c),
			nil,
		)
	}
	return Confidence(c), nil
}

// Float64 returns the underlying float64 value.
func (conf Confidence) Float64() float64 {
	return float64(conf)
}

// String returns a percentage representation of confidence.
func (conf Confidence) String() string {
	return fmt.Sprintf("%.1f%%", float64(conf)*100)
}

// MarshalJSON implements json.Marshaler for Confidence.
func (conf Confidence) MarshalJSON() ([]byte, error) {
	if conf < 0.0 || conf > 1.0 {
		return nil, errors.NewValidationError(
			fmt.Sprintf("confidence must be between 0.0 and 1.0, got: %f", conf),
			nil,
		)
	}
	return json.Marshal(float64(conf))
}

// UnmarshalJSON implements json.Unmarshaler for Confidence.
func (conf *Confidence) UnmarshalJSON(data []byte) error {
	var c float64
	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("failed to unmarshal Confidence: %w", err)
	}
	if c < 0.0 || c > 1.0 {
		return errors.NewValidationError(
			fmt.Sprintf("confidence must be between 0.0 and 1.0, got: %f", c),
			nil,
		)
	}
	*conf = Confidence(c)
	return nil
}

// ComplexityScore represents a complexity metric.
// Optimized: uint16 provides 0-65,535 range (sufficient for code complexity).
type ComplexityScore uint16

// NewComplexityScore creates a validated ComplexityScore from a uint16.
func NewComplexityScore(score uint16) ComplexityScore {
	return ComplexityScore(score)
}

// Uint16 returns the underlying uint16 value.
func (cs ComplexityScore) Uint16() uint16 {
	return uint16(cs)
}

// Uint returns the underlying uint value (for backward compatibility).
// Deprecated: Use Uint16() instead for type safety.
func (cs ComplexityScore) Uint() uint {
	return uint(cs)
}

// MarshalJSON implements json.Marshaler for ComplexityScore.
func (cs ComplexityScore) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint16(cs))
}

// UnmarshalJSON implements json.Unmarshaler for ComplexityScore.
func (cs *ComplexityScore) UnmarshalJSON(data []byte) error {
	var n uint16
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal ComplexityScore: %w", err)
	}
	*cs = ComplexityScore(n)
	return nil
}

// Hash represents a hash value (typically SHA256).
type Hash string

// NewHash creates a validated Hash from a string.
func NewHash(h string) (Hash, error) {
	if h == "" {
		return "", errors.NewValidationError("hash cannot be empty", nil)
	}
	return Hash(h), nil
}

// String returns the string representation of Hash.
func (h Hash) String() string {
	return string(h)
}

// MarshalJSON implements json.Marshaler for Hash.
func (h Hash) MarshalJSON() ([]byte, error) {
	return marshalStringID(string(h), "Hash", "hash cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for Hash.
func (h *Hash) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "Hash", "hash cannot be empty", func(s string) {
		*h = Hash(s)
	})
}

// FileCount represents the number of files.
type FileCount uint

// NewFileCount creates a validated FileCount from a uint.
func NewFileCount(count uint) FileCount {
	return FileCount(count)
}

// Uint returns the underlying uint value.
func (fc FileCount) Uint() uint {
	return uint(fc)
}

// MarshalJSON implements json.Marshaler for FileCount.
func (fc FileCount) MarshalJSON() ([]byte, error) {
	return marshalUint(uint(fc))
}

// UnmarshalJSON implements json.Unmarshaler for FileCount.
func (fc *FileCount) UnmarshalJSON(data []byte) error {
	return unmarshalUint(data, "FileCount", func(n uint) {
		*fc = FileCount(n)
	})
}

// CloneCount represents the number of clones.
type CloneCount uint

// NewCloneCount creates a validated CloneCount from a uint.
func NewCloneCount(count uint) CloneCount {
	return CloneCount(count)
}

// Uint returns the underlying uint value.
func (cc CloneCount) Uint() uint {
	return uint(cc)
}

// MarshalJSON implements json.Marshaler for CloneCount.
func (cc CloneCount) MarshalJSON() ([]byte, error) {
	return marshalUint(uint(cc))
}

// UnmarshalJSON implements json.Unmarshaler for CloneCount.
func (cc *CloneCount) UnmarshalJSON(data []byte) error {
	return unmarshalUint(data, "CloneCount", func(n uint) {
		*cc = CloneCount(n)
	})
}

// ProcessingTime represents processing time in milliseconds.
type ProcessingTime uint

// NewProcessingTime creates a validated ProcessingTime from a uint.
// Returns error if the time is 0 (invalid).
func NewProcessingTime(time uint) (ProcessingTime, error) {
	if time == 0 {
		return 0, errors.NewValidationError("processing time cannot be 0", nil)
	}
	return ProcessingTime(time), nil
}

// Uint returns the underlying uint value.
func (pt ProcessingTime) Uint() uint {
	return uint(pt)
}

// String returns a human-readable representation of time.
func (pt ProcessingTime) String() string {
	ms := uint64(pt)
	if ms < 1000 {
		return strconv.FormatUint(ms, 10) + "ms"
	}
	seconds := ms / 1000
	if seconds < 60 {
		return strconv.FormatUint(seconds, 10) + "s"
	}
	minutes := seconds / 60
	if minutes < 60 {
		return strconv.FormatUint(minutes, 10) + "m"
	}
	hours := minutes / 60
	return strconv.FormatUint(hours, 10) + "h"
}

// MarshalJSON implements json.Marshaler for ProcessingTime.
func (pt ProcessingTime) MarshalJSON() ([]byte, error) {
	if pt == 0 {
		return nil, errors.NewValidationError("processing time cannot be 0", nil)
	}
	return json.Marshal(uint(pt))
}

// UnmarshalJSON implements json.Unmarshaler for ProcessingTime.
func (pt *ProcessingTime) UnmarshalJSON(data []byte) error {
	return unmarshalUintNonZero(data, "ProcessingTime", "processing time cannot be 0", func(n uint) {
		*pt = ProcessingTime(n)
	})
}

// Threshold represents the minimum token threshold for clone detection.
type Threshold uint

// NewThreshold creates a validated Threshold from a uint.
// Returns error if the threshold is 0 (invalid).
func NewThreshold(t uint) (Threshold, error) {
	if t == 0 {
		return 0, errors.NewValidationError("threshold cannot be 0", nil)
	}
	return Threshold(t), nil
}

// Uint returns the underlying uint value.
func (t Threshold) Uint() uint {
	return uint(t)
}

// MarshalJSON implements json.Marshaler for Threshold.
func (t Threshold) MarshalJSON() ([]byte, error) {
	if t == 0 {
		return nil, errors.NewValidationError("threshold cannot be 0", nil)
	}
	return json.Marshal(uint(t))
}

// UnmarshalJSON implements json.Unmarshaler for Threshold.
func (t *Threshold) UnmarshalJSON(data []byte) error {
	return unmarshalUintNonZero(data, "Threshold", "threshold cannot be 0", func(n uint) {
		*t = Threshold(n)
	})
}
