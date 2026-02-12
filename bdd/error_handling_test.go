package bdd

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
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

// createTempTestFile is a helper that creates a temp directory with a test file
// and returns the temp directory path. The caller is responsible for cleanup.
func createTempTestFile(pattern string) string {
	tempDir, err := os.MkdirTemp("", pattern)
	Expect(err).NotTo(HaveOccurred())

	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\nfunc test() {}"), 0o644)
	Expect(err).NotTo(HaveOccurred())

	return tempDir
}

// runWithFlagsAndCheckOutput is a helper that runs art-dupl with given flags
// and verifies that it produces non-empty output without crashing.
func runWithFlagsAndCheckOutput(setup *testutil.BDDTestSetup, path string, flags ...string) {
	output, _ := setup.RunArtDuplOnDir(path, flags...)
	Expect(string(output)).ToNot(BeEmpty(), "Should handle gracefully and produce output")
}

var _ = Describe("Error Handling", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When analyzing non-existent paths", func() {
		It("should handle missing directory gracefully", func() {
			// Try to analyze non-existent directory
			nonExistentPath := "/tmp/art-dupl-test-nonexistent-xyz123"

			output, err := setup.RunArtDupl(nonExistentPath)

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

			output, err := setup.RunArtDupl(nonExistentFile)

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
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Create non-Go files
			nonGoFile := filepath.Join(tempDir, "test.txt")
			err = os.WriteFile(nonGoFile, []byte("not a go file"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl on directory with only non-Go files
			output, err := setup.RunArtDupl(tempDir)

			// Should not crash
			Expect(err).ToNot(HaveOccurred(), "Should handle non-Go files without error")
			// Output may be empty or show no results
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle mixed file types", func() {
			// Create temporary directory with mixed file types
			tempDir, err := os.MkdirTemp("", "art-dupl-mixed-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

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
			output, err := setup.RunArtDupl(tempDir)

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
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Create malformed JSON config
			configFile := filepath.Join(tempDir, "dupl.json")
			malformedJSON := `{ invalid json }`
			err = os.WriteFile(configFile, []byte(malformedJSON), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid config
			output, err := setup.RunArtDupl("--config", configFile, ".")
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

			output, err := setup.RunArtDupl("--config", nonExistentConfig, ".")
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
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Create config with invalid threshold
			configFile := filepath.Join(tempDir, "dupl.json")
			invalidConfig := `{
				"threshold": "not a number",
				"outputFormat": "json"
			}`
			err = os.WriteFile(configFile, []byte(invalidConfig), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid config
			output, err := setup.RunArtDupl("--config", configFile, ".")
			// Should handle gracefully (may use defaults or show error)
			outputStr := string(output)
			Expect(outputStr).ToNot(BeEmpty())
		})
	})

	Context("When using invalid flag combinations", func() {
		It("should handle conflicting output format flags gracefully", func() {
			tempDir := createTempTestFile("art-dupl-flag-conflict-bdd-*")
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Try to use multiple output format flags
			runWithFlagsAndCheckOutput(setup, tempDir, "--json", "--html", "--plumbing")
		})

		It("should handle invalid sorting option gracefully", func() {
			tempDir := createTempTestFile("art-dupl-sort-error-bdd-*")
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Use invalid sort option
			runWithFlagsAndCheckOutput(setup, tempDir, "--sort", "invalid_sort_option")
		})

		It("should handle invalid detection method gracefully", func() {
			tempDir := createTempTestFile("art-dupl-method-error-bdd-*")
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Use invalid detection method
			runWithFlagsAndCheckOutput(setup, tempDir, "--detection-methods", "invalid_method")
		})
	})

	Context("When dealing with permission issues", func() {
		It("should handle unreadable files gracefully", func() {
			tempDir := createTempTestFile("art-dupl-permission-bdd-*")
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Get the test file path
			testFile := filepath.Join(tempDir, "test.go")

			// Remove read permissions
			err := os.Chmod(testFile, 0o000)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl on directory
			output, err := setup.RunArtDupl(tempDir)
			// Restore permissions before cleanup
			_ = os.Chmod(testFile, 0o644)

			// Should handle gracefully
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle directories without execute permission gracefully", func() {
			// Create parent temp directory
			parentDir, err := os.MkdirTemp("", "art-dupl-permission-dir-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(parentDir) }() // test cleanup

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
			output, err := setup.RunArtDupl(parentDir)
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
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Run art-dupl on empty directory
			output, err := setup.RunArtDupl(tempDir)
			// Should not crash
			Expect(err).ToNot(HaveOccurred(), "Should handle empty directory")
			// Output may be empty or show no files analyzed
			Expect(len(output)).To(BeNumerically(">=", 0))
		})

		It("should handle directory with no Go files", func() {
			// Create temp directory with non-Go files
			tempDir, err := os.MkdirTemp("", "art-dupl-nogo-bdd-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tempDir) }() // test cleanup

			// Create non-Go files
			for _, name := range []string{"readme.txt", "doc.md", "data.json"} {
				file := filepath.Join(tempDir, name)
				err = os.WriteFile(file, []byte("content"), 0o644)
				Expect(err).NotTo(HaveOccurred())
			}

			// Run art-dupl
			output, err := setup.RunArtDupl(tempDir)
			// Should not crash
			Expect(err).ToNot(HaveOccurred(), "Should handle directory with no Go files")
			Expect(len(output)).To(BeNumerically(">=", 0))
		})
	})

	Context("When reading from stdin with invalid input", func() {
		It("should handle empty stdin gracefully", func() {
			// Use --files flag with empty stdin
			output, err := setup.RunArtDuplWithStdin("", map[string]string{
				"threshold": "10",
			})

			// Should handle gracefully
			Expect(err).ToNot(HaveOccurred(), "Should handle empty stdin")
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle stdin with invalid file paths gracefully", func() {
			// Use --files flag with invalid file paths
			invalidPaths := "/nonexistent/file1.go\n/nonexistent/file2.go\n"
			output, err := setup.RunArtDuplWithStdin(invalidPaths, map[string]string{
				"threshold": "10",
			})

			// Should handle gracefully
			Expect(err).ToNot(HaveOccurred(), "Should handle invalid file paths")
			Expect(string(output)).ToNot(BeEmpty())
		})
	})
})
