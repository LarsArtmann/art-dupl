package printer

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
)

func (p *TextPrinter) PrintFindings(findings []domain.Finding) error {
	if len(findings) == 0 {
		return nil
	}

	if _, err := fmt.Fprintf(p.w, "\n\U0001f4cb Findings (%d):\n", len(findings)); err != nil {
		return fmt.Errorf("write findings header: %w", err)
	}

	for _, f := range findings {
		if _, err := fmt.Fprintf(p.w, "  %s:%d [%s] %s\n", f.Filename, f.Line, f.Type, f.Message); err != nil {
			return fmt.Errorf("write finding %s:%d: %w", f.Filename, f.Line, err)
		}
	}

	return nil
}

func (p *plumbing) PrintFindings(findings []domain.Finding) error {
	for _, f := range findings {
		if _, err := fmt.Fprintf(p.w, "%s:%d\t%s\t%s\n", f.Filename, f.Line, f.Type, f.Message); err != nil {
			return fmt.Errorf("write finding %s:%d: %w", f.Filename, f.Line, err)
		}
	}

	return nil
}

func (p *JSONPrinter) PrintFindings(findings []domain.Finding) error {
	p.findings = findings

	return nil
}

func (p *htmlprinter) PrintFindings(findings []domain.Finding) error {
	if len(findings) == 0 {
		return nil
	}

	if _, err := fmt.Fprintf(p.w, "<h2>Findings (%d)</h2>\n<ul>\n", len(findings)); err != nil {
		return fmt.Errorf("write findings header: %w", err)
	}

	for _, f := range findings {
		if _, err := fmt.Fprintf(
			p.w,
			"<li><code>%s:%d</code> [%s] %s</li>\n",
			f.Filename, f.Line, f.Type, f.Message,
		); err != nil {
			return fmt.Errorf("write finding %s:%d: %w", f.Filename, f.Line, err)
		}
	}

	if _, err := fmt.Fprintln(p.w, "</ul>"); err != nil {
		return fmt.Errorf("write findings footer: %w", err)
	}

	return nil
}

func (p *sarifPrinter) PrintFindings(findings []domain.Finding) error {
	p.findings = findings

	return nil
}

func (p *stats) PrintFindings(_ []domain.Finding) error {
	return nil
}
