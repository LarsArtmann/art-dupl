package bdd

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// Templ suppression directives (issue #9): clone pairs inside templ markup
// need an in-source accept form that does not leak into the rendered page.
//
// Two forms are supported:
//
//   - The line-leading Go comment (`// art-dupl:accept ...`): templ parses
//     this as a GoComment and its generator drops it from the rendered
//     output entirely ("Do not render Go comments in the output HTML") —
//     the zero-leak, preferred form. It was already recognized by the
//     scanner regex; pinned here because the issue believed it rendered as
//     page text (it does not, in line-leading position).
//   - The HTML comment (`<!-- art-dupl:accept ... -->`, without the nested
//     // the old workaround required): browsers never display HTML
//     comments, so nothing leaks into the visible page either. Recognized
//     since issue #9.
//
// Fixture markers are assembled by concatenation: these _test.go sources
// are themselves scanned on self-runs, so a contiguous marker here would
// accept unrelated groups near these lines.
const templAcceptBody = `	<div class="badges">
		%s
		<span class="badge">{ label }{ fmt.Sprintf("%d", n) }</span>
		<span class="badge">{ label }{ fmt.Sprintf("%d", n) }</span>
	</div>
`

const templAcceptHeader = `package main

import "fmt"

templ Badges(label string, n int) {
`

func templWithDirective(directiveLine string) string {
	return templAcceptHeader + strings.Replace(templAcceptBody, "%s", directiveLine, 1) + "}"
}

var _ = Describe("Templ accept directives in markup", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	It("shows the markup clone group when no directive is present", func() {
		err := setup.CreateTestFile("control.templ", templWithDirective("\t<!-- filler comment -->"))
		Expect(err).NotTo(HaveOccurred())

		stdout, err := setup.RunArtDupl("--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(stdout)).To(ContainSubstring("found 2 clones"))
		Expect(string(stdout)).To(ContainSubstring("control.templ"))
	})

	It("suppresses via the HTML comment form without the nested // marker", func() {
		directive := "<!-- art-dupl:" + "accept templ markup conditional-span twin -->"
		err := setup.CreateTestFile("htmlform.templ", templWithDirective("\t"+directive))
		Expect(err).NotTo(HaveOccurred())

		stdout, err := setup.RunArtDupl("--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(stdout)).To(ContainSubstring("0 shown"))
		Expect(string(stdout)).To(ContainSubstring("filtered suppressed"))
	})

	It("suppresses via the line-leading templ Go comment form", func() {
		directive := "// art-dupl:" + "accept templ markup conditional-span twin"
		err := setup.CreateTestFile("gocomment.templ", templWithDirective("\t"+directive))
		Expect(err).NotTo(HaveOccurred())

		stdout, err := setup.RunArtDupl("--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(stdout)).To(ContainSubstring("0 shown"))
		Expect(string(stdout)).To(ContainSubstring("filtered suppressed"))
	})

	It("keeps suppressing with the legacy nested workaround form", func() {
		directive := "<!-- // art-dupl:" + "accept legacy workaround -->"
		err := setup.CreateTestFile("legacy.templ", templWithDirective("\t"+directive))
		Expect(err).NotTo(HaveOccurred())

		stdout, err := setup.RunArtDupl("--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(stdout)).To(ContainSubstring("0 shown"))
	})
})
