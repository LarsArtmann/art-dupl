// Package jsonv2gate hosts the AST quality gate that guards the ADR-0024
// JSON I/O invariant: production code must not import encoding/json/v2 (or
// jsontext) on paths that marshal time.Duration, because the pure v2 API has
// no default Duration representation ("no default representation" crash —
// the 2026-09-23 config save/load outage) and Go 1.27 removed the v2
// struct-tag grammar. The gate is a test so it runs in every `go test ./...`
// sweep, CI, and the nix flake check with zero extra plumbing.
package jsonv2gate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// toleratedV2Importers lists production (non-_test.go) files that may import
// encoding/json/v2 or encoding/json/v2/jsontext, each with the rationale.
// A NEW file showing up in the scan fails the test until it is either
// migrated to the v1 API (preferred; byte-identical output via
// internal/jsonutil for wire paths) or added here with a reviewed rationale.
var toleratedV2Importers = map[string]string{
	"config/enum_helpers.go":           "enum string-only payloads, no Duration",
	"config/config_migrate.go":         "raw jsontext.Value migration parsing, never marshals structs",
	"pkg/enum/enum.go":                 "enum string-only payloads, no Duration",
	"baseline/baseline.go":             "baseline format is string/int payloads, no Duration fields",
	"internal/testutil/assert.go":      "test-support package, string-only payloads",
	"internal/testutil/bdd_runners.go": "test-support package, string-only payloads",
}

// jsonv2ImportPaths are the banned-under-risk import paths.
var jsonv2ImportPaths = map[string]bool{
	"encoding/json/v2":          true,
	"encoding/json/v2/jsontext": true,
}

func TestNoDurationRiskyJSONV2Imports(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	var violations []string

	walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := entry.Name()

		if entry.IsDir() {
			switch name {
			case "vendor", ".git", "website", "examples", "node_modules", "testdata":
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}

		rel = filepath.ToSlash(rel)

		importsV2, durationRisk := scanFile(t, path)
		if !importsV2 {
			return nil
		}

		if durationRisk {
			violations = append(violations,
				rel+" imports encoding/json/v2 on a Duration-bearing path; "+
					"migrate to the v1 API (stdlib encoding/json or internal/jsonutil) - "+
					"the pure v2 API refuses time.Duration and Go 1.27 removed the format-tag grammar")

			return nil
		}

		if _, ok := toleratedV2Importers[rel]; !ok {
			violations = append(violations,
				rel+" imports encoding/json/v2 without a reviewed rationale; "+
					"migrate to the v1 API or add it to toleratedV2Importers with a rationale")
		}

		return nil
	})
	if walkErr != nil {
		t.Fatal(walkErr)
	}

	for _, violation := range violations {
		t.Error(violation)
	}
}

// scanFile parses one file and reports whether it imports a json/v2 path and
// whether it declares a struct field of type time.Duration with a json tag
// that is not "-" (i.e. a field that actually reaches the wire).
func scanFile(t *testing.T, path string) (bool, bool) {
	t.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	var (
		importsV2    bool
		durationRisk bool
	)

	for _, imp := range file.Imports {
		if jsonv2ImportPaths[strings.Trim(imp.Path.Value, `"`)] {
			importsV2 = true
		}
	}

	if !importsV2 {
		return false, false
	}

	ast.Inspect(file, func(node ast.Node) bool {
		field, ok := node.(*ast.Field)
		if !ok {
			return true
		}

		if !isTimeDuration(field.Type) {
			return true
		}

		tag := strings.Trim(field.Tag.Value, "`")
		if strings.Contains(tag, `json:"-"`) {
			return true
		}

		durationRisk = true

		return true
	})

	return importsV2, durationRisk
}

// isTimeDuration reports whether the expression is time.Duration.
func isTimeDuration(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)

	return ok && ident.Name == "time" && sel.Sel.Name == "Duration"
}
