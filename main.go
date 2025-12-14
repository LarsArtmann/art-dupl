package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() {
	// Create root command
	rootCmd := createRootCommand()

	// Execute with fang for enhanced CLI experience
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

// createRootCommand creates the basic Cobra command structure
func createRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dupl",
		Short: "Find code clones",
		Long: `Dupl is a tool for finding code clones in Go source files.

It analyzes abstract syntax trees (ASTs) to find structural code clones 
while ignoring literal values, using suffix tree algorithms for efficiency.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Call the existing Run function from cli.go
			os.Exit(Run())
		},
	}
	return cmd
}