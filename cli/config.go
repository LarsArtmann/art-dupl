package main

import (
	"flag"
	"path/filepath"

	"github.com/spf13/cobra"
)

// CLIConfig holds all CLI flag values
type CLIConfig struct {
	ConfigFile    *string
	Vendor        *bool
	Verbose       *bool
	VerboseLong   *bool
	Threshold     *int
	ThresholdLong *int
	Files         *bool
	HTML          *bool
	JSONFlag      *bool
	Plumbing      *bool
	SortBy        *string
	Paths         []string
}

// NewCLIConfig creates a new CLI configuration with default flags
func NewCLIConfig() *CLIConfig {
	return &CLIConfig{
		ConfigFile:    flag.String("config", "", "path to configuration file (JSON format)"),
		Vendor:        flag.Bool("vendor", false, "include vendor directory in analysis"),
		Verbose:       flag.Bool("v", false, "enable verbose logging to show processing progress"),
		VerboseLong:   flag.Bool("verbose", false, "enable verbose logging to show processing progress"),
		Threshold:     flag.Int("t", 15, "minimum token sequence size to consider as clone"),
		ThresholdLong: flag.Int("threshold", 15, "minimum token sequence size to consider as clone"),
		Files:         flag.Bool("files", false, "read file names from stdin, one per line"),
		HTML:          flag.Bool("html", false, "output results as HTML with syntax-highlighted code fragments"),
		JSONFlag:      flag.Bool("json", false, "output structured JSON format with metadata and statistics"),
		Plumbing:      flag.Bool("plumbing", false, "output machine-readable plumbing format for script integration"),
		SortBy:        flag.String("sort", "size", "sort clone groups by: size, occurrence, hash, total-tokens"),
	}
}

// AddFlagsToCommand adds CLI configuration flags to a Cobra command
func (c *CLIConfig) AddFlagsToCommand(cmd *cobra.Command) {
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
}

// GetThreshold returns the threshold value, preferring long form
func (c *CLIConfig) GetThreshold() int {
	if *c.ThresholdLong != 15 {
		return *c.ThresholdLong
	}
	return *c.Threshold
}

// IsVerbose returns true if verbose logging is enabled
func (c *CLIConfig) IsVerbose() bool {
	return *c.Verbose || *c.VerboseLong
}

// GetOutputFormats returns the output formats specified
func (c *CLIConfig) GetOutputFormats() []string {
	var formats []string
	if *c.HTML {
		formats = append(formats, "html")
	}
	if *c.JSONFlag {
		formats = append(formats, "json")
	}
	if *c.Plumbing {
		formats = append(formats, "plumbing")
	}
	if len(formats) == 0 {
		formats = append(formats, "text")
	}
	return formats
}

// Constants
const (
	DefaultThreshold = 15
	VendorDirPrefix  = "vendor" + string(filepath.Separator)
	VendorDirInPath  = string(filepath.Separator) + VendorDirPrefix
)