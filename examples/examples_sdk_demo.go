//nolint:forbidigo // Demo examples use fmt.Println for demonstration
package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LarsArtmann/art-dupl/pkg/artdupl"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
)

const demoTestFile = "test.go"

// newDemoDetector creates a detector and logs any creation error.
// Returns nil + false if creation fails; the caller should return early.
// The returned detector must be closed by the caller (typically via defer).
func newDemoDetector(opts *artdupl.Options) (artdupl.Detector, bool) {
	detector, err := artdupl.NewDetector(opts)
	if err != nil {
		log.Printf("Failed to create detector: %v", err)

		return nil, false
	}

	return detector, true
}

// newDemoCtx creates a demo detector paired with a background context.
// It folds the common "create + check ok + context.Background()" prologue
// shared by every example function. The returned detector must be closed
// by the caller (typically via defer).
func newDemoCtx(opts *artdupl.Options) (artdupl.Detector, context.Context, bool) {
	detector, ok := newDemoDetector(opts)
	if !ok {
		return nil, nil, false
	}

	return detector, context.Background(), true
}

// runDemo wraps the "create detector + defer close" boilerplate shared by
// every demo function. It runs fn with a fresh detector + background context
// when setup succeeds, and silently returns when setup fails.
func runDemo(opts *artdupl.Options, fn func(artdupl.Detector, context.Context)) {
	detector, ctx, ok := newDemoCtx(opts)
	if !ok {
		return
	}

	defer func() { _ = detector.Close() }()

	fn(detector, ctx)
}

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
	runDemo(nil, func(detector artdupl.Detector, ctx context.Context) {
		files := []string{
			"syntax/syntax.go",
			"config/config.go",
		}

		result, err := detector.FindClones(ctx, files)
		if err != nil {
			log.Printf("Analysis failed: %v", err)

			return
		}

		fmt.Printf(
			"Found %d clone cloneGroups in %v\n",
			len(result.CloneGroups),
			result.Summary.AnalysisTime,
		)
		fmt.Printf("Analyzed %d files, found %d total clones\n",
			result.Summary.TotalFiles, result.Summary.TotalClones)

		if len(result.CloneGroups) > 0 {
			cloneGroup := result.CloneGroups[0]
			fmt.Printf(
				"  First cloneGroup: %s with %d clones\n",
				cloneGroup.Hash[:8]+"...",
				len(cloneGroup.Clones),
			)
			fmt.Printf("    Method: %s, Size: %d tokens\n", cloneGroup.Method, cloneGroup.Size)
		}
	})
}

func progressExample() {
	opts := artdupl.DefaultOptions()
	//art-dupl:accept self-contained example: each demo must be independently readable
	opts.Threshold = 10
	opts.ProgressCallback = func(progress *artdupl.Progress) error {
		fmt.Printf("  [%s] %.1f%% - %s\n", progress.Stage, progress.Percentage, progress.Message)

		return nil
	}

	runDemo(opts, func(detector artdupl.Detector, ctx context.Context) {
		files := []string{"suffixtree/suffixtree.go", "detection/multidetector.go"}

		result, err := detector.FindClones(ctx, files)
		if err != nil {
			log.Printf("Analysis failed: %v", err)

			return
		}

		fmt.Printf("  Complete! Found %d clone groups\n", len(result.CloneGroups))
	})
}

func streamingExample() {
	opts := artdupl.DefaultOptions()
	opts.Threshold = 5
	opts.Timeout = 10 * time.Second

	runDemo(opts, func(detector artdupl.Detector, ctx context.Context) {
		files := []string{"cli.go", "main.go"}

		cloneChan, err := detector.FindClonesStreamResult(ctx, files)
		if err != nil {
			log.Printf("Stream setup failed: %v", err)

			return
		}

		groupCount := 0
		totalClones := 0

		for result := range cloneChan {
			if result.Err != nil {
				log.Printf("Stream error: %v", result.Err)

				return
			}

			if result.Group == nil {
				continue
			}

			groupCount++
			totalClones += len(result.Group.Clones)

			hashPreview := result.Group.Hash
			if len(hashPreview) > 8 {
				hashPreview = hashPreview[:8]
			}

			fmt.Printf("  Group %s...: %d clones\n", hashPreview, len(result.Group.Clones))
		}

		fmt.Printf("Streaming complete: %d groups, %d total clones\n", groupCount, totalClones)
	})
}

func configExample() {
	// Create detector with custom configuration
	opts := &artdupl.Options{
		Threshold:        artdupl.DefaultThreshold,
		DetectionMethods: []artdupl.DetectionMethod{artdupl.MethodArtDupl},
		IncludeFragments: true,
		MaxFileSize:      1024 * 1024, // 1MB
		MaxWorkers:       2,
		Timeout:          30 * time.Second,
		Logger:           logger.NewLogger(&logger.Config{Level: "debug"}),
	}

	runDemo(opts, func(detector artdupl.Detector, ctx context.Context) {
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
	})
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
			files: []string{demoTestFile},
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
			files: []string{demoTestFile},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("  Testing %s: ", tc.name)

		detector, ok := newDemoDetector(tc.opts)
		if !ok {
			continue
		}

		defer func() { _ = detector.Close() }()

		ctx := context.Background()

		_, err := detector.FindClones(ctx, tc.files)
		if err != nil {
			fmt.Printf("✅ Expected error: %v\n", err)
		} else {
			fmt.Printf("❌ Expected error but got none\n")
		}
	}
}
