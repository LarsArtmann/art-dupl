package actionability

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// signatureFragmentRoots lists the base AST node types that exist ONLY
// inside function and method signatures: parameter/result fields, their
// lists, and the enclosing FuncType. A clone fragment rooted at one of
// these is a signature fragment, not code.
var signatureFragmentRoots = map[int32]bool{ //nolint:gochecknoglobals // static set
	golang.FuncType:  true,
	golang.Field:     true,
	golang.FieldList: true,
}

// isSignatureFragmentMatch reports whether every clone is a Go-snippet
// signature fragment: a partial function signature (a shared parameter
// tail, a result list) parsed out of a .templ file's embedded Go.
//
// The suffix tree matches the longest common token substring, so two templ
// helper functions whose signatures merely END alike — `..., lang Lang)
// string` across nine view helpers — surface the shared tail as its own
// fragment. No deduplication follows from identical signature parts: the
// function bodies differ, and a signature cannot be "extracted". These
// groups drowned real findings at low thresholds (observed at -t 3 on
// webphone: two groups, 9 and 2 members, all `lang Lang) string` tails).
//
// Scope is deliberately narrow:
//   - GoSnippet gates the enum collision: templ markup nodes are a parallel
//     iota (golang.Field == templ.ConditionalAttributeIfStatement == 18),
//     so BaseType may only be read through the golang table when the flag
//     is set — the same file-kind-scoping rule as clone_classify (ADR-0027).
//   - Signature-part roots only. Statement fragments inside templ script
//     blocks (if/for/switch in Go snippets) stay visible — those carry
//     extractable logic. Whole type declarations (StructType, InterfaceType)
//     never appear as non-statement fragment roots: GenDecl specs are
//     fingerprinted as single composite tokens.
//   - Plain .go files are unaffected: non-statement signature roots there
//     are already rejected by FindSyntaxUnits' mixed-parser guard in
//     statement-bearing files.
//
// Still surfaced by --show-suppressed, disable-able via
// --disable-pattern signature-fragment.
func isSignatureFragmentMatch(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) == 0 {
			return false
		}

		root := seq[0]

		return root.GoSnippet && signatureFragmentRoots[root.BaseType]
	})
}
