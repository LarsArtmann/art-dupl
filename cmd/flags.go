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
		IntP("threshold", "t", config.DefaultThreshold, "minimum number of duplicated statements to report as clone (default: 5)")
	cmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	cmd.Flags().
		StringP("detection-methods", "m", "art-dupl", "detection methods (comma-separated): art-dupl, hash (default: art-dupl)")

	// Hidden flags for advanced features
	cmd.Flags().Bool("profile", false, "enable performance profiling")
	cmd.Flags().String("timeout", "30m", "maximum execution time (default: 30m)")
	_ = cmd.Flags().MarkHidden("profile")
	_ = cmd.Flags().MarkHidden("timeout")

	// Generated-code inclusion flag (replaces the old per-generator --include-* flags).
	cmd.Flags().
		StringArray("include-generated", []string{}, "include generated code categories (repeatable, comma-separated): sqlc, templ, protobuf, mockgen, stringer, generic, all")
	addDeprecatedIncludeGeneratedFlags(cmd)

	// Custom file pattern filters
	cmd.Flags().
		StringArray("include-pattern", []string{}, "file patterns to always include (takes precedence over filter)")
	cmd.Flags().
		StringArray("exclude-pattern", []string{}, "additional file patterns to exclude")

	// Semantic-aware detection flags
	cmd.Flags().
		Bool("semantic", true, "enable semantic detection (default): alpha-normalize local identifiers so renamed-variable clones (Type 2) are detected")
	cmd.Flags().
		Bool("exact", false, "match identifier names verbatim (copy-paste / Type 1 clones only); disables alpha-normalization")
	cmd.Flags().
		Bool("structural", false, "match by AST shape only, ignoring all names (loosest matching, most candidates)")

	cmd.Flags().Bool("type-aware", false,
		"enable go/types-based detection: encodes each variable's static type into the hash "+
			"so same-name methods on different types (e.g. time.Time.String vs *big.Int.String) "+
			"do not match. Requires full type checking (10-100x slower). Only effective with --semantic (default)")

	cmd.Flags().
		BoolP("quiet", "q", false, "suppress non-essential status output (progress messages, profiling notices)")
	cmd.Flags().Bool("no-color", false, "disable colored output")

	// File type filter
	cmd.Flags().
		String("only", "", "only analyze specific file type: 'go' or 'templ' (default: both)")
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

	// Root-only: scope flags
	rootCmd.Flags().
		Bool("include-node-modules", false, "include node_modules directory in hash-based detection (excluded by default)")

	// Root-only: incremental analysis flags
	rootCmd.Flags().
		Bool("incremental", false, "enable incremental analysis with AST caching")
	rootCmd.Flags().
		String("cache-dir", "", "cache directory for AST caching (requires --incremental, default: .cache/art-dupl)")
	rootCmd.Flags().
		Bool("clear-cache", false, "clear cache before running (requires --incremental)")
	rootCmd.Flags().
		Int("max-cache-entries", 0, "maximum number of cached AST files to keep on disk (0 = unlimited, requires --incremental)")

	// Root-only: concurrent processing flag
	rootCmd.Flags().
		Int("workers", 0, "number of concurrent workers for file parsing (0 = auto-detect based on CPU cores)")

	// Root-only: diff mode flag for HTML output
	rootCmd.Flags().
		String("diff", "", "enable diff visualization for HTML output (values: side-by-side, inline, or true for side-by-side)")

	// Root-only: rich text output with classification badges and priority
	rootCmd.Flags().
		Bool("rich-text", false, "enable enhanced text output with [PRIORITY] [category] badges and actionable suggestions")

	// Root-only: test clone filtering
	rootCmd.Flags().
		Bool("suppress-test-low", true, "suppress low-priority clones in test files (default: true)")
	rootCmd.Flags().
		Int("test-threshold", 0, "separate minimum token count for test files (0 = max(30, threshold))")
	rootCmd.Flags().
		Bool("ignore-tests", false, "exclude *_test.go files from analysis entirely")
	rootCmd.Flags().
		Bool("include-tests", false, "override --ignore-tests and analyze test files normally")
	rootCmd.Flags().
		Bool("dump-tokens", false, "dump the serialized token stream for debugging (skip clone detection)")
	rootCmd.Flags().
		Int("min-lines", 0, "suppress clone groups spanning fewer than N source lines (0 = disabled)")
	rootCmd.Flags().
		Bool("no-actionability", false, "show all clones including non-actionable boilerplate (disable actionability filtering)")
	rootCmd.Flags().
		Bool("explain", false, "explain why each clone group was reported (type, actionability, category, extractability)")
	rootCmd.Flags().
		Bool("no-accept-directives", false, "ignore //art-dupl:accept directives in source code (show all clones)")
	rootCmd.Flags().
		Bool("include-ignored", false, "include gitignored files in analysis (default: honor .gitignore)")
}

// addDeprecatedIncludeGeneratedFlags registers the old per-generator --include-*
// flags as hidden deprecated aliases. They still set the same config values but
// print a warning pointing to --include-generated.
func addDeprecatedIncludeGeneratedFlags(cmd *cobra.Command) {
	deprecatedFlags := []struct {
		name    string
		message string
	}{
		{"include-sqlc", "use --include-generated sqlc"},
		{"include-templ", "use --include-generated templ"},
		{"include-protobuf", "use --include-generated protobuf"},
		{"include-mockgen", "use --include-generated mockgen"},
		{"include-stringer", "use --include-generated stringer"},
		{"include-generic", "use --include-generated generic"},
	}

	for _, f := range deprecatedFlags {
		cmd.Flags().Bool(f.name, false, "")
		_ = cmd.Flags().MarkDeprecated(f.name, f.message)
	}
}
