package testutil

import (
	"bytes"
	"os"
	"sync"
)

// captureMu serializes stdout/stderr replacement so that concurrent captures
// don't clobber each other's global file descriptors.
//
// IMPORTANT: This mutex is NOT sufficient to make capture safe under t.Parallel.
// It only prevents two captures from overlapping each other. It does NOT
// synchronize against sibling tests that read os.Stdout/os.Stderr directly
// (e.g. fmt.Printf, fmt.Fprintln(os.Stderr)). While a capture holds these
// globals pointing at its pipes, a parallel direct reader observes a mutated
// pointer and the race detector flags the unsynchronized read/write of the
// os.Stdout/os.Stderr variables themselves.
//
// Therefore: tests that use CaptureStdoutStderr/CaptureCombinedOutput MUST be
// serial (do NOT call t.Parallel) whenever ANY sibling test in the same package
// touches os.Stdout or os.Stderr directly. The same applies to any test that
// reassigns os.Stdout/os.Stderr by hand.
//
//nolint:gochecknoglobals // mutex must be package-level to serialize captures
var captureMu sync.Mutex

// CaptureStdoutStderr runs fn with os.Stdout and os.Stderr temporarily replaced
// by pipes, returning whatever was written during the call.
//
// Captures are serialized against each other via captureMu, but this does NOT
// make them safe to run from a t.Parallel test: sibling tests that read
// os.Stdout/os.Stderr directly (fmt.Printf, fmt.Fprintln(os.Stderr), etc.) will
// race on the global pointer mutation. See the captureMu doc for the full
// contract. Callers must be serial.
//
//nolint:nonamedreturns // Multiple return values for stdout/stderr
func CaptureStdoutStderr(fn func() error) (stdout, stderr []byte, err error) {
	captureMu.Lock()
	defer captureMu.Unlock()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, pipeErr := os.Pipe()
	if pipeErr != nil {
		return nil, nil, pipeErr
	}

	stderrR, stderrW, pipeErr := os.Pipe()
	if pipeErr != nil {
		_ = stdoutR.Close()
		_ = stdoutW.Close()

		return nil, nil, pipeErr
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)

	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})

	CopyToBuffer(&stdoutBuf, stdoutR, stdoutDone)
	CopyToBuffer(&stderrBuf, stderrR, stderrDone)

	err = fn()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	_ = stdoutW.Close()
	_ = stderrW.Close()

	<-stdoutDone
	<-stderrDone

	return stdoutBuf.Bytes(), stderrBuf.Bytes(), err
}

// CaptureCombinedOutput is like CaptureStdoutStderr but returns stdout and
// stderr concatenated into a single byte slice (stdout first, then stderr).
func CaptureCombinedOutput(fn func() error) ([]byte, error) {
	stdout, stderr, err := CaptureStdoutStderr(fn)

	combined := make([]byte, 0, len(stdout)+len(stderr))
	combined = append(combined, stdout...)
	combined = append(combined, stderr...)

	return combined, err
}
