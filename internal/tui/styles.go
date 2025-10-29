package tui

import "github.com/charmbracelet/lipgloss"

var (
	YellowTxt = lipgloss.NewStyle().Foreground(lipgloss.Color("#F4991A"))
	YellowBg  = lipgloss.NewStyle().Background(lipgloss.Color("#F4991A"))
	YellowHex = "#F4991A"
)

var (
	codeBlockStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")). // Purple border
			Background(lipgloss.Color("235")).      // Dark gray background
			Foreground(lipgloss.Color("252")).      // Light gray text
			Padding(1, 2).
			Margin(1, 0)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")). // Pinkish header
			PaddingBottom(1)
)

var (
	toolCallHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("208")). // Orange
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("208")).
				Padding(0, 1).
				MarginBottom(1)

	toolCallBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")). // Gray
				Padding(0, 1)

	toolResultHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("69")). // Bright blue
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("69")).
				Padding(0, 1).
				MarginBottom(1)

	toolResultBodyStyle = lipgloss.NewStyle().
				Padding(0, 1)

	toolErrorHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196")). // Red
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Padding(0, 1).
				MarginBottom(1)
)
