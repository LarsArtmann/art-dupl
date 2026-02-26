package domain

import (
	"github.com/LarsArtmann/art-dupl/errors"
)

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

	return marshalUint(uint(t))
}

// UnmarshalJSON implements json.Unmarshaler for Threshold.
func (t *Threshold) UnmarshalJSON(data []byte) error {
	return unmarshalUintNonZero(data, "Threshold", "threshold cannot be 0", func(n uint) {
		*t = Threshold(n)
	})
}
