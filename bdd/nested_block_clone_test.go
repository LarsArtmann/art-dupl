package bdd

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Regression specs for the nested-block false-negative class (ADR-0023).
// Before the nested-statement token emission, two composite statements
// (for/if/switch/else-if) that shared leading statements but diverged deeper
// inside were invisible at EVERY threshold and detection mode — the tool
// reported "0 clones, Health A" on code containing a 26-line duplicated
// skeleton (docs/status/2026-09-13_15-45 report).

var _ = Describe("Nested block clone detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()
	})

	// loopSkeletonPair renders two functions whose loops share the first two
	// body statements and diverge in the third. loopKind picks the composite
	// statement so one table covers the whole masking class.
	loopSkeletonPair := func(loopKind string) (string, string) {
		const (
			fileA = `package main

func processA(items []int) int {
	total := 0
	for _, it := range items {
		total += it
		if it > 10 {
			total++
		}
		total *= 2
		total -= 7
	}
	return total
}`

			fileB = `package main

func processB(items []int) int {
	sum := 0
	for _, it := range items {
		sum += it
		if it > 10 {
			sum++
		}
		sum *= 2
		sum += 9
	}
	return sum
}`
		)

		switch loopKind {
		case "for loop":
			return fileA, fileB
		case "if statement":
			return `package main

func gradeA(score int) int {
	result := 0
	if score > 50 {
		result += 1
		result *= 2
		result++
	}
	return result
}`, `package main

func gradeB(score int) int {
	out := 0
	if score > 50 {
		out += 1
		out *= 2
		out -= 3
	}
	return out
}`
		case "switch statement":
			return `package main

func routeA(code int) int {
	r := 0
	switch code {
	case 1:
		r += 1
		r *= 2
		r++
	default:
		r += 1
		r *= 2
		r -= 7
	}
	return r
}`, `package main

func routeB(code int) int {
	out := 0
	switch code {
	case 1:
		out += 1
		out *= 2
		out--
	default:
		out += 1
		out *= 2
		out *= 9
	}
	return out
}`
		case "else-if chain":
			return `package main

func bucketA(v int) int {
	res := 0
	if v > 100 {
		res = 1
		res += v
	} else if v > 10 {
		res = 2
		res += v
	}
	return res
}`, `package main

func bucketB(v int) int {
	ans := 0
	if v > 100 {
		ans = 1
		ans += v
	} else if v > 10 {
		ans = 2
		ans += v
	}
	return ans
}`
		case "nested loops":
			return `package main

func sumGridA(grid [][]int) int {
	total := 0
	for _, row := range grid {
		for _, cell := range row {
			total += cell
			if cell%2 == 0 {
				total++
			}
			total *= 3
		}
	}
	return total
}`, `package main

func sumGridB(grid [][]int) int {
	sum := 0
	for _, row := range grid {
		for _, cell := range row {
			sum += cell
			if cell%2 == 0 {
				sum++
			}
			sum *= 3
		}
	}
	return sum
}`
		default:
			panic("unknown loop kind: " + loopKind)
		}
	}

	DescribeTable("clones sharing statements inside composite bodies are detected",
		func(loopKind string) {
			srcA, srcB := loopSkeletonPair(loopKind)

			err := setup.CreateTestFiles(map[string]string{"a.go": srcA, "b.go": srcB})
			Expect(err).NotTo(HaveOccurred())

			output, err := setup.RunArtDupl("--threshold", "2", "--no-actionability")
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("a.go"), "file A must appear in output for %s", loopKind)
			Expect(outputStr).To(ContainSubstring("b.go"), "file B must appear in output for %s", loopKind)
			Expect(outputStr).NotTo(ContainSubstring("Found total 0 clone groups"),
				"shared skeleton inside %s must produce clone groups", loopKind)
		},
		Entry("for loop body", "for loop"),
		Entry("if statement body", "if statement"),
		Entry("switch case bodies", "switch statement"),
		Entry("else-if chain links", "else-if chain"),
		Entry("nested loop bodies", "nested loops"),
	)

	It("detects the shared loop-body statements at the default threshold when the shared run is long enough", func() {
		// Five consecutive semantically distinct statements inside two loop
		// bodies survive -t 5; the divergence comes after the shared run.
		// The statements must differ in shape (operator encoding): literal
		// normalization would make `n += 1`..`n += 5` token-identical and
		// isCyclic would rightly reject the self-repeating run.
		loop := func(tail string) string {
			return fmt.Sprintf(`package main

func stretch(s []int) int {
	n := 0
	for range s {
		n += 1
		n -= 2
		n *= 3
		n /= 4
		n <<= 5
		%s
	}
	return n
}`, tail)
		}

		err := setup.CreateTestFiles(map[string]string{
			"a.go": loop("n += 6"),
			"b.go": loop("n *= 9"),
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "5", "--no-actionability")
		Expect(err).NotTo(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring("a.go"))
		Expect(outputStr).To(ContainSubstring("b.go"))
	})

	It("suppresses identical guard clones via actionability by default", func() {
		err := setup.CreateTestFiles(map[string]string{
			"guard1.go": `package main

func guard1(enabled bool) {
	if !enabled {
		return
	}
	use()
}

func use() {}`,
			"guard2.go": `package main

func guard2(active bool) {
	if !active {
		return
	}
	run()
}

func run() {}`,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1")
		Expect(err).NotTo(HaveOccurred())

		Expect(strings.Count(string(output), "found")).To(Equal(0),
			"identical guard-clause clones must stay suppressed by default, got: %s", string(output))
	})
})
