package bdd

import (
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testCodeSamples contains code samples used in plumbing tests.
var (
	duplicateTestCode = "package main\n" + testutil.DuplicateFuncSource("duplicate")
	commonTestCode    = "package main\n" + testutil.DuplicateFuncSource("common")
)

// BDD Test Suite for Plumbing Output and Multiple Paths
//
// These tests verify:
// - Plumbing output format (machine-readable)
// - Multiple path arguments
// - Edge cases with paths

// expectValidPlumbingOutput validates that each non-empty line in the output
// follows the plumbing format: filename.go:startline-endline.
func expectValidPlumbingOutput(output []byte) {
	lines := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for line := range lines {
		if line == "" {
			continue
		}
		// Should contain filename and line numbers
		Expect(line).To(MatchRegexp(`\.go:\d+-\d+$`))
	}
}

var _ = Describe("Plumbing Output Format", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Context("When using plumbing output", func() {
		It("should produce machine-readable output", func() {
			duplicateCode := "package main\n" + testutil.DuplicateFuncSource("common")

			err := setup.CreateDuplicateFiles([]string{goldenFile1, goldenFile2}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing output
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			// Each line should follow plumbing format: filename:startline-endline
			expectValidPlumbingOutput(output)
		})

		It("should be parseable by shell scripts", func() {
			duplicateCode := duplicateTestCode

			err := setup.CreateDuplicateFiles(
				[]string{"pkg/file1.go", "pkg/file2.go"},
				duplicateCode,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--plumbing", "--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Each line should be parseable with cut/awk
			lines := strings.SplitSeq(strings.TrimSpace(outputStr), "\n")
			for line := range lines {
				if line == "" {
					continue
				}
				// Format: path/to/file.go:start-end
				parts := strings.Split(line, ":")
				Expect(parts).To(HaveLen(2))
				Expect(parts[0]).To(MatchRegexp(`\.go$`))
				Expect(parts[1]).To(MatchRegexp(`\d+-\d+$`))
			}
		})

		It("should not include headers or formatting", func() {
			duplicateCode := duplicateTestCode

			err := setup.CreateDuplicateFiles([]string{goldenFile1, goldenFile2}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--plumbing", "--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should not contain HTML or text formatting
			Expect(outputStr).ToNot(ContainSubstring("<!DOCTYPE"))
			Expect(outputStr).ToNot(ContainSubstring("found"))
			Expect(outputStr).ToNot(ContainSubstring("clone"))
		})
	})

	Context("When combining plumbing with other options", func() {
		It("should respect threshold in plumbing output", func() {
			smallCode := `package main
func small() { println(1) }`
			largeCode := `package main

import "fmt"

func large() {
	count := 0
	for i := 0; i < 100; i++ {
		fmt.Println(i)
		count++
	}
	fmt.Println("total:", count)
}`

			// Create small duplicates
			err := setup.CreateDuplicateFiles([]string{goldenSmallFile1, smallFile2}, smallCode)
			Expect(err).NotTo(HaveOccurred())
			// Create large duplicates
			err = setup.CreateDuplicateFiles([]string{goldenLargeFile1, largeFile2}, largeCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with threshold that filters small but shows large
			// small: 1 statement, large: 3 statements
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "2")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should only show large files
			Expect(outputStr).To(ContainSubstring("large"))
			Expect(outputStr).ToNot(ContainSubstring("small"))
		})

		It("should work with sorting options", func() {
			duplicateCode := duplicateTestCode

			err := setup.CreateDuplicateFiles([]string{goldenFile1, goldenFile2}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing and sort
			output, err := setup.RunArtDupl("--plumbing", "--sort", "size", "--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			// Should still be valid plumbing format
			expectValidPlumbingOutput(output)
		})
	})
})

var _ = Describe("Multiple Path Arguments", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Context("When analyzing multiple directories", func() {
		It("should find duplicates across all paths", func() {
			// Create subdirectories
			err := setup.CreateSubdirectories("pkg1", "pkg2", "pkg3")
			Expect(err).NotTo(HaveOccurred())

			duplicateCode := "package main\n" + testutil.DuplicateFuncSource("common")

			// Create duplicates in different directories
			err = setup.CreateFileWithContent("pkg1/file1.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("pkg2/file2.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("pkg3/file3.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with multiple paths
			output, err := setup.RunArtDupl(
				setup.GetFilePath("pkg1"),
				setup.GetFilePath("pkg2"),
				setup.GetFilePath("pkg3"),
				"--threshold", "1",
			)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should find files from all paths
			Expect(outputStr).To(ContainSubstring("pkg1"))
			Expect(outputStr).To(ContainSubstring("pkg2"))
			Expect(outputStr).To(ContainSubstring("pkg3"))
		})

		It("should handle mix of directories and single files", func() {
			err := setup.CreateSubdirectories("pkg")
			Expect(err).NotTo(HaveOccurred())

			duplicateCode := commonTestCode

			// Create file in directory
			err = setup.CreateFileWithContent("pkg/file1.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			// Create standalone file
			err = setup.CreateTestFile("standalone.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with mixed paths
			output, err := setup.RunArtDupl(
				setup.GetFilePath("pkg"),
				setup.GetFilePath("standalone.go"),
				"--threshold", "1",
			)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should find both
			Expect(outputStr).To(ContainSubstring("pkg"))
			Expect(outputStr).To(ContainSubstring("standalone.go"))
		})

		It("should not find duplicates within excluded paths", func() {
			err := setup.CreateSubdirectories("include", "exclude")
			Expect(err).NotTo(HaveOccurred())

			duplicateCode := commonTestCode

			// Create duplicates in both directories
			err = setup.CreateFileWithContent("include/file1.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("include/file2.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("exclude/file3.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("exclude/file4.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run only on include directory (use RunArtDuplOnDir to avoid double-path issue)
			output, err := setup.RunArtDuplOnDir(
				setup.GetFilePath("include"),
				"--threshold", "1",
			)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should only show include directory
			Expect(outputStr).To(ContainSubstring("include"))
			Expect(outputStr).ToNot(ContainSubstring("exclude"))
		})
	})

	Context("When using exclude patterns with multiple paths", func() {
		It("should exclude patterns across all paths", func() {
			err := setup.CreateSubdirectories("pkg1", "pkg2")
			Expect(err).NotTo(HaveOccurred())

			duplicateCode := commonTestCode
			testCode := "package main\n" + testutil.DuplicateFuncSource("testCommon")

			// Create regular files
			err = setup.CreateFileWithContent("pkg1/file.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("pkg2/file.go", duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			// Create test files
			err = setup.CreateFileWithContent("pkg1/file_test.go", testCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("pkg2/file_test.go", testCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with exclude pattern
			output, err := setup.RunArtDupl(
				setup.GetFilePath("pkg1"),
				setup.GetFilePath("pkg2"),
				"--exclude-pattern", "*_test.go",
				"--threshold", "1",
			)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should show regular files but not test files
			Expect(outputStr).To(ContainSubstring("file.go"))
			Expect(outputStr).ToNot(ContainSubstring("_test.go"))
		})
	})
})

var _ = Describe("Path Edge Cases", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Context("When handling special path scenarios", func() {
		// Helper function for testing duplicate detection in subdirectories
		testSubdirectoryDuplicates := func(subDir, funcName, runPath, expectedSubstr string) {
			err := setup.CreateSubdirectories(subDir)
			Expect(err).NotTo(HaveOccurred())

			duplicateCode := "package main\n" + testutil.DuplicateFuncSource(funcName)

			err = setup.CreateFileWithContent(filepath.Join(subDir, goldenFile1), duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent(filepath.Join(subDir, goldenFile2), duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl(setup.GetFilePath(runPath), "--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			Expect(string(output)).To(ContainSubstring(expectedSubstr))
		}

		It("should handle nested directories correctly", func() {
			testSubdirectoryDuplicates("a/b/c/d", "deep", "a", "a/b/c/d")
		})

		It("should handle paths with special characters", func() {
			testSubdirectoryDuplicates("my-pkg", "hyphen", "my-pkg", "my-pkg")
		})

		// Helper function for testing duplicate detection in root directory
		testRootDuplicates := func(runPath string) {
			duplicateCode := "package main\n" + testutil.DuplicateFuncSource("root")

			err := setup.CreateDuplicateFiles([]string{goldenFile1, goldenFile2}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl(runPath, "--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			Expect(string(output)).To(ContainSubstring(goldenFile1))
		}

		It("should handle current directory", func() {
			testRootDuplicates(".")
		})

		It("should handle absolute paths", func() {
			testRootDuplicates(setup.TmpDir)
		})
	})

	Context("When no Go files exist in path", func() {
		It("should handle empty directory gracefully", func() {
			testutil.TestEmptyDirectory(setup, 5)
		})

		It("should handle directory with non-Go files", func() {
			// Create some non-Go files
			err := setup.CreateTestFile("README.md", "# Project")
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("config.json", `{}`)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			// Should complete without error
			Expect(output).ToNot(BeNil())
		})
	})
})
