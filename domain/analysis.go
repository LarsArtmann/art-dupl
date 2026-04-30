package domain

import (
	"fmt"
	"time"
)

// Analysis represents main analysis domain object.
type Analysis struct {
	ID          AnalysisID        `json:"id"`
	State       DetectionState    `json:"state"`
	Mode        AnalysisMode      `json:"mode"`
	Threshold   Threshold         `json:"threshold"`
	CloneGroups []CloneGroup      `json:"cloneGroups"`
	Stats       AnalysisStats     `json:"stats"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"createdAt"`
	CompletedAt *time.Time        `json:"completedAt,omitempty"`
}

func (a Analysis) IsValid() error {
	if !a.State.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidAnalysisState, a.State)
	}

	if !a.Mode.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidAnalysisMode, a.Mode)
	}

	if a.Threshold == 0 {
		return fmt.Errorf("%w: ID=%s", ErrAnalysisThresholdZero, a.ID)
	}

	if a.CreatedAt.IsZero() {
		return fmt.Errorf("%w: ID=%s", ErrAnalysisCreatedAtEmpty, a.ID)
	}

	for i, group := range a.CloneGroups {
		err := group.IsValid()
		if err != nil {
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
		validationRule{
			as.FilesAnalyzed > 0,
			fmt.Sprintf("files analyzed cannot be zero (actual=%d)", as.FilesAnalyzed),
		},
		validationRule{
			as.ProcessingTime > 0,
			fmt.Sprintf("processing time cannot be zero (actual=%d)", as.ProcessingTime),
		},
		validationRule{
			as.ComplexityScore >= 0,
			fmt.Sprintf("complexity score cannot be negative (actual=%f)", as.ComplexityScore),
		},
		validationRule{
			as.DuplicationRatio >= 0,
			fmt.Sprintf("duplication ratio cannot be negative (actual=%f)", as.DuplicationRatio),
		},
	)
}
