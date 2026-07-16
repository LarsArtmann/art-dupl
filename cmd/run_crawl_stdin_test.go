package cmd

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"
)

// TestStdinFeed_CancelUnblocksScanner verifies that cancelling the context
// unblocks the inherently blocking bufio.Scanner.Scan by closing the reader.
// Without the close-on-cancel watcher, this test would hang until the timeout.
func TestStdinFeed_CancelUnblocksScanner(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	t.Cleanup(func() {
		_ = w.Close()
		_ = r.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ch := feedFromStdin(ctx, r, nil, nil, generatorIncludes{}, "")

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel closed, got value")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out: stdin reader did not exit after context cancellation")
	}
}

// TestStdinFeed_ReadsPaths verifies that file paths piped to the reader are
// emitted on the channel, and that the channel closes on EOF.
func TestStdinFeed_ReadsPaths(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}

	want := []string{"foo.go", "bar.go", "baz.go"}
	for _, p := range want {
		if _, err := w.WriteString(p + "\n"); err != nil {
			t.Fatalf("WriteString error: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ch := feedFromStdin(ctx, r, nil, nil, generatorIncludes{}, "")

	var got []string
	for p := range ch {
		got = append(got, p)
	}

	if !slices.Equal(got, want) {
		t.Errorf("stdin paths = %v, want %v", got, want)
	}
}

// TestStdinFeed_PartialReadThenCancel verifies that paths already sent
// before cancellation are received, and that the goroutine exits promptly
// after cancel even if the pipe is still open.
func TestStdinFeed_PartialReadThenCancel(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	t.Cleanup(func() {
		_ = w.Close()
		_ = r.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ch := feedFromStdin(ctx, r, nil, nil, generatorIncludes{}, "")

	if _, err := w.WriteString("first.go\n"); err != nil {
		t.Fatalf("WriteString error: %v", err)
	}
	if _, err := w.WriteString("second.go\n"); err != nil {
		t.Fatalf("WriteString error: %v", err)
	}

	got := make([]string, 0, 2)
	readDone := make(chan struct{})
	go func() {
		for p := range ch {
			got = append(got, p)
		}
		close(readDone)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case <-readDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out: stdin reader did not exit after cancellation")
	}

	if len(got) < 2 {
		t.Errorf("expected at least 2 paths before cancellation, got %d: %v", len(got), got)
	}
}
