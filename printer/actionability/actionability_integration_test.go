package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestEvaluateActionabilityWithLabel_ErrorWrapping verifies the full
// EvaluateActionabilityWithLabel pipeline detects the error wrapping idiom:
// if err != nil { return fmt.Errorf("...") }.
//
// Note: isPureErrorPropagation is checked before isErrorWrappingReturn in the
// pipeline, so this tree gets PatternErrorPropagation. Both classify it as
// NonActionable, which is the important outcome.
func TestEvaluateActionabilityWithLabel_ErrorWrapping(t *testing.T) {
	t.Parallel()

	seqs := [][]*domain.CloneNode{
		{mustIfErrWrapReturn()},
		{mustIfErrWrapReturn()},
	}

	label, action := EvaluateActionabilityWithLabel(seqs)
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternErrorPropagation {
		t.Errorf("label = %q, want %q", label, PatternErrorPropagation)
	}
}

// TestEvaluateActionabilityWithLabel_CobraCommand verifies the full pipeline
// detects cobra.Command struct literals with the tightened isCommandLiteral
// check that validates the receiver is "cobra" or "fang".
func TestEvaluateActionabilityWithLabel_CobraCommand(t *testing.T) {
	t.Parallel()

	seqs := [][]*domain.CloneNode{
		{mustCobraCommandLit()},
		{mustCobraCommandLit()},
	}

	label, action := EvaluateActionabilityWithLabel(seqs)
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternCobraBoilerplate {
		t.Errorf("label = %q, want %q", label, PatternCobraBoilerplate)
	}
}

// TestEvaluateActionabilityWithLabel_CobraCommand_WrongReceiver verifies
// that a CompositeLit with a SelectorExpr "Command" but wrong receiver
// (e.g., "myapp") is NOT classified as cobra boilerplate.
func TestEvaluateActionabilityWithLabel_CobraCommand_WrongReceiver(t *testing.T) {
	t.Parallel()

	lit := &domain.CloneNode{
		BaseType: golang.CompositeLit,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.SelectorExpr,
				Name:     "Command",
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "myapp"},
				},
			},
		},
	}

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		{lit},
		{lit},
	})
	if action != domain.Actionable {
		t.Errorf("action = %q, want %q (wrong receiver should be actionable)", action, domain.Actionable)
	}

	if label != PatternNone {
		t.Errorf("label = %q, want %q", label, PatternNone)
	}
}

// TestEvaluateActionabilityWithLabel_BuilderCallback verifies the full
// pipeline detects builder/callback patterns: a chain of 3+ method calls
// on different receiver types.
func TestEvaluateActionabilityWithLabel_BuilderCallback(t *testing.T) {
	t.Parallel()

	clone := mustBuilderChain()

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		clone,
		clone,
	})
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternBuilderCallback {
		t.Errorf("label = %q, want %q", label, PatternBuilderCallback)
	}
}

// TestEvaluateActionabilityWithLabel_BuilderCallback_BelowThreshold verifies
// that only 2 calls with 1 receiver does NOT trigger the pattern.
func TestEvaluateActionabilityWithLabel_BuilderCallback_BelowThreshold(t *testing.T) {
	t.Parallel()

	clone := []*domain.CloneNode{
		mustCallExprWithSelector("Configure"),
		mustCallExprWithSelector("Configure"),
	}

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		clone,
		clone,
	})
	if action != domain.Actionable {
		t.Errorf("action = %q, want %q (below threshold should be actionable)", action, domain.Actionable)
	}

	if label != PatternNone {
		t.Errorf("label = %q, want %q", label, PatternNone)
	}
}

// TestEvaluateActionabilityWithLabel_TableDrivenTest_NonTestingReceiver
// verifies that a RangeStmt with t.Run but a NON-testing receiver variable
// (e.g., "x.Run") does NOT trigger the table-driven-test pattern.
// This tests the hasTestingReceiver guard against false positives.
func TestEvaluateActionabilityWithLabel_TableDrivenTest_NonTestingReceiver(t *testing.T) {
	t.Parallel()

	clone := mustRangeStmtWithNonTestingTRun("resolve_test.go")

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		{clone},
		{clone},
	})
	if action != domain.Actionable {
		t.Errorf("action = %q, want %q (non-testing receiver should be actionable)", action, domain.Actionable)
	}

	if label != PatternNone {
		t.Errorf("label = %q, want %q", label, PatternNone)
	}
}

// TestEvaluateActionabilityWithLabel_SingleCallExprStatement verifies that
// a lone function call used as a statement (ExprStmt wrapping CallExpr) is
// classified as non-actionable. This is the real-world shape produced by
// statement-level tokenization: t.Parallel(), fmt.Println("x"), etc. appear
// as ExprStmt(CallExpr), not bare CallExpr.
func TestEvaluateActionabilityWithLabel_SingleCallExprStatement(t *testing.T) {
	t.Parallel()

	clone := mustExprStmtCallExpr("t", "Parallel")

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		{clone},
		{clone},
	})
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternSingleCallExpr {
		t.Errorf("label = %q, want %q", label, PatternSingleCallExpr)
	}
}

// TestEvaluateActionabilityWithLabel_TestHelperDelegate verifies that a
// 2-statement t.Helper() + delegate body is classified as non-actionable.
// This is the irreducible Go test-helper idiom:
//
//	func AssertNotNil(t *testing.T, got any, what string) {
//	    t.Helper()
//	    failIfNilf(t, got, "expected non-nil")
//	}
func TestEvaluateActionabilityWithLabel_TestHelperDelegate(t *testing.T) {
	t.Parallel()

	helperCall := mustExprStmtCallExpr("t", "Helper")
	delegateCall := mustExprStmtBareCall("failIfNilf")

	seq := []*domain.CloneNode{helperCall, delegateCall}

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		seq,
		seq,
	})
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternTestHelperDelegate {
		t.Errorf("label = %q, want %q", label, PatternTestHelperDelegate)
	}
}

// TestEvaluateActionabilityWithLabel_TestHelperDelegate_SelectorDelegate
// verifies that t.Helper() + a method-call delegate (e.g. t.Errorf(...))
// is also classified as non-actionable test-helper-delegate.
func TestEvaluateActionabilityWithLabel_TestHelperDelegate_SelectorDelegate(t *testing.T) {
	t.Parallel()

	helperCall := mustExprStmtCallExpr("t", "Helper")
	delegateCall := mustExprStmtCallExpr("assert", "Equal") // selector-based delegate

	seq := []*domain.CloneNode{helperCall, delegateCall}

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		seq,
		seq,
	})
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternTestHelperDelegate {
		t.Errorf("label = %q, want %q", label, PatternTestHelperDelegate)
	}
}

// TestEvaluateActionabilityWithLabel_TestHelperDelegate_ThreeStmts verifies
// that a 3-statement body does NOT trigger the pattern — longer bodies may
// contain real actionable duplication beyond the t.Helper() boilerplate.
func TestEvaluateActionabilityWithLabel_TestHelperDelegate_ThreeStmts(t *testing.T) {
	t.Parallel()

	helperCall := mustExprStmtCallExpr("t", "Helper")
	delegateCall := mustExprStmtBareCall("failIfNilf")
	extraCall := mustExprStmtBareCall("doMore")

	seq := []*domain.CloneNode{helperCall, delegateCall, extraCall}

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		seq,
		seq,
	})
	if action != domain.Actionable {
		t.Errorf("action = %q, want %q (3 statements should be actionable)", action, domain.Actionable)
	}

	if label == PatternTestHelperDelegate {
		t.Errorf("label = %q, should NOT be test-helper-delegate", label)
	}
}

// TestEvaluateActionabilityWithLabel_TestHelperDelegate_NonHelperFirstStmt
// verifies that a 2-statement body where the first call is NOT .Helper()
// does NOT trigger the pattern.
func TestEvaluateActionabilityWithLabel_TestHelperDelegate_NonHelperFirstStmt(t *testing.T) {
	t.Parallel()

	firstCall := mustExprStmtCallExpr("t", "Parallel") // not .Helper()
	delegateCall := mustExprStmtBareCall("doSomething")

	seq := []*domain.CloneNode{firstCall, delegateCall}

	label, action := EvaluateActionabilityWithLabel([][]*domain.CloneNode{
		seq,
		seq,
	})
	if action != domain.Actionable {
		t.Errorf("action = %q, want %q (non-Helper first stmt should be actionable)", action, domain.Actionable)
	}

	if label == PatternTestHelperDelegate {
		t.Errorf("label = %q, should NOT be test-helper-delegate", label)
	}
}

// --- Helper builders for integration tests ---

// mustExprStmtBareCall constructs an ExprStmt wrapping a CallExpr whose
// function is a bare Ident (not a SelectorExpr). This represents a
// same-package function call used as a statement:
//
//	failIfNilf(t, got, "expected non-nil")
//	assertEqualMsgf(t, got, want, format)
func mustExprStmtBareCall(funcName string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.ExprStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: funcName},
				},
			},
		},
	}
}

// mustExprStmtCallExpr constructs an ExprStmt wrapping a CallExpr, representing
//
//	if err != nil {
//	    return fmt.Errorf("...")
//	}
func mustIfErrWrapReturn() *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.IfStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BinaryExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "err"},
					{BaseType: golang.Ident, Name: "nil"},
				},
			},
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ReturnStmt,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{
										BaseType: golang.SelectorExpr,
										Name:     "Errorf",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// mustCobraCommandLit constructs a CompositeLit representing:
//
//	cobra.Command{ Run: func() {...} }
func mustCobraCommandLit() *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.CompositeLit,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.SelectorExpr,
				Name:     "Command",
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "cobra"},
				},
			},
		},
	}
}

// mustBuilderChain constructs a sequence of 3 CallExpr nodes with different
// SelectorExpr names, simulating a builder pattern:
//
//	svc.WithName("x").WithPort(8080).Build()
func mustBuilderChain() []*domain.CloneNode {
	return []*domain.CloneNode{
		mustCallExprWithSelector("WithName"),
		mustCallExprWithSelector("WithPort"),
		mustCallExprWithSelector("Build"),
	}
}

// mustCallExprWithSelector builds a CallExpr wrapping a SelectorExpr with the
// given method name. Used to simulate chained method calls like builder patterns.
func mustCallExprWithSelector(methodName string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.CallExpr,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.SelectorExpr,
				Name:     methodName,
			},
		},
	}
}

// mustRangeStmtWithNonTestingTRun constructs a RangeStmt with a CallExpr
// to ".Run" where the receiver Ident is "x" (NOT a testing variable name).
// This should NOT trigger the table-driven-test pattern.
func mustRangeStmtWithNonTestingTRun(filename string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.RangeStmt,
		Filename: filename,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ExprStmt,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{
										BaseType: golang.SelectorExpr,
										Name:     "Run",
										Children: []*domain.CloneNode{
											{BaseType: golang.Ident, Name: "x"},
											{BaseType: golang.Ident, Name: "Run"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// mustExprStmtCallExpr constructs an ExprStmt wrapping a CallExpr, representing
// a standalone function call used as a statement. This is the real-world shape
// produced by statement-level tokenization for code like:
//
//	t.Parallel()
//	fmt.Println("hello")
//	log.Fatal(err)
func mustExprStmtCallExpr(receiver, method string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.ExprStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.SelectorExpr,
						Name:     method,
						Children: []*domain.CloneNode{
							{BaseType: golang.Ident, Name: receiver},
							{BaseType: golang.Ident, Name: method},
						},
					},
				},
			},
		},
	}
}
