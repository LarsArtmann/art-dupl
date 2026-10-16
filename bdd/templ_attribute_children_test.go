package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// Regression pins for templ clone-detection precision (issues #6 and #7).
//
// #6: constant attribute values participate in the fingerprint (markup
// identity) — paragraphs with different class strings must not group, while
// reordered class lists still match (CSS class order carries no meaning).
//
// #7: the bare `{ children... }` slot-injection idiom is never actionable —
// children IS templ's composition mechanism — and is suppressed under the
// templ-children-idiom pattern, still reachable via --show-suppressed.

var _ = Describe("templ constant attribute values (issue #6)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": `package main

templ alpha() {
	<p class="mt-4 text-sm text-gray-600 dark:text-gray-400">{ i18n.T(ctx, i18n.KeyAlpha) }</p>
}

templ beta() {
	<p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{ i18n.T(ctx, i18n.KeyAlpha) }</p>
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("does not group elements whose class values differ", func() {
		output, err := setup.RunArtDupl("--quiet", "--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("Found total 0 clone groups"),
			"differing class strings are markup identity and must not match")
	})

	It("still groups elements with identical class values", func() {
		err := setup.CreateTestFiles(map[string]string{
			"a.templ": `package main

templ alpha() {
	<p class="mt-4 text-sm">{ i18n.T(ctx, i18n.KeyAlpha) }</p>
}

templ beta() {
	<p class="mt-4 text-sm">{ i18n.T(ctx, i18n.KeyAlpha) }</p>
}
`,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--quiet", "--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("Found total 1 clone groups"))
	})
})

var _ = Describe("templ reordered class lists still clone (issue #6)", func() {
	It("groups class lists that differ only in token order", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": `package main

templ alpha() {
	<p class="mt-4 text-sm text-gray-500">{ i18n.T(ctx, i18n.KeyAlpha) }</p>
}

templ beta() {
	<p class="text-gray-500 mt-4 text-sm">{ i18n.T(ctx, i18n.KeyAlpha) }</p>
}
`,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--quiet", "--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("Found total 1 clone groups"),
			"reordered class lists are semantically identical markup and must match")
	})
})

var _ = Describe("templ { children... } wrapper idiom (issue #7)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": `package main

templ shellA() {
	<div class="wrap">{ children... }</div>
}

templ shellB() {
	<section class="wrap">{ children... }</section>
}
`,
			"page.templ": `package main

templ pageSubtitle() {
	<h2 class="subtitle">{ children... }</h2>
}

templ pageLink() {
	<a class="link">{ children... }</a>
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("shows no default-visible group for wrapper pairs", func() {
		output, err := setup.RunArtDupl("--quiet", "--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("0 shown"),
			"wrapper tags differ (a.templ) or the bare children idiom is suppressed (page.templ)")
	})

	It("surfaces the suppressed children groups via --show-suppressed with the pattern label", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--threshold", "1", "--show-suppressed", "--explain",
		)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("templ-children-idiom"),
			"--explain should attribute the suppression to the children-idiom pattern")
		Expect(string(output)).To(ContainSubstring("children..."))
	})

	It("does not suppress the children idiom once the pattern is disabled", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--threshold", "1", "--disable-pattern", "templ-children-idiom",
		)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).ToNot(ContainSubstring("0 shown"),
			"--disable-pattern templ-children-idiom must un-suppress the page.templ children pair")
	})
})
