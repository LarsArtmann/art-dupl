package config

// SortCriteria represents the sort criteria type.
type SortCriteria string

const (
	SortBySize        SortCriteria = "size"
	SortByOccurrence  SortCriteria = "occurrence"
	SortByHash        SortCriteria = "hash"
	SortByTotalTokens SortCriteria = "total-tokens"
)

//nolint:gochecknoglobals // Lookup table for valid sort criteria
var validSortCriteria = map[SortCriteria]bool{
	SortBySize:        true,
	SortByOccurrence:  true,
	SortByHash:        true,
	SortByTotalTokens: true,
}

// String implements fmt.Stringer.
func (sc SortCriteria) String() string {
	return string(sc)
}

// IsValid validates sort criteria.
func (sc SortCriteria) IsValid() bool {
	return isValidStringType(sc, validSortCriteria)
}

// MarshalJSON implements json.Marshaler.
func (sc SortCriteria) MarshalJSON() ([]byte, error) {
	return marshalStringType(sc, isValidMethod[SortCriteria](), "sort criteria")
}

// UnmarshalJSON implements json.Unmarshaler.
func (sc *SortCriteria) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		isValidMethod[SortCriteria](),
		SortBySize,
		"sort criteria",
		sc,
	)
}

// AllSortCriteria returns all supported sort criteria.
func AllSortCriteria() []SortCriteria {
	return []SortCriteria{
		SortBySize,
		SortByOccurrence,
		SortByHash,
		SortByTotalTokens,
	}
}

// DefaultSortCriteria returns the default sort criteria.
func DefaultSortCriteria() SortCriteria {
	return SortBySize
}
