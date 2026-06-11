package golang

const (
	BadNode = iota
	File
	ArrayType
	AssignStmt
	BasicLit
	BinaryExpr
	BlockStmt
	BranchStmt
	CallExpr
	CaseClause
	ChanType
	CommClause
	CompositeLit
	DeclStmt
	DeferStmt
	Ellipsis
	EmptyStmt
	ExprStmt
	Field
	FieldList
	ForStmt
	FuncDecl
	FuncLit
	FuncType
	GenDecl
	GoStmt
	Ident
	IfStmt
	IncDecStmt
	IndexExpr
	IndexListExpr
	InterfaceType
	KeyValueExpr
	LabeledStmt
	MapType
	ParenExpr
	RangeStmt
	ReturnStmt
	SelectStmt
	SelectorExpr
	SendStmt
	SliceExpr
	StarExpr
	StructType
	SwitchStmt
	TypeAssertExpr
	TypeSpec
	TypeSwitchStmt
	UnaryExpr
	ValueSpec
)

// nodeTypeNames maps AST node type constants to their string names.
var nodeTypeNames = map[int32]string{
	BadNode:        "BadNode",
	File:           "File",
	ArrayType:      "ArrayType",
	AssignStmt:     "AssignStmt",
	BasicLit:       "BasicLit",
	BinaryExpr:     "BinaryExpr",
	BlockStmt:      "BlockStmt",
	BranchStmt:     "BranchStmt",
	CallExpr:       "CallExpr",
	CaseClause:     "CaseClause",
	ChanType:       "ChanType",
	CommClause:     "CommClause",
	CompositeLit:   "CompositeLit",
	DeclStmt:       "DeclStmt",
	DeferStmt:      "DeferStmt",
	Ellipsis:       "Ellipsis",
	EmptyStmt:      "EmptyStmt",
	ExprStmt:       "ExprStmt",
	Field:          "Field",
	FieldList:      "FieldList",
	ForStmt:        "ForStmt",
	FuncDecl:       "FuncDecl",
	FuncLit:        "FuncLit",
	FuncType:       "FuncType",
	GenDecl:        "GenDecl",
	GoStmt:         "GoStmt",
	Ident:          "Ident",
	IfStmt:         "IfStmt",
	IncDecStmt:     "IncDecStmt",
	IndexExpr:      "IndexExpr",
	IndexListExpr:  "IndexListExpr",
	InterfaceType:  "InterfaceType",
	KeyValueExpr:   "KeyValueExpr",
	LabeledStmt:    "LabeledStmt",
	MapType:        "MapType",
	ParenExpr:      "ParenExpr",
	RangeStmt:      "RangeStmt",
	ReturnStmt:     "ReturnStmt",
	SelectStmt:     "SelectStmt",
	SelectorExpr:   "SelectorExpr",
	SendStmt:       "SendStmt",
	SliceExpr:      "SliceExpr",
	StarExpr:       "StarExpr",
	StructType:     "StructType",
	SwitchStmt:     "SwitchStmt",
	TypeAssertExpr: "TypeAssertExpr",
	TypeSpec:       "TypeSpec",
	TypeSwitchStmt: "TypeSwitchStmt",
	UnaryExpr:      "UnaryExpr",
	ValueSpec:      "ValueSpec",
}

// TypeName returns the human-readable name for a Go AST node type constant.
// Returns "Unknown" for unrecognized types.
func TypeName(nodeType int32) string {
	if name, ok := nodeTypeNames[nodeType]; ok {
		return name
	}

	return "Unknown"
}
