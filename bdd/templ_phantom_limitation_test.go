package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Fixtures for the templ `-t 1` whole-file phantom (plan T3, from the
// documented artifact in AGENTS.md → "Expression-aware templ detection"):
// two .templ files with the SAME component name and signature form a
// length-2 [File-token, ComponentDecl-token] sequence. The File token
// carries zero content, so at thresholds below the real shared token count
// this prefix reports as a whole-file clone group (preview "package main")
// even when the component bodies diverge completely (different elements,
// attributes, and expressions — nothing shared beyond the declaration).
const templPhantomA = `package main

templ Panel(title string) {
	<div class="side-a">{ title }</div>
}
`

const templPhantomB = `package main

templ Panel(title string) {
	<span data-kind="other">{ strings.ToUpper(title) }</span>
}
`

var _ = Describe("Templ -t 1 whole-file phantom (documented limitation)", func() {
	It("reports the phantom whole-file group at threshold 1", func() {
		// CURRENT: the length-2 [File, ComponentDecl] prefix matches and the
		// reported clone spans the whole file (preview "package main")
		// although the div bodies diverge.
		// EXPECTED: stays as-is until the File token is dropped from the
		// templ stream (ROADMAP candidate fix; needs a CacheVersion bump).
		// If this test starts failing with the group GONE at -t 1, the
		// artifact was fixed — update AGENTS.md's limitation note and this
		// suite together.
		// IMPACT: invisible at the default threshold (length 2 < 5); only
		// affects -t 1/-t 2 templ scans.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": templPhantomA,
			"b.templ": templPhantomB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("a.templ:1-"))
		Expect(string(output)).To(ContainSubstring("b.templ:1-"))
		Expect(string(output)).To(ContainSubstring("package main"))
	})

	It("does not report the phantom at the default threshold", func() {
		// The gate that keeps the artifact invisible in practice: the
		// length-2 prefix cannot reach the default threshold of 5.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": templPhantomA,
			"b.templ": templPhantomB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "5", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("Found total 0 clone groups"))
	})
})
