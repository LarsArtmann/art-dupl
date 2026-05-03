package cmd

import (
	"io"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/printer"
)

// withThreshold wraps a printer constructor that needs a threshold parameter.
func withThreshold(
	constructor func(io.Writer, printer.ReadFile, int) printer.Printer,
	threshold int,
) func(io.Writer, printer.ReadFile) printer.Printer {
	return func(w io.Writer, fread printer.ReadFile) printer.Printer {
		return constructor(w, fread, threshold)
	}
}

// createPrinter returns the appropriate printer based on output format.
func createPrinter(
	outputFormat config.OutputFormat,
	threshold int,
	diffMode config.DiffMode,
	metadata printer.ReportMetadata,
) func(io.Writer, printer.ReadFile) printer.Printer {
	switch outputFormat {
	case config.OutputFormatHTML:
		return func(w io.Writer, fread printer.ReadFile) printer.Printer {
			return printer.NewHTMLWithOptions(w, fread, diffMode, metadata, threshold)
		}
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	case config.OutputFormatSimpleJSON:
		return printer.NewJSON
	case config.OutputFormatSARIF:
		return withThreshold(printer.NewSARIF, threshold)
	case config.OutputFormatCSV:
		return withThreshold(printer.NewStats, threshold)
	case config.OutputFormatText:
		return printer.NewText
	default:
		return printer.NewText
	}
}
