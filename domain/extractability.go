package domain

// Extractability assesses whether a clone group can be cleanly extracted into a
// shared function, turning clone reports into actionable refactoring advice.
type Extractability struct {
	// CanExtract is true when the clone spans a complete syntactic unit (a
	// function body or a statement block) and duplicating it to a shared helper
	// would reduce total lines.
	CanExtract bool
	// EstimatedLinesSaved is the net line reduction from extracting: the lines
	// currently duplicated across all instances minus the one shared copy plus
	// its function signature overhead. Negative or zero means extraction would
	// not reduce line count.
	EstimatedLinesSaved int
	// Reason explains the verdict in one sentence for display in reports.
	Reason string
}

// extractOverhead is the fixed line cost of a new helper function (signature +
// braces + return), subtracted from the savings estimate.
const extractOverhead = 4

// AssessExtractability computes an Extractability verdict from clone metadata.
//
//	lines       — lines spanned by one clone instance
//	instances   — number of duplicate sites
//	isComplete  — true if the clone root is a complete function or block
func AssessExtractability(lines, instances int, isComplete bool) Extractability {
	if instances < 2 {
		return Extractability{
			CanExtract:          false,
			EstimatedLinesSaved: 0,
			Reason:              "single instance — nothing to deduplicate",
		}
	}

	if !isComplete {
		return Extractability{
			CanExtract:          false,
			EstimatedLinesSaved: lines * (instances - 1),
			Reason:              "partial fragment — extract requires manual boundary selection",
		}
	}

	saved := lines*(instances-1) - extractOverhead

	return Extractability{
		CanExtract:          saved > 0,
		EstimatedLinesSaved: saved,
		Reason:              extractHint(saved),
	}
}

func extractHint(saved int) string {
	if saved <= 0 {
		return "clone too small to benefit from extraction"
	}

	return "Extract to shared helper — saves lines across sites"
}

// IsCompleteUnit reports whether a clone classification category represents a
// complete, extractable syntactic unit (function/method/block) rather than a
// fragment or idiom.
func (c CloneCategory) IsCompleteUnit() bool {
	switch c { //nolint:exhaustive // default covers all non-extractable categories
	case CategoryFunction, CategoryMethod, CategoryLoop, CategoryConditional:
		return true
	default:
		return false
	}
}
