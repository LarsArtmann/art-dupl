package domain

// ExtractabilityAnalysis evaluates a clone group against 4 computable properties
// that define "harmful duplication." This replaces the pattern-denylist approach
// with property-based classification powered by go/types data.
//
// Properties default to true (harmful) when type info is unavailable or uncertain,
// preventing false negatives in syntax-only mode. Only suppress a clone when there
// is STRONG evidence it is non-harmful.
type ExtractabilityAnalysis struct {
	// MechanicallyExtractable: always true — any code can be wrapped in a function
	// and free variables captured as parameters. This property exists for completeness
	// but is always true by construction.
	MechanicallyExtractable bool

	// ControlFlowExtractable: false when the clone contains return/break/continue
	// statements that are forced by the enclosing function's signature. Example:
	// HTTP handler guards (if err != nil { writeError(); return }) cannot be extracted
	// because the helper cannot issue return on behalf of the void-return caller.
	ControlFlowExtractable bool

	// ROIPositive: false when extraction would not save tokens — the clone is too
	// small, or is dominated by a single call expression (the call IS the extraction).
	// Example: defer cancel(), helper invocations, single-statement wrappers.
	ROIPositive bool

	// Parameterizable: false when the only differences between clone instances are
	// in string-literal values that represent domain data. The clones ARE already
	// parameterized — the tool detected the structural match, but the variation is
	// intentional. Example: bool-to-string funcs, fmt.Sprintf format strings.
	Parameterizable bool

	// Confidence ranges from 0.0 to 1.0. Low confidence means the property analysis
	// lacked type information or encountered an ambiguous case.
	// Tiers: >= 0.8 Actionable, 0.5-0.8 LowConfidence, < 0.5 NonActionable.
	Confidence float64

	// Reason explains the verdict in one sentence for --explain output.
	Reason string
}

// DefaultExtractabilityAnalysis returns an analysis where all properties are true
// (harmful) with moderate confidence. This is the safe default when type info is
// unavailable — it prevents false negatives by treating unknown clones as harmful.
func DefaultExtractabilityAnalysis() ExtractabilityAnalysis {
	return ExtractabilityAnalysis{
		MechanicallyExtractable: true,
		ControlFlowExtractable:  true,
		ROIPositive:             true,
		Parameterizable:         true,
		Confidence:              0.5, //nolint:mnd // conservative default confidence
		Reason:                  "no type info available — treating as potentially harmful",
	}
}

// IsHarmful returns true when ALL 4 properties pass AND confidence >= 0.5.
// A clone is "harmful" when it represents real duplication that should be extracted.
func IsHarmful(e ExtractabilityAnalysis) bool {
	return e.MechanicallyExtractable &&
		e.ControlFlowExtractable &&
		e.ROIPositive &&
		e.Parameterizable &&
		e.Confidence >= 0.5
}

// ActionabilityTier converts a confidence score to a CloneActionability tier.
func ActionabilityTier(confidence float64) CloneActionability {
	if confidence >= 0.8 { //nolint:mnd // actionable threshold
		return Actionable
	}

	if confidence >= 0.5 { //nolint:mnd // low-confidence threshold
		return LowConfidence
	}

	return NonActionable
}

// ExtractabilityChecker evaluates clone sequences and produces an analysis.
// Each checker focuses on one property of the extractability model.
type ExtractabilityChecker interface {
	// Check evaluates the given clone sequences and returns a partial analysis.
	// The caller merges results from multiple checkers.
	Check(nodeSeqs [][]*CloneNode) ExtractabilityAnalysis

	// Name returns the checker's human-readable name for debugging.
	Name() string
}
