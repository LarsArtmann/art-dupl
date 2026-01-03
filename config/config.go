package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/errors"
)

// Config represents the dupl configuration with strong typing.
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
	}
}

// LoadConfig loads configuration from file.
func LoadConfig(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.NewConfigError("config file not found: "+filename, nil)
	}

	//nolint:gosec // G304: filename is controlled config path, not user input
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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.NewIOError(dir, "failed to create config directory", err)
	}

	data, err := errors.SafeMarshalIndent(config, "", "  ", "config")
	if err != nil {
		return err //nolint:wrapcheck // Error already wrapped by SafeMarshalIndent
	}

	err = os.WriteFile(filename, data, 0o644)
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

// MergeConfigs merges two configurations, with command line config taking precedence.
func MergeConfigs(fileConfig, cliConfig *Config) *Config {
	result := DefaultConfig()

	mergeFileConfig(result, fileConfig)
	mergeCLIConfig(result, cliConfig)

	return result
}

func mergeFileConfig(result, cfg *Config) {
	if cfg == nil {
		return
	}
	result.Threshold = cfg.Threshold
	result.IncludeVendor = cfg.IncludeVendor
	result.FilesFromStdin = cfg.FilesFromStdin
	if cfg.OutputFormat != "" {
		result.OutputFormat = cfg.OutputFormat
	}
	result.Verbose = cfg.Verbose
	if len(cfg.Paths) > 0 {
		result.Paths = cfg.Paths
	}
	if len(cfg.IgnoreFiles) > 0 {
		result.IgnoreFiles = cfg.IgnoreFiles
	}
	if cfg.MaxChildrenSerial != 0 {
		result.MaxChildrenSerial = cfg.MaxChildrenSerial
	}
	if cfg.OutputFile != "" {
		result.OutputFile = cfg.OutputFile
	}
	if len(cfg.DetectionMethods) > 0 {
		result.DetectionMethods = cfg.DetectionMethods
	}
}

func mergeCLIConfig(result, cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.Threshold != 0 {
		result.Threshold = cfg.Threshold
	}
	if cfg.IncludeVendor {
		result.IncludeVendor = cfg.IncludeVendor
	}
	if cfg.FilesFromStdin {
		result.FilesFromStdin = cfg.FilesFromStdin
	}
	if cfg.OutputFormat != "" {
		result.OutputFormat = cfg.OutputFormat
	}
	if cfg.Verbose {
		result.Verbose = cfg.Verbose
	}
	if len(cfg.Paths) > 0 {
		result.Paths = cfg.Paths
	}
	if len(cfg.IgnoreFiles) > 0 {
		result.IgnoreFiles = cfg.IgnoreFiles
	}
	if cfg.MaxChildrenSerial != 0 {
		result.MaxChildrenSerial = cfg.MaxChildrenSerial
	}
	if cfg.OutputFile != "" {
		result.OutputFile = cfg.OutputFile
	}
	if len(cfg.DetectionMethods) > 0 {
		result.DetectionMethods = cfg.DetectionMethods
	}
}
