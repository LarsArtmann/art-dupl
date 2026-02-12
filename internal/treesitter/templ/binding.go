package tree_sitter_templ

// #cgo CFLAGS: -std=c11 -fPIC -I${SRCDIR}/src
// #include "src/parser.c"
// #include "src/scanner.c"
import "C"

import "unsafe"

// Language returns the tree-sitter Language for templ grammar.
func Language() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_templ())
}
