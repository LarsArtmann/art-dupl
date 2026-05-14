package domain

// ProcessedClone represents a single clone instance with extracted fragment data.
// Decouples printer output from syntax.Node internals.
type ProcessedClone struct {
	Filename       string
	LineStart      int
	LineEnd        int
	Fragment       []byte
	Size           int                 // Token count / fragment size
	FileSize       int                 // Full file size in bytes
	Classification CloneClassification // Classification metadata for reports
}

// ProcessedCloneGroup represents a group of duplicate code fragments.
type ProcessedCloneGroup struct {
	Hash   string
	Size   int
	Clones []ProcessedClone
}

// CloneClassification categorizes clones for actionable reports.
type CloneClassification int

const (
	CloneUnknown       CloneClassification = iota
	CloneExact                             // Identical clones (byte-for-byte)
	CloneNearMiss                          // Slightly different (semantic)
	CloneStructural                        // Same structure, different names
	CloneFileDuplicate                     // Entire file is duplicate
)

// Name returns a human-readable name for the classification.
func (c CloneClassification) Name() string {
	switch c {
	case CloneUnknown:
		return "Unknown"
	case CloneExact:
		return "Exact"
	case CloneNearMiss:
		return "Near-Miss"
	case CloneStructural:
		return "Structural"
	case CloneFileDuplicate:
		return "File Duplicate"
	default:
		return "Unknown"
	}
}
