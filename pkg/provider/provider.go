// Package provider self-registers art-dupl as a BuildFlow tool provider via
// the go-finding toolsdk (github.com/larsartmann/go-finding/toolsdk).
//
// The registration is a package-level var so it runs at init time: BuildFlow
// blank-imports this package (tools/providers/sdk_imports.go pattern) and
// discovers the spec via toolsdk.All() — zero BuildFlow-side glue. This
// mirrors the go-structure-linter pkg/provider integration (B1).
//
// Detect runs the public SDK (pkg/artdupl) in semantic mode at the default
// threshold and converts groups to go-finding Findings through the
// printer/finding adapter, so GroupID, severity ladder, positions, snippets,
// and Related links come from the same adapter code the CLI's finding output
// uses. The SDK pipeline does not run actionability or generics analysis
// (CLI-only), so Classification is zero-valued: after stripEmptyMetadata the
// classification metadata keys (art-dupl/clone-type, -category, -priority,
// -actionability, -generics-*) are absent from these findings; severity
// gating in BuildFlow works off the threshold ladder regardless.
//
// File discovery mirrors the CLI's default exclusions: only .go/.templ files,
// no dot-directories, vendor/, node_modules/, examples/, demo/, demos/, and
// generated files (gogenfilter, all categories). .gitignore patterns are NOT
// honored here (the matcher lives in the cmd layer) — the directory skips
// above cover the dominant noise sources.
package provider

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/pkg/artdupl"
	"github.com/LarsArtmann/art-dupl/printer/finding"
	gofinding "github.com/larsartmann/go-finding"
	toolsdk "github.com/larsartmann/go-finding/toolsdk"
)

// modulePath is this repository's Go module path, used to resolve the
// provider version from build info when compiled as a dependency (the
// BuildFlow consumption mode).
const modulePath = "github.com/LarsArtmann/art-dupl"

// versionFallback matches the SDK's dev-version convention.
const versionFallback = "dev"

// semanticMode labels the SDK's default detection mode in finding metadata.
const semanticMode = "semantic"

// maxAdvisorySeverity is the highest severity the SDK interchange emits.
// The finding interchange feeds pipeline gates that key on error-or-above
// (BuildFlow's default findings gate); a detector-only tool with no Repairer
// must never produce gate-failing findings, or repos with large clone groups
// would fail every run with no automated fix path. The SARIF ladder's
// escalation survives in a tag; the CLI's own output paths are untouched.
const maxAdvisorySeverity = gofinding.SeverityWarning

// originalSeverityTagPrefix marks downgraded findings with the pre-cap
// severity. Hyphenated on purpose: a colonated tag fails report validation,
// which breaks consumers' `--format finding` output (found via PapDashboard
// 2026-09-22 in branching-flow, the same pattern mirrored here).
const originalSeverityTagPrefix = "original-severity-"

// Provider is the registered toolsdk spec. The var initializer performs the
// registration; keeping it as a package-level var (per the toolsdk contract)
// makes the blank import in BuildFlow the entire wiring step.
//
//nolint:gochecknoglobals // toolsdk self-registration by design (see package doc)
var Provider = toolsdk.Register(toolsdk.Spec{
	Name: finding.ToolName,
	Description: "Code duplication detection: suffix-tree + AST-hash clones " +
		"(Type 1 exact, Type 2 renamed, Type 3 near-miss) across Go and templ files",
	Trigger: toolsdk.OnFiles("go", "**/*.go", "**/*.templ"),
	Inputs:  []string{"**/*.go", "**/*.templ"},
	Detect:  cloneDetector{},
})

// cloneDetector implements the go-finding Detector contract on top of the
// art-dupl SDK.
type cloneDetector struct{}

// Name implements gofinding.Detector.
func (cloneDetector) Name() string { return finding.ToolName }

// Detect implements gofinding.Detector: analyze the working directory from
// the context (default ".") and return one finding per clone occurrence.
func (cloneDetector) Detect(ctx context.Context) ([]gofinding.Finding, error) {
	ctx = toolsdk.EnsureContext(ctx)

	dir := gofinding.WorkingDirFromContext(ctx)
	if dir == "" {
		dir = "."
	}

	files, err := collectSourceFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("collect source files under %s: %w", dir, err)
	}

	if len(files) == 0 {
		return []gofinding.Finding{}, nil
	}

	detector, err := artdupl.NewDetector(providerOptions())
	if err != nil {
		return nil, fmt.Errorf("create art-dupl detector: %w", err)
	}

	defer func() { _ = detector.Close() }()

	result, err := detector.FindClones(ctx, files)
	if err != nil {
		// A clone-free repo is a successful Detect run, not an error: the
		// SDK surfaces zero groups as a sentinel.
		if errors.Is(err, artdupl.ErrNoDuplicatesFound) {
			return []gofinding.Finding{}, nil
		}

		return nil, fmt.Errorf("detect clones in %s: %w", dir, err)
	}

	return findingsFromGroups(result.CloneGroups), nil
}

// providerOptions configures the SDK for finding interchange: snippets are
// required (they become Finding.Snippet) and per-group occurrence caps are
// lifted (BuildFlow gates per finding, so truncation would hide instances).
func providerOptions() *artdupl.Options {
	opts := artdupl.DefaultOptions()
	opts.IncludeFragments = true
	opts.MaxClonesPerGroup = 0

	return opts
}

// findingsFromGroups converts SDK clone groups into go-finding findings via
// the shared printer/finding adapter, keeping IDs, GroupIDs, and positions
// identical to the CLI's finding output for the same group input. Severity
// is advisory-capped (see capAdvisorySeverities). Classification metadata
// keys are intentionally absent: the SDK pipeline never computes them (see
// the package doc).
func findingsFromGroups(groups []*artdupl.CloneGroup) []gofinding.Finding {
	opts := finding.Options{
		Version:         providerVersion(),
		Threshold:       artdupl.DefaultThreshold,
		DetectionMethod: semanticMode,
	}

	findings := make([]gofinding.Finding, 0, len(groups))
	for _, group := range groups {
		findings = append(findings, finding.ToFindings(toProcessedGroup(group), opts)...)
	}

	capAdvisorySeverities(findings)
	stripEmptyMetadata(findings)

	return findings
}

// capAdvisorySeverities downgrades error findings to warning, appending an
// original-severity tag. Mutates the slice in place. The shared adapter's
// ladder mirrors the SARIF printer's error escalation, but the interchange
// consumer gates on error-or-above over findings nothing can auto-fix
// (detector-only provider), so the SDK surface tops out at warning.
func capAdvisorySeverities(findings []gofinding.Finding) {
	for i := range findings {
		if findings[i].Severity != gofinding.SeverityError && findings[i].Severity != gofinding.SeverityCritical {
			continue
		}

		findings[i].Tags = append(
			findings[i].Tags,
			gofinding.Tag(originalSeverityTagPrefix+findings[i].Severity.String()),
		)
		findings[i].Severity = maxAdvisorySeverity
	}
}

// stripEmptyMetadata drops empty-valued metadata entries. The shared adapter
// emits classification keys unconditionally; the SDK pipeline (unlike the
// CLI) has no classification, so those keys would all be empty strings.
func stripEmptyMetadata(findings []gofinding.Finding) {
	for i := range findings {
		for key, value := range findings[i].Metadata {
			if value == "" {
				delete(findings[i].Metadata, key)
			}
		}
	}
}

// toProcessedGroup maps an SDK CloneGroup onto the domain group shape the
// finding adapter consumes. Classification stays zero-valued: the SDK
// pipeline (unlike the CLI) does not run actionability analysis.
func toProcessedGroup(group *artdupl.CloneGroup) domain.ProcessedCloneGroup {
	clones := make([]domain.ProcessedClone, 0, len(group.Clones))
	for _, cl := range group.Clones {
		clones = append(clones, domain.ProcessedClone{
			CloneRef:   cl.CloneRef,
			StartPos:   int32(cl.StartPos),
			EndPos:     int32(cl.EndPos),
			TokenCount: cl.Size,
		})
	}

	return domain.NewProcessedCloneGroup(group.Hash, clones)
}

// providerVersion resolves the art-dupl version from build info. When
// compiled as a dependency (BuildFlow's consumption mode) the main-module
// version belongs to the consumer, so the dependency table is the source of
// truth there.
func providerVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return versionFallback
	}

	if info.Main.Path == modulePath && isReleaseVersion(info.Main.Version) {
		return info.Main.Version
	}

	for _, dep := range info.Deps {
		if dep.Path == modulePath && dep.Version != "" {
			return dep.Version
		}
	}

	return versionFallback
}

// isReleaseVersion filters the placeholder versions build info reports for
// local development builds.
func isReleaseVersion(version string) bool {
	return version != "" && version != "(devel)"
}
