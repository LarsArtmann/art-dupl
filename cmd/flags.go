package cmd

import (
	"github.com/spf13/cobra"
)

// AddFlags adds all flags to the root command.
func AddFlags(rootCmd *cobra.Command) {
	// Add all flags to root command with better descriptions
	rootCmd.Flags().StringP("config", "c", "", "path to configuration file (JSON format)")
	rootCmd.Flags().Bool("vendor", false, "include vendor directory in analysis")
	rootCmd.Flags().CountP("verbose", "v", "enable verbose logging (repeat for more verbosity)")
	rootCmd.Flags().
		IntP("threshold", "t", 15, "minimum token sequence size to consider as clone (default: 15)")
	rootCmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	rootCmd.Flags().
		Bool("html", false, "output results as HTML with syntax-highlighted code fragments")
	rootCmd.Flags().
		BoolP("json", "j", false, "output structured JSON format with metadata and statistics")
	rootCmd.Flags().
		BoolP("plumbing", "p", false, "output machine-readable plumbing format for script integration")
	rootCmd.Flags().
		StringP("sort", "s", "size", "sort clone groups: size (largest first), occurrence (most files first), hash (alphabetical), total-tokens (highest total token count) (default: size)")
	rootCmd.Flags().
		StringP("detection-methods", "m", "art-dupl", "detection methods: hash, art-dupl, or hash,art-dupl (default: art-dupl)")
	rootCmd.Flags().
		BoolP("all", "a", false, "generate all output formats for all detection methods")
	rootCmd.Flags().
		StringP("output-dir", "o", "reports/art-dupl", "output directory for generated files (used with --all)")

	// Add smart filtering flags
	rootCmd.Flags().
		Bool("filter-generated", false, "enable filtering of sqlc.dev and templ.guide generated code (auto-detects sqlc.yaml in parent directories)")
	rootCmd.Flags().
		Bool("include-sqlc", false, "include sqlc.dev generated files (override auto-detection)")
	rootCmd.Flags().
		Bool("include-templ", false, "include templ.guide generated files (override default filtering)")
	rootCmd.Flags().
		StringArray("include-pattern", []string{}, "file patterns to always include (takes precedence over filter)")
	rootCmd.Flags().
		StringArray("exclude-pattern", []string{}, "additional file patterns to exclude")

	// Add incremental analysis flags
	rootCmd.Flags().
		Bool("incremental", false, "enable incremental analysis with AST caching")
	rootCmd.Flags().
		String("since", "", "git reference for incremental mode (e.g., HEAD~1, main, commit-hash)")
	rootCmd.Flags().
		String("cache-dir", "", "cache directory for AST caching (requires --incremental, default: .cache/art-dupl)")
	rootCmd.Flags().Bool("clear-cache", false, "clear cache before running (requires --incremental)")

	// Add hidden flags for advanced features
	rootCmd.Flags().Bool("profile", false, "enable performance profiling")
	rootCmd.Flags().String("timeout", "30m", "maximum execution time (default: 30m)")
	_ = rootCmd.Flags().MarkHidden("profile")
	_ = rootCmd.Flags().MarkHidden("timeout")

	// Add semantic-aware detection flags
	// By default, semantic detection is ENABLED for better accuracy (see config.DefaultConfig)
	// Use --structural to disable and get structural-only matching (may increase false positives)
	rootCmd.Flags().
		Bool("semantic", false, "explicitly enable semantic-aware detection (already the default; use only to override config file)")
	rootCmd.Flags().
		Bool("structural", false, "disable semantic detection and use structural-only matching (may increase false positives) [opt-out from default]")

	// Add concurrent processing flag
	rootCmd.Flags().
		Int("workers", 0, "number of concurrent workers for file parsing (0 = auto-detect based on CPU cores)")
}
