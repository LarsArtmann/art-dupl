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

// unmarshalStringID is a helper function for unmarshaling string-based ID types.
// It handles the common pattern of unmarshaling JSON to string and validating emptiness.
func unmarshalStringID(data []byte, typeName string, validationMsg string, assign func(string)) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", typeName, err)
	}
	if s == "" {
		return errors.NewValidationError(validationMsg, nil)
	}
	assign(s)
	return nil
}

// CloneID represents a unique identifier for a code clone.
type CloneID string

// NewCloneID creates a validated CloneID from a string.
// Returns error if the ID is empty.
func NewCloneID(id string) (CloneID, error) {
	if id == "" {
		return "", errors.NewValidationError("clone ID cannot be empty", nil)
	}
	return CloneID(id), nil
}

// String returns the string representation of CloneID.
func (id CloneID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for CloneID.
func (id CloneID) MarshalJSON() ([]byte, error) {
	if id == "" {
		return nil, errors.NewValidationError("clone ID cannot be empty", nil)
	}
	return json.Marshal(string(id))
}

// UnmarshalJSON implements json.Unmarshaler for CloneID.
func (id *CloneID) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "CloneID", "clone ID cannot be empty", func(s string) {
		*id = CloneID(s)
	})
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
	if id == "" {
		return nil, errors.NewValidationError("clone group ID cannot be empty", nil)
	}
	return json.Marshal(string(id))
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
	if id == "" {
		return nil, errors.NewValidationError("analysis ID cannot be empty", nil)
	}
	return json.Marshal(string(id))
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
	if fp == "" {
		return nil, errors.NewValidationError("filepath cannot be empty", nil)
	}
	return json.Marshal(string(fp))
}

// UnmarshalJSON implements json.Unmarshaler for Filepath.
func (fp *Filepath) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "Filepath", "filepath cannot be empty", func(s string) {
		*fp = Filepath(s)
	})
}

// LineNumber represents a line number in a source file.
// Line numbers start at 1 (not 0) in most editors.
type LineNumber uint

// NewLineNumber creates a validated LineNumber from a uint.
// Returns error if the line number is 0 (invalid).
func NewLineNumber(n uint) (LineNumber, error) {
	if n == 0 {
		return 0, errors.NewValidationError("line number cannot be 0", nil)
	}
	return LineNumber(n), nil
}

// Uint returns the underlying uint value.
func (ln LineNumber) Uint() uint {
	return uint(ln)
}

// MarshalJSON implements json.Marshaler for LineNumber.
func (ln LineNumber) MarshalJSON() ([]byte, error) {
	if ln == 0 {
		return nil, errors.NewValidationError("line number cannot be 0", nil)
	}
	return json.Marshal(uint(ln))
}

// UnmarshalJSON implements json.Unmarshaler for LineNumber.
func (ln *LineNumber) UnmarshalJSON(data []byte) error {
	var n uint
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
type BytePosition uint

// NewBytePosition creates a validated BytePosition from a uint.
func NewBytePosition(pos uint) BytePosition {
	return BytePosition(pos)
}

// Uint returns the underlying uint value.
func (bp BytePosition) Uint() uint {
	return uint(bp)
}

// MarshalJSON implements json.Marshaler for BytePosition.
func (bp BytePosition) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint(bp))
}

// UnmarshalJSON implements json.Unmarshaler for BytePosition.
func (bp *BytePosition) UnmarshalJSON(data []byte) error {
	var n uint
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
	return json.Marshal(uint(tc))
}

// UnmarshalJSON implements json.Unmarshaler for TokenCount.
func (tc *TokenCount) UnmarshalJSON(data []byte) error {
	var n uint
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal TokenCount: %w", err)
	}
	*tc = TokenCount(n)
	return nil
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
type ComplexityScore uint

// NewComplexityScore creates a validated ComplexityScore from a uint.
func NewComplexityScore(score uint) ComplexityScore {
	return ComplexityScore(score)
}

// Uint returns the underlying uint value.
func (cs ComplexityScore) Uint() uint {
	return uint(cs)
}

// MarshalJSON implements json.Marshaler for ComplexityScore.
func (cs ComplexityScore) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint(cs))
}

// UnmarshalJSON implements json.Unmarshaler for ComplexityScore.
func (cs *ComplexityScore) UnmarshalJSON(data []byte) error {
	var n uint
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
	if h == "" {
		return nil, errors.NewValidationError("hash cannot be empty", nil)
	}
	return json.Marshal(string(h))
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
	return json.Marshal(uint(fc))
}

// UnmarshalJSON implements json.Unmarshaler for FileCount.
func (fc *FileCount) UnmarshalJSON(data []byte) error {
	var n uint
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal FileCount: %w", err)
	}
	*fc = FileCount(n)
	return nil
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
	return json.Marshal(uint(cc))
}

// UnmarshalJSON implements json.Unmarshaler for CloneCount.
func (cc *CloneCount) UnmarshalJSON(data []byte) error {
	var n uint
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal CloneCount: %w", err)
	}
	*cc = CloneCount(n)
	return nil
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
	var n uint
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal ProcessingTime: %w", err)
	}
	if n == 0 {
		return errors.NewValidationError("processing time cannot be 0", nil)
	}
	*pt = ProcessingTime(n)
	return nil
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
	var n uint
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("failed to unmarshal Threshold: %w", err)
	}
	if n == 0 {
		return errors.NewValidationError("threshold cannot be 0", nil)
	}
	*t = Threshold(n)
	return nil
}
