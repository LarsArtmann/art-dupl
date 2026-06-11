package printer

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
)

const (
	sizeRange1to5   = "1-5 lines"
	sizeRange6to10  = "6-10 lines"
	sizeRange11to20 = "11-20 lines"
	sizeRange21to50 = "21-50 lines"
	healthSmall     = "small"
	healthMedium    = "medium"
	healthLarge     = "large"
	healthHuge      = "huge"
)

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
func (p *stats) calculateHealthScore() domain.HealthScore {
	if p.isEmptyStats() {
		return domain.HealthScoreA
	}

	totalScore := p.calculateTotalHealthScore()

	return scoreToGrade(totalScore)
}

func (p *stats) isEmptyStats() bool {
	return p.statsData.DuplicationRatio == 0 &&
		p.statsData.ComplexityScore == 0 &&
		p.statsData.ImpactScore == 0
}

func (p *stats) calculateTotalHealthScore() float64 {
	duplicationScore := p.statsData.DuplicationRatio
	complexityScore := min((p.statsData.ComplexityScore/10.0)*20, 20)
	impactScore := min((float64(p.statsData.ImpactScore)/10000.0)*10, 10)

	return duplicationScore*0.7 + complexityScore*0.2 + impactScore*0.1
}

func scoreToGrade(score float64) domain.HealthScore {
	thresholds := []struct {
		limit float64
		grade domain.HealthScore
	}{
		{5, domain.HealthScoreA},
		{10, domain.HealthScoreB},
		{15, domain.HealthScoreC},
		{25, domain.HealthScoreD},
	}

	for _, t := range thresholds {
		if score < t.limit {
			return t.grade
		}
	}

	return domain.HealthScoreF
}

// getSizeRange returns a human-readable size range for a line count.
func (p *stats) getSizeRange(lines int) string {
	switch {
	case lines <= 5:
		return sizeRange1to5
	case lines <= 10:
		return sizeRange6to10
	case lines <= 20:
		return sizeRange11to20
	case lines <= 50:
		return sizeRange21to50
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
	multipliers := []struct {
		limit     int
		startMult int
		endMult   int
	}{
		{t, 0, 1},
		{t * 2, 1, 2},
		{t * 3, 2, 3},
		{t * 5, 3, 5},
		{t * 10, 5, 10},
	}

	for _, m := range multipliers {
		if tokens <= m.limit {
			return p.tokenRangeStr(t*m.startMult+1, t*m.endMult)
		}
	}

	return fmt.Sprintf("%d+ tokens", t*10+1)
}

// tokenRangeStr returns a formatted token range string.
func (p *stats) tokenRangeStr(start, end int) string {
	return fmt.Sprintf("%d-%d tokens", start, end)
}

// getSeverity returns a severity level based on token count relative to threshold.
// Categories: small (<=threshold*2), medium (<=threshold*5), large (<=threshold*10), huge (>threshold*10).
func (p *stats) getSeverity(tokens int) string {
	t := p.threshold
	switch {
	case tokens <= t*2:
		return healthSmall
	case tokens <= t*5:
		return healthMedium
	case tokens <= t*10:
		return healthLarge
	default:
		return healthHuge
	}
}
