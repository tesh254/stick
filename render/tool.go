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

func RenderToolResult(name, args, result string, isError bool) string {
	var header string
	if isError {
		header = toolErrorHeaderStyle.Render("Tool Error")
	} else {
		header = toolResultHeaderStyle.Render("Tool Result")
	}
	nameText := lipgloss.NewStyle().Bold(true).Render(name)
	argsText := toolCallBodyStyle.Render(args)
	resultText := toolResultBodyStyle.Render(result)
	content := fmt.Sprintf("%s %s\n%s", nameText, argsText, resultText)
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}
