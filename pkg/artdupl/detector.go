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
	if err := ValidateOptions(opts); err != nil {
		return nil, errors.WrapConfig(err, "invalid options") //nolint:wrapcheck // Error already wrapped by WrapConfig
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
		return nil, errors.WrapValidation(err, fmt.Sprintf("input validation failed for %d files", len(files))) //nolint:wrapcheck // Error already wrapped by WrapValidation
	}

	// Process files and build analysis pipeline
	data, _, err := d.buildAnalysisPipeline(ctx, files)
	if err != nil {
		return nil, errors.Wrap(err, errors.AnalysisError, fmt.Sprintf("analysis pipeline construction failed for %d files", len(files)))
	}

	// Run detection based on configured methods
	cloneGroups, err := d.runDetection(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, errors.DetectionError, "detection failed")
	}

	// Build and return result
	return d.buildResult(cloneGroups, 0), nil
}

// FindClonesStream provides streaming results for large projects.
func (d *detector) FindClonesStream(ctx context.Context, files []string) (<-chan *CloneGroup, error) {
	d.started = time.Now()

	// Validate inputs
	if err := d.validateInputs(ctx, files); err != nil {
		return nil, errors.WrapValidation(err, fmt.Sprintf("input validation failed for streaming with %d files", len(files))) //nolint:wrapcheck // Error already wrapped by WrapValidation
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
