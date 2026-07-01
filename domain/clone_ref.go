package domain

// CloneRef carries the location and source text of a single clone occurrence.
// It is designed for embedding into type-specific Clone structs (ProcessedClone,
// artdupl.Clone, etc.) to eliminate field-name drift without collapsing the
// DTO boundary between packages.
type CloneRef struct {
	Filename  string `json:"filename"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
	Fragment  string `json:"fragment,omitempty"`
}

// LineCount returns the number of source lines spanned by this clone.
func (r CloneRef) LineCount() int {
	return r.LineEnd - r.LineStart + 1
}
