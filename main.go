package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "art-dupl [command] [flags] [paths...]",
		Short: "Find code clones",
		Long: `art-dupl finds code clones in Go source files.

It analyzes abstract syntax trees (ASTs) to find structural code clones
while ignoring literal values using suffix tree algorithms.

Available commands:
  analyze    Default code analysis (default)
  json       Output JSON format
  html       Output HTML format
  plumbing   Output plumbing format

Examples:
  art-dupl analyze ./src                    # Default analysis
  art-dupl json -t 20 ./src              # JSON with threshold
  art-dupl html --vendor ./src            # HTML with vendor included
  art-dupl plumbing --sort occurrence ./src  # Plumbing sorted by occurrence`,
		Args: cobra.ArbitraryArgs, // Allow any number of positional arguments
	}

	// Add flags to root command (used by all subcommands)
	rootCmd.PersistentFlags().StringP("config", "c", "", "path to configuration file (JSON format)")
	rootCmd.PersistentFlags().BoolP("vendor", "", false, "include vendor directory in analysis")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging to show processing progress")
	rootCmd.PersistentFlags().IntP("threshold", "t", 15, "minimum token sequence size to consider as clone")
	rootCmd.PersistentFlags().BoolP("files", "f", false, "read file names from stdin, one per line")
	rootCmd.PersistentFlags().StringP("sort", "s", "size", "sort clone groups by: size, occurrence, hash")

	// Add subcommands
	rootCmd.AddCommand(newAnalyzeCommand())
	rootCmd.AddCommand(newJSONCommand())
	rootCmd.AddCommand(newHTMLCommand())
	rootCmd.AddCommand(newPlumbingCommand())

	if err := fang.Execute(context.Background(), rootCmd, fang.WithVersion(GetVersion())); err != nil {
		os.Exit(1)
	}
}

// runCmd implements the Cobra command execution
func runCmd(cmd *cobra.Command, args []string) error {
	return runCobraCommand(cmd, args)
}

// newAnalyzeCommand creates the analyze subcommand
func newAnalyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [flags] [paths...]",
		Short: "Default code analysis",
		Long:  "Analyze source files for code clones using default text output.",
		RunE:  runCmd,
		Args:  cobra.ArbitraryArgs,
	}
}

// newJSONCommand creates the json subcommand
func newJSONCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "json [flags] [paths...]",
		Short: "Output JSON format",
		Long:  "Analyze source files and output results in structured JSON format.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Force JSON output by setting the flag
			cmd.Flags().Set("json", "true")
			return runCobraCommand(cmd, args)
		},
		Args: cobra.ArbitraryArgs,
	}
}

// newHTMLCommand creates the html subcommand
func newHTMLCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "html [flags] [paths...]",
		Short: "Output HTML format",
		Long:  "Analyze source files and output results as HTML with syntax highlighting.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Force HTML output by setting the flag
			cmd.Flags().Set("html", "true")
			return runCobraCommand(cmd, args)
		},
		Args: cobra.ArbitraryArgs,
	}
}

// newPlumbingCommand creates the plumbing subcommand
func newPlumbingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "plumbing [flags] [paths...]",
		Short: "Output plumbing format",
		Long:  "Analyze source files and output results in machine-readable plumbing format.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Force plumbing output by setting the flag
			cmd.Flags().Set("plumbing", "true")
			return runCobraCommand(cmd, args)
		},
		Args: cobra.ArbitraryArgs,
	}
}
