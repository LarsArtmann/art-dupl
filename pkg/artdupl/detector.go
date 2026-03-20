package artdupl

import (
	"context"
	"fmt"
	"os"
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
		return nil, d.wrapValidationError(err, fmt.Sprintf("validation failed for %d files", len(files)), len(files))
	}

	// Process files and build analysis pipeline
	data, _, err := d.buildAnalysisPipeline(ctx, files)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.AnalysisError,
			fmt.Sprintf("analysis pipeline construction failed for %d files (methods=%v, threshold=%d)",
				len(files), d.config.DetectionMethods, d.config.Threshold),
		)
	}

	// Run detection based on configured methods
	cloneGroups, err := d.runDetection(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, errors.DetectionError,
			fmt.Sprintf("detection failed (methods=%v, nodes=%d)", d.config.DetectionMethods, len(data)))
	}

	// Build and return result
	return d.buildResult(cloneGroups, 0), nil
}

// FindClonesStream provides streaming results for large projects.
func (d *detector) FindClonesStream(
	ctx context.Context,
	files []string,
) (<-chan *CloneGroup, error) {
	d.started = time.Now()

	// Validate inputs
	err := d.validateInputs(ctx, files)
	if err != nil {
		return nil, d.wrapValidationError(err,
			fmt.Sprintf("streaming validation failed for %d files", len(files)), len(files))
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

// wrapValidationError wraps a validation error with context about the operation.
func (d *detector) wrapValidationError(err error, operation string, fileCount int) error {
	return errors.WrapValidation(
		err,
		fmt.Sprintf("input validation failed for "+operation+"%d files", fileCount),
	)
}

// Close releases any resources held by the detector.
func (d *detector) Close() error {
	// No persistent resources to clean up currently
	return nil
}
