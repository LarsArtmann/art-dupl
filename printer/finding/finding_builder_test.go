package finding_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/jsonutil"
	"github.com/LarsArtmann/art-dupl/printer/finding"
	gofinding "github.com/larsartmann/go-finding"
)

// legacyToFinding replicates the pre-Builder construction path (unvalidated
// NewFinding + direct field assignment) byte for byte. TestBuilderParity pins
// the Builder conversion to it: if a go-finding upgrade ever changes Builder
// semantics (ID generation, defaults, normalization), this test fails before
// the wire format can drift.
func legacyToFinding(
	group domain.ProcessedCloneGroup,
	cl domain.ProcessedClone,
	ids []gofinding.ID,
	index int,
	opts finding.Options,
) gofinding.Finding {
	size := group.TotalTokenCount()

	threshold := opts.Threshold
	if threshold <= 0 {
		threshold = config.DefaultThreshold
	}

	f := gofinding.NewFinding(
		gofinding.RuleName(finding.RuleCloneDetected),
		gofinding.ToolName(finding.ToolName),
		"Duplicate code: "+strconv.Itoa(size)+" tokens in "+strconv.Itoa(len(group.Clones))+" instances",
		severityForLegacy(size, threshold),
		positionOfLegacy(cl),
		confidenceOfLegacy(cl),
	)

	f.Category = gofinding.CategoryDuplication
	f.GroupID = finding.GroupIDOf(group)
	f.Tags = []gofinding.Tag{gofinding.Tag(gofinding.CategoryDuplication)}
	if t := cl.Classification.CloneType; t != "" {
		f.Tags = append(f.Tags, gofinding.Tag(t))
	}
	f.Range = rangeOfLegacy(cl)
	f.Snippet = cl.Fragment
	f.Metadata = metadataOfLegacy(group, cl, opts)
	f.Related = relatedOfLegacy(group, ids, index)

	if cl.Classification.Suggestion != "" {
		f.FixStrategy = gofinding.FixStrategySuggest
		f.Suggestion = cl.Classification.Suggestion
	}

	return f
}

func severityForLegacy(size, threshold int) gofinding.Severity {
	switch {
	case size >= threshold*4:
		return gofinding.SeverityError
	case size >= threshold*2:
		return gofinding.SeverityWarning
	default:
		return gofinding.SeverityInfo
	}
}

func confidenceOfLegacy(cl domain.ProcessedClone) gofinding.Confidence {
	if cl.Classification.Analysis == nil {
		return gofinding.ConfidenceNone
	}

	return gofinding.Confidence(cl.Classification.Analysis.Confidence).Clamp()
}

func positionOfLegacy(cl domain.ProcessedClone) gofinding.Position {
	return gofinding.Position{
		File:   gofinding.FilePath(cl.Filename),
		Line:   cl.LineStart,
		Offset: int(cl.StartPos),
	}
}

func rangeOfLegacy(cl domain.ProcessedClone) *gofinding.Range {
	return &gofinding.Range{
		Start: positionOfLegacy(cl),
		End: gofinding.Position{
			File:   gofinding.FilePath(cl.Filename),
			Line:   cl.LineEnd,
			Offset: int(cl.EndPos),
		},
	}
}

func metadataOfLegacy(group domain.ProcessedCloneGroup, cl domain.ProcessedClone, opts finding.Options) map[string]string {
	metadata := map[string]string{
		finding.MetadataKeyGroupSize:     strconv.Itoa(len(group.Clones)),
		finding.MetadataKeyGroupTokens:   strconv.Itoa(group.TotalTokenCount()),
		finding.MetadataKeyLines:         strconv.Itoa(cl.LineCount()),
		finding.MetadataKeyCloneType:     string(cl.Classification.CloneType),
		finding.MetadataKeyCategory:      string(cl.Classification.Category),
		finding.MetadataKeyPriority:      string(cl.Classification.Priority),
		finding.MetadataKeyActionability: string(cl.Classification.Actionability),
	}

	if opts.DetectionMethod != "" {
		metadata[finding.MetadataKeyDetectionMethod] = opts.DetectionMethod
	}

	if cl.Classification.NonActionablePattern != "" {
		metadata[finding.MetadataKeyNonActionablePattern] = cl.Classification.NonActionablePattern
	}

	if cl.Classification.GenericsCandidate {
		metadata[finding.MetadataKeyGenericsCandidate] = "true"
		if cl.Classification.GenericsHint != "" {
			metadata[finding.MetadataKeyGenericsHint] = cl.Classification.GenericsHint
		}
	}

	return metadata
}

func relatedOfLegacy(group domain.ProcessedCloneGroup, ids []gofinding.ID, index int) []gofinding.RelatedRef {
	refs := make([]gofinding.RelatedRef, 0, len(group.Clones)-1)

	for i, cl := range group.Clones {
		if i == index {
			continue
		}

		refs = append(refs, gofinding.RelatedRef{
			FindingID: ids[i],
			Relation:  gofinding.RelationCloneOf,
			Position:  positionOfLegacy(cl),
			Range:     rangeOfLegacy(cl),
		})
	}

	return refs
}

// TestBuilderConversionIsByteIdenticalToLegacyConstruction marshals both
// construction paths over fixture groups and requires identical bytes: the
// Builder conversion must be a pure validation upgrade, never a wire change.
func TestBuilderConversionIsByteIdenticalToLegacyConstruction(t *testing.T) {
	t.Parallel()

	groups := []domain.ProcessedCloneGroup{
		testGroup("abcdef0123456789", 10),
		testGroup("ffffffffffffffff", 4, 12),
		testGroup("", 40),
		testGroupWithSuggestion(),
	}

	for _, group := range groups {
		opts := finding.Options{Version: "test", Threshold: 5, DetectionMethod: "semantic"}

		built := finding.ToFindings(group, opts)

		ids := make([]gofinding.ID, 0, len(group.Clones))
		for _, cl := range group.Clones {
			ids = append(ids, gofinding.GenerateID(
				gofinding.ToolName(finding.ToolName),
				gofinding.RuleName(finding.RuleCloneDetected),
				gofinding.Position{
					File:   gofinding.FilePath(cl.Filename),
					Line:   cl.LineStart,
					Offset: int(cl.StartPos),
				},
			))
		}

		legacy := make([]gofinding.Finding, 0, len(group.Clones))
		for i, cl := range group.Clones {
			legacy = append(legacy, legacyToFinding(group, cl, ids, i, opts))
		}

		builtJSON, err := jsonutil.MarshalIndent(built, "", "  ")
		if err != nil {
			t.Fatalf("marshal built: %v", err)
		}

		legacyJSON, err := jsonutil.MarshalIndent(legacy, "", "  ")
		if err != nil {
			t.Fatalf("marshal legacy: %v", err)
		}

		if string(builtJSON) != string(legacyJSON) {
			t.Fatalf("Builder output diverged from legacy construction for group %q:\nbuilt:\n%s\nlegacy:\n%s",
				group.Hash, builtJSON, legacyJSON)
		}
	}
}

// testGroupWithSuggestion returns a group whose clones carry an extraction
// suggestion, exercising the FixStrategy/Suggestion branch of both paths.
func testGroupWithSuggestion() domain.ProcessedCloneGroup {
	group := testGroup("suggestion0000001", 25)
	for i := range group.Clones {
		group.Clones[i].Classification.Suggestion = "Extract the shared accumulator into a helper"
	}

	return group
}

// TestBuilderRejectsMalformedConstruction pins the validation boundary the
// adapter now relies on: Build must fail (and MustBuild panic) on the
// malformed classes go-finding's validators exist to catch. Our domain data
// can never produce these; the test documents why MustBuild is safe.
func TestBuilderRejectsMalformedConstruction(t *testing.T) {
	t.Parallel()

	pos := gofinding.Position{File: "a.go", Line: 1, Offset: 0}

	t.Run("empty rule", func(t *testing.T) {
		t.Parallel()

		if _, err := gofinding.NewBuilder("", "tool", "msg", gofinding.SeverityInfo, pos).Build(); err == nil {
			t.Fatal("empty rule must fail Build")
		} else if !strings.Contains(err.Error(), "Rule") {
			t.Fatalf("empty rule error should name Rule, got: %v", err)
		}
	})

	t.Run("invalid tag", func(t *testing.T) {
		t.Parallel()

		_, err := gofinding.NewBuilder("rule", "tool", "msg", gofinding.SeverityInfo, pos).
			WithTags(gofinding.Tag("Bad_Tag")).
			Build()
		if err == nil {
			t.Fatal("non-lowercase-hyphenated tag must fail Build")
		} else if !strings.Contains(err.Error(), "Tags[0]") {
			t.Fatalf("invalid tag error should name Tags[0], got: %v", err)
		}
	})

	t.Run("inverted range", func(t *testing.T) {
		t.Parallel()

		_, err := gofinding.NewBuilder("rule", "tool", "msg", gofinding.SeverityInfo, pos).
			WithRange(gofinding.Range{
				Start: gofinding.Position{File: "a.go", Line: 10, Offset: 100},
				End:   gofinding.Position{File: "a.go", Line: 1, Offset: 0},
			}).
			Build()
		if err == nil {
			t.Fatal("inverted range must fail Build")
		}
	})

	t.Run("adapter output passes validation", func(t *testing.T) {
		t.Parallel()

		for _, f := range finding.ToFindings(testGroup("valid1234567890", 10), finding.Options{Threshold: 5}) {
			if err := f.Validate(); err != nil {
				t.Fatalf("adapter finding must validate: %v", err)
			}
		}
	})
}
