package job

import (
	"github.com/LarsArtmann/art-dupl/pkg/logger"
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
				logger.Default.Error("failed to parse file", "file", file, "err", err)
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
