package bdd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// BDD Test Suite for Templ (.templ) File Clone Detection
//
// These tests verify that art-dupl can detect code clones in .templ source files
// using the github.com-a-h-templ parser.
//
// Note: This is different from filtering *_templ.go generated files.
// These tests verify actual clone detection IN .templ source files.

var _ = Describe("Templ Clone Detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	// runTemplOutputTest creates two .templ files with duplicate structures and
	// verifies the output format includes the expected file references.
	// This helper parameterizes the common pattern used across output format tests.
	runTemplOutputTest := func(templCode1, templCode2, filename1, filename2, formatFlag, threshold string, expectedSubstrings []string) {
		err := setup.CreateTestFile(filename1, templCode1)
		Expect(err).NotTo(HaveOccurred())
		err = setup.CreateTestFile(filename2, templCode2)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl(formatFlag, "--threshold", threshold)
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)

		for _, substr := range expectedSubstrings {
			Expect(outputStr).To(ContainSubstring(substr))
		}
	}

	// templComponentCode returns a standardized component template with custom HTML
	templComponentCode := func(componentName, paramName, htmlBody string) string {
		return fmt.Sprintf(`package main

templ %s(%s string) {
	%s
}`, componentName, paramName, htmlBody)
	}

	// createTemplComponent returns a templ component with the given HTML body template.
	// The htmlTemplate should be a format string with one %s placeholder for the parameter.
	createTemplComponent := func(componentName, paramName, htmlTemplate string) string {
		return templComponentCode(componentName, paramName,
			fmt.Sprintf(htmlTemplate, paramName))
	}

	// buttonTemplCode returns a standardized button component template
	buttonTemplCode := func(componentName, paramName string) string {
		return createTemplComponent(
			componentName,
			paramName,
			`<button type="button" class="btn">\n\t\t{ %s }\n\t</button>`,
		)
	}

	// inputTemplCode returns a standardized input field component template
	inputTemplCode := func(componentName, paramName string) string {
		return createTemplComponent(componentName, paramName, `<input type="text" name={ %s } />`)
	}

	Context("When analyzing .templ source files", func() {
		It("should detect duplicate components in .templ files", func() {
			// Create two .templ files with identical component structures
			templCode1 := `package main

templ Header(title string) {
	<header class="main-header">
		<h1>{ title }</h1>
		<nav>
			<a href="/">Home</a>
			<a href="/about">About</a>
		</nav>
	</header>
}

templ Footer() {
	<footer>
		<p>Copyright 2024</p>
	</footer>
}`
			templCode2 := `package main

templ PageHeader(name string) {
	<header class="main-header">
		<h1>{ name }</h1>
		<nav>
			<a href="/">Home</a>
			<a href="/about">About</a>
		</nav>
	</header>
}

templ PageFooter() {
	<footer>
		<p>Copyright 2024</p>
	</footer>
}`
			err := setup.CreateTestFile("header.templ", templCode1)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("page.templ", templCode2)
			Expect(err).NotTo(HaveOccurred())

			// Run art-dupl (templ files included by default)
			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should detect the duplicate header components
			Expect(outputStr).To(ContainSubstring("header.templ"))
			Expect(outputStr).To(ContainSubstring("page.templ"))
		})

		It("should detect duplicate nested element structures", func() {
			// Create .templ files with duplicate nested HTML structures
			cardCode := `package main

templ Card(title string, content string) {
	<div class="card">
		<div class="card-header">
			<h2>{ title }</h2>
		</div>
		<div class="card-body">
			<p>{ content }</p>
		</div>
		<div class="card-footer">
			<button type="button">Action</button>
		</div>
	</div>
}`
			panelCode := `package main

templ Panel(heading string, text string) {
	<div class="card">
		<div class="card-header">
			<h2>{ heading }</h2>
		</div>
		<div class="card-body">
			<p>{ text }</p>
		</div>
		<div class="card-footer">
			<button type="button">Action</button>
		</div>
	</div>
}`
			err := setup.CreateTestFile("card.templ", cardCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("panel.templ", panelCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "5")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Should detect the duplicate card/panel structures
			Expect(outputStr).To(ContainSubstring("card.templ"))
			Expect(outputStr).To(ContainSubstring("panel.templ"))
		})

		It("should detect duplicate for loops in templates", func() {
			listCode := `package main

templ ItemList(items []string) {
	<ul class="item-list">
		for _, item := range items {
			<li class="item">{ item }</li>
		}
	</ul>
}`
			menuCode := `package main

templ Menu(options []string) {
	<ul class="item-list">
		for _, option := range options {
			<li class="item">{ option }</li>
		}
	</ul>
}`
			err := setup.CreateTestFile("list.templ", listCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("menu.templ", menuCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "3")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			Expect(outputStr).To(ContainSubstring("list.templ"))
			Expect(outputStr).To(ContainSubstring("menu.templ"))
		})

		It("should detect duplicate if statements in templates", func() {
			conditionalCode := `package main

templ Conditional(show bool, message string) {
	if show {
		<div class="alert">
			<p>{ message }</p>
		</div>
	}
}`
			toggleCode := `package main

templ Toggle(visible bool, text string) {
	if visible {
		<div class="alert">
			<p>{ text }</p>
		</div>
	}
}`
			err := setup.CreateTestFile("conditional.templ", conditionalCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("toggle.templ", toggleCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "3")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			Expect(outputStr).To(ContainSubstring("conditional.templ"))
			Expect(outputStr).To(ContainSubstring("toggle.templ"))
		})
	})

	Context("When mixing .templ and .go files", func() {
		It("should exclude .templ files when --exclude-templ is used", func() {
			// Create a Go file with duplicate code
			goCode := `package main

func ProcessData(name string, count int) error {
	if name == "" {
		return nil
	}
	for i := 0; i < count; i++ {
		println(name, i)
	}
	return nil
}`
			// Create a .templ file (won't have Go function clones but should be analyzed)
			templCode := `package main

templ Display(name string) {
	<div>
		<h1>{ name }</h1>
	</div>
}`
			err := setup.CreateTestFile("handler.go", goCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("display.templ", templCode)
			Expect(err).NotTo(HaveOccurred())

			// Run with --exclude-templ - should only analyze .go files
			output, err := setup.RunArtDupl("--exclude-templ", "--threshold", "3")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// .templ files should not appear when excluded
			Expect(outputStr).ToNot(ContainSubstring("display.templ"))
		})
	})

	Context("When using different output formats with .templ files", func() {
		It("should produce valid JSON output including .templ files", func() {
			runTemplOutputTest(
				buttonTemplCode("Button", "text"),
				buttonTemplCode("Submit", "label"),
				"button.templ", "submit.templ",
				"--json", "3",
				[]string{"button.templ", "submit.templ", `"clone_groups"`},
			)
		})

		It("should produce HTML output including .templ files", func() {
			runTemplOutputTest(
				inputTemplCode("InputField", "name"),
				inputTemplCode("TextField", "id"),
				"input.templ", "text.templ",
				"--html", "2",
				[]string{"<!DOCTYPE html>", "input.templ", "text.templ"},
			)
		})
	})

	Context("When analyzing CSS and Script declarations", func() {
		It("should process CSS declarations but skip CSS content for clone detection", func() {
			// CSS content nodes (property names, values, identifiers) are intentionally
			// skipped in syntax/templ/templ.go:skipNodeTypes to avoid content-based false positives.
			// This test verifies CSS files are parsed without errors.
			cssCode1 := `package main

css baseStyles() {
	button {
		background: blue;
		color: white;
	}
	div {
		margin: 10px;
	}
}`
			cssCode2 := `package main

css buttonStyles() {
	button {
		background: red;
		color: black;
	}
	div {
		margin: 20px;
	}
}`
			err := setup.CreateTestFile("base.templ", cssCode1)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("button_styles.templ", cssCode2)
			Expect(err).NotTo(HaveOccurred())

			// Use a very low threshold to try to detect any clones
			output, err := setup.RunArtDupl("--threshold", "2")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			// Both files should be processed (appear in output or not cause errors)
			// The key assertion is that the tool runs successfully on CSS-containing files
			Expect(outputStr).ToNot(ContainSubstring("error"))
		})
	})

	Context("When analyzing complex templ structures", func() {
		It("should detect duplicates in components with nested elements", func() {
			// Create complex nested structures that are structurally identical
			formCode := `package main

templ Form(action string) {
	<form action={ action } method="POST">
		<div class="form-group">
			<label for="email">Email</label>
			<input type="email" id="email" name="email" />
		</div>
		<div class="form-group">
			<label for="password">Password</label>
			<input type="password" id="password" name="password" />
		</div>
		<button type="submit">Submit</button>
	</form>
}`
			loginCode := `package main

templ LoginForm(url string) {
	<form action={ url } method="POST">
		<div class="form-group">
			<label for="email">Email</label>
			<input type="email" id="email" name="email" />
		</div>
		<div class="form-group">
			<label for="password">Password</label>
			<input type="password" id="password" name="password" />
		</div>
		<button type="submit">Submit</button>
	</form>
}`
			err := setup.CreateTestFile("form.templ", formCode)
			Expect(err).NotTo(HaveOccurred())
			err = setup.CreateTestFile("login.templ", loginCode)
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "10")
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)

			Expect(outputStr).To(ContainSubstring("form.templ"))
			Expect(outputStr).To(ContainSubstring("login.templ"))
		})
	})
})
