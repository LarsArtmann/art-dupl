package bdd

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// genericsDivergentCode holds two structurally identical functions whose local
// variables carry different concrete types (First vs Second at two positions).
// This is the shape Go generics can eliminate, so --suggest-generics must
// annotate the group.
const genericsDivergentA = `package main

type First struct{ Value int }

func SumFirst(f First, g First, times int) int {
	total := f.Value + g.Value
	for i := 0; i < times; i++ {
		total += i
	}
	return total
}
`

const genericsDivergentB = `package main

type Second struct{ Value int }

func SumSecond(s Second, t Second, times int) int {
	total := s.Value + t.Value
	for i := 0; i < times; i++ {
		total += i
	}
	return total
}
`

// genericsShortCode has two divergent type positions but each clone is a
// single line, below the default suggest-generics minimum line gate.
const genericsShortA = `package main

type Alpha struct{ V int }

func AddA(x Alpha, y Alpha) int { return x.V + y.V }
`

const genericsShortB = `package main

type Beta struct{ V int }

func AddB(x Beta, y Beta) int { return x.V + y.V }
`

var _ = Describe("Suggest-Generics Enhancer", func() {
	It("should annotate type-divergent clones with a generics hint", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"generics_a.go": genericsDivergentA,
			"generics_b.go": genericsDivergentB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--suggest-generics")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("found"))
		Expect(string(output)).To(ContainSubstring("generics:"))
		Expect(string(output)).To(ContainSubstring("same algorithm, different types"))
	})

	It("should not annotate clones without --suggest-generics", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"generics_a.go": genericsDivergentA,
			"generics_b.go": genericsDivergentB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("found"))
		Expect(string(output)).ToNot(ContainSubstring("generics:"))
	})

	It("should not annotate clones below the minimum line gate", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"generics_short_a.go": genericsShortA,
			"generics_short_b.go": genericsShortB,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl(
			"--threshold", "1", "--suggest-generics", "--no-actionability",
		)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("found"))
		Expect(string(output)).ToNot(ContainSubstring("generics:"))
	})

	It("should not annotate identical-type clones", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles(
			[]string{"generics_identical1.go", "generics_identical2.go"},
			genericsDivergentA,
		)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--suggest-generics")
		Expect(err).ToNot(HaveOccurred())

		outputStr := strings.ToLower(string(output))
		if strings.Contains(outputStr, "found 2 clones") {
			Expect(outputStr).ToNot(ContainSubstring("generics:"))
		}
	})
})
