package bdd

import (
	"encoding/json/v2"
	"os"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Exit Codes and Version Subcommand", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Describe("Version Subcommand", func() {
		Context("when running 'version' without flags", func() {
			It("should print version information", func() {
				result, err := setup.ExecutorResult("version")
				Expect(err).ToNot(HaveOccurred())
				Expect(string(result.Stdout)).To(ContainSubstring("art-dupl"))
				Expect(string(result.Stdout)).To(ContainSubstring("version"))
			})
		})

		Context("when running 'version --json'", func() {
			It("should produce valid JSON with all fields", func() {
				result, err := setup.ExecutorResult("version", "--json")
				Expect(err).ToNot(HaveOccurred())

				var info map[string]string

				err = json.Unmarshal(result.Stdout, &info)
				Expect(err).ToNot(HaveOccurred())

				for _, field := range []string{"version", "commit", "date", "goVersion", "compiler", "platform", "arch"} {
					Expect(info).To(HaveKey(field), "JSON should contain field %q", field)
				}
			})
		})

		Context("when running 'version --short'", func() {
			It("should print just the version string", func() {
				result, err := setup.ExecutorResult("version", "--short")
				Expect(err).ToNot(HaveOccurred())

				output := string(result.Stdout)
				Expect(output).To(MatchRegexp(`^dev\n$|^v?\d+\.\d+`))
			})
		})
	})

	// art-dupl: accepted: each It block exercises a distinct exit-code path; the structural similarity is Ginkgo BDD style, not duplication.
	Describe("Exit Codes for Invalid Config", func() {
		Context("when threshold is negative", func() {
			It("should return an error", func() {
				tempDir := createTempTestFile("exit-bad-threshold-*")
				defer func() { _ = os.RemoveAll(tempDir) }()

				_, err := setup.RunArtDuplOnDir(tempDir, "--threshold", "-1")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when sort is invalid", func() {
			It("should return an error", func() {
				tempDir := createTempTestFile("exit-bad-sort-*")
				defer func() { _ = os.RemoveAll(tempDir) }()

				_, err := setup.RunArtDuplOnDir(tempDir, "--sort", "invalid")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Help and Exit Code Documentation", func() {
		It("should show exit codes in --help output", func() {
			output := getHelpOutput(setup)
			Expect(output).To(ContainSubstring("Exit codes"))
			Expect(output).To(ContainSubstring("130"))
		})
	})
})
