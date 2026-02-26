package printer

import (
	"fmt"
	"strings"
)

// SortBy represents a sorting criterion for clone groups.
type SortBy string

const (
	SortBySize        SortBy = "size"
	SortByOccurrence  SortBy = "occurrence"
	SortByHash        SortBy = "hash"
	SortByTotalTokens SortBy = "total-tokens"
)

// ParseSortBy converts a string to SortBy with validation.
// Returns an error if the value is not a valid sorting criterion.
func ParseSortBy(value string) (SortBy, error) {
	sortBy := SortBy(strings.ToLower(value))
	switch sortBy {
	case SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens:
		return sortBy, nil
	default:
		return "", fmt.Errorf(
			"invalid sort criteria '%s': must be one of (size|occurrence|hash|total-tokens)",
			value,
		)
	}
}

// IsValid returns true if the SortBy value is valid.
func (s SortBy) IsValid() bool {
	switch s {
	case SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens:
		return true
	default:
		return false
	}
}

// String returns the string representation of SortBy.
func (s SortBy) String() string {
	return string(s)
}
