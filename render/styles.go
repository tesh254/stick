package render

import "github.com/charmbracelet/lipgloss"

var PurpleText = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
var BoldText = lipgloss.NewStyle().Bold(true)
var RedText = lipgloss.NewStyle().Foreground(lipgloss.Color("#E43636"))
var CreamHighlight = lipgloss.NewStyle().Background(lipgloss.Color("#F6EFD2"))
var BlueText = lipgloss.NewStyle().Foreground(lipgloss.Color("#0046FF"))
var OrangeHighlight = lipgloss.NewStyle().Background(lipgloss.Color("#FF8040"))
var GrayText = lipgloss.NewStyle().Foreground(lipgloss.Color("#B7B7B7"))
var GreenText = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))

// Define styles for the code block
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

// Define lipgloss styles for markdown elements
var (
	// Document style for overall rendering
	docStyle = lipgloss.NewStyle().
			Margin(1, 2)

	// Header styles
	h1Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("69")).                           // Bright blue
		Border(lipgloss.NormalBorder(), false, false, true, false). // Underline
		BorderForeground(lipgloss.Color("69")).
		Padding(0, 1).
		MarginBottom(1)

	h2Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("208")). // Orange
		SetString("📌 ").
		MarginBottom(1)

	h3Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")). // Yellow
		SetString("➤ ").
		MarginBottom(1)

	// Inline styles
	boldStyle   = lipgloss.NewStyle().Bold(true)
	italicStyle = lipgloss.NewStyle().Italic(true)
	linkStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("27")). // Blue
			Underline(true)

	// Block quote with vertical line prefix
	quoteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("57")).                           // Purple
			Border(lipgloss.NormalBorder(), false, true, false, false). // Vertical line on left
			BorderForeground(lipgloss.Color("57")).
			PaddingLeft(1).
			MarginLeft(1)

	// Inline code style
	codeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")). // Reddish
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	// List style
	listItemStyle = lipgloss.NewStyle().
			SetString("• " + "%s"). // Creative bullet
			MarginLeft(2)

	// Horizontal rule
	hrStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Gray
		Margin(1, 0).
		Render("━━━━━")
)
