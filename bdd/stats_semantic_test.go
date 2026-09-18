package bdd

import (
	"encoding/json"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stats Semantic Detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	runStatsAndExpectSubstring := func(mode, threshold string, expected ...string) {
		files := []string{goldenTestFile1, testFile2}
		code := `package main

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`
		err := setup.CreateDuplicateFiles(files, code)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunSubcommand("stats", "--"+mode, "--threshold", threshold)
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)
		for _, exp := range expected {
			Expect(outputStr).To(ContainSubstring(exp))
		}
	}

	runStatsMinimal := func(mode, filePrefix, expectedSubstring string) {
		files := []string{filePrefix + "1.go", filePrefix + "2.go"}
		code := `package main

func minimalFunc() {}`
		err := setup.CreateDuplicateFiles(files, code)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunSubcommand("stats", "--"+mode, "--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring(expectedSubstring))
	}

	runStatsJSON := func(mode, expectedMode string, expectedSemantic bool) {
		code := `package main

import "fmt"

func jsonTest(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	fmt.Println(name)
	return nil
}`
		files := []string{"jsontest1.go", "jsontest2.go"}
		err := setup.CreateDuplicateFiles(files, code)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunSubcommandOutput(
			"stats", "--"+mode, "--format", "json", "--threshold", "1",
		)
		Expect(err).ToNot(HaveOccurred())

		var result map[string]any

		err = json.Unmarshal(output, &result)
		Expect(err).ToNot(HaveOccurred())

		config := result["configuration"].(map[string]any)
		Expect(config).To(HaveKeyWithValue("detectionMode", expectedMode))

		if expectedMode == "semantic" {
			Expect(config).To(HaveKeyWithValue("semanticDetection", true))
		} else {
			Expect(config).To(HaveKeyWithValue("semanticDetection", expectedSemantic))
		}
	}

	runStatsCSV := func(mode, expectedSubstring string) {
		code := `package main

func csvTest() {}`
		files := []string{"csvtest1.go", "csvtest2.go"}
		err := setup.CreateDuplicateFiles(files, code)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunSubcommand(
			"stats", "--"+mode, "--format", "csv", "--threshold", "1",
		)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring(expectedSubstring))
	}

	Context("When running stats with --semantic flag", func() {
		It("should show detection mode as semantic in text output", func() {
			runStatsAndExpectSubstring("semantic", "10", "Detection Mode", "semantic")
		})

		It("should include detection mode description in text output", func() {
			code := `package main

import "fmt"

func handleRequest(req string) error {
	if req == "" {
		return fmt.Errorf("empty request")
	}
	return nil
}`
			files := []string{"desc1.go", "desc2.go"}
			err := setup.CreateDuplicateFiles(files, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand("stats", "--semantic", "--threshold", "1")
			Expect(err).ToNot(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("identifier names"))
		})

		It("should recommend --structural in semantic mode", func() {
			runStatsMinimal("semantic", "rec", "--structural")
		})
	})

	Context("When running stats with --structural flag", func() {
		It("should show detection mode as structural in text output", func() {
			runStatsAndExpectSubstring("structural", "10", "Detection Mode", "structural")
		})

		It("should include detection mode description in text output", func() {
			runStatsMinimal("structural", "desc", "AST structure only")
		})

		It("should recommend --semantic in structural mode", func() {
			runStatsMinimal("structural", "srec", "--semantic")
		})
	})

	Context("When running stats --semantic with JSON format", func() {
		It("should include detectionMode in JSON configuration", func() {
			code := `package main

import "fmt"

func jsonSemanticTest(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	fmt.Println(name)
	return nil
}`
			files := []string{"jsonsem1.go", "jsonsem2.go"}
			err := setup.CreateDuplicateFiles(files, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommandOutput(
				"stats", "--semantic", "--format", "json", "--threshold", "1",
			)
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any

			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			Expect(result).To(HaveKey("configuration"))
			config := result["configuration"].(map[string]any)
			Expect(config).To(HaveKeyWithValue("detectionMode", "semantic"))
			Expect(config).To(HaveKey("detectionModeDescription"))
			Expect(config["detectionModeDescription"]).ToNot(BeEmpty())
		})

		It("should set semanticDetection to true", func() {
			runStatsJSON("semantic", "semantic", true)
		})
	})

	Context("When running stats --structural with JSON format", func() {
		It("should include detectionMode as structural in JSON", func() {
			code := `package main

import "fmt"

func jsonStructTest(data string) error {
	if data == "" {
		return fmt.Errorf("empty")
	}
	return nil
}`
			files := []string{"jsonst1.go", "jsonst2.go"}
			err := setup.CreateDuplicateFiles(files, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommandOutput(
				"stats", "--structural", "--format", "json", "--threshold", "1",
			)
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any

			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			config := result["configuration"].(map[string]any)
			Expect(config).To(HaveKeyWithValue("detectionMode", "structural"))
		})

		It("should set semanticDetection to false", func() {
			runStatsJSON("structural", "structural", false)
		})
	})

	Context("When running stats --semantic with CSV format", func() {
		It("should include Detection Mode in CSV output", func() {
			runStatsCSV("semantic", "Detection Mode,semantic")
		})

		It("should include Detection Mode Description in CSV output", func() {
			runStatsCSV("semantic", "Detection Mode Description,")
		})
	})

	Context("When running stats --structural with CSV format", func() {
		It("should include Detection Mode as structural in CSV output", func() {
			runStatsCSV("structural", "Detection Mode,structural")
		})
	})

	Context("When running stats without semantic or structural flag", func() {
		It("should not crash in default mode", func() {
			code := testutil.SimpleCodeTemplate("defaultMode")
			err := setup.CreateDuplicateFiles(
				[]string{"default1.go", "default2.go"},
				code,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--threshold", testutil.ThresholdMedium,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).ToNot(BeEmpty())
		})
	})
})
