package bdd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("Stats Subcommand", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
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

			err := setup.CreateDuplicateFiles([]string{goldenFile1, goldenFile2}, code)
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
			err := setup.CreateDuplicateFiles([]string{"a.go", "b.go", "c.go"}, simpleTestCode)
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

			// Run stats with JSON format
			result, err := setup.RunStatsSubcommandWithJSON("10")
			Expect(err).ToNot(HaveOccurred())

			// Should have expected structure
			Expect(result).To(SatisfyAny(
				HaveKey("overview"),
				HaveKey("summary"),
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
			result, err := setup.RunStatsSubcommandWithJSON("10")
			Expect(err).ToNot(HaveOccurred())

			// Verify overview contains clone-related fields
			Expect(result).To(HaveKey("overview"))
			statsOverview := result["overview"].(map[string]any)
			Expect(statsOverview).To(SatisfyAll(
				HaveKey("totalClones"),
				HaveKey("cloneGroups"),
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

			// Run stats with CSV format
			output, err := setup.RunSubcommand("stats", "--format", "csv", "--threshold", "10")
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

			err := setup.CreateDuplicateFiles([]string{goldenSmallFile1, smallFile2}, code)
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
			err := setup.CreateDuplicateFiles([]string{"main1.go", "main2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			_, err = setup.RunSubcommand("stats", "--threshold", testutil.ThresholdMedium)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should include vendor when --vendor flag is specified", func() {
			code := testutil.SimpleCodeTemplate("vendorCode")
			err := setup.CreateDuplicateFiles([]string{"vendor1.go", "vendor2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			_, err = setup.RunSubcommand(
				"stats",
				"--vendor",
				"--threshold",
				testutil.ThresholdMedium,
			)
			Expect(err).NotTo(HaveOccurred())
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

	Context("When running stats with detection methods", func() {
		DescribeTable(
			"should work with detection method",
			func(method, code string, filenames []string) {
				testDetectionMethod(setup, method, code, filenames)
			},
			Entry(
				"hash detection method",
				"hash",
				`package main
func hashTest() string {
	return "hash"
}`,
				[]string{"hash1.go", "hash2.go"},
			),
			Entry(
				"art-dupl detection method",
				"art-dupl",
				`package main
func artDuplTest() string {
	return "art-dupl"
}`,
				[]string{"art1.go", "art2.go"},
			),
		)
	})

	Context("When running stats with filter options", func() {
		DescribeTable(
			"should work with filter flag",
			func(testCode string, filenames []string, flag string) {
				testWithFilterFlag(setup, testCode, filenames, flag)
			},
			Entry(
				"include-sqlc",
				"// Code generated by sqlc. DO NOT EDIT.\npackage db\nfunc SQLCFunc() {}",
				[]string{"sqlc1.go", "sqlc2.go"},
				"--include-sqlc",
			),
		)
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
		DescribeTable(
			"should work with various verbose flag formats",
			func(funcName string, files []string, flags ...string) {
				code := fmt.Sprintf(`package main
func %s() {}`, funcName)
				err := setup.CreateDuplicateFiles(files, code)
				Expect(err).NotTo(HaveOccurred())

				args := append([]string{"stats"}, flags...)
				output, err := setup.RunSubcommand(args...)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())
			},
			Entry(
				"single verbose flag",
				"verboseTest",
				[]string{"verbose1.go", "verbose2.go"},
				"-v",
				"--threshold",
				"10",
			),
			Entry(
				"multiple verbose flags",
				"verboseTest2",
				[]string{"verbose3.go", "verbose4.go"},
				"-vv",
				"--threshold",
				"10",
			),
		)
	})
})

// testWithStats is a helper function to test stats subcommand with various options.
func testWithStats(
	setup *testutil.BDDTestSetup,
	testCode string,
	filenames []string,
	flags ...string,
) {
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

// containsThreshold checks if flags already contain threshold argument.
func containsThreshold(flags []string) bool {
	for _, flag := range flags {
		if flag == "--threshold" || flag == "-t" {
			return true
		}
	}

	return false
}

// testDetectionMethod is a helper function to test a specific detection method.
func testDetectionMethod(
	setup *testutil.BDDTestSetup,
	detectionMethod, testCode string,
	filenames []string,
) {
	testWithStats(setup, testCode, filenames, "--detection-methods", detectionMethod)
}

// testWithFilterFlag is a helper function to test filter flag functionality.
func testWithFilterFlag(
	setup *testutil.BDDTestSetup,
	testCode string,
	filenames []string,
	flag string,
) {
	testWithStats(setup, testCode, filenames, flag)
}

var _ = Describe("Stats Subcommand Edge Cases", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
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

			err := setup.CreateDuplicateFiles([]string{goldenTestFile1, "test2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid format
			output, err := setup.RunSubcommand(
				"stats",
				"--format",
				"invalid_format",
				"--threshold",
				"10",
			)
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
			result, err := setup.RunStatsSubcommandWithJSON("10")
			Expect(err).ToNot(HaveOccurred())

			// Verify overview contains clone-related fields
			Expect(result).To(HaveKey("overview"))
			overview := result["overview"].(map[string]any)
			Expect(overview).To(HaveKey("cloneGroups"))
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
			result, err := setup.RunStatsSubcommandWithJSON("5")
			Expect(err).ToNot(HaveOccurred())

			// Verify overview contains clone-related fields
			Expect(result).To(HaveKey("overview"))
			summaryData := result["overview"].(map[string]any)
			Expect(summaryData).To(SatisfyAll(
				HaveKey("filesScanned"),
				HaveKey("totalClones"),
			))
		})
	})
})
