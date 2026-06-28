// Package config re-exports domain enum types for backward compatibility.
// The canonical definitions live in domain/ — these aliases ensure existing
// code using config.SortCriteria continues to work without import changes.
package config

import "github.com/LarsArtmann/art-dupl/domain"

type SortCriteria = domain.SortCriteria

const (
	SortBySize        = domain.SortBySize
	SortByOccurrence  = domain.SortByOccurrence
	SortByHash        = domain.SortByHash
	SortByTotalTokens = domain.SortByTotalTokens
)

var (
	ErrInvalidSortCriteria = domain.ErrInvalidSortCriteria
	AllSortCriteria        = domain.AllSortCriteria
	DefaultSortCriteria    = domain.DefaultSortCriteria
	ParseSortCriteria      = domain.ParseSortCriteria
)
