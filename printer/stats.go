package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// stats provides aggregated statistics about code duplication.
type stats struct {
	ReadFile
	w         io.Writer
	threshold int
	format    Format
	statsData *StatsData
}

// StatsData holds all aggregated statistics.
type StatsData struct {
	TotalFilesScanned    int
	TotalCloneGroups     int
	TotalClones          int
	TotalDuplicateLines  int
	TotalTokens          int
	TotalEstimatedLines  int // Estimated total lines for duplication percentage
	AverageCloneSize     int
	ComplexityScore      float64
	ImpactScore         int
	DuplicationRatio   float64 // Percentage of duplicated code
	AnalysisDuration   string   // Time taken for analysis
	HealthScore        string   // A-F grade based on metrics
	FileDuplication      map[string]int // filename -> duplicate line count
	SizeDistribution     map[string]int // size range -> count
	DetectionMethods     string
	Timestamp          string // ISO 8601 timestamp
}

// NewStats creates a new stats printer.
func NewStats(w io.Writer, fread ReadFile, threshold int) Printer {
	return &stats{
		w:         w,
		ReadFile:  fread,
		threshold: threshold,
		statsData: &StatsData{
			FileDuplication:  make(map[string]int),
			SizeDistribution: make(map[string]int),
		},
	}
}

// SetFilesCount sets the total number of files scanned.
func (p *stats) SetFilesCount(count int) {
	p.statsData.TotalFilesScanned = count
}

// SetAnalysisDuration sets the analysis duration.
func (p *stats) SetAnalysisDuration(duration time.Duration) {
	p.statsData.AnalysisDuration = duration.String()
}

// SetTotalEstimatedLines sets the estimated total lines of code.
func (p *stats) SetTotalEstimatedLines(lines int) {
	p.statsData.TotalEstimatedLines = lines
}

// PrintHeader prints the stats header.
func (p *stats) PrintHeader() error {
	return nil
}

// PrintClones collects statistics from the clone groups.
func (p *stats) PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error {
	// Count clone group
	p.statsData.TotalCloneGroups++

	// Count tokens in this clone group
	tokensInGroup := 0

	// Process each clone in the group
	for _, dup := range dups {
		if len(dup) == 0 {
			continue
		}

		// Get file and line information using shared processor
		nstart := dup[0]
		nend := dup[len(dup)-1]

		fileInfo, err := ProcessNodeRange(p.ReadFile, nstart, nend)
		if err != nil {
			return fmt.Errorf("failed to process node range for file %s: %w", nstart.Filename, err)
		}

		lineCount := fileInfo.LineEnd - fileInfo.LineStart + 1
		tokensInGroup += len(dup)

		// Update statistics
		p.statsData.TotalClones++
		p.statsData.TotalDuplicateLines += lineCount
		p.statsData.TotalTokens += len(dup)

		// Track file-level duplication
		p.statsData.FileDuplication[nstart.Filename] += lineCount

		// Track size distribution
		sizeRange := p.getSizeRange(lineCount)
		p.statsData.SizeDistribution[sizeRange]++
	}

	// Update impact score (tokens × instances)
	p.statsData.ImpactScore += tokensInGroup * len(dups)

	return nil
}

// PrintFooter prints the aggregated statistics.
func (p *stats) PrintFooter() error {
	// Calculate average clone size
	if p.statsData.TotalClones > 0 {
		p.statsData.AverageCloneSize = p.statsData.TotalDuplicateLines / p.statsData.TotalClones
	}

	// Calculate complexity score (clones per group)
	if p.statsData.TotalCloneGroups > 0 {
		p.statsData.ComplexityScore = float64(p.statsData.TotalClones) / float64(p.statsData.TotalCloneGroups)
	}

	// Print statistics
	p.printStats()

	return nil
}

// printStats prints the collected statistics.
func (p *stats) printStats() {
	switch p.format {
	case FormatJSON:
		p.printJSON()
	case FormatCSV:
		p.printCSV()
	case FormatText:
		p.printText()
	}
}

// printCSV prints statistics in CSV format (not yet implemented).
func (p *stats) printCSV() {
	fmt.Fprintf(p.w, "CSV format is not yet implemented. Use --format text or --format json.\n")
}

// printText prints statistics in text format.
func (p *stats) printText() {
	fmt.Fprintf(p.w, "Code Duplication Statistics\n")
	fmt.Fprintf(p.w, "============================\n\n")

	fmt.Fprintf(p.w, "Configuration:\n")
	fmt.Fprintf(p.w, "  Threshold: %d tokens\n", p.threshold)
	fmt.Fprintf(p.w, "  Detection Methods: %s\n", p.statsData.DetectionMethods)
	fmt.Fprintf(p.w, "\n")

	fmt.Fprintf(p.w, "Overview:\n")
	fmt.Fprintf(p.w, "  Files Scanned: %d\n", p.statsData.TotalFilesScanned)
	fmt.Fprintf(p.w, "  Clone Groups: %d\n", p.statsData.TotalCloneGroups)
	fmt.Fprintf(p.w, "  Total Clones: %d\n", p.statsData.TotalClones)
	fmt.Fprintf(p.w, "\n")

	fmt.Fprintf(p.w, "Duplicate Code:\n")
	fmt.Fprintf(p.w, "  Total Duplicate Lines: %d\n", p.statsData.TotalDuplicateLines)
	fmt.Fprintf(p.w, "  Total Duplicate Tokens: %d\n", p.statsData.TotalTokens)
	fmt.Fprintf(p.w, "  Average Clone Size: %d lines\n", p.statsData.AverageCloneSize)
	fmt.Fprintf(p.w, "  Complexity Score: %.2f\n", p.statsData.ComplexityScore)
	fmt.Fprintf(p.w, "  Impact Score: %d\n", p.statsData.ImpactScore)
	fmt.Fprintf(p.w, "\n")

	// Print size distribution
	if len(p.statsData.SizeDistribution) > 0 {
		fmt.Fprintf(p.w, "Clone Size Distribution:\n")
		printSizeDistribution(p.w, p.statsData.SizeDistribution)
		fmt.Fprintf(p.w, "\n")
	}

	// Print top files with most duplicates
	if len(p.statsData.FileDuplication) > 0 {
		fmt.Fprintf(p.w, "Top Files by Duplicate Lines:\n")
		printTopFiles(p.w, p.statsData.FileDuplication, 10)
	}
}

// printJSON prints statistics in JSON format.
func (p *stats) printJSON() {
	// Create a struct for JSON output
	jsonData := struct {
		Configuration struct {
			Threshold         int    `json:"threshold"`
			DetectionMethods  string `json:"detectionMethods"`
		} `json:"configuration"`
		Overview struct {
			FilesScanned    int `json:"filesScanned"`
			CloneGroups     int `json:"cloneGroups"`
			TotalClones     int `json:"totalClones"`
		} `json:"overview"`
		DuplicateCode struct {
			TotalLines       int     `json:"totalDuplicateLines"`
			TotalTokens      int     `json:"totalDuplicateTokens"`
			AverageCloneSize int     `json:"averageCloneSize"`
			ComplexityScore  float64 `json:"complexityScore"`
			ImpactScore      int     `json:"impactScore"`
		} `json:"duplicateCode"`
		SizeDistribution map[string]int `json:"sizeDistribution"`
		TopFiles         []struct {
			Filename string `json:"filename"`
			Lines    int    `json:"duplicateLines"`
		} `json:"topFiles"`
	}{}

	// Fill configuration
	jsonData.Configuration.Threshold = p.threshold
	jsonData.Configuration.DetectionMethods = p.statsData.DetectionMethods

	// Fill overview
	jsonData.Overview.FilesScanned = p.statsData.TotalFilesScanned
	jsonData.Overview.CloneGroups = p.statsData.TotalCloneGroups
	jsonData.Overview.TotalClones = p.statsData.TotalClones

	// Fill duplicate code metrics
	jsonData.DuplicateCode.TotalLines = p.statsData.TotalDuplicateLines
	jsonData.DuplicateCode.TotalTokens = p.statsData.TotalTokens
	jsonData.DuplicateCode.AverageCloneSize = p.statsData.AverageCloneSize
	jsonData.DuplicateCode.ComplexityScore = p.statsData.ComplexityScore
	jsonData.DuplicateCode.ImpactScore = p.statsData.ImpactScore

	// Fill size distribution
	jsonData.SizeDistribution = p.statsData.SizeDistribution

	// Fill top files
	if len(p.statsData.FileDuplication) > 0 {
		// Convert map to slice and sort by duplicate lines (descending)
		type fileStat struct {
			filename string
			lines    int
		}
		files := make([]fileStat, 0, len(p.statsData.FileDuplication))
		for filename, lines := range p.statsData.FileDuplication {
			files = append(files, fileStat{filename, lines})
		}

		// Sort by lines descending
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[i].lines < files[j].lines {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// Take top 10
		limit := len(files)
		if limit > 10 {
			limit = 10
		}
		jsonData.TopFiles = make([]struct {
			Filename string `json:"filename"`
			Lines    int    `json:"duplicateLines"`
		}, limit)
		for i := 0; i < limit; i++ {
			jsonData.TopFiles[i].Filename = files[i].filename
			jsonData.TopFiles[i].Lines = files[i].lines
		}
	}

	encoder := json.NewEncoder(p.w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonData); err != nil {
		// In a real implementation, we'd handle this error properly
		fmt.Fprintf(p.w, "Error encoding JSON: %v\n", err)
	}
}

// getSizeRange returns a human-readable size range for a line count.
func (p *stats) getSizeRange(lines int) string {
	switch {
	case lines <= 5:
		return "1-5 lines"
	case lines <= 10:
		return "6-10 lines"
	case lines <= 20:
		return "11-20 lines"
	case lines <= 50:
		return "21-50 lines"
	case lines <= 100:
		return "51-100 lines"
	default:
		return "100+ lines"
	}
}

// printSizeDistribution prints the size distribution.
func printSizeDistribution(w io.Writer, distribution map[string]int) {
	ranges := make([]string, 0, len(distribution))
	for r := range distribution {
		ranges = append(ranges, r)
	}
	sort.Strings(ranges)

	for _, r := range ranges {
		fmt.Fprintf(w, "  %s: %d clones\n", r, distribution[r])
	}
}

// printTopFiles prints the top N files with most duplicate lines.
func printTopFiles(w io.Writer, fileDuplication map[string]int, topN int) {
	// Convert to slice for sorting
	type fileStat struct {
		filename string
		lines    int
	}
	files := make([]fileStat, 0, len(fileDuplication))
	for filename, lines := range fileDuplication {
		files = append(files, fileStat{filename, lines})
	}

	// Sort by duplicate lines (descending)
	sort.Slice(files, func(i, j int) bool {
		return files[i].lines > files[j].lines
	})

	// Print top N
	limit := len(files)
	if limit > topN {
		limit = topN
	}

	for i := 0; i < limit; i++ {
		fmt.Fprintf(w, "  %d lines in %s\n", files[i].lines, files[i].filename)
	}

	if len(files) > topN {
		fmt.Fprintf(w, "  ... and %d more files\n", len(files)-topN)
	}
}

// GetStatsData returns the collected statistics data.
func (p *stats) GetStatsData() *StatsData {
	return p.statsData
}

// SetDetectionMethods sets the detection methods used.
func (p *stats) SetDetectionMethods(methods string) {
	p.statsData.DetectionMethods = methods
}

// SetFormat sets the output format.
func (p *stats) SetFormat(format Format) {
	p.format = format
}
