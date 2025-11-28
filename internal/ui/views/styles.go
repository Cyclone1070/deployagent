package views

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	ColorPrimary = lipgloss.Color("63")  // Blue
	ColorSuccess = lipgloss.Color("42")  // Green
	ColorError   = lipgloss.Color("196") // Red
	ColorWarning = lipgloss.Color("214") // Orange
	ColorMuted   = lipgloss.Color("240") // Gray
	ColorPurple  = lipgloss.Color("141") // Purple

	// Styles
	UserMessageStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	AssistantMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	StatusExecutingStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	StatusDoneStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StatusThinkingStyle = lipgloss.NewStyle().
				Foreground(ColorPurple)

	StatusDefaultStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	PermissionBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorWarning).
				Padding(1, 2).
				Width(60)

	DiffContextStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	DiffAddStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	DiffRemoveStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)
