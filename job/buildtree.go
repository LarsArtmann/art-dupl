package job

import (
	"context"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func BuildTree(
	ctx context.Context,
	schan chan []*syntax.Node,
) (*suffixtree.STree, *[]*syntax.Node, chan error) {
	t := suffixtree.New()
	data := make([]*syntax.Node, 0, 100)
	done := make(chan error, 1)

	go func() {
		for seq := range schan {
			select {
			case <-ctx.Done():
				done <- nil

				return
			default:
			}

			data = append(data, seq...)
			for _, node := range seq {
				if err := t.Update(node); err != nil {
					logger.Default.Error("suffix tree update failed", "err", err)
					done <- err

					return
				}
			}
		}

		done <- nil
	}()

	return t, &data, done
}
