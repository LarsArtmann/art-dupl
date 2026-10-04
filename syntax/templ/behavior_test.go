package templ

import (
	"fmt"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// tokenStream serializes a parsed templ tree into its flat token values, the
// stream the suffix tree actually matches on.
func tokenStream(t *testing.T, src string, mode golang.DetectionMode) []suffixtree.TokenValue {
	t.Helper()

	if !strings.Contains(src, "package main") {
		src = fmt.Sprintf(wrapComponent, src)
	}

	node, _, err := ParseBytesWithMode("test.templ", []byte(src), mode)
	if err != nil {
		t.Fatalf("ParseBytesWithMode error = %v", err)
	}

	tokens := syntax.Serialize(node)

	stream := make([]suffixtree.TokenValue, 0, len(tokens))
	for _, tok := range tokens {
		stream = append(stream, tok.Val())
	}

	return stream
}

func assertStreamsEqual(t *testing.T, srcA, srcB string, mode golang.DetectionMode) {
	t.Helper()

	a := tokenStream(t, srcA, mode)
	b := tokenStream(t, srcB, mode)

	if len(a) != len(b) {
		t.Fatalf("token streams differ in length: %d vs %d\nA=%v\nB=%v", len(a), len(b), a, b)
	}

	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("token %d differs: %d vs %d\nA=%v\nB=%v", i, a[i], b[i], a, b)
		}
	}
}

func assertStreamsDiffer(t *testing.T, srcA, srcB string, mode golang.DetectionMode) {
	t.Helper()

	a := tokenStream(t, srcA, mode)
	b := tokenStream(t, srcB, mode)

	if len(a) == len(b) {
		same := true

		for i := range a {
			if a[i] != b[i] {
				same = false

				break
			}
		}

		if same {
			t.Fatalf("token streams are identical but must differ\nA=%v", a)
		}
	}
}

const wrapComponent = "package main\n\ntempl X(user User) {\n%s\n}\n"

// --- Conditions: the false-positive class ---

func TestConditionDifferencesPreventClone(t *testing.T) {
	t.Parallel()

	a := "if user.IsAdmin {\n<div>tools</div>\n}"
	b := "if user.IsCompletelyUnrelatedCondition {\n<div>tools</div>\n}"

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

func TestRenamedConditionLocalStillClones(t *testing.T) {
	t.Parallel()

	a := "if user.IsAdmin {\n<div>tools</div>\n}"
	b := "if person.IsAdmin {\n<div>tools</div>\n}"

	assertStreamsEqual(t, a, b, golang.DetectionModeSemantic)
}

func TestStringExpressionDifferencesPreventClone(t *testing.T) {
	t.Parallel()

	a := "<p>{ item.Name }</p>"
	b := "<p>{ item.DeletedAt }</p>"

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

func TestAttributeExpressionDifferencesPreventClone(t *testing.T) {
	t.Parallel()

	a := `<a href={ user.Profile }>link</a>`
	b := `<a href={ user.Settings }>link</a>`

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

// --- If/else structure: the flattening class ---

func TestElseStructureDiffersFromSequentialBodies(t *testing.T) {
	t.Parallel()

	ifElse := "if user.Ok {\n<div>A</div>\n} else {\n<div>B</div>\n}"
	bothInThen := "if user.Ok {\n<div>A</div>\n<div>B</div>\n}"

	assertStreamsDiffer(t, ifElse, bothInThen, golang.DetectionModeSemantic)
}

func TestElseIfChainStructurePreserved(t *testing.T) {
	t.Parallel()

	chain := "if a.B {\n<x/>\n} else if c.D {\n<y/>\n} else {\n<z/>\n}"
	flat := "if a.B {\n<x/>\n}\nif c.D {\n<y/>\n}\nif true {\n<z/>\n}"

	assertStreamsDiffer(t, chain, flat, golang.DetectionModeSemantic)
}

// --- Switch: default/case/fallthrough distinctions ---

func TestDefaultCaseDiffersFromValueCase(t *testing.T) {
	t.Parallel()

	srcDefault := `switch user.Kind {
default:
	<p>other</p>
}`
	srcCase := `switch user.Kind {
case "other":
	<p>other</p>
}`

	assertStreamsDiffer(t, srcDefault, srcCase, golang.DetectionModeSemantic)
}

func TestSwitchTagDifferencePreventsClone(t *testing.T) {
	t.Parallel()

	a := `switch user.Kind {
case "admin":
	<p>admin</p>
}`
	b := `switch user.Role {
case "admin":
	<p>admin</p>
}`

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

// --- Component calls: arguments participate ---

func TestComponentCallArgCountPreventsClone(t *testing.T) {
	t.Parallel()

	a := "@Card(user)"
	b := "@Card(user, admin, extra)"

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

func TestComponentCallRenamedArgsStillClone(t *testing.T) {
	t.Parallel()

	a := "@Card(user.Name, count)"
	b := "@Card(person.Name, total)"

	assertStreamsEqual(t, a, b, golang.DetectionModeSemantic)
}

func TestComponentCallFieldDifferencePreventsClone(t *testing.T) {
	t.Parallel()

	a := "@Card(user.Name)"
	b := "@Card(user.Email)"

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

// --- Conditional attributes ---

func TestConditionalAttributeBranchesVisible(t *testing.T) {
	t.Parallel()

	a := `<p if user.Ok { class="yes" } else { class="no" }>x</p>`
	b := `<p if user.Ok { class="yes" }>x</p>`

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

func TestConditionalAttributeConditionVisible(t *testing.T) {
	t.Parallel()

	a := `<p if user.Ok { class="yes" }>x</p>`
	b := `<p if user.Gone { class="yes" }>x</p>`

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

// --- CSS ---

func TestCSSPropertyNamesDistinguish(t *testing.T) {
	t.Parallel()

	cssA := "package main\n\ncss CardA() {\n\tcolor: red;\n\tbackground-color: blue;\n}\n"
	cssB := "package main\n\ncss CardB() {\n\tcolor: green;\n\tborder-radius: 4px;\n}\n"

	assertStreamsDiffer(t, cssA, cssB, golang.DetectionModeSemantic)
}

func TestCSSDeclarationStatementMarked(t *testing.T) {
	t.Parallel()

	node, _, err := ParseBytesWithMode("test.templ", []byte(
		"package main\n\ncss C() {\n\tcolor: red;\n}\n",
	), golang.DetectionModeSemantic)
	if err != nil {
		t.Fatalf("parse error = %v", err)
	}

	if len(node.Children) != 1 {
		t.Fatalf("expected 1 top-level child, got %d", len(node.Children))
	}

	if !node.Children[0].Statement {
		t.Error("CSSDeclaration must be Statement=true so it forms one composite unit")
	}

	if syntax.DecodeBaseType(node.Children[0].Type) != CSSDeclaration {
		t.Errorf("base type must stay CSSDeclaration for DecodeBaseType consumers, got %d", node.Children[0].Type)
	}
}

// --- GoCode ---

func TestGoCodeStatementsDetected(t *testing.T) {
	t.Parallel()

	a := "{{ x := user.Count; render(x) }}"
	b := "{{ y := person.Count; render(y) }}"

	assertStreamsEqual(t, a, b, golang.DetectionModeSemantic)
}

func TestGoCodeDifferencePreventsClone(t *testing.T) {
	t.Parallel()

	a := "{{ x := user.Count }}"
	b := "{{ x := user.Total }}"

	assertStreamsDiffer(t, a, b, golang.DetectionModeSemantic)
}

// --- Top-level Go code ---

func TestTopLevelGoHelpersVisible(t *testing.T) {
	t.Parallel()

	helper := "func formatLabel(s string) string {\n\treturn strings.ToUpper(s)\n}\n"

	a := "package main\n\nimport \"strings\"\n\n" + helper + "\ntempl A() {\n<div>x</div>\n}\n"
	b := "package main\n\nimport \"strings\"\n\n" + helper + "\ntempl B() {\n<div>y</div>\n}\n"

	// The duplicated helper body's statements must enter both token streams
	// identically (the declaration tokens legitimately differ by name).
	streamA := tokenStream(t, a, golang.DetectionModeSemantic)
	streamB := tokenStream(t, b, golang.DetectionModeSemantic)

	helperStream := tokenStream(t, "package main\n\n"+helper, golang.DetectionModeSemantic)
	helperBody := helperStream[1:] // drop the File token

	if !containsInOrder(streamA, helperBody) || !containsInOrder(streamB, helperBody) {
		t.Errorf(
			"helper body statements must appear in both files' token streams\nA=%v\nB=%v\nhelper=%v",
			streamA, streamB, helperBody,
		)
	}
}

// containsInOrder reports whether needle appears in haystack as a subsequence.
func containsInOrder(haystack, needle []suffixtree.TokenValue) bool {
	j := 0
	for i := range haystack {
		if haystack[i] == needle[j] {
			j++
			if j == len(needle) {
				return true
			}
		}
	}

	return false
}

// --- Fallback: hybrid templ-in-Go source stays conservative ---

func TestHybridGoCodeFallsBackToOpaqueToken(t *testing.T) {
	t.Parallel()

	// templ tolerates markup inside GoCode that go/parser rejects; the
	// transformer must not fail the whole parse.
	src := "{{ if user.Ok { <b>yes</b> } }}"

	wrapped := fmt.Sprintf(wrapComponent, "\t"+src)

	node, _, err := ParseBytesWithMode("test.templ", []byte(wrapped), golang.DetectionModeSemantic)
	if err != nil {
		t.Fatalf("hybrid GoCode must not fail the parse: %v", err)
	}

	if node == nil {
		t.Fatal("expected a tree")
	}
}

// --- Modes through the full pipeline ---

func TestExactModeConditionVerbatim(t *testing.T) {
	t.Parallel()

	a := "if user.IsAdmin {\n<div>tools</div>\n}"
	b := "if person.IsAdmin {\n<div>tools</div>\n}"

	// Exact mode: renamed locals do NOT match (verbatim hashing).
	assertStreamsDiffer(t, a, b, golang.DetectionModeExact)
}

func TestStructuralModeIgnoresConditionsAndNames(t *testing.T) {
	t.Parallel()

	a := "if user.IsAdmin {\n<div>tools</div>\n}"
	b := "if person.Whatever {\n<section>stuff</section>\n}"

	assertStreamsEqual(t, a, b, golang.DetectionModeStructural)
}
