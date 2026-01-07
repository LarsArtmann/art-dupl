package artdupl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/position"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/util"
)

// detector implements the Detector interface using existing dupl components.
type detector struct {
	opts    *Options
	config  *config.Config
	logger  Logger
	started time.Time
}

// NewDetector creates a new code duplication detector.
func NewDetector(opts *Options) (Detector, error) { //nolint:ireturn // Detector interface is the correct return type for factory pattern
	// Use default options if none provided
	if opts == nil {
		opts = DefaultOptions()
	}

	// Validate options
	if err := ValidateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Set default file reader if not provided
	if opts.FileReader == nil {
		opts.FileReader = os.ReadFile
	}

	// Set default logger if not provided
	if opts.Logger == nil {
		opts.Logger = &defaultLogger{}
	}

	// Convert SDK options to internal config
	cfg := convertOptionsToConfig(opts)

	return &detector{
		opts:   opts,
		config: cfg,
		logger: opts.Logger,
	}, nil
}

// FindClones performs complete duplication analysis.
func (d *detector) FindClones(ctx context.Context, files []string) (*Result, error) {
	d.started = time.Now()

	// Validate inputs
	if err := d.validateInputs(ctx, files); err != nil {
		return nil, fmt.Errorf("input validation failed for %d files: %w", len(files), err)
	}

	// Process files and build analysis pipeline
	data, _, err := d.buildAnalysisPipeline(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("analysis pipeline construction failed for %d files: %w", len(files), err)
	}

	// Run detection based on configured methods
	cloneGroups, err := d.runDetection(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("detection failed: %w", err)
	}

	// Build and return result
	return d.buildResult(cloneGroups, 0), nil
}

// FindClonesStream provides streaming results for large projects.
func (d *detector) FindClonesStream(ctx context.Context, files []string) (<-chan *CloneGroup, error) {
	d.started = time.Now()

	// Validate inputs
	if err := d.validateInputs(ctx, files); err != nil {
		return nil, fmt.Errorf("input validation failed for streaming with %d files: %w", len(files), err)
	}

	// Create output channel
	resultChan := make(chan *CloneGroup, 10)

	// Start streaming analysis
	go func() {
		defer close(resultChan)

		// Process files and build analysis pipeline
		data, _, err := d.buildAnalysisPipeline(ctx, files)
		if err != nil {
			d.logger.Error("Analysis pipeline error: %v", err)
			return
		}

		// Run detection with streaming
		if err := d.streamDetectionResults(ctx, data, resultChan); err != nil {
			d.logger.Error("Streaming detection error: %v", err)
		}

		d.reportProgress(100, "Analysis complete", "")
	}()

	return resultChan, nil
}

// Close releases any resources held by the detector.
func (d *detector) Close() error {
	// No persistent resources to clean up currently
	return nil
}

// validateInputs checks that inputs are valid for analysis.
func (d *detector) validateInputs(ctx context.Context, files []string) error {
	if len(files) == 0 {
		return ErrNoFilesProvided
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck // Context cancellation errors are already clear
	default:
	}

	return nil
}

// buildAnalysisPipeline processes files and prepares data for analysis.
func (d *detector) buildAnalysisPipeline(ctx context.Context, files []string) ([]*syntax.Node, int, error) {
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
			if err := d.validateFile(filename); err != nil {
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
	syntaxChan, fileCountChan := job.Parse(fileChan)
	tree, data, done := job.BuildTree(syntaxChan)

	// Wait for tree building to complete
	select {
	case <-done:
		// Tree building complete
	case <-ctx.Done():
		return nil, 0, ctx.Err() //nolint:wrapcheck // Context cancellation errors are already clear
	case <-time.After(d.opts.Timeout):
		return nil, 0, ErrAnalysisTimeout
	}

	// Get file count
	fileCount := <-fileCountChan

	// Finalize tree
	tree.Update(&syntax.Node{Type: -1})

	// Convert data slice to pointer
	nodeData := *data

	d.reportProgress(60, "Building suffix tree", "")

	return nodeData, fileCount, nil
}

// runDetection executes the configured detection methods.
func (d *detector) runDetection(ctx context.Context, data []*syntax.Node) ([]*CloneGroup, error) { //nolint:cyclop // Detection execution with multiple method paths
	d.reportProgress(70, "Starting duplicate detection", "")

	var allGroups []*CloneGroup

	// Note: Multi-detector would be used for multiple methods, but we handle single method for now
	_ = detection.NewMultiDetector(d.config, data, nil, false) // tree not needed for all methods

	// Get matches based on detection methods
	threshold := d.config.Threshold
	var matchesChan <-chan syntax.Match

	switch d.opts.DetectionMethods[0] { // Simplified - support single method for now
	case MethodArtDupl:
		// Build suffix tree for art-dupl method
		matchesChan = d.runArtDuplDetection(ctx, data, threshold)
	case MethodHash:
		matchesChan = d.runHashDetection(ctx, data, threshold)
	case MethodAll:
		// For now, use art-dupl method when MethodAll is specified
		// TODO: Implement multi-detection method support
		matchesChan = d.runArtDuplDetection(ctx, data, threshold)
	default:
		return nil, ErrUnsupportedMethod
	}

	// Collect and process matches
	groups := make(map[string][][]*syntax.Node)

	for match := range matchesChan {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err() //nolint:wrapcheck // Context cancellation errors are already clear
		default:
		}

		if len(match.Frags) > 0 {
			groups[match.Hash] = append(groups[match.Hash], match.Frags...)
		}
	}

	// Convert to CloneGroup format
	for hash, frags := range groups {
		uniq := util.Unique(frags)
		if len(uniq) > 1 {
			group := d.convertToCloneGroup(hash, uniq, d.opts.DetectionMethods[0])
			allGroups = append(allGroups, group)
		}
	}

	d.reportProgress(90, "Processing results", "")

	return allGroups, nil
}

// streamDetectionResults streams detection results to the provided channel.
func (d *detector) streamDetectionResults(ctx context.Context, data []*syntax.Node, resultChan chan<- *CloneGroup) error { //nolint:cyclop // Result streaming with multiple method paths
	// Similar to runDetection but streams results instead of collecting all
	threshold := d.config.Threshold
	var matchesChan <-chan syntax.Match

	switch d.opts.DetectionMethods[0] {
	case MethodArtDupl:
		matchesChan = d.runArtDuplDetection(ctx, data, threshold)
	case MethodHash:
		matchesChan = d.runHashDetection(ctx, data, threshold)
	case MethodAll:
		// For now, use art-dupl method when MethodAll is specified
		// TODO: Implement multi-detection method support
		matchesChan = d.runArtDuplDetection(ctx, data, threshold)
	default:
		return ErrUnsupportedMethod
	}

	groups := make(map[string][][]*syntax.Node)

	for match := range matchesChan {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err() //nolint:wrapcheck // Context cancellation errors are already clear
		default:
		}

		if len(match.Frags) > 0 {
			groups[match.Hash] = append(groups[match.Hash], match.Frags...)
		}
	}

	// Stream results
	for hash, frags := range groups {
		uniq := util.Unique(frags)
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

// runArtDuplDetection executes art-dupl (suffix tree) detection method.
func (d *detector) runArtDuplDetection(ctx context.Context, data []*syntax.Node, threshold int) <-chan syntax.Match {
	// Build suffix tree
	tree := d.buildSuffixTree(data)
	suffixMatches := tree.FindDuplOver(threshold)

	// Convert suffix tree matches to syntax matches
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

// runHashDetection executes hash-based detection method.
func (d *detector) runHashDetection(ctx context.Context, data []*syntax.Node, threshold int) <-chan syntax.Match {
	// Build a suffix tree for the hash detection method as well
	tree := d.buildSuffixTree(data)
	suffixMatches := tree.FindDuplOver(threshold)

	// Convert suffix tree matches to syntax matches
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

// buildSuffixTree creates a suffix tree from the provided data.
func (d *detector) buildSuffixTree(data []*syntax.Node) *suffixtree.STree {
	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}
	return tree
}

// convertToCloneGroup converts internal format to SDK CloneGroup format.
func (d *detector) convertToCloneGroup(hash string, frags [][]*syntax.Node, method DetectionMethod) *CloneGroup {
	clones := make([]*Clone, len(frags))
	totalSize := 0
	maxLines := 0

	for i, frag := range frags {
		clone := d.convertFragmentToClone(frag)
		clones[i] = clone
		totalSize += clone.Size
		lines := clone.EndLine - clone.StartLine + 1
		if lines > maxLines {
			maxLines = lines
		}
	}

	// Limit clones per group if specified
	if d.opts.MaxClonesPerGroup > 0 && len(clones) > d.opts.MaxClonesPerGroup {
		clones = clones[:d.opts.MaxClonesPerGroup]
	}

	return &CloneGroup{
		Hash:      hash,
		Clones:    clones,
		Size:      totalSize,
		LineCount: maxLines,
		Method:    method,
	}
}

// convertFragmentToClone converts a syntax fragment to SDK Clone format.
func (d *detector) convertFragmentToClone(frag []*syntax.Node) *Clone {
	if len(frag) == 0 {
		return &Clone{}
	}

	// Get file information from first node
	firstNode := frag[0]
	lastNode := frag[len(frag)-1]

	clone := &Clone{
		Filename:  firstNode.Filename,
		StartLine: firstNode.Pos,
		EndLine:   lastNode.End,
		StartPos:  firstNode.Pos,
		EndPos:    lastNode.End,
		Size:      len(frag),
	}

	// Include fragment content if requested
	if d.opts.IncludeFragments {
		clone.Fragment = d.extractFragmentContent(frag)
	}

	return clone
}

// extractFragmentContent extracts the actual source code for a fragment.
func (d *detector) extractFragmentContent(frag []*syntax.Node) string {
	if len(frag) == 0 {
		return ""
	}

	// Read the source file
	content, err := d.opts.FileReader(frag[0].Filename)
	if err != nil {
		d.logger.Warn("Failed to read file %s: %v", frag[0].Filename, err)
		return "[content unavailable]"
	}

	// Extract the relevant lines
	lines := position.SplitLines(content)
	start := frag[0].Pos - 1 // Convert to 0-based
	end := frag[len(frag)-1].End

	if start < 0 || end >= len(lines) {
		return "[content unavailable]"
	}

	var fragmentLines []string
	for i := start; i <= end && i < len(lines); i++ {
		fragmentLines = append(fragmentLines, lines[i])
	}

	return position.JoinLines(fragmentLines)
}

// validateFile checks if a file should be processed.
func (d *detector) validateFile(filename string) error {
	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		return ErrFileNotFound
	}

	// Check file size
	if d.opts.MaxFileSize > 0 && info.Size() > d.opts.MaxFileSize {
		return ErrFileTooLarge
	}

	// Check if file should be ignored
	for _, pattern := range d.opts.IgnoreFiles {
		if matched, _ := filepath.Match(pattern, filepath.Base(filename)); matched {
			return ErrParsingFailed // Use generic error for ignored files
		}
	}

	return nil
}

// buildResult creates final Result structure.
func (d *detector) buildResult(cloneGroups []*CloneGroup, fileCount int) *Result {
	analysisTime := time.Since(d.started)

	// Calculate summary statistics
	totalClones := 0
	for _, group := range cloneGroups {
		totalClones += len(group.Clones)
	}

	return &Result{
		CloneGroups: cloneGroups,
		Summary: &Summary{
			TotalFiles:    fileCount,
			TotalClones:   totalClones,
			TotalGroups:   len(cloneGroups),
			AnalysisTime:  analysisTime,
			MethodsUsed:   d.opts.DetectionMethods,
			LinesAnalyzed: 0, // Calculate actual lines analyzed
		},
		Metadata: &Metadata{
			Version:    "1.0.0", // Get from build info
			Timestamp:  time.Now(),
			ConfigHash: d.hashConfig(d.opts),
			Toolchain:  "go", // Get actual version
		},
	}
}

// reportProgress reports analysis progress if callback is provided.
func (d *detector) reportProgress(percentage float64, stage, currentFile string) {
	if d.opts.ProgressCallback != nil {
		progress := &Progress{
			Stage:       stage,
			Completed:   int(percentage),
			Total:       100,
			Percentage:  percentage,
			Message:     stage,
			CurrentFile: currentFile,
		}

		if err := d.opts.ProgressCallback(progress); err != nil {
			d.logger.Warn("Progress callback error: %v", err)
		}
	}
}

// hashConfig creates a hash of the configuration for metadata.
func (d *detector) hashConfig(opts *Options) string {
	// Simple hash - in real implementation use proper hashing
	return fmt.Sprintf("config-%d-%v", opts.Threshold, opts.DetectionMethods)
}

// convertOptionsToConfig converts SDK options to internal config format.
func convertOptionsToConfig(opts *Options) *config.Config {
	cfg := config.DefaultConfig()
	cfg.Threshold = opts.Threshold

	// Convert detection methods
	cfg.DetectionMethods = make(config.DetectionMethods, len(opts.DetectionMethods))
	for i, method := range opts.DetectionMethods {
		cfg.DetectionMethods[i] = config.DetectionMethod(method)
	}

	cfg.IncludeVendor = opts.IncludeVendor
	cfg.IgnoreFiles = opts.IgnoreFiles

	return cfg
}
