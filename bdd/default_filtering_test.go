package bdd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// BDD Test Suite for Default Filtering Behavior
//
// These tests verify that generated code is filtered by default in the main command,
// ensuring users don't see noise from auto-generated files.
//
// Generated code types covered:
// - Templ: *_templ.go files (templ.guide)
// - SQLC: *.sql.go, models.go, querier.go, batch.go files (sqlc.dev)

// assertGenFileFiltered is a helper to verify generated files are filtered out.
// It creates duplicate regular files, a generated file, runs art-dupl, and verifies
// that regular files appear in output while the generated file does not.
func assertGenFileFiltered(
	setup *testutil.BDDTestSetup,
	regularFiles []string,
	regularCode string,
	generatedFile string,
	generatedCode string,
	threshold string,
) string {
	err := setup.CreateDuplicateFiles(regularFiles, regularCode)
	Expect(err).NotTo(HaveOccurred())
	err = setup.CreateTestFile(generatedFile, generatedCode)
	Expect(err).NotTo(HaveOccurred())

	output, err := setup.RunArtDupl("--threshold", threshold)
	Expect(err).ToNot(HaveOccurred())

	outputStr := string(output)

	// Should find duplicates in regular files
	for _, file := range regularFiles {
		Expect(outputStr).To(ContainSubstring(file))
	}
	// Should NOT include generated file
	Expect(outputStr).ToNot(ContainSubstring(generatedFile))

	return outputStr
}

// assertFilteredTempl verifies templ files are filtered out by default.
func assertFilteredTempl(
	setup *testutil.BDDTestSetup,
	regularFiles []string,
	regularCode string,
	generatedFile string,
	generatedCode string,
	threshold string,
) string {
	err := setup.CreateDuplicateFiles(regularFiles, regularCode)
	Expect(err).NotTo(HaveOccurred())
	err = setup.CreateTestFile(generatedFile, generatedCode)
	Expect(err).NotTo(HaveOccurred())

	output, err := setup.RunArtDupl("--threshold", threshold)

	Expect(err).ToNot(HaveOccurred())

	outputStr := string(output)

	for _, file := range regularFiles {
		Expect(outputStr).To(ContainSubstring(file))
	}

	Expect(outputStr).ToNot(ContainSubstring(generatedFile))

	return outputStr
}

// assertSQLCFileFiltered tests that a specific SQLC generated file type is filtered out.
// It uses a standardized code pattern with the provided functionName.
func assertSQLCFileFiltered(setup *testutil.BDDTestSetup, filename, functionName string) {
	regularCode := fmt.Sprintf("package main\nfunc %s() { println(1) }", functionName)
	generatedCode := fmt.Sprintf("package db\nfunc %s() { println(1) }", functionName)

	assertGenFileFiltered(
		setup,
		[]string{"service1.go", "service2.go"},
		regularCode,
		filename,
		generatedCode,
		"3",
	)
}

// assertGeneratedFileIncluded is a helper to verify generated files are included when
// a specific flag is used. It creates duplicate regular files, a generated file, runs
// art-dupl with the include flag, and verifies that the generated file appears in output.
func assertGeneratedFileIncluded(
	setup *testutil.BDDTestSetup,
	regularFiles []string,
	regularCode string,
	generatedFile string,
	generatedCode string,
	includeFlag string,
	threshold string,
) string {
	err := setup.CreateDuplicateFiles(regularFiles, regularCode)
	Expect(err).NotTo(HaveOccurred())
	err = setup.CreateTestFile(generatedFile, generatedCode)
	Expect(err).NotTo(HaveOccurred())

	var output []byte
	if includeFlag != "" {
		output, err = setup.RunArtDupl(includeFlag, "--threshold", threshold)
	} else {
		output, err = setup.RunArtDupl("--threshold", threshold)
	}

	Expect(err).ToNot(HaveOccurred())

	outputStr := string(output)

	// Should now include generated file
	Expect(outputStr).To(ContainSubstring(generatedFile))

	return outputStr
}

var _ = Describe("Default Filtering Behavior", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	Context("When analyzing code with templ generated files", func() {
		It("should exclude *_templ.go files by default", func() {
			regularCode := `package main

import "fmt"

func process() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`
			templCode := `package main

import "github.com/a-h/templ"

func Page() templ.Component {
	return nil
}

func process() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`

			assertFilteredTempl(
				setup,
				[]string{"handler1.go", "handler2.go"},
				regularCode,
				"page_templ.go",
				templCode,
				"5",
			)
		})

		It("should exclude multiple templ files by default", func() {
			regularCode := `package main
func common() { println(1) }`

			templCode := `package main
import "github.com/a-h/templ"
func Component() templ.Component { return nil }
func common() { println(1) }`

			// Create TWO regular files with duplicate code (required for clone detection)
			err := setup.CreateTestFile("regular1.go", regularCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("regular2.go", regularCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("header_templ.go", templCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("footer_templ.go", templCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("layout_templ.go", templCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "3")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should only show regular files (duplicates between them)
			Expect(outputStr).To(ContainSubstring("regular"))
			Expect(outputStr).ToNot(ContainSubstring("header_templ.go"))
			Expect(outputStr).ToNot(ContainSubstring("footer_templ.go"))
			Expect(outputStr).ToNot(ContainSubstring("layout_templ.go"))
		})

		It("should include templ files when --include-templ is used", func() {
			assertGeneratedFileIncluded(
				setup,
				[]string{"regular1.go", "regular2.go"},
				testRegularCode,
				"page_templ.go",
				testTemplCode,
				"--include-templ",
				"3",
			)
		})
	})

	Context("When analyzing code with SQLC generated files", func() {
		It("should exclude *.sql.go files by default", func() {
			regularCode := `package main

import "fmt"

func query() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`
			sqlcCode := `// Code generated by sqlc. DO NOT EDIT.
package db

func query() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`

			assertGenFileFiltered(
				setup,
				[]string{"service1.go", "service2.go"},
				regularCode,
				"queries.sql.go",
				sqlcCode,
				"5",
			)
		})

		It("should exclude models.go files by default", func() {
			assertSQLCFileFiltered(setup, "models.go", "validate")
		})

		It("should exclude querier.go files by default", func() {
			assertSQLCFileFiltered(setup, "querier.go", "query")
		})

		It("should exclude batch.go files by default", func() {
			assertSQLCFileFiltered(setup, "batch.go", "batch")
		})

		It("should include SQLC files when --include-sqlc is used", func() {
			regularCode := `package main
func query() { println(1) }`
			sqlcCode := `// Code generated by sqlc. DO NOT EDIT.
package db
func query() { println(1) }`

			assertGeneratedFileIncluded(
				setup,
				[]string{"service1.go", "service2.go"},
				regularCode,
				"queries.sql.go",
				sqlcCode,
				"--include-sqlc",
				"3",
			)
		})
	})

	Context("When both templ and sqlc files are present", func() {
		It("should filter both templ and sqlc files by default", func() {
			regularCode := `package main
func process() { println(1) }`
			templCode := `package main
import "github.com/a-h/templ"
func Component() templ.Component { return nil }
func process() { println(1) }`
			sqlcCode := `// Code generated by sqlc. DO NOT EDIT.
package db
func process() { println(1) }`

			// Create TWO regular files with duplicate code (required for clone detection)
			err := setup.CreateTestFile("regular1.go", regularCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("regular2.go", regularCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("page_templ.go", templCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("queries.sql.go", sqlcCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "3")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should only show regular files (duplicates between them)
			Expect(outputStr).To(ContainSubstring("regular"))
			Expect(outputStr).ToNot(ContainSubstring("page_templ.go"))
			Expect(outputStr).ToNot(ContainSubstring("queries.sql.go"))
		})

		It("should include both when both include flags are used", func() {
			regularCode := `package main
func process() { println(1) }`
			templCode := `package main
import "github.com/a-h/templ"
func Component() templ.Component { return nil }
func process() { println(1) }`
			sqlcCode := `// Code generated by sqlc. DO NOT EDIT.
package db
func process() { println(1) }`

			err := setup.CreateTestFile("regular.go", regularCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("page_templ.go", templCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("queries.sql.go", sqlcCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with both include flags
			output, err := setup.RunArtDupl("--include-templ", "--include-sqlc", "--threshold", "3")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should show all files
			Expect(outputStr).To(ContainSubstring("regular.go"))
			Expect(outputStr).To(ContainSubstring("page_templ.go"))
			Expect(outputStr).To(ContainSubstring("queries.sql.go"))
		})
	})

	Context("When vendor directory is present", func() {
		It("should exclude vendor directory by default", func() {
			err := setup.CreateVendorDuplicateFiles(
				"vendor/github.com/example",
				testutil.SimpleVendorTestCode,
			)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "3")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should not mention vendor files
			Expect(outputStr).ToNot(ContainSubstring("vendor"))
		})

		It("should include vendor when --vendor flag is used", func() {
			output, err := setup.RunVendorTestWithOptions(
				"vendor/github.com/example",
				testutil.SimpleVendorTestCode,
				true,
				"",
				"--threshold", "3",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("vendor"))
		})
	})
})

// testCodeSamples contains code samples used in filtering tests.
const (
	testRegularCode = `package main
func process() { println(1) }`
	testTemplCode = `package main
import "github.com/a-h/templ"
func Component() templ.Component { return nil }
func process() { println(1) }`
)

// assertTemplFilteredWithFormat verifies that templ files are filtered out
// when using a specific output format. It creates duplicate regular files,
// a templ file, runs art-dupl with the specified format flag, and verifies
// that the templ file does not appear in the output.
func assertTemplFilteredWithFormat(
	setup *testutil.BDDTestSetup,
	outputFormatFlag string,
) {
	err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, testRegularCode)
	Expect(err).NotTo(HaveOccurred())
	err = setup.CreateTestFile("page_templ.go", testTemplCode)
	Expect(err).NotTo(HaveOccurred())

	output, err := setup.RunArtDupl(outputFormatFlag, "--threshold", "3")
	Expect(err).ToNot(HaveOccurred())

	Expect(string(output)).ToNot(ContainSubstring("page_templ.go"))
}

var _ = Describe("Filtering in Different Output Formats", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	DescribeTable("should not include templ files by default in various output formats",
		func(format string) {
			assertTemplFilteredWithFormat(setup, format)
		},
		Entry("with JSON output", "--json"),
		Entry("with HTML output", "--html"),
		Entry("with plumbing output", "--plumbing"),
	)
})
