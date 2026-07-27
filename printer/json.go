package printer

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
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
	domain.CloneRef

	Category             domain.CloneCategory      `json:"category,omitzero"`
	Priority             domain.ClonePriority      `json:"priority,omitzero"`
	Actionability        domain.CloneActionability `json:"actionability,omitzero"`
	NonActionablePattern string                    `json:"non_actionable_pattern,omitempty"`
	CloneType            domain.CloneType          `json:"clone_type,omitzero"`
	LinesSaved           int                       `json:"lines_saved,omitempty"`
	Extractable          bool                      `json:"extractable,omitempty"`
	Confidence           float64                   `json:"confidence,omitempty"`
}

type Summary struct {
	TotalCloneGroups int     `json:"total_clone_groups"`
	TotalClones      int     `json:"total_clones"`
	ComplexityScore  float64 `json:"complexity_score"`
	ImpactScore      int     `json:"impact_score,omitempty"`
}

// toJSONClone converts a domain.ProcessedClone to a JSONClone DTO.
// This is the single conversion point — all JSON output paths use it.
func toJSONClone(cl domain.ProcessedClone) JSONClone {
	clone := JSONClone{
		CloneRef:             cl.CloneRef,
		Category:             cl.Classification.Category,
		Priority:             cl.Classification.Priority,
		Actionability:        cl.Classification.Actionability,
		NonActionablePattern: cl.Classification.NonActionablePattern,
		CloneType:            cl.Classification.CloneType,
		LinesSaved:           cl.Classification.Extractability.EstimatedLinesSaved,
		Extractable:          cl.Classification.Extractability.CanExtract,
	}

	if cl.Classification.Analysis != nil {
		clone.Confidence = cl.Classification.Analysis.Confidence
	}

	return clone
}

type simpleJSONClone struct {
	domain.CloneRef

	TokenCount int `json:"token_count"`
}

type simpleCloneGroup struct {
	Hash   string            `json:"hash"`
	Size   int               `json:"score"`
	Clones []simpleJSONClone `json:"instances"`
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
		jsonClones = append(jsonClones, toJSONClone(cl))
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

	data, err := json.Marshal(&output, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
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
				CloneRef:   file.CloneRef,
				TokenCount: group.Size,
			})
		}

		simpleOutput = append(simpleOutput, simpleCloneGroup{
			Hash:   group.Hash,
			Size:   impactScore,
			Clones: simpleInstances,
		})
	}

	data, err := json.Marshal(simpleOutput, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return errors.HandleMarshalingError("encode", "simple JSON output", err)
	}

	return writeFormattedOutput(p.w, data, "simple JSON output")
}
