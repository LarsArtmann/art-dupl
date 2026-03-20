package artdupl

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/art-dupl/hash"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// buildAnalysisPipeline processes files and prepares data for analysis.
func (d *detector) buildAnalysisPipeline(
	ctx context.Context,
	files []string,
) ([]*syntax.Node, job.ParseStats, error) {
	d.reportProgress(0, "Starting file processing", "")

	// Create file channel
	fileChan := make(chan string, 100)

	// Feed files to channel in a goroutine
	go func() {
		defer close(fileChan)

		for i, filename := range files {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Validate file exists and isn't too large
			err := d.validateFile(filename)
			if err != nil {
				d.logger.Warn("Skipping file %s: %v", filename, err)

				continue
			}

			fileChan <- filename

			// Report progress
			progress := float64(i+1) / float64(len(files)) * 50 // Files are 50% of work
			d.reportProgress(progress, "Processing files", filename)
		}
	}()

	// Parse files and build syntax tree
	syntaxChan, fileCountChan := job.Parse(ctx, fileChan)
	tree, data, done := job.BuildTree(ctx, syntaxChan)

	// Wait for tree building to complete
	select {
	case <-done:
		// Tree building complete
	case <-ctx.Done():
		return nil, job.ParseStats{}, fmt.Errorf("pipeline canceled after processing %d files: %w", len(files), ctx.Err())
	case <-time.After(d.opts.Timeout):
		return nil, job.ParseStats{}, fmt.Errorf("analysis timed out after %v (processing %d files): %w", d.opts.Timeout, len(files), ErrAnalysisTimeout)
	}

	// Get file count
	fileCount := <-fileCountChan

	// Finalize tree
	tree.Update(&syntax.Node{Type: -1}) //nolint:exhaustruct

	// Convert data slice to pointer
	nodeData := *data

	d.reportProgress(60, "Building suffix tree", "")

	return nodeData, fileCount, nil
}

// runDetection executes the configured detection methods.
func (d *detector) runDetection(ctx context.Context, data []*syntax.Node) ([]*CloneGroup, error) {
	d.reportProgress(70, "Starting duplicate detection", "")

	var allGroups []*CloneGroup

	// Get matches based on detection methods
	threshold := d.config.Threshold

	var matchesChan <-chan syntax.Match

	switch d.opts.DetectionMethods[0] { // Simplified - support single method for now
	case MethodArtDupl:
		// Build suffix tree for art-dupl method
		matchesChan = d.runArtDuplDetection(ctx, data, threshold)
	case MethodHash:
		matchesChan = d.runHashDetection(ctx, data, threshold)
	default:
		return nil, ErrUnsupportedMethod
	}

	// Collect and process matches
	groups, err := collectMatchesIntoGroups(ctx, matchesChan)
	if err != nil {
		return nil, err
	}

	// Convert to CloneGroup format
	for hash, frags := range groups {
		uniq := syntax.Unique(frags)
		if len(uniq) > 1 {
			group := d.convertToCloneGroup(hash, uniq, d.opts.DetectionMethods[0])
			allGroups = append(allGroups, group)
		}
	}

	d.reportProgress(90, "Processing results", "")

	return allGroups, nil
}

// streamDetectionResults streams detection results to the provided channel.
func (d *detector) streamDetectionResults(
	ctx context.Context,
	data []*syntax.Node,
	resultChan chan<- *CloneGroup,
) error {
	// Similar to runDetection but streams results instead of collecting all
	threshold := d.config.Threshold

	var matchesChan <-chan syntax.Match

	switch d.opts.DetectionMethods[0] {
	case MethodArtDupl:
		matchesChan = d.runArtDuplDetection(ctx, data, threshold)
	case MethodHash:
		matchesChan = d.runHashDetection(ctx, data, threshold)
	default:
		return ErrUnsupportedMethod
	}

	groups, err := collectMatchesIntoGroups(ctx, matchesChan)
	if err != nil {
		return err
	}

	// Stream results
	for hash, frags := range groups {
		uniq := syntax.Unique(frags)
		if len(uniq) > 1 {
			group := d.convertToCloneGroup(hash, uniq, d.opts.DetectionMethods[0])

			select {
			case resultChan <- group:
			case <-ctx.Done():
				return ctx.Err() //nolint:wrapcheck // Context cancellation errors are already clear
			}
		}
	}

	return nil
}

// runSuffixTreeDetection executes suffix tree-based detection.
func (d *detector) runSuffixTreeDetection(
	data []*syntax.Node,
	threshold int,
) <-chan syntax.Match {
	tree := d.buildSuffixTree(data)
	suffixMatches := tree.FindDuplOver(threshold)

	syntaxMatches := make(chan syntax.Match)

	go func() {
		defer close(syntaxMatches)

		for match := range suffixMatches {
			syntaxMatch := syntax.FindSyntaxUnits(data, match, threshold)
			if len(syntaxMatch.Frags) > 0 {
				syntaxMatches <- syntaxMatch
			}
		}
	}()

	return syntaxMatches
}

// runArtDuplDetection executes art-dupl (suffix tree) detection method.
func (d *detector) runArtDuplDetection(
	_ context.Context,
	data []*syntax.Node,
	threshold int,
) <-chan syntax.Match {
	return d.runSuffixTreeDetection(data, threshold)
}

// runHashDetection executes hash-based detection method using SHA-256 file hashing.
func (d *detector) runHashDetection(
	_ context.Context,
	data []*syntax.Node,
	threshold int,
) <-chan syntax.Match {
	hashDetector := hash.NewHashDetector(threshold)

	return hashDetector.FindDuplOver(data, threshold)
}

// buildSuffixTree creates a suffix tree from the provided data.
func (d *detector) buildSuffixTree(data []*syntax.Node) *suffixtree.STree {
	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}

	return tree
}

// collectMatchesIntoGroups collects matches from a channel and groups them by hash.
func collectMatchesIntoGroups(
	ctx context.Context,
	matchesChan <-chan syntax.Match,
) (map[string][][]*syntax.Node, error) {
	groups := make(map[string][][]*syntax.Node)

	for match := range matchesChan {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err() //nolint:wrapcheck // Standard context cancellation
		default:
		}

		if len(match.Frags) > 0 {
			groups[match.Hash] = append(groups[match.Hash], match.Frags...)
		}
	}

	return groups, nil
}
