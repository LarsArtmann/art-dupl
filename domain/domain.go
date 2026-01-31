// Package domain provides core domain types and logic for art-dupl.
//
// This package defines the domain model for code duplication detection:
// - Value objects: LineNumber, BytePosition, TokenCount, Threshold, etc.
// - Entities: Clone, CloneGroup, Analysis
// - Enums: FileProcessingState, DetectionState, AnalysisMode, CloneSeverity
// - Services: StringInternPool for memory-efficient string storage
//
// All types in this package are designed for type safety and validation.
// Value objects are immutable and validated at construction time.
//
// Key Types:
//
//	// Line numbers with validation (must be > 0)
//	line, err := domain.NewLineNumber(10)
//
//	// File paths with validation (must be non-empty)
//	path, err := domain.NewFilepath("/path/to/file.go")
//
//	// Threshold for clone detection (must be > 0)
//	threshold, err := domain.NewThreshold(15)
//
//	// String interning for memory efficiency
//	pool := domain.GlobalPool()
//	id := pool.Intern("filename.go")
//
// Architecture:
// - domain_types.go: Value object definitions (LineNumber, Threshold, etc.)
// - clone.go: Domain entities (Clone, CloneGroup, Analysis) and enums
// - stringpool.go: String interning for memory optimization
package domain
