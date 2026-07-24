package bdd

import (
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// dupCode has 7+ separate top-level statements to exceed threshold 5.
// Each assignment is a separate statement token in the suffix tree.
const dupCode = `package main

func parseHeader(data []byte) (int, int, int) {
	header := data[:4]
	version := int(header[0])
	flags := int(header[1])
	length := int(header[2])<<8 | int(header[3])
	body := data[4 : 4+length]
	tail := data[4+length:]
	return version, flags, len(tail)
}
`

// dupCodeWithAccept adds the //art-dupl:accept directive.
const dupCodeWithAccept = `package main

func parseHeader(data []byte) (int, int, int) {
	//art-dupl:accept
	header := data[:4]
	version := int(header[0])
	flags := int(header[1])
	length := int(header[2])<<8 | int(header[3])
	body := data[4 : 4+length]
	tail := data[4+length:]
	return version, flags, len(tail)
}
`

var _ = Describe("Type-Aware Detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Context("when using --type-aware with --structural", func() {
		It("should error because type-aware is meaningless with structural mode", func() {
			err := setup.CreateTestFile("dummy.go", "package main\n")
			Expect(err).NotTo(HaveOccurred())

			_, err = setup.RunArtDupl("--type-aware", "--structural")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("type-aware"))
		})
	})

	Context("when using --type-aware with --exact", func() {
		It("should error because type-aware is meaningless with exact mode", func() {
			err := setup.CreateTestFile("dummy.go", "package main\n")
			Expect(err).NotTo(HaveOccurred())

			_, err = setup.RunArtDupl("--type-aware", "--exact")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("type-aware"))
		})
	})

	Context("when using --type-aware with --incremental", func() {
		It("should warn that type-aware is not compatible with incremental", func() {
			err := setup.CreateTestFile("dummy.go", "package main\n")
			Expect(err).NotTo(HaveOccurred())

			output, _ := setup.RunArtDupl("--type-aware", "--incremental", "--threshold", "5")
			Expect(string(output)).To(ContainSubstring("type-aware"))
		})
	})

	Context("when using --type-aware alone on valid Go code", func() {
		It("should not error", func() {
			err := setup.CreateTestFile("main.go", "package main\n\nfunc main() {}\n")
			Expect(err).NotTo(HaveOccurred())

			_, err = setup.RunArtDupl("--type-aware", "--threshold", "5")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when using --semantic flag explicitly", func() {
		It("should print a deprecation notice", func() {
			err := setup.CreateTestFile("dummy.go", "package main\n")
			Expect(err).NotTo(HaveOccurred())

			output, _ := setup.RunArtDupl("--semantic", "--threshold", "5")
			Expect(string(output)).To(ContainSubstring("default detection mode"))
		})
	})

	Context("when using //art-dupl:accept directive", func() {
		It("should suppress accepted groups and show others", func() {
			err := setup.CreateTestFile("accepted.go", dupCodeWithAccept)
			Expect(err).NotTo(HaveOccurred())

			err = setup.CreateTestFile("other.go", dupCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("other.go"))
		})

		It("should show all clones with --no-accept-directives", func() {
			err := setup.CreateTestFile("accepted2.go", dupCodeWithAccept)
			Expect(err).NotTo(HaveOccurred())

			err = setup.CreateTestFile("other2.go", dupCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "5", "--no-accept-directives")
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("accepted2.go"))
			Expect(outputStr).To(ContainSubstring("other2.go"))
		})
	})
})
