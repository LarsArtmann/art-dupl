package cli

import (
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
)

// RuntimeConfig represents the runtime configuration from CLI flags and config files.
type RuntimeConfig struct {
	// CLI Flags
	ConfigFile     string
	Vendor         bool
	Verbose        bool
	Threshold      int
	FilesFromStdin bool
	HTML           bool
	JSON           bool
	Plumbing       bool
	SortBy         string
	All            string // Output directory for "all" mode (empty means disabled)
	OutputDir      string // Custom output directory
	Paths          []string

	// Runtime configuration
	OutputWriter io.Writer
	ErrorWriter  io.Writer
}

// ToConfig converts RuntimeConfig to config.Config.
func (r *RuntimeConfig) ToConfig() *config.Config {
	cfg := &config.Config{
		Threshold:      r.Threshold,
		IncludeVendor:  r.Vendor,
		FilesFromStdin: r.FilesFromStdin,
		Verbose:        r.Verbose,
		Paths:          r.Paths,
	}

	if r.HTML {
		cfg.OutputFormat = config.OutputFormatHTML
	} else if r.Plumbing {
		cfg.OutputFormat = config.OutputFormatPlumbing
	} else if r.JSON {
		cfg.OutputFormat = config.OutputFormatJSON
	} else {
		// Default to text format when no output format is specified
		cfg.OutputFormat = config.OutputFormatText
	}

	return cfg
}

// DefaultRuntimeConfig returns a default runtime configuration.
func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		Threshold:    15, // defaultThreshold
		SortBy:       "size",
		OutputWriter: &cliStdout{},
		ErrorWriter:  &cliStderr{},
	}
}

// Interface wrappers for io.Writer to avoid import cycles.
type cliStdout struct{}

func (c *cliStdout) Write(p []byte) (n int, err error) { return os.Stdout.Write(p) }

type cliStderr struct{}

func (c *cliStderr) Write(p []byte) (n int, err error) { return os.Stderr.Write(p) }
