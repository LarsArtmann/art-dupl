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
		name        string
		filename    string
		content     string
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
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor()

		filePath := filepath.Join(tmpDir, "test.txt")
		err := fp.WriteFile(filePath, []byte("absolute path"), 0o644)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		content, err := os.ReadFile(filePath)
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(string(content)).To(gomega.Equal("absolute path"))
	})
}

func TestFileProcessorWriteTextFile(t *testing.T) {
	t.Parallel()

	t.Run("writes text file", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		err := fp.WriteTextFile("test.txt", "text content")
		g.Expect(err).ToNot(gomega.HaveOccurred())

		content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(string(content)).To(gomega.Equal("text content"))
	})
}

func TestFileProcessorReadFile(t *testing.T) {
	t.Parallel()

	t.Run("reads file from base directory", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		g.Expect(os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("read me"), 0o644)).
			To(gomega.Succeed())

		content, err := fp.ReadFile("test.txt")
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(string(content)).To(gomega.Equal("read me"))
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
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		fp := NewFileProcessor()

		filePath := filepath.Join(tmpDir, "test.txt")
		g.Expect(os.WriteFile(filePath, []byte("absolute"), 0o644)).To(gomega.Succeed())

		content, err := fp.ReadFile(filePath)
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(string(content)).To(gomega.Equal("absolute"))
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
			content, err := os.ReadFile(filepath.Join(tmpDir, filename))
			g.Expect(err).ToNot(gomega.HaveOccurred())
			g.Expect(string(content)).To(gomega.Equal(expectedContent))
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
			actualContent, err := os.ReadFile(filepath.Join(tmpDir, filename))
			g.Expect(err).ToNot(gomega.HaveOccurred())
			g.Expect(string(actualContent)).To(gomega.Equal(content))
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

func assertWriteFile(t *testing.T, filename, expectedContent string) {
	t.Helper()
	g := gomega.NewWithT(t)
	tmpDir := t.TempDir()
	fp := NewFileProcessor(tmpDir)

	err := fp.WriteFile(filename, []byte(expectedContent), 0o644)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	content, err := os.ReadFile(filepath.Join(tmpDir, filename))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(content)).To(gomega.Equal(expectedContent))
}

func TestApplyTimeout(t *testing.T) {
	t.Parallel()

	t.Run("returns original context when timeout is zero", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		resultCtx, cancel := ApplyTimeout(ctx, 0)
		defer cancel()

		g.Expect(resultCtx).To(gomega.Equal(ctx))
		g.Expect(resultCtx.Err()).ToNot(gomega.HaveOccurred())
	})

	t.Run("returns original context when timeout is negative", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		resultCtx, cancel := ApplyTimeout(ctx, -1)
		defer cancel()

		g.Expect(resultCtx).To(gomega.Equal(ctx))
		g.Expect(resultCtx.Err()).ToNot(gomega.HaveOccurred())
	})

	t.Run("creates timeout context when positive", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		resultCtx, cancel := ApplyTimeout(ctx, 5)
		defer cancel()

		g.Expect(resultCtx).ToNot(gomega.Equal(ctx))

		deadline, hasDeadline := resultCtx.Deadline()
		g.Expect(hasDeadline).To(gomega.BeTrue())
		g.Expect(time.Now().Before(deadline)).To(gomega.BeTrue())
	})

	t.Run("context expires after timeout", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()

		resultCtx, cancel := ApplyTimeout(ctx, 1)
		defer cancel()

		select {
		case <-resultCtx.Done():
			t.Error("context should not be done immediately")
		default:
		}

		time.Sleep(1100 * time.Millisecond)

		g.Expect(resultCtx.Err()).To(gomega.HaveOccurred())
		g.Expect(resultCtx.Err()).To(gomega.Equal(context.DeadlineExceeded))
	})

	t.Run("cancel function works", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		ctx := t.Context()
		resultCtx, cancel := ApplyTimeout(ctx, 10)

		cancel()

		g.Expect(resultCtx.Err()).To(gomega.HaveOccurred())
		g.Expect(resultCtx.Err()).To(gomega.Equal(context.Canceled))
	})
}
