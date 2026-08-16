package actionability

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// maxTestPreambleStatements bounds how many statements a test preamble clone
// may span and still count as boilerplate. Real test logic grows past this;
// setup-only preambles never do.
const maxTestPreambleStatements = 6

// testContextDirSuffixes are directory-name suffixes marking Go test helper
// packages (backuptest/, commandtest/, adttest/, testutil/). Clone preambles
// from these packages are as inextractable as _test.go preambles: the t
// receiver is still a *testing.T threaded through from real tests.
var testContextDirSuffixes = []string{ //nolint:gochecknoglobals // static name set
	"test", "testutil", "tests", "testing",
}

// allFromTestContext checks that all nodes in a sequence originate from
// _test.go files or from test-named helper packages. Broader than
// allFromTestFile: suite/harness helpers often live in regular files inside
// packages like foo_test or footest.
func allFromTestContext(seq []*domain.CloneNode) bool {
	for _, n := range seq {
		if !isTestContextFile(n.Filename) {
			return false
		}
	}

	return true
}

// isTestContextFile reports whether a path is a _test.go file or sits inside
// a test-named package directory.
func isTestContextFile(filename string) bool {
	if strings.HasSuffix(filename, "_test.go") {
		return true
	}

	dir := filepath.Dir(filepath.ToSlash(filename))

	for part := range strings.SplitSeq(dir, "/") {
		for _, suffix := range testContextDirSuffixes {
			if strings.HasSuffix(part, suffix) {
				return true
			}
		}
	}

	return false
}

// preambleDisallowedCalls are call names that indicate real test logic rather
// than setup. If any non-first statement of a candidate preamble calls one of
// these, the sequence stays actionable.
var preambleDisallowedCalls = []string{ //nolint:gochecknoglobals // static name set
	// Subtest launching is structure, not setup.
	"Run",
	// testing.T failure/logging methods.
	"Error", "Errorf", "Fatal", "Fatalf", "Fail", "FailNow", "Log", "Logf", //nolint:goconst
	"Skip", "Skipf",
	// Assertion families (testify, gomega, ginkgo, stdlib checks).
	"Expect", "Require", "Assert", "Check", "So", "Should", //nolint:goconst
	"NoError", "ErrorIs", "ErrorAs", "Equal", "NotEqual", "True", "False",
	"Nil", "NotNil", "Zero", "Len", "Empty", "Contains", "ElementsMatch",
	"HaveLen", "BeEmpty", "BeNil", "BeTrue", "BeFalse", "BeZero", "HaveOccurred",
	"NotTo", "To", "ToNot",
}

// isTestPreamblePattern reports whether every clone is a test-file preamble:
// a leading t.Parallel() / t.Helper() call followed only by trivial setup
// statements (assignments, declarations, non-assertion calls). This is the
// universal Go test opening; each test independently chooses its setup, so
// no extraction is possible or desirable.
//
// Source: go-cqrs-lite feedback (Finding 2) — 9 groups / ~190 occurrences of
// "t.Parallel() + one setup line" buried 24 actionable groups at -t 2.
//
// Guards against over-suppression:
//   - all clones must originate from test context files (_test.go or a
//     test-named helper package like backuptest/, commandtest/, testutil/)
//   - sequences containing control flow never match (statement-type allowlist)
//   - sequences containing assertion or subtest calls never match
func isTestPreamblePattern(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if !allFromTestContext(seq) || len(seq) < 2 || len(seq) > maxTestPreambleStatements {
			return false
		}

		unwrapped := unwrapExprStmt(seq[0])

		if !isTestFrameworkCall(unwrapped) {
			return false
		}

		for _, stmt := range seq[1:] {
			if !isTrivialSetupStatement(stmt) {
				return false
			}
		}

		return true
	})
}

// isTrivialSetupStatement reports whether a statement can appear in a test
// preamble: an assignment, a declaration, or a bare call that is not an
// assertion/subtest/failure method.
func isTrivialSetupStatement(stmt *domain.CloneNode) bool {
	switch stmt.BaseType {
	case golang.AssignStmt, golang.ValueSpec, golang.GenDecl, golang.DeclStmt:
		return true
	case golang.ExprStmt:
		call := unwrapExprStmt(stmt)

		if call.BaseType != golang.CallExpr {
			return false
		}

		return !callHasAnyName(call, preambleDisallowedCalls)
	default:
		return false
	}
}

// callHasAnyName reports whether a CallExpr's callee selector matches any of
// the given method names.
func callHasAnyName(call *domain.CloneNode, names []string) bool {
	for _, child := range call.Children {
		if child.BaseType == golang.SelectorExpr && slices.Contains(names, child.Name) {
			return true
		}
	}

	return false
}

// maxTestMainStatements bounds the statement count of a TestMain body clone.
const maxTestMainStatements = 5

// isTestMainBoilerplate reports whether every clone is the body of a
// TestMain(m *testing.M) function: an assignment from m.Run(), optional
// simple middle statements (flag.Parse, snaps.Clean, ...), ending in
// os.Exit(code). Go requires TestMain to be declared in-package, once per
// package — the shape is inextractable by language rules.
//
// Source: go-cqrs-lite feedback (Finding 3) — a 15-module snaps.Clean
// TestMain clone group that users had to accept via 15 hand-written
// //art-dupl:accept directives.
//
// Guards against over-suppression:
//   - all clones must originate from test context files
//   - the first statement must assign from an m.Run() call (bare receiver
//     Ident "m" — the conventional testing.M parameter name)
//   - the last statement must be an os.Exit(...) call
//   - middle statements may only be assignments or bare non-assertion calls
func isTestMainBoilerplate(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if !allFromTestContext(seq) || len(seq) < 3 || len(seq) > maxTestMainStatements {
			return false
		}

		if !isAssignFromMRun(seq[0]) {
			return false
		}

		if !isOSExitCall(seq[len(seq)-1]) {
			return false
		}

		for _, stmt := range seq[1 : len(seq)-1] {
			switch stmt.BaseType {
			case golang.AssignStmt:
				continue
			case golang.ExprStmt:
				call := unwrapExprStmt(stmt)

				if call.BaseType != golang.CallExpr || callHasAnyName(call, preambleDisallowedCalls) {
					return false
				}
			default:
				return false
			}
		}

		return true
	})
}

// isAssignFromMRun checks that a statement is `code := m.Run()` — an
// AssignStmt whose RHS call targets a SelectorExpr named "Run" with a bare
// receiver Ident "m".
func isAssignFromMRun(stmt *domain.CloneNode) bool {
	if stmt.BaseType != golang.AssignStmt {
		return false
	}

	for _, child := range stmt.Children {
		if child.BaseType != golang.CallExpr {
			continue
		}

		for _, cc := range child.Children {
			if cc.BaseType != golang.SelectorExpr || cc.Name != "Run" {
				continue
			}

			for _, sc := range cc.Children {
				if sc.BaseType == golang.Ident && sc.Name == "m" {
					return true
				}
			}
		}
	}

	return false
}

// isOSExitCall checks that a statement is `os.Exit(code)` — an ExprStmt
// wrapping a CallExpr whose callee SelectorExpr is named "Exit" with
// receiver Ident "os".
func isOSExitCall(stmt *domain.CloneNode) bool {
	unwrapped := unwrapExprStmt(stmt)
	if unwrapped.BaseType != golang.CallExpr {
		return false
	}

	for _, child := range unwrapped.Children {
		if child.BaseType != golang.SelectorExpr || child.Name != "Exit" {
			continue
		}

		for _, sc := range child.Children {
			if sc.BaseType == golang.Ident && sc.Name == "os" {
				return true
			}
		}
	}

	return false
}

// maxEmbedDirectiveStatements bounds the statement count of an embed bootstrap
// clone. Longer sequences contain real logic and stay actionable.
const maxEmbedDirectiveStatements = 4

// isEmbedDirectivePattern reports whether every clone is anchored on a
// `//go:embed`-populated `var ... embed.FS` declaration, optionally followed
// by the fs.Sub assignment and its error guard. The //go:embed compiler
// directive may only appear immediately before a var declaration in the same
// package — there is no mechanism to share an embedded embed.FS across
// packages, so this duplication is impossible to extract by language rules.
//
// Source: go-sse feedback — `var staticFiles embed.FS` + fs.Sub stanza
// reported as the only clone group at -t 1 across two example programs.
//
// Guards against over-suppression:
//   - the first statement must be a ValueSpec declaring a var whose type is
//     embed.FS (SelectorExpr "FS" on receiver "embed", or a bare "FS" ident
//     for dot-imported embed)
//   - at most maxEmbedDirectiveStatements statements total
//   - following statements may only be assignments (fs.Sub open) or
//     if-statements (the error guard)
func isEmbedDirectivePattern(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) < 1 || len(seq) > maxEmbedDirectiveStatements {
			return false
		}

		if !declaresEmbedFS(seq[0]) {
			return false
		}

		for _, stmt := range seq[1:] {
			switch stmt.BaseType {
			case golang.AssignStmt, golang.IfStmt:
				continue
			default:
				return false
			}
		}

		return true
	})
}

// declaresEmbedFS checks that a statement declares a variable of type
// embed.FS. Package-level `var x embed.FS` arrives as a ValueSpec token;
// in-function declarations arrive wrapped in DeclStmt/GenDecl.
func declaresEmbedFS(stmt *domain.CloneNode) bool {
	root := stmt

	if root.BaseType == golang.DeclStmt && len(root.Children) == 1 {
		root = root.Children[0]
	}

	if root.BaseType != golang.ValueSpec && root.BaseType != golang.GenDecl {
		return false
	}

	return subtreeDeclaresEmbedFS(root)
}

// subtreeDeclaresEmbedFS recursively searches a declaration subtree for an
// embed.FS type reference: a SelectorExpr "FS" on receiver Ident "embed"
// (the canonical `var x embed.FS` shape). A bare Ident "FS" is deliberately
// NOT accepted — it cannot be distinguished from a variable named FS without
// position info, and dot-importing embed is vanishingly rare.
func subtreeDeclaresEmbedFS(node *domain.CloneNode) bool {
	if node.BaseType == golang.SelectorExpr && node.Name == "FS" {
		for _, child := range node.Children {
			if child.BaseType == golang.Ident && child.Name == "embed" {
				return true
			}
		}
	}

	return slices.ContainsFunc(node.Children, subtreeDeclaresEmbedFS)
}
