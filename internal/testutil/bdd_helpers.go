package testutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateSubdirectories creates multiple directories in test temporary directory.
// Each directory name is a relative path that will be created under the temp directory.
func (s *BDDTestSetup) CreateSubdirectories(paths ...string) error {
	if s.T != nil {
		s.T.Helper()
	}

	for _, path := range paths {
		fullPath := filepath.Join(s.TmpDir, path)
		if err := os.MkdirAll(fullPath, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// CreateFileWithContent creates a file with specific content at a given subpath.
// The subpath is relative to the test temporary directory.
func (s *BDDTestSetup) CreateFileWithContent(subpath, content string) error {
	if s.T != nil {
		s.T.Helper()
	}

	fullPath := filepath.Join(s.TmpDir, subpath)
	dir := filepath.Dir(fullPath)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", subpath, err)
	}
	return nil
}

// CommonDuplicateCodeTemplate is a reusable code template for creating duplicate test files.
// Use with CreateDuplicateFilesAndRun to reduce test duplication.
const CommonDuplicateCodeTemplate = `package main

import "fmt"

func %s() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`

// CreateDuplicateFilesAndRun creates duplicate files with the given content and runs art-dupl.
// Returns the command output for assertions. This helper reduces boilerplate in BDD tests.
func (s *BDDTestSetup) CreateDuplicateFilesAndRun(content string, args ...string) ([]byte, error) {
	return s.CreateNamedDuplicateFilesAndRun([]string{"file1.go", "file2.go"}, content, args...)
}

// CreateNamedDuplicateFilesAndRun creates duplicate files with specific names and content, then runs art-dupl.
// Returns the command output for assertions. This helper reduces boilerplate in BDD tests.
func (s *BDDTestSetup) CreateNamedDuplicateFilesAndRun(filenames []string, content string, args ...string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	if err := s.CreateDuplicateFiles(filenames, content); err != nil {
		return nil, err
	}

	return s.RunArtDupl(args...)
}

// CreateAndRunDupl creates duplicate files with the given content and runs art-dupl with specified arguments.
// This is a convenience helper that combines CreateDuplicateFiles and RunArtDupl.
// Returns the command output and any error that occurred.
func (s *BDDTestSetup) CreateAndRunDupl(filenames []string, content string, args ...string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	if err := s.CreateDuplicateFiles(filenames, content); err != nil {
		return nil, fmt.Errorf("failed to create duplicate files: %w", err)
	}

	return s.RunArtDupl(args...)
}

// RunWithConfigFile creates a config file with custom filename, test files, and runs art-dupl.
// This is a convenience helper for config file tests that combines file creation and execution.
// Returns the command output and any error that occurred.
func (s *BDDTestSetup) RunWithConfigFile(configFileName, configContent, code string, fileNames []string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	configPath := filepath.Join(s.TmpDir, configFileName)
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	if err := s.CreateDuplicateFiles(fileNames, code); err != nil {
		return nil, fmt.Errorf("failed to create duplicate files: %w", err)
	}

	return s.RunArtDupl("--config", configPath, s.TmpDir)
}

// Common test thresholds for BDD tests.
const (
	// ThresholdSmall is used for tests with minimal code (5 tokens).
	ThresholdSmall = "5"
	// ThresholdMedium is used for tests with moderate code (10 tokens).
	ThresholdMedium = "10"
	// ThresholdLarge is used for tests with larger code (20 tokens).
	ThresholdLarge = "20"
)

// SimpleCodeTemplate generates a simple Go code template with a unique function name.
// This is useful for creating test files that need distinct function names to avoid
// false positives across different test cases.
func SimpleCodeTemplate(funcName string) string {
	return fmt.Sprintf(`package main
func %s() {}`, funcName)
}

// VendorTestCode is a reusable code sample for vendor directory filtering tests.
// It contains enough tokens to be detected as a duplicate while being simple and consistent.
const VendorTestCode = `package main

import "fmt"

func vendorFunc() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`

// SimpleVendorTestCode is a minimal code sample for vendor directory filtering tests.
// Use when a smaller code sample is sufficient (threshold of 3-5 tokens).
const SimpleVendorTestCode = `package main
func vendorFunc() { println(1) }`

// CreateVendorDuplicateFiles creates duplicate files in a vendor directory with the given vendor path and code content.
// This is a convenience helper for testing vendor directory filtering behavior.
// Returns an error if directory creation or file writing fails.
func (s *BDDTestSetup) CreateVendorDuplicateFiles(vendorPath, code string) error {
	if s.T != nil {
		s.T.Helper()
	}

	// Create the vendor directory structure
	if err := s.CreateSubdirectories(vendorPath); err != nil {
		return fmt.Errorf("failed to create vendor directory %s: %w", vendorPath, err)
	}

	// Create duplicate files in the vendor directory
	lib1Path := filepath.Join(vendorPath, "lib1.go")
	lib2Path := filepath.Join(vendorPath, "lib2.go")

	if err := s.CreateFileWithContent(lib1Path, code); err != nil {
		return fmt.Errorf("failed to create vendor file %s: %w", lib1Path, err)
	}

	if err := s.CreateFileWithContent(lib2Path, code); err != nil {
		return fmt.Errorf("failed to create vendor file %s: %w", lib2Path, err)
	}

	return nil
}

// CreateAndRunDuplExpectSuccess creates duplicate files, runs art-dupl, and asserts success.
// This helper reduces boilerplate by combining CreateAndRunDupl with common assertions.
// Returns the output for further assertions. Panics on error (suitable for Ginkgo tests).
func (s *BDDTestSetup) CreateAndRunDuplExpectSuccess(filenames []string, content string, args ...string) []byte {
	if s.T != nil {
		s.T.Helper()
	}

	output, err := s.CreateAndRunDupl(filenames, content, args...)
	if err != nil {
		if s.T != nil {
			s.T.Fatalf("art-dupl command failed: %v\nOutput: %s", err, string(output))
		}
		panic(fmt.Sprintf("art-dupl command failed: %v\nOutput: %s", err, string(output)))
	}
	if output == nil {
		if s.T != nil {
			s.T.Fatal("art-dupl output is nil")
		}
		panic("art-dupl output is nil")
	}

	return output
}

// RunVendorTest runs a vendor directory exclusion/inclusion test with the given configuration.
// This helper reduces duplication across BDD tests that verify vendor filtering behavior.
// Parameters:
//   - includeVendor: if true, runs with --vendor flag to include vendor directory
//   - subcommand: optional subcommand to run (e.g., "stats", "" for default)
//   - extraArgs: additional arguments to pass to art-dupl
//
// Returns the command output and any error that occurred.
func (s *BDDTestSetup) RunVendorTest(includeVendor bool, subcommand string, extraArgs ...string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	// Create vendor directory with duplicate files
	code := fmt.Sprintf(CommonDuplicateCodeTemplate, "vendorTestFunc")
	if err := s.CreateVendorDuplicateFiles("vendor/example", code); err != nil {
		return nil, fmt.Errorf("failed to create vendor duplicate files: %w", err)
	}

	// Build arguments
	args := []string{}
	if subcommand != "" {
		args = append(args, subcommand)
	}
	if includeVendor {
		args = append(args, "--vendor")
	}
	args = append(args, extraArgs...)

	return s.RunArtDupl(args...)
}

// RunVendorTestWithOptions runs a vendor directory test with full customization.
// This helper reduces duplication across BDD tests that verify vendor filtering behavior.
// Parameters:
//   - vendorPath: path to vendor directory (e.g., "vendor/example")
//   - code: the code content to use for duplicate files
//   - includeVendor: if true, runs with --vendor flag to include vendor directory
//   - subcommand: optional subcommand to run (e.g., "stats", "" for default)
//   - extraArgs: additional arguments to pass to art-dupl
//
// Returns the command output and any error that occurred.
func (s *BDDTestSetup) RunVendorTestWithOptions(vendorPath, code string, includeVendor bool, subcommand string, extraArgs ...string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	// Create vendor directory with duplicate files
	if err := s.CreateVendorDuplicateFiles(vendorPath, code); err != nil {
		return nil, fmt.Errorf("failed to create vendor duplicate files: %w", err)
	}

	// Build arguments
	args := []string{}
	if subcommand != "" {
		args = append(args, subcommand)
	}
	if includeVendor {
		args = append(args, "--vendor")
	}
	args = append(args, extraArgs...)

	// Use RunSubcommand when there's a subcommand (puts path at end), otherwise RunArtDupl
	if subcommand != "" {
		return s.RunSubcommand(args...)
	}
	return s.RunArtDupl(args...)
}
