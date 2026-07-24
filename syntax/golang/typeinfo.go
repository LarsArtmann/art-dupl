package golang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// PreloadedAST holds a pre-parsed AST file together with its type-checking
// results. When provided to the transformer, the pre-loaded AST is used instead
// of re-parsing the file, and the TypeInfo is consulted to encode variable
// types into identifier hashes.
type PreloadedAST struct {
	File     *ast.File
	TypeInfo *types.Info
	Fset     *token.FileSet
}

// TypeAwareData is a map from absolute file path to its pre-loaded AST and type info.
type TypeAwareData map[string]*PreloadedAST

// LookupPreloaded returns the pre-loaded data for the given file path, or nil
// if type-aware data was not loaded for this file.
func (td TypeAwareData) LookupPreloaded(file string) *PreloadedAST {
	if td == nil {
		return nil
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return nil
	}

	return td[absFile]
}

// LoadTypeAwareData loads type information for the given Go files using
// go/packages. Returns a map from absolute file path to PreloadedAST.
//
// This is significantly slower than parsing alone (10-100x) because it runs the
// full Go type checker with import resolution. Only call when --type-aware is enabled.
func LoadTypeAwareData(files []string) (TypeAwareData, error) {
	if len(files) == 0 {
		return nil, nil
	}

	absFiles := make([]string, 0, len(files))
	seen := make(map[string]bool, len(files))

	for _, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			abs = f
		}

		if !seen[abs] {
			seen[abs] = true
			absFiles = append(absFiles, abs)
		}
	}

	patterns := make([]string, len(absFiles))
	for i, f := range absFiles {
		patterns[i] = "file=" + f
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("go/packages.Load failed: %w", err)
	}

	result := make(TypeAwareData, len(absFiles))

	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}

		for _, syntaxFile := range pkg.Syntax {
			if syntaxFile == nil {
				continue
			}

			filename := pkg.Fset.File(syntaxFile.Pos()).Name()
			absName, err := filepath.Abs(filename)
			if err != nil {
				absName = filename
			}

			result[absName] = &PreloadedAST{
				File:     syntaxFile,
				TypeInfo: pkg.TypesInfo,
				Fset:     pkg.Fset,
			}
		}
	}

	return result, nil
}

// identTypeString returns the canonical type string for the object that ident
// refers to, or "" if type information is unavailable. Uses the type checker's
// Uses map (identifier references) and falls back to Defs (definitions).
func identTypeString(info *types.Info, ident *ast.Ident) string {
	if info == nil || ident == nil {
		return ""
	}

	if obj, ok := info.Uses[ident]; ok && obj != nil {
		if t := obj.Type(); t != nil {
			return t.String()
		}
	}

	if obj, ok := info.Defs[ident]; ok && obj != nil {
		if t := obj.Type(); t != nil {
			return t.String()
		}
	}

	return ""
}
