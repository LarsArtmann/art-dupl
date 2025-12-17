package examples

import (
	"fmt"
	"log"

	"github.com/LarsArtmann/art-dupl/lib"
)

func RunTestAPI() {
	// Test the current lib API
	files := []string{
		"syntax/syntax.go", // Analyze some of our own files
		"config/config.go", // Different file
	}

	issues, err := lib.Run(files, 10) // Low threshold to find duplicates
	if err != nil {
		log.Fatalf("Error running analysis: %v", err)
	}

	fmt.Printf("Found %d duplicate issues:\n\n", len(issues))

	for i, issue := range issues {
		fmt.Printf("Issue %d:\n", i+1)
		fmt.Printf("  From: %s:%d-%d\n", issue.From.Filename(), issue.From.LineStart(), issue.From.LineEnd())
		fmt.Printf("  To:   %s:%d-%d\n", issue.To.Filename(), issue.To.LineStart(), issue.To.LineEnd())
		fmt.Println()
	}

	// Also test the Issue structure for API usability
	fmt.Printf("API Analysis:\n")
	fmt.Printf("- Issue type: %T\n", issues)
	if len(issues) > 0 {
		fmt.Printf("- First issue From type: %T\n", issues[0].From)
		fmt.Printf("- Has Filename method: %t\n", issues[0].From.Filename() != "")
	}
}
