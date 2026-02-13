package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/errors"
)

// DetectionMethods is a slice of DetectionMethod for type safety.
type DetectionMethods []DetectionMethod

// IsDefault checks if methods have default value.
func (dm DetectionMethods) IsDefault() bool {
	return len(dm) == 1 && dm[0] == DetectionMethodArtDupl
}

// Contains checks if method is in methods list.
func (dm DetectionMethods) Contains(method DetectionMethod) bool {
	return slices.Contains(dm, method)
}

// IsEmpty checks if methods list is empty.
func (dm DetectionMethods) IsEmpty() bool {
	return len(dm) == 0
}

// Config represents the dupl configuration with strong typing.
//
// DOMAIN TYPES STATUS:
// ✅ Added domain package import
// ✅ Threshold field uses int for JSON compatibility
//
// Note: Config is used for loading from YAML/JSON files, so we use primitive types
// (int, string, bool) for JSON marshaling compatibility. For type-safe access,
// use the helper functions below to convert to/from domain types:
//
//	// Get typed threshold
//	:= cfg.GetThresholdAsDomain()
//
//	// Set threshold with validation
//	err := cfg.SetThresholdFromDomain(domainThreshold)
//
// This provides type safety where needed without breaking config file loading.
type Config struct {
	// Threshold sets the minimum token sequence size to consider as duplicate
	Threshold int `json:"threshold,omitempty"`

	// IncludeVendor includes vendor directory in analysis
	IncludeVendor bool `json:"includeVendor,omitempty"`

	// FilesFromStdin reads file paths from stdin when true
	FilesFromStdin bool `json:"filesFromStdin,omitempty"`

	// OutputFormat sets the output format with type safety
	OutputFormat OutputFormat `json:"outputFormat,omitempty"`

	// Verbose enables verbose output
	Verbose bool `json:"verbose,omitempty"`

	// Paths specifies the paths to analyze
	Paths []string `json:"paths,omitempty"`

	// IgnoreFiles specifies file patterns to ignore
	IgnoreFiles []string `json:"ignoreFiles,omitempty"`

	// MaxChildrenSerial sets the maximum children serial for large slices
	MaxChildrenSerial int `json:"maxChildrenSerial,omitempty"`

	// OutputFile specifies the output file (if not stdout)
	OutputFile string `json:"outputFile,omitempty"`

	// SortBy specifies sorting criteria for clone groups
	SortBy SortCriteria `json:"sortBy,omitempty"`

	// DetectionMethods specifies which detection methods to use
	DetectionMethods DetectionMethods `json:"detectionMethods,omitempty"`

	// Profile enables performance profiling output
	Profile bool `json:"profile,omitempty"`

	// Timeout specifies maximum execution time in seconds (0 = no timeout)
	Timeout int `json:"timeout,omitempty"`

	// FilterGenerated enables smart filtering of auto-generated code
	FilterGenerated bool `json:"filterGenerated,omitempty"`

	// IncludeSQLC includes sqlc.dev generated files (only when filterGenerated is true)
	IncludeSQLC bool `json:"includeSQLC,omitempty"`

	// IncludeTempl includes templ.guide generated files (only when filterGenerated is true)
	IncludeTempl bool `json:"includeTempl,omitempty"`

	// IncludePatterns specifies file patterns to always include (takes precedence over filter)
	IncludePatterns []string `json:"includePatterns,omitempty"`

	// ExcludePatterns specifies additional file patterns to exclude
	ExcludePatterns []string `json:"excludePatterns,omitempty"`
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Threshold:         15,
		IncludeVendor:     false,
		FilesFromStdin:    false,
		OutputFormat:      OutputFormatText,
		Verbose:           false,
		Paths:             []string{"."},
		IgnoreFiles:       []string{},
		MaxChildrenSerial: 10000,
		OutputFile:        "",
		SortBy:            SortBySize,
		DetectionMethods:  DetectionMethods{DetectionMethodArtDupl},
		Profile:           false,
		Timeout:           0,
		FilterGenerated:   false,
		IncludeSQLC:       false,
		IncludeTempl:      false,
		IncludePatterns:   []string{},
		ExcludePatterns:   []string{},
	}
}

// GetThresholdAsDomain converts config threshold to domain.Threshold.
//
// Usage:
//
//	cfg := config.DefaultConfig()
//	domainThreshold, err := cfg.GetThresholdAsDomain()
//	if err != nil { ... }
func (c *Config) GetThresholdAsDomain() (domain.Threshold, error) {
	threshold, err := domain.NewThreshold(uint(c.Threshold)) // #nosec G115 -- Threshold values are bounded (0-1000)
	if err != nil {
		return 0, errors.NewConfigError("invalid threshold in config", err)
	}
	return threshold, nil
}

// SetThresholdFromDomain sets threshold from domain.Threshold with validation.
//
// Usage:
//
//	domainThreshold, err := domain.NewThreshold(15)
//	if err != nil { return err }
//	err := cfg.SetThresholdFromDomain(domainThreshold)
func (c *Config) SetThresholdFromDomain(threshold domain.Threshold) error {
	// No validation needed - domain.Threshold already validated
	c.Threshold = int(threshold.Uint()) // #nosec G115 -- Threshold values are bounded (0-1000)
	return nil
}

// LoadConfig loads configuration from file.
func LoadConfig(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.NewConfigError("config file not found: "+filename, nil)
	}

	// #nosec G304 -- filename is controlled config path, not user input
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.NewIOError(filename, "failed to read config file", err)
	}

	config := DefaultConfig()
	if len(data) > 0 {
		err = errors.SafeUnmarshal(data, config, "config file: "+filename)
		if err != nil {
			return nil, errors.NewConfigError("failed to parse config file: "+filename, err)
		}
	}

	return config, nil
}

// SaveConfig saves configuration to file.
func SaveConfig(config *Config, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil { // #nosec G301 -- Config directory needs readable permission
		return errors.NewIOError(dir, "failed to create config directory", err)
	}

	data, err := errors.SafeMarshalIndent(config, "", "  ", "config")
	if err != nil {
		return err //nolint:wrapcheck // Error already wrapped by SafeMarshalIndent
	}

	err = os.WriteFile(filename, data, 0o644) // #nosec G306 -- Config file needs to be readable by user
	if err != nil {
		return errors.NewIOError(filename, "failed to write config file", err)
	}

	return nil
}

// ValidateConfig validates the configuration.
func ValidateConfig(config *Config) error {
	if config.Threshold < 1 {
		return errors.NewValidationError("threshold must be greater than 0", nil)
	}

	if config.Threshold > 1000 {
		return errors.NewValidationError("threshold seems too large (max 1000)", nil)
	}

	if config.MaxChildrenSerial < 1000 {
		return errors.NewValidationError("maxChildrenSerial should be at least 1000", nil)
	}

	if config.MaxChildrenSerial > 100000 {
		return errors.NewValidationError("maxChildrenSerial seems too large (max 100000)", nil)
	}

	if !config.OutputFormat.IsValid() {
		return errors.NewValidationError(fmt.Sprintf("invalid output format: %s (valid: text, html, json, plumbing)", config.OutputFormat), nil)
	}

	// Validate detection methods
	if len(config.DetectionMethods) == 0 {
		return errors.NewValidationError("at least one detection method must be specified", nil)
	}

	for _, method := range config.DetectionMethods {
		if !method.IsValid() {
			return errors.NewValidationError(fmt.Sprintf("invalid detection method: %s (valid: hash, art-dupl, todos, legacy)", method), nil)
		}
	}

	return nil
}
