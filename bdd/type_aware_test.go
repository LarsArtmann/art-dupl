package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

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

	Context("when using --type-aware alone", func() {
		It("should run without errors", func() {
			code := `package main

import "fmt"

func processA(items []string) {
	for _, item := range items {
		fmt.Println(item)
		fmt.Println(item)
		fmt.Println(item)
	}
}

func processB(items []string) {
	for _, item := range items {
		fmt.Println(item)
		fmt.Println(item)
		fmt.Println(item)
	}
}`
			err := setup.CreateDuplicateFiles(
				[]string{"handler1.go", "handler2.go"},
				code,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--type-aware", "--threshold", "5")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("handler1.go"))
		})
	})

	Context("when using //art-dupl:accept directive", func() {
		It("should suppress clone groups with the directive", func() {
			codeWithAccept := `package main

import "fmt"

func processA(items []string) {
	//art-dupl:accept
	for _, item := range items {
		fmt.Println(item)
		fmt.Println(item)
		fmt.Println(item)
	}
}`

			codeWithoutAccept := `package main

import "fmt"

func processB(items []string) {
	for _, item := range items {
		fmt.Println(item)
		fmt.Println(item)
		fmt.Println(item)
	}
}`

			err := setup.CreateTestFile("with_accept.go", codeWithAccept)
			Expect(err).NotTo(HaveOccurred())

			err = setup.CreateTestFile("without_accept.go", codeWithoutAccept)
			Expect(err).NotTo(HaveOccurred())

			// With accept directives active, the group should be suppressed
			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).NotTo(HaveOccurred())

			// The clone group should NOT appear because the accept directive suppresses it
			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("without_accept.go"))
		})

		It("should show all clones with --no-accept-directives", func() {
			codeWithAccept := `package main

import "fmt"

func processA(items []string) {
	//art-dupl:accept
	for _, item := range items {
		fmt.Println(item)
		fmt.Println(item)
		fmt.Println(item)
	}
}`

			codeWithoutAccept := `package main

import "fmt"

func processB(items []string) {
	for _, item := range items {
		fmt.Println(item)
		fmt.Println(item)
		fmt.Println(item)
	}
}`

			err := setup.CreateTestFile("with_accept2.go", codeWithAccept)
			Expect(err).NotTo(HaveOccurred())

			err = setup.CreateTestFile("without_accept2.go", codeWithoutAccept)
			Expect(err).NotTo(HaveOccurred())

			// With --no-accept-directives, the accept directive should be ignored
			output, err := setup.RunArtDupl("--threshold", "5", "--no-accept-directives")
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("with_accept2.go"))
			Expect(outputStr).To(ContainSubstring("without_accept2.go"))
		})
	})
})
