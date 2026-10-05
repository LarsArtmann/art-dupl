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

	// EmitSuppressedAccepted opts into surfacing reviewed duplicates: when
	// true, groups for which Accepted reports true are emitted with an
	// in-source Suppression annotation on every finding instead of relying
	// on the caller to drop them. Consumers decide whether suppressed
	// findings count (go-finding drops them from SARIF by default; opt in
	// via WithIncludeSuppressed). The zero value (false) never consults
	// Accepted, so the default output is byte-identical to the historical
	// behavior.
	EmitSuppressedAccepted bool

	// Accepted reports whether a clone group is covered by an in-source
	// //art-dupl:accept directive. It is the CLI's AcceptedSet.IsAccepted
	// (internal/accept) at the call site; declared as a plain predicate so
	// this adapter stays decoupled from cmd and internal/accept. Only
	// consulted when EmitSuppressedAccepted is true; nil accepts nothing.
	Accepted func(domain.ProcessedCloneGroup) bool
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
	tmpl := gofinding.NewTemplate(gofinding.ToolName(ToolName)).
		WithCategory(gofinding.CategoryDuplication)

	ids := findingIDs(group)

	accepted := opts.EmitSuppressedAccepted &&
		opts.Accepted != nil && opts.Accepted(group)

	findings := make([]gofinding.Finding, 0, len(group.Clones))
	for i, cl := range group.Clones {
		findings = append(findings, toFinding(tmpl, group, cl, ids, i, accepted, opts))
	}

	return findings
}

// ToReport converts clone groups into a go-finding Report ready for
// ToSARIF/ToLSP consumption or JSON interchange.
//
// The CLI pipeline does NOT call this — it streams groups to its printers
// (constant memory) — so ToReport exists for BATCH consumers (SDK callers,
// tests): it is the interchange entry point whose output is byte-equivalent
// to what the CLI's SARIF printer emits (pinned by the SARIF cross-check
// test) and whose suppression/tag semantics round-trip through go-finding.
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
	tmpl *gofinding.Template,
	group domain.ProcessedCloneGroup,
	cl domain.ProcessedClone,
	ids []gofinding.ID,
	index int,
	accepted bool,
	opts Options,
) gofinding.Finding {
	size := group.TotalTokenCount()

	threshold := opts.Threshold
	if threshold <= 0 {
		threshold = config.DefaultThreshold
	}

	b := tmpl.Builder(
		gofinding.RuleName(RuleCloneDetected),
		fmt.Sprintf("Duplicate code: %d tokens in %d instances", size, len(group.Clones)),
		severityFor(size, threshold),
		positionOf(cl),
	).
		WithGroupID(GroupIDOf(group)).
		WithTags(cloneTags(cl)...).
		WithConfidence(confidenceOf(cl)).
		WithRange(*rangeOf(cl)).
		WithSnippet(cl.Fragment).
		WithMetadata(metadataFor(group, cl, opts)).
		WithRelated(relatedOf(group, ids, index)...)

	if cl.Classification.Suggestion != "" {
		b = b.WithFixStrategy(gofinding.FixStrategySuggest).
			WithSuggestion(cl.Classification.Suggestion)
	}

	if accepted {
		b = b.WithSuppression(gofinding.Suppression{
			Kind:   gofinding.SuppressionInSource,
			Rule:   gofinding.RuleName(RuleCloneDetected),
			Reason: "//art-dupl:accept directive",
		})
	}

	// MustBuild, not BuildOrDefault: every input is derived from validated
	// domain data, so a construction-time validation failure is a programmer
	// error that must fail loudly (the wire goldens pin the bytes) rather than
	// silently emit a malformed finding.
	return b.MustBuild()
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

// cloneTags returns the semantic tags for a clone occurrence: the tool's
// standard category (which also satisfies go-finding's category-in-tags
// consistency rule when tags are set) plus, when the pipeline classified the
// clone, its Bellon taxonomy type (type-1/2/3 — the domain enum's string
// values already follow the validated lowercase-hyphenated tag convention).
// Unclassified findings (the SDK/provider path) still get the category tag.
func cloneTags(cl domain.ProcessedClone) []gofinding.Tag {
	tags := []gofinding.Tag{gofinding.Tag(gofinding.CategoryDuplication)}
	if t := cl.Classification.CloneType; t != "" {
		tags = append(tags, gofinding.Tag(t))
	}

	return tags
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
		Column: int(cl.ColumnStart), // 0 = unknown (no source at extraction)
		Offset: int(cl.StartPos),
	}
}

func rangeOf(cl domain.ProcessedClone) *gofinding.Range {
	return &gofinding.Range{
		Start: positionOf(cl),
		End: gofinding.Position{
			File:   gofinding.FilePath(cl.Filename),
			Line:   cl.LineEnd,
			Column: int(cl.ColumnEnd), // exclusive end; 0 = unknown
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
