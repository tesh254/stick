package functions

import (
	"fmt"
	"strconv"
	"strings"
)

// Add: args ["1","2"] -> sum numerically; missing args default to 0
func Add(args []string) (string, error) {
	var a, b float64
	var err error

	if len(args) >= 1 {
		a, err = parseNumber(args[0])
		if err != nil {
			return "", fmt.Errorf("add arg1: %w", err)
		}
	}
	if len(args) >= 2 {
		b, err = parseNumber(args[1])
		if err != nil {
			return "", fmt.Errorf("add arg2: %w", err)
		}
	}
	return fmt.Sprintf("%g", a+b), nil
}

func parseNumber(s string) (float64, error) {
	s = strings.TrimSpace(unquoteIfQuoted(s))
	if s == "" {
		return 0, nil // Return 0 for empty strings as per the function's spec
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	return f, nil
}

func unquoteIfQuoted(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		// Check for double quotes
		if s[0] == '"' && s[len(s)-1] == '"' {
			u, err := strconv.Unquote(s)
			if err == nil {
				return u
			}
		} else if s[0] == '\'' && s[len(s)-1] == '\'' { // Check for single quotes - simple removal
			return s[1 : len(s)-1] // Remove first and last characters
		}
	}
	return s
}

// Echo: args ["text1", "text2", ...] -> returns concatenated text with quotes removed from arguments
func Echo(args []string) (string, error) {
	// Remove quotes from each argument
	unquotedArgs := make([]string, len(args))
	for i, arg := range args {
		unquotedArgs[i] = unquoteIfQuoted(arg)
	}
	return strings.Join(unquotedArgs, " "), nil
}
