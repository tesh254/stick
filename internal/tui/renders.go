package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func renderFunctionName(name, args string) string {
	content := fmt.Sprintf("function call\n\n[INPUT]\nname => %s\nargs => %s", GreenText.Underline(true).Render(name), GreenText.Underline(true).Render(args))
	contentText := toolCallStyle.Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, contentText)
}

func renderFunctionOrToolResult(_, _, result string, isError bool) string {
	var header string
	content := fmt.Sprintf("[OUTPUT]\n %s", result)
	if isError {
		header = toolErrorStyle.Render("⛔️ error\n\n", content)
	} else {
		header = toolResultStyle.Render("✅ success\n\n", content)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header)
}
