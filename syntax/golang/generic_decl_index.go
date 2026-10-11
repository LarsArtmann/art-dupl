package golang

import (
	"fmt"
	"go/ast"
	"strings"
)

// CollectGenericDecls returns an index of every generic declaration across the
// pre-loaded ASTs: map key is the fully-qualified name (package path + "."
// + declaration name), value is the declaration's type parameter count. A nil
// or empty TypeAwareData yields an empty map.
//
// The index is syntax-level and bounded (plan Q3 decision): a TypeParams scan
// over ASTs that are already in memory from the go/packages load — no go/types
// queries, no additional crawl. It feeds the reuse-existing-generic verdict:
// when a generics candidate's divergent type positions are all instantiations
// of an indexed generic, the user should reuse that generic instead of
// extracting a new one ("reuse pkg.Box[T]" rather than "extract to generic").
func CollectGenericDecls(td TypeAwareData) map[string]int {
	index := make(map[string]int)

	for _, pre := range td {
		if pre == nil || pre.File == nil {
			continue
		}

		for _, decl := range pre.File.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Type != nil && d.Type.TypeParams != nil && len(d.Type.TypeParams.List) > 0 {
					index[qualifiedDeclName(pre.PkgPath, d.Name.Name)] = len(d.Type.TypeParams.List)
				}

			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.TypeParams == nil || len(ts.TypeParams.List) == 0 {
						continue
					}

					index[qualifiedDeclName(pre.PkgPath, ts.Name.Name)] = len(ts.TypeParams.List)
				}
			}
		}
	}

	return index
}

// qualifiedDeclName joins a package path and a declaration name into the same
// form go/types uses in type strings ("path.Name"), so VarType values like
// "github.com/x/pkg.Box[int]" resolve by cutting the type arguments.
func qualifiedDeclName(pkgPath, name string) string {
	if pkgPath == "" {
		return name
	}

	return pkgPath + "." + name
}

// LookupGenericDecl resolves a VarType type string against the generic-decl
// index. The string may carry pointer/slice prefixes ("*pkg.Box[int]",
// "[]pkg.Box[int]") and type arguments ("pkg.Box[int,string]"); the returned
// name is the fully-qualified declaration key when the base name is an indexed
// generic, "" otherwise.
func LookupGenericDecl(typeStr string, index map[string]int) string {
	base := typeStr

	// Strip type arguments: everything from the first '[' that opens the
	// argument list. Instantiation strings end with the matching ']' or are
	// nested ("Box[Pair[int,string]]") — cutting at the FIRST '[' leaves the
	// base name regardless of nesting.
	if cut := strings.Index(base, "["); cut >= 0 {
		base = base[:cut]
	}

	// Strip pointer and composite prefixes repeatedly: "*pkg.Box", "[]pkg.Box",
	// "[2]pkg.Box", "map[k]pkg.Box" all reduce by trimming leading qualifier
	// segments. Named generics cannot be the whole string after stripping
	// qualifiers, so unknown prefixes resolve to no match.
	for len(base) > 0 {
		switch base[0] {
		case '*', ' ':
			base = base[1:]
		case '[':
			cut := strings.Index(base, "]")
			if cut < 0 {
				return ""
			}

			base = base[cut+1:]
		default:
			var trimmed bool

			for _, p := range []string{"map[", "chan ", "chan<- ", "<-chan "} {
				if after, ok := strings.CutPrefix(base, p); ok {
					base = after
					trimmed = true

					break
				}
			}

			if !trimmed {
				return lookupBareOrQualified(base, index)
			}
		}
	}

	return lookupBareOrQualified(base, index)
}

// lookupBareOrQualified matches a base name against the index, preferring the
// exact qualified form and falling back to the bare declaration name. The
// fallback covers same-run packages whose PkgPath was unavailable.
func lookupBareOrQualified(base string, index map[string]int) string {
	if base == "" {
		return ""
	}

	if _, ok := index[base]; ok {
		return base
	}

	before, _, found := strings.CutLast(base, ".")
	if found {
		bare := base[len(before)+1:]
		if _, ok := index[bare]; ok {
			return bare
		}
	}

	return ""
}

// FormatGenericParams renders a type-parameter list for suggestion text:
// FormatGenericParams("Box", 1) → "Box[T]", FormatGenericParams("Pair", 2) →
// "Pair[T1,T2]". The parameter names are placeholders — the user's declaration
// may name them differently.
func FormatGenericParams(name string, count int) string {
	if count <= 0 {
		return name
	}

	params := make([]string, count)
	for i := range params {
		params[i] = fmt.Sprintf("T%d", i+1)
	}

	return name + "[" + strings.Join(params, ",") + "]"
}
