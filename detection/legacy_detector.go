package detection

import (
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
func (ld *LegacyDetector) FindLegacy(nodes []*syntax.Node) <-chan syntax.Match {
	return findIssuesGeneric(nodes, ld.findLegacyInFile, "LEGACY")
}

// findLegacyInFile finds legacy patterns in a specific file.
func (ld *LegacyDetector) findLegacyInFile(filename string, nodes []*syntax.Node) []LegacyIssue {
	var issues []LegacyIssue

	for _, node := range nodes {
		for _, pattern := range ld.patterns {
			for _, funcName := range pattern.Functions {
				if strings.Contains(fmt.Sprintf("%v", node), funcName) {
					lineNum, err := domain.NewLineNumber(
						uint16(node.Pos),
					) // #nosec G115 -- Node positions are within uint16 range
					if skipIfInvalidLineNumber(logger.Default, filename, node.Pos, err) {
						continue
					}

					file, err := domain.NewFilepath(filename)
					if skipIfInvalidFilepath(logger.Default, filename, err) {
						continue
					}

					issues = append(issues, LegacyIssue{
						Filename: file,
						Line:     lineNum,
						Type:     pattern.Type,
						Message:  fmt.Sprintf("%s: %s", pattern.Message, funcName),
						Severity: domain.CloneSeverity(pattern.Severity),
					})
				}
			}
		}
	}

	return issues
}

// getDefaultLegacyPatterns returns default legacy code patterns.
func getDefaultLegacyPatterns() []LegacyPattern {
	return []LegacyPattern{
		{ //nolint:exhaustruct
			Type:     "deprecated_function",
			Message:  "Use of deprecated function",
			Severity: "medium",
			Functions: []string{
				"io/ioutil.ReadFile",     // Deprecated in Go 1.16
				"io/ioutil.WriteFile",    // Deprecated in Go 1.16
				"io/ioutil.TempFile",     // Deprecated in Go 1.16
				"os/exec.CommandContext", // Actually not deprecated, example
			},
		},
		{ //nolint:exhaustruct
			Type:     "old_pattern",
			Message:  "Old code pattern that should be refactored",
			Severity: "low",
			Patterns: []string{
				`for.*range.*len\(.*\).*{.*\[\].*=.*append\(.*,.*\)`,
				`if.*err.*!=.*nil.*{.*return.*err}`,
			},
		},
		{ //nolint:exhaustruct
			Type:     "deprecated_import",
			Message:  "Use of deprecated import path",
			Severity: "high",
			Imports: []string{
				"golang.org/x/net/context",
				"gopkg.in/yaml.v1",
			},
		},
	}
}
