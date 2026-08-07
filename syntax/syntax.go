// Package syntax provides unified AST representation for code duplication detection.
//
// This package bridges the gap between language-specific AST parsers
// (golang/ast for Go code) and the language-agnostic suffix tree
// used by the detection algorithm.
//
// Core Types:
// - Node: Unified syntax tree node representing any language construct
// - Match: Represents a clone match with fragments (group of nodes)
// - Frags: Slice of node sequences (each fragment is a sequence of nodes)
//
// Design:
// - Language-agnostic: Works with any language that provides a parser
// - Type-safe: Uses int32 for types (see golang/constants for mapping)
// - Memory-optimized: Careful field ordering for cache efficiency
// - Position-aware: Tracks byte positions and line numbers for all nodes
//
// Usage Flow:
// 1. Parse source files -> language-specific AST (go/ast, etc.)
// 2. Transform AST -> unified syntax.Node tree (see syntax/golang/)
// 3. Build suffix tree from Node sequence (suffixtree.Update())
// 4. Find duplicates using suffix tree (FindDuplOver())
// 5. Convert matches to complete syntax units (FindSyntaxUnits())
//
// Key Functions:
// - FindSyntaxUnits(): Converts suffix tree matches to complete syntax units
// - hashSeq(): Creates hash of node sequence for duplicate detection
// - isCyclic/spansMultipleFiles(): Validation helpers
//
// Performance:
// - maxChildrenSerial constant prevents goroutine stack overflow
// - Node struct is 40B (37.5% reduction from 64B) via int32 fields
// - See MEMORY_LAYOUT_OPTIMIZATION_PLAN.md for details
package syntax

import (
	"github.com/LarsArtmann/art-dupl/suffixtree"
)

// To avoid "goroutine stack exceeds" with gigantic slices (Composite Literals).
// 10_000 => 0.89s
// 20_000 => 1.53s
// 30_000 => 2.57s
// 40_000 => 3.89s
// 50_000 => 5.58s
// 60_000 => 7.95s
// 70_000 => 10.15s
// 80_000 => 13.11s
// 90_000 => 16.62s
// 100_000 => 21.42s.
const maxChildrenSerial = 10_000

// Node represents a syntax tree node.
//
// Memory Layout Optimized with int32 fields:
// - int32 fields grouped for cache efficiency (4B each, 16B total)
// - pointer field (8B)
// - string headers at end (32B: Filename + Name)
// - Fingerprint int32 + Statement bool fit in trailing padding (4B+1B+3B)
// Total: 64B.
//
// The Name field stores the original Go identifier for Ident, SelectorExpr,
// and FuncDecl nodes. It enables actionability analysis to distinguish
// `nil` from `err`, `Unlock` from `Close`, etc.
// Empty string means "not an identifier-bearing node" or structural-only mode.
//
// The Statement field marks direct children of block-like nodes (BlockStmt,
// CaseClause.Body, CommClause.Body). When true, serial() fingerprints the
// entire subtree into a single composite Type token instead of emitting
// each descendant individually. This makes threshold mean "N duplicated
// statements" rather than "N arbitrary AST nodes.".
//
// The VarType field stores the go/types type string for Ident nodes when
// type-aware mode is active. Empty string means no type info available.
// The actionability layer uses this to detect helper-dominance and
// type-aware false positives.
//
// The EnclosingReturnArity field stores the number of return values in the
// enclosing function (0 for void, >0 for functions returning values).
// The property engine uses this to detect control-flow traps where bare
// returns are forced by the function signature (e.g., http.HandlerFunc).
//
// The InterfaceMethod field is set on FuncDecl nodes when type-aware mode is
// active and go/types confirms the method satisfies an interface declared in
// the same package. The flag is also propagated to body statement nodes by the
// transformer (same save/restore pattern as EnclosingReturnArity) because
// FuncDecl nodes are never clone roots in Go files. The actionability layer
// uses this to suppress interface-contract boilerplate.
type Node struct {
	Type                 int32
	Pos                  int32
	End                  int32
	Owns                 int32
	Children             []*Node
	Filename             string
	Name                 string
	VarType              string
	Statement            bool
	Fingerprint          int32
	EnclosingReturnArity int32
	InterfaceMethod      bool

	// IsAlias is set on TypeSpec nodes when the source uses `type X = Y`
	// (alias syntax) rather than `type X Y` (named type definition). The
	// actionability layer uses this to distinguish re-export shims (aliases,
	// non-actionable) from named type definitions (potentially actionable).
	IsAlias bool
}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) AddChildren(children ...*Node) {
	n.Children = append(n.Children, children...)
}

// Clone returns a deep copy of the node subtree. The incremental cache stores
// serialized node trees that are shared across goroutines on a cache hit, where
// each file rewrites Filename; Clone yields an independent copy safe to mutate.
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}

	clone := &Node{
		Type:                 n.Type,
		Pos:                  n.Pos,
		End:                  n.End,
		Owns:                 n.Owns,
		Children:             n.Children,
		Filename:             n.Filename,
		Name:                 n.Name,
		VarType:              n.VarType,
		Statement:            n.Statement,
		Fingerprint:          n.Fingerprint,
		EnclosingReturnArity: n.EnclosingReturnArity,
		InterfaceMethod:      n.InterfaceMethod,
		IsAlias:              n.IsAlias,
	}
	if len(n.Children) > 0 {
		clone.Children = make([]*Node, len(n.Children))
		for i, c := range n.Children {
			clone.Children[i] = c.Clone()
		}
	}

	return clone
}

// Val returns the token value for suffix tree compatibility.
// Implements the suffixtree.Token interface.
//
// For statement nodes, returns the composite Fingerprint so the suffix tree
// matches at statement granularity. For non-statement nodes, returns the
// semantic-encoded Type.
func (n *Node) Val() suffixtree.TokenValue {
	if n.Statement {
		return suffixtree.TokenValue(n.Fingerprint)
	}

	return suffixtree.TokenValue(n.Type)
}

// NewSyntheticFileNode creates a synthetic node representing an entire file.
// This is used for file-level duplicate detection where we want to match
// entire files rather than specific code fragments.
//

func NewSyntheticFileNode(filename string, size int) *Node {
	return &Node{
		Filename: InternFilename(filename),
		Pos:      0,
		End:      int32(size),
		Type:     1,
		// Owns and Children intentionally omitted - will be set later
	}
}

type Match struct {
	Hash  string
	Frags [][]*Node
}

func Serialize(n *Node) []*Node {
	return SerializeWithMaxChildren(n, maxChildrenSerial)
}

// SerializeWithMaxChildren serializes a node tree with a caller-specified
// maximum children cap. This allows the job layer to thread config.MaxChildrenSerial
// into the serialization without a global variable.
func SerializeWithMaxChildren(n *Node, maxChildren int) []*Node {
	if maxChildren <= 0 {
		maxChildren = maxChildrenSerial
	}

	stream := make([]*Node, 0, 10)
	serial(n, &stream, maxChildren)

	return stream
}

func serial(n *Node, stream *[]*Node, maxChildren int) int {
	// Shallow-copy the node so mutations (fingerprinting, Owns counting)
	// never corrupt the original tree. This makes Serialize idempotent and
	// safe for concurrent access to cached trees.
	node := &Node{
		Type:                 n.Type,
		Pos:                  n.Pos,
		End:                  n.End,
		Owns:                 n.Owns,
		Children:             n.Children,
		Filename:             n.Filename,
		Name:                 n.Name,
		VarType:              n.VarType,
		Statement:            n.Statement,
		Fingerprint:          n.Fingerprint,
		EnclosingReturnArity: n.EnclosingReturnArity,
		InterfaceMethod:      n.InterfaceMethod,
		IsAlias:              n.IsAlias,
	}
	*stream = append(*stream, node)

	if n.Statement {
		// Statement-level tokenization: fingerprint the entire subtree into
		// one composite value so the suffix tree matches at statement granularity.
		// Children remain in memory for classification/actionability but are not
		// emitted as individual tokens.
		//
		// The fingerprint is stored in the Fingerprint field, NOT in Type, so
		// that DecodeBaseType(Type) still returns the correct base AST type for
		// actionability analysis.
		node.Fingerprint = fingerprintSubtree(n)
		node.Owns = 0

		return 1
	}

	var count int

	for i, child := range n.Children {
		if i > maxChildren {
			break
		}

		count += serial(child, stream, maxChildren)
	}

	node.Owns = int32(count) // #nosec G115 -- Child count bounded by maxChildren

	return int(node.Owns) + 1
}

// fingerprintSubtree hashes the pre-order Type sequence of a node and all its
// descendants into a single int32 using FNV-1a. This produces a deterministic,
// order-sensitive composite token representing one complete statement.
//
// The function reads original Type values set during trans() (which already
// include semantic encoding: alpha-normalized identifiers, literal values,
// operator hashes). Two statements with identical semantic content produce
// identical fingerprints, enabling Type 1/Type 2 clone detection at statement
// granularity.
func fingerprintSubtree(n *Node) int32 {
	return int32(fingerprintSubtreeInto(n, fnvOffset32))
}

func fingerprintSubtreeInto(n *Node, hash uint32) uint32 {
	hash = fnvStep32(hash, uint32(n.Type))

	for _, child := range n.Children {
		hash = fingerprintSubtreeInto(child, hash)
	}

	return hash
}

func fnvStep32(hash, value uint32) uint32 {
	return (hash ^ value) * fnvPrime32
}

// FindSyntaxUnits finds all complete syntax units in the match group and returns them
// with the corresponding hash.
func FindSyntaxUnits(data []*Node, m suffixtree.Match, threshold int) Match {
	if len(m.Ps) == 0 {
		return Match{}
	}

	firstSeq := data[m.Ps[0] : m.Ps[0]+m.Len]
	indexes := getUnitsIndexes(firstSeq, threshold)

	// Statement-level threshold: when the match contains statement tokens,
	// require at least `threshold` statements. In legacy mode (no statement
	// tokens), the per-node threshold check in getUnitsIndexes is sufficient.
	if len(indexes) > 0 && firstSeq[indexes[0]].Statement && len(indexes) < threshold {
		return Match{}
	}

	// When this file uses statement-level tokenization but the match falls
	// entirely in the non-statement portion (structural wrappers like
	// FuncDecl/File), the legacy Owns-based indexes are not meaningful clone
	// units — skip them. Only applies to Go files; templ files may have
	// non-statement component declarations that are valid clone units.
	if len(indexes) > 0 && !firstSeq[indexes[0]].Statement &&
		fileContainsStatements(data, firstSeq[indexes[0]].Filename) &&
		!isTemplFile(firstSeq[indexes[0]].Filename) {
		return Match{}
	}

	if len(indexes) > 0 && len(m.Ps) > 1 {
		indexes = validateOwnershipConsistency(data, m, firstSeq, indexes)
	}

	if len(indexes) == 0 || isCyclic(indexes, firstSeq) || spansMultipleFiles(indexes, firstSeq) {
		return Match{}
	}

	return buildMatch(data, m, firstSeq, indexes)
}

// Unique removes duplicate entries from a group of syntax nodes based on file and range.
// Two clones are considered duplicates if they have the same filename and the same range (start-end).
func Unique(group [][]*Node) [][]*Node {
	type rangeKey struct {
		filename string
		start    int
		end      int
	}

	seen := make(map[rangeKey]struct{})

	var newGroup [][]*Node

	for _, seq := range group {
		if len(seq) == 0 {
			continue
		}

		first := seq[0]
		last := seq[len(seq)-1]

		key := rangeKey{filename: first.Filename, start: int(first.Pos), end: int(last.End)}
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}

			newGroup = append(newGroup, seq)
		}
	}

	return newGroup
}

// CountUniqueFiles returns the number of unique files in a clone group.
func CountUniqueFiles(group [][]*Node) int {
	uniqueFiles := make(map[string]bool)

	for _, seq := range group {
		if len(seq) > 0 {
			uniqueFiles[seq[0].Filename] = true
		}
	}

	return len(uniqueFiles)
}
