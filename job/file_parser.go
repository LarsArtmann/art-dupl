package job

import (
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"github.com/LarsArtmann/art-dupl/syntax/templ"
)

// ParseFileByExtensionWithConfig parses a file based on its extension with the
// given detection mode. The mode controls how identifier names participate in
// matching (Exact/Semantic/Structural). preloaded, when non-nil, supplies a
// pre-parsed AST with type-checking results for type-aware detection.
func ParseFileByExtensionWithConfig(
	file string,
	mode golang.DetectionMode,
	preloaded *golang.PreloadedAST,
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
		cfg := golang.ParseConfig{Mode: mode, Preloaded: preloaded}
		ast, lines, err = golang.ParseWithLineCountConfig(file, cfg)
	}

	return ast, lines, err
}
