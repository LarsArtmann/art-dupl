package hash

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// --- FindFileDuplicates ---

func TestFindFileDuplicates_IdenticalFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := []byte("package main\n\nfunc main() { println(\"hello\") }\n")

	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")

	if err := os.WriteFile(f1, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(f2, content, 0o644); err != nil {
		t.Fatal(err)
	}

	dups := FindFileDuplicates([]string{f1, f2}, 10)
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

	if err := os.WriteFile(f1, []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(f2, []byte("package b\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dups := FindFileDuplicates([]string{f1, f2}, 1)
	if len(dups) != 0 {
		t.Errorf("expected 0 duplicates for different files, got %d", len(dups))
	}
}

func TestFindFileDuplicates_SkipsSmallFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("tiny")
	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")

	if err := os.WriteFile(f1, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(f2, content, 0o644); err != nil {
		t.Fatal(err)
	}

	dups := FindFileDuplicates([]string{f1, f2}, 100)
	if len(dups) != 0 {
		t.Errorf("expected 0 duplicates (files below threshold), got %d", len(dups))
	}
}

func TestFindFileDuplicates_SkipsNonexistentFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("package main\nfunc main() {}\n")

	good := filepath.Join(dir, "exists.go")
	if err := os.WriteFile(good, content, 0o644); err != nil {
		t.Fatal(err)
	}

	bad := filepath.Join(dir, "nope.go")

	dups := FindFileDuplicates([]string{good, bad}, 1)
	if len(dups) != 0 {
		t.Errorf("expected 0 duplicates with missing file, got %d", len(dups))
	}
}

func TestFindFileDuplicates_EmptyInput(t *testing.T) {
	t.Parallel()

	dups := FindFileDuplicates(nil, 1)
	if len(dups) != 0 {
		t.Errorf("expected 0 duplicates for nil input, got %d", len(dups))
	}
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

	if err := os.WriteFile(files["a1.go"], contentA, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(files["a2.go"], contentA, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(files["b1.go"], contentB, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(files["b2.go"], contentB, 0o644); err != nil {
		t.Fatal(err)
	}

	all := []string{files["a1.go"], files["a2.go"], files["b1.go"], files["b2.go"]}
	dups := FindFileDuplicates(all, 1)

	if len(dups) != 2 {
		t.Errorf("expected 2 duplicate groups, got %d", len(dups))
	}
}

// --- FileDetector.FindDuplOver via node pipeline ---

func TestFileDetector_FindDuplOver_BasicDuplicate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("package main\n\nfunc main() { println(\"hello, world!\") }\n")

	f1 := filepath.Join(dir, "dup1.go")
	f2 := filepath.Join(dir, "dup2.go")

	if err := os.WriteFile(f1, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(f2, content, 0o644); err != nil {
		t.Fatal(err)
	}

	nodes := []*syntax.Node{
		syntax.NewSyntheticFileNode(f1, len(content)),
		syntax.NewSyntheticFileNode(f2, len(content)),
	}

	fd := NewFileDetector(1)
	matches := collectMatches(fd.FindDuplOver(nodes, 1))

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	if len(matches[0].Frags) != 2 {
		t.Errorf("expected 2 fragments, got %d", len(matches[0].Frags))
	}
}

func TestFileDetector_FindDuplOver_SkipsBelowThreshold(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("package main\n")

	f1 := filepath.Join(dir, "small1.go")
	f2 := filepath.Join(dir, "small2.go")

	if err := os.WriteFile(f1, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(f2, content, 0o644); err != nil {
		t.Fatal(err)
	}

	nodes := []*syntax.Node{
		syntax.NewSyntheticFileNode(f1, len(content)),
		syntax.NewSyntheticFileNode(f2, len(content)),
	}

	fd := NewFileDetector(1000)
	matches := collectMatches(fd.FindDuplOver(nodes, 1000))

	if len(matches) != 0 {
		t.Errorf("expected 0 matches (below threshold), got %d", len(matches))
	}
}

func TestFileDetector_FindDuplOver_EmptyNodes(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector(5)
	matches := collectMatches(fd.FindDuplOver(nil, 5))

	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty input, got %d", len(matches))
	}
}

func TestFileDetector_FindDuplOver_NodesWithoutFilename(t *testing.T) {
	t.Parallel()

	nodes := []*syntax.Node{
		{Filename: "", Type: 1, Pos: 0, End: 100},
		{Filename: "", Type: 1, Pos: 0, End: 100},
	}

	fd := NewFileDetector(1)
	matches := collectMatches(fd.FindDuplOver(nodes, 1))

	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty filenames, got %d", len(matches))
	}
}

func TestFileDetector_FindDuplOver_NonexistentFiles(t *testing.T) {
	t.Parallel()

	nodes := []*syntax.Node{
		syntax.NewSyntheticFileNode("/nonexistent/file1.go", 100),
		syntax.NewSyntheticFileNode("/nonexistent/file2.go", 100),
	}

	fd := NewFileDetector(1)
	matches := collectMatches(fd.FindDuplOver(nodes, 1))

	if len(matches) != 0 {
		t.Errorf("expected 0 matches for nonexistent files, got %d", len(matches))
	}
}

func TestFileDetector_FindDuplOver_DeduplicatesSameFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("package main\n\nfunc main() { println(\"dedup test content\") }\n")

	f1 := filepath.Join(dir, "same.go")
	if err := os.WriteFile(f1, content, 0o644); err != nil {
		t.Fatal(err)
	}

	nodes := []*syntax.Node{
		syntax.NewSyntheticFileNode(f1, len(content)),
		syntax.NewSyntheticFileNode(f1, len(content)),
		syntax.NewSyntheticFileNode(f1, len(content)),
	}

	fd := NewFileDetector(1)
	matches := collectMatches(fd.FindDuplOver(nodes, 1))

	if len(matches) != 0 {
		t.Errorf(
			"expected 0 matches for same file referenced 3x (only 1 unique file), got %d",
			len(matches),
		)
	}
}

// --- extractUniqueFiles ---

func TestExtractUniqueFiles_Deduplicates(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector(1)

	nodes := []*syntax.Node{
		{Filename: "a.go"},
		{Filename: "b.go"},
		{Filename: "a.go"},
		{Filename: "c.go"},
		{Filename: "b.go"},
	}

	files := fd.extractUniqueFiles(nodes)
	if len(files) != 3 {
		t.Errorf("expected 3 unique files, got %d: %v", len(files), files)
	}
}

func TestExtractUniqueFiles_SkipsEmptyFilenames(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector(1)

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

	fd := NewFileDetector(1)

	files := fd.extractUniqueFiles(nil)
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

// --- hashFile error paths ---

func TestHashFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector(1)

	fh, ok := fd.hashFile("/nonexistent/path/file.go")
	if ok {
		t.Error("expected ok=false for nonexistent file")
	}

	if fh.Hash != "" || fh.Filename != "" || fh.Size != 0 {
		t.Errorf("expected zero FileHash, got %+v", fh)
	}
}

func TestHashFile_UnreadableFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	f := filepath.Join(dir, "unreadable.go")

	if err := os.WriteFile(f, []byte("content"), 0o000); err != nil {
		t.Fatal(err)
	}

	fd := NewFileDetector(1)

	fh, ok := fd.hashFile(f)

	if ok {
		t.Error("expected ok=false for unreadable file")
	}

	if fh.Hash != "" {
		t.Errorf("expected empty hash, got %q", fh.Hash)
	}
}

func TestHashFile_ValidFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	f := filepath.Join(dir, "valid.go")
	content := []byte("package main\n\nfunc main() {}\n")

	if err := os.WriteFile(f, content, 0o644); err != nil {
		t.Fatal(err)
	}

	fd := NewFileDetector(1)

	fh, ok := fd.hashFile(f)
	if !ok {
		t.Fatal("expected ok=true for valid file")
	}

	if fh.Filename != f {
		t.Errorf("expected filename %q, got %q", f, fh.Filename)
	}

	if fh.Size != len(content) {
		t.Errorf("expected size %d, got %d", len(content), fh.Size)
	}

	if fh.Hash == "" {
		t.Error("expected non-empty hash")
	}
}

// --- fileError ---

func TestFileError_ReturnsZeroState(t *testing.T) {
	t.Parallel()

	fd := NewFileDetector(1)

	fh, ok := fd.fileError("test.go", os.ErrNotExist, "file not found")
	if ok {
		t.Error("expected ok=false")
	}

	if fh != (FileHash{}) {
		t.Errorf("expected zero FileHash, got %+v", fh)
	}
}

// --- FileHash / FileDuplicate struct coverage ---

func TestFileHash_Fields(t *testing.T) {
	t.Parallel()

	fh := FileHash{Hash: "abc123", Filename: "test.go", Size: 42}
	if fh.Hash != "abc123" || fh.Filename != "test.go" || fh.Size != 42 {
		t.Errorf("FileHash fields not set correctly: %+v", fh)
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

	if fd.Hash != "deadbeef" {
		t.Errorf("expected Hash 'deadbeef', got %q", fd.Hash)
	}

	if len(fd.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(fd.Files))
	}
}

// --- HashDetector wrapper ---

func TestNewHashDetector(t *testing.T) {
	t.Parallel()

	hd := NewHashDetector(42)
	if hd == nil {
		t.Fatal("expected non-nil HashDetector")
	}
}

func TestHashDetector_FindDuplOver_Delegates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := []byte("package main\n\nfunc main() { println(\"delegation test!\") }\n")

	f1 := filepath.Join(dir, "d1.go")
	f2 := filepath.Join(dir, "d2.go")

	if err := os.WriteFile(f1, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(f2, content, 0o644); err != nil {
		t.Fatal(err)
	}

	nodes := []*syntax.Node{
		syntax.NewSyntheticFileNode(f1, len(content)),
		syntax.NewSyntheticFileNode(f2, len(content)),
	}

	hd := NewHashDetector(1)
	matches := collectMatches(hd.FindDuplOver(nodes, 1))

	if len(matches) != 1 {
		t.Errorf("expected 1 match via HashDetector, got %d", len(matches))
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
