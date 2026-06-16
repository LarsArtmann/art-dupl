package detection

import (
	"context"
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// LegacyDetector finds legacy code patterns.
type LegacyDetector struct {
	patterns []LegacyPattern
}

// NewLegacyDetector creates a new legacy detector with default patterns.
func NewLegacyDetector() *LegacyDetector {
	return &LegacyDetector{
		patterns: getDefaultLegacyPatterns(),
	}
}

// FindLegacy finds all legacy patterns in provided nodes.
func (ld *LegacyDetector) FindLegacy(
	ctx context.Context,
	nodes []*syntax.Node,
) <-chan syntax.Match {
	return findIssuesGeneric(ctx, nodes, ld.findLegacyInFile, "LEGACY")
}

// FindFindings finds all legacy patterns and returns them as domain.Finding values.
func (ld *LegacyDetector) FindFindings(
	ctx context.Context,
	nodes []*syntax.Node,
) <-chan domain.Finding {
	return findFindingsInFile(ctx, nodes, ld.findLegacyInFile, legacyIssueToFinding)
}

func legacyIssueToFinding(issue LegacyIssue) domain.Finding {
	return domain.Finding{ //nolint:exhaustruct
		Filename: issue.Filename,
		Line:     issue.Line,
		Type:     domain.FindingTypeLegacy,
		Message:  issue.Message,
		Priority: issue.Severity,
	}
}

// findLegacyInFile finds legacy patterns in a specific file.
func (ld *LegacyDetector) findLegacyInFile(filename string, nodes []*syntax.Node) []LegacyIssue {
	var issues []LegacyIssue

	for _, node := range nodes {
		for _, pattern := range ld.patterns {
			for _, funcName := range pattern.Functions {
				if nodeContainsFunctionCall(node, funcName) {
					lineNum, file, ok := validateLocation(logger.Default, filename, int(node.Pos))
					if !ok {
						continue
					}

					issues = append(issues, LegacyIssue{
						Filename: file,
						Line:     lineNum,
						Type:     pattern.Type,
						Message:  fmt.Sprintf("%s: %s", pattern.Message, funcName),
						Severity: domain.ClonePriority(pattern.Severity),
					})
				}
			}
		}
	}

	return issues
}

// nodeContainsFunctionCall checks if a node or its children reference a function name.
// It matches against the node's Name field and recursively checks children,
// avoiding false positives from stringified struct dumps.
func nodeContainsFunctionCall(node *syntax.Node, funcName string) bool {
	if strings.Contains(node.Name, funcName) {
		return true
	}

	for _, child := range node.Children {
		if nodeContainsFunctionCall(child, funcName) {
			return true
		}
	}

	return false
}

// getDefaultLegacyPatterns returns default legacy code patterns.
func getDefaultLegacyPatterns() []LegacyPattern {
	return []LegacyPattern{
		{
			Type:     "deprecated_function",
			Message:  "Use of deprecated function",
			Severity: "medium",
			Functions: []string{
				"io/ioutil.ReadFile",  // Deprecated in Go 1.16
				"io/ioutil.WriteFile", // Deprecated in Go 1.16
				"io/ioutil.TempFile",  // Deprecated in Go 1.16
			},
		},
	}
}
