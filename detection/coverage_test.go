package detection

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestHasNonEmptyFrag(t *testing.T) {
	t.Parallel()

	nonEmpty := []*syntax.Node{{Type: int32(golang.File)}}

	cases := []struct {
		name  string
		frags [][]*syntax.Node
		want  bool
	}{
		{name: "nil", frags: nil, want: false},
		{name: "all empty inner slices", frags: [][]*syntax.Node{{}, {}}, want: false},
		{name: "mixed empty and non-empty", frags: [][]*syntax.Node{{}, nonEmpty}, want: true},
		{name: "all non-empty", frags: [][]*syntax.Node{nonEmpty, nonEmpty}, want: true},
		{name: "single non-empty", frags: [][]*syntax.Node{nonEmpty}, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := hasNonEmptyFrag(tc.frags); got != tc.want {
				t.Errorf("hasNonEmptyFrag() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStreamMatches_FiltersEmptyFrags(t *testing.T) {
	t.Parallel()

	src := make(chan syntax.Match, 3)
	src <- syntax.Match{Frags: [][]*syntax.Node{{}}} // empty → filtered

	src <- syntax.Match{Frags: [][]*syntax.Node{{}, {}}} // empty → filtered

	src <- syntax.Match{Frags: [][]*syntax.Node{
		{&syntax.Node{Type: int32(golang.File)}},
	}} // non-empty → forwarded

	close(src)

	dst := make(chan syntax.Match, 3)
	md := &MultiDetector{}
	md.streamMatches(context.Background(), src, dst)
	close(dst)

	var received []syntax.Match
	for m := range dst {
		received = append(received, m)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 forwarded match, got %d", len(received))
	}
}

func TestStreamMatches_StopsOnCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// streamMatches only checks ctx when a value arrives on src (the range
	// blocks otherwise). Buffer a value so the loop unblocks, observes the
	// cancelled context, and returns.
	src := make(chan syntax.Match, 1)
	src <- syntax.Match{Frags: [][]*syntax.Node{{}}}

	dst := make(chan syntax.Match)

	md := &MultiDetector{}

	done := make(chan struct{})
	go func() {
		defer close(done)

		md.streamMatches(ctx, src, dst)
	}()

	select {
	case <-done:
		// streamMatches returned due to cancelled ctx — expected.
	case <-time.After(time.Second):
		t.Fatal("streamMatches did not return after context cancellation")
	}
}

// fakeDetector is a MethodDetector used to verify the Name() contract for
// detectors outside the built-in adapters.
type fakeDetector struct{}

func (f *fakeDetector) FindDuplOver(_ context.Context, _ int) <-chan syntax.Match {
	ch := make(chan syntax.Match)
	close(ch)

	return ch
}

func (f *fakeDetector) Name() string {
	return "detection"
}

func TestDetectorName(t *testing.T) {
	t.Parallel()

	md := &MultiDetector{data: []*syntax.Node{}, tree: suffixtree.New()}

	cases := []struct {
		name string
		det  MethodDetector
		want string
	}{
		{"suffix tree adapter", &suffixTreeAdapter{tree: md.tree, data: md.data}, "suffix tree-based detection"},
		{"hash adapter", &hashAdapter{data: md.data}, "hash-based detection"},
		{"fake detector", &fakeDetector{}, "detection"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.det.Name(); got != tc.want {
				t.Errorf("Name() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildCloneDetectors(t *testing.T) {
	t.Parallel()

	md := &MultiDetector{data: []*syntax.Node{}, tree: suffixtree.New()}

	cases := []struct {
		name    string
		methods []domain.DetectionMethod
		want    int
	}{
		{"empty defaults to art-dupl", nil, 1},
		{"art-dupl only", []domain.DetectionMethod{MethodArtDupl}, 1},
		{"hash only", []domain.DetectionMethod{MethodHash}, 1},
		{"both methods", []domain.DetectionMethod{MethodArtDupl, MethodHash}, 2},
		{"unknown method ignored", []domain.DetectionMethod{"bogus"}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			md.cfg = Config{Methods: tc.methods}

			dets := md.buildCloneDetectors()
			if len(dets) != tc.want {
				t.Errorf("buildCloneDetectors(%v) = %d detectors, want %d", tc.methods, len(dets), tc.want)
			}
		})
	}
}

func TestSuffixTreeAdapter_StopsOnCancelledContext(t *testing.T) {
	t.Parallel()

	adapter := &suffixTreeAdapter{tree: suffixtree.New(), data: nil}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)

		for range adapter.FindDuplOver(ctx, 15) {
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("suffixTreeAdapter did not drain after context cancellation")
	}
}

func TestHashAdapter_StopsOnCancelledContext(t *testing.T) {
	t.Parallel()

	adapter := &hashAdapter{data: nil}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)

		for range adapter.FindDuplOver(ctx, 15) {
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("hashAdapter did not drain after context cancellation")
	}
}

// buildDuplicateData creates a node stream containing the same statement
// subtree twice (different filenames), so the suffix tree reports one match.
// A unique sentinel terminator is appended — Ukkonen's algorithm requires it
// to force every suffix into an explicit leaf, otherwise short repeats are
// missed. The returned data slice aligns positionally with the tree's data.
func buildDuplicateData() ([]*syntax.Node, *suffixtree.STree) {
	mk := func(filename string) []*syntax.Node {
		root := &syntax.Node{
			Filename: filename,
			Type:     int32(golang.FuncDecl),
			Children: []*syntax.Node{
				{Filename: filename, Type: int32(golang.BlockStmt)},
				{Filename: filename, Type: int32(golang.ReturnStmt)},
			},
		}

		return syntax.Serialize(root)
	}

	data := append(mk("a.go"), mk("b.go")...)
	// Unique sentinel terminator (value outside the AST type range).
	data = append(data, &syntax.Node{Type: 1 << 20, Filename: "sentinel.go"})

	tokens := make([]suffixtree.Token, 0, len(data))
	for _, n := range data {
		tokens = append(tokens, n)
	}

	tree := suffixtree.New()
	if err := tree.Update(tokens...); err != nil {
		panic("unexpected tree.Update error: " + err.Error())
	}

	return data, tree
}

func TestSuffixTreeAdapter_FindsDuplicates(t *testing.T) {
	t.Parallel()

	data, tree := buildDuplicateData()
	adapter := &suffixTreeAdapter{tree: tree, data: data}

	var count int

	for match := range adapter.FindDuplOver(context.Background(), 2) {
		if !hasNonEmptyFrag(match.Frags) {
			t.Error("adapter emitted a match with no fragments")
		}

		count++
	}

	if count == 0 {
		t.Fatal("expected at least one duplicate match, got none")
	}
}

func TestHashAdapter_FindsDuplicates(t *testing.T) {
	t.Parallel()

	// The hash detector reads files from disk and compares XXH3 hashes, so we
	// need two real byte-identical files.
	dir := t.TempDir()

	content := strings.Repeat("package main\n", 10) // > threshold bytes
	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")

	if err := os.WriteFile(f1, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(f2, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	data := []*syntax.Node{
		{Filename: f1, Type: int32(golang.File)},
		{Filename: f2, Type: int32(golang.File)},
	}

	adapter := &hashAdapter{data: data}

	var count int

	for match := range adapter.FindDuplOver(context.Background(), 50) {
		if !hasNonEmptyFrag(match.Frags) {
			t.Error("hash adapter emitted a match with no fragments")
		}

		count++
	}

	if count == 0 {
		t.Fatal("expected at least one hash duplicate match, got none")
	}
}

// TestStreamMatches_UnblocksOnDone exercises the select's ctx.Done branch:
// the destination is never read so the send blocks, then cancelling ctx must
// release the goroutine.
func TestStreamMatches_UnblocksOnDone(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	src := make(chan syntax.Match, 1)
	src <- syntax.Match{Frags: [][]*syntax.Node{
		{&syntax.Node{Type: int32(golang.File)}},
	}}

	close(src)

	dst := make(chan syntax.Match) // unbuffered + never read → send blocks

	md := &MultiDetector{}

	done := make(chan struct{})
	go func() {
		defer close(done)

		md.streamMatches(ctx, src, dst)
	}()

	// Allow the goroutine to reach the blocking select, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("streamMatches did not release via ctx.Done select branch")
	}
}
