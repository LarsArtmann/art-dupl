package bdd

import (
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Semantic-Aware Detection
//
// These tests verify the semantic-aware duplicate detection feature,
// which reduces false positives by including identifier names in matching.
//
// The scenarios cover:
// - Semantic-aware detection (default) prevents false positives
// - Structural-only matching with --structural flag
// - Config file support for semantic detection
// - Default behavior: semantic detection OFF (opt-in with --semantic)

var _ = Describe("Semantic Detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When semantic detection is disabled (--structural flag)", func() {
		It("should detect structural duplicates even with different method names", func() {
			// Create two files with identical AST structure but different method names
			// These should be flagged as duplicates WITH --structural
			err := setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "user_test.go"), []byte(structuralTestCode1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "order_test.go"), []byte(structuralTestCode2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITH --structural flag (structural-only matching)
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--structural")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should detect structural duplicate
			Expect(outputStr).To(ContainSubstring("user_test.go"))
			Expect(outputStr).To(ContainSubstring("order_test.go"))
		})
	})

	Context("When semantic detection is enabled", func() {
		It("should NOT flag structurally identical code with different method names", func() {
			// Create two files with identical AST structure but ALL different identifiers
			// These should NOT be flagged as duplicates WITH --semantic flag
			err := setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "user_test.go"), []byte(semanticDifferentCode1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "order_test.go"), []byte(semanticDifferentCode2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITH --semantic flag
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--semantic")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should NOT detect as duplicate because ALL identifiers differ
			Expect(outputStr).ToNot(ContainSubstring("user_test.go"))
			Expect(outputStr).ToNot(ContainSubstring("order_test.go"))
		})

		It("should still detect true duplicates when identifiers match", func() {
			// Create two files with identical code (true duplicates)
			err := setup.FileProcessor.WriteDuplicateFiles([]string{"test1_test.go", "test2_test.go"}, trueDuplicateCode)
			Expect(err).NotTo(HaveOccurred())

			// Run WITH --semantic flag
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--semantic")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should still detect true duplicates
			Expect(outputStr).To(ContainSubstring("test1_test.go"))
			Expect(outputStr).To(ContainSubstring("test2_test.go"))
		})
	})

	Context("When using config file with semantic setting", func() {
		It("should respect semantic: true in config file", func() {
			// Create test files with ALL different identifiers
			err := setup.FileProcessor.WriteFile("user.go", []byte(configTestDifferentCode1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile("order.go", []byte(configTestDifferentCode2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create config file with semantic: true
			configContent := `{
				"semantic": true,
				"threshold": 5
			}`
			configPath := filepath.Join(setup.TmpDir, "dupl.json")
			err = setup.FileProcessor.WriteFile("dupl.json", []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run with config file
			cmd := exec.Command(setup.BinaryPath, "-c", configPath, setup.TmpDir)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)
			Expect(err).ToNot(HaveOccurred())

			// Should NOT flag as duplicates due to semantic detection (ALL identifiers differ)
			Expect(outputStr).ToNot(ContainSubstring("user.go"))
			Expect(outputStr).ToNot(ContainSubstring("order.go"))
		})
	})

	Context("When analyzing Ginkgo test patterns", func() {
		It("should distinguish between different handler tests with semantic detection", func() {
			// This is the main use case: Ginkgo test blocks for different handlers
			// have identical structure but test different things
			err := setup.FileProcessor.WriteFile("user_handler_test.go", []byte(handlerTestCode1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile("order_handler_test.go", []byte(handlerTestCode2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITH --structural (opt-out from semantic): should detect as duplicate
			cmdStructural := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "15", "--structural")
			outputStructural, err := cmdStructural.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			structuralStr := string(outputStructural)
			Expect(structuralStr).To(ContainSubstring("user_handler_test.go"))

			// Run WITH --semantic: should NOT detect as duplicate
			cmdDefault := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "15", "--semantic")
			outputDefault, err := cmdDefault.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			defaultStr := string(outputDefault)
			// Different handler types should not match (NewUserHandler vs NewOrderHandler, GetUser vs GetOrder)
			Expect(defaultStr).ToNot(ContainSubstring("user_handler_test.go"))
			Expect(defaultStr).ToNot(ContainSubstring("order_handler_test.go"))
		})
	})

	Context("When enum pattern methods have same structure but different receiver types", func() {
		It("should NOT flag methods on different types as duplicates by default (semantic)", func() {
			// This is the exact pattern from auto-deduplicate that caused false positives
			err := setup.FileProcessor.WriteFile("crush_mode.go", []byte(enumPatternCode1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile("safety_mode.go", []byte(enumPatternCode2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITH --structural (opt-out from semantic): should detect as duplicate
			cmdStructural := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--structural")
			outputStructural, err := cmdStructural.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			structuralStr := string(outputStructural)
			Expect(structuralStr).To(ContainSubstring("crush_mode.go"))
			Expect(structuralStr).To(ContainSubstring("safety_mode.go"))

			// Run WITH --semantic: should NOT detect as duplicate
			// Because receiver types (CrushMode vs SafetyMode) and function names (ParseCrushMode vs ParseSafetyMode) differ
			cmdDefault := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--semantic")
			outputDefault, err := cmdDefault.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			defaultStr := string(outputDefault)
			Expect(defaultStr).ToNot(ContainSubstring("crush_mode.go"))
			Expect(defaultStr).ToNot(ContainSubstring("safety_mode.go"))
		})
	})
})
