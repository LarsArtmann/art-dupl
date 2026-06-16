package detection

import (
	"context"
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
	Severity domain.ClonePriority `json:"severity"`
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
}

// LineExtractor is an interface for issue types that have a line number.
type LineExtractor interface {
	GetLine() domain.LineNumber
}

// findIssuesInFile is a generic function that finds issues in a file and returns them.
// The ctx is checked on every channel send to prevent goroutine leaks.
func findIssuesInFile[T any](
	ctx context.Context,
	data []*syntax.Node,
	finder func(filePath string, nodeList []*syntax.Node) []T,
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
				if ctx.Err() != nil {
					return
				}

				match := matchCreator(issue, filename)

				select {
				case resultChan <- match:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return resultChan
}

// createIssueMatch creates a syntax.Match for an issue type that has a line number.
// Frags is left nil because issue detections (TODO, legacy) do not have AST
// node fragments — they represent code-quality findings, not code clones.
// The MultiDetector guard filters these out; proper issue output requires the
// MethodDetector pipeline (see docs/planning task M10).
func createIssueMatch(prefix, filename string, line int) syntax.Match {
	return syntax.Match{ //nolint:exhaustruct // Frags intentionally nil — issues have no AST fragments
		Hash: fmt.Sprintf("%s-%s-%d", prefix, filename, line),
	}
}

// findIssuesGeneric is a generic helper for finding issues and creating matches.
func findIssuesGeneric[T LineExtractor](
	ctx context.Context,
	data []*syntax.Node,
	finder func(string, []*syntax.Node) []T,
	matchType string,
) <-chan syntax.Match {
	return findIssuesInFile(ctx, data, finder, func(issue T, filename string) syntax.Match {
		return createIssueMatch(matchType, filename, int(issue.GetLine().Uint16()))
	})
}

// findFindingsInFile is a generic function that finds issues in files and
// converts them to domain.Finding values via the converter function.
// The ctx is checked on every channel send to prevent goroutine leaks.
func findFindingsInFile[T any](
	ctx context.Context,
	data []*syntax.Node,
	finder func(filePath string, nodeList []*syntax.Node) []T,
	converter func(issue T) domain.Finding,
) <-chan domain.Finding {
	resultChan := make(chan domain.Finding)

	go func() {
		defer close(resultChan)

		nodesByFile := make(map[string][]*syntax.Node)
		for _, node := range data {
			nodesByFile[node.Filename] = append(nodesByFile[node.Filename], node)
		}

		for filename, nodes := range nodesByFile {
			issues := finder(filename, nodes)
			for _, issue := range issues {
				if ctx.Err() != nil {
					return
				}

				finding := converter(issue)

				select {
				case resultChan <- finding:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return resultChan
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

// validateLocation validates and returns a LineNumber and Filepath pair.
// Returns ok=false if either validation fails (the caller should `continue`).
func validateLocation(
	l logger.Logger,
	filename string,
	line int,
) (domain.LineNumber, domain.Filepath, bool) {
	lineNum, err := domain.NewLineNumber(uint16(line)) // #nosec G115 -- positions within uint16 range
	if skipIfInvalidLineNumber(l, filename, line, err) {
		return 0, "", false
	}

	fp, err := domain.NewFilepath(filename)
	if skipIfInvalidFilepath(l, filename, err) {
		return 0, "", false
	}

	return lineNum, fp, true
}
