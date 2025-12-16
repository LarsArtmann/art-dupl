package config

// OutputFormat represents the supported output formats with type safety
type OutputFormat string

const (
	OutputFormatText     OutputFormat = "text"
	OutputFormatHTML     OutputFormat = "html"
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatPlumbing OutputFormat = "plumbing"
)

// String implements fmt.Stringer for OutputFormat
func (of OutputFormat) String() string {
	return string(of)
}

// IsValid checks if the output format is supported
func (of OutputFormat) IsValid() bool {
	switch of {
	case OutputFormatText, OutputFormatHTML, OutputFormatJSON, OutputFormatPlumbing:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler for OutputFormat
func (of OutputFormat) MarshalJSON() ([]byte, error) {
	// Explicitly type isValid function
	var isValid = func(o OutputFormat) bool { return o.IsValid() }
	return MarshalEnumJSON(of, isValid, "output format")
}

// UnmarshalJSON implements json.Unmarshaler for OutputFormat
func (of *OutputFormat) UnmarshalJSON(data []byte) error {
	candidate, err := UnmarshalEnumJSON(data, func(s string) OutputFormat { return OutputFormat(s) }, func(o OutputFormat) bool {
		return o.IsValid()
	}, "output format")
	if err != nil {
		return err
	}
	*of = candidate
	return nil
}

// SortCriteria represents supported sorting criteria with type safety
type SortCriteria string

const (
	SortBySize        SortCriteria = "size"
	SortByOccurrence  SortCriteria = "occurrence"
	SortByHash        SortCriteria = "hash"
	SortByTotalTokens SortCriteria = "total-tokens"
)

// String implements fmt.Stringer for SortCriteria
func (sc SortCriteria) String() string {
	return string(sc)
}

// IsValid checks if the sort criteria is supported
func (sc SortCriteria) IsValid() bool {
	switch sc {
	case SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler for SortCriteria
func (sc SortCriteria) MarshalJSON() ([]byte, error) {
	// Explicitly type isValid function
	var isValid = func(s SortCriteria) bool { return s.IsValid() }
	return MarshalEnumJSON(sc, isValid, "sort criteria")
}

// UnmarshalJSON implements json.Unmarshaler for SortCriteria
func (sc *SortCriteria) UnmarshalJSON(data []byte) error {
	candidate, err := UnmarshalEnumJSON(data, func(s string) SortCriteria { return SortCriteria(s) }, func(s SortCriteria) bool {
		return s.IsValid()
	}, "sort criteria")
	if err != nil {
		return err
	}
	*sc = candidate
	return nil
}

// AllOutputFormats returns list of all supported output formats
func AllOutputFormats() []OutputFormat {
	return []OutputFormat{
		OutputFormatText,
		OutputFormatHTML,
		OutputFormatJSON,
		OutputFormatPlumbing,
	}
}

// AllSortCriteria returns list of all supported sort criteria
func AllSortCriteria() []SortCriteria {
	return []SortCriteria{
		SortBySize,
		SortByOccurrence,
		SortByHash,
		SortByTotalTokens,
	}
}
