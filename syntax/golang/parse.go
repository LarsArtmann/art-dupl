package golang

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// Parse the given file and return uniform syntax tree.
func Parse(filename string) (*syntax.Node, error) {
	node, _, err := ParseWithLineCount(filename)
	return node, err
}

// ParseWithLineCount parses the given file and returns the syntax tree along with the line count.
func ParseWithLineCount(filename string) (*syntax.Node, int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, 0, err //nolint:wrapcheck // Parse errors are already clear
	}
	t := &transformer{
		fileset:  fset,
		filename: filename,
	}
	lineCount := fset.File(file.Pos()).LineCount()
	return t.trans(file), lineCount, nil
}

type transformer struct {
	fileset  *token.FileSet
	filename string
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
