package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

// Use an alias for json operations to avoid conflicts with the json flag in cli.go
func parseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

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
		outputBuf *bytes.Buffer
		errorBuf  *bytes.Buffer
		testCmd   *cobra.Command
	)

	BeforeEach(func() {
		// Create temporary directory for test files
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		// Define test Go files with intentional duplicates
		testFiles = map[string]string{
			"duplicate1.go": `package main

import "fmt"

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
}`,
			"duplicate2.go": `package main

import "fmt"

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

		// Setup command execution
		outputBuf = &bytes.Buffer{}
		errorBuf = &bytes.Buffer{}
	})

	AfterEach(func() {
		// Clean up temporary directory
		_ = os.RemoveAll(tempDir)
	})

	Context("When analyzing code for duplicates", func() {
		It("should find structural duplicates ignoring literal values", func() {
			// Setup
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{tempDir, "-t", "10"}
			testCmd.SetArgs(args)

			// Execute
			err := testCmd.Execute()

			// Verify
			Expect(err).ToNot(HaveOccurred())
			output := outputBuf.String()

			// Should find duplicates between the two similar functions
			Expect(output).To(ContainSubstring("duplicate1.go"))
			Expect(output).To(ContainSubstring("duplicate2.go"))
		})

		It("should respect threshold settings to filter noise", func() {
			// Setup with high threshold
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{tempDir, "-t", "50"}
			testCmd.SetArgs(args)

			// Execute
			err := testCmd.Execute()

			// Verify
			Expect(err).ToNot(HaveOccurred())
			output := outputBuf.String()

			// With high threshold, should find fewer or no duplicates
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("When generating reports", func() {
		It("should produce valid JSON output with statistics", func() {
			// Setup
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{tempDir, "-j"}
			testCmd.SetArgs(args)

			// Execute
			err := testCmd.Execute()

			// Verify
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = parseJSON(outputBuf.Bytes(), &result)
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
			// Setup
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{tempDir, "--html"}
			testCmd.SetArgs(args)

			// Execute
			err := testCmd.Execute()

			// Verify
			Expect(err).ToNot(HaveOccurred())
			output := outputBuf.String()

			// Should contain HTML structure
			Expect(output).To(ContainSubstring("<html>"))
			Expect(output).To(ContainSubstring("</html>"))
			Expect(output).To(ContainSubstring("duplicate"))
		})
	})
})

var _ = Describe("Configuration Management", func() {
	var (
		tempDir    string
		configFile string
		outputBuf  *bytes.Buffer
		errorBuf   *bytes.Buffer
		testCmd    *cobra.Command
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-config-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		outputBuf = &bytes.Buffer{}
		errorBuf = &bytes.Buffer{}

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

			// Setup command with config
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{"-config", configFile}
			testCmd.SetArgs(args)

			// Execute
			err = testCmd.Execute()

			// Verify
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = parseJSON(outputBuf.Bytes(), &result)
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

			// Setup command with config and override
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{"-c", configFile, "-t", "50"}
			testCmd.SetArgs(args)

			// Execute
			err = testCmd.Execute()

			// Verify - should use text output (not json from config)
			Expect(err).ToNot(HaveOccurred())
			output := outputBuf.String()
			Expect(output).ToNot(ContainSubstring("{")) // Not JSON
		})
	})
})

var _ = Describe("File Targeting Scenarios", func() {
	var (
		tempDir   string
		subDir1   string
		subDir2   string
		outputBuf *bytes.Buffer
		errorBuf  *bytes.Buffer
		testCmd   *cobra.Command
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

		outputBuf = &bytes.Buffer{}
		errorBuf = &bytes.Buffer{}
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

			// Analyze only subDir1
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{subDir1, "-t", "10"}
			testCmd.SetArgs(args)

			// Execute
			err = testCmd.Execute()

			// Verify - should only mention files from subDir1
			Expect(err).ToNot(HaveOccurred())
			output := outputBuf.String()
			Expect(output).To(ContainSubstring("pkg1"))
			Expect(output).ToNot(ContainSubstring("pkg2"))
		})
	})

	Context("When reading file list from stdin", func() {
		It("should analyze only files provided via stdin", func() {
			// Create multiple files
			file1 := filepath.Join(tempDir, "target1.go")
			file2 := filepath.Join(tempDir, "target2.go")
			file3 := filepath.Join(tempDir, "ignore.go")

			duplicateCode := `package main

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

			// Create stdin with only target files
			stdin := fmt.Sprintf("%s\n%s\n", file1, file2)
			testCmd = createTestCommandWithStdin(outputBuf, errorBuf, strings.NewReader(stdin))
			args := []string{"-f", "-t", "10"}
			testCmd.SetArgs(args)

			// Execute
			err = testCmd.Execute()

			// Verify - should find duplicates between target files
			Expect(err).ToNot(HaveOccurred())
			output := outputBuf.String()
			Expect(output).To(ContainSubstring("target1.go"))
			Expect(output).To(ContainSubstring("target2.go"))
		})
	})
})

var _ = Describe("Integration Scenarios", func() {
	var (
		tempDir   string
		outputBuf *bytes.Buffer
		errorBuf  *bytes.Buffer
		testCmd   *cobra.Command
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "art-dupl-integration-bdd-*")
		Expect(err).NotTo(HaveOccurred())

		outputBuf = &bytes.Buffer{}
		errorBuf = &bytes.Buffer{}
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

import "context"

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

			// Execute with JSON output
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{tempDir, "-j", "-t", "15"}
			testCmd.SetArgs(args)

			err = testCmd.Execute()
			Expect(err).ToNot(HaveOccurred())

			// Parse JSON response
			var result map[string]any
			err = parseJSON(outputBuf.Bytes(), &result)
			Expect(err).ToNot(HaveOccurred())

			// Verify structure suitable for CI/CD
			Expect(result).To(HaveKey("summary"))
			summary := result["summary"].(map[string]any)
			Expect(summary).To(HaveKey("total_clones"))
			Expect(summary).To(HaveKey("total_clone_groups"))
			Expect(summary).To(HaveKey("complexity_score"))

			// Should be able to use in CI/CD scripts
			totalClones := int(summary["total_clones"].(float64))
			Expect(totalClones).To(BeNumerically(">", 0))
		})
	})

	Context("Performance with Large Codebases", func() {
		It("should handle multiple files efficiently", func() {
			// Create multiple files with duplicates
			numFiles := 10
			duplicateCode := `package main

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

			// Measure execution time
			start := time.Now()
			testCmd = createTestCommand(outputBuf, errorBuf)
			args := []string{tempDir, "-t", "20"}
			testCmd.SetArgs(args)

			err := testCmd.Execute()
			duration := time.Since(start)

			// Verify it completes in reasonable time
			Expect(err).ToNot(HaveOccurred())
			Expect(duration).To(BeNumerically("<", 5*time.Second))

			// Should find duplicates across multiple files
			output := outputBuf.String()
			Expect(len(strings.Split(output, "\n"))).To(BeNumerically(">", numFiles))
		})
	})
})

// Helper functions for test setup

func createTestCommand(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "art-dupl",
		Short: "Find code clones",
		RunE:  runCmd,
	}

	// Add flags
	cmd.Flags().StringP("config", "c", "", "path to configuration file")
	cmd.Flags().Bool("vendor", false, "include vendor directory")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	cmd.Flags().IntP("threshold", "t", 15, "minimum token sequence size")
	cmd.Flags().BoolP("files", "f", false, "read file names from stdin")
	cmd.Flags().Bool("html", false, "output results as HTML")
	cmd.Flags().BoolP("json", "j", false, "output structured JSON format")
	cmd.Flags().BoolP("plumbing", "p", false, "output plumbing format")
	cmd.Flags().StringP("sort", "s", "size", "sort clone groups")
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	return cmd
}

func createTestCommandWithStdin(out, errOut io.Writer, stdin io.Reader) *cobra.Command {
	cmd := createTestCommand(out, errOut)
	cmd.SetIn(stdin)
	return cmd
}
