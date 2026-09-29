package testutil

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestCopyToBufferSurfacesCopyError(t *testing.T) {
	t.Parallel()

 boom := errors.New("read exploded")
	src := &errReader{err: boom, n: 3}
	var dst strings.Builder

	done := make(chan error)
	CopyToBuffer(&dst, src, done)

	copyErr := <-done
	if !errors.Is(copyErr, boom) {
		t.Fatalf("CopyToBuffer error = %v, want %v", copyErr, boom)
	}
	if dst.Len() == 0 {
		t.Error("partial data written before the error was lost")
	}
}

func TestCopyToBufferNilErrorOnSuccess(t *testing.T) {
	t.Parallel()

	var dst strings.Builder

	done := make(chan error)
	CopyToBuffer(&dst, strings.NewReader("hello"), done)

	if err := <-done; err != nil {
		t.Fatalf("unexpected error on clean copy: %v", err)
	}
	if dst.String() != "hello" {
		t.Errorf("dst = %q, want %q", dst.String(), "hello")
	}
}

// errReader yields n bytes then fails with err.
type errReader struct {
	err error
	n   int
}

func (r *errReader) Read(p []byte) (int, error) {
	if r.n <= 0 {
		return 0, r.err
	}
	n := min(r.n, len(p))
	r.n -= n

	for i := range n {
		p[i] = 'x'
	}

	return n, nil
}

var _ io.Reader = (*errReader)(nil)
