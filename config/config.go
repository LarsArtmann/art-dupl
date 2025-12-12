package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golangci/dupl/errors"
)

// Config represents the dupl configuration
type Config struct {
	// Threshold sets the minimum token sequence size to consider as duplicate
	Threshold int `json:"threshold,omitempty"`
	
	// IncludeVendor includes vendor directory in analysis
	IncludeVendor bool `json:"includeVendor,omitempty"`
	
	// FilesFromStdin reads file paths from stdin when true
	FilesFromStdin bool `json:"filesFromStdin,omitempty"`
	
	// OutputFormat sets the output format (text, html, json, plumbing)
	OutputFormat string `json:"outputFormat,omitempty"`
	
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
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Threshold:         15,
		IncludeVendor:     false,
		FilesFromStdin:    false,
		OutputFormat:      "text",
		Verbose:           false,
		Paths:             []string{"."},
		IgnoreFiles:       []string{},
		MaxChildrenSerial: 10000,
		OutputFile:        "",
	}
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.NewConfigError(fmt.Sprintf("config file not found: %s", filename), nil)
	}
	
	data, err := ioutil.ReadFile(filename)
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewIOError(dir, "failed to create config directory", err)
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.NewInternalError("failed to marshal config", err)
	}
	
	err = ioutil.WriteFile(filename, data, 0644)
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
	
	validFormats := map[string]bool{
		"text":     true,
		"html":     true,
		"json":     true,
		"plumbing": true,
	}
	
	if !validFormats[config.OutputFormat] {
		return errors.NewValidationError(fmt.Sprintf("invalid output format: %s (valid: text, html, json, plumbing)", config.OutputFormat), nil)
	}
	
	return nil
}

// MergeConfigs merges two configurations, with command line config taking precedence
func MergeConfigs(fileConfig, cliConfig *Config) *Config {
	result := DefaultConfig()
	
	// Start with file config
	if fileConfig != nil {
		if fileConfig.Threshold != 0 {
			result.Threshold = fileConfig.Threshold
		}
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
	}
	
	// Override with CLI config
	if cliConfig != nil {
		if cliConfig.Threshold != 0 {
			result.Threshold = cliConfig.Threshold
		}
		if cliConfig.IncludeVendor {
			result.IncludeVendor = cliConfig.IncludeVendor
		}
		if cliConfig.FilesFromStdin {
			result.FilesFromStdin = cliConfig.FilesFromStdin
		}
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
	}
	
	return result
}