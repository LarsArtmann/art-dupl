package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Pins for the templ whole-file "phantom" limitation documented 2026-10-04 in
// AGENTS.md ("Expression-aware templ detection"): two .templ files with the
// SAME component name+signature were claimed to form a length-2
// [File-token, ComponentDecl-token] sequence that reports as a whole-file
// clone group (preview "package main", e.g. `a.templ:1-6`) at thresholds
// below the real shared token count.
//
// INVESTIGATION VERDICT (2026-10-11): the whole-file claim is NOT
// REPRODUCIBLE on current code. Seven fixture shapes were tried — bare
// same-signature pairs with divergent bodies, shared body prefix +
// spread-attribute divergence (the original incident's exposing shape,
// cf. TestSpreadAttributeExpressionDifferencesPreventClone), and fully
// IDENTICAL component files. None produces a whole-file-spanning group:
// identical files report only the correctly line-scoped shared elements
// (e.g. `a.templ:4-6`, preview "<div ...>"), never `:1-N` with a
// "package main" preview. The original incident's exact fixture is
// unrecoverable (observed during the 2026-10-04 templ development; the
// suspected source repo has no .templ files), so the historical mechanism
// note stays in AGENTS.md as context while these tests pin the ACTUAL
// current behavior. If a spec here starts failing with a whole-file group
// appearing, the phantom is REAL again — update AGENTS.md and these pins
// together.
const templPhantomA = `package main

templ Panel(title string) {
	<div class="card">{ title }</div>
	<div { user.Attrs... }>x</div>
}
`

const templPhantomB = `package main

templ Panel(title string) {
	<div class="card">{ title }</div>
	<div { user.Different... }>x</div>
}
`

const templDivergentB = `package main

templ Panel(title string) {
	<span data-kind="other">{ strings.ToUpper(title) }</span>
	<em>{ user.Note }</em>
}
`

const templIdenticalA = `package main

templ Panel(title string) {
	<div class="card">{ title }</div>
	<div { user.Attrs... }>x</div>
}
`

var _ = Describe("Templ whole-file phantom limitation pins", func() {
	It("reports no group for same-signature components with fully divergent bodies at -t 1", func() {
		// CURRENT+PINNED: with nothing shared beyond the declaration, the
		// [File, ComponentDecl] prefix does not surface as a clone group
		// even at the lowest threshold — zero groups on current code.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": templPhantomA,
			"b.templ": templDivergentB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("Found total 0 clone groups"))
	})

	It("keeps groups line-scoped for shared bodies with spread-attribute divergence at -t 1", func() {
		// CURRENT+PINNED: the original incident's exposing shape — same
		// signature, shared body prefix, spread-attribute divergence. Only
		// the genuinely shared div is reported, correctly line-scoped; no
		// whole-file span, no "package main" preview.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": templPhantomA,
			"b.templ": templPhantomB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("a.templ:4-5"))
		Expect(string(output)).To(ContainSubstring("b.templ:4-5"))
		Expect(string(output)).ToNot(ContainSubstring("a.templ:1-"))
		Expect(string(output)).ToNot(ContainSubstring("package main"))
	})

	It("keeps groups line-scoped for fully identical templ files at -t 1", func() {
		// CURRENT+PINNED: even byte-identical component files report only
		// the shared elements with their real source range — no whole-file
		// span, no "package main" preview.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": templIdenticalA,
			"b.templ": templIdenticalA,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("a.templ:4-"))
		Expect(string(output)).To(ContainSubstring("b.templ:4-"))
		Expect(string(output)).ToNot(ContainSubstring("a.templ:1-"))
		Expect(string(output)).ToNot(ContainSubstring("package main"))
	})

	It("does not report the shared prefix at the default threshold", func() {
		// The gate that keeps any residual prefix artifact invisible in
		// practice: a length-2 prefix cannot reach the default threshold
		// of 5 duplicated statements.
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
