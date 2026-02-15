package adapter

import (
	"os"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// PrinterAdapter bridges domain types with existing printer interface.
type PrinterAdapter struct {
	// Existing printer will be wrapped
}

// NodeToDomainClone converts syntax node to domain clone.
func NodeToDomainClone(node *syntax.Node, filename string) domain.Clone {
	// Try to read file content
	var fileContent []byte
	if filename != "" {
		// #nosec G304 -- filename is controlled input from syntax tree, not user input
		if content, err := os.ReadFile(filename); err == nil {
			fileContent = content
		}
		// If file doesn't exist or can't be read, continue with empty content
	}
	return domain.NodeToClone(node, filename, fileContent)
}

// CloneGroupFromNodes creates domain clone group from syntax nodes.
func CloneGroupFromNodes(groupID string, nodes [][]*syntax.Node) domain.CloneGroup {
	var clones []domain.Clone
	var totalSize uint

	for _, nodeGroup := range nodes {
		if len(nodeGroup) == 0 {
			continue
		}

		// Calculate size based on first node
		startNode := nodeGroup[0]
		endNode := nodeGroup[len(nodeGroup)-1]
		size := uint(max(0, endNode.End-startNode.Pos)) // #nosec G115 -- size is always non-negative in valid clones
		totalSize += size

		// Convert each node to domain clone
		for _, node := range nodeGroup {
			clone := NodeToDomainClone(node, node.Filename)
			clones = append(clones, clone)
		}
	}

	// Calculate severity based on total size
	severity := domain.CalculateSeverity(totalSize, totalSize/10) // Approximate complexity

	return domain.CloneGroup{
		ID:       domain.CloneGroupID(groupID),
		Clones:   clones,
		Size:     totalSize,
		Hash:     domain.Hash(generateGroupHash(clones)),
		Severity: severity,
		Status:   domain.FileProcessingStateCompleted,
	}
}

// CreateAnalysisFromClones creates domain analysis from clone data.
func CreateAnalysisFromClones(cloneGroups []domain.CloneGroup, threshold uint) domain.Analysis {
	var totalClones uint
	var totalComplexity uint

	for _, group := range cloneGroups {
		totalClones += uint(len(group.Clones))
		totalComplexity += group.Size
	}

	return domain.Analysis{
		ID:          domain.AnalysisID(generateAnalysisID()),
		State:       domain.DetectionStateCompleted,
		Mode:        domain.AnalysisModeFull,
		Threshold:   domain.Threshold(threshold),
		CloneGroups: cloneGroups,
		Stats: domain.AnalysisStats{
			FilesAnalyzed:    domain.FileCount(countUniqueFiles(cloneGroups)),
			TotalClones:      domain.CloneCount(totalClones),
			TotalTokenSize:   domain.TokenCount(totalComplexity),
			ComplexityScore:  float64(totalComplexity) / float64(len(cloneGroups)+1),
			DuplicationRatio: calculateDuplicationRatio(cloneGroups),
			ProcessingTime:   domain.ProcessingTime(1000), // Calculate actual time
		},
		CreatedAt: currentTime(),
	}
}

// Helper functions

func generateGroupHash(_ []domain.Clone) string {
	// Implement proper hash generation
	return "group-hash"
}

func generateAnalysisID() string {
	// Implement proper ID generation
	return "analysis-id"
}

func countUniqueFiles(groups []domain.CloneGroup) uint {
	fileSet := make(map[string]bool)
	for _, group := range groups {
		for _, clone := range group.Clones {
			fileSet[clone.FilenameString()] = true
		}
	}
	return uint(len(fileSet))
}

func calculateDuplicationRatio(groups []domain.CloneGroup) float64 {
	if len(groups) == 0 {
		return 0.0
	}

	totalDuplicates := 0
	totalClones := 0
	for _, group := range groups {
		if len(group.Clones) > 1 {
			totalDuplicates += len(group.Clones) - 1
		}
		totalClones += len(group.Clones)
	}

	if totalClones == 0 {
		return 0.0
	}

	return float64(totalDuplicates) / float64(totalClones)
}

func currentTime() string {
	// Implement proper time generation
	return "2023-01-01T00:00:00Z"
}
