package detection

import (
	"context"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// TodoDetector finds TODO comments in Go source code.
type TodoDetector struct {
	patterns map[string]*regexp.Regexp
}

// NewTodoDetector creates a new TODO detector.
func NewTodoDetector() *TodoDetector {
	patterns := make(map[string]*regexp.Regexp)

	patterns["TODO"] = regexp.MustCompile(`(?i)TODO\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	patterns["FIXME"] = regexp.MustCompile(`(?i)FIXME\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	patterns["XXX"] = regexp.MustCompile(`(?i)XXX\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	patterns["HACK"] = regexp.MustCompile(`(?i)HACK\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)
	patterns["NOTE"] = regexp.MustCompile(`(?i)NOTE\s*(?:\(([^)]*)\))?\s*:\s*(.+)`)

	return &TodoDetector{patterns: patterns}
}

// FindTodos finds all TODO-style comments in the provided nodes.
func (td *TodoDetector) FindTodos(
	ctx context.Context,
	data []*syntax.Node,
) <-chan syntax.Match {
	return findIssuesGeneric(ctx, data, td.findTodosInFile, "TODO")
}

// FindFindings finds all TODO-style comments and returns them as domain.Finding values.
func (td *TodoDetector) FindFindings(
	ctx context.Context,
	data []*syntax.Node,
) <-chan domain.Finding {
	return findFindingsInFile(ctx, data, td.findTodosInFile, todoIssueToFinding)
}

func todoIssueToFinding(issue TodoIssue) domain.Finding {
	return domain.Finding{
		Filename: issue.Filename,
		Line:     issue.Line,
		Type:     domain.FindingTypeTodo,
		Message:  issue.Text,
		Priority: domain.PriorityLow,
		Tags:     issue.Tags,
	}
}

// findTodosInFile parses the file and finds TODO comments.
func (td *TodoDetector) findTodosInFile(path string, astNodes []*syntax.Node) []TodoIssue {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		logger.Default.Warn("failed to parse file for TODO detection", "file", path, "err", err)

		return nil
	}

	var todos []TodoIssue

	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			line := fset.Position(comment.Slash).Line
			text := strings.TrimSpace(comment.Text)

			text = strings.TrimPrefix(text, "//")
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
			text = strings.TrimSpace(text)

			for todoType, pattern := range td.patterns {
				matches := pattern.FindStringSubmatch(text)
				if len(matches) > 0 {
					tags := []string{}
					if matches[1] != "" {
						tags = strings.Split(matches[1], ",")
						for i, tag := range tags {
							tags[i] = strings.TrimSpace(tag)
						}
					}

					todoText := ""
					if len(matches) > 2 {
						todoText = strings.TrimSpace(matches[2])
					}

					lineNum, fp, ok := validateLocation(logger.Default, path, line)
					if !ok {
						continue
					}

					todos = append(todos, TodoIssue{
						Filename: fp,
						Line:     lineNum,
						Text:     todoText,
						Type:     todoType,
						Tags:     tags,
					})

					break
				}
			}
		}
	}

	return todos
}
