package syntax

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
)

// Tests for nested-statement token emission (ADR-0023). Before the fix,
// serial() emitted exactly one composite token per statement, so two composite
// statements sharing leading statements but diverging deeper inside (the
// "shared loop skeleton, divergent accumulator" shape) produced zero matching
// tokens and were invisible at every threshold and mode.

// nodeTree helpers build unified trees shaped like the Go transformer's output:
// statements carry Statement=true, blocks are plain container nodes.

func stmtNode(typ, pos, end int32) *Node {
	return &Node{Type: typ, Pos: pos, End: end, Statement: true}
}

func plainNode(typ, pos, end int32, children ...*Node) *Node {
	return &Node{Type: typ, Pos: pos, End: end, Children: children}
}

func blockNode(pos, end int32, stmts ...*Node) *Node {
	return plainNode(6, pos, end, stmts...) // golang.BlockStmt
}

// streamTypes returns the stream positions' statement markers and base types
// for assertions.
func streamStatements(stream []*Node) []*Node {
	var out []*Node

	for _, n := range stream {
		if n.Statement {
			out = append(out, n)
		}
	}

	return out
}

// TestNestedLoopBodyStatementsEmitted verifies that the statements inside a
// loop body are emitted as individual statement tokens after the loop's
// composite token, instead of being reachable only through the composite.
func TestNestedLoopBodyStatementsEmitted(t *testing.T) {
	t.Parallel()

	s1 := stmtNode(100, 10, 20)
	s2 := stmtNode(101, 21, 30)
	loop := &Node{
		Type:      20, // golang.ForStmt
		Pos:       5,
		End:       40,
		Statement: true,
		Children:  []*Node{blockNode(6, 35, s1, s2)},
	}
	fnBody := blockNode(4, 50, loop)
	fn := plainNode(21, 0, 60, fnBody) // golang.FuncDecl

	stream := Serialize(fn)
	stmts := streamStatements(stream)

	// Expected statement tokens: loop composite, s1, s2.
	if len(stmts) != 3 {
		t.Fatalf("expected 3 statement tokens (loop composite + 2 body statements), got %d", len(stmts))
	}

	if stmts[0].Fingerprint == 0 {
		t.Error("loop composite token must carry a fingerprint")
	}

	// serial() emits shallow arena copies, so compare by source range.
	if stmts[1].Pos != s1.Pos || stmts[1].End != s1.End || stmts[2].Pos != s2.Pos || stmts[2].End != s2.End {
		t.Error("body statements must be emitted in source order after the composite")
	}
}

// TestDeeplyNestedBlocksEmitRecursively verifies for→if→statement chains emit
// at every nesting level.
func TestDeeplyNestedBlocksEmitRecursively(t *testing.T) {
	t.Parallel()

	inner := stmtNode(102, 30, 40)
	ifStmt := &Node{
		Type:      27, // golang.IfStmt
		Pos:       20,
		End:       50,
		Statement: true,
		Children:  []*Node{blockNode(6, 45, inner)},
	}
	loop := &Node{
		Type:      20, // golang.ForStmt
		Pos:       10,
		End:       60,
		Statement: true,
		Children:  []*Node{blockNode(6, 55, ifStmt)},
	}
	fn := plainNode(21, 0, 70, blockNode(4, 65, loop))

	stmts := streamStatements(Serialize(fn))

	// Expected: loop composite, if composite, inner statement.
	if len(stmts) != 3 {
		t.Fatalf("expected 3 statement tokens (loop, if, inner), got %d", len(stmts))
	}
}

// TestElseIfChainLinkEmitted verifies that an `else if` link (flagged as a
// statement atom by the Go transformer) is emitted with its own composite plus
// its nested block statements.
func TestElseIfChainLinkEmitted(t *testing.T) {
	t.Parallel()

	thenStmt := stmtNode(103, 12, 20)
	elseIfBody := stmtNode(104, 32, 40)
	elseIf := &Node{
		Type:      27, // golang.IfStmt
		Pos:       25,
		End:       50,
		Statement: true,
		Children:  []*Node{blockNode(6, 45, elseIfBody)},
	}
	outer := &Node{
		Type:      27,
		Pos:       5,
		End:       60,
		Statement: true,
		Children: []*Node{
			blockNode(6, 22, thenStmt), // then block
			elseIf,                     // else-if chain link
		},
	}
	fn := plainNode(21, 0, 70, blockNode(4, 65, outer))

	stmts := streamStatements(Serialize(fn))

	// Expected: outer-if composite, then-block statement, else-if composite,
	// else-if body statement.
	if len(stmts) != 4 {
		t.Fatalf("expected 4 statement tokens, got %d", len(stmts))
	}

	if stmts[2].Pos != elseIf.Pos || stmts[2].End != elseIf.End {
		t.Error("else-if link must be emitted as its own statement token")
	}

	if stmts[3].Pos != elseIfBody.Pos || stmts[3].End != elseIfBody.End {
		t.Error("else-if body statement must be emitted")
	}
}

// TestGenDeclSpecsStayCompositeOnly verifies that GenDecl children (flagged
// ValueSpec/TypeSpec atoms) are NOT emitted as separate tokens: a single-spec
// `var x T = v` would otherwise contribute two tokens for one source
// statement.
func TestGenDeclSpecsStayCompositeOnly(t *testing.T) {
	t.Parallel()

	spec := stmtNode(48, 10, 20) // golang.ValueSpec
	genDecl := &Node{
		Type:      24, // golang.GenDecl (semantic encoding in real trees; base here suffices)
		Pos:       5,
		End:       30,
		Statement: true,
		Children:  []*Node{spec},
	}
	fn := plainNode(21, 0, 40, blockNode(4, 35, genDecl))

	stmts := streamStatements(Serialize(fn))

	if len(stmts) != 1 {
		t.Fatalf("expected exactly 1 statement token (the GenDecl composite), got %d", len(stmts))
	}
}

// TestSwitchCaseBodiesEmitted verifies switch case bodies (CaseClause statement
// atoms wrapping body statements) emit through the BlockStmt container descent.
func TestSwitchCaseBodiesEmitted(t *testing.T) {
	t.Parallel()

	case1Stmt := stmtNode(110, 20, 30)
	case1 := &Node{
		Type:      9, // golang.CaseClause
		Pos:       15,
		End:       35,
		Statement: true,
		Children:  []*Node{case1Stmt},
	}
	sw := &Node{
		Type:      43, // golang.SwitchStmt
		Pos:       5,
		End:       45,
		Statement: true,
		Children:  []*Node{blockNode(6, 40, case1)},
	}
	fn := plainNode(21, 0, 55, blockNode(4, 50, sw))

	stmts := streamStatements(Serialize(fn))

	// Expected: switch composite, case-clause composite, case body statement.
	if len(stmts) != 3 {
		t.Fatalf("expected 3 statement tokens, got %d", len(stmts))
	}
}

// TestNestedEmissionArenaCountMatchesStream guards the countSerializedNodes /
// serial mirror invariant for the nested-descent traversal: an undercount
// panics (arena index out of range), an overcount silently wastes memory.
func TestNestedEmissionArenaCountMatchesStream(t *testing.T) {
	t.Parallel()

	build := func() *Node {
		inner := stmtNode(102, 130, 140)
		elseIfBody := stmtNode(104, 120, 125)
		elseIf := &Node{
			Type: 27, Pos: 100, End: 130, Statement: true,
			Children: []*Node{blockNode(6, 128, elseIfBody)},
		}
		ifStmt := &Node{
			Type: 27, Pos: 90, End: 140, Statement: true,
			Children: []*Node{blockNode(6, 110, inner), elseIf},
		}
		case1Stmt := stmtNode(110, 160, 170)
		case1 := &Node{
			Type: 9, Pos: 150, End: 175, Statement: true,
			Children: []*Node{case1Stmt},
		}
		sw := &Node{
			Type: 43, Pos: 80, End: 180, Statement: true,
			Children: []*Node{blockNode(6, 178, case1)},
		}
		spec := stmtNode(48, 200, 210)
		decl := &Node{
			Type: 24, Pos: 190, End: 215, Statement: true,
			Children: []*Node{spec},
		}
		loop := &Node{
			Type: 20, Pos: 70, End: 220, Statement: true,
			Children: []*Node{blockNode(6, 85, ifStmt, sw, decl)},
		}

		return plainNode(21, 0, 240, blockNode(4, 60, loop))
	}

	for _, maxChildren := range []int{0, 1, 2, 10} {
		tree := build()
		stream := SerializeWithMaxChildren(tree, maxChildren)

		if got := len(stream); got == 0 {
			t.Fatalf("maxChildren=%d: empty stream", maxChildren)
		}

		// Serialize is deterministic: a second pass must produce the same
		// length (also re-verifies non-mutation, which would skew counts).
		again := SerializeWithMaxChildren(tree, maxChildren)
		if len(again) != len(stream) {
			t.Errorf("maxChildren=%d: stream length changed between runs: %d vs %d", maxChildren, len(stream), len(again))
		}
	}
}

// serializeFileWithSentinel appends a file's stream plus a unique sentinel
// statement token (mirrors job.BuildTree's file separation).
func serializeFileWithSentinel(t *testing.T, data *[]*Node, tree *suffixtree.STree, root *Node, sentinel int32) {
	t.Helper()

	for _, n := range Serialize(root) {
		*data = append(*data, n)

		if err := tree.Update(n); err != nil {
			t.Fatalf("suffix tree update: %v", err)
		}
	}

	sentinelNode := &Node{Type: -1, Statement: true, Fingerprint: sentinel}
	*data = append(*data, sentinelNode)

	if err := tree.Update(sentinelNode); err != nil {
		t.Fatalf("suffix tree update sentinel: %v", err)
	}
}

// TestLoopSkeletonWithDivergentTailIsDetected is THE regression for the
// false-negative class: two loops sharing the first two body statements but
// diverging in the third MUST produce a clone group (previously produced
// nothing at any threshold). Fixture proven on the real pipeline before the
// fix (docs/status/2026-09-13_15-45 report, experiments A/B).
func TestLoopSkeletonWithDivergentTailIsDetected(t *testing.T) {
	t.Parallel()

	const (
		tInit = 100 // shared initializer statement
		tS1   = 101 // shared loop-body statement 1
		tS2   = 102 // shared loop-body statement 2
		tS3a  = 103 // divergent tail in file A
		tS3b  = 104 // divergent tail in file B
	)

	buildFile := func(loopType, tailType, base int32) *Node {
		s1 := stmtNode(tS1, base+10, base+20)
		s2 := stmtNode(tS2, base+21, base+30)
		tail := stmtNode(tailType, base+31, base+40)
		loop := &Node{
			Type: loopType, Pos: base, End: base + 50, Statement: true,
			Children: []*Node{blockNode(6, base+45, s1, s2, tail)},
		}
		init := stmtNode(tInit, base-10, base-1)

		return plainNode(21, base-20, base+60, blockNode(4, base+55, init, loop))
	}

	var data []*Node
	tree := suffixtree.New()

	serializeFileWithSentinel(t, &data, tree, buildFile(20, tS3a, 100), -100)
	serializeFileWithSentinel(t, &data, tree, buildFile(20, tS3b, 200), -101)

	const threshold = 2

	found := false

	for m := range tree.FindDuplOver(t.Context(), threshold) {
		match := FindSyntaxUnits(data, m, threshold)
		if len(match.Frags) < 2 {
			continue
		}

		for _, frag := range match.Frags {
			stmts := streamStatements(frag)
			if len(stmts) < threshold {
				t.Fatalf("match fragment below threshold after trim: %d statements", len(stmts))
			}
		}

		found = true
	}

	if !found {
		t.Fatal("loop skeletons sharing two leading body statements must be detected at threshold 2")
	}
}

// TestIdenticalGuardCloneUnitsAreTrimmed verifies the subsumed-unit trim: an
// identical guard `if cond { return }` yields a maximal match of
// [if-composite, return-in-body]; the return's range lies inside the if's, so
// the fragment must collapse to the composite alone (pre-fix group shape).
func TestIdenticalGuardCloneUnitsAreTrimmed(t *testing.T) {
	t.Parallel()

	buildFile := func(base int32) *Node {
		ret := stmtNode(37, base+20, base+30) // golang.ReturnStmt
		ifStmt := &Node{
			Type: 27, Pos: base, End: base + 40, Statement: true,
			Children: []*Node{blockNode(6, base+35, ret)},
		}

		return plainNode(21, base-20, base+50, blockNode(4, base+45, ifStmt))
	}

	var data []*Node
	tree := suffixtree.New()

	serializeFileWithSentinel(t, &data, tree, buildFile(100), -100)
	serializeFileWithSentinel(t, &data, tree, buildFile(200), -101)

	const threshold = 1

	units := -1

	for m := range tree.FindDuplOver(t.Context(), threshold) {
		match := FindSyntaxUnits(data, m, threshold)
		if len(match.Frags) < 2 {
			continue
		}

		units = len(match.Frags[0])

		break
	}

	if units == -1 {
		t.Fatal("identical guard clones must be detected")
	}

	if units != 1 {
		t.Errorf("guard clone fragment must trim to 1 unit (the if composite), got %d", units)
	}
}
