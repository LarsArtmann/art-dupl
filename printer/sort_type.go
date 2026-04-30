package printer

import "github.com/LarsArtmann/art-dupl/config"

// SortBy is an alias for config.SortCriteria for backward compatibility.
// All sort logic and validation lives in config.SortCriteria.
type SortBy = config.SortCriteria

const (
	SortBySize        = config.SortBySize
	SortByOccurrence  = config.SortByOccurrence
	SortByHash        = config.SortByHash
	SortByTotalTokens = config.SortByTotalTokens
)

// ParseSortBy converts a string to SortBy with validation.
// Returns an error if the value is not a valid sorting criterion.
func ParseSortBy(value string) (SortBy, error) {
	return config.ParseSortCriteria(value)
}
