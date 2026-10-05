package printer

import (
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/jsonutil"
	"github.com/LarsArtmann/art-dupl/printer/finding"
	gofinding "github.com/larsartmann/go-finding"
)

// LSPDocument is the standalone equivalent of an LSP publishDiagnostics
// notification payload: one file's diagnostics under its URI. A bare
// LSPDiagnostic has no self-location (the protocol delivers diagnostics per
// document), so grouping by file is what makes the CLI output consumable —
// each entry feeds go-finding's FromLSP(fileURI, diag) directly.
type LSPDocument struct {
	URI         string                    `json:"uri"`
	Diagnostics []gofinding.LSPDiagnostic `json:"diagnostics"`
}

// lspPrinter emits clone groups as a JSON array of LSPDocument entries, one
// diagnostic per clone occurrence, converted through the go-finding adapter
// so the payload is exactly what editor/LSP consumers would get from the
// findings themselves: severity ladder, code/source, related information
// pointing at sibling occurrences, and the go-finding round-trip data
// (finding ID, GroupID, tags, metadata) in the diagnostic's data property.
type lspPrinter struct {
	w          io.Writer
	threshold  int
	byFile     map[string][]gofinding.LSPDiagnostic
	seenHashes map[string]bool
}

// NewLSP creates an LSP diagnostics printer at the given severity threshold
// (mirrors NewSARIF's ladder semantics; values <= 0 fall back to
// config.DefaultThreshold inside the adapter).
func NewLSP(w io.Writer, _ ReadFile, threshold int) Printer {
	return &lspPrinter{
		w:          w,
		threshold:  threshold,
		byFile:     make(map[string][]gofinding.LSPDiagnostic),
		seenHashes: make(map[string]bool),
	}
}

func (p *lspPrinter) PrintHeader() error { return nil }

func (p *lspPrinter) PrintClones(
	group domain.ProcessedCloneGroup,
	_ ...config.SortCriteria,
) error {
	if len(group.Clones) == 0 || group.Hash == "" {
		return nil
	}

	if p.seenHashes[group.Hash] {
		return nil
	}

	p.seenHashes[group.Hash] = true

	for _, f := range finding.ToFindings(group, finding.Options{Threshold: p.threshold}) {
		file := string(f.Position.File)
		p.byFile[file] = append(p.byFile[file], f.ToLSP())
	}

	return nil
}

func (p *lspPrinter) PrintFooter() error {
	files := make([]string, 0, len(p.byFile))
	for file := range p.byFile {
		files = append(files, file)
	}

	sort.Strings(files)

	documents := make([]LSPDocument, 0, len(files))
	for _, file := range files {
		documents = append(documents, LSPDocument{URI: file, Diagnostics: p.byFile[file]})
	}

	data, err := jsonutil.MarshalIndent(documents, "", "  ")
	if err != nil {
		return errors.HandleMarshalingError("encode", "LSP diagnostics output", err)
	}

	return writeFormattedOutput(p.w, data, "LSP diagnostics output")
}
