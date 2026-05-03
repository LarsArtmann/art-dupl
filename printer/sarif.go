package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// SARIFOutput represents the SARIF (Static Analysis Results Interchange Format) output.
// This format is used by security tools like GitHub Advanced Security, CodeQL, etc.
// Spec: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
type SARIFOutput struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents a single analysis run.
type SARIFRun struct {
	Tool        SARIFTool         `json:"tool"`
	Results     []SARIFResult     `json:"results"`
	Invocations []SARIFInvocation `json:"invocations,omitempty"`
}

// SARIFTool represents the tool that performed the analysis.
type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

// SARIFDriver represents the tool driver information.
type SARIFDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

// SARIFRule represents a rule/check that was violated.
type SARIFRule struct {
	ID                   string             `json:"id"`
	Name                 string             `json:"name"`
	ShortDescription     SARIFTextContent   `json:"shortDescription"`
	FullDescription      SARIFTextContent   `json:"fullDescription"`
	DefaultConfiguration SARIFConfiguration `json:"defaultConfiguration"`
	HelpURI              string             `json:"helpUri,omitempty"`
}

// SARIFTextContent represents text content in SARIF.
type SARIFTextContent struct {
	Text string `json:"text"`
}

// SARIFConfiguration represents rule configuration.
type SARIFConfiguration struct {
	Level string `json:"level"`
}

// SARIFResult represents a single result (finding).
type SARIFResult struct {
	RuleID       string            `json:"ruleId"`
	Level        string            `json:"level"`
	Message      SARIFMessage      `json:"message"`
	Locations    []SARIFLocation   `json:"locations"`
	Fingerprints SARIFFingerprints `json:"fingerprints"`
}

// SARIFMessage represents a message in a result.
type SARIFMessage struct {
	Text string `json:"text"`
}

// SARIFLocation represents a location in the code.
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation represents the physical file location.
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

// SARIFArtifactLocation represents the artifact (file) location.
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFRegion represents a region within a file.
type SARIFRegion struct {
	LineRangeMixin
}

// SARIFFingerprints represents fingerprints for deduplication.
type SARIFFingerprints struct {
	CloneHash string `json:"cloneHash,omitempty"`
}

// SARIFInvocation represents an invocation of the tool.
type SARIFInvocation struct {
	ExecutionSuccessful bool   `json:"executionSuccessful"`
	StartTimeUTC        string `json:"startTimeUtc,omitempty"`
	EndTimeUTC          string `json:"endTimeUtc,omitempty"`
}

type sarifPrinter struct {
	ReadFile

	w               io.Writer
	threshold       int
	results         []SARIFResult
	processedHashes map[string]bool // Track processed hashes to avoid duplicates
	startTime       time.Time
}

// NewSARIF creates a new SARIF format printer.
func NewSARIF(w io.Writer, fread ReadFile, threshold int) Printer {
	return &sarifPrinter{
		w:               w,
		ReadFile:        fread,
		threshold:       threshold,
		results:         []SARIFResult{},
		processedHashes: make(map[string]bool),
		startTime:       time.Now(),
	}
}

func (p *sarifPrinter) PrintHeader() error {
	return nil
}

func (p *sarifPrinter) PrintClones(dups [][]*syntax.Node, sortBy ...config.SortCriteria) error {
	if len(dups) == 0 {
		return nil
	}

	// Generate a hash for this clone group
	hash := generateCloneHash(dups)

	// Skip if we've already processed this hash
	if p.processedHashes[hash] {
		return nil
	}

	p.processedHashes[hash] = true

	// Calculate size (token count)
	size := 0
	if len(dups) > 0 && len(dups[0]) > 0 {
		size = len(dups[0])
	}

	// Create a result for each clone instance
	for _, dup := range dups {
		if len(dup) == 0 {
			continue
		}

		nstart := dup[0]
		nend := dup[len(dup)-1]

		// Get file and line information
		fileInfo, err := ProcessNodeRange(p.ReadFile, nstart, nend)
		if err != nil {
			// Log warning but continue
			continue
		}

		level := p.determineLevel(size)
		msg := fmt.Sprintf("Duplicate code: %d tokens in %d instances",
			size, len(dups))

		result := SARIFResult{
			RuleID: "art-dupl/duplicate-code",
			Level:  level,
			Message: SARIFMessage{
				Text: msg,
			},
			Locations: []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							URI: nstart.Filename,
						},
						Region: SARIFRegion{
							LineRangeMixin: LineRangeMixin{
								StartLine: fileInfo.LineStart,
								EndLine:   fileInfo.LineEnd,
							},
						},
					},
				},
			},
			Fingerprints: SARIFFingerprints{
				CloneHash: hash,
			},
		}

		p.results = append(p.results, result)
	}

	return nil
}

func (p *sarifPrinter) PrintFooter() error {
	return p.outputSARIF()
}

// determineLevel maps clone size to SARIF level.
func (p *sarifPrinter) determineLevel(size int) string {
	switch {
	case size >= p.threshold*4:
		return "error" // Large clones are errors
	case size >= p.threshold*2:
		return "warning" // Medium clones are warnings
	default:
		return "note" // Small clones are notes
	}
}

// generateCloneHash generates a hash for a clone group.
func generateCloneHash(dups [][]*syntax.Node) string {
	if len(dups) == 0 || len(dups[0]) == 0 {
		return ""
	}

	// Use filename and position of first clone as hash basis
	first := dups[0][0]

	return fmt.Sprintf("%s:%d:%d", first.Filename, first.Pos, first.End)
}

// outputSARIF generates and writes the SARIF output.
func (p *sarifPrinter) outputSARIF() error {
	output := SARIFOutput{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "art-dupl",
						Version:        "1.0.0",
						InformationURI: "https://github.com/LarsArtmann/art-dupl",
						Rules: []SARIFRule{
							{
								ID:   "art-dupl/duplicate-code",
								Name: "Duplicate Code Detection",
								ShortDescription: SARIFTextContent{
									Text: "Detects duplicate code fragments in source files",
								},
								FullDescription: SARIFTextContent{
									Text: "This rule identifies code duplication by analyzing abstract syntax trees (ASTs) and finding structural similarities between code fragments. Duplicated code increases maintenance burden and can lead to inconsistent bug fixes.",
								},
								DefaultConfiguration: SARIFConfiguration{
									Level: "warning",
								},
								HelpURI: "https://github.com/LarsArtmann/art-dupl#duplicate-code-detection",
							},
						},
					},
				},
				Results: p.results,
				Invocations: []SARIFInvocation{
					{
						ExecutionSuccessful: true,
						StartTimeUTC:        p.startTime.UTC().Format(time.RFC3339),
						EndTimeUTC:          time.Now().UTC().Format(time.RFC3339),
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(&output, "", "  ")
	if err != nil {
		return errors.HandleMarshalingError(
			"encode",
			"SARIF output",
			err,
		)
	}

	return writeFormattedOutput(p.w, data, "SARIF output")
}
