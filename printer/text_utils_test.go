package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// testNode creates a syntax.Node for testing.
func testNode(filename string, pos, end int) *syntax.Node {
	return &syntax.Node{Filename: filename, Pos: int32(pos), End: int32(end)}
}

func TestProcessFileContent(t *testing.T) {
	t.Parallel()

	node := testNode(testFilename, 0, 5)
	fread := func(name string) ([]byte, error) {
		return []byte("hello"), nil
	}

	meta, err := ProcessFileContent(fread, node)
	if err != nil {
		t.Fatalf("ProcessFileContent() error: %v", err)
	}

	if meta.Filename != testFilename {
		t.Errorf("Filename = %q, want %q", meta.Filename, testFilename)
	}

	if string(meta.Content) != "hello" {
		t.Errorf("Content = %q, want %q", string(meta.Content), "hello")
	}
}

func TestProcessFileContent_NilNode(t *testing.T) {
	t.Parallel()

	fread := func(fname string) ([]byte, error) { return nil, nil }

	_, err := ProcessFileContent(fread, nil)
	if err == nil {
		t.Error("ProcessFileContent(nil node) should return error")
	}
}

func TestProcessFileContent_ReadError(t *testing.T) {
	t.Parallel()

	node := testNode("missing.go", 0, 5)
	fread := errorReadFile("file not found")

	_, err := ProcessFileContent(fread, node)
	if err == nil {
		t.Error("ProcessFileContent() should return error on read failure")
	}
}

func TestProcessNodeRange(t *testing.T) {
	t.Parallel()

	startNode := testNode(testFilename, 0, 5)
	endNode := testNode(testFilename, 5, 10)
	fread := func(filename string) ([]byte, error) {
		return []byte("hello world"), nil
	}

	info, err := ProcessNodeRange(fread, startNode, endNode)
	if err != nil {
		t.Fatalf("ProcessNodeRange() error: %v", err)
	}

	if info.Filename != testFilename {
		t.Errorf("Filename = %q, want %q", info.Filename, testFilename)
	}
}

func TestProcessNodeRange_NilNodes(t *testing.T) {
	t.Parallel()

	fread := func(filename string) ([]byte, error) { return nil, nil }

	_, err := ProcessNodeRange(fread, nil, &syntax.Node{})
	if err == nil {
		t.Error("ProcessNodeRange(nil start) should return error")
	}

	_, err = ProcessNodeRange(fread, &syntax.Node{}, nil)
	if err == nil {
		t.Error("ProcessNodeRange(nil end) should return error")
	}
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    int
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range tests {
		got := formatBytes(tc.input)
		if got != tc.expected {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestExtractSortCriteria(t *testing.T) {
	t.Parallel()

	if got := ExtractSortCriteria(); got != config.SortBySize {
		t.Errorf("ExtractSortCriteria() = %v, want config.SortBySize", got)
	}

	if got := ExtractSortCriteria(config.SortByHash); got != config.SortByHash {
		t.Errorf("ExtractSortCriteria(config.SortByHash) = %v, want config.SortByHash", got)
	}
}
