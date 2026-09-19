// Package finding adapts art-dupl clone groups to the go-finding interchange
// model (github.com/larsartmann/go-finding), per the integration evaluation
// verdict: go-finding stays a consumer-side output layer, never an internal
// representation.
//
// The adapter assigns every finding of a clone group the same deterministic
// GroupID (the group's content hash), so consumers can aggregate and diff
// reports without re-deriving grouping from Related links. Clone-specific
// data travels in Finding.Metadata per the GAP-3 deferral (metadata keys use
// the "art-dupl/" namespace; "go-finding/" is reserved for the library).
//
// The three interchange paths are exercised by the tests: SARIF round-trip
// (go-finding/groupId property), Report.GroupFindings() reconstruction, and
// ToLSP/FromLSP (Data.GroupID).
package finding

import (
	"fmt"
	"strconv"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	gofinding "github.com/larsartmann/go-finding"
)

const (
	// ToolName identifies art-dupl as the finding source.
	ToolName = "art-dupl"

	// RuleCloneDetected is the go-finding rule name; it matches the ruleId
	// the SARIF printer emits so both output paths name the same rule.
	RuleCloneDetected = "art-dupl/duplicate-code"

	// SARIFPropGroupID is the reserved go-finding SARIF property key under
	// which a finding's GroupID round-trips (see go-finding sarif_import).
	SARIFPropGroupID = "go-finding/groupId"
)

// Metadata keys carry clone-specific data on each Finding (GAP-3 deferral:
// per-relationship metadata is deliberately not modeled by go-finding).
const (
	MetadataKeyGroupSize            = "art-dupl/group-size"
	MetadataKeyGroupTokens          = "art-dupl/group-tokens" //nolint:gosec // metadata key name, not a credential
	MetadataKeyDetectionMethod      = "art-dupl/detection-method"
	MetadataKeyCloneType            = "art-dupl/clone-type"
	MetadataKeyCategory             = "art-dupl/category"
	MetadataKeyPriority             = "art-dupl/priority"
	MetadataKeyActionability        = "art-dupl/actionability"
	MetadataKeyNonActionablePattern = "art-dupl/non-actionable-pattern"
	MetadataKeyGenericsCandidate    = "art-dupl/generics-candidate"
	MetadataKeyGenericsHint         = "art-dupl/generics-hint"
	MetadataKeyLines                = "art-dupl/lines"
)

// Options tunes the conversion. The zero value is usable: the threshold
// falls back to config.DefaultThreshold and unknown optional values are
// simply omitted from the output.
type Options struct {
	// Version is reported as the go-finding ToolInfo version.
	Version string

	// Threshold drives the severity mapping (identical to the SARIF
	// printer's level ladder). Values <= 0 mean config.DefaultThreshold.
	Threshold int

	// DetectionMethod, when non-empty, is attached to every finding as
	// MetadataKeyDetectionMethod.
	DetectionMethod string
}

// GroupIDOf returns the deterministic go-finding group id for a clone group:
// the group's content hash (16-char lowercase hex, machine-safe). Identical
// input always yields the identical id, so consumers can diff two reports.
// An empty group hash yields the empty GroupID, go-finding's "not grouped".
func GroupIDOf(group domain.ProcessedCloneGroup) gofinding.GroupID {
	return gofinding.GroupID(group.Hash)
}

// ToFindings converts one clone group into one go-finding Finding per clone
// occurrence. All findings share the group's GroupID; each finding points at
// one clone occurrence with its full source range, snippet, and classification
// metadata, and links its siblings via RelationCloneOf.
func ToFindings(group domain.ProcessedCloneGroup, opts Options) []gofinding.Finding {
	ids := findingIDs(group)

	findings := make([]gofinding.Finding, 0, len(group.Clones))
	for i, cl := range group.Clones {
		findings = append(findings, toFinding(group, cl, ids, i, opts))
	}

	return findings
}

// ToReport converts clone groups into a go-finding Report ready for
// ToSARIF/ToLSP consumption or JSON interchange.
func ToReport(groups []domain.ProcessedCloneGroup, opts Options) *gofinding.Report {
	findings := make([]gofinding.Finding, 0, len(groups))
	for _, group := range groups {
		findings = append(findings, ToFindings(group, opts)...)
	}

	return gofinding.NewReportFromFindings(
		gofinding.ToolInfo{Name: ToolName, Version: opts.Version},
		findings,
	)
}

// findingIDs precomputes the stable ID of every clone occurrence so Related
// links can reference sibling IDs without quadratic regeneration.
func findingIDs(group domain.ProcessedCloneGroup) []gofinding.ID {
	ids := make([]gofinding.ID, 0, len(group.Clones))
	for _, cl := range group.Clones {
		ids = append(ids, gofinding.GenerateID(
			gofinding.ToolName(ToolName),
			gofinding.RuleName(RuleCloneDetected),
			positionOf(cl),
		))
	}

	return ids
}

func toFinding(
	group domain.ProcessedCloneGroup,
	cl domain.ProcessedClone,
	ids []gofinding.ID,
	index int,
	opts Options,
) gofinding.Finding {
	size := group.TotalTokenCount()

	threshold := opts.Threshold
	if threshold <= 0 {
		threshold = config.DefaultThreshold
	}

	f := gofinding.NewFinding(
		gofinding.RuleName(RuleCloneDetected),
		gofinding.ToolName(ToolName),
		fmt.Sprintf("Duplicate code: %d tokens in %d instances", size, len(group.Clones)),
		severityFor(size, threshold),
		positionOf(cl),
		confidenceOf(cl),
	)

	f.Category = gofinding.CategoryDuplication
	f.GroupID = GroupIDOf(group)
	f.Range = rangeOf(cl)
	f.Snippet = cl.Fragment
	f.Metadata = metadataFor(group, cl, opts)
	f.Related = relatedOf(group, ids, index)

	if cl.Classification.Suggestion != "" {
		f.FixStrategy = gofinding.FixStrategySuggest
		f.Suggestion = cl.Classification.Suggestion
	}

	return f
}

// severityFor mirrors the SARIF printer's level ladder so both output paths
// report the same severity for the same group.
func severityFor(size, threshold int) gofinding.Severity {
	switch {
	case size >= threshold*4:
		return gofinding.SeverityError
	case size >= threshold*2:
		return gofinding.SeverityWarning
	default:
		return gofinding.SeverityInfo
	}
}

func confidenceOf(cl domain.ProcessedClone) gofinding.Confidence {
	if cl.Classification.Analysis == nil {
		return gofinding.ConfidenceNone
	}

	return gofinding.Confidence(cl.Classification.Analysis.Confidence).Clamp()
}

func positionOf(cl domain.ProcessedClone) gofinding.Position {
	return gofinding.Position{
		File:   gofinding.FilePath(cl.Filename),
		Line:   cl.LineStart,
		Offset: int(cl.StartPos),
	}
}

func rangeOf(cl domain.ProcessedClone) *gofinding.Range {
	return &gofinding.Range{
		Start: positionOf(cl),
		End: gofinding.Position{
			File:   gofinding.FilePath(cl.Filename),
			Line:   cl.LineEnd,
			Offset: int(cl.EndPos),
		},
	}
}

func metadataFor(group domain.ProcessedCloneGroup, cl domain.ProcessedClone, opts Options) map[string]string {
	metadata := map[string]string{
		MetadataKeyGroupSize:     strconv.Itoa(len(group.Clones)),
		MetadataKeyGroupTokens:   strconv.Itoa(group.TotalTokenCount()),
		MetadataKeyLines:         strconv.Itoa(cl.LineCount()),
		MetadataKeyCloneType:     string(cl.Classification.CloneType),
		MetadataKeyCategory:      string(cl.Classification.Category),
		MetadataKeyPriority:      string(cl.Classification.Priority),
		MetadataKeyActionability: string(cl.Classification.Actionability),
	}

	if opts.DetectionMethod != "" {
		metadata[MetadataKeyDetectionMethod] = opts.DetectionMethod
	}

	if cl.Classification.NonActionablePattern != "" {
		metadata[MetadataKeyNonActionablePattern] = cl.Classification.NonActionablePattern
	}

	if cl.Classification.GenericsCandidate {
		metadata[MetadataKeyGenericsCandidate] = "true"
		if cl.Classification.GenericsHint != "" {
			metadata[MetadataKeyGenericsHint] = cl.Classification.GenericsHint
		}
	}

	return metadata
}

func relatedOf(
	group domain.ProcessedCloneGroup,
	ids []gofinding.ID,
	index int,
) []gofinding.RelatedRef {
	related := make([]gofinding.RelatedRef, 0, len(group.Clones)-1)
	for i, other := range group.Clones {
		if i == index {
			continue
		}

		related = append(related, gofinding.RelatedRef{
			FindingID: ids[i],
			Relation:  gofinding.RelationCloneOf,
			Position:  positionOf(other),
			Range:     rangeOf(other),
		})
	}

	if len(related) == 0 {
		return nil
	}

	return related
}
