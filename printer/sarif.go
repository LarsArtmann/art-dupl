package printer

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	errors "github.com/LarsArtmann/art-dupl/errors"
)

// SARIF level constants.
const (
	sarifLevelError   = "error"
	sarifLevelWarning = "warning"
	sarifLevelNote    = "note"
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
	ID                   string              `json:"id"`
	Name                 string              `json:"name"`
	ShortDescription     SARIFTextContent    `json:"shortDescription"`
	FullDescription      SARIFTextContent    `json:"fullDescription"`
	DefaultConfiguration SARIFConfiguration  `json:"defaultConfiguration"`
	HelpURI              string              `json:"helpUri,omitempty"`
	Properties           SARIFRuleProperties `json:"properties"`
}

// SARIFRuleProperties carries tool-specific metadata recognised by
// GitHub Code Scanning, SonarQube, and other SARIF consumers.
type SARIFRuleProperties struct {
	Precision       string   `json:"precision,omitempty"`
	ProblemSeverity string   `json:"problem.severity,omitempty"`
	Tags            []string `json:"tags,omitempty"`
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
	Properties   map[string]string `json:"properties,omitempty"`
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
// Field names follow the SARIF 2.1.0 spec (region.startLine / region.endLine);
// they intentionally differ from CloneRef's line_start/line_end tags.
type SARIFRegion struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine,omitempty"`
}

// SARIFFingerprints represents fingerprints for deduplication.
// SARIF spec: fingerprints are short strings that can be used to relate related results.
type SARIFFingerprints struct {
	// ContentFingerprint is a hash that uniquely identifies the duplicate content.
	ContentFingerprint string `json:"contentFingerprint,omitempty"`
	// PartialFingerprint provides approximate matching for related content.
	PartialFingerprint string `json:"partialFingerprint,omitempty"`
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
	currentHash     string // Hash for the current clone group
	version         string // Tool version for SARIF output
}

// NewSARIF creates a new SARIF format printer with default settings.
func NewSARIF(w io.Writer, fread ReadFile, threshold int) Printer {
	return NewSARIFWithConfig(w, fread, SARIFPrinterOptions{
		Threshold: threshold,
		Version:   "dev",
	})
}

// SARIFPrinterOptions contains configuration for SARIF output.
type SARIFPrinterOptions struct {
	Threshold int
	Version   string
}

// NewSARIFWithConfig creates a new SARIF format printer with explicit config.
func NewSARIFWithConfig(w io.Writer, fread ReadFile, cfg SARIFPrinterOptions) Printer {
	return &sarifPrinter{
		w:               w,
		ReadFile:        fread,
		threshold:       cfg.Threshold,
		results:         []SARIFResult{},
		processedHashes: make(map[string]bool),
		startTime:       time.Now(),
		version:         cfg.Version,
	}
}

func (p *sarifPrinter) PrintHeader() error {
	return nil
}

func (p *sarifPrinter) PrintClones(
	group domain.ProcessedCloneGroup,
	sortBy ...config.SortCriteria,
) error {
	if len(group.Clones) == 0 {
		return nil
	}

	hash := p.currentHash
	if hash == "" {
		return nil
	}

	if p.processedHashes[hash] {
		return nil
	}

	p.processedHashes[hash] = true

	size := group.TotalTokenCount()

	for _, cl := range group.Clones {
		level := p.determineLevel(size)
		msg := fmt.Sprintf("Duplicate code: %d tokens in %d instances",
			size, len(group.Clones))

		properties := map[string]string{
			"clone_type": string(cl.Classification.CloneType),
			"category":   string(cl.Classification.Category),
		}

		if cl.Classification.NonActionablePattern != "" {
			properties["non_actionable_pattern"] = cl.Classification.NonActionablePattern
		}

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
							URI: cl.Filename,
						},
						Region: SARIFRegion{
							StartLine: cl.LineStart,
							EndLine:   cl.LineEnd,
						},
					},
				},
			},
			Fingerprints: SARIFFingerprints{
				ContentFingerprint: hash,
				PartialFingerprint: hash[:min(8, len(hash))],
			},
			Properties: properties,
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
		return sarifLevelError
	case size >= p.threshold*2:
		return sarifLevelWarning
	default:
		return sarifLevelNote
	}
}

// SetHash sets the hash for the current clone group.
func (p *sarifPrinter) SetHash(hash string) {
	p.currentHash = hash
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
						Version:        p.version,
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
								Properties: SARIFRuleProperties{
									Precision:       "high",
									ProblemSeverity: "warning",
									Tags:            []string{"maintainability", "duplicate-code", "design"},
								},
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

	data, err := json.Marshal(&output, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return errors.HandleMarshalingError(
			"encode",
			"SARIF output",
			err,
		)
	}

	return writeFormattedOutput(p.w, data, "SARIF output")
}
