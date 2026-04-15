package cli

import (
	"fmt"
	"sort"
	"testing"

	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// sortingTestCase represents a test case for sorting algorithms.
type sortingTestCase struct {
	name          string
	matches       []syntax.Match
	expectedOrder []string // Hashes in expected order
}

// TestOccurrenceSorting tests that occurrence sorting uses unique counts, not total counts.
func TestOccurrenceSorting(t *testing.T) {
	tests := []sortingTestCase{
		{
			name: "simple descending by unique count",
			matches: []syntax.Match{
				{
					Hash:  "hash3",
					Frags: createFragments(3), // 3 unique fragments
				},
				{
					Hash:  "hash1",
					Frags: createFragments(8), // 8 unique fragments (most, should be first)
				},
				{
					Hash:  "hash2",
					Frags: createFragments(5), // 5 unique fragments
				},
			},
			expectedOrder: []string{"hash1", "hash2", "hash3"}, // 8, 5, 3
		},
		{
			name: "with duplicates in fragments (bug regression test)",
			matches: []syntax.Match{
				{
					Hash:  "hash1",
					Frags: createFragmentsWithDuplicates(4, 6), // 4 unique, 6 total
				},
				{
					Hash:  "hash2",
					Frags: createFragmentsWithDuplicates(5, 5), // 5 unique, 5 total
				},
				{
					Hash:  "hash3",
					Frags: createFragmentsWithDuplicates(3, 9), // 3 unique, 9 total
				},
			},
			expectedOrder: []string{
				"hash2",
				"hash1",
				"hash3",
			}, // 5, 4, 3 unique counts (not 9, 6, 5 totals)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := buildGroupsFromMatches(tt.matches)

			// Get keys
			keys := make([]string, 0, len(groups))
			for k := range groups {
				keys = append(keys, k)
			}

			// Pre-compute unique counts for sorting (like printDupls)
			uniqueCounts := make(map[string]int)
			for k, v := range groups {
				uniqueCounts[k] = len(syntax.Unique(v))
			}

			// Sort by occurrence (descending unique count)
			sort.Slice(keys, func(i, j int) bool {
				return uniqueCounts[keys[i]] > uniqueCounts[keys[j]]
			})

			// Verify order
			assertSortedOrder(t, keys, tt.expectedOrder)

			// Also verify the unique counts are correct
			for hash, expectedUniqueCount := range map[string]int{
				"hash1": len(syntax.Unique(groups["hash1"])),
				"hash2": len(syntax.Unique(groups["hash2"])),
				"hash3": len(syntax.Unique(groups["hash3"])),
			} {
				t.Logf(
					"%s: %d unique out of %d total",
					hash,
					expectedUniqueCount,
					len(groups[hash]),
				)
			}
		})
	}
}

// buildGroupsFromMatches builds a groups map from matches (like printDupls does).
func buildGroupsFromMatches(matches []syntax.Match) map[string][][]*syntax.Node {
	groups := make(map[string][][]*syntax.Node)
	for _, match := range matches {
		groups[match.Hash] = append(groups[match.Hash], match.Frags...)
	}
	return groups
}

// assertSortedOrder verifies that the keys are in the expected order.
func assertSortedOrder(t *testing.T, keys, expectedOrder []string) {
	t.Helper()

	if len(keys) != len(expectedOrder) {
		t.Fatalf("Expected %d groups, got %d", len(expectedOrder), len(keys))
	}

	for i, expectedHash := range expectedOrder {
		if keys[i] != expectedHash {
			t.Errorf("Position %d: expected hash %s, got %s", i, expectedHash, keys[i])
		}
	}
}

// TestSizeSorting tests that size sorting works correctly.
func TestSizeSorting(t *testing.T) {
	tests := []sortingTestCase{
		{
			name: "descending by size",
			matches: []syntax.Match{
				{
					Hash:  "hash2",
					Frags: createFragmentsWithSize(50),
				},
				{
					Hash:  "hash1",
					Frags: createFragmentsWithSize(100), // Largest, should be first
				},
				{
					Hash:  "hash3",
					Frags: createFragmentsWithSize(25),
				},
			},
			expectedOrder: []string{"hash1", "hash2", "hash3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := buildGroupsFromMatches(tt.matches)

			keys := make([]string, 0, len(groups))
			for k := range groups {
				keys = append(keys, k)
			}

			// Sort by size (descending)
			sort.Slice(keys, func(i, j int) bool {
				return printer.GetCloneSize(groups[keys[i]]) > printer.GetCloneSize(groups[keys[j]])
			})

			// Verify order
			assertSortedOrder(t, keys, tt.expectedOrder)
		})
	}
}

// createFragments creates specified number of unique fragments.
func createFragments(count int) [][]*syntax.Node {
	fragments := make([][]*syntax.Node, count)
	for i := range count {
		node := &syntax.Node{
			Type:     int32(i),
			Filename: fmt.Sprintf("file%c.go", 'a'+i),
			Pos:      int32(i * 10),
			End:      int32(i*10 + 5),
			Owns:     5,
		}
		fragments[i] = []*syntax.Node{node}
	}

	return fragments
}

// createFragmentsWithDuplicates creates fragments with duplicates to test bug fix.
func createFragmentsWithDuplicates(uniqueCount, totalCount int) [][]*syntax.Node {
	fragments := make([][]*syntax.Node, totalCount)
	for i := range totalCount {
		// Create duplicates by reusing uniqueCount positions
		uniqueIndex := i % uniqueCount
		node := &syntax.Node{
			Type:     int32(uniqueIndex),
			Filename: fmt.Sprintf("file%c.go", 'a'+uniqueIndex),
			Pos:      int32(uniqueIndex * 10),
			End:      int32(uniqueIndex*10 + 5),
			Owns:     5,
		}
		fragments[i] = []*syntax.Node{node}
	}

	return fragments
}

// createFragmentsWithSize creates fragments with specified size (Owns value).
func createFragmentsWithSize(size int) [][]*syntax.Node {
	node := &syntax.Node{
		Type:     1,
		Filename: "test.go",
		Pos:      0,
		End:      int32(size),
		Owns:     int32(size),
	}

	return [][]*syntax.Node{{node}}
}
