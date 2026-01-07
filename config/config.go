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

	// Profile enables performance profiling output
	Profile bool `json:"profile,omitempty"`

	// Timeout specifies maximum execution time in seconds (0 = no timeout)
	Timeout int `json:"timeout,omitempty"`
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
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec //G301 Config directory needs readable permission
		return errors.NewIOError(dir, "failed to create config directory", err)
	}

	data, err := errors.SafeMarshalIndent(config, "", "  ", "config")
	if err != nil {
		return err //nolint:wrapcheck // Error already wrapped by SafeMarshalIndent
	}

	err = os.WriteFile(filename, data, 0o644) //nolint:gosec //G306 Config file needs to be readable by user
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



// mergeConfig merges source config into result config.
// If skipZeroValues is true, fields with zero/empty values are skipped.
// This provides a single source of truth for config merging.
func mergeConfig(result, cfg *Config, skipZeroValues bool) {
	if cfg == nil {
		return
	}

	// Threshold (int)
	if !skipZeroValues || cfg.Threshold != 0 {
		result.Threshold = cfg.Threshold
	}

	// IncludeVendor (bool)
	if !skipZeroValues || cfg.IncludeVendor {
		result.IncludeVendor = cfg.IncludeVendor
	}

	// FilesFromStdin (bool)
	if !skipZeroValues || cfg.FilesFromStdin {
		result.FilesFromStdin = cfg.FilesFromStdin
	}

	// OutputFormat (OutputFormat)
	if !skipZeroValues || cfg.OutputFormat != "" {
		result.OutputFormat = cfg.OutputFormat
	}

	// Verbose (bool)
	if !skipZeroValues || cfg.Verbose {
		result.Verbose = cfg.Verbose
	}

	// Paths ([]string)
	if !skipZeroValues || len(cfg.Paths) > 0 {
		result.Paths = cfg.Paths
	}

	// IgnoreFiles ([]string)
	if !skipZeroValues || len(cfg.IgnoreFiles) > 0 {
		result.IgnoreFiles = cfg.IgnoreFiles
	}

	// MaxChildrenSerial (int)
	if !skipZeroValues || cfg.MaxChildrenSerial != 0 {
		result.MaxChildrenSerial = cfg.MaxChildrenSerial
	}

	// OutputFile (string)
	if !skipZeroValues || cfg.OutputFile != "" {
		result.OutputFile = cfg.OutputFile
	}

	// SortBy (SortCriteria) - not in old functions, adding for completeness
	if !skipZeroValues || cfg.SortBy != "" {
		result.SortBy = cfg.SortBy
	}

	// DetectionMethods (DetectionMethods) - not in old functions, adding for completeness
	if !skipZeroValues || len(cfg.DetectionMethods) > 0 {
		result.DetectionMethods = cfg.DetectionMethods
	}

	// Profile (bool) - not in old functions, adding for completeness
	if !skipZeroValues || cfg.Profile {
		result.Profile = cfg.Profile
	}

	// Timeout (int) - not in old functions, adding for completeness
	if !skipZeroValues || cfg.Timeout != 0 {
		result.Timeout = cfg.Timeout
	}
}
func mergeFileConfig(result, cfg *Config) {
	mergeConfig(result, cfg, false)
}

func mergeCLIConfig(result, cfg *Config) { //nolint:cyclop // Config merging with multiple optional fields
	mergeConfig(result, cfg, true)
}
