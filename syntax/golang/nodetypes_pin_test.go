package golang

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// TestPinnedNodeTypeConstants guards the node-type constants that the syntax
// package pins as literals (the packages cannot share the iota block because
// golang imports syntax). Reordering or inserting into the nodetypes.go iota
// list breaks the serializer's nested-statement descent silently — this test
// makes it loud. The values must also stay unused by syntax/templ's parallel
// raw enum (templ assigns no node the BlockStmt value).
func TestPinnedNodeTypeConstants(t *testing.T) {
	t.Parallel()

	pins := map[string]struct {
		got  int32
		want int32
	}{
		"BlockStmt": {BlockStmt, 6},
		"GenDecl":   {GenDecl, 24},
	}

	for name, pin := range pins {
		if pin.got != pin.want {
			t.Errorf(
				"%s = %d, want %d (update syntax.genDeclNodeType/blockStmtNodeType in syntax.go when changing the iota block)",
				name,
				pin.got,
				pin.want,
			)
		}
	}

	if !syntax.IsStatementContainer(BlockStmt) {
		t.Error("IsStatementContainer(BlockStmt) must be true: block statements drive the nested-statement descent")
	}

	notContainers := map[string]int32{
		"GenDecl":      GenDecl,
		"FuncDecl":     FuncDecl,
		"File":         File,
		"CaseClause":   CaseClause,
		"CommClause":   CommClause,
		"IfStmt":       IfStmt,
		"Ident":        Ident,
		"CompositeLit": CompositeLit,
	}

	for name, typ := range notContainers {
		if syntax.IsStatementContainer(typ) {
			t.Errorf("IsStatementContainer(%s) must be false", name)
		}
	}
}
