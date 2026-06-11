package domain

import "errors"

// Static errors for domain type validation.
var (
	ErrInvalidCloneSeverity = errors.New("invalid clone severity")
	ErrInvalidSeverity      = errors.New("invalid severity value")
	ErrInvalidHealthScore   = errors.New("invalid health score")
)
