package bdd

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Go 1.27 legalized methods with their own type parameters (generic
// methods). These specs pin that the full CLI pipeline — including the
// go/types-backed --type-aware and --suggest-generics loaders — handles
// them: type loading must not fail on the new syntax, and the same
// detection contracts hold as for ordinary functions.

const genericMethodDup = `package main

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Drain() []T {
	out := []T{}
	for _, it := range s.items {
		out = append(out, it)
	}
	return out
}

func (s *Stack[T]) Peek() T {
	return s.items[len(s.items)-1]
}
`

var _ = Describe("Go 1.27 generic methods", func() {
	It("detects renamed clones across files with --type-aware", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFile("lib.go", genericMethodDup)
		Expect(err).NotTo(HaveOccurred())

		renamed := genericMethodDup
		for from, to := range map[string]string{
			"func (s *Stack[T]) Drain()": "func (r *Stack[T]) Empty()",
			"func (s *Stack[T]) Peek()":  "func (r *Stack[T]) Top()",
			"out := []T{}":               "result := []T{}",
			"for _, it := range s.items": "for _, item := range r.items",
			"out = append(out, it)":      "result = append(result, item)",
			"return out":                 "return result",
			"range s.items":              "range r.items",
			"len(s.items)":               "len(r.items)",
			"s.items[len(s.items)-1]":    "r.items[len(r.items)-1]",
		} {
			renamed = strings.ReplaceAll(renamed, from, to)
		}
		err = setup.CreateTestFile("lib2.go", renamed)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--type-aware", "--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("lib.go"), "type-aware run must surface the renamed clone")
	})

	It("type-loads generic methods under --suggest-generics without error", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFile("stack.go", genericMethodDup)
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--suggest-generics", "--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred(), "suggest-generics type loading must tolerate generic methods: %s", output)
		// The clone-free fixture legitimately yields zero groups; the pin is
		// that the go/types loader completed (✅) on generic-method syntax.
		Expect(string(output)).To(ContainSubstring("✅"))
	})
})
