package syntax

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LarsArtmann/art-dupl/pkg/position"
)

// memberLocation identifies one clone fragment by its DISPLAYED location:
// filename plus the 1-based line range a reader sees as `file:line-line`.
type memberLocation struct {
	filename  string
	lineStart int
	lineEnd   int
}

// groupMembers summarizes a clone group for member-set deduplication.
type groupMembers struct {
	locations  []memberLocation
	tokenSpan  int // stream tokens covered by all fragments (sum of Owns+1)
	byteExtent int // summed source extent of all fragments, in bytes
	readable   bool
}

// DeduplicateMemberSets removes groups whose displayed member sets are
// identical to another group's (issue #10): the suffix tree can match the
// same source ranges at different granularities — e.g. a templ declaration
// view and an expression view of one signature — producing two groups that
// print the exact same `file:line-line` entries. To a reader that is the
// same finding reported twice.
//
// For each duplicate member set only the RICHEST group survives: the one
// covering more stream tokens (more syntax structure), then the larger byte
// extent, then the lexicographically smallest hash as a deterministic
// tie-break. Groups with unreadable member files are exempt — a read error
// must never decide which finding disappears.
//
// readFile supplies file content for byte-offset→line conversion using the
// same position.ByteRangeToLines semantics the printers display, so the
// member key always matches what the output shows.
func DeduplicateMemberSets(
	readFile func(string) ([]byte, error),
	groups map[string][][]*Node,
) map[string][][]*Node {
	if len(groups) <= 1 {
		return groups
	}

	contentCache := make(map[string][]byte)

	membersByHash := make(map[string]groupMembers, len(groups))
	for hash, frags := range groups {
		membersByHash[hash] = computeGroupMembers(readFile, frags, contentCache)
	}

	keep := chooseDuplicateWinners(groups, membersByHash)

	result := make(map[string][][]*Node, len(keep))
	for hash := range keep {
		result[hash] = groups[hash]
	}

	return result
}

// computeGroupMembers derives the displayed member locations and richness
// metrics for one group. Files are read through contentCache at most once.
func computeGroupMembers(
	readFile func(string) ([]byte, error),
	frags [][]*Node,
	contentCache map[string][]byte,
) groupMembers {
	members := groupMembers{
		locations: make([]memberLocation, 0, len(frags)),
		readable:  true,
	}

	for _, frag := range frags {
		if len(frag) == 0 {
			continue
		}

		first := frag[0]
		last := frag[len(frag)-1]

		content, ok := contentCache[first.Filename]
		if !ok {
			data, err := readFile(first.Filename)
			if err != nil {
				contentCache[first.Filename] = nil
				members.readable = false
			} else {
				contentCache[first.Filename] = data
				content = data
			}
		} else if content == nil {
			members.readable = false
		}

		lineStart, lineEnd := position.ByteRangeToLines(content, int(first.Pos), int(last.End))
		members.locations = append(members.locations, memberLocation{
			filename:  first.Filename,
			lineStart: lineStart,
			lineEnd:   lineEnd,
		})

		for _, node := range frag {
			members.tokenSpan += int(node.Owns) + 1
		}

		members.byteExtent += int(last.End) - int(first.Pos)
	}

	sort.Slice(members.locations, func(i, j int) bool {
		a, b := members.locations[i], members.locations[j]
		if a.filename != b.filename {
			return a.filename < b.filename
		}

		if a.lineStart != b.lineStart {
			return a.lineStart < b.lineStart
		}

		return a.lineEnd < b.lineEnd
	})

	return members
}

// memberSetKey renders the sorted locations as a comparable string.
func (m groupMembers) memberSetKey() string {
	parts := make([]string, 0, len(m.locations))
	for _, loc := range m.locations {
		parts = append(parts, fmt.Sprintf("%s:%d-%d", loc.filename, loc.lineStart, loc.lineEnd))
	}

	return strings.Join(parts, "|")
}

// chooseDuplicateWinners picks, per identical member set, the richest group.
// Groups whose member files could not be read are exempt: their locations
// are unknowable, so they are always kept and never compared.
func chooseDuplicateWinners(
	groups map[string][][]*Node,
	membersByHash map[string]groupMembers,
) map[string]bool {
	byKey := make(map[string][]string)
	for hash := range groups {
		if !membersByHash[hash].readable {
			continue
		}

		key := membersByHash[hash].memberSetKey()
		byKey[key] = append(byKey[key], hash)
	}

	keep := make(map[string]bool, len(groups))
	for _, hashes := range byKey {
		if len(hashes) == 1 {
			keep[hashes[0]] = true
			continue
		}

		winner := hashes[0]
		for _, candidate := range hashes[1:] {
			if richerDuplicate(membersByHash[candidate], membersByHash[winner], candidate, winner) {
				winner = candidate
			}
		}

		keep[winner] = true
	}

	return keep
}

// richerDuplicate reports whether candidate a is a richer view than b for
// member-set dedup: more covered stream tokens first, then larger byte
// extent, then the smaller hash for determinism when nothing else differs.
func richerDuplicate(a, b groupMembers, hashA, hashB string) bool {
	if a.tokenSpan != b.tokenSpan {
		return a.tokenSpan > b.tokenSpan
	}

	if a.byteExtent != b.byteExtent {
		return a.byteExtent > b.byteExtent
	}

	return hashA < hashB
}
