package adapter

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/types"
)

// PrinterAdapter bridges domain types with existing printer interface
type PrinterAdapter struct {
	// Existing printer will be wrapped
}

// CloneToDomain converts printer clone to domain clone
func CloneToDomain(printerClone printer.Clone) domain.Clone {
	return domain.Clone{
		ID:        printerClone.Filename,
		Filename:  printerClone.Filename,
		StartLine: uint(printerClone.LineStart),
		EndLine:   uint(printerClone.LineEnd),
		Fragment:  string(printerClone.Fragment),
		Status:    types.FileProcessingStateCompleted,
	}
}

// NodeToDomainClone converts syntax node to domain clone
func NodeToDomainClone(node *syntax.Node, filename string) domain.Clone {
	return domain.NodeToClone(node, filename)
}

// CloneGroupFromNodes creates domain clone group from syntax nodes
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
		size := uint(endNode.End - startNode.Pos)
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
		ID:       groupID,
		Clones:   clones,
		Size:     totalSize,
		Hash:     generateGroupHash(clones),
		Severity: severity,
		Status:   types.FileProcessingStateCompleted,
	}
}

// CreateAnalysisFromClones creates domain analysis from clone data
func CreateAnalysisFromClones(cloneGroups []domain.CloneGroup, threshold uint) domain.Analysis {
	var totalClones uint
	var totalComplexity uint

	for _, group := range cloneGroups {
		totalClones += uint(len(group.Clones))
		totalComplexity += group.Size
	}

	return domain.Analysis{
		ID:          generateAnalysisID(),
		State:       types.DetectionStateCompleted,
		Mode:        types.AnalysisModeFull,
		Threshold:   threshold,
		CloneGroups: cloneGroups,
		Stats: domain.AnalysisStats{
			FilesAnalyzed:    countUniqueFiles(cloneGroups),
			TotalClones:      totalClones,
			TotalTokenSize:   totalComplexity,
			ComplexityScore:  float64(totalComplexity) / float64(len(cloneGroups)+1),
			DuplicationRatio: calculateDuplicationRatio(cloneGroups),
			ProcessingTime:   1000, // TODO: Calculate actual time
		},
		CreatedAt: currentTime(),
	}
}

// Helper functions

func generateGroupHash(clones []domain.Clone) string {
	// TODO: Implement proper hash generation
	return "group-hash"
}

func generateAnalysisID() string {
	// TODO: Implement proper ID generation
	return "analysis-id"
}

func countUniqueFiles(groups []domain.CloneGroup) uint {
	fileSet := make(map[string]bool)
	for _, group := range groups {
		for _, clone := range group.Clones {
			fileSet[clone.Filename] = true
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
	// TODO: Implement proper time generation
	return "2023-01-01T00:00:00Z"
}
