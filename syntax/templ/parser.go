package templ

import (
	"fmt"
	"os"
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax"
	templparser "github.com/a-h/templ/parser/v2"
)

// Parse parses the given templ file and returns the unified syntax tree.
func Parse(filename string) (*syntax.Node, error) {
	node, _, err := ParseWithLineCount(filename)

	return node, err
}

// ParseWithLineCount parses the given templ file and returns the syntax tree along with the line count.
func ParseWithLineCount(filepath string) (*syntax.Node, int, error) {
	return ParseWithLineCountWithMode(filepath, false)
}

// ParseWithLineCountWithMode parses with optional semantic encoding.
func ParseWithLineCountWithMode(filepath string, semantic bool) (*syntax.Node, int, error) {
	content, err := os.ReadFile(
		filepath,
	) // #nosec G304 -- filepath comes from controlled file system walk
	if err != nil {
		return nil, 0, fmt.Errorf("read templ file %s: %w", filepath, err)
	}

	return ParseBytesWithMode(filepath, content, semantic)
}

// ParseBytes parses templ content and returns the syntax tree along with the line count.
func ParseBytes(filename string, content []byte) (*syntax.Node, int, error) {
	return ParseBytesWithMode(filename, content, false)
}

// ParseBytesWithMode parses templ content with optional semantic encoding.
// When semantic is true, element tag names and attribute names are encoded
// into node Types, so <a href> and <div class> produce different tokens.
func ParseBytesWithMode(filename string, content []byte, semantic bool) (*syntax.Node, int, error) {
	// Parse using the official templ parser
	tf, err := templparser.ParseString(string(content))
	if err != nil {
		return nil, 0, fmt.Errorf("parse templ content from %s: %w", filename, err)
	}

	// Transform to unified syntax tree
	t := &transformer{
		filename:   syntax.InternFilename(filename),
		contentLen: len(content),
		semantic:   semantic,
		symbols:     make(map[string]string),
	}

	node := t.transformTemplateFile(tf)

	// Count lines from content
	lineCount := strings.Count(string(content), "\n") + 1

	return node, lineCount, nil
}

type transformer struct {
	filename   string
	contentLen int
	semantic   bool
	symbols    map[string]string // local-variable symbol table for expression normalization
}
