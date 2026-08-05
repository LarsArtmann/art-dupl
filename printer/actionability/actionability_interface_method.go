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

// isInterfaceMethodBody reports whether every clone is a FuncDecl that
// implements an interface contract AND whose body is small (≤4 nodes).
//
// Detection has two complementary paths:
//   - Type-aware path: when --type-aware is active, go/types sets the
//     InterfaceMethod flag on FuncDecl nodes that satisfy a same-package
//     interface. This catches custom interfaces beyond the static list.
//   - Static fallback: the commonInterfaceMethodNames list covers well-known
//     stdlib interfaces (fmt.Stringer, io.Reader, error, etc.) that are not
//     declared in the same package and thus invisible to same-package scanning.
//
// A method matching either path is suppressed as interface-driven boilerplate.
func isInterfaceMethodBody(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.FuncDecl {
			return false
		}

		if !root.InterfaceMethod && !slices.Contains(commonInterfaceMethodNames, root.Name) {
			return false
		}

		for _, child := range root.Children {
			if child.BaseType == golang.BlockStmt && len(child.Children) > maxInterfaceMethodBodyNodes {
				return false
			}
		}

		return true
	})
}
