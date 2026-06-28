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
//	cfg := detection.Config{Methods: []string{detection.MethodArtDupl}, Verbose: verbose}
//	md := detection.NewMultiDetector(cfg, data, tree)
//	matches := md.FindDuplOver(ctx, threshold)
package detection

import (
	"context"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// MultiDetector runs multiple detection methods and combines results.
type MultiDetector struct {
	cfg  Config
	data []*syntax.Node
	tree *suffixtree.STree
}

// NewMultiDetector creates a new multi-method detector.
func NewMultiDetector(
	cfg Config,
	data []*syntax.Node,
	tree *suffixtree.STree,
) *MultiDetector {
	return &MultiDetector{
		cfg:  cfg,
		data: data,
		tree: tree,
	}
}

// FindDuplOver runs all configured clone detection methods and streams matches.
// The ctx is checked on every channel send — if cancelled, the goroutine
// exits early to prevent goroutine leaks.
func (md *MultiDetector) FindDuplOver(
	ctx context.Context,
	threshold int,
) <-chan syntax.Match {
	detectors := md.buildCloneDetectors()

	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		for _, det := range detectors {
			md.logVerbose("Running " + det.Name() + "...")

			md.streamMatches(ctx, det.FindDuplOver(ctx, threshold), resultChan)
		}
	}()

	return resultChan
}

// hasNonEmptyFrag returns true if frags contains at least one non-empty
// node sequence. This prevents matches with empty inner slices (e.g.
// [][]*Node{{}}) from passing the guard.
func hasNonEmptyFrag(frags [][]*syntax.Node) bool {
	for _, frag := range frags {
		if len(frag) > 0 {
			return true
		}
	}

	return false
}

// logVerbose prints verbose output if enabled.
func (md *MultiDetector) logVerbose(message string) {
	if md.cfg.Verbose {
		logger.Default.Info(message)
	}
}

// streamMatches forwards matches from src to dst, checking ctx on every send.
// Filters out matches with no non-empty fragments.
func (md *MultiDetector) streamMatches(
	ctx context.Context,
	src <-chan syntax.Match,
	dst chan<- syntax.Match,
) {
	for match := range src {
		if ctx.Err() != nil {
			return
		}

		if hasNonEmptyFrag(match.Frags) {
			select {
			case dst <- match:
			case <-ctx.Done():
				return
			}
		}
	}
}
