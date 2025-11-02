package tui

import "github.com/charmbracelet/lipgloss"

var (
	YellowTxt = lipgloss.NewStyle().Foreground(lipgloss.Color("#F4991A"))
	YellowBg  = lipgloss.NewStyle().Background(lipgloss.Color("#F4991A"))
	YellowHex = "#F4991A"
	GreenText = lipgloss.NewStyle().Foreground(lipgloss.Color("#344F1F"))
	GreenBg   = lipgloss.NewStyle().Background(lipgloss.Color("#344F1F"))
	GreenHex  = "#344F1F"
	RedText   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E62727"))
	RedBg     = lipgloss.NewStyle().Background(lipgloss.Color("#E62727"))
	RedHex    = "#E62727"
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

// Tool styles
var (
	toolCallStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(YellowHex)). // Orange
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(YellowHex)).
			Padding(1, 2).
			MarginBottom(0)

	toolResultStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(GreenHex)). // Green
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(GreenHex)).
			Padding(1, 2).
			MarginBottom(0)

	toolErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(RedHex)). // Red
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(RedHex)).
			Padding(1, 2).
			MarginBottom(0)
)
