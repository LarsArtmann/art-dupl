package printer

import "github.com/LarsArtmann/art-dupl/syntax"

type ReadFile func(filename string) ([]byte, error)

type Printer interface {
	PrintHeader() error
	PrintClones(dups [][]*syntax.Node, sortBy ...string) error // Add optional sortBy parameter
	PrintFooter() error
}
