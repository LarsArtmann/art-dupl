package golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// Parse the given file and return uniform syntax tree using default configuration.
// For custom configuration, use ParseWithConfig.
func Parse(filename string) (*syntax.Node, error) {
	return ParseWithConfig(filename, DefaultParseConfig())
}

// ParseWithConfig parses the given file with the specified configuration.
func ParseWithConfig(filename string, cfg ParseConfig) (*syntax.Node, error) {
	node, _, err := ParseWithLineCountConfig(filename, cfg)

	return node, err
}

// ParseWithLineCount parses the given file and returns the syntax tree along with the line count.
// Uses default configuration (semantic mode).
func ParseWithLineCount(filename string) (*syntax.Node, int, error) {
	return ParseWithLineCountConfig(filename, DefaultParseConfig())
}

// ParseWithLineCountConfig parses the given file with the specified configuration
// and returns the syntax tree along with the line count.
func ParseWithLineCountConfig(filename string, cfg ParseConfig) (*syntax.Node, int, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, 0, fmt.Errorf("parse %s: %w", filename, err)
	}

	if cfg.Preloaded != nil {
		return parsePreloaded(filename, cfg)
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	t := &transformer{
		fileset:  fset,
		filename: syntax.InternFilename(filename),
		config:   cfg,
		norm:     newNormalizer(cfg.Mode.NormalizesLocals()),
	}
	lineCount := fset.File(file.Pos()).LineCount()

	return t.trans(file), lineCount, nil
}

// parsePreloaded transforms a pre-loaded AST with type-checking results.
// The AST comes from go/packages so the *types.Info maps use the same node pointers.
func parsePreloaded(filename string, cfg ParseConfig) (*syntax.Node, int, error) {
	pre := cfg.Preloaded

	t := &transformer{
		fileset:  pre.Fset,
		filename: syntax.InternFilename(filename),
		config:   cfg,
		norm:     newNormalizer(cfg.Mode.NormalizesLocals()),
		typeInfo: pre.TypeInfo,
	}
	lineCount := pre.Fset.File(pre.File.Pos()).LineCount()

	return t.trans(pre.File), lineCount, nil
}

type transformer struct {
	fileset              *token.FileSet
	filename             string
	config               ParseConfig
	inInterface          bool
	norm                 *normalizer
	typeInfo                 *types.Info
	enclosingReturnArity     int32
	enclosingInterfaceMethod bool
}

// funcReturnArity returns the number of return values in a function type.
// Returns 0 for nil types or functions with no return values (void).
func funcReturnArity(ft *ast.FuncType) int32 {
	if ft == nil || ft.Results == nil {
		return 0
	}

	return int32(ft.Results.NumFields()) // #nosec G115 -- return arity is bounded by field count
}

// addWithNilCheck adds a child to o if not nil and valid.
func (t *transformer) addWithNilCheck(o *syntax.Node, node ast.Node) {
	if node != nil {
		defer func() {
			if r := recover(); r != nil {
				// Intentionally ignore panics from invalid nodes
				// This can happen with malformed AST nodes during transformation
				// We use the recovered value to silence the linter
				_ = r // Explicitly ignore the recovered value
			}
		}()

		o.AddChildren(t.trans(node))
	}
}
