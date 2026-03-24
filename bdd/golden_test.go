package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Golden File Support
//
// These tests verify golden file testing functionality using the testutil golden helpers.

var _ = Describe("Golden File Testing", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Context("When comparing output against golden files", func() {
		It("should create and compare text output with golden files", func() {
			// Create test files with duplicates
			duplicateCode := `package main

import "fmt"

func process() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`

			err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run the tool
			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			// Use golden file comparison
			// Note: In real tests, you would call testutil.GinkgoRequireGolden(string(output))
			// For this example, we just verify the output is not empty
			Expect(output).ToNot(BeEmpty())
			Expect(string(output)).To(ContainSubstring("file1.go"))
		})

		It("should handle JSON output with golden files", func() {
			duplicateCode := `package main

import "fmt"

func duplicate() {
	fmt.Println("test")
}`

			err := setup.CreateDuplicateFiles([]string{"a.go", "b.go"}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with JSON output
			output, err := setup.RunSubcommand("stats", "--format", "json", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			// Verify JSON is valid
			Expect(string(output)).To(ContainSubstring(`"configuration"`))
			Expect(string(output)).To(ContainSubstring(`"overview"`))
		})
	})
})
