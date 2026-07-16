package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// ErrInvalidSortCriteria is returned when a sort criteria value is not recognized.
var ErrInvalidSortCriteria = errors.New("invalid sort criteria")

// SortCriteria controls how clone groups are ordered in output.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type SortCriteria string

const (
	SortBySize        SortCriteria = "size"
	SortByOccurrence  SortCriteria = "occurrence"
	SortByHash        SortCriteria = "hash"
	SortByTotalTokens SortCriteria = "total-tokens"
)

func (sc SortCriteria) String() string { return string(sc) }

func (sc SortCriteria) IsValid() bool {
	switch sc {
	case SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens:
		return true
	default:
		return false
	}
}

func (sc SortCriteria) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(sc, SortCriteria.IsValid, ErrInvalidSortCriteria)
}

func (sc *SortCriteria) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(sc, data, SortCriteria.IsValid, ErrInvalidSortCriteria)
}

func AllSortCriteria() []SortCriteria {
	return []SortCriteria{SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens}
}

func DefaultSortCriteria() SortCriteria {
	return SortBySize
}

func ParseSortCriteria(value string) (SortCriteria, error) {
	sc := SortCriteria(strings.ToLower(value))
	if sc.IsValid() {
		return sc, nil
	}

	return "", fmt.Errorf("%w: %q must be one of (size|occurrence|hash|total-tokens)",
		ErrInvalidSortCriteria, value)
}
