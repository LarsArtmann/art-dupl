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
// extracting a new one ("reuse pkg.Box[T1]" rather than "extract to generic").
func CollectGenericDecls(td TypeAwareData) map[string]int {
	index := make(map[string]int)

	for _, pre := range td {
		if pre == nil || pre.File == nil {
			continue
		}

		for _, decl := range pre.File.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Type != nil && d.Type.TypeParams != nil {
					if count := typeParamCount(d.Type.TypeParams); count > 0 {
						index[qualifiedDeclName(pre.PkgPath, d.Name.Name)] = count
					}
				}

			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.TypeParams == nil {
						continue
					}

					if count := typeParamCount(ts.TypeParams); count > 0 {
						index[qualifiedDeclName(pre.PkgPath, ts.Name.Name)] = count
					}
				}
			}
		}
	}

	return index
}

// typeParamCount counts type-parameter NAMES, not FieldList fields: a single
// field can declare several parameters ("[T, U any]" is one field, two names).
func typeParamCount(fl *ast.FieldList) int {
	count := 0

	for _, field := range fl.List {
		count += len(field.Names)
	}

	return count
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
// index. The string may carry pointer/slice/array/map/chan prefixes
// ("*pkg.Box[int]", "[]pkg.Box[int]", "map[string]pkg.Box[int]") and ends with
// type arguments ("pkg.Box[int,string]"). The returned name is the
// declaration key when the base name is an indexed generic AND the string
// carries type arguments (a bare "pkg.Box" is the uninstantiated generic
// itself, not an instantiation); "" otherwise.
func LookupGenericDecl(typeStr string, index map[string]int) string {
	base := stripQualifierPrefixes(typeStr)

	// Only instantiations count: the remaining base must end with type
	// arguments. Cut at the first '[' — nested arguments ("Box[Pair[int,T]]")
	// still leave the correct base name.
	cut := strings.Index(base, "[")
	if cut <= 0 {
		return ""
	}

	return lookupBareOrQualified(base[:cut], index)
}

// stripQualifierPrefixes consumes leading pointer, slice, array, map, and
// channel qualifiers from a type string, leaving the core named type.
// Bracket segments are matched by depth so "map[[2]int]pkg.Box" and
// "[2]pkg.Box" reduce correctly. Returns "" on unterminated brackets.
func stripQualifierPrefixes(s string) string {
	for {
		switch {
		case strings.HasPrefix(s, "*"), strings.HasPrefix(s, " "):
			s = s[1:]
		case strings.HasPrefix(s, "[]"):
			s = s[2:]
		case strings.HasPrefix(s, "map["):
			end := matchingBracket(s[len("map"):])
			if end < 0 {
				return ""
			}

			s = s[len("map")+end+1:]
		case strings.HasPrefix(s, "["):
			end := matchingBracket(s)
			if end < 0 {
				return ""
			}

			s = s[end+1:]
		case strings.HasPrefix(s, "chan "), strings.HasPrefix(s, "<-chan "):
			s = strings.TrimPrefix(strings.TrimPrefix(s, "<-"), "chan ")
		case strings.HasPrefix(s, "chan<- "):
			s = strings.TrimPrefix(s, "chan<- ")
		default:
			return s
		}
	}
}

// matchingBracket returns the index of the ']' matching the '[' at s[0], or -1
// when the brackets never close.
func matchingBracket(s string) int {
	if len(s) == 0 || s[0] != '[' {
		return -1
	}

	depth := 0

	for i := range s {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// lookupBareOrQualified matches a base name against the index, preferring the
// exact qualified form and falling back to the bare declaration name. The
// fallback covers same-run packages whose PkgPath was unavailable.
func lookupBareOrQualified(base string, index map[string]int) string {
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
// FormatGenericParams("Box", 1) → "Box[T1]", FormatGenericParams("Pair", 2) →
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
