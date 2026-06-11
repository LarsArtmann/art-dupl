package domain

import "errors"

// Static errors for domain type validation.
var (
	ErrInvalidCloneCategory      = errors.New("invalid clone category")
	ErrInvalidClonePriority      = errors.New("invalid clone priority")
	ErrInvalidCloneActionability = errors.New("invalid clone actionability")
	ErrInvalidCloneSeverity      = errors.New("invalid clone severity")
	ErrInvalidSeverity           = errors.New("invalid severity value")
	ErrInvalidHealthScore        = errors.New("invalid health score")
)
