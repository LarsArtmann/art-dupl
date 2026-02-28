package domain

import (
	"fmt"
)

// DetectionOptions represents configuration for detection.
type DetectionOptions struct {
	Threshold     Threshold    `json:"threshold"`
	Mode          AnalysisMode `json:"mode"`
	IncludeVendor bool         `json:"includeVendor"`
	Verbose       bool         `json:"verbose"`
	Paths         []Filepath   `json:"paths"`
	OutputFormat  string       `json:"outputFormat"`
}

func (do DetectionOptions) IsValid() error {
	if do.Threshold == 0 {
		return ErrThresholdInvalid
	}

	if !do.Mode.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidAnalysisMode, do.Mode)
	}

	if len(do.Paths) == 0 {
		return ErrNoPathsSpecified
	}

	return nil
}
