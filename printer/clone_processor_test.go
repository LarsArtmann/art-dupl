package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestClassifyCloneType_RenamedFunctions(t *testing.T) {
	// Two functions with identical structure but renamed identifiers.
	// In semantic mode, the suffix tree matches on Type (alpha-normalized),
	// so these match. classifyCloneType should detect Name divergence → Type 2.
	//
	//   func a(x int) int { return x + 1 }
	//   func b(y int) int { return y + 1 }

	makeFunc := func(funcName, paramName string) []*syntax.Node {
		// FuncDecl root
		funcDecl := &syntax.Node{
			Type: golang.FuncDecl,
			Name: funcName,
		}
		// Params: Ident(paramName)
		param := &syntax.Node{
			Type: golang.Ident,
			Name: paramName,
		}
		// Body: ReturnStmt → BinaryExpr(+) → Ident(paramName), Literal(1)
		bodyIdent := &syntax.Node{
			Type: golang.Ident,
			Name: paramName,
		}
		returnStmt := &syntax.Node{
			Type:     golang.ReturnStmt,
			Children: []*syntax.Node{bodyIdent},
		}
		funcDecl.Children = []*syntax.Node{param, returnStmt}

		return []*syntax.Node{funcDecl}
	}

	dups := [][]*syntax.Node{
		makeFunc("a", "x"),
		makeFunc("b", "y"),
	}

	result := classifyCloneType(dups)

	if result != domain.CloneType2 {
		t.Errorf("expected CloneType2 (renamed), got %s", result)
	}
}

func TestClassifyCloneType_IdenticalNames(t *testing.T) {
	// Same function, same names → Type 1 (exact)
	makeFunc := func() []*syntax.Node {
		funcDecl := &syntax.Node{
			Type: golang.FuncDecl,
			Name: "foo",
		}
		param := &syntax.Node{
			Type: golang.Ident,
			Name: "x",
		}
		funcDecl.Children = []*syntax.Node{param}
		return []*syntax.Node{funcDecl}
	}

	dups := [][]*syntax.Node{
		makeFunc(),
		makeFunc(),
	}

	result := classifyCloneType(dups)

	if result != domain.CloneType1 {
		t.Errorf("expected CloneType1 (exact), got %s", result)
	}
}

func TestClassifyCloneType_DoesNotMutateNodes(t *testing.T) {
	// Verify that classifyCloneType does NOT mutate node.Type values,
	// unlike the old syntax.Serialize-based approach.
	makeStmt := func(name string) []*syntax.Node {
		inner := &syntax.Node{
			Type: golang.Ident,
			Name: name,
		}
		stmt := &syntax.Node{
			Type:      golang.AssignStmt,
			Name:      "",
			Statement: true,
			Children:  []*syntax.Node{inner},
		}
		return []*syntax.Node{stmt}
	}

	dups := [][]*syntax.Node{
		makeStmt("a"),
		makeStmt("b"),
	}

	originalType := dups[0][0].Type

	_ = classifyCloneType(dups)

	if dups[0][0].Type != originalType {
		t.Errorf(
			"classifyCloneType mutated node.Type: was %d, now %d",
			originalType, dups[0][0].Type,
		)
	}
}
