package bdd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Fixture pair for the instantiation-vs-indexing collision (plan T2,
// docs/planning/2026-10-11_05-06_generics-precision-from-branching-flow-lessons.md).
//
// `feed[i]` (indexing a package-level slice of funcs) and `feed[T]`
// (instantiating a generic function to a func value) serialize to the SAME
// token stream once locals and type params are alpha-normalized: both become
// IndexExpr{Ident("feed"), Ident("v0")}. The two files live in different
// packages (a var and a type cannot share a name within one package scope),
// so this is the narrowest cross-package shape that collides.
//
// The canonical numbering must align too: type params are declared before
// params (see normalizer.collectFunctionLocals), so pkgA needs a second
// param `j` for its body local `h` to land on the same canonical index as
// pkgB's `h` (which is preceded by type param T and param v).
const collisionIndexingPkg = `package pkga

var feed []func(int) string

func pick(i int, j int) string {
	h := feed[i]
	return h(j)
}
`

const collisionInstantiationPkg = `package pkgb

func feed[T any](v T) T { return v }

func pick[T any](v T) T {
	h := feed[T]
	return h(v)
}
`

// Minimal structural-mode pair: identifiers are ignored entirely, so any
// IndexExpr with the same shape matches.
const collisionStructuralIndexing = `package structa

func first(xs []int, i int) int {
	h := xs[i]
	return h
}
`

const collisionStructuralInstantiation = `package structb

func Feed[T any](v T) T { return v }

func first[T any](v T) T {
	h := Feed[T]
	return h(v)
}
`

var _ = Describe("Instantiation vs indexing collision", func() {
	// Note: the semantic-mode collision always requires two packages — a
	// package-level var and a package-level type/func cannot share a name
	// within one package scope, so no single-package variant can be built.
	It("semantic mode: cross-package slice indexing collides with generic instantiation", func() {
		// CURRENT: the false clone group IS reported — indexing `feed[i]`
		// and instantiation `feed[T]` hash identically because go/parser
		// cannot distinguish an index expression from a type instantiation
		// without go/types (types.Info.Instances).
		// EXPECTED: default semantic mode stays ambiguity-blind (documented
		// limitation, plan decision Q2); --type-aware disambiguates once
		// Instances is wired (plan T7). If T7 changes this default-mode
		// behavior, update this test AND the FEATURES.md limitation note.
		// IMPACT: rare in practice — requires the same base name for a
		// package-level var (one package) and a generic func/type (another
		// package) plus aligned canonical numbering, at threshold <= 2.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"pkga/a.go": collisionIndexingPkg,
			"pkgb/b.go": collisionInstantiationPkg,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--semantic", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("pkga/a.go"))
		Expect(string(output)).To(ContainSubstring("pkgb/b.go"))
	})

	It("structural mode: indexing and instantiation are equivalent by construction", func() {
		// CURRENT: identical to semantic outcome, but for a stronger
		// reason — structural mode ignores identifier names entirely, so
		// EVERY IndexExpr shape collides (`xs[i]` vs `Feed[T]` here).
		// EXPECTED: documented equivalence class of structural mode; no
		// fix is possible without reintroducing identifier hashing.
		// IMPACT: any structural-mode run over code that mixes slice
		// indexing and generic instantiation reports false groups.
		setup := CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"structa/a.go": collisionStructuralIndexing,
			"structb/b.go": collisionStructuralInstantiation,
		})
		Expect(err).NotTo(HaveOccurred())

		output, err := setup.RunArtDupl("--threshold", "1", "--structural", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		Expect(string(output)).To(ContainSubstring("structa/a.go"))
		Expect(string(output)).To(ContainSubstring("structb/b.go"))
	})
})
