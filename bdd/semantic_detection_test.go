package bdd

import (
	"os/exec"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// BDD Test Suite for Semantic-Aware Detection
//
// These tests verify the semantic-aware duplicate detection feature,
// which reduces false positives by including identifier names in matching.

var _ = Describe("Semantic Detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Context("When semantic detection is disabled (--structural flag)", func() {
		It("should detect structural duplicates even with different method names", func() {
			err := setup.FileProcessor.WriteFile(
				filepath.Join(setup.TmpDir, "user_test.go"),
				[]byte(structuralTestCode1),
				0o644,
			)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile(
				filepath.Join(setup.TmpDir, "order_test.go"),
				[]byte(structuralTestCode2),
				0o644,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "10", "--structural")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("user_test.go"))
			Expect(outputStr).To(ContainSubstring("order_test.go"))
		})
	})

	Context("When semantic detection is enabled", func() {
		It("should NOT flag structurally identical code with different method names", func() {
			err := setup.FileProcessor.WriteFile(
				filepath.Join(setup.TmpDir, "user_test.go"),
				[]byte(semanticDifferentCode1),
				0o644,
			)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile(
				filepath.Join(setup.TmpDir, "order_test.go"),
				[]byte(semanticDifferentCode2),
				0o644,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "10", "--semantic")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).ToNot(ContainSubstring("user_test.go"))
			Expect(outputStr).ToNot(ContainSubstring("order_test.go"))
		})

		It("should still detect true duplicates when identifiers match", func() {
			err := setup.FileProcessor.WriteDuplicateFiles(
				[]string{"test1_test.go", "test2_test.go"},
				trueDuplicateCode,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "10", "--semantic")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("test1_test.go"))
			Expect(outputStr).To(ContainSubstring("test2_test.go"))
		})
	})

	Context("When using config file with semantic setting", func() {
		It("should respect semantic: true in config file", func() {
			err := setup.FileProcessor.WriteFile("user.go", []byte(configTestDifferentCode1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile("order.go", []byte(configTestDifferentCode2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			configContent := `{
				"semantic": true,
				"threshold": 5
			}`
			configPath := filepath.Join(setup.TmpDir, "dupl.json")
			err = setup.FileProcessor.WriteFile("dupl.json", []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command(setup.BinaryPath, "-c", configPath, setup.TmpDir)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).ToNot(ContainSubstring("user.go"))
			Expect(outputStr).ToNot(ContainSubstring("order.go"))
		})
	})

	Context("When comparing structural vs semantic detection", func() {
		DescribeTable(
			"should distinguish structural duplicates from semantic duplicates",
			func(file1, file2, code1, code2, threshold string, structExpected, semanticAbsent []string) {
				err := setup.FileProcessor.WriteFile(file1, []byte(code1), 0o644)
				Expect(err).NotTo(HaveOccurred())
				err = setup.FileProcessor.WriteFile(file2, []byte(code2), 0o644)
				Expect(err).NotTo(HaveOccurred())

				if len(structExpected) > 0 {
					structuralOutput, err := setup.RunArtDupl(
						"--threshold",
						threshold,
						"--structural",
					)
					Expect(err).ToNot(HaveOccurred())

					structuralStr := string(structuralOutput)
					for _, expected := range structExpected {
						Expect(structuralStr).To(ContainSubstring(expected))
					}
				}

				if len(semanticAbsent) > 0 {
					semanticOutput, err := setup.RunArtDupl("--threshold", threshold, "--semantic")
					Expect(err).ToNot(HaveOccurred())

					semanticStr := string(semanticOutput)
					for _, absent := range semanticAbsent {
						Expect(semanticStr).ToNot(ContainSubstring(absent))
					}
				}
			},
			Entry(
				"handler tests",
				"user_handler_test.go", "order_handler_test.go",
				handlerTestCode1, handlerTestCode2,
				"15",
				[]string{"user_handler_test.go"},
				[]string{"user_handler_test.go", "order_handler_test.go"},
			),
			Entry(
				"enum pattern methods",
				"crush_mode.go", "safety_mode.go",
				enumPatternCode1, enumPatternCode2,
				"10",
				[]string{"crush_mode.go", "safety_mode.go"},
				[]string{"crush_mode.go", "safety_mode.go"},
			),
		)
	})
})
