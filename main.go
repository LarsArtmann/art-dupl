package main

import (
	"context"
	"fmt"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()
	cmd := createRootCommand()
	
	if err := fang.Execute(ctx, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dupl [paths...]",
		Short: "Find code clones in Go source files",
		Long: `Dupl analyzes Go source files to find structural code clones
using suffix tree algorithms. It ignores literal values and focuses on
structural similarities in the abstract syntax trees.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeDuplicationAnalysis(cmd, args)
		},
	}
	
	// Add flags using cobra's persistent flags
	addPersistentFlags(cmd)
	
	return cmd
}

func addPersistentFlags(cmd *cobra.Command) {
	// Output format flags
	cmd.Flags().Bool("html", false, "output results as HTML with syntax-highlighted code fragments")
	cmd.Flags().Bool("json", false, "output structured JSON format with metadata and statistics")
	cmd.Flags().Bool("plumbing", false, "output machine-readable plumbing format for script integration")
	
	// Analysis flags
	cmd.Flags().IntP("threshold", "t", 15, "minimum token sequence size to consider as clone")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose logging to show processing progress")
	cmd.Flags().Bool("vendor", false, "include vendor directory in analysis")
	cmd.Flags().Bool("files", false, "read file names from stdin, one per line")
	
	// Configuration
	cmd.Flags().String("config", "", "path to configuration file (JSON format)")
	cmd.Flags().String("sort", "size", "sort clone groups by: size, occurrence, hash")
}

func executeDuplicationAnalysis(cmd *cobra.Command, args []string) error {
	// Parse flags into configuration
	cfg, err := parseFlagsToConfig(cmd, args)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	
	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// TODO: Execute actual analysis with proper DI
	// For now, just validate flags work
	fmt.Printf("Configuration: %+v\n", cfg)
	return nil
}

func parseFlagsToConfig(cmd *cobra.Command, args []string) (*config.Config, error) {
	cfg := &config.Config{}
	
	// Parse boolean flags
	html, _ := cmd.Flags().GetBool("html")
	json, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	verbose, _ := cmd.Flags().GetBool("verbose")
	vendor, _ := cmd.Flags().GetBool("vendor")
	files, _ := cmd.Flags().GetBool("files")
	
	// Parse integer flags
	threshold, _ := cmd.Flags().GetInt("threshold")
	
	// Parse string flags
	configFile, _ := cmd.Flags().GetString("config")
	sortBy, _ := cmd.Flags().GetString("sort")
	
	// Load config file if specified
	if configFile != "" {
		fileConfig, err := config.LoadConfig(configFile)
		if err != nil {
			return nil, fmt.Errorf("loading config file: %w", err)
		}
		cfg = fileConfig
	}
	
	// Override with CLI flags
	cfg.Threshold = threshold
	cfg.Verbose = verbose
	cfg.IncludeVendor = vendor
	cfg.FilesFromStdin = files
	cfg.Paths = args
	
	// Determine output format
	if html {
		cfg.OutputFormat = config.OutputFormatHTML
	} else if json {
		cfg.OutputFormat = config.OutputFormatJSON
	} else if plumbing {
		cfg.OutputFormat = config.OutputFormatPlumbing
	}
	
	// Parse sort criteria with type safety
	switch config.SortCriteria(sortBy) {
	case config.SortBySize, config.SortByOccurrence, config.SortByHash:
		cfg.SortBy = config.SortCriteria(sortBy)
	default:
		cfg.SortBy = config.SortBySize // Default
	}
	
	return cfg, nil
}