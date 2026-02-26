package templ

import (
	"os"
	"strings"

	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// Parse parses the given templ file and returns the unified syntax tree.
func Parse(filename string) (*syntax.Node, error) {
	node, _, err := ParseWithLineCount(filename)

	return node, err
}

// ParseWithLineCount parses the given templ file and returns the syntax tree along with the line count.
func ParseWithLineCount(filename string) (*syntax.Node, int, error) {
	content, err := os.ReadFile(
		filename,
	) // #nosec G304 -- filename comes from controlled file system walk
	if err != nil {
		return nil, 0, err //nolint:wrapcheck // Pass through os.ReadFile error
	}

	return ParseBytes(filename, content)
}

// ParseBytes parses templ content and returns the syntax tree along with the line count.
func ParseBytes(filename string, content []byte) (*syntax.Node, int, error) {
	// Parse using the official templ parser
	tf, err := templparser.ParseString(string(content))
	if err != nil {
		return nil, 0, err //nolint:wrapcheck // Pass through parser error
	}

	// Transform to unified syntax tree
	t := &transformer{
		filename: filename,
	}

	node := t.transformTemplateFile(tf)

	// Count lines from content
	lineCount := strings.Count(string(content), "\n") + 1

	return node, lineCount, nil
}

type transformer struct {
	filename string
}
