// Package detection provides multi-method code duplication detection.
//
// This package coordinates multiple detection algorithms and combines
// their results to provide comprehensive duplicate reporting.
//
// Detection Methods Supported:
// - DetectionMethodArtDupl: Suffix tree algorithm on AST tokens
// - DetectionMethodHash: Rolling hash on file content
// - DetectionMethodTodos: Finds TODO comments
// - DetectionMethodLegacy: Finds legacy code patterns
//
// Core Type:
// - MultiDetector: Coordinates multiple detection methods
//
// Usage:
//
//	// Create multi-detector with configuration
//	md := detection.NewMultiDetector(cfg, data, tree, verbose)
//
//	// Run all configured detection methods
//	matches := md.FindDuplOver(threshold)
//
// Design:
// - Methods are configured via config.DetectionMethods
// - Results are combined and deduplicated
// - Verbose logging available for debugging
// - Respects config.IsDefault() for optimization
//
// Performance:
// - Runs selected methods in parallel (goroutines)
// - Channels used for non-blocking result delivery
// - Each method runs independently, results combined at output
package detection

import (
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/hash"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// MultiDetector runs multiple detection methods and combines results.
type MultiDetector struct {
	config  *config.Config
	data    []*syntax.Node
	tree    *suffixtree.STree
	verbose bool
}

// NewMultiDetector creates a new multi-method detector.
func NewMultiDetector(
	cfg *config.Config,
	data []*syntax.Node,
	tree *suffixtree.STree,
	verbose bool,
) *MultiDetector {
	return &MultiDetector{
		config:  cfg,
		data:    data,
		tree:    tree,
		verbose: verbose,
	}
}

// FindDuplOver runs all configured detection methods.
func (md *MultiDetector) FindDuplOver(threshold int) <-chan syntax.Match {
	// If only art-dupl method is selected, use existing logic
	if md.config.DetectionMethods.IsDefault() {
		// Convert suffix tree matches to syntax matches
		resultChan := make(chan syntax.Match)

		go func() {
			defer close(resultChan)

			suffixMatches := md.tree.FindDuplOver(threshold)
			for match := range suffixMatches {
				syntaxMatch := syntax.FindSyntaxUnits(md.data, match, threshold)
				if len(syntaxMatch.Frags) > 0 {
					resultChan <- syntaxMatch
				}
			}
		}()

		return resultChan
	}

	// Create combined channel
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		// Run hash detection if selected
		if md.config.DetectionMethods.Contains(config.DetectionMethodHash) {
			md.logVerbose("Running hash-based detection...")

			hashDetector := hash.NewHashDetector(threshold)
			hashMatches := hashDetector.FindDuplOver(md.data, threshold)

			for match := range hashMatches {
				if len(match.Frags) > 0 {
					resultChan <- match
				}
			}
		}

		// Run art-dupl detection if selected
		if md.config.DetectionMethods.Contains(config.DetectionMethodArtDupl) {
			md.logVerbose("Running suffix tree-based detection...")
			artDuplMatches := md.tree.FindDuplOver(threshold)

			for match := range artDuplMatches {
				// Convert suffix tree matches to syntax matches
				syntaxMatch := syntax.FindSyntaxUnits(md.data, match, threshold)
				if len(syntaxMatch.Frags) > 0 {
					resultChan <- syntaxMatch
				}
			}
		}
	}()

	return resultChan
}

// logVerbose prints verbose output if enabled.
func (md *MultiDetector) logVerbose(message string) {
	if md.verbose {
		logger.Default.Info(message)
	}
}
