package config

// DetectionConfig holds the subset of configuration needed by the detection layer.
// This decouples detection from the full Config struct, following the Interface
// Segregation Principle — detectors only see what they need.
type DetectionConfig struct {
	// Methods specifies which detection algorithms to run.
	Methods DetectionMethods

	// Verbose enables detailed logging during detection.
	Verbose bool
}
