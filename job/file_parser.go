package job

import (
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"github.com/LarsArtmann/art-dupl/syntax/templ"
)

// ParseFileByExtension parses a file based on its extension.
func ParseFileByExtension(file string) (ast *syntax.Node, lines int, err error) {
	switch filepath.Ext(file) {
	case ".templ":
		ast, lines, err = templ.ParseWithLineCount(file)
	default:
		// Default to Go parser for .go files and any other files that reach here
		ast, lines, err = golang.ParseWithLineCount(file)
	}
	return ast, lines, err
}
