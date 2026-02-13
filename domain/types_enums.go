package domain

// FileProcessingState represents the processing state of a file.
type FileProcessingState string

const (
	FileProcessingStatePending    FileProcessingState = "pending"
	FileProcessingStateProcessing FileProcessingState = "processing"
	FileProcessingStateCompleted  FileProcessingState = "completed"
	FileProcessingStateFailed     FileProcessingState = "failed"
)

func (fps FileProcessingState) String() string { return string(fps) }

func (fps FileProcessingState) IsValid() bool {
	switch fps {
	case FileProcessingStatePending, FileProcessingStateProcessing, FileProcessingStateCompleted, FileProcessingStateFailed:
		return true
	default:
		return false
	}
}

// DetectionState represents the state of clone detection.
type DetectionState string

const (
	DetectionStateIdle      DetectionState = "idle"
	DetectionStateRunning   DetectionState = "running"
	DetectionStateCompleted DetectionState = "completed"
	DetectionStateFailed    DetectionState = "failed"
)

func (ds DetectionState) String() string { return string(ds) }

func (ds DetectionState) IsValid() bool {
	switch ds {
	case DetectionStateIdle, DetectionStateRunning, DetectionStateCompleted, DetectionStateFailed:
		return true
	default:
		return false
	}
}

// AnalysisMode represents the mode of code analysis.
type AnalysisMode string

const (
	AnalysisModeFull  AnalysisMode = "full"
	AnalysisModeQuick AnalysisMode = "quick"
	AnalysisModeDeep  AnalysisMode = "deep"
)

func (am AnalysisMode) String() string { return string(am) }

func (am AnalysisMode) IsValid() bool {
	switch am {
	case AnalysisModeFull, AnalysisModeQuick, AnalysisModeDeep:
		return true
	default:
		return false
	}
}
