package job

import (
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"github.com/LarsArtmann/art-dupl/syntax/templ"
)

// ParseFileByExtensionWithConfig parses a file based on its extension with the
// given detection mode. The mode controls how identifier names participate in
// matching (Exact/Semantic/Structural).
func ParseFileByExtensionWithConfig(
	file string,
	mode golang.DetectionMode,
) (*syntax.Node, int, error) {
	var (
		ast   *syntax.Node
		lines int
		err   error
	)

	switch filepath.Ext(file) {
	case ".templ":
		ast, lines, err = templ.ParseWithLineCountWithMode(
			file,
			mode == golang.DetectionModeSemantic,
		)
	default:
		// Default to Go parser for .go files and any other files that reach here
		cfg := golang.ParseConfig{Mode: mode}
		ast, lines, err = golang.ParseWithLineCountConfig(file, cfg)
	}

	return ast, lines, err
}
