package artdupl

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/config"
)

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
	copy(cfg.DetectionMethods, opts.DetectionMethods)

	cfg.IncludeVendor = opts.IncludeVendor
	cfg.IgnoreFiles = opts.IgnoreFiles

	return cfg
}

