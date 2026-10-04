package templ

import (
	"fmt"
	"os"
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templparser "github.com/a-h/templ/parser/v2"
)

// Parse parses the given templ file and returns the unified syntax tree.
func Parse(filename string) (*syntax.Node, error) {
	node, _, err := ParseWithLineCount(filename)

	return node, err
}

// ParseWithLineCount parses the given templ file and returns the syntax tree along with the line count.
func ParseWithLineCount(filepath string) (*syntax.Node, int, error) {
	return ParseWithLineCountWithMode(filepath, golang.DetectionModeSemantic)
}

// ParseWithLineCountWithMode parses with the given detection mode.
func ParseWithLineCountWithMode(filepath string, mode golang.DetectionMode) (*syntax.Node, int, error) {
	content, err := os.ReadFile(
		filepath,
	) // #nosec G304 -- filepath comes from controlled file system walk
	if err != nil {
		return nil, 0, fmt.Errorf("read templ file %s: %w", filepath, err)
	}

	return ParseBytesWithMode(filepath, content, mode)
}

// ParseBytes parses templ content and returns the syntax tree along with the line count.
func ParseBytes(filename string, content []byte) (*syntax.Node, int, error) {
	return ParseBytesWithMode(filename, content, golang.DetectionModeSemantic)
}

// ParseBytesWithMode parses templ content with the given detection mode.
// The mode controls how names participate in matching, mirroring the Go
// pipeline: structural mode ignores element/attribute/callee names, exact
// mode hashes them verbatim (Type 1), and semantic mode hashes them with
// local variables alpha-normalized (Type 2).
func ParseBytesWithMode(filename string, content []byte, mode golang.DetectionMode) (*syntax.Node, int, error) {
	// Parse using the official templ parser
	tf, err := templparser.ParseString(string(content))
	if err != nil {
		return nil, 0, fmt.Errorf("parse templ content from %s: %w", filename, err)
	}

	// Transform to unified syntax tree
	t := &transformer{
		filename:   syntax.InternFilename(filename),
		contentLen: len(content),
		mode:       mode,
	}

	node := t.transformTemplateFile(tf)

	// Count lines from content
	lineCount := strings.Count(string(content), "\n") + 1

	return node, lineCount, nil
}

type transformer struct {
	filename   string
	contentLen int
	mode       golang.DetectionMode
	symbols    map[string]string // local-variable symbol table for expression normalization
}
