package cmd

import (
	"io"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/printer"
)

// printerConstructor is the function signature for creating a Printer.
type printerConstructor func(io.Writer, printer.ReadFile) printer.Printer

// withThreshold wraps a printer constructor that needs a threshold parameter.
func withThreshold(
	constructor func(io.Writer, printer.ReadFile, int) printer.Printer,
	threshold int,
) func(io.Writer, printer.ReadFile) printer.Printer {
	return func(writer io.Writer, fileReader printer.ReadFile) printer.Printer {
		return constructor(writer, fileReader, threshold)
	}
}

// createPrinter returns the appropriate printer based on output format.
func createPrinter(
	outputFormat config.OutputFormat,
	threshold int,
	diffMode config.DiffMode,
	metadata printer.ReportMetadata,
	version string,
) printerConstructor {
	switch outputFormat {
	case config.OutputFormatHTML:
		return func(out io.Writer, reader printer.ReadFile) printer.Printer {
			return printer.NewHTMLWithOptions(out, reader, diffMode, metadata, threshold)
		}
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	case config.OutputFormatSimpleJSON:
		return printer.NewJSON
	case config.OutputFormatSARIF:
		return func(dst io.Writer, src printer.ReadFile) printer.Printer {
			return printer.NewSARIFWithConfig(dst, src, printer.SARIFPrinterOptions{
				Threshold: threshold,
				Version:   version,
			})
		}
	case config.OutputFormatCSV:
		return withThreshold(printer.NewStats, threshold)
	case config.OutputFormatText:
		return printer.NewText
	default:
		return printer.NewText
	}
}
