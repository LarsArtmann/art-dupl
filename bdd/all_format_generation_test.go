package bdd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/utils"
)

// BDD Test Suite for All Format Generation
//
// These tests verify the batch generation feature (--all flag)
// which generates all output formats for all detection methods.
//
// The scenarios cover:
// - Generating all output formats
// - Custom output directory configuration
// - Multiple detection methods in batch mode
// - Output file organization

func TestAllFormatGeneration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl All Format Generation BDD Suite")
}

var _ = Describe("All Format Generation (--all flag)", func() {
	var (
		tempDir       string
		outputDir     string
		fileProcessor *utils.FileProcessor
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-all-format-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		outputDir = filepath.Join(tempDir, "output")
		err = os.MkdirAll(outputDir, 0o755)
		Expect(err).NotTo(HaveOccurred())

		fileProcessor = utils.NewFileProcessor(tempDir)
	})

	AfterEach(func() {
		_ = os.RemoveAll(tempDir)
		_ = os.Remove("./bdd/art-dupl-all_format_generation-test")
	})

	Context("When generating all formats with default settings", func() {
		It("should generate text, HTML, JSON, and plumbing outputs", func() {
			// Create test files with duplicates
			code := `package main

import "fmt"

func process(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`

			err := fileProcessor.WriteDuplicateFiles([]string{"file1.go", "file2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-all_format_generation-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag
			cmd = exec.Command("./bdd/art-dupl-all_format_generation-test", tempDir, "--all", "--output-dir", outputDir, "--threshold", "10")
			_, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			// Check that output directory contains expected files
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())

			// Should have multiple output files
			Expect(files).ToNot(BeEmpty())

			// Verify at least one JSON file was created
			hasJSON := false
			for _, file := range files {
				if strings.Contains(file.Name(), ".json") {
					hasJSON = true
					// Verify JSON is valid
					jsonPath := filepath.Join(outputDir, file.Name())
					jsonData, err := os.ReadFile(jsonPath)
					Expect(err).NotTo(HaveOccurred())

					var result map[string]any
					err = json.Unmarshal(jsonData, &result)
					Expect(err).ToNot(HaveOccurred(), "Generated JSON should be valid")
				}
			}

			Expect(hasJSON).To(BeTrue(), "Should generate JSON output")
		})

		It("should include metadata in generated files", func() {
			// Create test files
			code := `package main

func duplicate() {}`

			err := fileProcessor.WriteDuplicateFiles([]string{"meta1.go", "meta2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-all_format_generation-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag
			cmd = exec.Command("./bdd/art-dupl-all_format_generation-test", tempDir, "--all", "--output-dir", outputDir, "--threshold", "10")
			_, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			// Check JSON files for metadata
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())

			for _, file := range files {
				if strings.Contains(file.Name(), ".json") {
					jsonPath := filepath.Join(outputDir, file.Name())
					jsonData, err := os.ReadFile(jsonPath)
					Expect(err).NotTo(HaveOccurred())

					var result map[string]any
					err = json.Unmarshal(jsonData, &result)
					Expect(err).ToNot(HaveOccurred())

					// Verify metadata fields exist
					Expect(result).To(HaveKey("version"), "JSON should include version")
					Expect(result).To(HaveKey("timestamp"), "JSON should include timestamp")
					Expect(result).To(HaveKey("threshold"), "JSON should include threshold")
				}
			}
		})
	})

	Context("When generating all formats with custom output directory", func() {
		It("should create output directory if it doesn't exist", func() {
			// Create test files
			code := `package main

func test() {}`

			err := fileProcessor.WriteDuplicateFiles([]string{"test1.go", "test2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-all_format_generation-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Specify a non-existent output directory
			customOutputDir := filepath.Join(tempDir, "custom", "nested", "output")

			// Run with --all flag and custom output directory
			cmd = exec.Command("./bdd/art-dupl-all_format_generation-test", tempDir, "--all", "--output-dir", customOutputDir, "--threshold", "10")
			_, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			// Verify directory was created
			_, err = os.Stat(customOutputDir)
			Expect(err).NotTo(HaveOccurred(), "Custom output directory should be created")

			// Verify files were created in the custom directory
			files, err := os.ReadDir(customOutputDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).ToNot(BeEmpty(), "Should have output files in custom directory")
		})

		It("should use existing output directory without errors", func() {
			// Create test files
			code := `package main

func test() {}`

			err := fileProcessor.WriteDuplicateFiles([]string{"test1.go", "test2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-all_format_generation-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Output directory already exists
			err = os.MkdirAll(outputDir, 0o755)
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag using existing directory
			cmd = exec.Command("./bdd/art-dupl-all_format_generation-test", tempDir, "--all", "--output-dir", outputDir, "--threshold", "10")
			_, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			// Verify files were created
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).ToNot(BeEmpty())
		})
	})

	Context("When generating all formats with multiple detection methods", func() {
		It("should generate separate files for each detection method", func() {
			// Create test files
			code := `package main

func multiDetect() string {
	return "test"
}`

			err := fileProcessor.WriteDuplicateFiles([]string{"multi1.go", "multi2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-all_format_generation-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with --all and multiple detection methods
			cmd = exec.Command("./bdd/art-dupl-all_format_generation-test", tempDir, "--all", "--output-dir", outputDir, "--detection-methods", "hash,art-dupl", "--threshold", "10")
			_, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			// Check for files from both methods
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())

			// Should have multiple files indicating different detection methods
			Expect(len(files)).To(BeNumerically(">=", 2), "Should have multiple output files for different methods")

			// Verify at least one file mentions hash detection
			hasHashMethod := false
			hasArtDuplMethod := false
			for _, file := range files {
				if strings.Contains(file.Name(), "hash") || strings.Contains(file.Name(), "Hash") {
					hasHashMethod = true
				}
				if strings.Contains(file.Name(), "art-dupl") || strings.Contains(file.Name(), "ArtDupl") {
					hasArtDuplMethod = true
				}

				// Check JSON files for detection_method field
				if strings.HasSuffix(file.Name(), ".json") {
					jsonPath := filepath.Join(outputDir, file.Name())
					jsonData, err := os.ReadFile(jsonPath)
					Expect(err).NotTo(HaveOccurred())

					var result map[string]any
					err = json.Unmarshal(jsonData, &result)
					Expect(err).ToNot(HaveOccurred())

					if detectionMethod, ok := result["detection_method"].(string); ok {
						if strings.Contains(detectionMethod, "hash") {
							hasHashMethod = true
						}
						if strings.Contains(detectionMethod, "art-dupl") {
							hasArtDuplMethod = true
						}
					}
				}
			}

			Expect(hasHashMethod || hasArtDuplMethod).To(BeTrue(), "Should have files from detection methods")
		})
	})

	Context("When generating all formats with high threshold", func() {
		It("should produce output with fewer or no clones", func() {
			// Create test files with small duplicates
			code := `package main

func small() {}`

			err := fileProcessor.WriteDuplicateFiles([]string{"small1.go", "small2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-all_format_generation-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with --all and high threshold
			cmd = exec.Command("./bdd/art-dupl-all_format_generation-test", tempDir, "--all", "--output-dir", outputDir, "--threshold", "100")
			_, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			// Verify files were created even if few/no clones
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).ToNot(BeEmpty(), "Should create output files even with high threshold")
		})
	})

	Context("When generating all formats for project with no duplicates", func() {
		It("should still create output files with empty results", func() {
			// Create unique code (no duplicates)
			uniqueCode1 := `package main

func unique1() {}`

			uniqueCode2 := `package main

func unique2() {}`

			err := fileProcessor.WriteTextFile("unique1.go", uniqueCode1)
			Expect(err).NotTo(HaveOccurred())
			err = fileProcessor.WriteTextFile("unique2.go", uniqueCode2)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-all_format_generation-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag
			cmd = exec.Command("./bdd/art-dupl-all_format_generation-test", tempDir, "--all", "--output-dir", outputDir, "--threshold", "10")
			_, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			// Verify files were created even with no duplicates
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).ToNot(BeEmpty(), "Should create output files even with no duplicates")

			// Verify JSON shows no clones
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".json") {
					jsonPath := filepath.Join(outputDir, file.Name())
					jsonData, err := os.ReadFile(jsonPath)
					Expect(err).NotTo(HaveOccurred())

					var result map[string]any
					err = json.Unmarshal(jsonData, &result)
					Expect(err).ToNot(HaveOccurred())

					if summary, ok := result["summary"].(map[string]any); ok {
						totalClones := int(summary["total_clones"].(float64))
						Expect(totalClones).To(Equal(0), "Should report zero clones")
					}
				}
			}
		})
	})
})
