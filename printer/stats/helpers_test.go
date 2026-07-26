package stats

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

const testFileContent = `package main

import "fmt"

func foo() {
	fmt.Println("hello")
}

func bar() {
	fmt.Println("hello")
}
`

func mockReadFile(content string) printer.ReadFile {
	return func(filename string) ([]byte, error) {
		return []byte(content), nil
	}
}

func processTestNodes(
	fread printer.ReadFile,
	hash string,
	dups [][]*syntax.Node,
) domain.ProcessedCloneGroup {
	group, err := printer.NodesToGroup(fread, hash, dups)
	if err != nil {
		return domain.ProcessedCloneGroup{Hash: hash, Clones: []domain.ProcessedClone{}}
	}

	return group
}

func printTestClones(p printer.Printer, fread printer.ReadFile, dups [][]*syntax.Node) error {
	return p.PrintClones(processTestNodes(fread, "test", dups))
}
