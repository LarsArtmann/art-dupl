package job

import (
	"log"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func Parse(fchan chan string) chan []*syntax.Node {
	// parse AST
	achan := make(chan *syntax.Node)
	go func() {
		for file := range fchan {
			ast, err := golang.Parse(file)
			if err != nil {
				log.Printf("%v", errors.NewParseError(file, 0, "failed to parse file", err))
				continue
			}
			achan <- ast
		}
		close(achan)
	}()

	// serialize
	schan := make(chan []*syntax.Node)
	go func() {
		for ast := range achan {
			seq := syntax.Serialize(ast)
			schan <- seq
		}
		close(schan)
	}()
	return schan
}
