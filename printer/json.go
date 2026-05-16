package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	errors "github.com/LarsArtmann/art-dupl/errors"
)

type JSONOutput struct {
	Version         string       `json:"version"`
	Timestamp       time.Time    `json:"timestamp"`
	Threshold       int          `json:"threshold"`
	FilesAnalyzed   int          `json:"files_analyzed"`
	DetectionMethod string       `json:"detection_method,omitempty"`
	CloneGroups     []CloneGroup `json:"clone_groups"`
	Summary         Summary      `json:"summary"`
}

type CloneGroup struct {
	Hash  string      `json:"hash"`
	Size  int         `json:"size"`
	Files []JSONClone `json:"files"`
}

type JSONClone struct {
	Filename      string `json:"filename"`
	LineStart     int    `json:"line_start"`
	LineEnd       int    `json:"line_end"`
	Fragment      string `json:"fragment"`
	Category      string `json:"category,omitempty"`
	Priority      string `json:"priority,omitempty"`
	Actionability string `json:"actionability,omitempty"`
}

type Summary struct {
	TotalCloneGroups int     `json:"total_clone_groups"`
	TotalClones      int     `json:"total_clones"`
	ComplexityScore  float64 `json:"complexity_score"`
	ImpactScore      int     `json:"impact_score,omitempty"`
}

type SimpleJSONClone struct {
	LineRangeMixin

	Filename   string `json:"filename"`
	TokenCount int    `json:"token_count"`
}

type SimpleCloneGroup struct {
	Hash      string            `json:"hash"`
	Score     int               `json:"score"`
	Instances []SimpleJSONClone `json:"instances"`
}

type LineRangeMixin struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine,omitempty"`
}

type SimpleJSONOutput []SimpleCloneGroup

type JSONPrinter struct {
	ReadFile

	iota        int
	w           io.Writer
	filesCount  int
	totalClones int
	cloneGroups []CloneGroup
	currentHash string
}

func NewJSON(w io.Writer, fread ReadFile) Printer {
	return &JSONPrinter{
		w:        w,
		ReadFile: fread,
	}
}

func (p *JSONPrinter) PrintHeader() error {
	p.iota = 0
	p.totalClones = 0
	p.cloneGroups = []CloneGroup{}

	return nil
}

func (p *JSONPrinter) SetHash(hash string) {
	p.currentHash = hash
}

func (p *JSONPrinter) SetFilesCount(count int) {
	p.filesCount = count
}

func (p *JSONPrinter) PrintClones(
	group domain.ProcessedCloneGroup,
	sortBy ...config.SortCriteria,
) error {
	p.iota++

	clones := group.Clones

	jsonClones := make([]JSONClone, len(clones))
	for i, cl := range clones {
		jsonClones[i] = JSONClone{
			Filename:      cl.Filename,
			LineStart:     cl.LineStart,
			LineEnd:       cl.LineEnd,
			Fragment:      string(deindent(cl.Fragment)),
			Category:      string(cl.Classification.Category),
			Priority:      string(cl.Classification.Priority),
			Actionability: string(cl.Classification.Actionability),
		}
	}

	sort.Slice(jsonClones, func(i, j int) bool {
		if jsonClones[i].Filename == jsonClones[j].Filename {
			return jsonClones[i].LineStart < jsonClones[j].LineStart
		}

		return jsonClones[i].Filename < jsonClones[j].Filename
	})

	size := 0
	for _, cl := range clones {
		size += cl.Size
	}

	cloneGroup := CloneGroup{
		Hash:  p.currentHash,
		Size:  size,
		Files: jsonClones,
	}

	p.cloneGroups = append(p.cloneGroups, cloneGroup)
	p.totalClones += len(jsonClones)

	return nil
}

func (*JSONPrinter) PrintFooter() error {
	return nil
}

func (p *JSONPrinter) OutputJSON(
	threshold int,
	sortBy config.SortCriteria,
	detectionMethod string,
) error {
	SortCloneGroups(p.cloneGroups, sortBy)

	output := JSONOutput{
		Version:       "1.0",
		Timestamp:     time.Now().UTC(),
		Threshold:     threshold,
		FilesAnalyzed: p.filesCount,
		CloneGroups:   p.cloneGroups,
		Summary: Summary{
			TotalCloneGroups: len(p.cloneGroups),
			TotalClones:      p.totalClones,
			ComplexityScore:  float64(p.totalClones) / float64(len(p.cloneGroups)+1),
		},
	}

	output.DetectionMethod = detectionMethod

	data, err := json.MarshalIndent(&output, "", "  ")
	if err != nil {
		return fmt.Errorf(
			"encode JSON output (threshold: %d, sortBy: %s, detection: %s): %w",
			threshold,
			sortBy.String(),
			detectionMethod,
			errors.HandleMarshalingError("encode", "JSON output", err),
		)
	}

	return writeFormattedOutput(p.w, data, "JSON output")
}

func (p *JSONPrinter) OutputSimpleJSON() error {
	simpleOutput := make(SimpleJSONOutput, len(p.cloneGroups))

	for i, group := range p.cloneGroups {
		impactScore := group.Size * len(group.Files)

		simpleInstances := make([]SimpleJSONClone, len(group.Files))
		for j, file := range group.Files {
			simpleInstances[j] = SimpleJSONClone{
				LineRangeMixin: LineRangeMixin{
					StartLine: file.LineStart,
					EndLine:   file.LineEnd,
				},
				Filename:   file.Filename,
				TokenCount: group.Size,
			}
		}

		simpleOutput[i] = SimpleCloneGroup{
			Hash:      group.Hash,
			Score:     impactScore,
			Instances: simpleInstances,
		}
	}

	data, err := json.MarshalIndent(simpleOutput, "", "  ")
	if err != nil {
		return errors.HandleMarshalingError("encode", "simple JSON output", err)
	}

	return writeFormattedOutput(p.w, data, "simple JSON output")
}
