package job

import (
	"context"
	"math"
	"sync/atomic"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// sentinelCounter generates unique sentinel token values to separate files
// in the suffix tree. Without sentinels, short files (especially .templ with
// statement-level tokenization) can fail to produce maximal repeats because
// overlapping suffixes from adjacent files collapse into the same context.
var sentinelCounter atomic.Int32

// BadNodeType is a node type value that never appears in real parsed nodes.
const BadNodeType int32 = -1

func init() {
	sentinelCounter.Store(math.MinInt32 / 2)
}

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

			sentinel := &syntax.Node{
				Type:        BadNodeType,
				Statement:   true,
				Fingerprint: sentinelCounter.Add(1),
			}
			data = append(data, sentinel)
			_ = t.Update(sentinel)
		}

		done <- nil
	}()

	return t, &data, done
}
