package bdd

import (
	"sort"
	"strings"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parallel Search (--search-workers)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	It("should produce identical results to sequential search", func() {
		dupCode := `package main

func process(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	result := 0
	for _, b := range data {
		result += int(b)
	}
	return result, nil
}

func validate(input string) bool {
	if len(input) < 3 {
		return false
	}
	count := 0
	for _, c := range input {
		if c == 'x' {
			count++
		}
	}
	return count > 0
}`

		err := setup.CreateDuplicateFiles(
			[]string{"alpha.go", "beta.go", "gamma.go"},
			dupCode,
		)
		Expect(err).NotTo(HaveOccurred())

		sequentialOutput, err := setup.RunArtDupl("--threshold", "5", "--quiet")
		Expect(err).ToNot(HaveOccurred())

		parallelOutput, err := setup.RunArtDupl("--threshold", "5", "--quiet", "--search-workers", "4")
		Expect(err).ToNot(HaveOccurred())

		seqSorted := sortLines(string(sequentialOutput))
		parSorted := sortLines(string(parallelOutput))

		Expect(parSorted).To(Equal(seqSorted),
			"parallel search output should match sequential search output (ignoring line order)")

		Expect(seqSorted).To(ContainSubstring("alpha.go"))
		Expect(parSorted).To(ContainSubstring("alpha.go"))
	})

	It("should produce identical results with workers=1 (sequential fallback)", func() {
		dupCode := `package main

func helper(x int) int {
	return x*2 + 1
}`

		err := setup.CreateDuplicateFiles(
			[]string{"h1.go", "h2.go"},
			dupCode,
		)
		Expect(err).NotTo(HaveOccurred())

		defaultOutput, err := setup.RunArtDupl("--threshold", "1", "--quiet")
		Expect(err).ToNot(HaveOccurred())

		workers1Output, err := setup.RunArtDupl("--threshold", "1", "--quiet", "--search-workers", "1")
		Expect(err).ToNot(HaveOccurred())

		Expect(sortLines(string(workers1Output))).To(Equal(sortLines(string(defaultOutput))),
			"--search-workers 1 should produce identical output to default sequential search")
	})

	It("should handle context cancellation gracefully", func() {
		dupCode := `package main

func process(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	result := 0
	for _, b := range data {
		result += int(b)
	}
	return result, nil
}

func validate(input string) bool {
	if len(input) < 3 {
		return false
	}
	count := 0
	for _, c := range input {
		if c == 'x' {
			count++
		}
	}
	return count > 0
}`

		err := setup.CreateDuplicateFiles(
			[]string{"cancel1.go", "cancel2.go"},
			dupCode,
		)
		Expect(err).NotTo(HaveOccurred())

		// Parallel search should complete without error even under cancellation pressure
		_, err = setup.RunArtDupl("--threshold", "5", "--quiet", "--search-workers", "4")
		Expect(err).ToNot(HaveOccurred())
	})
})

func sortLines(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	sort.Strings(lines)

	return strings.Join(lines, "\n")
}
