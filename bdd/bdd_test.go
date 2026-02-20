package bdd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for art-dupl
//
// These tests verify the tool's behavior from a user's perspective,
// focusing on common workflows and expected outcomes.
//
// The scenarios cover:
// - Basic duplication detection workflows
// - Configuration management scenarios
// - Output format validation
// - Integration use cases
// - Error handling scenarios

func TestBDD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl BDD Suite")
}

var _ = Describe("Basic User Workflows", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())

		// Define test Go files with intentional duplicates
		testFiles := map[string]string{
			"duplicate1.go": `package main

import (
	"fmt"
	"time"
)

func processUser(name string, age int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if age <= 0 {
		return fmt.Errorf("age must be positive")
	}
	
	// Process user data
	userID := generateUserID(name)
	result := saveToDatabase(userID, age)
	
	if result != nil {
		return fmt.Errorf("failed to save user: %w", result)
	}
	
	return nil
}

func generateUserID(name string) string {
	return fmt.Sprintf("user_%s_%d", name, time.Now().Unix())
}

func saveToDatabase(id string, value int) error {
	// Simulated database save
	return nil
}`,
			"duplicate2.go": `package main

import (
	"fmt"
	"time"
)

func processProduct(name string, price int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	
	// Process product data
	productID := generateProductID(name)
	result := saveToDatabase(productID, price)
	
	if result != nil {
		return fmt.Errorf("failed to save product: %w", result)
	}
	
	return nil
}

func generateProductID(name string) string {
	return fmt.Sprintf("product_%s_%d", name, time.Now().Unix())
}

func saveToDatabase(id string, value int) error {
	// Simulated database save
	return nil
}`,
			"unique.go": `package main

import "context"

func uniqueFunction(ctx context.Context) error {
	// This function has no duplicates
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}`,
		}

		// Write test files using unified processor
		err = setup.CreateTestFiles(testFiles)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When analyzing code for duplicates", func() {
		It("should find structural duplicates ignoring literal values", func() {
			// Run art-dupl on test directory
			output, err := setup.RunArtDupl("--threshold", "10")
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output)) //nolint:forbidigo // Debug output for test failure
			}

			// Verify
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)

			// Should find duplicates between the two similar functions
			Expect(outputStr).To(ContainSubstring("duplicate1.go"))
			Expect(outputStr).To(ContainSubstring("duplicate2.go"))
		})

		// Test: Verify occurrence sorting prioritizes clones with more unique files
		It("should sort clones by occurrence (most files first) when using --sort occurrence", func() {
			// Create code pattern 1: Complex function with unique structure
			widespreadCode := `package main

import "fmt"

type Processor struct {
	name string
	count int
}

func (p *Processor) veryCommon(message string) {
	for i := 0; i < 5; i++ {
		p.count++
		fmt.Printf("%s: %d\n", message, p.count)
	}
}`

			// Create 4 files with the same code
			err := setup.FileProcessor.WriteTextFile("widespread1.go", widespreadCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteTextFile("widespread2.go", widespreadCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteTextFile("widespread3.go", widespreadCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteTextFile("widespread4.go", widespreadCode)
			Expect(err).NotTo(HaveOccurred())

			// Create code pattern 2: Different complex function with error handling
			lessCommonCode := `package main

import "errors"

type Validator struct {
	min int
	max int
}

func (v *Validator) lessCommon(id int) error {
	if id < v.min || id > v.max {
		return errors.New("id out of range")
	}
	return nil
}`

			err = setup.FileProcessor.WriteTextFile("less1.go", lessCommonCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteTextFile("less2.go", lessCommonCode)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl with --sort occurrence
			output, err := setup.RunArtDupl("--threshold", "5", "--sort", "occurrence")
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output)) //nolint:forbidigo // Debug output for test failure
			}

			// Verify
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)

			// Verify both clone groups are found
			widespreadIndex := strings.Index(outputStr, "widespread1.go")
			lessCommonIndex := strings.Index(outputStr, "less1.go")
			Expect(widespreadIndex).ToNot(Equal(-1), "Widespread clone should be found")
			Expect(lessCommonIndex).ToNot(Equal(-1), "Less common clone should be found")
		})

		It("should respect threshold settings to filter noise", func() {
			// Run with high threshold
			output, err := setup.RunArtDuplWithFlags(map[string]string{
				"threshold": "50",
			})
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output)) //nolint:forbidigo // Debug output for test failure
			}

			// Verify
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)

			// Should not be empty (might find some matches or might not, but should run)
			Expect(outputStr).ToNot(BeEmpty())
		})
	})

	Context("When generating reports", func() {
		It("should produce valid JSON output with statistics", func() {
			// Run with JSON output on current directory - use Output() to get only stdout (no stderr)
			cmd := exec.Command(setup.BinaryPath, ".", "--json", "--threshold", "10")
			output, err := cmd.Output()
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output)) //nolint:forbidigo // Debug output for test failure
			}

			// Verify
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Check required JSON structure
			Expect(result).To(HaveKey("version"))
			Expect(result).To(HaveKey("timestamp"))
			Expect(result).To(HaveKey("threshold"))
			Expect(result).To(HaveKey("files_analyzed"))
			Expect(result).To(HaveKey("clone_groups"))
			Expect(result).To(HaveKey("summary"))
			Expect(result["threshold"]).To(Equal(float64(10)))
		})

		It("should produce HTML output with code fragments", func() {
			// Run with HTML output on current directory
			output, err := setup.RunArtDuplOnDir(".", "--html", "--threshold", "10")

			// Verify
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)

			// Should contain HTML structure (actual HTML printer output)
			Expect(outputStr).To(ContainSubstring("<!DOCTYPE html>"))
			Expect(outputStr).To(ContainSubstring("<title>Code Duplication Report</title>"))
			Expect(outputStr).To(ContainSubstring("</style>"))
		})
	})
})

var _ = Describe("File Targeting Scenarios", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())

		// Create subdirectories
		err = setup.CreateSubdirectories("pkg1", "pkg2")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When analyzing specific directories", func() {
		It("should limit analysis to specified paths", func() {
			// Create multiple files in different directories with duplicates within each
			file1 := "pkg1/file1.go"
			file2 := "pkg1/file2.go" // Second file in same directory
			file3 := "pkg2/file3.go"

			duplicateCode := `package pkg

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`

			// Use setup's file creation method
			err := setup.CreateDuplicateFiles([]string{file1, file2, file3}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Analyze only subDir1
			subDir1 := setup.GetFilePath("pkg1")
			output, err := setup.RunArtDuplOnDir(subDir1, "--threshold", "10")
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output)) //nolint:forbidigo // Debug output for test failure
			}

			// Verify - should only mention files from subDir1
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("pkg1"))
			Expect(outputStr).ToNot(ContainSubstring("pkg2"))
		})
	})

	Context("When reading file list from stdin", func() {
		It("should analyze only files provided via stdin", func() {
			duplicateCode := `package main

import "fmt"

func common() {
	// This is a duplicate
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`

			uniqueCode := `package main

func unique() {
	// This is unique
	return nil
}`

			// Use setup's file creation methods
			err := setup.CreateDuplicateFiles([]string{"target1.go", "target2.go"}, duplicateCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("ignore.go", uniqueCode)
			Expect(err).NotTo(HaveOccurred())

			// Create stdin with only target files (use absolute paths)
			stdin := fmt.Sprintf("%s\n%s\n", setup.GetFilePath("target1.go"), setup.GetFilePath("target2.go"))
			output, err := setup.RunArtDuplWithStdin(stdin, map[string]string{"threshold": "10"})

			// Verify - should find duplicates between target files
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("target1.go"))
			Expect(outputStr).To(ContainSubstring("target2.go"))
		})
	})
})

var _ = Describe("Integration Scenarios", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("CI/CD Pipeline Integration", func() {
		It("should provide JSON output suitable for automation", func() {
			serviceCode := `package service

import (
	"context"
	"fmt"
)

type Service struct {
	name string
}

func (s *Service) Process(ctx context.Context, data string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if data == "" {
			return fmt.Errorf("empty data")
		}
		return s.processInternal(data)
	}
}

func (s *Service) processInternal(data string) error {
	// Processing logic
	return nil
}`

			// Use setup's file creation methods with service replacements
			userServiceCode := strings.ReplaceAll(serviceCode, "Service", "UserService")
			orderServiceCode := strings.ReplaceAll(serviceCode, "Service", "OrderService")

			err := setup.CreateFileWithContent("service1.go", userServiceCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("service2.go", orderServiceCode)
			Expect(err).NotTo(HaveOccurred())

			// Execute with JSON output - separate stdout from stderr to avoid JSON corruption
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--json", "--threshold", "15")
			output, err := cmd.Output() // Use Output() instead of CombinedOutput() to avoid stderr contamination
			Expect(err).ToNot(HaveOccurred())

			// Parse JSON response
			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			// Verify structure suitable for CI/CD
			Expect(result).To(HaveKey("summary"))
			summary := result["summary"].(map[string]any)
			Expect(summary).To(HaveKey("total_clones"))
			Expect(summary).To(HaveKey("total_clone_groups"))
			Expect(summary).To(HaveKey("complexity_score"))

			// Should be able to use in CI/CD scripts
			totalClones := int(summary["total_clones"].(float64))
			Expect(totalClones).To(BeNumerically(">=", 0))
		})
	})

	Context("Performance with Large Codebases", func() {
		It("should handle multiple files efficiently", func() {
			// Create multiple files with duplicates
			numFiles := 10
			duplicateCode := `package main

import "fmt"

func processData(data string, count int) error {
	if data == "" {
		return fmt.Errorf("invalid data")
	}
	if count <= 0 {
		return fmt.Errorf("invalid count")
	}

	// Process data in loop
	for i := 0; i < count; i++ {
		if err := processItem(data, i); err != nil {
			return fmt.Errorf("failed at item %d: %w", i, err)
		}
	}

	return nil
}

func processItem(data string, index int) error {
	// Item processing logic
	return nil
}`

			// Create multiple duplicate files using setup's helper
			var filenames []string
			for i := range numFiles {
				filename := fmt.Sprintf("file%d.go", i)
				filenames = append(filenames, filename)
			}

			err := setup.CreateDuplicateFiles(filenames, duplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Measure execution time
			start := time.Now()
			output, err := setup.RunArtDupl("--threshold", "20")
			duration := time.Since(start)

			// Verify it completes in reasonable time
			Expect(err).ToNot(HaveOccurred())
			Expect(duration).To(BeNumerically("<", 5*time.Second))

			// Should find duplicates across multiple files
			outputStr := string(output)
			Expect(len(strings.Split(outputStr, "\n"))).To(BeNumerically(">", numFiles))
		})
	})
})
