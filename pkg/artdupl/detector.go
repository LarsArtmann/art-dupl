package artdupl

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
)

// detector implements the Detector interface using existing dupl components.
type detector struct {
	opts    *Options
	config  *config.Config
	logger  Logger
	started time.Time
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
		return nil, errors.WrapConfig(err, fmt.Sprintf("invalid options (threshold=%d, methods=%v)",
			opts.Threshold, opts.DetectionMethods))
	}

	// Set default file reader if not provided
	if opts.FileReader == nil {
		opts.FileReader = os.ReadFile
	}

	// Set default logger if not provided
	if opts.Logger == nil {
		opts.Logger = logger.Default
	}

	// Convert SDK options to internal config
	cfg := convertOptionsToConfig(opts)

	// Deep-copy caller-provided slices to prevent post-construction mutation
	opts.DetectionMethods = slices.Clone(opts.DetectionMethods)
	opts.IgnoreFiles = slices.Clone(opts.IgnoreFiles)

	return &detector{ //nolint:exhaustruct
		opts:   opts,
		config: cfg,
		logger: opts.Logger,
	}, nil
}

// FindClones performs complete duplication analysis.
// Not safe for concurrent use — see FindClonesStream for concurrent scenarios.
func (d *detector) FindClones(ctx context.Context, files []string) (*Result, error) {
	d.started = time.Now()

	// Validate inputs
	err := d.validateInputsOrError(ctx, files)
	if err != nil {
		return nil, err
	}

	// Process files and build analysis pipeline
	pipeline, err := d.buildAnalysisPipeline(ctx, files)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.AnalysisError,
			fmt.Sprintf(
				"analysis pipeline construction failed for %d files (methods=%v, threshold=%d)",
				len(files),
				d.config.DetectionMethods,
				d.config.Threshold,
			),
		)
	}

	// Run detection based on configured methods
	cloneGroups, err := d.runDetection(ctx, pipeline)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.DetectionError,
			fmt.Sprintf(
				"detection failed (methods=%v, nodes=%d)",
				d.config.DetectionMethods,
				len(pipeline.data),
			),
		)
	}

	// Build and return result
	result := d.buildResult(cloneGroups, pipeline.fileCount.FilesCount)

	if len(result.CloneGroups) == 0 {
		return nil, ErrNoDuplicatesFound
	}

	return result, nil
}

// FindClonesStream provides streaming results for large projects.
func (d *detector) FindClonesStream(
	ctx context.Context,
	files []string,
) (<-chan *CloneGroup, error) {
	d.started = time.Now()

	// Validate inputs
	err := d.validateInputsOrError(ctx, files)
	if err != nil {
		return nil, err
	}

	// Create output channel
	resultChan := make(chan *CloneGroup, 10)

	// Start streaming analysis
	go func() {
		defer close(resultChan)

		// Process files and build analysis pipeline
		pipeline, err := d.buildAnalysisPipeline(ctx, files)
		if err != nil {
			d.logger.Error("Analysis pipeline error: %v", err)

			return
		}

		// Run detection with streaming
		err = d.streamDetectionResults(ctx, pipeline, resultChan)
		if err != nil {
			d.logger.Error("Streaming detection error: %v", err)
		}

		d.reportProgress(100, "Analysis complete", "")
	}()

	return resultChan, nil
}

// FindClonesStreamResult provides streaming results with error propagation.
// The channel emits StreamResult values. A final StreamResult with Err != nil
// indicates pipeline failure. The channel is always closed after all results.
func (d *detector) FindClonesStreamResult(
	ctx context.Context,
	files []string,
) (<-chan StreamResult, error) {
	d.started = time.Now()

	err := d.validateInputsOrError(ctx, files)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan StreamResult, 10)

	go func() {
		defer close(resultChan)

		pipeline, err := d.buildAnalysisPipeline(ctx, files)
		if err != nil {
			resultChan <- StreamResult{Group: nil, Err: err}

			return
		}

		groupChan := make(chan *CloneGroup, 10)

		go func() {
			defer close(groupChan)

			streamErr := d.streamDetectionResults(ctx, pipeline, groupChan)
			if streamErr != nil {
				resultChan <- StreamResult{Group: nil, Err: streamErr}
			}
		}()

		for group := range groupChan {
			if group != nil {
				resultChan <- StreamResult{Group: group, Err: nil}
			}
		}

		d.reportProgress(100, "Analysis complete", "")
	}()

	return resultChan, nil
}

// wrapValidationError wraps a validation error with context about the operation.
func (d *detector) wrapValidationError(err error, operation string, fileCount int) error {
	return errors.WrapValidation(
		err,
		fmt.Sprintf("input validation failed for "+operation+"%d files", fileCount),
	)
}

// validateInputsOrError validates inputs and returns the error if validation fails.
// This helper deduplicates the common validation + error wrapping pattern.
func (d *detector) validateInputsOrError(ctx context.Context, files []string) error {
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
