package domain

import "errors"

// Static errors for domain type validation.
var (
	ErrInvalidCloneCategory      = errors.New("invalid clone category")
	ErrInvalidClonePriority      = errors.New("invalid clone priority")
	ErrInvalidCloneActionability = errors.New("invalid clone actionability")
	ErrInvalidCloneSeverity      = errors.New("invalid clone severity")
	ErrInvalidHealthScore        = errors.New("invalid health score")
	ErrEmptyFilename             = errors.New("filename cannot be empty")
	ErrLineEndBeforeStart        = errors.New("line end is before line start")
	ErrNegativeTokenCount        = errors.New("token count cannot be negative")
)
