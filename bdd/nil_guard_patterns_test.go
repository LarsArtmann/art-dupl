package bdd

import (
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// End-to-end specs for the nil-guard-after-helper pattern (plan T13) and the
// option-apply-loop seeding extension (gap families from the 2026-10-11
// survivor labeling). Unit-level real-source tests live in
// printer/actionability/actionability_nil_guard_test.go; these prove the
// suppression survives the full CLI pipeline with correct attribution.

var _ = Describe("nil-guard-after-helper pattern (e2e)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"guard_index.go": `package main

type node struct{ name string }

func indexIdentOf(n *node) *node { return n }

func isIndexBounded(idx *node) bool {
ident := indexIdentOf(idx)
if ident == nil {
	return false
}
return ident.name != ""
}
`,
			"guard_name.go": `package main

func isNameValid(target *node) bool {
ident := indexIdentOf(target)
if ident == nil {
	return false
}
return len(ident.name) > 0
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should suppress nil-guard-after-helper clones", func() {
		output, err := setup.RunArtDupl("--quiet", "--semantic", "--threshold", "2")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("0 shown"),
			"helper+nil-guard preludes are canonical early-return plumbing")
	})

	It("should attribute the suppression to nil-guard-after-helper", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--semantic", "--threshold", "2", "--show-suppressed", "--explain",
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("nil-guard-after-helper"))
	})
})

var _ = Describe("option-apply-loop seeded constructor (e2e)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"preset_bbot.go": `package main

type presetConfig struct{ syncWrites bool }

type presetOption func(*presetConfig)

func defaultPresetConfig() presetConfig {
	return presetConfig{syncWrites: true}
}

func newPreset(opts ...presetOption) presetConfig {
	cfg := defaultPresetConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
`,
			"engine_badger.go": `package main

type engineConfig struct{ syncWrites bool }

type engineOption func(*engineConfig)

func newEngine(opts ...engineOption) engineConfig {
	cfg := engineConfig{syncWrites: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should suppress seeded option-apply constructor clones", func() {
		output, err := setup.RunArtDupl("--quiet", "--semantic", "--threshold", "2")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("0 shown"),
			"seeded option-apply constructors are the standard options idiom")
	})

	It("should attribute the suppression to option-apply-loop", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--semantic", "--threshold", "2", "--show-suppressed", "--explain",
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("option-apply-loop"))
	})
})

var _ = Describe("nil-guard negative controls (e2e)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"rollback_a.go": `package main

type txA struct{ failed bool }

func (t *txA) Rollback() error { t.failed = true; return nil }

func closeA(tx *txA) error {
	if tx != nil {
		return tx.Rollback()
	}
	return nil
}
`,
			"rollback_b.go": `package main

type txB struct{ failed bool }

func (t *txB) Rollback() error { t.failed = true; return nil }

func closeB(tx *txB) error {
	if tx != nil {
		return tx.Rollback()
	}
	return nil
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should keep resource-closing nil guards actionable", func() {
		output, err := setup.RunArtDupl("--quiet", "--semantic", "--threshold", "2")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).NotTo(ContainSubstring("0 shown"),
			"guards that close resources are load-bearing, not boilerplate")
	})
})
