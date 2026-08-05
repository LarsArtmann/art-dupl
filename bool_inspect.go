package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func cloneNodeFromSyntax(n *syntax.Node) *domain.CloneNode {
	cn := &domain.CloneNode{
		BaseType: golang.DecodeBaseType(n.Type),
		Name:     n.Name,
	}
	if len(n.Children) > 0 {
		cn.Children = make([]*domain.CloneNode, len(n.Children))
		for i, c := range n.Children {
			cn.Children[i] = cloneNodeFromSyntax(c)
		}
	}
	return cn
}

func dump(cn *domain.CloneNode, depth int) {
	fmt.Printf("%sBaseType=%d Name=%q\n", indent(depth), cn.BaseType, cn.Name)
	for _, c := range cn.Children {
		dump(c, depth+1)
	}
}

func indent(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "  "
	}
	return s
}

func main() {
	src := `package fixture

func a() {
	hasFloatFormat := false
	hasSeparatorLoop := false
}

func b() {
	hasAll := false
	hasLinter := false
}
`
	dir, err := os.MkdirTemp("", "boolinspect")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "fixture.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		panic(err)
	}
	root, err := golang.Parse(path)
	if err != nil {
		panic(err)
	}
	var walk func(n *syntax.Node)
	walk = func(n *syntax.Node) {
		bt := golang.DecodeBaseType(n.Type)
		if bt == golang.AssignStmt {
			cn := cloneNodeFromSyntax(n)
			dump(cn, 0)
			fmt.Println("---")
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
}
