package artdupl

import (
	"fmt"
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

		err := d.opts.ProgressCallback(progress)
		if err != nil {
			d.logger.Warn("Progress callback error: %v", err)
		}
	}
}

// configDebugString returns a debug representation of the resolved config.
// This is NOT a cryptographic hash — it's a human-readable string for
// Metadata.ConfigHash, useful for distinguishing runs in output.
func (d *detector) configDebugString(opts *Options) string {
	return fmt.Sprintf("config-%d-%v", opts.Threshold, opts.DetectionMethods)
}

// detectorConfig holds the resolved configuration used by the detector pipeline.
// This is an SDK-internal type — no dependency on the config package.
type detectorConfig struct {
	Threshold        int
	DetectionMethods []DetectionMethod
	Semantic         bool
}

// convertOptionsToConfig resolves SDK Options into the internal detectorConfig.
// Semantic defaults to true (matching config.DefaultConfig().Semantic).
func convertOptionsToConfig(opts *Options) *detectorConfig {
	return &detectorConfig{
		Threshold:        opts.Threshold,
		DetectionMethods: opts.DetectionMethods,
		Semantic:         true,
	}
}

// methodsToStrings converts SDK DetectionMethod values to plain strings
// for packages that accept method names without importing the SDK types.
func methodsToStrings(methods []DetectionMethod) []string {
	result := make([]string, len(methods))
	for i, m := range methods {
		result[i] = string(m)
	}

	return result
}
