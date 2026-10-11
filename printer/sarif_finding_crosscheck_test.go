package printer

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer/finding"
)

// TestSARIFCrossCheckCLIvsGoFinding renders the same classified groups
// through BOTH SARIF producers — the CLI's hand-rolled printer/sarif.go and
// go-finding's Report.ToSARIF() via the printer/finding adapter — and
// requires identical (ruleId, level, go-finding/groupId) tuples per
// (file, line). The two paths are intentionally different implementations
// (plan T09): this cross-check is what makes their agreement a pinned
// contract instead of a coincidence, so a future change to either ladder or
// property emission fails here before it can desynchronize consumers that
// diff CLI SARIF against BuildFlow/go-finding SARIF.
func TestSARIFCrossCheckCLIvsGoFinding(t *testing.T) {
	t.Parallel()

	const threshold = 5

	// Three groups, one per severity-ladder rung at threshold 5:
	// 40 >= 4*5 error, 12 >= 2*5 warning, 6 < 10 note.
	clone := func(file string, tokens int) domain.ProcessedClone {
		return domain.ProcessedClone{
			Filename:   file,
			LineStart:  10,
			LineEnd:    29,
			Fragment:   "func a() {\n\tprintln(1)\n}",
			StartPos:   100,
			EndPos:     220,
			TokenCount: tokens,
			Classification: domain.CloneClassification{
				CloneType: domain.CloneType1,
			},
		}
	}

	groups := []domain.ProcessedCloneGroup{
		{Hash: "cross000000000001", Clones: []domain.ProcessedClone{clone("err_a.go", 40)}},
		{Hash: "cross000000000002", Clones: []domain.ProcessedClone{clone("warn_a.go", 12)}},
		{Hash: "cross000000000003", Clones: []domain.ProcessedClone{clone("note_a.go", 6)}},
	}

	var cli bytes.Buffer

	cliPrinter := NewSARIF(&cli, mockSARIFReadFile, threshold).(*sarifPrinter)
	if err := cliPrinter.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader: %v", err)
	}

	for _, group := range groups {
		cliPrinter.SetHash(group.Hash)

		if err := cliPrinter.PrintClones(group); err != nil {
			t.Fatalf("PrintClones(%s): %v", group.Hash, err)
		}
	}

	if err := cliPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter: %v", err)
	}

	gfData, err := finding.ToReport(groups, finding.Options{Threshold: threshold}).ToSARIF()
	if err != nil {
		t.Fatalf("ToReport().ToSARIF(): %v", err)
	}

	cliRows := extractSARIFRows(t, cli.Bytes(), "printer/sarif.go")
	gfRows := extractSARIFRows(t, gfData, "Report.ToSARIF()")

	if len(cliRows) == 0 {
		t.Fatal("CLI SARIF produced no rows")
	}

	if len(cliRows) != len(gfRows) {
		t.Fatalf("row count: printer/sarif.go=%d, Report.ToSARIF()=%d", len(cliRows), len(gfRows))
	}

	for key, want := range cliRows {
		got, ok := gfRows[key]
		if !ok {
			t.Errorf("Report.ToSARIF() lost the result for %v (CLI: %+v)", key, want)

			continue
		}

		if got != want {
			t.Errorf("divergence at %v: printer/sarif.go=%+v, Report.ToSARIF()=%+v", key, want, got)
		}
	}
}

// TestSARIFCrossCheckRelatedLocations extends the T09 contract to
// result.relatedLocations for multi-occurrence groups: for every result,
// both producers must list the same sibling occurrence locations with the
// same relation message. Region snippets are exact parity as of 2026-10-11
// (M5): NEITHER producer emits them on related regions — the CLI previously
// duplicated the fragment text per sibling (9.1% of SARIF bytes at -t 1 on
// go-cqrs-lite) and was trimmed to match go-finding's sarifRelatedLocs,
// whose import path drops related snippets anyway. The go-finding/id
// property on go-finding's related entries remains an accepted delta (the
// CLI result set has no per-occurrence ids; it carries fingerprints
// instead).
func TestSARIFCrossCheckRelatedLocations(t *testing.T) {
	t.Parallel()

	const threshold = 5

	clone := func(file string, start int) domain.ProcessedClone {
		return domain.ProcessedClone{
			Filename:   file,
			LineStart:  start,
			LineEnd:    start + 19,
			Fragment:   "func a() {\n\tprintln(1)\n}",
			StartPos:   100,
			EndPos:     220,
			TokenCount: 30,
			Classification: domain.CloneClassification{
				CloneType: domain.CloneType1,
			},
		}
	}

	group := domain.ProcessedCloneGroup{
		Hash: "relcross00000001",
		Clones: []domain.ProcessedClone{
			clone("rel_a.go", 10),
			clone("rel_b.go", 40),
			clone("rel_c.go", 70),
		},
	}

	var cli bytes.Buffer

	cliPrinter := NewSARIF(&cli, mockSARIFReadFile, threshold).(*sarifPrinter)
	if err := cliPrinter.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader: %v", err)
	}

	cliPrinter.SetHash(group.Hash)

	if err := cliPrinter.PrintClones(group); err != nil {
		t.Fatalf("PrintClones: %v", err)
	}

	if err := cliPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter: %v", err)
	}

	gfData, err := finding.ToReport(
		[]domain.ProcessedCloneGroup{group},
		finding.Options{Threshold: threshold},
	).ToSARIF()
	if err != nil {
		t.Fatalf("ToReport().ToSARIF(): %v", err)
	}

	cliRel := extractRelatedLocations(t, cli.Bytes(), "printer/sarif.go")
	gfRel := extractRelatedLocations(t, gfData, "Report.ToSARIF()")

	if len(cliRel) != len(gfRel) {
		t.Fatalf(
			"result count: printer/sarif.go=%d, Report.ToSARIF()=%d",
			len(cliRel),
			len(gfRel),
		)
	}

	for key, want := range cliRel {
		got, ok := gfRel[key]
		if !ok {
			t.Errorf("Report.ToSARIF() lost result %s (related: %v)", key, want)

			continue
		}

		if got.message != want.message {
			t.Errorf(
				"related message divergence at %s: printer/sarif.go=%q, Report.ToSARIF()=%q",
				key,
				want.message,
				got.message,
			)
		}

		if got.siblings == nil || want.siblings == nil {
			t.Fatalf("related set must not be nil at %s", key)
		}

		if !got.siblings.equal(want.siblings) {
			t.Errorf(
				"related location divergence at %s: printer/sarif.go=%v, Report.ToSARIF()=%v",
				key,
				want.siblings.slice(),
				got.siblings.slice(),
			)
		}
	}
}

// relatedSet is the set of sibling occurrence locations a result links to.
type relatedSet map[[2]string]bool

func (r relatedSet) equal(other relatedSet) bool {
	if len(r) != len(other) {
		return false
	}

	for k := range r {
		if !other[k] {
			return false
		}
	}

	return true
}

func (r relatedSet) slice() [][2]string {
	out := make([][2]string, 0, len(r))
	for k := range r {
		out = append(out, k)
	}

	return out
}

type relatedRow struct {
	message  string
	siblings relatedSet
}

func extractRelatedLocations(
	tb testing.TB,
	data []byte,
	label string,
) map[string]relatedRow {
	tb.Helper()

	var doc struct {
		Runs []struct {
			Results []struct {
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
				RelatedLocations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
					Message struct {
						Text string `json:"text"`
					} `json:"message"`
				} `json:"relatedLocations"`
			} `json:"results"`
		} `json:"runs"`
	}

	if err := json.Unmarshal(data, &doc); err != nil {
		tb.Fatalf("unmarshal %s SARIF: %v", label, err)
	}

	rows := make(map[string]relatedRow)

	for _, run := range doc.Runs {
		for _, res := range run.Results {
			if len(res.Locations) == 0 {
				tb.Fatalf("%s result has no location", label)
			}

			self := res.Locations[0].PhysicalLocation
			key := self.ArtifactLocation.URI + ":" +
				strconv.Itoa(self.Region.StartLine)

			siblings := make(relatedSet, len(res.RelatedLocations))

			message := ""
			for i, rel := range res.RelatedLocations {
				if i == 0 {
					message = rel.Message.Text
				} else if rel.Message.Text != message {
					tb.Fatalf(
						"%s result %s: mixed related messages %q vs %q",
						label,
						key,
						message,
						rel.Message.Text,
					)
				}

				siblings[[2]string{
					rel.PhysicalLocation.ArtifactLocation.URI,
					strconv.Itoa(rel.PhysicalLocation.Region.StartLine),
				}] = true
			}

			rows[key] = relatedRow{message: message, siblings: siblings}
		}
	}

	return rows
}

// sarifRow is the projection both SARIF paths must agree on.
type sarifRow struct {
	RuleID  string
	Level   string
	GroupID string
}

func extractSARIFRows(tb testing.TB, data []byte, label string) map[[2]string]sarifRow {
	tb.Helper()

	var doc struct {
		Runs []struct {
			Results []struct {
				RuleID    string `json:"ruleId"`
				Level     string `json:"level"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
				Properties map[string]any `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}

	if err := json.Unmarshal(data, &doc); err != nil {
		tb.Fatalf("unmarshal %s SARIF: %v", label, err)
	}

	rows := make(map[[2]string]sarifRow)

	for _, run := range doc.Runs {
		for _, res := range run.Results {
			if len(res.Locations) == 0 {
				tb.Fatalf("%s result %q has no location", label, res.RuleID)
			}

			key := [2]string{
				res.Locations[0].PhysicalLocation.ArtifactLocation.URI,
				strconv.Itoa(res.Locations[0].PhysicalLocation.Region.StartLine),
			}

			groupID, _ := res.Properties[finding.SARIFPropGroupID].(string)

			rows[key] = sarifRow{RuleID: res.RuleID, Level: res.Level, GroupID: groupID}
		}
	}

	return rows
}
