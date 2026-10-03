package artdupl

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isCombinedMode reports whether both TypeAware and SuggestGenerics are set.
// Instead of one option silently discarding the other, the combined mode runs
// two detection passes over one shared go/packages type-check: a type-aware
// pass and a type-erased pass. See ADR-0026.
func (d *detector) isCombinedMode() bool {
	return d.cfg.TypeAware && d.cfg.SuggestGenerics
}

// loadCombinedTypeData loads type information once for the combined mode and
// returns both hash dispositions. On load failure it falls back to syntax-only
// (nil, nil), in which case both passes produce identical results and the
// generics-pass filter drops everything, leaving the type-aware output.
func (d *detector) loadCombinedTypeData(files []string) (golang.TypeAwareData, golang.TypeAwareData) {
	goFiles := make([]string, 0, len(files))

	for _, f := range files {
		if filepath.Ext(f) == ".go" {
			goFiles = append(goFiles, f)
		}
	}

	if len(goFiles) == 0 {
		return nil, nil
	}

	typeData, err := golang.LoadTypeAwareData(goFiles, false)
	if err != nil {
		d.logger.Warn("Combined mode type loading failed, falling back to syntax-only", "err", err)

		return nil, nil
	}

	return typeData, typeData.WithEraseHash(true)
}

// buildCombinedPipelines builds both detection passes from one shared
// go/packages type-check. Only the parse + suffix-tree work runs twice.
func (d *detector) buildCombinedPipelines(
	ctx context.Context,
	files []string,
) (*pipelineResult, *pipelineResult, error) {
	typeAwareData, erasedData := d.loadCombinedTypeData(files)

	typeAwareResult, err := d.buildPipelineWithTypeData(ctx, files, typeAwareData)
	if err != nil {
		return nil, nil, fmt.Errorf("type-aware pass failed: %w", err)
	}

	genericsResult, err := d.buildPipelineWithTypeData(ctx, files, erasedData)
	if err != nil {
		return nil, nil, fmt.Errorf("generics pass failed: %w", err)
	}

	return typeAwareResult, genericsResult, nil
}

// runCombinedDetection runs both passes and merges their group maps:
// type-aware groups unchanged, plus generics-pass families with at least
// syntax.MinDivergentPositions divergent type positions. Zero-divergence
// generics families are exact content duplicates of type-aware groups;
// one-position divergence is the shallow receiver noise class type-aware
// matching exists to eliminate.
func (d *detector) runCombinedDetection(
	ctx context.Context,
	typeAwareResult, genericsResult *pipelineResult,
) (map[string][][]*syntax.Node, error) {
	typeAwareGroups, err := d.detectGroups(ctx, typeAwareResult)
	if err != nil {
		return nil, fmt.Errorf("type-aware detection failed: %w", err)
	}

	genericsGroups, err := d.detectGroups(ctx, genericsResult)
	if err != nil {
		return nil, fmt.Errorf("generics detection failed: %w", err)
	}

	for hash, frags := range genericsGroups {
		uniq := syntax.Unique(frags)

		if len(uniq) <= 1 || !syntax.IsGenericsCandidateStructure(uniq) {
			continue
		}

		typeAwareGroups[hash] = frags
	}

	return typeAwareGroups, nil
}

// runCombinedAnalysis is the combined-mode entry for FindClones: both passes,
// merged groups converted to CloneGroups.
func (d *detector) runCombinedAnalysis(
	ctx context.Context,
	files []string,
) ([]*CloneGroup, *pipelineResult, error) {
	d.reportProgress(0, "Starting combined type-aware + generics analysis", "")

	typeAwareResult, genericsResult, err := d.buildCombinedPipelines(ctx, files)
	if err != nil {
		return nil, nil, err
	}

	d.reportProgress(70, "Starting duplicate detection", "")

	groups, err := d.runCombinedDetection(ctx, typeAwareResult, genericsResult)
	if err != nil {
		return nil, nil, err
	}

	var allGroups []*CloneGroup

	d.processCloneGroups(groups, func(group *CloneGroup) {
		allGroups = append(allGroups, group)
	})

	d.reportProgress(90, "Processing results", "")

	return allGroups, typeAwareResult, nil
}

// streamCombinedInto runs the combined two-pass analysis and forwards its
// groups as StreamResults, emitting a terminal error result on failure.
func (d *detector) streamCombinedInto(
	ctx context.Context,
	files []string,
	resultChan chan<- StreamResult,
) {
	groupChan := make(chan *CloneGroup, 10)

	go func() {
		defer close(groupChan)

		streamErr := d.streamCombinedResults(ctx, files, groupChan)
		if streamErr != nil {
			select {
			case resultChan <- StreamResult{Group: nil, Err: streamErr}:
			case <-ctx.Done():
			}
		}
	}()

	for group := range groupChan {
		if group == nil {
			continue
		}

		select {
		case resultChan <- StreamResult{Group: group, Err: nil}:
		case <-ctx.Done():
			return
		}
	}

	d.reportProgress(100, "Analysis complete", "")
}

// streamCombinedResults is the combined-mode entry for
// FindClonesStreamResult: both passes, merged groups streamed to the channel.
func (d *detector) streamCombinedResults(
	ctx context.Context,
	files []string,
	resultChan chan<- *CloneGroup,
) error {
	typeAwareResult, genericsResult, err := d.buildCombinedPipelines(ctx, files)
	if err != nil {
		return err
	}

	groups, err := d.runCombinedDetection(ctx, typeAwareResult, genericsResult)
	if err != nil {
		return err
	}

	d.streamCloneGroups(ctx, groups, resultChan)

	return nil
}
