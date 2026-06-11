package printer

import "github.com/LarsArtmann/art-dupl/domain"

// printRecommendations prints actionable recommendations based on the health score.
func (p *stats) printRecommendations() {
	switch p.statsData.HealthScore {
	case domain.HealthScoreA:
		p.printSuccessf("✓ Excellent code health! Duplication is minimal.")
		p.printLinef("Keep up the good work. Maintain current practices.")
	case domain.HealthScoreB:
		p.printSuccessf("✓ Good code health with minor duplication.")
		p.printLinef("Consider extracting small duplicate patterns into shared functions.")
		p.printLinef("Review the 'Top Files' section to identify problem areas.")
	case domain.HealthScoreC:
		p.printWarningf("! Moderate code duplication detected.")
		p.printLinef("Prioritize refactoring duplicate code blocks:")
		p.printLinef("  1. Focus on large clones (50+ lines) first")
		p.printLinef("  2. Create shared utility functions or base classes")
		p.printLinef("  3. Consider domain-driven design patterns")
	case domain.HealthScoreD:
		p.printWarningf("⚠ High code duplication - action needed.")
		p.printLinef("Immediate actions recommended:")
		p.printLinef("  1. Extract all medium/large duplicate blocks (>20 lines)")
		p.printLinef("  2. Implement shared libraries or services")
		p.printLinef("  3. Establish code review guidelines to prevent new duplication")
		p.printLinef("  4. Consider architectural changes (e.g., introduce new abstractions)")
	case domain.HealthScoreF:
		p.printErrorf("✗ Critical code duplication - immediate action required!")
		p.printLinef("Urgent steps to take:")
		p.printLinef("  1. Prioritize ALL duplicate code extraction immediately")
		p.printLinef("  2. Halt new feature development until duplication is reduced")
		p.printLinef("  3. Create comprehensive refactoring plan")
		p.printLinef("  4. Invest in architectural review and design patterns")
		p.printLinef("  5. Consider team training on DRY principles")
	default:
		p.printLinef("No recommendations available.")
	}

	// Additional recommendations based on specific metrics
	p.printSection("Next Steps:")

	if p.statsData.TotalCloneGroups > 10 {
		p.printBulletf(
			"You have %d clone groups - focus on the largest ones first",
			p.statsData.TotalCloneGroups,
		)
	}

	if p.statsData.AverageCloneSize > 50 {
		p.printBulletf(
			"Average clone size is %d lines - prioritize extracting large blocks",
			p.statsData.AverageCloneSize,
		)
	}

	if p.statsData.ComplexityScore > 3.0 {
		p.printBulletf(
			"Complexity score of %.2f suggests multiple clones per group - consider patterns",
			p.statsData.ComplexityScore,
		)
	}

	switch p.statsData.DetectionMode {
	case "semantic":
		p.printBulletf("Using semantic mode: clones matched by structure AND identifier names")
		p.printBulletf(
			"Run with --structural to see all structural matches (may include more results)",
		)
	case "structural":
		p.printBulletf("Using structural mode: clones matched by AST structure only")
		p.printBulletf("Run with --semantic to reduce false positives by matching identifier names")
	default:
		p.printBulletf(
			"Run with --semantic for fewer false positives or --structural for more matches",
		)
	}

	p.printBulletf("Run with --threshold 50 to focus on large duplications only")
	p.printBulletf("Use --format json for machine-readable output")
	p.printBulletf("Integrate into CI/CD pipeline for continuous monitoring")
}
