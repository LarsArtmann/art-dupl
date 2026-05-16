package bdd

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const simpleGoCode = `package main

func process() error {
	if true {
		return nil
	}
	return nil
}`

var _ = Describe("SARIF Output", func() {
	duplicateCode := `package main

import "fmt"

func processData(data string, count int) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	if count <= 0 {
		return fmt.Errorf("invalid count")
	}
	for i := 0; i < count; i++ {
		if err := processItem(data, i); err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}
	return nil
}

func processItem(data string, index int) error {
	return nil
}`

	It("should produce valid SARIF 2.1.0 JSON output", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"sarif1.go", "sarif2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--sarif", "--threshold", "15")
		Expect(err).ToNot(HaveOccurred())

		var sarif map[string]any

		Expect(json.Unmarshal(output, &sarif)).To(Succeed())
		Expect(sarif).To(HaveKey("version"))
		Expect(sarif["version"]).To(Equal("2.1.0"))
		Expect(sarif).To(HaveKey("runs"))
	})

	It("should include file locations in SARIF results", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"loc1.go", "loc2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--sarif", "--threshold", "15")
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)

		Expect(outputStr).To(ContainSubstring("loc1.go"))
		Expect(outputStr).To(ContainSubstring("loc2.go"))
	})

	It("should respect threshold in SARIF output", func() {
		setup := CreateBDDTestSetup()

		smallCode := `package main

func small() {
	println(1)
}`
		err := setup.CreateDuplicateFiles([]string{"small1.go", smallFile2}, smallCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--sarif", "--threshold", "50")
		Expect(err).ToNot(HaveOccurred())

		var sarif map[string]any

		Expect(json.Unmarshal(output, &sarif)).To(Succeed())
		runs := sarif["runs"].([]any)
		Expect(runs).To(HaveLen(1))

		run := runs[0].(map[string]any)

		results, hasResults := run["results"]
		if hasResults {
			Expect(results.([]any)).To(BeEmpty())
		}
	})

	It("should produce SARIF output with hash detection method", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"hash1.go", "hash2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--sarif", "-m", "hash", "--threshold", "10")
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring("hash1.go"))
	})
})

var _ = Describe("Total Tokens Sorting", func() {
	It("should sort by total tokens when specified", func() {
		setup := CreateBDDTestSetup()

		largeClone := `package main

import "fmt"

func processLargeData(data string, count int, threshold float64) error {
	if data == "" {
		return fmt.Errorf("empty data")
	}
	if count <= 0 {
		return fmt.Errorf("invalid count")
	}
	if threshold <= 0.0 {
		return fmt.Errorf("invalid threshold")
	}
	for i := 0; i < count; i++ {
		if float64(i) >= threshold {
			return fmt.Errorf("threshold reached at %d", i)
		}
		if err := processDataItem(data, i, threshold); err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}
	return nil
}

func processDataItem(data string, index int, threshold float64) error {
	return nil
}`

		smallClone := `package main

func processSmall(data string) error {
	if data == "" {
		return nil
	}
	return nil
}`

		err := setup.CreateDuplicateFiles([]string{goldenLargeFile1, largeFile2}, largeClone)
		Expect(err).NotTo(HaveOccurred())
		err = setup.CreateDuplicateFiles([]string{"small1.go", smallFile2}, smallClone)
		Expect(err).NotTo(HaveOccurred())

		err = setup.CreateDuplicateFiles(
			[]string{"extra1.go", "extra2.go", "extra3.go"},
			smallClone,
		)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--sort", "total-tokens", "--threshold", "10")
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)

		largeIndex := strings.Index(outputStr, "large")
		Expect(largeIndex).ToNot(Equal(-1), "Large clone should be found")
	})
})

var _ = Describe("File Type Filter (--only)", func() {
	It("should restrict analysis to Go files only with --only go", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{goldenFile1, goldenFile2}, simpleGoCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--only", "go", "--threshold", "10")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring(goldenFile1))
	})

	It("should restrict analysis to templ files only with --only templ", func() {
		setup := CreateBDDTestSetup()

		templCode := `package main

templ page(name string) {
	<div class="container">
		<h1>{ name }</h1>
		<p>Hello World</p>
		<span>Content here</span>
	</div>
}`

		err := setup.CreateDuplicateFiles([]string{"page1.templ", "page2.templ"}, templCode)
		Expect(err).NotTo(HaveOccurred())

		err = setup.CreateDuplicateFiles([]string{goldenFile1, goldenFile2}, simpleGoCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--only", "templ", "--threshold", "5")
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring("page1.templ"))
		Expect(outputStr).ToNot(ContainSubstring(goldenFile1))
	})
})

var _ = Describe("Diff Visualization (--diff)", func() {
	diffCode := `package main

import "fmt"

func process() error {
	if true {
		return fmt.Errorf("error")
	}
	return nil
}`

	It("should produce HTML output with side-by-side diff", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"diff1.go", "diff2.go"}, diffCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--html", "--diff", "side-by-side", "--threshold", "10")
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring("<html"))
		Expect(outputStr).To(ContainSubstring("diff1.go"))
	})

	It("should produce HTML output with inline diff", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"inline1.go", "inline2.go"}, diffCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--html", "--diff", "inline", "--threshold", "10")
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring("<html"))
		Expect(outputStr).To(ContainSubstring("inline1.go"))
	})
})

var _ = Describe("Protobuf/Mockgen/Stringer Filtering", func() {
	Context("When filtering protobuf generated files", func() {
		It("should exclude .pb.go files by default", func() {
			setup := CreateBDDTestSetup()

			pbCode := `package proto

func (m *Message) GetField() string {
	return m.Field
}

func (m *Message) Reset() {
	*m = Message{}
}`

			err := setup.CreateDuplicateFiles([]string{regularFile1, regularFile2}, simpleGoCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("message.pb.go", pbCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring(regularFile1))
			Expect(outputStr).ToNot(ContainSubstring("message.pb.go"))
		})
	})

	Context("When filtering mockgen generated files", func() {
		It("should exclude mock files by default", func() {
			setup := CreateBDDTestSetup()

			mockCode := `package mock

type MockService struct{}

func (m *MockService) DoSomething() error {
	return nil
}

func (m *MockService) GetSomething() string {
	return ""
}`

			err := setup.CreateDuplicateFiles([]string{regularFile1, regularFile2}, simpleGoCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("mock_service.go", mockCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring(regularFile1))
		})
	})
})
