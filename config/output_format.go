package config

import "github.com/LarsArtmann/art-dupl/domain"

type OutputFormat = domain.OutputFormat

const (
	OutputFormatText       = domain.OutputFormatText
	OutputFormatHTML       = domain.OutputFormatHTML
	OutputFormatJSON       = domain.OutputFormatJSON
	OutputFormatCSV        = domain.OutputFormatCSV
	OutputFormatPlumbing   = domain.OutputFormatPlumbing
	OutputFormatSimpleJSON = domain.OutputFormatSimpleJSON
	OutputFormatSARIF      = domain.OutputFormatSARIF
)

var (
	ErrInvalidOutputFormat = domain.ErrInvalidOutputFormat
	AllOutputFormats       = domain.AllOutputFormats
	DefaultOutputFormat    = domain.DefaultOutputFormat
	ParseOutputFormat      = domain.ParseOutputFormat
)
