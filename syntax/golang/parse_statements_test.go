package golang

import (
	"testing"
)

func TestParse_ControlStatements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     string
		nodeType int
	}{
		{"ForStmt", `package main

func main() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}
`, ForStmt},
		{"RangeStmt", `package main

func main() {
	for _, v := range []int{1, 2, 3} {
		println(v)
	}
}
`, RangeStmt},
		{"IfStmt", `package main

func main() {
	if true {
		println("yes")
	}
}
`, IfStmt},
		{"SwitchStmt", `package main

func main() {
	switch 1 {
	case 1:
		println("one")
	default:
		println("other")
	}
}
`, SwitchStmt},
		{"SelectStmt", `package main

func main() {
	select {
	case <-make(chan int):
		println("received")
	default:
		println("default")
	}
}
`, SelectStmt},
		{"DeferStmt", `package main

func main() {
	defer println("deferred")
}
`, DeferStmt},
		{"GoStmt", `package main

func main() {
	go println("goroutine")
}
`, GoStmt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testParseNodeType(t, tt.name, tt.code, tt.nodeType)
		})
	}
}

func TestParse_ReturnStmt(t *testing.T) {
	t.Parallel()
	code := `package main

func foo() int {
	return 42
}
`
	testParseNodeType(t, "return", code, ReturnStmt)
}

func TestParse_ReturnStmt_Table(t *testing.T) {
	t.Parallel()
	code := `package main

func getValue() int {
	return 42
}
`
	testParseNodeType(t, "return_table", code, ReturnStmt)
}

func TestParse_IncDecStmt(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "incdec", `package main

func main() {
	i := 0
	i++
	i--
}
`, IncDecStmt)
}

func TestParse_AssignStmt(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "assign", `package main

func main() {
	a, b := 1, 2
	a, b = b, a
}
`, AssignStmt)
}

func TestParse_BranchStmt(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "branch", `package main

func main() {
	for {
		break
	}
	for {
		continue
	}
}
`, BranchStmt)
}

func TestParse_LabeledStmt(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "labeled", `package main

func main() {
outer:
	for {
		break outer
	}
}
`, LabeledStmt)
}

func TestParse_SendStmt(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "send", `package main

func main() {
	ch := make(chan int)
	ch <- 1
}
`, SendStmt)
}

func TestParse_TypeSwitchStmt(t *testing.T) {
	t.Parallel()
	testParseNodeType(t, "typeswitch", `package main

func main() {
	var x interface{}
	switch x.(type) {
	case int:
		println("int")
	case string:
		println("string")
	}
}
`, TypeSwitchStmt)
}
