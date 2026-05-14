package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
)

// assertMkdirAll creates a directory tree and asserts success.
func assertMkdirAll(t *testing.T, dir string) {
	t.Helper()
	g := gomega.NewWithT(t)
	g.Expect(os.MkdirAll(dir, 0o755)).To(gomega.Succeed())
}

// assertMkdir creates a directory and asserts success.
func assertMkdir(t *testing.T, dir string) {
	t.Helper()
	g := gomega.NewWithT(t)
	g.Expect(os.Mkdir(dir, 0o755)).To(gomega.Succeed())
}

// TestFindProjectRoot tests func FindProjectRoot function.
func TestFindProjectRoot(t *testing.T) {
	t.Parallel()

	t.Run("finds project root with go.mod", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")
		deepDir := filepath.Join(subDir, "deep")

		assertMkdirAll(t, deepDir)

		goModPath := filepath.Join(rootDir, "go.mod")

		err := os.WriteFile(goModPath, []byte("module test"), 0o644)
		if err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		foundRoot, err := FindProjectRoot(deepDir, []string{"go.mod"})
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(foundRoot).To(gomega.Equal(rootDir))
	})

	t.Run("finds project root with .git", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		assertMkdirAll(t, subDir)

		gitDir := filepath.Join(rootDir, ".git")
		assertMkdir(t, gitDir)

		foundRoot, err := FindProjectRoot(subDir, []string{".git"})
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(foundRoot).To(gomega.Equal(rootDir))
	})

	t.Run("finds project root with sqlc.yaml", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		dbDir := filepath.Join(rootDir, "db")

		assertMkdirAll(t, dbDir)

		sqlcPath := filepath.Join(rootDir, "sqlc.yaml")
		g.Expect(os.WriteFile(sqlcPath, []byte("version: 2"), 0o644)).To(gomega.Succeed())

		foundRoot, err := FindProjectRoot(dbDir, []string{"sqlc.yaml"})
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(foundRoot).To(gomega.Equal(rootDir))
	})

	t.Run("checks markers in order of preference", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		assertMkdirAll(t, subDir)

		goModPath := filepath.Join(rootDir, "go.mod")

		err := os.WriteFile(goModPath, []byte("module test"), 0o644)
		if err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		gitDir := filepath.Join(rootDir, ".git")
		assertMkdir(t, gitDir)

		foundRoot, err := FindProjectRoot(subDir, []string{"go.mod", ".git"})
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(foundRoot).To(gomega.Equal(rootDir))
	})

	t.Run("returns error when marker not found", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "project", "subdir")

		assertMkdirAll(t, subDir)

		_, err := FindProjectRoot(subDir, []string{"nonexistent-marker-test-only.xyz"})
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("returns project root when already at root", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")

		assertMkdir(t, rootDir)

		goModPath := filepath.Join(rootDir, "go.mod")

		err := os.WriteFile(goModPath, []byte("module test"), 0o644)
		if err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		foundRoot, err := FindProjectRoot(rootDir, []string{"go.mod"})
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(foundRoot).To(gomega.Equal(rootDir))
	})

	t.Run("stops at filesystem root", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()

		_, err := FindProjectRoot(tmpDir, []string{"nonexistent-marker-test-only.xyz"})
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("handles absolute paths", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		assertMkdirAll(t, subDir)

		goModPath := filepath.Join(rootDir, "go.mod")

		err := os.WriteFile(goModPath, []byte("module test"), 0o644)
		if err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		absSubDir, err := filepath.Abs(subDir)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		foundRoot, err := FindProjectRoot(absSubDir, []string{"go.mod"})
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(foundRoot).To(gomega.Equal(rootDir))
	})
}
