package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// Tests for DiscordSync false-positive patterns. Each test constructs a
// CloneNode tree matching the real DiscordSync AST shapes and verifies the
// actionability evaluator classifies them as NonActionable.

func TestDiscordSync_DeferCancelBareIdent(t *testing.T) {
	// defer cancel() — bare Ident callee, not SelectorExpr
	seqs := [][]*domain.CloneNode{
		{mustDeferBareIdentCall("cancel")},
		{mustDeferBareIdentCall("cancel")},
	}
	assertNonActionable(t, "defer cancel() should be non-actionable (raii-defer)", seqs)
}

func TestDiscordSync_DeferFuncLitWrapping(t *testing.T) {
	// defer func() { _ = rows.Close() }()
	seqs := [][]*domain.CloneNode{
		{mustDeferFuncLitCleanup("rows", "Close")},
		{mustDeferFuncLitCleanup("rows", "Close")},
	}
	assertNonActionable(t, "defer func() { _ = rows.Close() }() should be non-actionable (raii-defer)", seqs)
}

func TestDiscordSync_DeferFuncLitTxRollback(t *testing.T) {
	// defer func() { _ = transaction.Rollback() }()
	seqs := [][]*domain.CloneNode{
		{mustDeferFuncLitCleanup("transaction", "Rollback")},
		{mustDeferFuncLitCleanup("transaction", "Rollback")},
	}
	assertNonActionable(t, "defer func() { _ = tx.Rollback() }() should be non-actionable (raii-defer)", seqs)
}

func TestDiscordSync_HTTPErrorGuard(t *testing.T) {
	// if err != nil { writeError(w, r, err, ""); return }
	// The writeError call is NOT a logging method — it's a custom handler.
	seqs := [][]*domain.CloneNode{
		{mustHTTPErrorGuard()},
		{mustHTTPErrorGuard()},
	}
	assertNonActionable(t, "HTTP error guard (writeError + bare return) should be non-actionable (error-propagation)", seqs)
}

func TestDiscordSync_HTTPErrorGuardWithSlog(t *testing.T) {
	// if err != nil { slog.Error("msg", err); return }
	// This should already match (slog.Error IS a logging method), but test for regression safety.
	seqs := [][]*domain.CloneNode{
		{mustIfErrSlogAndReturn()},
		{mustIfErrSlogAndReturn()},
	}
	assertNonActionable(t, "if err != nil { slog.Error(...); return } should be non-actionable (error-propagation)", seqs)
}

func TestDiscordSync_QueryErrorWrapping(t *testing.T) {
	// if err != nil { return nil, queryError(err, "unique msg") }
	// queryError is a project-specific wrapper, not in the name allowlist.
	seqs := [][]*domain.CloneNode{
		{mustIfErrReturnWrappingCall("queryError")},
		{mustIfErrReturnWrappingCall("queryError")},
	}
	assertNonActionable(t, "return nil, queryError(err, ...) should be non-actionable (error-wrapping)", seqs)
}

func TestDiscordSync_ErrorfWrappingStillMatches(t *testing.T) {
	// Regression: fmt.Errorf should still match after structural rewrite
	seqs := [][]*domain.CloneNode{
		{mustIfErrReturnWrappingCall("Errorf")},
		{mustIfErrReturnWrappingCall("Errorf")},
	}
	assertNonActionable(t, "return fmt.Errorf(..., err) should be non-actionable (error-wrapping)", seqs)
}

func TestDiscordSync_BoolOkGuard(t *testing.T) {
	// X, ok := helper(); if !ok { return }
	seqs := [][]*domain.CloneNode{
		mustAssignWithBoolGuard(),
		mustAssignWithBoolGuard(),
	}
	assertNonActionable(t, "X, ok := helper(); if !ok { return } should be non-actionable (bool-guard)", seqs)
}

// --- Helpers for DiscordSync patterns ---

func mustDeferBareIdentCall(name string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.DeferStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: name},
				},
			},
		},
	}
}

func mustDeferFuncLitCleanup(receiver, method string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.DeferStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.FuncLit,
						Children: []*domain.CloneNode{
							{BaseType: golang.FuncType},
							{
								BaseType: golang.BlockStmt,
								Children: []*domain.CloneNode{
									{
										BaseType: golang.AssignStmt,
										Children: []*domain.CloneNode{
											{BaseType: golang.Ident, Name: "_"},
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

func mustHTTPErrorGuard() *domain.CloneNode {
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
						BaseType: golang.ExprStmt,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{
										BaseType: golang.SelectorExpr,
										Name:     "WriteError",
										Children: []*domain.CloneNode{
											{BaseType: golang.Ident, Name: "errorpage"},
											{BaseType: golang.Ident, Name: "WriteError"},
										},
									},
									{BaseType: golang.Ident, Name: "w"},
									{BaseType: golang.Ident, Name: "r"},
									{BaseType: golang.Ident, Name: "err"},
								},
							},
						},
					},
					{BaseType: golang.ReturnStmt},
				},
			},
		},
	}
}

func mustIfErrSlogAndReturn() *domain.CloneNode {
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
						BaseType: golang.ExprStmt,
						Children: []*domain.CloneNode{
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{
										BaseType: golang.SelectorExpr,
										Name:     "Error",
										Children: []*domain.CloneNode{
											{BaseType: golang.Ident, Name: "slog"},
											{BaseType: golang.Ident, Name: "Error"},
										},
									},
									{BaseType: golang.BasicLit},
									{BaseType: golang.Ident, Name: "err"},
								},
							},
						},
					},
					{BaseType: golang.ReturnStmt},
				},
			},
		},
	}
}

func mustIfErrReturnWrappingCall(calleeName string) *domain.CloneNode {
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
							{BaseType: golang.Ident, Name: "nil"},
							{
								BaseType: golang.CallExpr,
								Children: []*domain.CloneNode{
									{BaseType: golang.Ident, Name: calleeName},
									{BaseType: golang.Ident, Name: "err"},
									{BaseType: golang.BasicLit},
								},
							},
						},
					},
				},
			},
		},
	}
}

func mustAssignWithBoolGuard() []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType: golang.AssignStmt,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "guildID"},
				{BaseType: golang.Ident, Name: "ok"},
				{
					BaseType: golang.CallExpr,
					Children: []*domain.CloneNode{
						{
							BaseType: golang.SelectorExpr,
							Name:     "requireQueryParam",
							Children: []*domain.CloneNode{
								{BaseType: golang.Ident, Name: "s"},
								{BaseType: golang.Ident, Name: "requireQueryParam"},
							},
						},
						{BaseType: golang.Ident, Name: "w"},
						{BaseType: golang.Ident, Name: "r"},
						{BaseType: golang.BasicLit},
					},
				},
			},
		},
		{
			BaseType: golang.IfStmt,
			Children: []*domain.CloneNode{
				{
					BaseType: golang.UnaryExpr,
					Children: []*domain.CloneNode{
						{BaseType: golang.Ident, Name: "ok"},
					},
				},
				{
					BaseType: golang.BlockStmt,
					Children: []*domain.CloneNode{
						{BaseType: golang.ReturnStmt},
					},
				},
			},
		},
	}
}

func assertNonActionable(t *testing.T, msg string, seqs [][]*domain.CloneNode) {
	t.Helper()

	label, actionability := EvaluateActionabilityWithLabel(seqs)
	if actionability != domain.NonActionable {
		t.Errorf("%s\nexpected NonActionable, got %s (label=%s)", msg, actionability, label)
	}
}
