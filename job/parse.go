package job

import (
	"log"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func Parse(fchan chan string) (chan []*syntax.Node, chan int) {
	// parse AST
	achan := make(chan *syntax.Node)
	countChan := make(chan int, 1)
	go func() {
		fileCount := 0
		for file := range fchan {
			fileCount++
			ast, err := golang.Parse(file)
			if err != nil {
				log.Printf("%v", errors.NewParseError(file, 0, "failed to parse file", err))
				continue
			}
			achan <- ast
		}
		countChan <- fileCount
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
	return schan, countChan
}
