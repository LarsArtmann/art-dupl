package config

// FileConfig represents file-based configuration.
type FileConfig struct {
	Threshold       int          `json:"threshold"`
	Verbose         bool         `json:"verbose"`
	Vendor          bool         `json:"vendor"`
	Paths           []string     `json:"paths"`
	Profile         bool         `json:"profile"`
	Timeout         string       `json:"timeout"`
	OutputDir       string       `json:"outputDir"`
	OutputFormat    OutputFormat `json:"outputFormat"`
	SortBy          SortCriteria `json:"sortBy"`
	DetectionMethod string       `json:"detectionMethod"`
}
