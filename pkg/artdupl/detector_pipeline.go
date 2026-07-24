package artdupl

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type pipelineResult struct {
	data      []*syntax.Node
	tree      *suffixtree.STree
	fileCount job.ParseStats
}

// buildAnalysisPipeline processes files and prepares data for analysis.
func (d *detector) buildAnalysisPipeline(
	ctx context.Context,
	files []string,
) (*pipelineResult, error) {
	d.reportProgress(0, "Starting file processing", "")

	// Create file channel
	fileChan := make(chan string, 100)

	// Feed files to channel in a goroutine
	go func() {
		defer close(fileChan)

		for i, filename := range files {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := d.validateFile(filename)
			if err != nil {
				d.logger.Warn("Skipping file", "file", filename, "err", err)

				continue
			}

			select {
			case fileChan <- filename:
			case <-ctx.Done():
				return
			}

			progress := float64(i+1) / float64(len(files)) * 50
			d.reportProgress(progress, "Processing files", filename)
		}
	}()

	syntaxChan, fileCountChan := job.Parse(ctx, fileChan, d.cfg.toDetectionMode(), d.cfg.MaxChildrenSerial, nil)
	tree, data, done := job.BuildTree(ctx, syntaxChan)

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("suffix tree build failed: %w", err)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf(
			"pipeline canceled after processing %d files: %w",
			len(files),
			ctx.Err(),
		)
	case <-time.After(d.opts.Timeout):
		return nil, fmt.Errorf(
			"analysis timed out after %v (processing %d files): %w",
			d.opts.Timeout,
			len(files),
			ErrAnalysisTimeout,
		)
	}

	fileCount := <-fileCountChan

	if err := tree.Update(
		&syntax.Node{Type: -1},
	); err != nil {
		return nil, fmt.Errorf("suffix tree terminator update failed: %w", err)
	}

	d.reportProgress(60, "Building suffix tree", "")

	return &pipelineResult{
		data:      *data,
		tree:      tree,
		fileCount: fileCount,
	}, nil
}

// processCloneGroups iterates over clone groups and processes each one.
// The processFn is called for each valid group. Groups are sorted by hash
// before processing to ensure deterministic output regardless of map iteration order.
func (d *detector) processCloneGroups(
	groups map[string][][]*syntax.Node,
	processFn func(*CloneGroup),
) {
	hashes := make([]string, 0, len(groups))
	for hash := range groups {
		hashes = append(hashes, hash)
	}

	sort.Strings(hashes)

	for _, hash := range hashes {
		frags := groups[hash]

		uniq := syntax.Unique(frags)
		if len(uniq) > 1 {
			group := d.convertToCloneGroup(hash, uniq, d.opts.DetectionMethods[0])
			processFn(group)
		}
	}
}

// createMultiDetector creates a MultiDetector with the detector's configuration.
func (d *detector) createMultiDetector(
	data []*syntax.Node,
	tree *suffixtree.STree,
) *detection.MultiDetector {
	return detection.NewMultiDetector(detection.Config{
		Methods: d.cfg.DetectionMethods,
		Verbose: d.opts.Verbose,
	}, data, tree)
}

// runDetection executes the configured detection methods using MultiDetector.
func (d *detector) runDetection(
	ctx context.Context,
	result *pipelineResult,
) ([]*CloneGroup, error) {
	d.reportProgress(70, "Starting duplicate detection", "")

	md := d.createMultiDetector(result.data, result.tree)
	matchesChan := md.FindDuplOver(ctx, d.cfg.Threshold)

	groups, err := collectMatchesIntoGroups(ctx, matchesChan)
	if err != nil {
		return nil, err
	}

	var allGroups []*CloneGroup

	d.processCloneGroups(groups, func(group *CloneGroup) {
		allGroups = append(allGroups, group)
	})

	d.reportProgress(90, "Processing results", "")

	return allGroups, nil
}

// streamDetectionResults streams detection results to the provided channel.
func (d *detector) streamDetectionResults(
	ctx context.Context,
	result *pipelineResult,
	resultChan chan<- *CloneGroup,
) error {
	md := d.createMultiDetector(result.data, result.tree)
	matchesChan := md.FindDuplOver(ctx, d.cfg.Threshold)

	groups, err := collectMatchesIntoGroups(ctx, matchesChan)
	if err != nil {
		return err
	}

	d.processCloneGroups(groups, func(group *CloneGroup) {
		select {
		case resultChan <- group:
		case <-ctx.Done():
		}
	})

	return nil
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
			return nil, ctx.Err()
		default:
		}

		for _, frag := range match.Frags {
			if len(frag) > 0 {
				groups[match.Hash] = append(groups[match.Hash], frag)
			}
		}
	}

	return groups, nil
}
