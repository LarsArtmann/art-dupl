package actionability

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	templpkg "github.com/LarsArtmann/art-dupl/syntax/templ"
)

// isTemplChildrenIdiom reports whether every clone is a bare templ
// `{ children... }` slot-injection expression.
//
// `children` IS templ's composition mechanism: a wrapper component renders
// its slot through it by design, so any two wrappers produce an identical
// token and every wrapper pair becomes a clone group. The "extract to a
// component" suggestion is inapplicable — extracting the children
// expression is not a thing. Suppressed like other idiom patterns, still
// surfaced by --show-suppressed, and disable-able via
// --disable-pattern templ-children-idiom.
//
// The root type check is scoped to .templ files: templ node types are a
// parallel iota domain and ComponentChildrenExpression (22) numerically
// collides with golang.FuncLit — a bare Go FuncLit clone must never match
// (same file-kind-scoping rule as clone_classify's isTempl switch).
func isTemplChildrenIdiom(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]

		return strings.HasSuffix(root.Filename, ".templ") &&
			root.BaseType == templpkg.ComponentChildrenExpression
	})
}
