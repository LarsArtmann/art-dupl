package cli

import (
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
)

// DefaultThreshold is the default minimum token sequence size for clone detection.
const DefaultThreshold = 15

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
	DiffMode       bool   // Enable diff visualization for HTML output
	All            string // Output directory for "all" mode (empty means disabled)
	OutputDir      string // Custom output directory
	Paths          []string

	// Runtime configuration
	OutputWriter io.Writer
	ErrorWriter  io.Writer
}

// ToConfig converts RuntimeConfig to config.Config.
func (r *RuntimeConfig) ToConfig() *config.Config {
	cfg := &config.Config{ //nolint:exhaustruct
		Threshold:      r.Threshold,
		IncludeVendor:  r.Vendor,
		FilesFromStdin: r.FilesFromStdin,
		Verbose:        r.Verbose,
		Paths:          r.Paths,
	}

	switch {
	case r.HTML:
		cfg.OutputFormat = config.OutputFormatHTML
	case r.Plumbing:
		cfg.OutputFormat = config.OutputFormatPlumbing
	case r.JSON:
		cfg.OutputFormat = config.OutputFormatJSON
	default:
		// Default to text format when no output format is specified
		cfg.OutputFormat = config.OutputFormatText
	}

	return cfg
}

// DefaultRuntimeConfig returns a default runtime configuration.
func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{ //nolint:exhaustruct
		Threshold:    DefaultThreshold,
		SortBy:       "size",
		OutputWriter: &cliStdout{},
		ErrorWriter:  &cliStderr{},
	}
}

// Interface wrappers for io.Writer to avoid import cycles.
type cliStdout struct{}

//nolint:wrapcheck // Error wrapping not needed for IO Write operations
func (c *cliStdout) Write(p []byte) (n int, err error) { return os.Stdout.Write(p) }

type cliStderr struct{}

//nolint:wrapcheck // Error wrapping not needed for IO Write operations
func (c *cliStderr) Write(p []byte) (n int, err error) { return os.Stderr.Write(p) }
