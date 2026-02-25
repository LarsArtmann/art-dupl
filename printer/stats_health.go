package printer

// calculateHealthScore calculates an A-F grade based on duplication, complexity, and impact metrics.
func (p *stats) calculateHealthScore() string {
	if p.statsData.DuplicationRatio == 0 && p.statsData.ComplexityScore == 0 && p.statsData.ImpactScore == 0 {
		return "A"
	}

	// Normalize metrics to 0-100 scale (higher is worse)
	var duplicationScore float64
	if p.statsData.TotalEstimatedLines > 0 {
		duplicationScore = p.statsData.DuplicationRatio
	}

	// Complexity score: higher = worse. Normalize to 0-100 assuming max complexity of 10.0
	complexityScore := (p.statsData.ComplexityScore / 10.0) * 100
	if complexityScore > 100 {
		complexityScore = 100
	}

	// Impact score: higher = worse. Normalize to 0-100 assuming max impact of 10000
	impactScore := (float64(p.statsData.ImpactScore) / 10000.0) * 100
	if impactScore > 100 {
		impactScore = 100
	}

	// Weighted average: duplication 60%, complexity 25%, impact 15%
	totalScore := duplicationScore*0.6 + complexityScore*0.25 + impactScore*0.15

	// Convert to A-F grade based on total score
	switch {
	case totalScore < 3:
		return "A"
	case totalScore < 6:
		return "B"
	case totalScore < 10:
		return "C"
	case totalScore < 15:
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
func (p *stats) getTokenRange(tokens int) string {
	switch {
	case tokens <= 15:
		return "1-15 tokens"
	case tokens <= 30:
		return "16-30 tokens"
	case tokens <= 50:
		return "31-50 tokens"
	case tokens <= 100:
		return "51-100 tokens"
	case tokens <= 200:
		return "101-200 tokens"
	default:
		return "200+ tokens"
	}
}
