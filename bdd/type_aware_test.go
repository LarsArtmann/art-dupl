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
		It("should work without warning (compatible since M07)", func() {
			err := setup.CreateTestFile("dummy.go", "package main\n")
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--type-aware", "--incremental", "--threshold", "5")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).NotTo(ContainSubstring("not compatible"))
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
		It("should suppress accepted groups entirely", func() {
			err := setup.CreateTestFile("accepted.go", dupCodeWithAccept)
			Expect(err).NotTo(HaveOccurred())

			err = setup.CreateTestFile("other.go", dupCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			// The accept directive suppresses the entire clone group —
			// both the accepted clone and its duplicate partner disappear.
			Expect(outputStr).To(ContainSubstring("Found total 0 clone groups"))
			Expect(outputStr).NotTo(ContainSubstring("accepted.go"))
			Expect(outputStr).NotTo(ContainSubstring("other.go"))
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

	Context("interface-method suppression with --type-aware", func() {
		// Multi-statement bodies (2 statements each) avoid being caught by
		// single-simple-statement (#11), so the interface-method pattern (#20)
		// is the one that fires.
		const ifaceSrc = `package main

type Validator interface {
	Validate(x int) error
}
`
		const impl1Src = `package main

import "errors"

type Foo struct{}

func (f Foo) Validate(x int) error {
	if x == 0 {
		return errors.New("invalid")
	}
	return nil
}
`
		const impl2Src = `package main

import "errors"

type Bar struct{}

func (b Bar) Validate(x int) error {
	if x == 0 {
		return errors.New("invalid")
	}
	return nil
}
`
		const mainSrc = `package main

func main() {
	var _ Validator = Foo{}
	var _ Validator = Bar{}
}
`

		BeforeEach(func() {
			err := setup.CreateTestFiles(map[string]string{
				"iface.go": ifaceSrc,
				"impl1.go": impl1Src,
				"impl2.go": impl2Src,
				"main.go":  mainSrc,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should suppress interface-method clones with --type-aware", func() {
			output, err := setup.RunArtDupl("--type-aware", "--threshold", "2")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("Found total 0 clone groups"))
		})

		It("should classify suppressed clones as interface-method with --explain", func() {
			output, err := setup.RunArtDupl(
				"--type-aware", "--threshold", "2",
				"--no-actionability", "--explain",
			)
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("impl1.go"))
			Expect(outputStr).To(ContainSubstring("interface-method"))
		})

		It("should not suppress without --type-aware (no InterfaceMethod flag)", func() {
			output, err := setup.RunArtDupl("--threshold", "2")
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("impl1.go"))
			Expect(outputStr).NotTo(ContainSubstring("Found total 0 clone groups"))
		})
	})
})
