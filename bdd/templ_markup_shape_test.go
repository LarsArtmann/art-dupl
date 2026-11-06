package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// Templ markup-shape precision (issue #14): parallel-by-design UI sections
// share element structure while every meaning-carrying token (ids, classes,
// text content, expressions) differs. Before text content participated in the
// hash, such sections reported as clone groups at -t 1 and consumers
// adjudicated them ACCEPT one by one, growing the CI baseline with blessed
// noise. The fixtures below pin both directions: divergent markup must not
// clone, identical markup still must.

const markupDivergentSections = `package main

templ AdminSections() {
	<section id="settings-head">
		<h2 class="text-lg font-bold">Settings</h2>
		<p class="text-muted">Configure your workspace</p>
		<span class="chip">Active</span>
		<a href="/settings/edit" class="btn">Edit settings</a>
	</section>
	<section id="networks-head">
		<h2 class="text-lg font-bold">Networks</h2>
		<p class="text-muted">Manage your network profile</p>
		<span class="chip">Inactive</span>
		<a href="/networks/edit" class="btn">Edit networks</a>
	</section>
}
`

const markupIdenticalSections = `package main

templ TwinSections() {
	<section id="primary-head">
		<h2 class="text-lg font-bold">Primary</h2>
		<p class="text-muted">Managed upstream</p>
		<span class="chip">Active</span>
		<a href="/primary/edit" class="btn">Edit primary</a>
	</section>
	<section id="primary-head">
		<h2 class="text-lg font-bold">Primary</h2>
		<p class="text-muted">Managed upstream</p>
		<span class="chip">Active</span>
		<a href="/primary/edit" class="btn">Edit primary</a>
	</section>
}
`

var _ = Describe("Templ markup-shape precision", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	It("does not report parallel sections whose text, ids, and classes differ", func() {
		err := setup.CreateTestFile("admin.templ", markupDivergentSections)
		Expect(err).NotTo(HaveOccurred())

		stdout, err := setup.RunArtDupl("--threshold", "1", "--show-suppressed")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(stdout)).NotTo(ContainSubstring("settings-head"))
		Expect(string(stdout)).NotTo(ContainSubstring("networks-head"))
	})

	It("still reports markup that actually repeats", func() {
		err := setup.CreateTestFile("primary.templ", markupIdenticalSections)
		Expect(err).NotTo(HaveOccurred())

		stdout, err := setup.RunArtDupl("--threshold", "1", "--show-suppressed")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(stdout)).To(ContainSubstring("found 2 clones"))
		Expect(string(stdout)).To(ContainSubstring("primary.templ"))
	})
})
