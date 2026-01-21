package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// JSONOutput represents the structured JSON output.
type JSONOutput struct {
	Version         string       `json:"version"`
	Timestamp       time.Time    `json:"timestamp"`
	Threshold       int          `json:"threshold"`
	FilesAnalyzed   int          `json:"files_analyzed"`
	DetectionMethod string       `json:"detection_method,omitempty"`
	CloneGroups     []CloneGroup `json:"clone_groups"`
	Summary         Summary      `json:"summary"`
}

// CloneGroup represents a group of duplicate code fragments.
type CloneGroup struct {
	Hash  string      `json:"hash"`
	Size  int         `json:"size"`
	Files []JSONClone `json:"files"`
}

// JSONClone represents a single code fragment duplicate for JSON output.
type JSONClone struct {
	Filename  string `json:"filename"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
	Fragment  string `json:"fragment"`
}

// Summary provides analysis summary statistics.
type Summary struct {
	TotalCloneGroups int     `json:"total_clone_groups"`
	TotalClones      int     `json:"total_clones"`
	ComplexityScore  float64 `json:"complexity_score"`
	// ImpactScore represents total duplicated code volume (tokens × instances)
	// This is the simple scoring metric from the duplicates project
	ImpactScore int `json:"impact_score,omitempty"`
}

// SimpleJSONClone represents a single code clone instance in simple format (from duplicates project).
type SimpleJSONClone struct {
	Filename   string `json:"filename"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	TokenCount int    `json:"token_count"`
}

// SimpleCloneGroup represents a clone group in simple format (from duplicates project).
type SimpleCloneGroup struct {
	Hash      string            `json:"hash"`
	Score     int               `json:"score"` // Impact score: tokens × instances
	Instances []SimpleJSONClone `json:"instances"`
}

// SimpleJSONOutput represents the simple JSON output format (from duplicates project).
type SimpleJSONOutput []SimpleCloneGroup

type JSONPrinter struct {
	ReadFile

	iota        int
	w           io.Writer
	filesCount  int
	totalClones int
	cloneGroups []CloneGroup
	currentHash string // Hash for the current clone group
}

//nolint:ireturn // Printer interface is appropriate return type for factory function
func NewJSON(w io.Writer, fread ReadFile) Printer {
	return &JSONPrinter{
		w:        w,
		ReadFile: fread,
	}
}

// countLinesInFragment counts actual lines in a fragment by counting newline characters.
func countLinesInFragment(fragment string) int {
	if fragment == "" {
		return 1 // At least one line even for empty content
	}
	return strings.Count(fragment, "\n") + 1
}

func (p *JSONPrinter) PrintHeader() error {
	p.iota = 0
	// Don't reset filesCount - it's set once and should persist
	// p.filesCount = 0
	p.totalClones = 0
	p.cloneGroups = []CloneGroup{}
	return nil
}

// SetHash sets the current hash for the clone group being processed.
func (p *JSONPrinter) SetHash(hash string) {
	p.currentHash = hash
}

// SetFilesCount sets the total number of files analyzed.
func (p *JSONPrinter) SetFilesCount(count int) {
	p.filesCount = count
}

func (p *JSONPrinter) PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error {
	p.iota++

	clones := make([]JSONClone, len(dups))
	for i, dup := range dups {
		cnt := len(dup)
		if cnt == 0 {
			return errors.NewInternalError("zero length duplicate found", nil)
		}
		nstart := dup[0]
		nend := dup[cnt-1]

		// Use unified file processor
		fileInfo, err := ProcessNodeRange(p.ReadFile, nstart, nend)
		if err != nil {
			return fmt.Errorf("failed to process node range for file %s (clone %d of %d): %w", nstart.Filename, i+1, len(dups), err)
		}

		lineStart := fileInfo.LineStart
		content := extractContent(fileInfo, nstart, nend)
		clones[i] = JSONClone{
			Filename:  nstart.Filename,
			LineStart: lineStart,
			LineEnd:   0, // Will be calculated later
			Fragment:  string(deindent(content)),
		}
	}

	sort.Slice(clones, func(i, j int) bool {
		if clones[i].Filename == clones[j].Filename {
			return clones[i].LineStart < clones[j].LineStart
		}
		return clones[i].Filename < clones[j].Filename
	})

	// Calculate line ends - count actual lines, not characters
	for i := range clones {
		lines := countLinesInFragment(clones[i].Fragment)
		clones[i].LineEnd = clones[i].LineStart + lines - 1
	}

	// Calculate hash (use actual hash instead of counter)
	hash := p.currentHash

	// Calculate size (use actual token count)
	size := 0
	for _, dup := range dups {
		// Each dup is a sequence of []*Node, where each node is a token
		size += len(dup)
	}

	cloneGroup := CloneGroup{
		Hash:  hash,
		Size:  size,
		Files: clones,
	}

	p.cloneGroups = append(p.cloneGroups, cloneGroup)
	p.totalClones += len(clones)

	return nil
}

func (*JSONPrinter) PrintFooter() error {
	// This would be called at the end, but we need to access printer state
	// We'll handle JSON output in a separate flush method
	return nil
}

// OutputJSON generates the complete JSON output.
func (p *JSONPrinter) OutputJSON(threshold int, sortBy SortBy, detectionMethod string) error {
	// Sort clone groups before generating JSON
	SortCloneGroups(p.cloneGroups, sortBy)

	output := JSONOutput{
		Version:         "1.0",
		Timestamp:       time.Now().UTC(),
		Threshold:       threshold,
		FilesAnalyzed:   p.filesCount,
		DetectionMethod: detectionMethod,
		CloneGroups:     p.cloneGroups,
		Summary: Summary{
			TotalCloneGroups: len(p.cloneGroups),
			TotalClones:      p.totalClones,
			ComplexityScore:  float64(p.totalClones) / float64(len(p.cloneGroups)+1),
		},
	}

	encoder := json.NewEncoder(p.w)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(&output)
	if err != nil {
		return errors.HandleMarshalingError("encode", "JSON output", err) //nolint:wrapcheck // Error already wraps cause
	}
	return nil
}

// OutputSimpleJSON generates simple JSON output format (from duplicates project).
// This provides a simpler, more straightforward JSON format for users who prefer it.
func (p *JSONPrinter) OutputSimpleJSON() error {
	simpleOutput := make(SimpleJSONOutput, len(p.cloneGroups))

	for i, group := range p.cloneGroups {
		// Calculate impact score: tokens × instances
		impactScore := group.Size * len(group.Files)

		// Convert to simple format
		simpleInstances := make([]SimpleJSONClone, len(group.Files))
		for j, file := range group.Files {
			simpleInstances[j] = SimpleJSONClone{
				Filename:   file.Filename,
				StartLine:  file.LineStart,
				EndLine:    file.LineEnd,
				TokenCount: group.Size, // Each clone in group has same size
			}
		}

		simpleOutput[i] = SimpleCloneGroup{
			Hash:      group.Hash,
			Score:     impactScore,
			Instances: simpleInstances,
		}
	}

	encoder := json.NewEncoder(p.w)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(simpleOutput)
	if err != nil {
		return errors.HandleMarshalingError("encode", "simple JSON output", err) //nolint:wrapcheck // Error already wraps cause
	}
	return nil
}
