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
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// TodoIssue represents a TODO comment found in code.
type TodoIssue struct {
	Filename domain.Filepath   `json:"filename"`
	Line     domain.LineNumber `json:"line"`
	Text     string            `json:"text"`
	Type     string            `json:"type"`
	Tags     []string          `json:"tags,omitempty"` // @username, date, etc.
}

// GetLine returns the line number for this issue (implements LineExtractor interface).
func (t TodoIssue) GetLine() domain.LineNumber {
	return t.Line
}

// LegacyIssue represents a legacy code pattern.
type LegacyIssue struct {
	Filename domain.Filepath      `json:"filename"`
	Line     domain.LineNumber    `json:"line"`
	Type     string               `json:"type"` // deprecated function, old pattern, etc.
	Message  string               `json:"message"`
	Severity domain.CloneSeverity `json:"severity"` // low, medium, high
}

// GetLine returns the line number for this issue (implements LineExtractor interface).
func (l LegacyIssue) GetLine() domain.LineNumber {
	return l.Line
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

	// XXX pattern
	patterns["XXX"] = regexp.MustCompile(`(?i)XXX\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)

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

// LineExtractor is an interface for issue types that have a line number.
// This allows findIssuesGeneric to work with any issue type without needing
// a separate line extraction function.
type LineExtractor interface {
	GetLine() domain.LineNumber
}

// findIssuesGeneric is a generic helper for finding issues and creating matches.
// It avoids code duplication between FindTodos and FindLegacy methods.
// T must implement LineExtractor interface for line number access.
func findIssuesGeneric[T LineExtractor](
	data []*syntax.Node,
	finder func(string, []*syntax.Node) []T,
	matchType string,
) <-chan syntax.Match {
	return findIssuesInFile(data, finder, func(issue T, filename string) syntax.Match {
		return createIssueMatch(matchType, filename, int(issue.GetLine().Uint16()))
	})
}

// skipIfInvalidLineNumber returns true and logs if the line number is invalid.
func skipIfInvalidLineNumber(l logger.Logger, filename string, line any, err error) bool {
	if err != nil {
		l.Debug(
			"skipping issue with invalid line number",
			"file", filename,
			"line", line,
			"err", err,
		)
		return true
	}
	return false
}

// skipIfInvalidFilepath returns true and logs if the filepath is invalid.
func skipIfInvalidFilepath(l logger.Logger, filename string, err error) bool {
	if err != nil {
		l.Debug(
			"skipping issue with invalid filename",
			"file", filename,
			"err", err,
		)
		return true
	}
	return false
}

// FindTodos finds all TODO-style comments in the provided nodes.
func (td *TodoDetector) FindTodos(data []*syntax.Node) <-chan syntax.Match {
	return findIssuesGeneric(data, td.findTodosInFile, "TODO")
}

// findTodosInFile parses the file and finds TODO comments.
//
//nolint:gocognit // TODO parsing requires handling multiple comment formats and pattern matching
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

					lineNum, err := domain.NewLineNumber(
						uint16(line),
					) // #nosec G115 -- Line numbers from parser are within uint16 range
					if skipIfInvalidLineNumber(logger.Default, filename, line, err) {
						continue
					}

					file, err := domain.NewFilepath(filename)
					if skipIfInvalidFilepath(logger.Default, filename, err) {
						continue
					}

					todos = append(todos, TodoIssue{
						Filename: file,
						Line:     lineNum,
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
	return findIssuesGeneric(data, ld.findLegacyInFile, "LEGACY")
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
				`for.*range.*len\(.*\).*{.*\[\].*=.*append\(.*,.*\)`, // range append pattern
				`if.*err.*!=.*nil.*{.*return.*err}`,                  // error checking pattern (could be improved)
			},
		},
		{ //nolint:exhaustruct
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
