package domain

import (
	stderrors "errors"
	"fmt"
)

// Analysis represents main analysis domain object.
type Analysis struct {
	ID          AnalysisID         `json:"id"`
	State       DetectionState    `json:"state"`
	Mode        AnalysisMode      `json:"mode"`
	Threshold   Threshold         `json:"threshold"`
	CloneGroups []CloneGroup      `json:"cloneGroups"`
	Stats       AnalysisStats     `json:"stats"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   string            `json:"createdAt"`
	CompletedAt *string           `json:"completedAt,omitempty"`
}

func (a Analysis) IsValid() error {
	if !a.State.IsValid() {
		return fmt.Errorf("invalid analysis state: %s", a.State)
	}
	if !a.Mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", a.Mode)
	}
	if a.Threshold == 0 {
		return stderrors.New("analysis threshold cannot be zero")
	}
	if a.CreatedAt == "" {
		return stderrors.New("analysis created at cannot be empty")
	}

	for i, group := range a.CloneGroups {
		if err := group.IsValid(); err != nil {
			return fmt.Errorf("clone group %d in analysis is invalid: %w", i, err)
		}
	}
	return nil
}

// AnalysisStats represents analysis statistics.
type AnalysisStats struct {
	FilesAnalyzed    FileCount      `json:"filesAnalyzed"`
	TotalClones      CloneCount     `json:"totalClones"`
	TotalTokenSize   TokenCount     `json:"totalTokenSize"`
	ComplexityScore  float64        `json:"complexityScore"`
	DuplicationRatio float64        `json:"duplicationRatio"`
	ProcessingTime   ProcessingTime `json:"processingTime"`
}

func (as AnalysisStats) IsValid() error {
	return validateFields(
		validationRule{as.FilesAnalyzed > 0, "files analyzed cannot be zero"},
		validationRule{as.ProcessingTime > 0, "processing time cannot be zero"},
		validationRule{as.ComplexityScore >= 0, "complexity score cannot be negative"},
		validationRule{as.DuplicationRatio >= 0, "duplication ratio cannot be negative"},
	)
}
