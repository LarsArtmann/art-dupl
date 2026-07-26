package bdd

import (
	"encoding/json/v2"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Diff Report", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).To(Succeed())
	})

	Describe("comparing current scan against a baseline", func() {
		Context("when new clones are introduced after baseline", func() {
			It("should report them as new", func() {
				By("creating initial duplicate files")

				err := setup.CreateDuplicateFiles(
					[]string{"original1.go", "original2.go"},
					`package main

func processValue(x int) int {
	result := x * 2
	scaled := result + 100
	final := scaled / 3
	return final
}`,
				)
				Expect(err).NotTo(HaveOccurred())

				By("recording the baseline")

				_, err = setup.Executor(
					"baseline",
					setup.TmpDir,
					"-t",
					"1",
					"--baseline-path",
					filepath.Join(setup.TmpDir, ".art-dupl-baseline.json"),
				)
				Expect(err).NotTo(HaveOccurred())

				By("adding new duplicate files")

				err = setup.CreateDuplicateFiles(
					[]string{"newclone1.go", "newclone2.go"},
					`package main

func computeMetric(input int) int {
	transformed := input * 5
	adjusted := transformed - 50
	normalized := adjusted % 7
	return normalized
}`,
				)
				Expect(err).NotTo(HaveOccurred())

				By("running diff-report")

				output, err := setup.RunArtDupl(
					"--diff-report",
					filepath.Join(setup.TmpDir, ".art-dupl-baseline.json"),
					"-t", "1",
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(output)).To(ContainSubstring("New clones:      1"))
				Expect(string(output)).To(ContainSubstring("Suppressed:      1"))
			})
		})

		Context("when clones are removed after baseline", func() {
			It("should report them as resolved", func() {
				By("creating initial duplicate files")

				err := setup.CreateDuplicateFiles(
					[]string{"keep1.go", "keep2.go"},
					`package main

func keepFunc(x int) int {
	result := x * 2
	scaled := result + 100
	return scaled
}`,
				)
				Expect(err).NotTo(HaveOccurred())

				err = setup.CreateDuplicateFiles(
					[]string{"remove1.go", "remove2.go"},
					`package main

func removeFunc(x int) int {
	value := x * 3
	adjusted := value - 50
	final := adjusted / 7
	return final
}`,
				)
				Expect(err).NotTo(HaveOccurred())

				By("recording the baseline")

				_, err = setup.Executor(
					"baseline",
					setup.TmpDir,
					"-t",
					"1",
					"--baseline-path",
					filepath.Join(setup.TmpDir, ".art-dupl-baseline.json"),
				)
				Expect(err).NotTo(HaveOccurred())

				By("removing the second set of files")
				Expect(os.Remove(filepath.Join(setup.TmpDir, "remove1.go"))).To(Succeed())
				Expect(os.Remove(filepath.Join(setup.TmpDir, "remove2.go"))).To(Succeed())

				By("running diff-report")

				output, err := setup.RunArtDupl(
					"--diff-report",
					filepath.Join(setup.TmpDir, ".art-dupl-baseline.json"),
					"-t", "1",
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(output)).To(ContainSubstring("Resolved:"))
				Expect(string(output)).To(ContainSubstring("remove1.go"))
			})
		})

		Context("when JSON output is requested", func() {
			It("should produce valid JSON with new/suppressed/resolved arrays", func() {
				By("creating baseline with one clone")

				err := setup.CreateDuplicateFiles(
					[]string{"base1.go", "base2.go"},
					`package main

func baseMethod(x int) int {
	result := x * 2
	return result
}`,
				)
				Expect(err).NotTo(HaveOccurred())

				_, err = setup.Executor(
					"baseline",
					setup.TmpDir,
					"-t",
					"1",
					"--baseline-path",
					filepath.Join(setup.TmpDir, ".art-dupl-baseline.json"),
				)
				Expect(err).NotTo(HaveOccurred())

				By("adding a new clone")

				err = setup.CreateDuplicateFiles(
					[]string{"extra1.go", "extra2.go"},
					`package main

func extraMethod(x int) int {
	first := x + 10
	second := first * 3
	third := second - 5
	return third
}`,
				)
				Expect(err).NotTo(HaveOccurred())

				By("running diff-report with JSON")

				output, err := setup.RunArtDupl(
					"--diff-report",
					filepath.Join(setup.TmpDir, ".art-dupl-baseline.json"),
					"-t", "1", "--json", "--quiet",
				)
				Expect(err).NotTo(HaveOccurred())

				var report map[string]any

				err = json.Unmarshal(output, &report)
				Expect(err).NotTo(HaveOccurred())

				Expect(report).To(HaveKey("new"))
				Expect(report).To(HaveKey("suppressed"))
				Expect(report).To(HaveKey("resolved"))
			})
		})
	})
})
