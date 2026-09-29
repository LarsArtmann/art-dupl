package printer

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/jsonutil"
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
	Hash     string      `json:"hash"`
	AnchorID string      `json:"anchor_id"`
	Size     int         `json:"size"`
	Clones   []JSONClone `json:"files"`
}

type JSONClone struct {
	domain.CloneRef

	Category             domain.CloneCategory      `json:"category,omitzero"`
	Priority             domain.ClonePriority      `json:"priority,omitzero"`
	Actionability        domain.CloneActionability `json:"actionability,omitzero"`
	NonActionablePattern string                    `json:"non_actionable_pattern,omitempty"`
	CloneType            domain.CloneType          `json:"clone_type,omitzero"`
	LinesSaved           int                       `json:"lines_saved"`
	Extractable          bool                      `json:"extractable"`
	Confidence           float64                   `json:"confidence"`
	GenericsCandidate    bool                      `json:"generics_candidate"`
	GenericsHint         string                    `json:"generics_hint,omitempty"`
	Explanation          *Explanation              `json:"explanation,omitempty"`
}

// Explanation is the structured form of the text printer's `explain:` line.
// Present on every clone only when --explain is set. Field-for-field it
// mirrors TextPrinter.writeExplanation: clone type, actionability verdict
// (with the boilerplate-pattern label when suppressed), category, token/line
// counts, the extractability estimate, and the fix suggestion.
type Explanation struct {
	CloneType      domain.CloneType          `json:"clone_type"`
	Actionability  domain.CloneActionability `json:"actionability"`
	Pattern        string                    `json:"pattern,omitempty"`
	Category       domain.CloneCategory      `json:"category"`
	Tokens         int                       `json:"tokens"`
	Lines          int                       `json:"lines"`
	Extractability string                    `json:"extractability,omitempty"`
	Suggestion     string                    `json:"suggestion,omitempty"`
}

type Summary struct {
	TotalCloneGroups int     `json:"total_clone_groups"`
	TotalClones      int     `json:"total_clones"`
	ComplexityScore  float64 `json:"complexity_score"`
	ImpactScore      int     `json:"impact_score"`
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
		GenericsCandidate:    cl.Classification.GenericsCandidate,
		GenericsHint:         cl.Classification.GenericsHint,
	}

	if cl.Classification.Analysis != nil {
		clone.Confidence = cl.Classification.Analysis.Confidence
	}

	return clone
}

// buildExplanation mirrors TextPrinter.writeExplanation field-for-field so
// the JSON consumer gets the same "why" the text user sees, including the
// extractability estimate phrasing.
func buildExplanation(cl domain.ProcessedClone) *Explanation {
	cls := cl.Classification
	e := &Explanation{
		CloneType:     cls.CloneType,
		Actionability: cls.Actionability,
		Pattern:       cls.NonActionablePattern,
		Category:      cls.Category,
		Tokens:        cls.Tokens,
		Lines:         cls.Lines,
		Suggestion:    cls.Suggestion,
	}
	if cls.Extractability.CanExtract {
		e.Extractability = fmt.Sprintf("~%d lines saved across sites",
			cls.Extractability.EstimatedLinesSaved)
	}

	return e
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
	explain     bool
}

// SetExplain attaches a structured explanation object to every clone when
// --explain is set (JSONPrinter satisfies printer.ExplainSetter; cmd wires
// it via the same type assertion as the text printer).
func (p *JSONPrinter) SetExplain(enabled bool) {
	p.explain = enabled
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
		jc := toJSONClone(cl)
		if p.explain {
			jc.Explanation = buildExplanation(cl)
		}
		jsonClones = append(jsonClones, jc)
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

	// Anchors are derived AFTER the final sort so a JSON record deep-links to
	// the same group-<id> target as the HTML report (sanitized hash, with the
	// 1-based display position as the empty-hash fallback).
	for i := range p.cloneGroups {
		p.cloneGroups[i].AnchorID = groupAnchorID(p.cloneGroups[i].Hash, i+1)
	}

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

	data, err := jsonutil.MarshalIndent(&output, "", "  ")
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

	data, err := jsonutil.MarshalIndent(simpleOutput, "", "  ")
	if err != nil {
		return errors.HandleMarshalingError("encode", "simple JSON output", err)
	}

	return writeFormattedOutput(p.w, data, "simple JSON output")
}
