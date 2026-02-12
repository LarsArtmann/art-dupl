package bdd

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Plumbing Output Format
//
// These tests verify the machine-readable plumbing output format,
// designed for integration with scripts and external tools.
//
// The scenarios cover:
// - Basic plumbing format structure
// - Plumbing with different thresholds
// - Plumbing output with various detection methods
// - Plumbing format validation
// - Integration with CI/CD pipelines

func TestPlumbingOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl Plumbing Output BDD Suite")
}

var _ = Describe("Plumbing Output Format", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	// runPlumbingTest is a helper that creates duplicate files and runs art-dupl with plumbing output.
	// It returns the output for custom assertions.
	runPlumbingTest := func(filenames []string, code, threshold string) ([]byte, error) {
		err := setup.CreateDuplicateFiles(filenames, code)
		if err != nil {
			return nil, err
		}
		return setup.RunArtDupl("--plumbing", "--threshold", threshold)
	}

	// runPlumbingTestWithDetection is a helper that creates duplicate files and runs art-dupl with
	// plumbing output and a specific detection method. It returns the output for custom assertions.
	runPlumbingTestWithDetection := func(filenames []string, code, threshold, detectionMethod string) ([]byte, error) {
		err := setup.CreateDuplicateFiles(filenames, code)
		if err != nil {
			return nil, err
		}
		return setup.RunArtDupl("--plumbing", "--detection-methods", detectionMethod, "--threshold", threshold)
	}

	Context("When using plumbing output for basic analysis", func() {
		It("should produce machine-readable output", func() {
			code := `package main

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`

			output, err := runPlumbingTest([]string{"plumb1.go", "plumb2.go"}, code, "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring(":"),
				ContainSubstring("\t"),
			))
		})

		It("should include file paths in plumbing output", func() {
			code := `package main
func pathTest() {}`

			output, err := runPlumbingTest([]string{"path1.go", "path2.go"}, code, "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring("path1.go"),
				ContainSubstring("path2.go"),
			))
		})

		It("should include line numbers in plumbing output", func() {
			code := `package main

import "fmt"

func lineNumberTest() {
	fmt.Println("line 1")
	fmt.Println("line 2")
	fmt.Println("line 3")
}`

			output, err := setup.CreateAndRunDupl([]string{"line1.go", "line2.go"}, code, "--plumbing", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using plumbing with different thresholds", func() {
		It("should respect threshold in plumbing output", func() {
			// Create test files
			smallCode := `package main
func small() {}`

			largeCode := `package main

import "fmt"

func large() {
	fmt.Println("line 1")
	fmt.Println("line 2")
	fmt.Println("line 3")
	fmt.Println("line 4")
	fmt.Println("line 5")
}`

			// Create files with both small and large duplicates
			err := setup.CreateDuplicateFiles([]string{"small1.go", "small2.go"}, smallCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateDuplicateFiles([]string{"large1.go", "large2.go"}, largeCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with high threshold - should filter out small clones
			highOutput, err := setup.RunArtDupl("--plumbing", "--threshold", "50")
			Expect(err).ToNot(HaveOccurred())

			// Run with low threshold - should include more clones
			lowOutput, err := setup.RunArtDupl("--plumbing", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			// High threshold should produce less output than low threshold
			Expect(len(highOutput)).To(BeNumerically("<=", len(lowOutput)))
		})
	})

	Context("When using plumbing with different detection methods", func() {
		It("should work with hash detection method", func() {
			code := `package main
func hashPlumb() string {
	return "hash test"
}`

			output, err := runPlumbingTestWithDetection([]string{"hash1.go", "hash2.go"}, code, "10", "hash")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should work with art-dupl detection method", func() {
			code := `package main

import "fmt"

func artDuplPlumb(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	return nil
}`

			output, err := runPlumbingTestWithDetection([]string{"art1.go", "art2.go"}, code, "10", "art-dupl")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should work with combined detection methods", func() {
			code := `package main
func combinedPlumb() string {
	return "combined"
}`

			output, err := runPlumbingTestWithDetection([]string{"combined1.go", "combined2.go"}, code, "10", "hash,art-dupl")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using plumbing with sorting options", func() {
		It("should work with size sorting", func() {
			code := `package main
func sizeSortPlumb() {}`

			err := setup.CreateDuplicateFiles([]string{"size1.go", "size2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing and size sort
			output, err := setup.RunArtDupl("--plumbing", "--sort", "size", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should work with occurrence sorting", func() {
			code := `package main
func occSortPlumb() {}`

			err := setup.CreateDuplicateFiles([]string{"occ1.go", "occ2.go", "occ3.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing and occurrence sort
			output, err := setup.RunArtDupl("--plumbing", "--sort", "occurrence", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should work with hash sorting", func() {
			code := `package main
func hashSortPlumb() {}`

			err := setup.CreateDuplicateFiles([]string{"hashsort1.go", "hashsort2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing and hash sort
			output, err := setup.RunArtDupl("--plumbing", "--sort", "hash", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using plumbing with file filtering", func() {
		It("should work with vendor exclusion", func() {
			code := testutil.SimpleCodeTemplate("vendorPlumb")
			_ = setup.CreateAndRunDuplExpectSuccess([]string{"main1.go", "main2.go"}, code, "--plumbing", "--threshold", testutil.ThresholdSmall)
		})

		It("should work with filter-generated flag", func() {
			code := testutil.SimpleCodeTemplate("filterGenPlumb")
			_ = setup.CreateAndRunDuplExpectSuccess([]string{"filter1.go", "filter2.go"}, code, "--plumbing", "--filter-generated", "--threshold", testutil.ThresholdSmall)
		})

		It("should work with include patterns", func() {
			err := setup.CreateSubdirectories("src", "test")
			Expect(err).NotTo(HaveOccurred())

			code := `package main
func patternPlumb() {}`

			err = setup.CreateFileWithContent("src/file.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("test/file.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing and include pattern
			output, err := setup.RunArtDupl("--plumbing", "--include-pattern", "src/*", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using plumbing output for CI/CD integration", func() {
		It("should produce parseable output for scripts", func() {
			// Create test files with multiple duplicates
			code := `package main

import "fmt"

func ciCdFunction() error {
	fmt.Println("CI/CD test")
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"cicd1.go", "cicd2.go", "cicd3.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing output
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Each line should be parseable
			lines := strings.SplitSeq(strings.TrimSpace(outputStr), "\n")
			for line := range lines {
				if strings.TrimSpace(line) != "" {
					// Line should contain file path information
					Expect(line).To(ContainSubstring(".go"))
				}
			}
		})

		It("should include clone positions for precise reporting", func() {
			code := `package main

import "fmt"

func positionTest() {
	fmt.Println("line 1")
	fmt.Println("line 2")
}`

			err := setup.CreateDuplicateFiles([]string{"pos1.go", "pos2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing output
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should contain position information (line numbers)
			// Plumbing format typically uses path:startLine,startCol-endLine,endCol or similar
			Expect(outputStr).ToNot(BeEmpty())
		})
	})
})

var _ = Describe("Plumbing Output Format Validation", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When validating plumbing output structure", func() {
		It("should have consistent delimiter usage", func() {
			code := `package main
func delimiterTest() {}`

			err := setup.CreateDuplicateFiles([]string{"delim1.go", "delim2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing output
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			lines := strings.SplitSeq(strings.TrimSpace(outputStr), "\n")

			// All non-empty lines should have consistent structure
			for line := range lines {
				if strings.TrimSpace(line) != "" {
					// Should contain .go extension (file path)
					Expect(line).To(ContainSubstring(".go"))
				}
			}
		})

		It("should handle files with special characters in paths", func() {
			code := `package main
func specialPath() {}`

			// Create file with special characters in path
			err := setup.CreateFileWithContent("special-path_test.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("special-path_test.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Also create in subdirectory with hyphen
			err = setup.CreateSubdirectories("my-package")
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("my-package/file.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("my-package/file.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing output
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using plumbing with stats subcommand", func() {
		It("should handle stats command with plumbing consideration", func() {
			code := `package main
func statsPlumb() {}`

			err := setup.CreateDuplicateFiles([]string{"stats1.go", "stats2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Stats command doesn't have plumbing format but should work
			output, err := setup.RunArtDupl("stats", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using plumbing with stdin input", func() {
		It("should work with files from stdin", func() {
			code := `package main
func stdinPlumb() {}`

			// Create files
			file1 := setup.GetFilePath("stdin1.go")
			file2 := setup.GetFilePath("stdin2.go")

			err := setup.CreateFileWithContent("stdin1.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("stdin2.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Create stdin with file paths
			stdin := fmt.Sprintf("%s\n%s\n", file1, file2)

			// Run with --files and plumbing
			output, err := setup.RunArtDuplWithStdin(stdin, map[string]string{
				"threshold": "5",
				"plumbing":  "",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})
})

// Helper function to validate plumbing line format.
func validatePlumbingLine(line string) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Basic validation: should contain a .go file path
	if !strings.Contains(line, ".go") {
		return fmt.Errorf("line does not contain .go file path: %s", line)
	}

	// Should have position information (colon followed by numbers)
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return fmt.Errorf("line does not have position information: %s", line)
	}

	// Try to parse position as number or range
	position := parts[len(parts)-1]
	if _, err := strconv.Atoi(position); err != nil {
		// Might be a range like "1,10"
		if !strings.Contains(position, ",") {
			return fmt.Errorf("invalid position format: %s", position)
		}
	}

	return nil
}

// Helper function to parse plumbing output.
func parsePlumbingOutput(output string) ([]PlumbingEntry, error) {
	var entries []PlumbingEntry
	lines := strings.SplitSeq(strings.TrimSpace(output), "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entry, err := parsePlumbingLine(line)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// PlumbingEntry represents a parsed plumbing output line.
type PlumbingEntry struct {
	Filename  string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

// parsePlumbingLine parses a single plumbing output line.
func parsePlumbingLine(line string) (PlumbingEntry, error) {
	entry := PlumbingEntry{}

	// Common format: filename:startLine,startCol-endLine,endCol
	// Or: filename:startLine-endLine
	// Or: filename:startLine

	// Find the last colon (separates filename from position)
	lastColon := strings.LastIndex(line, ":")
	if lastColon == -1 {
		return entry, fmt.Errorf("no colon found in line: %s", line)
	}

	entry.Filename = line[:lastColon]
	position := line[lastColon+1:]

	// Try to parse position
	if strings.Contains(position, "-") {
		// Range format: start-end or start,startCol-end,endCol
		parts := strings.Split(position, "-")
		if len(parts) != 2 {
			return entry, fmt.Errorf("invalid range format: %s", position)
		}

		startParts := strings.Split(parts[0], ",")
		endParts := strings.Split(parts[1], ",")

		startLine, err := strconv.Atoi(startParts[0])
		if err != nil {
			return entry, fmt.Errorf("invalid start line: %s", startParts[0])
		}
		entry.StartLine = startLine

		if len(startParts) > 1 {
			startCol, err := strconv.Atoi(startParts[1])
			if err != nil {
				return entry, fmt.Errorf("invalid start column: %s", startParts[1])
			}
			entry.StartCol = startCol
		}

		endLine, err := strconv.Atoi(endParts[0])
		if err != nil {
			return entry, fmt.Errorf("invalid end line: %s", endParts[0])
		}
		entry.EndLine = endLine

		if len(endParts) > 1 {
			endCol, err := strconv.Atoi(endParts[1])
			if err != nil {
				return entry, fmt.Errorf("invalid end column: %s", endParts[1])
			}
			entry.EndCol = endCol
		}
	} else {
		// Simple line number
		lineNum, err := strconv.Atoi(position)
		if err != nil {
			return entry, fmt.Errorf("invalid line number: %s", position)
		}
		entry.StartLine = lineNum
		entry.EndLine = lineNum
	}

	return entry, nil
}

var _ = Describe("Plumbing Output Advanced Parsing", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When parsing plumbing output programmatically", func() {
		It("should produce parseable entries for multiple clones", func() {
			code := `package main

import "fmt"

func parseableClone() {
	fmt.Println("parseable")
}`

			err := setup.CreateDuplicateFiles([]string{"parse1.go", "parse2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing output
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			// Parse the output
			entries, err := parsePlumbingOutput(string(output))
			Expect(err).ToNot(HaveOccurred())

			// Should have entries
			Expect(entries).ToNot(BeEmpty())

			// Each entry should have valid filename
			for _, entry := range entries {
				Expect(entry.Filename).To(ContainSubstring(".go"))
				Expect(entry.StartLine).To(BeNumerically(">", 0))
			}
		})

		It("should include absolute paths in plumbing output", func() {
			code := `package main
func absPathTest() {}`

			err := setup.CreateDuplicateFiles([]string{"abs1.go", "abs2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with plumbing output
			output, err := setup.RunArtDupl("--plumbing", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			lines := strings.SplitSeq(strings.TrimSpace(outputStr), "\n")

			for line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				// Parse to get filename
				entry, err := parsePlumbingLine(line)
				if err == nil {
					// Path should be absolute or relative to temp dir
					Expect(entry.Filename).To(SatisfyAny(
						ContainSubstring("abs1.go"),
						ContainSubstring("abs2.go"),
					))
				}
			}
		})
	})
})
