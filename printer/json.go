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
	Hash   string      `json:"hash"`
	Size   int         `json:"size"`
	Clones []JSONClone `json:"files"`
}

type JSONClone struct {
	Filename      string                    `json:"filename"`
	LineStart     int                       `json:"line_start"`
	LineEnd       int                       `json:"line_end"`
	Fragment      string                    `json:"fragment"`
	Category      domain.CloneCategory      `json:"category,omitempty"`
	Priority      domain.ClonePriority      `json:"priority,omitempty"`
	Actionability domain.CloneActionability `json:"actionability,omitempty"`
	CloneType     domain.CloneType          `json:"clone_type,omitempty"`
	LinesSaved    int                       `json:"lines_saved,omitempty"`
	Extractable   bool                      `json:"extractable,omitempty"`
}

type Summary struct {
	TotalCloneGroups int     `json:"total_clone_groups"`
	TotalClones      int     `json:"total_clones"`
	ComplexityScore  float64 `json:"complexity_score"`
	ImpactScore      int     `json:"impact_score,omitempty"`
}

type simpleJSONClone struct {
	LineRangeMixin

	Filename   string `json:"filename"`
	TokenCount int    `json:"token_count"`
}

type simpleCloneGroup struct {
	Hash      string            `json:"hash"`
	Score     int               `json:"score"`
	Instances []simpleJSONClone `json:"instances"`
}

type LineRangeMixin struct {
	LineStart int `json:"line_start"`
	LineEnd   int `json:"line_end,omitempty"`
}

type simpleJSONOutput []simpleCloneGroup

type JSONPrinter struct {
	ReadFile

	cloneIndex  int
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
	p.cloneIndex = 0
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
	p.cloneIndex++

	clones := group.Clones

	jsonClones := make([]JSONClone, 0, len(clones))
	for _, cl := range clones {
		jsonClones = append(jsonClones, JSONClone{
			Filename:      cl.Filename,
			LineStart:     cl.LineStart,
			LineEnd:       cl.LineEnd,
			Fragment:      cl.Fragment,
			Category:      cl.Classification.Category,
			Priority:      cl.Classification.Priority,
			Actionability: cl.Classification.Actionability,
			CloneType:     cl.Classification.CloneType,
			LinesSaved:    cl.Classification.Extractability.EstimatedLinesSaved,
			Extractable:   cl.Classification.Extractability.CanExtract,
		})
	}

	sort.Slice(jsonClones, func(i, j int) bool {
		if jsonClones[i].Filename == jsonClones[j].Filename {
			return jsonClones[i].LineStart < jsonClones[j].LineStart
		}

		return jsonClones[i].Filename < jsonClones[j].Filename
	})

	size := 0
	for _, cl := range clones {
		size += cl.TokenCount
	}

	cloneGroup := CloneGroup{
		Hash:   p.currentHash,
		Size:   size,
		Clones: jsonClones,
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
	simpleOutput := make(simpleJSONOutput, 0, len(p.cloneGroups))

	for _, group := range p.cloneGroups {
		impactScore := group.Size * len(group.Clones)

		simpleInstances := make([]simpleJSONClone, 0, len(group.Clones))
		for _, file := range group.Clones {
			simpleInstances = append(simpleInstances, simpleJSONClone{
				LineRangeMixin: LineRangeMixin{
					LineStart: file.LineStart,
					LineEnd:   file.LineEnd,
				},
				Filename:   file.Filename,
				TokenCount: group.Size,
			})
		}

		simpleOutput = append(simpleOutput, simpleCloneGroup{
			Hash:      group.Hash,
			Score:     impactScore,
			Instances: simpleInstances,
		})
	}

	data, err := json.MarshalIndent(simpleOutput, "", "  ")
	if err != nil {
		return errors.HandleMarshalingError("encode", "simple JSON output", err)
	}

	return writeFormattedOutput(p.w, data, "simple JSON output")
}
