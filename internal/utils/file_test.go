package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFindProjectRoot tests func FindProjectRoot function.
func TestFindProjectRoot(t *testing.T) {
	t.Run("finds project root with go.mod", func(t *testing.T) {
		// Create temporary directory structure
		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")
		deepDir := filepath.Join(subDir, "deep")

		// Create directories
		require.NoError(t, os.MkdirAll(deepDir, 0o755))

		// Create go.mod in root
		goModPath := filepath.Join(rootDir, "go.mod")
		require.NoError(t, os.WriteFile(goModPath, []byte("module test"), 0o644))

		// Test from deep subdirectory
		foundRoot, err := FindProjectRoot(deepDir, []string{"go.mod"})
		require.NoError(t, err)
		assert.Equal(t, rootDir, foundRoot, "should find project root")
	})

	t.Run("finds project root with .git", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		// Create directories
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Create .git in root
		gitDir := filepath.Join(rootDir, ".git")
		require.NoError(t, os.Mkdir(gitDir, 0o755))

		// Test from subdirectory
		foundRoot, err := FindProjectRoot(subDir, []string{".git"})
		require.NoError(t, err)
		assert.Equal(t, rootDir, foundRoot, "should find project root")
	})

	t.Run("finds project root with sqlc.yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		dbDir := filepath.Join(rootDir, "db")

		// Create directories
		require.NoError(t, os.MkdirAll(dbDir, 0o755))

		// Create sqlc.yaml in root
		sqlcPath := filepath.Join(rootDir, "sqlc.yaml")
		require.NoError(t, os.WriteFile(sqlcPath, []byte("version: 2"), 0o644))

		// Test from db subdirectory
		foundRoot, err := FindProjectRoot(dbDir, []string{"sqlc.yaml"})
		require.NoError(t, err)
		assert.Equal(t, rootDir, foundRoot, "should find project root")
	})

	t.Run("checks markers in order of preference", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		// Create directories
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Create both markers in root
		goModPath := filepath.Join(rootDir, "go.mod")
		require.NoError(t, os.WriteFile(goModPath, []byte("module test"), 0o644))

		gitDir := filepath.Join(rootDir, ".git")
		require.NoError(t, os.Mkdir(gitDir, 0o755))

		// Test with go.mod first in marker list
		foundRoot, err := FindProjectRoot(subDir, []string{"go.mod", ".git"})
		require.NoError(t, err)
		assert.Equal(t, rootDir, foundRoot, "should find project root")
	})

	t.Run("returns error when marker not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "project", "subdir")

		// Create directory without any markers
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Test - should fail to find marker
		_, err := FindProjectRoot(subDir, []string{"go.mod", ".git", "sqlc.yaml"})
		assert.Error(t, err, "should return error when no marker found")
	})

	t.Run("returns project root when already at root", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")

		// Create directory
		require.NoError(t, os.Mkdir(rootDir, 0o755))

		// Create go.mod
		goModPath := filepath.Join(rootDir, "go.mod")
		require.NoError(t, os.WriteFile(goModPath, []byte("module test"), 0o644))

		// Test from root directory itself
		foundRoot, err := FindProjectRoot(rootDir, []string{"go.mod"})
		require.NoError(t, err)
		assert.Equal(t, rootDir, foundRoot, "should return current directory when already at root")
	})

	t.Run("stops at filesystem root", func(t *testing.T) {
		// Create a temporary directory without markers
		tmpDir := t.TempDir()

		// Test from temp directory (likely no markers in parents)
		_, err := FindProjectRoot(tmpDir, []string{"go.mod", ".git"})
		assert.Error(t, err, "should return error after reaching filesystem root")
	})

	t.Run("handles absolute paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		// Create directories
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Create go.mod
		goModPath := filepath.Join(rootDir, "go.mod")
		require.NoError(t, os.WriteFile(goModPath, []byte("module test"), 0o644))

		// Get absolute path
		absSubDir, err := filepath.Abs(subDir)
		require.NoError(t, err)

		// Test with absolute path
		foundRoot, err := FindProjectRoot(absSubDir, []string{"go.mod"})
		require.NoError(t, err)
		assert.Equal(t, rootDir, foundRoot, "should handle absolute paths")
	})
}
