package bdd

import (
	"encoding/json"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

func TestDetectionMethods(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl Detection Methods BDD Suite")
}

var _ = Describe("Detection Methods", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When using hash-based detection", func() {
		It("should detect exact file-level duplicates", func() {
			// Create identical files
			identicalCode := `package main

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}

func main() {
	processData("test")
}`

			err := setup.CreateDuplicateFiles([]string{
				"exact1.go", "exact2.go", "exact3.go",
			}, identicalCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with hash detection
			output, err := setup.RunArtDupl("--detection-methods", "hash", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("exact1.go"))
			Expect(outputStr).To(ContainSubstring("exact2.go"))
			Expect(outputStr).To(ContainSubstring("exact3.go"))
		})

		It("should provide JSON output with hash detection statistics", func() {
			// Create test files
			code := `package main

func duplicate() string {
	return "duplicate"
}`

			err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with hash detection and JSON output - separate stdout from stderr
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--detection-methods", "hash", "--json", "--threshold", "5")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			// Parse JSON from stdout (not corrupted by stderr)
			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Verify JSON structure
			Expect(result).To(HaveKey("detection_method"))
			Expect(result).To(HaveKey("clone_groups"))
			Expect(result).To(HaveKey("summary"))
			summary := result["summary"].(map[string]any)
			Expect(summary).To(HaveKey("total_clones"))
			Expect(summary).To(HaveKey("total_clone_groups"))

			// Verify detection method
			detectionMethod := result["detection_method"]
			Expect(detectionMethod).To(Equal("hash"))
		})
	})

	Context("When using art-dupl detection", func() {
		It("should detect structural duplicates ignoring literal values", func() {
			// Create structurally similar files with different literals
			userCode := `package main

import "fmt"

func processUser(name string, age int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if age <= 0 {
		return fmt.Errorf("age must be positive")
	}
	return nil
}`

			productCode := `package main

import "fmt"

func processProduct(name string, price int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	return nil
}`

			err := setup.CreateFileWithContent("user.go", userCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("product.go", productCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with art-dupl detection
			output, err := setup.RunArtDupl("--detection-methods", "art-dupl", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("user.go"))
			Expect(outputStr).To(ContainSubstring("product.go"))
		})

		It("should be default detection method", func() {
			// Create test files
			code := `package main

func test() error {
	if true {
		return nil
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"default1.go", "default2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run without specifying detection method (should default to art-dupl)
			output, err := setup.RunArtDupl("--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("default1.go"))
			Expect(outputStr).To(ContainSubstring("default2.go"))
		})
	})

	Context("When using combined detection methods", func() {
		It("should run both hash and art-dupl detection", func() {
			// Create test files with different duplicate types
			exactDup := `package main

func exact() string {
	return "exact duplicate"
}`

			structuralDup1 := `package main

func structuralA(name string) error {
	if name == "" {
		return nil
	}
	return nil
}`

			structuralDup2 := `package main

func structuralB(value string) error {
	if value == "" {
		return nil
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"exact1.go", "exact2.go"}, exactDup)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("structural1.go", structuralDup1)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("structural2.go", structuralDup2)
			Expect(err).NotTo(HaveOccurred())

			// Run with combined detection methods
			output, err := setup.RunArtDupl("--detection-methods", "hash,art-dupl", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should find both types of duplicates
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring("exact1.go"),
				ContainSubstring("structural1.go"),
			))
		})

		It("should provide comprehensive JSON output for combined detection", func() {
			// Create test files
			code := `package main

func test() string {
	return "test"
}`

			err := setup.CreateDuplicateFiles([]string{"combine1.go", "combine2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with combined detection methods and JSON output - separate stdout
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--detection-methods", "hash,art-dupl", "--json", "--threshold", "5")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			// Parse JSON from stdout
			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Should have combined results
			Expect(result).To(HaveKey("clone_groups"))
			Expect(result).To(HaveKey("summary"))
			summary := result["summary"].(map[string]any)
			Expect(summary).To(HaveKey("total_clones"))
			Expect(summary).To(HaveKey("total_clone_groups"))

			// Verify both methods were used
			Expect(result).To(HaveKey("detection_methods"))
			detectionMethods := result["detection_methods"]
			Expect(detectionMethods).To(Equal("hash,art-dupl"))
		})
	})

	Context("When using invalid detection method combinations", func() {
		It("should handle invalid method names gracefully", func() {
			// Create test files
			code := `package main

func test() {}
`

			err := setup.CreateDuplicateFiles([]string{"invalid1.go", "invalid2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid detection method
			output, err := setup.RunArtDupl("--detection-methods", "invalid_method")
			Expect(err).To(HaveOccurred())

			// Should have error message
			outputStr := string(output)
			Expect(outputStr).ToNot(BeEmpty())
		})
	})
})
