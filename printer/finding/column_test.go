package finding

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/domain"
	artdupl "github.com/LarsArtmann/art-dupl/pkg/artdupl"
)

// TestColumnsDerivedFromByteOffsets pins the column data flow end to end:
// the SDK detects clones over a fixture whose indentation is known, the
// clone positions must carry 1-based byte columns (start at the first token
// after the tab, end one past the last content column), and the finding
// adapter must copy them onto go-finding Position.Column (0 = unknown).
func TestColumnsDerivedFromByteOffsets(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	source := "package main\n\nfunc a() {\n\tprintln(1)\n\tprintln(2)\n}\n\nfunc b() {\n\tprintln(3)\n\tprintln(4)\n}\n"
	aPath := filepath.Join(dir, "a.go")
	bPath := filepath.Join(dir, "b.go")

	if err := os.WriteFile(aPath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(bPath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	detector, err := artdupl.NewDetector(&artdupl.Options{
		Threshold:        2,
		DetectionMethods: []artdupl.DetectionMethod{artdupl.MethodArtDupl},
		MaxWorkers:       1,
		Timeout:          30 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := detector.FindClones(context.Background(), []string{aPath, bPath})
	if err != nil {
		t.Fatal(err)
	}

	checked := 0

	for _, g := range result.CloneGroups {
		for _, cl := range g.Clones {
			// Both statements start after "\t" (tab = column 1), so the
			// clone starts at column 2 of its first line. The end is the
			// exclusive offset one past the last content byte of the last
			// statement: ')'+1 -> the column after ')'.
			if cl.ColumnStart != 2 {
				t.Errorf("%s lines %d-%d: ColumnStart = %d, want 2",
					filepath.Base(cl.Filename), cl.LineStart, cl.LineEnd, cl.ColumnStart)
			}

			if cl.ColumnEnd <= cl.ColumnStart {
				t.Errorf("%s: ColumnEnd = %d not past ColumnStart %d",
					filepath.Base(cl.Filename), cl.ColumnEnd, cl.ColumnStart)
			}

			checked++
		}
	}

	if checked == 0 {
		t.Fatal("no clones found in fixture")
	}

	// The adapter copies columns onto go-finding positions.
	group := domain.ProcessedCloneGroup{
		Hash: "col0000000000001",
		Clones: []domain.ProcessedClone{{
			CloneRef:    domain.CloneRef{Filename: "a.go", LineStart: 4, LineEnd: 5},
			StartPos:    26,
			EndPos:      48,
			ColumnStart: 2,
			ColumnEnd:   12,
			TokenCount:  2,
		}},
	}

	findings := ToFindings(group, Options{Threshold: 5})
	if len(findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(findings))
	}

	f := findings[0]
	if f.Position.Column != 2 {
		t.Errorf("Position.Column = %d, want 2", f.Position.Column)
	}

	if f.Range == nil || f.Range.End.Column != 12 {
		t.Errorf("Range.End.Column = %v, want 12", f.Range)
	}

	// Zero columns mean unknown and must not fabricate a position column.
	unknown := group
	unknown.Clones[0].ColumnStart = 0
	unknown.Clones[0].ColumnEnd = 0

	if f0 := ToFindings(unknown, Options{Threshold: 5}); f0[0].Position.Column != 0 {
		t.Errorf("Position.Column = %d, want 0 (unknown)", f0[0].Position.Column)
	}
}
