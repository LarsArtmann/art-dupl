package types

// DetectionState represents different detection states with type safety.
type DetectionState string

const (
	DetectionStateUnknown   DetectionState = "unknown"
	DetectionStatePending   DetectionState = "pending"
	DetectionStateRunning   DetectionState = "running"
	DetectionStateCompleted DetectionState = "completed"
	DetectionStateFailed    DetectionState = "failed"
)

// IsValid checks if detection state is valid.
func (ds DetectionState) IsValid() bool {
	switch ds {
	case DetectionStateUnknown, DetectionStatePending, DetectionStateRunning,
		DetectionStateCompleted, DetectionStateFailed:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (ds DetectionState) String() string {
	return string(ds)
}

// MarshalJSON implements json.Marshaler for API consistency.
func (ds DetectionState) MarshalJSON() ([]byte, error) {
	return MarshalEnumJSON(ds, "detection state")
}

// UnmarshalJSON implements json.Unmarshaler for API consistency.
func (ds *DetectionState) UnmarshalJSON(data []byte) error {
	enum, err := UnmarshalEnumJSON(data, func(s string) DetectionState { return DetectionState(s) }, "detection state")
	if err != nil {
		return err
	}
	*ds = *enum
	return nil
}

// AnalysisMode represents different analysis modes.
type AnalysisMode string

const (
	AnalysisModeFull        AnalysisMode = "full"
	AnalysisModeIncremental AnalysisMode = "incremental"
	AnalysisModeQuick       AnalysisMode = "quick"
	AnalysisModeDeep        AnalysisMode = "deep"
)

// IsValid checks if analysis mode is valid.
func (am AnalysisMode) IsValid() bool {
	switch am {
	case AnalysisModeFull, AnalysisModeIncremental, AnalysisModeQuick, AnalysisModeDeep:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (am AnalysisMode) String() string {
	return string(am)
}

// MarshalJSON implements json.Marshaler for API consistency.
func (am AnalysisMode) MarshalJSON() ([]byte, error) {
	return MarshalEnumJSON(am, "analysis mode")
}

// UnmarshalJSON implements json.Unmarshaler for API consistency.
func (am *AnalysisMode) UnmarshalJSON(data []byte) error {
	enum, err := UnmarshalEnumJSON(data, func(s string) AnalysisMode { return AnalysisMode(s) }, "analysis mode")
	if err != nil {
		return err
	}
	*am = *enum
	return nil
}

// FileProcessingState represents file processing states.
type FileProcessingState string

const (
	FileProcessingStateQueued     FileProcessingState = "queued"
	FileProcessingStateReading    FileProcessingState = "reading"
	FileProcessingStateProcessing FileProcessingState = "processing"
	FileProcessingStateCompleted  FileProcessingState = "completed"
	FileProcessingStateError      FileProcessingState = "error"
)

// IsValid checks if file processing state is valid.
func (fps FileProcessingState) IsValid() bool {
	switch fps {
	case FileProcessingStateQueued, FileProcessingStateReading,
		FileProcessingStateProcessing, FileProcessingStateCompleted, FileProcessingStateError:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (fps FileProcessingState) String() string {
	return string(fps)
}

// MarshalJSON implements json.Marshaler for API consistency.
func (fps FileProcessingState) MarshalJSON() ([]byte, error) {
	return MarshalEnumJSON(fps, "file processing state")
}

// UnmarshalJSON implements json.Unmarshaler for API consistency.
func (fps *FileProcessingState) UnmarshalJSON(data []byte) error {
	enum, err := UnmarshalEnumJSON(data, func(s string) FileProcessingState { return FileProcessingState(s) }, "file processing state")
	if err != nil {
		return err
	}
	*fps = *enum
	return nil
}
