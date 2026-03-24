package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestTextOutputGolden(t *testing.T) {
	var buf bytes.Buffer

	statsPrinter := NewStats(&buf, mockReadFile(testFileContent), 15).(*stats)
	statsPrinter.SetFilesCount(5)
	statsPrinter.SetDetectionMethods("art-dupl")

	if err := statsPrinter.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

func TestStatsJSONOutputGolden(t *testing.T) {
	var buf bytes.Buffer

	statsPrinter := NewStats(&buf, mockReadFile(testFileContent), 15).(*stats)
	statsPrinter.SetFilesCount(10)
	statsPrinter.SetDetectionMethods("art-dupl,hash")

	// Add some clone data
	dups := [][]*syntax.Node{
		{
			{Filename: "file1.go", Pos: 1, End: 3, Type: 1},
			{Filename: "file1.go", Pos: 2, End: 4, Type: 2},
		},
		{
			{Filename: "file2.go", Pos: 10, End: 12, Type: 1},
			{Filename: "file2.go", Pos: 11, End: 13, Type: 2},
		},
	}

	if err := statsPrinter.PrintClones(dups); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	statsPrinter.format = FormatJSON

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

func TestStatsCSVOutputGolden(t *testing.T) {
	var buf bytes.Buffer

	statsPrinter := NewStats(&buf, mockReadFile(testFileContent), 15).(*stats)
	statsPrinter.SetFilesCount(3)
	statsPrinter.SetDetectionMethods("art-dupl")
	statsPrinter.format = FormatCSV

	if err := statsPrinter.PrintClones(createTestCloneGroups()); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}
