package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Accent colors
	colorPrimary   = lipgloss.Color("#b47eff") // purple
	colorSecondary = lipgloss.Color("#ff79c6") // pink
	colorSuccess   = lipgloss.Color("#50fa7b") // green
	colorWarning   = lipgloss.Color("#f1fa8c") // yellow
	colorError     = lipgloss.Color("#ff5555") // red
	colorDim       = lipgloss.Color("#6272a4") // dim gray
	colorSubtle    = lipgloss.Color("#44475a") // subtle bg

	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(colorPrimary).
			Padding(0, 1)

	// Section headers
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// Normal text
	textStyle = lipgloss.NewStyle()

	// Dimmed / secondary text
	dimStyle = lipgloss.NewStyle().Foreground(colorDim)

	// Error text
	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	// Success text
	successStyle = lipgloss.NewStyle().Foreground(colorSuccess)

	// Highlighted / selected item
	selectedStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	// Help bar at bottom
	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)

	// Status line (e.g., "Playing: ...")
	statusStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	// Menu item (normal)
	menuItemStyle = lipgloss.NewStyle().PaddingLeft(2)

	// Menu item (selected)
	menuSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(colorSecondary).
				Bold(true)
)
