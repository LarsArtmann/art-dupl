package printer

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// ProcessClones converts raw syntax.Node groups into ProcessedClone slices.
// This is the single point where [][]*syntax.Node is decoded into domain types,
// eliminating the need for each printer to understand AST internals.
func ProcessClones(fread ReadFile, dups [][]*syntax.Node) ([]domain.ProcessedClone, error) {
	clones := make([]domain.ProcessedClone, len(dups))

	for i, dup := range dups {
		cnt := len(dup)
		if cnt == 0 {
			return nil, fmt.Errorf("zero length duplicate found at index %d", i)
		}

		nstart := dup[0]
		nend := dup[cnt-1]

		fileInfo, err := ProcessNodeRange(fread, nstart, nend)
		if err != nil {
			return nil, fmt.Errorf("failed to process node range for file %s: %w", nstart.Filename, err)
		}

		fragment := extractContent(fileInfo, nstart, nend)
		tokens := cnt
		lines := fileInfo.LineEnd - fileInfo.LineStart + 1

		clones[i] = domain.ProcessedClone{
			Filename:  fileInfo.Filename,
			LineStart: fileInfo.LineStart,
			LineEnd:   fileInfo.LineEnd,
			Fragment:  fragment,
			Size:      tokens,
			FileSize:  len(fileInfo.Content),
			Classification: ClassifyClone(
				fileInfo.Filename,
				nstart.Type,
				tokens,
				lines,
			),
		}
	}

	return clones, nil
}

// NodesToGroup converts raw syntax.Node groups into a ProcessedCloneGroup.
func NodesToGroup(fread ReadFile, hash string, dups [][]*syntax.Node) (domain.ProcessedCloneGroup, error) {
	clones, err := ProcessClones(fread, dups)
	if err != nil {
		return domain.ProcessedCloneGroup{}, err
	}

	size := 0
	for _, c := range clones {
		size += c.Size
	}

	return domain.ProcessedCloneGroup{
		Hash:   hash,
		Size:   size,
		Clones: clones,
	}, nil
}
