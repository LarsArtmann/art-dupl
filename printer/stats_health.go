package printer

import "fmt"

// calculateHealthScore calculates an A-F grade based on duplication, complexity, and impact metrics.
//
// Scoring philosophy:
// - Duplication ratio is the primary indicator (0-5% = A, 5-10% = B, etc.)
// - Complexity and impact are secondary modifiers
// - A project with <5% duplication should never score below C
//
// Grade thresholds (based on totalScore):
// - A: < 5 (excellent - minimal duplication)
// - B: < 10 (good - acceptable duplication)
// - C: < 15 (fair - needs attention)
// - D: < 25 (poor - significant cleanup needed)
// - F: >= 25 (critical - major refactoring required).
func (p *stats) calculateHealthScore() string {
	if p.statsData.DuplicationRatio == 0 && p.statsData.ComplexityScore == 0 &&
		p.statsData.ImpactScore == 0 {
		return "A"
	}

	// Duplication ratio is the primary score (direct percentage)
	duplicationScore := p.statsData.DuplicationRatio

	// Complexity score: normalize to 0-20 scale (complexity 10 = 20 points)
	// This ensures complexity doesn't dominate the score
	complexityScore := (p.statsData.ComplexityScore / 10.0) * 20
	if complexityScore > 20 {
		complexityScore = 20
	}

	// Impact score: normalize to 0-10 scale (impact 10000 = 10 points)
	// Impact is the least important factor
	impactScore := (float64(p.statsData.ImpactScore) / 10000.0) * 10
	if impactScore > 10 {
		impactScore = 10
	}

	// Weighted average: duplication 70%, complexity 20%, impact 10%
	// Duplication is the primary health indicator
	totalScore := duplicationScore*0.7 + complexityScore*0.2 + impactScore*0.1

	// Convert to A-F grade based on total score
	switch {
	case totalScore < 5:
		return "A"
	case totalScore < 10:
		return "B"
	case totalScore < 15:
		return "C"
	case totalScore < 25:
		return "D"
	default:
		return "F"
	}
}

// getSizeRange returns a human-readable size range for a line count.
func (p *stats) getSizeRange(lines int) string {
	switch {
	case lines <= 5:
		return "1-5 lines"
	case lines <= 10:
		return "6-10 lines"
	case lines <= 20:
		return "11-20 lines"
	case lines <= 50:
		return "21-50 lines"
	case lines <= 100:
		return "51-100 lines"
	default:
		return "100+ lines"
	}
}

// getTokenRange returns a human-readable range for a token count.
// Ranges are threshold-aware to provide meaningful distribution.
func (p *stats) getTokenRange(tokens int) string {
	t := p.threshold
	switch {
	case tokens <= t:
		return fmt.Sprintf("%d-%d tokens", 1, t)
	case tokens <= t*2:
		return fmt.Sprintf("%d-%d tokens", t+1, t*2)
	case tokens <= t*3:
		return fmt.Sprintf("%d-%d tokens", t*2+1, t*3)
	case tokens <= t*5:
		return fmt.Sprintf("%d-%d tokens", t*3+1, t*5)
	case tokens <= t*10:
		return fmt.Sprintf("%d-%d tokens", t*5+1, t*10)
	default:
		return fmt.Sprintf("%d+ tokens", t*10+1)
	}
}

// getSeverity returns a severity level based on token count.
// Categories: small (threshold-30), medium (31-50), large (51-100), huge (100+).
func (p *stats) getSeverity(tokens int) string {
	switch {
	case tokens <= 30:
		return "small"
	case tokens <= 50:
		return "medium"
	case tokens <= 100:
		return "large"
	default:
		return "huge"
	}
}
