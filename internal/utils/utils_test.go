package utils

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileProcessorWriteFile(t *testing.T) {
	t.Run("writes file to base directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		err := fp.WriteFile("test.txt", []byte("hello world"), 0o644)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(content))
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		err := fp.WriteFile("subdir/nested/test.txt", []byte("nested content"), 0o644)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "subdir/nested/test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "nested content", string(content))
	})

	t.Run("writes without base directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor()

		filePath := filepath.Join(tmpDir, "test.txt")
		err := fp.WriteFile(filePath, []byte("absolute path"), 0o644)
		require.NoError(t, err)

		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, "absolute path", string(content))
	})
}

func TestFileProcessorWriteTextFile(t *testing.T) {
	t.Run("writes text file", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		err := fp.WriteTextFile("test.txt", "text content")
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "text content", string(content))
	})
}

func TestFileProcessorReadFile(t *testing.T) {
	t.Run("reads file from base directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		require.NoError(
			t,
			os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("read me"), 0o644),
		)

		content, err := fp.ReadFile("test.txt")
		require.NoError(t, err)
		assert.Equal(t, "read me", string(content))
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		_, err := fp.ReadFile("nonexistent.txt")
		assert.Error(t, err)
	})

	t.Run("reads without base directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor()

		filePath := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("absolute"), 0o644))

		content, err := fp.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, "absolute", string(content))
	})
}

func TestFileProcessorWriteTestFiles(t *testing.T) {
	t.Run("writes multiple test files", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		files := map[string]string{
			"file1.go": `package main

func main() {}`,
			"file2.go": `package main

func helper() {}`,
		}

		err := fp.WriteTestFiles(files)
		require.NoError(t, err)

		for filename, expectedContent := range files {
			content, err := os.ReadFile(filepath.Join(tmpDir, filename))
			require.NoError(t, err)
			assert.Equal(t, expectedContent, string(content))
		}
	})
}

func TestFileProcessorWriteDuplicateFiles(t *testing.T) {
	t.Run("writes files with identical content", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		filenames := []string{"dup1.go", "dup2.go", "dup3.go"}
		content := `package main

func duplicate() {
	println("same")
}`

		err := fp.WriteDuplicateFiles(filenames, content)
		require.NoError(t, err)

		for _, filename := range filenames {
			actualContent, err := os.ReadFile(filepath.Join(tmpDir, filename))
			require.NoError(t, err)
			assert.Equal(t, content, string(actualContent))
		}
	})
}

func TestFileProcessorRoundTrip(t *testing.T) {
	t.Run("write and read round trip", func(t *testing.T) {
		tmpDir := t.TempDir()
		fp := NewFileProcessor(tmpDir)

		original := []byte("round trip content with binary\x00data")

		err := fp.WriteFile("roundtrip.bin", original, 0o644)
		require.NoError(t, err)

		read, err := fp.ReadFile("roundtrip.bin")
		require.NoError(t, err)
		assert.True(t, bytes.Equal(original, read))
	})
}

func TestApplyTimeout(t *testing.T) {
	t.Run("returns original context when timeout is zero", func(t *testing.T) {
		ctx := context.Background()

		resultCtx, cancel := ApplyTimeout(ctx, 0)
		defer cancel()

		assert.Equal(t, ctx, resultCtx)
		assert.NoError(t, resultCtx.Err())
	})

	t.Run("returns original context when timeout is negative", func(t *testing.T) {
		ctx := context.Background()

		resultCtx, cancel := ApplyTimeout(ctx, -1)
		defer cancel()

		assert.Equal(t, ctx, resultCtx)
		assert.NoError(t, resultCtx.Err())
	})

	t.Run("creates timeout context when positive", func(t *testing.T) {
		ctx := context.Background()

		resultCtx, cancel := ApplyTimeout(ctx, 5)
		defer cancel()

		assert.NotEqual(t, ctx, resultCtx)

		deadline, hasDeadline := resultCtx.Deadline()
		assert.True(t, hasDeadline)
		assert.True(t, time.Now().Before(deadline))
	})

	t.Run("context expires after timeout", func(t *testing.T) {
		ctx := context.Background()

		resultCtx, cancel := ApplyTimeout(ctx, 1)
		defer cancel()

		select {
		case <-resultCtx.Done():
			t.Error("context should not be done immediately")
		default:
		}

		time.Sleep(1100 * time.Millisecond)

		assert.Error(t, resultCtx.Err())
		assert.Equal(t, context.DeadlineExceeded, resultCtx.Err())
	})

	t.Run("cancel function works", func(t *testing.T) {
		ctx := context.Background()
		resultCtx, cancel := ApplyTimeout(ctx, 10)

		cancel()

		assert.Error(t, resultCtx.Err())
		assert.Equal(t, context.Canceled, resultCtx.Err())
	})
}
