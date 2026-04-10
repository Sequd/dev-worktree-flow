package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Palette — one accent, high contrast text, quiet chrome
	colorAccent   = lipgloss.Color("#5F9FFF")
	colorText     = lipgloss.Color("#FFFFFF")
	colorBright   = lipgloss.Color("#FFFFFF")
	colorMuted    = lipgloss.Color("#A0AEC0")
	colorFaint    = lipgloss.Color("#718096")
	colorBorder   = lipgloss.Color("#4A5568")
	colorSelectBg = lipgloss.Color("#2C5282")

	colorOk   = lipgloss.Color("#68D391")
	colorWarn = lipgloss.Color("#F6C950")
	colorErr  = lipgloss.Color("#FC8181")

	// ── Title ──
	titleStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 2)

	titleDimStyle = lipgloss.NewStyle().
			Foreground(colorFaint)

	// ── List rows ──
	rowSelectedStyle = lipgloss.NewStyle().
				Foreground(colorBright).
				Background(colorSelectBg).
				Bold(true).
				Padding(0, 2)

	rowNormalStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 2)

	rowDimStyle = lipgloss.NewStyle().
			Foreground(colorFaint).
			Padding(0, 0)

	// Main worktree accent
	rowMainSelectedStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Background(colorSelectBg).
				Bold(true).
				Padding(0, 2)

	rowMainNormalStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Padding(0, 2)

	// Tags
	tagStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Background(lipgloss.Color("#1A365D")).
			Padding(0, 1)

	tagCleanStyle = lipgloss.NewStyle().
			Foreground(colorOk)

	tagDirtyStyle = lipgloss.NewStyle().
			Foreground(colorWarn)

	// ── Detail area (below list) ──
	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorFaint).
				Width(10)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorText)

	// ── Status ──
	errorStyle   = lipgloss.NewStyle().Foreground(colorErr)
	successStyle = lipgloss.NewStyle().Foreground(colorOk)
	warningStyle = lipgloss.NewStyle().Foreground(colorWarn)

	statusCleanStyle = lipgloss.NewStyle().Foreground(colorOk)
	statusDirtyStyle = lipgloss.NewStyle().Foreground(colorWarn)

	// ── Help bar ──
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorFaint)

	helpSepStyle = lipgloss.NewStyle().
			Foreground(colorBorder)

	// ── Dialog ──
	inputLabelStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 3).
			Width(60)

	// ── Misc ──
	statusBarStyle = lipgloss.NewStyle().
			Padding(0, 2)

	mainTagStyle = tagStyle // alias for compatibility

	// Branch origin tags in suggestion list
	branchTagLocal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A202C")).
			Background(colorOk).
			Bold(true).
			Padding(0, 1)

	branchTagRemote = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A202C")).
			Background(colorAccent).
			Bold(true).
			Padding(0, 1)

	branchTagBoth = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A202C")).
			Background(colorWarn).
			Bold(true).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)
)
