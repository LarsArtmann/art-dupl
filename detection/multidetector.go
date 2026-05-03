// Package detection provides multi-method code duplication detection.
//
// This package coordinates multiple detection algorithms and combines
// their results to provide comprehensive duplicate reporting.
//
// Detection Methods Supported:
// - DetectionMethodArtDupl: Suffix tree algorithm on AST tokens
// - DetectionMethodHash: Rolling hash on file content
//
// Core Types:
// - MultiDetector: Coordinates multiple detection methods
// - MethodDetector: Interface for individual detection algorithms
//
// Usage:
//
//	cfg := config.DetectionConfig{Methods: methods, Verbose: verbose}
//	md := detection.NewMultiDetector(cfg, data, tree)
//	matches := md.FindDuplOver(threshold)
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
	detCfg config.DetectionConfig
	data   []*syntax.Node
	tree   *suffixtree.STree
}

// NewMultiDetector creates a new multi-method detector.
func NewMultiDetector(
	cfg config.DetectionConfig,
	data []*syntax.Node,
	tree *suffixtree.STree,
) *MultiDetector {
	return &MultiDetector{
		detCfg: cfg,
		data:   data,
		tree:   tree,
	}
}

// FindDuplOver runs all configured detection methods.
func (md *MultiDetector) FindDuplOver(threshold int) <-chan syntax.Match {
	if md.detCfg.Methods.IsDefault() {
		resultChan := make(chan syntax.Match)

		go func() {
			defer close(resultChan)

			suffixMatches := md.tree.FindDuplOver(threshold)
			md.processSuffixTreeMatches(suffixMatches, resultChan, threshold)
		}()

		return resultChan
	}

	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		if md.detCfg.Methods.Contains(config.DetectionMethodHash) {
			md.logVerbose("Running hash-based detection...")

			hashDetector := hash.NewFileDetector(threshold)
			hashMatches := hashDetector.FindDuplOver(md.data, threshold)

			for match := range hashMatches {
				if len(match.Frags) > 0 {
					resultChan <- match
				}
			}
		}

		if md.detCfg.Methods.Contains(config.DetectionMethodArtDupl) {
			md.logVerbose("Running suffix tree-based detection...")
			artDuplMatches := md.tree.FindDuplOver(threshold)

			md.processSuffixTreeMatches(artDuplMatches, resultChan, threshold)
		}
	}()

	return resultChan
}

// logVerbose prints verbose output if enabled.
func (md *MultiDetector) logVerbose(message string) {
	if md.detCfg.Verbose {
		logger.Default.Info(message)
	}
}

// processSuffixTreeMatches converts suffix tree matches to syntax matches and sends them to the result channel.
func (md *MultiDetector) processSuffixTreeMatches(
	matches <-chan suffixtree.Match,
	resultChan chan<- syntax.Match,
	threshold int,
) {
	for match := range matches {
		syntaxMatch := syntax.FindSyntaxUnits(md.data, match, threshold)
		if len(syntaxMatch.Frags) > 0 {
			resultChan <- syntaxMatch
		}
	}
}
