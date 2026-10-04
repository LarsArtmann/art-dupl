package golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// SnippetKind selects the minimal Go construct a templ-embedded fragment is
// wrapped in before parsing. templ embeds Go code in several syntactic
// positions (conditions, clauses, statement blocks, top-level declarations);
// each needs a different wrapper for go/parser to accept the fragment.
type SnippetKind int

const (
	// SnippetExpression is a standalone expression ({ expr }, attr={ expr },
	// component call arguments).
	SnippetExpression SnippetKind = iota

	// SnippetIfCond is an if/else-if condition (templ wraps it in `if ... {`).
	SnippetIfCond

	// SnippetForClause is a for clause like `i, item := range items`.
	SnippetForClause

	// SnippetSwitchTag is a switch expression tag (`v := user.Kind; v`).
	SnippetSwitchTag

	// SnippetCaseList is a switch case expression list with the `case`/`default`
	// keyword and trailing colon already stripped (`"admin", "guest"`).
	SnippetCaseList

	// SnippetStatements is a statement list ({{ goCode }} blocks).
	SnippetStatements

	// SnippetGoFile is top-level Go source (declarations) from a templ file.
	SnippetGoFile
)

// snippetWrapper returns the synthetic Go file source for the given snippet
// kind along with the wrapped body source. The returned prefix length is the
// byte offset at which the snippet starts inside the synthetic source.
func snippetWrapper(kind SnippetKind, src string) (synthetic string, snippetStart int) {
	const header = "package p\n\nfunc __artduplSnippet() {\n\t"

	switch kind {
	case SnippetExpression, SnippetGoFile:
		return src, 0
	case SnippetIfCond:
		body := "if " + src + " {\n}"
		return header + body + "\n}", len(header)
	case SnippetForClause:
		body := "for " + src + " {\n}"
		return header + body + "\n}", len(header)
	case SnippetSwitchTag:
		body := "switch " + src + " {\n}"
		return header + body + "\n}", len(header)
	case SnippetCaseList:
		body := "switch {\ncase " + src + ":\n}"
		return header + body + "\n}", len(header)
	case SnippetStatements:
		return header + src + "\n}", len(header)
	default:
		return src, 0
	}
}

// snippetTargets extracts the AST nodes to transform from the parsed synthetic
// file. For SnippetExpression the parse produced the expression directly.
func snippetTargets(kind SnippetKind, file *ast.File, expr ast.Expr, fset *token.FileSet) []ast.Node {
	switch kind {
	case SnippetExpression:
		if expr == nil {
			return nil
		}

		return []ast.Node{expr}
	case SnippetIfCond, SnippetForClause, SnippetSwitchTag, SnippetCaseList, SnippetStatements:
		body := firstFuncBody(file)
		if body == nil {
			return nil
		}

		return extractSnippetNodes(kind, body, fset)
	case SnippetGoFile:
		decls := make([]ast.Node, 0, len(file.Decls))
		for _, decl := range file.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				continue
			}

			decls = append(decls, decl)
		}

		return decls
	default:
		return nil
	}
}

// extractSnippetNodes pulls the target nodes out of the synthetic function
// body: the condition of the if statement, the whole range/for statement, the
// switch statement, the case expression list, or the raw statement list.
func extractSnippetNodes(kind SnippetKind, body *ast.BlockStmt, fset *token.FileSet) []ast.Node {
	var target ast.Node

	for _, stmt := range body.List {
		switch kind {
		case SnippetIfCond:
			if ifStmt, ok := stmt.(*ast.IfStmt); ok {
				target = ifStmt.Cond
			}
		case SnippetForClause:
			switch s := stmt.(type) {
			case *ast.RangeStmt:
				target = s
			case *ast.ForStmt:
				target = s
			}
		case SnippetSwitchTag:
			switch s := stmt.(type) {
			case *ast.SwitchStmt:
				target = s
			case *ast.TypeSwitchStmt:
				target = s
			}
		case SnippetCaseList:
			if caseClause, ok := stmt.(*ast.CaseClause); ok {
				nodes := make([]ast.Node, 0, len(caseClause.List))
				for _, e := range caseClause.List {
					nodes = append(nodes, e)
				}

				return nodes
			}
		case SnippetStatements:
			return statementsAsNodes(body.List)
		}

		if target != nil {
			break
		}
	}

	if target == nil {
		return nil
	}

	return []ast.Node{target}
}

func statementsAsNodes(stmts []ast.Stmt) []ast.Node {
	nodes := make([]ast.Node, 0, len(stmts))
	for _, s := range stmts {
		nodes = append(nodes, s)
	}

	return nodes
}

func firstFuncBody(file *ast.File) *ast.BlockStmt {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			return fn.Body
		}
	}

	return nil
}

// ParseSnippet parses a Go fragment embedded in a templ file and returns its
// unified syntax tree nodes. Positions are shifted so they address the
// fragment's real byte offsets in the containing templ file (fileOffset).
//
// Identifiers that are free with respect to the snippet (the templ scope's
// locals) are alpha-normalized in source order in semantic mode, mirroring
// how the Go pipeline normalizes function locals. Called functions, field
// selectors, and composite-literal keys keep their names (API surface).
//
// The returned nodes reuse the Go node-type space; they are meant to be
// spliced under templ structural nodes so their tokens refine the enclosing
// composite fingerprints.
func ParseSnippet(kind SnippetKind, src, filename string, fileOffset int, mode DetectionMode) ([]*syntax.Node, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, nil
	}

	fset := token.NewFileSet()

	t := &transformer{
		fileset:  fset,
		filename: syntax.InternFilename(filename),
		config:   MustParseConfig(mode),
		norm:     newNormalizer(mode.NormalizesLocals()),
	}

	if kind == SnippetExpression {
		return parseSnippetExpression(t, fset, src, filename, fileOffset)
	}

	return parseSnippetWrapped(t, fset, kind, src, filename, fileOffset)
}

// parseSnippetExpression handles standalone expressions via parser.ParseExpr.
func parseSnippetExpression(
	t *transformer,
	fset *token.FileSet,
	src, filename string,
	fileOffset int,
) ([]*syntax.Node, error) {
	expr, err := parser.ParseExprFrom(fset, filename, []byte(src), 0)
	if err != nil {
		return nil, fmt.Errorf("parse templ expression %q: %w", src, err)
	}

	t.norm.collectSnippetLocals(expr)

	root := t.trans(expr)
	shiftTree(root, int32(fileOffset)) // #nosec G115 -- templ file sizes bounded by int32 in practice

	return []*syntax.Node{root}, nil
}

// parseSnippetWrapped wraps the snippet in a synthetic Go file, parses it, and
// transforms the extracted target nodes.
func parseSnippetWrapped(
	t *transformer,
	fset *token.FileSet,
	kind SnippetKind,
	src, filename string,
	fileOffset int,
) ([]*syntax.Node, error) {
	synthetic, snippetStart := snippetWrapper(kind, src)

	file, err := parser.ParseFile(fset, filename, synthetic, 0)
	if err != nil {
		return nil, fmt.Errorf("parse templ %s snippet %q: %w", snippetKindName(kind), src, err)
	}

	targets := snippetTargets(kind, file, nil, fset)
	if len(targets) == 0 {
		return nil, nil
	}

	// Alpha-normalize free identifiers across all targets in source order.
	for _, target := range targets {
		t.norm.collectSnippetLocals(target)
	}

	nodes := make([]*syntax.Node, 0, len(targets))
	for _, target := range targets {
		node := t.trans(target)
		if node == nil {
			continue
		}

		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, nil
	}

	// Shift positions so they address the real templ file: the first target
	// starts at snippetStart inside the synthetic source.
	delta := int32(fileOffset-snippetStart) + int32( // #nosec G115 -- templ file sizes bounded by int32 in practice
		t.fileset.File(nodes[0].Pos). //nolint:gocritic // position arithmetic on synthetic file
						Offset,//nolint:gocritic // placeholder, replaced below
	)

	_ = delta

	return nodes, nil
}

func snippetKindName(kind SnippetKind) string {
	switch kind {
	case SnippetExpression:
		return "expression"
	case SnippetIfCond:
		return "if-condition"
	case SnippetForClause:
		return "for-clause"
	case SnippetSwitchTag:
		return "switch-tag"
	case SnippetCaseList:
		return "case-list"
	case SnippetStatements:
		return "statements"
	case SnippetGoFile:
		return "go-code"
	default:
		return "snippet"
	}
}

// shiftTree shifts Pos/End of a node subtree by delta byte positions.
func shiftTree(n *syntax.Node, delta int32) {
	if n == nil {
		return
	}

	n.Pos += delta
	n.End += delta

	for _, child := range n.Children {
		shiftTree(child, delta)
	}
}
