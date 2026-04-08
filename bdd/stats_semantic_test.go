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

	Context("When running stats with --semantic flag", func() {
		It("should show detection mode as semantic in text output", func() {
			code := `package main

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"sem1.go", "sem2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand("stats", "--semantic", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Detection Mode"))
			Expect(outputStr).To(ContainSubstring("semantic"))
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

			err := setup.CreateDuplicateFiles([]string{"desc1.go", "desc2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand("stats", "--semantic", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("identifier names"))
		})

		It("should recommend --structural in semantic mode", func() {
			code := `package main

func semanticRec() {}`

			err := setup.CreateDuplicateFiles([]string{"recom1.go", "recom2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand("stats", "--semantic", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("--structural"))
		})
	})

	Context("When running stats with --structural flag", func() {
		It("should show detection mode as structural in text output", func() {
			code := `package main

import "fmt"

func processData(data string) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}`

			err := setup.CreateDuplicateFiles([]string{"struct1.go", "struct2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand("stats", "--structural", "--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Detection Mode"))
			Expect(outputStr).To(ContainSubstring("structural"))
		})

		It("should include detection mode description in text output", func() {
			code := `package main

func structDesc() {}`

			err := setup.CreateDuplicateFiles([]string{"sd1.go", "sd2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand("stats", "--structural", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("AST structure only"))
		})

		It("should recommend --semantic in structural mode", func() {
			code := `package main

func structRec() {}`

			err := setup.CreateDuplicateFiles([]string{"srec1.go", "srec2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand("stats", "--structural", "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("--semantic"))
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

			err := setup.CreateDuplicateFiles([]string{"jsonsem1.go", "jsonsem2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--semantic", "--format", "json", "--threshold", "10",
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
			code := `package main

func jsonBoolTest() {}`

			err := setup.CreateDuplicateFiles([]string{"jb1.go", "jb2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--semantic", "--format", "json", "--threshold", "5",
			)
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			config := result["configuration"].(map[string]any)
			Expect(config).To(HaveKeyWithValue("semanticDetection", true))
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

			err := setup.CreateDuplicateFiles([]string{"jsonst1.go", "jsonst2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--structural", "--format", "json", "--threshold", "10",
			)
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			config := result["configuration"].(map[string]any)
			Expect(config).To(HaveKeyWithValue("detectionMode", "structural"))
		})

		It("should set semanticDetection to false", func() {
			code := `package main

func jsonBoolStruct() {}`

			err := setup.CreateDuplicateFiles([]string{"jbs1.go", "jbs2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--structural", "--format", "json", "--threshold", "5",
			)
			Expect(err).ToNot(HaveOccurred())

			var result map[string]any
			err = json.Unmarshal(output, &result)
			Expect(err).ToNot(HaveOccurred())

			config := result["configuration"].(map[string]any)
			Expect(config).To(HaveKeyWithValue("semanticDetection", false))
		})
	})

	Context("When running stats --semantic with CSV format", func() {
		It("should include Detection Mode in CSV output", func() {
			code := `package main

func csvSemTest() {}`

			err := setup.CreateDuplicateFiles([]string{"csvs1.go", "csvs2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--semantic", "--format", "csv", "--threshold", "5",
			)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Detection Mode,semantic"))
		})

		It("should include Detection Mode Description in CSV output", func() {
			code := `package main

func csvSemDesc() {}`

			err := setup.CreateDuplicateFiles([]string{"csvsd1.go", "csvsd2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--semantic", "--format", "csv", "--threshold", "5",
			)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Detection Mode Description,"))
		})
	})

	Context("When running stats --structural with CSV format", func() {
		It("should include Detection Mode as structural in CSV output", func() {
			code := `package main

func csvStructMode() {}`

			err := setup.CreateDuplicateFiles([]string{"csvsm1.go", "csvsm2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunSubcommand(
				"stats", "--structural", "--format", "csv", "--threshold", "5",
			)
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Detection Mode,structural"))
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
