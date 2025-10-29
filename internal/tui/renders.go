package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func renderFunctionName(name, args string) string {
	header := toolCallHeaderStyle.Render(fmt.Sprintf("Running Function: %s", name))
	nameText := lipgloss.NewStyle().Bold(true).Render(name)
	argsText := toolCallBodyStyle.Render(args)
	content := fmt.Sprintf("%s %s", nameText, argsText)
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func renderFunctionOrToolResult(name, args, result string, isError bool) string {
	var header string
	if isError {
		header = toolErrorHeaderStyle.Render("Function Error")
	} else {
		header = toolResultBodyStyle.Render("Function Result")
	}

	nameText := lipgloss.NewStyle().Bold(true).Render(name)
	argsText := toolCallBodyStyle.Render(args)
	resultText := toolResultBodyStyle.Render(result)
	content := fmt.Sprintf("%s %s\n%s", nameText, argsText, resultText)
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}
