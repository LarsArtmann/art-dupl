package golang

import (
	"fmt"
	"os"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestStatementTokenCount(t *testing.T) {
	src := `package toktest
import "fmt"
func processUser(name string, age int) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	if age < 0 {
		return fmt.Errorf("negative age")
	}
	fmt.Println(name, age)
	return nil
}
func processOrder(id string, total int) error {
	if id == "" {
		return fmt.Errorf("empty id")
	}
	if total < 0 {
		return fmt.Errorf("negative total")
	}
	fmt.Println(id, total)
	return nil
}
`

	tmpFile, _ := os.CreateTemp(t.TempDir(), "toktest*.go")
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(src)
	tmpFile.Close()

	node, err := ParseWithConfig(tmpFile.Name(), DefaultParseConfig())
	if err != nil {
		t.Fatal(err)
	}

	stream := syntax.Serialize(node)
	stmtCount := 0

	for _, n := range stream {
		if n.Statement {
			stmtCount++
		}
	}

	fmt.Printf("Total tokens: %d, Statement tokens: %d\n", len(stream), stmtCount)
}
