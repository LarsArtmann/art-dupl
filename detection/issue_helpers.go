package detection

import (
	"fmt"

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
	Tags     []string          `json:"tags,omitempty"`
}

// GetLine returns the line number for this issue (implements LineExtractor interface).
func (t TodoIssue) GetLine() domain.LineNumber {
	return t.Line
}

// LegacyIssue represents a legacy code pattern.
type LegacyIssue struct {
	Filename domain.Filepath      `json:"filename"`
	Line     domain.LineNumber    `json:"line"`
	Type     string               `json:"type"`
	Message  string               `json:"message"`
	Severity domain.CloneSeverity `json:"severity"`
}

// GetLine returns the line number for this issue (implements LineExtractor interface).
func (l LegacyIssue) GetLine() domain.LineNumber {
	return l.Line
}

// LegacyPattern represents a pattern to detect legacy code.
type LegacyPattern struct {
	Type      string   `json:"type"`
	Message   string   `json:"message"`
	Severity  string   `json:"severity"`
	Functions []string `json:"functions,omitempty"`
	Patterns  []string `json:"patterns,omitempty"`
	Imports   []string `json:"imports,omitempty"`
}

// LineExtractor is an interface for issue types that have a line number.
type LineExtractor interface {
	GetLine() domain.LineNumber
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

		nodesByFile := make(map[string][]*syntax.Node)
		for _, node := range data {
			nodesByFile[node.Filename] = append(nodesByFile[node.Filename], node)
		}

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
func createIssueMatch(prefix, filename string, line int) syntax.Match {
	return syntax.Match{
		Hash:  fmt.Sprintf("%s-%s-%d", prefix, filename, line),
		Frags: [][]*syntax.Node{{}},
	}
}

// findIssuesGeneric is a generic helper for finding issues and creating matches.
func findIssuesGeneric[T LineExtractor](
	data []*syntax.Node,
	finder func(string, []*syntax.Node) []T,
	matchType string,
) <-chan syntax.Match {
	return findIssuesInFile(data, finder, func(issue T, filename string) syntax.Match {
		return createIssueMatch(matchType, filename, int(issue.GetLine().Uint16()))
	})
}

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
