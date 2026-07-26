package templ

import (
	"fmt"
	"regexp"
	"strings"
)

// identifierBeforeDot matches a lowercase-initial identifier followed by a dot,
// indicating field/method access on a local variable (e.g., `user.Name`).
// Package qualifiers like `fmt.Sprintf` also match, but this is acceptable
// because the normalization is consistent within a single template.
var identifierBeforeDot = regexp.MustCompile(`\b([a-z][a-zA-Z0-9]*)\.`)

// normalizeExprValue canonicalizes local variable identifiers in a templ
// expression string. This enables Type-2 clone detection for templ
// expressions where only variable names differ:
//
//	Component(user.Name, count)    → Component(v0.Name, v1)
//	Component(person.Name, total)  → Component(v0.Name, v1)
//
// The callee name (uppercase-initial) is preserved. Only lowercase-initial
// identifiers before '.' are normalized, as these are local variables
// accessing fields/methods.
func normalizeExprValue(expr string, symbols map[string]string) string {
	if symbols == nil {
		return expr
	}

	return identifierBeforeDot.ReplaceAllStringFunc(expr, func(match string) string {
		name := match[:len(match)-1]
		if isReservedWord(name) {
			return match
		}

		canonical, ok := symbols[name]
		if !ok {
			canonical = fmt.Sprintf("v%d", len(symbols))
			symbols[name] = canonical
		}

		return canonical + "."
	})
}

// reservedGoWords are Go keywords/builtins that should not be normalized
// even if they appear before a dot (e.g., `len(x)` doesn't have a dot,
// but `copy(dst, src)` could theoretically match).
var reservedGoWords = map[string]bool{
	"len": true, "cap": true, "copy": true, "new": true, "make": true,
	"append": true, "delete": true, "panic": true, "print": true,
	"println": true, "complex": true, "real": true, "imag": true,
	"clear": true, "min": true, "max": true,
}

func isReservedWord(name string) bool {
	return reservedGoWords[name]
}

// extractCalleeNameNormalized extracts the callee name and normalizes the
// expression arguments for semantic comparison. Returns the callee name
// (unchanged) and the normalized full expression.
func extractCalleeNameNormalized(expr string, symbols map[string]string) (callee, normalized string) {
	normalized = normalizeExprValue(strings.TrimSpace(expr), symbols)
	callee = extractCalleeName(normalized)

	return callee, normalized
}
