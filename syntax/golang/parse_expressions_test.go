package golang

import (
	"testing"
)

func TestParse_BinaryExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "binary", `package main

func main() {
	_ = 1 + 2
	_ = 3 * 4
	_ = 5 > 6
}
`, BinaryExpr)
}

func TestParse_CallExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "call", simpleMainCode, CallExpr)
}

func TestParse_SelectorExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "selector", `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`, SelectorExpr)
}

func TestParse_IndexExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "index", `package main

func main() {
	arr := []int{1, 2, 3}
	_ = arr[0]
}
`, IndexExpr)
}

func TestParse_SliceExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "slice", `package main

func main() {
	arr := []int{1, 2, 3, 4, 5}
	_ = arr[1:3]
}
`, SliceExpr)
}

func TestParse_StarExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "star", `package main

func main() {
	x := 5
	p := &x
	_ = *p
}
`, StarExpr)
}

func TestParse_UnaryExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "unary", `package main

func main() {
	x := 5
	_ = -x
	_ = !true
}
`, UnaryExpr)
}

func TestParse_CompositeLit(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "composite", `package main

type Point struct {
	X, Y int
}

func main() {
	_ = Point{X: 1, Y: 2}
}
`, CompositeLit)
}

func TestParse_FuncLit(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "funclit", `package main

func main() {
	fn := func() {
		println("anonymous")
	}
	fn()
}
`, FuncLit)
}

func TestParse_TypeAssertExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "typeassert", `package main

func main() {
	var i interface{}
	_ = i.(string)
}
`, TypeAssertExpr)
}

func TestParse_Ellipsis(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "ellipsis", `package main

func printAll(args ...string) {
	for _, arg := range args {
		println(arg)
	}
}
`, Ellipsis)
}

func TestParse_KeyValueExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "keyvalue", `package main

type Config struct {
	Name string
	Port int
}

func main() {
	_ = Config{Name: "test", Port: 8080}
}
`, KeyValueExpr)
}

func TestParse_ParenExpr(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "paren", `package main

func main() {
	_ = (1 + 2) * 3
}
`, ParenExpr)
}
