package domain

import "errors"

// Static errors for domain type validation.
var (
	ErrInvalidCloneCategory      = errors.New("invalid clone category")
	ErrInvalidClonePriority      = errors.New("invalid clone priority")
	ErrInvalidCloneActionability = errors.New("invalid clone actionability")
	ErrInvalidHealthScore        = errors.New("invalid health score")
	ErrEmptyFilename             = errors.New("filename cannot be empty")
	ErrLineEndBeforeStart        = errors.New("line end is before line start")
	ErrNegativeTokenCount        = errors.New("token count cannot be negative")
	ErrEmptyCloneGroup           = errors.New("clone group must have at least one clone")
	ErrTokenCountMismatch        = errors.New("group token count does not match sum of clones")
)
