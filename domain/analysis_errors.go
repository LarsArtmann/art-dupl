package domain

import "errors"

// Static errors for Analysis validation.
var (
	ErrInvalidAnalysisState    = errors.New("invalid analysis state")
	ErrInvalidAnalysisMode     = errors.New("invalid analysis mode")
	ErrAnalysisThresholdZero   = errors.New("analysis threshold cannot be zero")
	ErrAnalysisCreatedAtEmpty  = errors.New("analysis created at cannot be empty")
	ErrCloneEndLineInvalid     = errors.New("clone end line must be >= start line")
	ErrCloneEndPositionInvalid = errors.New("clone end position must be > start position")
	ErrCloneGroupEmpty         = errors.New("clone group must have at least one clone")
	ErrThresholdInvalid        = errors.New("threshold must be > 0")
	ErrNoPathsSpecified        = errors.New("at least one path must be specified")
	ErrRepositoryPathEmpty     = errors.New("repository path cannot be empty")
	ErrRepositoryNameEmpty     = errors.New("repository name cannot be empty")
	ErrRepositoryLanguageEmpty = errors.New("repository language cannot be empty")
)
