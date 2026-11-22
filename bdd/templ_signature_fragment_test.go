package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Pins for the signature-fragment suppression (2026-11-22, webphone -t 3
// finding): .templ helper functions whose signatures merely END alike —
// `..., lang Lang) string` — surface the shared signature tail as its own
// clone fragment at low thresholds. Identical signature parts are not
// deduplicatable (the bodies differ; a signature cannot be extracted), so
// the group is suppressed as `signature-fragment`. With
// --disable-pattern signature-fragment the group returns.

const signatureTailA = `package main

type Lang string

func metaA(count int, lang Lang) string {
	if count > 3 {
		return "many"
	}
	return strings.ToUpper(string(lang))
}
`

const signatureTailB = `package main

type Lang string

func metaB(size int, lang Lang) string {
	switch size {
	case 1:
		return "one"
	}
	return strings.ToLower(string(lang))
}
`

var _ = Describe("Templ signature-fragment suppression", func() {
	It("suppresses shared Go-snippet signature tails at -t 3", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": signatureTailA,
			"b.templ": signatureTailB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "3")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("0 shown"))
		Expect(string(output)).ToNot(ContainSubstring("lang Lang"))
	})

	It("still surfaces the group with --show-suppressed and names the label", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": signatureTailA,
			"b.templ": signatureTailB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "3", "--show-suppressed", "--explain")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("lang Lang"))
		Expect(string(output)).To(ContainSubstring("signature-fragment"))
	})

	It("restores the group only with both fragment patterns disabled", func() {
		// Defense in depth: the one-line shape is also caught by the
		// trivial-fragment shape label, so restoring requires opting out of
		// both suppression layers.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"a.templ": signatureTailA,
			"b.templ": signatureTailB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl(
			"--threshold", "3",
			"--disable-pattern", "signature-fragment",
			"--disable-pattern", "trivial-fragment",
		)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("lang Lang"))
		Expect(string(output)).ToNot(ContainSubstring("0 shown"))
	})
})
