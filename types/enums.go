package types

import "fmt"

// DetectionState represents different detection states with type safety
type DetectionState string

const (
	DetectionStateUnknown   DetectionState = "unknown"
	DetectionStatePending   DetectionState = "pending"
	DetectionStateRunning   DetectionState = "running"
	DetectionStateCompleted DetectionState = "completed"
	DetectionStateFailed    DetectionState = "failed"
)

// IsValid checks if detection state is valid
func (ds DetectionState) IsValid() bool {
	switch ds {
	case DetectionStateUnknown, DetectionStatePending, DetectionStateRunning,
		DetectionStateCompleted, DetectionStateFailed:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer
func (ds DetectionState) String() string {
	return string(ds)
}

// MarshalJSON implements json.Marshaler for API consistency
func (ds DetectionState) MarshalJSON() ([]byte, error) {
	if !ds.IsValid() {
		return nil, fmt.Errorf("invalid detection state: %s", ds)
	}
	return []byte(`"` + string(ds) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for API consistency
func (ds *DetectionState) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	state := DetectionState(str)
	if !state.IsValid() {
		return fmt.Errorf("invalid detection state: %s", str)
	}
	*ds = state
	return nil
}

// AnalysisMode represents different analysis modes
type AnalysisMode string

const (
	AnalysisModeFull        AnalysisMode = "full"
	AnalysisModeIncremental AnalysisMode = "incremental"
	AnalysisModeQuick       AnalysisMode = "quick"
	AnalysisModeDeep        AnalysisMode = "deep"
)

// IsValid checks if analysis mode is valid
func (am AnalysisMode) IsValid() bool {
	switch am {
	case AnalysisModeFull, AnalysisModeIncremental, AnalysisModeQuick, AnalysisModeDeep:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer
func (am AnalysisMode) String() string {
	return string(am)
}

// MarshalJSON implements json.Marshaler for API consistency
func (am AnalysisMode) MarshalJSON() ([]byte, error) {
	if !am.IsValid() {
		return nil, fmt.Errorf("invalid analysis mode: %s", am)
	}
	return []byte(`"` + string(am) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for API consistency
func (am *AnalysisMode) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	mode := AnalysisMode(str)
	if !mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", str)
	}
	*am = mode
	return nil
}

// FileProcessingState represents file processing states
type FileProcessingState string

const (
	FileProcessingStateQueued     FileProcessingState = "queued"
	FileProcessingStateReading    FileProcessingState = "reading"
	FileProcessingStateProcessing FileProcessingState = "processing"
	FileProcessingStateCompleted  FileProcessingState = "completed"
	FileProcessingStateError      FileProcessingState = "error"
)

// IsValid checks if file processing state is valid
func (fps FileProcessingState) IsValid() bool {
	switch fps {
	case FileProcessingStateQueued, FileProcessingStateReading,
		FileProcessingStateProcessing, FileProcessingStateCompleted, FileProcessingStateError:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer
func (fps FileProcessingState) String() string {
	return string(fps)
}

// MarshalJSON implements json.Marshaler for API consistency
func (fps FileProcessingState) MarshalJSON() ([]byte, error) {
	if !fps.IsValid() {
		return nil, fmt.Errorf("invalid file processing state: %s", fps)
	}
	return []byte(`"` + string(fps) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for API consistency
func (fps *FileProcessingState) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	state := FileProcessingState(str)
	if !state.IsValid() {
		return fmt.Errorf("invalid file processing state: %s", str)
	}
	*fps = state
	return nil
}
