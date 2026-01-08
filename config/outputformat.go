package config

// OutputFormat represents the supported output formats with type safety.
type OutputFormat string

const (
	OutputFormatText     OutputFormat = "text"
	OutputFormatHTML     OutputFormat = "html"
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatPlumbing OutputFormat = "plumbing"
)

// String implements fmt.Stringer for OutputFormat.
func (of OutputFormat) String() string {
	return string(of)
}

// IsValid checks if the output format is supported.
func (of OutputFormat) IsValid() bool {
	switch of {
	case OutputFormatText, OutputFormatHTML, OutputFormatJSON, OutputFormatPlumbing:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler for OutputFormat.
func (of OutputFormat) MarshalJSON() ([]byte, error) {
	return MarshalEnumJSON(of, OutputFormat.IsValid, "output format")
}

// UnmarshalJSON implements json.Unmarshaler for OutputFormat.
func (of *OutputFormat) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONForEnum(of, data, "output format")
}

// SortCriteria represents supported sorting criteria with type safety.
type SortCriteria string

const (
	SortBySize        SortCriteria = "size"
	SortByOccurrence  SortCriteria = "occurrence"
	SortByHash        SortCriteria = "hash"
	SortByTotalTokens SortCriteria = "total-tokens"
)

// String implements fmt.Stringer for SortCriteria.
func (sc SortCriteria) String() string {
	return string(sc)
}

// IsValid checks if the sort criteria is supported.
func (sc SortCriteria) IsValid() bool {
	switch sc {
	case SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler for SortCriteria.
func (sc SortCriteria) MarshalJSON() ([]byte, error) {
	return MarshalEnumJSON(sc, SortCriteria.IsValid, "sort criteria")
}

// UnmarshalJSON implements json.Unmarshaler for SortCriteria.
func (sc *SortCriteria) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONForEnum(sc, data, "sort criteria")
}

// AllOutputFormats returns list of all supported output formats.
func AllOutputFormats() []OutputFormat {
	return []OutputFormat{
		OutputFormatText,
		OutputFormatHTML,
		OutputFormatJSON,
		OutputFormatPlumbing,
	}
}

// AllSortCriteria returns list of all supported sort criteria.
func AllSortCriteria() []SortCriteria {
	return []SortCriteria{
		SortBySize,
		SortByOccurrence,
		SortByHash,
		SortByTotalTokens,
	}
}
