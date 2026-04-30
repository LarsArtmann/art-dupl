package golang

import (
	"go/ast"
	"go/token"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// trans transforms given golang AST to uniform tree structure.
//
//nolint:funlen,maintidx,nonamedreturns // High complexity is inherent to AST transformation
func (t *transformer) trans(
	node ast.Node,
) (o *syntax.Node) {
	o = syntax.NewNode()
	o.Filename = t.filename
	st, end := node.Pos(), node.End()
	o.Pos, o.End = int32(
		t.fileset.File(st).Offset(st),
	), int32(
		t.fileset.File(end).Offset(end),
	) // #nosec G115 -- File offsets bounded by int32 in practice

	switch n := node.(type) {
	case *ast.ArrayType:
		o.Type = ArrayType
		t.addWithNilCheck(o, n.Len)
		o.AddChildren(t.trans(n.Elt))

	case *ast.AssignStmt:
		o.Type = AssignStmt
		for _, e := range n.Rhs {
			o.AddChildren(t.trans(e))
		}

		for _, e := range n.Lhs {
			o.AddChildren(t.trans(e))
		}

	case *ast.BasicLit:
		o.Type = BasicLit

	case *ast.BinaryExpr:
		o.Type = BinaryExpr
		o.AddChildren(t.trans(n.X), t.trans(n.Y))

	case *ast.BlockStmt:
		o.Type = BlockStmt
		for _, stmt := range n.List {
			o.AddChildren(t.trans(stmt))
		}

	case *ast.BranchStmt:
		o.Type = BranchStmt
		t.addWithNilCheck(o, n.Label)

	case *ast.CallExpr:
		o.Type = CallExpr
		o.AddChildren(t.trans(n.Fun))

		for _, arg := range n.Args {
			o.AddChildren(t.trans(arg))
		}

	case *ast.CaseClause:
		o.Type = CaseClause
		for _, e := range n.List {
			o.AddChildren(t.trans(e))
		}

		t.addBodyStatements(o, n.Body)

	case *ast.ChanType:
		o.Type = ChanType
		o.AddChildren(t.trans(n.Value))

	case *ast.CommClause:
		o.Type = CommClause
		t.addWithNilCheck(o, n.Comm)

		t.addBodyStatements(o, n.Body)

	case *ast.CompositeLit:
		o.Type = CompositeLit
		t.addWithNilCheck(o, n.Type)

		for _, e := range n.Elts {
			o.AddChildren(t.trans(e))
		}

	case *ast.DeclStmt:
		o.Type = DeclStmt
		o.AddChildren(t.trans(n.Decl))

	case *ast.DeferStmt:
		o.Type = DeferStmt
		o.AddChildren(t.trans(n.Call))

	case *ast.Ellipsis:
		o.Type = Ellipsis
		t.addWithNilCheck(o, n.Elt)

	case *ast.EmptyStmt:
		o.Type = EmptyStmt

	case *ast.ExprStmt:
		o.Type = ExprStmt
		o.AddChildren(t.trans(n.X))

	case *ast.Field:
		o.Type = Field
		t.addIdentifierNames(o, n.Names)

		o.AddChildren(t.trans(n.Type))

	case *ast.FieldList:
		o.Type = FieldList
		for _, field := range n.List {
			o.AddChildren(t.trans(field))
		}

	case *ast.File:
		o.Type = File

		for _, decl := range n.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				// skip import declarations
				continue
			}

			o.AddChildren(t.trans(decl))
		}

	case *ast.ForStmt:
		o.Type = ForStmt
		t.addWithNilCheck(o, n.Init)
		t.addWithNilCheck(o, n.Cond)
		t.addWithNilCheck(o, n.Post)
		o.AddChildren(t.trans(n.Body))

	case *ast.FuncDecl:
		// Semantic hashing: combine receiver type and function name
		// This makes methods on different types semantically distinct,
		// reducing false positives for template patterns like enums.
		receiverType := extractReceiverTypeName(n.Recv)
		funcName := n.Name.Name
		o.Type = encodeSemanticTypeMulti(
			FuncDecl,
			t.config.Mode.IsSemantic(),
			receiverType,
			funcName,
		)
		t.addWithNilCheck(o, n.Recv)
		o.AddChildren(t.trans(n.Name), t.trans(n.Type))
		t.addWithNilCheck(o, n.Body)

	case *ast.FuncLit:
		o.Type = FuncLit
		o.AddChildren(t.trans(n.Type), t.trans(n.Body))

	case *ast.FuncType:
		o.Type = FuncType
		t.addWithNilCheck(o, n.TypeParams)
		o.AddChildren(t.trans(n.Params))
		t.addWithNilCheck(o, n.Results)

	case *ast.GenDecl:
		o.Type = GenDecl
		for _, spec := range n.Specs {
			o.AddChildren(t.trans(spec))
		}

	case *ast.GoStmt:
		o.Type = GoStmt
		o.AddChildren(t.trans(n.Call))

	case *ast.Ident:
		o.Type = encodeSemanticType(Ident, n.Name, t.config.Mode.IsSemantic())

	case *ast.IfStmt:
		o.Type = IfStmt
		t.addWithNilCheck(o, n.Init)
		t.addWithNilCheck(o, n.Cond)
		t.addWithNilCheck(o, n.Body)
		t.addWithNilCheck(o, n.Else)

	case *ast.IncDecStmt:
		o.Type = IncDecStmt
		o.AddChildren(t.trans(n.X))

	case *ast.IndexExpr:
		o.Type = IndexExpr
		o.AddChildren(t.trans(n.X), t.trans(n.Index))

	case *ast.IndexListExpr:
		o.Type = IndexListExpr
		o.AddChildren(t.trans(n.X))

		for _, idx := range n.Indices {
			o.AddChildren(t.trans(idx))
		}

	case *ast.InterfaceType:
		o.Type = InterfaceType
		o.AddChildren(t.trans(n.Methods))

	case *ast.KeyValueExpr:
		o.Type = KeyValueExpr
		t.addKeyValue(o, n.Key, n.Value)

	case *ast.LabeledStmt:
		o.Type = LabeledStmt
		o.AddChildren(t.trans(n.Label), t.trans(n.Stmt))

	case *ast.MapType:
		o.Type = MapType
		t.addKeyValue(o, n.Key, n.Value)

	case *ast.ParenExpr:
		o.Type = ParenExpr
		o.AddChildren(t.trans(n.X))

	case *ast.RangeStmt:
		o.Type = RangeStmt
		t.addWithNilCheck(o, n.Key)
		t.addWithNilCheck(o, n.Value)
		o.AddChildren(t.trans(n.X), t.trans(n.Body))

	case *ast.ReturnStmt:
		o.Type = ReturnStmt
		for _, e := range n.Results {
			o.AddChildren(t.trans(e))
		}

	case *ast.SelectStmt:
		o.Type = SelectStmt
		o.AddChildren(t.trans(n.Body))

	case *ast.SelectorExpr:
		o.Type = encodeSemanticType(SelectorExpr, n.Sel.Name, t.config.Mode.IsSemantic())
		o.AddChildren(t.trans(n.X), t.trans(n.Sel))

	case *ast.SendStmt:
		o.Type = SendStmt
		o.AddChildren(t.trans(n.Chan), t.trans(n.Value))

	case *ast.SliceExpr:
		o.Type = SliceExpr
		o.AddChildren(t.trans(n.X))
		t.addWithNilCheck(o, n.Low)
		t.addWithNilCheck(o, n.High)
		t.addWithNilCheck(o, n.Max)

	case *ast.StarExpr:
		o.Type = StarExpr
		o.AddChildren(t.trans(n.X))

	case *ast.StructType:
		o.Type = StructType
		o.AddChildren(t.trans(n.Fields))

	case *ast.SwitchStmt:
		o.Type = SwitchStmt
		t.addWithNilCheck(o, n.Init)
		t.addWithNilCheck(o, n.Tag)
		o.AddChildren(t.trans(n.Body))

	case *ast.TypeAssertExpr:
		o.Type = TypeAssertExpr
		o.AddChildren(t.trans(n.X))
		t.addWithNilCheck(o, n.Type)

	case *ast.TypeSpec:
		// Semantic hashing: encode type name to differentiate type declarations
		o.Type = encodeSemanticType(TypeSpec, n.Name.Name, t.config.Mode.IsSemantic())
		o.AddChildren(t.trans(n.Name))
		t.addWithNilCheck(o, n.TypeParams)
		o.AddChildren(t.trans(n.Type))

	case *ast.TypeSwitchStmt:
		o.Type = TypeSwitchStmt
		t.addWithNilCheck(o, n.Init)
		o.AddChildren(t.trans(n.Assign), t.trans(n.Body))

	case *ast.UnaryExpr:
		o.Type = UnaryExpr
		o.AddChildren(t.trans(n.X))

	case *ast.ValueSpec:
		o.Type = ValueSpec
		t.addIdentifierNames(o, n.Names)

		t.addWithNilCheck(o, n.Type)

		for _, val := range n.Values {
			o.AddChildren(t.trans(val))
		}

	default:
		o.Type = BadNode
	}

	return o
}

// extractReceiverTypeName extracts the type name from a method receiver.
// Returns empty string if there's no receiver or if the type cannot be determined.
// Handles both value receivers (t Type) and pointer receivers (t *Type).
func extractReceiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	// Get the first field in the receiver (there's typically only one)
	field := recv.List[0]
	if field.Type == nil {
		return ""
	}

	// Handle pointer receiver: *Type
	if starExpr, ok := field.Type.(*ast.StarExpr); ok {
		if ident, ok := starExpr.X.(*ast.Ident); ok {
			return ident.Name
		}

		return ""
	}

	// Handle value receiver: Type
	if ident, ok := field.Type.(*ast.Ident); ok {
		return ident.Name
	}

	return ""
}

// addBodyStatements transforms all statements in a body and adds them as children.
func (t *transformer) addBodyStatements(o *syntax.Node, body []ast.Stmt) {
	for _, stmt := range body {
		o.AddChildren(t.trans(stmt))
	}
}

// addIdentifierNames transforms all identifier names and adds them as children.
func (t *transformer) addIdentifierNames(o *syntax.Node, names []*ast.Ident) {
	for _, name := range names {
		o.AddChildren(t.trans(name))
	}
}

// addKeyValue adds transformed key and value nodes as children.
func (t *transformer) addKeyValue(o *syntax.Node, key, value ast.Expr) {
	o.AddChildren(t.trans(key), t.trans(value))
}
