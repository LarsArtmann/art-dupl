package golang

import (
	"testing"
)

func TestParse_MapType(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "map", `package main

type StringMap map[string]string
`, MapType)
}

func TestParse_ChanType(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "chan", `package main

type IntChan chan int
`, ChanType)
}

func TestParse_ArrayType(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "array", `package main

type IntArray [5]int
`, ArrayType)
}

func TestParse_ValueSpec(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "valuespec", `package main

var (
	globalInt    int    = 42
	globalString string = "hello"
)
`, ValueSpec)
}

func TestParse_MethodDecl(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "method", `package main

type MyType struct{}

func (m *MyType) DoSomething() error {
	return nil
}
`, FuncDecl)
}
