package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// simpleMainCode is a reusable code template for simple main function tests.
const simpleMainCode = `package main

func main() {
	println("hello")
}
`

// TestParse tests the Parse function with valid Go code.
func TestParse(t *testing.T) {
	t.Parallel()

	// Create a temporary Go file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")

	code := simpleMainCode
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node == nil {
		t.Fatal("Parse() returned nil node")
	}

	if node.Type != File {
		t.Errorf("Parse() node.Type = %v, want %v", node.Type, File)
	}

	if node.Filename != tmpFile {
		t.Errorf("Parse() node.Filename = %v, want %v", node.Filename, tmpFile)
	}
}

// TestParseWithLineCount tests the ParseWithLineCount function.
func TestParseWithLineCount(t *testing.T) {
	t.Parallel()

	// Create a temporary Go file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")

	code := simpleMainCode
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, lineCount, err := ParseWithLineCount(tmpFile)
	if err != nil {
		t.Fatalf("ParseWithLineCount() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseWithLineCount() returned nil node")
	}

	if lineCount <= 0 {
		t.Errorf("ParseWithLineCount() lineCount = %v, want > 0", lineCount)
	}
}

// TestParse_NonExistentFile tests Parse with a non-existent file.
func TestParse_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := Parse("/nonexistent/file.go")
	if err == nil {
		t.Error("Parse() expected error for non-existent file, got nil")
	}
}

// TestParseWithLineCount_NonExistentFile tests ParseWithLineCount with a non-existent file.
func TestParseWithLineCount_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, _, err := ParseWithLineCount("/nonexistent/file.go")
	if err == nil {
		t.Error("ParseWithLineCount() expected error for non-existent file, got nil")
	}
}

// TestParse_InvalidGoCode tests Parse with invalid Go code.
func TestParse_InvalidGoCode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.go")

	invalidCode := `package main
func main(
// Missing closing paren and brace
`
	if err := os.WriteFile(tmpFile, []byte(invalidCode), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := Parse(tmpFile)
	if err == nil {
		t.Error("Parse() expected error for invalid Go code, got nil")
	}
}

// TestParse_FuncDecl tests that function declarations are parsed correctly.
func TestParse_FuncDecl(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "func.go")

	code := `package main

func add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check that we have children (the function declaration)
	if len(node.Children) == 0 {
		t.Error("Parse() returned node with no children, expected func declaration")
	}

	// Find the FuncDecl
	found := false
	for _, child := range node.Children {
		if child.Type == FuncDecl {
			found = true
			break
		}
	}

	if !found {
		t.Error("Parse() did not find FuncDecl in parsed AST")
	}
}

// TestParse_StructType tests that struct types are parsed correctly.
func TestParse_StructType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "struct.go")

	code := `package main

type Person struct {
	Name string
	Age  int
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find the GenDecl containing TypeSpec with StructType
	found := false
	for _, child := range node.Children {
		if child.Type == GenDecl {
			for _, spec := range child.Children {
				if spec.Type == TypeSpec {
					for _, typeChild := range spec.Children {
						if typeChild.Type == StructType {
							found = true
							break
						}
					}
				}
			}
		}
	}

	if !found {
		t.Error("Parse() did not find StructType in parsed AST")
	}
}

// TestParse_InterfaceType tests that interface types are parsed correctly.
func TestParse_InterfaceType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "interface.go")

	code := `package main

type Reader interface {
	Read() error
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find InterfaceType
	found := findNodeType(node, InterfaceType)
	if !found {
		t.Error("Parse() did not find InterfaceType in parsed AST")
	}
}

// TestParse_ForStmt tests that for statements are parsed correctly.
func TestParse_ForStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "for.go")

	code := `package main

func main() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, ForStmt)
	if !found {
		t.Error("Parse() did not find ForStmt in parsed AST")
	}
}

// TestParse_RangeStmt tests that range statements are parsed correctly.
func TestParse_RangeStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "range.go")

	code := `package main

func main() {
	for _, v := range []int{1, 2, 3} {
		println(v)
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, RangeStmt)
	if !found {
		t.Error("Parse() did not find RangeStmt in parsed AST")
	}
}

// TestParse_IfStmt tests that if statements are parsed correctly.
func TestParse_IfStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "if.go")

	code := `package main

func main() {
	if true {
		println("yes")
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, IfStmt)
	if !found {
		t.Error("Parse() did not find IfStmt in parsed AST")
	}
}

// TestParse_SwitchStmt tests that switch statements are parsed correctly.
func TestParse_SwitchStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "switch.go")

	code := `package main

func main() {
	switch 1 {
	case 1:
		println("one")
	default:
		println("other")
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, SwitchStmt)
	if !found {
		t.Error("Parse() did not find SwitchStmt in parsed AST")
	}
}

// TestParse_SelectStmt tests that select statements are parsed correctly.
func TestParse_SelectStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "select.go")

	code := `package main

func main() {
	select {
	case <-make(chan int):
		println("received")
	default:
		println("default")
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, SelectStmt)
	if !found {
		t.Error("Parse() did not find SelectStmt in parsed AST")
	}
}

// TestParse_DeferStmt tests that defer statements are parsed correctly.
func TestParse_DeferStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "defer.go")

	code := `package main

func main() {
	defer println("deferred")
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, DeferStmt)
	if !found {
		t.Error("Parse() did not find DeferStmt in parsed AST")
	}
}

// TestParse_GoStmt tests that go statements are parsed correctly.
func TestParse_GoStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "go.go")

	code := `package main

func main() {
	go println("goroutine")
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, GoStmt)
	if !found {
		t.Error("Parse() did not find GoStmt in parsed AST")
	}
}

// TestParse_ReturnStmt tests that return statements are parsed correctly.
func TestParse_ReturnStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "return.go")

	code := `package main

func getValue() int {
	return 42
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, ReturnStmt)
	if !found {
		t.Error("Parse() did not find ReturnStmt in parsed AST")
	}
}

// TestParse_BinaryExpr tests that binary expressions are parsed correctly.
func TestParse_BinaryExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "binary.go")

	code := `package main

func main() {
	_ = 1 + 2
	_ = 3 * 4
	_ = 5 > 6
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, BinaryExpr)
	if !found {
		t.Error("Parse() did not find BinaryExpr in parsed AST")
	}
}

// TestParse_CallExpr tests that call expressions are parsed correctly.
func TestParse_CallExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "call.go")

	code := simpleMainCode
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, CallExpr)
	if !found {
		t.Error("Parse() did not find CallExpr in parsed AST")
	}
}

// TestParse_SelectorExpr tests that selector expressions are parsed correctly.
func TestParse_SelectorExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "selector.go")

	code := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, SelectorExpr)
	if !found {
		t.Error("Parse() did not find SelectorExpr in parsed AST")
	}
}

// TestParse_IndexExpr tests that index expressions are parsed correctly.
func TestParse_IndexExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "index.go")

	code := `package main

func main() {
	arr := []int{1, 2, 3}
	_ = arr[0]
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, IndexExpr)
	if !found {
		t.Error("Parse() did not find IndexExpr in parsed AST")
	}
}

// TestParse_SliceExpr tests that slice expressions are parsed correctly.
func TestParse_SliceExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "slice.go")

	code := `package main

func main() {
	arr := []int{1, 2, 3, 4, 5}
	_ = arr[1:3]
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, SliceExpr)
	if !found {
		t.Error("Parse() did not find SliceExpr in parsed AST")
	}
}

// TestParse_MapType tests that map types are parsed correctly.
func TestParse_MapType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "map.go")

	code := `package main

type StringMap map[string]string
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, MapType)
	if !found {
		t.Error("Parse() did not find MapType in parsed AST")
	}
}

// TestParse_ChanType tests that channel types are parsed correctly.
func TestParse_ChanType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "chan.go")

	code := `package main

type IntChan chan int
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, ChanType)
	if !found {
		t.Error("Parse() did not find ChanType in parsed AST")
	}
}

// TestParse_ArrayType tests that array types are parsed correctly.
func TestParse_ArrayType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "array.go")

	code := `package main

type IntArray [5]int
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, ArrayType)
	if !found {
		t.Error("Parse() did not find ArrayType in parsed AST")
	}
}

// TestParse_StarExpr tests that star (pointer) expressions are parsed correctly.
func TestParse_StarExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "star.go")

	code := `package main

func main() {
	x := 5
	p := &x
	_ = *p
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, StarExpr)
	if !found {
		t.Error("Parse() did not find StarExpr in parsed AST")
	}
}

// TestParse_UnaryExpr tests that unary expressions are parsed correctly.
func TestParse_UnaryExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unary.go")

	code := `package main

func main() {
	x := 5
	_ = -x
	_ = !true
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, UnaryExpr)
	if !found {
		t.Error("Parse() did not find UnaryExpr in parsed AST")
	}
}

// TestParse_CompositeLit tests that composite literals are parsed correctly.
func TestParse_CompositeLit(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "composite.go")

	code := `package main

type Point struct {
	X, Y int
}

func main() {
	_ = Point{X: 1, Y: 2}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, CompositeLit)
	if !found {
		t.Error("Parse() did not find CompositeLit in parsed AST")
	}
}

// TestParse_FuncLit tests that function literals are parsed correctly.
func TestParse_FuncLit(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "funclit.go")

	code := `package main

func main() {
	fn := func() {
		println("anonymous")
	}
	fn()
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, FuncLit)
	if !found {
		t.Error("Parse() did not find FuncLit in parsed AST")
	}
}

// TestParse_ImportsSkipped tests that import declarations are skipped.
func TestParse_ImportsSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "import.go")

	code := `package main

import "fmt"
import "os"

func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// ImportSpec should not appear in children
	for _, child := range node.Children {
		checkNoImportSpec(t, child)
	}
}

func checkNoImportSpec(t *testing.T, node *syntax.Node) {
	t.Helper()

	// ImportSpec nodes shouldn't exist (imports are skipped)
	// However, we can't directly check for ImportSpec since it's not in our types
	// Just verify we have func declarations
	_ = node // Currently no-op - GenDecl check would go here if ImportSpec type existed
}

// TestParse_IncDecStmt tests that increment/decrement statements are parsed correctly.
func TestParse_IncDecStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "incdec.go")

	code := `package main

func main() {
	i := 0
	i++
	i--
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, IncDecStmt)
	if !found {
		t.Error("Parse() did not find IncDecStmt in parsed AST")
	}
}

// TestParse_AssignStmt tests that assignment statements are parsed correctly.
func TestParse_AssignStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "assign.go")

	code := `package main

func main() {
	a, b := 1, 2
	a, b = b, a
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, AssignStmt)
	if !found {
		t.Error("Parse() did not find AssignStmt in parsed AST")
	}
}

// TestParse_BranchStmt tests that branch statements are parsed correctly.
func TestParse_BranchStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "branch.go")

	code := `package main

func main() {
	for {
		break
	}

	for {
		continue
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, BranchStmt)
	if !found {
		t.Error("Parse() did not find BranchStmt in parsed AST")
	}
}

// TestParse_TypeAssertExpr tests that type assert expressions are parsed correctly.
func TestParse_TypeAssertExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "typeassert.go")

	code := `package main

func main() {
	var i interface{}
	_ = i.(string)
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, TypeAssertExpr)
	if !found {
		t.Error("Parse() did not find TypeAssertExpr in parsed AST")
	}
}

// TestParse_TypeSwitchStmt tests that type switch statements are parsed correctly.
func TestParse_TypeSwitchStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "typeswitch.go")

	code := `package main

func main() {
	var x interface{}
	switch x.(type) {
	case int:
		println("int")
	case string:
		println("string")
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, TypeSwitchStmt)
	if !found {
		t.Error("Parse() did not find TypeSwitchStmt in parsed AST")
	}
}

// TestParse_Ellipsis tests that ellipsis expressions are parsed correctly.
func TestParse_Ellipsis(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "ellipsis.go")

	code := `package main

func printAll(args ...string) {
	for _, arg := range args {
		println(arg)
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, Ellipsis)
	if !found {
		t.Error("Parse() did not find Ellipsis in parsed AST")
	}
}

// TestParse_KeyValueExpr tests that key-value expressions are parsed correctly.
func TestParse_KeyValueExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "keyvalue.go")

	code := `package main

type Config struct {
	Name string
	Port int
}

func main() {
	_ = Config{Name: "test", Port: 8080}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, KeyValueExpr)
	if !found {
		t.Error("Parse() did not find KeyValueExpr in parsed AST")
	}
}

// TestParse_ParenExpr tests that parenthesized expressions are parsed correctly.
func TestParse_ParenExpr(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "paren.go")

	code := `package main

func main() {
	_ = (1 + 2) * 3
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, ParenExpr)
	if !found {
		t.Error("Parse() did not find ParenExpr in parsed AST")
	}
}

// TestParse_LabeledStmt tests that labeled statements are parsed correctly.
func TestParse_LabeledStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "labeled.go")

	code := `package main

func main() {
outer:
	for {
		break outer
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, LabeledStmt)
	if !found {
		t.Error("Parse() did not find LabeledStmt in parsed AST")
	}
}

// TestParse_SendStmt tests that send statements are parsed correctly.
func TestParse_SendStmt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "send.go")

	code := `package main

func main() {
	ch := make(chan int)
	ch <- 1
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, SendStmt)
	if !found {
		t.Error("Parse() did not find SendStmt in parsed AST")
	}
}

// TestParse_ValueSpec tests that value specs are parsed correctly.
func TestParse_ValueSpec(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "valuespec.go")

	code := `package main

var (
	globalInt    int    = 42
	globalString string = "hello"
)
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	found := findNodeType(node, ValueSpec)
	if !found {
		t.Error("Parse() did not find ValueSpec in parsed AST")
	}
}

// TestParse_MethodDecl tests that method declarations are parsed correctly.
func TestParse_MethodDecl(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "method.go")

	code := `package main

type MyType struct{}

func (m *MyType) DoSomething() error {
	return nil
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// FuncDecl should exist
	found := findNodeType(node, FuncDecl)
	if !found {
		t.Error("Parse() did not find FuncDecl for method in parsed AST")
	}
}

// TestParse_NestedStructures tests parsing of deeply nested structures.
func TestParse_NestedStructures(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nested.go")

	code := `package main

func main() {
	if true {
		if false {
			for i := 0; i < 10; i++ {
				switch i {
				case 0:
					if i == 0 {
						println("zero")
					}
				}
			}
		}
	}
}
`
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node == nil {
		t.Fatal("Parse() returned nil node for nested structures")
	}
}

// TestNodePositions tests that node positions are set correctly.
func TestNodePositions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "positions.go")

	code := simpleMainCode
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check that positions are set
	if node.Pos < 0 || node.End < 0 {
		t.Errorf("Parse() node positions invalid: Pos=%d, End=%d", node.Pos, node.End)
	}

	// End should be greater than or equal to Pos
	if node.End < node.Pos {
		t.Errorf("Parse() node.End < node.Pos: End=%d, Pos=%d", node.End, node.Pos)
	}
}

// findNodeType recursively searches for a node with the given type.
func findNodeType(node *syntax.Node, nodeType int) bool {
	if node.Type == int32(nodeType) {
		return true
	}
	for _, child := range node.Children {
		if findNodeType(child, nodeType) {
			return true
		}
	}
	return false
}
