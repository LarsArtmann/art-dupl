package actionability

import (
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// commonInterfaceMethodNames are method names from standard library interfaces
// that types frequently implement with trivial delegation bodies. When multiple
// types implement the same interface with identical short bodies, the duplication
// is interface-driven boilerplate, not copy-paste.
var commonInterfaceMethodNames = []string{ //nolint:gochecknoglobals // static name set
	// fmt.Stringer / fmt.GoStringer
	"String", "GoString",
	// fmt.Formatter
	"Format",
	// error interface
	"Error", //nolint:goconst
	// io.Reader/Writer/Closer/Seeker
	"Read", "Write", "Close", "Seek",
	// sort.Interface
	"Len", "Less", "Swap",
	// json.Marshaler/Unmarshaler
	"MarshalJSON", "UnmarshalJSON",
	// encoding.TextMarshaler/TextUnmarshaler
	"MarshalText", "UnmarshalText",
	// encoding.BinaryMarshaler/BinaryUnmarshaler
	"MarshalBinary", "UnmarshalBinary",
	// driver.Valuer
	"Value",
	// flag.Value
	"Set",
}

// maxInterfaceMethodBodyNodes is the maximum body size for interface-method
// suppression. Bodies larger than this likely contain real logic worth
// extracting.
const maxInterfaceMethodBodyNodes = 4

// isInterfaceMethodBody reports whether every clone is the body of a function
// that implements an interface contract AND whose body is small (≤4 nodes).
//
// Detection has two paths depending on the clone root type:
//
//   - FuncDecl root (path 1): The FuncDecl itself is the clone root. This only
//     happens in edge-case files without statements. The InterfaceMethod flag
//     or the static name list identifies interface methods; body size is checked
//     via the BlockStmt child.
//
//   - Statement root (path 2, normal Go files): FuncDecl nodes can never be
//     clone roots in real Go files because the structural filter in
//     FindSyntaxUnits rejects non-Statement clone roots. The transformer
//     propagates the InterfaceMethod flag from the enclosing FuncDecl to body
//     statement nodes, so the flag is available on the statement clone root.
//     The number of clone units (len(seq)) IS the body size. The static name
//     list does not apply here because the function name is not on statement
//     nodes.
//
// Path 2 requires --type-aware mode (the flag is only set when go/types
// confirms interface satisfaction). Path 1 also works without type-aware
// via the static name list.
func isInterfaceMethodBody(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) == 0 {
			return false
		}

		root := seq[0]

		// Path 1: FuncDecl root (edge case — files without statements).
		if root.BaseType == golang.FuncDecl {
			if !root.InterfaceMethod && !slices.Contains(commonInterfaceMethodNames, root.Name) {
				return false
			}

			for _, child := range root.Children {
				if child.BaseType == golang.BlockStmt && len(child.Children) > maxInterfaceMethodBodyNodes {
					return false
				}
			}

			return true
		}

		// Path 2: Statement-level root (normal Go files).
		// The InterfaceMethod flag is propagated from the enclosing FuncDecl
		// by the transformer. The number of clone units IS the body size.
		return root.InterfaceMethod && len(seq) <= maxInterfaceMethodBodyNodes
	})
}
