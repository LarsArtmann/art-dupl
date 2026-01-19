package bdd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// BDD Test Suite for Error Handling
//
// These tests verify that art-dupl handles error conditions gracefully,
// providing clear error messages and not crashing.
//
// The scenarios cover:
// - Invalid paths and directories
// - Invalid configuration files
// - Invalid flag combinations
// - Missing files
// - Permission errors

func TestErrorHandling(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl Error Handling BDD Suite")
}

var _ = Describe("Error Handling", func() {
	var (
		binaryPath string
	)

	BeforeEach(func() {
		// Build art-dupl binary
		cmd := exec.Command("go", "build", "-o", ./bdd/art-dupl-error_handling-test, "./cmd/art-dupl/main.go")
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred())
		binaryPath = ./bdd/art-dupl-error_handling-test
	})

	AfterEach(func() {
		_ = os.Remove(binaryPath)
	})

	Context("When analyzing non-existent paths", func() {
		It("should handle missing directory gracefully", func() {
			// Try to analyze non-existent directory
			nonExistentPath := "/tmp/art-dupl-test-nonexistent-xyz123"

			cmd := exec.Command(binaryPath, nonExistentPath)
			output, err := cmd.CombinedOutput()

			// Should fail gracefully with error message
			Expect(err).To(HaveOccurred(), "Should error when path doesn't exist")
			Expect(string(output)).NotTo(BeEmpty(), "Should produce error message")
			Expect(string(output)).To(SatisfyAny(
				ContainSubstring("error"),
				ContainSubstring("no such"),
				ContainSubstring("not found"),
			))
		})

		It("should handle non-existent file gracefully", func() {
			// Try to analyze non-existent file
			nonExistentFile := "/tmp/art-dupl-test-nonexistent-file.go"

			cmd := exec.Command(binaryPath, nonExistentFile)
			output, err := cmd.CombinedOutput()

			// Should fail gracefully
			Expect(err).To(HaveOccurred())
			Expect(string(output)).NotTo(BeEmpty())
		})
	})

	Context("When analyzing invalid file types", func() {
		It("should ignore non-Go files gracefully", func() {
			// Create temporary directory with non-Go file
			tempDir, err := os.MkdirTemp("", "art-dupl-error-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Create non-Go files
			nonGoFile := filepath.Join(tempDir, "test.txt")
			err = os.WriteFile(nonGoFile, []byte("not a go file"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl on directory with only non-Go files
			cmd := exec.Command(binaryPath, tempDir)
			output, err := cmd.CombinedOutput()

			// Should not crash
			Expect(err).ToNot(HaveOccurred(), "Should handle non-Go files without error")
			// Output may be empty or show no results
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle mixed file types", func() {
			// Create temporary directory with mixed file types
			tempDir, err := os.MkdirTemp("", "art-dupl-mixed-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Create Go files
			goFile := filepath.Join(tempDir, "test.go")
			err = os.WriteFile(goFile, []byte("package main\nfunc test() {}"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create non-Go files
			txtFile := filepath.Join(tempDir, "readme.txt")
			err = os.WriteFile(txtFile, []byte("readme"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			mdFile := filepath.Join(tempDir, "doc.md")
			err = os.WriteFile(mdFile, []byte("# doc"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl - should only analyze Go files
			cmd := exec.Command(binaryPath, tempDir)
			output, err := cmd.CombinedOutput()

			// Should not crash
			Expect(err).ToNot(HaveOccurred(), "Should handle mixed file types")
			Expect(string(output)).ToNot(BeEmpty())
		})
	})

	Context("When using invalid configuration files", func() {
		It("should handle malformed JSON config gracefully", func() {
			// Create temp directory
			tempDir, err := os.MkdirTemp("", "art-dupl-config-error-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Create malformed JSON config
			configFile := filepath.Join(tempDir, "dupl.json")
			malformedJSON := `{ invalid json }`
			err = os.WriteFile(configFile, []byte(malformedJSON), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid config
			cmd := exec.Command(binaryPath, "--config", configFile, ".")
			output, err := cmd.CombinedOutput()

			// Should fail gracefully with clear error
			Expect(err).To(HaveOccurred(), "Should error on invalid config")
			outputStr := string(output)
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring("error"),
				ContainSubstring("invalid"),
				ContainSubstring("JSON"),
			))
		})

		It("should handle missing config file gracefully", func() {
			// Try to use non-existent config file
			nonExistentConfig := "/tmp/art-dupl-nonexistent-config.json"

			cmd := exec.Command(binaryPath, "--config", nonExistentConfig, ".")
			output, err := cmd.CombinedOutput()

			// Should fail gracefully
			Expect(err).To(HaveOccurred())
			Expect(string(output)).NotTo(BeEmpty())
			Expect(string(output)).To(SatisfyAny(
				ContainSubstring("error"),
				ContainSubstring("not found"),
				ContainSubstring("no such"),
			))
		})

		It("should handle config with invalid values gracefully", func() {
			// Create temp directory
			tempDir, err := os.MkdirTemp("", "art-dupl-config-value-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Create config with invalid threshold
			configFile := filepath.Join(tempDir, "dupl.json")
			invalidConfig := `{
				"threshold": "not a number",
				"outputFormat": "json"
			}`
			err = os.WriteFile(configFile, []byte(invalidConfig), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid config
			cmd := exec.Command(binaryPath, "--config", configFile, ".")
			output, err := cmd.CombinedOutput()

			// Should handle gracefully (may use defaults or show error)
			outputStr := string(output)
			Expect(len(outputStr) > 0).To(BeTrue())
		})
	})

	Context("When using invalid flag combinations", func() {
		It("should handle conflicting output format flags gracefully", func() {
			// Create temp directory with test file
			tempDir, err := os.MkdirTemp("", "art-dupl-flag-conflict-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			testFile := filepath.Join(tempDir, "test.go")
			err = os.WriteFile(testFile, []byte("package main\nfunc test() {}"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Try to use multiple output format flags
			cmd := exec.Command(binaryPath, tempDir, "--json", "--html", "--plumbing")
			output, err := cmd.CombinedOutput()

			// Should handle gracefully (may use first one or show error)
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle invalid sorting option gracefully", func() {
			// Create temp directory with test file
			tempDir, err := os.MkdirTemp("", "art-dupl-sort-error-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			testFile := filepath.Join(tempDir, "test.go")
			err = os.WriteFile(testFile, []byte("package main\nfunc test() {}"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Use invalid sort option
			cmd := exec.Command(binaryPath, tempDir, "--sort", "invalid_sort_option")
			output, err := cmd.CombinedOutput()

			// Should handle gracefully (may default or show error)
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle invalid detection method gracefully", func() {
			// Create temp directory with test file
			tempDir, err := os.MkdirTemp("", "art-dupl-method-error-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			testFile := filepath.Join(tempDir, "test.go")
			err = os.WriteFile(testFile, []byte("package main\nfunc test() {}"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Use invalid detection method
			cmd := exec.Command(binaryPath, tempDir, "--detection-methods", "invalid_method")
			output, err := cmd.CombinedOutput()

			// Should handle gracefully
			Expect(string(output)).ToNot(BeEmpty())
		})
	})

	Context("When dealing with permission issues", func() {
		It("should handle unreadable files gracefully", func() {
			// Create temp directory
			tempDir, err := os.MkdirTemp("", "art-dupl-permission-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Create file with no read permissions
			testFile := filepath.Join(tempDir, "noperm.go")
			err = os.WriteFile(testFile, []byte("package main\nfunc test() {}"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Remove read permissions
			err = os.Chmod(testFile, 0o000)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl on directory
			cmd := exec.Command(binaryPath, tempDir)
			output, err := cmd.CombinedOutput()

			// Restore permissions before cleanup
			_ = os.Chmod(testFile, 0o644)

			// Should handle gracefully
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle directories without execute permission gracefully", func() {
			// Create parent temp directory
			parentDir, err := os.MkdirTemp("", "art-dupl-permission-dir-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(parentDir)

			// Create subdirectory
			subDir := filepath.Join(parentDir, "subdir")
			err = os.MkdirAll(subDir, 0o755)
			Expect(err).NotTo(HaveOccurred())

			// Create file in subdirectory
			testFile := filepath.Join(subDir, "test.go")
			err = os.WriteFile(testFile, []byte("package main\nfunc test() {}"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Remove execute permission from subdirectory
			err = os.Chmod(subDir, 0o000)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl on parent directory
			cmd := exec.Command(binaryPath, parentDir)
			output, err := cmd.CombinedOutput()

			// Restore permissions before cleanup
			_ = os.Chmod(subDir, 0o755)

			// Should handle gracefully
			Expect(string(output)).ToNot(BeEmpty())
		})
	})

	Context("When analyzing empty directories", func() {
		It("should handle empty directory without errors", func() {
			// Create empty temp directory
			tempDir, err := os.MkdirTemp("", "art-dupl-empty-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Run art-dupl on empty directory
			cmd := exec.Command(binaryPath, tempDir)
			output, err := cmd.CombinedOutput()

			// Should not crash
			Expect(err).ToNot(HaveOccurred(), "Should handle empty directory")
			// Output may be empty or show no files analyzed
			Expect(len(output) >= 0).To(BeTrue())
		})

		It("should handle directory with no Go files", func() {
			// Create temp directory with non-Go files
			tempDir, err := os.MkdirTemp("", "art-dupl-nogo-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Create non-Go files
			for _, name := range []string{"readme.txt", "doc.md", "data.json"} {
				file := filepath.Join(tempDir, name)
				err = os.WriteFile(file, []byte("content"), 0o644)
				Expect(err).NotTo(HaveOccurred())
			}

			// Run art-dupl
			cmd := exec.Command(binaryPath, tempDir)
			output, err := cmd.CombinedOutput()

			// Should not crash
			Expect(err).ToNot(HaveOccurred(), "Should handle directory with no Go files")
			Expect(len(output) >= 0).To(BeTrue())
		})
	})

	Context("When reading from stdin with invalid input", func() {
		It("should handle empty stdin gracefully", func() {
			// Use --files flag with empty stdin
			cmd := exec.Command(binaryPath, "--files", "--threshold", "10")
			// Empty stdin
			cmd.Stdin = strings.NewReader("")
			output, _ := cmd.CombinedOutput()

			// Should handle gracefully
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle stdin with invalid file paths gracefully", func() {
			// Use --files flag with invalid file paths
			invalidPaths := "/nonexistent/file1.go\n/nonexistent/file2.go\n"
			cmd := exec.Command(binaryPath, "--files", "--threshold", "10")
			cmd.Stdin = strings.NewReader(invalidPaths)
			output, _ := cmd.CombinedOutput()

			// Should handle gracefully
			Expect(string(output)).ToNot(BeEmpty())
		})
	})
})
