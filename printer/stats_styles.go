package printer

import (
	"os"

	"charm.land/lipgloss/v2"
)

// StyleMixin holds common lipgloss style fields.
type StyleMixin struct {
	base    lipgloss.Style
	header  lipgloss.Style
	section lipgloss.Style
	metric  lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	error   lipgloss.Style
}

// styleConfig holds all lipgloss styles.
type styleConfig struct {
	StyleMixin
}

// initStyles initializes lipgloss styles, respecting NO_COLOR environment variable.
func initStyles() styleConfig {
	// Check for NO_COLOR environment variable
	noColor := os.Getenv("NO_COLOR") != ""

	// Create styles
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	sectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00E676")).Bold(true)
	metricStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#738ADB"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00C853"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E53E3"))

	// Disable all colors if NO_COLOR is set
	if noColor {
		return styleConfig{
			StyleMixin: StyleMixin{
				base:    lipgloss.NewStyle(),
				header:  lipgloss.NewStyle(),
				section: lipgloss.NewStyle(),
				metric:  lipgloss.NewStyle(),
				success: lipgloss.NewStyle(),
				warning: lipgloss.NewStyle(),
				error:   lipgloss.NewStyle(),
			},
		}
	}

	return styleConfig{
		StyleMixin: StyleMixin{
			base:    lipgloss.NewStyle().Bold(true),
			header:  headerStyle,
			section: sectionStyle,
			metric:  metricStyle,
			success: successStyle,
			warning: warningStyle,
			error:   errorStyle,
		},
	}
}

// healthScoreStyle returns the appropriate style for a health score grade.
func (p *stats) healthScoreStyle(grade string) lipgloss.Style {
	switch grade {
	case "A":
		return p.success
	case "B":
		return p.success
	case "C":
		return p.warning
	case "D":
		return p.warning
	case "F":
		return p.error
	default:
		return p.base
	}
}
