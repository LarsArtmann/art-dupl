package detection

//
// DOMAIN TYPES STATUS:
// ✅ Added domain package import
// ✅ Updated TodoIssue to use domain types
// ✅ Updated LegacyIssue to use domain types
//
// Note: Detection methods (todos, legacy) have similar structure to
// clone detection but don't share a common interface. Consider defining
// a Detector interface that all detection methods implement.

import (
	"fmt"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// TodoIssue represents a TODO comment found in code.
type TodoIssue struct {
	Filename domain.Filepath  `json:"filename"`
	Line     domain.LineNumber `json:"line"`
	Text     string           `json:"text"`
	Type     string           `json:"type"`           //nolint:godox // TODO, FIXME, XXX, etc.
	Tags     []string         `json:"tags,omitempty"` // @username, date, etc.
}

// LegacyIssue represents a legacy code pattern.
type LegacyIssue struct {
	Filename domain.Filepath `json:"filename"`
	Line     domain.LineNumber `json:"line"`
	Type     string           `json:"type"` // deprecated function, old pattern, etc.
	Message  string           `json:"message"`
	Severity domain.CloneSeverity `json:"severity"` // low, medium, high
}

// TodoDetector finds TODO comments in Go source code.
type TodoDetector struct {
	patterns map[string]*regexp.Regexp
}

// NewTodoDetector creates a new TODO detector.
func NewTodoDetector() *TodoDetector {
	// Patterns for TODO comments with optional tags
	// Examples:
	// // TODO: fix this
	// // FIXME(@user): handle error
	// // TODO(2024-01-01): replace with new API
	// // XXX: this is a hack
	patterns := make(map[string]*regexp.Regexp)

	// Pattern definitions below
	//nolint:godox
	// TODO pattern
	patterns["TODO"] = regexp.MustCompile(`(?i)TODO\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	//nolint:godox
	// FIXME pattern
	patterns["FIXME"] = regexp.MustCompile(`(?i)FIXME\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	//nolint:godox
	// XXX pattern
	patterns["XXX"] = regexp.MustCompile(`(?i)XXX\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	//nolint:godox
	// HACK pattern
	patterns["HACK"] = regexp.MustCompile(`(?i)HACK\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	// NOTE pattern
	patterns["NOTE"] = regexp.MustCompile(`(?i)NOTE\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)

	return &TodoDetector{patterns: patterns}
}

// findIssuesInFile is a generic function that finds issues in a file and returns them.
func findIssuesInFile[T any](
	data []*syntax.Node,
	finder func(filename string, nodes []*syntax.Node) []T,
	matchCreator func(issue T, filename string) syntax.Match,
) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		// Group nodes by filename
		nodesByFile := make(map[string][]*syntax.Node)
		for _, node := range data {
			nodesByFile[node.Filename] = append(nodesByFile[node.Filename], node)
		}

		// Process each file
		for filename, nodes := range nodesByFile {
			issues := finder(filename, nodes)
			for _, issue := range issues {
				resultChan <- matchCreator(issue, filename)
			}
		}
	}()

	return resultChan
}

// createIssueMatch creates a syntax.Match for an issue type that has a line number.
// This is used by FindTodos and FindLegacy to avoid duplicating the match creation logic.
func createIssueMatch(prefix, filename string, line int) syntax.Match {
	return syntax.Match{
		Hash:  fmt.Sprintf("%s-%s-%d", prefix, filename, line),
		Frags: [][]*syntax.Node{{}}, // Empty frag since issues aren't code fragments
	}
}

// findIssuesGeneric is a generic helper for finding issues and creating matches.
// It avoids code duplication between FindTodos and FindLegacy methods.
func findIssuesGeneric[T any](
	data []*syntax.Node,
	finder func(string, []*syntax.Node) []T,
	matchType string,
	lineExtractor func(T) int,
) <-chan syntax.Match {
	return findIssuesInFile(data, finder, func(issue T, filename string) syntax.Match {
		return createIssueMatch(matchType, filename, lineExtractor(issue))
	})
}

// FindTodos finds all TODO-style comments in the provided nodes.
func (td *TodoDetector) FindTodos(data []*syntax.Node) <-chan syntax.Match {
	return findIssuesGeneric(data, td.findTodosInFile, "TODO", func(t TodoIssue) int { return t.Line })
}

// findTodosInFile parses the file and finds TODO comments.
func (td *TodoDetector) findTodosInFile(filename string, nodes []*syntax.Node) []TodoIssue {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil
	}

	var todos []TodoIssue

	// Check all comment groups
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			line := fset.Position(comment.Slash).Line
			text := strings.TrimSpace(comment.Text)

			// Remove comment markers
			text = strings.TrimPrefix(text, "//")
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
			text = strings.TrimSpace(text)

			// Check against patterns
			for todoType, pattern := range td.patterns {
				matches := pattern.FindStringSubmatch(text)
				if len(matches) > 0 {
					tags := []string{}
					if matches[1] != "" {
						// Parse tags from parentheses
						tags = strings.Split(matches[1], ",")
						for i, tag := range tags {
							tags[i] = strings.TrimSpace(tag)
						}
					}

					todoText := ""
					if len(matches) > 2 {
						todoText = strings.TrimSpace(matches[2])
					}

					todos = append(todos, TodoIssue{
						Filename: filename,
						Line:     line,
						Text:     todoText,
						Type:     todoType,
						Tags:     tags,
					})
					break // Only match first pattern per comment
				}
			}
		}
	}

	return todos
}

// LegacyDetector finds legacy code patterns.
type LegacyDetector struct {
	patterns []LegacyPattern
}

// LegacyPattern represents a pattern to detect legacy code.
type LegacyPattern struct {
	Type      string   `json:"type"`
	Message   string   `json:"message"`
	Severity  string   `json:"severity"`
	Functions []string `json:"functions,omitempty"` // Deprecated function names
	Patterns  []string `json:"patterns,omitempty"`  // Regex patterns for code
	Imports   []string `json:"imports,omitempty"`   // Deprecated import paths
}

// NewLegacyDetector creates a new legacy detector with default patterns.
func NewLegacyDetector() *LegacyDetector {
	return &LegacyDetector{
		patterns: getDefaultLegacyPatterns(),
	}
}

// FindLegacy finds all legacy patterns in provided nodes.
func (ld *LegacyDetector) FindLegacy(data []*syntax.Node) <-chan syntax.Match {
	return findIssuesGeneric(data, ld.findLegacyInFile, "LEGACY", func(l LegacyIssue) int { return l.Line })
}

// findLegacyInFile finds legacy patterns in a specific file.
func (ld *LegacyDetector) findLegacyInFile(filename string, nodes []*syntax.Node) []LegacyIssue {
	var issues []LegacyIssue

	for _, node := range nodes {
		// Check each legacy pattern
		for _, pattern := range ld.patterns {
			// Check function names
			for _, funcName := range pattern.Functions {
				// This is simplified - in a real implementation,
				// we'd need to check if this node represents a call to the deprecated function
				if strings.Contains(fmt.Sprintf("%v", node), funcName) {
					issues = append(issues, LegacyIssue{
						Filename: filename,
						Line:     int(node.Pos),
						Type:     pattern.Type,
						Message:  fmt.Sprintf("%s: %s", pattern.Message, funcName),
						Severity: pattern.Severity,
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
		{
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
		{
			Type:     "old_pattern",
			Message:  "Old code pattern that should be refactored",
			Severity: "low",
			Patterns: []string{
				`for.*range.*len\(.*\).*{.*\[\].*=.*append\(.*,.*\)`, // range append pattern
				`if.*err.*!=.*nil.*{.*return.*err}`,                  // error checking pattern (could be improved)
			},
		},
		{
			Type:     "deprecated_import",
			Message:  "Use of deprecated import path",
			Severity: "high",
			Imports: []string{
				"golang.org/x/net/context", // Use context package instead
				"gopkg.in/yaml.v1",         // Use v2 or v3
			},
		},
	}
}
