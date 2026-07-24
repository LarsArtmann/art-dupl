package cmd

import (
	"context"
	"fmt"
	"io"
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
//	<filename>\t<position>\t<type>\t<name>\t<statement>
//
// Sentinels (Type=-1) are printed as "---" to visually separate files.
func dumpTokensOutput(ctx context.Context, cfg *config.Config, w io.Writer) error {
	err := validatePaths(cfg.Paths, cfg.FilesFromStdin)
	if err != nil {
		return err
	}

	filterParam, err := setupFilter(cfg)
	if err != nil {
		return err
	}

	var filterStats *FilterStats
	if filterParam != nil {
		filterStats = NewFilterStats(filterParam.FilterReasons())
	}

	params := buildParams{
		ctx:          ctx,
		paths:        cfg.Paths,
		cfg:          cfg,
		filterParam:  filterParam,
		filterStats:  filterStats,
		outputFormat: cfg.OutputFormat,
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

	for seq := range schan {
		for _, node := range seq {
			if node.Type == -1 {
				_, _ = fmt.Fprintln(w, "---")

				continue
			}

			baseType := node.Type & 0xFF
			semanticHash := (node.Type >> 8) & 0xFFFFFF

			if semanticHash > 0 {
				_, _ = fmt.Fprintf(w, "%s\t%d\tbase=%d+hash=%d\t%s\t%v\n",
					node.Filename, node.Pos, baseType, semanticHash, node.Name, node.Statement)
			} else {
				_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%v\n",
					node.Filename, node.Pos, strconv.Itoa(int(baseType)), node.Name, node.Statement)
			}
		}
	}

	<-statsChan

	return nil
}
