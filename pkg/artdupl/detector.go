package artdupl

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"
)

// detector implements the Detector interface using existing dupl components.
type detector struct {
	opts   *Options
	cfg    *detectorConfig
	logger Logger
}

// NewDetector creates a new code duplication detector.
func NewDetector(opts *Options) (Detector, error) {
	// Use default options if none provided
	if opts == nil {
		opts = DefaultOptions()
	}

	// Validate options
	err := ValidateOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("invalid options (threshold=%d, methods=%v): %w",
			opts.Threshold, opts.DetectionMethods, err)
	}

	// Set default file reader if not provided
	if opts.FileReader == nil {
		opts.FileReader = os.ReadFile
	}

	// Set default logger if not provided
	if opts.Logger == nil {
		opts.Logger = noOpLogger{}
	}

	// Convert SDK options to internal config
	cfg := convertOptionsToConfig(opts)

	// Deep-copy caller-provided slices to prevent post-construction mutation
	opts.DetectionMethods = slices.Clone(opts.DetectionMethods)
	opts.IgnoreFiles = slices.Clone(opts.IgnoreFiles)

	return &detector{
		opts:   opts,
		cfg:    cfg,
		logger: opts.Logger,
	}, nil
}

// FindClones performs complete duplication analysis.
// Not safe for concurrent use — see FindClonesStreamResult for concurrent scenarios.
func (d *detector) FindClones(ctx context.Context, files []string) (*Result, error) {
	startTime := time.Now()

	// Validate inputs
	err := d.validateInputsWithContext(ctx, files)
	if err != nil {
		return nil, err
	}

	// Process files and build analysis pipeline
	pipeline, err := d.buildAnalysisPipeline(ctx, files)
	if err != nil {
		return nil, fmt.Errorf(
			"analysis pipeline construction failed for %d files (methods=%v, threshold=%d): %w",
			len(files),
			d.cfg.DetectionMethods,
			d.cfg.Threshold,
			err,
		)
	}

	// Run detection based on configured methods
	cloneGroups, err := d.runDetection(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf(
			"detection failed (methods=%v, nodes=%d): %w",
			d.cfg.DetectionMethods,
			len(pipeline.data),
			err,
		)
	}

	// Build and return result
	result := d.buildResult(cloneGroups, pipeline.fileCount.FilesCount, pipeline.fileCount.LinesCount, startTime)

	d.reportProgress(100, "Analysis complete", "")

	if len(result.CloneGroups) == 0 {
		return nil, ErrNoDuplicatesFound
	}

	return result, nil
}

// FindClonesStreamResult provides streaming results with error propagation.
// The channel emits StreamResult values. A final StreamResult with Err != nil
// indicates pipeline failure. The channel is always closed after all results.
func (d *detector) FindClonesStreamResult(
	ctx context.Context,
	files []string,
) (<-chan StreamResult, error) {
	err := d.validateInputsWithContext(ctx, files)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan StreamResult, 10)

	go func() {
		defer close(resultChan)

		pipeline, err := d.buildAnalysisPipeline(ctx, files)
		if err != nil {
			select {
			case resultChan <- StreamResult{Group: nil, Err: err}:
			case <-ctx.Done():
			}

			return
		}

		groupChan := make(chan *CloneGroup, 10)

		go func() {
			defer close(groupChan)

			streamErr := d.streamDetectionResults(ctx, pipeline, groupChan)
			if streamErr != nil {
				select {
				case resultChan <- StreamResult{Group: nil, Err: streamErr}:
				case <-ctx.Done():
				}
			}
		}()

		for group := range groupChan {
			if group != nil {
				select {
				case resultChan <- StreamResult{Group: group, Err: nil}:
				case <-ctx.Done():
					return
				}
			}
		}

		d.reportProgress(100, "Analysis complete", "")
	}()

	return resultChan, nil
}

// wrapValidationError wraps a validation error with context about the operation.
func (d *detector) wrapValidationError(err error, operation string, fileCount int) error {
	return fmt.Errorf("input validation failed for "+operation+"%d files: %w", fileCount, err)
}

// validateInputsWithContext validates inputs and returns the error if validation fails.
// This helper deduplicates the common validation + error wrapping pattern.
func (d *detector) validateInputsWithContext(ctx context.Context, files []string) error {
	err := d.validateInputs(ctx, files)
	if err != nil {
		return d.wrapValidationError(err, "validation failed for ", len(files))
	}

	return nil
}

// Close releases any resources held by the detector.
func (d *detector) Close() error {
	// No persistent resources to clean up currently
	return nil
}
