package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary = lipgloss.Color("#7aa2f7")
	colorSuccess = lipgloss.Color("#9ece6a")
	colorWarning = lipgloss.Color("#e0af68")
	colorError   = lipgloss.Color("#f7768e")
	colorMuted   = lipgloss.Color("#565f89")

	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	// Table
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(lipgloss.Color("#1a1b26")).
			Bold(true)

	normalRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Status styles
	statusWorking = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	statusPending = lipgloss.NewStyle().Foreground(colorWarning).Bold(true)
	statusDone    = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	statusError   = lipgloss.NewStyle().Foreground(colorError).Bold(true)

	// Log pane
	logBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted)

	logTitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)

	// Help bar
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Flash message
	flashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1b26")).
			Background(colorPrimary).
			Padding(0, 1).
			Bold(true)

	flashErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1b26")).
			Background(colorError).
			Padding(0, 1).
			Bold(true)

	// Form
	formBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	formLabelStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Width(12)

	// Muted text
	mutedStyle = lipgloss.NewStyle().Foreground(colorMuted)

	// Log viewer tool lines
	toolStyle   = lipgloss.NewStyle().Foreground(colorWarning)
	resultStyle = lipgloss.NewStyle().Foreground(colorSuccess)
)
