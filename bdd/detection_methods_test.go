package bdd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Detection Methods", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	// Helper function to create command with detection methods
	createDetectionCmd := func(detectionMethods string) *exec.Cmd {
		return exec.Command(
			setup.BinaryPath,
			setup.TmpDir,
			"--detection-methods",
			detectionMethods,
			"--json",
			"--threshold",
			"5",
		)
	}

	Context("When using hash-based detection", func() {
		It("should detect exact file-level duplicates", func() {
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

			err := setup.CreateDuplicateFiles(
				[]string{"exact1.go", "exact2.go", "exact3.go"},
				identicalCode,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--detection-methods", "hash", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			runOutput := string(output)
			Expect(runOutput).To(ContainSubstring("exact1.go"))
			Expect(runOutput).To(ContainSubstring("exact2.go"))
			Expect(runOutput).To(ContainSubstring("exact3.go"))
		})

		It("should provide JSON output with hash detection statistics", func() {
			code := `package main

func duplicate() string {
	return "duplicate"
}`

			err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			cmd := createDetectionCmd("hash")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any

			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			Expect(result).To(HaveKey("detection_method"))
			Expect(result).To(HaveKey("clone_groups"))
			Expect(result).To(HaveKey("summary"))
			summary := result["summary"].(map[string]any)
			Expect(summary).To(HaveKey("total_clones"))
			Expect(summary).To(HaveKey("total_clone_groups"))
			Expect(result["detection_method"]).To(Equal("hash"))
		})
	})

	Context("When using art-dupl detection", func() {
		It("should detect structural duplicates ignoring literal values", func() {
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

			output, err := setup.RunArtDupl("--detection-methods", "art-dupl", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			runOutput := string(output)
			Expect(runOutput).To(ContainSubstring("user.go"))
			Expect(runOutput).To(ContainSubstring("product.go"))
		})

		It("should be default detection method", func() {
			code := `package main

func test() error {
	if true {
		return nil
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"default1.go", "default2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			runOutput := string(output)
			Expect(runOutput).To(ContainSubstring("default1.go"))
			Expect(runOutput).To(ContainSubstring("default2.go"))
		})
	})

	Context("When using combined detection methods", func() {
		It("should run both hash and art-dupl detection", func() {
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

			output, err := setup.RunArtDupl(
				"--detection-methods",
				"hash,art-dupl",
				"--threshold",
				"5",
			)
			Expect(err).ToNot(HaveOccurred())

			runOutput := string(output)
			Expect(runOutput).To(SatisfyAny(
				ContainSubstring("exact1.go"),
				ContainSubstring("structural1.go"),
			))
		})

		It("should provide comprehensive JSON output for combined detection", func() {
			code := `package main

func test() string {
	return "test"
}`

			err := setup.CreateDuplicateFiles([]string{"combine1.go", "combine2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			cmd := createDetectionCmd("hash,art-dupl")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any

			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			Expect(result).To(HaveKey("clone_groups"))
			Expect(result).To(HaveKey("summary"))
			summary := result["summary"].(map[string]any)
			Expect(summary).To(HaveKey("total_clones"))
			Expect(summary).To(HaveKey("total_clone_groups"))
			Expect(result).To(HaveKey("detection_method"))
			Expect(result["detection_method"]).To(Equal("hash,art-dupl"))
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
			runOutput := string(output)
			Expect(runOutput).ToNot(BeEmpty())
		})
	})

	Context("When using hash-based detection with node_modules", func() {
		var (
			nmDir         string
			generatedCode string
		)

		BeforeEach(func() {
			nmDir = filepath.Join(setup.TmpDir, "node_modules", "somepackage")
			err := os.MkdirAll(nmDir, 0o755)
			Expect(err).NotTo(HaveOccurred())

			generatedCode = `package somepackage

import "fmt"

func NodeModulesFunc() {
	for i := 0; i < 5; i++ {
		fmt.Println(i)
	}
}`

			err = os.WriteFile(
				filepath.Join(nmDir, "file1.go"),
				[]byte(generatedCode),
				0o644,
			)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(
				filepath.Join(nmDir, "file2.go"),
				[]byte(generatedCode),
				0o644,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should exclude node_modules by default", func() {
			// Create regular files with same content (duplicates)
			err := setup.CreateDuplicateFiles(
				[]string{"regular1.go", "regular2.go"},
				generatedCode,
			)
			Expect(err).NotTo(HaveOccurred())

			// Run hash detection in verbose mode
			output, err := setup.RunArtDupl(
				"--detection-methods",
				"hash",
				"--threshold",
				"10",
				"-v",
			)
			Expect(err).ToNot(HaveOccurred())

			runOutput := string(output)
			// Should contain the exclusion message
			Expect(runOutput).To(ContainSubstring("node_modules"))
			// Should detect regular files but not node_modules
			Expect(runOutput).To(ContainSubstring("regular1.go"))
			Expect(runOutput).To(ContainSubstring("regular2.go"))
		})

		It("should include node_modules when --include-node-modules is specified", func() {
			// Run hash detection with include-node-modules flag
			output, err := setup.RunArtDupl(
				"--detection-methods",
				"hash",
				"--threshold",
				"10",
				"--include-node-modules",
			)
			Expect(err).ToNot(HaveOccurred())

			runOutput := string(output)
			// Should detect files in node_modules
			Expect(runOutput).To(ContainSubstring("node_modules"))
		})
	})
})
