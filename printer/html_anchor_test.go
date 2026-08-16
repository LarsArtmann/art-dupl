package printer

import (
	"regexp"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestGroupAnchorIDDerivation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hash     string
		groupNum int
		want     string
	}{
		{
			name: "hex hash passes through with prefix", hash: "e0f6093241ba9931",
			groupNum: 1, want: "group-e0f6093241ba9931",
		},
		{
			name: "leading digit is fine after prefix", hash: "0123456789abcdef",
			groupNum: 3, want: "group-0123456789abcdef",
		},
		{name: "non-URL-safe chars are stripped", hash: "ab/cd!e", groupNum: 1, want: "group-abcde"},
		{name: "empty hash falls back to group number", hash: "", groupNum: 7, want: "group-7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := groupAnchorID(tt.hash, tt.groupNum); got != tt.want {
				t.Errorf("groupAnchorID(%q, %d) = %q, want %q", tt.hash, tt.groupNum, got, tt.want)
			}
		})
	}
}

func TestGroupAnchorIDStableAcrossReorder(t *testing.T) {
	t.Parallel()

	clones := []domain.ProcessedClone{
		{CloneRef: domain.CloneRef{Filename: "a.go", LineStart: 1, LineEnd: 2, Fragment: "x"}},
	}

	first := toCloneGroupView(1, "e0f6093241ba9931", clones)
	second := toCloneGroupView(5, "e0f6093241ba9931", clones)

	if first.AnchorID != second.AnchorID {
		t.Errorf("AnchorID changed with group position: %q (num %d) vs %q (num %d)",
			first.AnchorID, first.GroupNum, second.AnchorID, second.GroupNum)
	}

	if first.GroupNum == second.GroupNum {
		t.Error("test setup: group numbers should differ to prove stability")
	}
}

func mustPrintGroupWithHash(t *testing.T, p Printer, hash string, dups [][]*syntax.Node) {
	t.Helper()

	group, err := NodesToGroup(mockReadFile(testPlumbCode), hash, dups)
	if err != nil {
		t.Fatalf("NodesToGroup() error: %v", err)
	}

	if err := p.PrintClones(group); err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}
}

func TestHTMLGroupAnchorsAreDeepLinkableAndUnique(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithContent(testGoMultiCode)

	mustPrintHeader(t, p)

	nodes1 := nodesForHTML(testFilename)
	nodes2 := nodesForHTML(testFilename)

	mustPrintGroupWithHash(t, p, "aabbccdd11223344", [][]*syntax.Node{nodes1})
	mustPrintGroupWithHash(t, p, "0123456789abcdef", [][]*syntax.Node{nodes2})

	output := buf.String()

	for _, id := range []string{"group-aabbccdd11223344", "group-0123456789abcdef"} {
		if !strings.Contains(output, `id="`+id+`"`) {
			t.Errorf("output should contain group div id %q", id)
		}

		if !strings.Contains(output, `href="#`+id+`"`) {
			t.Errorf("output should contain deep-link href for %q", id)
		}
	}

	idRe := regexp.MustCompile(`id="(group-[^"]+)"`)
	matches := idRe.FindAllStringSubmatch(output, -1)

	seen := make(map[string]bool, len(matches))
	for _, m := range matches {
		if seen[m[1]] {
			t.Errorf("duplicate group id %q — anchors must be unique", m[1])
		}

		seen[m[1]] = true
	}

	if got := len(matches); got != 2 {
		t.Errorf("expected exactly 2 group ids, got %d: %v", got, matches)
	}
}

func TestHTMLAnchorMatchesCloneGroupHash(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithContent(testGoMultiCode)

	mustPrintHeader(t, p)

	mustPrintGroupWithHash(t, p, "deadbeefdeadbeef", [][]*syntax.Node{nodesForHTML(testFilename)})

	output := buf.String()

	// The anchor is derived from the group's content hash, so a deep link
	// stays valid for the same duplicated code across re-runs and re-sorts.
	if !strings.Contains(output, `id="group-deadbeefdeadbeef"`) {
		t.Error("anchor should be derived from the group hash")
	}
}
