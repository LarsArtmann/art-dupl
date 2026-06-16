// Package detection provides multi-method code duplication detection.
//
// This package coordinates multiple detection algorithms and combines
// their results to provide comprehensive duplicate reporting.
//
// Detection Methods Supported:
// - DetectionMethodArtDupl: Suffix tree algorithm on AST tokens
// - DetectionMethodHash: Rolling hash on file content
// - DetectionMethodTodos: Find TODO/FIXME/HACK comments
// - DetectionMethodLegacy: Find deprecated functions and legacy patterns
//
// Core Types:
// - MultiDetector: Coordinates multiple detection methods
// - MethodDetector: Interface for individual detection algorithms
//
// Usage:
//
//	cfg := config.DetectionConfig{Methods: methods, Verbose: verbose}
//	md := detection.NewMultiDetector(cfg, data, tree)
//	matches := md.FindDuplOver(ctx, threshold)
package detection

import (
	"context"

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

// FindDuplOver runs all configured detection methods and streams matches.
// The ctx is checked on every channel send — if cancelled, the goroutine
// exits early to prevent goroutine leaks.
func (md *MultiDetector) FindDuplOver(
	ctx context.Context,
	threshold int,
) <-chan syntax.Match {
	if md.detCfg.Methods.IsDefault() {
		resultChan := make(chan syntax.Match)

		go func() {
			defer close(resultChan)

			suffixMatches := md.tree.FindDuplOver(threshold)
			md.processSuffixTreeMatches(ctx, suffixMatches, resultChan, threshold)
		}()

		return resultChan
	}

	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		md.runMultiMethodDetection(ctx, resultChan, threshold)
	}()

	return resultChan
}

// runMultiMethodDetection runs each configured detection method sequentially,
// streaming matches to resultChan. Respects ctx cancellation.
func (md *MultiDetector) runMultiMethodDetection(
	ctx context.Context,
	resultChan chan<- syntax.Match,
	threshold int,
) {
	if md.detCfg.Methods.Contains(config.DetectionMethodHash) {
		md.logVerbose("Running hash-based detection...")

		hashDetector := hash.NewFileDetector(threshold)
		hashMatches := hashDetector.FindDuplOver(md.data, threshold)
		md.streamMatches(ctx, hashMatches, resultChan)
	}

	if md.detCfg.Methods.Contains(config.DetectionMethodArtDupl) {
		md.logVerbose("Running suffix tree-based detection...")

		artDuplMatches := md.tree.FindDuplOver(threshold)
		md.processSuffixTreeMatches(ctx, artDuplMatches, resultChan, threshold)
	}

	if md.detCfg.Methods.Contains(config.DetectionMethodTodos) {
		md.logVerbose("Running TODO detection...")

		todoMatches := NewTodoDetector().FindTodos(ctx, md.data)
		md.streamMatches(ctx, todoMatches, resultChan)
	}

	if md.detCfg.Methods.Contains(config.DetectionMethodLegacy) {
		md.logVerbose("Running legacy pattern detection...")

		legacyMatches := NewLegacyDetector().FindLegacy(ctx, md.data)
		md.streamMatches(ctx, legacyMatches, resultChan)
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
	if md.detCfg.Verbose {
		logger.Default.Info(message)
	}
}

// processSuffixTreeMatches converts suffix tree matches to syntax matches
// and sends them to the result channel. Respects ctx cancellation.
func (md *MultiDetector) processSuffixTreeMatches(
	ctx context.Context,
	matches <-chan suffixtree.Match,
	resultChan chan<- syntax.Match,
	threshold int,
) {
	for match := range matches {
		if ctx.Err() != nil {
			return
		}

		syntaxMatch := syntax.FindSyntaxUnits(md.data, match, threshold)
		if hasNonEmptyFrag(syntaxMatch.Frags) {
			select {
			case resultChan <- syntaxMatch:
			case <-ctx.Done():
				return
			}
		}
	}
}
