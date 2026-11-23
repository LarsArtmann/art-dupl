package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Pins for the statement-threshold semantics (verified 2026-11-23, webphone
// follow-up): a duplicated function body counts only its EFFECTIVE top-level
// statements — trimSubsumedIndexes drops statement tokens whose source range
// lies inside a composite statement's (an if swallows the return in its
// body), and FindSyntaxUnits requires len(units) >= threshold whenever the
// match window contains statement tokens. Two consequences pinned here:
//
//   - two functions identical except the name, body = if + trailing return
//     (2 effective statements), are invisible at -t 3 and detected at -t 2
//     (verified identical for .go files and .templ Go snippets);
//   - the identical body MASKS the shared signature tail: it extends the
//     match window past the signature, the window then contains statements,
//     and getUnitsIndexes discards the non-statement signature tokens as
//     noise. Divergent bodies surface the signature fragment instead (see
//     templ_signature_fragment_test.go). --show-suppressed cannot restore
//     what was never built — the group is masked, not suppressed.
//
// This is deliberate threshold semantics, not a bug. Any change to unit
// extraction or the statement gate must move these pins consciously.

const identicalBodyA = `package main

type Lang string

func metaC(count int, lang Lang) string {
	if count > 3 {
		return "many"
	}
	return strings.ToUpper(string(lang))
}
`

const identicalBodyB = `package main

type Lang string

func metaD(size int, lang Lang) string {
	if size > 3 {
		return "many"
	}
	return strings.ToUpper(string(lang))
}
`

var _ = Describe("Statement threshold counts effective statements", func() {
	It("reports nothing for identical 2-statement bodies at -t 3, even with --show-suppressed", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"c.templ": identicalBodyA,
			"d.templ": identicalBodyB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "3", "--show-suppressed")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("Found total 0 clone groups"))
	})

	It("detects the same pair at -t 2", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"c.templ": identicalBodyA,
			"d.templ": identicalBodyB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "2")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("1 clone group"))
	})
})
