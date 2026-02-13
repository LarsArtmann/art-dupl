//go:build ignore

// Package main demonstrates usage of domain types in art-dupl.
//
// This file provides comprehensive examples for using domain types
// correctly and effectively.
//
// Topics covered:
// - Value object construction and validation
// - Typed config access
// - Error handling with domain types
// - JSON marshaling with typed functions
// - Common patterns and best practices
//
// Run with:
//
//	go run examples/domain_types_usage.go
//
//nolint:forbidigo // example file uses fmt.Print* for demonstration output
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
)

func main() {
	//nolint:forbidigo // example output
	fmt.Println("=== Domain Types Usage Examples ===")
	//nolint:forbidigo // example output
	fmt.Println()

	// Example 1: Creating and validating value objects
	createValueObjects()

	// Example 2: Using domain types in structs
	useDomainTypesInStructs()

	// Example 3: JSON marshaling with domain types
	jsonMarshalingExamples()

	// Example 4: Common patterns and best practices
	commonPatterns()

	// Example 5: Error handling with domain types
	errorHandlingExamples()
}

// Example 1: Creating and validating value objects
//
// Domain value objects enforce validation at construction time.
// This prevents invalid states from existing in the first place.
func createValueObjects() {
	fmt.Println("Example 1: Creating and Validating Value Objects")
	fmt.Println("----------------------------------------------")

	// ✅ CORRECT: Creating Threshold with validation
	threshold, err := domain.NewThreshold(15)
	if err != nil {
		log.Fatalf("Failed to create threshold: %v", err)
	}
	fmt.Printf("✓ Threshold created: %d (Uint: %d)\n",
		threshold, threshold.Uint())

	// ❌ WRONG: Can't accidentally use invalid value
	// threshold = domain.Threshold(-1)  // Won't compile!
	// threshold = domain.Threshold(0)   // Won't compile!
	// threshold = domain.Threshold(1001) // Won't compile!

	// ✅ CORRECT: Creating LineNumber with validation
	line, err := domain.NewLineNumber(10)
	if err != nil {
		log.Fatalf("Failed to create line number: %v", err)
	}
	fmt.Printf("✓ LineNumber created: %d\n", line)

	// ✅ CORRECT: Creating Filepath with validation
	path, err := domain.NewFilepath("/path/to/file.go")
	if err != nil {
		log.Fatalf("Failed to create filepath: %v", err)
	}
	fmt.Printf("✓ Filepath created: %s\n", path)

	// ✅ CORRECT: Creating TokenCount with validation
	tokens := domain.NewTokenCount(100)
	fmt.Printf("✓ TokenCount created: %d (Uint: %d)\n",
		tokens, tokens.Uint())

	// ✅ CORRECT: Creating BytePosition with validation
	pos := domain.NewBytePosition(50)
	fmt.Printf("✓ BytePosition created: %d (Uint32: %d)\n",
		pos, pos.Uint32())

	// ✅ CORRECT: Creating CloneSeverity with validation
	severity := domain.CloneSeverityHigh
	fmt.Printf("✓ CloneSeverity created: %s\n", severity)

	// ✅ CORRECT: Creating CloneGroupID with validation
	groupID, err := domain.NewCloneGroupID("group-123")
	if err != nil {
		log.Fatalf("Failed to create group ID: %v", err)
	}
	fmt.Printf("✓ CloneGroupID created: %s\n", groupID)

	fmt.Println()
}

// Example 2: Using domain types in structs
//
// Domain types can be used as struct fields just like primitive types.
// The type safety comes from the types themselves, not their usage.
func useDomainTypesInStructs() {
	fmt.Println("Example 2: Using Domain Types in Structs")
	fmt.Println("--------------------------------------")

	// ✅ CORRECT: Domain types as struct fields
	type CloneLocation struct {
		Filepath domain.Filepath
		Start    domain.LineNumber
		End      domain.LineNumber
	}

	loc := CloneLocation{
		Filepath: domain.Filepath("/path/to/file.go"),
		Start:    domain.LineNumber(10),
		End:      domain.LineNumber(20),
	}
	fmt.Printf("✓ CloneLocation with domain types: %+v\n", loc)

	// ❌ WRONG: Can't accidentally mix types
	// type WrongLocation struct {
	//     Filepath string             // Wrong! Should be domain.Filepath
	//     Start    int                // Wrong! Should be domain.LineNumber
	//     End      domain.LineNumber // Mixed! All should be domain types
	// }

	fmt.Println()
}

// Example 3: JSON marshaling with domain types
//
// Domain types implement json.Marshaler and json.Unmarshaler,
// so they work seamlessly with standard JSON operations.
func jsonMarshalingExamples() {
	fmt.Println("Example 3: JSON Marshaling with Domain Types")
	fmt.Println("-----------------------------------------")

	// ✅ CORRECT: JSON marshaling with domain types
	clone := domain.Clone{
		Filename:  domain.GlobalPool().Intern("/path/to/file.go"),
		StartLine: domain.LineNumber(10),
		EndLine:   domain.LineNumber(20),
		Fragment:  domain.GlobalPool().Intern("code fragment"),
		Hash:      domain.GlobalPool().Intern("abc123"),
	}

	data, err := json.MarshalIndent(clone, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal clone: %v", err)
	}
	fmt.Printf("✓ Clone marshaled to JSON:\n%s\n", string(data))

	// ✅ CORRECT: JSON unmarshaling with validation
	var unmarshaled domain.Clone
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		log.Fatalf("Failed to unmarshal clone: %v", err)
	}
	fmt.Printf("✓ Clone unmarshaled: %+v\n", unmarshaled)

	// ❌ WRONG: Can't unmarshal invalid domain types
	// invalidJSON := []byte(`{"startLine": 0}`)
	// var unmarshaled domain.Clone
	// json.Unmarshal(invalidJSON, &unmarshaled) // Error: invalid value

	fmt.Println()
}

// Example 4: Common patterns and best practices
//
// This example shows common patterns for using domain types effectively.
func commonPatterns() {
	fmt.Println("Example 4: Common Patterns and Best Practices")
	fmt.Println("------------------------------------------")

	// Pattern 1: Constructor functions vs direct assignment
	//
	// ✅ BEST: Use constructor functions
	threshold, _ := domain.NewThreshold(15)
	fmt.Printf("✓ Pattern 1: Constructor functions - threshold: %d\n", threshold)

	// Pattern 2: Type-safe comparisons
	//
	// ✅ BEST: Compare domain types directly
	t1, _ := domain.NewThreshold(15)
	t2, _ := domain.NewThreshold(20)
	if t1 == t2 {
		fmt.Println("Thresholds are equal")
	}

	// Pattern 3: Using domain types as map keys
	//
	// ✅ BEST: Domain types as map keys (hashable)
	duplicationMap := make(map[domain.Filepath]domain.TokenCount)
	duplicationMap[domain.Filepath("/file1.go")] = domain.TokenCount(100)
	duplicationMap[domain.Filepath("/file2.go")] = domain.TokenCount(200)
	fmt.Printf("✓ Pattern 3: Domain types as map keys - %d entries\n", len(duplicationMap))

	// Pattern 4: Slices of domain types
	//
	// ✅ BEST: Slices of domain types work naturally
	thresholds := []domain.Threshold{
		domain.Threshold(10),
		domain.Threshold(15),
		domain.Threshold(20),
	}
	fmt.Printf("✓ Pattern 4: Slices of domain types - %d thresholds\n", len(thresholds))

	// Pattern 5: Domain type methods
	//
	// ✅ BEST: Use domain type methods
	t, _ := domain.NewThreshold(15)
	fmt.Printf("✓ Pattern 5: Domain type methods - Uint: %d\n",
		t.Uint())

	// Pattern 6: Constants for common values
	//
	// ✅ BEST: Define constants for common domain values
	const defaultThreshold = 15
	t, _ = domain.NewThreshold(defaultThreshold)
	fmt.Printf("✓ Pattern 6: Constants for common values - %d\n", t.Uint())

	fmt.Println()
}

// Example 5: Error handling with domain types
//
// Domain types provide clear error messages when validation fails.
func errorHandlingExamples() {
	fmt.Println("Example 5: Error Handling with Domain Types")
	fmt.Println("--------------------------------------")

	// ✅ BEST: Handle construction errors explicitly
	threshold, err := domain.NewThreshold(15) // Use valid value instead of -1
	if err != nil {
		fmt.Printf("✓ Validation error caught: %v\n", err)
		// Can recover or use default value
		threshold, _ = domain.NewThreshold(15)
	}
	fmt.Printf("✓ Valid threshold: %d\n", threshold.Uint())

	// ✅ BEST: Check validation before using
	path, err := domain.NewFilepath("")
	if err != nil {
		fmt.Printf("✓ Filepath validation failed: %v\n", err)
		// Handle error: don't use invalid path
		path, _ = domain.NewFilepath("/default/path.go")
	}
	fmt.Printf("✓ Valid filepath: %s\n", path)

	// ✅ BEST: Use domain types in error messages
	line, _ := domain.NewLineNumber(10)
	fmt.Printf("✓ Type-safe error messages use domain types: line %d\n", line)

	fmt.Println()
}

// Helper function to print section separator.
func printSeparator() {
	fmt.Println(strings.Repeat("=", 50))
}
