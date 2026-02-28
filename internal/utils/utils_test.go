package utils

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

// TestFindProjectRoot tests func FindProjectRoot function.
func TestFindProjectRoot(t *testing.T) {
	t.Parallel()

	t.Run("finds project root with go.mod", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")
		deepDir := filepath.Join(subDir, "deep")

		g.Expect(os.MkdirAll(deepDir, 0o755)).To(Succeed())

		goModPath := filepath.Join(rootDir, "go.mod")
		g.Expect(os.WriteFile(goModPath, []byte("module test"), 0o644)).To(Succeed())

		foundRoot, err := FindProjectRoot(deepDir, []string{"go.mod"})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundRoot).To(Equal(rootDir))
	})

	t.Run("finds project root with .git", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		g.Expect(os.MkdirAll(subDir, 0o755)).To(Succeed())

		gitDir := filepath.Join(rootDir, ".git")
		g.Expect(os.Mkdir(gitDir, 0o755)).To(Succeed())

		foundRoot, err := FindProjectRoot(subDir, []string{".git"})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundRoot).To(Equal(rootDir))
	})

	t.Run("finds project root with sqlc.yaml", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		dbDir := filepath.Join(rootDir, "db")

		g.Expect(os.MkdirAll(dbDir, 0o755)).To(Succeed())

		sqlcPath := filepath.Join(rootDir, "sqlc.yaml")
		g.Expect(os.WriteFile(sqlcPath, []byte("version: 2"), 0o644)).To(Succeed())

		foundRoot, err := FindProjectRoot(dbDir, []string{"sqlc.yaml"})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundRoot).To(Equal(rootDir))
	})

	t.Run("checks markers in order of preference", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		g.Expect(os.MkdirAll(subDir, 0o755)).To(Succeed())

		goModPath := filepath.Join(rootDir, "go.mod")
		g.Expect(os.WriteFile(goModPath, []byte("module test"), 0o644)).To(Succeed())

		gitDir := filepath.Join(rootDir, ".git")
		g.Expect(os.Mkdir(gitDir, 0o755)).To(Succeed())

		foundRoot, err := FindProjectRoot(subDir, []string{"go.mod", ".git"})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundRoot).To(Equal(rootDir))
	})

	t.Run("returns error when marker not found", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "project", "subdir")

		g.Expect(os.MkdirAll(subDir, 0o755)).To(Succeed())

		_, err := FindProjectRoot(subDir, []string{"go.mod", ".git", "sqlc.yaml"})
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("returns project root when already at root", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")

		g.Expect(os.Mkdir(rootDir, 0o755)).To(Succeed())

		goModPath := filepath.Join(rootDir, "go.mod")
		g.Expect(os.WriteFile(goModPath, []byte("module test"), 0o644)).To(Succeed())

		foundRoot, err := FindProjectRoot(rootDir, []string{"go.mod"})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundRoot).To(Equal(rootDir))
	})

	t.Run("stops at filesystem root", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()

		_, err := FindProjectRoot(tmpDir, []string{"go.mod", ".git"})
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("handles absolute paths", func(t *testing.T) {
		g := NewWithT(t)
		t.Parallel()

		tmpDir := t.TempDir()
		rootDir := filepath.Join(tmpDir, "project")
		subDir := filepath.Join(rootDir, "subdir")

		g.Expect(os.MkdirAll(subDir, 0o755)).To(Succeed())

		goModPath := filepath.Join(rootDir, "go.mod")
		g.Expect(os.WriteFile(goModPath, []byte("module test"), 0o644)).To(Succeed())

		absSubDir, err := filepath.Abs(subDir)
		g.Expect(err).ToNot(HaveOccurred())

		foundRoot, err := FindProjectRoot(absSubDir, []string{"go.mod"})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundRoot).To(Equal(rootDir))
	})
}
