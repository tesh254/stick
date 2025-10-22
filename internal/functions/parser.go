package functions

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

func (Parser) Parse(input string) (string, []string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", nil, errors.New("empty input")
	}

	// Look for function call pattern anywhere in the string
	// Find the first opening parenthesis
	open := strings.Index(s, "(")
	if open == -1 {
		// If no parentheses found, check if the entire string is a function name
		if !isIdentifier(s) {
			return "", nil, fmt.Errorf("invalid function name: %q", s)
		}
		return s, []string{}, nil
	}

	// Find the function name before the opening parenthesis
	nameStart := 0
	// Look backwards from the open parenthesis to find the start of the function name
	for i := open - 1; i >= 0; i-- {
		char := rune(s[i])
		if i == 0 {
			// If we reach the beginning, the function name starts at 0
			nameStart = 0
			break
		} else if !(unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_') {
			// Found a non-identifier character, function name starts after it
			nameStart = i + 1
			break
		}
	}

	// Extract and validate the function name
	name := strings.TrimSpace(s[nameStart:open])

	// Count parentheses to handle nested ones properly
	parenCount := 0
	close := -1
	startLooking := open
	for i := startLooking; i < len(s); i++ {
		char := rune(s[i])
		if char == '(' {
			parenCount++
		} else if char == ')' {
			parenCount--
			if parenCount == 0 {
				close = i
				break
			}
		}
	}

	if open <= nameStart || close < 0 || close < open {
		return "", nil, fmt.Errorf("invalid call syntax: %q", s)
	}

	if !isIdentifier(name) {
		return "", nil, fmt.Errorf("invalid function name: %q", name)
	}

	// Extract arguments from between the parentheses
	argStr := strings.TrimSpace(s[open+1 : close])
	if argStr == "" {
		return name, []string{}, nil
	}

	args, err := splitArgs(argStr)
	if err != nil {
		return "", nil, err
	}

	return name, args, nil
}

func isIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	runes := []rune(s)
	if !(unicode.IsLetter(runes[0]) || runes[0] == '_') {
		return false
	}

	for _, r := range runes[1:] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

func splitArgs(s string) ([]string, error) {
	var out []string
	var buf strings.Builder
	inQuotes := false
	escape := false

	for i, r := range s {
		if escape {
			buf.WriteRune(r)
			escape = false
			continue
		}
		switch r {
		case '\\':
			if inQuotes {
				escape = true
			} else {
				buf.WriteRune(r)
			}
		case '"':
			inQuotes = !inQuotes
			buf.WriteRune(r)
		case ',':
			if inQuotes {
				buf.WriteRune(r)
			} else {
				token := strings.TrimSpace(buf.String())
				if token != "" {
					out = append(out, token)
				}
				buf.Reset()
			}
		default:
			buf.WriteRune(r)
		}
		// Detect unterminated quotes at end
		if i == len([]rune(s))-1 && inQuotes {
			return nil, errors.New("unterminated string literal")
		}
	}
	last := strings.TrimSpace(buf.String())
	if last != "" {
		out = append(out, last)
	}
	return out, nil
}
