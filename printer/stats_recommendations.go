package printer

import "fmt"

// printRecommendations prints actionable recommendations based on the health score.
func (p *stats) printRecommendations() {
	switch p.statsData.HealthScore {
	case "A":
		_, _ = fmt.Fprintf(p.w, "%s\n", p.success.Render("✓ Excellent code health! Duplication is minimal."))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("Keep up the good work. Maintain current practices."))
	case "B":
		_, _ = fmt.Fprintf(p.w, "%s\n", p.success.Render("✓ Good code health with minor duplication."))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("Consider extracting small duplicate patterns into shared functions."))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("Review the 'Top Files' section to identify problem areas."))
	case "C":
		_, _ = fmt.Fprintf(p.w, "%s\n", p.warning.Render("! Moderate code duplication detected."))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("Prioritize refactoring duplicate code blocks:"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  1. Focus on large clones (50+ lines) first"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  2. Create shared utility functions or base classes"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  3. Consider domain-driven design patterns"))
	case "D":
		_, _ = fmt.Fprintf(p.w, "%s\n", p.warning.Render("⚠ High code duplication - action needed."))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("Immediate actions recommended:"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  1. Extract all medium/large duplicate blocks (>20 lines)"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  2. Implement shared libraries or services"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  3. Establish code review guidelines to prevent new duplication"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  4. Consider architectural changes (e.g., introduce new abstractions)"))
	case "F":
		_, _ = fmt.Fprintf(p.w, "%s\n", p.error.Render("✗ Critical code duplication - immediate action required!"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("Urgent steps to take:"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  1. Prioritize ALL duplicate code extraction immediately"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  2. Halt new feature development until duplication is reduced"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  3. Create comprehensive refactoring plan"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  4. Invest in architectural review and design patterns"))
		_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("  5. Consider team training on DRY principles"))
	default:
		_, _ = fmt.Fprintf(p.w, "%s\n", p.base.Render("No recommendations available."))
	}

	// Additional recommendations based on specific metrics
	_, _ = fmt.Fprintf(p.w, "\n%s\n", p.section.Render("Next Steps:"))

	if p.statsData.TotalCloneGroups > 10 {
		_, _ = fmt.Fprintf(p.w, "  • %s\n", p.base.Render(fmt.Sprintf("You have %d clone groups - focus on the largest ones first", p.statsData.TotalCloneGroups)))
	}

	if p.statsData.AverageCloneSize > 50 {
		_, _ = fmt.Fprintf(p.w, "  • %s\n", p.base.Render(fmt.Sprintf("Average clone size is %d lines - prioritize extracting large blocks", p.statsData.AverageCloneSize)))
	}

	if p.statsData.ComplexityScore > 3.0 {
		_, _ = fmt.Fprintf(p.w, "  • %s\n", p.base.Render(fmt.Sprintf("Complexity score of %.2f suggests multiple clones per group - consider patterns", p.statsData.ComplexityScore)))
	}

	_, _ = fmt.Fprintf(p.w, "  • %s\n", p.base.Render("Run with --threshold 50 to focus on large duplications only"))
	_, _ = fmt.Fprintf(p.w, "  • %s\n", p.base.Render("Use --format json for machine-readable output"))
	_, _ = fmt.Fprintf(p.w, "  • %s\n", p.base.Render("Integrate into CI/CD pipeline for continuous monitoring"))
}
