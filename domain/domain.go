// Package domain provides core domain types and logic for art-dupl.
//
// This package defines value objects for code duplication detection:
// - LineNumber, BytePosition, Filepath: file location types
// - Threshold, TokenCount, FileCount, CloneCount: metric types
// - CloneSeverity: classification of duplicate importance
// - Validation helpers for type-safe construction
//
// All value objects are immutable and validated at construction time.
package domain
