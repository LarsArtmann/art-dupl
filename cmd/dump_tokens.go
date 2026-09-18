package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// dumpTokensOutput parses all files, serializes their ASTs to the flat token
// stream, prints each token to stdout, and exits without running clone detection.
// This is useful for debugging false positives/negatives: it shows exactly what
// the suffix tree sees after transformation and normalization.
//
// Output format (tab-separated, one token per line):
//
//	<filename>\t<offset>\t<line:col[-line:col]>\t<type>\t<name>\t<statement>
//
// <offset> is the byte offset of the token start (matching Node.Pos, useful
// for correlating with serialized stream ranges). <line:col> is the 1-based
// source position of the token start (columns count bytes, like go/token);
// the filename is column 0. Multi-line tokens (composite statement tokens)
// append "-<endline>:<endcol>" so the covered source span is visible without
// opening the file. Positions degrade to "?" when the file cannot be read
// (deleted between walk and dump).
//
// Sentinels (Type=-1) are printed as "---" to visually separate files.
func dumpTokensOutput(ctx context.Context, cfg *config.Config, w io.Writer, stderr io.Writer) error {
	err := validatePaths(cfg.Paths, cfg.FilesFromStdin)
	if err != nil {
		return err
	}

	filterParam, err := setupFilter(stderr, cfg)
	if err != nil {
		return err
	}

	var filterStats *FilterStats = newTrackedFilterStats(filterParam, cfg)

	params := buildParams{
		ctx:          ctx,
		paths:        cfg.Paths,
		cfg:          cfg,
		filterParam:  filterParam,
		filterStats:  filterStats,
		outputFormat: cfg.OutputFormat,
		stderr:       stderr,
	}

	filesChan := params.getFilesChan()

	var (
		schan     chan []*syntax.Node
		statsChan chan job.ParseStats
	)

	if cfg.Workers != 1 {
		schan, statsChan = job.ParseParallel(
			ctx, filesChan, cfg.Workers, detectionMode(cfg), cfg.MaxChildrenSerial, nil,
		)
	} else {
		schan, statsChan = job.Parse(ctx, filesChan, detectionMode(cfg), cfg.MaxChildrenSerial, nil)
	}

	lineTables := make(map[string]*sourceLineTable)

	for seq := range schan {
		for _, node := range seq {
			if node.Type == -1 {
				_, _ = fmt.Fprintln(w, "---")

				continue
			}

			baseType := node.Type & 0xFF
			semanticHash := (node.Type >> 8) & 0xFFFFFF

			typeLabel := strconv.Itoa(int(baseType))
			if semanticHash > 0 {
				typeLabel = fmt.Sprintf("base=%d+hash=%d", baseType, semanticHash)
			}

			_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%v\n",
				node.Filename, node.Pos, formatTokenPosition(lineTables, node), typeLabel, node.Name, node.Statement)
		}
	}

	<-statsChan

	return nil
}

// sourceLineTable converts byte offsets into 1-based line:column pairs for one
// source file. The dump path re-reads the file from disk because the serialized
// Node stream carries byte offsets only; a missing file degrades to "?" instead
// of failing the dump.
type sourceLineTable struct {
	starts []int32
}

func newSourceLineTable(path string) *sourceLineTable {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	starts := make([]int32, 0, 64)
	starts = append(starts, 0)
	for i, b := range content {
		if b == '\n' {
			starts = append(starts, int32(i+1)) // #nosec G115 -- file offsets bounded by int32, same as transformer
		}
	}

	return &sourceLineTable{starts: starts}
}

// lineCol returns the 1-based line and byte column for a byte offset, or
// (0, 0) when the table is unavailable or the offset is out of range.
func (t *sourceLineTable) lineCol(offset int32) (int32, int32) {
	if t == nil || offset < 0 {
		return 0, 0
	}

	i := sort.Search(len(t.starts), func(i int) bool { return t.starts[i] > offset })
	if i == 0 {
		return 0, 0
	}
	i--

	return int32(i + 1), offset - t.starts[i] + 1 // #nosec G115 -- index and offset bounded by int32 file size
}

// formatTokenPosition renders a node's source position ("line:col", with a
// "-endline:endcol" span suffix for multi-line tokens) using a per-file line
// table cache, or "?" when the file cannot be read.
func formatTokenPosition(tables map[string]*sourceLineTable, node *syntax.Node) string {
	tbl, ok := tables[node.Filename]
	if !ok {
		tbl = newSourceLineTable(node.Filename)
		tables[node.Filename] = tbl
	}

	line, col := tbl.lineCol(node.Pos)
	if line == 0 {
		return "?"
	}

	position := fmt.Sprintf("%d:%d", line, col)

	endLine, endCol := tbl.lineCol(node.End)
	if endLine > line {
		position = fmt.Sprintf("%s-%d:%d", position, endLine, endCol)
	}

	return position
}
