package job

import (
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"github.com/LarsArtmann/art-dupl/syntax/templ"
)

// ParseFileByExtension parses a file based on its extension.
// Uses default configuration (semantic mode).
func ParseFileByExtension(file string) (*syntax.Node, int, error) {
	return ParseFileByExtensionWithConfig(file, true)
}

// ParseFileByExtensionWithConfig parses a file based on its extension with semantic mode configuration.
// When semantic is true, identifier names are included in the type hash (reduces false positives).
// When semantic is false, only AST structure is considered (structural matching).
func ParseFileByExtensionWithConfig(
	file string,
	semantic bool,
) (*syntax.Node, int, error) {
	var (
		ast   *syntax.Node
		lines int
		err   error
	)

	switch filepath.Ext(file) {
	case ".templ":
		ast, lines, err = templ.ParseWithLineCount(file)
	default:
		// Default to Go parser for .go files and any other files that reach here
		mode := golang.DetectionModeStructural
		if semantic {
			mode = golang.DetectionModeSemantic
		}

		cfg := golang.ParseConfig{Mode: mode}
		ast, lines, err = golang.ParseWithLineCountConfig(file, cfg)
	}

	return ast, lines, err
}
