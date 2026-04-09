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

	err := statsPrinter.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	err = statsPrinter.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

func makeTestNodes(filename string, positions ...struct{ pos, end, typ int }) []*syntax.Node {
	nodes := make([]*syntax.Node, len(positions))
	for i, p := range positions {
		nodes[i] = &syntax.Node{Filename: filename, Pos: int32(p.pos), End: int32(p.end), Type: int32(p.typ)}
	}
	return nodes
}

func makeTestNodePair(filename string, p1, p2 struct{ pos, end, typ int }) []*syntax.Node {
	return makeTestNodes(filename, p1, p2)
}

func pos(p, e, t int) struct{ pos, end, typ int } { return struct{ pos, end, typ int }{p, e, t} }

func TestStatsJSONOutputGolden(t *testing.T) {
	var buf bytes.Buffer

	statsPrinter := NewStats(&buf, mockReadFile(testFileContent), 15).(*stats)
	statsPrinter.SetFilesCount(10)
	statsPrinter.SetDetectionMethods("art-dupl,hash")

	dups := [][]*syntax.Node{
		makeTestNodePair("file1.go", pos(1, 3, 1), pos(2, 4, 2)),
		makeTestNodePair("file2.go", pos(10, 12, 1), pos(11, 13, 2)),
	}

	err := statsPrinter.PrintClones(dups)
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	statsPrinter.format = FormatJSON

	err = statsPrinter.PrintFooter()
	if err != nil {
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

	err := statsPrinter.PrintClones(createTestCloneGroups())
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	err = statsPrinter.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}
