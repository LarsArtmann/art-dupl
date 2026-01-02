package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LarsArtmann/art-dupl/pkg/artdupl"
)

func RunSDKDemo() {
	fmt.Println("=== dupl SDK Demo ===")

	// Example 1: Basic SDK Usage
	fmt.Println("1. Basic SDK Usage:")
	basicExample()

	fmt.Println("\n2. Advanced Usage with Progress Callback:")
	progressExample()

	fmt.Println("\n3. Streaming Results for Large Projects:")
	streamingExample()

	fmt.Println("\n4. Custom Configuration:")
	configExample()

	fmt.Println("\n5. Error Handling:")
	errorExample()
}

func basicExample() {
	// Create detector with default options
	detector, err := artdupl.NewDetector(nil)
	if err != nil {
		log.Printf("Failed to create detector: %v", err)
		return
	}
	defer func() { _ = detector.Close() }()

	// Analyze some files
	files := []string{
		"syntax/syntax.go",
		"config/config.go",
	}

	ctx := context.Background()
	result, err := detector.FindClones(ctx, files)
	if err != nil {
		log.Printf("Analysis failed: %v", err)
		return
	}

	fmt.Printf("Found %d clone groups in %v\n", len(result.CloneGroups), result.Summary.AnalysisTime)
	fmt.Printf("Analyzed %d files, found %d total clones\n",
		result.Summary.TotalFiles, result.Summary.TotalClones)

	if len(result.CloneGroups) > 0 {
		group := result.CloneGroups[0]
		fmt.Printf("  First group: %s with %d clones\n", group.Hash[:8]+"...", len(group.Clones))
		fmt.Printf("    Method: %s, Size: %d tokens\n", group.Method, group.Size)
	}
}

func progressExample() {
	// Create detector with progress callback
	opts := artdupl.DefaultOptions()
	opts.Threshold = 10
	opts.ProgressCallback = func(progress *artdupl.Progress) error {
		fmt.Printf("  [%s] %.1f%% - %s\n", progress.Stage, progress.Percentage, progress.Message)
		return nil
	}

	detector, err := artdupl.NewDetector(opts)
	if err != nil {
		log.Printf("Failed to create detector: %v", err)
		return
	}
	defer func() { _ = detector.Close() }()

	ctx := context.Background()
	files := []string{"suffixtree/suffixtree.go", "detection/multidetector.go"}

	result, err := detector.FindClones(ctx, files)
	if err != nil {
		log.Printf("Analysis failed: %v", err)
		return
	}

	fmt.Printf("  Complete! Found %d clone groups\n", len(result.CloneGroups))
}

func streamingExample() {
	// Create detector optimized for streaming large results
	opts := artdupl.DefaultOptions()
	opts.Threshold = 5
	opts.Timeout = 10 * time.Second

	detector, err := artdupl.NewDetector(opts)
	if err != nil {
		log.Printf("Failed to create detector: %v", err)
		return
	}
	defer func() { _ = detector.Close() }()

	ctx := context.Background()
	files := []string{"cli.go", "main.go"}

	cloneChan, err := detector.FindClonesStream(ctx, files)
	if err != nil {
		log.Printf("Stream setup failed: %v", err)
		return
	}

	groupCount := 0
	totalClones := 0

	for group := range cloneChan {
		groupCount++
		totalClones += len(group.Clones)
		fmt.Printf("  Group %s: %d clones\n", group.Hash[:8]+"...", len(group.Clones))
	}

	fmt.Printf("Streaming complete: %d groups, %d total clones\n", groupCount, totalClones)
}

func configExample() {
	// Create detector with custom configuration
	opts := &artdupl.Options{
		Threshold:        15,
		DetectionMethods: []artdupl.DetectionMethod{artdupl.MethodArtDupl},
		IncludeVendor:    false,
		IncludeFragments: true,
		MaxFileSize:      1024 * 1024, // 1MB
		MaxWorkers:       2,
		Timeout:          30 * time.Second,
		Logger:           &verboseLogger{},
	}

	detector, err := artdupl.NewDetector(opts)
	if err != nil {
		log.Printf("Failed to create detector: %v", err)
		return
	}
	defer func() { _ = detector.Close() }()

	ctx := context.Background()
	files := []string{"printer/printer.go"}

	result, err := detector.FindClones(ctx, files)
	if err != nil {
		log.Printf("Analysis failed: %v", err)
		return
	}

	fmt.Printf("Config analysis: %d groups found\n", len(result.CloneGroups))
	if len(result.CloneGroups) > 0 && len(result.CloneGroups[0].Clones) > 0 {
		fmt.Printf("  Sample fragment: %s\n", result.CloneGroups[0].Clones[0].Fragment)
	}
}

func errorExample() {
	// Demonstrate error handling
	testCases := []struct {
		name  string
		opts  *artdupl.Options
		files []string
	}{
		{
			name:  "Invalid threshold",
			opts:  &artdupl.Options{Threshold: 0},
			files: []string{"test.go"},
		},
		{
			name:  "No files",
			opts:  artdupl.DefaultOptions(),
			files: []string{},
		},
		{
			name: "Invalid detection method",
			opts: &artdupl.Options{
				Threshold:        10,
				DetectionMethods: []artdupl.DetectionMethod{"invalid"},
			},
			files: []string{"test.go"},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("  Testing %s: ", tc.name)
		detector, err := artdupl.NewDetector(tc.opts)
		if err != nil {
			fmt.Printf("✅ Expected error: %v\n", err)
			continue
		}
		defer func() { _ = detector.Close() }()

		ctx := context.Background()
		_, err = detector.FindClones(ctx, tc.files)
		if err != nil {
			fmt.Printf("✅ Expected error: %v\n", err)
		} else {
			fmt.Printf("❌ Expected error but got none\n")
		}
	}
}

// verboseLogger implements artdupl.Logger with verbose output
type verboseLogger struct{}

func (l *verboseLogger) Debug(msg string, args ...any) {
	fmt.Printf("[DEBUG] %s\n", fmt.Sprintf(msg, args...))
}

func (l *verboseLogger) Info(msg string, args ...any) {
	fmt.Printf("[INFO] %s\n", fmt.Sprintf(msg, args...))
}

func (l *verboseLogger) Warn(msg string, args ...any) {
	fmt.Printf("[WARN] %s\n", fmt.Sprintf(msg, args...))
}

func (l *verboseLogger) Error(msg string, args ...any) {
	fmt.Printf("[ERROR] %s\n", fmt.Sprintf(msg, args...))
}
