package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// combinedSameTypeA/B are structurally identical with IDENTICAL local types
// (int, int, int): the type-aware pass matches them as a regular clone group.
const combinedSameTypeA = `package main

func pairSumInts(a int, b int) int {
	total := a + b
	steps := 3
	for i := 0; i < steps; i++ {
		total += i
	}
	return total
}
`

const combinedSameTypeB = `package main

func pairSumIntsAgain(x int, y int) int {
	total := x + y
	steps := 3
	for i := 0; i < steps; i++ {
		total += i
	}
	return total
}
`

// combinedReceiverNoiseClock/Meter differ in exactly ONE divergent type
// position (the receiver `c Clock` vs `m Meter`, used once at line 6). This is
// the shallow receiver-noise class --type-aware exists to eliminate. The
// combined mode must drop the full-body cross-type match (lines 6-9) while
// still reporting the same-type statement core (lines 7-9) from the
// type-aware pass; suggest-generics alone shows the full 6-9 group.
const combinedReceiverNoiseClock = `package main

type Clock struct{ Name string }

func renderClock(c Clock) string {
	parts := c.Name
	report := "ts:" + parts
	width := len(report) + 1
	return report[:width-1]
}
`

const combinedReceiverNoiseMeter = `package main

type Meter struct{ Name string }

func renderMeter(m Meter) string {
	parts := m.Name
	report := "ts:" + parts
	width := len(report) + 1
	return report[:width-1]
}
`

var _ = Describe("Combined Type-Aware + Suggest-Generics Mode", func() {
	It("reports same-type clones and generics candidates in one run", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"combined_pair_a.go":   combinedSameTypeA,
			"combined_pair_b.go":   combinedSameTypeB,
			"combined_family_a.go": genericsDivergentA,
			"combined_family_b.go": genericsDivergentB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--type-aware", "--suggest-generics")
		Expect(err).ToNot(HaveOccurred())

		outputStr := string(output)

		// The type-aware pass reports the identical-type pair as a regular group.
		Expect(outputStr).To(ContainSubstring("combined_pair_a.go"))
		Expect(outputStr).To(ContainSubstring("combined_pair_b.go"))

		// The type-erased pass reports the cross-type family with a generics hint.
		Expect(outputStr).To(ContainSubstring("combined_family_a.go"))
		Expect(outputStr).To(ContainSubstring("combined_family_b.go"))
		Expect(outputStr).To(ContainSubstring("generics:"))
		Expect(outputStr).To(ContainSubstring("same algorithm, different types"))
	})

	It("suppresses single-position receiver noise that suggest-generics alone shows", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"combined_noise_a.go": combinedReceiverNoiseClock,
			"combined_noise_b.go": combinedReceiverNoiseMeter,
		})
		Expect(err).NotTo(HaveOccurred())

		combined, err := setup.RunArtDupl("--threshold", "1", "--type-aware", "--suggest-generics")
		Expect(err).ToNot(HaveOccurred())

		// The full-body cross-type group (including the divergent receiver
		// statement at line 6) is dropped; the same-type core (lines 7-9)
		// survives via the type-aware pass.
		Expect(string(combined)).ToNot(ContainSubstring("combined_noise_a.go:6-9"))
		Expect(string(combined)).To(ContainSubstring("combined_noise_a.go:7-9"))

		// Suggest-generics alone matches the full body across the divergent
		// receiver types (single divergence: no generics hint).
		genericsOnly, err := setup.RunArtDupl("--threshold", "1", "--suggest-generics")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(genericsOnly)).To(ContainSubstring("combined_noise_a.go:6-9"))
	})

	It("falls back gracefully when type loading fails", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"combined_broken.go": "package main\n\nfunc broken() {\n\tx := undefinedType.Method()\n\t_ = x\n}\n",
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = setup.RunArtDupl("--threshold", "1", "--type-aware", "--suggest-generics")
		Expect(err).ToNot(HaveOccurred())
	})
})
