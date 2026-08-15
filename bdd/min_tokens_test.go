package bdd

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Min-Tokens Filtering", func() {
	It("should report duplicate clones without --min-tokens", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"mintokens1.go", "mintokens2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("found"))
	})

	It("should suppress clone groups when --min-tokens exceeds every clone", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"mintokens1.go", "mintokens2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--min-tokens", "100000")
		Expect(err).ToNot(HaveOccurred())

		outputStr := strings.ToLower(string(output))
		Expect(outputStr).To(ContainSubstring("0 shown"))
		Expect(outputStr).To(ContainSubstring("filtered"))
	})

	It("should not suppress clone groups when --min-tokens is below the clone size", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"mintokens1.go", "mintokens2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--min-tokens", "2")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("found"))
		Expect(strings.ToLower(string(output))).ToNot(ContainSubstring("0 shown"))
	})

	It("should reject negative --min-tokens values", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"mintokens1.go", "mintokens2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--min-tokens", "-5")
		Expect(err).To(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("min-tokens"))
	})
})
