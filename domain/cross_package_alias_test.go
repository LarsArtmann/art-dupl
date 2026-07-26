package domain_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/pkg/artdupl"
	syntaxgolang "github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestAliasedSentinelsAreIdentical verifies that every sentinel re-exported
// from domain/ into config/, pkg/artdupl/, or syntax/golang/ points to the
// exact same error value. This prevents the class of bug where two
// errors.New("same message") in different packages cause errors.Is to
// silently return false across package boundaries.
func TestAliasedSentinelsAreIdentical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		canon   error
		aliased error
	}{
		{"config.ErrInvalidDetectionMethod", domain.ErrInvalidDetectionMethod, config.ErrInvalidDetectionMethod},
		{"config.ErrInvalidDetectionMode", domain.ErrInvalidDetectionMode, config.ErrInvalidDetectionMode},
		{"config.ErrInvalidDiffMode", domain.ErrInvalidDiffMode, config.ErrInvalidDiffMode},
		{"config.ErrInvalidThreshold", domain.ErrInvalidThreshold, config.ErrInvalidThreshold},
		{"config.ErrThresholdTooLarge", domain.ErrThresholdTooLarge, config.ErrThresholdTooLarge},
		{"config.ErrInvalidOutputFormat", domain.ErrInvalidOutputFormat, config.ErrInvalidOutputFormat},
		{"config.ErrInvalidSortCriteria", domain.ErrInvalidSortCriteria, config.ErrInvalidSortCriteria},
		{"pkg/artdupl.ErrInvalidThreshold", domain.ErrInvalidThreshold, artdupl.ErrInvalidThreshold},
		{"pkg/artdupl.ErrThresholdTooLarge", domain.ErrThresholdTooLarge, artdupl.ErrThresholdTooLarge},
		{"pkg/artdupl.ErrCloneLineEndBeforeStart", domain.ErrLineEndBeforeStart, artdupl.ErrCloneLineEndBeforeStart},
		{"pkg/artdupl.ErrUnsupportedMethod", domain.ErrInvalidDetectionMethod, artdupl.ErrUnsupportedMethod},
		{"syntax/golang.ErrInvalidDetectionMode", domain.ErrInvalidDetectionMode, syntaxgolang.ErrInvalidDetectionMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if !errors.Is(tt.aliased, tt.canon) {
				t.Errorf("aliased sentinel %s does not match canonical via errors.Is", tt.name)
			}

			if !errors.Is(tt.canon, tt.aliased) {
				t.Errorf("canonical does not match aliased %s via errors.Is (reverse)", tt.name)
			}

			//nolint:errorlint // intentional pointer-identity check: aliases must be the same value
			if tt.canon != tt.aliased {
				t.Errorf("%s is a distinct error value, not an alias", tt.name)
			}
		})
	}
}

// TestNoDuplicateErrorNewMessages scans ALL production .go files in the repo
// for errors.New("literal") calls and fails if any two share the same message
// string. Duplicate messages with different pointers cause silent errors.Is
// failures across package boundaries — this was the root cause of the original
// ErrInvalidDetectionMode bug.
//
// This test is self-maintaining: it requires no hardcoded list. New sentinels
// are automatically discovered via AST parsing.
func TestNoDuplicateErrorNewMessages(t *testing.T) {
	t.Parallel()

	messages := map[string]string{} // message -> first location (file:line)

	rootDir := "../" // repo root relative to domain/

	walkErr := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}

		if d.IsDir() {
			switch d.Name() {
			case "vendor", ".git", "node_modules", "tmp":
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files — they may contain ad-hoc errors.New for assertions
		if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/bdd/") {
			return nil
		}

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil // skip unparseable files
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "New" {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "errors" {
				return true
			}

			if len(call.Args) == 0 {
				return true
			}

			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			msg, unquoteErr := strconv.Unquote(lit.Value)
			if unquoteErr != nil {
				return true
			}

			pos := fset.Position(call.Pos())
			location := fmt.Sprintf("%s:%d", pos.Filename, pos.Line)

			if prev, exists := messages[msg]; exists {
				t.Errorf("duplicate errors.New(%q):\n  first:  %s\n  second: %s\n"+
					"Two errors.New with the same message create distinct pointers — errors.Is will silently fail across packages. "+
					"Consolidate to a single definition in domain/ and alias it.",
					msg, prev, location)
			}

			messages[msg] = location

			return true
		})

		return nil
	})

	if walkErr != nil {
		t.Fatalf("filepath.WalkDir failed: %v", walkErr)
	}
}
