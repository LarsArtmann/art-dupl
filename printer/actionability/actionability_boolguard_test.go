package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestFindOkVarName(t *testing.T) {
	mkAssign := func(name string) *domain.CloneNode {
		return &domain.CloneNode{
			BaseType: golang.AssignStmt,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "val"},
				{BaseType: golang.Ident, Name: name},
			},
		}
	}

	for _, tc := range []struct {
		name string
		want string
	}{
		{"ok", "ok"},
		{"found", "found"},
		{"exists", "exists"},
		{"success", "success"},
		{"present", "present"},
		{"count", ""},
		{"err", ""},
		{"result", ""},
		{"", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := findOkVarName(mkAssign(tc.name))
			if got != tc.want {
				t.Fatalf("findOkVarName(name=%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestIsAssignWithBoolGuard_AllNames(t *testing.T) {
	names := []string{"ok", "found", "exists", "success", "present"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			seqs := [][]*domain.CloneNode{
				buildBoolGuardSeq(name),
				buildBoolGuardSeq(name),
			}
			if !isAssignWithBoolGuard(seqs) {
				t.Fatalf("isAssignWithBoolGuard should match guard with var %q", name)
			}
		})
	}
}

func TestIsAssignWithBoolGuard_RejectsNonBoolName(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		buildBoolGuardSeq("count"),
		buildBoolGuardSeq("count"),
	}
	if isAssignWithBoolGuard(seqs) {
		t.Fatal("isAssignWithBoolGuard should NOT match guard with var 'count'")
	}
}

func buildBoolGuardSeq(okName string) []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType: golang.AssignStmt,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "val"},
				{BaseType: golang.Ident, Name: okName},
				{
					BaseType: golang.CallExpr,
					Children: []*domain.CloneNode{
						{BaseType: golang.Ident, Name: "lookup"},
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
						{BaseType: golang.Ident, Name: okName},
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
