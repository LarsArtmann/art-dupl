package printer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TypeDivergence describes a single position where corresponding nodes across
// clone instances have different VarType values. This is the signal that a clone
// group is a generics-extraction candidate.
type TypeDivergence struct {
	Position int
	TypeA    string
	TypeB    string
}

// MinDivergentPositions is the minimum number of distinct structural positions
// that must have differing concrete types across clone instances before a group
// qualifies as a generics-extraction candidate. Single-position differences
// are dominated by shallow idiom noise (one differently-typed call result)
// rather than genuine same-algorithm-different-types duplication.
//
// Canonical definition: syntax.MinDivergentPositions (shared with the
// combined type-aware + suggest-generics mode's structural gate).
const MinDivergentPositions = syntax.MinDivergentPositions

// ClassifyGenericsCandidate examines the VarType fields across corresponding
// CloneNode positions in a clone group. When instances have different
// non-empty VarType values at the same structural positions, the group is a
// generics-extraction candidate: the algorithm is identical but the concrete
// types differ, which is exactly what Go generics can eliminate.
//
// At least MinDivergentPositions distinct positions must diverge; fewer is
// treated as noise.
//
// Returns (true, hint) when the group is a generics-extraction candidate.
// The hint summarizes the type differences for user-facing output.
func ClassifyGenericsCandidate(seqs [][]*domain.CloneNode) (bool, string) {
	if len(seqs) < 2 {
		return false, ""
	}

	base := flattenCloneNodes(seqs[0])

	var divergences []TypeDivergence

	for i := 1; i < len(seqs); i++ {
		other := flattenCloneNodes(seqs[i])

		maxLen := min(len(base), len(other))

		for j := range maxLen {
			a := base[j].VarType
			b := other[j].VarType

			if a == "" || b == "" {
				continue
			}

			if a != b {
				divergences = append(divergences, TypeDivergence{
					Position: j,
					TypeA:    a,
					TypeB:    b,
				})
			}
		}
	}

	if len(distinctPositions(divergences)) < MinDivergentPositions {
		return false, ""
	}

	if allInstantiationPairs(divergences) {
		return true, formatInstantiationOnlyHint(divergences)
	}

	return true, formatGenericsHint(divergences)
}

// allInstantiationPairs reports whether every divergence is a pair of
// instantiations of the SAME base type with different type arguments
// (list[int] vs list[string]). Detectable without the generic-decl index: the
// strings share their pre-bracket prefix and both carry type arguments.
func allInstantiationPairs(divs []TypeDivergence) bool {
	if len(divs) == 0 {
		return false
	}

	for _, d := range divs {
		baseA, argsA := splitInstantiation(d.TypeA)
		baseB, argsB := splitInstantiation(d.TypeB)

		if baseA == "" || baseA != baseB || argsA == argsB {
			return false
		}
	}

	return true
}

// splitInstantiation splits a type string into its pre-bracket base and type
// arguments ("[]list[int]" → "list", "int"; qualifier prefixes are stripped
// first). Returns ("", "") when the string carries no type arguments.
func splitInstantiation(typeStr string) (string, string) {
	qualified := golang.StripQualifierPrefixes(typeStr)

	cut := strings.Index(qualified, "[")
	if cut <= 0 || !strings.HasSuffix(qualified, "]") {
		return "", ""
	}

	return qualified[:cut], qualified[cut+1 : len(qualified)-1]
}

// formatInstantiationOnlyHint renders the T9 hint for groups whose only
// divergence is the type arguments of one generic: actionable phrasing that
// names the instantiations instead of listing them as unrelated differences.
func formatInstantiationOnlyHint(divs []TypeDivergence) string {
	parts := make([]string, 0, len(divs))
	seen := make(map[string]bool, len(divs))

	for _, d := range divs {
		key := divergenceKey(d)
		if seen[key] {
			continue
		}

		seen[key] = true

		parts = append(parts, fmt.Sprintf(
			"`%s` vs `%s`",
			shortenTypeString(d.TypeA),
			shortenTypeString(d.TypeB),
		))

		if len(parts) >= 3 {
			break
		}
	}

	return "differs only in generic instantiation: " + strings.Join(parts, "; ")
}

// usesIndexedGenericInstantiation reports whether any type carried by the
// sequences instantiates a generic declared in the run-wide index (plan T6).
// Clones that already sit on shared generics (the per-exporter
// `type Option = catalog.Option[Exporter]` wrapper class) do not need a NEW
// abstraction — suggesting one risks over-engineering, so the hint carries a
// note instead. Reuse-classified groups (every divergence is an
// instantiation of one indexed generic) are excluded by the caller: they get
// the stronger "Reuse existing generic" suggestion.
func usesIndexedGenericInstantiation(seqs [][]*domain.CloneNode, index map[string]int) bool {
	if len(index) == 0 {
		return false
	}

	for _, seq := range seqs {
		for _, n := range seq {
			if nodeInstantiatesIndexed(n, index) {
				return true
			}
		}
	}

	return false
}

func nodeInstantiatesIndexed(n *domain.CloneNode, index map[string]int) bool {
	if n.VarType != "" && golang.LookupGenericDecl(n.VarType, index) != "" {
		return true
	}

	for _, child := range n.Children {
		if nodeInstantiatesIndexed(child, index) {
			return true
		}
	}

	return false
}

// alreadyGenericNote is appended to the generics hint when the clones already
// instantiate shared generics but are not pure reuse candidates.
const alreadyGenericNote = "; already instantiates shared generics — further extraction may be over-engineering"

// ClassifyReuseExistingGeneric reports whether a generics candidate's type
// divergences are dominated by instantiations of one already-declared generic.
// When they are, the user should reuse that generic instead of extracting a
// new abstraction: the clones differ only in type arguments
// (Box[int] vs Box[string]), which the existing parameterization already
// covers. This is the reuse-vs-extract distinction from the
// generics-precision cluster (plan T5): "reuse pkg.Box[T1]" is actionable with
// zero new abstraction; "extract to generic" for the same group would
// duplicate an existing generic.
//
// Verdict rules, evaluated over every divergent position:
//   - A position whose two sides are instantiations of the SAME indexed
//     generic anchors the verdict and records the observed argument pair.
//   - A plain-type divergence is allowed only when it matches a recorded
//     argument pair — that is how a type argument propagates to related
//     locals (Box[int] naturally pairs with a value of type int).
//   - Any other divergence (two unrelated plain types, instantiations of
//     different generics, unindexed names) means the group parameterizes
//     genuinely different types: extraction, not reuse.
//
// index is the generic-declaration index (name → type-param count) collected
// from the run's pre-loaded ASTs (golang.CollectGenericDecls). An empty index
// disables the verdict. Returns the resolved declaration key and its type
// parameter count when the reuse verdict fires.
func ClassifyReuseExistingGeneric(
	seqs [][]*domain.CloneNode,
	index map[string]int,
) (string, int, bool) {
	if len(index) == 0 || len(seqs) < 2 {
		return "", 0, false
	}

	divs := collectDivergences(seqs)
	if len(divs) == 0 {
		return "", 0, false
	}

	// Pass 1: anchor positions — instantiations of one indexed generic.
	resolved := ""

	argPairs := make([][2]string, 0, len(divs))

	for _, d := range divs {
		keyA := golang.LookupGenericDecl(d.TypeA, index)

		keyB := golang.LookupGenericDecl(d.TypeB, index)
		if keyA == "" || keyA != keyB {
			continue
		}

		if resolved == "" {
			resolved = keyA
		} else if resolved != keyA {
			return "", 0, false
		}

		argPairs = append(argPairs, [2]string{typeArgsOf(d.TypeA), typeArgsOf(d.TypeB)})
	}

	if resolved == "" {
		return "", 0, false
	}

	// Pass 2: every non-anchor divergence must be a propagated argument pair.
	for _, d := range divs {
		if golang.LookupGenericDecl(d.TypeA, index) != "" {
			continue
		}

		if !matchesArgPair(d.TypeA, d.TypeB, argPairs) {
			return "", 0, false
		}
	}

	return resolved, index[resolved], true
}

// collectDivergences gathers every position where corresponding nodes across
// clone instances carry different non-empty VarType values.
func collectDivergences(seqs [][]*domain.CloneNode) []TypeDivergence {
	base := flattenCloneNodes(seqs[0])

	var divergences []TypeDivergence

	for i := 1; i < len(seqs); i++ {
		other := flattenCloneNodes(seqs[i])

		for j := range min(len(base), len(other)) {
			a := base[j].VarType
			b := other[j].VarType

			if a == "" || b == "" || a == b {
				continue
			}

			divergences = append(divergences, TypeDivergence{Position: j, TypeA: a, TypeB: b})
		}
	}

	return divergences
}

// typeArgsOf returns the type-argument substring of an instantiation
// ("pkg.Box[int]" → "int"); the input is assumed to end with a bracketed
// argument list (the caller resolved it via LookupGenericDecl).
func typeArgsOf(typeStr string) string {
	base := typeStr
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '[' {
			return base[i+1 : len(base)-1]
		}
	}

	return ""
}

// matchesArgPair reports whether a plain type divergence (a vs b) equals a
// recorded argument pair in either orientation.
func matchesArgPair(a, b string, argPairs [][2]string) bool {
	for _, p := range argPairs {
		if (a == p[0] && b == p[1]) || (a == p[1] && b == p[0]) {
			return true
		}
	}

	return false
}

// distinctPositions returns the set of structural positions that appear in
// the given divergences.
func distinctPositions(divs []TypeDivergence) map[int]bool {
	positions := make(map[int]bool, len(divs))
	for _, d := range divs {
		positions[d.Position] = true
	}

	return positions
}

// pkgPathSegmentRe matches a Go import path segment followed by a slash,
// e.g. "github.com/larsartmann/" in "github.com/larsartmann/erraudit/pkg.Type".
// Used to strip package paths to just the last segment for readability.
var pkgPathSegmentRe = regexp.MustCompile(`[\w.-]+/`)

// shortenTypeString strips Go import path prefixes from type strings,
// keeping only the last path segment. For example:
//   - github.com/pkg.Type → pkg.Type
//   - []github.com/pkg.Type → []pkg.Type
//   - *github.com/pkg.Type → *pkg.Type
func shortenTypeString(s string) string {
	for {
		loc := pkgPathSegmentRe.FindStringIndex(s)
		if loc == nil {
			return s
		}

		s = s[:loc[0]] + s[loc[1]:]
	}
}

// formatGenericsHint produces a concise summary of the type differences.
// Shows up to 3 unique type-pair divergences to keep the hint readable.
// Pair ordering is canonicalized (A vs B == B vs A) so reversed pairs in
// different instance comparisons do not duplicate.
func formatGenericsHint(divs []TypeDivergence) string {
	seen := make(map[string]bool)

	var parts []string

	for _, d := range divs {
		key := divergenceKey(d)
		if seen[key] {
			continue
		}

		seen[key] = true

		parts = append(parts, fmt.Sprintf("%s vs %s", shortenTypeString(d.TypeA), shortenTypeString(d.TypeB)))

		if len(parts) >= 3 {
			break
		}
	}

	hint := "same algorithm, different types: " + strings.Join(parts, "; ")
	if len(divs) > 3 {
		hint += fmt.Sprintf(" (%d type differences total)", len(divs))
	}

	return hint
}

// divergenceKey builds a canonical dedup key for a type divergence so that
// "A vs B" and "B vs A" (observed from different instance pairings) collapse
// to a single entry.
func divergenceKey(d TypeDivergence) string {
	if d.TypeB < d.TypeA {
		return d.TypeB + " | " + d.TypeA
	}

	return d.TypeA + " | " + d.TypeB
}

// flattenCloneNodes walks a CloneNode subtree in pre-order and returns a flat
// slice. Corresponding positions in flattened slices from structurally-identical
// clones represent the same AST node — so VarType comparison at the same index
// is meaningful.
func flattenCloneNodes(nodes []*domain.CloneNode) []*domain.CloneNode {
	result := make([]*domain.CloneNode, 0, len(nodes))

	for _, n := range nodes { //art-dupl:accept recursive twin of flattenChildren
		result = append(result, n)
		result = append(result, flattenChildren(n)...)
	}

	return result
}

func flattenChildren(n *domain.CloneNode) []*domain.CloneNode {
	result := make([]*domain.CloneNode, 0, len(n.Children))

	for _, c := range n.Children { //art-dupl:accept recursive twin of flattenCloneNodes
		result = append(result, c)
		result = append(result, flattenChildren(c)...)
	}

	return result
}
