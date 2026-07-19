package hash

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// --- FindFileDuplicates ---

func TestFindFileDuplicates_IdenticalFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := []byte("package main\n\nfunc main() { println(\"hello\") }\n")

	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")
	writeDuplicateFiles(t, f1, f2, content)

	dups := FindFileDuplicates(context.Background(), []string{f1, f2}, 10)
	if len(dups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(dups))
	}

	if len(dups[0].Files) != 2 {
		t.Errorf("expected 2 files in group, got %d", len(dups[0].Files))
	}
}

func TestFindFileDuplicates_DifferentFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")

	writeTestFile(t, f1, []byte("package a\n"))
	writeTestFile(t, f2, []byte("package b\n"))

	dups := FindFileDuplicates(context.Background(), []string{f1, f2}, 1)
	testutil.AssertCountf(t, len(dups), 0, "expected 0 duplicates for different files, got %d")
}

func TestFindFileDuplicates_SkipsSmallFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("tiny")
	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")
	writeDuplicateFiles(t, f1, f2, content)

	dups := FindFileDuplicates(context.Background(), []string{f1, f2}, 100)
	testutil.AssertCountf(t, len(dups), 0, "expected 0 duplicates (files below threshold), got %d")
}

func TestFindFileDuplicates_SkipsNonexistentFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("package main\nfunc main() {}\n")

	good := filepath.Join(dir, "exists.go")
	writeTestFile(t, good, content)

	bad := filepath.Join(dir, "nope.go")

	dups := FindFileDuplicates(context.Background(), []string{good, bad}, 1)
	testutil.AssertCountf(t, len(dups), 0, "expected 0 duplicates with missing file, got %d")
}

func TestFindFileDuplicates_EmptyInput(t *testing.T) {
	t.Parallel()

	dups := FindFileDuplicates(context.Background(), nil, 1)
	testutil.AssertCountf(t, len(dups), 0, "expected 0 duplicates for nil input, got %d")
}

func TestFindFileDuplicates_MultipleGroups(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	contentA := []byte("package main\n\nfunc a() { println(\"aaaaaaaaaaaaaaaa\") }\n")
	contentB := []byte("package main\n\nfunc b() { println(\"bbbbbbbbbbbbbbbb\") }\n")

	files := map[string]string{
		"a1.go": filepath.Join(dir, "a1.go"),
		"a2.go": filepath.Join(dir, "a2.go"),
		"b1.go": filepath.Join(dir, "b1.go"),
		"b2.go": filepath.Join(dir, "b2.go"),
	}

	writeDuplicateFiles(t, files["a1.go"], files["a2.go"], contentA)
	writeDuplicateFiles(t, files["b1.go"], files["b2.go"], contentB)

	all := []string{files["a1.go"], files["a2.go"], files["b1.go"], files["b2.go"]}
	dups := FindFileDuplicates(context.Background(), all, 1)

	testutil.AssertCountf(t, len(dups), 2, "expected 2 duplicate groups, got %d")
}

// --- FileDetector.FindDuplOver via node pipeline ---

func TestFileDetector_FindDuplOver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		fileNames [2]string
		threshold int
		wantCount int
	}{
		{
			name:      "BasicDuplicate",
			content:   "package main\n\nfunc main() { println(\"hello, world!\") }\n",
			fileNames: [2]string{"dup1.go", "dup2.go"},
			threshold: 1,
			wantCount: 1,
		},
		{
			name:      "SkipsBelowThreshold",
			content:   "package main\n",
			fileNames: [2]string{"small1.go", "small2.go"},
			threshold: 1000,
			wantCount: 0,
		},
		{
			name:      "Delegation",
			content:   "package main\n\nfunc main() { println(\"delegation test!\") }\n",
			fileNames: [2]string{"d1.go", "d2.go"},
			threshold: 1,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			f1 := filepath.Join(dir, tt.fileNames[0])
			f2 := filepath.Join(dir, tt.fileNames[1])
			writeDuplicateFiles(t, f1, f2, []byte(tt.content))

			nodes := newSyntheticNodePair(f1, f2, len(tt.content))

			fd := NewFileDetector()
			matches := collectMatches(fd.FindDuplOver(context.Background(), nodes, tt.threshold))

			testutil.AssertCountf(t, len(matches), tt.wantCount, "expected %d match, got %%d", tt.wantCount)
		})
	}
}

func TestFileDetector_FindDuplOver_EmptyNodes(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector()
	matches := collectMatches(fd.FindDuplOver(context.Background(), nil, 5))

	testutil.AssertCountf(t, len(matches), 0, "expected 0 matches for empty input, got %d")
}

func TestFileDetector_FindDuplOver_NodesWithoutFilename(t *testing.T) {
	t.Parallel()

	nodes := []*syntax.Node{
		{Filename: "", Type: 1, Pos: 0, End: 100},
		{Filename: "", Type: 1, Pos: 0, End: 100},
	}

	fd := NewFileDetector()
	matches := collectMatches(fd.FindDuplOver(context.Background(), nodes, 1))

	testutil.AssertCountf(t, len(matches), 0, "expected 0 matches for empty filenames, got %d")
}

func TestFileDetector_FindDuplOver_NonexistentFiles(t *testing.T) {
	t.Parallel()

	nodes := []*syntax.Node{
		syntax.NewSyntheticFileNode("/nonexistent/file1.go", 100),
		syntax.NewSyntheticFileNode("/nonexistent/file2.go", 100),
	}

	fd := NewFileDetector()
	matches := collectMatches(fd.FindDuplOver(context.Background(), nodes, 1))

	testutil.AssertCountf(t, len(matches), 0, "expected 0 matches for nonexistent files, got %d")
}

func TestFileDetector_FindDuplOver_DeduplicatesSameFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("package main\n\nfunc main() { println(\"dedup test content\") }\n")

	f1 := filepath.Join(dir, "same.go")
	writeTestFile(t, f1, content)

	nodes := []*syntax.Node{
		syntax.NewSyntheticFileNode(f1, len(content)),
		syntax.NewSyntheticFileNode(f1, len(content)),
		syntax.NewSyntheticFileNode(f1, len(content)),
	}

	fd := NewFileDetector()
	matches := collectMatches(fd.FindDuplOver(context.Background(), nodes, 1))

	testutil.AssertCountf(
		t,
		len(matches),
		0,
		"expected 0 matches for same file referenced 3x (only 1 unique file), got %d",
	)
}

// --- extractUniqueFiles ---

func TestExtractUniqueFiles_Deduplicates(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector()

	nodes := []*syntax.Node{
		{Filename: "a.go"},
		{Filename: "b.go"},
		{Filename: "a.go"},
		{Filename: "c.go"},
		{Filename: "b.go"},
	}

	uniqueFiles := fd.extractUniqueFiles(nodes)
	if len(uniqueFiles) != 3 {
		t.Errorf("expected 3 unique uniqueFiles, got %d: %v", len(uniqueFiles), uniqueFiles)
	}
}

func TestExtractUniqueFiles_SkipsEmptyFilenames(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector()

	nodes := []*syntax.Node{
		{Filename: "a.go"},
		{Filename: ""},
		{Filename: "b.go"},
	}

	files := fd.extractUniqueFiles(nodes)
	if len(files) != 2 {
		t.Errorf("expected 2 files (empty skipped), got %d: %v", len(files), files)
	}
}

func TestExtractUniqueFiles_Empty(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector()

	files := fd.extractUniqueFiles(nil)
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

// --- hashFile error paths ---

func TestHashFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector()

	hashEntry, ok := fd.hashFile("/nonexistent/path/file.go")
	if ok {
		t.Error("expected ok=false for nonexistent file")
	}

	if hashEntry.Hash != "" || hashEntry.Filename != "" || hashEntry.Size != 0 {
		t.Errorf("expected zero FileHash, got %+v", hashEntry)
	}
}

func TestHashFile_UnreadableFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	f := filepath.Join(dir, "unreadable.go")

	if err := os.WriteFile(f, []byte("content"), 0o000); err != nil {
		t.Fatal(err)
	}

	fd := NewFileDetector()

	fileHash, ok := fd.hashFile(f)

	if ok {
		t.Error("expected ok=false for unreadable file")
	}

	testutil.AssertFieldValue(t, fileHash.Hash, "", "Hash")
}

func TestHashFile_ValidFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	f := filepath.Join(dir, "valid.go")
	content := []byte("package main\n\nfunc main() {}\n")

	writeTestFile(t, f, content)

	fd := NewFileDetector()

	fileHash, ok := fd.hashFile(f)
	if !ok {
		t.Fatal("expected ok=true for valid file")
	}

	if fileHash.Filename != f {
		t.Errorf("expected filename %q, got %q", f, fileHash.Filename)
	}

	if fileHash.Size != len(content) {
		t.Errorf("expected size %d, got %d", len(content), fileHash.Size)
	}

	if fileHash.Hash == "" {
		t.Error("expected non-empty hash")
	}
}

// --- fileError ---

func TestFileError_ReturnsZeroState(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector()

	ok := fd.fileError("test.go", os.ErrNotExist, "file not found")
	if ok {
		t.Error("expected ok=false")
	}
}

// --- FileHash / FileDuplicate struct coverage ---

func TestFileHash_Fields(t *testing.T) {
	t.Parallel()

	fileHash := FileHash{Hash: "abc123", Filename: "test.go", Size: 42}
	if fileHash.Hash != "abc123" || fileHash.Filename != "test.go" || fileHash.Size != 42 {
		t.Errorf("FileHash fields not set correctly: %+v", fileHash)
	}
}

func TestFileDuplicate_Fields(t *testing.T) {
	t.Parallel()

	fd := FileDuplicate{
		Hash: "deadbeef",
		Files: []FileHash{
			{Hash: "deadbeef", Filename: "a.go", Size: 100},
			{Hash: "deadbeef", Filename: "b.go", Size: 100},
		},
	}

	testutil.AssertFieldValue(t, fd.Hash, "deadbeef", "Hash")

	if len(fd.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(fd.Files))
	}
}

// --- FileDetector ---

func TestNewFileDetector(t *testing.T) {
	t.Parallel()

	hd := NewFileDetector()
	if hd == nil {
		t.Fatal("expected non-nil FileDetector")
	}
}

// --- helper ---

func collectMatches(ch <-chan syntax.Match) []syntax.Match {
	var matches []syntax.Match

	for m := range ch {
		matches = append(matches, m)
	}

	return matches
}

func writeTestFile(t *testing.T, path string, fileContent []byte) {
	t.Helper()

	if err := os.WriteFile(path, fileContent, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeDuplicateFiles(t *testing.T, f1, f2 string, content []byte) {
	t.Helper()
	writeTestFile(t, f1, content)
	writeTestFile(t, f2, content)
}

func newSyntheticNodePair(f1, f2 string, contentLen int) []*syntax.Node {
	return []*syntax.Node{
		syntax.NewSyntheticFileNode(f1, contentLen),
		syntax.NewSyntheticFileNode(f2, contentLen),
	}
}
