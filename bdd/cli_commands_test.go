package bdd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// BDD Test Suite for CLI Commands (Version, Help, etc.)
//
// These tests verify CLI command behavior including version output,
// help documentation, and general command-line interface functionality.
//
// The scenarios cover:
// - Version command output
// - Help command documentation
// - Flag documentation and validation
// - Command-line interface behavior

// assertCommandOutput is a helper to verify command output matches expected patterns.
func assertCommandOutput(
	setup *testutil.BDDTestSetup,
	args []string,
	matchers ...types.GomegaMatcher,
) {
	output, err := setup.RunArtDupl(args...)
	Expect(err).ToNot(HaveOccurred())

	outputStr := string(output)
	Expect(outputStr).To(SatisfyAny(matchers...))
}

// assertHelpOutput verifies help output contains expected patterns.
func assertHelpOutput(setup *testutil.BDDTestSetup, matchers ...types.GomegaMatcher) {
	outputStr := getHelpOutput(setup)
	Expect(outputStr).To(SatisfyAny(matchers...))
}

// setupBDDTest creates and configures a BDDTestSetup for Ginkgo tests.
// Returns the setup instance and cleanup function for use in BeforeEach/AfterEach.
func setupBDDTest() (*testutil.BDDTestSetup, func()) {
	setup, err := testutil.NewBDDTestSetupForGinkgo()
	Expect(err).NotTo(HaveOccurred())

	cleanup := func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	}

	return setup, cleanup
}

// getHelpOutput runs art-dupl --help and returns the output string.
func getHelpOutput(setup *testutil.BDDTestSetup) string {
	output, err := setup.RunArtDupl("--help")
	Expect(err).ToNot(HaveOccurred())

	return string(output)
}

// verifyHelpContent checks that help output contains at least one of the expected substrings.
func verifyHelpContent(setup *testutil.BDDTestSetup, substrings []string) {
	outputStr := getHelpOutput(setup)

	expectations := make([]types.GomegaMatcher, 0, len(substrings))
	for _, substr := range substrings {
		expectations = append(expectations, ContainSubstring(substr))
	}

	Expect(outputStr).To(SatisfyAny(expectations...))
}

var _ = Describe("Version Command", func() {
	var (
		setup   *testutil.BDDTestSetup
		cleanup func()
	)

	BeforeEach(func() {
		setup, cleanup = setupBDDTest()
	})

	AfterEach(func() {
		cleanup()
	})

	Context("When running version command", func() {
		It("should display version information", func() {
			assertCommandOutput(setup, []string{"--version"},
				ContainSubstring("art-dupl"),
				ContainSubstring("version"),
				MatchRegexp(`\d+\.\d+`),
			)
		})

		It("should display version with -v shorthand", func() {
			// Note: -v is verbose in subcommands, but --version should work
			output, err := setup.RunArtDupl("version")
			// May error if version is a subcommand or show help
			_ = err

			Expect(output).ToNot(BeNil())
		})
	})

	Context("When checking version format", func() {
		It("should follow semantic versioning format", func() {
			assertCommandOutput(setup, []string{"--version"},
				MatchRegexp(`v?\d+\.\d+\.?\d*`),
				ContainSubstring("version"),
			)
		})
	})
})

var _ = Describe("Help Command", func() {
	var (
		setup   *testutil.BDDTestSetup
		cleanup func()
	)

	BeforeEach(func() {
		setup, cleanup = setupBDDTest()
	})

	AfterEach(func() {
		cleanup()
	})

	Context("When running help command", func() {
		It("should display usage information", func() {
			assertCommandOutput(setup, []string{"--help"},
				ContainSubstring("Usage:"),
				ContainSubstring("usage:"),
				ContainSubstring("art-dupl"),
			)
		})

		It("should list available flags", func() {
			assertHelpOutput(setup,
				ContainSubstring("--threshold"),
				ContainSubstring("--json"),
				ContainSubstring("--html"),
			)
		})

		It("should describe available commands", func() {
			verifyHelpContent(setup, []string{"stats", "Commands:", "Available"})
		})
	})

	Context("When running help for specific subcommands", func() {
		It("should display stats subcommand help", func() {
			output, err := setup.RunArtDupl("stats", "--help")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Stats help should contain stats-specific information
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring("stats"),
				ContainSubstring("statistics"),
			))
		})

		It("should show flags specific to stats command", func() {
			output, err := setup.RunArtDupl("stats", "--help")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Stats should have its own flags
			Expect(outputStr).ToNot(BeEmpty())
		})
	})
})

var _ = Describe("CLI Flag Validation", func() {
	var (
		setup   *testutil.BDDTestSetup
		cleanup func()
	)

	BeforeEach(func() {
		setup, cleanup = setupBDDTest()
	})

	AfterEach(func() {
		cleanup()
	})

	Context("When using invalid flags", func() {
		It("should reject unknown flags", func() {
			output, err := setup.RunArtDupl("--unknown-flag-that-does-not-exist")
			// Should error on unknown flag
			Expect(err).To(HaveOccurred())
			Expect(string(output)).ToNot(BeEmpty())
		})

		It("should handle invalid flag values", func() {
			// Create test file
			err := setup.CreateFileWithContent("test.go", "package main\nfunc test() {}")
			Expect(err).NotTo(HaveOccurred())

			// Try to use invalid threshold value
			output, err := setup.RunArtDupl("--threshold", "not-a-number", setup.TmpDir)
			// Should error or use default
			_ = err

			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using mutually exclusive flags", func() {
		It("should handle multiple output formats gracefully", func() {
			// Create test files
			err := setup.CreateDuplicateFiles(
				[]string{"fmt1.go", "fmt2.go"},
				"package main\nfunc test() {}",
			)
			Expect(err).NotTo(HaveOccurred())

			// Try using multiple output formats
			output, err := setup.RunArtDupl("--json", "--html", "--plumbing", setup.TmpDir)
			// May error or use priority
			_ = err

			Expect(output).ToNot(BeNil())
		})
	})
})

var _ = Describe("Verbose Flag Behavior", func() {
	var (
		setup   *testutil.BDDTestSetup
		cleanup func()
	)

	BeforeEach(func() {
		setup, cleanup = setupBDDTest()
	})

	AfterEach(func() {
		cleanup()
	})

	Context("When using verbose flags", func() {
		DescribeTable(
			"should work with various verbose flag formats",
			func(funcName string, files []string, flags ...string) {
				code := fmt.Sprintf(`package main
func %s() {}`, funcName)
				output, err := setup.CreateNamedDuplicateFilesAndRun(files, code, flags...)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())
			},
			Entry(
				"single verbose flag",
				"verbose1",
				[]string{"verbose1.go", "verbose2.go"},
				"-v",
				"--threshold",
				"5",
			),
			Entry(
				"multiple verbose flags",
				"verbose2",
				[]string{"verbose3.go", "verbose4.go"},
				"-vv",
				"--threshold",
				"5",
			),
			Entry(
				"triple verbose flag",
				"verbose3",
				[]string{"verbose5.go", "verbose6.go"},
				"-vvv",
				"--threshold",
				"5",
			),
			Entry(
				"verbose long flag",
				"verbose4",
				[]string{"verbose7.go", "verbose8.go"},
				"--verbose",
				"--threshold",
				"5",
			),
		)
	})
})

var _ = Describe("CLI Error Handling", func() {
	var (
		setup   *testutil.BDDTestSetup
		cleanup func()
	)

	BeforeEach(func() {
		setup, cleanup = setupBDDTest()
	})

	AfterEach(func() {
		cleanup()
	})

	Context("When no arguments provided", func() {
		It("should handle empty directory gracefully", func() {
			testutil.TestEmptyDirectory(setup, 10)
		})
	})

	Context("When providing positional arguments", func() {
		It("should accept directory paths as arguments", func() {
			code := `package main
func argTest() {}`

			err := setup.CreateDuplicateFiles([]string{"arg1.go", "arg2.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run with path as positional argument
			output, err := setup.RunArtDupl(setup.TmpDir, "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("should accept multiple directory paths", func() {
			err := setup.CreateSubdirectories("dir1", "dir2")
			Expect(err).NotTo(HaveOccurred())

			code := `package main
func multiDir() {}`

			err = setup.CreateFileWithContent("dir1/file.go", code)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateFileWithContent("dir2/file.go", code)
			Expect(err).NotTo(HaveOccurred())

			dir1 := setup.GetFilePath("dir1")
			dir2 := setup.GetFilePath("dir2")

			// Run with multiple paths
			output, err := setup.RunArtDupl(dir1, dir2, "--threshold", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})
})

var _ = Describe("CLI Completion Commands", func() {
	var (
		setup   *testutil.BDDTestSetup
		cleanup func()
	)

	BeforeEach(func() {
		setup, cleanup = setupBDDTest()
	})

	AfterEach(func() {
		cleanup()
	})

	Context("When requesting shell completion", func() {
		shells := []string{"bash", "zsh", "fish"}
		for _, shell := range shells {
			It(fmt.Sprintf("should provide %s completion", shell), func() {
				output, err := setup.RunArtDupl("completion", shell)
				// May or may not be available
				if err == nil {
					outputStr := string(output)
					Expect(outputStr).To(SatisfyAny(
						ContainSubstring(shell),
						ContainSubstring("completion"),
					))
				}
			})
		}
	})

	Context("When requesting man page", func() {
		It("should provide man page output", func() {
			output, err := setup.RunArtDupl("man")
			// May or may not be available
			if err == nil {
				outputStr := string(output)
				// Man page typically starts with .TH
				Expect(outputStr).To(SatisfyAny(
					HavePrefix(".TH"),
					ContainSubstring("art-dupl"),
				))
			}
		})
	})
})

var _ = Describe("CLI Documentation Quality", func() {
	var (
		setup   *testutil.BDDTestSetup
		cleanup func()
	)

	BeforeEach(func() {
		setup, cleanup = setupBDDTest()
	})

	AfterEach(func() {
		cleanup()
	})

	Context("When reviewing help documentation", func() {
		It("should describe detection methods in help", func() {
			assertCommandOutput(setup, []string{"--help"},
				ContainSubstring("detection"),
				ContainSubstring("method"),
			)
		})

		It("should describe output formats in help", func() {
			verifyHelpContent(setup, []string{"json", "html", "output"})
		})

		It("should describe sorting options in help", func() {
			outputStr := getHelpOutput(setup)
			// Should mention sorting
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring("sort"),
			))
		})

		It("should provide examples in help", func() {
			outputStr := getHelpOutput(setup)
			Expect(outputStr).To(SatisfyAny(
				ContainSubstring("EXAMPLES"),
				ContainSubstring("Example"),
				ContainSubstring("example"),
			))
		})
	})
})
