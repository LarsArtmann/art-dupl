package golang

import (
	"go/ast"
	"go/types"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestIsInterfaceMethod_ValueReceiver(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Validator interface {
	Validate() error
}

type User struct{ name string }
func (u User) Validate() error { return nil }

type Product struct{ sku string }
func (p Product) Validate() error { return nil }

func (u User) GetName() string { return u.name }
`
	info, file := loadTypeInfoForTest(t, src)

	validateUser := findFuncDecl(t, file, "Validate", "User")
	if !IsInterfaceMethod(info, validateUser) {
		t.Error("User.Validate should be detected as an interface method")
	}

	validateProduct := findFuncDecl(t, file, "Validate", "Product")
	if !IsInterfaceMethod(info, validateProduct) {
		t.Error("Product.Validate should be detected as an interface method")
	}

	getName := findFuncDecl(t, file, "GetName", "User")
	if IsInterfaceMethod(info, getName) {
		t.Error("User.GetName should NOT be detected as an interface method")
	}
}

func TestIsInterfaceMethod_PointerReceiver(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Updater interface {
	Update() error
}

type Account struct{ id int }
func (a *Account) Update() error { return nil }
`
	info, file := loadTypeInfoForTest(t, src)

	update := findFuncDecl(t, file, "Update", "Account")
	if !IsInterfaceMethod(info, update) {
		t.Error("Account.Update (pointer receiver) should be detected as an interface method")
	}
}

func TestIsInterfaceMethod_NonInterfaceMethod(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Worker struct{ id int }
func (w Worker) Process() error { return nil }
func (w Worker) String() string { return "worker" }
`
	info, file := loadTypeInfoForTest(t, src)

	// Process is not part of any interface in this package.
	// String is in the static name list but there's no fmt.Stringer interface
	// declared in this package, so type-aware detection should return false
	// for both.
	process := findFuncDecl(t, file, "Process", "Worker")
	if IsInterfaceMethod(info, process) {
		t.Error("Worker.Process should NOT be detected as an interface method (no matching interface)")
	}

	stringMethod := findFuncDecl(t, file, "String", "Worker")
	if IsInterfaceMethod(info, stringMethod) {
		t.Error("Worker.String should NOT be detected as interface method (fmt.Stringer is not same-package)")
	}
}

func TestIsInterfaceMethod_EmbeddedInterface(t *testing.T) {
	t.Parallel()

	src := `package testpkg

import "io"

type ReadCloser interface {
	io.Reader
	io.Closer
}

type FileHandle struct{ path string }
func (f *FileHandle) Read(p []byte) (int, error) { return 0, nil }
func (f *FileHandle) Close() error               { return nil }
`
	info, file := loadTypeInfoForTest(t, src)

	// ReadCloser embeds io.Reader and io.Closer. go/types flattens embedded
	// methods, so Read and Close should be detected.
	read := findFuncDecl(t, file, "Read", "FileHandle")
	if !IsInterfaceMethod(info, read) {
		t.Error("FileHandle.Read should be detected via embedded interface")
	}

	closeMethod := findFuncDecl(t, file, "Close", "FileHandle")
	if !IsInterfaceMethod(info, closeMethod) {
		t.Error("FileHandle.Close should be detected via embedded interface")
	}
}

func TestIsInterfaceMethod_NilGuard(t *testing.T) {
	t.Parallel()

	if IsInterfaceMethod(nil, nil) {
		t.Error("IsInterfaceMethod(nil, nil) should return false")
	}
}

func TestIsInterfaceMethod_NoReceiver(t *testing.T) {
	t.Parallel()

	src := `package testpkg

func standalone() error { return nil }
`
	info, file := loadTypeInfoForTest(t, src)

	fn := findFuncDecl(t, file, "standalone", "")
	if IsInterfaceMethod(info, fn) {
		t.Error("standalone function (no receiver) should not be an interface method")
	}
}

func TestTransformer_SetsInterfaceMethodFlag(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Validator interface {
	Validate() error
}

type User struct{ name string }
func (u User) Validate() error { return nil }
func (u User) GetName() string { return u.name }
`
	dir := t.TempDir()
	path := dir + "/src.go"
	writeFile(t, path, src)

	typeData, err := LoadTypeAwareData([]string{path})
	if err != nil {
		t.Fatalf("LoadTypeAwareData failed: %v", err)
	}

	pre := typeData.LookupPreloaded(path)
	if pre == nil {
		t.Fatal("LookupPreloaded returned nil")
	}

	root := parsePreloadedTest(t, path, pre)

	validateNode := findFuncDeclNode(t, root, "Validate")
	if !validateNode.InterfaceMethod {
		t.Error("Validate node should have InterfaceMethod=true")
	}

	getNameNode := findFuncDeclNode(t, root, "GetName")
	if getNameNode.InterfaceMethod {
		t.Error("GetName node should have InterfaceMethod=false")
	}
}

func TestTransformer_NoTypeInfo_InterfaceMethodFalse(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Stringer interface {
	String() string
}

type Foo struct{}
func (f Foo) String() string { return "foo" }
`
	// parseSemantic uses nil typeInfo — InterfaceMethod must stay false.
	root := parseSemantic(t, src)

	node := findFuncDeclNode(t, root, "String")
	if node.InterfaceMethod {
		t.Error("InterfaceMethod should be false when typeInfo is nil (no --type-aware)")
	}
}

func TestTransformer_PropagatesInterfaceMethodToBody(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Validator interface {
	Validate() error
}

type User struct{ name string }
func (u User) Validate() error { return nil }
func (u User) GetName() string { return u.name }`
	dir := t.TempDir()
	path := dir + "/src.go"
	writeFile(t, path, src)

	typeData, err := LoadTypeAwareData([]string{path})
	if err != nil {
		t.Fatalf("LoadTypeAwareData failed: %v", err)
	}

	pre := typeData.LookupPreloaded(path)
	if pre == nil {
		t.Fatal("LookupPreloaded returned nil")
	}

	root := parsePreloadedTest(t, path, pre)

	validateNode := findFuncDeclNode(t, root, "Validate")
	if !validateNode.InterfaceMethod {
		t.Fatal("Validate FuncDecl should have InterfaceMethod=true")
	}

	for _, stmt := range findBodyStatements(validateNode) {
		if !stmt.InterfaceMethod {
			t.Error("body statement of interface method Validate should have InterfaceMethod=true")
		}
	}

	getNameNode := findFuncDeclNode(t, root, "GetName")
	if getNameNode.InterfaceMethod {
		t.Fatal("GetName FuncDecl should have InterfaceMethod=false")
	}

	for _, stmt := range findBodyStatements(getNameNode) {
		if stmt.InterfaceMethod {
			t.Error("body statement of non-interface method GetName should have InterfaceMethod=false")
		}
	}
}

func findBodyStatements(funcDecl *syntax.Node) []*syntax.Node {
	for _, child := range funcDecl.Children {
		if DecodeBaseType(child.Type) == BlockStmt {
			return child.Children
		}
	}

	return nil
}

// loadTypeInfoForTest writes the source to a temp file, loads type info via
// go/packages, and returns the types.Info and parsed AST file.
func loadTypeInfoForTest(t *testing.T, src string) (*types.Info, *ast.File) {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/src.go"
	writeFile(t, path, src)

	typeData, err := LoadTypeAwareData([]string{path})
	if err != nil {
		t.Fatalf("LoadTypeAwareData failed: %v", err)
	}

	pre := typeData.LookupPreloaded(path)
	if pre == nil {
		t.Fatal("LookupPreloaded returned nil")
	}

	return pre.TypeInfo, pre.File
}

// findFuncDecl walks the AST to find a FuncDecl by name and receiver type.
// Pass empty recvType for functions without a receiver.
func findFuncDecl(t *testing.T, file *ast.File, name, recvType string) *ast.FuncDecl {
	t.Helper()

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || fn.Name.Name != name {
			continue
		}

		if recvType == "" {
			if fn.Recv == nil {
				return fn
			}

			continue
		}

		if fn.Recv != nil && len(fn.Recv.List) > 0 {
			// Strip pointer indirection from receiver type name.
			rt := fn.Recv.List[0].Type

			if star, ok := rt.(*ast.StarExpr); ok {
				rt = star.X
			}

			if ident, ok := rt.(*ast.Ident); ok && ident.Name == recvType {
				return fn
			}
		}
	}

	t.Fatalf("FuncDecl %s on %q not found", name, recvType)

	return nil
}

// findFuncDeclNode walks the syntax.Node tree to find a FuncDecl by its
// decoded base type and Name field. Returns the first match.
func findFuncDeclNode(t *testing.T, root *syntax.Node, name string) *syntax.Node {
	t.Helper()

	var found *syntax.Node

	var walk func(n *syntax.Node)

	walk = func(n *syntax.Node) {
		if found != nil {
			return
		}

		if DecodeBaseType(n.Type) == FuncDecl && n.Name == name {
			found = n

			return
		}

		for _, c := range n.Children {
			walk(c)
		}
	}

	walk(root)

	if found == nil {
		t.Fatalf("syntax.Node FuncDecl %q not found in tree", name)
	}

	return found
}

// findFuncLitNodes walks the syntax.Node subtree and returns all FuncLit nodes.
func findFuncLitNodes(root *syntax.Node) []*syntax.Node {
	var found []*syntax.Node

	var walk func(n *syntax.Node)

	walk = func(n *syntax.Node) {
		if DecodeBaseType(n.Type) == FuncLit {
			found = append(found, n)
		}

		for _, c := range n.Children {
			walk(c)
		}
	}

	walk(root)

	return found
}

// TestTransformer_FuncLitResetsInterfaceMethodFlag verifies that closures
// (FuncLit) inside interface method bodies do NOT inherit the
// InterfaceMethod flag. The transformer resets enclosingInterfaceMethod to
// false in the FuncLit case because closures are not interface methods.
func TestTransformer_FuncLitResetsInterfaceMethodFlag(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Handler interface {
	Process(data []byte) error
}

type Worker struct{}

func (w Worker) Process(data []byte) error {
	check := func(b byte) bool {
		return b > 0
	}
	for _, b := range data {
		if check(b) {
			return nil
		}
	}
	return nil
}`
	dir := t.TempDir()
	path := dir + "/src.go"
	writeFile(t, path, src)

	typeData, err := LoadTypeAwareData([]string{path})
	if err != nil {
		t.Fatalf("LoadTypeAwareData failed: %v", err)
	}

	pre := typeData.LookupPreloaded(path)
	if pre == nil {
		t.Fatal("LookupPreloaded returned nil")
	}

	root := parsePreloadedTest(t, path, pre)

	// Find the Process FuncDecl — it satisfies Handler, so InterfaceMethod=true.
	processNode := findFuncDeclNode(t, root, "Process")
	if !processNode.InterfaceMethod {
		t.Fatal("Process FuncDecl should have InterfaceMethod=true")
	}

	// Body statements of Process should inherit InterfaceMethod=true.
	processBodyStmts := findBodyStatements(processNode)
	if len(processBodyStmts) == 0 {
		t.Fatal("Process body has no statements")
	}

	for _, stmt := range processBodyStmts {
		if !stmt.InterfaceMethod {
			t.Error("body statement of interface method Process should have InterfaceMethod=true")
		}
	}

	// Find the FuncLit (closure) inside Process.
	funcLits := findFuncLitNodes(processNode)
	if len(funcLits) == 0 {
		t.Fatal("no FuncLit found inside Process body")
	}

	for _, fl := range funcLits {
		// The FuncLit node itself inherits the flag from its parent context
		// (set at line 19 before the case-specific reset). The reset to false
		// only applies to the FuncLit's CHILDREN — body statements, params, etc.
		// This is correct: the actionability layer matches at the statement level,
		// and the FuncLit's body statements must NOT be flagged as interface methods.

		// The FuncLit's body statements should have InterfaceMethod=false.
		for _, stmt := range findBodyStatements(fl) {
			if stmt.InterfaceMethod {
				t.Error("FuncLit body statement should have InterfaceMethod=false " +
					"(closures inside interface methods are not themselves interface methods)")
			}
		}
	}
}
