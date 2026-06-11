// Package artdupl provides a Go SDK for detecting code clones via suffix tree
// and hash-based detection on Go ASTs. It supports both .go and .templ files.
//
// # Quick Start
//
// Create a detector with default options and analyze files:
//
//	detector, err := artdupl.NewDetector(artdupl.DefaultOptions())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer detector.Close()
//
//	result, err := detector.FindClones(ctx, []string{"service.go", "handler.go"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, group := range result.CloneGroups {
//	    fmt.Printf("Clone: %d tokens, %d occurrences\n", group.Size, len(group.Clones))
//	    for _, clone := range group.Clones {
//	        fmt.Printf("  %s:%d-%d\n", clone.Filename, clone.StartLine, clone.EndLine)
//	    }
//	}
//
// # Configuration
//
// Options control detection behavior:
//
//	opts := &artdupl.Options{
//	    Threshold:         15,                       // Minimum token count for a clone
//	    DetectionMethods:  []artdupl.DetectionMethod{artdupl.MethodArtDupl},
//	    IncludeVendor:     false,                    // Skip vendor/
//	    IncludeFragments:  false,                    // Don't include source code in results
//	    MaxWorkers:        4,                        // Concurrency level
//	    Timeout:           30 * time.Minute,         // Hard timeout
//	    ProgressCallback:  func(p *artdupl.Progress) error { return nil },
//	}
//
// # Streaming API
//
// For large codebases, use FindClonesStream for incremental results:
//
//	ch, err := detector.FindClonesStream(ctx, files)
//	for group := range ch {
//	    processGroup(group)
//	}
//
// # Detection Methods
//
// Two detection methods are available:
//   - MethodArtDupl: suffix tree algorithm on AST tokens (default, most accurate)
//   - MethodHash: rolling hash on file content (faster, less precise)
//
// # Error Handling
//
// All errors are sentinels comparable via errors.Is():
//
//	result, err := detector.FindClones(ctx, files)
//	if errors.Is(err, artdupl.ErrNoFilesProvided) {
//	    // handle empty input
//	}
package artdupl
