package testutil

import (
	"bytes"
	"os"
	"sync"
)

// captureMu serializes stdout/stderr replacement so that concurrent captures
// (e.g. parallel tests) don't clobber each other's global file descriptors.
//
//nolint:gochecknoglobals // mutex must be package-level to serialize captures
var captureMu sync.Mutex

// CaptureStdoutStderr runs fn with os.Stdout and os.Stderr temporarily replaced
// by pipes, returning whatever was written during the call.
// It is safe for concurrent use (captures are serialized via a mutex).
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
