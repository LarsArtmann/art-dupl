package domain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/LarsArtmann/art-dupl/errors"
)

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

	return json.Marshal(float64(conf)) //nolint:wrapcheck // Standard JSON marshaling
}

// UnmarshalJSON implements json.Unmarshaler for Confidence.
func (conf *Confidence) UnmarshalJSON(data []byte) error {
	var c float64

	err := json.Unmarshal(data, &c)
	if err != nil {
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
//
// Deprecated: Use Uint16() instead for type safety.
func (cs ComplexityScore) Uint() uint {
	return uint(cs)
}

// MarshalJSON implements json.Marshaler for ComplexityScore.
func (cs ComplexityScore) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint16(cs)) //nolint:wrapcheck // Standard JSON marshaling
}

// UnmarshalJSON implements json.Unmarshaler for ComplexityScore.
func (cs *ComplexityScore) UnmarshalJSON(data []byte) error {
	return unmarshalUintGeneric(data, "ComplexityScore", func(n uint16) {
		*cs = ComplexityScore(n)
	})
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
	return marshalStringID(string(h), "hash cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for Hash.
func (h *Hash) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "Hash", "hash cannot be empty", func(s string) {
		*h = Hash(s)
	})
}

// ProcessingTime represents processing time in milliseconds.
type ProcessingTime uint

// NewProcessingTime creates a validated ProcessingTime from a uint.
// Returns error if the time is 0 (invalid).
func NewProcessingTime(time uint) (ProcessingTime, error) {
	if time == 0 {
		return 0, errors.NewValidationError(
			fmt.Sprintf("processing time cannot be 0 (actual=%d)", time),
			nil,
		)
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
		return nil, errors.NewValidationError(
			fmt.Sprintf("processing time cannot be 0 (actual=%d)", pt),
			nil,
		)
	}

	data, err := json.Marshal(uint(pt))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ProcessingTime: %w", err)
	}

	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for ProcessingTime.
func (pt *ProcessingTime) UnmarshalJSON(data []byte) error {
	return unmarshalUintNonZero(
		data,
		"ProcessingTime",
		"processing time cannot be 0",
		func(n uint) {
			*pt = ProcessingTime(n)
		},
	)
}
