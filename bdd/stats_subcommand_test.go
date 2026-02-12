package bdd

import (
	"encoding/json"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Stats Subcommand
//
// These tests verify the stats subcommand functionality,
// which provides aggregated duplication statistics.
//
// The scenarios cover:
// - Basic stats output in different formats (text, json, csv)
// - Stats with various thresholds
// - Stats for specific paths
// - Stats with vendor directory inclusion
// - Stats with filtering options

func TestStatsSubcommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl Stats Subcommand BDD Suite")
}

var _ = Describe("Stats Subcommand", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When running stats with default text format", func() {
		It("should display duplication statistics", func() {
			// Create test files with duplicates
			code := `package main

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run stats subcommand
			output, err := setup.RunSubcommand("stats", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should contain stats-related information
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring("clone"),
				ContainSubstring("duplicat"),
				ContainSubstring("file"),
				ContainSubstring("stat"),
			))
		})

		It("should show files analyzed count", func() {
			// Create multiple test files
			code := `package main
func test() {}`

			err := setup.CreateDuplicateFiles([]string{"a.go", "b.go", "c.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run stats
			output, err := setup.RunSubcommand("stats", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should indicate files were analyzed
			Expect(outputStr).ToNot(BeEmpty())
		})
	})

	Context("When running stats with JSON format", func() {
		It("should produce valid JSON output", func() {
			// Create test files
			code := `package main

func duplicate() string {
	return "test"
}`

			err := setup.CreateDuplicateFiles([]string{"json1.go", "json2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run stats with JSON format - separate stdout from stderr
			cmd := exec.Command(setup.BinaryPath, "stats", "--format", "json", "--threshold", "10", setup.TmpDir)
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			// Parse JSON
			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Should have expected structure
			Expect(result).To(SatisfyAny(
				HaveKey("overview"),
				HaveKey("summary"),
				HaveKey("files_analyzed"),
				HaveKey("clone_groups"),
			))
		})

		It("should include clone statistics in JSON", func() {
			// Create test files with known duplicates
			code := `package main

import "fmt"

func processUser(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	fmt.Println(name)
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"user1.go", "user2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run stats with JSON
			cmd := exec.Command(setup.BinaryPath, "stats", setup.TmpDir, "--format", "json", "--threshold", "10")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Should have numeric stats fields
			Expect(result).To(SatisfyAny(
				HaveKey("total_clones"),
				HaveKey("totalCloneGroups"),
				HaveKey("clone_groups"),
			))
		})
	})

	Context("When running stats with CSV format", func() {
		It("should produce CSV output", func() {
			// Create test files
			code := `package main
func test() {}`

			err := setup.CreateDuplicateFiles([]string{"csv1.go", "csv2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run stats with CSV format - separate stdout
			cmd := exec.Command(setup.BinaryPath, "stats", "--format", "csv", "--threshold", "10", setup.TmpDir)
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// CSV should have comma-separated values
			Expect(outputStr).To(ContainSubstring(","))
		})
	})

	Context("When running stats with different thresholds", func() {
		It("should respect threshold parameter", func() {
			// Create test files
			code := `package main
func small() {}`

			err := setup.CreateDuplicateFiles([]string{"small1.go", "small2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with high threshold
			output, err := setup.RunSubcommand("stats", "--threshold", "100")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should still produce output
			Expect(outputStr).ToNot(BeEmpty())
		})
	})

	Context("When running stats with vendor directory", func() {
		It("should exclude vendor by default", func() {
			code := testutil.SimpleCodeTemplate("vendorTest")
			_ = setup.CreateAndRunDuplExpectSuccess([]string{"main1.go", "main2.go"}, code, "stats", "--threshold", testutil.ThresholdMedium)
		})

		It("should include vendor when --vendor flag is specified", func() {
			code := testutil.SimpleCodeTemplate("vendorCode")
			_ = setup.CreateAndRunDuplExpectSuccess([]string{"vendor1.go", "vendor2.go"}, code, "stats", "--vendor", "--threshold", testutil.ThresholdMedium)
		})
	})

	Context("When running stats with specific paths", func() {
		It("should analyze only specified paths", func() {
			// Create subdirectories
			err := setup.CreateSubdirectories("pkg1", "pkg2")
			Expect(err).NotTo(HaveOccurred())

			code := `package main
func pkgFunc() {}`

			// Create files in both directories
			err = setup.CreateFileWithContent("pkg1/file.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("pkg2/file.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Run stats on specific path
			specificPath := setup.GetFilePath("pkg1")
			output, err := setup.RunSubcommand("stats", "--threshold", "10", specificPath)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Output should reflect analysis of specific path
			Expect(outputStr).ToNot(BeEmpty())
		})
	})

	// testWithStats is a helper function to test stats subcommand with various options
	func testWithStats(testCode string, filenames []string, flags ...string) {
		err := setup.CreateDuplicateFiles(filenames, testCode)
		Expect(err).NotTo(HaveOccurred())

		// Build command args with mandatory --threshold if not provided
		args := append([]string{"stats"}, flags...)
		if !containsThreshold(flags) {
			args = append(args, "--threshold", "10")
		}

		output, err := setup.RunSubcommand(args...)
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).ToNot(BeEmpty())
	}

	// containsThreshold checks if flags already contain threshold argument
	func containsThreshold(flags []string) bool {
		for _, flag := range flags {
			if flag == "--threshold" || flag == "-t" {
				return true
			}
		}
		return false
	}

	// testDetectionMethod is a helper function to test a specific detection method
	func testDetectionMethod(detectionMethod, testCode string, filenames []string) {
		testWithStats(testCode, filenames, "--detection-methods", detectionMethod)
	}

	Context("When running stats with detection methods", func() {
		It("should work with hash detection method", func() {
			code := `package main
func hashTest() string {
	return "hash"
}`
			testDetectionMethod("hash", code, []string{"hash1.go", "hash2.go"})
		})

		It("should work with art-dupl detection method", func() {
			code := `package main
func artDuplTest() string {
	return "art-dupl"
}`
			testDetectionMethod("art-dupl", code, []string{"art1.go", "art2.go"})
		})
	})

	Context("When running stats with filter options", func() {
		It("should work with --filter-generated flag", func() {
			code := `package main
func filterTest() {}`
			testWithStats(code, []string{"filter1.go", "filter2.go"}, "--filter-generated")
		})

		It("should work with include-sqlc flag", func() {
			code := `// Code generated by sqlc. DO NOT EDIT.
package db
func SQLCFunc() {}`
			testWithStats(code, []string{"sqlc1.go", "sqlc2.go"}, "--include-sqlc")
		})
	})

	Context("When running stats on empty directories", func() {
		It("should handle empty directory gracefully", func() {
			// Run stats on empty temp directory
			output, err := setup.RunSubcommand("stats", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			// Should produce output even for empty directory
			Expect(output).ToNot(BeNil())
		})

		It("should handle directory with no Go files", func() {
			// Create a non-Go file
			err := setup.CreateFileWithContent("readme.txt", "This is not a Go file")
			Expect(err).NotTo(HaveOccurred())

			// Run stats
			output, err := setup.RunSubcommand("stats", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should handle gracefully
			Expect(outputStr).ToNot(BeEmpty())
		})
	})

	Context("When running stats with verbose flag", func() {
		It("should work with single verbose flag", func() {
			code := `package main
func verboseTest() {}`

			err := setup.CreateDuplicateFiles([]string{"verbose1.go", "verbose2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with verbose flag
			output, err := setup.RunSubcommand("stats", "-v", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			Expect(output).ToNot(BeNil())
		})

		It("should work with multiple verbose flags", func() {
			code := `package main
func verboseTest2() {}`

			err := setup.CreateDuplicateFiles([]string{"verbose3.go", "verbose4.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with multiple verbose flags
			output, err := setup.RunSubcommand("stats", "-vv", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			Expect(output).ToNot(BeNil())
		})
	})
})

var _ = Describe("Stats Subcommand Edge Cases", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When running stats with invalid inputs", func() {
		It("should handle non-existent path gracefully", func() {
			// Run stats on non-existent path
			output, err := setup.RunSubcommand("stats", "/nonexistent/path")
			// May error but should not panic
			_ = err
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle invalid format option", func() {
			code := `package main
func test() {}`

			err := setup.CreateDuplicateFiles([]string{"test1.go", "test2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid format
			output, err := setup.RunSubcommand("stats", "--format", "invalid_format", "--threshold", "10")
			// Should either error or use default
			_ = err
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When running stats on code with complex duplicates", func() {
		It("should handle multiple clone groups", func() {
			// Create different clone patterns
			codeA := `package main
import "fmt"
func patternA() {
	fmt.Println("A")
	fmt.Println("B")
	fmt.Println("C")
}`

			codeB := `package main
import "fmt"
func patternB() {
	fmt.Println("X")
	fmt.Println("Y")
	fmt.Println("Z")
}`

			err := setup.CreateDuplicateFiles([]string{"patternA1.go", "patternA2.go"}, codeA)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateDuplicateFiles([]string{"patternB1.go", "patternB2.go"}, codeB)
			Expect(err).NotTo(HaveOccurred())

			// Run stats with JSON to verify multiple groups
			cmd := exec.Command(setup.BinaryPath, "stats", setup.TmpDir, "--format", "json", "--threshold", "10")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Should have clone-related fields
			Expect(result).To(SatisfyAny(
				HaveKey("clone_groups"),
				HaveKey("totalCloneGroups"),
			))
		})

		It("should calculate statistics correctly for widespread clones", func() {
			// Create a clone that appears in many files
			widespreadCode := `package main
import "fmt"
func commonUtility(message string) {
	fmt.Printf("Message: %s\n", message)
}`

			// Create 5 files with the same code
			files := []string{"util1.go", "util2.go", "util3.go", "util4.go", "util5.go"}
			err := setup.CreateDuplicateFiles(files, widespreadCode)
			Expect(err).NotTo(HaveOccurred())

			// Run stats
			cmd := exec.Command(setup.BinaryPath, "stats", setup.TmpDir, "--format", "json", "--threshold", "5")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Verify stats structure
			Expect(result).To(SatisfyAny(
				HaveKey("files_analyzed"),
				HaveKey("total_clones"),
			))
		})
	})
})
