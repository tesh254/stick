package render

import (
	"log"
	"strings"
)

// RenderMarkdown parses markdown content and returns a styled string using lipgloss
func RenderMarkdown(content string) string {
	// Log the content
	log.Printf("Rendering markdown content:\n%s\n", content)

	lines := strings.Split(content, "\n")
	var output []string
	var inCodeBlock bool
	var codeBlockContent strings.Builder
	var inList bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				continue
			} else {
				inCodeBlock = false
				// Render collected code block
				output = append(output, codeBlockStyle.Render(codeBlockContent.String()))
				codeBlockContent.Reset()
				continue
			}
		}
		if inCodeBlock {
			codeBlockContent.WriteString(line + "\n")
			continue
		}

		// Handle empty lines
		if trimmed == "" {
			output = append(output, "")
			inList = false
			continue
		}

		// Handle headers
		if strings.HasPrefix(trimmed, "# ") {
			output = append(output, h1Style.Render(strings.TrimPrefix(trimmed, "# ")))
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			output = append(output, h2Style.Render(strings.TrimPrefix(trimmed, "## ")))
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			output = append(output, h3Style.Render(strings.TrimPrefix(trimmed, "### ")))
			continue
		}

		// Handle block quotes
		if strings.HasPrefix(trimmed, "> ") {
			output = append(output, quoteStyle.Render(strings.TrimPrefix(trimmed, "> ")))
			continue
		}

		// Handle horizontal rules
		if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "***") {
			output = append(output, hrStyle)
			continue
		}

		// Handle lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			inList = true
			output = append(output, listItemStyle.Render(strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")))
			continue
		}

		// Handle inline formatting for non-list lines
		if !inList {
			line = applyInlineStyles(trimmed)
			output = append(output, line)
		} else {
			// Continue list item with indentation
			output = append(output, listItemStyle.Render(trimmed))
		}
	}

	// Join output with newlines and apply document style
	return docStyle.Render(strings.Join(output, "\n"))
}

// applyInlineStyles handles inline markdown (bold, italic, code, links)
func applyInlineStyles(line string) string {
	// Handle links: [text](url)
	for {
		start := strings.Index(line, "[")
		if start == -1 {
			break
		}
		endText := strings.Index(line[start:], "]")
		if endText == -1 {
			break
		}
		endText += start
		if endText+1 >= len(line) || line[endText+1] != '(' {
			break
		}
		endURL := strings.Index(line[endText:], ")")
		if endURL == -1 {
			break
		}
		endURL += endText
		text := line[start+1 : endText]
		url := line[endText+2 : endURL]
		styledLink := linkStyle.Render(text)
		line = line[:start] + styledLink + " (" + url + ")" + line[endURL+1:]
	}

	// Handle bold: **text** or __text__
	for {
		start := strings.Index(line, "**")
		if start == -1 {
			start = strings.Index(line, "__")
		}
		if start == -1 {
			break
		}
		end := strings.Index(line[start+2:], "**")
		if end == -1 {
			end = strings.Index(line[start+2:], "__")
		}
		if end == -1 {
			break
		}
		end += start + 2
		boldText := boldStyle.Render(line[start+2 : end])
		line = line[:start] + boldText + line[end+2:]
	}

	// Handle italic: *text* or _text_
	for {
		start := strings.Index(line, "*")
		if start == -1 {
			start = strings.Index(line, "_")
		}
		if start == -1 || start+1 >= len(line) {
			break
		}
		end := strings.Index(line[start+1:], "*")
		if end == -1 {
			end = strings.Index(line[start+1:], "_")
		}
		if end == -1 {
			break
		}
		end += start + 1
		italicText := italicStyle.Render(line[start+1 : end])
		line = line[:start] + italicText + line[end+1:]
	}

	// Handle inline code: `code`
	for {
		start := strings.Index(line, "`")
		if start == -1 {
			break
		}
		end := strings.Index(line[start+1:], "`")
		if end == -1 {
			break
		}
		end += start + 1
		codeText := codeStyle.Render(line[start+1 : end])
		line = line[:start] + codeText + line[end+1:]
	}

	return line
}
