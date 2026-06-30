package job

import (
	"context"

	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func BuildTree(
	ctx context.Context,
	schan chan []*syntax.Node,
) (*suffixtree.STree, *[]*syntax.Node, chan bool) {
	t := suffixtree.New()
	data := make([]*syntax.Node, 0, 100)
	done := make(chan bool, 1)

	go func() {
		for seq := range schan {
			select {
			case <-ctx.Done():
				done <- true

				return
			default:
			}

			data = append(data, seq...)
			for _, node := range seq {
				t.Update(node)
			}
		}

		done <- true
	}()

	return t, &data, done
}
