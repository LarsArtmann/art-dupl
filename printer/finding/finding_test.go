package finding_test

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer/finding"
	gofinding "github.com/larsartmann/go-finding"
)

func testGroup(hash string, tokensPerClone ...int) domain.ProcessedCloneGroup {
	tokens := func(i int) int {
		if i < len(tokensPerClone) {
			return tokensPerClone[i]
		}

		return 10
	}

	clones := []domain.ProcessedClone{
		{
			Filename:   "a.go",
			LineStart:  10,
			LineEnd:    19,
			Fragment:   "func a() {\n\tprintln(1)\n}",
			StartPos:   100,
			EndPos:     220,
			TokenCount: tokens(0),
			FileSize:   1000,
			Classification: domain.CloneClassification{
				Category:      domain.CategoryFunction,
				Priority:      domain.PriorityMedium,
				Actionability: domain.Actionable,
				CloneType:     domain.CloneType1,
				Tokens:        tokens(0),
				Lines:         10,
			},
		},
		{
			Filename:   "b.go",
			LineStart:  30,
			LineEnd:    39,
			Fragment:   "func b() {\n\tprintln(2)\n}",
			StartPos:   300,
			EndPos:     420,
			TokenCount: tokens(1),
			FileSize:   2000,
			Classification: domain.CloneClassification{
				Category:             domain.CategoryFunction,
				Priority:             domain.PriorityMedium,
				Actionability:        domain.NonActionable,
				CloneType:            domain.CloneType2,
				NonActionablePattern: "signature-only",
				Tokens:               tokens(1),
				Lines:                10,
			},
		},
	}

	return domain.NewProcessedCloneGroup(hash, clones)
}

func TestGroupIDOfIsDeterministicAndMachineSafe(t *testing.T) {
	group := testGroup("abcdef0123456789", 10)

	got := finding.GroupIDOf(group)
	if got != gofinding.GroupID("abcdef0123456789") {
		t.Errorf("GroupIDOf() = %q, want the group hash", got)
	}

	if got != finding.GroupIDOf(group) {
		t.Error("GroupIDOf() must be deterministic for identical input")
	}

	if !got.IsValid() {
		t.Errorf("GroupIDOf() = %q must pass go-finding machine-safety validation", got)
	}
}

func TestGroupIDOfEmptyHashIsNotGrouped(t *testing.T) {
	group := testGroup("", 10)

	if got := finding.GroupIDOf(group); got != "" {
		t.Errorf("GroupIDOf(empty hash) = %q, want empty (not grouped)", got)
	}
}

func TestToFindingsOneFindingPerCloneSharingGroupID(t *testing.T) {
	group := testGroup("group0123456789a", 10)

	findings := finding.ToFindings(group, finding.Options{Version: "test"})
	if len(findings) != len(group.Clones) {
		t.Fatalf("ToFindings() = %d findings, want %d (one per clone)", len(findings), len(group.Clones))
	}

	wantID := finding.GroupIDOf(group)
	for i, f := range findings {
		if f.GroupID != wantID {
			t.Errorf("findings[%d].GroupID = %q, want %q", i, f.GroupID, wantID)
		}

		if f.ToolName != finding.ToolName {
			t.Errorf("findings[%d].ToolName = %q, want %q", i, f.ToolName, finding.ToolName)
		}

		if f.Rule != finding.RuleCloneDetected {
			t.Errorf("findings[%d].Rule = %q, want %q", i, f.Rule, finding.RuleCloneDetected)
		}

		if f.Category != gofinding.CategoryDuplication {
			t.Errorf("findings[%d].Category = %q, want %q", i, f.Category, gofinding.CategoryDuplication)
		}

		if f.Position.File != gofinding.FilePath(group.Clones[i].Filename) {
			t.Errorf("findings[%d].Position.File = %q, want %q", i, f.Position.File, group.Clones[i].Filename)
		}

		if f.Position.Line != group.Clones[i].LineStart {
			t.Errorf("findings[%d].Position.Line = %d, want %d", i, f.Position.Line, group.Clones[i].LineStart)
		}

		if f.Range == nil || f.Range.End.Line != group.Clones[i].LineEnd {
			t.Errorf("findings[%d].Range must span to LineEnd %d", i, group.Clones[i].LineEnd)
		}

		if f.Snippet != group.Clones[i].Fragment {
			t.Errorf("findings[%d].Snippet = %q, want the clone fragment", i, f.Snippet)
		}
	}
}

func TestToFindingsIDsAreDeterministicForReportDiffing(t *testing.T) {
	group := testGroup("group0123456789a", 10)

	first := finding.ToFindings(group, finding.Options{Version: "test"})
	second := finding.ToFindings(group, finding.Options{Version: "test"})

	for i := range first {
		if first[i].ID != second[i].ID {
			t.Errorf("finding[%d].ID differs across conversions: %q vs %q", i, first[i].ID, second[i].ID)
		}
	}
}

func TestToFindingsMetadataCarriesCloneData(t *testing.T) {
	group := testGroup("group0123456789a", 10)

	findings := finding.ToFindings(group, finding.Options{
		Version:         "test",
		DetectionMethod: "semantic",
	})

	first := findings[0].Metadata

	want := map[string]string{
		finding.MetadataKeyGroupSize:       "2",
		finding.MetadataKeyGroupTokens:     "20",
		finding.MetadataKeyLines:           "10",
		finding.MetadataKeyCloneType:       string(domain.CloneType1),
		finding.MetadataKeyCategory:        string(domain.CategoryFunction),
		finding.MetadataKeyPriority:        string(domain.PriorityMedium),
		finding.MetadataKeyActionability:   string(domain.Actionable),
		finding.MetadataKeyDetectionMethod: "semantic",
	}
	for key, wantValue := range want {
		if first[key] != wantValue {
			t.Errorf("metadata[%q] = %q, want %q", key, first[key], wantValue)
		}
	}

	second := findings[1].Metadata
	if second[finding.MetadataKeyNonActionablePattern] != "signature-only" {
		t.Errorf("metadata[%q] = %q, want %q",
			finding.MetadataKeyNonActionablePattern,
			second[finding.MetadataKeyNonActionablePattern], "signature-only")
	}

	if _, has := first[finding.MetadataKeyNonActionablePattern]; has {
		t.Error("actionable finding must not carry a non-actionable-pattern key")
	}

	if _, has := first[finding.MetadataKeyGenericsCandidate]; has {
		t.Error("finding without generics candidacy must not carry the key")
	}
}

func TestToFindingsGenericsCandidateMetadata(t *testing.T) {
	group := testGroup("group0123456789a", 10)
	group.Clones[0].Classification.GenericsCandidate = true
	group.Clones[0].Classification.GenericsHint = "same algorithm, different types: A vs B"

	findings := finding.ToFindings(group, finding.Options{})
	metadata := findings[0].Metadata

	if metadata[finding.MetadataKeyGenericsCandidate] != "true" {
		t.Errorf("metadata[%q] = %q, want %q",
			finding.MetadataKeyGenericsCandidate,
			metadata[finding.MetadataKeyGenericsCandidate], "true")
	}

	if metadata[finding.MetadataKeyGenericsHint] != "same algorithm, different types: A vs B" {
		t.Errorf("metadata[%q] = %q, want the hint",
			finding.MetadataKeyGenericsHint, metadata[finding.MetadataKeyGenericsHint])
	}
}

func TestToFindingsSeverityMirrorsSARIFLadder(t *testing.T) {
	tests := []struct {
		name      string
		groupSize int
		level     gofinding.Severity
	}{
		{"below first escalation", config.DefaultThreshold, gofinding.SeverityInfo},
		{"at warning escalation", config.DefaultThreshold * 2, gofinding.SeverityWarning},
		{"at error escalation", config.DefaultThreshold * 4, gofinding.SeverityError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := testGroup("group0123456789a", tt.groupSize)
			group := domain.NewProcessedCloneGroup(pair.Hash, pair.Clones[:1])

			findings := finding.ToFindings(group, finding.Options{})
			for i, f := range findings {
				if f.Severity != tt.level {
					t.Errorf("findings[%d].Severity = %q, want %q for size %d", i, f.Severity, tt.level, tt.groupSize)
				}
			}
		})
	}
}

func TestToFindingsZeroThresholdUsesDefault(t *testing.T) {
	group := testGroup("group0123456789a", config.DefaultThreshold*4)

	findings := finding.ToFindings(group, finding.Options{})
	if findings[0].Severity != gofinding.SeverityError {
		t.Errorf(
			"Severity = %q, want %q when size reaches the default threshold escalation",
			findings[0].Severity,
			gofinding.SeverityError,
		)
	}
}

func TestToFindingsConfidenceFromAnalysis(t *testing.T) {
	group := testGroup("group0123456789a", 10)
	group.Clones[0].Classification.Analysis = &domain.ExtractabilityAnalysis{Confidence: 0.9}

	findings := finding.ToFindings(group, finding.Options{})
	if findings[0].Confidence != gofinding.Confidence(0.9) {
		t.Errorf("Confidence = %v, want 0.9 from the extractability analysis", findings[0].Confidence)
	}

	if findings[1].Confidence != gofinding.ConfidenceNone {
		t.Errorf("Confidence = %v, want none when no analysis is present", findings[1].Confidence)
	}
}

func TestToFindingsSuggestionSetsFixStrategy(t *testing.T) {
	group := testGroup("group0123456789a", 10)
	group.Clones[0].Classification.Suggestion = "extract to a shared helper"

	findings := finding.ToFindings(group, finding.Options{})
	if findings[0].FixStrategy != gofinding.FixStrategySuggest {
		t.Errorf(
			"FixStrategy = %q, want %q when a suggestion exists",
			findings[0].FixStrategy,
			gofinding.FixStrategySuggest,
		)
	}

	if findings[0].Suggestion != "extract to a shared helper" {
		t.Errorf("Suggestion = %q, want the classification suggestion", findings[0].Suggestion)
	}

	if findings[1].FixStrategy != gofinding.FixStrategyNone {
		t.Errorf("FixStrategy = %q, want %q without a suggestion", findings[1].FixStrategy, gofinding.FixStrategyNone)
	}
}

func TestToFindingsRelatedLinksAllSiblings(t *testing.T) {
	group := testGroup("group0123456789a", 10)

	findings := finding.ToFindings(group, finding.Options{})
	if len(findings[0].Related) != 1 {
		t.Fatalf("findings[0] has %d related refs, want 1 sibling", len(findings[0].Related))
	}

	rel := findings[0].Related[0]
	if rel.Relation != gofinding.RelationCloneOf {
		t.Errorf("Relation = %q, want %q", rel.Relation, gofinding.RelationCloneOf)
	}

	if rel.FindingID != findings[1].ID {
		t.Errorf("Related FindingID = %q, want the sibling finding ID %q", rel.FindingID, findings[1].ID)
	}

	if rel.Range == nil || rel.Range.End.Line != group.Clones[1].LineEnd {
		t.Error("Related ref must carry the sibling's full range (GAP-1)")
	}

	if len(findings[1].Related) != 1 || findings[1].Related[0].FindingID != findings[0].ID {
		t.Error("findings[1] must link back to findings[0]")
	}

	single := domain.NewProcessedCloneGroup("solo", group.Clones[:1])

	soloFindings := finding.ToFindings(single, finding.Options{})
	if soloFindings[0].Related != nil {
		t.Errorf("single-clone finding has %d related refs, want none", len(soloFindings[0].Related))
	}
}

func TestToReportValidates(t *testing.T) {
	groups := []domain.ProcessedCloneGroup{
		testGroup("group0123456789a", 10),
		testGroup("groupbcdef01234567", 40),
	}

	report := finding.ToReport(groups, finding.Options{Version: "test"})
	if err := report.Validate(); err != nil {
		t.Fatalf("Report.Validate() = %v, want nil", err)
	}
}

// TestReportGroupFindingsReconstructsCloneGroups is issue #1's second
// verification criterion: Report.GroupFindings() on an emitted report
// reconstructs exactly the clone groups art-dupl found.
func TestReportGroupFindingsReconstructsCloneGroups(t *testing.T) {
	groups := []domain.ProcessedCloneGroup{
		testGroup("group0123456789a", 10),
		testGroup("groupbcdef01234567", 40),
	}

	report := finding.ToReport(groups, finding.Options{Version: "test"})

	grouped := report.GroupFindings()
	if len(grouped) != len(groups) {
		t.Fatalf("GroupFindings() = %d groups, want %d", len(grouped), len(groups))
	}

	for _, group := range groups {
		wantID := finding.GroupIDOf(group)

		findings, ok := grouped[wantID]
		if !ok {
			t.Fatalf("GroupFindings() missing group %q", wantID)
		}

		if len(findings) != len(group.Clones) {
			t.Fatalf("group %q reconstructed %d findings, want %d", wantID, len(findings), len(group.Clones))
		}

		wantOccurrences := make([]string, 0, len(group.Clones))
		for _, cl := range group.Clones {
			wantOccurrences = append(wantOccurrences, cl.Filename+":"+strconv.Itoa(cl.LineStart))
		}

		slices.Sort(wantOccurrences)

		gotOccurrences := make([]string, 0, len(findings))
		for _, f := range findings {
			gotOccurrences = append(gotOccurrences, string(f.Position.File)+":"+strconv.Itoa(f.Position.Line))
		}

		slices.Sort(gotOccurrences)

		if !slices.Equal(wantOccurrences, gotOccurrences) {
			t.Errorf("group %q reconstructed occurrences %v, want %v", wantID, gotOccurrences, wantOccurrences)
		}
	}
}

// TestSARIFRoundTripPreservesGroupID is issue #1's first verification
// criterion: every finding of a clone group carries go-finding/groupId with
// the same value, and the property survives the SARIF round-trip.
func TestSARIFRoundTripPreservesGroupID(t *testing.T) {
	groups := []domain.ProcessedCloneGroup{
		testGroup("group0123456789a", 10),
		testGroup("groupbcdef01234567", 40),
	}

	report := finding.ToReport(groups, finding.Options{Version: "test"})

	data, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF() = %v, want nil", err)
	}

	imported, err := gofinding.FindingsFromSARIF(context.Background(), data)
	if err != nil {
		t.Fatalf("FindingsFromSARIF() = %v, want nil", err)
	}

	byGroup := map[gofinding.GroupID][]string{}

	for _, f := range imported {
		if f.GroupID == "" {
			t.Errorf("imported finding %q lost its GroupID", f.ID)

			continue
		}

		byGroup[f.GroupID] = append(byGroup[f.GroupID], string(f.ID))
	}

	if len(byGroup) != len(groups) {
		t.Errorf("SARIF round-trip reconstructed %d groups, want %d", len(byGroup), len(groups))
	}

	for _, group := range groups {
		ids, ok := byGroup[finding.GroupIDOf(group)]
		if !ok {
			t.Errorf("SARIF round-trip lost group %q", finding.GroupIDOf(group))

			continue
		}

		if len(ids) != len(group.Clones) {
			t.Errorf("group %q carries %d findings after round-trip, want %d",
				finding.GroupIDOf(group), len(ids), len(group.Clones))
		}
	}
}

// TestToLSPRoundTripPreservesGroupID is issue #1's third verification
// criterion: finding.ToLSP() round-trip preserves GroupID via Data.GroupID.
func TestToLSPRoundTripPreservesGroupID(t *testing.T) {
	group := testGroup("group0123456789a", 10)

	for i, f := range finding.ToFindings(group, finding.Options{Version: "test"}) {
		diag := f.ToLSP()
		if diag.Data == nil {
			t.Fatalf("findings[%d].ToLSP() has no Data payload", i)
		}

		if diag.Data.GroupID != f.GroupID {
			t.Errorf("findings[%d] ToLSP Data.GroupID = %q, want %q", i, diag.Data.GroupID, f.GroupID)
		}

		back := gofinding.FromLSP(gofinding.FilePath(f.Position.File), diag)
		if back.GroupID != f.GroupID {
			t.Errorf("findings[%d] FromLSP GroupID = %q, want %q", i, back.GroupID, f.GroupID)
		}
	}
}

func TestFindingJSONCarriesGroupID(t *testing.T) {
	group := testGroup("group0123456789a", 10)

	data, err := json.Marshal(finding.ToFindings(group, finding.Options{Version: "test"})[0])
	if err != nil {
		t.Fatalf("json.Marshal() = %v, want nil", err)
	}

	want := `"groupId":"group0123456789a"`
	if !strings.Contains(string(data), want) {
		t.Errorf("finding JSON = %s, want it to contain %s", data, want)
	}
}
