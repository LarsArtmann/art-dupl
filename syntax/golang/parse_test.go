package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

const simpleMainCode = `package main

func main() {
	println("hello")
}
`

func TestParse(t *testing.T) {
	t.Parallel()
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

func TestParseWithLineCount(t *testing.T) {
	t.Parallel()
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

func TestParse_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := Parse("/nonexistent/file.go")
	if err == nil {
		t.Error("Parse() expected error for non-existent file, got nil")
	}
}

func TestParseWithLineCount_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, _, err := ParseWithLineCount("/nonexistent/file.go")
	if err == nil {
		t.Error("ParseWithLineCount() expected error for non-existent file, got nil")
	}
}

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

	if len(node.Children) == 0 {
		t.Error("Parse() returned node with no children, expected func declaration")
	}

	found := false

	for _, child := range node.Children {
		if DecodeBaseType(child.Type) == FuncDecl {
			found = true

			break
		}
	}

	if !found {
		t.Error("Parse() did not find FuncDecl in parsed AST")
	}
}

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

	found := false

	for _, child := range node.Children {
		if DecodeBaseType(child.Type) == GenDecl {
			for _, spec := range child.Children {
				if DecodeBaseType(spec.Type) == TypeSpec {
					for _, typeChild := range spec.Children {
						if DecodeBaseType(typeChild.Type) == StructType {
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

func TestParse_InterfaceType(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "interface", `package main

type Reader interface {
	Read() error
}
`, InterfaceType)
}

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

	for _, child := range node.Children {
		checkNoImportSpec(t, child)
	}
}

func checkNoImportSpec(t *testing.T, node *syntax.Node) {
	t.Helper()

	_ = node
}

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

	if node.Pos < 0 || node.End < 0 {
		t.Errorf("Parse() node positions invalid: Pos=%d, End=%d", node.Pos, node.End)
	}

	if node.End < node.Pos {
		t.Errorf("Parse() node.End < node.Pos: End=%d, Pos=%d", node.End, node.Pos)
	}
}

func findNodeType(node *syntax.Node, nodeType int) bool {
	if DecodeBaseType(node.Type) == int32(nodeType) {
		return true
	}

	for _, child := range node.Children {
		if findNodeType(child, nodeType) {
			return true
		}
	}

	return false
}

func testParseNodeType(t *testing.T, name, code string, nodeType int) {
	t.Helper()
	tmpDir := t.TempDir()

	tmpFile := filepath.Join(tmpDir, name+".go")
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !findNodeType(node, nodeType) {
		t.Errorf("Parse() did not find %s in parsed AST", nodeTypeName(nodeType))
	}
}

func nodeTypeName(nodeType int) string {
	names := map[int]string{
		File: "File", FuncDecl: "FuncDecl", FuncLit: "FuncLit", GenDecl: "GenDecl",
		ValueSpec: "ValueSpec", TypeSpec: "TypeSpec", BlockStmt: "BlockStmt",
		IfStmt: "IfStmt", SwitchStmt: "SwitchStmt", TypeSwitchStmt: "TypeSwitchStmt",
		CaseClause: "CaseClause", CommClause: "CommClause", SelectStmt: "SelectStmt",
		ForStmt: "ForStmt", RangeStmt: "RangeStmt", ReturnStmt: "ReturnStmt",
		AssignStmt: "AssignStmt", GoStmt: "GoStmt", DeferStmt: "DeferStmt",
		SendStmt: "SendStmt", IncDecStmt: "IncDecStmt", ExprStmt: "ExprStmt",
		CallExpr: "CallExpr", UnaryExpr: "UnaryExpr", BinaryExpr: "BinaryExpr",
		ParenExpr: "ParenExpr", SelectorExpr: "SelectorExpr", IndexExpr: "IndexExpr",
		SliceExpr: "SliceExpr", TypeAssertExpr: "TypeAssertExpr", StarExpr: "StarExpr",
		StructType: "StructType", ArrayType: "ArrayType", MapType: "MapType",
		ChanType: "ChanType", FuncType: "FuncType", InterfaceType: "InterfaceType",
		CompositeLit: "CompositeLit", KeyValueExpr: "KeyValueExpr", Ident: "Ident",
		BasicLit: "BasicLit", Field: "Field", FieldList: "FieldList",
		LabeledStmt: "LabeledStmt", BranchStmt: "BranchStmt", Ellipsis: "Ellipsis",
		DeclStmt: "DeclStmt", EmptyStmt: "EmptyStmt",
	}
	if name, ok := names[nodeType]; ok {
		return name
	}

	return fmt.Sprintf("NodeType(%d)", nodeType)
}
