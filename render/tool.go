package render

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func RenderToolCall(name, args string) string {
	header := toolCallHeaderStyle.Render("Running Tool")
	nameText := lipgloss.NewStyle().Bold(true).Render(name)
	argsText := toolCallBodyStyle.Render(args)
	content := fmt.Sprintf("%s %s", nameText, argsText)
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func RenderToolResult(name, result string, isError bool) string {
	var header string
	if isError {
		header = toolErrorHeaderStyle.Render("Tool Error")
	} else {
		header = toolResultHeaderStyle.Render("Tool Result")
	}
	nameText := lipgloss.NewStyle().Bold(true).Render(name)
	resultText := toolResultBodyStyle.Render(result)
	content := fmt.Sprintf("%s\n%s", nameText, resultText)
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}
