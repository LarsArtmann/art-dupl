package bdd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

var _ = Describe("All Format Generation (--all flag)", func() {
	var (
		setup     *testutil.BDDTestSetup
		outputDir string
	)

	BeforeEach(func() {
		var err error

		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())

		outputDir = filepath.Join(setup.TmpDir, "output")
		err = os.MkdirAll(outputDir, 0o755)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
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

			err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag
			_, err = setup.RunArtDuplAllFormat(outputDir, "10")
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

			err := setup.CreateDuplicateFiles([]string{"meta1.go", "meta2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag
			_, err = setup.RunArtDuplAllFormat(outputDir, "10")
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

			err := setup.CreateDuplicateFiles([]string{"test1.go", "test2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Specify a non-existent output directory
			customOutputDir := filepath.Join(setup.TmpDir, "custom", "nested", "output")

			// Run with --all flag and custom output directory
			_, err = setup.RunArtDuplOnDir(
				setup.TmpDir,
				"--all",
				"--output-dir",
				customOutputDir,
				"--threshold",
				"10",
			)
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

			err := setup.CreateDuplicateFiles([]string{"test1.go", "test2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Output directory already exists
			err = os.MkdirAll(outputDir, 0o755)
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag using existing directory
			_, err = setup.RunArtDuplOnDir(
				setup.TmpDir,
				"--all",
				"--output-dir",
				outputDir,
				"--threshold",
				"10",
			)
			Expect(err).ToNot(HaveOccurred())

			// Verify files were created
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).ToNot(BeEmpty())
		})
	})

	Context("When generating all formats with multiple detection methods", func() {
		It("should generate reports with combined detection methods", func() {
			// Create test files
			code := `package main

func multiDetect() string {
	return "test"
}`

			err := setup.CreateDuplicateFiles([]string{"multi1.go", "multi2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag (uses all detection methods by default)
			_, err = setup.RunArtDuplOnDir(
				setup.TmpDir,
				"--all",
				"--output-dir",
				outputDir,
				"--threshold",
				"10",
			)
			Expect(err).ToNot(HaveOccurred())

			// Check for generated files
			files, err := os.ReadDir(outputDir)
			Expect(err).NotTo(HaveOccurred())

			// Should have multiple output files (one per format)
			Expect(
				len(files),
			).To(BeNumerically(">=", 2), "Should have multiple output files for different formats")

			// Verify JSON file contains detection_method field with combined methods
			foundJSON := false

			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".json") {
					foundJSON = true
					jsonPath := filepath.Join(outputDir, file.Name())
					jsonData, err := os.ReadFile(jsonPath)
					Expect(err).NotTo(HaveOccurred())

					var result map[string]any

					err = json.Unmarshal(jsonData, &result)
					Expect(err).ToNot(HaveOccurred())

					// Should have detection_method field indicating combined methods
					Expect(result).To(HaveKey("detection_method"))
					detectionMethod := result["detection_method"].(string)
					// Should contain both methods or "all" indicator
					Expect(detectionMethod).To(SatisfyAny(
						ContainSubstring("hash"),
						ContainSubstring("art-dupl"),
						Equal("all"),
					))
				}
			}

			Expect(foundJSON).To(BeTrue(), "Should generate JSON output file")
		})
	})

	Context("When generating all formats with high threshold", func() {
		It("should produce output with fewer or no clones", func() {
			// Create test files with small duplicates
			code := `package main

func small() {}`

			err := setup.CreateDuplicateFiles([]string{"small1.go", "small2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with --all and high threshold
			_, err = setup.RunArtDuplOnDir(
				setup.TmpDir,
				"--all",
				"--output-dir",
				outputDir,
				"--threshold",
				"100",
			)
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

			err := setup.CreateTestFile("unique1.go", uniqueCode1)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("unique2.go", uniqueCode2)
			Expect(err).NotTo(HaveOccurred())

			// Run with --all flag
			_, err = setup.RunArtDuplAllFormat(outputDir, "10")
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
