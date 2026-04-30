package printer

import (
	"errors"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestProcessFileContent(t *testing.T) {
	t.Parallel()

	node := &syntax.Node{Filename: "test.go", Pos: 0, End: 5}
	fread := func(filename string) ([]byte, error) {
		return []byte("hello"), nil
	}

	info, err := ProcessFileContent(fread, node)
	if err != nil {
		t.Fatalf("ProcessFileContent() error: %v", err)
	}

	if info.Filename != "test.go" {
		t.Errorf("Filename = %q, want %q", info.Filename, "test.go")
	}

	if string(info.Content) != "hello" {
		t.Errorf("Content = %q, want %q", string(info.Content), "hello")
	}
}

func TestProcessFileContent_NilNode(t *testing.T) {
	t.Parallel()

	fread := func(filename string) ([]byte, error) { return nil, nil }

	_, err := ProcessFileContent(fread, nil)
	if err == nil {
		t.Error("ProcessFileContent(nil node) should return error")
	}
}

func TestProcessFileContent_ReadError(t *testing.T) {
	t.Parallel()

	node := &syntax.Node{Filename: "missing.go", Pos: 0, End: 5}
	fread := func(filename string) ([]byte, error) {
		return nil, errors.New("file not found")
	}

	_, err := ProcessFileContent(fread, node)
	if err == nil {
		t.Error("ProcessFileContent() should return error on read failure")
	}
}

func TestProcessNodeRange(t *testing.T) {
	t.Parallel()

	startNode := &syntax.Node{Filename: "test.go", Pos: 0, End: 5}
	endNode := &syntax.Node{Filename: "test.go", Pos: 5, End: 10}
	fread := func(filename string) ([]byte, error) {
		return []byte("hello world"), nil
	}

	info, err := ProcessNodeRange(fread, startNode, endNode)
	if err != nil {
		t.Fatalf("ProcessNodeRange() error: %v", err)
	}

	if info.Filename != "test.go" {
		t.Errorf("Filename = %q, want %q", info.Filename, "test.go")
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

func TestSumFragmentLengths(t *testing.T) {
	t.Parallel()

	clones := []clone{
		{fragment: []byte("hello")},
		{fragment: []byte("world!")},
	}

	if got := sumFragmentLengths(clones); got != 11 {
		t.Errorf("sumFragmentLengths() = %d, want 11", got)
	}

	if got := sumFragmentLengths(nil); got != 0 {
		t.Errorf("sumFragmentLengths(nil) = %d, want 0", got)
	}
}

func TestSortCloneGroupsBySize(t *testing.T) {
	t.Parallel()

	groups := [][]clone{
		{{fragment: []byte("short")}},
		{{fragment: []byte("this is a longer fragment")}},
		{{fragment: []byte("medium")}},
	}

	sortCloneGroupsBySize(groups)

	frag0 := string(groups[0][0].fragment)
	if frag0 != "this is a longer fragment" {
		t.Errorf("sortCloneGroupsBySize: first = %q, want longest", frag0)
	}
}

func TestExtractSortCriteria(t *testing.T) {
	t.Parallel()

	if got := ExtractSortCriteria(); got != SortBySize {
		t.Errorf("ExtractSortCriteria() = %v, want SortBySize", got)
	}

	if got := ExtractSortCriteria(SortByHash); got != SortByHash {
		t.Errorf("ExtractSortCriteria(SortByHash) = %v, want SortByHash", got)
	}
}
