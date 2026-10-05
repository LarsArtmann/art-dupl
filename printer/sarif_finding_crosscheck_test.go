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
				RuleID     string `json:"ruleId"`
				Level      string `json:"level"`
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
