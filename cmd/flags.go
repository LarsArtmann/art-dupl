package cmd

import (
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/spf13/cobra"
)

// addSharedFlags adds flags common to both the root command and the stats subcommand.
// This eliminates duplication of config, vendor, verbose, threshold, files,
// detection-methods, profile, timeout, filter, include/exclude, and semantic flags.
func addSharedFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "", "path to configuration file (JSON format)")
	cmd.Flags().Bool("vendor", false, "include vendor directory in analysis")
	cmd.Flags().CountP("verbose", "v", "enable verbose logging (repeat for more verbosity)")
	cmd.Flags().
		IntP("threshold", "t", config.DefaultThreshold, "minimum token sequence size to consider as clone (default: 15)")
	cmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	cmd.Flags().
		StringP("detection-methods", "m", "art-dupl", "detection methods: hash, art-dupl, or hash,art-dupl (default: art-dupl)")

	// Hidden flags for advanced features
	cmd.Flags().Bool("profile", false, "enable performance profiling")
	cmd.Flags().String("timeout", "30m", "maximum execution time (default: 30m)")
	_ = cmd.Flags().MarkHidden("profile")
	_ = cmd.Flags().MarkHidden("timeout")

	// Smart filtering flags
	cmd.Flags().
		Bool("include-sqlc", false, "include sqlc.dev generated files (override auto-detection)")
	cmd.Flags().
		Bool("include-templ", false, "include templ-generated *_templ.go files in analysis (default: filtered)")
	cmd.Flags().
		Bool("include-protobuf", false, "include protobuf generated files (.pb.go, _grpc.pb.go)")
	cmd.Flags().
		Bool("include-mockgen", false, "include mockgen generated files")
	cmd.Flags().
		Bool("include-stringer", false, "include stringer generated files")
	cmd.Flags().
		StringArray("include-pattern", []string{}, "file patterns to always include (takes precedence over filter)")
	cmd.Flags().
		StringArray("exclude-pattern", []string{}, "additional file patterns to exclude")

	// Semantic-aware detection flags
	cmd.Flags().
		Bool("semantic", true, "enable semantic-aware detection (on by default; matches by structure AND identifier names)")
	cmd.Flags().
		Bool("structural", false, "disable semantic detection; match by structure only (increases false positives from similar-looking but semantically different code)")
}

// AddFlags adds all flags to the root command.
func AddFlags(rootCmd *cobra.Command) {
	addSharedFlags(rootCmd)

	// Root-only: output format flags
	rootCmd.Flags().
		Bool("html", false, "output results as HTML with syntax-highlighted code fragments")
	rootCmd.Flags().
		BoolP("json", "j", false, "output structured JSON format with metadata and statistics")
	rootCmd.Flags().
		BoolP("plumbing", "p", false, "output machine-readable plumbing format for script integration")
	rootCmd.Flags().
		Bool("sarif", false, "output SARIF format for security tool integration (GitHub Advanced Security, CodeQL)")
	rootCmd.Flags().
		Bool("simple-json", false, "output simplified JSON format with impact scores")
	rootCmd.Flags().
		StringP("sort", "s", "size", "sort clone groups: size (largest first), occurrence (most files first), hash (alphabetical), total-tokens (highest total token count) (default: size)")
	rootCmd.Flags().
		BoolP("all", "a", false, "generate all output formats for all detection methods")
	rootCmd.Flags().
		StringP("output-dir", "o", "reports/art-dupl", "output directory for generated files (used with --all)")

	// Root-only: scope and file type flags
	rootCmd.Flags().
		Bool("include-node-modules", false, "include node_modules directory in hash-based detection (excluded by default)")
	rootCmd.Flags().
		String("only", "", "only analyze specific file type: 'go' or 'templ' (default: both)")

	// Root-only: incremental analysis flags
	rootCmd.Flags().
		Bool("incremental", false, "enable incremental analysis with AST caching")
	rootCmd.Flags().
		String("since", "", "git reference for incremental mode (e.g., HEAD~1, main, commit-hash)")
	rootCmd.Flags().
		String("cache-dir", "", "cache directory for AST caching (requires --incremental, default: .cache/art-dupl)")
	rootCmd.Flags().
		Bool("clear-cache", false, "clear cache before running (requires --incremental)")

	// Root-only: concurrent processing flag
	rootCmd.Flags().
		Int("workers", 0, "number of concurrent workers for file parsing (0 = auto-detect based on CPU cores)")

	// Root-only: diff mode flag for HTML output
	rootCmd.Flags().
		String("diff", "", "enable diff visualization for HTML output (values: side-by-side, inline, or true for side-by-side)")

	// Root-only: rich text output with classification badges and priority
	rootCmd.Flags().
		Bool("rich-text", false, "enable enhanced text output with [PRIORITY] [category] badges and actionable suggestions")
}
