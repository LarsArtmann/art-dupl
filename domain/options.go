package domain

import (
	stderrors "errors"
	"fmt"
)

// DetectionOptions represents configuration for detection.
type DetectionOptions struct {
	Threshold     uint         `json:"threshold"`
	Mode          AnalysisMode `json:"mode"`
	IncludeVendor bool         `json:"includeVendor"`
	Verbose       bool         `json:"verbose"`
	Paths         []string     `json:"paths"`
	OutputFormat  string       `json:"outputFormat"`
}

func (do DetectionOptions) IsValid() error {
	if do.Threshold == 0 {
		return stderrors.New("threshold must be > 0")
	}
	if !do.Mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", do.Mode)
	}
	if len(do.Paths) == 0 {
		return stderrors.New("at least one path must be specified")
	}
	return nil
}
