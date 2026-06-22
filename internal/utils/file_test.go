package utils

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestFileProcessorWriteFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{"writes file to base directory", "test.txt", "hello world"},
		{"creates nested directories", "subdir/nested/test.txt", "nested content"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertWriteFile(t, tt.filename, tt.content)
		})
	}

	t.Run("writes without base directory", func(t *testing.T) {
		t.Parallel()
		assertFileContentAbsolute(t, "test.txt", "absolute path")
	})
}

// assertFileContentAbsolute writes a file without base directory and asserts content.
func assertFileContentAbsolute(t *testing.T, filename, expectedContent string) {
	t.Helper()
	g := gomega.NewWithT(t)
	tmpDir := t.TempDir()
	fp := NewFileProcessor()

	filePath := filepath.Join(tmpDir, filename)
	err := fp.WriteFile(filePath, []byte(expectedContent), 0o644)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	rawBytes, err := os.ReadFile(filePath)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(rawBytes)).To(gomega.Equal(expectedContent))
}

// assertFileContentWithBase writes a file with base directory and asserts rawBytes.
func assertFileContentWithBase(t *testing.T, filename, expectedContent string) {
	t.Helper()
	g := gomega.NewWithT(t)
	tmpDir := t.TempDir()
	fp := NewFileProcessor(tmpDir)

	err := fp.WriteTextFile(filename, expectedContent)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	readContent, err := os.ReadFile(filepath.Join(tmpDir, filename))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(readContent)).To(gomega.Equal(expectedContent))
}

// assertReadFile writes a file to base dir and reads it back via fp.ReadFile.
func assertReadFile(t *testing.T, filename, fileData string) {
	t.Helper()
	g := gomega.NewWithT(t)
	tmpDir := t.TempDir()
	fp := NewFileProcessor(tmpDir)

	g.Expect(os.WriteFile(filepath.Join(tmpDir, filename), []byte(fileData), 0o644)).
		To(gomega.Succeed())

	loaded, err := fp.ReadFile(filename)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(loaded)).To(gomega.Equal(fileData))
}

// assertReadFileAbsolute writes a file without base dir and reads it back via fp.ReadFile.
func assertReadFileAbsolute(t *testing.T, filename, loadedBytes string) {
	t.Helper()
	g := gomega.NewWithT(t)
	tmpDir := t.TempDir()
	fp := NewFileProcessor()

	filePath := filepath.Join(tmpDir, filename)
	g.Expect(os.WriteFile(filePath, []byte(loadedBytes), 0o644)).To(gomega.Succeed())

	read, err := fp.ReadFile(filePath)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(read)).To(gomega.Equal(loadedBytes))
}

func TestFileProcessorWriteTextFile(t *testing.T) {
	t.Parallel()

	t.Run("writes text file", func(t *testing.T) {
		t.Parallel()
		assertFileContentWithBase(t, "test.txt", "text content")
	})
}

func TestFileProcessorReadFile(t *testing.T) {
	t.Parallel()

	t.Run("reads file from base directory", func(t *testing.T) {
		t.Parallel()
		assertReadFile(t, "test.txt", "read me")
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		_, err := fp.ReadFile("nonexistent.txt")
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("reads without base directory", func(t *testing.T) {
		t.Parallel()
		assertReadFileAbsolute(t, "test.txt", "absolute")
	})
}

func TestFileProcessorWriteTestFiles(t *testing.T) {
	t.Parallel()

	t.Run("writes multiple test files", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		files := map[string]string{
			"file1.go": `package main

func main() {}`,
			"file2.go": `package main

func helper() {}`,
		}

		err := fp.WriteTestFiles(files)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		for filename, expectedContent := range files {
			assertFileContent(t, filepath.Join(tmpDir, filename), expectedContent)
		}
	})
}

func TestFileProcessorWriteDuplicateFiles(t *testing.T) {
	t.Parallel()

	t.Run("writes files with identical content", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		filenames := []string{"dup1.go", "dup2.go", "dup3.go"}
		content := `package main

func duplicate() {
	println("same")
}`

		err := fp.WriteDuplicateFiles(filenames, content)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		for _, filename := range filenames {
			assertFileContent(t, filepath.Join(tmpDir, filename), content)
		}
	})
}

func TestFileProcessorRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("write and read round trip", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		original := []byte("round trip content with binary\x00data")

		err := fp.WriteFile("roundtrip.bin", original, 0o644)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		read, err := fp.ReadFile("roundtrip.bin")
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(bytes.Equal(original, read)).To(gomega.BeTrue())
	})
}

// assertFileContent reads a file and asserts its content equals expectedContent.
func assertFileContent(t *testing.T, filePath, expectedContent string) {
	t.Helper()
	g := gomega.NewWithT(t)
	diskContent, err := os.ReadFile(filePath)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(diskContent)).To(gomega.Equal(expectedContent))
}

func assertWriteFile(t *testing.T, filename, expectedContent string) {
	t.Helper()
	tmpDir := t.TempDir()
	fp := NewFileProcessor(tmpDir)

	err := fp.WriteFile(filename, []byte(expectedContent), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	assertFileContent(t, filepath.Join(tmpDir, filename), expectedContent)
}

func TestApplyTimeout(t *testing.T) {
	t.Parallel()

	t.Run("returns original context when timeout is zero", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		zeroTimeoutCtx, cancel := ApplyTimeout(ctx, 0)
		defer cancel()

		g.Expect(zeroTimeoutCtx).To(gomega.Equal(ctx))
		g.Expect(zeroTimeoutCtx.Err()).ToNot(gomega.HaveOccurred())
	})

	t.Run("returns original context when timeout is negative", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		negTimeoutCtx, cancel := ApplyTimeout(ctx, -1)
		defer cancel()

		g.Expect(negTimeoutCtx).To(gomega.Equal(ctx))
		g.Expect(negTimeoutCtx.Err()).ToNot(gomega.HaveOccurred())
	})

	t.Run("creates timeout context when positive", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		deadlineCtx, cancel := ApplyTimeout(ctx, 5*time.Second)
		defer cancel()

		g.Expect(deadlineCtx).ToNot(gomega.Equal(ctx))

		deadline, hasDeadline := deadlineCtx.Deadline()
		g.Expect(hasDeadline).To(gomega.BeTrue())
		g.Expect(time.Now().Before(deadline)).To(gomega.BeTrue())
	})

	t.Run("context expires after timeout", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		deadlineCtx, cancel := ApplyTimeout(ctx, 1*time.Second)
		defer cancel()

		select {
		case <-deadlineCtx.Done():
			t.Error("context should not be done immediately")
		default:
		}

		time.Sleep(1100 * time.Millisecond)

		g.Expect(deadlineCtx.Err()).To(gomega.HaveOccurred())
		g.Expect(deadlineCtx.Err()).To(gomega.Equal(context.DeadlineExceeded))
	})

	t.Run("cancel function works", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()
		cancelableCtx, cancel := ApplyTimeout(ctx, 10*time.Second)

		cancel()

		g.Expect(cancelableCtx.Err()).To(gomega.HaveOccurred())
		g.Expect(cancelableCtx.Err()).To(gomega.Equal(context.Canceled))
	})
}
