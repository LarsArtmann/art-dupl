package golang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
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

	// PkgPath is the import path of the package the file belongs to. Used to
	// qualify generic-declaration names so the reuse-existing-generic verdict
	// can match them against fully-qualified type strings in VarType
	// ("github.com/x/pkg.Box[int]" resolves to the indexed key
	// "github.com/x/pkg.Box"). Empty when unknown.
	PkgPath string

	// EraseHash, when true, populates VarType on nodes but does NOT encode the
	// type string into the identifier hash. Used by --suggest-generics so that
	// clones with different concrete types still match, while the type info is
	// preserved for post-detection generics-extraction classification.
	EraseHash bool
}

// TypeAwareData is a map from absolute file path to its pre-loaded AST and type info.
type TypeAwareData map[string]*PreloadedAST

// WithEraseHash returns a shallow copy of the map with every PreloadedAST's
// EraseHash flag set to the given value. The underlying ASTs, file sets, and
// type info are shared (read-only by the transformer), so this never
// re-type-checks — it exists so the combined type-aware + suggest-generics
// mode can derive both hash dispositions from a single go/packages load,
// which is the expensive part (10-100x parsing).
func (td TypeAwareData) WithEraseHash(eraseHash bool) TypeAwareData {
	if td == nil {
		return nil
	}

	result := make(TypeAwareData, len(td))

	for path, pre := range td {
		shallow := *pre
		shallow.EraseHash = eraseHash
		result[path] = &shallow
	}

	return result
}

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
// When eraseHash is true, each PreloadedAST has EraseHash set so that the
// transformer populates VarType on nodes without encoding the type into the
// identifier hash. This is used by --suggest-generics to detect generics-
// extraction candidates that --type-aware would suppress.
//
// Files are grouped by their nearest enclosing module (nearest ancestor
// directory with a go.mod) and one go/packages load runs per module with
// Config.Dir set to the module root. Without this, `file=` patterns for files
// from DIFFERENT packages collapse into a single synthesized
// "command-line-arguments" package with type errors whenever the load's
// working directory is outside the target module — cross-package identifiers
// then resolve against one merged symbol table and types.Info.Instances stays
// empty or wrong (found while wiring the instantiation marker,
// generics-precision cluster T7).
//
// This is significantly slower than parsing alone (10-100x) because it runs the
// full Go type checker with import resolution.
func LoadTypeAwareData(files []string, eraseHash bool) (TypeAwareData, error) {
	if len(files) == 0 {
		return TypeAwareData{}, nil
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

	result := make(TypeAwareData, len(absFiles))

	for moduleDir, moduleFiles := range groupFilesByModule(absFiles) {
		if err := loadModuleTypeAware(moduleDir, moduleFiles, eraseHash, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// groupFilesByModule buckets absolute files by their nearest ancestor
// directory containing a go.mod. Files with no enclosing module share the ""
// bucket and load relative to the process working directory (legacy
// behavior). The module lookup is cached per directory — repository trees
// share directory prefixes, so the walk runs once per distinct ancestor
// chain, not once per file.
func groupFilesByModule(absFiles []string) map[string][]string {
	groups := make(map[string][]string)

	cache := make(map[string]string)

	moduleFor := func(dir string) string {
		if m, ok := cache[dir]; ok {
			return m
		}

		m := nearestModuleDir(dir)
		cache[dir] = m

		return m
	}

	for _, f := range absFiles {
		module := moduleFor(filepath.Dir(f))
		groups[module] = append(groups[module], f)
	}

	return groups
}

// nearestModuleDir walks up from dir and returns the first ancestor holding a
// go.mod, or "" when none exists up to the filesystem root.
func nearestModuleDir(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}

		dir = parent
	}
}

// loadModuleTypeAware loads one module's files with Config.Dir anchored at
// the module root so `file=` patterns resolve to their REAL packages instead
// of a synthesized command-line-arguments merge. An empty moduleDir keeps the
// legacy cwd-relative load.
func loadModuleTypeAware(
	moduleDir string,
	files []string,
	eraseHash bool,
	result TypeAwareData,
) error {
	patterns := make([]string, len(files))
	for i, f := range files {
		patterns[i] = "file=" + f
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	if moduleDir != "" {
		cfg.Dir = moduleDir
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return fmt.Errorf("go/packages.Load failed: %w", err)
	}

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
				File:      syntaxFile,
				TypeInfo:  pkg.TypesInfo,
				Fset:      pkg.Fset,
				PkgPath:   pkg.PkgPath,
				EraseHash: eraseHash,
			}
		}
	}

	return nil
}

// identTypeString returns the canonical type string for the object that ident
// refers to, or "" if type information is unavailable. Uses the type checker's
// Uses map (identifier references) and falls back to Defs (definitions).
//
// Aliases are resolved via types.Unalias: a variable declared through an alias
// (`type MyDur = time.Duration`) and one declared directly must produce the
// SAME string, otherwise the generics classifier would count them as a type
// divergence (false candidate — plan T4, generics-precision cluster). Type
// identity, not alias spelling, is what divergence comparison means.
func identTypeString(info *types.Info, ident *ast.Ident) string {
	if info == nil || ident == nil {
		return ""
	}

	if obj, ok := info.Uses[ident]; ok && obj != nil {
		if t := obj.Type(); t != nil {
			return types.Unalias(t).String()
		}
	}

	if obj, ok := info.Defs[ident]; ok && obj != nil {
		if t := obj.Type(); t != nil {
			return types.Unalias(t).String()
		}
	}

	return ""
}
