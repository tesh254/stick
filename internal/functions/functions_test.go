package functions

import (
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		hasError bool
	}{
		{
			name:     "basic addition",
			args:     []string{"1", "2"},
			expected: "3",
			hasError: false,
		},
		{
			name:     "addition with float",
			args:     []string{"1.5", "2.7"},
			expected: "4.2",
			hasError: false,
		},
		{
			name:     "addition with negative numbers",
			args:     []string{"-1", "2"},
			expected: "1",
			hasError: false,
		},
		{
			name:     "first argument only",
			args:     []string{"5"},
			expected: "5",
			hasError: false,
		},
		{
			name:     "second argument only",
			args:     []string{"", "3"},
			expected: "3",
			hasError: false,
		},
		{
			name:     "no arguments",
			args:     []string{},
			expected: "0",
			hasError: false,
		},
		{
			name:     "one empty argument",
			args:     []string{"", "5"},
			expected: "5",
			hasError: false,
		},
		{
			name:     "both empty arguments",
			args:     []string{"", ""},
			expected: "0",
			hasError: false,
		},
		{
			name:     "float result with trailing zeros",
			args:     []string{"1.0", "2.0"},
			expected: "3",
			hasError: false,
		},
		{
			name:     "large numbers",
			args:     []string{"1000000", "2000000"},
			expected: "3e+06", // %g formatter will use scientific notation for large integers
			hasError: false,
		},
		{
			name:     "decimal precision",
			args:     []string{"0.1", "0.2"},
			expected: "0.30000000000000004", // Due to floating-point precision
			hasError: false,
		},
		{
			name:     "invalid first argument",
			args:     []string{"abc", "2"},
			expected: "",
			hasError: true,
		},
		{
			name:     "invalid second argument",
			args:     []string{"1", "xyz"},
			expected: "",
			hasError: true,
		},
		{
			name:     "both invalid arguments",
			args:     []string{"abc", "xyz"},
			expected: "",
			hasError: true,
		},
		{
			name:     "mixed valid and invalid",
			args:     []string{"1", "invalid"},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Add(tt.args)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{
			input:    "123",
			expected: 123.0,
			hasError: false,
		},
		{
			input:    "123.45",
			expected: 123.45,
			hasError: false,
		},
		{
			input:    "-45.67",
			expected: -45.67,
			hasError: false,
		},
		{
			input:    "  789  ",
			expected: 789.0,
			hasError: false,
		},
		{
			input:    "\"123\"",
			expected: 123.0,
			hasError: false,
		},
		{
			input:    "invalid",
			expected: 0.0,
			hasError: true,
		},
		{
			input:    "",
			expected: 0.0, // Empty string should return 0 based on the function spec
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result, err := parseNumber(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for input %q but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				} else if result != tt.expected {
					t.Errorf("for input %q: expected %f, got %f", tt.input, tt.expected, result)
				}
			}
		})
	}
}

func TestUnquoteIfQuoted(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `"hello"`,
			expected: `hello`, // Function removes the quotes after unquoting
		},
		{
			input:    `  "world"  `,
			expected: `world`, // Function removes quotes and trims whitespace
		},
		{
			input:    "unquoted",
			expected: "unquoted",
		},
		{
			input:    `"quoted with spaces"`,
			expected: `quoted with spaces`,
		},
		{
			input:    `"quote with \"escaped\" content"`,
			expected: `quote with "escaped" content`,
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    `"`,
			expected: `"`, // Unterminated quote returns as is
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := unquoteIfQuoted(tt.input)
			if result != tt.expected {
				t.Errorf("for input %q: expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	p := Parser{}
	
	tests := []struct {
		name     string
		input    string
		expectedName string
		expectedArgs []string
		hasError bool
	}{
		{
			name:     "valid function call",
			input:    "add(2, 2)",
			expectedName: "add",
			expectedArgs: []string{"2", "2"},
			hasError: false,
		},
		{
			name:     "function with extra text",
			input:    "add(2, 2) please",
			expectedName: "add",
			expectedArgs: []string{"2", "2"},
			hasError: false, // This should now succeed as we allow trailing text
		},
		{
			name:     "function with wrong bracket",
			input:    "add{2, 2) please",
			expectedName: "",
			expectedArgs: nil,
			hasError: true, // This should fail due to invalid syntax
		},
		{
			name:     "no arguments",
			input:    "add()",
			expectedName: "add",
			expectedArgs: []string{},
			hasError: false,
		},
		{
			name:     "empty input",
			input:    "",
			expectedName: "",
			expectedArgs: nil,
			hasError: true,
		},
		{
			name:     "function without parentheses",
			input:    "add",
			expectedName: "add",
			expectedArgs: []string{},
			hasError: false,
		},
		{
			name:     "malformed parentheses",
			input:    "add(2, 2",
			expectedName: "",
			expectedArgs: nil,
			hasError: true,
		},
		{
			name:     "unterminated string in args",
			input:    `add("hello)`,
			expectedName: "",
			expectedArgs: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, args, err := p.Parse(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else {
					if name != tt.expectedName {
						t.Errorf("expected name %q, got %q", tt.expectedName, name)
					}
					if len(args) != len(tt.expectedArgs) {
						t.Errorf("expected %d args, got %d", len(tt.expectedArgs), len(args))
					} else {
						for i, arg := range args {
							if arg != tt.expectedArgs[i] {
								t.Errorf("expected arg[%d] %q, got %q", i, tt.expectedArgs[i], arg)
							}
						}
					}
				}
			}
		})
	}
}

func TestCompleteFunctionFlow(t *testing.T) {
	// Test the complete flow: parsing a string and executing the function
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "complete add function call",
			input:    "add(2, 2)",
			expected: "4",
			hasError: false,
		},
		{
			name:     "complete add function call with floats",
			input:    "add(1.5, 2.5)",
			expected: "4",
			hasError: false,
		},
		{
			name:     "complete add function call with one arg",
			input:    "add(5)",
			expected: "5",
			hasError: false,
		},
		{
			name:     "complete add function call with no args",
			input:    "add()",
			expected: "0",
			hasError: false,
		},
		{
			name:     "valid function call with trailing text",
			input:    "add(2, 2) please",
			expected: "4",
			hasError: false,
		},
		{
			name:     "malformed function call",
			input:    "add{2, 2) please",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input string to extract function name and arguments
			p := Parser{}
			name, args, err := p.Parse(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error during parsing but got none")
				}
				return // Exit early if we expected an error during parsing
			}
			
			if err != nil {
				t.Errorf("unexpected error during parsing: %v", err)
				return
			}
			
			// Create a registry and register the function
			r := NewRegistry()
			r.Register("add", Add, 0, 2)
			
			// Call the function using the registry
			result, err := r.Call(name, args)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error during function call but got none, result was: %s", result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error during function call: %v", err)
				} else if result != tt.expected {
					t.Errorf("expected result %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestRegistry_Call(t *testing.T) {
	r := NewRegistry()
	r.Register("add", Add, 0, 2) // Register the add function

	tests := []struct {
		name     string
		function string
		args     []string
		expected string
		hasError bool
	}{
		{
			name:     "valid add call",
			function: "add",
			args:     []string{"2", "2"},
			expected: "4",
			hasError: false,
		},
		{
			name:     "add with no args",
			function: "add",
			args:     []string{},
			expected: "0",
			hasError: false,
		},
		{
			name:     "unknown function",
			function: "subtract",
			args:     []string{"2", "1"},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := r.Call(tt.function, tt.args)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestUnknownFunction(t *testing.T) {
	r := NewRegistry()
	// Only register add, not other functions
	
	// Test calling an unregistered function
	result, err := r.Call("unknownFunction", []string{})
	
	if err == nil {
		t.Errorf("expected error for unknown function, but got none")
	} else {
		expectedError := "unknown function: unknownFunction"
		if err.Error() != expectedError {
			t.Errorf("expected error %q, got %q", expectedError, err.Error())
		}
	}
	
	if result != "" {
		t.Errorf("expected empty result for unknown function, got %q", result)
	}
}

func TestEcho(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		hasError bool
	}{
		{
			name:     "echo with single argument",
			args:     []string{"hello"},
			expected: "hello",
			hasError: false,
		},
		{
			name:     "echo with multiple arguments",
			args:     []string{"hello", "world"},
			expected: "hello world",
			hasError: false,
		},
		{
			name:     "echo with no arguments",
			args:     []string{},
			expected: "",
			hasError: false,
		},
		{
			name:     "echo with multiple words",
			args:     []string{"this", "is", "a", "test"},
			expected: "this is a test",
			hasError: false,
		},
		{
			name:     "echo with empty strings",
			args:     []string{"", "hello", ""},
			expected: " hello ",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Echo(tt.args)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}