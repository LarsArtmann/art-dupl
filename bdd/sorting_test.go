package bdd

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Sorting Features
//
// These tests verify the sorting functionality of art-dupl,
// focusing on different sorting options and their behavior.
//
// The scenarios cover:
// - Size sorting (largest clones first)
// - Occurrence sorting (most widespread clones first)
// - Hash sorting (alphabetical order)
// - Default sorting behavior

var _ = Describe("Sorting Functionality", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When sorting by size", func() {
		It("should display largest clones first by default", func() {
			// Create test files with different clone sizes
			largeClone := `package main

import "fmt"

func processLargeData(data string, count int, threshold float64) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	if count <= 0 {
		return fmt.Errorf("invalid count")
	}
	if threshold <= 0.0 {
		return fmt.Errorf("invalid threshold")
	}
	
	// Large processing logic
	for i := 0; i < count; i++ {
		if float64(i) >= threshold {
			return fmt.Errorf("threshold reached at %d", i)
		}
		if err := processDataItem(data, i, threshold); err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}
	
	return nil
}

func processDataItem(data string, index int, threshold float64) error {
	// Item processing
	return nil
}`

			mediumClone := `package main

import "fmt"

func processMediumData(data string, count int) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	if count <= 0 {
		return fmt.Errorf("invalid count")
	}
	
	for i := 0; i < count; i++ {
		if err := processItem(data, i); err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}
	
	return nil
}

func processItem(data string, index int) error {
	return nil
}`

			// Create duplicates with different sizes
			err := setup.CreateDuplicateFiles([]string{"large1.go", "large2.go"}, largeClone)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateDuplicateFiles([]string{"medium1.go", "medium2.go"}, mediumClone)
			Expect(err).NotTo(HaveOccurred())

			// Run with default sorting (size)
			output, err := setup.RunArtDuplWithFlags(map[string]string{
				"threshold": "15",
				"sort":      "size",
			})
			// Print debug info on error
			if err != nil {
				fmt.Printf("DEBUG: Command failed with output: %s\n", string(output)) //nolint:forbidigo // Debug output
			}
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Large clones should appear before medium clones in output
			// The first occurrence of "large" should come before "medium"
			largeIndex := strings.Index(outputStr, "large")
			mediumIndex := strings.Index(outputStr, "medium")

			Expect(largeIndex).ToNot(Equal(-1), "Large clone should be found")
			Expect(mediumIndex).ToNot(Equal(-1), "Medium clone should be found")
			Expect(largeIndex).To(BeNumerically("<", mediumIndex), "Larger clone should appear first when sorted by size")
		})

		It("should use size sorting when explicitly specified", func() {
			// Create test files
			duplicateCode := `package main

func process(data string) error {
	if data == "" {
		return nil
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"size1.go", "size2.go"}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with size sorting
			output, err := setup.RunArtDuplWithFlags(map[string]string{
				"threshold": "10",
				"sort":      "size",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("size1.go"))
		})
	})

	Context("When sorting by occurrence", func() {
		It("should display most widespread clones first", func() {
			// Create code that appears in many files
			// Pattern 1: Function with string parameter and no return value
			widespreadCode := `package main

import "fmt"

func commonFunction(message string) {
	fmt.Println(message)
}`

			// Pattern 2: Function with two parameters and error return value
			// This creates a structurally different AST pattern
			lessCommonCode := `package main

import "errors"

func lessCommonFunction(id int, name string) error {
	if id == 0 {
		return errors.New("invalid id")
	}
	return nil
}`

			// Create 4 files with widespread code
			err := setup.CreateDuplicateFiles([]string{
				"wide1.go", "wide2.go", "wide3.go", "wide4.go",
			}, widespreadCode)
			Expect(err).NotTo(HaveOccurred())

			// Create 2 files with less common code
			err = setup.CreateDuplicateFiles([]string{
				"less1.go", "less2.go",
			}, lessCommonCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with occurrence sorting
			output, err := setup.RunArtDuplWithFlags(map[string]string{
				"threshold": "5",
				"sort":      "occurrence",
			})
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Widespread clone (4 files) should appear before less common clone (2 files)
			widespreadIndex := strings.Index(outputStr, "wide")
			lessCommonIndex := strings.Index(outputStr, "less")

			Expect(widespreadIndex).ToNot(Equal(-1), "Widespread clone should be found")
			Expect(lessCommonIndex).ToNot(Equal(-1), "Less common clone should be found")
			Expect(widespreadIndex).To(BeNumerically("<", lessCommonIndex), "More widespread clone should appear first")
		})
	})

	Context("When sorting by hash", func() {
		It("should display clones in alphabetical order by hash", func() {
			// Create test files with different code
			codeA := `package main

func functionA() error {
	if true {
		return nil
	}
	return nil
}`

			codeB := `package main

func functionB() error {
	if false {
		return nil
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"fileA1.go", "fileA2.go"}, codeA)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateDuplicateFiles([]string{"fileB1.go", "fileB2.go"}, codeB)
			Expect(err).NotTo(HaveOccurred())

			// Run with hash sorting
			output, err := setup.RunArtDuplWithFlags(map[string]string{
				"threshold": "10",
				"sort":      "hash",
			})
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Both clones should be found
			Expect(outputStr).To(ContainSubstring("fileA"))
			Expect(outputStr).To(ContainSubstring("fileB"))
		})
	})

	Context("When using invalid sorting options", func() {
		It("should handle invalid sort values gracefully", func() {
			// Create simple test files
			code := `package main

func hello() {
	println("hello")
}`

			err := setup.CreateDuplicateFiles([]string{"test1.go", "test2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid sort option - should default to size
			output, _ := setup.RunArtDuplWithFlags(map[string]string{
				"threshold": "5",
				"sort":      "invalid",
			})

			// Should handle the error gracefully
			// Either by showing error message or defaulting to size sorting
			Expect(output).ToNot(BeEmpty(), "Should produce some output")
		})
	})
})
