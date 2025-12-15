package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/errors"
)

// Config represents the dupl configuration with strong typing
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
}

// DefaultConfig returns a default configuration
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
	}
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.NewConfigError(fmt.Sprintf("config file not found: %s", filename), nil)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.NewIOError(filename, "failed to read config file", err)
	}

	config := DefaultConfig()
	if len(data) > 0 {
		err = json.Unmarshal(data, config)
		if err != nil {
			return nil, errors.NewConfigError(fmt.Sprintf("failed to parse config file: %s", filename), err)
		}
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.NewIOError(dir, "failed to create config directory", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.NewInternalError("failed to marshal config", err)
	}

	err = os.WriteFile(filename, data, 0o644)
	if err != nil {
		return errors.NewIOError(filename, "failed to write config file", err)
	}

	return nil
}

// ValidateConfig validates the configuration
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
			return errors.NewValidationError(fmt.Sprintf("invalid detection method: %s (valid: hash, art-dupl)", method), nil)
		}
	}

	return nil
}

// MergeConfigs merges two configurations, with command line config taking precedence
func MergeConfigs(fileConfig, cliConfig *Config) *Config {
	result := DefaultConfig()

	// Start with file config
	if fileConfig != nil {
		result.Threshold = fileConfig.Threshold
		result.IncludeVendor = fileConfig.IncludeVendor
		result.FilesFromStdin = fileConfig.FilesFromStdin
		if fileConfig.OutputFormat != "" {
			result.OutputFormat = fileConfig.OutputFormat
		}
		result.Verbose = fileConfig.Verbose
		if len(fileConfig.Paths) > 0 {
			result.Paths = fileConfig.Paths
		}
		if len(fileConfig.IgnoreFiles) > 0 {
			result.IgnoreFiles = fileConfig.IgnoreFiles
		}
		if fileConfig.MaxChildrenSerial != 0 {
			result.MaxChildrenSerial = fileConfig.MaxChildrenSerial
		}
		if fileConfig.OutputFile != "" {
			result.OutputFile = fileConfig.OutputFile
		}
		if len(fileConfig.DetectionMethods) > 0 {
			result.DetectionMethods = fileConfig.DetectionMethods
		}
	}

	// Override with CLI config
	if cliConfig != nil {
		// Note: Only set CLI values if they're different from defaults
		// to allow file config values to take precedence
		if cliConfig.Threshold != 15 {
			result.Threshold = cliConfig.Threshold
		}
		if cliConfig.IncludeVendor {
			result.IncludeVendor = cliConfig.IncludeVendor
		}
		if cliConfig.FilesFromStdin {
			result.FilesFromStdin = cliConfig.FilesFromStdin
		}
		// Only override output format if CLI provided it
		if cliConfig.OutputFormat != "" {
			result.OutputFormat = cliConfig.OutputFormat
		}
		if cliConfig.Verbose {
			result.Verbose = cliConfig.Verbose
		}
		if len(cliConfig.Paths) > 0 {
			result.Paths = cliConfig.Paths
		}
		if len(cliConfig.IgnoreFiles) > 0 {
			result.IgnoreFiles = cliConfig.IgnoreFiles
		}
		if cliConfig.MaxChildrenSerial != 0 {
			result.MaxChildrenSerial = cliConfig.MaxChildrenSerial
		}
		if cliConfig.OutputFile != "" {
			result.OutputFile = cliConfig.OutputFile
		}
		if len(cliConfig.DetectionMethods) > 0 {
			result.DetectionMethods = cliConfig.DetectionMethods
		}
	}

	return result
}
