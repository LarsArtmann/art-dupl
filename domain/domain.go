// Package domain provides core domain types and logic for art-dupl.
//
// This package defines value objects for code duplication detection:
// - ProcessedClone / ProcessedCloneGroup: canonical clone data with validation
// - CloneClassification: category, priority, and actionability metadata
// - ClonePriority, CloneCategory, CloneActionability: typed enums with JSON marshaling
// - Error sentinels for clone validation failures
package domain
