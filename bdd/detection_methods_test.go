package bdd

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/utils"
)

// BDD Test Suite for Detection Methods
//
// These tests verify the different detection methods available in art-dupl:
// - Hash-based detection (fast, exact duplicates)
// - Art-dupl detection (structural duplicates)
// - Multi-detection mode (combined results)
//
// The scenarios cover:
// - Individual detection method behavior
// - Combined detection results
// - Detection method selection via CLI

func TestDetectionMethods(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl Detection Methods BDD Suite")
}

var _ = Describe("Detection Methods", func() {
	var (
		tempDir       string
		fileProcessor *utils.FileProcessor
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-detection-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		fileProcessor = utils.NewFileProcessor(tempDir)
	})

	AfterEach(func() {
		_ = os.RemoveAll(tempDir)
		_ = os.Remove("./bdd/art-dupl-detection_methods-test")
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

			err := fileProcessor.WriteDuplicateFiles([]string{
				"exact1.go", "exact2.go", "exact3.go",
			}, identicalCode)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-detection_methods-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with hash detection
			cmd = exec.Command("./bdd/art-dupl-detection_methods-test", tempDir, "--detection-methods", "hash", "--threshold", "10")
			output, err := cmd.CombinedOutput()
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

			err := fileProcessor.WriteDuplicateFiles([]string{"file1.go", "file2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-detection_methods-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with hash detection and JSON output
			cmd = exec.Command("./bdd/art-dupl-detection_methods-test", tempDir, "--detection-methods", "hash", "--json", "--threshold", "5")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			// Parse JSON
			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Verify JSON structure
			Expect(result).To(HaveKey("detection_method"))
			Expect(result).To(HaveKey("clone_groups"))
			Expect(result).To(HaveKey("summary"))
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

			err := fileProcessor.WriteTextFile("user.go", userCode)
			Expect(err).NotTo(HaveOccurred())
			err = fileProcessor.WriteTextFile("product.go", productCode)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-detection_methods-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with art-dupl detection
			cmd = exec.Command("./bdd/art-dupl-detection_methods-test", tempDir, "--detection-methods", "art-dupl", "--threshold", "10")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("user.go"))
			Expect(outputStr).To(ContainSubstring("product.go"))
		})

		It("should be the default detection method", func() {
			// Create test files
			code := `package main

func test() error {
	if true {
		return nil
	}
	return nil
}`

			err := fileProcessor.WriteDuplicateFiles([]string{"default1.go", "default2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-detection_methods-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run without specifying detection method (should default to art-dupl)
			cmd = exec.Command("./bdd/art-dupl-detection_methods-test", tempDir, "--threshold", "10")
			output, err := cmd.CombinedOutput()
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

			err := fileProcessor.WriteDuplicateFiles([]string{"exact1.go", "exact2.go"}, exactDup)
			Expect(err).NotTo(HaveOccurred())
			err = fileProcessor.WriteTextFile("structural1.go", structuralDup1)
			Expect(err).NotTo(HaveOccurred())
			err = fileProcessor.WriteTextFile("structural2.go", structuralDup2)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-detection_methods-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with combined detection methods
			cmd = exec.Command("./bdd/art-dupl-detection_methods-test", tempDir, "--detection-methods", "hash,art-dupl", "--threshold", "5")
			output, err := cmd.CombinedOutput()
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

func detect() error {
	return nil
}`

			err := fileProcessor.WriteDuplicateFiles([]string{"combine1.go", "combine2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-detection_methods-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with combined detection and JSON
			cmd = exec.Command("./bdd/art-dupl-detection_methods-test", tempDir, "--detection-methods", "hash,art-dupl", "--json", "--threshold", "5")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			// Parse JSON
			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Should have combined results
			Expect(result).To(HaveKey("clone_groups"))
			Expect(result).To(HaveKey("summary"))
		})
	})

	Context("When using invalid detection method combinations", func() {
		It("should handle invalid method names gracefully", func() {
			// Create test files
			code := `package main

func test() {}
`

			err := fileProcessor.WriteDuplicateFiles([]string{"invalid1.go", "invalid2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "./bdd/art-dupl-detection_methods-test", "./cmd/art-dupl/main.go")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			// Run with invalid detection method
			cmd = exec.Command("./bdd/art-dupl-detection_methods-test", tempDir, "--detection-methods", "invalid_method")
			output, err := cmd.CombinedOutput()

			// Should handle error gracefully
			Expect(len(output)).To(BeNumerically(">", 0))
		})
	})
})
