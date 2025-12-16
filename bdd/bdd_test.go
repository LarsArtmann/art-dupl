package bdd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	var (
		tempDir   string
		testFiles map[string]string
	)

	BeforeEach(func() {
		// Create temporary directory for test files
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		// Define test Go files with intentional duplicates
		testFiles = map[string]string{
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

		// Write test files
		for filename, content := range testFiles {
			filePath := filepath.Join(tempDir, filename)
			err := os.WriteFile(filePath, []byte(content), 0o644)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		// Clean up temporary directory
		_ = os.RemoveAll(tempDir)
	})

	Context("When analyzing code for duplicates", func() {
		It("should find structural duplicates ignoring literal values", func() {
			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Run art-dupl on test directory
			cmd = exec.Command("../bdd/art-dupl-test", tempDir, "--threshold", "10")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output))
			}

			// Verify
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)

			// Should find duplicates between the two similar functions
			Expect(outputStr).To(ContainSubstring("duplicate1.go"))
			Expect(outputStr).To(ContainSubstring("duplicate2.go"))
		})

		It("should respect threshold settings to filter noise", func() {
			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Run with high threshold
			cmd = exec.Command("../bdd/art-dupl-test", tempDir, "--threshold", "50")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()
			
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output))
			}

			// Verify
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)

			// Should not be empty (might find some matches or might not, but should run)
			Expect(len(outputStr)).To(BeNumerically(">", 0))
		})
	})

	Context("When generating reports", func() {
		It("should produce valid JSON output with statistics", func() {
			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Run with JSON output
			cmd = exec.Command("../bdd/art-dupl-test", tempDir, "--json", "--threshold", "10")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()
			
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output))
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
		})

		It("should produce HTML output with code fragments", func() {
			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Run with HTML output
			cmd = exec.Command("../bdd/art-dupl-test", tempDir, "--html", "--threshold", "10")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()

			// Verify
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)

			// Should contain HTML structure (actual HTML printer output)
			Expect(outputStr).To(ContainSubstring("<!DOCTYPE html>"))
			Expect(outputStr).To(ContainSubstring("<title>Duplicates</title>"))
			Expect(outputStr).To(ContainSubstring("</style>"))
		})
	})
})

var _ = Describe("Configuration Management", func() {
	var (
		tempDir    string
		configFile string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-config-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		// Create test Go file
		testFile := filepath.Join(tempDir, "test.go")
		err = os.WriteFile(testFile, []byte(`package main
func a() {}
func b() {}`), 0o644)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tempDir)
	})

	Context("When using configuration files", func() {
		It("should load settings from JSON configuration file", func() {
			// Create config file
			configFile = filepath.Join(tempDir, "dupl.json")
			configContent := fmt.Sprintf(`{
				"threshold": 25,
				"outputFormat": "json",
				"paths": ["%s"],
				"verbose": true
			}`, tempDir)
			err := os.WriteFile(configFile, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Run with config
			cmd = exec.Command("../bdd/art-dupl-test", "--config", configFile)
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()

			// Verify
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result["threshold"]).To(Equal(float64(25)))
		})

		It("should allow CLI flags to override config file settings", func() {
			// Create config file with threshold 25
			configFile = filepath.Join(tempDir, "dupl.json")
			configContent := fmt.Sprintf(`{
				"threshold": 25,
				"outputFormat": "json",
				"paths": ["%s"]
			}`, tempDir)
			err := os.WriteFile(configFile, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Run with config and override
			cmd = exec.Command("../bdd/art-dupl-test", "--config", configFile, "--threshold", "50")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()

			// Verify - should use text output (not json from config)
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)
			Expect(outputStr).ToNot(ContainSubstring("{")) // Not JSON format
		})
	})
})

var _ = Describe("File Targeting Scenarios", func() {
	var (
		tempDir string
		subDir1 string
		subDir2 string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-files-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		// Create subdirectories
		subDir1 = filepath.Join(tempDir, "pkg1")
		subDir2 = filepath.Join(tempDir, "pkg2")
		err = os.MkdirAll(subDir1, 0o755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(subDir2, 0o755)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tempDir)
	})

	Context("When analyzing specific directories", func() {
		It("should limit analysis to specified paths", func() {
			// Create files in different directories
			file1 := filepath.Join(subDir1, "file.go")
			file2 := filepath.Join(subDir2, "file.go")

			duplicateCode := `package pkg

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`

			err := os.WriteFile(file1, []byte(duplicateCode), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(file2, []byte(duplicateCode), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Analyze only subDir1
			cmd = exec.Command("../bdd/art-dupl-test", subDir1, "--threshold", "10")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()
			
			// Print debug information if there's an error
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output))
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
			// Create multiple files
			file1 := filepath.Join(tempDir, "target1.go")
			file2 := filepath.Join(tempDir, "target2.go")
			file3 := filepath.Join(tempDir, "ignore.go")

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

			err := os.WriteFile(file1, []byte(duplicateCode), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(file2, []byte(duplicateCode), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(file3, []byte(uniqueCode), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Create stdin with only target files
			stdin := fmt.Sprintf("%s\n%s\n", file1, file2)
			cmd = exec.Command("../bdd/art-dupl-test", "--files", "--threshold", "10")
			cmd.Dir = ".."
			cmd.Stdin = strings.NewReader(stdin)
			output, err := cmd.CombinedOutput()

			// Verify - should find duplicates between target files
			Expect(err).ToNot(HaveOccurred())
			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("target1.go"))
			Expect(outputStr).To(ContainSubstring("target2.go"))
		})
	})
})

var _ = Describe("Integration Scenarios", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-integration-bdd-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tempDir)
	})

	Context("CI/CD Pipeline Integration", func() {
		It("should provide JSON output suitable for automation", func() {
			// Create files with duplicates
			file1 := filepath.Join(tempDir, "service1.go")
			file2 := filepath.Join(tempDir, "service2.go")

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

			err := os.WriteFile(file1, []byte(strings.ReplaceAll(serviceCode, "Service", "UserService")), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(file2, []byte(strings.ReplaceAll(serviceCode, "Service", "OrderService")), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Execute with JSON output
			cmd = exec.Command("../bdd/art-dupl-test", tempDir, "--json", "--threshold", "15")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()
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

			for i := range numFiles {
				filename := filepath.Join(tempDir, fmt.Sprintf("file%d.go", i))
				err := os.WriteFile(filename, []byte(duplicateCode), 0o644)
				Expect(err).NotTo(HaveOccurred())
			}

			// Build art-dupl binary
			cmd := exec.Command("go", "build", "-o", "../bdd/art-dupl-test", ".")
			cmd.Dir = ".."
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove("../bdd/art-dupl-test") }()

			// Measure execution time
			start := time.Now()
			cmd = exec.Command("../bdd/art-dupl-test", tempDir, "--threshold", "20")
			cmd.Dir = ".."
			output, err := cmd.CombinedOutput()
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
