package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderCodeBlock formats content as a styled code block and logs it
func RenderCodeBlock(content string) string {
	// Format the content as a code block
	header := headerStyle.Render("Code Block")
	codeContent := codeBlockStyle.Render(strings.TrimSpace(content))
	return lipgloss.JoinVertical(lipgloss.Left, header, codeContent)
}
