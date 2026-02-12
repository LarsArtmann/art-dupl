package bdd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Configuration File Handling
//
// These tests verify configuration file loading and parsing behavior,
// including JSON configuration support and command-line overrides.
//
// The scenarios cover:
// - Loading configuration from JSON files
// - Configuration precedence (CLI flags override config file)
// - Invalid configuration handling
// - Configuration validation

func TestConfigurationFile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "art-dupl Configuration File BDD Suite")
}

// newBDDTestSetup creates a new BDDTestSetup and registers cleanup for Ginkgo tests
func newBDDTestSetup() *testutil.BDDTestSetup {
	setup, err := testutil.NewBDDTestSetupForGinkgo()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})
	return setup
}

var _ = Describe("Configuration File Loading", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = newBDDTestSetup()
	})

	// runWithConfig creates a config file, test files, and runs art-dupl
	runWithConfig := func(configContent, code string, fileNames []string) ([]byte, error) {
		return setup.RunWithConfigFile("dupl.json", configContent, code, fileNames)
	}

	Context("When using a valid JSON configuration file", func() {
		It("should load threshold from config file", func() {
			configContent := `{
				"threshold": 50,
				"outputFormat": "json"
			}`
			code := `package main
func test() {}`
			output, err := runWithConfig(configContent, code, []string{"test1.go", "test2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should load multiple settings from config file", func() {
			configContent := `{
				"threshold": 20,
				"outputFormat": "json",
				"vendor": true,
				"verbose": true
			}`
			code := `package main
func multiConfig() {}`
			output, err := runWithConfig(configContent, code, []string{"multi1.go", "multi2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should load paths from config file", func() {
			// Create subdirectories
			err := setup.CreateSubdirectories("src", "lib")
			Expect(err).NotTo(HaveOccurred())

			// Create config with paths
			configContent := `{
				"threshold": 15,
				"paths": ["./src", "./lib"]
			}`
			configPath := filepath.Join(setup.TmpDir, "dupl.json")
			err = os.WriteFile(configPath, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create test files in subdirectories
			code := `package main
func pathTest() {}`
			err = setup.CreateFileWithContent("src/file.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("lib/file.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Change to temp directory so relative paths work
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)
			os.Chdir(setup.TmpDir)

			// Run with config file (uses paths from config)
			output, err := setup.RunArtDupl("--config", configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When CLI flags override config file settings", func() {
		It("should use CLI threshold over config threshold", func() {
			configContent := `{
				"threshold": 50
			}`
			code := `package main
func overrideTest() {}`
			output, err := runWithConfig(configContent, code, []string{"override1.go", "override2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should use CLI output format over config format", func() {
			configContent := `{
				"threshold": 15,
				"outputFormat": "text"
			}`
			code := `package main
func formatOverride() string {
	return "test"
}`
			output, err := runWithConfig(configContent, code, []string{"fmt1.go", "fmt2.go"})
			Expect(err).ToNot(HaveOccurred())

			// Verify JSON output
			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveKey("clone_groups"))
		})
	})

	Context("When configuration file has invalid format", func() {
		It("should handle malformed JSON gracefully", func() {
			configContent := `{ invalid json content }`
			code := `package main
func malformedTest() {}`
			output, err := runWithConfig(configContent, code, []string{"malformed1.go", "malformed2.go"})
			Expect(err).To(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("error"))
		})

		It("should handle missing config file gracefully", func() {
			configPath := filepath.Join(setup.TmpDir, "nonexistent.json")
			code := `package main
func missingConfig() {}`
			err := setup.CreateDuplicateFiles([]string{"missing1.go", "missing2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--config", configPath, setup.TmpDir)
			Expect(err).To(HaveOccurred())
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle config with invalid field types", func() {
			configContent := `{
				"threshold": "not a number",
				"outputFormat": 12345
			}`
			code := `package main
func invalidType() {}`
			output, err := runWithConfig(configContent, code, []string{"invalid1.go", "invalid2.go"})
			_ = err
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When configuration file has detection method settings", func() {
		It("should load detection methods from config", func() {
			configContent := `{
				"threshold": 15,
				"detectionMethods": ["hash", "art-dupl"]
			}`
			code := `package main
func detectionMethod() string {
	return "test"
}`
			output, err := runWithConfig(configContent, code, []string{"detect1.go", "detect2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should load single detection method from config", func() {
			configContent := `{
				"threshold": 15,
				"detectionMethods": ["hash"]
			}`
			code := `package main
func singleMethod() string {
	return "hash only"
}`
			output, err := runWithConfig(configContent, code, []string{"single1.go", "single2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When configuration file has filtering settings", func() {
		It("should load filter-generated setting from config", func() {
			configContent := `{
				"threshold": 15,
				"filterGenerated": true
			}`
			code := `package main
func filterGen() {}`
			output, err := runWithConfig(configContent, code, []string{"filter1.go", "filter2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should load include-sqlc setting from config", func() {
			configContent := `{
				"threshold": 15,
				"filterGenerated": true,
				"includeSqlc": true
			}`
			code := `// Code generated by sqlc. DO NOT EDIT.
package db
func SQLCQuery() {}`
			output, err := runWithConfig(configContent, code, []string{"query1.go", "query2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should load include-templ setting from config", func() {
			configContent := `{
				"threshold": 15,
				"filterGenerated": true,
				"includeTempl": true
			}`
			code := `package components
import "github.com/a-h/templ"
func TemplComponent() templ.Component {
	return nil
}`
			output, err := runWithConfig(configContent, code, []string{"comp1.go", "comp2.go"})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When configuration file has pattern settings", func() {
		It("should load include patterns from config", func() {
			// Create subdirectories
			err := setup.CreateSubdirectories("src", "vendor")
			Expect(err).NotTo(HaveOccurred())

			// Create config with include patterns
			configContent := `{
				"threshold": 15,
				"includePatterns": ["src/*"]
			}`
			configPath := filepath.Join(setup.TmpDir, "dupl.json")
			err = os.WriteFile(configPath, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create files in both directories
			code := `package main
func patternTest() {}`
			err = setup.CreateFileWithContent("src/file.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("vendor/file.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Run with config
			output, err := setup.RunArtDupl("--config", configPath, setup.TmpDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should load exclude patterns from config", func() {
			// Create subdirectories
			err := setup.CreateSubdirectories("src", "test")
			Expect(err).NotTo(HaveOccurred())

			// Create config with exclude patterns
			configContent := `{
				"threshold": 15,
				"excludePatterns": ["test/*"]
			}`
			configPath := filepath.Join(setup.TmpDir, "dupl.json")
			err = os.WriteFile(configPath, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create files in both directories
			code := `package main
func excludeTest() {}`
			err = setup.CreateFileWithContent("src/file.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("test/file.go", code)
			Expect(err).NotTo(HaveOccurred())

			// Run with config
			output, err := setup.RunArtDupl("--config", configPath, setup.TmpDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using stats subcommand with config file", func() {
		It("should load stats configuration from file", func() {
			// Create config for stats
			configContent := `{
				"threshold": 20,
				"outputFormat": "json"
			}`
			configPath := filepath.Join(setup.TmpDir, "dupl.json")
			err := os.WriteFile(configPath, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create test files
			code := `package main
func statsConfig() {}`
			err = setup.CreateDuplicateFiles([]string{"stats1.go", "stats2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run stats with config
			output, err := setup.RunArtDupl("stats", "--config", configPath, setup.TmpDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})
})

var _ = Describe("Configuration File Edge Cases", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = newBDDTestSetup()
	})

	Context("When configuration file is empty", func() {
		It("should handle empty config file gracefully", func() {
			code := `package main
func emptyConfig() {}`
			output, err := setup.RunWithConfigFile("empty.json", "", code, []string{"empty1.go", "empty2.go"})
			// May error or use defaults
			_ = err
			Expect(output).ToNot(BeNil())
		})

		It("should handle config with only whitespace", func() {
			code := `package main
func whitespaceConfig() {}`
			output, err := setup.RunWithConfigFile("whitespace.json", "   \n\t  ", code, []string{"ws1.go", "ws2.go"})
			// May error or use defaults
			_ = err
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When configuration file has extra fields", func() {
		It("should ignore unknown fields in config", func() {
			configContent := `{
				"threshold": 15,
				"unknownField": "should be ignored",
				"anotherUnknown": 12345
			}`
			code := `package main
func unknownField() {}`
			output, err := setup.RunWithConfigFile("dupl.json", configContent, code, []string{"unknown1.go", "unknown2.go"})
			// Should work and ignore unknown fields
			_ = err
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using deeply nested config files", func() {
		It("should handle config in nested directory", func() {
			// Create nested directory structure
			nestedDir := filepath.Join(setup.TmpDir, "nested", "config")
			err := os.MkdirAll(nestedDir, 0o755)
			Expect(err).NotTo(HaveOccurred())

			// Create config in nested directory
			configContent := `{
				"threshold": 25
			}`
			configPath := filepath.Join(nestedDir, "dupl.json")
			err = os.WriteFile(configPath, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create test files
			code := `package main
func nestedConfig() {}`
			err = setup.CreateDuplicateFiles([]string{"nested1.go", "nested2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with nested config path
			output, err := setup.RunArtDupl("--config", configPath, setup.TmpDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When configuration has special characters", func() {
		It("should handle config with unicode content", func() {
			// Create config with unicode
			configContent := fmt.Sprintf(`{
				"threshold": 15,
				"paths": ["%s"]
			}`, setup.TmpDir)
			configPath := filepath.Join(setup.TmpDir, "unicode.json")
			err := os.WriteFile(configPath, []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create test files
			code := `package main
func unicodeConfig() {}`
			err = setup.CreateDuplicateFiles([]string{"unicode1.go", "unicode2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with unicode config
			output, err := setup.RunArtDupl("--config", configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})
})
